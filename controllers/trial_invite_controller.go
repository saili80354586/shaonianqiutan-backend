package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/config"
	"github.com/shaonianqiutan/backend/middleware"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/utils"
)

// TrialInviteController 试训邀请控制器
type TrialInviteController struct{}

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
	go func() {
		notification := models.Notification{
			UserID:  req.PlayerID,
			Type:    "trial_invite",
			Title:   "收到试训邀请",
			Content: "您收到一份新的试训邀请，请尽快查看并确认。",
			Data:    "",
		}
		db.Create(&notification)
	}()

	utils.Success(c, "邀请发送成功", gin.H{"invite": invite})
}
