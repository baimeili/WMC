<#
.SYNOPSIS
    分阶段验收脚本：编译 → 静态检查 → 单元测试 → 集成测试（可选） → 前端构建。
.EXAMPLE
    pwsh ./scripts/test-all.ps1
    pwsh ./scripts/test-all.ps1 -SkipFrontend
#>
param(
    [switch]$SkipFrontend,
    [switch]$SkipIntegration,
    [string]$TestDatabaseUrl = $env:TEST_DATABASE_URL
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$failed = @()

function Step([string]$name, [scriptblock]$action) {
    Write-Host ""
    Write-Host "==== $name ====" -ForegroundColor Cyan
    try {
        & $action
        if ($LASTEXITCODE -ne 0 -and $null -ne $LASTEXITCODE) { throw "退出码 $LASTEXITCODE" }
        Write-Host "[通过] $name" -ForegroundColor Green
    } catch {
        Write-Host "[失败] $name : $_" -ForegroundColor Red
        $script:failed += $name
    }
}

Push-Location (Join-Path $root "backend")
Step "go build" { go build ./... }
Step "go vet" { go vet ./... }
Step "go test（单元测试）" { go test ./... }
if (-not $SkipIntegration) {
    if ($TestDatabaseUrl) {
        $env:TEST_DATABASE_URL = $TestDatabaseUrl
        Step "go test（集成测试）" { go test -tags=integration ./... }
    } else {
        Write-Host "[跳过] 集成测试：未设置 TEST_DATABASE_URL（可用 scripts/dev-db.ps1 启动本地库）" -ForegroundColor Yellow
    }
}
Pop-Location

if (-not $SkipFrontend) {
    Push-Location (Join-Path $root "admin/vue-admin")
    Step "admin 构建" { npm run build }
    Pop-Location

    Push-Location (Join-Path $root "app/uni-app")
    Step "uni-app H5 构建" { npm run build:h5 }
    Pop-Location
}

Write-Host ""
if ($failed.Count -eq 0) {
    Write-Host "全部检查通过" -ForegroundColor Green
    exit 0
} else {
    Write-Host ("以下检查失败：" + ($failed -join "、")) -ForegroundColor Red
    exit 1
}
