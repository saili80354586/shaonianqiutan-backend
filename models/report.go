package models

import (
	"time"

	"gorm.io/gorm"
)

// ReportStatus 报告状态
type ReportStatus string

const (
	ReportStatusProcessing ReportStatus = "processing"
	ReportStatusCompleted  ReportStatus = "completed"
	ReportStatusFailed     ReportStatus = "failed"
)

// Report 球探报告模型
type Report struct {
	ID              uint         `json:"id" gorm:"primaryKey"`
	OrderID         uint         `json:"order_id" gorm:"not null;index"`
	UserID          uint         `json:"user_id" gorm:"not null;index"`
	AnalystID       uint         `json:"analyst_id" gorm:"not null;index"`
	PlayerName      string       `json:"player_name" gorm:"size:50;not null"`
	PlayerBirthDate string       `json:"player_birth_date" gorm:"size:10"`
	PlayerPosition  string       `json:"player_position" gorm:"size:50"`
	PlayerProvince  string       `json:"player_province" gorm:"size:50"`
	PlayerCity      string       `json:"player_city" gorm:"size:50"`
	Content         string       `json:"content" gorm:"type:text;not null"`
	PdfURL          string       `json:"pdf_url" gorm:"size:255"`
	Status          ReportStatus `json:"status" gorm:"size:20;default:'processing'"`
	ReviewRemark    string       `json:"review_remark" gorm:"type:text"`

	// ===== 新增：评分结构化字段 =====
	OverallRating float64 `json:"overall_rating" gorm:"type:decimal(3,1)"`
	OffenseRating float64 `json:"offense_rating" gorm:"type:decimal(3,1)"`
	DefenseRating float64 `json:"defense_rating" gorm:"type:decimal(3,1)"`
	Summary       string  `json:"summary" gorm:"type:text"`
	Strengths     string  `json:"strengths" gorm:"type:text"`  // JSON 数组
	Weaknesses    string  `json:"weaknesses" gorm:"type:text"` // JSON 数组
	Suggestions   string  `json:"suggestions" gorm:"type:text"`
	Potential     string  `json:"potential" gorm:"size:20"`       // top | high | medium | low
	ClipVideoURL  string  `json:"clip_video_url" gorm:"size:500"` // 视频版剪辑地址

	// ===== 新增：19 项评分明细（统一 JSON 存储）=====
	RatingDetails string `json:"rating_details" gorm:"type:text"` // JSON 对象

	// ===== 新增：MD 文档路径 =====
	RatingReportMD string `json:"rating_report_md" gorm:"size:500"` // 球员评分报告 MD 文件路径
	PlayerInfoMD   string `json:"player_info_md" gorm:"size:500"`   // 球员基础信息 MD 文件路径

	// ===== 新增：AI 生成报告路径 =====
	AIReportURL string `json:"ai_report_url" gorm:"size:500"` // AI 生成的 Word 报告 URL
	AIVideoURL  string `json:"ai_video_url" gorm:"size:500"`  // AI 生成的视频分析 URL

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// ReportListResult 报告列表查询结果
type ReportListResult struct {
	List     []Report `json:"list"`
	Total    int64    `json:"total"`
	Page     int      `json:"page"`
	PageSize int      `json:"pageSize"`
}

// ReportRepository 报告数据访问层
type ReportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

// Create 创建报告
func (r *ReportRepository) Create(report *Report) error {
	return r.db.Create(report).Error
}

// Update 更新报告
func (r *ReportRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&Report{}).Where("id = ?", id).Updates(updates).Error
}

// FindByID 根据ID查询报告
func (r *ReportRepository) FindByID(id uint) (*Report, error) {
	var report Report
	result := r.db.First(&report, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &report, nil
}

// FindByOrderID 根据订单ID查询报告
func (r *ReportRepository) FindByOrderID(orderID uint) (*Report, error) {
	var report Report
	result := r.db.Where("order_id = ?", orderID).First(&report)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &report, nil
}

// FindByUserID 获取用户报告列表
func (r *ReportRepository) FindByUserID(userID uint, page, pageSize int) (*ReportListResult, error) {
	var result ReportListResult
	var reports []Report

	offset := (page - 1) * pageSize

	query := r.db.Model(&Report{}).Where("user_id = ? AND status = ?", userID, ReportStatusCompleted)
	var total int64
	query.Count(&total)

	query = query.Order("created_at DESC").Limit(pageSize).Offset(offset)
	if err := query.Find(&reports).Error; err != nil {
		return nil, err
	}

	result.List = reports
	result.Total = total
	result.Page = page
	result.PageSize = pageSize

	return &result, nil
}

// FindByAnalystID 获取分析师报告列表
func (r *ReportRepository) FindByAnalystID(analystID uint, page, pageSize int) (*ReportListResult, error) {
	var result ReportListResult
	var reports []Report

	offset := (page - 1) * pageSize

	query := r.db.Model(&Report{}).Where("analyst_id = ?", analystID)
	var total int64
	query.Count(&total)

	query = query.Order("created_at DESC").Limit(pageSize).Offset(offset)
	if err := query.Find(&reports).Error; err != nil {
		return nil, err
	}

	result.List = reports
	result.Total = total
	result.Page = page
	result.PageSize = pageSize

	return &result, nil
}

// FindByStatus 根据状态获取报告列表
func (r *ReportRepository) FindByStatus(status ReportStatus, page, pageSize int) ([]Report, int64, error) {
	var reports []Report
	var total int64

	query := r.db.Model(&Report{}).Where("status = ?", status).Order("created_at DESC")
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Find(&reports).Error
	return reports, total, err
}

// ReportStatistics 报告统计数据
type ReportStatistics struct {
	TotalCount     int64
	TodayCount     int64
	PendingCount   int64
	CompletedCount int64
}

// GetStatistics 获取报告统计数据
func (r *ReportRepository) GetStatistics() (*ReportStatistics, error) {
	stats := &ReportStatistics{}

	// 总报告数
	if err := r.db.Model(&Report{}).Count(&stats.TotalCount).Error; err != nil {
		return nil, err
	}

	// 今日报告数
	today := time.Now().Format("2006-01-02")
	if err := r.db.Model(&Report{}).Where("DATE(created_at) = ?", today).Count(&stats.TodayCount).Error; err != nil {
		return nil, err
	}

	// 处理中报告数
	if err := r.db.Model(&Report{}).Where("status = ?", ReportStatusProcessing).Count(&stats.PendingCount).Error; err != nil {
		return nil, err
	}

	// 已完成报告数
	if err := r.db.Model(&Report{}).Where("status = ?", ReportStatusCompleted).Count(&stats.CompletedCount).Error; err != nil {
		return nil, err
	}

	return stats, nil
}
