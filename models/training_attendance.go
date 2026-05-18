package models

import (
	"time"

	"gorm.io/gorm"
)

type TrainingAttendanceStatus string

const (
	TrainingAttendancePresent TrainingAttendanceStatus = "present"
	TrainingAttendanceLeave   TrainingAttendanceStatus = "leave"
	TrainingAttendanceAbsent  TrainingAttendanceStatus = "absent"
	TrainingAttendanceLate    TrainingAttendanceStatus = "late"
)

type TrainingAttendance struct {
	ID             uint                     `json:"id" gorm:"primaryKey"`
	TrainingPlanID uint                     `json:"training_plan_id" gorm:"index;not null;uniqueIndex:idx_training_attendance_player"`
	ClubID         uint                     `json:"club_id" gorm:"index;not null"`
	TeamID         uint                     `json:"team_id" gorm:"index;not null"`
	PlayerID       uint                     `json:"player_id" gorm:"index;not null;uniqueIndex:idx_training_attendance_player"`
	Status         TrainingAttendanceStatus `json:"status" gorm:"size:20;not null"`
	Remark         string                   `json:"remark" gorm:"size:255"`
	RecordedBy     uint                     `json:"recorded_by" gorm:"index;not null"`
	RecordedAt     time.Time                `json:"recorded_at"`
	CreatedAt      time.Time                `json:"created_at"`
	UpdatedAt      time.Time                `json:"updated_at"`
	DeletedAt      gorm.DeletedAt           `json:"-" gorm:"index"`

	TrainingPlan *TrainingPlan `json:"trainingPlan,omitempty" gorm:"foreignKey:TrainingPlanID"`
	Player       *User         `json:"player,omitempty" gorm:"foreignKey:PlayerID"`
}

func (TrainingAttendance) TableName() string {
	return "training_attendances"
}
