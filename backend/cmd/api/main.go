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

	"wmc/backend/internal/auth"
	"wmc/backend/internal/baseline"
	"wmc/backend/internal/calculator"
	"wmc/backend/internal/meta"
	"wmc/backend/internal/platform/config"
	"wmc/backend/internal/platform/dbx"
	"wmc/backend/internal/platform/httpx"
	"wmc/backend/internal/platform/logx"
	"wmc/backend/internal/platform/metrics"
	"wmc/backend/internal/platform/ratelimit"
	"wmc/backend/internal/platform/redisx"
	"wmc/backend/internal/settings"
	"wmc/backend/internal/store"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "启动失败:", err)
		os.Exit(1)
	}
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

	settingsSvc := settings.NewService(st, 15*time.Second)
	tokens := auth.NewTokenManager(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	authSvc := auth.NewService(st, tokens, cfg.PhoneHashSecret, cfg.PhoneEncryptionKey)
	calcSvc := calculator.NewService(st, pool, settingsSvc, kv)
	metaSvc := meta.NewService(st, settingsSvc)
	baselineSvc := baseline.NewService(st, pool, settingsSvc, string(cfg.Env))
	limiter := ratelimit.New(kv)

	m := metrics.New()
	m.AttachPool(pool)

	router := buildRouter(cfg, logger, pool, kv, m, tokens, authSvc, calcSvc, metaSvc, baselineSvc, limiter, settingsSvc)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
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

func buildRouter(
	cfg *config.Config,
	logger *slog.Logger,
	pool *pgxpool.Pool,
	kv redisx.Store,
	m *metrics.Metrics,
	tokens *auth.TokenManager,
	authSvc *auth.Service,
	calcSvc *calculator.Service,
	metaSvc *meta.Service,
	baselineSvc *baseline.Service,
	limiter *ratelimit.Limiter,
	settingsSvc *settings.Service,
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

	calcHandler := calculator.NewHandler(calcSvc)
	metaHandler := meta.NewHandler(metaSvc)
	baselineHandler := baseline.NewHandler(baselineSvc)
	authHandler := auth.NewHandler(authSvc, cfg.AdminDevLoginEnabled)

	r.Route("/api/v1", func(api chi.Router) {
		api.Use(httpx.CORS(corsOrigins(cfg)))
		api.Use(httpx.Timeout(30 * time.Second))

		// 公开接口（游客可用，带可选身份）
		api.Group(func(pub chi.Router) {
			pub.Use(auth.Optional(tokens, authSvc))
			pub.Use(limiter.Middleware("calculator", 60, time.Minute))
			authHandler.Register(pub)
			calcHandler.Register(pub)
			metaHandler.RegisterPublic(pub)
		})

		// 后台接口（需管理员身份）
		api.Group(func(adm chi.Router) {
			adm.Use(auth.Middleware(tokens, authSvc))
			adm.Use(auth.RequireRole(2))
			baselineHandler.Register(adm)
			metaHandler.RegisterAdmin(adm)
		})
	})
	_ = settingsSvc
	return r
}

func corsOrigins(cfg *config.Config) []string {
	if cfg.IsProduction() {
		return nil // 生产由反向代理配置同源访问
	}
	return []string{"http://localhost:5173", "http://localhost:5174", "http://127.0.0.1:5173", "http://127.0.0.1:5174"}
}