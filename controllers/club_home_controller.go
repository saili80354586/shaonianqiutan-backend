package controllers

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/repositories"
	"github.com/shaonianqiutan/backend/utils"
	"gorm.io/gorm"
)

// ClubHomeController 俱乐部主页控制器
type ClubHomeController struct {
	clubHomeRepo *repositories.ClubHomeRepository
	db           *gorm.DB
}

// NewClubHomeController 创建俱乐部主页控制器
func NewClubHomeController(clubHomeRepo *repositories.ClubHomeRepository, db *gorm.DB) *ClubHomeController {
	return &ClubHomeController{clubHomeRepo: clubHomeRepo, db: db}
}

// GetClubHome 获取俱乐部主页配置
func (c *ClubHomeController) GetClubHome(ctx *gin.Context) {
	clubIDStr := ctx.Param("clubId")
	clubID, err := strconv.ParseUint(clubIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的俱乐部ID")
		return
	}

	// 获取俱乐部基础信息
	var club models.Club
	if err := c.db.First(&club, clubID).Error; err != nil {
		utils.NotFoundError(ctx, "俱乐部不存在")
		return
	}

	// 获取主页配置
	home, err := c.clubHomeRepo.FindByClubID(uint(clubID))
	if err != nil {
		home = models.DefaultClubHome(uint(clubID))
		if err := c.clubHomeRepo.Create(home); err != nil {
			utils.ServerError(ctx, "获取失败")
			return
		}
	}

	// 解析模块排序
	var moduleOrder []string
	if home.ModuleOrder != "" {
		json.Unmarshal([]byte(home.ModuleOrder), &moduleOrder)
	}
	if len(moduleOrder) == 0 {
		moduleOrder = models.DefaultModuleOrder
	}
	if home.ModuleVisibility == nil {
		home.ModuleVisibility = models.DefaultModuleVisibility
	}

	// 获取成就列表
	achievements, _ := c.clubHomeRepo.GetAchievements(uint(clubID))

	// 获取球队列表（真实数据 + 展示配置）
	teamConfigs, _ := c.clubHomeRepo.GetClubHomeTeams(uint(clubID))
	teamConfigMap := make(map[uint]models.ClubHomeTeam)
	showTeamIDs := make([]uint, 0)
	for _, tc := range teamConfigs {
		teamConfigMap[tc.TeamID] = tc
		showTeamIDs = append(showTeamIDs, tc.TeamID)
	}

	var allTeams []models.Team
	c.db.Where("club_id = ? AND status = ?", clubID, "active").Order("age_group ASC, created_at DESC").Find(&allTeams)

	teamsData := make([]map[string]interface{}, 0)
	for _, t := range allTeams {
		playerCount, _ := repositories.NewTeamRepository(c.db).CountPlayers(t.ID)
		showPlayerCount := true
		if cfg, ok := teamConfigMap[t.ID]; ok {
			showPlayerCount = cfg.ShowPlayerCount
		}
		teamsData = append(teamsData, map[string]interface{}{
			"id":              t.ID,
			"name":            t.Name,
			"ageGroup":        t.AgeGroup,
			"description":     t.Description,
			"playerCount":     playerCount,
			"showPlayerCount": showPlayerCount,
			"isShown":         len(showTeamIDs) == 0 || containsUint(showTeamIDs, t.ID),
		})
	}

	// 获取教练列表（真实数据 + 展示配置）
	coachConfigs, _ := c.clubHomeRepo.GetClubHomeCoaches(uint(clubID))
	coachConfigMap := make(map[uint]models.ClubHomeCoach)
	showCoachIDs := make([]uint, 0)
	for _, cc := range coachConfigs {
		coachConfigMap[cc.CoachID] = cc
		showCoachIDs = append(showCoachIDs, cc.CoachID)
	}

	var coachesData []map[string]interface{}
	c.db.Table("team_coaches tc").
		Select("tc.id, tc.user_id, u.name, u.nickname, u.avatar, tc.role, c.license_level").
		Joins("JOIN teams t ON t.id = tc.team_id").
		Joins("JOIN users u ON u.id = tc.user_id").
		Joins("LEFT JOIN coaches c ON c.user_id = tc.user_id").
		Where("t.club_id = ? AND tc.status = ?", clubID, "active").
		Scan(&coachesData)

	// 去重教练
	seenCoachUsers := make(map[uint]bool)
	uniqueCoaches := make([]map[string]interface{}, 0)
	for _, coach := range coachesData {
		userID := uint(0)
		if uid, ok := coach["user_id"].(uint); ok {
			userID = uid
		} else if uid, ok := coach["user_id"].(uint64); ok {
			userID = uint(uid)
		} else if uid, ok := coach["user_id"].(int64); ok {
			userID = uint(uid)
		}
		if userID > 0 && seenCoachUsers[userID] {
			continue
		}
		seenCoachUsers[userID] = true
		coachID := uint(0)
		if cid, ok := coach["id"].(uint); ok {
			coachID = cid
		} else if cid, ok := coach["id"].(uint64); ok {
			coachID = uint(cid)
		} else if cid, ok := coach["id"].(int64); ok {
			coachID = uint(cid)
		}
		coach["isShown"] = len(showCoachIDs) == 0 || containsUint(showCoachIDs, coachID)
		uniqueCoaches = append(uniqueCoaches, coach)
	}

	// 获取球员列表（真实数据 + 展示配置）
	playerConfigs, _ := c.clubHomeRepo.GetClubHomePlayers(uint(clubID))
	playerConfigMap := make(map[uint]models.ClubHomePlayer)
	showPlayerIDs := make([]uint, 0)
	for _, pc := range playerConfigs {
		playerConfigMap[pc.PlayerID] = pc
		showPlayerIDs = append(showPlayerIDs, pc.PlayerID)
	}

	var playersData []map[string]interface{}
	c.db.Table("club_players cp").
		Select("cp.id, cp.user_id, u.name, u.nickname, u.avatar, u.age, cp.position, cp.age_group").
		Joins("JOIN users u ON u.id = cp.user_id").
		Where("cp.club_id = ? AND cp.status = ?", clubID, "active").
		Scan(&playersData)

	for _, p := range playersData {
		userID := uint(0)
		if uid, ok := p["user_id"].(uint); ok {
			userID = uid
		} else if uid, ok := p["user_id"].(uint64); ok {
			userID = uint(uid)
		} else if uid, ok := p["user_id"].(int64); ok {
			userID = uint(uid)
		}
		recommendText := ""
		if cfg, ok := playerConfigMap[userID]; ok {
			recommendText = cfg.RecommendText
		}
		p["recommendText"] = recommendText
		p["isShown"] = len(showPlayerIDs) == 0 || containsUint(showPlayerIDs, userID)
	}

	// 获取最近比赛
	var matches []map[string]interface{}
	c.db.Table("match_summaries").
		Select("id, match_name as title, match_date as date, match_result as result, opponent, our_score, opponent_score").
		Where("club_id = ?", clubID).
		Order("match_date DESC").
		Limit(3).
		Scan(&matches)
	for _, m := range matches {
		m["type"] = "match"
	}

	// 获取最近体测
	var tests []map[string]interface{}
	c.db.Table("physical_test_activities").
		Select("id, name, start_date as date, status").
		Where("club_id = ?", clubID).
		Order("start_date DESC").
		Limit(2).
		Scan(&tests)
	for _, t := range tests {
		t["type"] = "physical_test"
	}

	// 获取活动
	var activities []map[string]interface{}
	c.db.Table("club_activities").
		Select("id, title, type, status, description, cover_image as coverImage, start_time as startTime, end_time as endTime, location, max_participants as maxParticipants, contact_phone as contactPhone, contact_wechat as contactWechat, is_review as isReview, review_content as reviewContent, review_images as reviewImages").
		Where("club_id = ?", clubID).
		Order("start_time DESC").
		Scan(&activities)

	// 组装响应
	utils.SuccessResponse(ctx, gin.H{
		"club": gin.H{
			"id":              club.ID,
			"name":            club.Name,
			"logo":            club.Logo,
			"description":     club.Description,
			"address":         club.Address,
			"contactPhone":    club.ContactPhone,
			"contactName":     club.ContactName,
			"establishedYear": club.EstablishedYear,
			"clubSize":        club.ClubSize,
			"memberLevel":     club.MemberLevel,
		},
		"modules": gin.H{
			"order":      moduleOrder,
			"visibility": home.ModuleVisibility,
		},
		"hero":         home.Hero,
		"about":        home.About,
		"achievements": achievements,
		"teams":        teamsData,
		"coaches":      uniqueCoaches,
		"players":      playersData,
		"facilities":   home.Facilities,
		"recruitment":  home.Recruitment,
		"contact":      home.Contact,
		"socialLinks":  home.SocialLinks,
		"news": gin.H{
			"matches": matches,
			"tests":   tests,
			"manualItems": home.NewsItems,
		},
		"activities": activities,
	})
}

func containsUint(arr []uint, val uint) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}

// SaveClubHome 保存俱乐部主页配置
func (c *ClubHomeController) SaveClubHome(ctx *gin.Context) {
	clubIDStr := ctx.Param("clubId")
	clubID, err := strconv.ParseUint(clubIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的俱乐部ID")
		return
	}

	var req struct {
		Hero             *models.ClubHomeHero        `json:"hero"`
		About            *models.ClubHomeAbout       `json:"about"`
		Contact          *models.ClubHomeContact     `json:"contact"`
		Facilities       *models.ClubHomeFacilities  `json:"facilities"`
		Recruitment      *models.ClubHomeRecruitment `json:"recruitment"`
		SocialLinks      *models.ClubHomeSocialLinks `json:"socialLinks"`
		ModuleOrder      []string                    `json:"moduleOrder"`
		ModuleVisibility map[string]bool             `json:"moduleVisibility"`
		Achievements     []models.Achievement        `json:"achievements"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	home, err := c.clubHomeRepo.FindByClubID(uint(clubID))
	if err != nil {
		home = models.DefaultClubHome(uint(clubID))
	}

	if req.Hero != nil {
		home.Hero = *req.Hero
	}
	if req.About != nil {
		home.About = *req.About
	}
	if req.Contact != nil {
		home.Contact = *req.Contact
	}
	if req.Facilities != nil {
		home.Facilities = *req.Facilities
	}
	if req.Recruitment != nil {
		home.Recruitment = *req.Recruitment
	}
	if req.SocialLinks != nil {
		home.SocialLinks = *req.SocialLinks
	}
	if req.ModuleOrder != nil {
		orderJSON, _ := json.Marshal(req.ModuleOrder)
		home.ModuleOrder = string(orderJSON)
	}
	if req.ModuleVisibility != nil {
		home.ModuleVisibility = req.ModuleVisibility
	}

	if err := c.clubHomeRepo.Save(home); err != nil {
		utils.ServerError(ctx, "保存失败")
		return
	}

	if req.Achievements != nil {
		c.clubHomeRepo.SaveAchievements(uint(clubID), req.Achievements)
	}

	utils.SuccessResponseWithMessage(ctx, gin.H{"id": home.ID}, "保存成功")
}

// UpdateHero 更新 Hero 配置
func (c *ClubHomeController) UpdateHero(ctx *gin.Context) {
	clubIDStr := ctx.Param("clubId")
	clubID, err := strconv.ParseUint(clubIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的俱乐部ID")
		return
	}

	var req models.ClubHomeHero
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	if err := c.clubHomeRepo.UpdateHero(uint(clubID), &req); err != nil {
		utils.ServerError(ctx, "更新失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "更新成功")
}

// UpdateAbout 更新 About 配置
func (c *ClubHomeController) UpdateAbout(ctx *gin.Context) {
	clubIDStr := ctx.Param("clubId")
	clubID, err := strconv.ParseUint(clubIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的俱乐部ID")
		return
	}

	var req models.ClubHomeAbout
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	if err := c.clubHomeRepo.UpdateAbout(uint(clubID), &req); err != nil {
		utils.ServerError(ctx, "更新失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "更新成功")
}

// UpdateContact 更新联系方式
func (c *ClubHomeController) UpdateContact(ctx *gin.Context) {
	clubIDStr := ctx.Param("clubId")
	clubID, err := strconv.ParseUint(clubIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的俱乐部ID")
		return
	}

	var req models.ClubHomeContact
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	if err := c.clubHomeRepo.UpdateContact(uint(clubID), &req); err != nil {
		utils.ServerError(ctx, "更新失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "更新成功")
}

// UpdateFacilities 更新训练环境配置
func (c *ClubHomeController) UpdateFacilities(ctx *gin.Context) {
	clubIDStr := ctx.Param("clubId")
	clubID, err := strconv.ParseUint(clubIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的俱乐部ID")
		return
	}

	var req models.ClubHomeFacilities
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	if err := c.clubHomeRepo.UpdateFacilities(uint(clubID), &req); err != nil {
		utils.ServerError(ctx, "更新失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "更新成功")
}

// UpdateRecruitment 更新招生信息配置
func (c *ClubHomeController) UpdateRecruitment(ctx *gin.Context) {
	clubIDStr := ctx.Param("clubId")
	clubID, err := strconv.ParseUint(clubIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的俱乐部ID")
		return
	}

	var req models.ClubHomeRecruitment
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	if err := c.clubHomeRepo.UpdateRecruitment(uint(clubID), &req); err != nil {
		utils.ServerError(ctx, "更新失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "更新成功")
}

// UpdateSocialLinks 更新社交媒体链接
func (c *ClubHomeController) UpdateSocialLinks(ctx *gin.Context) {
	clubIDStr := ctx.Param("clubId")
	clubID, err := strconv.ParseUint(clubIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的俱乐部ID")
		return
	}

	var req models.ClubHomeSocialLinks
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	if err := c.clubHomeRepo.UpdateSocialLinks(uint(clubID), &req); err != nil {
		utils.ServerError(ctx, "更新失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "更新成功")
}

// UpdateNews 更新手工置顶公告
func (c *ClubHomeController) UpdateNews(ctx *gin.Context) {
	clubIDStr := ctx.Param("clubId")
	clubID, err := strconv.ParseUint(clubIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的俱乐部ID")
		return
	}

	var req []models.ClubHomeNewsItem
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	if err := c.clubHomeRepo.UpdateNews(uint(clubID), req); err != nil {
		utils.ServerError(ctx, "更新失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "更新成功")
}

// UpdateModules 更新模块排序和可见性
func (c *ClubHomeController) UpdateModules(ctx *gin.Context) {
	clubIDStr := ctx.Param("clubId")
	clubID, err := strconv.ParseUint(clubIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的俱乐部ID")
		return
	}

	var req struct {
		ModuleOrder      []string        `json:"moduleOrder"`
		ModuleVisibility map[string]bool `json:"moduleVisibility"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	orderJSON, _ := json.Marshal(req.ModuleOrder)
	if err := c.clubHomeRepo.UpdateModules(uint(clubID), string(orderJSON), req.ModuleVisibility); err != nil {
		utils.ServerError(ctx, "更新失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "更新成功")
}

// UpdateTeams 保存主页展示的球队配置
func (c *ClubHomeController) UpdateTeams(ctx *gin.Context) {
	clubIDStr := ctx.Param("clubId")
	clubID, err := strconv.ParseUint(clubIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的俱乐部ID")
		return
	}

	var req []models.ClubHomeTeam
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	if err := c.clubHomeRepo.SaveClubHomeTeams(uint(clubID), req); err != nil {
		utils.ServerError(ctx, "保存失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "保存成功")
}

// UpdateCoaches 保存主页展示的教练配置
func (c *ClubHomeController) UpdateCoaches(ctx *gin.Context) {
	clubIDStr := ctx.Param("clubId")
	clubID, err := strconv.ParseUint(clubIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的俱乐部ID")
		return
	}

	var req []models.ClubHomeCoach
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	if err := c.clubHomeRepo.SaveClubHomeCoaches(uint(clubID), req); err != nil {
		utils.ServerError(ctx, "保存失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "保存成功")
}

// UpdatePlayers 保存主页展示的球员配置
func (c *ClubHomeController) UpdatePlayers(ctx *gin.Context) {
	clubIDStr := ctx.Param("clubId")
	clubID, err := strconv.ParseUint(clubIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的俱乐部ID")
		return
	}

	var req []models.ClubHomePlayer
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	if err := c.clubHomeRepo.SaveClubHomePlayers(uint(clubID), req); err != nil {
		utils.ServerError(ctx, "保存失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "保存成功")
}

// GetNews 获取最新动态（比赛 + 体测 + 手工公告）
func (c *ClubHomeController) GetNews(ctx *gin.Context) {
	clubIDStr := ctx.Param("clubId")
	clubID, err := strconv.ParseUint(clubIDStr, 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的俱乐部ID")
		return
	}

	// 获取最近比赛
	var matches []map[string]interface{}
	c.db.Table("match_summaries").
		Select("id, match_name as title, match_date as date, match_result as result, opponent, our_score, opponent_score").
		Where("club_id = ?", clubID).
		Order("match_date DESC").
		Limit(3).
		Scan(&matches)

	for _, m := range matches {
		m["type"] = "match"
	}

	// 获取最近体测
	var tests []map[string]interface{}
	c.db.Table("physical_test_activities").
		Select("id, name, start_date as date, status").
		Where("club_id = ?", clubID).
		Order("start_date DESC").
		Limit(2).
		Scan(&tests)

	for _, t := range tests {
		t["type"] = "physical_test"
	}

	utils.SuccessResponse(ctx, gin.H{
		"autoItems": gin.H{
			"matches": matches,
			"tests":   tests,
		},
		"manualItems": []map[string]interface{}{},
	})
}
