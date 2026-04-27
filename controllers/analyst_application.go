package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/middleware"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/services"
	"github.com/shaonianqiutan/backend/utils"
)

// AnalystApplicationController 分析师申请控制器
type AnalystApplicationController struct {
	appService *services.AnalystApplicationService
}

func NewAnalystApplicationController(
	appService *services.AnalystApplicationService,
) *AnalystApplicationController {
	return &AnalystApplicationController{appService: appService}
}

// CreateApplication 创建分析师申请
func (ctrl *AnalystApplicationController) CreateApplication(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	var req services.CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	app, err := ctrl.appService.CreateApplication(userID, &req)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.Success(c, "申请提交成功，请等待审核", gin.H{"application": app})
}

// GetMyApplication 获取我的申请
func (ctrl *AnalystApplicationController) GetMyApplication(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	app, err := ctrl.appService.GetMyApplication(userID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取申请信息失败")
		return
	}

	utils.Success(c, "", gin.H{"application": app})
}

// GetApplicationList 获取申请列表（管理后台）
func (ctrl *AnalystApplicationController) GetApplicationList(c *gin.Context) {
	pagination := utils.ParsePaginationWithSize(c, 10)
	page := pagination.Page
	pageSize := pagination.PageSize

	statusStr := c.Query("status")
	var status *models.ApplicationStatus
	if statusStr != "" {
		s := models.ApplicationStatus(statusStr)
		status = &s
	}

	apps, total, err := ctrl.appService.GetApplicationList(page, pageSize, status)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取申请列表失败")
		return
	}

	utils.Success(c, "", gin.H{
		"list":     apps,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// ReviewApplicationRequest 审核申请请求
type ReviewApplicationRequest struct {
	Status models.ApplicationStatus `json:"status" binding:"required,oneof=pending approved rejected"`
	Remark string                   `json:"remark"`
}

// ReviewApplication 审核申请（管理后台）
func (ctrl *AnalystApplicationController) ReviewApplication(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的申请ID")
		return
	}

	var req ReviewApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	err = ctrl.appService.ReviewApplication(uint(id), req.Status, req.Remark)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.Success(c, "审核完成", nil)
}
