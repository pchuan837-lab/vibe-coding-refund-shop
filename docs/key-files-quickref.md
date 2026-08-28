# key-files-quickref · 关键文件速查表（S-A 第 4 步 · 新项目简化版承诺卡）

> 对齐 vibeCoding 共享单元 S-A「第 4 步 承诺速查表」。
> 5 个关键文件：改之前先查这张表，避免改错层、改错入口。

## 4 列速查表

| 模块 | 位置 | 角色 | 进入端点 / 入口 |
|---|---|---|---|
| 后端入口 | `main.go` | 装配：连库→建路由→挂静态 | `go run .` 第一行；端口 3000 |
| 订单 | `internal/routes/orders.go` | `POST/GET /api/orders`、`GET /api/orders/:id` | 前端下单页 `index.html` |
| 退款 | `internal/routes/refunds.go` | `POST/GET /api/refunds`、`PATCH /api/refunds/:id/approve` | 前端订单页 `orders.html` / 审核页 `admin.html` |
| 规则 | `internal/domain/refund_rules.go` | `CalcRemaining` 纯函数，算可退金额 | 被 `refunds.go` 的申请退款调用 |
| 数据 | `internal/db/db.go` + `schema.sql` | 打开 SQLite、建 `orders/refunds` 两表 | 所有路由的写操作 |

## 一句话承诺卡

- **数据放哪** → `schema.sql`（建表）+ `internal/db/db.go`（连库）。
- **入口在哪** → `main.go`，一条 `go run .` 起全栈。
- **坑在哪** → 金额用分存储；`CHECK(amount>0)` 约束别删；测试用 `:memory:` 别碰 `data.db`。