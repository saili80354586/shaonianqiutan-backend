package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/repositories"
	"github.com/shaonianqiutan/backend/utils"
)

// CoachTeamHomeController 教练视角球队主页控制器
type CoachTeamHomeController struct {
	teamHomeRepo *repositories.TeamHomeRepository
	teamRepo     *repositories.TeamRepository
}

// NewCoachTeamHomeController 创建控制器
func NewCoachTeamHomeController(teamHomeRepo *repositories.TeamHomeRepository, teamRepo *repositories.TeamRepository) *CoachTeamHomeController {
	return &CoachTeamHomeController{
		teamHomeRepo: teamHomeRepo,
		teamRepo:     teamRepo,
	}
}

// GetTeamHome 获取球队主页配置（教练视角）
func (c *CoachTeamHomeController) GetTeamHome(ctx *gin.Context) {
	teamIDStr := ctx.Param("teamId")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	// 获取球队信息
	team, err := c.teamRepo.FindByID(uint(teamID))
	if err != nil || team == nil {
		utils.ValidationError(ctx, "球队不存在")
		return
	}

	// 获取或创建主页配置
	home, err := c.teamHomeRepo.FindOrCreate(uint(teamID))
	if err != nil {
		utils.ServerError(ctx, "获取失败")
		return
	}

	// 获取荣誉列表
	honors, _ := c.teamHomeRepo.GetHonors(uint(teamID))

	// 获取球员和教练数量
	playerCount := 0
	coachCount := 0
	if players, _, err := c.teamRepo.GetPlayers(uint(teamID), "", "", ""); err == nil {
		playerCount = len(players)
	}
	if coaches, err := c.teamRepo.GetCoaches(uint(teamID), ""); err == nil {
		coachCount = len(coaches)
	}

	utils.SuccessResponse(ctx, models.CoachTeamHomeResponse{
		TeamID:      uint(teamID),
		TeamName:    team.Name,
		AgeGroup:    team.AgeGroup,
		Hero:        home.Hero,
		About:       home.About,
		Honors:      honors,
		Contact:     home.Contact,
		PlayerCount: playerCount,
		CoachCount:  coachCount,
	})
}

// SaveTeamHome 保存球队主页配置
func (c *CoachTeamHomeController) SaveTeamHome(ctx *gin.Context) {
	teamIDStr := ctx.Param("teamId")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	var req struct {
		Hero        *models.TeamHomeHero    `json:"hero"`
		About       *models.TeamHomeAbout   `json:"about"`
		Honors      []models.TeamHonor      `json:"honors"`
		Contact     *models.TeamHomeContact `json:"contact"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	home, err := c.teamHomeRepo.FindOrCreate(uint(teamID))
	if err != nil {
		utils.ServerError(ctx, "获取失败")
		return
	}

	// 更新字段
	if req.Hero != nil {
		home.Hero = *req.Hero
	}
	if req.About != nil {
		home.About = *req.About
	}
	if req.Contact != nil {
		home.Contact = *req.Contact
	}
	if req.Honors != nil {
		home.Honors = req.Honors
	}

	if err := c.teamHomeRepo.Save(home); err != nil {
		utils.ServerError(ctx, "保存失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, gin.H{"id": home.ID}, "保存成功")
}

// UpdateHero 更新 Hero 配置
func (c *CoachTeamHomeController) UpdateHero(ctx *gin.Context) {
	teamIDStr := ctx.Param("teamId")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	var req models.TeamHomeHero
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	if err := c.teamHomeRepo.UpdateHero(uint(teamID), &req); err != nil {
		utils.ServerError(ctx, "更新失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "更新成功")
}

// UpdateAbout 更新 About 配置
func (c *CoachTeamHomeController) UpdateAbout(ctx *gin.Context) {
	teamIDStr := ctx.Param("teamId")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	var req models.TeamHomeAbout
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	if err := c.teamHomeRepo.UpdateAbout(uint(teamID), &req); err != nil {
		utils.ServerError(ctx, "更新失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "更新成功")
}

// UpdateContact 更新联系方式
func (c *CoachTeamHomeController) UpdateContact(ctx *gin.Context) {
	teamIDStr := ctx.Param("teamId")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	var req models.TeamHomeContact
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	if err := c.teamHomeRepo.UpdateContact(uint(teamID), &req); err != nil {
		utils.ServerError(ctx, "更新失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "更新成功")
}

// AddHonor 添加荣誉
func (c *CoachTeamHomeController) AddHonor(ctx *gin.Context) {
	teamIDStr := ctx.Param("teamId")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	var req models.TeamHonor
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	home, err := c.teamHomeRepo.FindOrCreate(uint(teamID))
	if err != nil {
		utils.ServerError(ctx, "获取失败")
		return
	}

	// 添加到列表
	req.ID = uint(len(home.Honors) + 1)
	home.Honors = append(home.Honors, req)

	if err := c.teamHomeRepo.Save(home); err != nil {
		utils.ServerError(ctx, "保存失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, req, "添加成功")
}

// DeleteHonor 删除荣誉
func (c *CoachTeamHomeController) DeleteHonor(ctx *gin.Context) {
	teamIDStr := ctx.Param("teamId")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	honorIDStr := ctx.Param("honorId")
	honorID, err := strconv.ParseUint(honorIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的荣誉ID")
		return
	}

	home, err := c.teamHomeRepo.FindByTeamID(uint(teamID))
	if err != nil || home == nil {
		utils.ValidationError(ctx, "球队主页不存在")
		return
	}

	// 过滤掉要删除的荣誉
	var newHonors []models.TeamHonor
	for _, h := range home.Honors {
		if h.ID != uint(honorID) {
			newHonors = append(newHonors, h)
		}
	}
	home.Honors = newHonors

	if err := c.teamHomeRepo.Save(home); err != nil {
		utils.ServerError(ctx, "删除失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "删除成功")
}

// GetDynamics 获取动态列表
func (c *CoachTeamHomeController) GetDynamics(ctx *gin.Context) {
	teamIDStr := ctx.Param("teamId")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	dynamics, err := c.teamHomeRepo.GetDynamics(uint(teamID))
	if err != nil {
		utils.ServerError(ctx, "获取失败")
		return
	}

	utils.SuccessResponse(ctx, dynamics)
}

// AddDynamic 添加动态
func (c *CoachTeamHomeController) AddDynamic(ctx *gin.Context) {
	teamIDStr := ctx.Param("teamId")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	var req models.TeamDynamic
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	if err := c.teamHomeRepo.AddDynamic(uint(teamID), &req); err != nil {
		utils.ServerError(ctx, "添加失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, req, "添加成功")
}

// DeleteDynamic 删除动态
func (c *CoachTeamHomeController) DeleteDynamic(ctx *gin.Context) {
	teamIDStr := ctx.Param("teamId")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	dynamicIDStr := ctx.Param("dynamicId")
	dynamicID, err := strconv.ParseUint(dynamicIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的动态ID")
		return
	}

	if err := c.teamHomeRepo.DeleteDynamic(uint(teamID), uint(dynamicID)); err != nil {
		utils.ServerError(ctx, "删除失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "删除成功")
}

// UpdateDynamics 更新动态列表
func (c *CoachTeamHomeController) UpdateDynamics(ctx *gin.Context) {
	teamIDStr := ctx.Param("teamId")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	var req []models.TeamDynamic
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	if err := c.teamHomeRepo.UpdateDynamics(uint(teamID), req); err != nil {
		utils.ServerError(ctx, "更新失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "更新成功")
}
