package models

import "time"

// PlayerFilterPreset 球员筛选方案预设
type PlayerFilterPreset struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ClubID    uint      `json:"club_id" gorm:"index;not null"`
	Name      string    `json:"name" gorm:"size:100;not null"`
	Filters   string    `json:"filters" gorm:"type:text"` // JSON: { ageGroup, position, minHeight, maxHeight, minWeight, maxWeight }
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
