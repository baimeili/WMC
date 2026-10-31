// Command devdb 在未安装 Docker 的开发机上启动本地 PostgreSQL。
// 二进制来自 Maven 镜像（默认阿里云），首次运行会自动下载并缓存。
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
)

func main() {
	var (
		port     = flag.Int("port", envInt("EMBEDDED_PG_PORT", 55432), "监听端口")
		dataDir  = flag.String("data", envStr("EMBEDDED_PG_DATA_DIR", ".data/pg"), "数据目录")
		user     = flag.String("user", "wmc", "数据库用户")
		password = flag.String("password", "wmc_local_password", "数据库密码（仅本地开发）")
		database = flag.String("database", "wmc", "数据库名")
		repo     = flag.String("repo", envStr("PG_BINARY_REPO", "https://maven.aliyun.com/repository/public"), "PostgreSQL 二进制镜像地址")
	)
	flag.Parse()

	abs, err := filepath.Abs(*dataDir)
	if err != nil {
		fail(fmt.Errorf("解析数据目录失败: %w", err))
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		fail(fmt.Errorf("创建数据目录失败: %w", err))
	}

	cfg := embeddedpostgres.DefaultConfig().
		Port(uint32(*port)).
		Database(*database).
		Username(*user).
		Password(*password).
		DataPath(filepath.Join(abs, "data")).
		RuntimePath(filepath.Join(abs, "runtime")).
		BinariesPath(filepath.Join(abs, "bin")).
		BinaryRepositoryURL(*repo).
		StartTimeout(120 * time.Second)

	db := embeddedpostgres.NewDatabase(cfg)
	fmt.Println("正在准备本地 PostgreSQL（首次运行需要下载二进制，请稍候）...")
	if err := db.Start(); err != nil {
		fail(fmt.Errorf("启动本地 PostgreSQL 失败: %w", err))
	}

	dsn := fmt.Sprintf("postgres://%s:%s@localhost:%d/%s?sslmode=disable", *user, *password, *port, *database)
	fmt.Println("本地 PostgreSQL 已启动")
	fmt.Println("DATABASE_URL=" + dsn)
	fmt.Println("按 Ctrl+C 停止服务")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	fmt.Println("正在停止本地 PostgreSQL ...")
	if err := db.Stop(); err != nil {
		fmt.Fprintln(os.Stderr, "停止失败:", err)
		os.Exit(1)
	}
	fmt.Println("已停止")
}

func envStr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	var n int
	if _, err := fmt.Sscanf(v, "%d", &n); err != nil {
		return def
	}
	return n
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "错误:", err)
	os.Exit(1)
}
