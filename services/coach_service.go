package services

import (
	"encoding/json"
	"time"

	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/repositories"
	"gorm.io/gorm"
)

// CoachService 教练服务层
type CoachService struct {
	db                 *gorm.DB
	coachRepo          *repositories.CoachRepository
	teamRepo           *repositories.TeamRepository
	weeklyReportRepo   *repositories.WeeklyReportRepository
	matchSummaryRepo   *repositories.MatchSummaryRepository
}

// NewCoachService 创建教练服务
func NewCoachService(db *gorm.DB, teamRepo *repositories.TeamRepository, weeklyReportRepo *repositories.WeeklyReportRepository, matchSummaryRepo *repositories.MatchSummaryRepository) *CoachService {
	return &CoachService{
		db:               db,
		coachRepo:        repositories.NewCoachRepository(db),
		teamRepo:         teamRepo,
		weeklyReportRepo: weeklyReportRepo,
		matchSummaryRepo: matchSummaryRepo,
	}
}

// GetCoachByUserID 根据用户ID获取教练
func (s *CoachService) GetCoachByUserID(userID uint) (*models.Coach, error) {
	return s.coachRepo.GetCoachByUserID(userID)
}

// GetOrCreateCoach 获取或创建教练资料
func (s *CoachService) GetOrCreateCoach(userID uint) (*models.Coach, error) {
	coach, err := s.coachRepo.GetCoachByUserID(userID)
	if err == gorm.ErrRecordNotFound {
		// 创建新教练资料
		coach = &models.Coach{
			UserID:        userID,
			LicenseType:   "",
			Specialties:   "[]",
			CoachingYears: 0,
			Verified:      false,
		}
		if err := s.coachRepo.CreateCoach(coach); err != nil {
			return nil, err
		}
		return coach, nil
	}
	return coach, err
}

// UpdateCoachProfile 更新教练资料
// position 用于同步更新 users.position（球探地图筛选用）
func (s *CoachService) UpdateCoachProfile(userID uint, licenseType, licenseNumber, specialties, style, ageGroups, bio, city, currentClub string, coachingYears int, position string) (*models.Coach, error) {
	coach, err := s.coachRepo.GetCoachByUserID(userID)
	if err == gorm.ErrRecordNotFound {
		// 创建新教练资料
		specialtiesJSON, _ := json.Marshal(specialties)
		styleJSON, _ := json.Marshal(style)
		ageGroupsJSON, _ := json.Marshal(ageGroups)
		coach = &models.Coach{
			UserID:        userID,
			LicenseType:   licenseType,
			LicenseNumber: licenseNumber,
			Specialties:   string(specialtiesJSON),
			Style:         string(styleJSON),
			AgeGroups:     string(ageGroupsJSON),
			Bio:           bio,
			City:          city,
			CoachingYears: coachingYears,
			CurrentClub:   currentClub,
			Verified:      false,
		}
		err = s.coachRepo.CreateCoach(coach)
		return coach, err
	} else if err != nil {
		return nil, err
	}

	// 更新现有教练资料
	if licenseType != "" {
		coach.LicenseType = licenseType
	}
	if licenseNumber != "" {
		coach.LicenseNumber = licenseNumber
	}
	if specialties != "" {
		specialtiesJSON, _ := json.Marshal(specialties)
		coach.Specialties = string(specialtiesJSON)
	}
	if style != "" {
		styleJSON, _ := json.Marshal(style)
		coach.Style = string(styleJSON)
	}
	if ageGroups != "" {
		ageGroupsJSON, _ := json.Marshal(ageGroups)
		coach.AgeGroups = string(ageGroupsJSON)
	}
	if bio != "" {
		coach.Bio = bio
	}
	if city != "" {
		coach.City = city
	}
	if coachingYears > 0 {
		coach.CoachingYears = coachingYears
	}
	if currentClub != "" {
		coach.CurrentClub = currentClub
	}

	err = s.coachRepo.UpdateCoach(coach)

	// 同步更新 users.position（球探地图筛选用）
	if err == nil && position != "" {
		s.db.Model(&models.User{}).Where("id = ?", userID).Update("position", position)
	}

	return coach, err
}

// GetDashboardStats 获取工作台统计数据
func (s *CoachService) GetDashboardStats(coachID uint, userID uint) map[string]interface{} {
	// 获取关注球员数
	players, totalPlayers, _ := s.coachRepo.GetFollowedPlayers(coachID, 1, 100)
	starredCount := 0
	for _, p := range players {
		if p.IsStarred {
			starredCount++
		}
	}

	// 获取训练笔记数
	_, totalNotes, _ := s.coachRepo.GetTrainingNotes(coachID, 1, 1000, nil, "")

	// 获取执教球队
	teams, _ := s.teamRepo.GetCoachTeams(userID)
	teamCount := len(teams)
	totalManagedPlayers := int64(0)
	teamIDs := make([]uint, 0, len(teams))
	for _, t := range teams {
		teamIDs = append(teamIDs, t.ID)
		count, _ := s.teamRepo.CountPlayers(t.ID)
		totalManagedPlayers += count
	}

	// 获取本周周报提交率（跨所有执教球队）
	weeklySubmitRate := 0
	weeklyTotalPlayers := 0
	weeklySubmitted := 0
	pendingWeeklyReports := int64(0)
	var recentPendingWeeklyReports []map[string]interface{}
	if len(teamIDs) > 0 {
		now := time.Now()
		weekStart := now.AddDate(0, 0, -int(now.Weekday())+1)
		if now.Weekday() == time.Sunday {
			weekStart = now.AddDate(0, 0, -6)
		}
		weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())
		s.db.Model(&models.WeeklyReportPeriod{}).
			Where("team_id IN ? AND week_start >= ?", teamIDs, weekStart).
			Select("COALESCE(SUM(total_players), 0) AS total, COALESCE(SUM(submitted_count), 0) AS submitted").
			Row().Scan(&weeklyTotalPlayers, &weeklySubmitted)
		if weeklyTotalPlayers > 0 {
			weeklySubmitRate = weeklySubmitted * 100 / weeklyTotalPlayers
		}

		s.db.Model(&models.WeeklyReport{}).
			Where("team_id IN ? AND review_status = ?", teamIDs, "pending").
			Count(&pendingWeeklyReports)

		var recentReports []models.WeeklyReport
		s.db.Where("team_id IN ? AND review_status = ?", teamIDs, "pending").
			Order("created_at DESC").
			Limit(5).
			Find(&recentReports)

		for _, r := range recentReports {
			var player models.User
			s.db.First(&player, r.PlayerID)
			var team models.Team
			s.db.First(&team, r.TeamID)
			recentPendingWeeklyReports = append(recentPendingWeeklyReports, map[string]interface{}{
				"id":         r.ID,
				"playerId":   r.PlayerID,
				"playerName": player.Name,
				"teamId":     r.TeamID,
				"teamName":   team.Name,
				"weekStart":  r.WeekStart.Format("2006-01-02"),
				"weekEnd":    r.WeekEnd.Format("2006-01-02"),
				"status":     r.ReviewStatus,
			})
		}
	}

	// 获取待点评比赛总结数（跨所有执教球队）
	pendingMatchSummaries := int64(0)
	var recentPendingMatchSummaries []map[string]interface{}
	if len(teamIDs) > 0 {
		s.db.Model(&models.MatchSummary{}).
			Where("team_id IN ? AND status = ?", teamIDs, "player_submitted").
			Count(&pendingMatchSummaries)

		var recentMatches []models.MatchSummary
		s.db.Where("team_id IN ? AND status = ?", teamIDs, "player_submitted").
			Order("created_at DESC").
			Limit(5).
			Find(&recentMatches)

		for _, m := range recentMatches {
			var team models.Team
			s.db.First(&team, m.TeamID)
			recentPendingMatchSummaries = append(recentPendingMatchSummaries, map[string]interface{}{
				"id":              m.ID,
				"teamId":          m.TeamID,
				"teamName":        team.Name,
				"matchName":       m.MatchName,
				"matchDate":       m.MatchDate,
				"opponent":        m.Opponent,
				"ourScore":        m.OurScore,
				"opponentScore":   m.OppScore,
				"matchResult":     m.Result,
				"status":          m.Status,
			})
		}
	}

	return map[string]interface{}{
		"followedPlayers":             totalPlayers,
		"totalReports":                int(totalNotes) * 3, // 估算
		"trainingNotes":               totalNotes,
		"monthlyViews":                30,                  // 保留字段，避免前端报错
		"starredPlayers":              starredCount,
		"teamCount":                   teamCount,
		"totalPlayers":                totalManagedPlayers,
		"weeklyReportSubmitRate":      weeklySubmitRate,
		"weeklyReportTotal":           weeklyTotalPlayers,
		"weeklyReportSubmitted":       weeklySubmitted,
		"pendingWeeklyReports":        pendingWeeklyReports,
		"pendingMatchSummaries":       pendingMatchSummaries,
		"recentPendingWeeklyReports":  recentPendingWeeklyReports,
		"recentPendingMatchSummaries": recentPendingMatchSummaries,
	}
}

// GetFollowedPlayers 获取关注的球员列表
func (s *CoachService) GetFollowedPlayers(coachID uint, page, pageSize int, keyword string) ([]map[string]interface{}, int64, error) {
	follows, total, err := s.coachRepo.GetFollowedPlayers(coachID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	result := make([]map[string]interface{}, 0, len(follows))
	for _, f := range follows {
		if f.User == nil {
			continue
		}

		// 过滤关键词
		if keyword != "" {
			found := false
			if f.User.Name != "" && contains(keyword, f.User.Name) {
				found = true
			}
			if f.User.Nickname != "" && contains(keyword, f.User.Nickname) {
				found = true
			}
			if !found {
				continue
			}
		}

		var playerSpecialties []string
		_ = json.Unmarshal([]byte(f.User.Nickname), &playerSpecialties) // ignore error

		result = append(result, map[string]interface{}{
			"id":              f.User.ID,
			"userId":          f.User.ID,
			"name":            f.User.Name,
			"avatar":          f.User.Avatar,
			"age":             f.User.Age,
			"position":        f.User.Position,
			"positionName":    models.GetPositionName(f.User.Position),
			"clubName":        "", // 从球员信息获取
			"reportCount":     0,  // TODO: 从报告统计获取
			"lastReportDate":  nil,
			"overallRating":   0, // TODO: 从报告统计获取
			"isStarred":       f.IsStarred,
			"notes":           f.Notes,
			"followedAt":      f.FollowedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return result, total, nil
}

// FollowPlayer 关注球员
func (s *CoachService) FollowPlayer(coachID, playerID uint) error {
	return s.coachRepo.FollowPlayer(coachID, playerID)
}

// UnfollowPlayer 取消关注球员
func (s *CoachService) UnfollowPlayer(coachID, playerID uint) error {
	return s.coachRepo.UnfollowPlayer(coachID, playerID)
}

// UpdateFollowNotes 更新关注备注
func (s *CoachService) UpdateFollowNotes(coachID, playerID uint, notes string, isStarred bool) error {
	return s.coachRepo.UpdateFollowNotes(coachID, playerID, notes, isStarred)
}

// GetTrainingNotes 获取训练笔记列表
func (s *CoachService) GetTrainingNotes(coachID uint, page, pageSize int, playerID *uint, category string) ([]map[string]interface{}, int64, error) {
	notes, total, err := s.coachRepo.GetTrainingNotes(coachID, page, pageSize, playerID, category)
	if err != nil {
		return nil, 0, err
	}

	result := make([]map[string]interface{}, 0, len(notes))
	for _, n := range notes {
		var tags []string
		json.Unmarshal([]byte(n.Tags), &tags)

		playerName := ""
		if n.Player != nil {
			playerName = n.Player.Name
		}

		result = append(result, map[string]interface{}{
			"id":         n.ID,
			"playerId":   n.PlayerID,
			"playerName": playerName,
			"title":      n.Title,
			"content":    n.Content,
			"category":   n.Category,
			"tags":       tags,
			"rating":     n.Rating,
			"isPublic":   n.IsPublic,
			"viewCount":  n.ViewCount,
			"createdAt":  n.CreatedAt.Format("2006-01-02T15:04:05Z"),
			"updatedAt":  n.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return result, total, nil
}

// CreateTrainingNote 创建训练笔记
func (s *CoachService) CreateTrainingNote(coachID uint, playerID uint, title, content, category string, tags []string, rating int, isPublic bool) (*models.TrainingNote, error) {
	tagsJSON, _ := json.Marshal(tags)
	note := &models.TrainingNote{
		CoachID:  coachID,
		PlayerID: playerID,
		Title:    title,
		Content:  content,
		Category: category,
		Tags:     string(tagsJSON),
		Rating:   rating,
		IsPublic: isPublic,
	}

	err := s.coachRepo.CreateTrainingNote(note)
	return note, err
}

// UpdateTrainingNote 更新训练笔记
func (s *CoachService) UpdateTrainingNote(coachID, noteID uint, title, content, category string, tags []string, rating int, isPublic bool) error {
	note, err := s.coachRepo.GetTrainingNoteByID(coachID, noteID)
	if err != nil {
		return err
	}

	note.Title = title
	note.Content = content
	note.Category = category
	tagsJSON, _ := json.Marshal(tags)
	note.Tags = string(tagsJSON)
	note.Rating = rating
	note.IsPublic = isPublic

	return s.coachRepo.UpdateTrainingNote(note)
}

// DeleteTrainingNote 删除训练笔记
func (s *CoachService) DeleteTrainingNote(coachID, noteID uint) error {
	return s.coachRepo.DeleteTrainingNote(coachID, noteID)
}

// GetPlayerProgress 获取球员进度
func (s *CoachService) GetPlayerProgress(coachID, playerID uint) ([]map[string]interface{}, error) {
	return s.coachRepo.GetPlayerProgress(coachID, playerID)
}

// GetFootballExperiences 获取足球经历列表
func (s *CoachService) GetFootballExperiences(coachID uint) ([]models.FootballExperience, error) {
	return s.coachRepo.GetFootballExperiences(coachID)
}

// CreateFootballExperience 创建足球经历
func (s *CoachService) CreateFootballExperience(coachID uint, stage, teamName, position string, startYear, endYear int, level, honors string) (*models.FootballExperience, error) {
	exp := &models.FootballExperience{
		CoachID:   coachID,
		Stage:     stage,
		TeamName:  teamName,
		Position:  position,
		StartYear: startYear,
		EndYear:   endYear,
		Level:     level,
		Honors:    honors,
	}
	err := s.coachRepo.CreateFootballExperience(exp)
	return exp, err
}

// UpdateFootballExperience 更新足球经历
func (s *CoachService) UpdateFootballExperience(coachID, expID uint, stage, teamName, position string, startYear, endYear int, level, honors string) error {
	exp, err := s.coachRepo.GetFootballExperienceByID(coachID, expID)
	if err != nil {
		return err
	}

	if stage != "" {
		exp.Stage = stage
	}
	if teamName != "" {
		exp.TeamName = teamName
	}
	exp.Position = position
	exp.StartYear = startYear
	exp.EndYear = endYear
	exp.Level = level
	exp.Honors = honors

	return s.coachRepo.UpdateFootballExperience(exp)
}

// DeleteFootballExperience 删除足球经历
func (s *CoachService) DeleteFootballExperience(coachID, expID uint) error {
	return s.coachRepo.DeleteFootballExperience(coachID, expID)
}

// GetRecentActivities 获取最近动态
func (s *CoachService) GetRecentActivities(coachID uint, limit int) ([]map[string]interface{}, error) {
	activities := make([]map[string]interface{}, 0)

	// 获取最近的训练笔记
	notes, _, _ := s.coachRepo.GetTrainingNotes(coachID, 1, limit, nil, "")
	for _, n := range notes {
		playerName := ""
		if n.Player != nil {
			playerName = n.Player.Name
		}
		activities = append(activities, map[string]interface{}{
			"id":          n.ID,
			"type":        "add_note",
			"playerName":  playerName,
			"description": "添加了训练笔记",
			"time":        formatTimeAgo(n.CreatedAt),
		})
	}

	return activities, nil
}

// GetCoachPublicProfile 获取教练公开主页数据
type CoachPublicProfile struct {
	Coach               *CoachPublicInfo             `json:"coach"`
	User                *models.User                 `json:"user,omitempty"`
	Stats               CoachPublicStats              `json:"stats"`
	SampleNotes         []CoachPublicNote             `json:"sample_notes"`
	CoachingTeams       []CoachPublicTeam             `json:"coaching_teams"`
	FootballExperiences []CoachPublicFootballExp      `json:"football_experiences"`
}

// CoachPublicInfo API响应专用的教练信息结构
type CoachPublicInfo struct {
	ID              uint     `json:"id"`
	UserID          uint     `json:"user_id"`
	LicenseType     string   `json:"license_type"`
	LicenseNumber   string   `json:"license_number"`
	Specialties     []string `json:"specialties"`
	Style           []string `json:"style"`
	AgeGroups       []string `json:"age_groups"`
	Bio             string   `json:"bio"`
	CoachingYears   int      `json:"coaching_years"`
	CurrentClub     string   `json:"current_club"`
	City            string   `json:"city"`
	Verified        bool     `json:"verified"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
}

type CoachPublicStats struct {
	FollowedPlayers  int64 `json:"followed_players"`
	TrainingNotes   int64 `json:"training_notes"`
	CoachingYears  int    `json:"coaching_years"`
	TeamCount      int64 `json:"team_count"`
}

type CoachPublicNote struct {
	ID        uint   `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Category  string `json:"category"`
	CreatedAt string `json:"created_at"`
}

type CoachPublicTeam struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Logo string `json:"logo"`
}

type CoachPublicFootballExp struct {
	ID         uint   `json:"id"`
	Stage      string `json:"stage"`
	StageName  string `json:"stage_name"`
	TeamName   string `json:"team_name"`
	Position   string `json:"position"`
	StartYear  int    `json:"start_year"`
	EndYear    int    `json:"end_year"`
	Level      string `json:"level"`
	Honors     string `json:"honors"`
}

func (s *CoachService) GetCoachPublicProfile(coachID uint) (*CoachPublicProfile, error) {
	coach, err := s.coachRepo.GetCoachByID(coachID)
	if err != nil {
		return nil, err
	}
	if coach == nil {
		return nil, nil
	}

	// 解析专长
	var specialties []string
	json.Unmarshal([]byte(coach.Specialties), &specialties)

	// 获取统计
	followedPlayers, totalFollowed, _ := s.coachRepo.GetFollowedPlayers(coachID, 1, 1)
	notes, totalNotes, _ := s.coachRepo.GetTrainingNotes(coachID, 1, 10, nil, "")
	starredCount := int64(0)
	for _, f := range followedPlayers {
		if f.IsStarred {
			starredCount++
		}
	}

	// 获取执教的球队（需要从 team_coaches 表获取）

	// 示例笔记（取最新的3条公开笔记）
	sampleNotes := make([]CoachPublicNote, 0)
	for _, n := range notes {
		if n.IsPublic {
			sampleNotes = append(sampleNotes, CoachPublicNote{
				ID:        n.ID,
				Title:     n.Title,
				Content:   n.Content,
				Category:  n.Category,
				CreatedAt: n.CreatedAt.Format("2006-01-02"),
			})
			if len(sampleNotes) >= 3 {
				break
			}
		}
	}

	// 获取足球经历
	footballExps, _ := s.coachRepo.GetFootballExperiences(coachID)
	publicExps := make([]CoachPublicFootballExp, 0)
	for _, exp := range footballExps {
		publicExps = append(publicExps, CoachPublicFootballExp{
			ID:        exp.ID,
			Stage:     exp.Stage,
			StageName: models.StageNameMap[models.FootballStage(exp.Stage)],
			TeamName:  exp.TeamName,
			Position:  exp.Position,
			StartYear: exp.StartYear,
			EndYear:   exp.EndYear,
			Level:     exp.Level,
			Honors:    exp.Honors,
		})
	}

	return &CoachPublicProfile{
		Coach: &CoachPublicInfo{
			ID:            coach.ID,
			UserID:        coach.UserID,
			LicenseType:   coach.LicenseType,
			LicenseNumber: coach.LicenseNumber,
			Specialties:   parseJSONArray(coach.Specialties),
			Style:         parseJSONArray(coach.Style),
			AgeGroups:     parseJSONArray(coach.AgeGroups),
			Bio:           coach.Bio,
			CoachingYears: coach.CoachingYears,
			CurrentClub:   coach.CurrentClub,
			City:          coach.City,
			Verified:      coach.Verified,
			CreatedAt:     coach.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:     coach.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		},
		User: coach.User,
		Stats: CoachPublicStats{
			FollowedPlayers: totalFollowed,
			TrainingNotes:   totalNotes,
			CoachingYears:  coach.CoachingYears,
			TeamCount:      0, // TODO: 从球队关联表获取
		},
		SampleNotes:         sampleNotes,
		CoachingTeams:       []CoachPublicTeam{}, // TODO: 从球队关联表获取
		FootballExperiences: publicExps,
	}, nil
}

// Helper functions
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func formatTimeAgo(t time.Time) string {
	diff := time.Since(t)
	if diff < time.Minute {
		return "刚刚"
	}
	if diff < time.Hour {
		return string(rune(int(diff.Minutes()))) + "分钟前"
	}
	if diff < 24*time.Hour {
		return string(rune(int(diff.Hours()))) + "小时前"
	}
	if diff < 7*24*time.Hour {
		return string(rune(int(diff.Hours()/24))) + "天前"
	}
	return t.Format("2006-01-02")
}