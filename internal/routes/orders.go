// Package routes 实现 HTTP 接入层：订单 + 退款 6 个端点。
//
// B3 坏味道 1「Handler 三段式重复」锚点：orders/refunds 多处出现
// "ShouldBindJSON -> err!=nil -> 400 JSON -> return" 的重复样板。
package routes

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// OrderMinAmount 订单金额合法下限（分），默认 1（即 amount>=1 才合法，等价于 amount>0）。
// B2 教学注入点：b2-bug 分支的 internal/b2bug 包会在 init() 里改它，从而可复现
// "0 元订单能创建"的 Bug；main 分支不 import 该包，此值保持默认，行为始终正确。
var OrderMinAmount = 1

// OrderReq 下单请求体。字段名与前端 JSON 严格一致（NFR：金额为「分」）。
type OrderReq struct {
	ProductName string `json:"product_name"`
	Amount      int    `json:"amount"`
	Shipping    int    `json:"shipping"`
	CouponUsed  int    `json:"coupon_used"`
}

// Order 订单返回对象。
type Order struct {
	ID          int64  `json:"id"`
	ProductName string `json:"product_name"`
	Amount      int    `json:"amount"`
	Shipping    int    `json:"shipping"`
	CouponUsed  int    `json:"coupon_used"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

// RegisterOrders 挂载订单 3 端点。
func RegisterOrders(r *gin.RouterGroup, database *sql.DB) {
	r.POST("/orders", createOrder(database))
	r.GET("/orders", listOrders(database))
	r.GET("/orders/:id", getOrderDetail(database))
}

func createOrder(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req OrderReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
			return
		}
		// B2 Bug1 预埋点（orders.go 内）：b2-bug 分支会通过注入包把 OrderMinAmount
		// 从默认 1 改成 0，导致 amount=0 也能创建订单。main 分支保持默认(>=1)，行为正确。
		if req.Amount < OrderMinAmount {
			c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be > 0 (cents)"})
			return
		}
		if req.ProductName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "product_name must not be empty"})
			return
		}

		res, err := database.Exec(
			"INSERT INTO orders (product_name, amount, shipping, coupon_used) VALUES (?, ?, ?, ?)",
			req.ProductName, req.Amount, req.Shipping, req.CouponUsed,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		id, _ := res.LastInsertId()
		o, err := fetchOrder(database, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, o)
	}
}

func listOrders(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := database.Query(
			"SELECT id, product_name, amount, shipping, coupon_used, status, created_at" +
				" FROM orders ORDER BY created_at DESC, id DESC")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		orders := []Order{}
		for rows.Next() {
			var o Order
			if err := rows.Scan(&o.ID, &o.ProductName, &o.Amount, &o.Shipping, &o.CouponUsed, &o.Status, &o.CreatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			orders = append(orders, o)
		}
		c.JSON(http.StatusOK, orders)
	}
}

// RefundBrief 订单详情内嵌的退款摘要。
type RefundBrief struct {
	ID     int64  `json:"id"`
	Amount int    `json:"amount"`
	Status string `json:"status"`
	Reason string `json:"reason"`
}

func getOrderDetail(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
			return
		}
		o, err := fetchOrder(database, id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		refunds, err := fetchRefundsByOrder(database, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"order": o, "refunds": refunds})
	}
}

func fetchOrder(database *sql.DB, id int64) (*Order, error) {
	var o Order
	err := database.QueryRow(
		"SELECT id, product_name, amount, shipping, coupon_used, status, created_at"+
			" FROM orders WHERE id = ?", id).
		Scan(&o.ID, &o.ProductName, &o.Amount, &o.Shipping, &o.CouponUsed, &o.Status, &o.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func fetchRefundsByOrder(database *sql.DB, orderID int64) ([]RefundBrief, error) {
	rows, err := database.Query(
		"SELECT id, amount, status, reason FROM refunds WHERE order_id = ? ORDER BY created_at ASC", orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	refunds := []RefundBrief{}
	for rows.Next() {
		var rf RefundBrief
		if err := rows.Scan(&rf.ID, &rf.Amount, &rf.Status, &rf.Reason); err != nil {
			return nil, err
		}
		refunds = append(refunds, rf)
	}
	return refunds, nil
}