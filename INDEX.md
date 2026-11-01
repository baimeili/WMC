# 项目索引（PROJECT INDEX）

> 面向新开发者与 AI Agent 的入口文档。读完本文即可知道：**系统是什么、代码在哪、怎么跑、怎么改、怎么验**。
> 更详细的设计与验收记录见文末「文档地图」。

## 0. 30 秒速览

| 项 | 值 |
|---|---|
| 产品 | 婚恋市场计算器（理论市场估算）+ 双向婚恋匹配（本站真实用户池）MVP |
| 形态 | Modular Monolith 后端 + 两个前端（H5 客户端 / 运营后台），Monorepo |
| 后端 | Go 1.26（工具链）、chi、pgx/v5、sqlc、goose、Redis、JWT、slog、Prometheus |
| 数据库 | PostgreSQL 18（本机用 embedded-postgres 便携版，端口 55432） |
| 客户端 | uni-app（Vue3 + TS + Vite + Pinia + SCSS），H5 已验证构建 |
| 后台 | Vue3 + TS + Vite + Element Plus |
| 启动 | `powershell -File ./scripts/start.ps1`（幂等，不会重复起进程） |
| 冒烟验收 | `node scripts/smoke.mjs`（46 项端到端检查，当前 46/46 通过） |
| 端口 | API 8080 / PG 55432 / H5 5173 / 后台 5174 |
| 演示账号 | 管理员 `13800000000`（后台开发登录）；测试用户 `13700000150` 等，验证码 `123456` |
| 当前状态 | Phase 1-4 已实现并验收；chat/payments/扩展认证、Docker 端到端、移动端打包待验证 |

## 1. 目录地图

```
WMC/
├── INDEX.md                  ← 本文（新手/Agent 入口）
├── README.md                 安装、配置、迁移、seed、构建、部署、备份
├── 提示词.txt                原始产品需求（v0.1 规格）
├── backend/                  Go 后端
│   ├── cmd/
│   │   ├── api/              HTTP 服务入口（依赖装配在 main.go 的 buildRouter）
│   │   ├── worker/           后台任务：过期 Token 清理、注销保留期清理
│   │   ├── seed/             演示数据：DEMO 基准数据集、管理员、200 个 TEST DATA 用户
│   │   └── devdb/            无 Docker 环境下的便携 PostgreSQL 启动器
│   ├── db/
│   │   ├── migrations/       7 个 goose 迁移（全量建表 + 种子问题/参数/地区）
│   │   ├── queries/          13 个 sqlc SQL 文件（生成 internal/db/sqlcdb）
│   │   └── embed.go          迁移文件内嵌进二进制
│   ├── openapi/openapi.yaml  API 契约（OpenAPI 3.1）
│   ├── internal/             领域模块（见 §3）
│   └── sqlc.yaml             sqlc 生成配置（uuid/数组/jsonb 类型映射）
├── app/uni-app/              H5/App 客户端
│   └── src/{api,components,pages,stores,styles,utils}
├── admin/vue-admin/          运营后台
│   └── src/{api,components,layouts,router,stores,styles,utils,views}
├── deploy/                   docker-compose、Caddyfile、备份/恢复脚本
├── docs/                     需求/架构/数据库/API/引擎/部署/工具链/验收/界面设计
│   └── screenshots/          H5 页面真实截图（Playwright 390×844 @2x）
├── scripts/                  一键启动/停止、迁移、全量验收、端到端冒烟
└── .data/                    运行时数据（数据库、上传文件、日志、PID），已 gitignore
```

## 2. 快速开始

```powershell
# 1) 启动数据库 + 后端 + H5 + 运营后台（已运行的组件自动跳过）
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/start.ps1

# 2) 端到端冒烟（46 项，覆盖游客计算 → 注册 → 入池 → 匹配 → 反馈 → 后台）
node scripts/smoke.mjs

# 3) 停止全部（数据库优雅关闭，数据保留）
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/stop.ps1
```

访问：

| 组件 | 地址 | 说明 |
|---|---|---|
| H5 客户端 | http://localhost:5173 | 必须用 `localhost`（Vite 只监听 IPv6 ::1） |
| 运营后台 | http://localhost:5174/admin/ | 管理员手机号 13800000000 |
| 后端 API | http://127.0.0.1:8080 | `/health/live`、`/health/ready`、`/metrics` |
| PostgreSQL | 127.0.0.1:55432 | `postgres://wmc:wmc_local_password@localhost:55432/wmc` |

## 3. 后端模块职责

| 包 | 职责 | 关键文件 |
|---|---|---|
| `internal/platform/*` | 基础设施：配置、DB 连接池/迁移、HTTP 响应与中间件、错误码、日志、指标、Redis/内存存储、限流、手机号加密、枚举、游标 | `platform/config/config.go`、`platform/httpx/*`、`platform/dbx/*` |
| `internal/auth` | 短信验证码登录（mock Provider）、JWT 双 Token 轮换/重用检测、管理员开发登录、RBAC 中间件 | `auth/service.go`、`auth/sms.go`、`auth/middleware.go` |
| `internal/user` | 资料、完成度、择偶条件、价值观规则、入池状态与校验、注销 | `user/service.go` |
| `internal/profile` | 资料完成度算法与年龄计算（纯函数，含单测） | `profile/completion.go` |
| `internal/calculator` | **理论市场引擎**：加权计数、9 层顺序漏斗、Leave-One-Out 敏感性、缓存、计算记录与分享 | `calculator/engine.go`、`calculator/sql_counter.go` |
| `internal/matching` | **双向匹配引擎**：硬条件双向判定、软分、解释（纯函数，重点单测） | `matching/engine.go` |
| `internal/discovery` | 真实候选池：SQL 双向硬筛、软分排序、K 匿名、漏斗、敏感性、候选人详情 | `discovery/service.go` |
| `internal/interaction` | Like/Pass、Match（唯一约束）、结果反馈、Block、举报、产品事件 | `interaction/service.go` |
| `internal/photo` | 照片上传：MIME/尺寸校验、重编码去 EXIF/GPS、签名 URL、主图逻辑 | `photo/service.go` |
| `internal/baseline` | 基准数据集：创建、CSV 校验/导入/激活/回滚、DEMO 数据 | `baseline/service.go` |
| `internal/meta` | 选项字典、行政区划、价值观问题、前端配置 | `meta/service.go` |
| `internal/admin` | 运营后台：看板、用户冻结、举报处理、图片审核、系统参数、审计日志 | `admin/service.go`、`admin/dashboard.go` |
| `internal/settings` | 系统参数与 Feature Flag（带 15s 缓存） | `settings/service.go` |
| `internal/storage` | 对象存储抽象：`fs`（本地+签名代理）与 `s3`（预签名） | `storage/storage.go`、`storage/s3.go` |
| `internal/sms` | 短信供应商接口 + mock 实现（生产禁用 mock） | `sms/sms.go` |
| `internal/seed` | TEST DATA 生成（200 用户、DEMO 数据集） | `seed/users.go`、`seed/demo.go` |
| `internal/store` | sqlc Queries 聚合 + 事务/CopyFrom 辅助 + 参数加载 | `store/store.go` |
| `internal/db/sqlcdb` | sqlc 生成代码（勿手改，改 `db/queries/*.sql` 后重新生成） | — |
| `internal/events` | 产品事件写入 | `events/events.go` |

### 请求链路

```
HTTP → chi 路由 → middleware(RequestID → AccessLog → Recover → Timeout → CORS → RateLimit → Auth)
     → handler（解析/校验）→ service（业务规则）→ store.Store（sqlc + 动态 SQL）→ PostgreSQL
                                             ↘ 纯函数引擎：calculator.Engine / matching.Engine
```

- 统一响应：`{ code, message, data, request_id }`；错误码定义在 `platform/apperr`。
- 鉴权：`Authorization: Bearer <access_token>`；签名入口 `cmd/api/main.go` → `buildRouter`。
- 后台路由统一加 `RequireRole(2)`（moderator 及以上）与限流。

## 4. 数据库

- 迁移：`backend/db/migrations/00001..00007`（users/auth → profiles → compatibility → baseline → social → ops → indexes）。
- 一次性建全所有表；`go run ./cmd/api` 启动时自动迁移（`MIGRATE_ON_START=true`）。
- 关键约束：`interactions(actor,target)` 唯一；`matches` 有序对唯一（永不重复）；`baseline_datasets` 激活唯一（部分索引）；手机号只存 HMAC + AES-GCM 密文。
- 改表流程：新增迁移文件 → `sqlc generate`（或改 `db/queries` 后生成）→ 改 service → 补集成测试。

## 5. 前端（uni-app）

| 层 | 位置 | 规则 |
|---|---|---|
| API | `src/api/*.ts` | 唯一网络出口 `http.ts`；页面禁止裸 `uni.request`；统一解包 envelope；401 自动 refresh 重放一次 |
| Store | `src/stores/*.ts` | auth / calculator / profile / discovery；Token 持久化到 uni storage |
| 页面 | `src/pages/**` + `src/pages.json` | 23 个页面；路由必须同时存在于 pages.json 与磁盘 |
| 组件 | `src/components/*.vue` | PageShell、StepHeader、MarketBadge、FunnelChart、SensitivityList、CandidateCard、EmptyState、ScoreBar、OptionChips、HardToggle |
| 样式 | `src/uni.scss`（设计令牌）+ `src/styles/global.scss`（全局类） | **全局类必须在 App.vue 引入**，放进普通组件会被 uni-app 加 data-v 作用域导致失效 |

产品文案规则：理论市场禁止精确人数/双向率；真实池 <10 显示 `<10`；分数固定叫「条件契合度/关系观念兼容度/综合匹配度」并附免责声明；不使用价值判断词汇。

## 6. 运营后台

- 视图：`admin/vue-admin/src/views/`（Dashboard、Funnel、Users、UserDetail、Reports、Photos、Baseline、BaselineDetail、Questions、Settings、AuditLogs、Login）。
- API 封装：`admin/vue-admin/src/api/*.ts`（统一 axios 实例与 401 处理）。
- 角色：moderator/admin/super_admin 可登录；路由守卫校验本地 token 与角色。

## 7. 测试与验收

| 类型 | 命令 | 覆盖 |
|---|---|---|
| 后端单元 | `cd backend && go test ./...` | 计算器引擎、匹配引擎、令牌、完成度、限流、游标、设置缓存 |
| 后端集成（真实 PG） | `$env:TEST_DATABASE_URL='postgres://wmc:wmc_local_password@localhost:55432/wmc?sslmode=disable'; go test -tags=integration ./internal/...` | SQL 与内存实现一致性、漏斗单调、LOO 单调、双向硬筛、Block、并发唯一 Match、确认见面 |
| 端到端冒烟 | `node scripts/smoke.mjs` | 游客计算→分享→注册→资料→偏好→价值观→照片审核→入池→候选池→Like/Match→反馈→Block/举报→后台看板（46 项） |
| 前端构建 | `cd app/uni-app && npm run build:h5`；`cd admin/vue-admin && npm run build` | H5 与后台生产构建 |
| 一键回归 | `powershell -File ./scripts/test-all.ps1` | 编译+静态检查+单测+前端构建 |

截图复核：`docs/screenshots/`（Playwright 390×844 @2x）；截图脚本在 `.tools/ui-shot/`（本地工具，不入库）。

## 8. 调试手册（高频问题）

| 现象 | 原因 | 处理 |
|---|---|---|
| H5 页面打不开 / 空白 | 访问 `127.0.0.1:5173`，而 Vite 只监听 IPv6 | 改用 `http://localhost:5173` |
| 全局样式不生效 | 样式写进普通组件被 scoped | 移到 `src/styles/global.scss`，由 `App.vue` 引入 |
| SCSS 报 Undefined variable | 在 `App.vue` 使用 `$wmc-*` 但未注入 uni.scss | `App.vue` 顶部 `@import './uni.scss'` |
| 短信验证码发不出 | 同号 60 秒冷却 / 每日 5 次；或测试期已用掉 | 等 60 秒或换手机号；开发环境验证码固定 123456 |
| 登录后仍是未登录态 | localStorage 缺少 `wmc:auth:*` 或 token 过期 | 重新登录；检查 `utils/storage.ts` 键名 |
| 候选池为空但后台显示有用户 | 未入池 / 资料完成度 <80 / 未同意三项协议 / 城市等硬条件不满足 | 看 `GET /me/market/status` 或前端入池提示 |
| 点击候选卡进详情 id=undefined | 组件 tap 事件与原生冒泡冲突（历史缺陷） | 组件用 `select` 事件 + `@tap.stop`（已修复，勿回退） |
| 接口 409 DATASET_UNAVAILABLE | 没有激活的基准数据集 | `go run ./cmd/seed -demo` 或后台激活数据集 |
| 接口 403 FEATURE_DISABLED | 系统参数关闭了对应功能 | 后台「系统参数」或 `system_settings` 打开 |
| 迁移失败 | goose 语句分隔/残留 postmaster.pid | 看 `.data/logs/api.log`；重启脚本会自动清理残留 pid |
| sqlc 生成失败 | 类型映射不支持 | 调整 `backend/sqlc.yaml` overrides，别手改生成代码 |

日志：`.data/logs/{api,db,h5,admin}.log`；PID：`.data/pids/`；数据库文件：`backend/.data/pg`。

## 9. 修改指南（常见任务）

| 任务 | 步骤 |
|---|---|
| 新增 API | openapi.yaml 加契约 → handler 注册路由 → service 业务 → 前端 api 模块 → 冒烟/测试 |
| 新增数据库字段 | 新迁移文件 → 改 queries → `sqlc generate` → service/DTO → 前端类型 → 集成测试 |
| 修改匹配规则 | `internal/matching/engine.go`（纯函数）+ 单测；同步 `docs/04-匹配与计算引擎.md` |
| 新增价值观问题 | 后台「问题配置」或迁移插入；前端自动渲染（问题不进客户端代码） |
| 新增 H5 页面 | `src/pages/**` + `src/pages.json` 注册 + 导航入口 + 截图复核 |
| 调整主题 | `src/uni.scss` 令牌 + `src/styles/global.scss` 组件类 |
| 调整后台指标 | `internal/admin/dashboard.go` + `admin/vue-admin/src/views/DashboardView.vue` |

## 10. Agent 工作约定

1. 注释、文档、提交信息一律中文；提交按功能粒度（`feat(模块): ...`）。
2. 改后端后至少跑：`go build ./...`、`go vet ./...`、`go test ./...`；涉及 SQL 再跑 `-tags=integration`。
3. 改前端后至少跑对应 `npm run build`；UI 改动需用 Playwright 截图自查。
4. 不要手改 `internal/db/sqlcdb`；不要提交 `.data/`、`node_modules/`、`backend/bin/`。
5. 生产环境硬性校验：mock 短信、fs 存储、开发登录、弱密钥都会拒绝启动，不要为了本地方便放宽。
6. 起服务用 `scripts/start.ps1`（幂等）；用完 `scripts/stop.ps1`，避免重复进程占用端口。
7. 本轮已知契约差异/保留接口：`GET /me/market/status`、`GET /admin/funnel`、`POST /admin/baseline/datasets/demo` 前端未直接调用（保留给脚本/第三方）。

## 11. 文档地图

| 文档 | 内容 |
|---|---|
| [README.md](README.md) | 安装、配置、迁移、seed、启动、测试、构建、部署、备份恢复 |
| [docs/00-需求范围.md](docs/00-需求范围.md) | 需求范围与阶段划分 |
| [docs/01-架构设计.md](docs/01-架构设计.md) | 架构与分层 |
| [docs/02-数据库设计.md](docs/02-数据库设计.md) | 21 张表、字段编码、约束与索引 |
| [docs/03-API契约.md](docs/03-API契约.md) | 响应规范与错误码 |
| [docs/04-匹配与计算引擎.md](docs/04-匹配与计算引擎.md) | 两个引擎算法、口径、边界用例 |
| [docs/05-部署运维.md](docs/05-部署运维.md) | 环境变量、compose、备份恢复、排障 |
| [docs/06-工具链与依赖安装记录.md](docs/06-工具链与依赖安装记录.md) | 本机实测环境、安装命令、镜像与回退 |
| [docs/07-阶段验收记录.md](docs/07-阶段验收记录.md) | 每阶段命令、结果、缺陷与未验证项 |
| [docs/08-H5界面设计说明.md](docs/08-H5界面设计说明.md) | 设计令牌、页面改造、截图与缺陷修复 |
| [docs/09-闭环检查报告.md](docs/09-闭环检查报告.md) | 本轮闭环检查方法、46 项结果、剩余缺口 |
| [backend/openapi/openapi.yaml](backend/openapi/openapi.yaml) | API 契约 |
| [docs/screenshots/](docs/screenshots/) | H5 页面截图 |

## 12. 当前边界（未闭环/未验证）

- 即时聊天（`chat_enabled=false`）、支付、扩展认证（身份/学历/收入/婚史）按规划未实现；
- Docker Compose 全链路、真实 Redis/S3/MinIO/短信供应商未实机联调；
- Android/iOS 原生打包未执行（仅 H5 构建验证）；
- 10 万级数据下的 P95 压测未做（已给出索引与 EXPLAIN ANALYZE 方法）；
- 浏览器端交互仅做页面渲染与截图验证，未做断言式 E2E 测试套件。
