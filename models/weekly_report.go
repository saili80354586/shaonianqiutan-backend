package models

import (
	"fmt"
	"time"
)

// WeeklyReport 球员周报
type WeeklyReport struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	TeamID    uint       `gorm:"index;not null" json:"teamId"`
	PlayerID  uint       `gorm:"index;not null" json:"playerId"`
	CoachID   uint       `gorm:"not null" json:"coachId"`
	WeekStart time.Time  `gorm:"type:date;not null" json:"weekStart"` // 周起始日期(周一)
	WeekEnd   time.Time  `gorm:"type:date;not null" json:"weekEnd"`   // 周结束日期(周日)
	Deadline  *time.Time `gorm:"type:datetime" json:"deadline"`       // 填写截止时间

	// ==================== 球员自评 ====================

	// 训练出勤情况（新增）
	TrainingCount    int    `gorm:"default:0" json:"trainingCount"`    // 本周训练次数
	TrainingDuration int    `gorm:"default:0" json:"trainingDuration"` // 训练时长（分钟）
	AbsenceCount     int    `gorm:"default:0" json:"absenceCount"`     // 请假/缺勤次数
	AbsenceReason    string `gorm:"type:text" json:"absenceReason"`    // 请假原因

	// 训练内容反馈（扩展）
	KnowledgeSummary  string `gorm:"type:text" json:"knowledgeSummary"`  // 本周训练知识点总结 ⭐
	TechnicalContent  string `gorm:"type:text" json:"technicalContent"`  // 技术训练内容
	TacticalContent   string `gorm:"type:text" json:"tacticalContent"`   // 战术训练内容
	PhysicalCondition string `gorm:"type:text" json:"physicalCondition"` // 体能训练情况
	MatchPerformance  string `gorm:"type:text" json:"matchPerformance"`  // 比赛/对抗表现

	// 自我评价 - 多维度评分（新增）
	SelfAttitudeRating  int    `gorm:"default:0" json:"selfAttitudeRating"`  // 训练态度自评 1-5
	SelfTechniqueRating int    `gorm:"default:0" json:"selfTechniqueRating"` // 技术表现自评 1-5
	SelfTeamworkRating  int    `gorm:"default:0" json:"selfTeamworkRating"`  // 团队协作自评 1-5
	ImprovementsDetail  string `gorm:"type:text" json:"improvementsDetail"`  // 本周进步点
	Weaknesses          string `gorm:"type:text" json:"weaknesses"`          // 待改进方面

	// 身体状态反馈（新增）
	FatigueLevel  int    `gorm:"default:3" json:"fatigueLevel"`  // 疲劳程度 1-5
	Injuries      string `gorm:"type:text" json:"injuries"`      // 伤病情况
	SleepQuality  int    `gorm:"default:3" json:"sleepQuality"`  // 睡眠质量 1-5
	DietCondition string `gorm:"type:text" json:"dietCondition"` // 饮食情况

	// 其他信息（新增）
	MessageToCoach string `gorm:"type:text" json:"messageToCoach"` // 想对教练说的话
	Attachments    string `gorm:"type:text" json:"attachments"`    // 附件JSON数组

	// 提交状态
	SubmitStatus string     `gorm:"type:varchar(20);default:'draft'" json:"submitStatus"` // draft/submitted/overdue
	SubmittedAt  *time.Time `json:"submittedAt"`

	// ==================== 教练审核 ====================

	// 教练评价 - 多维度评分（新增）
	CoachAttitudeRating  int `gorm:"default:0" json:"coachAttitudeRating"`  // 训练态度评分 1-5
	CoachTechniqueRating int `gorm:"default:0" json:"coachTechniqueRating"` // 技术执行评分 1-5
	CoachTacticsRating   int `gorm:"default:0" json:"coachTacticsRating"`   // 战术理解评分 1-5
	CoachKnowledgeRating int `gorm:"default:0" json:"coachKnowledgeRating"` // 知识点掌握度 1-5

	// 教练评语（扩展）
	ReviewStatus            string     `gorm:"type:varchar(20);default:'pending'" json:"reviewStatus"` // pending/approved/rejected
	ReviewCoachID           uint       `json:"reviewCoachId"`
	ReviewComment           string     `gorm:"type:text" json:"reviewComment"`           // 整体表现评价
	StrengthsAcknowledgment string     `gorm:"type:text" json:"strengthsAcknowledgment"` // 优点肯定
	Suggestions             string     `gorm:"type:text" json:"suggestions"`             // 改进建议
	KnowledgeFeedback       string     `gorm:"type:text" json:"knowledgeFeedback"`       // 知识点理解偏差
	NextWeekFocus           string     `gorm:"type:text" json:"nextWeekFocus"`           // 下周训练重点
	RecommendAward          bool       `gorm:"default:false" json:"recommendAward"`      // 是否推荐表彰
	ReviewedAt              *time.Time `json:"reviewedAt"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// 关联
	Player *User `gorm:"foreignKey:PlayerID" json:"player,omitempty"`
	Coach  *User `gorm:"foreignKey:CoachID" json:"coach,omitempty"`
	Team   *Team `gorm:"foreignKey:TeamID" json:"team,omitempty"`
}

// TableName 表名
func (WeeklyReport) TableName() string {
	return "weekly_reports"
}

// WeeklyReportResponse 周报响应结构
type WeeklyReportResponse struct {
	ID         uint   `json:"id"`
	ReportID   uint   `json:"reportId"` // 兼容前端PlayerReportStatus接口
	TeamID     uint   `json:"teamId"`
	TeamName   string `json:"teamName"`
	PlayerID   uint   `json:"playerId"`
	PlayerName string `json:"playerName"`
	CoachID    uint   `json:"coachId"`
	CoachName  string `json:"coachName"`
	WeekStart  string `json:"weekStart"`
	WeekEnd    string `json:"weekEnd"`
	WeekLabel  string `json:"weekLabel"` // 如 "2026年第14周"
	Deadline   string `json:"deadline"`  // 填写截止时间

	// ==================== 球员自评 ====================

	// 训练出勤情况
	TrainingCount    int    `json:"trainingCount"`    // 本周训练次数
	TrainingDuration int    `json:"trainingDuration"` // 训练时长（分钟）
	AbsenceCount     int    `json:"absenceCount"`     // 请假/缺勤次数
	AbsenceReason    string `json:"absenceReason"`    // 请假原因

	// 训练内容反馈
	KnowledgeSummary  string `json:"knowledgeSummary"`  // 本周训练知识点总结
	TechnicalContent  string `json:"technicalContent"`  // 技术训练内容
	TacticalContent   string `json:"tacticalContent"`   // 战术训练内容
	PhysicalCondition string `json:"physicalCondition"` // 体能训练情况
	MatchPerformance  string `json:"matchPerformance"`  // 比赛/对抗表现

	// 自我评价 - 多维度评分
	SelfAttitudeRating  int    `json:"selfAttitudeRating"`  // 训练态度自评 1-5
	SelfTechniqueRating int    `json:"selfTechniqueRating"` // 技术表现自评 1-5
	SelfTeamworkRating  int    `json:"selfTeamworkRating"`  // 团队协作自评 1-5
	ImprovementsDetail  string `json:"improvementsDetail"`  // 本周进步点
	Weaknesses          string `json:"weaknesses"`          // 待改进方面

	// 身体状态反馈
	FatigueLevel  int    `json:"fatigueLevel"`  // 疲劳程度 1-5
	Injuries      string `json:"injuries"`      // 伤病情况
	SleepQuality  int    `json:"sleepQuality"`  // 睡眠质量 1-5
	DietCondition string `json:"dietCondition"` // 饮食情况

	// 其他信息
	MessageToCoach string   `json:"messageToCoach"` // 想对教练说的话
	Attachments    []string `json:"attachments"`    // 附件数组

	// 提交状态
	SubmitStatus string `json:"submitStatus"` // draft/submitted/overdue
	SubmittedAt  string `json:"submittedAt"`

	// ==================== 教练审核 ====================

	// 教练评价 - 多维度评分
	CoachAttitudeRating  int `json:"coachAttitudeRating"`  // 训练态度评分 1-5
	CoachTechniqueRating int `json:"coachTechniqueRating"` // 技术执行评分 1-5
	CoachTacticsRating   int `json:"coachTacticsRating"`   // 战术理解评分 1-5
	CoachKnowledgeRating int `json:"coachKnowledgeRating"` // 知识点掌握度 1-5

	// 教练评语
	ReviewStatus            string `json:"reviewStatus"`            // pending/approved/rejected
	ReviewComment           string `json:"reviewComment"`           // 整体表现评价
	StrengthsAcknowledgment string `json:"strengthsAcknowledgment"` // 优点肯定
	Suggestions             string `json:"suggestions"`             // 改进建议
	KnowledgeFeedback       string `json:"knowledgeFeedback"`       // 知识点理解偏差
	NextWeekFocus           string `json:"nextWeekFocus"`           // 下周训练重点
	RecommendAward          bool   `json:"recommendAward"`          // 是否推荐表彰
	ReviewedAt              string `json:"reviewedAt"`

	CreatedAt string `json:"createdAt"`
}

// ToResponse 转换为响应结构
func (w *WeeklyReport) ToResponse() WeeklyReportResponse {
	// 使用ISO 8601标准计算周数（周一是一周的第一天）
	year, week := w.WeekStart.ISOWeek()
	weekLabel := fmt.Sprintf("%d年第%d周", year, week)

	resp := WeeklyReportResponse{
		ID:           w.ID,
		ReportID:     w.ID, // 添加ReportID字段供前端使用
		TeamID:       w.TeamID,
		PlayerID:     w.PlayerID,
		CoachID:      w.CoachID,
		WeekStart:    w.WeekStart.Format("2006-01-02"),
		WeekEnd:      w.WeekEnd.Format("2006-01-02"),
		WeekLabel:    weekLabel,
		SubmitStatus: w.SubmitStatus,

		// 训练出勤情况
		TrainingCount:    w.TrainingCount,
		TrainingDuration: w.TrainingDuration,
		AbsenceCount:     w.AbsenceCount,
		AbsenceReason:    w.AbsenceReason,

		// 训练内容反馈
		KnowledgeSummary:  w.KnowledgeSummary,
		TechnicalContent:  w.TechnicalContent,
		TacticalContent:   w.TacticalContent,
		PhysicalCondition: w.PhysicalCondition,
		MatchPerformance:  w.MatchPerformance,

		// 自我评价 - 多维度评分
		SelfAttitudeRating:  w.SelfAttitudeRating,
		SelfTechniqueRating: w.SelfTechniqueRating,
		SelfTeamworkRating:  w.SelfTeamworkRating,
		ImprovementsDetail:  w.ImprovementsDetail,
		Weaknesses:          w.Weaknesses,

		// 身体状态反馈
		FatigueLevel:  w.FatigueLevel,
		Injuries:      w.Injuries,
		SleepQuality:  w.SleepQuality,
		DietCondition: w.DietCondition,

		// 其他信息
		MessageToCoach: w.MessageToCoach,

		// 教练评价 - 多维度评分
		CoachAttitudeRating:  w.CoachAttitudeRating,
		CoachTechniqueRating: w.CoachTechniqueRating,
		CoachTacticsRating:   w.CoachTacticsRating,
		CoachKnowledgeRating: w.CoachKnowledgeRating,

		// 教练评语
		ReviewStatus:            w.ReviewStatus,
		ReviewComment:           w.ReviewComment,
		StrengthsAcknowledgment: w.StrengthsAcknowledgment,
		Suggestions:             w.Suggestions,
		KnowledgeFeedback:       w.KnowledgeFeedback,
		NextWeekFocus:           w.NextWeekFocus,
		RecommendAward:          w.RecommendAward,

		CreatedAt: w.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	// 关联数据
	if w.Player != nil {
		resp.PlayerName = w.Player.Name
	}
	if w.Coach != nil {
		resp.CoachName = w.Coach.Name
	}
	if w.Team != nil {
		resp.TeamName = w.Team.Name
	}

	// 时间字段处理
	if w.Deadline != nil {
		resp.Deadline = w.Deadline.Format("2006-01-02 15:04:05")
	}
	if w.SubmittedAt != nil {
		resp.SubmittedAt = w.SubmittedAt.Format("2006-01-02 15:04:05")
	}
	if w.ReviewedAt != nil {
		resp.ReviewedAt = w.ReviewedAt.Format("2006-01-02 15:04:05")
	}

	// 附件解析
	if w.Attachments != "" {
		// 简单按逗号分割，实际可能需要JSON解析
		resp.Attachments = parseAttachments(w.Attachments)
	}

	return resp
}

// parseAttachments 解析附件字符串为数组
func parseAttachments(s string) []string {
	if s == "" {
		return []string{}
	}
	// 简单实现，实际可能需要JSON解析
	return []string{s}
}

// itoa 简单数字转字符串
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}

// WeeklyReportSubmit 球员提交周报请求
type WeeklyReportSubmit struct {
	// 基础信息
	TeamID    uint   `json:"teamId" binding:"required"`
	WeekStart string `json:"weekStart" binding:"required"` // 格式: 2006-01-02
	WeekEnd   string `json:"weekEnd"`                      // 格式: 2006-01-02，默认自动计算

	// ==================== 训练出勤情况 ====================
	TrainingCount    int    `json:"trainingCount" binding:"min=0"`    // 本周训练次数
	TrainingDuration int    `json:"trainingDuration" binding:"min=0"` // 训练时长（分钟）
	AbsenceCount     int    `json:"absenceCount" binding:"min=0"`     // 请假/缺勤次数
	AbsenceReason    string `json:"absenceReason"`                    // 请假原因

	// ==================== 训练内容反馈 ====================
	KnowledgeSummary  string `json:"knowledgeSummary" binding:"required,min=10,max=500"` // 本周训练知识点总结 ⭐
	TechnicalContent  string `json:"technicalContent" binding:"max=500"`                 // 技术训练内容
	TacticalContent   string `json:"tacticalContent" binding:"max=500"`                  // 战术训练内容
	PhysicalCondition string `json:"physicalCondition" binding:"max=300"`                // 体能训练情况
	MatchPerformance  string `json:"matchPerformance"`                                   // 比赛/对抗表现

	// ==================== 自我评价 - 多维度评分 ====================
	SelfAttitudeRating  int    `json:"selfAttitudeRating" binding:"required,min=1,max=5"`  // 训练态度自评 1-5
	SelfTechniqueRating int    `json:"selfTechniqueRating" binding:"required,min=1,max=5"` // 技术表现自评 1-5
	SelfTeamworkRating  int    `json:"selfTeamworkRating" binding:"required,min=1,max=5"`  // 团队协作自评 1-5
	ImprovementsDetail  string `json:"improvementsDetail" binding:"max=300"`               // 本周进步点
	Weaknesses          string `json:"weaknesses" binding:"max=300"`                       // 待改进方面

	// ==================== 身体状态反馈 ====================
	FatigueLevel  int    `json:"fatigueLevel" binding:"required,min=1,max=5"` // 疲劳程度 1-5
	Injuries      string `json:"injuries" binding:"max=200"`                  // 伤病情况
	SleepQuality  int    `json:"sleepQuality" binding:"required,min=1,max=5"` // 睡眠质量 1-5
	DietCondition string `json:"dietCondition" binding:"max=200"`             // 饮食情况

	// ==================== 其他信息 ====================
	MessageToCoach string   `json:"messageToCoach" binding:"max=300"` // 想对教练说的话
	Attachments    []string `json:"attachments"`                      // 附件URL数组
}

// WeeklyReportReview 教练审核请求
type WeeklyReportReview struct {
	// 审核状态
	Status string `json:"status" binding:"required,oneof=approved rejected"` // pending/approved/rejected

	// ==================== 多维度评分 ====================
	CoachAttitudeRating  int `json:"coachAttitudeRating" binding:"required,min=1,max=5"`  // 训练态度评分 1-5
	CoachTechniqueRating int `json:"coachTechniqueRating" binding:"required,min=1,max=5"` // 技术执行评分 1-5
	CoachTacticsRating   int `json:"coachTacticsRating" binding:"required,min=1,max=5"`   // 战术理解评分 1-5
	CoachKnowledgeRating int `json:"coachKnowledgeRating" binding:"required,min=1,max=5"` // 知识点掌握度 1-5

	// ==================== 教练评语 ====================
	ReviewComment           string `json:"reviewComment" binding:"required,min=10,max=500"`           // 整体表现评价
	StrengthsAcknowledgment string `json:"strengthsAcknowledgment" binding:"required,min=5,max=300"`  // 优点肯定
	Suggestions             string `json:"suggestions" binding:"required_if=Status rejected,max=500"` // 改进建议（退回时必填）
	KnowledgeFeedback       string `json:"knowledgeFeedback" binding:"max=300"`                       // 知识点理解偏差
	NextWeekFocus           string `json:"nextWeekFocus" binding:"max=300"`                           // 下周训练重点
	RecommendAward          bool   `json:"recommendAward"`                                            // 是否推荐表彰
}

// CreateWeeklyReportInput 教练发起周报的请求（为多个球员创建）
type CreateWeeklyReportInput struct {
	PlayerIDs []uint `json:"playerIds" binding:"required,min=1"` // 球员ID数组
	WeekStart string `json:"weekStart" binding:"required"`       // 周起始日期 格式: 2006-01-02
	WeekEnd   string `json:"weekEnd"`                            // 周结束日期 格式: 2006-01-02
}

// ==================== 周报周期管理 ====================

// WeeklyReportPeriod 周报周期（用于统计和查询）
type WeeklyReportPeriod struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	TeamID    uint       `gorm:"index;not null" json:"teamId"`
	WeekStart time.Time  `gorm:"type:date;not null" json:"weekStart"` // 周起始日期
	WeekEnd   time.Time  `gorm:"type:date;not null" json:"weekEnd"`   // 周结束日期
	Deadline  *time.Time `json:"deadline"`                            // 填写截止时间

	// 统计信息
	TotalPlayers   int `json:"totalPlayers"`   // 应填人数
	SubmittedCount int `json:"submittedCount"` // 已提交数
	PendingCount   int `json:"pendingCount"`   // 未提交数
	OverdueCount   int `json:"overdueCount"`   // 逾期数
	ReviewedCount  int `json:"reviewedCount"`  // 已审核数

	// 状态
	Status string `gorm:"type:varchar(20);default:'active'" json:"status"` // active/closed/archived

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName 表名
func (WeeklyReportPeriod) TableName() string {
	return "weekly_report_periods"
}

// WeeklyReportPeriodResponse 周报周期响应
type WeeklyReportPeriodResponse struct {
	ID        uint   `json:"id"`
	TeamID    uint   `json:"teamId"`
	TeamName  string `json:"teamName,omitempty"`
	WeekStart string `json:"weekStart"`
	WeekEnd   string `json:"weekEnd"`
	WeekLabel string `json:"weekLabel"`
	Deadline  string `json:"deadline"`

	// 统计信息
	TotalPlayers   int `json:"totalPlayers"`
	SubmittedCount int `json:"submittedCount"`
	PendingCount   int `json:"pendingCount"`
	OverdueCount   int `json:"overdueCount"`
	ReviewedCount  int `json:"reviewedCount"`

	// 计算字段
	SubmissionRate float64 `json:"submissionRate"` // 提交率（百分比）
	ReviewRate     float64 `json:"reviewRate"`     // 审核完成率

	Status        string                             `json:"status"`
	TrainingPlans []WeeklyPeriodTrainingPlanResponse `json:"trainingPlans,omitempty"`
}

type WeeklyPeriodTrainingPlanResponse struct {
	ID        uint   `json:"id"`
	Title     string `json:"title"`
	Theme     string `json:"theme,omitempty"`
	Location  string `json:"location,omitempty"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime,omitempty"`
	Status    string `json:"status"`
}

// ToResponse 转换为响应结构
func (p *WeeklyReportPeriod) ToResponse() WeeklyReportPeriodResponse {
	// 使用ISO 8601标准计算周数（周一是一周的第一天）
	year, week := p.WeekStart.ISOWeek()
	weekLabel := fmt.Sprintf("%d年第%d周", year, week)

	// 计算率
	submissionRate := 0.0
	reviewRate := 0.0
	if p.TotalPlayers > 0 {
		submissionRate = float64(p.SubmittedCount) / float64(p.TotalPlayers) * 100
		reviewRate = float64(p.ReviewedCount) / float64(p.TotalPlayers) * 100
	}

	resp := WeeklyReportPeriodResponse{
		ID:             p.ID,
		TeamID:         p.TeamID,
		WeekStart:      p.WeekStart.Format("2006-01-02"),
		WeekEnd:        p.WeekEnd.Format("2006-01-02"),
		WeekLabel:      weekLabel,
		TotalPlayers:   p.TotalPlayers,
		SubmittedCount: p.SubmittedCount,
		PendingCount:   p.PendingCount,
		OverdueCount:   p.OverdueCount,
		ReviewedCount:  p.ReviewedCount,
		SubmissionRate: submissionRate,
		ReviewRate:     reviewRate,
		Status:         p.Status,
	}

	if p.Deadline != nil {
		resp.Deadline = p.Deadline.Format("2006-01-02 15:04:05")
	}

	return resp
}
