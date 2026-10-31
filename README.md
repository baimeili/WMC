# 婚恋市场计算器（WMC）

> 先让 1000 个人真正可以使用，再考虑 100 万人。

婚恋市场计算器 + 双向婚恋匹配 MVP。它不是刷脸 Dating App，而是先回答两个问题：

1. 你的择偶条件，在**理论人口市场**中筛掉了多少人？（统计估算）
2. 在**本站真实用户池**里，有多少人和你**双向**满足硬条件？

## 目录结构

```
WMC/
├── backend/            Go 1.26 + chi + pgx + sqlc + goose 的后端服务
├── app/uni-app/        uni-app(Vue3 + TS + Vite + Pinia + uni-ui) 客户端（H5/Android/iOS）
├── admin/vue-admin/    Vue3 + TypeScript + Vite + Element Plus 运营后台
├── deploy/             docker-compose、Caddyfile、备份恢复脚本
├── docs/               需求、架构、数据库、API、引擎、部署、验收文档
└── scripts/            开发脚本（本地数据库、迁移、全量验收）
```

## 1. 环境要求

| 组件 | 版本 | 说明 |
|---|---|---|
| Go | 1.26+ | 规格原文为 1.27.x，当前工具链为 1.25/1.26 系列，使用 1.26 编译 |
| Node.js | 20+ | 前端构建 |
| PostgreSQL | 18 | 生产/测试；本机无 Docker 时可用内置便携版（见下文） |
| Redis | 7 | 验证码、限流、缓存；开发环境可降级为内存实现 |
| Docker | 可选 | 一键启动 PostgreSQL/Redis/MinIO/后端 |

## 2. 配置

所有配置来自环境变量，**禁止**在代码中硬编码数据库密码、JWT Secret、OSS 密钥。

```bash
cp backend/.env.example backend/.env    # 按需修改
```

关键变量：

| 变量 | 说明 |
|---|---|
| `APP_ENV` | development / staging / production |
| `DATABASE_URL` | PostgreSQL 连接串 |
| `REDIS_URL` | Redis 连接串（为空则用内存实现） |
| `JWT_SECRET` | Access Token 签名密钥（生产必须 ≥32 字节） |
| `PHONE_HASH_SECRET` | 手机号 HMAC 密钥 |
| `PHONE_ENC_KEY` | 手机号 AES-256-GCM 密钥（base64，32 字节） |
| `SMS_PROVIDER` | mock（仅开发）/ 云厂商名称 |
| `STORAGE_DRIVER` | fs（开发）/ s3（生产） |
| `ADMIN_DEV_LOGIN_ENABLED` | 开发环境管理员登录，生产必须关闭 |

生产环境启动时会强制校验：mock 短信、fs 存储、开发登录、弱密钥一律拒绝启动。

## 3. 启动数据库

### 方式 A：Docker Compose（推荐）

```bash
cd deploy
cp .env.example .env
docker compose up -d postgres redis minio
```

### 方式 B：本机便携 PostgreSQL（无 Docker 环境）

```bash
# 首次运行会自动下载 PostgreSQL 二进制并初始化（Maven 镜像）
pwsh ./scripts/dev-db.ps1
# 输出：DATABASE_URL=postgres://wmc:wmc_local_password@localhost:55432/wmc?sslmode=disable
```

## 4. 数据库迁移

迁移文件位于 `backend/db/migrations`（goose 格式）。后端启动时默认自动迁移（`MIGRATE_ON_START=true`），也可手动执行：

```bash
go install github.com/pressly/goose/v3/cmd/goose@v3.29.0
pwsh ./scripts/migrate.ps1 up        # 应用迁移
pwsh ./scripts/migrate.ps1 status    # 查看状态
pwsh ./scripts/migrate.ps1 down      # 回滚最近一次
```

## 5. 代码生成（sqlc）

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1
cd backend && sqlc generate
```

生成的代码位于 `backend/internal/db/sqlcdb`，已提交入库（无 sqlc 也能编译）。

## 6. 启动后端

```bash
cd backend
go run ./cmd/api
# 监听 :8080
curl http://127.0.0.1:8080/health/live
curl http://127.0.0.1:8080/health/ready
```

## 7. 初始化数据

```bash
cd backend
# 生成演示基准数据集（DEMO，非真实统计）+ 200 个测试用户（TEST DATA）
go run ./cmd/seed -demo -users 200
```

生产环境（`APP_ENV=production`）执行 seed 会被拒绝。

## 8. 前端

```bash
# H5
cd app/uni-app
npm install
npm run dev:h5        # 开发服务器 http://localhost:5173
npm run build:h5      # 构建产物 dist/build/h5

# 运营后台
cd admin/vue-admin
npm install
npm run dev           # http://localhost:5174
npm run build         # 构建产物 dist（部署到 /admin）
```

### 构建 Android / iOS

uni-app 的 APP 打包需要原生工具链，二选一：

- **HBuilderX**：导入 `app/uni-app`，选择「发行 → 原生App-云打包」，配置包名与证书。
- **CLI 本地打包**：安装 Android SDK / Xcode 后执行 `npm run build:app`，再用 uni-app 离线打包 SDK 生成 APK/IPA。

H5 构建已在本仓库验证；原生打包需在具备对应工具链的机器上执行。

## 9. 测试

```bash
pwsh ./scripts/test-all.ps1                 # 编译 + 静态检查 + 单元测试 + 前端构建
cd backend && go test ./...                 # 仅单元测试
cd backend && go test -tags=integration ./...  # 集成测试（需要真实 PostgreSQL）
```

集成测试要求真实 PostgreSQL（不使用 SQLite 冒充）。设置 `TEST_DATABASE_URL` 或先启动本地库。

测试覆盖核心规则：
- A 满足 B 但 B 不满足 A → 不 Eligible
- 硬条件冲突 → 不 Eligible（软条件不得补偿）
- 软条件冲突 → 仍 Eligible
- 年龄/身高边界、空答案、数组条件、Block 后不可见
- 漏斗逐层收紧候选数不增加
- 敏感性分析取消条件后候选数不减少

## 10. 部署

```bash
cd deploy
docker compose up -d --build
# Caddy: /api → backend:8080，/admin → 运营后台静态文件，/ → H5
```

构建产物放置：

- H5：`app/uni-app/dist/build/h5` → `deploy/static/h5`
- 运营后台：`admin/vue-admin/dist` → `deploy/static/admin`

## 11. 备份与恢复

```bash
# 备份（自定义格式，默认保留 7 日 + 4 周）
pwsh ./deploy/scripts/backup.ps1 -DatabaseUrl "postgres://wmc:pass@localhost:5432/wmc?sslmode=disable"

# 恢复
pwsh ./deploy/scripts/restore.ps1 -File deploy/backups/wmc_daily_20260101_030000.dump -DatabaseUrl "postgres://wmc:pass@localhost:5432/wmc?sslmode=disable"
```

生产建议：每日备份上传对象存储，并定期演练恢复。

## 12. 主要接口

详见 `backend/openapi/openapi.yaml` 与 `docs/03-API契约.md`。

- `POST /api/v1/calculator/estimate` 理论市场估算（游客可用）
- `POST /api/v1/auth/sms/send|verify`、`POST /api/v1/auth/refresh`
- `GET/PUT /api/v1/me/profile|preferences|compatibility`
- `GET /api/v1/discovery/{summary,funnel,sensitivity,candidates}`
- `PUT /api/v1/interactions/{user_id}`、`GET /api/v1/matches`
- `GET /health/live|ready`、`GET /metrics`

## 13. 产品与合规约定

- 理论市场只展示估算比例，**禁止**展示精确人数与双向匹配率
- 真实候选数小于 K 匿名阈值（默认 10）统一显示 `<10`
- 理论市场与真实市场在页面与接口中严格区分
- 不采集身份证号、精确住址、GPS、银行卡、工资流水
- 日志禁止输出手机号、JWT、验证码、完整个人资料
- 文案禁止价值判断（“剩女”“高价值男人”等），只陈述数据

## 14. 文档索引

- [需求范围](docs/00-需求范围.md)
- [架构设计](docs/01-架构设计.md)
- [数据库设计](docs/02-数据库设计.md)
- [API 契约](docs/03-API契约.md)
- [匹配与计算引擎](docs/04-匹配与计算引擎.md)
- [部署与运维](docs/05-部署运维.md)
- [工具链与依赖安装记录](docs/06-工具链与依赖安装记录.md)
- [阶段验收记录](docs/07-阶段验收记录.md)
