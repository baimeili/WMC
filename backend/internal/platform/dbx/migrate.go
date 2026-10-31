package dbx

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// up 使用 goose 执行迁移。
func up(ctx context.Context, pool *pgxpool.Pool, fsys fs.FS, dir string) error {
	goose.SetBaseFS(fsys)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("设置 goose 方言失败: %w", err)
	}
	goose.SetLogger(goose.NopLogger())

	sqldb := stdlib.OpenDBFromPool(pool)
	defer sqldb.Close()

	if err := goose.UpContext(ctx, sqldb, dir); err != nil {
		return fmt.Errorf("执行数据库迁移失败: %w", err)
	}
	return nil
}

// Down 回滚最近一次迁移（仅开发/测试使用）。
func Down(ctx context.Context, pool *pgxpool.Pool, fsys fs.FS, dir string) error {
	goose.SetBaseFS(fsys)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	goose.SetLogger(goose.NopLogger())
	sqldb := stdlib.OpenDBFromPool(pool)
	defer sqldb.Close()
	return goose.DownContext(ctx, sqldb, dir)
}
