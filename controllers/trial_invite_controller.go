package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/config"
	"github.com/shaonianqiutan/backend/middleware"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/utils"
	"gorm.io/gorm"
)

// TrialInviteController 试训邀请控制器
type TrialInviteController struct{}

type trialInviteResponse struct {
	ID           uint                     `json:"id"`
	SenderID     uint                     `json:"sender_id"`
	SenderName   string                   `json:"sender_name"`
	SenderRole   models.UserRole          `json:"sender_role"`
	SenderAvatar string                   `json:"sender_avatar"`
	PlayerID     uint                     `json:"player_id"`
	PlayerName   string                   `json:"player_name"`
	TrialDate    string                   `json:"trial_date"`
	TrialTime    string                   `json:"trial_time"`
	Location     string                   `json:"location"`
	ContactName  string                   `json:"contact_name"`
	ContactPhone string                   `json:"contact_phone"`
	Note         string                   `json:"note"`
	Status       models.TrialInviteStatus `json:"status"`
	ResponseNote string                   `json:"response_note"`
	CreatedAt    time.Time                `json:"created_at"`
	UpdatedAt    time.Time                `json:"updated_at"`
	RespondedAt  *time.Time               `json:"responded_at"`
}

// NewTrialInviteController 创建试训邀请控制器
func NewTrialInviteController() *TrialInviteController {
	return &TrialInviteController{}
}

// CreateTrialInvite 创建试训邀请
func (ctrl *TrialInviteController) CreateTrialInvite(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	var req struct {
		PlayerID     uint   `json:"player_id" binding:"required"`
		TrialDate    string `json:"trial_date" binding:"required"`
		TrialTime    string `json:"trial_time"`
		Location     string `json:"location" binding:"required"`
		ContactName  string `json:"contact_name" binding:"required"`
		ContactPhone string `json:"contact_phone" binding:"required"`
		Note         string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	db := config.GetDB()

	// 校验球员存在
	var player models.User
	if err := db.Where("id = ? AND role = ? AND status = ?", req.PlayerID, "user", "active").First(&player).Error; err != nil {
		utils.Error(c, http.StatusBadRequest, "球员不存在")
		return
	}

	invite := models.TrialInvite{
		SenderID:     userID,
		PlayerID:     req.PlayerID,
		TrialDate:    req.TrialDate,
		TrialTime:    req.TrialTime,
		Location:     req.Location,
		ContactName:  req.ContactName,
		ContactPhone: req.ContactPhone,
		Note:         req.Note,
		Status:       models.TrialInvitePending,
	}

	if err := db.Create(&invite).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "创建邀请失败")
		return
	}

	// 发送通知给球员
	var sender models.User
	senderName := "邀请方"
	if err := db.First(&sender, userID).Error; err == nil {
		senderName = ctrl.userDisplayName(sender)
	}
	notification := models.Notification{
		UserID:   req.PlayerID,
		Type:     models.NotificationTypeTrialInvite,
		Title:    "收到试训邀请",
		Content:  "您收到一份新的试训邀请，请尽快查看并确认。",
		IsRead:   false,
		Priority: 3,
	}
	notification.SetData(&models.NotificationData{
		TriggerUserID:   userID,
		TriggerUserName: senderName,
		TargetType:      "trial_invite",
		TargetID:        invite.ID,
		Link:            "/user-dashboard?tab=my_invitations",
	})
	_ = db.Create(&notification).Error

	utils.Success(c, "邀请发送成功", gin.H{"invite": invite})
}

// GetMyTrialInvites 获取当前球员收到的试训邀请
func (ctrl *TrialInviteController) GetMyTrialInvites(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	db := config.GetDB()
	var invites []models.TrialInvite
	if err := db.Where("player_id = ?", userID).Order("created_at DESC").Find(&invites).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取试训邀请失败")
		return
	}

	responses := ctrl.buildTrialInviteResponses(db, invites)
	utils.Success(c, "获取成功", gin.H{
		"list":  responses,
		"total": len(responses),
	})
}

// RespondTrialInvite 处理试训邀请
func (ctrl *TrialInviteController) RespondTrialInvite(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	inviteID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || inviteID == 0 {
		utils.Error(c, http.StatusBadRequest, "邀请ID无效")
		return
	}

	var req struct {
		Status       models.TrialInviteStatus `json:"status" binding:"required"`
		ResponseNote string                   `json:"response_note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.Status != models.TrialInviteAccepted && req.Status != models.TrialInviteDeclined {
		utils.Error(c, http.StatusBadRequest, "仅支持接受或拒绝试训邀请")
		return
	}

	db := config.GetDB()
	var invite models.TrialInvite
	if err := db.Where("id = ? AND player_id = ?", uint(inviteID), userID).First(&invite).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Error(c, http.StatusNotFound, "试训邀请不存在")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "获取试训邀请失败")
		return
	}
	if invite.Status != models.TrialInvitePending {
		utils.Error(c, http.StatusBadRequest, "该试训邀请已处理")
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":        req.Status,
		"response_note": req.ResponseNote,
		"responded_at":  &now,
	}
	if err := db.Model(&invite).Updates(updates).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "处理试训邀请失败")
		return
	}
	invite.Status = req.Status
	invite.ResponseNote = req.ResponseNote
	invite.RespondedAt = &now

	ctrl.notifyTrialInviteResponse(db, invite, userID)

	responses := ctrl.buildTrialInviteResponses(db, []models.TrialInvite{invite})
	if len(responses) == 0 {
		utils.Success(c, "处理成功", gin.H{"invite": invite})
		return
	}
	utils.Success(c, "处理成功", gin.H{"invite": responses[0]})
}

func (ctrl *TrialInviteController) buildTrialInviteResponses(db *gorm.DB, invites []models.TrialInvite) []trialInviteResponse {
	if len(invites) == 0 {
		return []trialInviteResponse{}
	}

	userIDs := make([]uint, 0, len(invites)*2)
	seen := map[uint]bool{}
	for _, invite := range invites {
		if invite.SenderID != 0 && !seen[invite.SenderID] {
			userIDs = append(userIDs, invite.SenderID)
			seen[invite.SenderID] = true
		}
		if invite.PlayerID != 0 && !seen[invite.PlayerID] {
			userIDs = append(userIDs, invite.PlayerID)
			seen[invite.PlayerID] = true
		}
	}

	var users []models.User
	_ = db.Where("id IN ?", userIDs).Find(&users).Error
	userMap := make(map[uint]models.User, len(users))
	for _, user := range users {
		userMap[user.ID] = user
	}

	responses := make([]trialInviteResponse, 0, len(invites))
	for _, invite := range invites {
		sender := userMap[invite.SenderID]
		player := userMap[invite.PlayerID]
		responses = append(responses, trialInviteResponse{
			ID:           invite.ID,
			SenderID:     invite.SenderID,
			SenderName:   ctrl.userDisplayName(sender),
			SenderRole:   sender.Role,
			SenderAvatar: sender.Avatar,
			PlayerID:     invite.PlayerID,
			PlayerName:   ctrl.userDisplayName(player),
			TrialDate:    invite.TrialDate,
			TrialTime:    invite.TrialTime,
			Location:     invite.Location,
			ContactName:  invite.ContactName,
			ContactPhone: invite.ContactPhone,
			Note:         invite.Note,
			Status:       invite.Status,
			ResponseNote: invite.ResponseNote,
			CreatedAt:    invite.CreatedAt,
			UpdatedAt:    invite.UpdatedAt,
			RespondedAt:  invite.RespondedAt,
		})
	}
	return responses
}

func (ctrl *TrialInviteController) notifyTrialInviteResponse(db *gorm.DB, invite models.TrialInvite, playerID uint) {
	var player models.User
	playerName := "球员"
	if err := db.First(&player, playerID).Error; err == nil {
		playerName = ctrl.userDisplayName(player)
	}

	statusText := "已拒绝"
	if invite.Status == models.TrialInviteAccepted {
		statusText = "已接受"
	}

	notification := models.Notification{
		UserID:   invite.SenderID,
		Type:     models.NotificationTypeTrialInvite,
		Title:    "试训邀请" + statusText,
		Content:  fmt.Sprintf("%s %s了您的试训邀请。", playerName, statusText),
		IsRead:   false,
		Priority: 3,
	}
	notification.SetData(&models.NotificationData{
		TriggerUserID:   playerID,
		TriggerUserName: playerName,
		TargetType:      "trial_invite",
		TargetID:        invite.ID,
		Link:            "/scout-map",
	})
	_ = db.Create(&notification).Error
}

func (ctrl *TrialInviteController) userDisplayName(user models.User) string {
	if user.Nickname != "" {
		return user.Nickname
	}
	if user.Name != "" {
		return user.Name
	}
	if user.ID == 0 {
		return ""
	}
	return fmt.Sprintf("用户%d", user.ID)
}
