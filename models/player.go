package models

import (
	"time"

	"gorm.io/gorm"
)

// Player 球员信息模型
type Player struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint      `gorm:"index" json:"user_id"`
	Name       string    `gorm:"size:50;not null" json:"name"`
	Nickname   string    `gorm:"size:50" json:"nickname"`
	Province   string    `gorm:"size:50" json:"province"`
	City       string    `gorm:"size:50" json:"city"`
	District   string    `gorm:"size:50" json:"district"`
	Position   string    `gorm:"size:20" json:"position"`
	Age        int       `json:"age"`
	BirthDate  string    `gorm:"size:20" json:"birth_date"`
	Height     int       `json:"height"`
	Weight     int       `json:"weight"`
	Foot       string    `gorm:"size:10" json:"foot"`
	Club       string    `gorm:"size:100" json:"club"`
	School     string    `gorm:"size:100" json:"school"`
	Phone      string    `gorm:"size:20" json:"phone"`
	Avatar     string    `gorm:"size:255" json:"avatar"`
	VideoURL   string    `gorm:"size:255" json:"video_url"`
	Status     int       `gorm:"default:1" json:"status"` // 1:正常 0:禁用
	CreateTime time.Time `gorm:"autoCreateTime" json:"create_time"`
	UpdateTime time.Time `gorm:"autoUpdateTime" json:"update_time"`
}

// TableName 设置表名
func (Player) TableName() string {
	return "players"
}

// PlayerRepository 球员数据仓库
type PlayerRepository struct {
	db *gorm.DB
}

// NewPlayerRepository 创建球员仓库
func NewPlayerRepository(db *gorm.DB) *PlayerRepository {
	return &PlayerRepository{db: db}
}

// GetAllPlayers 获取所有球员
func (r *PlayerRepository) GetAllPlayers() ([]Player, error) {
	var players []Player
	err := r.db.Where("status = ?", 1).Find(&players).Error
	return players, err
}

// GetPlayersByProvince 按省份获取球员
func (r *PlayerRepository) GetPlayersByProvince(province string) ([]Player, error) {
	var players []Player
	err := r.db.Where("province = ? AND status = ?", province, 1).Find(&players).Error
	return players, err
}

// GetPlayersByCity 按城市获取球员
func (r *PlayerRepository) GetPlayersByCity(city string) ([]Player, error) {
	var players []Player
	err := r.db.Where("city = ? AND status = ?", city, 1).Find(&players).Error
	return players, err
}

// CreatePlayer 创建球员
func (r *PlayerRepository) CreatePlayer(player *Player) error {
	return r.db.Create(player).Error
}

// UpdatePlayer 更新球员
func (r *PlayerRepository) UpdatePlayer(player *Player) error {
	return r.db.Save(player).Error
}

// DeletePlayer 删除球员（软删除）
func (r *PlayerRepository) DeletePlayer(id uint) error {
	return r.db.Model(&Player{}).Where("id = ?", id).Update("status", 0).Error
}

// GetPlayerCount 获取球员总数
func (r *PlayerRepository) GetPlayerCount() (int64, error) {
	var count int64
	err := r.db.Model(&Player{}).Where("status = ?", 1).Count(&count).Error
	return count, err
}

// GetProvinceStats 获取省份统计
func (r *PlayerRepository) GetProvinceStats() (map[string]int64, error) {
	var results []struct {
		Province string
		Count    int64
	}

	err := r.db.Model(&Player{}).
		Select("province, COUNT(*) as count").
		Where("status = ? AND province != ?", 1, "").
		Group("province").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	stats := make(map[string]int64)
	for _, r := range results {
		stats[r.Province] = r.Count
	}

	return stats, nil
}
