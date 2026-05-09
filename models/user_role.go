package models

import (
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UserRoleRecord 是账号多身份的统一状态表，用于承接“一个账号拥有多个业务身份”的申请、审核和权限读取。
type UserRoleRecord struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	UserID        uint           `json:"user_id" gorm:"uniqueIndex:idx_user_role;not null"`
	Role          UserRole       `json:"role" gorm:"size:20;uniqueIndex:idx_user_role;not null"`
	Status        string         `json:"status" gorm:"size:20;index;default:'pending'"`
	Source        string         `json:"source" gorm:"size:50;default:'self_apply'"`
	ProfileID     uint           `json:"profile_id"`
	PublicVisible bool           `json:"public_visible" gorm:"default:true"`
	RejectReason  string         `json:"reject_reason" gorm:"type:text"`
	ApprovedAt    *time.Time     `json:"approved_at"`
	ApprovedBy    uint           `json:"approved_by"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

func (UserRoleRecord) TableName() string {
	return "user_roles"
}

func IsMissingUserRolesTableError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "no such table: user_roles") ||
		strings.Contains(message, "table 'user_roles' doesn't exist")
}

func UpsertUserRoleRecord(db *gorm.DB, record UserRoleRecord) error {
	values := map[string]interface{}{
		"status":         record.Status,
		"source":         record.Source,
		"profile_id":     record.ProfileID,
		"public_visible": record.PublicVisible,
		"reject_reason":  record.RejectReason,
		"approved_at":    record.ApprovedAt,
		"approved_by":    record.ApprovedBy,
		"updated_at":     time.Now(),
	}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "role"}},
		DoUpdates: clause.Assignments(values),
	}).Create(&record).Error
}

// BackfillUserRoleRecords 将旧的单角色字段和既有业务资料补入统一身份表，避免上线后老账号无法切换身份。
func BackfillUserRoleRecords(db *gorm.DB) error {
	var users []User
	if err := db.Where("status = ?", StatusActive).Find(&users).Error; err != nil {
		return err
	}
	now := time.Now()
	for _, user := range users {
		role := user.Role
		if role == "" {
			role = RoleUser
		}
		if err := UpsertUserRoleRecord(db, UserRoleRecord{
			UserID:        user.ID,
			Role:          role,
			Status:        "active",
			Source:        "legacy_user_role",
			PublicVisible: true,
			ApprovedAt:    &now,
		}); err != nil {
			return err
		}
	}

	var analysts []Analyst
	if err := db.Where("status = ?", AnalystStatusActive).Find(&analysts).Error; err != nil {
		return err
	}
	for _, analyst := range analysts {
		if err := UpsertUserRoleRecord(db, UserRoleRecord{
			UserID:        analyst.UserID,
			Role:          RoleAnalyst,
			Status:        "active",
			Source:        "legacy_analyst_profile",
			ProfileID:     analyst.ID,
			PublicVisible: true,
			ApprovedAt:    &now,
		}); err != nil {
			return err
		}
	}

	var scouts []Scout
	if err := db.Find(&scouts).Error; err != nil {
		return err
	}
	for _, scout := range scouts {
		if err := UpsertUserRoleRecord(db, UserRoleRecord{
			UserID:        scout.UserID,
			Role:          RoleScout,
			Status:        "active",
			Source:        "legacy_scout_profile",
			ProfileID:     scout.ID,
			PublicVisible: true,
			ApprovedAt:    &now,
		}); err != nil {
			return err
		}
	}

	var clubs []Club
	if err := db.Find(&clubs).Error; err != nil {
		return err
	}
	for _, club := range clubs {
		if err := UpsertUserRoleRecord(db, UserRoleRecord{
			UserID:        club.UserID,
			Role:          RoleClub,
			Status:        "active",
			Source:        "legacy_club_profile",
			ProfileID:     club.ID,
			PublicVisible: true,
			ApprovedAt:    &now,
		}); err != nil {
			return err
		}
	}

	var clubCoaches []ClubCoach
	if err := db.Where("status = ?", ClubCoachStatusActive).Find(&clubCoaches).Error; err != nil {
		return err
	}
	for _, coach := range clubCoaches {
		if err := UpsertUserRoleRecord(db, UserRoleRecord{
			UserID:        coach.UserID,
			Role:          RoleCoach,
			Status:        "active",
			Source:        "legacy_club_coach",
			ProfileID:     coach.ID,
			PublicVisible: true,
			ApprovedAt:    &now,
		}); err != nil {
			return err
		}
	}

	var teamCoaches []TeamCoach
	if err := db.Where("status = ?", "active").Find(&teamCoaches).Error; err != nil {
		return err
	}
	for _, coach := range teamCoaches {
		if err := UpsertUserRoleRecord(db, UserRoleRecord{
			UserID:        coach.UserID,
			Role:          RoleCoach,
			Status:        "active",
			Source:        "legacy_team_coach",
			ProfileID:     coach.ID,
			PublicVisible: true,
			ApprovedAt:    &now,
		}); err != nil {
			return err
		}
	}

	return nil
}
