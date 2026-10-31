// Command worker 运行后台定时任务：清理过期 Token、执行注销保留期清理。
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"wmc/backend/internal/platform/config"
	"wmc/backend/internal/platform/dbx"
	"wmc/backend/internal/platform/logx"
	"wmc/backend/internal/settings"
	"wmc/backend/internal/store"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "worker 启动失败:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.DatabaseURL == "" {
		return fmt.Errorf("必须配置 DATABASE_URL")
	}
	logger := logx.New(cfg.LogLevel)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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
	settingsSvc := settings.NewService(st, 30*time.Second)

	logger.Info("worker 已启动")
	tokenTicker := time.NewTicker(time.Hour)
	purgeTicker := time.NewTicker(6 * time.Hour)
	defer tokenTicker.Stop()
	defer purgeTicker.Stop()

	runTokenCleanup(ctx, st, logger)
	runDeletionPurge(ctx, st, settingsSvc, logger)

	for {
		select {
		case <-ctx.Done():
			logger.Info("worker 收到退出信号")
			return nil
		case <-tokenTicker.C:
			runTokenCleanup(ctx, st, logger)
		case <-purgeTicker.C:
			runDeletionPurge(ctx, st, settingsSvc, logger)
		}
	}
}

func runTokenCleanup(ctx context.Context, st *store.Store, logger *slog.Logger) {
	n, err := st.Q.DeleteExpiredRefreshTokens(ctx)
	if err != nil {
		logger.Error("清理过期 Refresh Token 失败", "error", err)
		return
	}
	logger.Info("清理过期 Refresh Token", "deleted", n)
}

// runDeletionPurge 处理超过保留期的注销账号：匿名化个人资料与手机号，保留必要的审计数据。
func runDeletionPurge(ctx context.Context, st *store.Store, settingsSvc *settings.Service, logger *slog.Logger) {
	retentionDays := settingsSvc.Int(ctx, settings.KeyDeletionRetentionDays, 30)
	cutoff := time.Now().AddDate(0, 0, -retentionDays)

	tx, err := st.Pool.Begin(ctx)
	if err != nil {
		logger.Error("开启注销清理事务失败", "error", err)
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	rows, err := tx.Query(ctx, `
SELECT id FROM users
WHERE status = 5 AND deletion_requested_at IS NOT NULL AND deletion_requested_at < $1
FOR UPDATE SKIP LOCKED
LIMIT 200`, cutoff)
	if err != nil {
		logger.Error("查询待清理账号失败", "error", err)
		return
	}
	ids := make([]string, 0, 200)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	rows.Close()
	if len(ids) == 0 {
		_ = tx.Commit(ctx)
		return
	}

	for _, id := range ids {
		// 手机号不可恢复：清除密文并把摘要替换为随机值，避免继续被检索
		if _, err := tx.Exec(ctx, `
UPDATE users
SET status = 4, deleted_at = now(), discovery_enabled = FALSE,
    phone_encrypted = ''::bytea,
    phone_hash = digest(uuid_send(id), 'sha256')
WHERE id = $1`, id); err != nil {
			logger.Error("匿名化账号失败", "user_id", id, "error", err)
		}
		_, _ = tx.Exec(ctx, `
UPDATE user_profiles
SET nickname = '已注销用户', bio = '', height_cm = NULL, birth_date = NULL, district_code = NULL
WHERE user_id = $1`, id)
		_, _ = tx.Exec(ctx, `UPDATE user_photos SET deleted_at = now(), is_primary = FALSE WHERE user_id = $1 AND deleted_at IS NULL`, id)
		_, _ = tx.Exec(ctx, `DELETE FROM user_compatibility_rules WHERE user_id = $1`, id)
		_, _ = tx.Exec(ctx, `DELETE FROM partner_preferences WHERE user_id = $1`, id)
	}
	if err := tx.Commit(ctx); err != nil {
		logger.Error("提交注销清理失败", "error", err)
		return
	}
	logger.Info("已完成注销保留期清理", "count", len(ids), "retention_days", retentionDays)
}
