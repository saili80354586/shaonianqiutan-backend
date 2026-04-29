package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/middleware"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/services"
	"github.com/shaonianqiutan/backend/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserController 用户控制器
type UserController struct {
	authService         *services.AuthService
	physicalTestService *services.PhysicalTestService
	db                  *gorm.DB
}

func NewUserController(authService *services.AuthService, physicalTestService *services.PhysicalTestService, db *gorm.DB) *UserController {
	return &UserController{
		authService:         authService,
		physicalTestService: physicalTestService,
		db:                  db,
	}
}

// GetProfile 获取用户资料
func (ctrl *UserController) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	user, err := ctrl.authService.GetUserByID(userID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取用户信息失败")
		return
	}

	utils.Success(c, "", gin.H{"user": user})
}

// UpdateProfile 更新用户资料
func (ctrl *UserController) UpdateProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	var req services.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := ctrl.authService.UpdateUser(userID, &req)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "更新用户信息失败")
		return
	}

	utils.Success(c, "用户信息更新成功", gin.H{"user": user})
}

// GetPublicProfile 获取公开资料
func (ctrl *UserController) GetPublicProfile(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的用户ID")
		return
	}

	user, err := ctrl.authService.GetUserByID(uint(userID))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取用户信息失败")
		return
	}
	if user == nil {
		utils.Error(c, http.StatusNotFound, "用户不存在")
		return
	}

	privacy := parsePrivacySettings(user.PrivacySettings)
	publicName := ""
	if privacy.ShowRealName {
		publicName = user.Name
	}

	// 返回公开资料（只包含公开字段）
	publicUser := gin.H{
		"id":         user.ID,
		"nickname":   user.Nickname,
		"name":       publicName,
		"avatar":     user.Avatar,
		"role":       user.Role,
		"position":   user.Position,
		"age":        user.Age,
		"height":     user.Height,
		"weight":     user.Weight,
		"foot":       user.Foot,
		"province":   user.Province,
		"city":       user.City,
		"club":       user.Club,
		"bio":        "", // User模型没有bio字段，可以留空或从其他表获取
		"created_at": user.CreatedAt,
	}

	utils.Success(c, "", gin.H{"user": publicUser})
}

// GetPlayerProfile 获取球员档案
func (ctrl *UserController) GetPlayerProfile(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的用户ID")
		return
	}

	user, err := ctrl.authService.GetUserByID(uint(userID))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取用户信息失败")
		return
	}
	if user == nil {
		utils.Error(c, http.StatusNotFound, "用户不存在")
		return
	}

	// 查询最新体测记录
	physicalTests := gin.H{}
	if ctrl.physicalTestService != nil {
		if record, err := ctrl.physicalTestService.GetLatestPhysicalTestRecordByPlayer(uint(userID)); err == nil && record != nil {
			data := services.GetTestDataMapFromRecord(record)
			// 将下划线命名转为 camelCase 以兼容前端 PersonalHomepage 的 PhysicalTests 组件
			physicalTests = gin.H{
				"height":           data["height"],
				"weight":           data["weight"],
				"bmi":              data["bmi"],
				"sprint30m":        data["sprint_30m"],
				"sprint50m":        data["sprint_50m"],
				"sprint100m":       data["sprint_100m"],
				"agilityLadder":    data["agility_ladder"],
				"tTest":            data["t_test"],
				"shuttleRun":       data["shuttle_run"],
				"standingLongJump": data["standing_long_jump"],
				"verticalJump":     data["vertical_jump"],
				"sitAndReach":      data["sit_and_reach"],
				"pushUp":           data["push_up"],
				"sitUp":            data["sit_up"],
				"plank":            data["plank"],
				"testDate":         record.TestDate.Format("2006-01-02"),
			}
		}
	}

	privacy := parsePrivacySettings(user.PrivacySettings)
	publicName := ""
	if privacy.ShowRealName {
		publicName = user.Name
	}

	// 构建球员完整档案
	playerProfile := gin.H{
		"id":             user.ID,
		"name":           publicName,
		"nickname":       user.Nickname,
		"avatar":         user.Avatar,
		"position":       user.Position,
		"secondPosition": user.SecondPosition,
		"age":            user.Age,
		"height":         user.Height,
		"weight":         user.Weight,
		"foot":           user.Foot,
		"gender":         user.Gender,
		"province":       user.Province,
		"city":           user.City,
		"club":           user.Club,
		"startYear":      user.StartYear,
		"faRegistered":   user.FARegistered,
		"association":    user.Association,
		"jerseyColor":    user.JerseyColor,
		"jerseyNumber":   user.JerseyNumber,
		"physicalTests":  physicalTests,
		"createdAt":      user.CreatedAt,
	}

	utils.Success(c, "", gin.H{"player": playerProfile})
}

// GetPublicReports 获取公开报告
func (ctrl *UserController) GetPublicReports(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的用户ID")
		return
	}

	var reports []models.Report
	if err := ctrl.db.
		Where("user_id = ? AND status = ?", uint(userID), models.ReportStatusCompleted).
		Order("created_at DESC").
		Limit(10).
		Find(&reports).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取报告失败")
		return
	}

	list := make([]gin.H, 0, len(reports))
	for _, report := range reports {
		list = append(list, gin.H{
			"id":              report.ID,
			"order_id":        report.OrderID,
			"user_id":         report.UserID,
			"analyst_id":      report.AnalystID,
			"player_name":     report.PlayerName,
			"player_position": report.PlayerPosition,
			"title":           "视频分析报告",
			"description":     firstNonEmpty(report.Summary, report.Suggestions),
			"content":         firstNonEmpty(report.Summary, report.Suggestions),
			"rating":          report.OverallRating,
			"overall_rating":  report.OverallRating,
			"offense_rating":  report.OffenseRating,
			"defense_rating":  report.DefenseRating,
			"strengths":       homepageStringList(report.Strengths),
			"weaknesses":      homepageStringList(report.Weaknesses),
			"suggestions":     report.Suggestions,
			"potential":       report.Potential,
			"status":          report.Status,
			"created_at":      report.CreatedAt,
			"updated_at":      report.UpdatedAt,
		})
	}

	utils.Success(c, "", gin.H{"reports": list})
}

// GrowthRecord 成长记录
type GrowthRecord struct {
	ID          string   `json:"id"`
	Date        string   `json:"date"`
	Type        string   `json:"type"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Content     string   `json:"content"`
	Location    string   `json:"location"`
	Opponent    string   `json:"opponent"`
	Score       string   `json:"score"`
	VideoURL    string   `json:"videoUrl"`
	Images      []string `json:"images"`
	Tags        []string `json:"tags"`
	Height      float64  `json:"height"`
	Weight      float64  `json:"weight"`
	Speed       float64  `json:"speed"`
	Strength    float64  `json:"strength"`
	Technique   float64  `json:"technique"`
	Tactics     float64  `json:"tactics"`
	Note        string   `json:"note"`
	MatchName   string   `json:"matchName"`
	Result      string   `json:"result"`
	Goals       int      `json:"goals"`
	Assists     int      `json:"assists"`
	PlayTime    int      `json:"playTime"`
	Feeling     string   `json:"feeling"`
	Photos      []string `json:"photos"`
	Videos      []string `json:"videos"`
}

func normalizeGrowthRecordType(recordType string) models.GrowthRecordType {
	switch models.GrowthRecordType(recordType) {
	case models.GrowthRecordTypeMilestone,
		models.GrowthRecordTypeAchievement,
		models.GrowthRecordTypeTraining,
		models.GrowthRecordTypeMatch,
		models.GrowthRecordTypePhysical:
		return models.GrowthRecordType(recordType)
	default:
		return models.GrowthRecordTypePhysical
	}
}

func growthRecordToDTO(r models.GrowthRecord) GrowthRecord {
	var stats struct {
		Height    float64  `json:"height"`
		Weight    float64  `json:"weight"`
		Speed     float64  `json:"speed"`
		Strength  float64  `json:"strength"`
		Technique float64  `json:"technique"`
		Tactics   float64  `json:"tactics"`
		Note      string   `json:"note"`
		Location  string   `json:"location"`
		Opponent  string   `json:"opponent"`
		Score     string   `json:"score"`
		VideoURL  string   `json:"videoUrl"`
		Images    []string `json:"images"`
		Tags      []string `json:"tags"`
		MatchName string   `json:"matchName"`
		Result    string   `json:"result"`
		Goals     int      `json:"goals"`
		Assists   int      `json:"assists"`
		PlayTime  int      `json:"playTime"`
		Feeling   string   `json:"feeling"`
		Photos    []string `json:"photos"`
		Videos    []string `json:"videos"`
	}
	if r.StatsJSON != "" {
		_ = json.Unmarshal([]byte(r.StatsJSON), &stats)
	}
	return GrowthRecord{
		ID:          strconv.FormatUint(uint64(r.ID), 10),
		Date:        r.RecordDate.Format("2006-01-02"),
		Type:        string(r.RecordType),
		Title:       r.Title,
		Description: r.Content,
		Content:     r.Content,
		Location:    stats.Location,
		Opponent:    stats.Opponent,
		Score:       stats.Score,
		VideoURL:    stats.VideoURL,
		Images:      stats.Images,
		Tags:        stats.Tags,
		Height:      stats.Height,
		Weight:      stats.Weight,
		Speed:       stats.Speed,
		Strength:    stats.Strength,
		Technique:   stats.Technique,
		Tactics:     stats.Tactics,
		Note:        stats.Note,
		MatchName:   stats.MatchName,
		Result:      stats.Result,
		Goals:       stats.Goals,
		Assists:     stats.Assists,
		PlayTime:    stats.PlayTime,
		Feeling:     stats.Feeling,
		Photos:      stats.Photos,
		Videos:      stats.Videos,
	}
}

func growthRecordModel(userID uint, rec GrowthRecord) (models.GrowthRecord, error) {
	recordDate, err := time.Parse("2006-01-02", rec.Date)
	if err != nil || recordDate.IsZero() {
		recordDate = time.Now()
	}

	title := rec.Title
	if title == "" {
		title = "成长记录"
	}
	content := rec.Content
	if content == "" {
		content = rec.Description
	}
	if rec.Note == "" {
		rec.Note = content
	}

	stats := map[string]interface{}{
		"height":    rec.Height,
		"weight":    rec.Weight,
		"speed":     rec.Speed,
		"strength":  rec.Strength,
		"technique": rec.Technique,
		"tactics":   rec.Tactics,
		"note":      rec.Note,
		"location":  rec.Location,
		"opponent":  rec.Opponent,
		"score":     rec.Score,
		"videoUrl":  rec.VideoURL,
		"images":    rec.Images,
		"tags":      rec.Tags,
		"matchName": rec.MatchName,
		"result":    rec.Result,
		"goals":     rec.Goals,
		"assists":   rec.Assists,
		"playTime":  rec.PlayTime,
		"feeling":   rec.Feeling,
		"photos":    rec.Photos,
		"videos":    rec.Videos,
	}
	statsJSON, err := json.Marshal(stats)
	if err != nil {
		return models.GrowthRecord{}, err
	}

	return models.GrowthRecord{
		UserID:     userID,
		RecordDate: recordDate,
		RecordType: normalizeGrowthRecordType(rec.Type),
		Title:      title,
		Content:    content,
		StatsJSON:  string(statsJSON),
	}, nil
}

// GetGrowthRecords 获取成长记录
func (ctrl *UserController) GetGrowthRecords(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	var dbRecords []models.GrowthRecord
	if err := ctrl.db.Where("user_id = ?", userID).
		Order("record_date DESC").Find(&dbRecords).Error; err != nil {
		utils.ServerError(c, "查询失败")
		return
	}

	records := make([]GrowthRecord, 0, len(dbRecords))
	for _, r := range dbRecords {
		records = append(records, growthRecordToDTO(r))
	}

	utils.Success(c, "", gin.H{"records": records})
}

// SaveGrowthRecords 批量保存成长记录
func (ctrl *UserController) SaveGrowthRecords(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "读取请求失败")
		return
	}

	var req struct {
		Records []GrowthRecord `json:"records"`
	}
	if err := json.Unmarshal(body, &req); err == nil && req.Records != nil {
		for _, rec := range req.Records {
			gr, err := growthRecordModel(userID, rec)
			if err != nil {
				utils.Error(c, http.StatusBadRequest, "成长记录格式错误")
				return
			}
			if err := ctrl.db.Create(&gr).Error; err != nil {
				utils.ServerError(c, "保存失败")
				return
			}
		}

		utils.Success(c, "成长记录保存成功", nil)
		return
	}

	var single GrowthRecord
	if err := json.Unmarshal(body, &single); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	gr, err := growthRecordModel(userID, single)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "成长记录格式错误")
		return
	}
	if gr.Title == "" {
		utils.Error(c, http.StatusBadRequest, "标题不能为空")
		return
	}
	if err := ctrl.db.Create(&gr).Error; err != nil {
		utils.ServerError(c, "保存失败")
		return
	}
	utils.Success(c, "成长记录保存成功", growthRecordToDTO(gr))
}

func (ctrl *UserController) UpdateGrowthRecord(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	recordID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的成长记录ID")
		return
	}

	var req GrowthRecord
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	var existing models.GrowthRecord
	if err := ctrl.db.Where("id = ? AND user_id = ?", recordID, userID).First(&existing).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "成长记录不存在")
		return
	}

	updated, err := growthRecordModel(userID, req)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "成长记录格式错误")
		return
	}

	if err := ctrl.db.Model(&existing).Updates(map[string]interface{}{
		"record_date": updated.RecordDate,
		"record_type": updated.RecordType,
		"title":       updated.Title,
		"content":     updated.Content,
		"stats_json":  updated.StatsJSON,
	}).Error; err != nil {
		utils.ServerError(c, "更新失败")
		return
	}

	if err := ctrl.db.First(&existing, existing.ID).Error; err != nil {
		utils.ServerError(c, "查询失败")
		return
	}
	utils.Success(c, "成长记录已更新", growthRecordToDTO(existing))
}

func (ctrl *UserController) DeleteGrowthRecord(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	recordID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的成长记录ID")
		return
	}

	result := ctrl.db.Where("id = ? AND user_id = ?", recordID, userID).Delete(&models.GrowthRecord{})
	if result.Error != nil {
		utils.ServerError(c, "删除失败")
		return
	}
	if result.RowsAffected == 0 {
		utils.Error(c, http.StatusNotFound, "成长记录不存在")
		return
	}

	utils.Success(c, "成长记录已删除", nil)
}

// ========== 账号设置相关 API ==========

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ChangePassword 修改密码
func (ctrl *UserController) ChangePassword(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := ctrl.authService.GetUserByID(userID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取用户信息失败")
		return
	}
	if user == nil {
		utils.Error(c, http.StatusNotFound, "用户不存在")
		return
	}

	// 验证原密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		utils.Error(c, http.StatusBadRequest, "原密码错误")
		return
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "密码加密失败")
		return
	}

	// 更新密码
	updates := map[string]interface{}{
		"password": string(hashedPassword),
	}
	if err := ctrl.db.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "修改密码失败")
		return
	}

	utils.Success(c, "密码修改成功", nil)
}

// ChangePhoneRequest 修改手机号请求
type ChangePhoneRequest struct {
	NewPhone string `json:"new_phone" binding:"required"`
	Code     string `json:"code" binding:"required,len=6"`
}

// ChangePhone 修改手机号
func (ctrl *UserController) ChangePhone(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	var req ChangePhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 验证新手机号是否已被注册
	existing, err := ctrl.authService.CheckPhoneExists(req.NewPhone)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "服务器错误")
		return
	}
	if existing {
		utils.Error(c, http.StatusBadRequest, "该手机号已被注册")
		return
	}

	// 验证验证码（使用 reset 类型，因为已注册用户的手机号变更需要验证）
	valid, err := ctrl.authService.VerifyCode(&services.VerifyCodeRequest{
		Phone: req.NewPhone,
		Code:  req.Code,
		Type:  models.SmsCodeTypeReset,
	})
	if err != nil || valid == nil || !valid.Valid {
		utils.Error(c, http.StatusBadRequest, "验证码无效或已过期")
		return
	}

	// 更新手机号
	updates := map[string]interface{}{
		"phone": req.NewPhone,
	}
	if err := ctrl.db.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "修改手机号失败")
		return
	}

	utils.Success(c, "手机号修改成功", gin.H{"phone": req.NewPhone})
}

// NotificationSettings 通知设置
type NotificationSettings struct {
	SystemAnnouncements bool `json:"system_announcements"`
	OrderStatus         bool `json:"order_status"`
	WeeklyReport        bool `json:"weekly_report"`
	SocialInteraction   bool `json:"social_interaction"`
	PrivateMessage      bool `json:"private_message"`
	EmailNotification   bool `json:"email_notification"`
}

// PrivacySettings 隐私设置
type PrivacySettings struct {
	ProfileVisible bool `json:"profile_visible"`
	PhoneVisible   bool `json:"phone_visible"`
	AllowSearch    bool `json:"allow_search"`
	ShowRealName   bool `json:"show_real_name"`
}

func defaultNotificationSettings() NotificationSettings {
	return NotificationSettings{
		SystemAnnouncements: true,
		OrderStatus:         true,
		WeeklyReport:        true,
		SocialInteraction:   true,
		PrivateMessage:      true,
		EmailNotification:   true,
	}
}

func defaultPrivacySettings() PrivacySettings {
	return PrivacySettings{
		ProfileVisible: true,
		PhoneVisible:   true,
		AllowSearch:    true,
		ShowRealName:   true,
	}
}

func parseNotificationSettings(raw string) NotificationSettings {
	notif := defaultNotificationSettings()
	if raw != "" {
		_ = json.Unmarshal([]byte(raw), &notif)
	}
	return notif
}

func parsePrivacySettings(raw string) PrivacySettings {
	privacy := defaultPrivacySettings()
	if raw != "" {
		_ = json.Unmarshal([]byte(raw), &privacy)
		var compat struct {
			ProfileVisible *bool `json:"profileVisible"`
			PhoneVisible   *bool `json:"phoneVisible"`
			AllowSearch    *bool `json:"allowSearch"`
			Searchable     *bool `json:"searchable"`
			ShowRealName   *bool `json:"showRealName"`
		}
		if err := json.Unmarshal([]byte(raw), &compat); err == nil {
			if compat.ProfileVisible != nil {
				privacy.ProfileVisible = *compat.ProfileVisible
			}
			if compat.PhoneVisible != nil {
				privacy.PhoneVisible = *compat.PhoneVisible
			}
			if compat.AllowSearch != nil {
				privacy.AllowSearch = *compat.AllowSearch
			}
			if compat.Searchable != nil {
				privacy.AllowSearch = *compat.Searchable
			}
			if compat.ShowRealName != nil {
				privacy.ShowRealName = *compat.ShowRealName
			}
		}
	}
	return privacy
}

// GetSettings 获取用户设置（通知+隐私）
func (ctrl *UserController) GetSettings(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	user, err := ctrl.authService.GetUserByID(userID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取用户信息失败")
		return
	}

	// 解析通知设置
	notif := parseNotificationSettings(user.NotificationSettings)

	// 解析隐私设置
	privacy := parsePrivacySettings(user.PrivacySettings)

	utils.Success(c, "", gin.H{
		"notification": notif,
		"privacy":      privacy,
	})
}

// UpdateSettingsRequest 更新设置请求
type UpdateSettingsRequest struct {
	Notification *NotificationSettings `json:"notification"`
	Privacy      *PrivacySettings      `json:"privacy"`
}

// UpdateSettings 更新用户设置
func (ctrl *UserController) UpdateSettings(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	updates := make(map[string]interface{})

	if req.Notification != nil {
		notifJSON, _ := json.Marshal(req.Notification)
		updates["notification_settings"] = string(notifJSON)
	}

	if req.Privacy != nil {
		privacyJSON, _ := json.Marshal(req.Privacy)
		updates["privacy_settings"] = string(privacyJSON)
	}

	if len(updates) == 0 {
		utils.Error(c, http.StatusBadRequest, "无更新内容")
		return
	}

	if err := ctrl.db.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "更新设置失败")
		return
	}

	utils.Success(c, "设置更新成功", nil)
}

// LoginDevice 登录设备信息
type LoginDevice struct {
	ID         uint   `json:"id"`
	DeviceName string `json:"device_name"`
	DeviceType string `json:"device_type"`
	Browser    string `json:"browser"`
	OS         string `json:"os"`
	IP         string `json:"ip"`
	Location   string `json:"location"`
	LoginAt    string `json:"login_at"`
	IsCurrent  bool   `json:"is_current"`
}

func loginLogDeviceType(log models.LoginLog) string {
	osName := log.OS
	if osName == "iOS" || osName == "Android" {
		return "mobile"
	}
	return "desktop"
}

// GetLoginDevices 获取登录设备列表
func (ctrl *UserController) GetLoginDevices(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	var logs []models.LoginLog
	if err := ctrl.db.Where("user_id = ? AND status = ?", userID, "success").
		Order("created_at DESC").
		Limit(20).
		Find(&logs).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取登录设备失败")
		return
	}

	devices := make([]LoginDevice, 0, len(logs))
	for i, logItem := range logs {
		deviceName := logItem.Device
		if deviceName == "" {
			deviceName = "未知设备"
		}
		devices = append(devices, LoginDevice{
			ID:         logItem.ID,
			DeviceName: deviceName,
			DeviceType: loginLogDeviceType(logItem),
			Browser:    logItem.Browser,
			OS:         logItem.OS,
			IP:         logItem.IP,
			Location:   logItem.Location,
			LoginAt:    logItem.CreatedAt.Format("2006-01-02 15:04"),
			IsCurrent:  i == 0,
		})
	}

	if len(devices) == 0 {
		devices = append(devices, LoginDevice{
			ID:         0,
			DeviceName: "当前设备",
			DeviceType: "desktop",
			Browser:    "Unknown",
			OS:         "Unknown",
			IP:         c.ClientIP(),
			Location:   "当前请求",
			LoginAt:    time.Now().Format("2006-01-02 15:04"),
			IsCurrent:  true,
		})
	}

	utils.Success(c, "", gin.H{"devices": devices})
}

// LogoutDevice 登出指定设备
func (ctrl *UserController) LogoutDevice(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	deviceID, err := strconv.ParseUint(c.Param("deviceId"), 10, 64)
	if err != nil || deviceID == 0 {
		utils.Error(c, http.StatusBadRequest, "无效的设备ID")
		return
	}

	var latest models.LoginLog
	if err := ctrl.db.Where("user_id = ? AND status = ?", userID, "success").
		Order("created_at DESC").
		First(&latest).Error; err == nil && latest.ID == uint(deviceID) {
		utils.Error(c, http.StatusBadRequest, "当前设备不能在此登出")
		return
	}

	result := ctrl.db.Where("id = ? AND user_id = ?", uint(deviceID), userID).Delete(&models.LoginLog{})
	if result.Error != nil {
		utils.Error(c, http.StatusInternalServerError, "移除设备记录失败")
		return
	}
	if result.RowsAffected == 0 {
		utils.Error(c, http.StatusNotFound, "设备记录不存在")
		return
	}

	utils.Success(c, "设备记录已移除", nil)
}
