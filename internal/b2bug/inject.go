// Package b2bug 为 B2 教学轨道注入两个可复现 Bug。
//
// 仅存在于 b2-bug 分支；main 分支不 import 本包。init() 会修改共享全局变量，
// 从而让 lines 行为偏离正确实现，形成「0 元订单能创建」与「累计退款计算写反」两个 Bug。
//
// 学习使用：
//   1. 在 main 分支用 `git checkout b2-bug` 切到本分支。
//   2. 运行 `go run .` 起服务，用 curl POST amount=0 试图 0 元下单 → 会成功(200)，
//      而 main 分支同样请求应返回 400。这就是 Bug1。
//   3. 修复目标见 PITFALLS-B2.md；修复时【不要改本文件】，应修正 orders.go 的
//      OrderMinAmount 或 domain 逻辑，让 Bug 从根上消失。
package b2bug

import (
	"refund-shop/internal/domain"
	"refund-shop/internal/routes"
)

func init() {
	// Bug1：把订单金额下限从默认 1 改成 0 → amount=0 也能创建订单。
	routes.OrderMinAmount = 0

	// Bug2：把「可退余额」用错误减法 → 累计退款当时会导致 remaining 异常。
	domain.ReverseRefundRemaining = true
}