package models

import (
	"time"
)

// TeamHome 球队主页配置模型
type TeamHome struct {
	ID        uint          `json:"id" gorm:"primaryKey"`
	TeamID    uint          `json:"team_id" gorm:"uniqueIndex;not null"`
	Hero      TeamHomeHero  `json:"hero" gorm:"serializer:json"`
	About     TeamHomeAbout `json:"about" gorm:"serializer:json"`
	Honors    []TeamHonor   `json:"honors" gorm:"serializer:json"`
	Dynamics  []TeamDynamic `json:"dynamics" gorm:"-"`
	Contact   TeamHomeContact `json:"contact" gorm:"serializer:json"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

func (TeamHome) TableName() string {
	return "team_homes"
}

// TeamHomeHero Hero 区域配置
type TeamHomeHero struct {
	Title           string `json:"title"`            // 主标题
	Subtitle        string `json:"subtitle"`         // 副标题
	BackgroundImage string `json:"backgroundImage"`  // 背景图
	Logo            string `json:"logo"`             // 球队Logo
	AgeGroup        string `json:"ageGroup"`         // 年龄组
	FoundedYear     string `json:"foundedYear"`      // 成立年份
	ShowStats       bool   `json:"showStats"`        // 是否显示统计
}

// TeamHomeAbout About 区域配置
type TeamHomeAbout struct {
	Enabled  bool     `json:"enabled"`  // 是否启用
	Title    string   `json:"title"`   // 标题
	Content  string   `json:"content"` // 内容（富文本）
	Images   []string `json:"images"`   // 图片列表
}

// TeamHonor 球队荣誉
type TeamHonor struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`        // 荣誉名称
	Description string `json:"description"`  // 描述
	Icon        string `json:"icon"`         // 图标
	Year        string `json:"year"`         // 获奖年份
	Count       string `json:"count"`        // 数量
	Sort        int    `json:"sort"`         // 排序
}

// TeamDynamic 球队动态
type TeamDynamic struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`     // 动态标题
	Content   string    `json:"content"`   // 内容
	Images    []string  `json:"images"`    // 图片
	Type      string    `json:"type"`      // 类型: training/match/activity
	CreatedAt time.Time `json:"createdAt"`
}

// TeamHomeContact 联系方式配置
type TeamHomeContact struct {
	Enabled bool   `json:"enabled"` // 是否启用
	Phone   string `json:"phone"`   // 电话
	Wechat  string `json:"wechat"`  // 微信
	Address string `json:"address"` // 地址
}

// DefaultTeamHome 创建默认配置
func DefaultTeamHome(teamID uint) *TeamHome {
	return &TeamHome{
		TeamID: teamID,
		Hero: TeamHomeHero{
			Title:      "球队主页",
			Subtitle:   "青少年足球球队",
			ShowStats:  true,
		},
		About: TeamHomeAbout{
			Enabled: true,
			Title:   "关于我们",
			Content: "球队成立于2020年，致力于青少年足球培训...",
			Images:  []string{},
		},
		Honors:  []TeamHonor{},
		Contact: TeamHomeContact{
			Enabled: true,
		},
	}
}

// CoachTeamHomeResponse 教练视角的球队主页响应
type CoachTeamHomeResponse struct {
	TeamID      uint            `json:"teamId"`
	TeamName    string          `json:"teamName"`
	AgeGroup    string          `json:"ageGroup"`
	Hero        TeamHomeHero    `json:"hero"`
	About       TeamHomeAbout   `json:"about"`
	Honors      []TeamHonor     `json:"honors"`
	Contact     TeamHomeContact `json:"contact"`
	PlayerCount int             `json:"playerCount"`
	CoachCount  int             `json:"coachCount"`
}