package models

import (
	"encoding/json"
	"time"
)

// MatchSummary 比赛总结
type MatchSummary struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TeamID    uint      `gorm:"index;not null" json:"teamId"`
	CoachID   uint      `gorm:"not null" json:"coachId"` // 创建教练

	// 比赛信息
	MatchName   string `gorm:"type:varchar(255)" json:"matchName"`
	MatchDate   string `gorm:"type:varchar(50)" json:"matchDate"` // 比赛日期字符串
	Opponent    string `gorm:"type:varchar(255)" json:"opponent"`
	Location    string `gorm:"type:varchar(50);default:'home'" json:"location"` // home/away/neutral

	// 赛制: 5人制/8人制/11人制
	MatchFormat string `gorm:"type:varchar(20);default:'11人制'" json:"matchFormat"`

	OurScore int    `gorm:"default:0" json:"ourScore"`
	OppScore int    `gorm:"default:0" json:"oppScore"`
	Result   string `gorm:"type:varchar(20);default:'pending'" json:"result"` // win/draw/lose/pending

	// 封面图
	CoverImage string `gorm:"type:varchar(500)" json:"coverImage"`

	// 视频链接(JSON数组): [{"platform":"baidu","url":"","code":"","name":"","note":""}]
	Videos string `gorm:"type:json" json:"videos"`

	// 参赛球员（JSON数组）- 用于精确控制球员自评权限
	PlayerIDs   string `gorm:"type:json" json:"playerIds"`
	PlayerCount int    `gorm:"default:0" json:"playerCount"`

	// 教练整体点评
	CoachOverall    string `gorm:"type:text" json:"coachOverall"`
	CoachTactic     string `gorm:"type:text" json:"coachTactic"`
	CoachKeyMoments string `gorm:"type:text" json:"coachKeyMoments"`

	// 状态: pending(待自评) -> player_submitted(待点评) -> completed(已完成)
	Status    string    `gorm:"type:varchar(30);default:'pending'" json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// 关联
	Team  *Team `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	Coach *User `gorm:"foreignKey:CoachID" json:"coach,omitempty"`
}

// TableName 表名
func (MatchSummary) TableName() string {
	return "match_summaries"
}

// GetPlayerIDs 获取关联的球员ID列表
func (m *MatchSummary) GetPlayerIDs() []uint {
	if m.PlayerIDs == "" || m.PlayerIDs == "null" {
		return []uint{}
	}
	var ids []uint
	json.Unmarshal([]byte(m.PlayerIDs), &ids)
	return ids
}

// SetPlayerIDs 设置球员ID列表
func (m *MatchSummary) SetPlayerIDs(ids []uint) {
	data, _ := json.Marshal(ids)
	m.PlayerIDs = string(data)
	m.PlayerCount = len(ids)
}

// GetVideos 获取视频链接列表
func (m *MatchSummary) GetVideos() []MatchVideoResponse {
	if m.Videos == "" || m.Videos == "null" {
		return []MatchVideoResponse{}
	}
	var videos []MatchVideoResponse
	json.Unmarshal([]byte(m.Videos), &videos)
	return videos
}

// CalcResult 根据比分计算比赛结果
func (m *MatchSummary) CalcResult() {
	if m.OurScore > m.OppScore {
		m.Result = "win"
	} else if m.OurScore < m.OppScore {
		m.Result = "lose"
	} else if m.OurScore == 0 && m.OppScore == 0 {
		m.Result = "pending" // 赛前创建，比分都为0
	} else {
		m.Result = "draw"
	}
}

// ============================================================
// 请求结构体
// ============================================================

// MatchSummaryCreate 创建比赛请求
type MatchSummaryCreate struct {
	TeamID        uint   `json:"teamId" binding:"required"`
	MatchName     string `json:"matchName" binding:"required"`
	MatchDate     string `json:"matchDate" binding:"required"`
	Opponent      string `json:"opponent" binding:"required"`
	Location      string `json:"location"`      // home/away/neutral，默认home
	MatchFormat   string `json:"matchFormat"`   // 5人制/8人制/11人制，默认11人制
	OurScore      int    `json:"ourScore"`
	OppScore      int    `json:"oppScore"`
	OpponentScore int    `json:"opponentScore"` // 兼容旧字段名
	Result        string `json:"result"`        // win/draw/lose/pending，为空时自动计算
	CoverImage    string `json:"coverImage"`    // 封面图URL
	PlayerIDs     []uint `json:"playerIds"`     // 参赛球员列表（可为空）
}

// MatchSummaryUpdate 更新比赛请求
type MatchSummaryUpdate struct {
	MatchName   string `json:"matchName"`
	MatchDate   string `json:"matchDate"`
	Opponent    string `json:"opponent"`
	Location    string `json:"location"`
	MatchFormat string `json:"matchFormat"`
	OurScore    *int   `json:"ourScore"`   // 指针区分0值和未传
	OppScore    *int   `json:"oppScore"`
	Result      string `json:"result"`
	CoverImage  string `json:"coverImage"`
	PlayerIDs   []uint `json:"playerIds"`
}

// CoachSummarySubmit 教练提交整体点评请求
type CoachSummarySubmit struct {
	CoachOverall    string              `json:"coachOverall" binding:"required"`
	CoachTactic     string              `json:"coachTactic"`
	CoachKeyMoments string              `json:"coachKeyMoments"`
	PlayerReviews   []CoachPlayerReview `json:"playerReviews"` // 教练对每个球员的评分+点评
}

// CoachPlayerReview 教练对单个球员的评分点评
type CoachPlayerReview struct {
	PlayerID     uint    `json:"playerId" binding:"required"`
	Rating       float64 `json:"rating"`    // 1-5
	CoachComment string  `json:"coachComment"`
	CoachReply   string  `json:"coachReply"` // 回复球员疑问
}

// CoverImageUpdate 封面图更新请求
type CoverImageUpdate struct {
	CoverImage string `json:"coverImage" binding:"required"`
}

// ============================================================
// 响应结构体
// ============================================================

// MatchSummaryResponse 比赛响应
type MatchSummaryResponse struct {
	ID          uint   `json:"id"`
	TeamID      uint   `json:"teamId"`
	TeamName    string `json:"teamName"`
	CoachID     uint   `json:"coachId"`
	CoachName   string `json:"coachName"`
	Status      string `json:"status"` // pending/player_submitted/completed

	// 比赛信息
	MatchName   string `json:"matchName"`
	MatchDate   string `json:"matchDate"`
	Opponent    string `json:"opponent"`
	Location    string `json:"location"`
	MatchFormat string `json:"matchFormat"`
	OurScore    int    `json:"ourScore"`
	OppScore    int    `json:"oppScore"`
	Result      string `json:"result"`
	CoverImage  string `json:"coverImage"`

	// 视频
	Videos []MatchVideoResponse `json:"videos"`

	// 参赛球员
	PlayerIDs   []uint               `json:"playerIds"`
	PlayerCount int                  `json:"playerCount"`
	Players     []PlayerInfoResponse `json:"players,omitempty"` // 球员详细信息

	// 教练点评（整体）
	CoachOverall    string `json:"coachOverall"`
	CoachTactic     string `json:"coachTactic"`
	CoachKeyMoments string `json:"coachKeyMoments"`

	// 球员自评列表
	PlayerReviews []PlayerReviewResponse `json:"playerReviews,omitempty"`

	// 统计
	SubmittedCount int `json:"submittedCount"` // 已提交自评数

	CreatedAt string `json:"createdAt"`
}

// PlayerInfoResponse 球员基本信息响应
type PlayerInfoResponse struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
	Number   int    `json:"number"`   // 球衣号码
	Position string `json:"position"` // 位置
}

// MatchSummaryListResponse 比赛列表项响应（精简版）
type MatchSummaryListResponse struct {
	ID          uint   `json:"id"`
	TeamID      uint   `json:"teamId"`
	TeamName    string `json:"teamName"`
	CoachID     uint   `json:"coachId"`
	CoachName   string `json:"coachName"`
	Status      string `json:"status"`
	MatchName   string `json:"matchName"`
	MatchDate   string `json:"matchDate"`
	Opponent    string `json:"opponent"`
	Location    string `json:"location"`
	MatchFormat string `json:"matchFormat"`
	OurScore    int    `json:"ourScore"`
	OppScore    int    `json:"oppScore"`
	Result      string `json:"result"`
	CoverImage  string `json:"coverImage"`
	PlayerCount int    `json:"playerCount"`

	SubmittedCount int `json:"submittedCount"`

	CreatedAt string `json:"createdAt"`
}

// ToListResponse 转换为列表项响应
func (m *MatchSummary) ToListResponse() MatchSummaryListResponse {
	resp := MatchSummaryListResponse{
		ID:          m.ID,
		TeamID:      m.TeamID,
		CoachID:     m.CoachID,
		Status:      m.Status,
		MatchName:   m.MatchName,
		MatchDate:   m.MatchDate,
		Opponent:    m.Opponent,
		Location:    m.Location,
		MatchFormat: m.MatchFormat,
		OurScore:    m.OurScore,
		OppScore:    m.OppScore,
		Result:      m.Result,
		CoverImage:  m.CoverImage,
		PlayerCount: m.PlayerCount,
		CreatedAt:   m.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if m.Team != nil {
		resp.TeamName = m.Team.Name
	}
	if m.Coach != nil {
		resp.CoachName = m.Coach.Name
	}

	return resp
}

// MatchStatsResponse 比赛统计响应
type MatchStatsResponse struct {
	TotalCount      int64 `json:"totalCount"`
	PendingCount    int64 `json:"pendingCount"`
	SubmittedCount  int64 `json:"submittedCount"`
	CompletedCount  int64 `json:"completedCount"`
	WinCount        int64 `json:"winCount"`
	DrawCount       int64 `json:"drawCount"`
	LoseCount       int64 `json:"loseCount"`
}
