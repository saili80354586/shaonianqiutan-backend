package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/services"
	"github.com/shaonianqiutan/backend/utils"
)

type AnalystLevelController struct {
	service *services.AnalystLevelService
}

func NewAnalystLevelController(service *services.AnalystLevelService) *AnalystLevelController {
	return &AnalystLevelController{service: service}
}

func (ctrl *AnalystLevelController) ListLevels(c *gin.Context) {
	levels, err := ctrl.service.ListLevels()
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取分析师等级失败")
		return
	}
	utils.Success(c, "", gin.H{"list": levels})
}

func (ctrl *AnalystLevelController) GetMyLevel(c *gin.Context) {
	analyst, latestApplication, growth, histories, err := ctrl.service.GetAnalystLevelProfileByUserID(c.GetUint("userId"))
	if err != nil {
		ctrl.writeAnalystLevelError(c, err)
		return
	}
	utils.Success(c, "", gin.H{
		"analyst":            analyst,
		"latest_application": latestApplication,
		"growth":             growth,
		"level_histories":    histories,
	})
}

func (ctrl *AnalystLevelController) SubmitMyApplication(c *gin.Context) {
	var req services.AnalystLevelApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	app, err := ctrl.service.SubmitApplication(c.GetUint("userId"), req, time.Now())
	if err != nil {
		ctrl.writeAnalystLevelError(c, err)
		return
	}
	utils.Success(c, "等级申请已提交", gin.H{"application": app})
}

func (ctrl *AnalystLevelController) ListMyApplications(c *gin.Context) {
	pagination := utils.ParsePaginationWithSize(c, 20)
	apps, total, err := ctrl.service.ListMyApplications(c.GetUint("userId"), pagination.Page, pagination.PageSize)
	if err != nil {
		ctrl.writeAnalystLevelError(c, err)
		return
	}
	utils.Success(c, "", gin.H{
		"list":     apps,
		"total":    total,
		"page":     pagination.Page,
		"pageSize": pagination.PageSize,
	})
}

func (ctrl *AnalystLevelController) ListAdminApplications(c *gin.Context) {
	pagination := utils.ParsePaginationWithSize(c, 20)
	apps, total, err := ctrl.service.ListApplications(pagination.Page, pagination.PageSize, c.Query("status"))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取等级申请失败")
		return
	}
	utils.Success(c, "", gin.H{
		"list":     apps,
		"total":    total,
		"page":     pagination.Page,
		"pageSize": pagination.PageSize,
	})
}

func (ctrl *AnalystLevelController) GetApplication(c *gin.Context) {
	appID, ok := parseAnalystLevelUintParam(c, "id", "无效的等级申请ID")
	if !ok {
		return
	}
	app, err := ctrl.service.GetApplication(appID)
	if err != nil {
		ctrl.writeAnalystLevelError(c, err)
		return
	}
	utils.Success(c, "", gin.H{"application": app})
}

func (ctrl *AnalystLevelController) ReviewApplication(c *gin.Context) {
	appID, ok := parseAnalystLevelUintParam(c, "id", "无效的等级申请ID")
	if !ok {
		return
	}
	var req services.AnalystLevelReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	app, err := ctrl.service.ReviewApplication(appID, c.GetUint("userId"), req, time.Now())
	if err != nil {
		ctrl.writeAnalystLevelError(c, err)
		return
	}
	utils.Success(c, "等级申请已审核", gin.H{"application": app})
}

func (ctrl *AnalystLevelController) ListAdminAnalysts(c *gin.Context) {
	pagination := utils.ParsePaginationWithSize(c, 20)
	analysts, total, err := ctrl.service.ListAnalysts(pagination.Page, pagination.PageSize)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取分析师等级列表失败")
		return
	}
	utils.Success(c, "", gin.H{
		"list":     analysts,
		"total":    total,
		"page":     pagination.Page,
		"pageSize": pagination.PageSize,
	})
}

func (ctrl *AnalystLevelController) SetAnalystLevel(c *gin.Context) {
	analystID, ok := parseAnalystLevelUintParam(c, "id", "无效的分析师ID")
	if !ok {
		return
	}
	var req services.AnalystLevelSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	analyst, err := ctrl.service.SetAnalystLevel(analystID, c.GetUint("userId"), req, time.Now())
	if err != nil {
		ctrl.writeAnalystLevelError(c, err)
		return
	}
	utils.Success(c, "分析师等级已更新", gin.H{"analyst": analyst})
}

func (ctrl *AnalystLevelController) SetOfficialPartnership(c *gin.Context) {
	analystID, ok := parseAnalystLevelUintParam(c, "id", "无效的分析师ID")
	if !ok {
		return
	}
	var req services.AnalystOfficialPartnershipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	analyst, err := ctrl.service.SetOfficialPartnership(analystID, c.GetUint("userId"), req, time.Now())
	if err != nil {
		ctrl.writeAnalystLevelError(c, err)
		return
	}
	utils.Success(c, "官方合作状态已更新", gin.H{"analyst": analyst})
}

func (ctrl *AnalystLevelController) RefreshAnalystGrowth(c *gin.Context) {
	analystID, ok := parseAnalystLevelUintParam(c, "id", "无效的分析师ID")
	if !ok {
		return
	}
	growth, err := ctrl.service.RefreshGrowthSnapshot(analystID, time.Now())
	if err != nil {
		ctrl.writeAnalystLevelError(c, err)
		return
	}
	utils.Success(c, "成长分已刷新", gin.H{"growth": growth})
}

func (ctrl *AnalystLevelController) ApplyLevelSuggestion(c *gin.Context) {
	analystID, ok := parseAnalystLevelUintParam(c, "id", "无效的分析师ID")
	if !ok {
		return
	}
	var req services.AnalystLevelSuggestionActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	analyst, err := ctrl.service.ApplyLevelSuggestion(analystID, c.GetUint("userId"), req, time.Now())
	if err != nil {
		ctrl.writeAnalystLevelError(c, err)
		return
	}
	utils.Success(c, "等级建议已采纳", gin.H{"analyst": analyst})
}

func (ctrl *AnalystLevelController) IgnoreLevelSuggestion(c *gin.Context) {
	analystID, ok := parseAnalystLevelUintParam(c, "id", "无效的分析师ID")
	if !ok {
		return
	}
	var req services.AnalystLevelSuggestionActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	growth, err := ctrl.service.IgnoreLevelSuggestion(analystID, c.GetUint("userId"), req, time.Now())
	if err != nil {
		ctrl.writeAnalystLevelError(c, err)
		return
	}
	utils.Success(c, "等级建议已忽略", gin.H{"growth": growth})
}

func (ctrl *AnalystLevelController) ListLevelHistories(c *gin.Context) {
	analystID, ok := parseAnalystLevelUintParam(c, "id", "无效的分析师ID")
	if !ok {
		return
	}
	pagination := utils.ParsePaginationWithSize(c, 20)
	histories, err := ctrl.service.ListLevelHistories(analystID, pagination.Page, pagination.PageSize)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取等级历史失败")
		return
	}
	utils.Success(c, "", gin.H{
		"list":     histories,
		"page":     pagination.Page,
		"pageSize": pagination.PageSize,
	})
}

func (ctrl *AnalystLevelController) writeAnalystLevelError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrAnalystLevelInvalid):
		utils.Error(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, services.ErrAnalystLevelApplicationConflict):
		utils.Error(c, http.StatusConflict, err.Error())
	case errors.Is(err, services.ErrAnalystLevelApplicationNotFound):
		utils.Error(c, http.StatusNotFound, err.Error())
	case errors.Is(err, services.ErrAnalystLevelAnalystInvalid):
		utils.Error(c, http.StatusForbidden, err.Error())
	default:
		utils.Error(c, http.StatusInternalServerError, "分析师等级操作失败")
	}
}

func parseAnalystLevelUintParam(c *gin.Context, key, message string) (uint, bool) {
	raw := c.Param(key)
	id, err := strconv.ParseUint(raw, 10, 32)
	if err != nil || id == 0 {
		utils.Error(c, http.StatusBadRequest, message)
		return 0, false
	}
	return uint(id), true
}
