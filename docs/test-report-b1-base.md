# test-report-b1-base · S-B 测试与回归（四段空模板 + 预填范围）

> 对齐 vibeCoding 共享单元 S-B「测试与回归」。
> 这是一份「测试报告」模板。**范围基线段已预填**（refund-shop B1 基线要覆盖哪些），其余三段待你跑完测试后填写。

---

## 1. 范围基线（本次改动的范围）
- 下单：POST `/api/orders`（含 0/负数/超大金额边界）。
- 退款申请：POST `/api/refunds`（amount ≤ 订单可退额）。
- 退款审批：PATCH `/api/refunds/:id/approve`（通过→订单状态联动）。
- 纯函数 `CalcRemaining`：正常 / 已 part 部分退 / 超可退额 / 一次退满。

## 2. 用例代码
（贴你的测试用例代码 / 或写明引用了哪个 `_test.go`）
```
___
```

## 3. 运行结果
```powershell
go test -count=1 -cover ./...
```
```
___
```
（期望：`internal/domain` + `internal/routes` 两行 ok；routes 覆盖率落在 40–60%（当前 51.2%），domain 覆盖率允许更高（当前 85.7%）。）

## 4. 回归结论
- [ ] 既有功能无回归（`go test ./...` 全绿）
- [ ] routes 覆盖率教学区间达成（domain 允许更高）
- [ ] 结论：___

---

## 模板备注
- 四段固定：范围基线 / 用例代码 / 运行结果 / 回归结论。
- "范围基线"一改代码前就写死，避免"只测改的那一行"的偷懒回归。