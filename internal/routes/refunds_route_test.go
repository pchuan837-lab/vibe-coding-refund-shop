package routes

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// 本文件为 C-08 HTTP 层测试：5 用例覆盖资金关键链路
// （createRefund / approveRefund / 订单状态联动），断言 C-05 冻结 400 文案。
// 不铺并发模拟（:memory: 并发脆弱），聚焦关键业务路径。

// createOrderInDB 建订单（退款测试预处理），返回订单 id。
func createOrderInDB(t *testing.T, r *gin.Engine, amount int) int64 {
	t.Helper()
	w := doJSON(t, r, http.MethodPost, "/api/orders", map[string]any{
		"product_name": "TestOrder", "amount": amount, "shipping": 0, "coupon_used": 0,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("create order: want 200, got %d body=%s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal order: %v", err)
	}
	id, _ := resp["id"].(float64)
	return int64(id)
}

// createRefundInDB 申请退款（预处理），返回 refund id。
func createRefundInDB(t *testing.T, r *gin.Engine, orderID int64, amount int) int64 {
	t.Helper()
	w := doJSON(t, r, http.MethodPost, "/api/refunds", map[string]any{
		"order_id": orderID, "amount": amount, "reason": "test refund",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("create refund: want 200, got %d body=%s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal refund: %v", err)
	}
	id, _ := resp["id"].(float64)
	return int64(id)
}

// approveRefundInDB 审批退款（预处理）。
func approveRefundInDB(t *testing.T, r *gin.Engine, refundID int64, approved bool) {
	t.Helper()
	path := "/api/refunds/" + strconv.FormatInt(refundID, 10) + "/approve"
	w := doJSON(t, r, http.MethodPatch, path, map[string]any{
		"approved": approved, "comment": "ok",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("approve refund %d: want 200, got %d body=%s", refundID, w.Code, w.Body.String())
	}
}

// assertOrderStatus 通过详情接口断言订单状态（C-07 聚合推导）。
// 详情接口返回 {"order": {..., "status": ...}, "refunds": [...]}。
func assertOrderStatus(t *testing.T, r *gin.Engine, orderID int64, want string) {
	t.Helper()
	path := "/api/orders/" + strconv.FormatInt(orderID, 10)
	w := doJSON(t, r, http.MethodGet, path, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("get order %d: want 200, got %d body=%s", orderID, w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal order: %v", err)
	}
	order, ok := resp["order"].(map[string]any)
	if !ok {
		t.Fatalf("resp.order missing: %v", resp)
	}
	if order["status"] != want {
		t.Fatalf("order %d status: want %s, got %v", orderID, want, order["status"])
	}
}

// 用例1：正常申请退款 → 200 + status=pending（C-03 事务化创建）
func TestCreateRefund_Normal_Returns200AndPending(t *testing.T) {
	r := newTestRouter(t)
	orderID := createOrderInDB(t, r, 10000) // 100 元

	w := doJSON(t, r, http.MethodPost, "/api/refunds", map[string]any{
		"order_id": orderID, "amount": 3000, "reason": "damaged",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp["status"] != "pending" {
		t.Fatalf("want status=pending, got %v", resp["status"])
	}
	if int(resp["amount"].(float64)) != 3000 {
		t.Fatalf("want amount=3000, got %v", resp["amount"])
	}
}

// 用例2：已全退再申请 → 400 + 冻结文案（C-03 超限拒绝）
// 场景：订单 10000，全额申请并审批通过（totalRefunded=10000），
// 再申请退 → remaining=0 → CalcRemaining 报错 → 400 冻结文案。
func TestCreateRefund_AfterFullyRefunded_Returns400WithFrozenMsg(t *testing.T) {
	r := newTestRouter(t)
	orderID := createOrderInDB(t, r, 10000)
	rid := createRefundInDB(t, r, orderID, 10000) // 截断为 10000，pending
	approveRefundInDB(t, r, rid, true)            // totalRefunded=10000

	w := doJSON(t, r, http.MethodPost, "/api/refunds", map[string]any{
		"order_id": orderID, "amount": 100, "reason": "extra",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "refund amount exceeds refundable quota") {
		t.Fatalf("want frozen quota msg, got %s", w.Body.String())
	}
}

// 用例3：审批通过 → 200 + status=approved（C-04 重验 + C-06 原子更新）
func TestApproveRefund_Normal_Returns200AndApproved(t *testing.T) {
	r := newTestRouter(t)
	orderID := createOrderInDB(t, r, 10000)
	rid := createRefundInDB(t, r, orderID, 3000)

	path := "/api/refunds/" + strconv.FormatInt(rid, 10) + "/approve"
	w := doJSON(t, r, http.MethodPatch, path, map[string]any{
		"approved": true, "comment": "ok",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp["status"] != "approved" {
		t.Fatalf("want status=approved, got %v", resp["status"])
	}
}

// 用例4：审批超限被拒 → 400 + 冻结文案（C-04 审批重验额度，防 TOCTOU）
// 场景：订单 10000，创建时两笔退款（rid1=8000 / rid2=5000）额度都够（remaining=10000），
// 先审批 rid1 → totalRefunded=8000，再审批 rid2 时重算 remaining=2000 < amount=5000
// → 400 冻结文案。这模拟"创建时额度够、审批前被占用"的 TOCTOU 场景。
func TestApproveRefund_OverQuota_Returns400WithFrozenMsg(t *testing.T) {
	r := newTestRouter(t)
	orderID := createOrderInDB(t, r, 10000)
	rid1 := createRefundInDB(t, r, orderID, 8000) // 创建时 remaining=10000，amount=8000
	rid2 := createRefundInDB(t, r, orderID, 5000) // 创建时 remaining=10000，amount=5000
	approveRefundInDB(t, r, rid1, true)            // totalRefunded=8000

	// 审批 rid2：重算 remaining=2000 < amount=5000 → 400 冻结文案
	path := "/api/refunds/" + strconv.FormatInt(rid2, 10) + "/approve"
	w := doJSON(t, r, http.MethodPatch, path, map[string]any{
		"approved": true, "comment": "ok",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "refund amount exceeds refundable quota") {
		t.Fatalf("want frozen quota msg, got %s", w.Body.String())
	}
}

// 用例5：订单状态联动 → paid/partial_refunded/fully_refunded（C-07 聚合推导）
func TestOrderStatus_Linkage_PaidPartialFully(t *testing.T) {
	r := newTestRouter(t)
	orderID := createOrderInDB(t, r, 10000)

	// 初始：paid（无 approved 退款）
	assertOrderStatus(t, r, orderID, "paid")

	// 退 3000 并审批 → partial_refunded
	rid1 := createRefundInDB(t, r, orderID, 3000)
	approveRefundInDB(t, r, rid1, true)
	assertOrderStatus(t, r, orderID, "partial_refunded")

	// 退剩余 7000 并审批 → fully_refunded
	rid2 := createRefundInDB(t, r, orderID, 7000)
	approveRefundInDB(t, r, rid2, true)
	assertOrderStatus(t, r, orderID, "fully_refunded")
}
