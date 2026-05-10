package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// ReportVersionSource marks where a report version came from.
type ReportVersionSource string

const (
	ReportVersionSourceAI          ReportVersionSource = "ai_generated"
	ReportVersionSourceSystem      ReportVersionSource = "system_generated"
	ReportVersionSourceOnlineEdit  ReportVersionSource = "online_edit"
	ReportVersionSourceAnalystWord ReportVersionSource = "analyst_word_upload"
	ReportVersionSourceAdminWord   ReportVersionSource = "admin_word_upload"
	ReportVersionSourceAdminReview ReportVersionSource = "admin_review"
)

// ReportVersionStatus is the auditable lifecycle status for a version.
type ReportVersionStatus string

const (
	ReportVersionStatusAIGenerating     ReportVersionStatus = "ai_generating"
	ReportVersionStatusAIDraft          ReportVersionStatus = "ai_draft"
	ReportVersionStatusAnalystEditing   ReportVersionStatus = "analyst_editing"
	ReportVersionStatusAnalystSubmitted ReportVersionStatus = "analyst_submitted"
	ReportVersionStatusAdminRejected    ReportVersionStatus = "admin_rejected"
	ReportVersionStatusApproved         ReportVersionStatus = "approved"
)

// ReportVersion stores every important AI/Word/PDF report handoff.
type ReportVersion struct {
	ID                      uint                `json:"id" gorm:"primaryKey"`
	ReportID                uint                `json:"report_id" gorm:"index;not null"`
	OrderID                 uint                `json:"order_id" gorm:"index;not null"`
	AnalysisID              *uint               `json:"analysis_id" gorm:"index"`
	VersionNo               int                 `json:"version_no" gorm:"not null;index"`
	SourceType              ReportVersionSource `json:"source_type" gorm:"size:32;not null;index"`
	Status                  ReportVersionStatus `json:"status" gorm:"size:32;not null;index"`
	Content                 string              `json:"content" gorm:"type:longtext"`
	WordURL                 string              `json:"word_url" gorm:"size:500"`
	PDFURL                  string              `json:"pdf_url" gorm:"size:500"`
	InputSnapshot           string              `json:"input_snapshot" gorm:"type:longtext"`
	TemplateVersion         string              `json:"template_version" gorm:"size:80;index"`
	DocumentTemplateVersion string              `json:"document_template_version" gorm:"size:80;index"`
	OriginalFileName        string              `json:"original_file_name" gorm:"size:255"`
	ReviewRemark            string              `json:"review_remark" gorm:"type:text"`
	CreatedByUserID         *uint               `json:"created_by_user_id" gorm:"index"`
	CreatedByRole           string              `json:"created_by_role" gorm:"size:32"`
	CreatedAt               time.Time           `json:"created_at"`
}

// TableName returns the report version table name.
func (ReportVersion) TableName() string {
	return "report_versions"
}

// CreateReportVersion appends a version record. If VersionNo is empty, the next
// number for that report is assigned inside the current transaction.
func CreateReportVersion(db *gorm.DB, version *ReportVersion) error {
	if db == nil || version == nil || version.ReportID == 0 {
		return nil
	}
	if !db.Migrator().HasTable(&ReportVersion{}) {
		return nil
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if version.VersionNo <= 0 {
			next, err := NextReportVersionNo(tx, version.ReportID)
			if err != nil {
				return err
			}
			version.VersionNo = next
		}
		return tx.Create(version).Error
	})
}

// NextReportVersionNo returns max(version_no)+1 for a report.
func NextReportVersionNo(db *gorm.DB, reportID uint) (int, error) {
	if db == nil || reportID == 0 {
		return 1, nil
	}
	var maxVersion int
	if err := db.Model(&ReportVersion{}).
		Where("report_id = ?", reportID).
		Select("COALESCE(MAX(version_no), 0)").
		Scan(&maxVersion).Error; err != nil {
		return 0, err
	}
	return maxVersion + 1, nil
}

// FindReportVersionsByReportID lists report versions newest first.
func FindReportVersionsByReportID(db *gorm.DB, reportID uint) ([]ReportVersion, error) {
	if db == nil || reportID == 0 {
		return nil, nil
	}
	if !db.Migrator().HasTable(&ReportVersion{}) {
		return nil, nil
	}
	var versions []ReportVersion
	err := db.Where("report_id = ?", reportID).
		Order("version_no DESC, created_at DESC").
		Find(&versions).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return versions, err
}
