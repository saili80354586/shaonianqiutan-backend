package models

import "time"

// PlayerShortlist 候选名单（选材决策）
type PlayerShortlist struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ClubID    uint      `gorm:"index:idx_club_player;not null" json:"clubId"`
	PlayerID  uint      `gorm:"index:idx_club_player;not null" json:"playerId"`
	Note      string    `gorm:"type:text" json:"note"`
	CreatedBy uint      `gorm:"not null" json:"createdBy"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// PlayerShortlistResponse 前端响应结构
type PlayerShortlistResponse struct {
	ID        uint   `json:"id"`
	PlayerID  uint   `json:"playerId"`
	Note      string `json:"note"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}
