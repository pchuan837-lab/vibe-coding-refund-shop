package domain

import "testing"

// ---- 骨架单元用例（S-B 基线，对应 spec 附录-A 7 条）----

// L12 全额退款无历史
func TestCalcRemaining_FullRefund_NoHistory(t *testing.T) {
	got, err := CalcRemaining(1, 10000, 0, 10000)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != 10000 {
		t.Fatalf("want 10000, got %d", got)
	}
}

// L21 部分退款 30% 合法
func TestCalcRemaining_Partial_WithinRange(t *testing.T) {
	got, err := CalcRemaining(1, 10000, 0, 3000)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != 3000 {
		t.Fatalf("want 3000, got %d", got)
	}
}

// L29 二次部分退款（累计 80% 仍合法）
func TestCalcRemaining_SecondPartial_CumulativeUnderTotal(t *testing.T) {
	got, err := CalcRemaining(1, 10000, 5000, 3000)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != 3000 {
		t.Fatalf("want 3000, got %d", got)
	}
}

// L37 超额申请 → 截断退剩余（策略 A）
func TestCalcRemaining_OverApply_TruncatedToRemaining(t *testing.T) {
	got, err := CalcRemaining(1, 10000, 8000, 5000)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != 2000 { // 剩余 2000 截断
		t.Fatalf("want 2000 (truncated), got %d", got)
	}
}

// L46 申请 0 元 → 报错
func TestCalcRemaining_ApplyZero_ReturnsError(t *testing.T) {
	_, err := CalcRemaining(1, 10000, 0, 0)
	if err == nil {
		t.Fatal("want err for apply 0, got nil")
	}
}

// L53 订单金额无效 → 报错
func TestCalcRemaining_InvalidOrderAmount_ReturnsError(t *testing.T) {
	_, err := CalcRemaining(1, 0, 0, 100)
	if err == nil {
		t.Fatal("want err for orderAmount<=0, got nil")
	}
}

// L60 已全退完再申请 → 报错
func TestCalcRemaining_AfterFullRefunded_ReturnsError(t *testing.T) {
	_, err := CalcRemaining(1, 10000, 10000, 100)
	if err == nil {
		t.Fatal("want err for fully refunded, got nil")
	}
}

// L74 以下为 S-B③「失败分析」练习用的注释示范（默认保留为注释，不参与编译）
/*
func TestCalc_OverApply_Legacy(t *testing.T) {
	// 原实现曾选策略 B（报错拒整笔），后按澄清改为策略 A（截断）
	// 这段用于读者练习「失败用例 -> 原因分析 -> 改断言」的 S-B③ 流程
}
*/