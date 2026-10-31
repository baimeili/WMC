<#
.SYNOPSIS
    停止「婚恋市场计算器」的全部本地进程（数据库、后端 API、前端开发服务器）。

.DESCRIPTION
    1) 优先按 .data/pids 中记录的 PID 精确停止；
    2) 数据库通过 pg_ctl 优雅停止，避免数据损坏；
    3) 兜底按端口清理未被 PID 文件记录的进程；
    4) 不删除 .data 数据目录。

.EXAMPLE
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/stop.ps1
#>
[CmdletBinding()]
param()

$ErrorActionPreference = 'Continue'

$Root   = Split-Path -Parent $PSScriptRoot
$PidDir = Join-Path $Root '.data/pids'
$PgData = Join-Path $Root 'backend/.data/pg/data'
$PgCtl  = Join-Path $Root 'backend/.data/pg/bin/bin/pg_ctl.exe'

$script:ListenMap = @{}

function Update-ListenMap {
    $map = @{}
    $rows = & netstat.exe -ano 2>$null
    foreach ($row in $rows) {
        if ($row -match ':(\d+)\s+\S+\s+LISTENING\s+(\d+)') {
            $port = [int]$Matches[1]
            $procId = [int]$Matches[2]
            if (-not $map.ContainsKey($port)) { $map[$port] = $procId }
        }
    }
    $script:ListenMap = $map
}

function Test-Port {
    param([int]$Port)
    if ($script:ListenMap.ContainsKey($Port)) { return $true }
    Update-ListenMap
    return $script:ListenMap.ContainsKey($Port)
}

function Stop-ProcessTree {
    param([int]$ProcessId, [string]$Name)
    $proc = Get-Process -Id $ProcessId -ErrorAction SilentlyContinue
    if (-not $proc) { return }
    Write-Host ('  停止 ' + $Name + ' (PID ' + $ProcessId + '，' + $proc.ProcessName + ')') -ForegroundColor Yellow
    & taskkill.exe /PID $ProcessId /T /F 2>&1 | Out-Null
}

Write-Host ''
Write-Host '=== 婚恋市场计算器 · 停止 ===' -ForegroundColor Cyan

# 1) 按 PID 文件停止
if (Test-Path $PidDir) {
    foreach ($file in Get-ChildItem $PidDir -Filter '*.pid' -ErrorAction SilentlyContinue) {
        $raw = Get-Content $file.FullName -ErrorAction SilentlyContinue
        if ($raw) {
            $procId = 0
            if ([int]::TryParse([string]$raw, [ref]$procId)) {
                Stop-ProcessTree -ProcessId $procId -Name $file.BaseName
            }
        }
        Remove-Item $file.FullName -Force -ErrorAction SilentlyContinue
    }
}

# 2) 优雅停止 PostgreSQL
if ((Test-Path $PgCtl) -and (Test-Path (Join-Path $PgData 'postmaster.pid'))) {
    Write-Host '  停止 PostgreSQL（pg_ctl fast）...' -ForegroundColor Yellow
    & $PgCtl -D $PgData stop -m fast 2>&1 | Out-Null
    Start-Sleep -Seconds 2
}

# 3) 按端口兜底清理（含历史 go run 启动的进程）
$portNames = @{ 8080 = '后端 API'; 5173 = 'H5 开发服务器'; 5174 = '运营后台'; 55432 = 'PostgreSQL' }
Update-ListenMap
foreach ($port in @(8080, 5173, 5174, 55432)) {
    if ($script:ListenMap.ContainsKey($port)) {
        Stop-ProcessTree -ProcessId $script:ListenMap[$port] -Name ($portNames[$port] + '（端口 ' + $port + '）')
        Start-Sleep -Milliseconds 500
    }
}

# 4) 结果核对
Update-ListenMap
Write-Host ''
Write-Host '=== 停止结果 ===' -ForegroundColor Cyan
foreach ($port in @(55432, 8080, 5173, 5174)) {
    if ($script:ListenMap.ContainsKey($port)) {
        Write-Host ('  [仍在监听] 端口 ' + $port) -ForegroundColor Red
    } else {
        Write-Host ('  [已停止]   端口 ' + $port) -ForegroundColor Green
    }
}
Write-Host ''
Write-Host '数据目录 .data 保留，下次 start.ps1 会直接复用。' -ForegroundColor DarkGray
Write-Host ''
exit 0