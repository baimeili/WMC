<#
.SYNOPSIS
    在未安装 Docker 的开发机上启动本地 PostgreSQL（embedded-postgres 便携版）。
.DESCRIPTION
    首次运行会从 Maven 镜像下载 PostgreSQL 二进制并初始化数据目录。
    按 Ctrl+C 停止数据库。
.EXAMPLE
    pwsh ./scripts/dev-db.ps1
#>
param(
    [int]$Port = 55432,
    [string]$DataDir = ".data/pg"
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Push-Location (Join-Path $root "backend")
try {
    Write-Host "启动本地 PostgreSQL（端口 $Port，数据目录 $DataDir）" -ForegroundColor Cyan
    go run ./cmd/devdb -port $Port -data $DataDir
} finally {
    Pop-Location
}
