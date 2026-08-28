package routes

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"refund-shop/internal/domain"
)

// querier 被 *sql.DB 和 *sql.Tx 同时满足，使 fetchRefund 等查询函数
// 在事务内外均可复用（C-03/C-06 事务化后需要在 tx 上查）。
type querier interface {
	QueryRow(query string, args ...any) *sql.Row
}

// validateAndOccupyQuota 在事务内重验退款额度（C-04 公共函数）。
// 查订单实付 + 累计已批 → CalcRefundable → 返回可批金额 / 错误。
// createRefund（申请=预占）与 approveRefund（审批=重验，最后闸门）均调用，
// 但语义差分：申请时返回值用于写入退款单金额；审批时仅判 err + approved<amount 防并发超退。
func validateAndOccupyQuota(tx *sql.Tx, orderID int64, applyAmount int) (int, error) {
	var orderAmount, totalRefunded int
	err := tx.QueryRow("SELECT amount FROM orders WHERE id = ?", orderID).Scan(&orderAmount)
	if err != nil {
		return 0, err // sql.ErrNoRows 或其他
	}
	if err := tx.QueryRow(
		"SELECT COALESCE(SUM(amount),0) FROM refunds WHERE order_id = ? AND status='approved'", orderID,
	).Scan(&totalRefunded); err != nil {
		return 0, err
	}
	return domain.CalcRefundable(orderID, orderAmount, totalRefunded, applyAmount)
}

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

		// C-03：读-算-写包进事务（_txlock=immediate 在 DSN 层已配，
		// BeginTx 即取写锁，防并发超退/Lost Update）。
		tx, err := database.BeginTx(c.Request.Context(), nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		defer tx.Rollback() // commit 后 noop

		// C-04：公共函数统一查订单+算累计，申请=预占额度。
		approved, err := validateAndOccupyQuota(tx, req.OrderID, req.Amount)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "refund amount exceeds refundable quota"})
			}
			return
		}

		res, err := tx.Exec(
			"INSERT INTO refunds (order_id, amount, status, reason) VALUES (?, ?, 'pending', ?)",
			req.OrderID, approved, req.Reason,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		id, _ := res.LastInsertId()
		rf, err := fetchRefund(tx, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}

		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		defer rows.Close()

		refunds := []Refund{}
		for rows.Next() {
			var rf Refund
			if err := rows.Scan(&rf.ID, &rf.OrderID, &rf.Amount, &rf.Status, &rf.Reason, &rf.CreatedAt, &rf.UpdatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
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

		// C-06：审批包进事务，原子更新防 TOCTOU。
		tx, err := database.BeginTx(c.Request.Context(), nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		defer tx.Rollback()

		// 查退款单的 order_id + amount（需 orderID 才能重验额度）。
		var orderID int64
		var amount int
		err = tx.QueryRow("SELECT order_id, amount FROM refunds WHERE id = ?", id).Scan(&orderID, &amount)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "refund not found"})
			return
		}

		var newStatus string
		if req.Approved {
			// C-04：审批=重验额度（最后闸门），不信任创建时算好的结果。
			approved, err := validateAndOccupyQuota(tx, orderID, amount)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "refund amount exceeds refundable quota"})
				return
			}
			// approved < amount 说明期间有并发批准导致额度减少，无法全额批。
			if approved < amount {
				c.JSON(http.StatusBadRequest, gin.H{"error": "refund amount exceeds refundable quota"})
				return
			}
			newStatus = "approved"
		} else {
			newStatus = "rejected"
		}

		// C-06：原子更新 WHERE id=? AND status='pending' + RowsAffected 判并发。
		res, err := tx.Exec(
			"UPDATE refunds SET status = ?, updated_at = datetime('now') WHERE id = ? AND status = 'pending'",
			newStatus, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		rows, _ := res.RowsAffected()
		if rows == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "refund already processed"})
			return
		}

		// C-07：删 syncOrderStatus——订单状态由查询时聚合推导，不再写入。

		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}

		// commit 后用 database 读返回。
		rf, err := fetchRefund(database, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
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

// C-07：syncOrderStatus 已删除——订单状态不再可变写入，
// 改由 orders.go 的查询接口用 LEFT JOIN refunds + SUM(amount) 实时推导。

func fetchRefund(q querier, id int64) (*Refund, error) {
	var rf Refund
	err := q.QueryRow(
		"SELECT id, order_id, amount, status, reason, created_at, updated_at FROM refunds WHERE id = ?", id).
		Scan(&rf.ID, &rf.OrderID, &rf.Amount, &rf.Status, &rf.Reason, &rf.CreatedAt, &rf.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &rf, nil
}