package services

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/repositories"
	"gorm.io/gorm"
)

// MatchSummaryService 比赛总结服务
type MatchSummaryService struct {
	db          *gorm.DB
	summaryRepo *repositories.MatchSummaryRepository
	reviewRepo  *repositories.PlayerReviewRepository
	videoRepo   *repositories.MatchVideoRepository
	teamRepo    *repositories.TeamRepository
	userRepo    *models.UserRepository
}

// NewMatchSummaryService 创建比赛总结服务
func NewMatchSummaryService(
	db *gorm.DB,
	summaryRepo *repositories.MatchSummaryRepository,
	reviewRepo *repositories.PlayerReviewRepository,
	videoRepo *repositories.MatchVideoRepository,
	teamRepo *repositories.TeamRepository,
	userRepo *models.UserRepository,
) *MatchSummaryService {
	return &MatchSummaryService{
		db:          db,
		summaryRepo: summaryRepo,
		reviewRepo:  reviewRepo,
		videoRepo:   videoRepo,
		teamRepo:    teamRepo,
		userRepo:    userRepo,
	}
}

// ============================================================
// 权限检查
// ============================================================

// isTeamCoach 检查用户是否是球队教练
func (s *MatchSummaryService) isTeamCoach(teamID, userID uint) bool {
	var count int64
	s.db.Model(&models.TeamCoach{}).
		Where("team_id = ? AND user_id = ? AND status = ?", teamID, userID, "active").
		Count(&count)
	return count > 0
}

// isClubAdminOfTeam 检查用户是否是该球队所属俱乐部的管理员
func (s *MatchSummaryService) isClubAdminOfTeam(teamID, userID uint) bool {
	var club models.Club
	if err := s.db.Where("user_id = ?", userID).First(&club).Error; err != nil {
		return false
	}
	var count int64
	s.db.Model(&models.Team{}).
		Where("id = ? AND club_id = ?", teamID, club.ID).
		Count(&count)
	return count > 0
}

// isPlayerInTeam 检查球员是否属于球队
func (s *MatchSummaryService) isPlayerInTeam(teamID, playerID uint) bool {
	var count int64
	s.db.Model(&models.TeamPlayer{}).
		Where("team_id = ? AND user_id = ? AND status = ?", teamID, playerID, "active").
		Count(&count)
	return count > 0
}

// CanAccessMatchSummary 检查用户是否有权访问比赛总结
func (s *MatchSummaryService) CanAccessMatchSummary(summaryID, userID uint) bool {
	summary, err := s.summaryRepo.GetByID(summaryID)
	if err != nil {
		return false
	}
	if s.isTeamCoach(summary.TeamID, userID) {
		return true
	}
	if s.isClubAdminOfTeam(summary.TeamID, userID) {
		return true
	}
	playerIDs := summary.GetPlayerIDs()
	if len(playerIDs) > 0 {
		for _, pid := range playerIDs {
			if pid == userID {
				return true
			}
		}
		return false
	}
	return s.isPlayerInTeam(summary.TeamID, userID)
}

// ============================================================
// 比赛CRUD
// ============================================================

// Create 创建比赛
func (s *MatchSummaryService) Create(coachID uint, input *models.MatchSummaryCreate) (*models.MatchSummary, error) {
	if !s.isTeamCoach(input.TeamID, coachID) {
		return nil, errors.New("只有球队教练才能创建比赛")
	}

	team, err := s.teamRepo.FindByID(input.TeamID)
	if err != nil {
		return nil, errors.New("球队不存在")
	}

	// 验证参赛球员都属于该球队
	for _, pid := range input.PlayerIDs {
		if !s.isPlayerInTeam(input.TeamID, pid) {
			return nil, errors.New("部分球员不属于该球队")
		}
	}

	location := input.Location
	if location == "" {
		location = "home"
	}

	matchFormat := input.MatchFormat
	if matchFormat == "" {
		matchFormat = "11人制"
	}

	// 兼容旧字段 opponentScore
	oppScore := input.OppScore
	if oppScore == 0 && input.OpponentScore > 0 {
		oppScore = input.OpponentScore
	}

	result := input.Result
	if result == "" {
		if input.OurScore > oppScore {
			result = "win"
		} else if input.OurScore < oppScore {
			result = "lose"
		} else if input.OurScore == 0 && oppScore == 0 {
			result = "pending"
		} else {
			result = "draw"
		}
	}

	playerIDsJSON, _ := json.Marshal(input.PlayerIDs)

	summary := &models.MatchSummary{
		TeamID:      input.TeamID,
		CoachID:     coachID,
		MatchName:   input.MatchName,
		MatchDate:   input.MatchDate,
		Opponent:    input.Opponent,
		Location:    location,
		MatchFormat: matchFormat,
		OurScore:    input.OurScore,
		OppScore:    oppScore,
		Result:      result,
		CoverImage:  input.CoverImage,
		PlayerIDs:   string(playerIDsJSON),
		PlayerCount: len(input.PlayerIDs),
		Status:      "pending",
	}

	if err := s.summaryRepo.Create(summary); err != nil {
		return nil, err
	}

	summary.Team = team
	loaded, _ := s.summaryRepo.GetByID(summary.ID)
	if loaded != nil {
		return loaded, nil
	}
	return summary, nil
}

// Update 更新比赛
func (s *MatchSummaryService) Update(id, coachID uint, input *models.MatchSummaryUpdate) (*models.MatchSummary, error) {
	summary, err := s.summaryRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("比赛不存在")
	}

	if !s.isTeamCoach(summary.TeamID, coachID) {
		return nil, errors.New("只有球队教练才能编辑比赛")
	}

	if input.MatchName != "" {
		summary.MatchName = input.MatchName
	}
	if input.MatchDate != "" {
		summary.MatchDate = input.MatchDate
	}
	if input.Opponent != "" {
		summary.Opponent = input.Opponent
	}
	if input.Location != "" {
		summary.Location = input.Location
	}
	if input.MatchFormat != "" {
		summary.MatchFormat = input.MatchFormat
	}
	if input.OurScore != nil {
		summary.OurScore = *input.OurScore
	}
	if input.OppScore != nil {
		summary.OppScore = *input.OppScore
	}
	if input.CoverImage != "" {
		summary.CoverImage = input.CoverImage
	}
	if input.PlayerIDs != nil {
		playerIDsJSON, _ := json.Marshal(input.PlayerIDs)
		summary.PlayerIDs = string(playerIDsJSON)
		summary.PlayerCount = len(input.PlayerIDs)
	}

	if input.Result != "" {
		summary.Result = input.Result
	} else if input.OurScore != nil || input.OppScore != nil {
		summary.CalcResult()
	}

	if err := s.summaryRepo.Update(summary); err != nil {
		return nil, err
	}
	return s.summaryRepo.GetByID(id)
}

// Delete 删除比赛
func (s *MatchSummaryService) Delete(id, coachID uint) error {
	summary, err := s.summaryRepo.GetByID(id)
	if err != nil {
		return errors.New("比赛不存在")
	}

	if !s.isTeamCoach(summary.TeamID, coachID) {
		return errors.New("只有球队教练才能删除比赛")
	}

	s.reviewRepo.DeleteByMatch(id)
	s.videoRepo.DeleteByMatch(id)
	return s.summaryRepo.Delete(id)
}

// GetByID 获取比赛详情
func (s *MatchSummaryService) GetByID(id uint) (*models.MatchSummary, error) {
	return s.summaryRepo.GetByID(id)
}

// ============================================================
// 列表查询
// ============================================================

// ListByTeam 列出球队比赛
func (s *MatchSummaryService) ListByTeam(teamID uint, status string, page, pageSize int) ([]models.MatchSummary, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}
	return s.summaryRepo.ListByTeam(teamID, status, page, pageSize)
}

// ListByCoach 列出教练发起的比赛
func (s *MatchSummaryService) ListByCoach(coachID uint, status string, page, pageSize int) ([]models.MatchSummary, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}
	return s.summaryRepo.ListByCoach(coachID, status, page, pageSize)
}

// ListByPlayer 列出球员参与的比赛
func (s *MatchSummaryService) ListByPlayer(playerID uint, page, pageSize int) ([]models.MatchSummary, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}
	return s.summaryRepo.ListByPlayer(playerID, page, pageSize)
}

// ListByClub 列出俱乐部所有球队的比赛
func (s *MatchSummaryService) ListByClub(clubID uint, status string, page, pageSize int) ([]models.MatchSummary, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}
	return s.summaryRepo.ListByClub(clubID, status, page, pageSize)
}

// GetPendingCount 获取待处理数量
func (s *MatchSummaryService) GetPendingCount(teamID uint) (int64, error) {
	return s.summaryRepo.CountPending(teamID)
}

// ============================================================
// 球员自评
// ============================================================

// SubmitPlayerReview 球员提交自评
func (s *MatchSummaryService) SubmitPlayerReview(summaryID, playerID uint, input *models.PlayerReviewSubmit) (*models.PlayerReview, error) {
	summary, err := s.summaryRepo.GetByID(summaryID)
	if err != nil {
		return nil, errors.New("比赛不存在")
	}

	playerIDs := summary.GetPlayerIDs()
	if len(playerIDs) > 0 {
		found := false
		for _, pid := range playerIDs {
			if pid == playerID {
				found = true
				break
			}
		}
		if !found {
			return nil, errors.New("您不在该比赛的参赛球员名单中")
		}
	} else {
		if !s.isPlayerInTeam(summary.TeamID, playerID) {
			return nil, errors.New("您无权提交该比赛的自评")
		}
	}

	// 检查是否已提交
	existing, err := s.reviewRepo.GetByMatchAndPlayer(summaryID, playerID)
	if err == nil && existing != nil {
		existing.Performance = input.Performance
		existing.Goals = input.Goals
		existing.Assists = input.Assists
		existing.Saves = input.Saves
		existing.Highlights = input.Highlights
		existing.Improvements = input.Improvements
		existing.NextGoals = input.NextGoals
		if input.Tactics != nil {
			tacticsJSON, _ := json.Marshal(input.Tactics)
			existing.Tactics = string(tacticsJSON)
		}
		if err := s.reviewRepo.Update(existing); err != nil {
			return nil, err
		}
		if summary.Status == "pending" {
			summary.Status = "player_submitted"
			if err := s.summaryRepo.Update(summary); err != nil {
				return nil, err
			}
		}
		return s.reviewRepo.GetByID(existing.ID)
	}

	tacticsJSON := ""
	if input.Tactics != nil {
		data, _ := json.Marshal(input.Tactics)
		tacticsJSON = string(data)
	}

	review := &models.PlayerReview{
		MatchID:      summaryID,
		PlayerID:     playerID,
		TeamID:       summary.TeamID,
		Performance:  input.Performance,
		Goals:        input.Goals,
		Assists:      input.Assists,
		Saves:        input.Saves,
		Tactics:      tacticsJSON,
		Highlights:   input.Highlights,
		Improvements: input.Improvements,
		NextGoals:    input.NextGoals,
		Status:       "submitted",
		SubmittedAt:  time.Now(),
	}

	if err := s.reviewRepo.Create(review); err != nil {
		return nil, err
	}

	// 更新比赛状态
	if summary.Status == "pending" {
		summary.Status = "player_submitted"
		s.summaryRepo.Update(summary)
	}

	return s.reviewRepo.GetByID(review.ID)
}

// GetPlayerReview 获取球员自评详情
func (s *MatchSummaryService) GetPlayerReview(summaryID, playerID uint) (*models.PlayerReview, error) {
	return s.reviewRepo.GetByMatchAndPlayer(summaryID, playerID)
}

// ListPlayerReviews 获取比赛的所有球员自评
func (s *MatchSummaryService) ListPlayerReviews(matchID uint) ([]models.PlayerReview, error) {
	return s.reviewRepo.ListByMatch(matchID)
}

// ============================================================
// 教练点评
// ============================================================

// SubmitCoachSummary 教练提交整体点评
func (s *MatchSummaryService) SubmitCoachSummary(summaryID, coachID uint, input *models.CoachSummarySubmit) (*models.MatchSummary, error) {
	summary, err := s.summaryRepo.GetByID(summaryID)
	if err != nil {
		return nil, errors.New("比赛不存在")
	}

	if !s.isTeamCoach(summary.TeamID, coachID) {
		return nil, errors.New("只有球队教练才能提交点评")
	}

	summary.CoachOverall = input.CoachOverall
	summary.CoachTactic = input.CoachTactic
	summary.CoachKeyMoments = input.CoachKeyMoments

	for _, pr := range input.PlayerReviews {
		review, err := s.reviewRepo.GetByMatchAndPlayer(summaryID, pr.PlayerID)
		if err != nil {
			continue
		}
		review.CoachRating = pr.Rating
		review.CoachComment = pr.CoachComment
		review.CoachReply = pr.CoachReply
		review.Status = "coach_reviewed"
		s.reviewRepo.Update(review)
	}

	summary.Status = "completed"

	if err := s.summaryRepo.Update(summary); err != nil {
		return nil, err
	}
	return s.summaryRepo.GetByID(summaryID)
}

// SubmitCoachPlayerReview 教练对单个球员提交评分点评
func (s *MatchSummaryService) SubmitCoachPlayerReview(summaryID, coachID uint, input *models.CoachPlayerReviewSubmit) (*models.PlayerReview, error) {
	summary, err := s.summaryRepo.GetByID(summaryID)
	if err != nil {
		return nil, errors.New("比赛不存在")
	}

	if !s.isTeamCoach(summary.TeamID, coachID) {
		return nil, errors.New("只有球队教练才能提交点评")
	}

	review, err := s.reviewRepo.GetByMatchAndPlayer(summaryID, input.PlayerID)
	if err != nil {
		return nil, errors.New("该球员尚未提交自评")
	}

	review.CoachRating = input.Rating
	review.CoachComment = input.CoachComment
	review.CoachReply = input.CoachReply
	review.Status = "coach_reviewed"

	if err := s.reviewRepo.Update(review); err != nil {
		return nil, err
	}
	return s.reviewRepo.GetByID(review.ID)
}

// ============================================================
// 视频管理
// ============================================================

// AddVideo 添加视频链接
func (s *MatchSummaryService) AddVideo(matchID, teamID, uploaderID uint, input *models.MatchVideoCreate) (*models.MatchVideo, error) {
	_, err := s.summaryRepo.GetByID(matchID)
	if err != nil {
		return nil, errors.New("比赛不存在")
	}

	video := &models.MatchVideo{
		MatchID:    matchID,
		TeamID:     teamID,
		UploaderID: uploaderID,
		Platform:   input.Platform,
		URL:        input.URL,
		Code:       input.Code,
		Name:       input.Name,
		Note:       input.Note,
		SortOrder:  input.SortOrder,
		Status:     "active",
	}

	if err := s.videoRepo.Create(video); err != nil {
		return nil, err
	}
	return s.videoRepo.GetByID(video.ID)
}

// DeleteVideo 删除视频链接
func (s *MatchSummaryService) DeleteVideo(videoID, coachID uint) error {
	video, err := s.videoRepo.GetByID(videoID)
	if err != nil {
		return errors.New("视频不存在")
	}

	summary, err := s.summaryRepo.GetByID(video.MatchID)
	if err != nil {
		return errors.New("比赛不存在")
	}
	if !s.isTeamCoach(summary.TeamID, coachID) {
		return errors.New("只有球队教练才能删除视频")
	}

	return s.videoRepo.Delete(videoID)
}

// ListVideos 获取比赛的视频列表
func (s *MatchSummaryService) ListVideos(matchID uint) ([]models.MatchVideo, error) {
	return s.videoRepo.ListByMatch(matchID)
}

// ============================================================
// 封面图
// ============================================================

// UpdateCoverImage 更新封面图
func (s *MatchSummaryService) UpdateCoverImage(summaryID, coachID uint, coverImage string) (*models.MatchSummary, error) {
	summary, err := s.summaryRepo.GetByID(summaryID)
	if err != nil {
		return nil, errors.New("比赛不存在")
	}

	if !s.isTeamCoach(summary.TeamID, coachID) {
		return nil, errors.New("只有球队教练才能更新封面图")
	}

	summary.CoverImage = coverImage
	if err := s.summaryRepo.Update(summary); err != nil {
		return nil, err
	}
	return s.summaryRepo.GetByID(summaryID)
}

// ============================================================
// 统计
// ============================================================

// GetMatchStats 获取比赛统计
func (s *MatchSummaryService) GetMatchStats(clubID uint) (*models.MatchStatsResponse, error) {
	statusCounts, err := s.summaryRepo.CountByStatus(clubID)
	if err != nil {
		return nil, err
	}

	resultCounts, err := s.summaryRepo.CountByResult(clubID)
	if err != nil {
		return nil, err
	}

	var total int64
	for _, v := range statusCounts {
		total += v
	}

	return &models.MatchStatsResponse{
		TotalCount:     total,
		PendingCount:   statusCounts["pending"],
		SubmittedCount: statusCounts["player_submitted"],
		CompletedCount: statusCounts["completed"],
		WinCount:       resultCounts["win"],
		DrawCount:      resultCounts["draw"],
		LoseCount:      resultCounts["lose"],
	}, nil
}

// GetSubmittedCount 获取比赛已提交自评数量
func (s *MatchSummaryService) GetSubmittedCount(matchID uint) (int64, error) {
	return s.reviewRepo.CountByMatch(matchID)
}

// RemindPlayers 催办未提交自评的球员，返回待通知的球员ID列表
func (s *MatchSummaryService) RemindPlayers(matchID, userID uint, playerIDs []uint, message string) ([]uint, string, error) {
	summary, err := s.summaryRepo.GetByID(matchID)
	if err != nil {
		return nil, "", errors.New("比赛不存在")
	}

	if !s.isTeamCoach(summary.TeamID, userID) && !s.isClubAdminOfTeam(summary.TeamID, userID) {
		return nil, "", errors.New("无权催办")
	}

	allPlayerIDs := summary.GetPlayerIDs()
	if len(allPlayerIDs) == 0 {
		return nil, "", errors.New("该比赛没有指定参赛球员")
	}

	// 获取已提交球员
	submittedReviews, _ := s.reviewRepo.ListByMatch(matchID)
	submittedMap := make(map[uint]bool)
	for _, r := range submittedReviews {
		submittedMap[r.PlayerID] = true
	}

	// 找出未提交球员
	var pendingIDs []uint
	for _, pid := range allPlayerIDs {
		if !submittedMap[pid] {
			pendingIDs = append(pendingIDs, pid)
		}
	}

	// 如果指定了 playerIDs，则只提醒指定玩家中且未提交的
	if len(playerIDs) > 0 {
		filtered := make([]uint, 0)
		targetSet := make(map[uint]bool)
		for _, pid := range playerIDs {
			targetSet[pid] = true
		}
		for _, pid := range pendingIDs {
			if targetSet[pid] {
				filtered = append(filtered, pid)
			}
		}
		pendingIDs = filtered
	}

	return pendingIDs, summary.MatchName, nil
}
