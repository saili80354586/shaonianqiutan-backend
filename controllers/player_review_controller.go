package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/services"
	"github.com/shaonianqiutan/backend/utils"
	"gorm.io/gorm"
)

// PlayerReviewController 球员自评控制器
type PlayerReviewController struct {
	service *services.MatchSummaryService
	db      *gorm.DB
}

// NewPlayerReviewController 创建球员自评控制器
func NewPlayerReviewController(service *services.MatchSummaryService, db *gorm.DB) *PlayerReviewController {
	return &PlayerReviewController{service: service, db: db}
}

// SubmitReview 球员提交自评 P3
// POST /api/match-summaries/:id/player-review
func (c *PlayerReviewController) SubmitReview(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "无效的比赛ID"})
		return
	}

	var input models.PlayerReviewSubmit
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	userID := ctx.GetUint("userId")
	review, err := c.service.SubmitPlayerReview(uint(id), userID, &input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 异步通知教练
	notificationHelper := NewNotificationHelper(c.db)
	go func() {
		summary, _ := c.service.GetByID(uint(id))
		if summary != nil {
			notificationHelper.NotifyMatchCoachReminder(summary.CoachID, summary.MatchName, summary.ID)
		}
	}()

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    review.ToResponse(),
	})
}

// GetReview 获取球员自评详情 P2
// GET /api/match-summaries/:id/player-review
func (c *PlayerReviewController) GetReview(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "无效的比赛ID"})
		return
	}

	userID := ctx.GetUint("userId")
	review, err := c.service.GetPlayerReview(uint(id), userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"success": false, "error": "自评不存在"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    review.ToResponse(),
	})
}

// ListByPlayer 获取球员比赛列表 P1
// GET /api/players/:playerId/match-summaries
func (c *PlayerReviewController) ListByPlayer(ctx *gin.Context) {
	playerIDStr := ctx.Param("playerId")
	playerID, err := strconv.ParseUint(playerIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "无效的球员ID"})
		return
	}

	// 权限校验：只能查看自己的
	userID := ctx.GetUint("userId")
	if uint(playerID) != userID {
		ctx.JSON(http.StatusForbidden, gin.H{"success": false, "error": "无权查看其他球员的比赛"})
		return
	}

	pagination := utils.ParsePaginationWithSize(ctx, 10)
	page := pagination.Page
	pageSize := pagination.PageSize

	summaries, total, err := c.service.ListByPlayer(uint(playerID), page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	list := make([]models.MatchSummaryListResponse, len(summaries))
	for i, s := range summaries {
		list[i] = s.ToListResponse()
		submittedCount, _ := c.service.GetSubmittedCount(s.ID)
		list[i].SubmittedCount = int(submittedCount)
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

// SubmitCoachPlayerReview 教练对单个球员提交评分点评
// POST /api/match-summaries/:id/coach-player-review
func (c *PlayerReviewController) SubmitCoachPlayerReview(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "无效的比赛ID"})
		return
	}

	var input models.CoachPlayerReviewSubmit
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	userID := ctx.GetUint("userId")
	review, err := c.service.SubmitCoachPlayerReview(uint(id), userID, &input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    review.ToResponse(),
	})
}
