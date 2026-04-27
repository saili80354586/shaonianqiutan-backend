package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/services"
	"github.com/shaonianqiutan/backend/utils"
)

// FootballExperienceController 足球经历控制器
type FootballExperienceController struct {
	coachService *services.CoachService
}

// NewFootballExperienceController 创建足球经历控制器
func NewFootballExperienceController(coachService *services.CoachService) *FootballExperienceController {
	return &FootballExperienceController{coachService: coachService}
}

// GetFootballExperiences 获取教练的足球经历列表
func (c *FootballExperienceController) GetFootballExperiences(ctx *gin.Context) {
	userID := ctx.GetUint("userId")

	coach, err := c.coachService.GetOrCreateCoach(userID)
	if err != nil {
		utils.ServerError(ctx, "获取数据失败")
		return
	}

	experiences, err := c.coachService.GetFootballExperiences(coach.ID)
	if err != nil {
		utils.ServerError(ctx, "获取足球经历失败")
		return
	}

	utils.Success(ctx, "", experiences)
}

// CreateFootballExperience 创建足球经历
func (c *FootballExperienceController) CreateFootballExperience(ctx *gin.Context) {
	userID := ctx.GetUint("userId")

	var req struct {
		Stage     string `json:"stage" binding:"required"`
		TeamName  string `json:"teamName" binding:"required"`
		Position  string `json:"position"`
		StartYear int    `json:"startYear" binding:"required"`
		EndYear   int    `json:"endYear"`
		Level     string `json:"level"`
		Honors    string `json:"honors"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	// 验证阶段
	validStages := map[string]bool{
		string(models.StagePrimary):      true,
		string(models.StageMiddle):       true,
		string(models.StageHigh):         true,
		string(models.StageUniversity):   true,
		string(models.StageProfessional): true,
	}
	if !validStages[req.Stage] {
		utils.Error(ctx, 400, "无效的阶段类型")
		return
	}

	coach, err := c.coachService.GetOrCreateCoach(userID)
	if err != nil {
		utils.ServerError(ctx, "获取数据失败")
		return
	}

	exp, err := c.coachService.CreateFootballExperience(
		coach.ID,
		req.Stage,
		req.TeamName,
		req.Position,
		req.StartYear,
		req.EndYear,
		req.Level,
		req.Honors,
	)
	if err != nil {
		utils.ServerError(ctx, "创建失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, gin.H{
		"id": exp.ID,
	}, "创建成功")
}

// UpdateFootballExperience 更新足球经历
func (c *FootballExperienceController) UpdateFootballExperience(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	expID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的经历ID")
		return
	}

	var req struct {
		Stage     string `json:"stage"`
		TeamName  string `json:"teamName"`
		Position  string `json:"position"`
		StartYear int    `json:"startYear"`
		EndYear   int    `json:"endYear"`
		Level     string `json:"level"`
		Honors    string `json:"honors"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	coach, err := c.coachService.GetOrCreateCoach(userID)
	if err != nil {
		utils.ServerError(ctx, "获取数据失败")
		return
	}

	err = c.coachService.UpdateFootballExperience(
		coach.ID,
		uint(expID),
		req.Stage,
		req.TeamName,
		req.Position,
		req.StartYear,
		req.EndYear,
		req.Level,
		req.Honors,
	)
	if err != nil {
		utils.ServerError(ctx, "更新失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "更新成功")
}

// DeleteFootballExperience 删除足球经历
func (c *FootballExperienceController) DeleteFootballExperience(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	expID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的经历ID")
		return
	}

	coach, err := c.coachService.GetOrCreateCoach(userID)
	if err != nil {
		utils.ServerError(ctx, "获取数据失败")
		return
	}

	err = c.coachService.DeleteFootballExperience(coach.ID, uint(expID))
	if err != nil {
		utils.ServerError(ctx, "删除失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "删除成功")
}
