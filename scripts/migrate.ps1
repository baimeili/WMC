<#
.SYNOPSIS
    数据库迁移管理（goose）。
.EXAMPLE
    pwsh ./scripts/migrate.ps1 up
    pwsh ./scripts/migrate.ps1 status
    pwsh ./scripts/migrate.ps1 down
#>
param(
    [ValidateSet("up", "down", "status", "version", "reset")]
    [string]$Command = "up",
    [string]$DatabaseUrl = $env:DATABASE_URL
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$migrations = Join-Path $root "backend/db/migrations"

if (-not $DatabaseUrl) {
    $DatabaseUrl = "postgres://wmc:wmc_local_password@localhost:55432/wmc?sslmode=disable"
}

$goose = Join-Path $env:USERPROFILE "go/bin/goose.exe"
if (-not (Test-Path $goose)) { $goose = "goose" }

switch ($Command) {
    "up"      { & $goose -dir $migrations postgres $DatabaseUrl up }
    "down"    { & $goose -dir $migrations postgres $DatabaseUrl down }
    "status"  { & $goose -dir $migrations postgres $DatabaseUrl status }
    "version" { & $goose -dir $migrations postgres $DatabaseUrl version }
    "reset"   { & $goose -dir $migrations postgres $DatabaseUrl reset }
}
