package utils

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ErrorCode 错误码常量
const (
	CodeSuccess         = "SUCCESS"
	CodeValidationError = "VALIDATION_ERROR"
	CodeNotFound       = "NOT_FOUND"
	CodeForbidden      = "FORBIDDEN"
	CodeServerError    = "SERVER_ERROR"
	CodeUnauthorized   = "UNAUTHORIZED"
)

// SuccessResponse 返回成功响应
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// SuccessResponseWithMessage 返回成功响应带消息
func SuccessResponseWithMessage(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
		"message": message,
	})
}

// ErrorResponse 返回错误响应
func ErrorResponse(c *gin.Context, httpStatus int, code, message string) {
	c.JSON(httpStatus, gin.H{
		"success": false,
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	})
}

// ValidationError 返回参数错误
func ValidationError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusBadRequest, CodeValidationError, message)
}

// NotFoundError 返回未找到错误
func NotFoundError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusNotFound, CodeNotFound, message)
}

// ForbiddenError 返回无权限错误
func ForbiddenError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusForbidden, CodeForbidden, message)
}

// ServerError 返回服务器错误
func ServerError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusInternalServerError, CodeServerError, message)
}

// PaginatedResponse 返回分页响应
func PaginatedResponse(c *gin.Context, list interface{}, page, pageSize int, total int64) {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}
	SuccessResponse(c, gin.H{
		"list": list,
		"pagination": gin.H{
			"page":       page,
			"pageSize":   pageSize,
			"total":      total,
			"totalPages": totalPages,
		},
	})
}

// Success 成功响应别名 (支持message参数)
func Success(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "ERROR",
			"message": message,
		},
	})
}

// FormatTime 格式化时间
func FormatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

// FormatDate 格式化日期
func FormatDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

// FormatDateTime 格式化日期时间
func FormatDateTime(t time.Time) string {
	return t.Format("2006-01-02T15:04:05Z")
}
