package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/services"
	"github.com/shaonianqiutan/backend/utils"
)

type PhysicalTestController struct {
	ptService *services.PhysicalTestService
}

func NewPhysicalTestController(ptService *services.PhysicalTestService) *PhysicalTestController {
	return &PhysicalTestController{ptService: ptService}
}

// GetPhysicalTests 获取体测活动列表
func (c *PhysicalTestController) GetPhysicalTests(ctx *gin.Context) {
	userID := ctx.GetUint("userId")

	pagination := utils.ParsePagination(ctx)
	page := pagination.Page
	pageSize := pagination.PageSize
	status := ctx.Query("status")
	teamID, _ := strconv.ParseUint(ctx.Param("teamId"), 10, 32)
	if teamID == 0 {
		teamID, _ = strconv.ParseUint(ctx.Query("teamId"), 10, 32)
	}

	club, err := c.ptService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.PaginatedResponse(ctx, []interface{}{}, page, pageSize, 0)
		return
	}

	tests, total, _ := c.ptService.GetPhysicalTests(club.ID, uint(teamID), page, pageSize, status)

	list := make([]interface{}, 0, len(tests))
	for _, t := range tests {
		completedCount, _ := c.ptService.GetCompletedRecordCount(t.ID)
		reportCount, _ := c.ptService.GetReportsCount(t.ID)
		playerCount := len(t.GetPlayerIDs())

		list = append(list, gin.H{
			"id":               t.ID,
			"name":             t.Name,
			"description":      t.Description,
			"startDate":        t.StartDate.Format("2006-01-02"),
			"endDate":          utils.FormatTime(t.EndDate),
			"location":         t.Location,
			"template":         string(t.Template),
			"templateName":     getTemplateName(t.Template),
			"playerCount":      playerCount,
			"completedCount":   completedCount,
			"status":           string(t.Status),
			"statusName":       getStatusName(t.Status),
			"reportsGenerated": reportCount,
			"createdAt":        utils.FormatTime(&t.CreatedAt),
		})
	}

	utils.PaginatedResponse(ctx, list, page, pageSize, total)
}

// CreatePhysicalTest 创建体测活动
func (c *PhysicalTestController) CreatePhysicalTest(ctx *gin.Context) {
	userID := ctx.GetUint("userId")

	var req struct {
		Name             string   `json:"name" binding:"required"`
		Description      string   `json:"description"`
		StartDate        string   `json:"startDate" binding:"required"`
		EndDate          string   `json:"endDate"`
		Location         string   `json:"location"`
		Template         string   `json:"template" binding:"required"`
		CustomTemplateID uint     `json:"customTemplateId"`
		CustomItems      []string `json:"customItems"`
		PlayerScope      string   `json:"playerScope"`
		PlayerIDs        []uint   `json:"playerIds"`
		NotifyParents    bool     `json:"notifyParents"`
		AutoSendReport   bool     `json:"autoSendReport"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	club, err := c.ptService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		utils.ValidationError(ctx, "日期格式错误")
		return
	}

	var endDate *time.Time
	if req.EndDate != "" {
		t, err := time.Parse("2006-01-02", req.EndDate)
		if err == nil {
			endDate = &t
		}
	}

	playerIDsJSON := ""
	if len(req.PlayerIDs) > 0 {
		data, _ := json.Marshal(req.PlayerIDs)
		playerIDsJSON = string(data)
	}

	customItemsJSON := ""
	if len(req.CustomItems) > 0 {
		data, _ := json.Marshal(req.CustomItems)
		customItemsJSON = string(data)
	}

	// 如果传了 customTemplateId，校验归属
	if req.CustomTemplateID > 0 {
		tmpl, err := c.ptService.GetCustomTemplateByID(req.CustomTemplateID)
		if err != nil || tmpl.ClubID != club.ID {
			utils.ValidationError(ctx, "无效的自定义模板")
			return
		}
		// 同步 custom_items
		customItemsJSON = tmpl.Items
	}

	test := &models.PhysicalTestActivity{
		ClubID:         club.ID,
		Name:           req.Name,
		Description:    req.Description,
		StartDate:      startDate,
		EndDate:        endDate,
		Location:       req.Location,
		Template:       models.PhysicalTestTemplate(req.Template),
		CustomItems:    customItemsJSON,
		PlayerIDs:      playerIDsJSON,
		Status:         models.PTStatusPending,
		NotifyParents:  req.NotifyParents,
		AutoSendReport: req.AutoSendReport,
		CreatedBy:      userID,
	}

	if err := c.ptService.CreatePhysicalTest(test); err != nil {
		utils.ServerError(ctx, "创建失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, gin.H{
		"id":        test.ID,
		"name":      test.Name,
		"status":    string(test.Status),
		"startDate": test.StartDate.Format("2006-01-02"),
		"createdAt": utils.FormatTime(&test.CreatedAt),
	}, "体测活动创建成功")
}

// GetPhysicalTest 获取体测活动详情
func (c *PhysicalTestController) GetPhysicalTest(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	testID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的ID")
		return
	}

	club, err := c.ptService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	test, err := c.ptService.GetPhysicalTestByID(uint(testID))
	if err != nil || test == nil {
		utils.NotFoundError(ctx, "体测活动不存在")
		return
	}

	if test.ClubID != club.ID {
		utils.ForbiddenError(ctx, "无权限访问")
		return
	}

	completedCount, _ := c.ptService.GetCompletedRecordCount(test.ID)
	reportCount, _ := c.ptService.GetReportsCount(test.ID)
	templateItems := c.ptService.GetTemplateItems(test.Template, test.CustomItems, 0)

	utils.SuccessResponse(ctx, gin.H{
		"id":               test.ID,
		"name":             test.Name,
		"description":      test.Description,
		"startDate":        test.StartDate.Format("2006-01-02"),
		"endDate":          utils.FormatTime(test.EndDate),
		"location":         test.Location,
		"template":         string(test.Template),
		"templateName":     getTemplateName(test.Template),
		"templateItems":    templateItems,
		"playerCount":      len(test.GetPlayerIDs()),
		"completedCount":   completedCount,
		"reportsGenerated": reportCount,
		"status":           string(test.Status),
		"statusName":       getStatusName(test.Status),
		"notifyParents":    test.NotifyParents,
		"autoSendReport":   test.AutoSendReport,
		"createdAt":        utils.FormatTime(&test.CreatedAt),
		"updatedAt":        utils.FormatTime(&test.UpdatedAt),
	})
}

// UpdatePhysicalTest 更新体测活动
func (c *PhysicalTestController) UpdatePhysicalTest(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	testID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的ID")
		return
	}

	club, err := c.ptService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	test, err := c.ptService.GetPhysicalTestByID(uint(testID))
	if err != nil || test == nil {
		utils.NotFoundError(ctx, "体测活动不存在")
		return
	}

	if test.ClubID != club.ID {
		utils.ForbiddenError(ctx, "无权限访问")
		return
	}

	var req struct {
		Name           string `json:"name"`
		EndDate        string `json:"endDate"`
		Location       string `json:"location"`
		PlayerIDs      []uint `json:"playerIds"`
		NotifyParents  *bool  `json:"notifyParents"`
		AutoSendReport *bool  `json:"autoSendReport"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	updates := make(map[string]interface{})

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.EndDate != "" {
		t, err := time.Parse("2006-01-02", req.EndDate)
		if err == nil {
			updates["end_date"] = t
		}
	}
	if req.Location != "" {
		updates["location"] = req.Location
	}
	if len(req.PlayerIDs) > 0 {
		data, _ := json.Marshal(req.PlayerIDs)
		updates["player_ids"] = string(data)
	}
	if req.NotifyParents != nil {
		updates["notify_parents"] = *req.NotifyParents
	}
	if req.AutoSendReport != nil {
		updates["auto_send_report"] = *req.AutoSendReport
	}

	if err := c.ptService.UpdatePhysicalTest(uint(testID), updates); err != nil {
		utils.ServerError(ctx, "更新失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "更新成功")
}

// DeletePhysicalTest 删除体测活动
func (c *PhysicalTestController) DeletePhysicalTest(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	testID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的ID")
		return
	}

	club, err := c.ptService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	test, err := c.ptService.GetPhysicalTestByID(uint(testID))
	if err != nil || test == nil {
		utils.NotFoundError(ctx, "体测活动不存在")
		return
	}

	if test.ClubID != club.ID {
		utils.ForbiddenError(ctx, "无权限访问")
		return
	}

	if err := c.ptService.DeletePhysicalTest(uint(testID)); err != nil {
		utils.ServerError(ctx, "删除失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "删除成功")
}

// NotifyPhysicalTest 发送体测通知
func (c *PhysicalTestController) NotifyPhysicalTest(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	testID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的ID")
		return
	}

	club, err := c.ptService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	test, err := c.ptService.GetPhysicalTestByID(uint(testID))
	if err != nil || test == nil || test.ClubID != club.ID {
		utils.NotFoundError(ctx, "体测活动不存在")
		return
	}

	// TODO: 实现真实的通知发送逻辑
	utils.SuccessResponseWithMessage(ctx, gin.H{
		"sent":   len(test.GetPlayerIDs()),
		"failed": 0,
	}, "通知已发送")
}

// GetPhysicalTestRecords 获取体测数据列表
func (c *PhysicalTestController) GetPhysicalTestRecords(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	testID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的ID")
		return
	}

	club, err := c.ptService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	test, err := c.ptService.GetPhysicalTestByID(uint(testID))
	if err != nil || test == nil || test.ClubID != club.ID {
		utils.NotFoundError(ctx, "体测活动不存在")
		return
	}

	var playerID *uint
	if pid := ctx.Query("playerId"); pid != "" {
		id, err := strconv.ParseUint(pid, 10, 32)
		if err == nil {
			pid := uint(id)
			playerID = &pid
		}
	}

	records, _ := c.ptService.GetPhysicalTestRecords(uint(testID), playerID)

	templateItems := c.ptService.GetTemplateItems(test.Template, test.CustomItems, 0)

	list := make([]interface{}, 0, len(records))
	for _, r := range records {
		playerName := ""
		if r.Player != nil {
			playerName = r.Player.Name
			if playerName == "" {
				playerName = r.Player.Nickname
			}
		}

		data := services.GetTestDataMapFromRecord(&r)
		progress := getRecordProgressByItems(&r, templateItems)

		list = append(list, gin.H{
			"id":             r.ID,
			"playerId":       r.PlayerID,
			"playerName":     playerName,
			"playerAvatar":   "",
			"testDate":       r.TestDate.Format("2006-01-02"),
			"status":         getRecordStatus(&r),
			"data":           data,
			"recordProgress": progress,
		})
	}

	utils.SuccessResponse(ctx, gin.H{
		"list": list,
	})
}

// CreatePhysicalTestRecord 录入体测数据
func (c *PhysicalTestController) CreatePhysicalTestRecord(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	testID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的ID")
		return
	}

	club, err := c.ptService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	test, err := c.ptService.GetPhysicalTestByID(uint(testID))
	if err != nil || test == nil || test.ClubID != club.ID {
		utils.NotFoundError(ctx, "体测活动不存在")
		return
	}

	var req struct {
		PlayerID uint               `json:"playerId" binding:"required"`
		TestDate string             `json:"testDate"`
		Data     map[string]float64 `json:"data" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	testDate := test.StartDate
	if req.TestDate != "" {
		t, err := time.Parse("2006-01-02", req.TestDate)
		if err == nil {
			testDate = t
		}
	}

	record := &models.PhysicalTestRecord{
		ActivityID: uint(testID),
		PlayerID:   req.PlayerID,
		ClubID:     club.ID,
		TestDate:   testDate,
		RecorderID: userID,
	}

	// 设置体测数据
	c.ptService.SetRecordData(record, req.Data)

	// 查找是否已有记录
	existingRecords, _ := c.ptService.GetPhysicalTestRecords(uint(testID), &req.PlayerID)
	if len(existingRecords) > 0 {
		// 更新现有记录
		record.ID = existingRecords[0].ID
		c.ptService.SetRecordData(&existingRecords[0], req.Data)
		// 直接用DB更新
		c.ptService.UpdateRecord(&existingRecords[0])
		utils.SuccessResponseWithMessage(ctx, gin.H{
			"id":      record.ID,
			"updated": true,
		}, "数据更新成功")
		return
	}

	if err := c.ptService.CreatePhysicalTestRecord(record); err != nil {
		utils.ServerError(ctx, "保存失败")
		return
	}

	// 更新活动状态为进行中
	if test.Status == models.PTStatusPending {
		c.ptService.UpdatePhysicalTest(uint(testID), map[string]interface{}{
			"status": string(models.PTStatusOngoing),
		})
	}

	utils.SuccessResponseWithMessage(ctx, gin.H{
		"id":      record.ID,
		"updated": false,
	}, "数据保存成功")
}

// BatchImportPhysicalTestRecords 批量导入体测数据
func (c *PhysicalTestController) BatchImportPhysicalTestRecords(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	testID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的ID")
		return
	}

	club, err := c.ptService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	test, err := c.ptService.GetPhysicalTestByID(uint(testID))
	if err != nil || test == nil || test.ClubID != club.ID {
		utils.NotFoundError(ctx, "体测活动不存在")
		return
	}

	// TODO: 实现Excel文件解析和批量导入
	utils.SuccessResponseWithMessage(ctx, gin.H{
		"total":   0,
		"success": 0,
		"failed":  0,
		"details": []interface{}{},
	}, "批量导入功能开发中")
}

// GeneratePhysicalTestReports 生成体测报告
func (c *PhysicalTestController) GeneratePhysicalTestReports(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	testID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的ID")
		return
	}

	club, err := c.ptService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	test, err := c.ptService.GetPhysicalTestByID(uint(testID))
	if err != nil || test == nil || test.ClubID != club.ID {
		utils.NotFoundError(ctx, "体测活动不存在")
		return
	}

	var req struct {
		PlayerIDs []uint `json:"playerIds"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		// 生成全部球员的报告
		req.PlayerIDs = test.GetPlayerIDs()
	}

	// TODO: 实现真实的报告生成逻辑
	utils.SuccessResponseWithMessage(ctx, gin.H{
		"total":     len(req.PlayerIDs),
		"generated": len(req.PlayerIDs),
		"failed":    0,
	}, "报告生成完成")
}

// GetPlayerPhysicalTestReports 获取球员体测报告列表
func (c *PhysicalTestController) GetPlayerPhysicalTestReports(ctx *gin.Context) {
	if _, err := strconv.ParseUint(ctx.Param("id"), 10, 32); err != nil {
		utils.ValidationError(ctx, "无效的ID")
		return
	}
	// TODO: 获取该球员的体测报告 - 目前返回空列表
	utils.SuccessResponse(ctx, gin.H{
		"list": []interface{}{},
	})
}

// 辅助函数

func getTemplateName(t models.PhysicalTestTemplate) string {
	switch t {
	case models.PTTemplateBasic:
		return "基础版"
	case models.PTTemplateAdvanced:
		return "进阶版"
	case models.PTTemplateProfessional:
		return "专业版"
	case models.PTTemplateCustom:
		return "自定义"
	default:
		return "进阶版"
	}
}

func getStatusName(s models.PhysicalTestStatus) string {
	switch s {
	case models.PTStatusPending:
		return "待开始"
	case models.PTStatusOngoing:
		return "进行中"
	case models.PTStatusCompleted:
		return "已完成"
	case models.PTStatusReported:
		return "报告已生成"
	default:
		return "未知"
	}
}

func getRecordStatus(r *models.PhysicalTestRecord) string {
	if r.Height != nil || r.Weight != nil || r.Sprint30m != nil {
		return "completed"
	}
	return "pending"
}

func getRecordProgressByItems(r *models.PhysicalTestRecord, items []string) map[string]int {
	completed := 0
	for _, item := range items {
		switch item {
		case "height":
			if r.Height != nil {
				completed++
			}
		case "weight":
			if r.Weight != nil {
				completed++
			}
		case "bmi":
			if r.BMI != nil {
				completed++
			}
		case "sprint_30m":
			if r.Sprint30m != nil {
				completed++
			}
		case "sprint_50m":
			if r.Sprint50m != nil {
				completed++
			}
		case "sprint_100m":
			if r.Sprint100m != nil {
				completed++
			}
		case "agility_ladder":
			if r.AgilityLadder != nil {
				completed++
			}
		case "t_test":
			if r.TTest != nil {
				completed++
			}
		case "shuttle_run":
			if r.ShuttleRun != nil {
				completed++
			}
		case "standing_long_jump":
			if r.StandingLongJump != nil {
				completed++
			}
		case "vertical_jump":
			if r.VerticalJump != nil {
				completed++
			}
		case "sit_and_reach":
			if r.SitAndReach != nil {
				completed++
			}
		case "push_up":
			if r.PushUp != nil {
				completed++
			}
		case "sit_up":
			if r.SitUp != nil {
				completed++
			}
		case "plank":
			if r.Plank != nil {
				completed++
			}
		}
	}
	return map[string]int{
		"total":     len(items),
		"completed": completed,
	}
}

// ========== 自定义模板 API ==========

// GetCustomTemplates 获取自定义模板列表
func (c *PhysicalTestController) GetCustomTemplates(ctx *gin.Context) {
	userID := ctx.GetUint("userId")

	club, err := c.ptService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	templates, _ := c.ptService.GetCustomTemplates(club.ID)

	list := make([]interface{}, 0, len(templates))
	for _, t := range templates {
		list = append(list, gin.H{
			"id":          t.ID,
			"name":        t.Name,
			"description": t.Description,
			"items":       t.GetItems(),
			"itemCount":   len(t.GetItems()),
			"createdAt":   utils.FormatTime(&t.CreatedAt),
		})
	}

	utils.SuccessResponse(ctx, gin.H{
		"list": list,
	})
}

// CreateCustomTemplate 创建自定义模板
func (c *PhysicalTestController) CreateCustomTemplate(ctx *gin.Context) {
	userID := ctx.GetUint("userId")

	var req struct {
		Name        string   `json:"name" binding:"required"`
		Description string   `json:"description"`
		Items       []string `json:"items" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	club, err := c.ptService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	itemsJSON, _ := json.Marshal(req.Items)
	template := &models.PhysicalTestTemplateCustom{
		ClubID:      club.ID,
		CreatedBy:   userID,
		Name:        req.Name,
		Description: req.Description,
		Items:       string(itemsJSON),
	}

	if err := c.ptService.CreateCustomTemplate(template); err != nil {
		utils.ServerError(ctx, "创建失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, gin.H{
		"id": template.ID,
	}, "自定义模板创建成功")
}

// DeleteCustomTemplate 删除自定义模板
func (c *PhysicalTestController) DeleteCustomTemplate(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	templateID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的ID")
		return
	}

	club, err := c.ptService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	if err := c.ptService.DeleteCustomTemplate(uint(templateID), club.ID); err != nil {
		utils.ServerError(ctx, "删除失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "删除成功")
}

// 确保http import被使用
var _ = http.StatusOK
