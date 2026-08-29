package domain_test

import (
	"strings"
	"testing"

	domain "refund-shop/internal/domain"
)

// resetB2Globals 复位 b2bug.init() 的副作用，保证 S-B 基线测试在 b2-bug 上仍然通过。
// （Bug1 注入 routes.OrderMinAmount=0；Bug2 注入 ReverseRefundRemaining=true）
func resetB2Globals(t *testing.T) {
	t.Helper()
	savedMin := 0 // 占位：OrderMinAmount 在 routes 包，本包看不到
	savedRv := domain.ReverseRefundRemaining
	domain.ReverseRefundRemaining = false // 清除 Bug2 注入
	t.Cleanup(func() { _ = savedMin; domain.ReverseRefundRemaining = savedRv })
}

// -------- S-B 基线：spec 附录-A 7 条关键路径（在注入环境下 reset 后仍要全通过）--------

// L12 全额退款无历史
func TestCalcRemaining_FullRefund_NoHistory(t *testing.T) {
	resetB2Globals(t)
	got, err := domain.CalcRemaining(1, 10000, 0, 10000)
	if err != nil { t.Fatalf("unexpected err: %v", err) }
	if got != 10000 { t.Fatalf("want 10000, got %d", got) }
}

// L21 部分退款 30%
func TestCalcRemaining_Partial_WithinRange(t *testing.T) {
	resetB2Globals(t)
	got, err := domain.CalcRemaining(1, 10000, 0, 3000)
	if err != nil { t.Fatalf("unexpected err: %v", err) }
	if got != 3000 { t.Fatalf("want 3000, got %d", got) }
}

// 二次部分退款（累计 80% 仍合法）
func TestCalcRemaining_SecondPartial_CumulativeUnderTotal(t *testing.T) {
	resetB2Globals(t)
	got, err := domain.CalcRemaining(1, 10000, 3000, 5000)
	if err != nil { t.Fatalf("unexpected err: %v", err) }
	if got != 5000 { t.Fatalf("want 5000, got %d", got) }
}

// 完全退款后再申请 → 拒绝
func TestCalcRemaining_AfterFullyRefunded_Rejects(t *testing.T) {
	resetB2Globals(t)
	_, err := domain.CalcRemaining(1, 10000, 10000, 1)
	if err == nil || !strings.Contains(err.Error(), "fully refunded") {
		t.Fatalf("want fully refunded err, got err=%v", err)
	}
}

// 超额申请 → 截断（策略 A）
func TestCalcRemaining_OverApply_Truncates(t *testing.T) {
	resetB2Globals(t)
	got, err := domain.CalcRemaining(1, 10000, 3000, 99999)
	if err != nil { t.Fatalf("unexpected err: %v", err) }
	if got != 7000 { t.Fatalf("truncate: want 7000, got %d", got) }
}

// 0 金额拒绝（输入校验）
func TestCalcRemaining_ZeroApply_Rejects(t *testing.T) {
	resetB2Globals(t)
	_, err := domain.CalcRemaining(1, 1000, 0, 0)
	if err == nil || !strings.Contains(err.Error(), "positive") {
		t.Fatalf("want positive err, got %v", err)
	}
}

// 负数 totalRefunded 拒绝
func TestCalcRemaining_NegativeTotalRefunded_Rejects(t *testing.T) {
	resetB2Globals(t)
	_, err := domain.CalcRemaining(1, 1000, -1, 100)
	if err == nil || !strings.Contains(err.Error(), "negative totalRefunded") {
		t.Fatalf("want negative totalRefunded err, got %v", err)
	}
}

// ====== Bug2 红线（R4-① 修复关键：不 reset，不 switch；直接断言正确行为。b2-bug 注入环境下真实 FAIL 红，修复后绿。）======

// Bug2 红线-1：订单 10000 已退 3000，再批 5000 → 应该批准 5000，无 err。
// Bug2 注入时（ReverseRefundRemaining=true）写反 remaining = totalRefunded-(order-approve)=3000-5000=-2000→负数→触发 fully refunded 错误。
func TestCalcRemaining_Bug2ReverseRemaining_Apply5000After3000_Approved(t *testing.T) {
	// 不调用 resetB2Globals：保留 Bug2 注入标志 ReverseRefundRemaining=true
	approved, err := domain.CalcRemaining(1, 10000, 3000, 5000)
	if err != nil {
		t.Fatalf("Bug2 真实 FAIL：批 5000 不应报错。b2-bug 注入写反 remaining，err=%v", err)
	}
	if approved != 5000 {
		t.Fatalf("Bug2 真实 FAIL：批 5000 应全额通过，want approved=5000, got %d err=%v", approved, err)
	}
}

// Bug2 红线-2：订单 10000 未退，申请 8000 → 应批准 8000，无 err。
// Bug2 注入时：remaining=0-(10000-8000)= -2000 → 触发 fully refunded 错判。
func TestCalcRemaining_Bug2ReverseRemaining_Apply8000OnFresh_Approved(t *testing.T) {
	approved, err := domain.CalcRemaining(1, 10000, 0, 8000)
	if err != nil {
		t.Fatalf("Bug2 真实 FAIL：申请 8000 不应报错。b2-bug 注入写反 remaining，err=%v", err)
	}
	if approved != 8000 {
		t.Fatalf("Bug2 真实 FAIL：want approved=8000, got %d err=%v", approved, err)
	}
}

// Bug2 红线-3：订单 10000 已退 3000，再申请 1（极小额）→ 应批准 1，无 err。
// Bug2 注入时：remaining=3000-(10000-1)=3000-9999= -6999 → fully refunded。
func TestCalcRemaining_Bug2ReverseRemaining_Apply1After3000_Approved(t *testing.T) {
	approved, err := domain.CalcRemaining(1, 10000, 3000, 1)
	if err != nil {
		t.Fatalf("Bug2 真实 FAIL：申请 1 不应报错。b2-bug 注入写反 remaining，err=%v", err)
	}
	if approved != 1 {
		t.Fatalf("Bug2 真实 FAIL：want approved=1, got %d err=%v", approved, err)
	}
}