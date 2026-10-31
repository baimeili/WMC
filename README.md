# 婚恋市场计算器（WMC）

> 先让 1000 个人真正可以使用，再考虑 100 万人。

婚恋市场计算器 + 双向婚恋匹配 MVP。它不是刷脸 Dating App，而是先回答两个问题：

1. 你的择偶条件，在**理论人口市场**中筛掉了多少人？（基于基准数据集的统计估算）
2. 在**本站真实用户池**里，有多少人和你**双向**满足硬条件？（严格的双向硬筛 + 可解释软评分）

两个市场严格区分：理论市场只给估算比例，绝不输出精确人数或「双向匹配率」；真实市场只统计已注册、主动入池、账户正常的用户。

## 当前状态（2026-11-01 验收）

| 阶段 | 内容 | 状态 |
|---|---|---|
| Phase 1 | 理论市场计算器、基准数据集 CSV 管理、顺序漏斗、Leave-One-Out 敏感性、结果分享 | 已实现并验收通过 |
| Phase 2 | 手机号 + 短信验证码登录、JWT 轮换、资料/择偶条件/价值观、照片、加入婚恋池、管理端问题配置 | 已实现并验收通过（短信为 mock，真实供应商未接入） |
| Phase 3 | 真实候选池、双向硬筛、软评分与解释、真实漏斗/敏感性、K 匿名与冷启动、游标分页 | 已实现并验收通过 |
| Phase 4 | Like/Pass、Match、结果反馈、Block、举报、产品事件、运营后台（仪表盘/漏斗/设置/审计/图片审核） | 已实现并验收通过；聊天按规划未实现（chat_enabled=false） |
| 部署与工程化 | docker compose、Android/iOS 打包、真实 Redis/S3、备份恢复、10 万级性能压测 | **待验证**（本机无 Docker / 原生工具链 / pg 客户端，详见 [07-阶段验收记录](docs/07-阶段验收记录.md)） |

已实现代码：后端 87 个 Go 文件、68 个 HTTP 路由、7 个 goose 迁移；uni-app 23 个页面；运营后台 11 个视图。验收命令与实测结果见 [07-阶段验收记录](docs/07-阶段验收记录.md)。

## 目录结构

```
WMC/
├── backend/            Go 后端（chi + pgx + sqlc + goose + embedded-postgres）
│   ├── cmd/            api / worker / seed / devdb
│   ├── internal/       领域模块（calculator、matching、discovery、interaction、admin ...）+ platform 平台层
│   ├── db/             migrations（内嵌进二进制）、queries、embed.go
│   └── openapi/        openapi.yaml（API 契约）
├── app/uni-app/        uni-app（Vue3 + TS + Vite + Pinia）客户端（H5 已验证构建；Android/iOS 待验证）
├── admin/vue-admin/    Vue3 + TS + Vite + Element Plus 运营后台
├── deploy/             docker-compose、Caddyfile、备份恢复脚本
├── docs/               需求/架构/数据库/API/引擎/部署/工具链/验收文档
└── scripts/            开发脚本（本地数据库、迁移、全量验收）
```

## 1. 环境要求（实测版本）

| 组件 | 版本 | 说明 |
|---|---|---|
| Go | 本机 1.25.6，模块声明 go 1.26.0 | `GOTOOLCHAIN=auto` 会自动下载并使用 Go 1.26 工具链（无需手动安装） |
| Node.js | 25.2.1（要求 ≥ 20） | 前端构建 |
| npm | 11.6.2 | **必须使用国内镜像** `https://registry.npmmirror.com`（仓库已内置 `.npmrc`；本机 npmjs.org 不可达） |
| PostgreSQL | 18.3（embedded-postgres，端口 55432） | 生产/测试 PostgreSQL 18；无 Docker 时用内置便携版，见 §3 |
| Redis | 可选（开发环境可降级内存实现） | 验证码、限流、计算器缓存 |
| Docker | 可选（本机未安装，compose 未实跑） | 一键启动 PostgreSQL/Redis/MinIO/backend |
| PowerShell | Windows PowerShell 5.1 | 仓库脚本用 `powershell -File` 执行；如已安装 PowerShell 7 也可用 `pwsh` |

Go 工具链：goose/sqlc 等工具安装时可能进一步切换 go1.26.10（由 Go 自动管理）。详细安装与失败回退记录见 [06-工具链与依赖安装记录](docs/06-工具链与依赖安装记录.md)。

## 2. 配置

所有配置来自环境变量，**禁止**在代码中硬编码数据库密码、JWT Secret、OSS 密钥。

```bash
cp backend/.env.example backend/.env    # 按需修改
```

关键变量（完整 30 项见 [05-部署运维](docs/05-部署运维.md)）：

| 变量 | 默认值 | 说明 |
|---|---|---|
| `APP_ENV` | development | development / staging / production |
| `DATABASE_URL` | 空（必填） | 本地 devdb：`postgres://wmc:wmc_local_password@localhost:55432/wmc?sslmode=disable` |
| `REDIS_URL` | 空 | 为空则使用内存实现 |
| `RATE_LIMIT_STORE` | auto | redis / memory / auto |
| `JWT_SECRET` | 空 | Access Token 签名密钥；生产必须 ≥32 字节 |
| `PHONE_HASH_SECRET` / `PHONE_ENC_KEY` | 空 | 手机号 HMAC 密钥 / AES-256-GCM 密钥（base64 32 字节）；生产必填 |
| `ACCESS_TOKEN_TTL` / `REFRESH_TOKEN_TTL` | 15m / 720h | Access 15 分钟 / Refresh 30 天 |
| `SMS_PROVIDER` / `SMS_MOCK_CODE` | mock / 123456 | 开发环境固定验证码；生产禁止 mock |
| `STORAGE_DRIVER` | fs | fs（本地盘）/ s3（生产必须 s3） |
| `ADMIN_DEV_LOGIN_ENABLED` | true | 开发管理员登录；生产必须关闭 |
| `MIGRATE_ON_START` | true | 启动时自动迁移 |
| `PG_BINARY_REPO` | `https://maven.aliyun.com/repository/public` | devdb 下载 PostgreSQL 二进制的镜像 |

生产环境启动时会强制校验：弱 JWT/手机号密钥、mock 短信、fs 存储、管理员开发登录一律拒绝启动（已实测）。

## 3. 启动数据库

### 方式 A：Docker Compose（推荐；本机无 Docker，未实跑，待验证）

```bash
cd deploy
cp .env.example .env
docker compose up -d postgres redis minio
# 或全部服务（含 backend、caddy）：
docker compose up -d --build
```

### 方式 B：本机便携 PostgreSQL（无 Docker，已实测可用）

```powershell
# 首次运行会从 Maven 镜像下载 PostgreSQL 二进制并初始化（约 150MB+），保持前台运行
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/dev-db.ps1

# 输出（另开终端使用）：
# DATABASE_URL=postgres://wmc:wmc_local_password@localhost:55432/wmc?sslmode=disable
```

- 实测版本：PostgreSQL **18.3**；端口 **55432**；数据目录 **`backend/.data/pg`**；Ctrl+C 优雅停止，数据保留。
- 便携发行版**不含 pg_dump/psql**，备份恢复需另装 PostgreSQL 客户端工具。
- 也可直接：`cd backend && go run ./cmd/devdb -port 55432 -data .data/pg`。

## 4. 数据库迁移

迁移文件：`backend/db/migrations/00001~00007`（goose 格式），通过 `backend/db/embed.go` 内嵌进二进制。

自动迁移：`MIGRATE_ON_START=true`（默认）时，`cmd/api`、`cmd/worker`、`cmd/seed` 启动都会执行迁移（已实测迁移 00001~00007 全部成功）。

手动迁移：

```powershell
# 前置：已安装 goose v3.29.0（%USERPROFILE%\go\bin\goose.exe）
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/migrate.ps1 up
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/migrate.ps1 status
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/migrate.ps1 version
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/migrate.ps1 down      # 回滚最近一个版本（仅开发）
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/migrate.ps1 reset     # 回滚全部（仅开发，危险）

# 或直接调用 goose：
goose -dir backend/db/migrations postgres "$env:DATABASE_URL" up
```

## 5. 代码生成（sqlc）

已安装 sqlc **v1.31.1**（`C:\Users\<user>\go\bin\sqlc.exe`，官方预编译二进制；本机源码编译因 CGO/TDM-GCC 链接问题失败，详见 06 文档）。生成的代码位于 `backend/internal/db/sqlcdb`，**已提交入库，无 sqlc 也能编译**。

```bash
cd backend
sqlc generate        # 依据 sqlc.yaml 重新生成 internal/db/sqlcdb
```

## 6. 初始化数据（Seed）

```bash
cd backend
$env:DATABASE_URL="postgres://wmc:wmc_local_password@localhost:55432/wmc?sslmode=disable"
go run ./cmd/seed -users 200
```

参数：

| 参数 | 默认 | 说明 |
|---|---|---|
| `-demo` | true | 生成 DEMO 基准数据集并激活 |
| `-demo-records` | 3000 | DEMO 数据集加权样本数 |
| `-admin` | true | 创建开发管理员（手机号 13800000000，super_admin） |
| `-admin-phone` | 13800000000 | 管理员手机号 |
| `-users` | 0 | 生成 N 个 TEST DATA 用户（含完整资料/偏好/价值观并主动入池；手机号 137xxxxxxxx，验证码 123456） |

- 所有 Seed 数据明确标记 TEST DATA / DEMO，**生产环境（APP_ENV=production）执行会被拒绝**。
- Seed 可重复执行：已存在的手机号会跳过；但每次运行会新建一个 DEMO 数据集（开发库会累积，见待办）。

## 7. 启动后端

```powershell
cd backend
$env:DATABASE_URL="postgres://wmc:wmc_local_password@localhost:55432/wmc?sslmode=disable"
go run ./cmd/api      # 默认 :8080
```

- 运维接口：`GET /health/live`、`GET /health/ready`（含 DB 与 Redis 检查）、`GET /metrics`（Prometheus）。
- 后台任务：`go run ./cmd/worker`（每小时清理过期 Refresh Token、每 6 小时执行注销保留期清理）。
- 结构化 JSON 日志输出到 stdout，敏感字段（手机号、JWT、验证码、完整资料）禁止打印。

## 8. 前端

> 本机访问 npmjs.org 超时，仓库已在根目录与各前端目录内置 `.npmrc` 指向 npmmirror。

```bash
# 运营后台（Vue3 + Element Plus）
cd admin/vue-admin
npm install
npm run dev            # http://localhost:5174
npm run build          # 产物 dist/（部署到 /admin）

# H5 客户端（uni-app）
cd app/uni-app
npm install
npm run dev:h5         # 开发服务器 http://localhost:5173
npm run build:h5       # 产物 dist/build/h5
```

实测（2026-11-01）：admin `npm run build` 与 uni-app `npm run build:h5` 均成功（仅有 chunk 体积与 Dart Sass 弃用警告）。

### 构建 Android / iOS（待验证）

uni-app 原生打包需要原生工具链，本机未安装，未验证：

- **HBuilderX**：导入 `app/uni-app`，选择「发行 → 原生App-云打包」，配置包名与证书；
- **CLI 本地打包**：安装 Android SDK / Xcode 后执行 `npm run build:app`，再用 uni-app 离线打包 SDK 生成 APK/IPA。

## 9. 测试

```powershell
# 一键验收：编译 + 静态检查 + 单元测试 + 集成测试（可选） + 前端构建
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/test-all.ps1 `
  -TestDatabaseUrl "postgres://wmc:wmc_local_password@localhost:55432/wmc?sslmode=disable"

# 仅后端
cd backend
go build ./...
go vet ./...
go test -count=1 ./...
$env:TEST_DATABASE_URL="postgres://wmc:wmc_local_password@localhost:55432/wmc?sslmode=disable"
go test -tags=integration -count=1 ./...
```

集成测试使用真实 PostgreSQL（每个测试创建独立数据库，不伪造 SQLite），覆盖：SQL 加权计数与内存实现一致性、漏斗单调、敏感性不减少、分享往返、双向硬筛、Block 隐藏、K 匿名 <10 等。实测结果（全部通过）见 [07-阶段验收记录](docs/07-阶段验收记录.md)。

## 10. 部署（docker compose；本机无 Docker，待验证）

```bash
cd deploy
cp .env.example .env
docker compose up -d --build
# Caddy：/api、/health、/metrics → backend:8080；/admin → 后台静态；/ → H5 静态
```

静态产物放置：

- H5：`app/uni-app/dist/build/h5` → `deploy/static/h5`
- 运营后台：`admin/vue-admin/dist` → `deploy/static/admin`

`backend/Dockerfile` 为多阶段构建，产出 `wmc-api`、`wmc-worker`、`wmc-seed`；入口为 api。生产必须配置真实密钥、Redis、S3 兼容对象存储与真实短信供应商。

## 11. 备份与恢复（脚本已提供；本机无 pg_dump，待实测）

```powershell
# 备份（PostgreSQL custom format；默认保留 7 个日备份 + 4 个周备份）
powershell -NoProfile -ExecutionPolicy Bypass -File ./deploy/scripts/backup.ps1 `
  -DatabaseUrl "postgres://wmc:wmc_local_password@localhost:55432/wmc?sslmode=disable"

# 恢复（--clean --if-exists，会先删除同名对象，属于破坏性操作）
powershell -NoProfile -ExecutionPolicy Bypass -File ./deploy/scripts/restore.ps1 `
  -File deploy/backups/wmc_daily_20260101_030000.dump `
  -DatabaseUrl "postgres://wmc:wmc_local_password@localhost:55432/wmc?sslmode=disable"
```

需要 PATH 中存在 `pg_dump`/`pg_restore`（便携 PostgreSQL 发行版不含客户端工具）；生产建议每日备份上传对象存储并定期演练恢复。

## 12. 主要接口

契约文件：`backend/openapi/openapi.yaml`（OpenAPI 3.1），说明见 [03-API契约](docs/03-API契约.md)。统一响应 `{ code, message, data, request_id }`。

已实现接口按模块（业务前缀 `/api/v1`）：

- **元数据**：`GET /meta/options|regions|compatibility-questions|app-config`
- **认证**：`POST /auth/sms/send|verify`、`POST /auth/refresh|logout`、`POST /admin/auth/dev-login`（开发环境）
- **理论市场**：`POST /calculator/estimate`、`POST /calculator/share`、`GET /calculator/share/{token}`
- **我的资料**：`GET/PUT /me/profile`、`GET/PUT /me/preferences`、`GET/PUT /me/compatibility`、`GET/POST /me/consents`、`GET/POST/DELETE /me/market/{status,join}`、`DELETE /me/account`
- **照片**：`GET/POST /me/photos`、`DELETE /me/photos/{id}`、`GET /media/*`（本地存储签名代理）
- **真实候选池**：`GET /discovery/summary|funnel|sensitivity|candidates|candidates/{id}`
- **互动/Match**：`PUT /interactions/{user_id}`、`GET /matches`、`GET/DELETE /matches/{id}`、`POST /matches/{id}/outcomes`、`POST/DELETE /users/{id}/block`、`POST /users/{id}/report`、`POST /events`
- **运营后台**：`/admin/users`（列表/详情/举报/冻结/解冻）、`/admin/dashboard|funnel|reports|photos|settings|audit-logs`、`/admin/baseline/datasets`（列表/创建/导入/激活/回滚/演示数据）、`/admin/questions`
- **运维**：`GET /health/live|ready`、`GET /metrics`

未实现（规划内）：聊天 IM、推送、支付、实名/学历/收入认证、LLM 推荐等 Phase 5 内容。

## 13. 产品与合规约定

- 理论市场只展示估算比例，**禁止**展示精确人数与双向匹配率；演示数据集必须显著标注 DEMO；
- 真实候选数小于 K 匿名阈值（默认 10）统一显示 `<10`；冷启动不显示「0 个人」，展示早期测试文案；
- 理论市场与真实市场在页面与接口中严格区分；
- 不采集身份证号、精确住址、GPS 坐标、银行卡、工资流水；收入/资产/负债使用区间；
- 手机号只存 HMAC 摘要与 AES-GCM 密文；
- 日志禁止输出手机号、JWT、验证码、完整个人资料；
- 文案禁止价值判断（「剩女」「高价值男人」等），只陈述数据。

## 14. 已知问题与待办

完整清单见 [07-阶段验收记录 §4](docs/07-阶段验收记录.md)。当前主要项：

1. `CandidateDetail` 未返回 OpenAPI 定义的 `disclaimer` 字段；
2. worker 注销清理 SQL 依赖 `pgcrypto.digest`，但迁移未创建该扩展（保留期清理暂不可用）；
3. Block 后已存在的 Match 未同步为「已屏蔽」，Match 列表未按 Block 过滤；
4. 候选列表仅在 `scanLimit=500` 内计算软分并内存排序（规模化优化待做）；
5. Seed 每次运行会新建 DEMO 数据集（不幂等）；
6. 管理接口无限流；生产建议网关层补充；
7. docker compose、Android/iOS、真实 Redis/S3/短信、备份恢复、10 万级 EXPLAIN ANALYZE 与 P95 压测均待验证。

## 15. 文档索引

- [00-需求范围](docs/00-需求范围.md)
- [01-架构设计](docs/01-架构设计.md)
- [02-数据库设计](docs/02-数据库设计.md)
- [03-API契约](docs/03-API契约.md)
- [04-匹配与计算引擎](docs/04-匹配与计算引擎.md)
- [05-部署运维](docs/05-部署运维.md)
- [06-工具链与依赖安装记录](docs/06-工具链与依赖安装记录.md)
- [07-阶段验收记录](docs/07-阶段验收记录.md)
