# api-reference · refund-shop 接口文档（S-C 文档交付 · 与代码可断言）

> 本文档与 `internal/routes/` 源码**逐字一致**。改任何路由字段前，请同步更新这里。
> 金额单位：**所有 amount / shipping / coupon_used 一律「分」（整数）**，前端展示时再 /100。

---

## 一、路由清单（6 条）

| 方法 | 路径 | 用途 | 权限 |
|---|---|---|---|
| POST | `/api/orders` | 下单 | - |
| GET | `/api/orders` | 订单列表 | - |
| GET | `/api/orders/:id` | 订单详情（含关联退款） | - |
| POST | `/api/refunds` | 申请退款 | - |
| GET | `/api/refunds` | 退款列表（支持 `?status=` / `?order_id=`） | - |
| PATCH | `/api/refunds/:id/approve` | 审批退款请求体 | 通过/拒绝 |

## 二、使用说明（前置条件）
1. 服务：`go run .`，默认 `http://localhost:3000`，API 前缀 `/api`。
2. Content-Type：`application/json`。
3. 订单状态机：`paid → partial_refunded → fully_refunded`。
   退款状态机：`pending → approved / rejected`。
4. 退货前提：订单已存在，且累计已批退款额 < 订单 `amount`（再多会被规则拦截或截断，见 `domain` 规则）。

## 三、金额 / 字段注意事项
- **分**：`amount:9900` 表示 99.00 元。
- `shipping`（运费）默认 0；`coupon_used`（用券）默认 0。两者语义边界是 B1 教学模糊点，`domain/refund_rules.go` 顶部有 TODO。
- 超卖防护：下单 `amount > 0` 且非空 `product_name` 才接受；否则 400。

---

## 逐条接口详表

### 1. POST /api/orders
请求体：
```json
{ "product_name": "保温杯", "amount": 9900, "shipping": 500, "coupon_used": 0 }
```
响应 `200`：
```json
{ "id": 1, "product_name": "保温杯", "amount": 9900, "shipping": 500, "coupon_used": 0, "status": "paid", "created_at": "..." }
```
`400`：`{"error":"amount must be > 0 (cents)"}` / `{"error":"product_name must not be empty"}`

### 2. GET /api/orders
响应 `200`：`Order` 数组（按 `created_at DESC, id DESC`）。

### 3. GET /api/orders/:id
响应 `200`：
```json
{
  "order": { "id":1,"product_name":"保温杯","amount":9900,"shipping":500,"coupon_used":0,"status":"paid","created_at":"..." },
  "refunds": [ { "id":1,"amount":2000,"status":"pending","reason":"..." } ]
}
```
`404`：`{"error":"order not found"}`

### 4. POST /api/refunds
请求体：
```json
{ "order_id": 1, "amount": 2000, "reason": "试喝不满意" }
```
响应 `200`：完整 `Refund`（含 `id/order_id/amount/status='pending'/reason/created_at/updated_at`）。
`400`：refund amount 非正 / 订单不存在 / 超额被规则拦截。

### 5. GET /api/refunds
查询参数（可组合）：`?status=approved` / `?order_id=1`。
响应 `200`：`Refund` 数组（`created_at DESC, id DESC`）。

### 6. PATCH /api/refunds/:id/approve
请求体：`{ "approved": true }`（`false` = 拒绝，可带 `"comment"`）。
行为：通过 → 退款 `status='approved'` 且联动订单状态（`partial/fully_refunded`）；拒绝 → `rejected`。
响应 `200`：`{ "id":.., "order_id":.., "status":"approved|rejected", "updated_at":".." }`。
`400`：非 `pending` 状态不能改；`404`：退款不存在。

---

## 四、测试锚定
`internal/routes/*_test.go` 使用 `httptest`；断言与本文档字段一致。改动契约请同步三处：**本文件 ↔ *_test.go ↔ 前端 fetch**。