// Command api 启动婚恋市场计算器后端 HTTP 服务。
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"wmc/backend/internal/admin"
	"wmc/backend/internal/auth"
	"wmc/backend/internal/baseline"
	"wmc/backend/internal/calculator"
	"wmc/backend/internal/discovery"
	"wmc/backend/internal/interaction"
	"wmc/backend/internal/meta"
	"wmc/backend/internal/photo"
	"wmc/backend/internal/platform/config"
	"wmc/backend/internal/platform/dbx"
	"wmc/backend/internal/platform/httpx"
	"wmc/backend/internal/platform/logx"
	"wmc/backend/internal/platform/metrics"
	"wmc/backend/internal/platform/ratelimit"
	"wmc/backend/internal/platform/redisx"
	"wmc/backend/internal/settings"
	"wmc/backend/internal/sms"
	"wmc/backend/internal/storage"
	"wmc/backend/internal/store"
	"wmc/backend/internal/user"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "启动失败:", err)
		os.Exit(1)
	}
}

type services struct {
	auth     *auth.Handler
	calc     *calculator.Handler
	meta     *meta.Handler
	baseline *baseline.Handler
	user      *user.Handler
	photo     *photo.Handler
	admin       *admin.Handler
	discovery   *discovery.Handler
	interaction *interaction.Handler
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	logger := logx.New(cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if cfg.DatabaseURL == "" {
		return errors.New("必须配置 DATABASE_URL（可使用 scripts/dev-db.ps1 启动本地数据库）")
	}
	pool, err := dbx.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if cfg.MigrateOnStart {
		if err := dbx.Migrate(ctx, pool, logger); err != nil {
			return err
		}
	}

	st := store.New(pool)
	kv, err := redisx.New(ctx, cfg.RateLimitStore, cfg.RedisURL)
	if err != nil {
		return err
	}
	defer kv.Close()
	if kv.Kind() == "memory" && cfg.RedisURL != "" {
		logger.Warn("Redis 不可用，已降级为内存实现（仅适用于开发环境）")
	}

	obj, fsStore, err := newStorage(ctx, cfg)
	if err != nil {
		return err
	}

	settingsSvc := settings.NewService(st, 15*time.Second)
	tokens := auth.NewTokenManager(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	authSvc := auth.NewService(st, tokens, cfg.PhoneHashSecret, cfg.PhoneEncryptionKey)
	smsProvider, err := sms.New(cfg.SMSProvider, cfg.SMSMockCode, logger)
	if err != nil {
		return err
	}
	smsSvc := auth.NewSMSService(authSvc, st, smsProvider, kv, settingsSvc)

	calcSvc := calculator.NewService(st, pool, settingsSvc, kv)
	discoverySvc := discovery.NewService(st, pool, settingsSvc, obj)
	metaSvc := meta.NewService(st, settingsSvc)
	baselineSvc := baseline.NewService(st, pool, settingsSvc, string(cfg.Env))
	userSvc := user.NewService(st, settingsSvc)
	photoSvc := photo.NewService(st, obj)
	adminSvc := admin.NewService(st, settingsSvc, obj, cfg.PhoneEncryptionKey)
	interactionSvc := interaction.NewService(st, settingsSvc, obj)
	limiter := ratelimit.New(kv)

	svc := &services{
		auth:     auth.NewHandler(authSvc, smsSvc, cfg.AdminDevLoginEnabled),
		calc:     calculator.NewHandler(calcSvc),
		meta:     meta.NewHandler(metaSvc),
		baseline: baseline.NewHandler(baselineSvc),
		user:     user.NewHandler(userSvc),
		photo:    photo.NewHandler(photoSvc),
		admin:    admin.NewHandler(adminSvc),
		discovery:   discovery.NewHandler(discoverySvc),
		interaction: interaction.NewHandler(interactionSvc, st),
	}

	m := metrics.New()
	m.AttachPool(pool)

	router := buildRouter(cfg, logger, pool, kv, m, tokens, authSvc, svc, limiter, fsStore)
	_ = settingsSvc

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      90 * time.Second,
		IdleTimeout:       90 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("HTTP 服务启动", "addr", cfg.HTTPAddr, "env", string(cfg.Env))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info("收到退出信号，开始优雅关闭")
	case err := <-errCh:
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func newStorage(ctx context.Context, cfg *config.Config) (storage.ObjectStorage, *storage.FSStorage, error) {
	if cfg.StorageDriver == "s3" {
		s3Store, err := storage.NewS3(ctx, storage.S3Options{
			Endpoint: cfg.S3Endpoint, Region: cfg.S3Region, Bucket: cfg.S3Bucket,
			AccessKey: cfg.S3AccessKey, SecretKey: cfg.S3SecretKey, UseSSL: cfg.S3UseSSL,
			ForcePathStyle: true, PublicURL: cfg.S3PublicURL,
		})
		if err != nil {
			return nil, nil, err
		}
		return s3Store, nil, nil
	}
	fsStore, err := storage.NewFS(cfg.StorageFSRoot, "http://127.0.0.1"+cfg.HTTPAddr, cfg.PhoneHashSecret)
	if err != nil {
		return nil, nil, err
	}
	return fsStore, fsStore, nil
}

func buildRouter(
	cfg *config.Config,
	logger *slog.Logger,
	pool *pgxpool.Pool,
	kv redisx.Store,
	m *metrics.Metrics,
	tokens *auth.TokenManager,
	authSvc *auth.Service,
	svc *services,
	limiter *ratelimit.Limiter,
	fsStore *storage.FSStorage,
) http.Handler {
	r := chi.NewRouter()
	r.Use(httpx.RequestID)
	r.Use(httpx.AccessLog(logger))
	r.Use(httpx.Recover)
	r.Use(m.Middleware)

	r.Get("/health/live", func(w http.ResponseWriter, req *http.Request) {
		httpx.OK(w, req, map[string]string{"status": "ok"})
	})
	r.Get("/health/ready", func(w http.ResponseWriter, req *http.Request) {
		dbOK := dbx.Ping(req.Context(), pool) == nil
		kvOK := kv.Ping(req.Context()) == nil
		m.SetChecks(dbOK, kvOK)
		if !dbOK || !kvOK {
			httpx.Fail(w, req, fmt.Errorf("依赖未就绪"))
			return
		}
		httpx.OK(w, req, map[string]string{"status": "ready", "database": "ok", "redis": kv.Kind()})
	})
	r.Handle("/metrics", m.Handler())

	r.Route("/api/v1", func(api chi.Router) {
		api.Use(httpx.CORS(corsOrigins(cfg)))
		api.Use(httpx.Timeout(60 * time.Second))

		// 公开接口（游客可用）
		api.Group(func(pub chi.Router) {
			pub.Use(auth.Optional(tokens, authSvc))
			pub.Use(limiter.Middleware("public", 120, time.Minute))
			svc.auth.Register(pub)
			svc.calc.Register(pub)
			svc.meta.RegisterPublic(pub)
			svc.interaction.RegisterPublic(pub)

			// 需要登录的用户接口
			pub.Group(func(me chi.Router) {
				me.Use(auth.Middleware(tokens, authSvc))
				svc.user.Register(me)
				svc.photo.Register(me)
				svc.discovery.Register(me)
				svc.interaction.Register(me)
			})
		})

		// 本地存储的媒体代理
		if fsStore != nil {
			api.Group(func(media chi.Router) {
				svc.photo.RegisterMedia(media, fsStore)
			})
		}

		// 后台接口
		api.Group(func(adm chi.Router) {
			adm.Use(auth.Middleware(tokens, authSvc))
			adm.Use(auth.RequireRole(2))
			adm.Use(limiter.Middleware("admin", 300, time.Minute))
			svc.baseline.Register(adm)
			svc.meta.RegisterAdmin(adm)
			svc.admin.Register(adm)
		})
	})
	return r
}

func corsOrigins(cfg *config.Config) []string {
	if cfg.IsProduction() {
		return nil
	}
	return []string{"http://localhost:5173", "http://localhost:5174", "http://127.0.0.1:5173", "http://127.0.0.1:5174"}
}