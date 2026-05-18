package models

import (
	"time"

	"gorm.io/gorm"
)

// TeamMonthlyReportArchive stores a locked monthly report snapshot for a team.
type TeamMonthlyReportArchive struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	TeamID     uint           `json:"teamId" gorm:"not null;uniqueIndex:idx_team_monthly_report_archive"`
	Month      string         `json:"month" gorm:"size:7;not null;uniqueIndex:idx_team_monthly_report_archive"`
	Version    int            `json:"version" gorm:"not null;default:1"`
	Snapshot   string         `json:"snapshot" gorm:"type:longtext;not null"`
	ArchivedBy uint           `json:"archivedBy" gorm:"not null"`
	ArchivedAt time.Time      `json:"archivedAt" gorm:"not null;index"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

func (TeamMonthlyReportArchive) TableName() string {
	return "team_monthly_report_archives"
}

// TeamMonthlyReportArchiveVersion stores each saved version for audit/history.
type TeamMonthlyReportArchiveVersion struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	ArchiveID    uint           `json:"archiveId" gorm:"not null;uniqueIndex:idx_team_monthly_report_archive_version"`
	TeamID       uint           `json:"teamId" gorm:"not null;index"`
	Month        string         `json:"month" gorm:"size:7;not null;index"`
	Version      int            `json:"version" gorm:"not null;uniqueIndex:idx_team_monthly_report_archive_version"`
	Snapshot     string         `json:"snapshot" gorm:"type:longtext;not null"`
	ArchivedBy   uint           `json:"archivedBy" gorm:"not null"`
	ArchivedAt   time.Time      `json:"archivedAt" gorm:"not null;index"`
	ReviewStatus string         `json:"reviewStatus" gorm:"size:32;default:'pending';index"`
	ReviewNote   string         `json:"reviewNote" gorm:"type:text"`
	ReviewedBy   *uint          `json:"reviewedBy" gorm:"index"`
	ReviewedAt   *time.Time     `json:"reviewedAt"`
	CreatedAt    time.Time      `json:"createdAt"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

func (TeamMonthlyReportArchiveVersion) TableName() string {
	return "team_monthly_report_archive_versions"
}

// TeamMonthlyReportArchiveReviewEvent stores each review action for a monthly archive version.
type TeamMonthlyReportArchiveReviewEvent struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	ArchiveID uint           `json:"archiveId" gorm:"not null;index"`
	VersionID uint           `json:"versionId" gorm:"not null;index"`
	TeamID    uint           `json:"teamId" gorm:"not null;index"`
	Month     string         `json:"month" gorm:"size:7;not null;index"`
	Version   int            `json:"version" gorm:"not null;index"`
	Status    string         `json:"status" gorm:"size:32;not null;index"`
	Note      string         `json:"note" gorm:"type:text"`
	ActorID   uint           `json:"actorId" gorm:"not null;index"`
	CreatedAt time.Time      `json:"createdAt" gorm:"not null;index"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (TeamMonthlyReportArchiveReviewEvent) TableName() string {
	return "team_monthly_report_archive_review_events"
}

// TeamMonthlyReportArchiveAdjustmentItem stores a trackable revision item for a monthly archive version.
type TeamMonthlyReportArchiveAdjustmentItem struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	ArchiveID   uint           `json:"archiveId" gorm:"not null;index"`
	VersionID   uint           `json:"versionId" gorm:"not null;index"`
	TeamID      uint           `json:"teamId" gorm:"not null;index"`
	Month       string         `json:"month" gorm:"size:7;not null;index"`
	Version     int            `json:"version" gorm:"not null;index"`
	Content     string         `json:"content" gorm:"type:text;not null"`
	Status      string         `json:"status" gorm:"size:20;not null;default:'open';index"`
	CreatedBy   uint           `json:"createdBy" gorm:"not null;index"`
	CompletedBy *uint          `json:"completedBy" gorm:"index"`
	CompletedAt *time.Time     `json:"completedAt"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (TeamMonthlyReportArchiveAdjustmentItem) TableName() string {
	return "team_monthly_report_archive_adjustment_items"
}
