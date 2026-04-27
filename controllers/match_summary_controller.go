package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/services"
	"github.com/shaonianqiutan/backend/utils"
	"gorm.io/gorm"
)

// MatchSummaryController 比赛总结控制器
type MatchSummaryController struct {
	service *services.MatchSummaryService
	db      *gorm.DB
}

// NewMatchSummaryController 创建比赛总结控制器
func NewMatchSummaryController(service *services.MatchSummaryService, db *gorm.DB) *MatchSummaryController {
	return &MatchSummaryController{service: service, db: db}
}

// Create 创建比赛 M1
func (c *MatchSummaryController) Create(ctx *gin.Context) {
	var input models.MatchSummaryCreate
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	teamIDStr := ctx.Param("teamId")
	if teamIDStr != "" {
		teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
		if err == nil {
			input.TeamID = uint(teamID)
		}
	}

	userID := ctx.GetUint("userId")
	summary, err := c.service.Create(userID, &input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 异步通知参赛球员
	go func() {
		notificationHelper := NewNotificationHelper(c.db)
		coachName := "教练"
		if summary.Coach != nil {
			coachName = summary.Coach.Name
		}
		if summary.PlayerIDs != "" && summary.PlayerIDs != "null" {
			var playerIDs []uint
			if err := json.Unmarshal([]byte(summary.PlayerIDs), &playerIDs); err == nil {
				for _, playerID := range playerIDs {
					notificationHelper.NotifyMatchSummaryCreated(
						playerID,
						coachName,
						summary.MatchName,
						summary.ID,
					)
				}
			}
		}
	}()

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    c.buildDetailResponse(summary),
	})
}

// ListByTeam 列出球队比赛 M2
func (c *MatchSummaryController) ListByTeam(ctx *gin.Context) {
	teamIDStr := ctx.Param("teamId")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "无效的球队ID"})
		return
	}

	status := ctx.Query("status")
	pagination := utils.ParsePaginationWithSize(ctx, 10)
	page := pagination.Page
	pageSize := pagination.PageSize

	summaries, total, err := c.service.ListByTeam(uint(teamID), status, page, pageSize)
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

// Get 获取比赛详情 M3
func (c *MatchSummaryController) Get(ctx *gin.Context) {
	id, err := c.parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "无效的ID"})
		return
	}

	summary, err := c.service.GetByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"success": false, "error": "比赛不存在"})
		return
	}

	userID := ctx.GetUint("userId")
	if !c.service.CanAccessMatchSummary(uint(id), userID) {
		ctx.JSON(http.StatusForbidden, gin.H{"success": false, "error": "无权查看该比赛"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    c.buildDetailResponse(summary),
	})
}

// Update 更新比赛 M4
func (c *MatchSummaryController) Update(ctx *gin.Context) {
	id, err := c.parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "无效的ID"})
		return
	}

	var input models.MatchSummaryUpdate
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	userID := ctx.GetUint("userId")
	summary, err := c.service.Update(uint(id), userID, &input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    c.buildDetailResponse(summary),
	})
}

// Delete 删除比赛 M5
func (c *MatchSummaryController) Delete(ctx *gin.Context) {
	id, err := c.parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "无效的ID"})
		return
	}

	userID := ctx.GetUint("userId")
	if err := c.service.Delete(uint(id), userID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "删除成功",
	})
}

// UpdateCover 更新封面图 M8
func (c *MatchSummaryController) UpdateCover(ctx *gin.Context) {
	id, err := c.parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "无效的ID"})
		return
	}

	var input models.CoverImageUpdate
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	userID := ctx.GetUint("userId")
	summary, err := c.service.UpdateCoverImage(uint(id), userID, input.CoverImage)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    c.buildDetailResponse(summary),
	})
}

// SubmitCoachSummary 教练提交点评 M9
func (c *MatchSummaryController) SubmitCoachSummary(ctx *gin.Context) {
	id, err := c.parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "无效的ID"})
		return
	}

	var input models.CoachSummarySubmit
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	userID := ctx.GetUint("userId")
	summary, err := c.service.SubmitCoachSummary(uint(id), userID, &input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 异步通知球员
	notificationHelper := NewNotificationHelper(c.db)
	go func() {
		reviews, _ := c.service.ListPlayerReviews(uint(id))
		coachName := "教练"
		if summary.Coach != nil {
			coachName = summary.Coach.Name
		}
		for _, review := range reviews {
			notificationHelper.NotifyMatchSummaryComplete(
				review.PlayerID,
				coachName,
				summary.MatchName,
				summary.ID,
			)
		}
	}()

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    c.buildDetailResponse(summary),
	})
}

// GetPendingCount 获取待处理数量
func (c *MatchSummaryController) GetPendingCount(ctx *gin.Context) {
	teamIDStr := ctx.Param("teamId")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "无效的球队ID"})
		return
	}

	count, err := c.service.GetPendingCount(uint(teamID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"count": count,
		},
	})
}

// ListByCoach 列出教练发起的比赛总结
func (c *MatchSummaryController) ListByCoach(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	status := ctx.Query("status")
	pagination := utils.ParsePaginationWithSize(ctx, 10)
	page := pagination.Page
	pageSize := pagination.PageSize

	summaries, total, err := c.service.ListByCoach(userID, status, page, pageSize)
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

// ============================================================
// 辅助方法
// ============================================================

// buildDetailResponse 构建比赛详情响应
func (c *MatchSummaryController) buildDetailResponse(m *models.MatchSummary) models.MatchSummaryResponse {
	resp := models.MatchSummaryResponse{
		ID:              m.ID,
		TeamID:          m.TeamID,
		CoachID:         m.CoachID,
		Status:          m.Status,
		MatchName:       m.MatchName,
		MatchDate:       m.MatchDate,
		Opponent:        m.Opponent,
		Location:        m.Location,
		MatchFormat:     m.MatchFormat,
		OurScore:        m.OurScore,
		OppScore:        m.OppScore,
		Result:          m.Result,
		CoverImage:      m.CoverImage,
		CoachOverall:    m.CoachOverall,
		CoachTactic:     m.CoachTactic,
		CoachKeyMoments: m.CoachKeyMoments,
		CreatedAt:       m.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if m.Team != nil {
		resp.TeamName = m.Team.Name
	}
	if m.Coach != nil {
		resp.CoachName = m.Coach.Name
	}

	// PlayerIDs
	if m.PlayerIDs != "" && m.PlayerIDs != "null" {
		var ids []uint
		json.Unmarshal([]byte(m.PlayerIDs), &ids)
		resp.PlayerIDs = ids
		resp.PlayerCount = len(ids)
	} else {
		resp.PlayerIDs = []uint{}
	}

	// Videos JSON
	if m.Videos != "" && m.Videos != "null" {
		var videos []models.MatchVideoResponse
		json.Unmarshal([]byte(m.Videos), &videos)
		resp.Videos = videos
	} else {
		resp.Videos = []models.MatchVideoResponse{}
	}

	// 球员自评列表
	reviews, err := c.service.ListPlayerReviews(m.ID)
	if err == nil && len(reviews) > 0 {
		reviewResponses := make([]models.PlayerReviewResponse, len(reviews))
		for i, r := range reviews {
			reviewResponses[i] = r.ToResponse()
		}
		resp.PlayerReviews = reviewResponses
		resp.SubmittedCount = len(reviews)
	}

	return resp
}

// SubmitCoachReview 教练提交整体点评（别名，对应新路由）
// POST /api/match-summaries/:id/coach-review
func (c *MatchSummaryController) SubmitCoachReview(ctx *gin.Context) {
	c.SubmitCoachSummary(ctx)
}

// UpdateCoverImage 更新封面图（别名，对应新路由）
// POST /api/match-summaries/:id/cover-image
func (c *MatchSummaryController) UpdateCoverImage(ctx *gin.Context) {
	c.UpdateCover(ctx)
}

// GetStats 获取比赛统计
// GET /api/match-summaries/stats
func (c *MatchSummaryController) GetStats(ctx *gin.Context) {
	userID := ctx.GetUint("userId")

	// 获取用户的俱乐部 clubID
	var club struct {
		ID uint
	}
	if err := c.db.Table("clubs").Select("id").Where("user_id = ?", userID).First(&club).Error; err != nil {
		// 非俱乐部管理员，返回个人统计
		stats, err := c.service.GetMatchStats(0)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"success": true, "data": stats})
		return
	}

	stats, err := c.service.GetMatchStats(club.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true, "data": stats})
}

// Remind 催办未提交自评的球员
// POST /api/match-summaries/:id/remind
func (c *MatchSummaryController) Remind(ctx *gin.Context) {
	id, err := c.parseID(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "无效的ID"})
		return
	}

	var input struct {
		PlayerIDs []uint `json:"playerIds"`
		Message   string `json:"message"`
	}
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	userID := ctx.GetUint("userId")
	pendingIDs, matchName, err := c.service.RemindPlayers(uint(id), userID, input.PlayerIDs, input.Message)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	sent := 0
	failed := 0
	if len(pendingIDs) > 0 {
		helper := NewNotificationHelper(c.db)
		for _, pid := range pendingIDs {
			if err := helper.NotifyMatchPlayerReminder(pid, matchName, uint(id)); err != nil {
				failed++
			} else {
				sent++
			}
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": models.RemindResult{
			Sent:   sent,
			Failed: failed,
		},
	})
}

// parseID 解析路径参数ID
func (c *MatchSummaryController) parseID(ctx *gin.Context) (uint64, error) {
	idStr := ctx.Param("id")
	return strconv.ParseUint(idStr, 10, 32)
}
