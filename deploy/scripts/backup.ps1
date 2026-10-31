<#
.SYNOPSIS
    PostgreSQL 备份脚本（自定义格式，可恢复到任意时间点快照）。
.DESCRIPTION
    默认保留 7 个日备份 + 4 个周备份，备份目录默认 deploy/backups。
.EXAMPLE
    pwsh ./deploy/scripts/backup.ps1 -DatabaseUrl "postgres://wmc:pass@localhost:5432/wmc?sslmode=disable"
#>
param(
    [string]$DatabaseUrl = $env:DATABASE_URL,
    [string]$OutDir = "deploy/backups",
    [int]$KeepDaily = 7,
    [int]$KeepWeekly = 4
)

$ErrorActionPreference = "Stop"
if (-not $DatabaseUrl) { throw "必须提供 DATABASE_URL" }
New-Item -ItemType Directory -Force -Path $OutDir | Out-Null

$stamp = Get-Date -Format "yyyyMMdd_HHmmss"
$isWeekly = (Get-Date).DayOfWeek -eq "Sunday"
$prefix = if ($isWeekly) { "weekly" } else { "daily" }
$file = Join-Path $OutDir ("wmc_" + $prefix + "_" + $stamp + ".dump")

Write-Host "开始备份到 $file" -ForegroundColor Cyan
$pgDump = (Get-Command pg_dump -ErrorAction SilentlyContinue)
if ($pgDump) {
    & $pgDump.Source --dbname=$DatabaseUrl --format=custom --no-owner --no-privileges --file=$file
} else {
    # 回退：在 compose 容器内执行
    docker compose -f deploy/docker-compose.yml exec -T postgres pg_dump --dbname="$DatabaseUrl" --format=custom --no-owner --no-privileges > $file
}

if (-not (Test-Path $file)) { throw "备份失败" }
Write-Host ("备份完成：" + $file + " 大小 " + [math]::Round((Get-Item $file).Length/1MB, 2) + " MB") -ForegroundColor Green

# 清理过期备份
foreach ($kind in @(@{ Prefix = "daily"; Keep = $KeepDaily }, @{ Prefix = "weekly"; Keep = $KeepWeekly })) {
    $old = Get-ChildItem $OutDir -Filter ("wmc_" + $kind.Prefix + "_*.dump") | Sort-Object LastWriteTime -Descending | Select-Object -Skip $kind.Keep
    foreach ($f in $old) { Remove-Item $f.FullName -Force; Write-Host ("已删除旧备份 " + $f.Name) }
}
