package models

import "time"

// VideoClipExportJobStatus 批量片段导出任务状态
type VideoClipExportJobStatus string

const (
	VideoClipExportQueued     VideoClipExportJobStatus = "queued"
	VideoClipExportProcessing VideoClipExportJobStatus = "processing"
	VideoClipExportReady      VideoClipExportJobStatus = "ready"
	VideoClipExportFailed     VideoClipExportJobStatus = "failed"
)

// VideoClipExportJob 记录分析师批量导出片段的后台任务
type VideoClipExportJob struct {
	ID          uint                     `json:"-" gorm:"primaryKey"`
	JobID       string                   `json:"id" gorm:"uniqueIndex;size:64;not null"`
	AnalysisID  uint                     `json:"analysis_id" gorm:"index;not null"`
	AnalystID   uint                     `json:"-" gorm:"index;not null"`
	Status      VideoClipExportJobStatus `json:"status" gorm:"size:32;index;not null"`
	Progress    int                      `json:"progress"`
	Processed   int                      `json:"processed"`
	Total       int                      `json:"total"`
	FileName    string                   `json:"filename" gorm:"size:255"`
	ZipPath     string                   `json:"-" gorm:"size:500"`
	RequestJSON string                   `json:"-" gorm:"type:text"`
	Error       string                   `json:"error,omitempty" gorm:"type:text"`
	ExpiresAt   *time.Time               `json:"expires_at,omitempty"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
}
