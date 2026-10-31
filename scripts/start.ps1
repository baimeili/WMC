<#
.SYNOPSIS
    一键启动「婚恋市场计算器」系统（PostgreSQL + 后端 API + H5 + 运营后台）。

.DESCRIPTION
    幂等设计：已经在运行的组件会被直接复用，不会重复启动进程。
    - 数据库：本地便携版 PostgreSQL（无 Docker 环境），监听 55432，数据目录 backend/.data/pg
    - 后端 API：监听 8080
    - H5 客户端：uni-app 开发服务器，监听 5173
    - 运营后台：Vue 开发服务器，监听 5174
    后台输出写入 .data/logs，PID 写入 .data/pids，便于 stop.ps1 精确停止。

.PARAMETER SkipFrontend
    只启动数据库与后端 API，不启动两个前端开发服务器。
.PARAMETER SkipDb
    跳过数据库启动（例如使用外部 PostgreSQL，通过 DATABASE_URL 指定）。
.PARAMETER SkipApi
    跳过后端 API 启动。
.PARAMETER SkipBuild
    跳过 go build（使用 backend/bin 下已有二进制，加快启动）。
.PARAMETER Force
    先执行 stop.ps1 停止全部组件，再重新启动。

.EXAMPLE
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/start.ps1
.EXAMPLE
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/start.ps1 -SkipFrontend
#>
[CmdletBinding()]
param(
    [switch]$SkipFrontend,
    [switch]$SkipDb,
    [switch]$SkipApi,
    [switch]$SkipBuild,
    [switch]$Force
)

$ErrorActionPreference = 'Stop'

$Root      = Split-Path -Parent $PSScriptRoot
$Backend   = Join-Path $Root 'backend'
$BinDir    = Join-Path $Backend 'bin'
$LogDir    = Join-Path $Root '.data/logs'
$PidDir    = Join-Path $Root '.data/pids'
$AdminDir  = Join-Path $Root 'admin/vue-admin'
$AppDir    = Join-Path $Root 'app/uni-app'
$DbDataDir = Join-Path $Backend '.data/pg'

foreach ($d in @($BinDir, $LogDir, $PidDir)) {
    if (-not (Test-Path $d)) { New-Item -ItemType Directory -Force -Path $d | Out-Null }
}

if (-not $env:DATABASE_URL) {
    $env:DATABASE_URL = 'postgres://wmc:wmc_local_password@localhost:55432/wmc?sslmode=disable'
}
if (-not $env:APP_ENV) { $env:APP_ENV = 'development' }
if (-not $env:HTTP_ADDR) { $env:HTTP_ADDR = ':8080' }

$DbPort    = 55432
$ApiPort   = 8080
$H5Port    = 5173
$AdminPort = 5174

# 本机 Get-NetTCPConnection 单次调用约 13 秒，因此改为一次 netstat 解析 + 内存查询。
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

function Get-PortOwnerName {
    param([int]$Port)
    if (-not $script:ListenMap.ContainsKey($Port)) { Update-ListenMap }
    if (-not $script:ListenMap.ContainsKey($Port)) { return '' }
    $procId = $script:ListenMap[$Port]
    $proc = Get-Process -Id $procId -ErrorAction SilentlyContinue
    if ($proc) { return ($proc.ProcessName + ' (PID ' + $procId + ')') }
    return ('PID ' + $procId)
}

function Save-ProcessId {
    param([string]$Name, [int]$ProcessId)
    Set-Content -Path (Join-Path $PidDir ($Name + '.pid')) -Value $ProcessId -Encoding ascii
}

function Start-Component {
    param([string]$Name, [string]$FilePath, [string[]]$Arguments, [string]$WorkingDirectory)
    $out = Join-Path $LogDir ($Name + '.log')
    $err = Join-Path $LogDir ($Name + '.err.log')
    $argList = @($Arguments | Where-Object { $null -ne $_ -and $_ -ne '' })
    if ($argList.Count -gt 0) {
        $p = Start-Process -FilePath $FilePath -ArgumentList $argList -WorkingDirectory $WorkingDirectory -WindowStyle Hidden -PassThru -RedirectStandardOutput $out -RedirectStandardError $err
    } else {
        $p = Start-Process -FilePath $FilePath -WorkingDirectory $WorkingDirectory -WindowStyle Hidden -PassThru -RedirectStandardOutput $out -RedirectStandardError $err
    }
    Save-ProcessId -Name $Name -ProcessId $p.Id
    return $p
}

function Wait-Http {
    param([string]$Url, [int]$TimeoutSeconds = 60)
    $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
    while ((Get-Date) -lt $deadline) {
        try {
            $resp = Invoke-WebRequest -Uri $Url -UseBasicParsing -TimeoutSec 20
            if ($resp.StatusCode -ge 200 -and $resp.StatusCode -lt 500) { return $true }
        } catch { }
        Start-Sleep -Milliseconds 800
    }
    return $false
}

function Wait-Port {
    param([int]$Port, [int]$TimeoutSeconds = 60)
    $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
    while ((Get-Date) -lt $deadline) {
        if (Test-Port -Port $Port) { return $true }
        Start-Sleep -Milliseconds 500
    }
    return $false
}

function Build-Binary {
    param([string]$OutputName, [string]$Package)
    $outFile = Join-Path $BinDir $OutputName
    Write-Host ('  编译 ' + $Package + ' ...') -ForegroundColor DarkGray
    Push-Location $Backend
    try {
        & go build -o $outFile $Package
        if ($LASTEXITCODE -ne 0) { throw ('go build 失败：' + $Package) }
    } finally {
        Pop-Location
    }
}

Write-Host ''
Write-Host '=== 婚恋市场计算器 · 一键启动 ===' -ForegroundColor Cyan

if ($Force) {
    Write-Host '--force：先停止现有组件' -ForegroundColor Yellow
    & (Join-Path $PSScriptRoot 'stop.ps1')
    Start-Sleep -Seconds 2
}

# ---------- 1. 数据库 ----------
if (-not $SkipDb) {
    if (Test-Port -Port $DbPort) {
        Write-Host ('[跳过] 数据库已在运行：端口 ' + $DbPort + '（' + (Get-PortOwnerName -Port $DbPort) + '）') -ForegroundColor Yellow
    } else {
        if (-not $SkipBuild) { Build-Binary -OutputName 'wmc-devdb.exe' -Package './cmd/devdb' }
        $exe = Join-Path $BinDir 'wmc-devdb.exe'
        if (-not (Test-Path $exe)) { throw '找不到 wmc-devdb.exe，请去掉 -SkipBuild 重新运行' }
        $stalePidFile = Join-Path $DbDataDir 'data/postmaster.pid'
        if (Test-Path $stalePidFile) {
            $firstLine = Get-Content $stalePidFile -TotalCount 1 -ErrorAction SilentlyContinue
            $stalePid = 0
            if ($firstLine -and [int]::TryParse([string]$firstLine, [ref]$stalePid)) {
                if (-not (Get-Process -Id $stalePid -ErrorAction SilentlyContinue)) {
                    Remove-Item $stalePidFile -Force -ErrorAction SilentlyContinue
                    Write-Host '[清理] 发现残留 postmaster.pid（进程已退出），已删除' -ForegroundColor Yellow
                }
            }
        }
        Write-Host '[启动] 本地 PostgreSQL ...' -ForegroundColor Green
        Start-Component -Name 'db' -FilePath $exe -Arguments @('-port', [string]$DbPort, '-data', $DbDataDir) -WorkingDirectory $Backend | Out-Null
        if (-not (Wait-Port -Port $DbPort -TimeoutSeconds 120)) {
            throw ('数据库启动超时，请查看日志：' + (Join-Path $LogDir 'db.log'))
        }
        Write-Host ('[就绪] 数据库 127.0.0.1:' + $DbPort) -ForegroundColor Green
    }
} else {
    Write-Host '[跳过] 数据库（-SkipDb）' -ForegroundColor Yellow
}

# ---------- 2. 后端 API ----------
if (-not $SkipApi) {
    if (Test-Port -Port $ApiPort) {
        Write-Host ('[跳过] 后端 API 已在运行：端口 ' + $ApiPort + '（' + (Get-PortOwnerName -Port $ApiPort) + '）') -ForegroundColor Yellow
    } else {
        if (-not $SkipBuild) { Build-Binary -OutputName 'wmc-api.exe' -Package './cmd/api' }
        $exe = Join-Path $BinDir 'wmc-api.exe'
        if (-not (Test-Path $exe)) { throw '找不到 wmc-api.exe，请去掉 -SkipBuild 重新运行' }
        Write-Host '[启动] 后端 API ...' -ForegroundColor Green
        Start-Component -Name 'api' -FilePath $exe -Arguments @() -WorkingDirectory $Backend | Out-Null
        if (-not (Wait-Http -Url ('http://127.0.0.1:' + $ApiPort + '/health/ready') -TimeoutSeconds 90)) {
            throw ('后端 API 启动超时，请查看日志：' + (Join-Path $LogDir 'api.log'))
        }
        Write-Host ('[就绪] API http://127.0.0.1:' + $ApiPort) -ForegroundColor Green
    }
} else {
    Write-Host '[跳过] 后端 API（-SkipApi）' -ForegroundColor Yellow
}

# ---------- 3. 前端开发服务器 ----------
if (-not $SkipFrontend) {
    if (Test-Port -Port $H5Port) {
        Write-Host ('[跳过] H5 开发服务器已在运行：端口 ' + $H5Port) -ForegroundColor Yellow
    } elseif (-not (Test-Path (Join-Path $AppDir 'node_modules'))) {
        Write-Host '[警告] app/uni-app 缺少 node_modules，请先在该目录执行 npm install' -ForegroundColor Red
    } else {
        Write-Host '[启动] H5 客户端（uni-app）...' -ForegroundColor Green
        Start-Component -Name 'h5' -FilePath 'cmd.exe' -Arguments @('/c', 'npm run dev:h5') -WorkingDirectory $AppDir | Out-Null
        if (Wait-Http -Url ('http://localhost:' + $H5Port + '/') -TimeoutSeconds 150) {
            Write-Host ('[就绪] H5 http://localhost:' + $H5Port) -ForegroundColor Green
        } else {
            Write-Host '[警告] H5 启动较慢或失败，请查看 .data/logs/h5.log（可继续使用后端）' -ForegroundColor Red
        }
    }

    if (Test-Port -Port $AdminPort) {
        Write-Host ('[跳过] 运营后台已在运行：端口 ' + $AdminPort) -ForegroundColor Yellow
    } elseif (-not (Test-Path (Join-Path $AdminDir 'node_modules'))) {
        Write-Host '[警告] admin/vue-admin 缺少 node_modules，请先在该目录执行 npm install' -ForegroundColor Red
    } else {
        Write-Host '[启动] 运营后台（Vue）...' -ForegroundColor Green
        Start-Component -Name 'admin' -FilePath 'cmd.exe' -Arguments @('/c', 'npm run dev') -WorkingDirectory $AdminDir | Out-Null
        if (Wait-Http -Url ('http://localhost:' + $AdminPort + '/admin/') -TimeoutSeconds 150) {
            Write-Host ('[就绪] 运营后台 http://localhost:' + $AdminPort + '/admin/') -ForegroundColor Green
        } else {
            Write-Host '[警告] 运营后台启动较慢或失败，请查看 .data/logs/admin.log' -ForegroundColor Red
        }
    }
} else {
    Write-Host '[跳过] 前端开发服务器（-SkipFrontend）' -ForegroundColor Yellow
}

# ---------- 4. 汇总 ----------
Write-Host ''
Write-Host '=== 当前状态 ===' -ForegroundColor Cyan
$rows = @(
    @{ Name = '数据库 PostgreSQL'; Port = $DbPort; Url = ('127.0.0.1:' + $DbPort) },
    @{ Name = '后端 API'; Port = $ApiPort; Url = ('http://127.0.0.1:' + $ApiPort) },
    @{ Name = 'H5 客户端'; Port = $H5Port; Url = ('http://localhost:' + $H5Port) },
    @{ Name = '运营后台'; Port = $AdminPort; Url = ('http://localhost:' + $AdminPort + '/admin/') }
)
foreach ($row in $rows) {
    if (Test-Port -Port $row.Port) {
        Write-Host ('  [运行] ' + $row.Name.PadRight(18) + $row.Url + '  (' + (Get-PortOwnerName -Port $row.Port) + ')') -ForegroundColor Green
    } else {
        Write-Host ('  [停止] ' + $row.Name.PadRight(18) + $row.Url) -ForegroundColor DarkGray
    }
}

Write-Host ''
Write-Host '开发环境账号：' -ForegroundColor Cyan
Write-Host '  管理员手机号 13800000000（super_admin，后台登录页可使用）'
Write-Host '  测试用户手机号 137xxxxxxxx（如 13700000199），短信验证码 123456'
Write-Host ''
Write-Host '日志目录： .data/logs    进程 PID： .data/pids' -ForegroundColor DarkGray
Write-Host '停止全部： powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/stop.ps1' -ForegroundColor DarkGray
Write-Host ''

if (-not (Test-Port -Port $ApiPort)) {
    Write-Host '后端 API 未运行，请检查上面的错误信息。' -ForegroundColor Red
    exit 1
}
exit 0