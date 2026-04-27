package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/config"
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
	// userID := c.Param("userId")
	utils.Success(c, "", gin.H{"reports": []interface{}{}})
}

// GrowthRecord 成长记录
type GrowthRecord struct {
	ID        uint    `json:"id"`
	Date      string  `json:"date"`
	Height    float64 `json:"height"`
	Weight    float64 `json:"weight"`
	Speed     float64 `json:"speed"`
	Strength  float64 `json:"strength"`
	Technique float64 `json:"technique"`
	Tactics   float64 `json:"tactics"`
	Note      string  `json:"note"`
}

// GetGrowthRecords 获取成长记录（体测数据）
func (ctrl *UserController) GetGrowthRecords(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	var dbRecords []models.GrowthRecord
	if err := ctrl.db.Where("user_id = ? AND record_type = ?", userID, models.GrowthRecordTypePhysical).
		Order("record_date DESC").Find(&dbRecords).Error; err != nil {
		utils.ServerError(c, "查询失败")
		return
	}

	records := make([]GrowthRecord, 0, len(dbRecords))
	for _, r := range dbRecords {
		var stats struct {
			Height    float64 `json:"height"`
			Weight    float64 `json:"weight"`
			Speed     float64 `json:"speed"`
			Strength  float64 `json:"strength"`
			Technique float64 `json:"technique"`
			Tactics   float64 `json:"tactics"`
			Note      string  `json:"note"`
		}
		if r.StatsJSON != "" {
			_ = json.Unmarshal([]byte(r.StatsJSON), &stats)
		}
		records = append(records, GrowthRecord{
			ID:        r.ID,
			Date:      r.RecordDate.Format("2006-01-02"),
			Height:    stats.Height,
			Weight:    stats.Weight,
			Speed:     stats.Speed,
			Strength:  stats.Strength,
			Technique: stats.Technique,
			Tactics:   stats.Tactics,
			Note:      stats.Note,
		})
	}

	utils.Success(c, "", gin.H{"records": records})
}

// SaveGrowthRecords 保存成长记录（体测数据）
func (ctrl *UserController) SaveGrowthRecords(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	var req struct {
		Records []GrowthRecord `json:"records"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	for _, rec := range req.Records {
		stats := map[string]interface{}{
			"height":    rec.Height,
			"weight":    rec.Weight,
			"speed":     rec.Speed,
			"strength":  rec.Strength,
			"technique": rec.Technique,
			"tactics":   rec.Tactics,
			"note":      rec.Note,
		}
		statsJSON, _ := json.Marshal(stats)

		recordDate, _ := time.Parse("2006-01-02", rec.Date)
		if recordDate.IsZero() {
			recordDate = time.Now()
		}

		gr := models.GrowthRecord{
			UserID:     userID,
			RecordDate: recordDate,
			RecordType: models.GrowthRecordTypePhysical,
			Title:      "体测记录",
			Content:    rec.Note,
			StatsJSON:  string(statsJSON),
		}
		if err := ctrl.db.Create(&gr).Error; err != nil {
			utils.ServerError(c, "保存失败")
			return
		}
	}

	utils.Success(c, "成长记录保存成功", nil)
}

// VideoAnalysisResult 视频分析结果
type VideoAnalysisResult struct {
	ID           uint                   `json:"id"`
	VideoURL     string                 `json:"videoUrl"`
	Status       string                 `json:"status"`
	Progress     int                    `json:"progress"`
	AnalysisData map[string]interface{} `json:"analysisData"`
	Suggestions  []string               `json:"suggestions"`
}

// UploadAndAnalyzeVideo 上传并分析视频
func (ctrl *UserController) UploadAndAnalyzeVideo(c *gin.Context) {
	// 返回模拟分析结果
	result := VideoAnalysisResult{
		ID:       1,
		VideoURL: "https://example.com/video.mp4",
		Status:   "analyzing",
		Progress: 50,
		AnalysisData: map[string]interface{}{
			"speed":     85,
			"strength":  78,
			"technique": 82,
		},
		Suggestions: []string{
			"建议加强射门练习",
			"体能储备需要提升",
		},
	}

	utils.Success(c, "视频分析已启动", gin.H{"result": result})
}

// GetScoutMapData 获取球探地图数据（从数据库查询）
func (ctrl *UserController) GetScoutMapData(c *gin.Context) {
	// 从数据库查询所有球员用户
	var users []models.User
	if err := config.GetDB().Where("role = ?", "user").Find(&users).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询球员数据失败")
		return
	}

	// 组织数据
	var players []gin.H
	for _, user := range users {
		players = append(players, gin.H{
			"id":       user.ID,
			"name":     user.Name,
			"nickname": user.Nickname,
			"province": user.Province,
			"city":     user.City,
			"position": user.Position,
			"age":      user.Age,
			"height":   user.Height,
			"weight":   user.Weight,
			"foot":     user.Foot,
			"club":     user.Club,
			"avatar":   user.Avatar,
		})
	}

	// 提取省份列表
	provinceMap := make(map[string]bool)
	for _, user := range users {
		if user.Province != "" {
			provinceMap[user.Province] = true
		}
	}
	var provinces []string
	for p := range provinceMap {
		provinces = append(provinces, p)
	}

	utils.Success(c, "", gin.H{
		"players":   players,
		"total":     len(players),
		"provinces": provinces,
	})
}

// GetPlayersByProvince 按省份获取球员
func (ctrl *UserController) GetPlayersByProvince(c *gin.Context) {
	// 返回按省份分组的数据
	data := map[string][]gin.H{
		"广东": {
			{"id": 1, "name": "张三", "city": "广州", "position": "前锋", "age": 15},
			{"id": 2, "name": "李四", "city": "深圳", "position": "中场", "age": 16},
		},
		"北京": {
			{"id": 3, "name": "王五", "city": "北京", "position": "后卫", "age": 14},
		},
	}

	utils.Success(c, "", gin.H{"data": data})
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

// GetLoginDevices 获取登录设备列表（模拟数据）
func (ctrl *UserController) GetLoginDevices(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	// 返回模拟设备数据（实际应从 login_logs 表查询）
	devices := []LoginDevice{
		{
			ID:         1,
			DeviceName: "当前设备",
			DeviceType: "desktop",
			Browser:    "Chrome",
			OS:         "macOS",
			IP:         "192.168.1.xxx",
			Location:   "本地网络",
			LoginAt:    time.Now().Format("2006-01-02 15:04"),
			IsCurrent:  true,
		},
		{
			ID:         2,
			DeviceName: "iPhone 15 Pro",
			DeviceType: "mobile",
			Browser:    "Safari",
			OS:         "iOS 17",
			IP:         "117.136.xxx.xxx",
			Location:   "北京市",
			LoginAt:    time.Now().Add(-24 * time.Hour).Format("2006-01-02 15:04"),
			IsCurrent:  false,
		},
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

	deviceID := c.Param("deviceId")
	// 实际应删除对应设备的登录记录/token
	_ = deviceID

	utils.Success(c, "设备已登出", nil)
}
