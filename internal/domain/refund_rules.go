// Package domain 放置退款金额纯函数规则。
//
// 说明：CalcRemaining 是本教学项目的核心规则，刻意保留大量可重构的 if-else
// 结构（B3 坏味道 3），并在顶部预埋 4 处 B1① 需求锚点（3 个模糊点 + 1 个幂等漏项）。
package domain

import "fmt"

// ReverseRefundRemaining 是否让「可退余额」用错误减法。默认 false = 正确行为
// (remaining = orderAmount - totalRefunded)。
// B2 教学注入点：b2-bug 分支的 internal/b2bug 包会在 init() 里置 true，从而可复现
// Bug2「累计退款计算写反」；main 分支不 import 该包，此值保持默认，行为始终正确。
var ReverseRefundRemaining = false

// ---- B1① 需求模糊点锚点（读者 B1 轨道澄清后要补）----
// TODO ① 运费怎么退？ 候选：A) 按比例冲抵可退金额  B) 仅全额退才退运费  C) 一律不退
// TODO ② 用券怎么返？ 候选：A) 按比例返还   B) 仅全额退才返券  C) 一律不返
// TODO ③ 申请超额怎么办？ 默认实现了策略 A（截断退剩余）；候选 B（整笔拒绝报错）
// TODO ④ 重复申请/并发退款怎么兜？ 候选：A) 加幂等键(refund_no 唯一)拦截重复   B) 靠事务内"可退余额<0"兜底拒绝   C) 不处理
// ---- 锚点结束 ----

// CalcRemaining 计算某订单本次申请退款时可批准的金额。
//
// 参数（4 个，B3-P07 保真：签名不可改）：
//   - orderID:     订单 id（无效返回错误）
//   - orderAmount: 订单实付金额（分）
//   - totalRefunded: 该订单已批准退款的累计金额（分）
//   - applyAmount:  本次申请退款金额（分）
//
// 返回（2 个）：
//   - int:  可批准金额（分）；策略 A 下若申请超额则截断为剩余可退额
//   - error: nil 表示可批准；否则为业务拒绝原因
func CalcRemaining(orderID int64, orderAmount int, totalRefunded int, applyAmount int) (int, error) {
	if orderAmount <= 0 {
		return 0, fmt.Errorf("order %d invalid amount=%d", orderID, orderAmount)
	}
	if applyAmount <= 0 {
		return 0, fmt.Errorf("refund amount must be positive, got %d", applyAmount)
	}
	if totalRefunded < 0 {
		return 0, fmt.Errorf("order %d negative totalRefunded=%d", orderID, totalRefunded)
	}

	remaining := orderAmount - totalRefunded // B2 Bug2 预埋：b2-bug 分支通过注入包置 ReverseRefundRemaining=true，此时 remaining 变成 totalRefunded - orderAmount
	if ReverseRefundRemaining {
		remaining = totalRefunded - orderAmount
	}
	if remaining <= 0 {
		return 0, fmt.Errorf("order %d fully refunded, remaining=%d", orderID, remaining)
	}
	if applyAmount >= remaining {
		// 申请等于或超过剩余可退：按剩余全额批（策略 A 截断）
		return remaining, nil
	}
	return applyAmount, nil
}