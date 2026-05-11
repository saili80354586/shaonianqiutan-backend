package models

import (
	"time"

	"gorm.io/gorm"
)

type AnalysisOperationEvent struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	OrderID       uint           `json:"order_id" gorm:"not null;index:idx_analysis_operation_order_time"`
	AnalysisID    uint           `json:"analysis_id" gorm:"not null;index:idx_analysis_operation_analysis_time"`
	AnalystID     uint           `json:"analyst_id" gorm:"not null;index:idx_analysis_operation_analyst_time"`
	EventType     string         `json:"event_type" gorm:"size:64;not null;index"`
	Section       string         `json:"section" gorm:"size:32;not null;index"`
	FieldKey      string         `json:"field_key" gorm:"size:128;index"`
	FieldLabel    string         `json:"field_label" gorm:"size:128"`
	BeforeSummary string         `json:"before_summary" gorm:"type:text"`
	AfterSummary  string         `json:"after_summary" gorm:"type:text"`
	Metadata      string         `json:"metadata" gorm:"type:text"`
	CreatedAt     time.Time      `json:"created_at" gorm:"index:idx_analysis_operation_order_time;index:idx_analysis_operation_analysis_time;index:idx_analysis_operation_analyst_time"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

func (AnalysisOperationEvent) TableName() string {
	return "analysis_operation_events"
}

type AnalysisOperationEventRepository struct {
	db *gorm.DB
}

func NewAnalysisOperationEventRepository(db *gorm.DB) *AnalysisOperationEventRepository {
	return &AnalysisOperationEventRepository{db: db}
}

func (r *AnalysisOperationEventRepository) Create(event *AnalysisOperationEvent) error {
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}
	return r.db.Create(event).Error
}

func (r *AnalysisOperationEventRepository) CreateWithTx(tx *gorm.DB, event *AnalysisOperationEvent) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}
	return db.Create(event).Error
}

func (r *AnalysisOperationEventRepository) FindByOrderID(orderID uint, limit int) ([]AnalysisOperationEvent, error) {
	var events []AnalysisOperationEvent
	query := r.db.Where("order_id = ?", orderID).Order("created_at DESC, id DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&events).Error
	return events, err
}

func (r *AnalysisOperationEventRepository) FindByAnalysisID(analysisID uint, limit int) ([]AnalysisOperationEvent, error) {
	var events []AnalysisOperationEvent
	query := r.db.Where("analysis_id = ?", analysisID).Order("created_at DESC, id DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&events).Error
	return events, err
}
