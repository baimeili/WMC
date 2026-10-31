// Command seed 生成开发与测试数据（DEMO 基准数据集与开发管理员账号）。
// 生产环境（APP_ENV=production）禁止执行。
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"wmc/backend/internal/auth"
	"wmc/backend/internal/baseline"
	"wmc/backend/internal/platform/config"
	"wmc/backend/internal/platform/dbx"
	"wmc/backend/internal/platform/logx"
	"wmc/backend/internal/platform/redisx"
	"wmc/backend/internal/seed"
	"wmc/backend/internal/settings"
	"wmc/backend/internal/store"
)

func main() {
	var (
		demo       = flag.Bool("demo", true, "生成 DEMO 基准数据集并激活")
		demoCount  = flag.Int("demo-records", 3000, "DEMO 数据集记录数")
		admin      = flag.Bool("admin", true, "创建开发管理员账号")
		adminPhone = flag.String("admin-phone", "13800000000", "开发管理员手机号")
		userCount  = flag.Int("users", 0, "生成 TEST DATA 测试用户数量（0 表示不生成）")
	)
	flag.Parse()

	if err := run(*demo, *demoCount, *admin, *adminPhone, *userCount); err != nil {
		fmt.Fprintln(os.Stderr, "生成数据失败:", err)
		os.Exit(1)
	}
}

func run(demo bool, demoCount int, createAdmin bool, adminPhone string, userCount int) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.IsProduction() {
		return fmt.Errorf("生产环境禁止执行测试数据 Seed")
	}
	if cfg.DatabaseURL == "" {
		return fmt.Errorf("必须配置 DATABASE_URL")
	}
	logger := logx.New(cfg.LogLevel)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	pool, err := dbx.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	if err := dbx.Migrate(ctx, pool, logger); err != nil {
		return err
	}

	st := store.New(pool)
	kv, err := redisx.New(ctx, cfg.RateLimitStore, cfg.RedisURL)
	if err != nil {
		return err
	}
	defer kv.Close()

	settingsSvc := settings.NewService(st, 5*time.Second)
	tokens := auth.NewTokenManager(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	authSvc := auth.NewService(st, tokens, cfg.PhoneHashSecret, cfg.PhoneEncryptionKey)
	baselineSvc := baseline.NewService(st, pool, settingsSvc, string(cfg.Env))

	if createAdmin {
		user, err := authSvc.CreateUser(ctx, adminPhone, 4, true)
		if err != nil {
			logger.Warn("管理员账号创建跳过", "reason", err.Error())
		} else {
			fmt.Printf("开发管理员：手机号 %s（super_admin），开发验证码 123456\n", adminPhone)
		}
		_ = user
	}

	if userCount > 0 {
		createdUsers, err := seed.Users(ctx, st, cfg.PhoneHashSecret, cfg.PhoneEncryptionKey, userCount, 20260101)
		if err != nil {
			return fmt.Errorf("生成测试用户失败: %w", err)
		}
		fmt.Printf("已生成 %d 个 TEST DATA 测试用户（手机号 137xxxxxxxx，验证码 123456）\n", createdUsers)
	}

	if demo {
		ds, err := baselineSvc.CreateDemoDataset(ctx, demoCount, nil)
		if err != nil {
			return fmt.Errorf("生成 DEMO 数据集失败: %w", err)
		}
		if err := baselineSvc.Activate(ctx, ds.ID); err != nil {
			return fmt.Errorf("激活 DEMO 数据集失败: %w", err)
		}
		fmt.Printf("DEMO 数据集已生成并激活：%s（%d 条加权样本）\n", ds.Name, demoCount)
	}
	return nil
}