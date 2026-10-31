# API 契约

契约文件：`backend/openapi/openapi.yaml`（OpenAPI 3.1）。前端 TypeScript 类型由 `openapi-typescript` 生成，禁止手写重复类型。

## 统一响应

成功：

```json
{ "code": "OK", "message": "", "data": {}, "request_id": "..." }
```

失败：

```json
{ "code": "PROFILE_INCOMPLETE", "message": "请先完善个人资料", "request_id": "..." }
```

## 错误码

| code | HTTP | 说明 |
|---|---|---|
| INVALID_ARGUMENT | 400 | 参数校验失败 |
| UNAUTHORIZED | 401 | 未登录 / Token 失效 |
| FORBIDDEN | 403 | 权限不足 / 未入池 / 账号受限 |
| NOT_FOUND | 404 | 资源不存在或不可见 |
| CONFLICT | 409 | 状态冲突 |
| PROFILE_INCOMPLETE | 409 | 资料完成度不足 |
| DATASET_UNAVAILABLE | 409 | 无可用基准数据集 |
| RATE_LIMITED | 429 | 触发限流 |
| INTERNAL | 500 | 服务端错误 |
| FEATURE_DISABLED | 403 | Feature Flag 已关闭 |

## 端点清单

见 `docs/00-需求范围.md` 对应阶段与 `backend/openapi/openapi.yaml` 全量定义。所有业务接口前缀 `/api/v1`；运维接口 `/health/*`、`/metrics` 不带前缀。
