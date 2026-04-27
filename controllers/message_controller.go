package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/services"
	"github.com/shaonianqiutan/backend/utils"
)

// MessageController 私信控制器
type MessageController struct {
	service *services.MessageService
}

// NewMessageController 创建私信控制器
func NewMessageController(service *services.MessageService) *MessageController {
	return &MessageController{service: service}
}

// SendMessage 发送私信
func (c *MessageController) SendMessage(ctx *gin.Context) {
	userId := ctx.GetUint("userId")

	var req struct {
		ReceiverID uint   `json:"receiver_id" binding:"required"`
		Content    string `json:"content" binding:"required,min=1,max=2000"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Error(ctx, http.StatusBadRequest, "请求参数错误")
		return
	}

	// 禁止给自己发私信
	if userId == req.ReceiverID {
		utils.Error(ctx, http.StatusBadRequest, "不能给自己发送私信")
		return
	}

	msg, err := c.service.SendMessage(userId, req.ReceiverID, req.Content)
	if err != nil {
		// 检查是否是私信限制错误
		if _, ok := err.(*services.MessageLimitError); ok {
			utils.Error(ctx, http.StatusForbidden, err.Error())
			return
		}
		utils.Error(ctx, http.StatusInternalServerError, "发送失败")
		return
	}

	utils.Success(ctx, "发送成功", msg)
}

// GetMessages 获取与某用户的私信列表
func (c *MessageController) GetMessages(ctx *gin.Context) {
	userId := ctx.GetUint("userId")

	otherUserID, err := strconv.ParseUint(ctx.Param("userId"), 10, 32)
	if err != nil {
		utils.Error(ctx, http.StatusBadRequest, "用户ID格式错误")
		return
	}

	pagination := utils.ParsePagination(ctx)
	page := pagination.Page
	pageSize := pagination.PageSize

	messages, total, err := c.service.GetMessages(userId, uint(otherUserID), pageSize, (page-1)*pageSize)
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "获取消息失败")
		return
	}

	utils.Success(ctx, "查询成功", gin.H{
		"list":  messages,
		"total": total,
		"page":  page,
	})
}

// GetConversations 获取会话列表
func (c *MessageController) GetConversations(ctx *gin.Context) {
	userId := ctx.GetUint("userId")

	conversations, err := c.service.GetConversations(userId)
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "获取会话列表失败")
		return
	}

	utils.Success(ctx, "查询成功", conversations)
}

// MarkAsRead 标记单条消息已读
func (c *MessageController) MarkAsRead(ctx *gin.Context) {
	userId := ctx.GetUint("userId")

	messageID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.Error(ctx, http.StatusBadRequest, "消息ID格式错误")
		return
	}

	if err := c.service.MarkAsRead(uint(messageID), userId); err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "标记已读失败")
		return
	}

	utils.Success(ctx, "已标记为已读", nil)
}

// MarkConversationAsRead 标记整个会话已读
func (c *MessageController) MarkConversationAsRead(ctx *gin.Context) {
	userId := ctx.GetUint("userId")

	otherUserID, err := strconv.ParseUint(ctx.Param("userId"), 10, 32)
	if err != nil {
		utils.Error(ctx, http.StatusBadRequest, "用户ID格式错误")
		return
	}

	if err := c.service.MarkConversationAsRead(userId, uint(otherUserID)); err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "标记已读失败")
		return
	}

	utils.Success(ctx, "会话已标记为已读", nil)
}

// GetUnreadCount 获取未读私信数
func (c *MessageController) GetUnreadCount(ctx *gin.Context) {
	userId := ctx.GetUint("userId")

	count, err := c.service.GetUnreadCount(userId)
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "获取未读数失败")
		return
	}

	utils.Success(ctx, "查询成功", gin.H{"count": count})
}

// DeleteMessage 删除私信
func (c *MessageController) DeleteMessage(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.Error(ctx, http.StatusBadRequest, "消息ID格式错误")
		return
	}

	if err := c.service.DeleteMessage(uint(id)); err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "删除失败")
		return
	}

	utils.Success(ctx, "删除成功", nil)
}
