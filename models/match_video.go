package models

import (
	"time"
)

// MatchVideo 比赛视频链接表
type MatchVideo struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	MatchID    uint   `gorm:"index;not null" json:"matchId"`
	TeamID     uint   `gorm:"index;not null" json:"teamId"`
	UploaderID uint   `gorm:"index;not null" json:"uploaderId"` // 上传者ID

	// 视频信息
	Platform string `gorm:"type:varchar(50);not null" json:"platform"` // baidu/aliyun/weiyun/bilibili/douyin/other
	URL      string `gorm:"type:varchar(500);not null" json:"url"`
	Code     string `gorm:"type:varchar(20)" json:"code"`       // 提取码
	Name     string `gorm:"type:varchar(255)" json:"name"`      // 链接名称
	Note     string `gorm:"type:varchar(500)" json:"note"`      // 备注
	SortOrder int   `gorm:"default:0" json:"sortOrder"`         // 排序

	Status    string    `gorm:"type:varchar(20);default:'active'" json:"status"` // active/deleted
	CreatedAt time.Time `json:"createdAt"`

	// 关联
	Match *MatchSummary `gorm:"foreignKey:MatchID" json:"match,omitempty"`
}

// TableName 表名
func (MatchVideo) TableName() string {
	return "match_videos"
}

// ============================================================
// 视频平台常量
// ============================================================

// VideoPlatform 视频平台常量
var VideoPlatform = struct {
	Baidu    string
	Aliyun   string
	Weiyun   string
	Bilibili string
	Douyin   string
	Other    string
}{
	Baidu:    "baidu",
	Aliyun:   "aliyun",
	Weiyun:   "weiyun",
	Bilibili: "bilibili",
	Douyin:   "douyin",
	Other:    "other",
}

// ============================================================
// 请求结构体
// ============================================================

// MatchVideoCreate 创建视频链接请求
type MatchVideoCreate struct {
	Platform string `json:"platform" binding:"required"` // baidu/aliyun/weiyun/bilibili/douyin/other
	URL      string `json:"url" binding:"required"`      // 视频URL
	Code     string `json:"code"`                        // 提取码
	Name     string `json:"name" binding:"required"`     // 链接名称
	Note     string `json:"note"`                        // 备注
	SortOrder int   `json:"sortOrder"`                   // 排序
}

// MatchVideoUpdate 更新视频链接请求
type MatchVideoUpdate struct {
	Platform  string `json:"platform"`
	URL       string `json:"url"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	Note      string `json:"note"`
	SortOrder *int   `json:"sortOrder"` // 指针区分0值和未传
	Status    string `json:"status"`    // active/deleted
}

// ============================================================
// 响应结构体
// ============================================================

// MatchVideoResponse 视频链接响应
type MatchVideoResponse struct {
	ID         uint   `json:"id"`
	MatchID    uint   `json:"matchId"`
	TeamID     uint   `json:"teamId"`
	UploaderID uint   `json:"uploaderId"`
	Platform   string `json:"platform"`
	URL        string `json:"url"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	Note       string `json:"note"`
	SortOrder  int    `json:"sortOrder"`
	Status     string `json:"status"`
	CreatedAt  string `json:"createdAt"`
}

// ToResponse 转换为响应结构
func (v *MatchVideo) ToResponse() MatchVideoResponse {
	return MatchVideoResponse{
		ID:         v.ID,
		MatchID:    v.MatchID,
		TeamID:     v.TeamID,
		UploaderID: v.UploaderID,
		Platform:   v.Platform,
		URL:        v.URL,
		Code:       v.Code,
		Name:       v.Name,
		Note:       v.Note,
		SortOrder:  v.SortOrder,
		Status:     v.Status,
		CreatedAt:  v.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
