package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/services"
	"github.com/shaonianqiutan/backend/utils"
)

// NotificationController 通知控制器
type NotificationController struct {
	service *services.NotificationService
}

// NewNotificationController 创建通知控制器
func NewNotificationController(service *services.NotificationService) *NotificationController {
	return &NotificationController{service: service}
}

// NotificationResponse 通知响应结构
type NotificationResponse struct {
	ID        uint        `json:"id"`
	Type      string      `json:"type"`
	Title     string      `json:"title"`
	Content   string      `json:"content"`
	Data      interface{} `json:"data,omitempty"`
	IsRead    bool        `json:"isRead"`
	Priority  int         `json:"priority"`
	CreatedAt string      `json:"createdAt"`
}

// toResponse 转换为响应结构
func toResponse(n *models.Notification) NotificationResponse {
	resp := NotificationResponse{
		ID:        n.ID,
		Type:      string(n.Type),
		Title:     n.Title,
		Content:   n.Content,
		IsRead:    n.IsRead,
		Priority:  n.Priority,
		CreatedAt: n.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	// 将 Data JSON 字符串解析为对象
	if n.Data != "" {
		var dataMap map[string]interface{}
		if err := json.Unmarshal([]byte(n.Data), &dataMap); err == nil {
			resp.Data = dataMap
		}
	}
	return resp
}

// List 获取通知列表
// GET /api/notifications
func (c *NotificationController) List(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	pagination := utils.ParsePagination(ctx)
	page := pagination.Page
	pageSize := pagination.PageSize

	notifications, total, err := c.service.ListByUser(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	list := make([]NotificationResponse, len(notifications))
	for i, n := range notifications {
		list[i] = toResponse(&n)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"list":  list,
			"total": total,
			"page":  page,
		},
	})
}

// ListUnread 获取未读通知列表
// GET /api/notifications/unread
func (c *NotificationController) ListUnread(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	pagination := utils.ParsePagination(ctx)
	page := pagination.Page
	pageSize := pagination.PageSize

	notifications, total, err := c.service.ListUnread(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	list := make([]NotificationResponse, len(notifications))
	for i, n := range notifications {
		list[i] = toResponse(&n)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"list":  list,
			"total": total,
			"page":  page,
		},
	})
}

// GetUnreadCount 获取未读数量
// GET /api/notifications/unread-count
func (c *NotificationController) GetUnreadCount(ctx *gin.Context) {
	userID := ctx.GetUint("userId")

	count, err := c.service.CountUnread(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"count": count,
		},
	})
}

// MarkAsRead 标记通知为已读
// PUT /api/notifications/:id/read
func (c *NotificationController) MarkAsRead(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的通知ID"})
		return
	}

	if err := c.service.MarkAsRead(userID, uint(id)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "标记成功",
	})
}

// MarkAllAsRead 标记全部为已读
// PUT /api/notifications/read-all
func (c *NotificationController) MarkAllAsRead(ctx *gin.Context) {
	userID := ctx.GetUint("userId")

	if err := c.service.MarkAllAsRead(userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "全部已读",
	})
}

// Delete 删除通知
// DELETE /api/notifications/:id
func (c *NotificationController) Delete(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的通知ID"})
		return
	}

	if err := c.service.Delete(userID, uint(id)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "删除成功",
	})
}
