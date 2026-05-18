package services

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/shaonianqiutan/backend/models"
	"gorm.io/gorm"
)

var (
	ErrOfficialTaskNotFound          = errors.New("官方选题单不存在")
	ErrOfficialTaskUnavailable       = errors.New("官方选题单不可接取")
	ErrOfficialTaskFull              = errors.New("官方选题单名额已满")
	ErrOfficialTaskDuplicate         = errors.New("已接取过该官方选题单")
	ErrOfficialTaskLevelDenied       = errors.New("分析师等级不满足接取要求")
	ErrOfficialTaskDailyLimit        = errors.New("今日官方选题接单额度已用完")
	ErrOfficialTaskAnalystInvalid    = errors.New("分析师不存在或不可用")
	ErrOfficialTaskInvalid           = errors.New("官方选题单参数不合法")
	ErrOfficialSubmissionNotFound    = errors.New("官方任务提交不存在")
	ErrOfficialSubmissionInvalid     = errors.New("官方任务提交参数不合法")
	ErrOfficialAdoptionNotFound      = errors.New("官方采用记录不存在")
	ErrOfficialPublishRecordNotFound = errors.New("官方发布记录不存在")
	ErrAnalystRewardNotFound         = errors.New("分析师官方奖励不存在")
	ErrAnalystRewardInvalid          = errors.New("分析师官方奖励状态不允许该操作")
)

type OfficialAnalysisTaskService struct {
	db *gorm.DB
}

func NewOfficialAnalysisTaskService(db *gorm.DB) *OfficialAnalysisTaskService {
	return &OfficialAnalysisTaskService{db: db}
}

type OfficialAnalysisTaskRequest struct {
	Title                string     `json:"title"`
	MatchName            string     `json:"match_name"`
	AgeGroup             string     `json:"age_group"`
	MatchDate            string     `json:"match_date"`
	VideoURL             string     `json:"video_url"`
	VideoFirstHalfURL    string     `json:"video_first_half_url"`
	VideoSecondHalfURL   string     `json:"video_second_half_url"`
	VideoSource          string     `json:"video_source"`
	AuthorizationStatus  string     `json:"authorization_status"`
	TargetPlayerUserID   uint       `json:"target_player_user_id"`
	TargetPlayerName     string     `json:"target_player_name"`
	TargetPlayerTeam     string     `json:"target_player_team"`
	TargetJerseyColor    string     `json:"target_jersey_color"`
	TargetJerseyNumber   string     `json:"target_jersey_number"`
	TargetPlayerPosition string     `json:"target_player_position"`
	TaskType             string     `json:"task_type"`
	BaseRewardAmount     float64    `json:"base_reward_amount"`
	AdoptionRewardRule   string     `json:"adoption_reward_rule"`
	BonusRule            string     `json:"bonus_rule"`
	Requirements         string     `json:"requirements"`
	MaxAcceptCount       int        `json:"max_accept_count"`
	VisibleLevelMin      string     `json:"visible_level_min"`
	VisibleLevelCodes    []string   `json:"visible_level_codes"`
	PriorityLevelMin     string     `json:"priority_level_min"`
	PriorityUntil        *time.Time `json:"priority_until"`
	Deadline             *time.Time `json:"deadline"`
}

type OfficialAnalysisTaskBatchCreateRequest struct {
	EventName          string                        `json:"event_name"`
	Common             OfficialAnalysisTaskRequest   `json:"common"`
	Tasks              []OfficialAnalysisTaskRequest `json:"tasks"`
	PublishAfterCreate bool                          `json:"publish_after_create"`
}

type OfficialAnalysisTaskBatchCreateResult struct {
	Created []models.OfficialAnalysisTask `json:"created"`
	Total   int                           `json:"total"`
}

type OfficialTaskListFilters struct {
	Status          string
	TaskID          uint
	Keyword         string
	MatchName       string
	AgeGroup        string
	VisibleLevelMin string
	AnalystID       uint
}

type OfficialRewardListFilters struct {
	Status     string
	RewardType string
	SourceType string
	AnalystID  uint
	BatchID    uint
}

type OfficialMaterialListFilters struct {
	Keyword        string
	MatchName      string
	AgeGroup       string
	Channel        string
	AdoptionStatus string
	AssetKind      string
	PublishReady   bool
	PublishStatus  string
	MetricsStatus  string
	BonusStatus    string
	SortBy         string
	AnalystID      uint
}

type OfficialEventTopicFilters struct {
	Keyword      string
	AgeGroup     string
	FeaturedOnly bool
}

type OfficialEventTopic struct {
	MatchName         string             `json:"match_name"`
	DisplayName       string             `json:"display_name"`
	Summary           string             `json:"summary"`
	CoverURL          string             `json:"cover_url"`
	AliasNames        []string           `json:"alias_names"`
	PinnedAdoptionID  uint               `json:"pinned_adoption_id"`
	IsFeatured        bool               `json:"is_featured"`
	SortOrder         int                `json:"sort_order"`
	ConflictWarnings  []string           `json:"conflict_warnings"`
	AgeGroups         []string           `json:"age_groups"`
	WorkCount         int                `json:"work_count"`
	AnalystCount      int                `json:"analyst_count"`
	FeaturedWork      *OfficialEventWork `json:"featured_work,omitempty"`
	LatestPublishedAt string             `json:"latest_published_at"`
}

type OfficialEventTopicDetail struct {
	MatchName        string              `json:"match_name"`
	DisplayName      string              `json:"display_name"`
	Summary          string              `json:"summary"`
	CoverURL         string              `json:"cover_url"`
	AliasNames       []string            `json:"alias_names"`
	PinnedAdoptionID uint                `json:"pinned_adoption_id"`
	IsFeatured       bool                `json:"is_featured"`
	SortOrder        int                 `json:"sort_order"`
	AgeGroupFilter   string              `json:"age_group_filter"`
	AgeGroups        []string            `json:"age_groups"`
	WorkCount        int                 `json:"work_count"`
	AnalystCount     int                 `json:"analyst_count"`
	FeaturedWork     *OfficialEventWork  `json:"featured_work,omitempty"`
	Works            []OfficialEventWork `json:"works"`
}

type OfficialPublishRecordRequest struct {
	Channel       string     `json:"channel"`
	AccountName   string     `json:"account_name"`
	PublishURL    string     `json:"publish_url"`
	PublishTitle  string     `json:"publish_title"`
	ReusePurpose  string     `json:"reuse_purpose"`
	PublishedAt   *time.Time `json:"published_at"`
	PlayCount     int64      `json:"play_count"`
	LikeCount     int64      `json:"like_count"`
	CommentCount  int64      `json:"comment_count"`
	ShareCount    int64      `json:"share_count"`
	FavoriteCount int64      `json:"favorite_count"`
	MetricsAt     *time.Time `json:"metrics_recorded_at"`
	Note          string     `json:"note"`
}

type OfficialEventWork struct {
	ID             uint    `json:"id"`
	TaskID         uint    `json:"task_id"`
	SubmissionID   uint    `json:"submission_id"`
	AnalystID      uint    `json:"analyst_id"`
	AnalystUserID  uint    `json:"analyst_user_id"`
	AnalystName    string  `json:"analyst_name"`
	AnalystLevel   string  `json:"analyst_level"`
	IsPartner      bool    `json:"is_partner"`
	MatchName      string  `json:"match_name"`
	AgeGroup       string  `json:"age_group"`
	Channel        string  `json:"channel"`
	AdoptionStatus string  `json:"adoption_status"`
	WorkTitle      string  `json:"work_title"`
	WorkSummary    string  `json:"work_summary"`
	CoverURL       string  `json:"cover_url"`
	RewardAmount   float64 `json:"reward_amount"`
	CreatedAt      string  `json:"created_at"`
}

type OfficialEventTopicConfigRequest struct {
	MatchName        string   `json:"match_name"`
	DisplayName      string   `json:"display_name"`
	Summary          string   `json:"summary"`
	CoverURL         string   `json:"cover_url"`
	AliasNames       []string `json:"alias_names"`
	PinnedAdoptionID uint     `json:"pinned_adoption_id"`
	IsFeatured       bool     `json:"is_featured"`
	SortOrder        int      `json:"sort_order"`
}

type OfficialRewardActionRequest struct {
	Note string `json:"note"`
}

type OfficialRewardBatchSettleRequest struct {
	RewardIDs []uint `json:"reward_ids"`
	Note      string `json:"note"`
}

type OfficialRewardBatchListFilters struct {
	Status string
}

type OfficialRewardBatchSettleResult struct {
	Count int64                                `json:"count"`
	Batch *models.AnalystRewardSettlementBatch `json:"batch,omitempty"`
}

type OfficialTaskSubmitRequest struct {
	ReportID                 uint   `json:"report_id"`
	AnalysisID               uint   `json:"analysis_id"`
	VideoFileURL             string `json:"video_file_url"`
	VideoAuthorizationStatus string `json:"video_authorization_status"`
	ScriptText               string `json:"script_text"`
	Summary                  string `json:"summary"`
	SubmitNote               string `json:"submit_note"`
}

type OfficialSubmissionReviewRequest struct {
	Status     models.OfficialAnalysisSubmissionStatus `json:"status"`
	ReviewNote string                                  `json:"review_note"`
	IsPublic   bool                                    `json:"is_public"`
}

type OfficialSubmissionAdoptionRequest struct {
	AdoptionStatus models.OfficialContentAdoptionStatus `json:"adoption_status"`
	Channel        string                               `json:"channel"`
	WorkTitle      string                               `json:"work_title"`
	WorkSummary    string                               `json:"work_summary"`
	CoverURL       string                               `json:"cover_url"`
	BonusBasis     string                               `json:"bonus_basis"`
	RewardAmount   float64                              `json:"reward_amount"`
	AdoptionNote   string                               `json:"adoption_note"`
	IsPublic       bool                                 `json:"is_public"`
}

type OfficialPlaybackBonusRequest struct {
	Amount     float64 `json:"amount"`
	BonusBasis string  `json:"bonus_basis"`
	Note       string  `json:"note"`
}

type OfficialAdoptionPublicRequest struct {
	IsPublic bool `json:"is_public"`
}

func (s *OfficialAnalysisTaskService) CreateTask(req OfficialAnalysisTaskRequest, adminID uint, now time.Time) (*models.OfficialAnalysisTask, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("db is nil")
	}
	task, err := buildOfficialTask(req, adminID, now)
	if err != nil {
		return nil, err
	}
	task.TaskNo = generateOfficialTaskNo(adminID, now)
	if err := s.db.Create(task).Error; err != nil {
		return nil, err
	}
	return task, nil
}

func (s *OfficialAnalysisTaskService) CreateTasksBatch(req OfficialAnalysisTaskBatchCreateRequest, adminID uint, now time.Time) (*OfficialAnalysisTaskBatchCreateResult, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("db is nil")
	}
	if len(req.Tasks) == 0 {
		return nil, fmt.Errorf("%w: 至少需要添加一场比赛", ErrOfficialTaskInvalid)
	}
	if len(req.Tasks) > 50 {
		return nil, fmt.Errorf("%w: 单次最多创建50个官方选题", ErrOfficialTaskInvalid)
	}

	created := make([]models.OfficialAnalysisTask, 0, len(req.Tasks))
	err := s.db.Transaction(func(tx *gorm.DB) error {
		for index, item := range req.Tasks {
			merged := mergeOfficialBatchTask(req.Common, item, req.EventName, index+1)
			taskNow := now.Add(time.Duration(index) * time.Microsecond)
			task, err := buildOfficialTask(merged, adminID, taskNow)
			if err != nil {
				return err
			}
			task.TaskNo = generateOfficialTaskNo(adminID, taskNow)
			if req.PublishAfterCreate {
				if err := validateOfficialTaskForPublish(task); err != nil {
					return err
				}
				task.Status = models.OfficialAnalysisTaskPublished
			}
			if err := tx.Create(task).Error; err != nil {
				return err
			}
			created = append(created, *task)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &OfficialAnalysisTaskBatchCreateResult{Created: created, Total: len(created)}, nil
}

func (s *OfficialAnalysisTaskService) UpdateTask(taskID uint, req OfficialAnalysisTaskRequest, now time.Time) (*models.OfficialAnalysisTask, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("db is nil")
	}

	var task models.OfficialAnalysisTask
	if err := s.db.First(&task, taskID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrOfficialTaskNotFound
		}
		return nil, err
	}
	if task.Status == models.OfficialAnalysisTaskClosed || task.Status == models.OfficialAnalysisTaskExpired {
		return nil, ErrOfficialTaskUnavailable
	}

	updates, err := buildOfficialTaskUpdates(req)
	if err != nil {
		return nil, err
	}
	updates["updated_at"] = now
	if err := s.db.Model(&task).Updates(updates).Error; err != nil {
		return nil, err
	}
	if err := s.db.First(&task, taskID).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *OfficialAnalysisTaskService) PublishTask(taskID, adminID uint, now time.Time) (*models.OfficialAnalysisTask, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("db is nil")
	}

	var task models.OfficialAnalysisTask
	if err := s.db.First(&task, taskID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrOfficialTaskNotFound
		}
		return nil, err
	}
	if task.Status != models.OfficialAnalysisTaskDraft {
		return nil, ErrOfficialTaskUnavailable
	}
	if err := validateOfficialTaskForPublish(&task); err != nil {
		return nil, err
	}
	if err := s.db.Model(&task).Updates(map[string]interface{}{
		"status":     models.OfficialAnalysisTaskPublished,
		"updated_at": now,
		"created_by": firstNonZero(task.CreatedBy, adminID),
	}).Error; err != nil {
		return nil, err
	}
	if err := s.db.First(&task, taskID).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *OfficialAnalysisTaskService) CloseTask(taskID, adminID uint, now time.Time) (*models.OfficialAnalysisTask, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("db is nil")
	}

	var task models.OfficialAnalysisTask
	if err := s.db.First(&task, taskID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrOfficialTaskNotFound
		}
		return nil, err
	}
	if task.Status == models.OfficialAnalysisTaskClosed {
		return &task, nil
	}
	if err := s.db.Model(&task).Updates(map[string]interface{}{
		"status":     models.OfficialAnalysisTaskClosed,
		"closed_by":  adminID,
		"closed_at":  &now,
		"updated_at": now,
	}).Error; err != nil {
		return nil, err
	}
	if err := s.db.First(&task, taskID).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *OfficialAnalysisTaskService) ListAdminTasks(page, pageSize int, filters OfficialTaskListFilters) ([]models.OfficialAnalysisTask, int64, error) {
	if s == nil || s.db == nil {
		return nil, 0, errors.New("db is nil")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	query := s.db.Model(&models.OfficialAnalysisTask{})
	if strings.TrimSpace(filters.Status) != "" {
		query = query.Where("status = ?", strings.TrimSpace(filters.Status))
	}
	if filters.TaskID != 0 {
		query = query.Where("id = ?", filters.TaskID)
	}
	if strings.TrimSpace(filters.MatchName) != "" {
		query = query.Where("match_name LIKE ?", "%"+strings.TrimSpace(filters.MatchName)+"%")
	}
	if strings.TrimSpace(filters.AgeGroup) != "" {
		query = query.Where("age_group = ?", strings.TrimSpace(filters.AgeGroup))
	}
	if strings.TrimSpace(filters.VisibleLevelMin) != "" {
		query = query.Where("visible_level_min = ?", strings.TrimSpace(filters.VisibleLevelMin))
	}
	if keyword := strings.TrimSpace(filters.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("title LIKE ? OR match_name LIKE ? OR target_player_name LIKE ?", like, like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var tasks []models.OfficialAnalysisTask
	if err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&tasks).Error; err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}

func (s *OfficialAnalysisTaskService) GetAdminTask(taskID uint) (*models.OfficialAnalysisTask, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("db is nil")
	}
	var task models.OfficialAnalysisTask
	if err := s.db.First(&task, taskID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrOfficialTaskNotFound
		}
		return nil, err
	}
	return &task, nil
}

func (s *OfficialAnalysisTaskService) ListTaskAcceptances(taskID uint, page, pageSize int) ([]models.OfficialAnalysisTaskAcceptance, int64, error) {
	if s == nil || s.db == nil {
		return nil, 0, errors.New("db is nil")
	}
	if taskID == 0 {
		return nil, 0, ErrOfficialTaskNotFound
	}
	var taskCount int64
	if err := s.db.Model(&models.OfficialAnalysisTask{}).Where("id = ?", taskID).Count(&taskCount).Error; err != nil {
		return nil, 0, err
	}
	if taskCount == 0 {
		return nil, 0, ErrOfficialTaskNotFound
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	query := s.db.Model(&models.OfficialAnalysisTaskAcceptance{}).Where("task_id = ?", taskID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var acceptances []models.OfficialAnalysisTaskAcceptance
	if err := query.Preload("Task").Preload("Analyst").Preload("Analyst.User").
		Preload("Submissions", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Order("accepted_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&acceptances).Error; err != nil {
		return nil, 0, err
	}
	return acceptances, total, nil
}

func (s *OfficialAnalysisTaskService) GetSubmission(submissionID uint) (*models.OfficialAnalysisSubmission, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("db is nil")
	}
	var submission models.OfficialAnalysisSubmission
	if err := s.db.Preload("Task").Preload("Analyst").Preload("Analyst.User").First(&submission, submissionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrOfficialSubmissionNotFound
		}
		return nil, err
	}
	return &submission, nil
}

func (s *OfficialAnalysisTaskService) SubmitTask(analystID, taskID uint, req OfficialTaskSubmitRequest, now time.Time) (*models.OfficialAnalysisSubmission, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("db is nil")
	}
	if err := validateOfficialSubmissionRequest(req); err != nil {
		return nil, err
	}

	var created models.OfficialAnalysisSubmission
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if _, err := s.findActiveAnalystWithTx(tx, analystID); err != nil {
			return err
		}

		var task models.OfficialAnalysisTask
		if err := tx.First(&task, taskID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrOfficialTaskNotFound
			}
			return err
		}
		if task.Status != models.OfficialAnalysisTaskPublished && task.Status != models.OfficialAnalysisTaskFull {
			return ErrOfficialTaskUnavailable
		}
		if task.Deadline != nil && !task.Deadline.After(now) {
			return ErrOfficialTaskUnavailable
		}

		var acceptance models.OfficialAnalysisTaskAcceptance
		if err := tx.Where("task_id = ? AND analyst_id = ? AND status IN ?", taskID, analystID, []models.OfficialAnalysisAcceptanceStatus{
			models.OfficialAnalysisAcceptanceAccepted,
			models.OfficialAnalysisAcceptanceSubmitted,
		}).First(&acceptance).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrOfficialTaskUnavailable
			}
			return err
		}

		var blockingCount int64
		if err := tx.Model(&models.OfficialAnalysisSubmission{}).
			Where("acceptance_id = ? AND status IN ?", acceptance.ID, []models.OfficialAnalysisSubmissionStatus{
				models.OfficialAnalysisSubmissionSubmitted,
				models.OfficialAnalysisSubmissionApproved,
				models.OfficialAnalysisSubmissionAdopted,
			}).
			Count(&blockingCount).Error; err != nil {
			return err
		}
		if blockingCount > 0 {
			return ErrOfficialSubmissionInvalid
		}

		created = models.OfficialAnalysisSubmission{
			TaskID:                   taskID,
			AnalystID:                analystID,
			AcceptanceID:             acceptance.ID,
			ReportID:                 req.ReportID,
			AnalysisID:               req.AnalysisID,
			VideoFileURL:             strings.TrimSpace(req.VideoFileURL),
			VideoAuthorizationStatus: strings.TrimSpace(req.VideoAuthorizationStatus),
			ScriptText:               strings.TrimSpace(req.ScriptText),
			Summary:                  strings.TrimSpace(req.Summary),
			SubmitNote:               strings.TrimSpace(req.SubmitNote),
			Status:                   models.OfficialAnalysisSubmissionSubmitted,
			CreatedAt:                now,
			UpdatedAt:                now,
		}
		if err := tx.Create(&created).Error; err != nil {
			return err
		}

		if err := tx.Model(&acceptance).Updates(map[string]interface{}{
			"status":       models.OfficialAnalysisAcceptanceSubmitted,
			"submitted_at": &now,
			"updated_at":   now,
		}).Error; err != nil {
			return err
		}

		if task.BaseRewardAmount > 0 {
			var existingReward int64
			if err := tx.Model(&models.AnalystRewardRecord{}).
				Where("source_type = ? AND source_id = ? AND reward_type = ?", "official_task_acceptance", acceptance.ID, "base").
				Count(&existingReward).Error; err != nil {
				return err
			}
			if existingReward == 0 {
				reward := models.AnalystRewardRecord{
					AnalystID:  analystID,
					SourceType: "official_task_acceptance",
					SourceID:   acceptance.ID,
					RewardType: "base",
					Amount:     task.BaseRewardAmount,
					Status:     models.AnalystRewardPending,
					Note:       "官方选题基础报酬",
					CreatedAt:  now,
					UpdatedAt:  now,
				}
				if err := tx.Create(&reward).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (s *OfficialAnalysisTaskService) ListAdminSubmissions(page, pageSize int, filters OfficialTaskListFilters) ([]models.OfficialAnalysisSubmission, int64, error) {
	if s == nil || s.db == nil {
		return nil, 0, errors.New("db is nil")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	query := s.db.Model(&models.OfficialAnalysisSubmission{})
	if filters.TaskID != 0 {
		query = query.Where("task_id = ?", filters.TaskID)
	}
	if strings.TrimSpace(filters.Status) != "" {
		query = query.Where("status = ?", strings.TrimSpace(filters.Status))
	}
	if filters.AnalystID != 0 {
		query = query.Where("analyst_id = ?", filters.AnalystID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var submissions []models.OfficialAnalysisSubmission
	if err := query.Preload("Task").Preload("Analyst").Preload("Analyst.User").
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&submissions).Error; err != nil {
		return nil, 0, err
	}
	return submissions, total, nil
}

func (s *OfficialAnalysisTaskService) ReviewSubmission(submissionID, adminID uint, req OfficialSubmissionReviewRequest, now time.Time) (*models.OfficialAnalysisSubmission, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("db is nil")
	}
	if req.Status != models.OfficialAnalysisSubmissionApproved &&
		req.Status != models.OfficialAnalysisSubmissionRevisionRequired &&
		req.Status != models.OfficialAnalysisSubmissionRejected {
		return nil, ErrOfficialSubmissionInvalid
	}

	var submission models.OfficialAnalysisSubmission
	if err := s.db.First(&submission, submissionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrOfficialSubmissionNotFound
		}
		return nil, err
	}
	if submission.Status != models.OfficialAnalysisSubmissionSubmitted && submission.Status != models.OfficialAnalysisSubmissionRevisionRequired {
		return nil, ErrOfficialSubmissionInvalid
	}

	if err := s.db.Model(&submission).Updates(map[string]interface{}{
		"status":      req.Status,
		"reviewed_by": adminID,
		"reviewed_at": &now,
		"review_note": strings.TrimSpace(req.ReviewNote),
		"is_public":   req.IsPublic,
		"updated_at":  now,
	}).Error; err != nil {
		return nil, err
	}
	if err := s.db.First(&submission, submissionID).Error; err != nil {
		return nil, err
	}
	return &submission, nil
}

func (s *OfficialAnalysisTaskService) AdoptSubmission(submissionID, adminID uint, req OfficialSubmissionAdoptionRequest, now time.Time) (*models.OfficialContentAdoption, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("db is nil")
	}
	if err := validateOfficialAdoptionRequest(req); err != nil {
		return nil, err
	}

	var adoption models.OfficialContentAdoption
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var submission models.OfficialAnalysisSubmission
		if err := tx.First(&submission, submissionID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrOfficialSubmissionNotFound
			}
			return err
		}
		if submission.Status != models.OfficialAnalysisSubmissionSubmitted &&
			submission.Status != models.OfficialAnalysisSubmissionApproved {
			return ErrOfficialSubmissionInvalid
		}

		var existing int64
		if err := tx.Model(&models.OfficialContentAdoption{}).Where("submission_id = ?", submissionID).Count(&existing).Error; err != nil {
			return err
		}
		if existing > 0 {
			return ErrOfficialTaskDuplicate
		}

		adoption = models.OfficialContentAdoption{
			TaskID:         submission.TaskID,
			SubmissionID:   submission.ID,
			AnalystID:      submission.AnalystID,
			AdoptionStatus: req.AdoptionStatus,
			Channel:        strings.TrimSpace(req.Channel),
			WorkTitle:      strings.TrimSpace(req.WorkTitle),
			WorkSummary:    strings.TrimSpace(req.WorkSummary),
			CoverURL:       strings.TrimSpace(req.CoverURL),
			BonusBasis:     strings.TrimSpace(req.BonusBasis),
			RewardAmount:   req.RewardAmount,
			AdoptionNote:   strings.TrimSpace(req.AdoptionNote),
			IsPublic:       req.IsPublic,
			CreatedBy:      adminID,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := tx.Create(&adoption).Error; err != nil {
			return err
		}

		if err := tx.Model(&submission).Updates(map[string]interface{}{
			"status":      models.OfficialAnalysisSubmissionAdopted,
			"reviewed_by": adminID,
			"reviewed_at": &now,
			"review_note": strings.TrimSpace(req.AdoptionNote),
			"is_public":   req.IsPublic,
			"updated_at":  now,
		}).Error; err != nil {
			return err
		}

		statUpdates := map[string]interface{}{
			"official_adoption_count": gorm.Expr("official_adoption_count + ?", 1),
			"updated_at":              now,
		}
		if req.AdoptionStatus == models.OfficialContentAdoptionMaterial {
			statUpdates["official_material_count"] = gorm.Expr("official_material_count + ?", 1)
		}
		if req.AdoptionStatus == models.OfficialContentAdoptionOfficialPublished ||
			req.AdoptionStatus == models.OfficialContentAdoptionKeySpread ||
			req.AdoptionStatus == models.OfficialContentAdoptionLongTerm {
			statUpdates["official_publish_count"] = gorm.Expr("official_publish_count + ?", 1)
		}
		if err := tx.Model(&models.Analyst{}).Where("id = ?", submission.AnalystID).Updates(statUpdates).Error; err != nil {
			return err
		}

		if req.RewardAmount > 0 {
			reward := models.AnalystRewardRecord{
				AnalystID:  submission.AnalystID,
				SourceType: "official_adoption",
				SourceID:   adoption.ID,
				RewardType: "adoption",
				Amount:     req.RewardAmount,
				Status:     models.AnalystRewardPending,
				Note:       strings.TrimSpace(req.BonusBasis),
				CreatedAt:  now,
				UpdatedAt:  now,
			}
			if err := tx.Create(&reward).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return &adoption, nil
}

func (s *OfficialAnalysisTaskService) ListAvailableTasksForAnalyst(analystID uint, page, pageSize int, now time.Time) ([]models.OfficialAnalysisTask, int64, error) {
	if s == nil || s.db == nil {
		return nil, 0, errors.New("db is nil")
	}
	analyst, err := s.findActiveAnalyst(analystID)
	if err != nil {
		return nil, 0, err
	}
	levelCode := analyst.LevelCode
	if levelCode == "" {
		levelCode = models.DefaultAnalystLevelCode
	}

	allowedLevels := make([]string, 0, len(models.DefaultAnalystLevels()))
	for _, level := range models.DefaultAnalystLevels() {
		if models.AnalystLevelMeets(levelCode, level.Code) {
			allowedLevels = append(allowedLevels, level.Code)
		}
	}
	visibleConditions := make([]string, 0, len(allowedLevels)+1)
	visibleArgs := make([]interface{}, 0, len(allowedLevels)*4+1)
	visibleConditions = append(visibleConditions, "((visible_level_codes IS NULL OR visible_level_codes = '') AND visible_level_min IN ?)")
	visibleArgs = append(visibleArgs, allowedLevels)
	for _, level := range allowedLevels {
		visibleConditions = append(visibleConditions, "visible_level_codes = ? OR visible_level_codes LIKE ? OR visible_level_codes LIKE ? OR visible_level_codes LIKE ?")
		visibleArgs = append(visibleArgs, level, level+",%", "%,"+level+",%", "%,"+level)
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	query := s.db.Model(&models.OfficialAnalysisTask{}).
		Where("status = ?", models.OfficialAnalysisTaskPublished).
		Where("max_accept_count > current_accept_count").
		Where("("+strings.Join(visibleConditions, " OR ")+")", visibleArgs...).
		Where("(priority_until IS NULL OR priority_until <= ? OR priority_level_min = '' OR priority_level_min IN ?)", now, allowedLevels).
		Where("(deadline IS NULL OR deadline > ?)", now).
		Where("id NOT IN (?)",
			s.db.Model(&models.OfficialAnalysisTaskAcceptance{}).
				Select("task_id").
				Where("analyst_id = ? AND status IN ?", analystID, []models.OfficialAnalysisAcceptanceStatus{
					models.OfficialAnalysisAcceptanceAccepted,
					models.OfficialAnalysisAcceptanceSubmitted,
				}),
		)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var tasks []models.OfficialAnalysisTask
	if err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&tasks).Error; err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}

func (s *OfficialAnalysisTaskService) GetAvailableTaskForAnalyst(analystID, taskID uint, now time.Time) (*models.OfficialAnalysisTask, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("db is nil")
	}
	analyst, err := s.findActiveAnalyst(analystID)
	if err != nil {
		return nil, err
	}
	levelCode := analyst.LevelCode
	if levelCode == "" {
		levelCode = models.DefaultAnalystLevelCode
	}
	var task models.OfficialAnalysisTask
	if err := s.db.First(&task, taskID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrOfficialTaskNotFound
		}
		return nil, err
	}
	if task.Status != models.OfficialAnalysisTaskPublished && task.Status != models.OfficialAnalysisTaskFull {
		return nil, ErrOfficialTaskUnavailable
	}
	if task.Deadline != nil && !task.Deadline.After(now) {
		return nil, ErrOfficialTaskUnavailable
	}
	if !officialTaskVisibleToLevel(&task, levelCode) {
		return nil, ErrOfficialTaskLevelDenied
	}
	if task.PriorityUntil != nil && task.PriorityUntil.After(now) && strings.TrimSpace(task.PriorityLevelMin) != "" &&
		!models.AnalystLevelMeets(levelCode, task.PriorityLevelMin) {
		return nil, ErrOfficialTaskLevelDenied
	}
	return &task, nil
}

func (s *OfficialAnalysisTaskService) ListMyAcceptances(analystID uint, page, pageSize int) ([]models.OfficialAnalysisTaskAcceptance, int64, error) {
	if s == nil || s.db == nil {
		return nil, 0, errors.New("db is nil")
	}
	if _, err := s.findActiveAnalyst(analystID); err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	query := s.db.Model(&models.OfficialAnalysisTaskAcceptance{}).
		Where("analyst_id = ?", analystID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var acceptances []models.OfficialAnalysisTaskAcceptance
	if err := query.Preload("Task").
		Preload("Submissions", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Order("accepted_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&acceptances).Error; err != nil {
		return nil, 0, err
	}
	return acceptances, total, nil
}

func (s *OfficialAnalysisTaskService) ListAdminRewards(page, pageSize int, filters OfficialRewardListFilters) ([]models.AnalystRewardRecord, int64, error) {
	if s == nil || s.db == nil {
		return nil, 0, errors.New("db is nil")
	}
	query := s.buildRewardQuery(filters)
	return s.listRewards(query, page, pageSize)
}

func (s *OfficialAnalysisTaskService) ListRewardSettlementBatches(page, pageSize int, filters OfficialRewardBatchListFilters) ([]models.AnalystRewardSettlementBatch, int64, error) {
	if s == nil || s.db == nil {
		return nil, 0, errors.New("db is nil")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	query := s.db.Model(&models.AnalystRewardSettlementBatch{})
	if strings.TrimSpace(filters.Status) != "" {
		query = query.Where("status = ?", strings.TrimSpace(filters.Status))
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var batches []models.AnalystRewardSettlementBatch
	if err := query.Order("settled_at DESC, id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&batches).Error; err != nil {
		return nil, 0, err
	}
	return batches, total, nil
}

func (s *OfficialAnalysisTaskService) GetRewardSettlementBatch(batchID uint) (*models.AnalystRewardSettlementBatch, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("db is nil")
	}
	var batch models.AnalystRewardSettlementBatch
	if err := s.db.Preload("Rewards", func(db *gorm.DB) *gorm.DB {
		return db.Order("id ASC")
	}).Preload("Rewards.Analyst").Preload("Rewards.Analyst.User").First(&batch, batchID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrAnalystRewardNotFound
		}
		return nil, err
	}
	return &batch, nil
}

func (s *OfficialAnalysisTaskService) ListAnalystRewards(analystID uint, page, pageSize int) ([]models.AnalystRewardRecord, int64, error) {
	if s == nil || s.db == nil {
		return nil, 0, errors.New("db is nil")
	}
	if _, err := s.findActiveAnalyst(analystID); err != nil {
		return nil, 0, err
	}
	query := s.buildRewardQuery(OfficialRewardListFilters{AnalystID: analystID})
	return s.listRewards(query, page, pageSize)
}

func (s *OfficialAnalysisTaskService) ListAnalystAdoptions(analystID uint, page, pageSize int) ([]models.OfficialContentAdoption, int64, error) {
	if s == nil || s.db == nil {
		return nil, 0, errors.New("db is nil")
	}
	if _, err := s.findActiveAnalyst(analystID); err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	query := s.db.Model(&models.OfficialContentAdoption{}).Where("analyst_id = ?", analystID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var adoptions []models.OfficialContentAdoption
	if err := query.Preload("Task").Preload("Submission").
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&adoptions).Error; err != nil {
		return nil, 0, err
	}
	return adoptions, total, nil
}

func (s *OfficialAnalysisTaskService) ListOfficialMaterials(page, pageSize int, filters OfficialMaterialListFilters) ([]models.OfficialContentAdoption, int64, error) {
	if s == nil || s.db == nil {
		return nil, 0, errors.New("db is nil")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	query := s.db.Model(&models.OfficialContentAdoption{}).
		Joins("LEFT JOIN official_analysis_tasks ON official_analysis_tasks.id = official_content_adoptions.task_id").
		Joins("LEFT JOIN official_analysis_submissions ON official_analysis_submissions.id = official_content_adoptions.submission_id").
		Joins("LEFT JOIN analysts ON analysts.id = official_content_adoptions.analyst_id")
	if strings.TrimSpace(filters.Keyword) != "" {
		keyword := "%" + strings.TrimSpace(filters.Keyword) + "%"
		query = query.Where(
			"official_content_adoptions.work_title LIKE ? OR official_content_adoptions.work_summary LIKE ? OR official_analysis_submissions.summary LIKE ? OR official_analysis_submissions.script_text LIKE ?",
			keyword, keyword, keyword, keyword,
		)
	}
	if strings.TrimSpace(filters.MatchName) != "" {
		query = query.Where("official_analysis_tasks.match_name LIKE ?", "%"+strings.TrimSpace(filters.MatchName)+"%")
	}
	if strings.TrimSpace(filters.AgeGroup) != "" {
		query = query.Where("official_analysis_tasks.age_group = ?", strings.TrimSpace(filters.AgeGroup))
	}
	if strings.TrimSpace(filters.Channel) != "" {
		query = query.Where("official_content_adoptions.channel = ?", strings.TrimSpace(filters.Channel))
	}
	if strings.TrimSpace(filters.AdoptionStatus) != "" {
		query = query.Where("official_content_adoptions.adoption_status = ?", strings.TrimSpace(filters.AdoptionStatus))
	}
	switch strings.TrimSpace(filters.AssetKind) {
	case "video":
		query = query.Where("official_analysis_submissions.video_file_url <> ''")
	case "script":
		query = query.Where("official_analysis_submissions.script_text <> ''")
	case "report":
		query = query.Where("official_analysis_submissions.report_id > 0")
	case "analysis":
		query = query.Where("official_analysis_submissions.analysis_id > 0")
	case "summary":
		query = query.Where("official_content_adoptions.work_summary <> '' OR official_analysis_submissions.summary <> ''")
	}
	if filters.PublishReady {
		query = query.Where("official_content_adoptions.adoption_status IN ?", []models.OfficialContentAdoptionStatus{
			models.OfficialContentAdoptionOfficialPublished,
			models.OfficialContentAdoptionKeySpread,
			models.OfficialContentAdoptionLongTerm,
		}).
			Where("official_analysis_submissions.video_authorization_status NOT IN ?", []string{"", "unknown", "rejected"}).
			Where("official_content_adoptions.work_summary <> '' OR official_analysis_submissions.summary <> '' OR official_analysis_submissions.script_text <> '' OR official_analysis_submissions.video_file_url <> ''")
	}
	switch strings.TrimSpace(filters.PublishStatus) {
	case "published":
		query = query.Where("EXISTS (SELECT 1 FROM official_content_publish_records WHERE official_content_publish_records.adoption_id = official_content_adoptions.id)")
	case "unpublished":
		query = query.Where("NOT EXISTS (SELECT 1 FROM official_content_publish_records WHERE official_content_publish_records.adoption_id = official_content_adoptions.id)")
	}
	switch strings.TrimSpace(filters.MetricsStatus) {
	case "has_metrics":
		query = query.Where(`EXISTS (
			SELECT 1 FROM official_content_publish_records
			WHERE official_content_publish_records.adoption_id = official_content_adoptions.id
				AND (play_count > 0 OR like_count > 0 OR comment_count > 0 OR share_count > 0 OR favorite_count > 0)
		)`)
	case "missing_metrics":
		query = query.Where("EXISTS (SELECT 1 FROM official_content_publish_records WHERE official_content_publish_records.adoption_id = official_content_adoptions.id)").
			Where(`NOT EXISTS (
				SELECT 1 FROM official_content_publish_records
				WHERE official_content_publish_records.adoption_id = official_content_adoptions.id
					AND (play_count > 0 OR like_count > 0 OR comment_count > 0 OR share_count > 0 OR favorite_count > 0)
			)`)
	case "stale_metrics":
		staleBefore := time.Now().AddDate(0, 0, -7)
		query = query.Where("EXISTS (SELECT 1 FROM official_content_publish_records WHERE official_content_publish_records.adoption_id = official_content_adoptions.id)").
			Where(`NOT EXISTS (
				SELECT 1 FROM official_content_publish_records
				WHERE official_content_publish_records.adoption_id = official_content_adoptions.id
					AND metrics_recorded_at IS NOT NULL
					AND metrics_recorded_at >= ?
			)`, staleBefore)
	}
	switch strings.TrimSpace(filters.BonusStatus) {
	case "with_playback_bonus":
		query = query.Where(`EXISTS (
			SELECT 1 FROM analyst_reward_records
			WHERE analyst_reward_records.source_type = 'official_adoption'
				AND analyst_reward_records.source_id = official_content_adoptions.id
				AND analyst_reward_records.reward_type = 'playback_bonus'
				AND analyst_reward_records.status NOT IN ?
		)`, []models.AnalystRewardStatus{models.AnalystRewardCancelled, models.AnalystRewardReversed})
	case "without_playback_bonus":
		query = query.Where(`NOT EXISTS (
			SELECT 1 FROM analyst_reward_records
			WHERE analyst_reward_records.source_type = 'official_adoption'
				AND analyst_reward_records.source_id = official_content_adoptions.id
				AND analyst_reward_records.reward_type = 'playback_bonus'
				AND analyst_reward_records.status NOT IN ?
		)`, []models.AnalystRewardStatus{models.AnalystRewardCancelled, models.AnalystRewardReversed})
	}
	if filters.AnalystID != 0 {
		query = query.Where("official_content_adoptions.analyst_id = ?", filters.AnalystID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var materials []models.OfficialContentAdoption
	orderExpr := officialMaterialOrderExpr(filters.SortBy)
	if err := query.Preload("Task").Preload("Submission").Preload("Analyst").Preload("Analyst.User").
		Preload("PublishRecords", func(db *gorm.DB) *gorm.DB {
			return db.Order("published_at DESC, created_at DESC")
		}).
		Order(orderExpr).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&materials).Error; err != nil {
		return nil, 0, err
	}
	return materials, total, nil
}

func officialMaterialOrderExpr(sortBy string) string {
	switch strings.TrimSpace(sortBy) {
	case "total_play_desc":
		return "(SELECT COALESCE(SUM(play_count), 0) FROM official_content_publish_records WHERE official_content_publish_records.adoption_id = official_content_adoptions.id) DESC, official_content_adoptions.created_at DESC"
	case "max_play_desc":
		return "(SELECT COALESCE(MAX(play_count), 0) FROM official_content_publish_records WHERE official_content_publish_records.adoption_id = official_content_adoptions.id) DESC, official_content_adoptions.created_at DESC"
	case "latest_publish_desc":
		return "(SELECT MAX(COALESCE(published_at, created_at)) FROM official_content_publish_records WHERE official_content_publish_records.adoption_id = official_content_adoptions.id) DESC, official_content_adoptions.created_at DESC"
	default:
		return "official_content_adoptions.created_at DESC"
	}
}

func (s *OfficialAnalysisTaskService) CreatePublishRecord(adoptionID, adminID uint, req OfficialPublishRecordRequest, now time.Time) (*models.OfficialContentPublishRecord, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("db is nil")
	}
	if hasNegativePublishMetrics(req) {
		return nil, ErrOfficialTaskInvalid
	}
	if adoptionID == 0 {
		return nil, ErrOfficialAdoptionNotFound
	}
	var adoption models.OfficialContentAdoption
	if err := s.db.First(&adoption, adoptionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrOfficialAdoptionNotFound
		}
		return nil, err
	}
	channel := strings.TrimSpace(req.Channel)
	if channel == "" {
		channel = strings.TrimSpace(adoption.Channel)
	}
	if channel == "" {
		return nil, ErrOfficialTaskInvalid
	}
	publishedAt := req.PublishedAt
	if publishedAt == nil {
		publishedAt = &now
	}
	record := models.OfficialContentPublishRecord{
		AdoptionID:    adoption.ID,
		Channel:       channel,
		AccountName:   strings.TrimSpace(req.AccountName),
		PublishURL:    strings.TrimSpace(req.PublishURL),
		PublishTitle:  strings.TrimSpace(req.PublishTitle),
		ReusePurpose:  strings.TrimSpace(req.ReusePurpose),
		PublishedAt:   publishedAt,
		PlayCount:     req.PlayCount,
		LikeCount:     req.LikeCount,
		CommentCount:  req.CommentCount,
		ShareCount:    req.ShareCount,
		FavoriteCount: req.FavoriteCount,
		MetricsAt:     req.MetricsAt,
		Note:          strings.TrimSpace(req.Note),
		CreatedBy:     adminID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.db.Create(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (s *OfficialAnalysisTaskService) UpdatePublishRecord(adoptionID, recordID uint, req OfficialPublishRecordRequest, now time.Time) (*models.OfficialContentPublishRecord, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("db is nil")
	}
	if adoptionID == 0 || recordID == 0 {
		return nil, ErrOfficialPublishRecordNotFound
	}
	if hasNegativePublishMetrics(req) {
		return nil, ErrOfficialTaskInvalid
	}
	var adoption models.OfficialContentAdoption
	if err := s.db.First(&adoption, adoptionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrOfficialAdoptionNotFound
		}
		return nil, err
	}
	channel := strings.TrimSpace(req.Channel)
	if channel == "" {
		channel = strings.TrimSpace(adoption.Channel)
	}
	if channel == "" {
		return nil, ErrOfficialTaskInvalid
	}
	var record models.OfficialContentPublishRecord
	if err := s.db.Where("id = ? AND adoption_id = ?", recordID, adoptionID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrOfficialPublishRecordNotFound
		}
		return nil, err
	}
	updates := map[string]interface{}{
		"channel":             channel,
		"account_name":        strings.TrimSpace(req.AccountName),
		"publish_url":         strings.TrimSpace(req.PublishURL),
		"publish_title":       strings.TrimSpace(req.PublishTitle),
		"reuse_purpose":       strings.TrimSpace(req.ReusePurpose),
		"published_at":        req.PublishedAt,
		"play_count":          req.PlayCount,
		"like_count":          req.LikeCount,
		"comment_count":       req.CommentCount,
		"share_count":         req.ShareCount,
		"favorite_count":      req.FavoriteCount,
		"metrics_recorded_at": req.MetricsAt,
		"note":                strings.TrimSpace(req.Note),
		"updated_at":          now,
	}
	if err := s.db.Model(&record).Updates(updates).Error; err != nil {
		return nil, err
	}
	if err := s.db.First(&record, record.ID).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (s *OfficialAnalysisTaskService) DeletePublishRecord(adoptionID, recordID uint) error {
	if s == nil || s.db == nil {
		return errors.New("db is nil")
	}
	if adoptionID == 0 || recordID == 0 {
		return ErrOfficialPublishRecordNotFound
	}
	result := s.db.Where("id = ? AND adoption_id = ?", recordID, adoptionID).Delete(&models.OfficialContentPublishRecord{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrOfficialPublishRecordNotFound
	}
	return nil
}

func hasNegativePublishMetrics(req OfficialPublishRecordRequest) bool {
	return req.PlayCount < 0 || req.LikeCount < 0 || req.CommentCount < 0 || req.ShareCount < 0 || req.FavoriteCount < 0
}

func (s *OfficialAnalysisTaskService) CreatePlaybackBonus(adoptionID, adminID uint, req OfficialPlaybackBonusRequest, now time.Time) (*models.AnalystRewardRecord, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("db is nil")
	}
	if adoptionID == 0 || req.Amount <= 0 {
		return nil, ErrAnalystRewardInvalid
	}
	var adoption models.OfficialContentAdoption
	if err := s.db.First(&adoption, adoptionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrOfficialAdoptionNotFound
		}
		return nil, err
	}
	note := strings.TrimSpace(req.Note)
	basis := strings.TrimSpace(req.BonusBasis)
	if note == "" {
		note = basis
	} else if basis != "" {
		note = fmt.Sprintf("%s；依据：%s", note, basis)
	}
	reward := models.AnalystRewardRecord{
		AnalystID:  adoption.AnalystID,
		SourceType: "official_adoption",
		SourceID:   adoption.ID,
		RewardType: "playback_bonus",
		Amount:     req.Amount,
		Status:     models.AnalystRewardPending,
		Note:       note,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.db.Create(&reward).Error; err != nil {
		return nil, err
	}
	return &reward, nil
}

func (s *OfficialAnalysisTaskService) UpdateAdoptionPublic(adoptionID uint, req OfficialAdoptionPublicRequest, now time.Time) (*models.OfficialContentAdoption, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("db is nil")
	}
	var adoption models.OfficialContentAdoption
	if err := s.db.First(&adoption, adoptionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrOfficialAdoptionNotFound
		}
		return nil, err
	}
	if err := s.db.Model(&adoption).Updates(map[string]interface{}{
		"is_public":  req.IsPublic,
		"updated_at": now,
	}).Error; err != nil {
		return nil, err
	}
	if err := s.db.Preload("Task").Preload("Submission").Preload("Analyst").Preload("Analyst.User").First(&adoption, adoptionID).Error; err != nil {
		return nil, err
	}
	return &adoption, nil
}

func (s *OfficialAnalysisTaskService) ListAdminEventTopics(page, pageSize int, filters OfficialEventTopicFilters) ([]OfficialEventTopic, int64, error) {
	return s.ListPublicEventTopics(page, pageSize, filters)
}

func (s *OfficialAnalysisTaskService) SaveEventTopicConfig(req OfficialEventTopicConfigRequest, adminID uint, now time.Time) (*models.OfficialEventTopicConfig, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("db is nil")
	}
	matchName := strings.TrimSpace(req.MatchName)
	if matchName == "" {
		return nil, ErrOfficialTaskInvalid
	}
	aliasNames := sanitizeTopicAliases(req.AliasNames, matchName)
	if err := s.validateTopicAliases(matchName, aliasNames); err != nil {
		return nil, err
	}
	config := models.OfficialEventTopicConfig{}
	err := s.db.Where("match_name = ?", matchName).First(&config).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if req.PinnedAdoptionID != 0 {
		valid, err := s.isPinnedAdoptionValid(matchName, aliasNames, req.PinnedAdoptionID)
		if err != nil {
			return nil, err
		}
		if !valid {
			return nil, ErrOfficialTaskInvalid
		}
	}
	if err == gorm.ErrRecordNotFound {
		config = models.OfficialEventTopicConfig{
			MatchName:        matchName,
			DisplayName:      strings.TrimSpace(req.DisplayName),
			Summary:          strings.TrimSpace(req.Summary),
			CoverURL:         strings.TrimSpace(req.CoverURL),
			AliasNames:       aliasNames,
			PinnedAdoptionID: req.PinnedAdoptionID,
			IsFeatured:       req.IsFeatured,
			SortOrder:        req.SortOrder,
			CreatedBy:        adminID,
			UpdatedBy:        adminID,
		}
		if err := s.db.Create(&config).Error; err != nil {
			return nil, err
		}
		return &config, nil
	}
	if err := s.db.Model(&config).Updates(map[string]interface{}{
		"display_name":       strings.TrimSpace(req.DisplayName),
		"summary":            strings.TrimSpace(req.Summary),
		"cover_url":          strings.TrimSpace(req.CoverURL),
		"alias_names":        aliasNames,
		"pinned_adoption_id": req.PinnedAdoptionID,
		"is_featured":        req.IsFeatured,
		"sort_order":         req.SortOrder,
		"updated_by":         adminID,
		"updated_at":         now,
	}).Error; err != nil {
		return nil, err
	}
	if err := s.db.First(&config, config.ID).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

func (s *OfficialAnalysisTaskService) ListPublicEventTopics(page, pageSize int, filters OfficialEventTopicFilters) ([]OfficialEventTopic, int64, error) {
	if s == nil || s.db == nil {
		return nil, 0, errors.New("db is nil")
	}
	filters.Keyword = strings.TrimSpace(filters.Keyword)
	filters.AgeGroup = strings.TrimSpace(filters.AgeGroup)
	configs, err := s.loadEventTopicConfigs()
	if err != nil {
		return nil, 0, err
	}
	configByMatch, aliasToCanonical := buildEventTopicConfigIndexes(configs)
	adoptions, err := s.listPublicEventAdoptions(filters)
	if err != nil {
		return nil, 0, err
	}
	grouped := make(map[string]*OfficialEventTopic)
	analystsByMatch := make(map[string]map[uint]bool)
	ageByMatch := make(map[string]map[string]bool)
	order := make([]string, 0)
	for _, adoption := range adoptions {
		work := buildOfficialEventWork(adoption)
		if work.MatchName == "" {
			continue
		}
		canonical := resolveEventTopicMatchName(work.MatchName, aliasToCanonical)
		work.MatchName = canonical
		topic, ok := grouped[canonical]
		if !ok {
			topic = &OfficialEventTopic{
				MatchName:         canonical,
				DisplayName:       canonical,
				FeaturedWork:      &work,
				LatestPublishedAt: work.CreatedAt,
			}
			applyEventTopicConfig(topic, configByMatch[canonical])
			grouped[canonical] = topic
			analystsByMatch[canonical] = make(map[uint]bool)
			ageByMatch[canonical] = make(map[string]bool)
			order = append(order, canonical)
		}
		topic.WorkCount++
		if work.AnalystID != 0 {
			analystsByMatch[canonical][work.AnalystID] = true
		}
		if work.AgeGroup != "" {
			ageByMatch[canonical][work.AgeGroup] = true
		}
		if topic.PinnedAdoptionID == work.ID {
			pinned := work
			topic.FeaturedWork = &pinned
		}
	}

	topics := make([]OfficialEventTopic, 0, len(order))
	for _, matchName := range order {
		topic := grouped[matchName]
		topic.AnalystCount = len(analystsByMatch[matchName])
		topic.AgeGroups = mapKeys(ageByMatch[matchName])
		topic.ConflictWarnings = buildTopicConflictWarnings(topic.MatchName, topic.AliasNames, configs)
		if filters.FeaturedOnly && !topic.IsFeatured {
			continue
		}
		if filters.Keyword != "" && !topicMatchesKeyword(*topic, filters.Keyword) {
			continue
		}
		topics = append(topics, *topic)
	}
	sortEventTopics(topics)
	total := int64(len(topics))
	start, end := paginationBounds(page, pageSize, len(topics))
	if start >= end {
		return []OfficialEventTopic{}, total, nil
	}
	return topics[start:end], total, nil
}

func (s *OfficialAnalysisTaskService) GetPublicEventTopic(matchName string, page, pageSize int, ageGroup string) (*OfficialEventTopicDetail, int64, error) {
	if s == nil || s.db == nil {
		return nil, 0, errors.New("db is nil")
	}
	matchName = strings.TrimSpace(matchName)
	ageGroup = strings.TrimSpace(ageGroup)
	if matchName == "" {
		return nil, 0, ErrOfficialTaskNotFound
	}
	configs, err := s.loadEventTopicConfigs()
	if err != nil {
		return nil, 0, err
	}
	configByMatch, aliasToCanonical := buildEventTopicConfigIndexes(configs)
	canonical := resolveEventTopicMatchName(matchName, aliasToCanonical)
	adoptions, err := s.listPublicEventAdoptions(OfficialEventTopicFilters{})
	if err != nil {
		return nil, 0, err
	}
	works := make([]OfficialEventWork, 0, len(adoptions))
	analysts := make(map[uint]bool)
	ageGroups := make(map[string]bool)
	var featuredWork *OfficialEventWork
	for _, adoption := range adoptions {
		work := buildOfficialEventWork(adoption)
		if work.MatchName == "" {
			continue
		}
		resolved := resolveEventTopicMatchName(work.MatchName, aliasToCanonical)
		if resolved != canonical {
			continue
		}
		if work.AgeGroup != "" {
			ageGroups[work.AgeGroup] = true
		}
		if ageGroup != "" && !strings.EqualFold(work.AgeGroup, ageGroup) {
			continue
		}
		work.MatchName = canonical
		works = append(works, work)
		if work.AnalystID != 0 {
			analysts[work.AnalystID] = true
		}
		if featuredWork == nil {
			copyWork := work
			featuredWork = &copyWork
		}
	}
	if len(works) == 0 {
		return nil, 0, ErrOfficialTaskNotFound
	}
	if config := configByMatch[canonical]; config != nil && config.PinnedAdoptionID != 0 {
		for _, work := range works {
			if work.ID == config.PinnedAdoptionID {
				copyWork := work
				featuredWork = &copyWork
				break
			}
		}
	}
	total := int64(len(works))
	start, end := paginationBounds(page, pageSize, len(works))
	pagedWorks := []OfficialEventWork{}
	if start < end {
		pagedWorks = works[start:end]
	}
	detail := &OfficialEventTopicDetail{
		MatchName:      canonical,
		DisplayName:    canonical,
		AgeGroupFilter: ageGroup,
		AgeGroups:      mapKeys(ageGroups),
		WorkCount:      len(works),
		AnalystCount:   len(analysts),
		FeaturedWork:   featuredWork,
		Works:          pagedWorks,
	}
	applyEventTopicDetailConfig(detail, configByMatch[canonical])
	return detail, total, nil
}

func (s *OfficialAnalysisTaskService) listPublicEventAdoptions(filters OfficialEventTopicFilters) ([]models.OfficialContentAdoption, error) {
	query := s.db.Model(&models.OfficialContentAdoption{}).
		Joins("JOIN official_analysis_tasks ON official_analysis_tasks.id = official_content_adoptions.task_id").
		Where("official_content_adoptions.is_public = ?", true).
		Where("official_analysis_tasks.match_name <> ''")
	if filters.AgeGroup != "" {
		query = query.Where("official_analysis_tasks.age_group = ?", filters.AgeGroup)
	}
	var adoptions []models.OfficialContentAdoption
	if err := query.Preload("Task").Preload("Analyst").Preload("Analyst.User").
		Order("official_content_adoptions.created_at DESC").
		Limit(500).
		Find(&adoptions).Error; err != nil {
		return nil, err
	}
	return adoptions, nil
}

func (s *OfficialAnalysisTaskService) SettleReward(rewardID, adminID uint, req OfficialRewardActionRequest, now time.Time) (*models.AnalystRewardRecord, error) {
	result, err := s.settleRewardsWithBatch([]uint{rewardID}, adminID, strings.TrimSpace(req.Note), now)
	if err != nil {
		return nil, err
	}
	if result.Batch == nil || len(result.Batch.Rewards) == 0 {
		return nil, ErrAnalystRewardInvalid
	}
	reward := result.Batch.Rewards[0]
	return &reward, nil
}

func (s *OfficialAnalysisTaskService) BatchSettleRewards(req OfficialRewardBatchSettleRequest, adminID uint, now time.Time) (*OfficialRewardBatchSettleResult, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("db is nil")
	}
	if len(req.RewardIDs) == 0 {
		return nil, ErrAnalystRewardInvalid
	}
	return s.settleRewardsWithBatch(req.RewardIDs, adminID, strings.TrimSpace(req.Note), now)
}

func (s *OfficialAnalysisTaskService) ReverseReward(rewardID, adminID uint, req OfficialRewardActionRequest, now time.Time) (*models.AnalystRewardRecord, error) {
	return s.updateRewardStatus(rewardID, adminID, models.AnalystRewardReversed, strings.TrimSpace(req.Note), now, models.AnalystRewardPending, models.AnalystRewardSettled)
}

func (s *OfficialAnalysisTaskService) buildRewardQuery(filters OfficialRewardListFilters) *gorm.DB {
	query := s.db.Model(&models.AnalystRewardRecord{})
	if filters.AnalystID != 0 {
		query = query.Where("analyst_id = ?", filters.AnalystID)
	}
	if strings.TrimSpace(filters.Status) != "" {
		query = query.Where("status = ?", strings.TrimSpace(filters.Status))
	}
	if strings.TrimSpace(filters.RewardType) != "" {
		query = query.Where("reward_type = ?", strings.TrimSpace(filters.RewardType))
	}
	if strings.TrimSpace(filters.SourceType) != "" {
		query = query.Where("source_type = ?", strings.TrimSpace(filters.SourceType))
	}
	if filters.BatchID != 0 {
		query = query.Where("settlement_batch_id = ?", filters.BatchID)
	}
	return query
}

func (s *OfficialAnalysisTaskService) listRewards(query *gorm.DB, page, pageSize int) ([]models.AnalystRewardRecord, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rewards []models.AnalystRewardRecord
	if err := query.Preload("Analyst").Preload("Analyst.User").Preload("SettlementBatch").
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&rewards).Error; err != nil {
		return nil, 0, err
	}
	return rewards, total, nil
}

func (s *OfficialAnalysisTaskService) settleRewardsWithBatch(rewardIDs []uint, adminID uint, note string, now time.Time) (*OfficialRewardBatchSettleResult, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("db is nil")
	}
	if len(rewardIDs) == 0 {
		return nil, ErrAnalystRewardInvalid
	}

	var batchID uint
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var rewards []models.AnalystRewardRecord
		if err := tx.Where("id IN ? AND status = ?", rewardIDs, models.AnalystRewardPending).
			Order("id ASC").
			Find(&rewards).Error; err != nil {
			return err
		}
		if len(rewards) == 0 {
			return ErrAnalystRewardInvalid
		}

		rewardIDsToSettle := make([]uint, 0, len(rewards))
		totalAmount := 0.0
		for _, reward := range rewards {
			rewardIDsToSettle = append(rewardIDsToSettle, reward.ID)
			totalAmount += reward.Amount
		}

		batch := models.AnalystRewardSettlementBatch{
			BatchNo:     buildOfficialRewardBatchNo(now, adminID, rewardIDsToSettle[0]),
			RewardCount: len(rewards),
			TotalAmount: totalAmount,
			Status:      models.AnalystRewardSettled,
			SettledAt:   now,
			SettledBy:   adminID,
			Note:        note,
		}
		if err := tx.Create(&batch).Error; err != nil {
			return err
		}
		updated := tx.Model(&models.AnalystRewardRecord{}).
			Where("id IN ? AND status = ?", rewardIDsToSettle, models.AnalystRewardPending).
			Updates(map[string]interface{}{
				"status":              models.AnalystRewardSettled,
				"settlement_batch_id": batch.ID,
				"settled_at":          &now,
				"settled_by":          adminID,
				"note":                note,
				"updated_at":          now,
			})
		if updated.Error != nil {
			return updated.Error
		}
		if updated.RowsAffected != int64(len(rewardIDsToSettle)) {
			return ErrAnalystRewardInvalid
		}
		batchID = batch.ID
		return nil
	})
	if err != nil {
		return nil, err
	}

	batch, err := s.GetRewardSettlementBatch(batchID)
	if err != nil {
		return nil, err
	}
	return &OfficialRewardBatchSettleResult{
		Count: int64(batch.RewardCount),
		Batch: batch,
	}, nil
}

func buildOfficialRewardBatchNo(now time.Time, adminID, firstRewardID uint) string {
	return fmt.Sprintf("ORB%s%09dA%dR%d", now.Format("20060102150405"), now.Nanosecond(), adminID, firstRewardID)
}

func (s *OfficialAnalysisTaskService) updateRewardStatus(rewardID, adminID uint, status models.AnalystRewardStatus, note string, now time.Time, allowed ...models.AnalystRewardStatus) (*models.AnalystRewardRecord, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("db is nil")
	}
	var reward models.AnalystRewardRecord
	if err := s.db.First(&reward, rewardID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrAnalystRewardNotFound
		}
		return nil, err
	}
	allowedMap := make(map[models.AnalystRewardStatus]bool, len(allowed))
	for _, item := range allowed {
		allowedMap[item] = true
	}
	if !allowedMap[reward.Status] {
		return nil, ErrAnalystRewardInvalid
	}
	updates := map[string]interface{}{
		"status":     status,
		"note":       note,
		"updated_at": now,
	}
	if status == models.AnalystRewardSettled {
		updates["settled_at"] = &now
		updates["settled_by"] = adminID
	}
	if status == models.AnalystRewardReversed {
		updates["settled_by"] = firstNonZero(reward.SettledBy, adminID)
	}
	if err := s.db.Model(&reward).Updates(updates).Error; err != nil {
		return nil, err
	}
	if err := s.db.Preload("Analyst").Preload("Analyst.User").First(&reward, rewardID).Error; err != nil {
		return nil, err
	}
	return &reward, nil
}

func buildOfficialEventWork(adoption models.OfficialContentAdoption) OfficialEventWork {
	work := OfficialEventWork{
		ID:             adoption.ID,
		TaskID:         adoption.TaskID,
		SubmissionID:   adoption.SubmissionID,
		AnalystID:      adoption.AnalystID,
		Channel:        adoption.Channel,
		AdoptionStatus: string(adoption.AdoptionStatus),
		WorkTitle:      adoption.WorkTitle,
		WorkSummary:    adoption.WorkSummary,
		CoverURL:       adoption.CoverURL,
		RewardAmount:   adoption.RewardAmount,
		CreatedAt:      adoption.CreatedAt.Format("2006-01-02"),
	}
	if adoption.Task != nil {
		work.MatchName = adoption.Task.MatchName
		work.AgeGroup = adoption.Task.AgeGroup
	}
	if adoption.Analyst != nil {
		work.AnalystName = adoption.Analyst.Name
		work.AnalystLevel = adoption.Analyst.LevelCode
		work.IsPartner = adoption.Analyst.IsOfficialPartner
		work.AnalystUserID = adoption.Analyst.UserID
		if work.AnalystName == "" {
			work.AnalystName = adoption.Analyst.User.Nickname
		}
	}
	return work
}

func (s *OfficialAnalysisTaskService) loadEventTopicConfigs() ([]models.OfficialEventTopicConfig, error) {
	var configs []models.OfficialEventTopicConfig
	if err := s.db.Order("updated_at DESC").Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

func buildEventTopicConfigIndexes(configs []models.OfficialEventTopicConfig) (map[string]*models.OfficialEventTopicConfig, map[string]string) {
	configByMatch := make(map[string]*models.OfficialEventTopicConfig, len(configs))
	aliasToCanonical := make(map[string]string)
	for i := range configs {
		config := configs[i]
		canonical := strings.TrimSpace(config.MatchName)
		if canonical == "" {
			continue
		}
		configCopy := config
		configByMatch[canonical] = &configCopy
		aliasToCanonical[canonical] = canonical
		for _, alias := range config.AliasNames {
			normalized := strings.TrimSpace(alias)
			if normalized == "" {
				continue
			}
			aliasToCanonical[normalized] = canonical
		}
	}
	return configByMatch, aliasToCanonical
}

func resolveEventTopicMatchName(matchName string, aliasToCanonical map[string]string) string {
	matchName = strings.TrimSpace(matchName)
	if matchName == "" {
		return ""
	}
	if canonical, ok := aliasToCanonical[matchName]; ok && canonical != "" {
		return canonical
	}
	return matchName
}

func applyEventTopicConfig(topic *OfficialEventTopic, config *models.OfficialEventTopicConfig) {
	if topic == nil || config == nil {
		return
	}
	topic.DisplayName = firstNonEmptyTopic(strings.TrimSpace(config.DisplayName), topic.DisplayName, topic.MatchName)
	topic.Summary = strings.TrimSpace(config.Summary)
	topic.CoverURL = strings.TrimSpace(config.CoverURL)
	topic.AliasNames = sanitizeTopicAliases(config.AliasNames, topic.MatchName)
	topic.PinnedAdoptionID = config.PinnedAdoptionID
	topic.IsFeatured = config.IsFeatured
	topic.SortOrder = config.SortOrder
}

func applyEventTopicDetailConfig(detail *OfficialEventTopicDetail, config *models.OfficialEventTopicConfig) {
	if detail == nil || config == nil {
		return
	}
	detail.DisplayName = firstNonEmptyTopic(strings.TrimSpace(config.DisplayName), detail.DisplayName, detail.MatchName)
	detail.Summary = strings.TrimSpace(config.Summary)
	detail.CoverURL = strings.TrimSpace(config.CoverURL)
	detail.AliasNames = sanitizeTopicAliases(config.AliasNames, detail.MatchName)
	detail.PinnedAdoptionID = config.PinnedAdoptionID
	detail.IsFeatured = config.IsFeatured
	detail.SortOrder = config.SortOrder
}

func sortEventTopics(topics []OfficialEventTopic) {
	sort.SliceStable(topics, func(i, j int) bool {
		if topics[i].IsFeatured != topics[j].IsFeatured {
			return topics[i].IsFeatured
		}
		if topics[i].SortOrder != topics[j].SortOrder {
			if topics[i].SortOrder == 0 {
				return false
			}
			if topics[j].SortOrder == 0 {
				return true
			}
			return topics[i].SortOrder < topics[j].SortOrder
		}
		return topics[i].LatestPublishedAt > topics[j].LatestPublishedAt
	})
}

func topicMatchesKeyword(topic OfficialEventTopic, keyword string) bool {
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	if keyword == "" {
		return true
	}
	values := []string{
		topic.MatchName,
		topic.DisplayName,
		topic.Summary,
		topic.FeaturedWorkText(),
		strings.Join(topic.AliasNames, " "),
		strings.Join(topic.AgeGroups, " "),
	}
	return strings.Contains(strings.ToLower(strings.Join(values, " ")), keyword)
}

func (topic OfficialEventTopic) FeaturedWorkText() string {
	if topic.FeaturedWork == nil {
		return ""
	}
	return topic.FeaturedWork.WorkTitle + " " + topic.FeaturedWork.WorkSummary
}

func buildTopicConflictWarnings(matchName string, aliases []string, configs []models.OfficialEventTopicConfig) []string {
	warnings := []string{}
	for _, config := range configs {
		otherMatch := strings.TrimSpace(config.MatchName)
		if otherMatch == "" || otherMatch == matchName {
			continue
		}
		for _, alias := range aliases {
			alias = strings.TrimSpace(alias)
			if alias == "" {
				continue
			}
			if alias == otherMatch {
				warnings = append(warnings, fmt.Sprintf("别名「%s」也是另一个专题的标准赛事名", alias))
			}
			for _, otherAlias := range config.AliasNames {
				if alias == strings.TrimSpace(otherAlias) {
					warnings = append(warnings, fmt.Sprintf("别名「%s」已被专题「%s」使用", alias, otherMatch))
				}
			}
		}
	}
	return warnings
}

func (s *OfficialAnalysisTaskService) validateTopicAliases(matchName string, aliases []string) error {
	var configs []models.OfficialEventTopicConfig
	if err := s.db.Where("match_name <> ?", matchName).Find(&configs).Error; err != nil {
		return err
	}
	for _, config := range configs {
		otherMatch := strings.TrimSpace(config.MatchName)
		for _, alias := range aliases {
			alias = strings.TrimSpace(alias)
			if alias == "" {
				continue
			}
			for _, otherAlias := range config.AliasNames {
				if alias == strings.TrimSpace(otherAlias) {
					return fmt.Errorf("%w: 别名「%s」已被专题「%s」使用", ErrOfficialTaskInvalid, alias, otherMatch)
				}
			}
		}
	}
	return nil
}

func sanitizeTopicAliases(values []string, matchName string) []string {
	if len(values) == 0 {
		return []string{}
	}
	canonical := strings.TrimSpace(matchName)
	seen := make(map[string]bool, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		alias := strings.TrimSpace(value)
		if alias == "" || alias == canonical || seen[alias] {
			continue
		}
		seen[alias] = true
		result = append(result, alias)
	}
	return result
}

func firstNonEmptyTopic(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func (s *OfficialAnalysisTaskService) isPinnedAdoptionValid(matchName string, aliasNames []string, adoptionID uint) (bool, error) {
	var adoption models.OfficialContentAdoption
	if err := s.db.Preload("Task").Where("id = ? AND is_public = ?", adoptionID, true).First(&adoption).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	if adoption.Task == nil {
		return false, nil
	}
	targets := make(map[string]bool, len(aliasNames)+1)
	targets[strings.TrimSpace(matchName)] = true
	for _, alias := range aliasNames {
		targets[strings.TrimSpace(alias)] = true
	}
	return targets[strings.TrimSpace(adoption.Task.MatchName)], nil
}

func mapKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}

func paginationBounds(page, pageSize, total int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	start := (page - 1) * pageSize
	if start > total {
		return total, total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return start, end
}

func (s *OfficialAnalysisTaskService) FindActiveAnalystByUserID(userID uint) (*models.Analyst, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("db is nil")
	}
	var analyst models.Analyst
	if err := s.db.Where("user_id = ?", userID).First(&analyst).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrOfficialTaskAnalystInvalid
		}
		return nil, err
	}
	if analyst.Status != models.AnalystStatusActive {
		return nil, ErrOfficialTaskAnalystInvalid
	}
	return &analyst, nil
}

// AcceptTask 接取官方选题单，事务内控制名额，防止多人并发超额接单。
func (s *OfficialAnalysisTaskService) AcceptTask(analystID, taskID uint, now time.Time) (*models.OfficialAnalysisTaskAcceptance, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("db is nil")
	}

	var created models.OfficialAnalysisTaskAcceptance
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var analyst models.Analyst
		if err := tx.First(&analyst, analystID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrOfficialTaskAnalystInvalid
			}
			return err
		}
		if analyst.Status != models.AnalystStatusActive {
			return ErrOfficialTaskAnalystInvalid
		}
		if analyst.LevelCode == "" {
			analyst.LevelCode = models.DefaultAnalystLevelCode
		}

		var task models.OfficialAnalysisTask
		if err := tx.First(&task, taskID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrOfficialTaskNotFound
			}
			return err
		}
		if task.Status != models.OfficialAnalysisTaskPublished {
			if task.Status == models.OfficialAnalysisTaskFull {
				return ErrOfficialTaskFull
			}
			return ErrOfficialTaskUnavailable
		}
		if task.Deadline != nil && !task.Deadline.After(now) {
			return ErrOfficialTaskUnavailable
		}
		if task.MaxAcceptCount <= 0 {
			return ErrOfficialTaskUnavailable
		}
		if !officialTaskVisibleToLevel(&task, analyst.LevelCode) {
			return ErrOfficialTaskLevelDenied
		}
		if task.PriorityUntil != nil && task.PriorityUntil.After(now) && strings.TrimSpace(task.PriorityLevelMin) != "" &&
			!models.AnalystLevelMeets(analyst.LevelCode, task.PriorityLevelMin) {
			return ErrOfficialTaskLevelDenied
		}
		if err := s.ensureDailyOfficialTaskQuota(tx, analyst.ID, analyst.LevelCode, now); err != nil {
			return err
		}

		var existing int64
		if err := tx.Model(&models.OfficialAnalysisTaskAcceptance{}).
			Where("task_id = ? AND analyst_id = ? AND status IN ?", taskID, analystID, []models.OfficialAnalysisAcceptanceStatus{
				models.OfficialAnalysisAcceptanceAccepted,
				models.OfficialAnalysisAcceptanceSubmitted,
			}).
			Count(&existing).Error; err != nil {
			return err
		}
		if existing > 0 {
			return ErrOfficialTaskDuplicate
		}

		updated := tx.Model(&models.OfficialAnalysisTask{}).
			Where("id = ? AND status = ? AND current_accept_count < max_accept_count", taskID, models.OfficialAnalysisTaskPublished).
			Update("current_accept_count", gorm.Expr("current_accept_count + ?", 1))
		if updated.Error != nil {
			return updated.Error
		}
		if updated.RowsAffected == 0 {
			return ErrOfficialTaskFull
		}

		analysisOrderID, err := s.createOfficialAnalysisOrderForAcceptance(tx, &task, &analyst, now)
		if err != nil {
			return err
		}

		created = models.OfficialAnalysisTaskAcceptance{
			TaskID:          taskID,
			AnalystID:       analystID,
			AnalysisOrderID: analysisOrderID,
			AcceptedAt:      now,
			Status:          models.OfficialAnalysisAcceptanceAccepted,
		}
		if err := tx.Create(&created).Error; err != nil {
			return fmt.Errorf("创建官方选题接单记录失败: %w", err)
		}

		if err := tx.First(&task, taskID).Error; err != nil {
			return err
		}
		if task.CurrentAcceptCount >= task.MaxAcceptCount {
			if err := tx.Model(&models.OfficialAnalysisTask{}).
				Where("id = ? AND status = ?", taskID, models.OfficialAnalysisTaskPublished).
				Update("status", models.OfficialAnalysisTaskFull).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &created, nil
}

func (s *OfficialAnalysisTaskService) ensureDailyOfficialTaskQuota(tx *gorm.DB, analystID uint, levelCode string, now time.Time) error {
	limit := defaultOfficialDailyTaskLimit(levelCode)
	var level models.AnalystLevel
	if err := tx.Where("code = ?", levelCode).First(&level).Error; err == nil && level.DailyTaskLimit > 0 {
		limit = level.DailyTaskLimit
	}
	var acceptedToday int64
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dayEnd := dayStart.Add(24 * time.Hour)
	if err := tx.Table("official_analysis_task_acceptances AS a").
		Joins("JOIN official_analysis_tasks AS t ON t.id = a.task_id").
		Where("a.analyst_id = ? AND a.accepted_at >= ? AND a.accepted_at < ?", analystID, dayStart, dayEnd).
		Where("t.status NOT IN ?", []models.OfficialAnalysisTaskStatus{
			models.OfficialAnalysisTaskClosed,
			models.OfficialAnalysisTaskExpired,
		}).
		Count(&acceptedToday).Error; err != nil {
		return err
	}
	if acceptedToday >= int64(limit) {
		return ErrOfficialTaskDailyLimit
	}
	return nil
}

func (s *OfficialAnalysisTaskService) createOfficialAnalysisOrderForAcceptance(tx *gorm.DB, task *models.OfficialAnalysisTask, analyst *models.Analyst, now time.Time) (uint, error) {
	if task == nil || analyst == nil || task.TargetPlayerUserID == 0 {
		return 0, nil
	}

	var player models.User
	if err := tx.Where("id = ? AND role = ? AND status = ?", task.TargetPlayerUserID, models.RoleUser, models.StatusActive).First(&player).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, fmt.Errorf("%w: 关联球员账号不存在或不可用", ErrOfficialTaskInvalid)
		}
		return 0, err
	}

	videoURL := strings.TrimSpace(firstNonEmptyString(task.VideoFirstHalfURL, task.VideoURL, task.VideoSecondHalfURL))
	analystID := analyst.ID
	playerName := strings.TrimSpace(firstNonEmptyString(task.TargetPlayerName, player.Name, player.Nickname))
	playerTeam := strings.TrimSpace(firstNonEmptyString(task.TargetPlayerTeam, player.CurrentTeam, player.Club))
	position := strings.TrimSpace(firstNonEmptyString(task.TargetPlayerPosition, player.Position))
	jerseyColor := strings.TrimSpace(firstNonEmptyString(task.TargetJerseyColor, player.JerseyColor))
	jerseyNumber := strings.TrimSpace(task.TargetJerseyNumber)
	if jerseyNumber == "" && player.JerseyNumber > 0 {
		jerseyNumber = fmt.Sprintf("%d", player.JerseyNumber)
	}

	order := models.Order{
		UserID:             player.ID,
		AnalystID:          &analystID,
		OrderNo:            generateOfficialOrderNo(task.ID, analyst.ID, now),
		Amount:             0,
		Status:             models.OrderStatusProcessing,
		PaymentMethod:      models.PaymentMethodBalance,
		PaymentTime:        &now,
		PaidAt:             &now,
		VideoURL:           videoURL,
		VideoSecondHalfURL: strings.TrimSpace(task.VideoSecondHalfURL),
		Remark:             fmt.Sprintf("官方选题 #%d：%s", task.ID, task.Title),
		OrderType:          "official_analysis",
		PlayerName:         playerName,
		PlayerAge:          player.Age,
		PlayerPosition:     position,
		JerseyColor:        jerseyColor,
		JerseyNumber:       jerseyNumber,
		MatchName:          strings.TrimSpace(task.MatchName),
		MatchDate:          strings.TrimSpace(task.MatchDate),
		Deadline:           task.Deadline,
		AssignedAt:         &now,
		AcceptedAt:         &now,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if playerTeam != "" {
		if strings.TrimSpace(order.Remark) != "" {
			order.Remark += "；"
		}
		order.Remark += "目标球队：" + playerTeam
	}
	if err := tx.Create(&order).Error; err != nil {
		return 0, err
	}

	assignment := models.OrderAssignment{
		OrderID:     order.ID,
		AnalystID:   analyst.ID,
		AssignedAt:  now,
		Status:      models.OrderAssignmentStatusAccepted,
		RespondedAt: &now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if task.CreatedBy != 0 {
		assignedBy := task.CreatedBy
		assignment.AssignedBy = &assignedBy
	}
	if err := tx.Create(&assignment).Error; err != nil {
		return 0, err
	}

	return order.ID, nil
}

func defaultOfficialDailyTaskLimit(levelCode string) int {
	for _, level := range models.DefaultAnalystLevels() {
		if level.Code == levelCode && level.DailyTaskLimit > 0 {
			return level.DailyTaskLimit
		}
	}
	return 1
}

func (s *OfficialAnalysisTaskService) findActiveAnalyst(analystID uint) (*models.Analyst, error) {
	return s.findActiveAnalystWithTx(s.db, analystID)
}

func (s *OfficialAnalysisTaskService) findActiveAnalystWithTx(tx *gorm.DB, analystID uint) (*models.Analyst, error) {
	var analyst models.Analyst
	if err := tx.First(&analyst, analystID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrOfficialTaskAnalystInvalid
		}
		return nil, err
	}
	if analyst.Status != models.AnalystStatusActive {
		return nil, ErrOfficialTaskAnalystInvalid
	}
	return &analyst, nil
}

func buildOfficialTask(req OfficialAnalysisTaskRequest, adminID uint, now time.Time) (*models.OfficialAnalysisTask, error) {
	updates, err := buildOfficialTaskUpdates(req)
	if err != nil {
		return nil, err
	}
	task := &models.OfficialAnalysisTask{
		CreatedBy: adminID,
		Status:    models.OfficialAnalysisTaskDraft,
		CreatedAt: now,
		UpdatedAt: now,
	}
	applyOfficialTaskUpdates(task, updates)
	return task, nil
}

func buildOfficialTaskUpdates(req OfficialAnalysisTaskRequest) (map[string]interface{}, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, fmt.Errorf("%w: 标题不能为空", ErrOfficialTaskInvalid)
	}
	if req.MaxAcceptCount <= 0 {
		return nil, fmt.Errorf("%w: 最多接单人数必须大于0", ErrOfficialTaskInvalid)
	}
	if req.BaseRewardAmount < 0 {
		return nil, fmt.Errorf("%w: 基础报酬不能小于0", ErrOfficialTaskInvalid)
	}

	visibleLevel := strings.TrimSpace(req.VisibleLevelMin)
	if visibleLevel == "" {
		visibleLevel = models.DefaultAnalystLevelCode
	}
	if models.AnalystLevelRank(visibleLevel) == 0 {
		return nil, fmt.Errorf("%w: 可见等级无效", ErrOfficialTaskInvalid)
	}
	visibleLevelCodes, err := normalizeOfficialVisibleLevelCodes(req.VisibleLevelCodes)
	if err != nil {
		return nil, err
	}
	if len(visibleLevelCodes) > 0 {
		visibleLevel = visibleLevelCodes[0]
	}
	priorityLevel := strings.TrimSpace(req.PriorityLevelMin)
	if priorityLevel != "" && models.AnalystLevelRank(priorityLevel) == 0 {
		return nil, fmt.Errorf("%w: 优先等级无效", ErrOfficialTaskInvalid)
	}

	taskType := strings.TrimSpace(req.TaskType)
	if taskType == "" {
		taskType = "composite"
	}
	authStatus := strings.TrimSpace(req.AuthorizationStatus)
	if authStatus == "" {
		authStatus = "pending"
	}

	return map[string]interface{}{
		"title":                  title,
		"match_name":             strings.TrimSpace(req.MatchName),
		"age_group":              strings.TrimSpace(req.AgeGroup),
		"match_date":             strings.TrimSpace(req.MatchDate),
		"video_url":              strings.TrimSpace(req.VideoURL),
		"video_first_half_url":   strings.TrimSpace(firstNonEmptyString(req.VideoFirstHalfURL, req.VideoURL)),
		"video_second_half_url":  strings.TrimSpace(req.VideoSecondHalfURL),
		"video_source":           strings.TrimSpace(req.VideoSource),
		"authorization_status":   authStatus,
		"target_player_user_id":  req.TargetPlayerUserID,
		"target_player_name":     strings.TrimSpace(req.TargetPlayerName),
		"target_player_team":     strings.TrimSpace(req.TargetPlayerTeam),
		"target_jersey_color":    strings.TrimSpace(req.TargetJerseyColor),
		"target_jersey_number":   strings.TrimSpace(req.TargetJerseyNumber),
		"target_player_position": strings.TrimSpace(req.TargetPlayerPosition),
		"task_type":              taskType,
		"base_reward_amount":     req.BaseRewardAmount,
		"adoption_reward_rule":   strings.TrimSpace(req.AdoptionRewardRule),
		"bonus_rule":             strings.TrimSpace(req.BonusRule),
		"requirements":           strings.TrimSpace(req.Requirements),
		"max_accept_count":       req.MaxAcceptCount,
		"visible_level_min":      visibleLevel,
		"visible_level_codes":    strings.Join(visibleLevelCodes, ","),
		"priority_level_min":     priorityLevel,
		"priority_until":         req.PriorityUntil,
		"deadline":               req.Deadline,
	}, nil
}

func applyOfficialTaskUpdates(task *models.OfficialAnalysisTask, updates map[string]interface{}) {
	task.Title = updates["title"].(string)
	task.MatchName = updates["match_name"].(string)
	task.AgeGroup = updates["age_group"].(string)
	task.MatchDate = updates["match_date"].(string)
	task.VideoURL = updates["video_url"].(string)
	task.VideoFirstHalfURL = updates["video_first_half_url"].(string)
	task.VideoSecondHalfURL = updates["video_second_half_url"].(string)
	task.VideoSource = updates["video_source"].(string)
	task.AuthorizationStatus = updates["authorization_status"].(string)
	task.TargetPlayerUserID = updates["target_player_user_id"].(uint)
	task.TargetPlayerName = updates["target_player_name"].(string)
	task.TargetPlayerTeam = updates["target_player_team"].(string)
	task.TargetJerseyColor = updates["target_jersey_color"].(string)
	task.TargetJerseyNumber = updates["target_jersey_number"].(string)
	task.TargetPlayerPosition = updates["target_player_position"].(string)
	task.TaskType = updates["task_type"].(string)
	task.BaseRewardAmount = updates["base_reward_amount"].(float64)
	task.AdoptionRewardRule = updates["adoption_reward_rule"].(string)
	task.BonusRule = updates["bonus_rule"].(string)
	task.Requirements = updates["requirements"].(string)
	task.MaxAcceptCount = updates["max_accept_count"].(int)
	task.VisibleLevelMin = updates["visible_level_min"].(string)
	task.VisibleLevelCodes = updates["visible_level_codes"].(string)
	task.PriorityLevelMin = updates["priority_level_min"].(string)
	task.PriorityUntil = updates["priority_until"].(*time.Time)
	task.Deadline = updates["deadline"].(*time.Time)
}

func mergeOfficialBatchTask(common, item OfficialAnalysisTaskRequest, eventName string, index int) OfficialAnalysisTaskRequest {
	eventName = strings.TrimSpace(eventName)
	merged := item
	if strings.TrimSpace(merged.Title) == "" {
		parts := []string{}
		if eventName != "" {
			parts = append(parts, eventName)
		}
		if strings.TrimSpace(item.MatchName) != "" && strings.TrimSpace(item.MatchName) != eventName {
			parts = append(parts, strings.TrimSpace(item.MatchName))
		}
		if strings.TrimSpace(item.TargetPlayerName) != "" {
			parts = append(parts, strings.TrimSpace(item.TargetPlayerName))
		}
		if len(parts) == 0 {
			parts = append(parts, fmt.Sprintf("第%d场", index))
		}
		merged.Title = strings.Join(parts, "｜")
	}
	if strings.TrimSpace(merged.MatchName) == "" {
		merged.MatchName = firstNonEmptyString(item.MatchName, common.MatchName, eventName)
	}
	if strings.TrimSpace(merged.AgeGroup) == "" {
		merged.AgeGroup = common.AgeGroup
	}
	if strings.TrimSpace(merged.MatchDate) == "" {
		merged.MatchDate = common.MatchDate
	}
	if strings.TrimSpace(merged.VideoSource) == "" {
		merged.VideoSource = common.VideoSource
	}
	if strings.TrimSpace(merged.AuthorizationStatus) == "" {
		merged.AuthorizationStatus = common.AuthorizationStatus
	}
	if strings.TrimSpace(merged.TaskType) == "" {
		merged.TaskType = common.TaskType
	}
	if merged.BaseRewardAmount == 0 && common.BaseRewardAmount > 0 {
		merged.BaseRewardAmount = common.BaseRewardAmount
	}
	if strings.TrimSpace(merged.AdoptionRewardRule) == "" {
		merged.AdoptionRewardRule = common.AdoptionRewardRule
	}
	if strings.TrimSpace(merged.BonusRule) == "" {
		merged.BonusRule = common.BonusRule
	}
	if strings.TrimSpace(merged.Requirements) == "" {
		merged.Requirements = common.Requirements
	}
	if merged.MaxAcceptCount <= 0 {
		merged.MaxAcceptCount = common.MaxAcceptCount
	}
	if strings.TrimSpace(merged.VisibleLevelMin) == "" {
		merged.VisibleLevelMin = common.VisibleLevelMin
	}
	if len(merged.VisibleLevelCodes) == 0 {
		merged.VisibleLevelCodes = common.VisibleLevelCodes
	}
	if strings.TrimSpace(merged.PriorityLevelMin) == "" {
		merged.PriorityLevelMin = common.PriorityLevelMin
	}
	if merged.PriorityUntil == nil {
		merged.PriorityUntil = common.PriorityUntil
	}
	if merged.Deadline == nil {
		merged.Deadline = common.Deadline
	}
	return merged
}

func normalizeOfficialVisibleLevelCodes(values []string) ([]string, error) {
	if len(values) == 0 {
		return nil, nil
	}
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, raw := range values {
		code := strings.TrimSpace(raw)
		if code == "" {
			continue
		}
		if models.AnalystLevelRank(code) == 0 {
			return nil, fmt.Errorf("%w: 可见等级无效", ErrOfficialTaskInvalid)
		}
		if !seen[code] {
			seen[code] = true
			result = append(result, code)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return models.AnalystLevelRank(result[i]) < models.AnalystLevelRank(result[j])
	})
	return result, nil
}

func officialTaskVisibleToLevel(task *models.OfficialAnalysisTask, levelCode string) bool {
	if task == nil {
		return false
	}
	levelCode = strings.TrimSpace(levelCode)
	if levelCode == "" {
		levelCode = models.DefaultAnalystLevelCode
	}
	visibleCodes := strings.Split(strings.TrimSpace(task.VisibleLevelCodes), ",")
	for _, raw := range visibleCodes {
		if strings.TrimSpace(raw) == levelCode {
			return true
		}
	}
	if strings.TrimSpace(task.VisibleLevelCodes) != "" {
		return false
	}
	return models.AnalystLevelMeets(levelCode, task.VisibleLevelMin)
}

func validateOfficialTaskForPublish(task *models.OfficialAnalysisTask) error {
	if strings.TrimSpace(task.Title) == "" {
		return fmt.Errorf("%w: 标题不能为空", ErrOfficialTaskInvalid)
	}
	if task.MaxAcceptCount <= 0 {
		return fmt.Errorf("%w: 最多接单人数必须大于0", ErrOfficialTaskInvalid)
	}
	authStatus := strings.TrimSpace(task.AuthorizationStatus)
	if authStatus == "" || authStatus == "unknown" || authStatus == "pending" {
		return fmt.Errorf("%w: 视频授权状态必须填写后才能发布", ErrOfficialTaskInvalid)
	}
	if strings.TrimSpace(firstNonEmptyString(task.VideoFirstHalfURL, task.VideoURL, task.VideoSecondHalfURL)) == "" {
		return fmt.Errorf("%w: 至少需要上传一段比赛视频后才能发布", ErrOfficialTaskInvalid)
	}
	if task.TargetPlayerUserID == 0 {
		return fmt.Errorf("%w: 必须关联平台球员账号后才能发布", ErrOfficialTaskInvalid)
	}
	if strings.TrimSpace(task.TargetPlayerName) == "" {
		return fmt.Errorf("%w: 目标球员姓名不能为空", ErrOfficialTaskInvalid)
	}
	if strings.TrimSpace(task.TargetPlayerTeam) == "" {
		return fmt.Errorf("%w: 目标球员球队不能为空", ErrOfficialTaskInvalid)
	}
	if strings.TrimSpace(task.TargetJerseyColor) == "" {
		return fmt.Errorf("%w: 目标球员队服颜色不能为空", ErrOfficialTaskInvalid)
	}
	if strings.TrimSpace(task.TargetJerseyNumber) == "" {
		return fmt.Errorf("%w: 目标球员号码不能为空", ErrOfficialTaskInvalid)
	}
	if strings.TrimSpace(task.TargetPlayerPosition) == "" {
		return fmt.Errorf("%w: 目标球员位置不能为空", ErrOfficialTaskInvalid)
	}
	if strings.TrimSpace(task.VisibleLevelCodes) != "" {
		for _, code := range strings.Split(task.VisibleLevelCodes, ",") {
			if models.AnalystLevelRank(strings.TrimSpace(code)) == 0 {
				return fmt.Errorf("%w: 可见等级无效", ErrOfficialTaskInvalid)
			}
		}
		return nil
	}
	if models.AnalystLevelRank(task.VisibleLevelMin) == 0 {
		return fmt.Errorf("%w: 可见等级无效", ErrOfficialTaskInvalid)
	}
	return nil
}

func validateOfficialSubmissionRequest(req OfficialTaskSubmitRequest) error {
	authStatus := strings.TrimSpace(req.VideoAuthorizationStatus)
	if authStatus == "" || authStatus == "unknown" {
		return fmt.Errorf("%w: 视频授权状态必须填写后才能提交", ErrOfficialSubmissionInvalid)
	}
	if strings.TrimSpace(req.VideoFileURL) == "" &&
		strings.TrimSpace(req.ScriptText) == "" &&
		strings.TrimSpace(req.Summary) == "" &&
		req.ReportID == 0 &&
		req.AnalysisID == 0 {
		return fmt.Errorf("%w: 至少需要填写脚本、摘要、视频文件或关联报告", ErrOfficialSubmissionInvalid)
	}
	return nil
}

func validateOfficialAdoptionRequest(req OfficialSubmissionAdoptionRequest) error {
	if req.AdoptionStatus != models.OfficialContentAdoptionMaterial &&
		req.AdoptionStatus != models.OfficialContentAdoptionOfficialPublished &&
		req.AdoptionStatus != models.OfficialContentAdoptionKeySpread &&
		req.AdoptionStatus != models.OfficialContentAdoptionLongTerm {
		return fmt.Errorf("%w: 采用状态无效", ErrOfficialSubmissionInvalid)
	}
	if req.RewardAmount < 0 {
		return fmt.Errorf("%w: 奖励金额不能小于0", ErrOfficialSubmissionInvalid)
	}
	if strings.TrimSpace(req.WorkTitle) == "" {
		return fmt.Errorf("%w: 作品标题不能为空", ErrOfficialSubmissionInvalid)
	}
	return nil
}

func generateOfficialTaskNo(adminID uint, now time.Time) string {
	return fmt.Sprintf("OAT%s%06d%03d", now.Format("20060102150405"), now.Nanosecond()/1000, adminID%1000)
}

func generateOfficialOrderNo(taskID, analystID uint, now time.Time) string {
	return fmt.Sprintf("OA%s%06d%04d%04d", now.Format("20060102150405"), now.Nanosecond()/1000, taskID%10000, analystID%10000)
}

func firstNonZero(current, fallback uint) uint {
	if current != 0 {
		return current
	}
	return fallback
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
