package models

import "time"

// TeamApplication 入队/试训申请模型
type TeamApplication struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	TeamID       uint       `json:"team_id" gorm:"index;not null"`
	ClubID       uint       `json:"club_id" gorm:"index;not null"`
	PlayerID     uint       `json:"player_id" gorm:"index;not null"`
	Type         string     `json:"type" gorm:"size:20;not null"`       // join / trial
	Status       string     `json:"status" gorm:"size:20;default:'pending';index"` // pending / approved / rejected / cancelled
	Reason       string     `json:"reason" gorm:"type:text"`
	ResponseNote string     `json:"response_note" gorm:"type:text"`
	ReviewedBy   *uint      `json:"reviewed_by" gorm:"index"`
	ReviewedAt   *time.Time `json:"reviewed_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// 关联
	Team     *Team     `json:"team,omitempty" gorm:"foreignKey:TeamID"`
	Club     *Club     `json:"club,omitempty" gorm:"foreignKey:ClubID"`
	Player   *User     `json:"player,omitempty" gorm:"foreignKey:PlayerID"`
	Reviewer *User     `json:"reviewer,omitempty" gorm:"foreignKey:ReviewedBy"`
}

// TableName 表名
func (TeamApplication) TableName() string {
	return "team_applications"
}
