package models

import (
	"encoding/json"
	"time"
)

// PlayerReview 球员自评表（独立表）
type PlayerReview struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	MatchID  uint   `gorm:"uniqueIndex:uk_match_player;index;not null" json:"matchId"`
	PlayerID uint   `gorm:"uniqueIndex:uk_match_player;index;not null" json:"playerId"`
	TeamID   uint   `gorm:"index;not null" json:"teamId"`

	// 基础数据
	Performance string `gorm:"type:varchar(50)" json:"performance"` // 优秀/良好/一般/需改进
	Goals       int    `gorm:"default:0" json:"goals"`
	Assists     int    `gorm:"default:0" json:"assists"`
	Saves       int    `gorm:"default:0" json:"saves"` // 扑救（门将）

	// 战术还原（JSON数组）
	Tactics string `gorm:"type:json" json:"tactics"` // TacticScenario数组

	// 文字描述
	Highlights   string `gorm:"type:text" json:"highlights"`
	Improvements string `gorm:"type:text" json:"improvements"`
	NextGoals    string `gorm:"type:text" json:"nextGoals"`

	// 教练对球员点评
	CoachRating  float64 `gorm:"type:decimal(2,1)" json:"coachRating"` // 1-5
	CoachComment string  `gorm:"type:text" json:"coachComment"`
	CoachReply   string  `gorm:"type:text" json:"coachReply"` // 教练回复球员疑问

	Status      string    `gorm:"type:varchar(20);default:'submitted'" json:"status"` // submitted/coach_reviewed
	SubmittedAt time.Time `json:"submittedAt"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	// 关联
	Match  *MatchSummary `gorm:"foreignKey:MatchID" json:"match,omitempty"`
	Player *User         `gorm:"foreignKey:PlayerID" json:"player,omitempty"`
}

// TableName 表名
func (PlayerReview) TableName() string {
	return "player_reviews"
}

// ============================================================
// 战术情景结构
// ============================================================

// TacticScenario 战术情景
type TacticScenario struct {
	Index       int        `json:"index"`       // 情景编号 1, 2, 3...
	Title       string     `json:"title"`       // "情景 1：第15分钟 - 进球瞬间"
	Description string     `json:"description"` // 情景描述
	Question    string     `json:"question"`    // 球员疑问
	ImageURL    string     `json:"imageUrl"`    // 生成的战术图URL
	Positions   []Position `json:"positions"`   // 位置数据
	Format      string     `json:"format"`      // 赛制
}

// Position 位置
type Position struct {
	Type   string  `json:"type"`   // "our" 我方 / "opp" 对方
	Number int     `json:"number"` // 球员编号
	X      float64 `json:"x"`      // X坐标 (0-100%)
	Y      float64 `json:"y"`      // Y坐标 (0-100%)
	Label  string  `json:"label"`  // 标签
}

// ============================================================
// 请求结构体
// ============================================================

// PlayerReviewSubmit 球员提交自评请求
type PlayerReviewSubmit struct {
	Performance  string           `json:"performance" binding:"required"`
	Goals        int              `json:"goals"`
	Assists      int              `json:"assists"`
	Saves        int              `json:"saves"`
	Tactics      []TacticScenario `json:"tactics"`
	Highlights   string           `json:"highlights"`
	Improvements string           `json:"improvements"`
	NextGoals    string           `json:"nextGoals"`
}

// CoachPlayerReviewSubmit 教练对单个球员提交评分点评请求
type CoachPlayerReviewSubmit struct {
	PlayerID     uint    `json:"playerId" binding:"required"`
	Rating       float64 `json:"rating"`     // 1-5
	CoachComment string  `json:"coachComment"`
	CoachReply   string  `json:"coachReply"` // 回复球员疑问
}

// ============================================================
// 响应结构体
// ============================================================

// PlayerReviewResponse 球员自评响应
type PlayerReviewResponse struct {
	ID           uint             `json:"id"`
	MatchID      uint             `json:"matchId"`
	PlayerID     uint             `json:"playerId"`
	PlayerName   string           `json:"playerName"`
	TeamID       uint             `json:"teamId"`
	Performance  string           `json:"performance"`
	Goals        int              `json:"goals"`
	Assists      int              `json:"assists"`
	Saves        int              `json:"saves"`
	Tactics      []TacticScenario `json:"tactics"`
	Highlights   string           `json:"highlights"`
	Improvements string           `json:"improvements"`
	NextGoals    string           `json:"nextGoals"`
	CoachRating  float64          `json:"coachRating"`
	CoachComment string           `json:"coachComment"`
	CoachReply   string           `json:"coachReply"`
	Status       string           `json:"status"`
	SubmittedAt  string           `json:"submittedAt"`
	CreatedAt    string           `json:"createdAt"`
}

// ToResponse 转换为响应结构
func (p *PlayerReview) ToResponse() PlayerReviewResponse {
	resp := PlayerReviewResponse{
		ID:           p.ID,
		MatchID:      p.MatchID,
		PlayerID:     p.PlayerID,
		TeamID:       p.TeamID,
		Performance:  p.Performance,
		Goals:        p.Goals,
		Assists:      p.Assists,
		Saves:        p.Saves,
		Highlights:   p.Highlights,
		Improvements: p.Improvements,
		NextGoals:    p.NextGoals,
		CoachRating:  p.CoachRating,
		CoachComment: p.CoachComment,
		CoachReply:   p.CoachReply,
		Status:       p.Status,
		CreatedAt:    p.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if p.Player != nil {
		resp.PlayerName = p.Player.Name
	}

	if !p.SubmittedAt.IsZero() {
		resp.SubmittedAt = p.SubmittedAt.Format("2006-01-02 15:04:05")
	}

	// 解析 Tactics JSON
	if p.Tactics != "" && p.Tactics != "null" {
		var tactics []TacticScenario
		if err := json.Unmarshal([]byte(p.Tactics), &tactics); err == nil {
			resp.Tactics = tactics
		}
	}

	return resp
}
