package models

import (
	"time"
)

// ClubHome 俱乐部主页配置模型
type ClubHome struct {
	ID               uint                `json:"id" gorm:"primaryKey"`
	ClubID           uint                `json:"club_id" gorm:"uniqueIndex;not null"`
	Hero             ClubHomeHero        `json:"hero" gorm:"serializer:json"`
	About            ClubHomeAbout       `json:"about" gorm:"serializer:json"`
	Contact          ClubHomeContact     `json:"contact" gorm:"serializer:json"`
	Facilities       ClubHomeFacilities  `json:"facilities" gorm:"serializer:json"`
	Recruitment      ClubHomeRecruitment `json:"recruitment" gorm:"serializer:json"`
	SocialLinks      ClubHomeSocialLinks `json:"socialLinks" gorm:"serializer:json"`
	NewsItems        []ClubHomeNewsItem  `json:"newsItems" gorm:"serializer:json"`
	ModuleOrder      string              `json:"moduleOrder" gorm:"type:text"`
	ModuleVisibility map[string]bool     `json:"moduleVisibility" gorm:"serializer:json"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
}

func (ClubHome) TableName() string {
	return "club_homes"
}

// ClubHomeHero Hero 区域配置
type ClubHomeHero struct {
	Title           string `json:"title"`           // 主标题
	Subtitle        string `json:"subtitle"`        // 副标题
	BackgroundImage string `json:"backgroundImage"` // 背景图
	ShowStats       bool   `json:"showStats"`       // 是否显示统计
}

// ClubHomeAbout About 区域配置
type ClubHomeAbout struct {
	Enabled  bool              `json:"enabled"`  // 是否启用
	Title    string            `json:"title"`    // 标题
	Content  string            `json:"content"`  // 内容
	Images   []string          `json:"images"`   // 图片列表
	Features []ClubHomeFeature `json:"features"` // 特色标签
}

// ClubHomeFeature 特色标签
type ClubHomeFeature struct {
	Icon        string `json:"icon"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// ClubHomeContact 联系方式配置
type ClubHomeContact struct {
	Enabled bool   `json:"enabled"` // 是否启用
	Phone   string `json:"phone"`   // 电话
	Wechat  string `json:"wechat"`  // 微信
	Address string `json:"address"` // 地址
	Email   string `json:"email"`   // 邮箱
}

// ClubHomeFacilities 训练环境配置
type ClubHomeFacilities struct {
	Enabled     bool                   `json:"enabled"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Images      []string               `json:"images"`
	Schedule    []ClubHomeScheduleItem `json:"schedule"`
}

// ClubHomeScheduleItem 训练时间安排
type ClubHomeScheduleItem struct {
	Day       string `json:"day"`
	TimeRange string `json:"timeRange"`
	Group     string `json:"group"`
}

// ClubHomeRecruitment 招生信息配置
type ClubHomeRecruitment struct {
	Enabled       bool   `json:"enabled"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	TrialDate     string `json:"trialDate"`
	ContactPhone  string `json:"contactPhone"`
	ContactWechat string `json:"contactWechat"`
	QRCode        string `json:"qrCode"`
}

// ClubHomeSocialLinks 社交媒体链接
type ClubHomeSocialLinks struct {
	Weibo       string `json:"weibo"`
	Wechat      string `json:"wechat"`
	Douyin      string `json:"douyin"`
	Xiaohongshu string `json:"xiaohongshu"`
	Website     string `json:"website"`
}

// ClubHomeNewsItem 手工置顶公告
type ClubHomeNewsItem struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	Link        string `json:"link"`
	Image       string `json:"image"`
	IsPinned    bool   `json:"isPinned"`
	PublishDate string `json:"publishDate"`
}

// DefaultModuleOrder 默认模块排序
var DefaultModuleOrder = []string{
	"hero", "about", "achievements", "teams", "coaches",
	"players", "facilities", "news", "activities", "recruitment", "contact",
}

// DefaultModuleVisibility 默认模块可见性
var DefaultModuleVisibility = map[string]bool{
	"hero":         true,
	"about":        true,
	"achievements": true,
	"teams":        true,
	"coaches":      true,
	"players":      false,
	"facilities":   false,
	"news":         true,
	"activities":   true,
	"recruitment":  false,
	"contact":      true,
}

// DefaultClubHome 创建默认配置
func DefaultClubHome(clubID uint) *ClubHome {
	return &ClubHome{
		ClubID: clubID,
		Hero: ClubHomeHero{
			Title:     "青少年足球俱乐部",
			Subtitle:  "专注青少年足球培养",
			ShowStats: true,
		},
		About: ClubHomeAbout{
			Enabled: true,
			Title:   "关于我们",
			Content: "俱乐部成立于2020年，致力于青少年足球培训...",
			Images:  []string{},
			Features: []ClubHomeFeature{
				{Icon: "users", Title: "专业教练", Description: "持证教练团队"},
				{Icon: "target", Title: "科学训练", Description: "系统化培养体系"},
				{Icon: "award", Title: "赛事平台", Description: "丰富比赛机会"},
				{Icon: "heart", Title: "健康成长", Description: "德智体全面发展"},
			},
		},
		Contact: ClubHomeContact{
			Enabled: true,
			Phone:   "",
			Wechat:  "",
			Address: "",
			Email:   "",
		},
		Facilities: ClubHomeFacilities{
			Enabled:     false,
			Title:       "训练环境",
			Description: "",
			Images:      []string{},
			Schedule:    []ClubHomeScheduleItem{},
		},
		Recruitment: ClubHomeRecruitment{
			Enabled:       false,
			Title:         "招生信息",
			Description:   "",
			TrialDate:     "",
			ContactPhone:  "",
			ContactWechat: "",
			QRCode:        "",
		},
		SocialLinks: ClubHomeSocialLinks{
			Weibo:       "",
			Wechat:      "",
			Douyin:      "",
			Xiaohongshu: "",
			Website:     "",
		},
		ModuleOrder:      "[\"hero\",\"about\",\"achievements\",\"teams\",\"coaches\",\"players\",\"facilities\",\"news\",\"activities\",\"recruitment\",\"contact\"]",
		ModuleVisibility: DefaultModuleVisibility,
	}
}

// Achievement 成就/荣誉
type Achievement struct {
	ID          uint      `json:"id"`
	ClubID      uint      `json:"club_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	Count       string    `json:"count"`
	Sort        int       `json:"sort"`
	CreatedAt   time.Time `json:"created_at"`
}

func (Achievement) TableName() string {
	return "achievements"
}

// ClubHomeTeam 主页展示的球队配置
type ClubHomeTeam struct {
	ID              uint `json:"id" gorm:"primaryKey"`
	ClubID          uint `json:"club_id" gorm:"index;not null"`
	TeamID          uint `json:"team_id" gorm:"index;not null"`
	Sort            int  `json:"sort"`
	ShowPlayerCount bool `json:"showPlayerCount"`
}

func (ClubHomeTeam) TableName() string {
	return "club_home_teams"
}

// ClubHomeCoach 主页展示的教练配置
type ClubHomeCoach struct {
	ID      uint `json:"id" gorm:"primaryKey"`
	ClubID  uint `json:"club_id" gorm:"index;not null"`
	CoachID uint `json:"coach_id" gorm:"index;not null"`
	Sort    int  `json:"sort"`
}

func (ClubHomeCoach) TableName() string {
	return "club_home_coaches"
}

// ClubHomePlayer 主页展示的球员配置
type ClubHomePlayer struct {
	ID            uint   `json:"id" gorm:"primaryKey"`
	ClubID        uint   `json:"club_id" gorm:"index;not null"`
	PlayerID      uint   `json:"player_id" gorm:"index;not null"`
	Sort          int    `json:"sort"`
	RecommendText string `json:"recommendText" gorm:"size:200"`
}

func (ClubHomePlayer) TableName() string {
	return "club_home_players"
}
