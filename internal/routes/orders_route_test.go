package routes_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	dbpkg "refund-shop/internal/db"
	routes "refund-shop/internal/routes"
)

func init() { gin.SetMode(gin.TestMode) }

// resetB2Globals 复位 b2bug.init() 注入的 Bug1（OrderMinAmount=0）与 Bug2 标志。
// 调用后：OrderMinAmount=1  ReverseRefundRemaining=false。
func resetB2Globals(t *testing.T) {
	t.Helper()
	saved := routes.OrderMinAmount
	routes.OrderMinAmount = 1        // 清 Bug1
	// domain 包反向标志由本包测试体在必要时单独复位（避免循环依赖此导入）
	t.Cleanup(func() { routes.OrderMinAmount = saved })
}

// newTestRouter 用导出 API 组装；外部包不可见内部 helper。
func newTestRouter(t *testing.T) (*gin.Engine, *sql.DB) {
	t.Helper()
	d, err := dbpkg.NewDB(":memory:")
	if err != nil { t.Fatalf("NewDB: %v", err) }
	t.Cleanup(func() { _ = d.Close() })
	eng := gin.New()
	api := eng.Group("/api")
	routes.RegisterOrders(api, d)
	routes.RegisterRefunds(api, d)
	return eng, d
}

// doJSON 做 JSON 收发；本文件与 refunds_route_test.go 同包共享此符号定义。
func doJSON(t *testing.T, eng *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var raw []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil { t.Fatalf("marshal: %v", err) }
		raw = b
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	eng.ServeHTTP(w, req)
	return w
}

// -------- S-B 基线（reset 后在 b2-bug 仍必须全部通过）--------

// 正常下单 → 200 + id + status=paid
func TestCreateOrder_Success_Returns200AndID(t *testing.T) {
	resetB2Globals(t)
	eng, _ := newTestRouter(t)
	w := doJSON(t, eng, http.MethodPost, "/api/orders", map[string]any{
		"product_name": "Cup", "amount": 9900, "shipping": 0, "coupon_used": 0,
	})
	if w.Code != http.StatusOK { t.Fatalf("want 200, got %d body=%s", w.Code, w.Body.String()) }
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil { t.Fatalf("unmarshal: %v", err) }
	if resp["id"] == nil { t.Fatal("resp.id missing") }
	if resp["status"] != "paid" { t.Fatalf("want paid, got status=%v", resp["status"]) }
}

// 负数金额 → 400
func TestCreateOrder_NegativeAmount_Returns400(t *testing.T) {
	resetB2Globals(t)
	eng, _ := newTestRouter(t)
	w := doJSON(t, eng, http.MethodPost, "/api/orders", map[string]any{
		"product_name": "Bad", "amount": -1, "shipping": 0, "coupon_used": 0,
	})
	if w.Code != http.StatusBadRequest { t.Fatalf("want 400, got %d body=%s", w.Code, w.Body.String()) }
	if !strings.Contains(w.Body.String(), "amount must be") {
		t.Fatalf("want amount msg, got %s", w.Body.String())
	}
}

// Bug1 红线（R4-①）：0 元订单必须拒绝 → 400（直接断言，不 switch 自适应）。
// b2-bug 注入 OrderMinAmount=0 时：本测试会真实 FAIL（HTTP 200 创建成功 → 红）；
// main / 修复后 OrderMinAmount=1：HTTP 400 → PASS（绿）。
// 注：本测试 NOT 调用 resetB2Globals——保留注入的 Bug1 标志让它真实 FAIL。
func TestCreateOrder_ZeroAmount_Returns400(t *testing.T) {
	eng, _ := newTestRouter(t)
	w := doJSON(t, eng, http.MethodPost, "/api/orders", map[string]any{
		"product_name": "Free", "amount": 0, "shipping": 0, "coupon_used": 0,
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("Bug1 真实 FAIL：amount=0 应被拒绝(HTTP 400)，实际 HTTP %d body=%s。b2-bug 注入 OrderMinAmount=0 已生效。",
			w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "amount must be") {
		t.Fatalf("Bug1 修复后绿态仍必须命中 amount must be 文案，got %s", w.Body.String())
	}
}