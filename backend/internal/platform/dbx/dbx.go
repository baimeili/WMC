// Package dbx 负责 PostgreSQL 连接池与迁移。
package dbx

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"wmc/backend/db"
)

// NewPool 创建连接池并做一次连通性检查。
func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("解析数据库连接串失败: %w", err)
	}
	cfg.MaxConns = 20
	cfg.MinConns = 2
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("创建数据库连接池失败: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("数据库不可用: %w", err)
	}
	return pool, nil
}

// Migrate 执行全部迁移（幂等）。
func Migrate(ctx context.Context, pool *pgxpool.Pool, logger *slog.Logger) error {
	if err := up(ctx, pool, db.Migrations, db.MigrationsDir); err != nil {
		return err
	}
	if logger != nil {
		logger.InfoContext(ctx, "数据库迁移完成")
	}
	return nil
}

// Ping 健康检查。
func Ping(ctx context.Context, pool *pgxpool.Pool) error {
	if pool == nil {
		return fmt.Errorf("数据库未初始化")
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return pool.Ping(ctx)
}
