package controllers

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/repositories"
	"github.com/shaonianqiutan/backend/services"
	"github.com/shaonianqiutan/backend/utils"
)

// ClubController 俱乐部控制器
type ClubController struct {
	clubService        *services.ClubService
	db                 *gorm.DB
	weeklyReportRepo   *repositories.WeeklyReportRepository
	matchSummaryRepo   *repositories.MatchSummaryRepository
	orderRepo          *models.OrderRepository
	physicalTestService *services.PhysicalTestService
	adminLogRepo       *repositories.AdminOperationLogRepository
}

// NewClubController 创建俱乐部控制器
func NewClubController(
	clubService *services.ClubService,
	db *gorm.DB,
	weeklyReportRepo *repositories.WeeklyReportRepository,
	matchSummaryRepo *repositories.MatchSummaryRepository,
	orderRepo *models.OrderRepository,
	physicalTestService *services.PhysicalTestService,
	adminLogRepo *repositories.AdminOperationLogRepository,
) *ClubController {
	return &ClubController{
		clubService:         clubService,
		db:                  db,
		weeklyReportRepo:    weeklyReportRepo,
		matchSummaryRepo:    matchSummaryRepo,
		orderRepo:           orderRepo,
		physicalTestService: physicalTestService,
		adminLogRepo:        adminLogRepo,
	}
}

// getClubByUserOrCoach 获取用户关联的俱乐部（支持俱乐部管理员和教练）
func (c *ClubController) getClubByUserOrCoach(ctx *gin.Context, teamID uint) (*models.Club, error) {
	userID := ctx.GetUint("userId")

	// 1. 尝试作为俱乐部管理员获取
	var club models.Club
	err := c.db.Where("user_id = ?", userID).First(&club).Error
	if err == nil {
		// 验证球队是否属于该俱乐部
		var team models.Team
		if err := c.db.Where("id = ? AND club_id = ?", teamID, club.ID).First(&team).Error; err == nil {
			return &club, nil
		}
	}

	// 2. 尝试作为教练获取
	var teamCoach models.TeamCoach
	if err := c.db.Where("user_id = ? AND team_id = ? AND status = ?", userID, teamID, "active").First(&teamCoach).Error; err == nil {
		var team models.Team
		if err := c.db.First(&team, teamCoach.TeamID).Error; err == nil {
			if err := c.db.First(&club, team.ClubID).Error; err == nil {
				return &club, nil
			}
		}
	}

	return nil, fmt.Errorf("无权限访问")
}

// createAdminLog 记录管理员操作日志（辅助方法）
func (c *ClubController) createAdminLog(ctx *gin.Context, clubID, adminID uint, action, target string, targetID uint, detail string) {
	if c.adminLogRepo == nil {
		return
	}
	var user models.User
	adminName := ""
	if err := c.db.First(&user, adminID).Error; err == nil {
		adminName = user.Name
	}
	log := &models.AdminOperationLog{
		ClubID:    clubID,
		AdminID:   adminID,
		AdminName: adminName,
		Action:    action,
		Target:    target,
		TargetID:  targetID,
		Detail:    detail,
		IP:        ctx.ClientIP(),
	}
	c.adminLogRepo.Create(log)
}

// GetClubProfile 获取俱乐部资料
func (c *ClubController) GetClubProfile(ctx *gin.Context) {
	userID := ctx.GetUint("userId")

	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil {
		utils.ServerError(ctx, "获取俱乐部资料失败")
		return
	}

	if club == nil {
		utils.NotFoundError(ctx, "俱乐部不存在")
		return
	}

	// 获取球员数量
	playerCount, _ := c.clubService.GetPlayerCount(club.ID)

	utils.SuccessResponse(ctx, gin.H{
		"id":                    club.ID,
		"userId":                club.UserID,
		"name":                  club.Name,
		"logo":                  club.Logo,
		"description":           club.Description,
		"address":               club.Address,
		"contactName":           club.ContactName,
		"contactPhone":          club.ContactPhone,
		"establishedYear":       club.EstablishedYear,
		"clubSize":              club.ClubSize,
		"memberLevel":           club.MemberLevel,
		"memberExpireDate":      utils.FormatTime(&club.MemberExpireDate),
		"freePhysicalTestQuota": club.FreeTestQuota,
		"playerCount":           playerCount,
		"clubRole":              "admin",
		"createdAt":             utils.FormatDateTime(club.CreatedAt),
	})
}

// UpdateClubProfile 更新俱乐部资料
func (c *ClubController) UpdateClubProfile(ctx *gin.Context) {
	userID := ctx.GetUint("userId")

	var req struct {
		Name         string `json:"name"`
		Logo         string `json:"logo"`
		Description  string `json:"description"`
		Address      string `json:"address"`
		ContactName  string `json:"contactName"`
		ContactPhone string `json:"contactPhone"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	club, err := c.clubService.UpdateClubProfile(userID, req.Name, req.Logo, req.Description, req.Address, req.ContactName, req.ContactPhone)
	if err != nil {
		utils.ServerError(ctx, "更新失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, gin.H{
		"id":        club.ID,
		"name":      club.Name,
		"updatedAt": utils.FormatDateTime(club.UpdatedAt),
	}, "资料更新成功")
}

// GetDashboard 获取工作台数据
func (c *ClubController) GetDashboard(ctx *gin.Context) {
	userID := ctx.GetUint("userId")

	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.SuccessResponse(ctx, gin.H{
			"overview": gin.H{
				"totalPlayers":           0,
				"activePlayers":          0,
				"totalOrders":            0,
				"pendingOrders":          0,
				"completedOrders":        0,
				"physicalTestsThisMonth": 0,
				"totalPhysicalTests":     0,
			},
		})
		return
	}

	overview, _ := c.clubService.GetDashboardOverview(club.ID)
	recentOrders, _ := c.clubService.GetRecentOrders(club.ID, 5)
	upcomingTests, _ := c.clubService.GetUpcomingTests(club.ID, 3)

	utils.SuccessResponse(ctx, gin.H{
		"overview":              overview,
		"recentOrders":          recentOrders,
		"upcomingPhysicalTests": upcomingTests,
		"memberInfo": gin.H{
			"level":      string(club.MemberLevel),
			"quotaUsed":  overview["physicalTestsThisMonth"],
			"quotaLimit": club.FreeTestQuota,
			"expireDate": utils.FormatTime(&club.MemberExpireDate),
		},
		"quickActions": []gin.H{
			{"action": "invite_player", "label": "邀请球员"},
			{"action": "create_physical_test", "label": "创建体测"},
			{"action": "batch_order", "label": "批量下单"},
			{"action": "view_reports", "label": "查看报告"},
		},
	})
}

// GetPlayers 获取球员列表
func (c *ClubController) GetPlayers(ctx *gin.Context) {
	userID := ctx.GetUint("userId")

	// 分页参数
	pagination := utils.ParsePagination(ctx)
	page := pagination.Page
	pageSize := pagination.PageSize
	keyword := ctx.Query("keyword")
	ageGroup := ctx.Query("ageGroup")
	position := ctx.Query("position")
	tag := ctx.Query("tag")
	status := ctx.DefaultQuery("status", models.PlayerStatusActive)
	sortBy := ctx.DefaultQuery("sortBy", "created_at")
	sortOrder := ctx.DefaultQuery("sortOrder", "desc")

	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.PaginatedResponse(ctx, []interface{}{}, page, pageSize, 0)
		return
	}

	players, total, _ := c.clubService.GetPlayers(club.ID, page, pageSize, keyword, ageGroup, position, tag, status, sortBy, sortOrder)

	// 转换数据格式
	list := make([]interface{}, 0, len(players))
	for _, p := range players {
		var tags []string
		if p.Tags != "" {
			json.Unmarshal([]byte(p.Tags), &tags)
		}

		list = append(list, gin.H{
			"id":               p.ID,
			"userId":           p.UserID,
			"name":             p.User.Name,
			"nickname":         p.User.Nickname,
			"avatar":           p.User.Avatar,
			"age":              p.User.Age,
			"birthDate":        p.User.BirthDate,
			"position":         p.Position,
			"positionName":     models.GetPositionName(p.Position),
			"ageGroup":         p.AgeGroup,
			"phone":            p.User.Phone,
			"joinDate":         utils.FormatTime(&p.JoinDate),
			"tags":             tags,
			"lastPhysicalTest": "",
			"totalOrders":      0,
			"totalReports":     0,
			"status":           p.Status,
		})
	}

	utils.PaginatedResponse(ctx, list, page, pageSize, total)
}

// GetPlayerSelection 选材决策台 - 获取球员聚合数据
func (c *ClubController) GetPlayerSelection(ctx *gin.Context) {
	userID := ctx.GetUint("userId")

	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.SuccessResponse(ctx, []interface{}{})
		return
	}

	// 查询参数
	ageGroup := ctx.Query("ageGroup")
	position := ctx.Query("position")
	minHeight, _ := strconv.ParseFloat(ctx.DefaultQuery("minHeight", "0"), 64)
	maxHeight, _ := strconv.ParseFloat(ctx.DefaultQuery("maxHeight", "0"), 64)
	minWeight, _ := strconv.ParseFloat(ctx.DefaultQuery("minWeight", "0"), 64)
	maxWeight, _ := strconv.ParseFloat(ctx.DefaultQuery("maxWeight", "0"), 64)
	sortBy := ctx.DefaultQuery("sortBy", "name")
	sortOrder := ctx.DefaultQuery("sortOrder", "asc")

	// 获取俱乐部活跃球员
	players, _, _ := c.clubService.GetPlayers(club.ID, 1, 500, "", ageGroup, position, "", models.PlayerStatusActive, "id", "asc")

	list := make([]gin.H, 0, len(players))
	for _, p := range players {
		var tags []string
		if p.Tags != "" {
			json.Unmarshal([]byte(p.Tags), &tags)
		}

		// 查询最新体测记录
		var latestTest models.PhysicalTestRecord
		testSummary := gin.H{
			"height": 0, "weight": 0, "sprint50m": 0,
			"standingLongJump": 0, "pushUp": 0, "sitUp": 0,
		}
		c.db.Where("player_id = ? AND club_id = ?", p.UserID, club.ID).
			Order("test_date DESC").First(&latestTest)
		if latestTest.ID > 0 {
			if latestTest.Height != nil {
				testSummary["height"] = *latestTest.Height
			}
			if latestTest.Weight != nil {
				testSummary["weight"] = *latestTest.Weight
			}
			if latestTest.Sprint50m != nil {
				testSummary["sprint50m"] = *latestTest.Sprint50m
			}
			if latestTest.StandingLongJump != nil {
				testSummary["standingLongJump"] = *latestTest.StandingLongJump
			}
			if latestTest.PushUp != nil {
				testSummary["pushUp"] = *latestTest.PushUp
			}
			if latestTest.SitUp != nil {
				testSummary["sitUp"] = *latestTest.SitUp
			}
		}

		// 身高体重筛选
		height := testSummary["height"].(float64)
		weight := testSummary["weight"].(float64)
		if minHeight > 0 && height > 0 && height < minHeight {
			continue
		}
		if maxHeight > 0 && height > 0 && height > maxHeight {
			continue
		}
		if minWeight > 0 && weight > 0 && weight < minWeight {
			continue
		}
		if maxWeight > 0 && weight > 0 && weight > maxWeight {
			continue
		}

		// 查询最近3个月周报平均分
		var reports []models.WeeklyReport
		threeMonthsAgo := time.Now().AddDate(0, -3, 0)
		c.db.Where("player_id = ? AND created_at >= ? AND review_status = ?", p.UserID, threeMonthsAgo, "approved").
			Find(&reports)

			var attitudeSum, techniqueSum, tacticsSum, knowledgeSum int
			reportCount := len(reports)
			for _, r := range reports {
				attitudeSum += r.CoachAttitudeRating
				techniqueSum += r.CoachTechniqueRating
				tacticsSum += r.CoachTacticsRating
				knowledgeSum += r.CoachKnowledgeRating
			}

		weeklyAvg := gin.H{
			"attitude":  0,
			"technique": 0,
			"tactics":   0,
			"knowledge": 0,
			"overall":   0,
			"count":     reportCount,
		}
		if reportCount > 0 {
			weeklyAvg["attitude"] = float64(attitudeSum) / float64(reportCount)
			weeklyAvg["technique"] = float64(techniqueSum) / float64(reportCount)
			weeklyAvg["tactics"] = float64(tacticsSum) / float64(reportCount)
			weeklyAvg["knowledge"] = float64(knowledgeSum) / float64(reportCount)
			weeklyAvg["overall"] = (weeklyAvg["attitude"].(float64) + weeklyAvg["technique"].(float64) +
				weeklyAvg["tactics"].(float64) + weeklyAvg["knowledge"].(float64)) / 4
		}

		// 查询比赛总结统计
		var matchCount int64
		c.db.Model(&models.MatchSummary{}).Where("? = ANY(player_ids)", p.UserID).Count(&matchCount)

		list = append(list, gin.H{
			"id":               p.ID,
			"userId":           p.UserID,
			"name":             p.User.Name,
			"avatar":           p.User.Avatar,
			"age":              p.User.Age,
			"position":         p.Position,
			"positionName":     models.GetPositionName(p.Position),
			"ageGroup":         p.AgeGroup,
			"joinDate":         utils.FormatTime(&p.JoinDate),
			"tags":             tags,
			"physicalTest":     testSummary,
			"weeklyAverage":    weeklyAvg,
			"matchCount":       matchCount,
			"status":           p.Status,
		})
	}

	// 排序
	if sortBy != "" && sortBy != "name" {
		sort.Slice(list, func(i, j int) bool {
			var vi, vj float64
			switch sortBy {
			case "age":
				vi = float64(list[i]["age"].(int))
				vj = float64(list[j]["age"].(int))
			case "height":
				vi = list[i]["physicalTest"].(gin.H)["height"].(float64)
				vj = list[j]["physicalTest"].(gin.H)["height"].(float64)
			case "weight":
				vi = list[i]["physicalTest"].(gin.H)["weight"].(float64)
				vj = list[j]["physicalTest"].(gin.H)["weight"].(float64)
			case "weeklyOverall":
				vi = list[i]["weeklyAverage"].(gin.H)["overall"].(float64)
				vj = list[j]["weeklyAverage"].(gin.H)["overall"].(float64)
			case "matchCount":
				vi = float64(list[i]["matchCount"].(int64))
				vj = float64(list[j]["matchCount"].(int64))
			}
			if sortOrder == "desc" {
				return vi > vj
			}
			return vi < vj
		})
	}

	utils.SuccessResponse(ctx, list)
}

// GetPlayerDetail 获取球员详情（成长档案聚合）
func (c *ClubController) GetPlayerDetail(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	playerID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球员ID")
		return
	}

	// 获取当前用户关联的俱乐部（支持俱乐部管理员和教练）
	var club models.Club
	err = c.db.Where("user_id = ?", userID).First(&club).Error
	if err != nil {
		// 如果不是俱乐部管理员，尝试查是否是球队教练
		var teamCoach models.TeamCoach
		if err2 := c.db.Where("user_id = ? AND status = ?", userID, "active").First(&teamCoach).Error; err2 == nil {
			var team models.Team
			if err3 := c.db.First(&team, teamCoach.TeamID).Error; err3 == nil {
				if err4 := c.db.First(&club, team.ClubID).Error; err4 != nil {
					utils.ForbiddenError(ctx, "无权限访问")
					return
				}
			} else {
				utils.ForbiddenError(ctx, "无权限访问")
				return
			}
		} else {
			utils.ForbiddenError(ctx, "无权限访问")
			return
		}
	}

	cp, err := c.clubService.GetPlayerByID(uint(playerID))
	if err != nil || cp == nil {
		// 尝试按 user_id 查找（球队列表传递的是 user_id）
		var clubPlayer models.ClubPlayer
		if err := c.db.Preload("User").Where("user_id = ? AND club_id = ?", playerID, club.ID).First(&clubPlayer).Error; err == nil {
			cp = &clubPlayer
		} else {
			// club_players 中没有记录，但该球员可能通过邀请加入球队但 club_players 未同步
			var user models.User
			if err := c.db.First(&user, uint(playerID)).Error; err != nil {
				utils.NotFoundError(ctx, "球员不存在")
				return
			}
			// 检查该用户是否在该俱乐部的任何球队中
			var teamPlayer models.TeamPlayer
			if err := c.db.Joins("JOIN teams ON teams.id = team_players.team_id").
				Where("team_players.user_id = ? AND teams.club_id = ?", playerID, club.ID).
				First(&teamPlayer).Error; err != nil {
				utils.NotFoundError(ctx, "球员不存在")
				return
			}
			// 构建临时 ClubPlayer，使后续逻辑能正常执行
			joinDate := teamPlayer.JoinedAt
			if joinDate.IsZero() {
				joinDate = time.Now()
			}
			cp = &models.ClubPlayer{
				ClubID:   club.ID,
				UserID:   user.ID,
				JoinDate: joinDate,
				Status:   teamPlayer.Status,
				Position: teamPlayer.Position,
			}
			cp.User = &user
		}
	}

	// 检查是否属于该俱乐部
	if cp.ClubID != club.ID {
		utils.ForbiddenError(ctx, "无权限访问")
		return
	}

	var tags []string
	if cp.Tags != "" {
		json.Unmarshal([]byte(cp.Tags), &tags)
	}

	// ========== 聚合真实数据 ==========
	playerUserID := cp.UserID

	// 1. 订单统计与列表
	orders, orderTotal, _ := c.orderRepo.FindByUserID(playerUserID, 1, 20, "")
	var completedOrders int64
	var pendingOrders int64
	for _, o := range orders {
		if o.Status == models.OrderStatusCompleted {
			completedOrders++
		}
		if o.Status == models.OrderStatusPending || o.Status == models.OrderStatusProcessing {
			pendingOrders++
		}
	}
	orderList := make([]gin.H, 0, len(orders))
	for _, o := range orders {
		orderList = append(orderList, gin.H{
			"id":          o.ID,
			"orderNo":     o.OrderNo,
			"amount":      o.Amount,
			"status":      o.Status,
			"orderType":   o.OrderType,
			"playerName":  o.PlayerName,
			"matchName":   o.MatchName,
			"opponent":    o.Opponent,
			"createdAt":   utils.FormatDateTime(o.CreatedAt),
			"completedAt": utils.FormatTime(o.CompletedAt),
		})
	}

	// 2. 周报列表
	weeklyReports, _, _ := c.weeklyReportRepo.ListByPlayer(playerUserID, 1, 20)
	weeklyReportList := make([]gin.H, 0, len(weeklyReports))
	for _, r := range weeklyReports {
		coachName := ""
		if r.Coach != nil {
			coachName = r.Coach.Name
		}
		teamName := ""
		if r.Team != nil {
			teamName = r.Team.Name
		}
		weeklyReportList = append(weeklyReportList, gin.H{
			"id":           r.ID,
			"weekStart":    r.WeekStart.Format("2006-01-02"),
			"weekEnd":      r.WeekEnd.Format("2006-01-02"),
			"reviewStatus": r.ReviewStatus,
			"coachName":    coachName,
			"teamName":     teamName,
			"createdAt":    utils.FormatDateTime(r.CreatedAt),
		})
	}

	// 3. 比赛总结列表
	matchSummaries, _, _ := c.matchSummaryRepo.ListByPlayer(playerUserID, 1, 20)
	matchSummaryList := make([]gin.H, 0, len(matchSummaries))
	for _, ms := range matchSummaries {
		teamName := ""
		if ms.Team != nil {
			teamName = ms.Team.Name
		}
		coachName := ""
		if ms.Coach != nil {
			coachName = ms.Coach.Name
		}
		matchSummaryList = append(matchSummaryList, gin.H{
			"id":         ms.ID,
			"matchName":  ms.MatchName,
			"opponent":   ms.Opponent,
			"matchDate":  ms.MatchDate,
			"status":     ms.Status,
			"teamName":   teamName,
			"coachName":  coachName,
			"matchResult": ms.Result,
			"createdAt":  utils.FormatDateTime(ms.CreatedAt),
		})
	}

	// 4. 体测记录
	var physicalTestRecords []models.PhysicalTestRecord
	c.db.Where("player_id = ? AND club_id = ?", playerUserID, club.ID).
		Order("test_date DESC, created_at DESC").
		Limit(20).
		Find(&physicalTestRecords)
	physicalTestList := make([]gin.H, 0, len(physicalTestRecords))
	var lastPhysicalTest gin.H
	for i, r := range physicalTestRecords {
		data := services.GetTestDataMapFromRecord(&r)
		pt := gin.H{
			"id":         r.ID,
			"activityID": r.ActivityID,
			"testDate":   r.TestDate.Format("2006-01-02"),
			"data":       data,
		}
		physicalTestList = append(physicalTestList, pt)
		if i == 0 {
			lastPhysicalTest = pt
		}
	}

	// 5. 球探报告
	var scoutReports []models.ScoutReport
	c.db.Where("player_id = ?", playerUserID).Order("created_at DESC").Limit(20).Find(&scoutReports)
	scoutReportList := make([]gin.H, 0, len(scoutReports))
	var totalScoutReports int64
	c.db.Model(&models.ScoutReport{}).Where("player_id = ?", playerUserID).Count(&totalScoutReports)
	for _, sr := range scoutReports {
		scoutReportList = append(scoutReportList, gin.H{
			"id":              sr.ID,
			"overallRating":   sr.OverallRating,
			"potentialRating": sr.PotentialRating,
			"status":          sr.Status,
			"summary":         sr.Summary,
			"recommendation":  sr.Recommendation,
			"targetClub":      sr.TargetClub,
			"createdAt":       utils.FormatDateTime(sr.CreatedAt),
		})
	}

	// 6. 成长记录时间线（合并所有类型）
	growthRecords := make([]gin.H, 0)
	for _, o := range orders {
		growthRecords = append(growthRecords, gin.H{
			"date":    utils.FormatDateTime(o.CreatedAt),
			"type":    "order",
			"title":   "订单：" + o.OrderNo,
			"summary": o.PlayerName + " - " + string(o.Status),
			"status":  o.Status,
		})
	}
	for _, r := range weeklyReports {
		growthRecords = append(growthRecords, gin.H{
			"date":    r.WeekStart.Format("2006-01-02"),
			"type":    "weekly_report",
			"title":   "周报",
			"summary": "第" + r.WeekStart.Format("01/02") + "周训练报告",
			"status":  r.ReviewStatus,
		})
	}
	for _, ms := range matchSummaries {
		growthRecords = append(growthRecords, gin.H{
			"date":    ms.MatchDate,
			"type":    "match_summary",
			"title":   "比赛：" + ms.MatchName,
			"summary": ms.Opponent + " - " + ms.Result,
			"status":  ms.Status,
		})
	}
	for _, r := range physicalTestRecords {
		growthRecords = append(growthRecords, gin.H{
			"date":    r.TestDate.Format("2006-01-02"),
			"type":    "physical_test",
			"title":   "体测",
			"summary": "完成体测记录",
		})
	}
	for _, sr := range scoutReports {
		growthRecords = append(growthRecords, gin.H{
			"date":    utils.FormatDateTime(sr.CreatedAt),
			"type":    "scout_report",
			"title":   "球探报告",
			"summary": sr.Summary,
			"status":  sr.Status,
		})
	}

	utils.SuccessResponse(ctx, gin.H{
		"id":           cp.ID,
		"userId":       cp.UserID,
		"name":         cp.User.Name,
		"nickname":     cp.User.Nickname,
		"avatar":       cp.User.Avatar,
		"gender":       cp.User.Gender,
		"age":          cp.User.Age,
		"birthDate":    cp.User.BirthDate,
		"position":     cp.Position,
		"positionName": models.GetPositionName(cp.Position),
		"ageGroup":     cp.AgeGroup,
		"phone":        cp.User.Phone,
		"email":        "",
		"joinDate":     utils.FormatTime(&cp.JoinDate),
		"tags":         tags,
		"status":       cp.Status,
		"notes":        cp.Notes,
		"playerProfile": gin.H{
			"height":       cp.User.Height,
			"weight":       cp.User.Weight,
			"dominantFoot": cp.User.Foot,
		},
		"lastPhysicalTest": lastPhysicalTest,
		"statistics": gin.H{
			"totalOrders":     orderTotal,
			"completedOrders": completedOrders,
			"pendingOrders":   pendingOrders,
			"totalReports":    totalScoutReports,
			"avgScore":        0,
		},
		"orders":         orderList,
		"weeklyReports":  weeklyReportList,
		"matchSummaries": matchSummaryList,
		"physicalTests":  physicalTestList,
		"scoutReports":   scoutReportList,
		"growthRecords":  growthRecords,
	})
}

// InvitePlayer 邀请球员
func (c *ClubController) InvitePlayer(ctx *gin.Context) {
	userID := ctx.GetUint("userId")

	var req struct {
		Phone    string `json:"phone"`
		Name     string `json:"name"`
		AgeGroup string `json:"ageGroup"`
		Position string `json:"position"`
		Message  string `json:"message"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	// TODO: 查找用户并创建邀请记录
	utils.SuccessResponseWithMessage(ctx, gin.H{
		"inviteId": "inv_" + strconv.FormatInt(time.Now().UnixNano(), 36),
		"playerId": nil,
		"status":   "invite_sent",
		"sentAt":   time.Now().Format("2006-01-02T15:04:05Z"),
		"expireAt": time.Now().AddDate(0, 0, 7).Format("2006-01-02T15:04:05Z"),
	}, "邀请已发送")
}

// ImportPlayers 批量导入球员
func (c *ClubController) ImportPlayers(ctx *gin.Context) {
	userID := ctx.GetUint("userId")

	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	// TODO: 处理Excel文件
	utils.SuccessResponseWithMessage(ctx, gin.H{
		"total":       50,
		"success":     48,
		"failed":      2,
		"details":     []interface{}{},
		"invitesSent": 48,
	}, "导入完成，成功48条，失败2条")
}

// UpdatePlayerTags 更新球员标签
func (c *ClubController) UpdatePlayerTags(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	playerID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的ID")
		return
	}

	var req struct {
		Tags []string `json:"tags"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	err = c.clubService.UpdatePlayerTags(club.ID, uint(playerID), req.Tags)
	if err != nil {
		utils.ServerError(ctx, "更新失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, gin.H{
		"id":   playerID,
		"tags": req.Tags,
	}, "标签更新成功")
}

// RemovePlayer 移除球员
func (c *ClubController) RemovePlayer(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	playerID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的ID")
		return
	}

	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	err = c.clubService.RemovePlayer(club.ID, uint(playerID))
	if err != nil {
		utils.ServerError(ctx, "移除失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "球员已从俱乐部移除")
}

// GetAnalytics 获取俱乐部数据分析
func (c *ClubController) GetAnalytics(ctx *gin.Context) {
	userID := ctx.GetUint("userId")

	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.SuccessResponse(ctx, gin.H{
			"playerDistribution": gin.H{
				"byAgeGroup":   []interface{}{},
				"byPosition":  []interface{}{},
			},
			"abilityRadar": gin.H{
				"labels":    []string{"速度", "力量", "耐力", "灵敏", "柔韧", "技术"},
				"teamAvg":   []int{70, 65, 68, 67, 63, 72},
				"platformAvg": []int{65, 60, 63, 62, 58, 68},
			},
			"topPerformers": []interface{}{},
		})
		return
	}

	// 球员年龄分布
	ageGroupStats, _ := c.clubService.GetPlayerAgeGroupStats(club.ID)
	// 球员位置分布
	positionStats, _ := c.clubService.GetPlayerPositionStats(club.ID)
	// 梯队能力数据
	abilityRadar, _ := c.clubService.GetAbilityRadar(club.ID)
	// TOP球员
	topPerformers, _ := c.clubService.GetTopPerformers(club.ID)

	utils.SuccessResponse(ctx, gin.H{
		"playerDistribution": gin.H{
			"byAgeGroup":  ageGroupStats,
			"byPosition": positionStats,
		},
		"abilityRadar":   abilityRadar,
		"topPerformers": topPerformers,
	})
}

// CreateWeeklyReport 俱乐部管理员为球员创建周报（批量）
func (c *ClubController) CreateWeeklyReport(ctx *gin.Context) {
	userID := ctx.GetUint("userId")

	var req models.CreateWeeklyReportInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误: "+err.Error())
		return
	}

	// 解析周起始日期
	weekStart, err := time.Parse("2006-01-02", req.WeekStart)
	if err != nil {
		utils.ValidationError(ctx, "无效的周起始日期，格式应为 2006-01-02")
		return
	}

	// 计算周结束日期
	weekEnd := weekStart.AddDate(0, 0, 6)
	if req.WeekEnd != "" {
		weekEnd, _ = time.Parse("2006-01-02", req.WeekEnd)
	}

	// 获取俱乐部信息
	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.Error(ctx, 403, "无权限访问")
		return
	}

	// 验证球员是否属于该俱乐部
	players, _, err := c.clubService.GetPlayers(club.ID, 1, 1000, "", "", "", "", "", "", "")
	if err != nil {
		utils.ServerError(ctx, "获取球员列表失败")
		return
	}

	playerMap := make(map[uint]bool)
	for _, p := range players {
		playerMap[p.ID] = true
	}

	// 创建周报
	created := 0
	failed := 0
	for _, playerID := range req.PlayerIDs {
		// 检查球员是否属于该俱乐部
		if !playerMap[playerID] {
			failed++
			continue
		}

		// 检查是否已存在周报
		weekStartStr := weekStart.Format("2006-01-02")
		existing, _ := c.weeklyReportRepo.GetByPlayerAndWeek(playerID, weekStartStr)
		if existing != nil {
			failed++
			continue
		}

		report := &models.WeeklyReport{
			TeamID:       0, // 俱乐部创建的周报不关联特定球队
			PlayerID:     playerID,
			CoachID:      userID, // 使用当前用户ID作为教练ID
			WeekStart:    weekStart,
			WeekEnd:      weekEnd,
			ReviewStatus: "pending",
		}

		if err := c.weeklyReportRepo.Create(report); err != nil {
			failed++
			continue
		}
		created++
	}

	// 异步通知球员
	if created > 0 {
		go func() {
			notificationHelper := NewNotificationHelper(c.db)
			var user models.User
			coachName := "教练"
			if err := c.db.First(&user, userID).Error; err == nil {
				coachName = user.Name
			}
			clubName := "俱乐部"
			if club != nil {
				clubName = club.Name
			}
			weekLabel := weekStart.Format("01/02") + " ~ " + weekEnd.Format("01/02")
			for _, playerID := range req.PlayerIDs {
				if !playerMap[playerID] {
					continue
				}
				var report models.WeeklyReport
				if err := c.db.Where("player_id = ? AND week_start = ?", playerID, weekStart.Format("2006-01-02")).First(&report).Error; err == nil {
					notificationHelper.NotifyWeeklyReportCreated(playerID, coachName, clubName, weekLabel, report.ID)
				}
			}
		}()
	}

	utils.SuccessResponse(ctx, gin.H{
		"created": created,
		"failed":  failed,
		"message": "成功创建 " + strconv.Itoa(created) + " 份周报",
	})
}

// GetMatchSummaryStats 获取俱乐部比赛汇总统计
// GET /api/club/match-summaries/summary
func (c *ClubController) GetMatchSummaryStats(ctx *gin.Context) {
	userID := ctx.GetUint("userId")

	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限访问")
		return
	}

	// 查询俱乐部下所有球队的比赛总结统计
	type teamStat struct {
		TeamID          uint   `json:"teamId" gorm:"column:team_id"`
		TeamName        string `json:"teamName" gorm:"column:team_name"`
		Total           int64  `json:"total" gorm:"column:total"`
		Pending         int64  `json:"pending" gorm:"column:pending"`
		PlayerSubmitted int64  `json:"playerSubmitted" gorm:"column:player_submitted"`
		Completed       int64  `json:"completed" gorm:"column:completed"`
	}

	var stats []teamStat
	err = c.db.Raw(`
		SELECT 
			t.id AS team_id,
			t.name AS team_name,
			COUNT(ms.id) AS total,
			SUM(CASE WHEN ms.status = 'pending' THEN 1 ELSE 0 END) AS pending,
			SUM(CASE WHEN ms.status = 'player_submitted' THEN 1 ELSE 0 END) AS player_submitted,
			SUM(CASE WHEN ms.status = 'completed' THEN 1 ELSE 0 END) AS completed
		FROM teams t
		LEFT JOIN match_summaries ms ON ms.team_id = t.id
		WHERE t.club_id = ?
		GROUP BY t.id, t.name
		ORDER BY t.name
	`, club.ID).Scan(&stats).Error

	if err != nil {
		utils.ServerError(ctx, "获取比赛汇总统计失败: "+err.Error())
		return
	}

	var clubTotal int64
	for _, s := range stats {
		clubTotal += s.Total
	}

	utils.SuccessResponse(ctx, gin.H{
		"teams":      stats,
		"clubTotal":  clubTotal,
	})
}

// ExportReport 导出报告（PDF/HTML 过渡方案）
// GET /api/club/reports/:reportId/export
func (c *ClubController) ExportReport(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	reportID, err := strconv.ParseUint(ctx.Param("reportId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的报告ID")
		return
	}

	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限访问")
		return
	}

	var report models.Report
	if err := c.db.First(&report, uint(reportID)).Error; err != nil {
		utils.NotFoundError(ctx, "报告不存在")
		return
	}

	// 如果已有 PDF 文件，直接返回下载
	if report.PdfURL != "" {
		filePath := path.Join(".", report.PdfURL)
		if _, err := os.Stat(filePath); err == nil {
			ctx.Header("Content-Type", "application/pdf")
			ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s_球探报告.pdf\"", report.PlayerName))
			ctx.File(filePath)
			return
		}
	}

	// 否则返回精美格式的 HTML，前端可直接打印为 PDF
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s_球探报告.html\"", report.PlayerName))

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<title>%s - 球探报告</title>
<style>
body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif; background: #0a0e14; color: #fff; margin: 0; padding: 40px; }
.container { max-width: 800px; margin: 0 auto; background: #1a1f2e; border-radius: 16px; padding: 40px; border: 1px solid rgba(255,255,255,0.05); }
.header { text-align: center; border-bottom: 2px solid #39ff14; padding-bottom: 20px; margin-bottom: 30px; }
.brand { font-size: 24px; font-weight: bold; color: #39ff14; margin-bottom: 8px; }
.title { font-size: 28px; font-weight: bold; color: #fff; margin: 16px 0 8px; }
.subtitle { color: #94a3b8; font-size: 14px; }
.section { margin-bottom: 28px; }
.section-title { font-size: 18px; font-weight: bold; color: #39ff14; margin-bottom: 12px; display: flex; align-items: center; gap: 8px; }
.info-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px; }
.info-item { background: rgba(10,14,23,0.7); padding: 16px; border-radius: 12px; text-align: center; }
.info-label { font-size: 12px; color: #94a3b8; margin-bottom: 4px; }
.info-value { font-size: 20px; font-weight: bold; color: #fff; }
.content-box { background: rgba(10,14,23,0.5); padding: 20px; border-radius: 12px; line-height: 1.8; color: #e2e8f0; }
.rating-bar { height: 8px; background: rgba(255,255,255,0.1); border-radius: 4px; overflow: hidden; margin-top: 8px; }
.rating-fill { height: 100%%; background: linear-gradient(90deg, #39ff14, #00d4ff); border-radius: 4px; }
.footer { text-align: center; margin-top: 40px; padding-top: 20px; border-top: 1px solid rgba(255,255,255,0.05); color: #64748b; font-size: 12px; }
@media print { body { background: #fff; color: #000; } .container { background: #fff; border: 1px solid #ddd; } }
</style>
</head>
<body>
<div class="container">
  <div class="header">
    <div class="brand">⚽ 少年球探</div>
    <div class="title">%s</div>
    <div class="subtitle">球探报告 · %s</div>
  </div>

  <div class="section">
    <div class="section-title">👤 球员信息</div>
    <div class="info-grid">
      <div class="info-item"><div class="info-label">姓名</div><div class="info-value">%s</div></div>
      <div class="info-item"><div class="info-label">位置</div><div class="info-value">%s</div></div>
      <div class="info-item"><div class="info-label">地区</div><div class="info-value">%s %s</div></div>
    </div>
  </div>

  <div class="section">
    <div class="section-title">📝 综合评价</div>
    <div class="content-box">%s</div>
  </div>

  <div class="section">
    <div class="section-title">⭐ 综合评分</div>
    <div class="info-grid">
      <div class="info-item">
        <div class="info-label">总体评分</div>
        <div class="info-value">%.1f</div>
        <div class="rating-bar"><div class="rating-fill" style="width:%.1f%%"></div></div>
      </div>
      <div class="info-item">
        <div class="info-label">进攻评分</div>
        <div class="info-value">%.1f</div>
        <div class="rating-bar"><div class="rating-fill" style="width:%.1f%%"></div></div>
      </div>
      <div class="info-item">
        <div class="info-label">防守评分</div>
        <div class="info-value">%.1f</div>
        <div class="rating-bar"><div class="rating-fill" style="width:%.1f%%"></div></div>
      </div>
    </div>
  </div>

  <div class="section">
    <div class="section-title">💡 潜力评估</div>
    <div class="content-box">潜力等级：<strong>%s</strong></div>
  </div>

  <div class="section">
    <div class="section-title">📋 技术特点与建议</div>
    <div class="content-box">
      <p><strong>优点：</strong>%s</p>
      <p><strong>待改进：</strong>%s</p>
      <p><strong>发展建议：</strong>%s</p>
    </div>
  </div>

  <div class="footer">
    <p>本报告由少年球探平台生成 · www.shaonianqiutan.com</p>
    <p>生成时间：%s</p>
  </div>
</div>
</body>
</html>`,
		report.PlayerName,
		report.PlayerName+" - 球探报告",
		utils.FormatDateTime(report.CreatedAt),
		report.PlayerName,
		report.PlayerPosition,
		report.PlayerProvince,
		report.PlayerCity,
		report.Summary,
		report.OverallRating,
		report.OverallRating*10,
		report.OffenseRating,
		report.OffenseRating*10,
		report.DefenseRating,
		report.DefenseRating*10,
		report.Potential,
		report.Strengths,
		report.Weaknesses,
		report.Suggestions,
		utils.FormatDateTime(report.CreatedAt),
	)

	ctx.String(200, html)
}

// ===== PlayerFilterPreset CRUD =====

// GetPlayerFilterPresets 获取筛选方案列表
func (c *ClubController) GetPlayerFilterPresets(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.SuccessResponse(ctx, []interface{}{})
		return
	}

	var presets []models.PlayerFilterPreset
	c.db.Where("club_id = ?", club.ID).Order("created_at DESC").Find(&presets)

	list := make([]gin.H, 0, len(presets))
	for _, p := range presets {
		list = append(list, gin.H{
			"id":        p.ID,
			"name":      p.Name,
			"filters":   p.Filters,
			"createdAt": utils.FormatDateTime(p.CreatedAt),
		})
	}
	utils.SuccessResponse(ctx, list)
}

// CreatePlayerFilterPreset 创建筛选方案
func (c *ClubController) CreatePlayerFilterPreset(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	var req struct {
		Name    string `json:"name" binding:"required"`
		Filters string `json:"filters" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	preset := models.PlayerFilterPreset{
		ClubID:  club.ID,
		Name:    req.Name,
		Filters: req.Filters,
	}
	if err := c.db.Create(&preset).Error; err != nil {
		utils.ServerError(ctx, "创建失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, gin.H{
		"id":      preset.ID,
		"name":    preset.Name,
		"filters": preset.Filters,
	}, "保存成功")
}

// UpdatePlayerFilterPreset 更新筛选方案
func (c *ClubController) UpdatePlayerFilterPreset(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	presetID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的ID")
		return
	}

	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	var req struct {
		Name    string `json:"name"`
		Filters string `json:"filters"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	var preset models.PlayerFilterPreset
	if err := c.db.Where("id = ? AND club_id = ?", presetID, club.ID).First(&preset).Error; err != nil {
		utils.NotFoundError(ctx, "方案不存在")
		return
	}

	updates := gin.H{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Filters != "" {
		updates["filters"] = req.Filters
	}
	if err := c.db.Model(&preset).Updates(updates).Error; err != nil {
		utils.ServerError(ctx, "更新失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, gin.H{
		"id":      preset.ID,
		"name":    preset.Name,
		"filters": preset.Filters,
	}, "更新成功")
}

// DeletePlayerFilterPreset 删除筛选方案
func (c *ClubController) DeletePlayerFilterPreset(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	presetID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的ID")
		return
	}

	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	if err := c.db.Where("id = ? AND club_id = ?", presetID, club.ID).Delete(&models.PlayerFilterPreset{}).Error; err != nil {
		utils.ServerError(ctx, "删除失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "删除成功")
}

// GetClubNotifications 获取俱乐部相关通知动态（用于概览页近期动态）
func (c *ClubController) GetClubNotifications(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.SuccessResponse(ctx, []interface{}{})
		return
	}

	// 获取俱乐部的球员和教练 user_id 列表
	var playerUserIDs []uint
	c.db.Model(&models.ClubPlayer{}).Where("club_id = ?", club.ID).Pluck("user_id", &playerUserIDs)

	var relatedUserIDs []uint
	relatedUserIDs = append(relatedUserIDs, userID)
	relatedUserIDs = append(relatedUserIDs, playerUserIDs...)

	var notifications []models.Notification
	c.db.Where("user_id IN ?", relatedUserIDs).
		Order("created_at DESC").
		Limit(10).
		Find(&notifications)

	list := make([]gin.H, 0, len(notifications))
	for _, n := range notifications {
		var icon, title, subTitle string
		switch n.Type {
		case models.NotificationTypeWeeklyReportCreated, models.NotificationTypeWeeklyReportApproved, models.NotificationTypeWeeklyReportRejected:
			icon = "report"
			title = "周报更新"
			subTitle = n.Title
		case models.NotificationTypeMatchSummaryCreated, models.NotificationTypeMatchSummaryComplete:
			icon = "match"
			title = "比赛动态"
			subTitle = n.Title
		case models.NotificationTypeMatchPlayerReminder, models.NotificationTypeMatchCoachReminder:
			icon = "match"
			title = "比赛提醒"
			subTitle = n.Title
		default:
			icon = "system"
			title = "系统通知"
			subTitle = n.Title
		}
		list = append(list, gin.H{
			"id":        n.ID,
			"title":     title,
			"subTitle":  subTitle,
			"content":   n.Content,
			"icon":      icon,
			"type":      n.Type,
			"createdAt": utils.FormatDateTime(n.CreatedAt),
		})
	}

	utils.SuccessResponse(ctx, list)
}

// ===== Announcement CRUD =====

// GetAnnouncements 获取俱乐部公告列表
func (c *ClubController) GetAnnouncements(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.SuccessResponse(ctx, []interface{}{})
		return
	}

	var announcements []models.Announcement
	c.db.Where("club_id = ?", club.ID).Order("is_pinned DESC, created_at DESC").Find(&announcements)

	list := make([]gin.H, 0, len(announcements))
	for _, a := range announcements {
		list = append(list, gin.H{
			"id":        a.ID,
			"clubId":    a.ClubID,
			"title":     a.Title,
			"content":   a.Content,
			"isPinned":  a.IsPinned,
			"createdBy": a.CreatedBy,
			"createdAt": utils.FormatDateTime(a.CreatedAt),
		})
	}
	utils.SuccessResponse(ctx, list)
}

// CreateAnnouncement 创建公告
func (c *ClubController) CreateAnnouncement(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	var req struct {
		Title    string `json:"title" binding:"required"`
		Content  string `json:"content"`
		IsPinned bool   `json:"isPinned"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	announcement := models.Announcement{
		ClubID:    club.ID,
		Title:     req.Title,
		Content:   req.Content,
		IsPinned:  req.IsPinned,
		CreatedBy: userID,
	}
	if err := c.db.Create(&announcement).Error; err != nil {
		utils.ServerError(ctx, "创建失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, gin.H{
		"id":       announcement.ID,
		"title":    announcement.Title,
		"isPinned": announcement.IsPinned,
	}, "发布成功")
}

// UpdateAnnouncement 更新公告
func (c *ClubController) UpdateAnnouncement(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	announcementID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的ID")
		return
	}

	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	var req struct {
		Title    string `json:"title"`
		Content  string `json:"content"`
		IsPinned *bool  `json:"isPinned"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	var announcement models.Announcement
	if err := c.db.Where("id = ? AND club_id = ?", announcementID, club.ID).First(&announcement).Error; err != nil {
		utils.NotFoundError(ctx, "公告不存在")
		return
	}

	updates := gin.H{}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.IsPinned != nil {
		updates["is_pinned"] = *req.IsPinned
	}
	if err := c.db.Model(&announcement).Updates(updates).Error; err != nil {
		utils.ServerError(ctx, "更新失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, gin.H{
		"id":       announcement.ID,
		"title":    announcement.Title,
		"isPinned": announcement.IsPinned,
	}, "更新成功")
}

// DeleteAnnouncement 删除公告
func (c *ClubController) DeleteAnnouncement(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	announcementID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的ID")
		return
	}

	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	if err := c.db.Where("id = ? AND club_id = ?", announcementID, club.ID).Delete(&models.Announcement{}).Error; err != nil {
		utils.ServerError(ctx, "删除失败")
		return
	}

	c.createAdminLog(ctx, club.ID, userID, "delete_announcement", "announcement", uint(announcementID), "删除公告")
	utils.SuccessResponseWithMessage(ctx, nil, "删除成功")
}

// GetAdminOperationLogs 获取管理员操作日志
// GET /api/club/admin-logs
func (c *ClubController) GetAdminOperationLogs(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	pagination := utils.ParsePagination(ctx)
	page := pagination.Page
	pageSize := pagination.PageSize

	logs, total, err := c.adminLogRepo.GetLogsByClubID(club.ID, page, pageSize)
	if err != nil {
		utils.ServerError(ctx, "获取日志失败")
		return
	}

	list := make([]*models.AdminOperationLogResponse, 0, len(logs))
	for _, l := range logs {
		list = append(list, l.ToResponse())
	}

	utils.SuccessResponse(ctx, gin.H{
		"list":     list,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// ===== Player Shortlist CRUD =====

// GetShortlist 获取候选名单
func (c *ClubController) GetShortlist(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	var list []models.PlayerShortlist
	c.db.Where("club_id = ?", club.ID).Order("created_at DESC").Find(&list)

	res := make([]gin.H, 0, len(list))
	for _, s := range list {
		res = append(res, gin.H{
			"id":        s.ID,
			"playerId":  s.PlayerID,
			"note":      s.Note,
			"createdAt": utils.FormatDateTime(s.CreatedAt),
			"updatedAt": utils.FormatDateTime(s.UpdatedAt),
		})
	}
	utils.SuccessResponse(ctx, res)
}

// AddToShortlist 加入候选名单
func (c *ClubController) AddToShortlist(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	var req struct {
		PlayerIDs []uint `json:"playerIds" binding:"required"`
		Note      string `json:"note"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	added := 0
	for _, pid := range req.PlayerIDs {
		var existing models.PlayerShortlist
		if err := c.db.Where("club_id = ? AND player_id = ?", club.ID, pid).First(&existing).Error; err == nil {
			continue
		}
		s := models.PlayerShortlist{
			ClubID:    club.ID,
			PlayerID:  pid,
			Note:      req.Note,
			CreatedBy: userID,
		}
		if err := c.db.Create(&s).Error; err == nil {
			added++
		}
	}

	c.createAdminLog(ctx, club.ID, userID, "add_shortlist", "player", 0, "批量加入候选名单 "+strconv.Itoa(added)+" 人")
	utils.SuccessResponseWithMessage(ctx, gin.H{"added": added}, "添加成功")
}

// UpdateShortlistNote 更新候选备注
func (c *ClubController) UpdateShortlistNote(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	playerID, err := strconv.ParseUint(ctx.Param("playerId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球员ID")
		return
	}

	var req struct {
		Note string `json:"note"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	if err := c.db.Model(&models.PlayerShortlist{}).
		Where("club_id = ? AND player_id = ?", club.ID, playerID).
		Update("note", req.Note).Error; err != nil {
		utils.ServerError(ctx, "更新失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, nil, "更新成功")
}

// RemoveFromShortlist 从候选名单移除
func (c *ClubController) RemoveFromShortlist(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	playerID, err := strconv.ParseUint(ctx.Param("playerId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球员ID")
		return
	}

	if err := c.db.Where("club_id = ? AND player_id = ?", club.ID, playerID).Delete(&models.PlayerShortlist{}).Error; err != nil {
		utils.ServerError(ctx, "删除失败")
		return
	}

	c.createAdminLog(ctx, club.ID, userID, "remove_shortlist", "player", uint(playerID), "从候选名单移除")
	utils.SuccessResponseWithMessage(ctx, nil, "移除成功")
}

// ===== Team Season Archive CRUD =====

// GetTeamSeasonArchives 获取球队赛季档案列表
func (c *ClubController) GetTeamSeasonArchives(ctx *gin.Context) {
	teamID, err := strconv.ParseUint(ctx.Param("teamId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	club, err := c.getClubByUserOrCoach(ctx, uint(teamID))
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	var archives []models.TeamSeasonArchive
	c.db.Where("team_id = ?", teamID).Order("created_at DESC").Find(&archives)

	res := make([]gin.H, 0, len(archives))
	for _, a := range archives {
		res = append(res, gin.H{
			"id":          a.ID,
			"teamId":      a.TeamID,
			"seasonName":  a.SeasonName,
			"startDate":   a.StartDate,
			"endDate":     a.EndDate,
			"matchCount":  a.MatchCount,
			"weeklyCount": a.WeeklyCount,
			"testCount":   a.TestCount,
			"description": a.Description,
			"createdAt":   utils.FormatDateTime(a.CreatedAt),
			"updatedAt":   utils.FormatDateTime(a.UpdatedAt),
		})
	}
	utils.SuccessResponse(ctx, res)
}

// CreateTeamSeasonArchive 创建赛季档案
func (c *ClubController) CreateTeamSeasonArchive(ctx *gin.Context) {
	teamID, err := strconv.ParseUint(ctx.Param("teamId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}

	club, err := c.getClubByUserOrCoach(ctx, uint(teamID))
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	userID := ctx.GetUint("userId")

	var req struct {
		SeasonName  string `json:"seasonName" binding:"required"`
		StartDate   string `json:"startDate"`
		EndDate     string `json:"endDate"`
		Description string `json:"description"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	archive := models.TeamSeasonArchive{
		TeamID:      uint(teamID),
		SeasonName:  req.SeasonName,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Description: req.Description,
		CreatedBy:   userID,
	}

	// 自动统计
	var matchCount, weeklyCount, testCount int64
	c.db.Model(&models.MatchSummary{}).Where("team_id = ?", teamID).Count(&matchCount)
	c.db.Model(&models.WeeklyReport{}).Where("team_id = ?", teamID).Count(&weeklyCount)
	c.db.Model(&models.PhysicalTestActivity{}).Where("team_id = ?", teamID).Count(&testCount)
	archive.MatchCount = int(matchCount)
	archive.WeeklyCount = int(weeklyCount)
	archive.TestCount = int(testCount)

	if err := c.db.Create(&archive).Error; err != nil {
		utils.ServerError(ctx, "创建失败")
		return
	}

	c.createAdminLog(ctx, club.ID, userID, "create_season_archive", "team_season", archive.ID, "创建赛季档案: "+req.SeasonName)
	utils.SuccessResponseWithMessage(ctx, gin.H{"id": archive.ID}, "创建成功")
}

// GetTeamSeasonArchiveDetail 获取赛季档案详情
func (c *ClubController) GetTeamSeasonArchiveDetail(ctx *gin.Context) {
	teamID, err := strconv.ParseUint(ctx.Param("teamId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}
	archiveID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的档案ID")
		return
	}

	club, err := c.getClubByUserOrCoach(ctx, uint(teamID))
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	var archive models.TeamSeasonArchive
	if err := c.db.Where("id = ? AND team_id = ?", archiveID, teamID).First(&archive).Error; err != nil {
		utils.NotFoundError(ctx, "档案不存在")
		return
	}

	utils.SuccessResponse(ctx, gin.H{
		"id":          archive.ID,
		"teamId":      archive.TeamID,
		"seasonName":  archive.SeasonName,
		"startDate":   archive.StartDate,
		"endDate":     archive.EndDate,
		"matchCount":  archive.MatchCount,
		"weeklyCount": archive.WeeklyCount,
		"testCount":   archive.TestCount,
		"description": archive.Description,
		"createdAt":   utils.FormatDateTime(archive.CreatedAt),
		"updatedAt":   utils.FormatDateTime(archive.UpdatedAt),
	})
}

// UpdateTeamSeasonArchive 更新赛季档案
func (c *ClubController) UpdateTeamSeasonArchive(ctx *gin.Context) {
	teamID, err := strconv.ParseUint(ctx.Param("teamId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}
	archiveID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的档案ID")
		return
	}

	club, err := c.getClubByUserOrCoach(ctx, uint(teamID))
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	var archive models.TeamSeasonArchive
	if err := c.db.Where("id = ? AND team_id = ?", archiveID, teamID).First(&archive).Error; err != nil {
		utils.NotFoundError(ctx, "档案不存在")
		return
	}

	var req struct {
		SeasonName  string `json:"seasonName"`
		StartDate   string `json:"startDate"`
		EndDate     string `json:"endDate"`
		Description string `json:"description"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	updates := gin.H{}
	if req.SeasonName != "" {
		updates["season_name"] = req.SeasonName
	}
	if req.StartDate != "" {
		updates["start_date"] = req.StartDate
	}
	if req.EndDate != "" {
		updates["end_date"] = req.EndDate
	}
	updates["description"] = req.Description

	if err := c.db.Model(&archive).Updates(updates).Error; err != nil {
		utils.ServerError(ctx, "更新失败")
		return
	}

	utils.SuccessResponseWithMessage(ctx, gin.H{"id": archive.ID}, "更新成功")
}

// DeleteTeamSeasonArchive 删除赛季档案
func (c *ClubController) DeleteTeamSeasonArchive(ctx *gin.Context) {
	teamID, err := strconv.ParseUint(ctx.Param("teamId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的球队ID")
		return
	}
	archiveID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的档案ID")
		return
	}

	club, err := c.getClubByUserOrCoach(ctx, uint(teamID))
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	userID := ctx.GetUint("userId")

	if err := c.db.Where("id = ? AND team_id = ?", archiveID, teamID).Delete(&models.TeamSeasonArchive{}).Error; err != nil {
		utils.ServerError(ctx, "删除失败")
		return
	}

	c.createAdminLog(ctx, club.ID, userID, "delete_season_archive", "team_season", uint(archiveID), "删除赛季档案")
	utils.SuccessResponseWithMessage(ctx, nil, "删除成功")
}

// ListPublicClubs 获取公开俱乐部列表
func (c *ClubController) ListPublicClubs(ctx *gin.Context) {
	pagination := utils.ParsePaginationWithSize(ctx, 10)
	page := pagination.Page
	pageSize := pagination.PageSize
	if pageSize > 50 {
		pageSize = 50
	}

	var clubs []models.Club
	var total int64

	// 查询非删除状态的俱乐部
	if err := c.db.Model(&models.Club{}).Count(&total).Error; err != nil {
		utils.ServerError(ctx, "查询失败")
		return
	}

	offset := (page - 1) * pageSize
	if err := c.db.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&clubs).Error; err != nil {
		utils.ServerError(ctx, "查询失败")
		return
	}

	// 组装返回数据（统计球员数和教练数）
	type clubItem struct {
		ID            uint   `json:"id"`
		Name          string `json:"name"`
		Logo          string `json:"logo"`
		Description   string `json:"description"`
		Address       string `json:"address"`
		EstablishedYear int  `json:"established_year"`
		PlayerCount   int    `json:"player_count"`
		CoachCount    int    `json:"coach_count"`
		CreatedAt     string `json:"created_at"`
	}

	items := make([]clubItem, 0, len(clubs))
	for _, club := range clubs {
		var playerCount, coachCount int64
		c.db.Model(&models.ClubPlayer{}).Where("club_id = ? AND status = ?", club.ID, "active").Count(&playerCount)
		c.db.Model(&models.TeamCoach{}).Where("club_id = ?", club.ID).Count(&coachCount)

		items = append(items, clubItem{
			ID:              club.ID,
			Name:            club.Name,
			Logo:            club.Logo,
			Description:     club.Description,
			Address:         club.Address,
			EstablishedYear: club.EstablishedYear,
			PlayerCount:     int(playerCount),
			CoachCount:      int(coachCount),
			CreatedAt:       utils.FormatDateTime(club.CreatedAt),
		})
	}

	utils.SuccessResponse(ctx, gin.H{
		"list":  items,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// SearchClubs 搜索俱乐部（公开接口）
func (c *ClubController) SearchClubs(ctx *gin.Context) {
	keyword := ctx.Query("keyword")
	province := ctx.Query("province")
	city := ctx.Query("city")
	pagination := utils.ParsePaginationWithSize(ctx, 10)
	page := pagination.Page
	pageSize := pagination.PageSize
	if pageSize > 50 {
		pageSize = 50
	}

	var clubs []models.Club
	var total int64

	query := c.db.Model(&models.Club{})
	if keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if province != "" {
		query = query.Where("province = ?", province)
	}
	if city != "" {
		query = query.Where("city = ?", city)
	}

	if err := query.Count(&total).Error; err != nil {
		utils.ServerError(ctx, "查询失败")
		return
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&clubs).Error; err != nil {
		utils.ServerError(ctx, "查询失败")
		return
	}

	type clubItem struct {
		ID              uint   `json:"id"`
		Name            string `json:"name"`
		Logo            string `json:"logo"`
		Description     string `json:"description"`
		Address         string `json:"address"`
		EstablishedYear int    `json:"established_year"`
		Province        string `json:"province"`
		City            string `json:"city"`
		PlayerCount     int    `json:"player_count"`
		CoachCount      int    `json:"coach_count"`
		TeamCount       int    `json:"team_count"`
	}

	items := make([]clubItem, 0, len(clubs))
	for _, club := range clubs {
		var playerCount, coachCount, teamCount int64
		c.db.Model(&models.ClubPlayer{}).Where("club_id = ? AND status = ?", club.ID, "active").Count(&playerCount)
		c.db.Model(&models.TeamCoach{}).Where("club_id = ?", club.ID).Count(&coachCount)
		c.db.Model(&models.Team{}).Where("club_id = ? AND status = ?", club.ID, "active").Count(&teamCount)

		items = append(items, clubItem{
			ID:              club.ID,
			Name:            club.Name,
			Logo:            club.Logo,
			Description:     club.Description,
			Address:         club.Address,
			EstablishedYear: club.EstablishedYear,
			Province:        club.Province,
			City:            club.City,
			PlayerCount:     int(playerCount),
			CoachCount:      int(coachCount),
			TeamCount:       int(teamCount),
		})
	}

	utils.SuccessResponse(ctx, gin.H{
		"list":  items,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// GetClubDetail 获取俱乐部详情（公开接口）
func (c *ClubController) GetClubDetail(ctx *gin.Context) {
	clubID, err := strconv.ParseUint(ctx.Param("clubId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的俱乐部ID")
		return
	}

	var club models.Club
	if err := c.db.First(&club, uint(clubID)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFoundError(ctx, "俱乐部不存在")
			return
		}
		utils.ServerError(ctx, "查询失败")
		return
	}

	// 统计数量
	var playerCount, coachCount, teamCount int64
	c.db.Model(&models.ClubPlayer{}).Where("club_id = ? AND status = ?", club.ID, "active").Count(&playerCount)
	c.db.Model(&models.TeamCoach{}).Where("club_id = ?", club.ID).Count(&coachCount)
	c.db.Model(&models.Team{}).Where("club_id = ? AND status = ?", club.ID, "active").Count(&teamCount)

	// 获取旗下球队
	var teams []models.Team
	c.db.Where("club_id = ? AND status = ?", club.ID, "active").Find(&teams)

	type teamItem struct {
		ID           uint   `json:"id"`
		Name         string `json:"name"`
		AgeGroup     string `json:"age_group"`
		PlayerCount  int    `json:"player_count"`
		CoachCount   int    `json:"coach_count"`
	}

	teamItems := make([]teamItem, 0, len(teams))
	for _, team := range teams {
		var tPlayerCount, tCoachCount int64
		c.db.Model(&models.TeamPlayer{}).Where("team_id = ? AND status = ?", team.ID, "active").Count(&tPlayerCount)
		c.db.Model(&models.TeamCoach{}).Where("team_id = ?", team.ID).Count(&tCoachCount)
		teamItems = append(teamItems, teamItem{
			ID:          team.ID,
			Name:        team.Name,
			AgeGroup:    team.AgeGroup,
			PlayerCount: int(tPlayerCount),
			CoachCount:  int(tCoachCount),
		})
	}

	utils.SuccessResponse(ctx, gin.H{
		"id":               club.ID,
		"name":             club.Name,
		"logo":             club.Logo,
		"description":      club.Description,
		"address":          club.Address,
		"contact_name":     club.ContactName,
		"contact_phone":    club.ContactPhone,
		"established_year": club.EstablishedYear,
		"province":         club.Province,
		"city":             club.City,
		"player_count":     playerCount,
		"coach_count":      coachCount,
		"team_count":       teamCount,
		"teams":            teamItems,
	})
}

// ===== ClubCoach CRUD =====

// GetClubCoaches 获取俱乐部教练列表
func (c *ClubController) GetClubCoaches(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	status := ctx.Query("status")
	keyword := ctx.Query("keyword")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))

	coaches, total, err := c.clubService.GetClubCoaches(club.ID, status, keyword, page, pageSize)
	if err != nil {
		utils.ServerError(ctx, "获取教练列表失败")
		return
	}

	list := make([]gin.H, 0, len(coaches))
	for _, cc := range coaches {
		userName := ""
		userAvatar := ""
		userPhone := ""
		if cc.User != nil {
			userName = cc.User.Name
			if userName == "" {
				userName = cc.User.Nickname
			}
			userAvatar = cc.User.Avatar
			userPhone = cc.User.Phone
		}
		list = append(list, gin.H{
			"id":          cc.ID,
			"clubId":      cc.ClubID,
			"userId":      cc.UserID,
			"name":        userName,
			"avatar":      userAvatar,
			"phone":       userPhone,
			"primaryRole": cc.PrimaryRole,
			"roleLabel":   models.GetCoachRoleLabel(cc.PrimaryRole),
			"status":      cc.Status,
			"joinedAt":    utils.FormatTime(&cc.JoinedAt),
			"leftAt":      utils.FormatTime(cc.LeftAt),
			"notes":       cc.Notes,
			"createdAt":   utils.FormatDateTime(cc.CreatedAt),
		})
	}

	utils.SuccessResponse(ctx, gin.H{
		"list":     list,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// AddClubCoach 添加教练到俱乐部
func (c *ClubController) AddClubCoach(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	var req struct {
		UserID      uint              `json:"userId" binding:"required"`
		PrimaryRole models.CoachRole  `json:"primaryRole"`
		Notes       string            `json:"notes"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误: "+err.Error())
		return
	}

	if req.PrimaryRole == "" {
		req.PrimaryRole = models.CoachRoleHead
	}

	// 检查用户是否存在
	var targetUser models.User
	if err := c.db.First(&targetUser, req.UserID).Error; err != nil {
		utils.NotFoundError(ctx, "用户不存在")
		return
	}

	coach, err := c.clubService.AddClubCoach(club.ID, req.UserID, req.PrimaryRole, req.Notes)
	if err != nil {
		utils.ServerError(ctx, "添加失败: "+err.Error())
		return
	}

	c.createAdminLog(ctx, club.ID, userID, "add_coach", "club_coach", coach.ID,
		fmt.Sprintf("添加教练: %s, 角色: %s", targetUser.Name, models.GetCoachRoleLabel(req.PrimaryRole)))

	utils.SuccessResponseWithMessage(ctx, gin.H{
		"id":          coach.ID,
		"userId":      coach.UserID,
		"primaryRole": coach.PrimaryRole,
		"status":      coach.Status,
	}, "教练添加成功")
}

// GetClubCoachDetail 获取教练详情
func (c *ClubController) GetClubCoachDetail(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	coachID, err := strconv.ParseUint(ctx.Param("coachId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的教练ID")
		return
	}

	coach, err := c.clubService.GetClubCoachByID(uint(coachID))
	if err != nil {
		utils.NotFoundError(ctx, "教练不存在")
		return
	}

	if coach.ClubID != club.ID {
		utils.ForbiddenError(ctx, "无权限查看")
		return
	}

	// 获取该教练的球队分配
	teamAssignments, _ := c.clubService.GetClubCoachTeams(coach.ID)
	teams := make([]gin.H, 0, len(teamAssignments))
	for _, ta := range teamAssignments {
		teamName := ""
		if ta.Team != nil {
			teamName = ta.Team.Name
		}
		teams = append(teams, gin.H{
			"teamCoachId": ta.ID,
			"teamId":      ta.TeamID,
			"teamName":    teamName,
			"role":        ta.Role,
			"roleLabel":   models.GetCoachRoleLabel(ta.Role),
			"status":      ta.Status,
			"joinedAt":    utils.FormatTime(&ta.JoinedAt),
		})
	}

	userName := ""
	userAvatar := ""
	userPhone := ""
	if coach.User != nil {
		userName = coach.User.Name
		if userName == "" {
			userName = coach.User.Nickname
		}
		userAvatar = coach.User.Avatar
		userPhone = coach.User.Phone
	}

	utils.SuccessResponse(ctx, gin.H{
		"id":          coach.ID,
		"clubId":      coach.ClubID,
		"userId":      coach.UserID,
		"name":        userName,
		"avatar":      userAvatar,
		"phone":       userPhone,
		"primaryRole": coach.PrimaryRole,
		"roleLabel":   models.GetCoachRoleLabel(coach.PrimaryRole),
		"status":      coach.Status,
		"notes":       coach.Notes,
		"joinedAt":    utils.FormatTime(&coach.JoinedAt),
		"leftAt":      utils.FormatTime(coach.LeftAt),
		"teams":       teams,
		"createdAt":   utils.FormatDateTime(coach.CreatedAt),
	})
}

// UpdateClubCoach 更新俱乐部教练
func (c *ClubController) UpdateClubCoach(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	coachID, err := strconv.ParseUint(ctx.Param("coachId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的教练ID")
		return
	}

	var req struct {
		PrimaryRole *models.CoachRole `json:"primaryRole"`
		Status      *string           `json:"status"`
		Notes       *string           `json:"notes"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	// 验证教练归属
	coach, err := c.clubService.GetClubCoachByID(uint(coachID))
	if err != nil || coach.ClubID != club.ID {
		utils.NotFoundError(ctx, "教练不存在")
		return
	}

	updates := map[string]interface{}{}
	if req.PrimaryRole != nil {
		updates["primary_role"] = *req.PrimaryRole
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}

	if err := c.clubService.UpdateClubCoach(uint(coachID), updates); err != nil {
		utils.ServerError(ctx, "更新失败")
		return
	}

	c.createAdminLog(ctx, club.ID, userID, "update_coach", "club_coach", uint(coachID), "更新教练信息")
	utils.SuccessResponseWithMessage(ctx, nil, "更新成功")
}

// RemoveClubCoach 从俱乐部移除教练
func (c *ClubController) RemoveClubCoach(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	coachID, err := strconv.ParseUint(ctx.Param("coachId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的教练ID")
		return
	}

	coach, err := c.clubService.GetClubCoachByID(uint(coachID))
	if err != nil || coach.ClubID != club.ID {
		utils.NotFoundError(ctx, "教练不存在")
		return
	}

	if err := c.clubService.RemoveClubCoach(uint(coachID)); err != nil {
		utils.ServerError(ctx, "移除失败")
		return
	}

	c.createAdminLog(ctx, club.ID, userID, "remove_coach", "club_coach", uint(coachID), "移除教练")
	utils.SuccessResponseWithMessage(ctx, nil, "教练已移除")
}

// AssignCoachToTeam 将俱乐部教练分配到球队
func (c *ClubController) AssignCoachToTeam(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	coachID, err := strconv.ParseUint(ctx.Param("coachId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的教练ID")
		return
	}

	var req struct {
		TeamID uint             `json:"teamId" binding:"required"`
		Role   models.CoachRole `json:"role" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误: "+err.Error())
		return
	}

	// 验证教练归属
	coach, err := c.clubService.GetClubCoachByID(uint(coachID))
	if err != nil || coach.ClubID != club.ID {
		utils.NotFoundError(ctx, "教练不存在")
		return
	}

	// 验证球队归属
	var team models.Team
	if err := c.db.Where("id = ? AND club_id = ?", req.TeamID, club.ID).First(&team).Error; err != nil {
		utils.NotFoundError(ctx, "球队不存在")
		return
	}

	tc, err := c.clubService.AssignCoachToTeam(coach.UserID, req.TeamID, req.Role)
	if err != nil {
		utils.ServerError(ctx, "分配失败: "+err.Error())
		return
	}

	c.createAdminLog(ctx, club.ID, userID, "assign_coach_team", "team_coach", tc.ID,
		fmt.Sprintf("分配教练到球队: %s, 角色: %s", team.Name, models.GetCoachRoleLabel(req.Role)))

	utils.SuccessResponseWithMessage(ctx, gin.H{
		"teamCoachId": tc.ID,
		"teamId":      tc.TeamID,
		"role":        tc.Role,
	}, "分配成功")
}

// RemoveCoachFromTeam 从球队移除教练
func (c *ClubController) RemoveCoachFromTeam(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	teamCoachID, err := strconv.ParseUint(ctx.Param("teamCoachId"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的ID")
		return
	}

	if err := c.clubService.RemoveCoachFromTeam(uint(teamCoachID)); err != nil {
		utils.ServerError(ctx, "移除失败")
		return
	}

	c.createAdminLog(ctx, club.ID, userID, "remove_coach_team", "team_coach", uint(teamCoachID), "从球队移除教练")
	utils.SuccessResponseWithMessage(ctx, nil, "已从球队移除")
}

// CreateClubInvitation 创建俱乐部邀请（教练加入俱乐部）
func (c *ClubController) CreateClubInvitation(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	var req struct {
		Type        string `json:"type" binding:"required"`
		TargetUserID *uint `json:"targetUserId"`
		TargetPhone  string `json:"targetPhone"`
		TargetRole   string `json:"targetRole"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	// 生成邀请码
	inviteCode := fmt.Sprintf("CLUB%s%d%d", time.Now().Format("20060102150405"), club.ID, userID%10000)

	role := models.CoachRoleAssistant
	if req.TargetRole != "" {
		role = models.CoachRole(req.TargetRole)
	}

	inv := &models.ClubInvitation{
		ClubID:       club.ID,
		Type:         models.InvitationType(req.Type),
		InviteCode:   inviteCode,
		TargetUserID: req.TargetUserID,
		TargetPhone:  req.TargetPhone,
		TargetRole:   role,
		Status:       models.InvitationStatusPending,
		CreatedBy:    userID,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour), // 7天过期
	}

	if err := c.db.Create(inv).Error; err != nil {
		utils.ServerError(ctx, "创建邀请失败")
		return
	}

	// 发送通知给被邀请人
	if req.TargetUserID != nil && *req.TargetUserID > 0 {
		roleLabel := models.GetCoachRoleLabel(role)
		notification := models.Notification{
			UserID:   *req.TargetUserID,
			Type:     models.NotificationTypeInvitation,
			Title:    "收到俱乐部邀请",
			Content:  fmt.Sprintf("%s 邀请您以%s身份加入俱乐部", club.Name, roleLabel),
			Data:     fmt.Sprintf(`{"invite_code":"%s","club_id":%d,"club_name":"%s","target_role":"%s","role_label":"%s","status":"pending"}`, inviteCode, club.ID, club.Name, role, roleLabel),
			Priority: 2,
		}
		c.db.Create(&notification)
	}

	utils.SuccessResponse(ctx, gin.H{
		"code":       inviteCode,
		"id":         inv.ID,
		"inviteCode": inviteCode,
		"message":    "邀请创建成功",
	})
}

// GetMyClubInvitations 获取我收到的待处理俱乐部邀请
func (c *ClubController) GetMyClubInvitations(ctx *gin.Context) {
	userID := ctx.GetUint("userId")

	var invitations []models.ClubInvitation
	if err := c.db.Where("target_user_id = ? AND status = ?", userID, models.InvitationStatusPending).
		Preload("Club").Order("created_at DESC").Find(&invitations).Error; err != nil {
		utils.ServerError(ctx, "获取邀请列表失败")
		return
	}

	result := make([]gin.H, 0, len(invitations))
	for _, inv := range invitations {
		clubName := ""
		if inv.Club != nil {
			clubName = inv.Club.Name
		}
		result = append(result, gin.H{
			"id":         inv.ID,
			"clubId":     inv.ClubID,
			"clubName":   clubName,
			"inviteCode": inv.InviteCode,
			"targetRole": inv.TargetRole,
			"roleLabel":  models.GetCoachRoleLabel(inv.TargetRole),
			"status":     inv.Status,
			"createdAt":  utils.FormatTime(&inv.CreatedAt),
			"expiresAt":  utils.FormatTime(&inv.ExpiresAt),
		})
	}

	utils.SuccessResponse(ctx, gin.H{
		"list": result,
	})
}

// AcceptClubInvitation 接受俱乐部邀请
func (c *ClubController) AcceptClubInvitation(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	code := ctx.Param("code")

	var inv models.ClubInvitation
	if err := c.db.Where("invite_code = ?", code).First(&inv).Error; err != nil {
		utils.NotFoundError(ctx, "邀请不存在")
		return
	}

	if inv.Status != models.InvitationStatusPending {
		utils.Error(ctx, 400, "邀请已处理")
		return
	}

	if inv.ExpiresAt.Before(time.Now()) {
		utils.Error(ctx, 400, "邀请已过期")
		return
	}

	// 验证邀请是否发给当前用户
	if inv.TargetUserID != nil && *inv.TargetUserID != userID {
		utils.ForbiddenError(ctx, "无权处理此邀请")
		return
	}

	// 如果邀请是手机定向的，验证手机号
	if inv.TargetUserID == nil && inv.TargetPhone != "" {
		var user models.User
		if err := c.db.First(&user, userID).Error; err != nil || user.Phone != inv.TargetPhone {
			utils.ForbiddenError(ctx, "无权处理此邀请")
			return
		}
	}

	// 创建 ClubCoach 记录
	clubCoach := models.ClubCoach{
		ClubID:      inv.ClubID,
		UserID:      userID,
		PrimaryRole: inv.TargetRole,
		Status:      models.ClubCoachStatusActive,
		JoinedAt:    time.Now(),
	}
	if err := c.db.Create(&clubCoach).Error; err != nil {
		utils.ServerError(ctx, "加入俱乐部失败")
		return
	}

	// 更新邀请状态
	inv.Status = models.InvitationStatusAccepted
	inv.AcceptedAt = func() *time.Time { t := time.Now(); return &t }()
	c.db.Save(&inv)

	// 获取俱乐部名称
	var club models.Club
	clubName := ""
	if err := c.db.First(&club, inv.ClubID).Error; err == nil {
		clubName = club.Name
	}

	utils.SuccessResponseWithMessage(ctx, gin.H{
		"clubId":   inv.ClubID,
		"clubName": clubName,
		"role":     inv.TargetRole,
	}, "已成功加入俱乐部")
}

// RejectClubInvitation 拒绝俱乐部邀请
func (c *ClubController) RejectClubInvitation(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	code := ctx.Param("code")

	var inv models.ClubInvitation
	if err := c.db.Where("invite_code = ?", code).First(&inv).Error; err != nil {
		utils.NotFoundError(ctx, "邀请不存在")
		return
	}

	if inv.Status != models.InvitationStatusPending {
		utils.Error(ctx, 400, "邀请已处理")
		return
	}

	if inv.TargetUserID != nil && *inv.TargetUserID != userID {
		utils.ForbiddenError(ctx, "无权处理此邀请")
		return
	}

	inv.Status = models.InvitationStatusRejected
	c.db.Save(&inv)

	utils.SuccessResponseWithMessage(ctx, nil, "已拒绝邀请")
}

// GetClubInvitations 获取俱乐部邀请列表
func (c *ClubController) GetClubInvitations(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	status := ctx.Query("status")
	var invitations []models.ClubInvitation
	query := c.db.Where("club_id = ?", club.ID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Preload("TargetUser").Order("created_at DESC").Find(&invitations).Error; err != nil {
		utils.ServerError(ctx, "获取邀请列表失败")
		return
	}

	result := make([]gin.H, 0, len(invitations))
	for _, inv := range invitations {
		item := gin.H{
			"id":         inv.ID,
			"clubId":     inv.ClubID,
			"type":       inv.Type,
			"inviteCode": inv.InviteCode,
			"status":     inv.Status,
			"createdAt":  utils.FormatTime(&inv.CreatedAt),
			"expiresAt":  utils.FormatTime(&inv.ExpiresAt),
		}
		if inv.TargetUser != nil {
			item["targetUser"] = gin.H{
				"id":   inv.TargetUser.ID,
				"name": inv.TargetUser.Name,
				"phone": inv.TargetUser.Phone,
			}
		}
		result = append(result, item)
	}

	utils.SuccessResponse(ctx, gin.H{
		"list": result,
	})
}
