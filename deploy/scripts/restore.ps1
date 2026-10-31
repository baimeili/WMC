<#
.SYNOPSIS
    从自定义格式备份恢复数据库。
.EXAMPLE
    pwsh ./deploy/scripts/restore.ps1 -File deploy/backups/wmc_daily_20260101_030000.dump
#>
param(
    [Parameter(Mandatory = $true)][string]$File,
    [string]$DatabaseUrl = $env:DATABASE_URL
)

$ErrorActionPreference = "Stop"
if (-not $DatabaseUrl) { throw "必须提供 DATABASE_URL" }
if (-not (Test-Path $File)) { throw "备份文件不存在：$File" }

Write-Host "从 $File 恢复数据库（请确认目标库可被覆盖）" -ForegroundColor Yellow
$pgRestore = (Get-Command pg_restore -ErrorAction SilentlyContinue)
if ($pgRestore) {
    & $pgRestore.Source --dbname=$DatabaseUrl --clean --if-exists --no-owner --no-privileges $File
} else {
    Get-Content -Raw $File | docker compose -f deploy/docker-compose.yml exec -T postgres pg_restore --dbname="$DatabaseUrl" --clean --if-exists --no-owner --no-privileges
}
Write-Host "恢复完成" -ForegroundColor Green
