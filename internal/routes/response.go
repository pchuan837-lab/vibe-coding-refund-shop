package routes

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrQuotaExceeded 是 sentinel error：退款申请/审批时金额超出可退额度。
// 调用方用 errors.Is 判定后返回 400；其他 error 返回 500。
var ErrQuotaExceeded = errors.New("refund amount exceeds refundable quota")

// respondOK 返回 200 + JSON data。
func respondOK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, data)
}

// respondBadRequest 返回 400 + 业务文案（不泄露内部细节）。
func respondBadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, gin.H{"error": msg})
}

// respondNotFound 返回 404。
func respondNotFound(c *gin.Context, msg string) {
	c.JSON(http.StatusNotFound, gin.H{"error": msg})
}

// respondQuotaExceeded 返回 400 + 冻结超限文案（C-05 衔接锁，教程 D 组引用）。
// 文案 = "refund amount exceeds refundable quota"，一字不改。
func respondQuotaExceeded(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{"error": "refund amount exceeds refundable quota"})
}

// respondInternalError 返回 500 + 固定文案（不泄露 err.Error()）。
func respondInternalError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
}
