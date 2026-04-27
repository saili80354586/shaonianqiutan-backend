package models

import "time"

// TrialInviteStatus 试训邀请状态
type TrialInviteStatus string

const (
	TrialInvitePending   TrialInviteStatus = "pending"
	TrialInviteAccepted  TrialInviteStatus = "accepted"
	TrialInviteDeclined  TrialInviteStatus = "declined"
	TrialInviteCompleted TrialInviteStatus = "completed"
)

// TrialInvite 试训邀请模型
type TrialInvite struct {
	ID            uint              `json:"id" gorm:"primaryKey"`
	SenderID      uint              `json:"sender_id" gorm:"not null;index"`
	PlayerID      uint              `json:"player_id" gorm:"not null;index"`
	TrialDate     string            `json:"trial_date" gorm:"size:10;not null"`
	TrialTime     string            `json:"trial_time" gorm:"size:10"`
	Location      string            `json:"location" gorm:"size:255;not null"`
	ContactName   string            `json:"contact_name" gorm:"size:50;not null"`
	ContactPhone  string            `json:"contact_phone" gorm:"size:20;not null"`
	Note          string            `json:"note" gorm:"type:text"`
	Status        TrialInviteStatus `json:"status" gorm:"size:20;default:'pending'"`
	ResponseNote  string            `json:"response_note" gorm:"type:text"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	RespondedAt   *time.Time        `json:"responded_at"`
}
