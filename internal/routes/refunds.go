package routes

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"refund-shop/internal/domain"
)

// RefundReq 申请退款请求体。
type RefundReq struct {
	OrderID int64  `json:"order_id"`
	Amount  int    `json:"amount"`
	Reason  string `json:"reason"`
}

// Refund 退款返回对象。
type Refund struct {
	ID        int64  `json:"id"`
	OrderID   int64  `json:"order_id"`
	Amount    int    `json:"amount"`
	Status    string `json:"status"`
	Reason    string `json:"reason"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// RegisterRefunds 挂载退款 3 端点。
func RegisterRefunds(r *gin.RouterGroup, database *sql.DB) {
	r.POST("/refunds", createRefund(database))
	r.GET("/refunds", listRefunds(database))
	r.PATCH("/refunds/:id/approve", approveRefund(database))
}

func createRefund(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// ③ 三段式重复样板（B3 坏味道 1 锚点）
		var req RefundReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
			return
		}
		if req.Amount <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "refund amount must be > 0 (cents)"})
			return
		}

		// 读取订单实付金额 + 累计已批金额
		var orderAmount, totalRefunded int
		err := database.QueryRow("SELECT amount FROM orders WHERE id = ?", req.OrderID).Scan(&orderAmount)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		_ = database.QueryRow(
			"SELECT COALESCE(SUM(amount),0) FROM refunds WHERE order_id = ? AND status='approved'", req.OrderID).
			Scan(&totalRefunded)

		approved, err := domain.CalcRefundable(req.OrderID, orderAmount, totalRefunded, req.Amount)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "refund rule: " + err.Error()})
			return
		}

		res, err := database.Exec(
			"INSERT INTO refunds (order_id, amount, status, reason) VALUES (?, ?, 'pending', ?)",
			req.OrderID, approved, req.Reason,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		id, _ := res.LastInsertId()
		rf, err := fetchRefund(database, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, rf)
	}
}

func listRefunds(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := c.Query("status")
		orderID := c.Query("order_id")

		q := "SELECT id, order_id, amount, status, reason, created_at, updated_at FROM refunds"
		cond := ""
		args := []any{}
		if status != "" {
			cond = " WHERE status = ?"
			args = append(args, status)
		}
		if orderID != "" {
			if cond == "" {
				cond = " WHERE "
			} else {
				cond += " AND "
			}
			cond += "order_id = ?"
			args = append(args, orderID)
		}
		q += cond + " ORDER BY created_at DESC, id DESC"

		rows, err := database.Query(q, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		refunds := []Refund{}
		for rows.Next() {
			var rf Refund
			if err := rows.Scan(&rf.ID, &rf.OrderID, &rf.Amount, &rf.Status, &rf.Reason, &rf.CreatedAt, &rf.UpdatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			refunds = append(refunds, rf)
		}
		c.JSON(http.StatusOK, refunds)
	}
}

// ApproveReq 审批请求体。
type ApproveReq struct {
	Approved bool   `json:"approved"`
	Comment  string `json:"comment"`
}

func approveRefund(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid refund id"})
			return
		}
		var req ApproveReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
			return
		}

		// 校验当前退款存在且是 pending
		var orderID int64
		var status, updatedAt string
		err = database.QueryRow("SELECT order_id, status, updated_at FROM refunds WHERE id = ?", id).
			Scan(&orderID, &status, &updatedAt)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "refund not found"})
			return
		}
		if status != "pending" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "refund status is " + status + ", not pending"})
			return
		}

		var newStatus string
		if req.Approved {
			newStatus = "approved"
		} else {
			newStatus = "rejected"
		}
		if _, err := database.Exec(
			"UPDATE refunds SET status = ?, updated_at = datetime('now') WHERE id = ?", newStatus, id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 审批通过时联动订单状态
		if req.Approved {
			if err := syncOrderStatus(database, orderID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		// 重新读取返回
		rf, err := fetchRefund(database, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"id":         rf.ID,
			"order_id":   rf.OrderID,
			"status":     rf.Status,
			"updated_at": rf.UpdatedAt,
		})
	}
}

// syncOrderStatus 按累计已批金额联动订单状态：==amount→fully; <→partial; 0→paid。
func syncOrderStatus(database *sql.DB, orderID int64) error {
	var orderAmount, totalRefunded int
	if err := database.QueryRow("SELECT amount FROM orders WHERE id = ?", orderID).Scan(&orderAmount); err != nil {
		return err
	}
	_ = database.QueryRow(
		"SELECT COALESCE(SUM(amount),0) FROM refunds WHERE order_id = ? AND status='approved'", orderID).
		Scan(&totalRefunded)

	var newStatus string
	switch {
	case totalRefunded >= orderAmount:
		newStatus = "fully_refunded"
	case totalRefunded > 0:
		newStatus = "partial_refunded"
	default:
		newStatus = "paid"
	}
	_, err := database.Exec("UPDATE orders SET status = ? WHERE id = ?", newStatus, orderID)
	return err
}

func fetchRefund(database *sql.DB, id int64) (*Refund, error) {
	var rf Refund
	err := database.QueryRow(
		"SELECT id, order_id, amount, status, reason, created_at, updated_at FROM refunds WHERE id = ?", id).
		Scan(&rf.ID, &rf.OrderID, &rf.Amount, &rf.Status, &rf.Reason, &rf.CreatedAt, &rf.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &rf, nil
}