// Package config 负责从环境变量加载全部配置。
// 原则：配置只来自环境变量，代码中不得硬编码数据库密码、密钥等信息。
package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Env 运行环境。
type Env string

const (
	EnvDevelopment Env = "development"
	EnvStaging     Env = "staging"
	EnvProduction  Env = "production"
)

// Config 应用总配置。
type Config struct {
	Env      Env
	HTTPAddr string

	DatabaseURL string
	MigrateOnStart bool

	RedisURL       string
	RateLimitStore string // redis | memory | auto

	JWTSecret          string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	PhoneHashSecret    string
	PhoneEncryptionKey []byte

	SMSProvider  string // mock | aliyun | tencent ...
	SMSMockCode  string
	SMSResend    time.Duration
	SMSDailyLimit int

	StorageDriver  string // fs | s3
	StorageFSRoot  string
	S3Endpoint     string
	S3Region       string
	S3Bucket       string
	S3AccessKey    string
	S3SecretKey    string
	S3UseSSL       bool
	S3PublicURL    string

	AdminDevLoginEnabled bool

	// 本地嵌入式 PostgreSQL（无 Docker 环境使用）
	EmbeddedPGEnabled bool
	EmbeddedPGPort    int
	EmbeddedPGDataDir string
	EmbeddedPGBinaryRepo string

	LogLevel string
}

// IsProduction 是否生产环境。
func (c *Config) IsProduction() bool { return c.Env == EnvProduction }

// Load 从环境变量加载配置并校验关键项。
func Load() (*Config, error) {
	cfg := &Config{
		Env:                  Env(getenv("APP_ENV", string(EnvDevelopment))),
		HTTPAddr:             getenv("HTTP_ADDR", ":8080"),
		DatabaseURL:          getenv("DATABASE_URL", ""),
		MigrateOnStart:       getbool("MIGRATE_ON_START", true),
		RedisURL:             getenv("REDIS_URL", ""),
		RateLimitStore:       getenv("RATE_LIMIT_STORE", "auto"),
		JWTSecret:            getenv("JWT_SECRET", ""),
		AccessTokenTTL:       getduration("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL:      getduration("REFRESH_TOKEN_TTL", 30*24*time.Hour),
		PhoneHashSecret:      getenv("PHONE_HASH_SECRET", ""),
		SMSProvider:          getenv("SMS_PROVIDER", "mock"),
		SMSMockCode:          getenv("SMS_MOCK_CODE", "123456"),
		SMSResend:            getduration("SMS_RESEND_INTERVAL", 60*time.Second),
		SMSDailyLimit:        getint("SMS_DAILY_LIMIT", 5),
		StorageDriver:        getenv("STORAGE_DRIVER", "fs"),
		StorageFSRoot:        getenv("STORAGE_FS_ROOT", ".data/uploads"),
		S3Endpoint:           getenv("S3_ENDPOINT", ""),
		S3Region:             getenv("S3_REGION", ""),
		S3Bucket:             getenv("S3_BUCKET", ""),
		S3AccessKey:          getenv("S3_ACCESS_KEY", ""),
		S3SecretKey:          getenv("S3_SECRET_KEY", ""),
		S3UseSSL:             getbool("S3_USE_SSL", true),
		S3PublicURL:          getenv("S3_PUBLIC_URL", ""),
		AdminDevLoginEnabled: getbool("ADMIN_DEV_LOGIN_ENABLED", true),
		EmbeddedPGEnabled:    getbool("EMBEDDED_PG", false),
		EmbeddedPGPort:       getint("EMBEDDED_PG_PORT", 55432),
		EmbeddedPGDataDir:    getenv("EMBEDDED_PG_DATA_DIR", ".data/pg"),
		EmbeddedPGBinaryRepo: getenv("PG_BINARY_REPO", "https://maven.aliyun.com/repository/public"),
		LogLevel:             getenv("LOG_LEVEL", "info"),
	}

	if key := getenv("PHONE_ENC_KEY", ""); key != "" {
		raw, err := base64.StdEncoding.DecodeString(key)
		if err != nil || len(raw) != 32 {
			return nil, errors.New("PHONE_ENC_KEY 必须是 base64 编码的 32 字节密钥")
		}
		cfg.PhoneEncryptionKey = raw
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	switch c.Env {
	case EnvDevelopment, EnvStaging, EnvProduction:
	default:
		return fmt.Errorf("APP_ENV 取值非法：%s", c.Env)
	}

	if c.IsProduction() {
		if c.JWTSecret == "" || len(c.JWTSecret) < 32 {
			return errors.New("生产环境必须配置至少 32 字节的 JWT_SECRET")
		}
		if c.PhoneHashSecret == "" {
			return errors.New("生产环境必须配置 PHONE_HASH_SECRET")
		}
		if len(c.PhoneEncryptionKey) != 32 {
			return errors.New("生产环境必须配置 32 字节的 PHONE_ENC_KEY")
		}
		if c.SMSProvider == "mock" {
			return errors.New("生产环境禁止使用 mock 短信服务")
		}
		if c.StorageDriver != "s3" {
			return errors.New("生产环境必须使用 s3 对象存储")
		}
		if c.AdminDevLoginEnabled {
			return errors.New("生产环境禁止开启管理员开发登录")
		}
	}

	// 开发环境缺省值，仅用于本地调试
	if !c.IsProduction() {
		if c.JWTSecret == "" {
			c.JWTSecret = "dev-only-jwt-secret-please-change-0123456789"
		}
		if c.PhoneHashSecret == "" {
			c.PhoneHashSecret = "dev-only-phone-hash-secret"
		}
		if len(c.PhoneEncryptionKey) != 32 {
			// 仅用于本地开发：32 字节固定密钥，生产必须通过 PHONE_ENC_KEY 覆盖
			c.PhoneEncryptionKey = []byte("0123456789abcdef0123456789abcdef")
		}
	}

	switch c.RateLimitStore {
	case "redis", "memory", "auto":
	default:
		return fmt.Errorf("RATE_LIMIT_STORE 取值非法：%s", c.RateLimitStore)
	}
	switch c.StorageDriver {
	case "fs", "s3":
	default:
		return fmt.Errorf("STORAGE_DRIVER 取值非法：%s", c.StorageDriver)
	}
	return nil
}

func getenv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return def
}

func getbool(key string, def bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	b, err := strconv.ParseBool(strings.TrimSpace(v))
	if err != nil {
		return def
	}
	return b
}

func getint(key string, def int) int {
	v, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil {
		return def
	}
	return n
}

func getduration(key string, def time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	d, err := time.ParseDuration(strings.TrimSpace(v))
	if err != nil {
		return def
	}
	return d
}
