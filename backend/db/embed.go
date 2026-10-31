// Package db 内嵌数据库迁移脚本，保证二进制自带 Schema 定义。
package db

import "embed"

// Migrations 内嵌 goose 迁移文件。
//
//go:embed migrations/*.sql
var Migrations embed.FS

// MigrationsDir 迁移文件在 embed FS 中的目录。
const MigrationsDir = "migrations"
