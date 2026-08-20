# PITFALLS-B2 · 阅卷答案本（严禁练习前打开）

> ## ⚠️⚠️⚠️ 三重泄题警告 ⚠️⚠️⚠️
> **1.** 本文件是 B2「Bug 修复」赛道的**标准答案**（含 Bug 根因），练习前打开 = 白练。
> **2.** 请先闭卷上 `b2-bug` 分支修复，再用 `PROMPT-AI-REVIEWER.md` 让 AI 按此判定。
> **3.** 公开教学仓库前**必须移除本文件**。
> 使用纪律：仅阅卷时供 AI/讲师使用。

---

**判定符号**：✅ 达标 / ⚠ 部分达成 / ❌ 未达成
B2 赛道 = 可复现 Bug 修复。共 **10 条**：Bug1 修复 1 / Bug2 修复 1 / 过程纪律 7。源：`docs/~/{bug-prompt-b2}.md`（老王凌晨 2 点拉群话术）。

## 一、Bug1：下单 amount 边界（精确答案）

| 条目 | 判定标准（✅ 需达成） | 修复标准 |
|---|---|---|
| B2-P01 | 定位到 `internal/routes/orders.go` 下单处理的下单金额边界条件 | `b2-bug` 分支故意把 `<= 0` 写成 `< 0`；修复 = 恢复 `<= 0`，使 `amount=0` 拒单 |
| B2-P02 | 用复现步骤证明：发 `amount:0` 的下单请求能被拒 || POST /api/orders {amount:0} → 期望 400 |

## 二、Bug2：累计退款减数（精确答案）

| 条目 | 判定标准（✅ 需达成） | 修复标准 |
|---|---|---|
| B2-P03 | 定位到 `internal/domain/refund_rules.go` 计算可退余额处的减数写法 | `b2-bug` 分支把 `orderAmount - totalRefunded` 写成 `totalRefunded - ...` 之类；修复 = 恢复「可退余额 = orderAmount - totalRefunded ≥ 0」 |

## 三、过程纪律（How 修对的姿势）共 7 条

| 条目 | 判定标准（✅ 需达成） | 漏了去哪补 |
|---|---|---|
| B2-P04 | 先写/运行可复现失败的测试，再改代码（红→绿） | `test-report-b1-base.md` §2；`*_test.go` |
| B2-P05 | 修复后跑回归 `go test ./...` 全绿，旧用例不破坏 | 同上 |
| B2-P06 | 写下「Bug 根因一句话 + 触发条件 + 修复后行为」 | `implementation-sample-record.md` |
| B2-P07 | 注明 Bug 是仅演示用 pre-intended，非生产混淆 | `CHANGELOG.md` 备注 |
| B2-P08 | 修复局限在最小 diff，不顺手重构无关代码 | diff 审查 |
| B2-P09 | 对「0 元能下单」这个行为建了防再犯测试 | `orders_route_test.go` |
| B2-P10 | 修复方案让 AI 复核是否真的覆盖了根因（非只盖症状） | AI-REVIEWER 留存 |

---

### 阅卷口径（AI 用，读者勿读）
- Bug 是否"真的修了根因"，看 P01/P03 的**代码恢复点**是否命中。
- 过程纪律看测试红/绿顺序与 diff 最小性。
- 防再犯测试（P09）是"是否会复发"的关键证据。