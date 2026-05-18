package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
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
	ID                  uint              `json:"id"`
	Nickname            string            `json:"nickname"`
	RealName            string            `json:"real_name"`
	BirthDate           string            `json:"birth_date"`
	Gender              string            `json:"gender"`
	Age                 int               `json:"age"`
	Avatar              string            `json:"avatar"`
	Position            string            `json:"position"`
	SecondPosition      string            `json:"second_position,omitempty"`
	DominantFoot        string            `json:"dominant_foot"`
	Height              float64           `json:"height,omitempty"`
	Weight              float64           `json:"weight,omitempty"`
	PlayingStyle        []string          `json:"playing_style"`
	JerseyNumber        int               `json:"jersey_number,omitempty"`
	JerseyColor         string            `json:"jersey_color,omitempty"`
	CurrentTeam         string            `json:"current_team,omitempty"`
	StartYear           int               `json:"start_year,omitempty"`
	FARegistered        bool              `json:"fa_registered"`
	Association         string            `json:"association,omitempty"`
	Province            string            `json:"province"`
	City                string            `json:"city"`
	Wechat              string            `json:"wechat,omitempty"`
	School              string            `json:"school,omitempty"`
	FatherHeight        float64           `json:"father_height,omitempty"`
	FatherPhone         string            `json:"father_phone,omitempty"`
	FatherJob           string            `json:"father_job,omitempty"`
	FatherAthlete       bool              `json:"father_athlete"`
	MotherHeight        float64           `json:"mother_height,omitempty"`
	MotherPhone         string            `json:"mother_phone,omitempty"`
	MotherJob           string            `json:"mother_job,omitempty"`
	MotherAthlete       bool              `json:"mother_athlete"`
	TechnicalTags       []string          `json:"technical_tags"`
	MentalTags          []string          `json:"mental_tags"`
	Experiences         []ExperienceItem  `json:"experiences"`
	LatestPhysicalTest  *PhysicalTestInfo `json:"latest_physical_test,omitempty"`
	ProfileCompleteness int               `json:"profile_completeness"`
}

// ExperienceItem 足球经历项
type ExperienceItem struct {
	ID          string `json:"id,omitempty"`
	Period      string `json:"period"`
	Team        string `json:"team"`
	Position    string `json:"position"`
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
	age := effectivePlayerAge(user.Age, user.BirthDate)

	response := PlayerProfileResponse{
		ID:                  user.ID,
		Nickname:            user.Nickname,
		RealName:            user.Name,
		BirthDate:           user.BirthDate,
		Gender:              user.Gender,
		Age:                 age,
		Avatar:              user.Avatar,
		Position:            user.Position,
		SecondPosition:      user.SecondPosition,
		DominantFoot:        user.Foot,
		Height:              user.Height,
		Weight:              user.Weight,
		PlayingStyle:        playingStyles,
		JerseyNumber:        user.JerseyNumber,
		JerseyColor:         user.JerseyColor,
		CurrentTeam:         user.CurrentTeam,
		StartYear:           user.StartYear,
		FARegistered:        user.FARegistered,
		Association:         user.Association,
		Province:            user.Province,
		City:                user.City,
		Wechat:              user.Wechat,
		School:              user.School,
		FatherHeight:        user.FatherHeight,
		FatherPhone:         user.FatherPhone,
		FatherJob:           user.FatherJob,
		FatherAthlete:       user.FatherAthlete,
		MotherHeight:        user.MotherHeight,
		MotherPhone:         user.MotherPhone,
		MotherJob:           user.MotherJob,
		MotherAthlete:       user.MotherAthlete,
		TechnicalTags:       technicalTags,
		MentalTags:          mentalTags,
		Experiences:         experiences,
		LatestPhysicalTest:  latestTest,
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
		Nickname       *string  `json:"nickname"`
		Avatar         *string  `json:"avatar"`
		Name           *string  `json:"name"`
		BirthDate      *string  `json:"birth_date"`
		Gender         *string  `json:"gender"`
		Height         *float64 `json:"height"`
		Weight         *float64 `json:"weight"`
		Position       *string  `json:"position"`
		SecondPosition *string  `json:"second_position"`
		Foot           *string  `json:"foot"`
		Province       *string  `json:"province"`
		City           *string  `json:"city"`
		CurrentTeam    *string  `json:"current_team"`
		PlayingStyle   *string  `json:"playing_style"`
		Wechat         *string  `json:"wechat"`
		School         *string  `json:"school"`
		TechnicalTags  *string  `json:"technical_tags"`
		MentalTags     *string  `json:"mental_tags"`
		Experiences    *string  `json:"experiences"`
		StartYear      *int     `json:"start_year"`
		FARegistered   *bool    `json:"fa_registered"`
		Association    *string  `json:"association"`
		JerseyNumber   *int     `json:"jersey_number"`
		JerseyColor    *string  `json:"jersey_color"`
		FatherHeight   *float64 `json:"father_height"`
		FatherPhone    *string  `json:"father_phone"`
		FatherJob      *string  `json:"father_job"`
		FatherAthlete  *bool    `json:"father_athlete"`
		MotherHeight   *float64 `json:"mother_height"`
		MotherPhone    *string  `json:"mother_phone"`
		MotherJob      *string  `json:"mother_job"`
		MotherAthlete  *bool    `json:"mother_athlete"`
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
	if req.BirthDate != nil {
		updates["age"] = calculatePlayerAgeFromBirthDate(*req.BirthDate)
	}
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
			"source":    source,   // "personal" | "club"
			"club_name": clubName, // 个人为 nil，俱乐部为俱乐部名
			"activity_id": func() *uint {
				v := r.ActivityID
				if v == 0 {
					return nil
				}
				return &v
			}(),
			"recorder_role": recorderRole, // "player" | "coach"
			"created_at":    r.CreatedAt.Format("2006-01-02 15:04"),
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
		PlayerID:         userID,
		TestDate:         testDate,
		Sprint30m:        req.Sprint30m,
		Sprint50m:        req.Sprint50m,
		Sprint100m:       req.Sprint100m,
		StandingLongJump: req.StandingLongJump,
		PushUp:           req.PushUp,
		SitAndReach:      req.SitAndReach,
		Height:           req.Height,
		Weight:           req.Weight,
		BMI:              bmi,
		AgilityLadder:    req.AgilityLadder,
		TTest:            req.TTest,
		ShuttleRun:       req.ShuttleRun,
		VerticalJump:     req.VerticalJump,
		SitUp:            req.SitUp,
		Plank:            req.Plank,
		ExtraData:        req.ExtraData,
		RecorderID:       userID, // 球员自己录入
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

// GetTeamCalendar 获取球员所属球队的只读日历
func (ctrl *PlayerController) GetTeamCalendar(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	startDate, endDate := parseCalendarRange(c.Query("startDate"), c.Query("endDate"))
	typeSet := parseCalendarTypes(c.Query("types"))
	statusFilter := strings.TrimSpace(c.Query("status"))
	teamIDQuery := strings.TrimSpace(c.Query("teamId"))

	query := ctrl.db.Preload("Team.Club").
		Where("user_id = ? AND status = ?", userID, "active").
		Order("joined_at DESC, id DESC")
	if teamIDQuery != "" {
		teamID, err := strconv.ParseUint(teamIDQuery, 10, 32)
		if err != nil {
			utils.ValidationError(c, "无效的球队ID")
			return
		}
		query = query.Where("team_id = ?", uint(teamID))
	}

	var memberships []models.TeamPlayer
	if err := query.Find(&memberships).Error; err != nil {
		utils.ServerError(c, "获取球队日历失败")
		return
	}

	teams := make([]gin.H, 0, len(memberships))
	items := make([]gin.H, 0)
	stats := gin.H{
		"trainingCount":           0,
		"matchCount":              0,
		"physicalTestCount":       0,
		"weeklyPeriodCount":       0,
		"weeklyPendingCount":      0,
		"matchReviewPendingCount": 0,
	}

	teamCtrl := &TeamController{db: ctrl.db}
	for _, membership := range memberships {
		if membership.Team == nil {
			continue
		}
		team := membership.Team
		clubName := ""
		if team.Club != nil {
			clubName = team.Club.Name
		}
		teams = append(teams, gin.H{
			"id":           team.ID,
			"name":         team.Name,
			"ageGroup":     team.AgeGroup,
			"clubId":       team.ClubID,
			"clubName":     clubName,
			"jerseyNumber": membership.JerseyNumber,
			"position":     membership.Position,
		})

		payload, err := teamCtrl.buildTeamCalendarPayload(team.ID, team.ClubID, startDate, endDate, typeSet, statusFilter, "player")
		if err != nil {
			utils.ServerError(c, err.Error())
			return
		}
		for _, rawItem := range payload["items"].([]gin.H) {
			ctrl.enrichPlayerCalendarLinks(rawItem, userID)
			rawItem["teamName"] = team.Name
			rawItem["clubName"] = clubName
			items = append(items, rawItem)
		}
		if payloadStats, ok := payload["stats"].(gin.H); ok {
			accumulateCalendarStats(stats, payloadStats)
		}
	}

	sort.Slice(items, func(i, j int) bool {
		left, _ := time.Parse(time.RFC3339, fmt.Sprint(items[i]["startTime"]))
		right, _ := time.Parse(time.RFC3339, fmt.Sprint(items[j]["startTime"]))
		return left.Before(right)
	})

	var currentTeamID uint
	if len(teams) > 0 {
		currentTeamID, _ = teams[0]["id"].(uint)
	}

	utils.SuccessResponse(c, gin.H{
		"teamId": currentTeamID,
		"teams":  teams,
		"range": gin.H{
			"startDate": startDate.Format("2006-01-02"),
			"endDate":   endDate.AddDate(0, 0, -1).Format("2006-01-02"),
		},
		"items": items,
		"stats": stats,
	})
}

func (ctrl *PlayerController) enrichPlayerCalendarLinks(item gin.H, playerID uint) {
	links, ok := item["links"].(gin.H)
	if !ok {
		links = gin.H{}
		item["links"] = links
	}

	sourceID := calendarItemUint(item["sourceId"])
	switch fmt.Sprint(item["type"]) {
	case "weekly":
		links["weeklyPeriodId"] = sourceID
		if sourceID == 0 {
			return
		}
		var period models.WeeklyReportPeriod
		if err := ctrl.db.Select("id, team_id, week_start").
			First(&period, sourceID).Error; err != nil {
			links["weeklyReportId"] = nil
			return
		}
		var report models.WeeklyReport
		if err := ctrl.db.Select("id").
			Where("team_id = ? AND player_id = ? AND week_start = ?", period.TeamID, playerID, period.WeekStart).
			First(&report).Error; err != nil {
			links["weeklyReportId"] = nil
			return
		}
		links["weeklyReportId"] = report.ID
	case "physical":
		links["physicalActivityId"] = sourceID
		if sourceID == 0 {
			return
		}
		var record models.PhysicalTestRecord
		if err := ctrl.db.Select("id").
			Where("player_id = ? AND activity_id = ?", playerID, sourceID).
			Order("test_date DESC, created_at DESC").
			First(&record).Error; err == nil {
			links["physicalRecordId"] = record.ID
		} else {
			links["physicalRecordId"] = nil
		}
	}
}

func calendarItemUint(value interface{}) uint {
	switch typed := value.(type) {
	case uint:
		return typed
	case int:
		if typed > 0 {
			return uint(typed)
		}
	case int64:
		if typed > 0 {
			return uint(typed)
		}
	case float64:
		if typed > 0 {
			return uint(typed)
		}
	}
	return 0
}

func accumulateCalendarStats(target gin.H, source gin.H) {
	for _, key := range []string{
		"trainingCount",
		"matchCount",
		"physicalTestCount",
		"weeklyPeriodCount",
		"weeklyPendingCount",
		"matchReviewPendingCount",
	} {
		target[key] = calendarStatInt(target[key]) + calendarStatInt(source[key])
	}
}

func calendarStatInt(value interface{}) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case uint:
		return int(typed)
	default:
		return 0
	}
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

	age := effectivePlayerAge(user.Age, user.BirthDate)
	response := gin.H{
		"id":                   user.ID,
		"nickname":             user.Nickname,
		"real_name":            user.Name,
		"avatar":               user.Avatar,
		"age":                  age,
		"gender":               user.Gender,
		"position":             user.Position,
		"second_position":      user.SecondPosition,
		"dominant_foot":        user.Foot,
		"height":               user.Height,
		"weight":               user.Weight,
		"playing_style":        parseJSONArray(user.PlayingStyle),
		"current_team":         user.CurrentTeam,
		"start_year":           user.StartYear,
		"jersey_number":        user.JerseyNumber,
		"jersey_color":         user.JerseyColor,
		"province":             user.Province,
		"city":                 user.City,
		"school":               user.School,
		"fa_registered":        user.FARegistered,
		"association":          user.Association,
		"technical_tags":       parseJSONArray(user.TechnicalTags),
		"mental_tags":          parseJSONArray(user.MentalTags),
		"experiences":          parseExperiences(user.Experiences),
		"latest_physical_test": latestTest,
		"created_at":           user.CreatedAt,
	}

	utils.Success(c, "", gin.H{"player": response})
}

// GetHomepage 获取球员个人主页聚合数据（公开可访问，登录后补充访客权限）
func (ctrl *PlayerController) GetHomepage(c *gin.Context) {
	idStr := c.Param("playerId")
	var playerID uint
	if _, err := fmt.Sscanf(idStr, "%d", &playerID); err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的用户ID")
		return
	}

	user, err := models.NewUserRepository(ctrl.db).FindByID(playerID)
	if err != nil || user == nil || user.Status != models.StatusActive {
		utils.Error(c, http.StatusNotFound, "球员不存在")
		return
	}

	currentUser := ctrl.getOptionalCurrentUser(c)
	isOwnProfile := currentUser != nil && currentUser.ID == playerID
	privacy := parsePrivacySettings(user.PrivacySettings)
	if !privacy.ProfileVisible && !isOwnProfile {
		utils.Error(c, http.StatusForbidden, "该球员主页未公开")
		return
	}

	showRealName := privacy.ShowRealName || isOwnProfile
	displayName := user.Nickname
	if displayName == "" && showRealName {
		displayName = user.Name
	}
	if displayName == "" {
		displayName = fmt.Sprintf("球员%d", user.ID)
	}

	affiliation := ctrl.getHomepageAffiliation(playerID)
	physicalRecords := ctrl.getHomepagePhysicalRecords(playerID)
	weeklyReports := ctrl.getHomepageWeeklyReports(playerID)
	matches := ctrl.getHomepageMatches(playerID)
	reportList, reportsTotal, completedReports, avgReportRating := ctrl.getHomepageReports(playerID)
	scoutReports := ctrl.getHomepageScoutReports(playerID)
	posts := ctrl.getHomepagePosts(playerID)
	followersCount, followingCount, isFollowing, isMutual := ctrl.getHomepageSocial(playerID, currentUser)
	actions := ctrl.getHomepageActions(currentUser, playerID)
	timeline := ctrl.buildHomepageTimeline(physicalRecords, weeklyReports, matches, reportList, scoutReports, posts)
	age := effectivePlayerAge(user.Age, user.BirthDate)

	showSchool := isOwnProfile
	showReportDetails := isOwnProfile || currentUser != nil
	response := gin.H{
		"profile": gin.H{
			"id":          user.ID,
			"nickname":    user.Nickname,
			"displayName": displayName,
			"realName": func() string {
				if showRealName {
					return user.Name
				}
				return ""
			}(),
			"avatar":         user.Avatar,
			"age":            age,
			"ageGroup":       firstNonEmpty(affiliation["ageGroup"], getHomepageAgeGroup(age)),
			"gender":         user.Gender,
			"position":       user.Position,
			"secondPosition": user.SecondPosition,
			"height":         user.Height,
			"weight":         user.Weight,
			"dominantFoot":   firstNonEmpty(user.DominantFoot, user.Foot),
			"province":       user.Province,
			"city":           user.City,
			"school": func() string {
				if showSchool {
					return user.School
				}
				return ""
			}(),
			"currentTeam":         firstNonEmpty(affiliation["teamName"], user.CurrentTeam, user.Club),
			"startYear":           user.StartYear,
			"jerseyNumber":        user.JerseyNumber,
			"jerseyColor":         user.JerseyColor,
			"faRegistered":        user.FARegistered,
			"association":         user.Association,
			"playingStyle":        parseJSONArray(user.PlayingStyle),
			"technicalTags":       parseJSONArray(user.TechnicalTags),
			"mentalTags":          parseJSONArray(user.MentalTags),
			"experiences":         parseExperiences(user.Experiences),
			"profileCompleteness": ctrl.calculateCompleteness(user),
			"updatedAt":           user.UpdatedAt.Format("2006-01-02 15:04:05"),
		},
		"affiliation": affiliation,
		"visibility": gin.H{
			"profileVisible":    privacy.ProfileVisible,
			"showRealName":      showRealName,
			"showSchool":        showSchool,
			"showFamily":        isOwnProfile,
			"showReportDetails": showReportDetails,
			"canMessage":        actions["canMessage"],
			"canInvite":         actions["canInviteTrial"],
			"canEdit":           actions["canEdit"],
		},
		"stats": gin.H{
			"followersCount":        followersCount,
			"followingCount":        followingCount,
			"reportsCount":          reportsTotal,
			"completedReportsCount": completedReports,
			"averageReportRating":   avgReportRating,
			"physicalTestCount":     len(physicalRecords),
			"weeklyReportCount":     len(weeklyReports),
			"matchCount":            len(matches),
			"scoutReportsCount":     len(scoutReports),
			"postCount":             len(posts),
		},
		"physicalTests": gin.H{
			"records": physicalRecords,
			"latest":  firstHomepageItem(physicalRecords),
		},
		"weeklyReports": gin.H{
			"total": len(weeklyReports),
			"list":  weeklyReports,
		},
		"matches": gin.H{
			"total": len(matches),
			"list":  matches,
		},
		"reports": gin.H{
			"total": reportsTotal,
			"list":  reportList,
		},
		"scoutReports": gin.H{
			"total": len(scoutReports),
			"list":  scoutReports,
		},
		"posts": gin.H{
			"total": len(posts),
			"list":  posts,
		},
		"timeline": timeline,
		"social": gin.H{
			"followersCount": followersCount,
			"followingCount": followingCount,
			"isFollowing":    isFollowing,
			"isMutual":       isMutual,
		},
		"actions": actions,
	}

	utils.Success(c, "", response)
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
			"activity_id": func() *uint {
				v := r.ActivityID
				if v == 0 {
					return nil
				}
				return &v
			}(),
			"recorder_role": recorderRole,
			"created_at":    r.CreatedAt.Format("2006-01-02 15:04"),
		}
	}

	utils.Success(c, "", gin.H{
		"records": list,
		"total":   len(list),
	})
}

// ============ 辅助函数 ============

func (ctrl *PlayerController) getOptionalCurrentUser(c *gin.Context) *models.User {
	authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
	if authHeader == "" {
		return nil
	}
	token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if token == "" || token == authHeader {
		return nil
	}
	claims, err := middleware.ParseToken(token)
	if err != nil {
		return nil
	}
	user, err := models.NewUserRepository(ctrl.db).FindByID(claims.UserID)
	if err != nil || user == nil || user.Status != models.StatusActive {
		return nil
	}
	return user
}

func (ctrl *PlayerController) getHomepageAffiliation(playerID uint) gin.H {
	var teamPlayer models.TeamPlayer
	if err := ctrl.db.Preload("Team.Club").
		Where("user_id = ? AND status = ?", playerID, "active").
		Order("joined_at DESC, id DESC").
		First(&teamPlayer).Error; err == nil {
		teamName := ""
		ageGroup := ""
		var clubID uint
		clubName := ""
		clubLogo := ""
		clubMemberLevel := ""
		if teamPlayer.Team != nil {
			teamName = teamPlayer.Team.Name
			ageGroup = teamPlayer.Team.AgeGroup
			clubID = teamPlayer.Team.ClubID
			if teamPlayer.Team.Club != nil {
				clubName = teamPlayer.Team.Club.Name
				clubLogo = teamPlayer.Team.Club.Logo
				clubMemberLevel = string(teamPlayer.Team.Club.MemberLevel)
			}
		}
		return gin.H{
			"clubId":          clubID,
			"clubName":        clubName,
			"clubLogo":        clubLogo,
			"clubMemberLevel": clubMemberLevel,
			"teamId":          teamPlayer.TeamID,
			"teamName":        teamName,
			"ageGroup":        ageGroup,
			"jerseyNumber":    teamPlayer.JerseyNumber,
			"position":        teamPlayer.Position,
			"joinedAt":        formatHomepageTime(teamPlayer.JoinedAt),
			"status":          teamPlayer.Status,
		}
	}

	var clubPlayer models.ClubPlayer
	if err := ctrl.db.Preload("Club").
		Where("user_id = ? AND status = ?", playerID, "active").
		Order("join_date DESC, id DESC").
		First(&clubPlayer).Error; err == nil {
		clubName := ""
		clubLogo := ""
		clubMemberLevel := ""
		if clubPlayer.Club != nil {
			clubName = clubPlayer.Club.Name
			clubLogo = clubPlayer.Club.Logo
			clubMemberLevel = string(clubPlayer.Club.MemberLevel)
		}
		return gin.H{
			"clubId":          clubPlayer.ClubID,
			"clubName":        clubName,
			"clubLogo":        clubLogo,
			"clubMemberLevel": clubMemberLevel,
			"ageGroup":        clubPlayer.AgeGroup,
			"position":        clubPlayer.Position,
			"joinedAt":        formatHomepageTime(clubPlayer.JoinDate),
			"status":          clubPlayer.Status,
		}
	}

	return gin.H{}
}

func (ctrl *PlayerController) getHomepagePhysicalRecords(playerID uint) []gin.H {
	var records []models.PhysicalTestRecord
	ctrl.db.Preload("Activity.Club").
		Where("player_id = ?", playerID).
		Order("test_date DESC, created_at DESC").
		Limit(20).
		Find(&records)

	items := make([]gin.H, 0, len(records))
	for _, record := range records {
		source := "personal"
		sourceLabel := "球员/家长自测"
		clubName := ""
		activityName := ""
		if record.ActivityID > 0 && record.ClubID > 0 {
			source = "club"
			sourceLabel = "俱乐部官方体测"
			if record.Activity != nil {
				activityName = record.Activity.Name
				if record.Activity.Club != nil {
					clubName = record.Activity.Club.Name
				}
			}
		}
		recorderRole := "player"
		if record.RecorderID != 0 && record.RecorderID != playerID {
			recorderRole = "coach"
		}
		items = append(items, gin.H{
			"id":               record.ID,
			"testDate":         formatHomepageTime(record.TestDate),
			"source":           source,
			"sourceLabel":      sourceLabel,
			"clubName":         clubName,
			"activityName":     activityName,
			"recorderRole":     recorderRole,
			"height":           record.Height,
			"weight":           record.Weight,
			"bmi":              record.BMI,
			"sprint30m":        record.Sprint30m,
			"sprint50m":        record.Sprint50m,
			"sprint100m":       record.Sprint100m,
			"agilityLadder":    record.AgilityLadder,
			"tTest":            record.TTest,
			"shuttleRun":       record.ShuttleRun,
			"standingLongJump": record.StandingLongJump,
			"verticalJump":     record.VerticalJump,
			"sitAndReach":      record.SitAndReach,
			"pushUp":           record.PushUp,
			"sitUp":            record.SitUp,
			"plank":            record.Plank,
		})
	}
	return items
}

func (ctrl *PlayerController) getHomepageWeeklyReports(playerID uint) []gin.H {
	var reports []models.WeeklyReport
	ctrl.db.Preload("Team").
		Where("player_id = ?", playerID).
		Order("week_start DESC, created_at DESC").
		Limit(6).
		Find(&reports)

	items := make([]gin.H, 0, len(reports))
	for _, report := range reports {
		year, week := report.WeekStart.ISOWeek()
		teamName := ""
		if report.Team != nil {
			teamName = report.Team.Name
		}
		items = append(items, gin.H{
			"id":            report.ID,
			"teamName":      teamName,
			"weekLabel":     fmt.Sprintf("%d年第%d周", year, week),
			"weekStart":     formatHomepageTime(report.WeekStart),
			"weekEnd":       formatHomepageTime(report.WeekEnd),
			"submitStatus":  report.SubmitStatus,
			"reviewStatus":  report.ReviewStatus,
			"selfAverage":   homepageAverageInts(report.SelfAttitudeRating, report.SelfTechniqueRating, report.SelfTeamworkRating),
			"coachAverage":  homepageAverageInts(report.CoachAttitudeRating, report.CoachTechniqueRating, report.CoachTacticsRating, report.CoachKnowledgeRating),
			"reviewComment": report.ReviewComment,
			"suggestions":   report.Suggestions,
			"nextWeekFocus": report.NextWeekFocus,
			"createdAt":     formatHomepageTime(report.CreatedAt),
		})
	}
	return items
}

func (ctrl *PlayerController) getHomepageMatches(playerID uint) []gin.H {
	var reviews []models.PlayerReview
	ctrl.db.Preload("Match.Team").
		Where("player_id = ?", playerID).
		Order("submitted_at DESC, created_at DESC").
		Limit(6).
		Find(&reviews)

	items := make([]gin.H, 0, len(reviews))
	for _, review := range reviews {
		matchName := "比赛记录"
		matchDate := formatHomepageTime(review.CreatedAt)
		opponent := ""
		score := ""
		result := ""
		status := review.Status
		teamName := ""
		if review.Match != nil {
			matchName = review.Match.MatchName
			matchDate = review.Match.MatchDate
			opponent = review.Match.Opponent
			score = fmt.Sprintf("%d:%d", review.Match.OurScore, review.Match.OppScore)
			result = review.Match.Result
			status = review.Match.Status
			if review.Match.Team != nil {
				teamName = review.Match.Team.Name
			}
		}
		items = append(items, gin.H{
			"id":           review.MatchID,
			"reviewId":     review.ID,
			"matchName":    matchName,
			"matchDate":    matchDate,
			"opponent":     opponent,
			"score":        score,
			"result":       result,
			"status":       status,
			"teamName":     teamName,
			"performance":  review.Performance,
			"goals":        review.Goals,
			"assists":      review.Assists,
			"saves":        review.Saves,
			"coachRating":  review.CoachRating,
			"coachComment": review.CoachComment,
			"highlights":   review.Highlights,
			"createdAt":    formatHomepageTime(review.CreatedAt),
		})
	}
	return items
}

func (ctrl *PlayerController) getHomepageReports(playerID uint) ([]gin.H, int64, int64, float64) {
	var total int64
	ctrl.db.Model(&models.Report{}).Where("user_id = ?", playerID).Count(&total)

	var completedTotal int64
	ctrl.db.Model(&models.Report{}).
		Where("user_id = ? AND status = ?", playerID, models.ReportStatusCompleted).
		Count(&completedTotal)

	var avg float64
	ctrl.db.Model(&models.Report{}).
		Select("COALESCE(AVG(overall_rating), 0)").
		Where("user_id = ? AND status = ? AND overall_rating > 0", playerID, models.ReportStatusCompleted).
		Row().
		Scan(&avg)

	var reports []models.Report
	ctrl.db.Where("user_id = ? AND status = ?", playerID, models.ReportStatusCompleted).
		Order("created_at DESC").
		Limit(5).
		Find(&reports)

	analystIDs := make([]uint, 0)
	seenAnalysts := map[uint]bool{}
	for _, report := range reports {
		if report.AnalystID > 0 && !seenAnalysts[report.AnalystID] {
			seenAnalysts[report.AnalystID] = true
			analystIDs = append(analystIDs, report.AnalystID)
		}
	}
	analystNames := map[uint]string{}
	if len(analystIDs) > 0 {
		var users []models.User
		ctrl.db.Select("id, name, nickname").Where("id IN ?", analystIDs).Find(&users)
		for _, u := range users {
			analystNames[u.ID] = firstNonEmpty(u.Nickname, u.Name)
		}
	}

	items := make([]gin.H, 0, len(reports))
	for _, report := range reports {
		items = append(items, gin.H{
			"id":             report.ID,
			"createdAt":      formatHomepageTime(report.CreatedAt),
			"status":         report.Status,
			"playerName":     report.PlayerName,
			"playerPosition": report.PlayerPosition,
			"overallRating":  report.OverallRating,
			"offenseRating":  report.OffenseRating,
			"defenseRating":  report.DefenseRating,
			"summary":        report.Summary,
			"strengths":      homepageStringList(report.Strengths),
			"weaknesses":     homepageStringList(report.Weaknesses),
			"suggestions":    report.Suggestions,
			"potential":      report.Potential,
			"analystName":    analystNames[report.AnalystID],
		})
	}

	return items, total, completedTotal, avg
}

func (ctrl *PlayerController) getHomepageScoutReports(playerID uint) []gin.H {
	var reports []models.ScoutReport
	ctrl.db.Preload("Scout.User").
		Where("player_id = ? AND status = ?", playerID, "published").
		Order("published_at DESC, created_at DESC").
		Limit(6).
		Find(&reports)

	items := make([]gin.H, 0, len(reports))
	for _, report := range reports {
		reportDate := report.CreatedAt
		if report.PublishedAt != nil {
			reportDate = *report.PublishedAt
		}

		scoutName := "球探"
		scoutOrganization := ""
		scoutUserID := uint(0)
		if report.Scout != nil {
			scoutOrganization = report.Scout.CurrentOrganization
			if report.Scout.User != nil {
				scoutUserID = report.Scout.User.ID
				scoutName = firstNonEmpty(report.Scout.User.Nickname, report.Scout.User.Name, scoutName)
			}
		}

		items = append(items, gin.H{
			"id":              report.ID,
			"createdAt":       formatHomepageTime(reportDate),
			"status":          report.Status,
			"scoutId":         report.ScoutID,
			"scoutUserId":     scoutUserID,
			"scoutName":       scoutName,
			"organization":    scoutOrganization,
			"overallRating":   report.OverallRating,
			"potentialRating": report.PotentialRating,
			"strengths":       homepageStringList(report.Strengths),
			"weaknesses":      homepageStringList(report.Weaknesses),
			"technicalSkills": homepageObject(report.TechnicalSkills),
			"summary":         report.Summary,
			"recommendation":  report.Recommendation,
			"targetClub":      report.TargetClub,
			"views":           report.Views,
			"likes":           report.Likes,
			"publishedAt":     formatHomepageTime(reportDate),
		})
	}
	return items
}

func (ctrl *PlayerController) getHomepagePosts(playerID uint) []gin.H {
	var posts []models.Post
	ctrl.db.Where("user_id = ?", playerID).
		Order("created_at DESC").
		Limit(6).
		Find(&posts)

	items := make([]gin.H, 0, len(posts))
	for _, post := range posts {
		items = append(items, gin.H{
			"id":            post.ID,
			"content":       post.Content,
			"images":        post.GetImagesArray(),
			"roleTag":       post.RoleTag,
			"likesCount":    post.LikesCount,
			"commentsCount": post.CommentsCount,
			"createdAt":     formatHomepageTime(post.CreatedAt),
		})
	}
	return items
}

func (ctrl *PlayerController) getHomepageSocial(playerID uint, currentUser *models.User) (int64, int64, bool, bool) {
	var followersCount int64
	var followingCount int64
	ctrl.db.Model(&models.Follow{}).Where("following_id = ?", playerID).Count(&followersCount)
	ctrl.db.Model(&models.Follow{}).Where("follower_id = ?", playerID).Count(&followingCount)

	isFollowing := false
	isMutual := false
	if currentUser != nil && currentUser.ID != playerID {
		var count int64
		ctrl.db.Model(&models.Follow{}).
			Where("follower_id = ? AND following_id = ?", currentUser.ID, playerID).
			Count(&count)
		isFollowing = count > 0
		ctrl.db.Model(&models.Follow{}).
			Where("follower_id = ? AND following_id = ?", playerID, currentUser.ID).
			Count(&count)
		isMutual = isFollowing && count > 0
	}

	return followersCount, followingCount, isFollowing, isMutual
}

func (ctrl *PlayerController) getHomepageActions(currentUser *models.User, playerID uint) gin.H {
	isOwn := currentUser != nil && currentUser.ID == playerID
	role := ""
	if currentUser != nil {
		role = string(currentUser.CurrentRole)
		if role == "" {
			role = string(currentUser.Role)
		}
	}
	return gin.H{
		"canEdit":              isOwn,
		"canFollow":            currentUser != nil && !isOwn,
		"canMessage":           currentUser != nil && !isOwn,
		"canInviteTrial":       role == "club" || role == "coach" || role == "scout" || role == "admin",
		"canCreateScoutReport": role == "scout" || role == "admin",
	}
}

func (ctrl *PlayerController) buildHomepageTimeline(physical []gin.H, weekly []gin.H, matches []gin.H, reports []gin.H, scoutReports []gin.H, posts []gin.H) []gin.H {
	items := make([]gin.H, 0, len(physical)+len(weekly)+len(matches)+len(reports)+len(scoutReports)+len(posts))
	for _, item := range physical {
		items = append(items, gin.H{
			"id":      fmt.Sprintf("physical-%v", item["id"]),
			"type":    "physical_test",
			"date":    item["testDate"],
			"title":   firstNonEmpty(item["activityName"], "体测记录"),
			"summary": fmt.Sprintf("%s · %s", firstNonEmpty(item["sourceLabel"], "体测数据"), firstNonEmpty(item["clubName"])),
			"source": func() string {
				if item["source"] == "club" {
					return "club"
				}
				return "player"
			}(),
			"sourceLabel": firstNonEmpty(item["sourceLabel"], "体测数据"),
		})
	}
	for _, item := range weekly {
		items = append(items, gin.H{
			"id":          fmt.Sprintf("weekly-%v", item["id"]),
			"type":        "weekly_report",
			"date":        item["weekEnd"],
			"title":       fmt.Sprintf("%s 周报", firstNonEmpty(item["weekLabel"], "训练")),
			"summary":     firstNonEmpty(item["reviewComment"], item["suggestions"], item["submitStatus"]),
			"source":      "coach",
			"sourceLabel": "周报反馈",
		})
	}
	for _, item := range matches {
		items = append(items, gin.H{
			"id":          fmt.Sprintf("match-%v", item["id"]),
			"type":        "match",
			"date":        item["matchDate"],
			"title":       firstNonEmpty(item["matchName"], "比赛总结"),
			"summary":     firstNonEmpty(item["coachComment"], item["performance"], item["score"]),
			"source":      "coach",
			"sourceLabel": "比赛点评",
		})
	}
	for _, item := range reports {
		items = append(items, gin.H{
			"id":          fmt.Sprintf("report-%v", item["id"]),
			"type":        "report",
			"date":        item["createdAt"],
			"title":       "视频分析报告",
			"summary":     firstNonEmpty(item["summary"], item["suggestions"]),
			"source":      "analyst",
			"sourceLabel": "平台分析师",
		})
	}
	for _, item := range scoutReports {
		items = append(items, gin.H{
			"id":          fmt.Sprintf("scout-report-%v", item["id"]),
			"type":        "scout_report",
			"date":        item["publishedAt"],
			"title":       "球探观察报告",
			"summary":     firstNonEmpty(item["summary"], item["recommendation"], item["targetClub"]),
			"source":      "scout",
			"sourceLabel": firstNonEmpty(item["scoutName"], "球探"),
		})
	}
	for _, item := range posts {
		items = append(items, gin.H{
			"id":          fmt.Sprintf("post-%v", item["id"]),
			"type":        "post",
			"date":        item["createdAt"],
			"title":       "训练动态",
			"summary":     item["content"],
			"source":      "player",
			"sourceLabel": "球员动态",
		})
	}
	sort.SliceStable(items, func(i, j int) bool {
		return firstNonEmpty(items[i]["date"]) > firstNonEmpty(items[j]["date"])
	})
	if len(items) > 12 {
		return items[:12]
	}
	return items
}

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

func firstHomepageItem(items []gin.H) interface{} {
	if len(items) == 0 {
		return nil
	}
	return items[0]
}

func firstNonEmpty(values ...interface{}) string {
	for _, value := range values {
		switch v := value.(type) {
		case string:
			if strings.TrimSpace(v) != "" {
				return v
			}
		case fmt.Stringer:
			if strings.TrimSpace(v.String()) != "" {
				return v.String()
			}
		}
	}
	return ""
}

func formatHomepageTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

func getHomepageAgeGroup(age int) string {
	if age <= 0 {
		return ""
	}
	if age <= 12 {
		return "U12"
	}
	if age <= 14 {
		return "U14"
	}
	if age <= 16 {
		return "U16"
	}
	return "U18"
}

func calculatePlayerAgeFromBirthDate(birthDate string) int {
	birthDate = strings.TrimSpace(birthDate)
	if birthDate == "" {
		return 0
	}
	layouts := []string{"2006-01-02", "2006/01/02", "01-02-2006"}
	var parsed time.Time
	for _, layout := range layouts {
		t, err := time.Parse(layout, birthDate)
		if err == nil {
			parsed = t
			break
		}
	}
	if parsed.IsZero() {
		return 0
	}
	now := time.Now()
	age := now.Year() - parsed.Year()
	if now.Month() < parsed.Month() || (now.Month() == parsed.Month() && now.Day() < parsed.Day()) {
		age--
	}
	if age < 0 {
		return 0
	}
	return age
}

func effectivePlayerAge(age int, birthDate string) int {
	if age > 0 {
		return age
	}
	return calculatePlayerAgeFromBirthDate(birthDate)
}

func homepageAverageInts(values ...int) float64 {
	total := 0
	count := 0
	for _, value := range values {
		if value > 0 {
			total += value
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return float64(total) / float64(count)
}

func homepageStringList(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{}
	}
	items := parseJSONArray(raw)
	if len(items) > 0 {
		return items
	}
	return []string{raw}
}

func homepageObject(raw string) gin.H {
	if strings.TrimSpace(raw) == "" {
		return gin.H{}
	}
	var item map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &item); err != nil {
		return gin.H{}
	}
	return gin.H(item)
}

func calculateBMI(height, weight *float64) *float64 {
	if height == nil || weight == nil || *height <= 0 || *weight <= 0 {
		return nil
	}
	m := *height / 100.0
	val := *weight / (m * m)
	return &val
}
