package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/middleware"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/utils"
	"gorm.io/gorm"
)

// PlayerController 球员资料控制器
type PlayerController struct {
	db *gorm.DB
}

// NewPlayerController 创建球员资料控制器
func NewPlayerController(db *gorm.DB) *PlayerController {
	return &PlayerController{db: db}
}

// PlayerProfileResponse 球员完整资料响应
type PlayerProfileResponse struct {
	ID                  uint     `json:"id"`
	Nickname            string   `json:"nickname"`
	RealName            string   `json:"real_name"`
	BirthDate           string   `json:"birth_date"`
	Gender              string   `json:"gender"`
	Age                 int      `json:"age"`
	Avatar              string   `json:"avatar"`
	Position            string   `json:"position"`
	SecondPosition      string   `json:"second_position,omitempty"`
	DominantFoot        string   `json:"dominant_foot"`
	Height              float64  `json:"height,omitempty"`
	Weight              float64  `json:"weight,omitempty"`
	PlayingStyle        []string `json:"playing_style"`
	JerseyNumber        int      `json:"jersey_number,omitempty"`
	JerseyColor         string   `json:"jersey_color,omitempty"`
	CurrentTeam         string   `json:"current_team,omitempty"`
	StartYear           int      `json:"start_year,omitempty"`
	FARegistered        bool     `json:"fa_registered"`
	Association         string   `json:"association,omitempty"`
	Province            string   `json:"province"`
	City                string   `json:"city"`
	Wechat              string   `json:"wechat,omitempty"`
	School              string   `json:"school,omitempty"`
	FatherHeight        float64  `json:"father_height,omitempty"`
	FatherPhone         string   `json:"father_phone,omitempty"`
	FatherJob           string   `json:"father_job,omitempty"`
	FatherAthlete       bool     `json:"father_athlete"`
	MotherHeight        float64  `json:"mother_height,omitempty"`
	MotherPhone         string   `json:"mother_phone,omitempty"`
	MotherJob           string   `json:"mother_job,omitempty"`
	MotherAthlete       bool     `json:"mother_athlete"`
	TechnicalTags       []string `json:"technical_tags"`
	MentalTags          []string `json:"mental_tags"`
	Experiences         []ExperienceItem `json:"experiences"`
	LatestPhysicalTest  *PhysicalTestInfo `json:"latest_physical_test,omitempty"`
	ProfileCompleteness int      `json:"profile_completeness"`
}

// ExperienceItem 足球经历项
type ExperienceItem struct {
	ID         string `json:"id,omitempty"`
	Period     string `json:"period"`
	Team       string `json:"team"`
	Position   string `json:"position"`
	Achievement string `json:"achievement,omitempty"`
}

// PhysicalTestInfo 最新体测信息
type PhysicalTestInfo struct {
	TestDate         string  `json:"test_date"`
	Sprint30m        float64 `json:"sprint_30m,omitempty"`
	StandingLongJump float64 `json:"standing_long_jump,omitempty"`
	PushUp           int     `json:"push_up,omitempty"`
	SitAndReach      float64 `json:"sit_and_reach,omitempty"`
}

// GetProfile 获取球员个人完整资料
func (ctrl *PlayerController) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	user, err := models.NewUserRepository(ctrl.db).FindByID(userID)
	if err != nil || user == nil {
		utils.Error(c, http.StatusInternalServerError, "获取用户信息失败")
		return
	}

	// 获取最新体测记录
	var latestTest *PhysicalTestInfo
	latestRecord, err := ctrl.getLatestPhysicalTest(userID)
	if err == nil && latestRecord != nil {
		latestTest = &PhysicalTestInfo{
			TestDate:         latestRecord.TestDate.Format("2006-01-02"),
			Sprint30m:        ptrFloat(latestRecord.Sprint30m),
			StandingLongJump: ptrFloat(latestRecord.StandingLongJump),
			PushUp:           ptrInt(latestRecord.PushUp),
			SitAndReach:      ptrFloat(latestRecord.SitAndReach),
		}
	} else if user.Sprint30m > 0 || user.StandingLongJump > 0 || user.PushUp > 0 || user.SitAndReach > 0 {
		// 回退：注册时填写的体测数据存储在 users 表中
		latestTest = &PhysicalTestInfo{
			TestDate:         user.CreatedAt.Format("2006-01-02"),
			Sprint30m:        ptrFloat(&user.Sprint30m),
			StandingLongJump: ptrFloat(&user.StandingLongJump),
			PushUp:           ptrInt(&user.PushUp),
			SitAndReach:      ptrFloat(&user.SitAndReach),
		}
	}

	// 计算资料完成度
	completeness := ctrl.calculateCompleteness(user)

	// 解析 JSON 字段
	playingStyles := parseJSONArray(user.PlayingStyle)
	technicalTags := parseJSONArray(user.TechnicalTags)
	mentalTags := parseJSONArray(user.MentalTags)
	experiences := parseExperiences(user.Experiences)

	response := PlayerProfileResponse{
		ID:                  user.ID,
		Nickname:           user.Nickname,
		RealName:           user.Name,
		BirthDate:          user.BirthDate,
		Gender:             user.Gender,
		Age:                user.Age,
		Avatar:             user.Avatar,
		Position:           user.Position,
		SecondPosition:     user.SecondPosition,
		DominantFoot:       user.Foot,
		Height:             user.Height,
		Weight:             user.Weight,
		PlayingStyle:       playingStyles,
		JerseyNumber:       user.JerseyNumber,
		JerseyColor:        user.JerseyColor,
		CurrentTeam:        user.CurrentTeam,
		StartYear:          user.StartYear,
		FARegistered:       user.FARegistered,
		Association:        user.Association,
		Province:           user.Province,
		City:               user.City,
		Wechat:             user.Wechat,
		School:             user.School,
		FatherHeight:       user.FatherHeight,
		FatherPhone:        user.FatherPhone,
		FatherJob:          user.FatherJob,
		FatherAthlete:      user.FatherAthlete,
		MotherHeight:       user.MotherHeight,
		MotherPhone:        user.MotherPhone,
		MotherJob:          user.MotherJob,
		MotherAthlete:      user.MotherAthlete,
		TechnicalTags:      technicalTags,
		MentalTags:         mentalTags,
		Experiences:        experiences,
		LatestPhysicalTest: latestTest,
		ProfileCompleteness: completeness,
	}

	utils.Success(c, "", gin.H{"profile": response})
}

// UpdateProfile 更新球员个人资料
func (ctrl *PlayerController) UpdateProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	var req struct {
		Nickname       *string `json:"nickname"`
		Avatar         *string `json:"avatar"`
		Name           *string `json:"name"`
		BirthDate      *string `json:"birth_date"`
		Gender         *string `json:"gender"`
		Height         *float64 `json:"height"`
		Weight         *float64 `json:"weight"`
		Position       *string `json:"position"`
		SecondPosition *string `json:"second_position"`
		Foot           *string `json:"foot"`
		Province       *string `json:"province"`
		City           *string `json:"city"`
		CurrentTeam    *string `json:"current_team"`
		PlayingStyle   *string `json:"playing_style"`
		Wechat         *string `json:"wechat"`
		School         *string `json:"school"`
		TechnicalTags  *string `json:"technical_tags"`
		MentalTags     *string `json:"mental_tags"`
		Experiences    *string `json:"experiences"`
		StartYear      *int    `json:"start_year"`
		FARegistered   *bool   `json:"fa_registered"`
		Association    *string `json:"association"`
		JerseyNumber   *int    `json:"jersey_number"`
		JerseyColor    *string `json:"jersey_color"`
		FatherHeight   *float64 `json:"father_height"`
		FatherPhone    *string `json:"father_phone"`
		FatherJob      *string `json:"father_job"`
		FatherAthlete  *bool   `json:"father_athlete"`
		MotherHeight   *float64 `json:"mother_height"`
		MotherPhone    *string `json:"mother_phone"`
		MotherJob      *string `json:"mother_job"`
		MotherAthlete  *bool   `json:"mother_athlete"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 构建更新字段
	updates := make(map[string]interface{})
	addStringField(updates, "nickname", req.Nickname)
	addStringField(updates, "avatar", req.Avatar)
	addStringField(updates, "name", req.Name)
	addStringField(updates, "birth_date", req.BirthDate)
	addStringField(updates, "gender", req.Gender)
	addFloatField(updates, "height", req.Height)
	addFloatField(updates, "weight", req.Weight)
	addStringField(updates, "position", req.Position)
	addStringField(updates, "second_position", req.SecondPosition)
	addStringField(updates, "foot", req.Foot)
	addStringField(updates, "province", req.Province)
	addStringField(updates, "city", req.City)
	addStringField(updates, "current_team", req.CurrentTeam)
	addStringField(updates, "playing_style", req.PlayingStyle)
	addStringField(updates, "wechat", req.Wechat)
	addStringField(updates, "school", req.School)
	addStringField(updates, "technical_tags", req.TechnicalTags)
	addStringField(updates, "mental_tags", req.MentalTags)
	addStringField(updates, "experiences", req.Experiences)
	addStringField(updates, "association", req.Association)
	addStringField(updates, "jersey_color", req.JerseyColor)
	addStringField(updates, "father_phone", req.FatherPhone)
	addStringField(updates, "father_job", req.FatherJob)
	addStringField(updates, "mother_phone", req.MotherPhone)
	addStringField(updates, "mother_job", req.MotherJob)

	if req.StartYear != nil {
		updates["start_year"] = *req.StartYear
	}
	if req.FARegistered != nil {
		updates["fa_registered"] = *req.FARegistered
	}
	if req.JerseyNumber != nil {
		updates["jersey_number"] = *req.JerseyNumber
	}
	if req.FatherHeight != nil {
		updates["father_height"] = *req.FatherHeight
	}
	if req.FatherAthlete != nil {
		updates["father_athlete"] = *req.FatherAthlete
	}
	if req.MotherHeight != nil {
		updates["mother_height"] = *req.MotherHeight
	}
	if req.MotherAthlete != nil {
		updates["mother_athlete"] = *req.MotherAthlete
	}

	if len(updates) == 0 {
		utils.Error(c, http.StatusBadRequest, "没有需要更新的字段")
		return
	}

	// 执行更新
	userRepo := models.NewUserRepository(ctrl.db)
	if err := userRepo.Update(userID, updates); err != nil {
		utils.Error(c, http.StatusInternalServerError, "更新失败")
		return
	}

	// 返回更新后的资料
	ctrl.GetProfile(c)
}

// PatchProfile 部分更新球员资料（单一字段快速更新）
func (ctrl *PlayerController) PatchProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.Error(c, http.StatusBadRequest, "请求体格式错误")
		return
	}

	if len(body) == 0 {
		utils.Error(c, http.StatusBadRequest, "没有需要更新的字段")
		return
	}

	// 白名单字段映射（前端 camelCase → 数据库 snake_case）
	fieldWhitelist := map[string]string{
		"nickname":       "nickname",
		"realName":       "name",
		"birthDate":      "birth_date",
		"gender":         "gender",
		"height":         "height",
		"weight":         "weight",
		"position":       "position",
		"secondPosition": "second_position",
		"foot":           "foot",
		"province":       "province",
		"city":           "city",
		"currentTeam":    "current_team",
		"wechat":         "wechat",
		"school":         "school",
		"jerseyNumber":   "jersey_number",
		"jerseyColor":    "jersey_color",
		"startYear":      "start_year",
		"faRegistered":   "fa_registered",
		"association":    "association",
		"playingStyle":   "playing_style",
		"technicalTags":  "technical_tags",
		"mentalTags":     "mental_tags",
		"experiences":    "experiences",
		"fatherHeight":   "father_height",
		"fatherPhone":    "father_phone",
		"fatherJob":      "father_job",
		"fatherAthlete":  "father_athlete",
		"motherHeight":   "mother_height",
		"motherPhone":    "mother_phone",
		"motherJob":      "mother_job",
		"motherAthlete":  "mother_athlete",
	}

	updates := make(map[string]interface{})
	for camelKey, val := range body {
		dbKey, ok := fieldWhitelist[camelKey]
		if !ok {
			continue // 忽略不在白名单的字段
		}
		updates[dbKey] = val
	}

	if len(updates) == 0 {
		utils.Error(c, http.StatusBadRequest, "没有可更新的有效字段")
		return
	}

	userRepo := models.NewUserRepository(ctrl.db)
	if err := userRepo.Update(userID, updates); err != nil {
		utils.Error(c, http.StatusInternalServerError, "更新失败")
		return
	}

	utils.Success(c, "更新成功", nil)
}

// GetPhysicalTests 获取体测记录列表（含来源区分）
func (ctrl *PlayerController) GetPhysicalTests(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	var records []models.PhysicalTestRecord
	result := ctrl.db.Where("player_id = ?", userID).
		Order("test_date DESC, created_at DESC").
		Find(&records)
	if result.Error != nil {
		utils.Error(c, http.StatusInternalServerError, "获取体测记录失败")
		return
	}

	// 预加载俱乐部名称（避免 N+1 查询）
	activityIDs := make([]uint, 0)
	clubIDs := make(map[uint]bool)
	for _, r := range records {
		if r.ActivityID > 0 && r.ClubID > 0 {
			activityIDs = append(activityIDs, r.ActivityID)
			clubIDs[r.ClubID] = true
		}
	}

	// 批量查俱乐部名称
	clubNames := make(map[uint]string) // clubID -> name
	if len(clubIDs) > 0 {
		var cids []uint
		for cid := range clubIDs {
			cids = append(cids, cid)
		}
		var clubs []models.Club
		ctrl.db.Select("id, name").Where("id IN ?", cids).Find(&clubs)
		for _, cl := range clubs {
			clubNames[cl.ID] = cl.Name
		}
	}

	// 转换为响应格式（含来源区分字段）
	list := make([]gin.H, len(records))
	for i, r := range records {
		// 判定数据来源: 有活动ID且关联俱乐部 → 俱乐部体测，否则个人自测
		source := "personal"
		var clubName *string
		if r.ActivityID > 0 && r.ClubID > 0 {
			source = "club"
			if name, ok := clubNames[r.ClubID]; ok {
				clubName = &name
			}
		}

		// 判定录入者角色
		recorderRole := "player" // 默认球员自己
		if r.RecorderID != 0 && r.RecorderID != userID {
			recorderRole = "coach"
		}
		if source == "club" && r.RecorderID != userID {
			recorderRole = "coach"
		}

		list[i] = gin.H{
			"id":                 r.ID,
			"test_date":          r.TestDate.Format("2006-01-02"),
			"height":             r.Height,
			"weight":             r.Weight,
			"bmi":                r.BMI,
			"sprint_30m":         r.Sprint30m,
			"sprint_50m":         r.Sprint50m,
			"sprint_100m":        r.Sprint100m,
			"agility_ladder":     r.AgilityLadder,
			"t_test":             r.TTest,
			"shuttle_run":        r.ShuttleRun,
			"standing_long_jump": r.StandingLongJump,
			"vertical_jump":      r.VerticalJump,
			"sit_and_reach":      r.SitAndReach,
			"push_up":            r.PushUp,
			"sit_up":             r.SitUp,
			"plank":              r.Plank,
			"extra_data":         r.ExtraData,
			// 来源区分字段（新增）
			"source":            source,       // "personal" | "club"
			"club_name":         clubName,     // 个人为 nil，俱乐部为俱乐部名
			"activity_id":       func() *uint { v := r.ActivityID; if v == 0 { return nil }; return &v }(),
			"recorder_role":     recorderRole, // "player" | "coach"
			"created_at":        r.CreatedAt.Format("2006-01-02 15:04"),
		}
	}

	utils.Success(c, "", gin.H{
		"records":   list,
		"total":     len(list),
		"page":      1,
		"page_size": 20,
	})
}

// CreatePhysicalTest 创建体测记录（写入 physical_test_records 独立表）
func (ctrl *PlayerController) CreatePhysicalTest(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	var req struct {
		TestDate         string   `json:"test_date"`
		Sprint30m        *float64 `json:"sprint_30m"`
		Sprint50m        *float64 `json:"sprint_50m"`
		Sprint100m       *float64 `json:"sprint_100m"`
		StandingLongJump *float64 `json:"standing_long_jump"`
		PushUp           *int     `json:"push_up"`
		SitAndReach      *float64 `json:"sit_and_reach"`
		Height           *float64 `json:"height"`
		Weight           *float64 `json:"weight"`
		AgilityLadder    *float64 `json:"agility_ladder"`
		TTest            *float64 `json:"t_test"`
		ShuttleRun       *float64 `json:"shuttle_run"`
		VerticalJump     *float64 `json:"vertical_jump"`
		SitUp            *int     `json:"sit_up"`
		Plank            *int     `json:"plank"`
		ExtraData        string   `json:"extra_data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 解析测试日期，默认今天
	testDate := time.Now()
	if req.TestDate != "" {
		if parsed, err := parseDate(req.TestDate); err == nil {
			testDate = parsed
		}
	}

	// 自动计算 BMI
	var bmi *float64
	if req.Height != nil && req.Weight != nil && *req.Height > 0 {
		m := *req.Height / 100.0
		val := *req.Weight / (m * m)
		bmi = &val
	}

	record := models.PhysicalTestRecord{
		PlayerID:          userID,
		TestDate:          testDate,
		Sprint30m:         req.Sprint30m,
		Sprint50m:         req.Sprint50m,
		Sprint100m:        req.Sprint100m,
		StandingLongJump:  req.StandingLongJump,
		PushUp:            req.PushUp,
		SitAndReach:       req.SitAndReach,
		Height:            req.Height,
		Weight:            req.Weight,
		BMI:               bmi,
		AgilityLadder:     req.AgilityLadder,
		TTest:             req.TTest,
		ShuttleRun:        req.ShuttleRun,
		VerticalJump:      req.VerticalJump,
		SitUp:             req.SitUp,
		Plank:             req.Plank,
		ExtraData:         req.ExtraData,
		RecorderID:        userID, // 球员自己录入
	}

	if err := ctrl.db.Create(&record).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "保存体测数据失败")
		return
	}

	utils.Success(c, "体测数据已保存", gin.H{"id": record.ID})
}

// UpdatePhysicalTest 编辑体测记录
func (ctrl *PlayerController) UpdatePhysicalTest(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	// 获取记录ID并校验归属
	idStr := c.Param("id")
	var id uint
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的记录ID")
		return
	}

	var existing models.PhysicalTestRecord
	if err := ctrl.db.Where("id = ? AND player_id = ?", id, userID).First(&existing).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.Error(c, http.StatusNotFound, "体测记录不存在")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}

	var req struct {
		TestDate         string   `json:"test_date"`
		Sprint30m        *float64 `json:"sprint_30m"`
		Sprint50m        *float64 `json:"sprint_50m"`
		Sprint100m       *float64 `json:"sprint_100m"`
		StandingLongJump *float64 `json:"standing_long_jump"`
		PushUp           *int     `json:"push_up"`
		SitAndReach      *float64 `json:"sit_and_reach"`
		Height           *float64 `json:"height"`
		Weight           *float64 `json:"weight"`
		AgilityLadder    *float64 `json:"agility_ladder"`
		TTest            *float64 `json:"t_test"`
		ShuttleRun       *float64 `json:"shuttle_run"`
		VerticalJump     *float64 `json:"vertical_jump"`
		SitUp            *int     `json:"sit_up"`
		Plank            *int     `json:"plank"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	updates := make(map[string]interface{})
	if req.TestDate != "" {
		if parsed, err := parseDate(req.TestDate); err == nil {
			updates["test_date"] = parsed
		}
	}
	addPtrFloat(updates, "sprint30m", req.Sprint30m)
	addPtrFloat(updates, "sprint50m", req.Sprint50m)
	addPtrFloat(updates, "sprint100m", req.Sprint100m)
	addPtrFloat(updates, "standing_long_jump", req.StandingLongJump)
	addPtrInt(updates, "push_up", req.PushUp)
	addPtrFloat(updates, "sit_and_reach", req.SitAndReach)
	addPtrFloat(updates, "height", req.Height)
	addPtrFloat(updates, "weight", req.Weight)
	addPtrFloat(updates, "bmi", calculateBMI(req.Height, req.Weight))
	addPtrFloat(updates, "agility_ladder", req.AgilityLadder)
	addPtrFloat(updates, "t_test", req.TTest)
	addPtrFloat(updates, "shuttle_run", req.ShuttleRun)
	addPtrFloat(updates, "vertical_jump", req.VerticalJump)
	addPtrInt(updates, "sit_up", req.SitUp)
	addPtrInt(updates, "plank", req.Plank)

	if len(updates) == 0 {
		utils.Error(c, http.StatusBadRequest, "没有需要更新的字段")
		return
	}

	if err := ctrl.db.Model(&existing).Updates(updates).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "更新失败")
		return
	}

	utils.Success(c, "体测数据已更新", nil)
}

// DeletePhysicalTest 删除体测记录（软删除）
func (ctrl *PlayerController) DeletePhysicalTest(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	idStr := c.Param("id")
	var id uint
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的记录ID")
		return
	}

	result := ctrl.db.Where("id = ? AND player_id = ?", id, userID).Delete(&models.PhysicalTestRecord{})
	if result.Error != nil {
		utils.Error(c, http.StatusInternalServerError, "删除失败")
		return
	}
	if result.RowsAffected == 0 {
		utils.Error(c, http.StatusNotFound, "体测记录不存在")
		return
	}

	utils.Success(c, "体测记录已删除", nil)
}

// GetPlayerPublicProfile 获取球员公开资料（无需登录）
func (ctrl *PlayerController) GetPlayerPublicProfile(c *gin.Context) {
	idStr := c.Param("playerId")
	var userID uint
	if _, err := fmt.Sscanf(idStr, "%d", &userID); err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的用户ID")
		return
	}

	user, err := models.NewUserRepository(ctrl.db).FindByID(userID)
	if err != nil || user == nil {
		utils.Error(c, http.StatusNotFound, "球员不存在")
		return
	}

	// 获取最新体测（简要）
	var latestTest *PhysicalTestInfo
	record, err := ctrl.getLatestPhysicalTest(userID)
	if err == nil && record != nil {
		latestTest = &PhysicalTestInfo{
			TestDate:         record.TestDate.Format("2006-01-02"),
			Sprint30m:        ptrFloat(record.Sprint30m),
			StandingLongJump: ptrFloat(record.StandingLongJump),
			PushUp:           ptrInt(record.PushUp),
			SitAndReach:      ptrFloat(record.SitAndReach),
		}
	}

	response := gin.H{
		"id":              user.ID,
		"nickname":        user.Nickname,
		"real_name":       user.Name,
		"avatar":          user.Avatar,
		"age":             user.Age,
		"gender":          user.Gender,
		"position":        user.Position,
		"second_position": user.SecondPosition,
		"dominant_foot":   user.Foot,
		"height":          user.Height,
		"weight":          user.Weight,
		"playing_style":   parseJSONArray(user.PlayingStyle),
		"current_team":    user.CurrentTeam,
		"start_year":      user.StartYear,
		"jersey_number":   user.JerseyNumber,
		"jersey_color":    user.JerseyColor,
		"province":        user.Province,
		"city":            user.City,
		"school":          user.School,
		"fa_registered":   user.FARegistered,
		"association":     user.Association,
		"technical_tags":  parseJSONArray(user.TechnicalTags),
		"mental_tags":     parseJSONArray(user.MentalTags),
		"experiences":     parseExperiences(user.Experiences),
		"latest_physical_test": latestTest,
		"created_at":      user.CreatedAt,
	}

	utils.Success(c, "", gin.H{"player": response})
}

// GetPublicPhysicalTests 获取球员公开体测记录（无需登录，供主页展示用）
func (ctrl *PlayerController) GetPublicPhysicalTests(c *gin.Context) {
	idStr := c.Param("playerId")
	var userID uint
	if _, err := fmt.Sscanf(idStr, "%d", &userID); err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的用户ID")
		return
	}

	// 验证用户存在
	var user models.User
	if err := ctrl.db.Select("id").Where("id = ? AND role IN (?) AND status = ?", userID, []string{"user", "player"}, "active").First(&user).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "球员不存在")
		return
	}

	var records []models.PhysicalTestRecord
	result := ctrl.db.Where("player_id = ?", userID).
		Order("test_date DESC, created_at DESC").
		Limit(20). // 公开接口限制返回条数
		Find(&records)
	if result.Error != nil {
		utils.Error(c, http.StatusInternalServerError, "获取体测数据失败")
		return
	}

	// 批量查俱乐部名称（复用逻辑）
	clubIDs := make(map[uint]bool)
	for _, r := range records {
		if r.ClubID > 0 {
			clubIDs[r.ClubID] = true
		}
	}
	clubNames := make(map[uint]string)
	if len(clubIDs) > 0 {
		var cids []uint
		for cid := range clubIDs {
			cids = append(cids, cid)
		}
		var clubs []models.Club
		ctrl.db.Select("id, name").Where("id IN ?", cids).Find(&clubs)
		for _, cl := range clubs {
			clubNames[cl.ID] = cl.Name
		}
	}

	// 转换为响应格式 + 来源区分
	list := make([]gin.H, len(records))
	for i, r := range records {
		source := "personal"
		var clubName *string
		if r.ActivityID > 0 && r.ClubID > 0 {
			source = "club"
			if name, ok := clubNames[r.ClubID]; ok {
				clubName = &name
			}
		}
		recorderRole := "player"
		if r.RecorderID != 0 && r.RecorderID != userID {
			recorderRole = "coach"
		}
		if source == "club" && r.RecorderID != userID {
			recorderRole = "coach"
		}

		list[i] = gin.H{
			"id":                 r.ID,
			"test_date":          r.TestDate.Format("2006-01-02"),
			"height":             r.Height,
			"weight":             r.Weight,
			"bmi":                r.BMI,
			"sprint_30m":         r.Sprint30m,
			"sprint_50m":         r.Sprint50m,
			"sprint_100m":        r.Sprint100m,
			"agility_ladder":     r.AgilityLadder,
			"t_test":             r.TTest,
			"shuttle_run":        r.ShuttleRun,
			"standing_long_jump": r.StandingLongJump,
			"vertical_jump":      r.VerticalJump,
			"sit_and_reach":      r.SitAndReach,
			"push_up":            r.PushUp,
			"sit_up":             r.SitUp,
			"plank":              r.Plank,
			"source":             source,
			"club_name":          clubName,
			"activity_id":        func() *uint { v := r.ActivityID; if v == 0 { return nil }; return &v }(),
			"recorder_role":      recorderRole,
			"created_at":         r.CreatedAt.Format("2006-01-02 15:04"),
		}
	}

	utils.Success(c, "", gin.H{
		"records": list,
		"total":   len(list),
	})
}

// ============ 辅助函数 ============

func (ctrl *PlayerController) getLatestPhysicalTest(userID uint) (*models.PhysicalTestRecord, error) {
	var record models.PhysicalTestRecord
	err := ctrl.db.Where("player_id = ?", userID).
		Order("test_date DESC, created_at DESC").
		First(&record).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &record, err
}

func (ctrl *PlayerController) calculateCompleteness(user *models.User) int {
	fields := 0
	total := 20

	if user.Nickname != "" {
		fields++
	}
	if user.Name != "" {
		fields++
	}
	if user.BirthDate != "" {
		fields++
	}
	if user.Gender != "" {
		fields++
	}
	if user.Province != "" && user.City != "" {
		fields++
	}
	if user.Position != "" {
		fields++
	}
	if user.Height > 0 {
		fields++
	}
	if user.Weight > 0 {
		fields++
	}
	if user.Foot != "" {
		fields++
	}
	if user.PlayingStyle != "" {
		fields++
	}
	if user.CurrentTeam != "" {
		fields++
	}
	if user.StartYear > 0 {
		fields++
	}
	if user.Wechat != "" {
		fields++
	}
	if user.School != "" {
		fields++
	}
	if user.FatherHeight > 0 || user.MotherHeight > 0 {
		fields++
	}
	if user.TechnicalTags != "" {
		fields++
	}
	if user.MentalTags != "" {
		fields++
	}
	if user.Experiences != "" {
		fields++
	}
	if user.Avatar != "" {
		fields++
	}

	return int(float64(fields) / float64(total) * 100)
}

func parseJSONArray(s string) []string {
	if s == "" {
		return []string{}
	}
	var arr []string
	json.Unmarshal([]byte(s), &arr)
	return arr
}

func parseExperiences(s string) []ExperienceItem {
	if s == "" {
		return []ExperienceItem{}
	}
	var items []ExperienceItem
	json.Unmarshal([]byte(s), &items)
	return items
}

func parseDate(dateStr string) (time.Time, error) {
	formats := []string{"2006-01-02", "2006/01/02", "2006-01-02T15:04:05Z"}
	for _, format := range formats {
		if t, err := parseDateFormat(dateStr, format); err == nil {
			return t, nil
		}
	}
	return parseDateFormat(dateStr, "2006-01-02")
}

func parseDateFormat(dateStr, format string) (time.Time, error) {
	// 简化实现
	return time.Parse("2006-01-02", dateStr)
}

func addStringField(m map[string]interface{}, key string, val *string) {
	if val != nil {
		m[key] = *val
	}
}

func addFloatField(m map[string]interface{}, key string, val *float64) {
	if val != nil {
		m[key] = *val
	}
}

func addIntField(m map[string]interface{}, key string, val *int) {
	if val != nil {
		m[key] = *val
	}
}

func ptrFloat(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}

func ptrInt(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

func addPtrFloat(m map[string]interface{}, key string, val *float64) {
	if val != nil {
		m[key] = *val
	}
}

func addPtrInt(m map[string]interface{}, key string, val *int) {
	if val != nil {
		m[key] = *val
	}
}

func calculateBMI(height, weight *float64) *float64 {
	if height == nil || weight == nil || *height <= 0 || *weight <= 0 {
		return nil
	}
	m := *height / 100.0
	val := *weight / (m * m)
	return &val
}