# design-skeleton · B1② 七段方案文档（样本 + 空模板）

> 对齐 vibeCoding 共享单元 B「需求澄清」之「② 七段方案文档」。
> 用途两种：**①样本**（本次 refund-shop 完整做好的那份）给你参照；**②空模板**（下划线 `___` 处）留给学员自己项目用。
> 你上手时：先读样本段 4/5（已预填），其余段对比本项目的实际做法理解即可。

---

## 样本：refund-shop 本次实现（已填好）

### 1. 问题 / 目标
给新手一个「能 3 条命令跑起来、能照着 S-A/B/C 三章练」的售后退款教学项目。目标是演示"从需求到验收"全流程，刻意留出 B1 模糊点 / B2 Bug / B3 坏味道供练习。

### 2. 范围（本次做 / 不做）
- **做**：下单 → 申请退款 → 审批 → 订单状态联动；6 条 REST 接口；三页前端；SQLite 持久化。
- **不做**：真实支付 / 库存 / 运费单 / 分页 / 多用户权限 / 退款到账。

### 3. 约束
- Go 1.21+；Gin；modernc SQLite（零 CGO）。金额一律用**分**存整数。
- 验收线：新克隆 3 命令可跑（`go mod download` / `go run .` / `go test ./...`）。

### 4. 关键数据模型（已预填）
`orders` 表：

| 字段 | 类型 | 说明 |
|---|---|---|
| id | INTEGER PK | 自增 |
| product_name | TEXT | 商品名 |
| amount | INTEGER | 订单金额（分） |
| shipping | INTEGER DEFAULT 0 | 运费（分） |
| coupon_used | INTEGER DEFAULT 0 | 用券（分） |
| status | TEXT DEFAULT 'paid' | paid / partial_refunded / fully_refunded |
| created_at | DATETIME DEFAULT CURRENT_TIMESTAMP | 创建时间 |

`refunds` 表：

| 字段 | 类型 | 说明 |
|---|---|---|
| id | INTEGER PK | 自增 |
| order_id | INTEGER FK→orders.id | 关联订单 |
| amount | INTEGER | 申请/已退金额（分） |
| reason | TEXT | 原因 |
| status | TEXT DEFAULT 'pending' | pending / approved / rejected |
| created_at | DATETIME ... | 申请时间 |

### 5. 接口定义（已预填，6 条）
| 方法 | 路径 | 请求体 | 响应 |
|---|---|---|---|
| POST | /api/orders | {product_name,amount,shipping,coupon_used} | Order |
| GET | /api/orders | - | 全量订单 |
| GET | /api/orders/:id | - | Order（含已关联 refunds） |
| POST | /api/refunds | {order_id,amount,reason} | Refund |
| GET | /api/refunds | - | 全量退款 |
| PATCH | /api/refunds/:id/approve | {approved:bool} | 更新后 Refund |

对应完整请求/响应字段见 `api-reference.md`。

### 6. 横评表（比较几种方案选一个）—— 已填样本
| 方案 | 优点 | 缺点 | 结论 |
|---|---|---|---|
| Node+Express+better-sqlite3 | 生态熟 | Win/Node24 编译冲突 | ✗ |
| **Go+Gin+modernc SQLite** | 零 CGO、一条命令、测试友好 | 无 hot reload | ✓ 采用 |
| Go+标准库 net/http | 更简 | 路由样板手写多 | ✗ |

### 7. 风险
Windows 下 CGO 编译（已用纯 Go 驱动规避）；契约变更影响前端（先看 api-reference）。

---

## 空模板：给你自己项目用（复制下面内容开写）

```
### 1. 问题 / 目标
___
### 2. 范围（做 / 不做）
___
### 3. 约束
___
### 4. 关键数据模型
___
### 5. 接口定义
___
### 6. 横评表
___
### 7. 风险
___
```