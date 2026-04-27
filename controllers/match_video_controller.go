package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/services"
	"gorm.io/gorm"
)

// MatchVideoController 比赛视频控制器
type MatchVideoController struct {
	service *services.MatchSummaryService
	db      *gorm.DB
}

// NewMatchVideoController 创建比赛视频控制器
func NewMatchVideoController(service *services.MatchSummaryService, db *gorm.DB) *MatchVideoController {
	return &MatchVideoController{service: service, db: db}
}

// AddVideo 添加视频链接 M6
// POST /api/match-summaries/:id/videos
func (c *MatchVideoController) AddVideo(ctx *gin.Context) {
	idStr := ctx.Param("id")
	matchID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "无效的比赛ID"})
		return
	}

	var input models.MatchVideoCreate
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	userID := ctx.GetUint("userId")

	// 获取比赛信息以获得teamId
	summary, err := c.service.GetByID(uint(matchID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"success": false, "error": "比赛不存在"})
		return
	}

	video, err := c.service.AddVideo(uint(matchID), summary.TeamID, userID, &input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    video.ToResponse(),
	})
}

// DeleteVideo 删除视频链接 M7
// DELETE /api/match-summaries/:id/videos/:videoId
func (c *MatchVideoController) DeleteVideo(ctx *gin.Context) {
	videoIDStr := ctx.Param("videoId")
	videoID, err := strconv.ParseUint(videoIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "无效的视频ID"})
		return
	}

	userID := ctx.GetUint("userId")
	if err := c.service.DeleteVideo(uint(videoID), userID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "删除成功",
	})
}

// ListVideos 获取比赛视频列表
// GET /api/match-summaries/:id/videos
func (c *MatchVideoController) ListVideos(ctx *gin.Context) {
	idStr := ctx.Param("id")
	matchID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "无效的比赛ID"})
		return
	}

	videos, err := c.service.ListVideos(uint(matchID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	list := make([]models.MatchVideoResponse, len(videos))
	for i, v := range videos {
		list[i] = v.ToResponse()
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    list,
	})
}
