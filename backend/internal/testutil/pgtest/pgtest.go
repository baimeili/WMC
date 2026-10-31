//go:build integration

// Package pgtest 为集成测试提供真实 PostgreSQL 数据库。
// 要求 TEST_DATABASE_URL 指向一个可创建数据库的 PostgreSQL 账号；未设置时跳过测试。
package pgtest

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"wmc/backend/internal/platform/dbx"
	"wmc/backend/internal/settings"
	"wmc/backend/internal/store"
)

// Env 返回测试数据库的测试上下文。
type Env struct {
	Pool     *pgxpool.Pool
	Store    *store.Store
	Settings *settings.Service
}

// New 创建一个全新的测试数据库并执行全部迁移。
func New(t *testing.T) *Env {
	t.Helper()
	base := os.Getenv("TEST_DATABASE_URL")
	if base == "" {
		t.Skip("未设置 TEST_DATABASE_URL，跳过集成测试（可先运行 scripts/dev-db.ps1）")
	}
	ctx := context.Background()

	adminDSN, err := dsnWithDatabase(base, "postgres")
	if err != nil {
		t.Fatalf("解析 TEST_DATABASE_URL 失败: %v", err)
	}
	adminPool, err := pgxpool.New(ctx, adminDSN)
	if err != nil {
		t.Fatalf("连接维护数据库失败: %v", err)
	}
	defer adminPool.Close()

	name := fmt.Sprintf("wmc_test_%d", time.Now().UnixNano())
	if _, err := adminPool.Exec(ctx, "CREATE DATABASE "+name); err != nil {
		t.Fatalf("创建测试数据库失败: %v", err)
	}

	testDSN, err := dsnWithDatabase(base, name)
	if err != nil {
		t.Fatalf("生成测试 DSN 失败: %v", err)
	}
	pool, err := pgxpool.New(ctx, testDSN)
	if err != nil {
		t.Fatalf("连接测试数据库失败: %v", err)
	}
	t.Cleanup(func() {
		pool.Close()
		cleanupPool, err := pgxpool.New(context.Background(), adminDSN)
		if err != nil {
			return
		}
		defer cleanupPool.Close()
		_, _ = cleanupPool.Exec(context.Background(), "DROP DATABASE IF EXISTS "+name+" WITH (FORCE)")
	})

	if err := dbx.Migrate(ctx, pool, nil); err != nil {
		t.Fatalf("迁移失败: %v", err)
	}

	st := store.New(pool)
	return &Env{
		Pool:     pool,
		Store:    st,
		Settings: settings.NewService(st, time.Millisecond),
	}
}

func dsnWithDatabase(raw, database string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", err
	}
	u.Path = "/" + database
	return u.String(), nil
}
