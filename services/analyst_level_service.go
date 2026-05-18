package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shaonianqiutan/backend/models"
	"gorm.io/gorm"
)

var (
	ErrAnalystLevelInvalid             = errors.New("分析师等级参数不合法")
	ErrAnalystLevelAnalystInvalid      = errors.New("分析师不存在或不可用")
	ErrAnalystLevelApplicationNotFound = errors.New("分析师等级申请不存在")
	ErrAnalystLevelApplicationConflict = errors.New("已有待审核的等级申请")
)

type AnalystLevelService struct {
	db *gorm.DB
}

func NewAnalystLevelService(db *gorm.DB) *AnalystLevelService {
	return &AnalystLevelService{db: db}
}

type AnalystLevelApplicationRequest struct {
	RequestedLevelCode string `json:"requested_level_code"`
	ApplicationReason  string `json:"application_reason"`
	ExperienceSummary  string `json:"experience_summary"`
	CaseMaterials      string `json:"case_materials"`
	Specialties        string `json:"specialties"`
	SelfAssessment     string `json:"self_assessment"`
}

type AnalystLevelReviewRequest struct {
	Status            models.AnalystLevelApplicationStatus `json:"status"`
	ReviewedLevelCode string                               `json:"reviewed_level_code"`
	ReviewNote        string                               `json:"review_note"`
}

type AnalystLevelSetRequest struct {
	LevelCode string `json:"level_code"`
	Note      string `json:"note"`
}

type AnalystOfficialPartnershipRequest struct {
	IsOfficialPartner   bool   `json:"is_official_partner"`
	PartnershipNote     string `json:"partnership_note"`
	PartnershipBenefits string `json:"partnership_benefits"`
}

type AnalystLevelSuggestionActionRequest struct {
	Note string `json:"note"`
}

func (s *AnalystLevelService) ListLevels() ([]models.AnalystLevel, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("db is nil")
	}
	var levels []models.AnalystLevel
	if err := s.db.Order("priority_weight ASC, code ASC").Find(&levels).Error; err != nil {
		return nil, err
	}
	return levels, nil
}

func (s *AnalystLevelService) GetAnalystLevelProfileByUserID(userID uint) (*models.Analyst, *models.AnalystLevelApplication, *models.AnalystGrowthSnapshot, []models.AnalystLevelHistory, error) {
	if s == nil || s.db == nil {
		return nil, nil, nil, nil, errors.New("db is nil")
	}
	analyst, err := s.findActiveAnalystByUserID(userID)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	growth, err := s.RefreshGrowthSnapshot(analyst.ID, time.Now())
	if err != nil {
		return nil, nil, nil, nil, err
	}

	var app models.AnalystLevelApplication
	err = s.db.Where("analyst_id = ?", analyst.ID).Order("created_at DESC").First(&app).Error
	if err == gorm.ErrRecordNotFound {
		histories, historyErr := s.ListLevelHistories(analyst.ID, 1, 5)
		return analyst, nil, growth, histories, historyErr
	}
	if err != nil {
		return nil, nil, nil, nil, err
	}
	histories, err := s.ListLevelHistories(analyst.ID, 1, 5)
	return analyst, &app, growth, histories, err
}

func (s *AnalystLevelService) SubmitApplication(userID uint, req AnalystLevelApplicationRequest, now time.Time) (*models.AnalystLevelApplication, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("db is nil")
	}
	levelCode, err := normalizeLevelCode(req.RequestedLevelCode)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.ApplicationReason) == "" {
		return nil, fmt.Errorf("%w: 申请理由不能为空", ErrAnalystLevelInvalid)
	}

	analyst, err := s.findActiveAnalystByUserID(userID)
	if err != nil {
		return nil, err
	}

	var pendingCount int64
	if err := s.db.Model(&models.AnalystLevelApplication{}).
		Where("analyst_id = ? AND status = ?", analyst.ID, models.AnalystLevelApplicationPending).
		Count(&pendingCount).Error; err != nil {
		return nil, err
	}
	if pendingCount > 0 {
		return nil, ErrAnalystLevelApplicationConflict
	}

	app := models.AnalystLevelApplication{
		AnalystID:          analyst.ID,
		RequestedLevelCode: levelCode,
		ApplicationReason:  strings.TrimSpace(req.ApplicationReason),
		ExperienceSummary:  strings.TrimSpace(req.ExperienceSummary),
		CaseMaterials:      strings.TrimSpace(req.CaseMaterials),
		Specialties:        strings.TrimSpace(req.Specialties),
		SelfAssessment:     strings.TrimSpace(req.SelfAssessment),
		Status:             models.AnalystLevelApplicationPending,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := s.db.Create(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

func (s *AnalystLevelService) ListApplications(page, pageSize int, status string) ([]models.AnalystLevelApplication, int64, error) {
	if s == nil || s.db == nil {
		return nil, 0, errors.New("db is nil")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	query := s.db.Model(&models.AnalystLevelApplication{})
	if strings.TrimSpace(status) != "" {
		query = query.Where("status = ?", strings.TrimSpace(status))
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var apps []models.AnalystLevelApplication
	if err := query.Preload("Analyst").Preload("Analyst.User").
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&apps).Error; err != nil {
		return nil, 0, err
	}
	return apps, total, nil
}

func (s *AnalystLevelService) GetApplication(appID uint) (*models.AnalystLevelApplication, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("db is nil")
	}
	var app models.AnalystLevelApplication
	if err := s.db.Preload("Analyst").Preload("Analyst.User").First(&app, appID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrAnalystLevelApplicationNotFound
		}
		return nil, err
	}
	return &app, nil
}

func (s *AnalystLevelService) ListMyApplications(userID uint, page, pageSize int) ([]models.AnalystLevelApplication, int64, error) {
	if s == nil || s.db == nil {
		return nil, 0, errors.New("db is nil")
	}
	analyst, err := s.findActiveAnalystByUserID(userID)
	if err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	query := s.db.Model(&models.AnalystLevelApplication{}).Where("analyst_id = ?", analyst.ID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var apps []models.AnalystLevelApplication
	if err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&apps).Error; err != nil {
		return nil, 0, err
	}
	return apps, total, nil
}

func (s *AnalystLevelService) ReviewApplication(appID, adminID uint, req AnalystLevelReviewRequest, now time.Time) (*models.AnalystLevelApplication, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("db is nil")
	}
	if req.Status != models.AnalystLevelApplicationApproved &&
		req.Status != models.AnalystLevelApplicationAdjusted &&
		req.Status != models.AnalystLevelApplicationRejected {
		return nil, ErrAnalystLevelInvalid
	}

	var reviewed models.AnalystLevelApplication
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var app models.AnalystLevelApplication
		if err := tx.First(&app, appID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrAnalystLevelApplicationNotFound
			}
			return err
		}
		if app.Status != models.AnalystLevelApplicationPending {
			return ErrAnalystLevelInvalid
		}

		levelCode := app.RequestedLevelCode
		if req.Status == models.AnalystLevelApplicationAdjusted {
			var err error
			levelCode, err = normalizeLevelCode(req.ReviewedLevelCode)
			if err != nil {
				return err
			}
		}
		if req.Status == models.AnalystLevelApplicationApproved {
			var err error
			levelCode, err = normalizeLevelCode(levelCode)
			if err != nil {
				return err
			}
		}
		if req.Status == models.AnalystLevelApplicationRejected {
			levelCode = ""
		}

		if err := tx.Model(&app).Updates(map[string]interface{}{
			"status":              req.Status,
			"reviewed_level_code": levelCode,
			"review_note":         strings.TrimSpace(req.ReviewNote),
			"reviewed_by":         adminID,
			"reviewed_at":         &now,
			"updated_at":          now,
		}).Error; err != nil {
			return err
		}

		if req.Status == models.AnalystLevelApplicationApproved || req.Status == models.AnalystLevelApplicationAdjusted {
			if err := s.setAnalystLevelWithTx(tx, app.AnalystID, levelCode, adminID, strings.TrimSpace(req.ReviewNote), "application", "", "applied", now); err != nil {
				return err
			}
		}

		return tx.First(&reviewed, appID).Error
	})
	if err != nil {
		return nil, err
	}
	return &reviewed, nil
}

func (s *AnalystLevelService) SetAnalystLevel(analystID, adminID uint, req AnalystLevelSetRequest, now time.Time) (*models.Analyst, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("db is nil")
	}
	levelCode, err := normalizeLevelCode(req.LevelCode)
	if err != nil {
		return nil, err
	}
	if err := s.setAnalystLevelWithTx(s.db, analystID, levelCode, adminID, strings.TrimSpace(req.Note), "manual", "", "applied", now); err != nil {
		return nil, err
	}
	var analyst models.Analyst
	if err := s.db.Preload("User").First(&analyst, analystID).Error; err != nil {
		return nil, err
	}
	return &analyst, nil
}

func (s *AnalystLevelService) ListAnalysts(page, pageSize int) ([]models.Analyst, int64, error) {
	if s == nil || s.db == nil {
		return nil, 0, errors.New("db is nil")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	query := s.db.Model(&models.Analyst{})
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var analysts []models.Analyst
	if err := query.Preload("User").Preload("GrowthSnapshot").
		Order("is_official_partner DESC, level_code DESC, official_adoption_count DESC, created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&analysts).Error; err != nil {
		return nil, 0, err
	}
	if err := s.attachOfficialPublishMetrics(analysts); err != nil {
		return nil, 0, err
	}
	return analysts, total, nil
}

func (s *AnalystLevelService) SetOfficialPartnership(analystID, adminID uint, req AnalystOfficialPartnershipRequest, now time.Time) (*models.Analyst, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("db is nil")
	}
	var analyst models.Analyst
	if err := s.db.First(&analyst, analystID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrAnalystLevelAnalystInvalid
		}
		return nil, err
	}
	if analyst.Status != models.AnalystStatusActive {
		return nil, ErrAnalystLevelAnalystInvalid
	}

	updates := map[string]interface{}{
		"is_official_partner":    req.IsOfficialPartner,
		"partnership_updated_by": adminID,
		"partnership_note":       strings.TrimSpace(req.PartnershipNote),
		"partnership_benefits":   strings.TrimSpace(req.PartnershipBenefits),
		"updated_at":             now,
	}
	if req.IsOfficialPartner {
		if analyst.PartnershipStartedAt == nil {
			updates["partnership_started_at"] = &now
		}
	} else {
		updates["partnership_started_at"] = gorm.Expr("NULL")
	}
	if err := s.db.Model(&models.Analyst{}).
		Where("id = ?", analystID).
		Select("is_official_partner", "partnership_started_at", "partnership_updated_by", "partnership_note", "partnership_benefits", "updated_at").
		Updates(updates).Error; err != nil {
		return nil, err
	}
	if !req.IsOfficialPartner {
		if err := s.db.Model(&models.Analyst{}).Where("id = ?", analystID).UpdateColumn("partnership_started_at", gorm.Expr("NULL")).Error; err != nil {
			return nil, err
		}
	}
	analyst = models.Analyst{}
	if err := s.db.Preload("User").Preload("GrowthSnapshot").First(&analyst, analystID).Error; err != nil {
		return nil, err
	}
	return &analyst, nil
}

func (s *AnalystLevelService) RefreshGrowthSnapshot(analystID uint, now time.Time) (*models.AnalystGrowthSnapshot, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("db is nil")
	}
	var analyst models.Analyst
	if err := s.db.First(&analyst, analystID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrAnalystLevelAnalystInvalid
		}
		return nil, err
	}

	var totalSubmissions int64
	if err := s.db.Model(&models.OfficialAnalysisSubmission{}).Where("analyst_id = ?", analystID).Count(&totalSubmissions).Error; err != nil {
		return nil, err
	}
	var approvedSubmissions int64
	if err := s.db.Model(&models.OfficialAnalysisSubmission{}).
		Where("analyst_id = ? AND status IN ?", analystID, []models.OfficialAnalysisSubmissionStatus{
			models.OfficialAnalysisSubmissionApproved,
			models.OfficialAnalysisSubmissionAdopted,
		}).Count(&approvedSubmissions).Error; err != nil {
		return nil, err
	}
	var revisionSubmissions int64
	if err := s.db.Model(&models.OfficialAnalysisSubmission{}).
		Where("analyst_id = ? AND status IN ?", analystID, []models.OfficialAnalysisSubmissionStatus{
			models.OfficialAnalysisSubmissionRevisionRequired,
			models.OfficialAnalysisSubmissionRejected,
		}).Count(&revisionSubmissions).Error; err != nil {
		return nil, err
	}
	var onTimeSubmissions int64
	if err := s.db.Table("official_analysis_submissions AS s").
		Joins("JOIN official_analysis_tasks AS t ON t.id = s.task_id").
		Where("s.analyst_id = ? AND (t.deadline IS NULL OR s.created_at <= t.deadline)", analystID).
		Count(&onTimeSubmissions).Error; err != nil {
		return nil, err
	}
	var completedOrders int64
	if err := s.db.Model(&models.Order{}).
		Where("analyst_id = ? AND status = ?", analystID, models.OrderStatusCompleted).
		Count(&completedOrders).Error; err != nil {
		return nil, err
	}

	qualityScore := analyst.QualityScore
	if totalSubmissions > 0 {
		approvedRate := float64(approvedSubmissions) / float64(totalSubmissions)
		revisionPenalty := float64(revisionSubmissions) / float64(totalSubmissions) * 20
		qualityScore = clampScore(approvedRate*85 + minFloat(analyst.Rating*3, 15) - revisionPenalty)
	}
	deliveryScore := analyst.DeliveryScore
	if totalSubmissions > 0 || completedOrders > 0 {
		onTimeRate := 1.0
		if totalSubmissions > 0 {
			onTimeRate = float64(onTimeSubmissions) / float64(totalSubmissions)
		}
		deliveryScore = clampScore(onTimeRate*80 + minFloat(float64(totalSubmissions+completedOrders)*4, 20))
	}
	contentScore := clampScore(float64(analyst.OfficialMaterialCount)*10 + float64(analyst.OfficialPublishCount)*25 + float64(analyst.OfficialAdoptionCount)*8)
	businessScore := clampScore(float64(completedOrders)*12 + minFloat(analyst.Rating*4, 20) + minFloat(float64(analyst.ReviewCount)*2, 20))
	growthScore := clampScore(qualityScore*0.35 + deliveryScore*0.25 + contentScore*0.25 + businessScore*0.15)
	suggestedLevel, nextLevel, gap := suggestAnalystLevel(growthScore)
	reason := fmt.Sprintf("成长分 %.1f；质量 %.1f、履约 %.1f、内容 %.1f、商业 %.1f", growthScore, qualityScore, deliveryScore, contentScore, businessScore)
	publishMetrics, err := s.getOfficialPublishMetrics(analystID)
	if err != nil {
		return nil, err
	}
	if publishMetrics.PublishRecordCount > 0 {
		reason = fmt.Sprintf("%s；官方发布 %d 条，累计播放 %d，最高播放 %d", reason, publishMetrics.PublishRecordCount, publishMetrics.TotalPlayCount, publishMetrics.MaxPlayCount)
	}

	snapshot := models.AnalystGrowthSnapshot{
		AnalystID:             analystID,
		QualityScore:          qualityScore,
		DeliveryScore:         deliveryScore,
		ContentScore:          contentScore,
		BusinessScore:         businessScore,
		GrowthScore:           growthScore,
		SuggestedLevelCode:    suggestedLevel,
		SuggestionReason:      reason,
		NextLevelCode:         nextLevel,
		NextLevelGap:          gap,
		OfficialSubmissionNum: int(totalSubmissions),
		OfficialApprovedNum:   int(approvedSubmissions),
		OfficialAdoptionNum:   analyst.OfficialAdoptionCount,
		PaidCompletedNum:      int(completedOrders),
		SuggestionStatus:      "pending",
		CalculatedAt:          now,
		UpdatedAt:             now,
	}

	var existing models.AnalystGrowthSnapshot
	err = s.db.Where("analyst_id = ?", analystID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		snapshot.CreatedAt = now
		if err := s.db.Create(&snapshot).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else {
		if existing.SuggestedLevelCode == suggestedLevel && (existing.SuggestionStatus == "ignored" || existing.SuggestionStatus == "applied") {
			snapshot.SuggestionStatus = existing.SuggestionStatus
			snapshot.SuggestionReviewedBy = existing.SuggestionReviewedBy
			snapshot.SuggestionReviewedAt = existing.SuggestionReviewedAt
			snapshot.SuggestionReviewNote = existing.SuggestionReviewNote
		}
		if err := s.db.Model(&existing).Updates(snapshot).Error; err != nil {
			return nil, err
		}
		snapshot.ID = existing.ID
		snapshot.CreatedAt = existing.CreatedAt
	}

	if err := s.db.Model(&models.Analyst{}).Where("id = ?", analystID).Updates(map[string]interface{}{
		"quality_score":  qualityScore,
		"delivery_score": deliveryScore,
		"content_score":  contentScore,
		"business_score": businessScore,
		"growth_score":   growthScore,
		"updated_at":     now,
	}).Error; err != nil {
		return nil, err
	}

	return &snapshot, nil
}

func (s *AnalystLevelService) attachOfficialPublishMetrics(analysts []models.Analyst) error {
	if len(analysts) == 0 {
		return nil
	}
	ids := make([]uint, 0, len(analysts))
	for _, analyst := range analysts {
		ids = append(ids, analyst.ID)
	}
	var rows []officialPublishMetricsRow
	if err := s.db.Table("official_content_adoptions AS a").
		Select(`a.analyst_id,
			COUNT(r.id) AS publish_record_count,
			COALESCE(SUM(r.play_count), 0) AS total_play_count,
			COALESCE(SUM(r.like_count), 0) AS total_like_count,
			COALESCE(SUM(r.comment_count), 0) AS total_comment_count,
			COALESCE(SUM(r.share_count), 0) AS total_share_count,
			COALESCE(SUM(r.favorite_count), 0) AS total_favorite_count,
			COALESCE(MAX(r.play_count), 0) AS max_play_count`).
		Joins("JOIN official_content_publish_records AS r ON r.adoption_id = a.id").
		Where("a.analyst_id IN ?", ids).
		Group("a.analyst_id").
		Scan(&rows).Error; err != nil {
		return err
	}
	byAnalystID := make(map[uint]models.OfficialPublishMetricsSummary, len(rows))
	for _, row := range rows {
		byAnalystID[row.AnalystID] = models.OfficialPublishMetricsSummary{
			PublishRecordCount: row.PublishRecordCount,
			TotalPlayCount:     row.TotalPlayCount,
			TotalLikeCount:     row.TotalLikeCount,
			TotalCommentCount:  row.TotalCommentCount,
			TotalShareCount:    row.TotalShareCount,
			TotalFavoriteCount: row.TotalFavoriteCount,
			MaxPlayCount:       row.MaxPlayCount,
		}
	}
	for index := range analysts {
		metrics := byAnalystID[analysts[index].ID]
		analysts[index].OfficialPublishMetrics = &metrics
	}
	return nil
}

func (s *AnalystLevelService) getOfficialPublishMetrics(analystID uint) (models.OfficialPublishMetricsSummary, error) {
	var row officialPublishMetricsRow
	err := s.db.Table("official_content_adoptions AS a").
		Select(`a.analyst_id,
			COUNT(r.id) AS publish_record_count,
			COALESCE(SUM(r.play_count), 0) AS total_play_count,
			COALESCE(SUM(r.like_count), 0) AS total_like_count,
			COALESCE(SUM(r.comment_count), 0) AS total_comment_count,
			COALESCE(SUM(r.share_count), 0) AS total_share_count,
			COALESCE(SUM(r.favorite_count), 0) AS total_favorite_count,
			COALESCE(MAX(r.play_count), 0) AS max_play_count`).
		Joins("JOIN official_content_publish_records AS r ON r.adoption_id = a.id").
		Where("a.analyst_id = ?", analystID).
		Group("a.analyst_id").
		Scan(&row).Error
	if err != nil {
		return models.OfficialPublishMetricsSummary{}, err
	}
	return models.OfficialPublishMetricsSummary{
		PublishRecordCount: row.PublishRecordCount,
		TotalPlayCount:     row.TotalPlayCount,
		TotalLikeCount:     row.TotalLikeCount,
		TotalCommentCount:  row.TotalCommentCount,
		TotalShareCount:    row.TotalShareCount,
		TotalFavoriteCount: row.TotalFavoriteCount,
		MaxPlayCount:       row.MaxPlayCount,
	}, nil
}

type officialPublishMetricsRow struct {
	AnalystID          uint
	PublishRecordCount int64
	TotalPlayCount     int64
	TotalLikeCount     int64
	TotalCommentCount  int64
	TotalShareCount    int64
	TotalFavoriteCount int64
	MaxPlayCount       int64
}

func (s *AnalystLevelService) ApplyLevelSuggestion(analystID, adminID uint, req AnalystLevelSuggestionActionRequest, now time.Time) (*models.Analyst, error) {
	snapshot, err := s.RefreshGrowthSnapshot(analystID, now)
	if err != nil {
		return nil, err
	}
	if snapshot.SuggestedLevelCode == "" {
		return nil, ErrAnalystLevelInvalid
	}
	if err := s.setAnalystLevelWithTx(s.db, analystID, snapshot.SuggestedLevelCode, adminID, strings.TrimSpace(req.Note), "system_suggestion", snapshot.SuggestedLevelCode, "applied", now); err != nil {
		return nil, err
	}
	if err := s.db.Model(&models.AnalystGrowthSnapshot{}).Where("analyst_id = ?", analystID).Updates(map[string]interface{}{
		"suggestion_status":      "applied",
		"suggestion_reviewed_by": adminID,
		"suggestion_reviewed_at": &now,
		"suggestion_review_note": strings.TrimSpace(req.Note),
		"updated_at":             now,
	}).Error; err != nil {
		return nil, err
	}
	var analyst models.Analyst
	if err := s.db.Preload("User").Preload("GrowthSnapshot").First(&analyst, analystID).Error; err != nil {
		return nil, err
	}
	return &analyst, nil
}

func (s *AnalystLevelService) IgnoreLevelSuggestion(analystID, adminID uint, req AnalystLevelSuggestionActionRequest, now time.Time) (*models.AnalystGrowthSnapshot, error) {
	snapshot, err := s.RefreshGrowthSnapshot(analystID, now)
	if err != nil {
		return nil, err
	}
	if err := s.db.Model(&models.AnalystGrowthSnapshot{}).Where("analyst_id = ?", analystID).Updates(map[string]interface{}{
		"suggestion_status":      "ignored",
		"suggestion_reviewed_by": adminID,
		"suggestion_reviewed_at": &now,
		"suggestion_review_note": strings.TrimSpace(req.Note),
		"updated_at":             now,
	}).Error; err != nil {
		return nil, err
	}
	history := models.AnalystLevelHistory{
		AnalystID:          analystID,
		SuggestedLevelCode: snapshot.SuggestedLevelCode,
		Source:             "system_suggestion",
		Action:             "ignored",
		Note:               strings.TrimSpace(req.Note),
		OperatorID:         adminID,
		CreatedAt:          now,
	}
	if err := s.db.Create(&history).Error; err != nil {
		return nil, err
	}
	if err := s.db.Where("analyst_id = ?", analystID).First(snapshot).Error; err != nil {
		return nil, err
	}
	return snapshot, nil
}

func (s *AnalystLevelService) ListLevelHistories(analystID uint, page, pageSize int) ([]models.AnalystLevelHistory, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	var histories []models.AnalystLevelHistory
	err := s.db.Preload("Analyst").Preload("Analyst.User").
		Where("analyst_id = ?", analystID).
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&histories).Error
	return histories, err
}

func (s *AnalystLevelService) findActiveAnalystByUserID(userID uint) (*models.Analyst, error) {
	var analyst models.Analyst
	if err := s.db.Preload("User").Where("user_id = ?", userID).First(&analyst).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrAnalystLevelAnalystInvalid
		}
		return nil, err
	}
	if analyst.Status != models.AnalystStatusActive {
		return nil, ErrAnalystLevelAnalystInvalid
	}
	return &analyst, nil
}

func (s *AnalystLevelService) setAnalystLevelWithTx(tx *gorm.DB, analystID uint, levelCode string, adminID uint, note, source, suggestedLevel, action string, now time.Time) error {
	var analyst models.Analyst
	if err := tx.First(&analyst, analystID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrAnalystLevelAnalystInvalid
		}
		return err
	}
	updated := tx.Model(&models.Analyst{}).
		Where("id = ?", analystID).
		Updates(map[string]interface{}{
			"level_code":       levelCode,
			"level_updated_at": &now,
			"level_updated_by": adminID,
			"level_note":       note,
			"updated_at":       now,
		})
	if updated.Error != nil {
		return updated.Error
	}
	if updated.RowsAffected == 0 {
		return ErrAnalystLevelAnalystInvalid
	}
	history := models.AnalystLevelHistory{
		AnalystID:          analystID,
		FromLevelCode:      analyst.LevelCode,
		ToLevelCode:        levelCode,
		SuggestedLevelCode: suggestedLevel,
		Source:             source,
		Action:             action,
		Note:               note,
		OperatorID:         adminID,
		CreatedAt:          now,
	}
	if history.Source == "" {
		history.Source = "manual"
	}
	if history.Action == "" {
		history.Action = "applied"
	}
	if err := tx.Create(&history).Error; err != nil {
		return err
	}
	return nil
}

func normalizeLevelCode(code string) (string, error) {
	levelCode := strings.ToUpper(strings.TrimSpace(code))
	if levelCode == "" {
		levelCode = models.DefaultAnalystLevelCode
	}
	if models.AnalystLevelRank(levelCode) == 0 {
		return "", fmt.Errorf("%w: 等级无效", ErrAnalystLevelInvalid)
	}
	return levelCode, nil
}

func suggestAnalystLevel(growthScore float64) (string, string, float64) {
	thresholds := []struct {
		code  string
		score float64
	}{
		{code: "L5", score: 85},
		{code: "L4", score: 70},
		{code: "L3", score: 55},
		{code: "L2", score: 35},
		{code: "L1", score: 0},
	}
	for _, threshold := range thresholds {
		if growthScore >= threshold.score {
			next, gap := nextLevelGap(threshold.code, growthScore)
			return threshold.code, next, gap
		}
	}
	return models.DefaultAnalystLevelCode, "L2", 35
}

func nextLevelGap(current string, score float64) (string, float64) {
	switch current {
	case "L1":
		return "L2", maxFloat(0, 35-score)
	case "L2":
		return "L3", maxFloat(0, 55-score)
	case "L3":
		return "L4", maxFloat(0, 70-score)
	case "L4":
		return "L5", maxFloat(0, 85-score)
	default:
		return "", 0
	}
}

func clampScore(score float64) float64 {
	return maxFloat(0, minFloat(score, 100))
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
