package routes_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	domain "refund-shop/internal/domain"
)

// 同包 orders_route_test.go 已提供：resetB2Globals / newTestRouter / doJSON（同名符号共享）。
// 仅在此占位避免 import 未使用告警：
var (
	_ = httptest.NewRecorder
	_ = gin.New
)

// ---------- helpers ----------

func createOrderInDB(t *testing.T, eng *gin.Engine, amount int) int64 {
	t.Helper()
	w := doJSON(t, eng, http.MethodPost, "/api/orders", map[string]any{
		"product_name": "TestOrder", "amount": amount, "shipping": 0, "coupon_used": 0,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("create order: want 200 got %d body=%s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil { t.Fatalf("unmarshal order: %v", err) }
	return int64(resp["id"].(float64))
}

func createRefundInDB(t *testing.T, eng *gin.Engine, orderID int64, amount int) int64 {
	t.Helper()
	w := doJSON(t, eng, http.MethodPost, "/api/refunds", map[string]any{
		"order_id": orderID, "amount": amount, "reason": "test refund",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("create refund: want 200 got %d body=%s (order=%d amount=%d)",
			w.Code, w.Body.String(), orderID, amount)
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil { t.Fatalf("unmarshal refund: %v", err) }
	return int64(resp["id"].(float64))
}

func approveRefundInDB(t *testing.T, eng *gin.Engine, refundID int64, approved bool) {
	t.Helper()
	path := "/api/refunds/" + strconv.FormatInt(refundID, 10) + "/approve"
	w := doJSON(t, eng, http.MethodPatch, path, map[string]any{
		"approved": approved, "comment": "ok",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("approve refund %d: want 200 got %d body=%s", refundID, w.Code, w.Body.String())
	}
}

// 详情接口返回 {"order": {...,"status": ...}, "refunds": [...]}
func assertOrderStatus(t *testing.T, eng *gin.Engine, orderID int64, want string) {
	t.Helper()
	path := "/api/orders/" + strconv.FormatInt(orderID, 10)
	w := doJSON(t, eng, http.MethodGet, path, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("get order %d: want 200 got %d body=%s", orderID, w.Code, w.Body.String())
	}
	var wrap struct {
		Order struct {
			Status string `json:"status"`
		} `json:"order"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &wrap); err != nil {
		t.Fatalf("unmarshal order detail: %v body=%s", err, w.Body.String())
	}
	if wrap.Order.Status != want {
		t.Fatalf("order id=%d status: want %q got %q body=%s",
			orderID, want, wrap.Order.Status, w.Body.String())
	}
}

// ========== S-B 基线（reset 后必须全通过）==========

// 用例1：正常退款申请 → 200 + status=pending
func TestCreateRefund_Normal_Returns200AndPending(t *testing.T) {
	resetB2Globals(t)
	saved := domain.ReverseRefundRemaining
	domain.ReverseRefundRemaining = false
	t.Cleanup(func() { domain.ReverseRefundRemaining = saved })
	eng, _ := newTestRouter(t)
	orderID := createOrderInDB(t, eng, 10000)

	w := doJSON(t, eng, http.MethodPost, "/api/refunds", map[string]any{
		"order_id": orderID, "amount": 3000, "reason": "damaged",
	})
	if w.Code != http.StatusOK { t.Fatalf("want 200 got %d body=%s", w.Code, w.Body.String()) }
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil { t.Fatalf("unmarshal: %v", err) }
	if resp["status"] != "pending" { t.Fatalf("want pending got %v", resp["status"]) }
	if int(resp["amount"].(float64)) != 3000 { t.Fatalf("want amount=3000 got %v", resp["amount"]) }
}

// 用例2：已全退再申请 → 400 + 额度冻结文案
func TestCreateRefund_AfterFullyRefunded_Returns400Frozen(t *testing.T) {
	resetB2Globals(t)
	saved := domain.ReverseRefundRemaining
	domain.ReverseRefundRemaining = false
	t.Cleanup(func() { domain.ReverseRefundRemaining = saved })
	eng, _ := newTestRouter(t)
	orderID := createOrderInDB(t, eng, 10000)
	rid := createRefundInDB(t, eng, orderID, 10000)
	approveRefundInDB(t, eng, rid, true)

	w := doJSON(t, eng, http.MethodPost, "/api/refunds", map[string]any{
		"order_id": orderID, "amount": 100, "reason": "extra",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400 got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "refund amount exceeds refundable quota") {
		t.Fatalf("want frozen quota msg, got %s", w.Body.String())
	}
}

// 用例3：审批通过 → 200 + status=approved
func TestApproveRefund_Normal_Returns200AndApproved(t *testing.T) {
	resetB2Globals(t)
	saved := domain.ReverseRefundRemaining
	domain.ReverseRefundRemaining = false
	t.Cleanup(func() { domain.ReverseRefundRemaining = saved })
	eng, _ := newTestRouter(t)
	orderID := createOrderInDB(t, eng, 10000)
	rid := createRefundInDB(t, eng, orderID, 3000)

	path := "/api/refunds/" + strconv.FormatInt(rid, 10) + "/approve"
	w := doJSON(t, eng, http.MethodPatch, path, map[string]any{
		"approved": true, "comment": "ok",
	})
	if w.Code != http.StatusOK { t.Fatalf("want 200 got %d body=%s", w.Code, w.Body.String()) }
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil { t.Fatalf("unmarshal: %v", err) }
	if resp["status"] != "approved" { t.Fatalf("want approved got %v", resp["status"]) }
}

// 用例4：审批超限被拒 → 400 冻结文案（防 TOCTOU）
func TestApproveRefund_OverQuota_Returns400Frozen(t *testing.T) {
	resetB2Globals(t)
	saved := domain.ReverseRefundRemaining
	domain.ReverseRefundRemaining = false
	t.Cleanup(func() { domain.ReverseRefundRemaining = saved })
	eng, _ := newTestRouter(t)
	orderID := createOrderInDB(t, eng, 10000)
	rid1 := createRefundInDB(t, eng, orderID, 8000)
	rid2 := createRefundInDB(t, eng, orderID, 5000)
	approveRefundInDB(t, eng, rid1, true)

	path := "/api/refunds/" + strconv.FormatInt(rid2, 10) + "/approve"
	w := doJSON(t, eng, http.MethodPatch, path, map[string]any{
		"approved": true, "comment": "ok",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400 got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "refund amount exceeds refundable quota") {
		t.Fatalf("want frozen quota msg got %s", w.Body.String())
	}
}

// 用例5：订单状态联动 → paid → partial → fully
func TestOrderStatus_Linkage_PaidPartialFully(t *testing.T) {
	resetB2Globals(t)
	saved := domain.ReverseRefundRemaining
	domain.ReverseRefundRemaining = false
	t.Cleanup(func() { domain.ReverseRefundRemaining = saved })
	eng, _ := newTestRouter(t)
	orderID := createOrderInDB(t, eng, 10000)
	assertOrderStatus(t, eng, orderID, "paid")
	rid1 := createRefundInDB(t, eng, orderID, 3000)
	approveRefundInDB(t, eng, rid1, true)
	assertOrderStatus(t, eng, orderID, "partial_refunded")
	rid2 := createRefundInDB(t, eng, orderID, 7000)
	approveRefundInDB(t, eng, rid2, true)
	assertOrderStatus(t, eng, orderID, "fully_refunded")
}

// ========== Bug2 HTTP 红线（R4-① 修复关键：NOT reset，NOT switch；直接断言正确行为，b2-bug 注入真实 FAIL）==========

// Bug2 HTTP 红线-1：先申请 10000 退款再审批 → 应 200 approved + 订单 fully_refunded。
// Bug2 注入 ReverseRefundRemaining=true 时：CalcRemaining 写反 remaining 导致审批 10000 时
// 走到「remaining=-10000」路径被误认为 fully refunded，返回 400。
func TestCreateRefund_Bug2ReverseRemaining_ApproveFull_LeadsFully(t *testing.T) {
	// 不 reset：保留 Bug2 ReverseRefundRemaining=true 的注入态
	eng, _ := newTestRouter(t)
	orderID := createOrderInDB(t, eng, 10000)
	rid := createRefundInDB(t, eng, orderID, 10000)
	path := "/api/refunds/" + strconv.FormatInt(rid, 10) + "/approve"
	w := doJSON(t, eng, http.MethodPatch, path, map[string]any{
		"approved": true, "comment": "ok",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("Bug2 真实 FAIL：批 10000 应 200 approved，实际 HTTP %d body=%s（b2-bug 写反 remaining 触发）",
			w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil { t.Fatalf("unmarshal: %v", err) }
	if resp["status"] != "approved" {
		t.Fatalf("Bug2 修复后绿态 want approved, got %v", resp["status"])
	}
	assertOrderStatus(t, eng, orderID, "fully_refunded")
}

// Bug2 HTTP 红线-2：先申请 8000 再审批 → 应 200 approved + 订单 partial_refunded。
// Bug2 注入时：remaining 被写反为负值 → 触发"已退满"逻辑，返回 400 或状态错位
func TestOrderStatus_Bug2ReverseRemaining_Approve8k_LeadsPartial(t *testing.T) {
	eng, _ := newTestRouter(t)
	orderID := createOrderInDB(t, eng, 10000)
	rid := createRefundInDB(t, eng, orderID, 8000)
	path := "/api/refunds/" + strconv.FormatInt(rid, 10) + "/approve"
	w := doJSON(t, eng, http.MethodPatch, path, map[string]any{
		"approved": true, "comment": "ok",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("Bug2 真实 FAIL：批 8000 应 200 approved，实际 HTTP %d body=%s（b2-bug 写反 remaining 触发）",
			w.Code, w.Body.String())
	}
	assertOrderStatus(t, eng, orderID, "partial_refunded")
}