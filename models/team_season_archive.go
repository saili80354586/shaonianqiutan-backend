package models

import "time"

// TeamSeasonArchive 球队赛季档案
type TeamSeasonArchive struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TeamID      uint      `gorm:"index;not null" json:"teamId"`
	SeasonName  string    `gorm:"size:64;not null" json:"seasonName"`
	StartDate   string    `gorm:"size:32" json:"startDate"`
	EndDate     string    `gorm:"size:32" json:"endDate"`
	MatchCount  int       `gorm:"default:0" json:"matchCount"`
	WeeklyCount int       `gorm:"default:0" json:"weeklyCount"`
	TestCount   int       `gorm:"default:0" json:"testCount"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedBy   uint      `gorm:"not null" json:"createdBy"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// TeamSeasonArchiveResponse 前端响应结构
type TeamSeasonArchiveResponse struct {
	ID          uint   `json:"id"`
	TeamID      uint   `json:"teamId"`
	SeasonName  string `json:"seasonName"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	MatchCount  int    `json:"matchCount"`
	WeeklyCount int    `json:"weeklyCount"`
	TestCount   int    `json:"testCount"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}
