package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// ========== 评分维度定义 ==========

// RatingDimension 评分维度
type RatingDimension struct {
	Score   float64 `json:"score"`   // 分数 1-10
	Weight  float64 `json:"weight"`  // 权重百分比
	Comment string  `json:"comment"` // 评语
}

// VideoAnalysisScores 20项评分结构（对齐前端 VideoAnalysisWorkspace）
// 分为三大类：整体表现(4) + 进攻能力(8) + 防守能力(8)
type VideoAnalysisScores struct {
	// ===== 整体表现 (Overall) =====
	BallControl       RatingDimension `json:"ball_control"`       // 控球能力
	OffBallMovement   RatingDimension `json:"off_ball_movement"`  // 无球跑动
	PressingAwareness RatingDimension `json:"pressing_awareness"` // 逼抢意识
	Positioning       RatingDimension `json:"positioning"`        // 站位/选位

	// ===== 进攻能力 (Offense) =====
	WidthParticipation RatingDimension `json:"width_participation"` // 拉开宽度参与
	OffBallSupport     RatingDimension `json:"off_ball_support"`    // 无球支援
	OneVOne            RatingDimension `json:"one_v_one"`           // 1v1过人能力
	CrossingAssist     RatingDimension `json:"crossing_assist"`     // 传中/助攻
	CombatAbility      RatingDimension `json:"combat_ability"`      // 对抗能力
	PaceRhythm         RatingDimension `json:"pace_rhythm"`         // 节奏把控
	PassVision         RatingDimension `json:"pass_vision"`         // 传球视野
	BodyPosture        RatingDimension `json:"body_posture"`        // 身体姿态

	// ===== 防守能力 (Defense) =====
	DefensiveCommitment  RatingDimension `json:"defensive_commitment"`  // 防守投入度
	LossRecovery         RatingDimension `json:"loss_recovery"`         // 丢球回追
	TeammateCoordination RatingDimension `json:"teammate_coordination"` // 队友协防配合
	SecondBall           RatingDimension `json:"second_ball"`           // 二点球争抢
	AerialDuel           RatingDimension `json:"aerial_duel"`           // 空中争顶
	DefensiveShape       RatingDimension `json:"defensive_shape"`       // 防守阵型保持
	RoleAdjustment       RatingDimension `json:"role_adjustment"`       // 角色调整能力
	DefensiveRhythm      RatingDimension `json:"defensive_rhythm"`      // 防守节奏
}

// NewDefaultScores 创建默认评分（7分制）
func NewDefaultScores() *VideoAnalysisScores {
	return &VideoAnalysisScores{
		// 整体表现
		BallControl:       RatingDimension{Score: 7.0, Weight: 0.05, Comment: ""},
		OffBallMovement:   RatingDimension{Score: 7.0, Weight: 0.05, Comment: ""},
		PressingAwareness: RatingDimension{Score: 7.0, Weight: 0.05, Comment: ""},
		Positioning:       RatingDimension{Score: 7.0, Weight: 0.05, Comment: ""},
		// 进攻能力
		WidthParticipation: RatingDimension{Score: 7.0, Weight: 0.05, Comment: ""},
		OffBallSupport:     RatingDimension{Score: 7.0, Weight: 0.05, Comment: ""},
		OneVOne:            RatingDimension{Score: 7.0, Weight: 0.05, Comment: ""},
		CrossingAssist:     RatingDimension{Score: 7.0, Weight: 0.05, Comment: ""},
		CombatAbility:      RatingDimension{Score: 7.0, Weight: 0.05, Comment: ""},
		PaceRhythm:         RatingDimension{Score: 7.0, Weight: 0.05, Comment: ""},
		PassVision:         RatingDimension{Score: 7.0, Weight: 0.05, Comment: ""},
		BodyPosture:        RatingDimension{Score: 7.0, Weight: 0.05, Comment: ""},
		// 防守能力
		DefensiveCommitment:  RatingDimension{Score: 7.0, Weight: 0.05, Comment: ""},
		LossRecovery:         RatingDimension{Score: 7.0, Weight: 0.05, Comment: ""},
		TeammateCoordination: RatingDimension{Score: 7.0, Weight: 0.05, Comment: ""},
		SecondBall:           RatingDimension{Score: 7.0, Weight: 0.05, Comment: ""},
		AerialDuel:           RatingDimension{Score: 7.0, Weight: 0.05, Comment: ""},
		DefensiveShape:       RatingDimension{Score: 7.0, Weight: 0.05, Comment: ""},
		RoleAdjustment:       RatingDimension{Score: 7.0, Weight: 0.05, Comment: ""},
		DefensiveRhythm:      RatingDimension{Score: 7.0, Weight: 0.05, Comment: ""},
	}
}

// CalculateOverallScore 计算综合评分 (20项平均分，映射到1-100)
func (s *VideoAnalysisScores) CalculateOverallScore() float64 {
	if s == nil {
		return 0
	}
	total := s.BallControl.Score +
		s.OffBallMovement.Score +
		s.PressingAwareness.Score +
		s.Positioning.Score +
		s.WidthParticipation.Score +
		s.OffBallSupport.Score +
		s.OneVOne.Score +
		s.CrossingAssist.Score +
		s.CombatAbility.Score +
		s.PaceRhythm.Score +
		s.PassVision.Score +
		s.BodyPosture.Score +
		s.DefensiveCommitment.Score +
		s.LossRecovery.Score +
		s.TeammateCoordination.Score +
		s.SecondBall.Score +
		s.AerialDuel.Score +
		s.DefensiveShape.Score +
		s.RoleAdjustment.Score +
		s.DefensiveRhythm.Score
	avg := total / 20.0
	return avg * 10 // 10分制 → 100分制，保留1位小数
}

// MarshalJSON 自定义序列化：输出嵌套格式（与前端对齐）
func (s *VideoAnalysisScores) MarshalJSON() ([]byte, error) {
	nested := struct {
		Overall struct {
			BallControl       RatingDimension `json:"ball_control"`
			OffBallMovement   RatingDimension `json:"off_ball_movement"`
			PressingAwareness RatingDimension `json:"pressing_awareness"`
			Positioning       RatingDimension `json:"positioning"`
		} `json:"overall"`
		Offense struct {
			WidthParticipation RatingDimension `json:"width_participation"`
			OffBallSupport     RatingDimension `json:"off_ball_support"`
			OneVOne            RatingDimension `json:"one_v_one"`
			CrossingAssist     RatingDimension `json:"crossing_assist"`
			CombatAbility      RatingDimension `json:"combat_ability"`
			PaceRhythm         RatingDimension `json:"pace_rhythm"`
			PassVision         RatingDimension `json:"pass_vision"`
			BodyPosture        RatingDimension `json:"body_posture"`
		} `json:"offense"`
		Defense struct {
			DefensiveCommitment  RatingDimension `json:"defensive_commitment"`
			LossRecovery         RatingDimension `json:"loss_recovery"`
			TeammateCoordination RatingDimension `json:"teammate_coordination"`
			SecondBall           RatingDimension `json:"second_ball"`
			AerialDuel           RatingDimension `json:"aerial_duel"`
			DefensiveShape       RatingDimension `json:"defensive_shape"`
			RoleAdjustment       RatingDimension `json:"role_adjustment"`
			DefensiveRhythm      RatingDimension `json:"defensive_rhythm"`
		} `json:"defense"`
	}{
		Overall: struct {
			BallControl       RatingDimension `json:"ball_control"`
			OffBallMovement   RatingDimension `json:"off_ball_movement"`
			PressingAwareness RatingDimension `json:"pressing_awareness"`
			Positioning       RatingDimension `json:"positioning"`
		}{
			BallControl:       s.BallControl,
			OffBallMovement:   s.OffBallMovement,
			PressingAwareness: s.PressingAwareness,
			Positioning:       s.Positioning,
		},
		Offense: struct {
			WidthParticipation RatingDimension `json:"width_participation"`
			OffBallSupport     RatingDimension `json:"off_ball_support"`
			OneVOne            RatingDimension `json:"one_v_one"`
			CrossingAssist     RatingDimension `json:"crossing_assist"`
			CombatAbility      RatingDimension `json:"combat_ability"`
			PaceRhythm         RatingDimension `json:"pace_rhythm"`
			PassVision         RatingDimension `json:"pass_vision"`
			BodyPosture        RatingDimension `json:"body_posture"`
		}{
			WidthParticipation: s.WidthParticipation,
			OffBallSupport:     s.OffBallSupport,
			OneVOne:            s.OneVOne,
			CrossingAssist:     s.CrossingAssist,
			CombatAbility:      s.CombatAbility,
			PaceRhythm:         s.PaceRhythm,
			PassVision:         s.PassVision,
			BodyPosture:        s.BodyPosture,
		},
		Defense: struct {
			DefensiveCommitment  RatingDimension `json:"defensive_commitment"`
			LossRecovery         RatingDimension `json:"loss_recovery"`
			TeammateCoordination RatingDimension `json:"teammate_coordination"`
			SecondBall           RatingDimension `json:"second_ball"`
			AerialDuel           RatingDimension `json:"aerial_duel"`
			DefensiveShape       RatingDimension `json:"defensive_shape"`
			RoleAdjustment       RatingDimension `json:"role_adjustment"`
			DefensiveRhythm      RatingDimension `json:"defensive_rhythm"`
		}{
			DefensiveCommitment:  s.DefensiveCommitment,
			LossRecovery:         s.LossRecovery,
			TeammateCoordination: s.TeammateCoordination,
			SecondBall:           s.SecondBall,
			AerialDuel:           s.AerialDuel,
			DefensiveShape:       s.DefensiveShape,
			RoleAdjustment:       s.RoleAdjustment,
			DefensiveRhythm:      s.DefensiveRhythm,
		},
	}
	return json.Marshal(nested)
}

// UnmarshalJSON 自定义反序列化：支持嵌套格式（前端）和扁平格式（旧数据兼容）
func (s *VideoAnalysisScores) UnmarshalJSON(data []byte) error {
	// 先尝试嵌套格式
	var nested struct {
		Overall struct {
			BallControl       RatingDimension `json:"ball_control"`
			OffBallMovement   RatingDimension `json:"off_ball_movement"`
			PressingAwareness RatingDimension `json:"pressing_awareness"`
			Positioning       RatingDimension `json:"positioning"`
		} `json:"overall"`
		Offense struct {
			WidthParticipation RatingDimension `json:"width_participation"`
			OffBallSupport     RatingDimension `json:"off_ball_support"`
			OneVOne            RatingDimension `json:"one_v_one"`
			CrossingAssist     RatingDimension `json:"crossing_assist"`
			CombatAbility      RatingDimension `json:"combat_ability"`
			PaceRhythm         RatingDimension `json:"pace_rhythm"`
			PassVision         RatingDimension `json:"pass_vision"`
			BodyPosture        RatingDimension `json:"body_posture"`
		} `json:"offense"`
		Defense struct {
			DefensiveCommitment  RatingDimension `json:"defensive_commitment"`
			LossRecovery         RatingDimension `json:"loss_recovery"`
			TeammateCoordination RatingDimension `json:"teammate_coordination"`
			SecondBall           RatingDimension `json:"second_ball"`
			AerialDuel           RatingDimension `json:"aerial_duel"`
			DefensiveShape       RatingDimension `json:"defensive_shape"`
			RoleAdjustment       RatingDimension `json:"role_adjustment"`
			DefensiveRhythm      RatingDimension `json:"defensive_rhythm"`
		} `json:"defense"`
	}
	if err := json.Unmarshal(data, &nested); err == nil && (nested.Overall.BallControl.Score != 0 || nested.Offense.OneVOne.Score != 0 || nested.Defense.DefensiveCommitment.Score != 0) {
		s.BallControl = nested.Overall.BallControl
		s.OffBallMovement = nested.Overall.OffBallMovement
		s.PressingAwareness = nested.Overall.PressingAwareness
		s.Positioning = nested.Overall.Positioning
		s.WidthParticipation = nested.Offense.WidthParticipation
		s.OffBallSupport = nested.Offense.OffBallSupport
		s.OneVOne = nested.Offense.OneVOne
		s.CrossingAssist = nested.Offense.CrossingAssist
		s.CombatAbility = nested.Offense.CombatAbility
		s.PaceRhythm = nested.Offense.PaceRhythm
		s.PassVision = nested.Offense.PassVision
		s.BodyPosture = nested.Offense.BodyPosture
		s.DefensiveCommitment = nested.Defense.DefensiveCommitment
		s.LossRecovery = nested.Defense.LossRecovery
		s.TeammateCoordination = nested.Defense.TeammateCoordination
		s.SecondBall = nested.Defense.SecondBall
		s.AerialDuel = nested.Defense.AerialDuel
		s.DefensiveShape = nested.Defense.DefensiveShape
		s.RoleAdjustment = nested.Defense.RoleAdjustment
		s.DefensiveRhythm = nested.Defense.DefensiveRhythm
		return nil
	}
	// 回退到扁平格式（兼容旧数据）
	type alias VideoAnalysisScores
	var flat alias
	if err := json.Unmarshal(data, &flat); err != nil {
		return err
	}
	*s = VideoAnalysisScores(flat)
	return nil
}

// ToJSON 序列化为JSON字符串
func (s *VideoAnalysisScores) ToJSON() (string, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ParseScoresFromJSON 从JSON解析评分
func ParseScoresFromJSON(jsonStr string) (*VideoAnalysisScores, error) {
	var scores VideoAnalysisScores
	if jsonStr == "" {
		return NewDefaultScores(), nil
	}
	err := json.Unmarshal([]byte(jsonStr), &scores)
	if err != nil {
		return nil, err
	}
	return &scores, nil
}

// PotentialLevel 潜力等级
type PotentialLevel string

const (
	PotentialS PotentialLevel = "S" // 90-100 天才球员
	PotentialA PotentialLevel = "A" // 80-89 优秀球员
	PotentialB PotentialLevel = "B" // 70-79 良好球员
	PotentialC PotentialLevel = "C" // 60-69 一般球员
	PotentialD PotentialLevel = "D" // <60 需重点培养
)

// GetPotentialLevel 根据分数获取潜力等级
func GetPotentialLevel(score float64) PotentialLevel {
	switch {
	case score >= 90:
		return PotentialS
	case score >= 80:
		return PotentialA
	case score >= 70:
		return PotentialB
	case score >= 60:
		return PotentialC
	default:
		return PotentialD
	}
}

// ========== 高光时刻标签类型 ==========
type HighlightTagType string

const (
	HighlightGoal             HighlightTagType = "goal"              // 进球
	HighlightAssist           HighlightTagType = "assist"            // 助攻
	HighlightSteal            HighlightTagType = "steal"             // 抢断
	HighlightSave             HighlightTagType = "save"              // 扑救
	HighlightDribble          HighlightTagType = "dribble"           // 过人
	HighlightPass             HighlightTagType = "pass"              // 关键传球
	HighlightDefense          HighlightTagType = "defense"           // 防守关键
	HighlightPositioningError HighlightTagType = "positioning_error" // 站位问题
	HighlightDecisionError    HighlightTagType = "decision_error"    // 决策问题
	HighlightTurnover         HighlightTagType = "turnover"          // 失误
	HighlightRecoverySlow     HighlightTagType = "recovery_slow"     // 回防不及时
	HighlightTacticalNote     HighlightTagType = "tactical_note"     // 战术观察
	HighlightOffBallRun       HighlightTagType = "off_ball_run"      // 无球跑动
)

type HighlightMarkerType string

const (
	HighlightMarkerHighlight   HighlightMarkerType = "highlight"   // 精彩表现
	HighlightMarkerIssue       HighlightMarkerType = "issue"       // 待改进问题
	HighlightMarkerObservation HighlightMarkerType = "observation" // 战术观察
)

type HighlightMode string

const (
	HighlightModePoint HighlightMode = "point" // 单点
	HighlightModeRange HighlightMode = "range" // 时间段
)

type HighlightClipStatus string

const (
	HighlightClipNone       HighlightClipStatus = "none"       // 未生成
	HighlightClipQueued     HighlightClipStatus = "queued"     // 已排队
	HighlightClipProcessing HighlightClipStatus = "processing" // 处理中
	HighlightClipReady      HighlightClipStatus = "ready"      // 已生成
	HighlightClipFailed     HighlightClipStatus = "failed"     // 生成失败
)

// ========== VideoAnalysis 主分析表 ==========
type VideoAnalysisStatus string

const (
	AnalysisStatusScoring    VideoAnalysisStatus = "scoring"    // 评分中
	AnalysisStatusDraft      VideoAnalysisStatus = "draft"      // 草稿
	AnalysisStatusGenerating VideoAnalysisStatus = "generating" // AI生成中
	AnalysisStatusCompleted  VideoAnalysisStatus = "completed"  // 已完成
	AnalysisStatusSubmitted  VideoAnalysisStatus = "submitted"  // 已提交
)

// VideoAnalysis 视频分析模型
type VideoAnalysis struct {
	ID        uint `json:"id" gorm:"primaryKey"`
	OrderID   uint `json:"order_id" gorm:"uniqueIndex;not null"`
	AnalystID uint `json:"analyst_id" gorm:"index;not null"`
	UserID    uint `json:"user_id" gorm:"index;not null"` // 球员用户ID

	// 球员基本信息
	PlayerName     string  `json:"player_name" gorm:"size:50"`
	PlayerAge      int     `json:"player_age"`
	PlayerPosition string  `json:"player_position" gorm:"size:20"`
	PlayerFoot     string  `json:"player_foot" gorm:"size:10"` // 惯用脚
	PlayerHeight   float64 `json:"player_height"`
	PlayerWeight   float64 `json:"player_weight"`
	PlayerTeam     string  `json:"player_team" gorm:"size:100"`

	// 比赛信息
	MatchName     string `json:"match_name" gorm:"size:200"`
	MatchDate     string `json:"match_date" gorm:"size:10"`
	MatchType     string `json:"match_type" gorm:"size:50"`     // 正式比赛/友谊赛
	OpponentLevel string `json:"opponent_level" gorm:"size:20"` // 对手实力
	Opponent      string `json:"opponent" gorm:"size:100"`
	PlayTime      int    `json:"play_time"` // 出场时间(分钟)
	Goals         int    `json:"goals"`
	Assists       int    `json:"assists"`

	// 视频信息
	VideoURL string `json:"video_url" gorm:"size:500"`

	// ========== 核心评分数据 ==========
	// 综合评分（1-100）
	OverallScore float64 `json:"overall_score"`
	// 潜力等级
	PotentialLevel PotentialLevel `json:"potential_level" gorm:"size:1"`
	// 10项评分JSON
	Scores string `json:"scores" gorm:"type:text"`

	// ========== 摘要与建议 ==========
	Summary      string `json:"summary" gorm:"type:text"`       // 综合评价摘要
	Improvements string `json:"improvements" gorm:"type:text"`  // 重点改进建议
	AnalystNotes string `json:"analyst_notes" gorm:"type:text"` // 分析师补充说明

	// ========== AI生成报告 ==========
	AIReport        string `json:"ai_report" gorm:"type:longtext"`  // AI生成的完整报告
	AIReportStatus  string `json:"ai_report_status" gorm:"size:20"` // draft/regenerating/confirmed
	AIReportVersion int    `json:"ai_report_version"`               // 报告版本号（用于追踪修改）

	// ========== 状态与时间 ==========
	Status    VideoAnalysisStatus `json:"status" gorm:"size:20;default:'scoring'"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
	DeletedAt gorm.DeletedAt      `json:"-" gorm:"index"`

	// MD 文档路径（GORM 列名自动映射 snake_case）
	RatingReportMD string `json:"rating_report_md" gorm:"column:rating_report_md"` // 评分报告 MD 路径
	PlayerInfoMD   string `json:"player_info_md" gorm:"column:player_info_md"`     // 球员基础信息 MD 路径
}

// TableName 表名
func (VideoAnalysis) TableName() string {
	return "video_analyses"
}

// ========== AnalysisHighlight 高光时刻表 ==========
type AnalysisHighlight struct {
	ID              uint                `json:"id" gorm:"primaryKey"`
	AnalysisID      uint                `json:"analysis_id" gorm:"index;not null"`
	Timestamp       string              `json:"timestamp" gorm:"size:32"`                       // 展示时间 "12:30" 或 "12:30-12:45"
	MarkerType      HighlightMarkerType `json:"marker_type" gorm:"size:20;default:'highlight'"` // 标记类型
	Mode            HighlightMode       `json:"mode" gorm:"size:10;default:'point'"`            // 单点/时间段
	StartTimeMs     int                 `json:"start_time_ms" gorm:"default:0"`                 // 开始时间（毫秒）
	EndTimeMs       *int                `json:"end_time_ms"`                                    // 结束时间（毫秒）
	TagType         HighlightTagType    `json:"tag_type" gorm:"size:20"`                        // 标签类型
	Description     string              `json:"description" gorm:"type:text"`                   // 描述
	VideoClipURL    string              `json:"video_clip_url" gorm:"size:500"`                 // 视频片段URL
	ClipStatus      HighlightClipStatus `json:"clip_status" gorm:"size:20;default:'none'"`      // 剪辑状态
	ClipError       string              `json:"clip_error" gorm:"type:text"`                    // 剪辑失败原因
	ClipVersion     int                 `json:"clip_version" gorm:"default:0"`                  // 剪辑版本
	ClipGeneratedAt *time.Time          `json:"clip_generated_at"`                              // 剪辑生成时间
	IncludeInReport bool                `json:"include_in_report" gorm:"default:true"`          // 是否包含在报告中
	SortOrder       int                 `json:"sort_order" gorm:"default:0"`                    // 排序
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
}

// TableName 表名
func (AnalysisHighlight) TableName() string {
	return "analysis_highlights"
}

// ========== VideoAnalysisRepository ==========
type VideoAnalysisRepository struct {
	db *gorm.DB
}

func NewVideoAnalysisRepository(db *gorm.DB) *VideoAnalysisRepository {
	return &VideoAnalysisRepository{db: db}
}

// Create 创建分析记录
func (r *VideoAnalysisRepository) Create(analysis *VideoAnalysis) error {
	return r.db.Create(analysis).Error
}

// Update 更新分析记录
func (r *VideoAnalysisRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&VideoAnalysis{}).Where("id = ?", id).Updates(updates).Error
}

// FindByID 根据ID查询
func (r *VideoAnalysisRepository) FindByID(id uint) (*VideoAnalysis, error) {
	var analysis VideoAnalysis
	err := r.db.First(&analysis, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &analysis, nil
}

// FindByOrderID 根据订单ID查询
func (r *VideoAnalysisRepository) FindByOrderID(orderID uint) (*VideoAnalysis, error) {
	var analysis VideoAnalysis
	err := r.db.Where("order_id = ?", orderID).First(&analysis).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &analysis, nil
}

// FindByAnalystID 根据分析师ID查询分析列表
func (r *VideoAnalysisRepository) FindByAnalystID(analystID uint, page, pageSize int) ([]VideoAnalysis, int64, error) {
	var analyses []VideoAnalysis
	var total int64

	query := r.db.Model(&VideoAnalysis{}).Where("analyst_id = ?", analystID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&analyses).Error
	return analyses, total, err
}

// FindByUserID 根据用户ID查询分析列表（球员查看自己的报告）
func (r *VideoAnalysisRepository) FindByUserID(userID uint, page, pageSize int) ([]VideoAnalysis, int64, error) {
	var analyses []VideoAnalysis
	var total int64

	query := r.db.Model(&VideoAnalysis{}).Where("user_id = ?", userID).Where("status = ?", AnalysisStatusCompleted)
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&analyses).Error
	return analyses, total, err
}

// UpdateScores 更新评分
func (r *VideoAnalysisRepository) UpdateScores(id uint, scores *VideoAnalysisScores, overallScore float64, potentialLevel PotentialLevel) error {
	scoresJSON, err := scores.ToJSON()
	if err != nil {
		return err
	}
	return r.db.Model(&VideoAnalysis{}).Where("id = ?", id).Updates(map[string]interface{}{
		"scores":          scoresJSON,
		"overall_score":   overallScore,
		"potential_level": potentialLevel,
	}).Error
}

// ========== AnalysisHighlightRepository ==========
type AnalysisHighlightRepository struct {
	db *gorm.DB
}

func NewAnalysisHighlightRepository(db *gorm.DB) *AnalysisHighlightRepository {
	return &AnalysisHighlightRepository{db: db}
}

// Create 创建高光记录
func (r *AnalysisHighlightRepository) Create(highlight *AnalysisHighlight) error {
	if err := r.db.Create(highlight).Error; err != nil {
		return err
	}
	if !highlight.IncludeInReport {
		return r.db.Model(&AnalysisHighlight{}).Where("id = ?", highlight.ID).Update("include_in_report", false).Error
	}
	return nil
}

// BatchCreate 批量创建
func (r *AnalysisHighlightRepository) BatchCreate(highlights []AnalysisHighlight) error {
	if err := r.db.Create(&highlights).Error; err != nil {
		return err
	}
	for _, highlight := range highlights {
		if !highlight.IncludeInReport {
			if err := r.db.Model(&AnalysisHighlight{}).Where("id = ?", highlight.ID).Update("include_in_report", false).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// Delete 删除高光
func (r *AnalysisHighlightRepository) Delete(id uint) error {
	return r.db.Delete(&AnalysisHighlight{}, id).Error
}

// DeleteByAnalysisID 删除某分析的所有高光
func (r *AnalysisHighlightRepository) DeleteByAnalysisID(analysisID uint) error {
	return r.db.Where("analysis_id = ?", analysisID).Delete(&AnalysisHighlight{}).Error
}

// FindByID 根据ID查询高光
func (r *AnalysisHighlightRepository) FindByID(id uint) (*AnalysisHighlight, error) {
	var highlight AnalysisHighlight
	err := r.db.First(&highlight, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &highlight, nil
}

// FindByAnalysisID 查询某分析的所有高光
func (r *AnalysisHighlightRepository) FindByAnalysisID(analysisID uint) ([]AnalysisHighlight, error) {
	var highlights []AnalysisHighlight
	err := r.db.Where("analysis_id = ?", analysisID).Order("start_time_ms ASC, timestamp ASC").Find(&highlights).Error
	return highlights, err
}

// FindIncludedInReport 查询包含在报告中的高光
func (r *AnalysisHighlightRepository) FindIncludedInReport(analysisID uint) ([]AnalysisHighlight, error) {
	var highlights []AnalysisHighlight
	err := r.db.Where("analysis_id = ? AND include_in_report = ?", analysisID, true).Order("start_time_ms ASC, timestamp ASC").Find(&highlights).Error
	return highlights, err
}

// Update 更新高光
func (r *AnalysisHighlightRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&AnalysisHighlight{}).Where("id = ?", id).Updates(updates).Error
}
