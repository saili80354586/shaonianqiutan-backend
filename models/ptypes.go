package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// PhysicalTestStatus 体测活动状态
type PhysicalTestStatus string

const (
	PTStatusPending   PhysicalTestStatus = "pending"
	PTStatusOngoing   PhysicalTestStatus = "ongoing"
	PTStatusCompleted PhysicalTestStatus = "completed"
	PTStatusReported  PhysicalTestStatus = "report_generated"
)

// PhysicalTestTemplate 体测模板类型
type PhysicalTestTemplate string

const (
	PTTemplateBasic        PhysicalTestTemplate = "basic"
	PTTemplateAdvanced     PhysicalTestTemplate = "advanced"
	PTTemplateProfessional PhysicalTestTemplate = "professional"
	PTTemplateCustom       PhysicalTestTemplate = "custom"
)

// PhysicalTestTemplateCustom 自定义体测模板
type PhysicalTestTemplateCustom struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	ClubID      uint      `json:"club_id" gorm:"index;not null"`
	CreatedBy   uint      `json:"created_by" gorm:"index;not null"`
	Name        string    `json:"name" gorm:"size:100;not null"`
	Description string    `json:"description" gorm:"type:text"`
	Items       string    `json:"items" gorm:"type:text;not null"` // JSON数组，项目key列表
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 表名
func (PhysicalTestTemplateCustom) TableName() string {
	return "physical_test_template_customs"
}

// GetItems 获取自定义项目列表
func (p *PhysicalTestTemplateCustom) GetItems() []string {
	if p.Items == "" {
		return []string{}
	}
	var items []string
	json.Unmarshal([]byte(p.Items), &items)
	return items
}

// PhysicalTestActivity 体测活动模型
type PhysicalTestActivity struct {
	ID             uint                 `json:"id" gorm:"primaryKey"`
	ClubID         uint                 `json:"club_id" gorm:"index;not null"`
	Name           string               `json:"name" gorm:"size:100;not null"`
	Description    string               `json:"description" gorm:"type:text"`
	StartDate      time.Time            `json:"start_date" gorm:"not null"`
	EndDate        *time.Time           `json:"end_date"`
	Location       string               `json:"location" gorm:"size:200"`
	Template       PhysicalTestTemplate `json:"template" gorm:"size:20;default:'advanced'"`
	CustomItems    string               `json:"custom_items" gorm:"type:text"` // JSON数组，自定义项目ID列表
	PlayerIDs      string               `json:"player_ids" gorm:"type:text"`   // JSON数组，参与球员ID列表
	Status         PhysicalTestStatus   `json:"status" gorm:"size:20;default:'pending'"`
	NotifyParents  bool                 `json:"notify_parents" gorm:"default:true"`
	AutoSendReport bool                 `json:"auto_send_report" gorm:"default:true"`
	CreatedBy      uint                 `json:"created_by"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
	DeletedAt      gorm.DeletedAt       `json:"-" gorm:"index"`

	// 关联
	Club    *Club                `json:"club,omitempty" gorm:"foreignKey:ClubID"`
	Records []PhysicalTestRecord `json:"records,omitempty" gorm:"foreignKey:ActivityID"`
}

// TableName 表名
func (PhysicalTestActivity) TableName() string {
	return "physical_test_activities"
}

// GetCustomItems 获取自定义项目列表
func (p *PhysicalTestActivity) GetCustomItems() []string {
	if p.CustomItems == "" {
		return []string{}
	}
	var items []string
	json.Unmarshal([]byte(p.CustomItems), &items)
	return items
}

// GetPlayerIDs 获取参与球员ID列表
func (p *PhysicalTestActivity) GetPlayerIDs() []uint {
	if p.PlayerIDs == "" {
		return []uint{}
	}
	var ids []uint
	json.Unmarshal([]byte(p.PlayerIDs), &ids)
	return ids
}

// PhysicalTestRecord 体测记录模型
type PhysicalTestRecord struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	ActivityID uint      `json:"activity_id" gorm:"index;not null"`
	PlayerID   uint      `json:"player_id" gorm:"index;not null"`
	ClubID     uint      `json:"club_id" gorm:"index;not null"`
	TestDate   time.Time `json:"test_date" gorm:"not null"`

	// 基础指标
	Height *float64 `json:"height"` // 身高 cm
	Weight *float64 `json:"weight"` // 体重 kg
	BMI    *float64 `json:"bmi"`    // BMI (自动计算)

	// 速度类
	Sprint30m  *float64 `json:"sprint_30m"`  // 30米跑 秒
	Sprint50m  *float64 `json:"sprint_50m"`  // 50米跑 秒
	Sprint100m *float64 `json:"sprint_100m"` // 100米跑 秒

	// 灵敏类
	AgilityLadder *float64 `json:"agility_ladder"` // 敏捷梯 秒
	TTest         *float64 `json:"t_test"`         // T型跑 秒
	ShuttleRun    *float64 `json:"shuttle_run"`    // 折返跑 秒

	// 爆发类
	StandingLongJump *float64 `json:"standing_long_jump"` // 立定跳远 cm
	VerticalJump     *float64 `json:"vertical_jump"`      // 纵跳 cm

	// 柔韧类
	SitAndReach *float64 `json:"sit_and_reach"` // 坐位体前屈 cm

	// 力量类
	PushUp *int `json:"push_up"` // 俯卧撑 个
	SitUp  *int `json:"sit_up"`  // 仰卧起坐 个/分钟
	Plank  *int `json:"plank"`   // 平板支撑 秒

	// 其他数据 (JSON格式存储额外项目)
	ExtraData  string         `json:"extra_data" gorm:"type:text"`
	RecorderID uint           `json:"recorder_id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Activity *PhysicalTestActivity `json:"activity,omitempty" gorm:"foreignKey:ActivityID"`
	Player   *User                 `json:"player,omitempty" gorm:"foreignKey:PlayerID"`
}

// TableName 表名
func (PhysicalTestRecord) TableName() string {
	return "physical_test_records"
}

// GetExtraData 获取额外数据
func (p *PhysicalTestRecord) GetExtraData() map[string]interface{} {
	if p.ExtraData == "" {
		return map[string]interface{}{}
	}
	var data map[string]interface{}
	json.Unmarshal([]byte(p.ExtraData), &data)
	return data
}

// PhysicalTestReport 体测报告模型
type PhysicalTestReport struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	RecordID   uint           `json:"record_id" gorm:"index;not null"`
	PlayerID   uint           `json:"player_id" gorm:"index;not null"`
	ClubID     uint           `json:"club_id" gorm:"index;not null"`
	ActivityID uint           `json:"activity_id" gorm:"index;not null"`
	ReportData string         `json:"report_data" gorm:"type:text;not null"` // JSON格式的报告数据
	PDFURL     string         `json:"pdf_url" gorm:"size:500"`
	ShareToken string         `json:"share_token" gorm:"size:100;uniqueIndex"`
	ShareCount int            `json:"share_count" gorm:"default:0"`
	ViewCount  int            `json:"view_count" gorm:"default:0"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Record   *PhysicalTestRecord   `json:"record,omitempty" gorm:"foreignKey:RecordID"`
	Player   *User                 `json:"player,omitempty" gorm:"foreignKey:PlayerID"`
	Activity *PhysicalTestActivity `json:"activity,omitempty" gorm:"foreignKey:ActivityID"`
}

// TableName 表名
func (PhysicalTestReport) TableName() string {
	return "physical_test_reports"
}

// ReportData 体测报告数据结构
type PhysicalTestReportData struct {
	// 基础信息
	PlayerName     string `json:"player_name"`
	PlayerAge      int    `json:"player_age"`
	PlayerAgeGroup string `json:"player_age_group"`
	Position       string `json:"position"`
	TestDate       string `json:"test_date"`

	// 综合评级
	OverallRating string `json:"overall_rating"` // 优秀/良好/平均/需加强
	Percentile    int    `json:"percentile"`     // 百分位

	// 单项数据
	TestData map[string]TestItemData `json:"test_data"`

	// 成长趋势
	GrowthTrend map[string][]TrendPoint `json:"growth_trend"`

	// 百分位对比
	PercentileComparison map[string]PercentileInfo `json:"percentile_comparison"`

	// 优势与待提升
	Strengths    []string `json:"strengths"`
	Improvements []string `json:"improvements"`

	// 建议
	TrainingSuggestions  []string `json:"training_suggestions"`
	NutritionSuggestions []string `json:"nutrition_suggestions"`
	RestSuggestions      []string `json:"rest_suggestions"`

	// 下次建议
	NextTestSuggestion string `json:"next_test_suggestion"`
}

// TestItemData 单项体测数据
type TestItemData struct {
	Value      float64 `json:"value"`
	Unit       string  `json:"unit"`
	Percentile int     `json:"percentile"`
	Rating     string  `json:"rating"`      // 优秀/良好/平均/需加强
	Change     string  `json:"change"`      // 变化：+5cm / -0.3s
	ChangeType string  `json:"change_type"` // improved/declined/stable
}

// TrendPoint 趋势点
type TrendPoint struct {
	Date       string  `json:"date"`
	Value      float64 `json:"value"`
	Percentile int     `json:"percentile"`
}

// PercentileInfo 百分位信息
type PercentileInfo struct {
	Value int    `json:"value"` // 百分位值
	Level string `json:"level"` // top5/top10/top15/top25/top50等
}
