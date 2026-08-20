# implementation-sample-record · B1⑤ 分步实现记录样本

> 对齐 vibeCoding 共享单元 B「需求澄清」之「⑤ 分步实现」。
> 这是一份「做完一条任务后怎么留痕」的样板。本次以 **建表 + 连接（schema.sql + db.go）** 那条任务为例填写，你照着这个节奏记录自己的每一步。

---

## 任务：初始化数据库层（schema.sql + db.go）

### 改了啥
- 新增 `internal/db/schema.sql`：定义 `orders`、`refunds` 两表 + CHECK 约束。
- 新增 `internal/db/db.go`：`NewDB(path)` 打开 SQLite、执行 schema.sql 建表、强制外键。
- `go.mod` 引入 `modernc.org/sqlite`。

### 为啥这么改
- 两表关系：一条订单可有多条退款记录（`refunds.order_id` 外键）。
- 金额用**分**整数存，避免浮点误差；`CHECK(amount>0)` 拦截非法数据。
- 用纯 Go 驱动规避 Windows CGO 编译问题（见 ADR-001）。

### 验了啥（这条任务的验证）
```powershell
go build ./internal/db/...
```
通过；`data.db` 首次连接自动建表成功。

### 无回归
- `go test ./internal/...` 全绿（本轮暂无外围测试受影响）。

### 状态
✅ 通过 → 数据层可被上层路由调用。

---

## 五段记录模板（复制用）

```
## 任务：___
### 改了啥
___
### 为啥这么改
___
### 验了啥
___
### 无回归
___
### 状态
用 ✅/⚠/❌ 标注
```