package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"refund-shop/internal/db"
)

// newTestRouter 用内存库建路由，供集成测隔离。
func newTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	database, err := db.NewDB(":memory:")
	if err != nil {
		t.Fatalf("new memory db: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api")
	RegisterOrders(api, database)
	RegisterRefunds(api, database)
	return r
}

func doJSON(t *testing.T, r *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// L43 正常下单返回 200 + id + status=paid
func TestCreateOrder_Success_Returns200AndID(t *testing.T) {
	r := newTestRouter(t)
	w := doJSON(t, r, http.MethodPost, "/api/orders", map[string]any{
		"product_name": "Cup", "amount": 9900, "shipping": 0, "coupon_used": 0,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp["id"] == nil {
		t.Fatal("resp.id missing")
	}
	if resp["status"] != "paid" {
		t.Fatalf("want status=paid, got %v", resp["status"])
	}
}

// L73 负数金额返回 400
func TestCreateOrder_NegativeAmount_Returns400(t *testing.T) {
	r := newTestRouter(t)
	w := doJSON(t, r, http.MethodPost, "/api/orders", map[string]any{
		"product_name": "Bad", "amount": -1, "shipping": 0, "coupon_used": 0,
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "amount must be") {
		t.Fatalf("want amount error msg, got %s", w.Body.String())
	}
}

// 0 元订单应被拒绝（main 分支正确行为；此即 B2 Bug1 的防再犯红线，见 PITFALLS-B2 B2-P09）
func TestCreateOrder_ZeroAmount_Returns400(t *testing.T) {
	r := newTestRouter(t)
	w := doJSON(t, r, http.MethodPost, "/api/orders", map[string]any{
		"product_name": "Free", "amount": 0, "shipping": 0, "coupon_used": 0,
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "amount must be") {
		t.Fatalf("want amount error msg, got %s", w.Body.String())
	}
}