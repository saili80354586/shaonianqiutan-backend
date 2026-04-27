package models

import (
	"time"

	"gorm.io/gorm"
)

// MemberLevel 会员等级
type MemberLevel string

const (
	MemberLevelFree         MemberLevel = "free"
	MemberLevelBasic        MemberLevel = "basic"
	MemberLevelProfessional MemberLevel = "professional"
	MemberLevelEnterprise   MemberLevel = "enterprise"
)

// Club 俱乐部模型
type Club struct {
	ID               uint           `json:"id" gorm:"primaryKey"`
	UserID           uint           `json:"user_id" gorm:"index;not null"`
	Name             string         `json:"name" gorm:"size:100;not null"`
	Logo             string         `json:"logo" gorm:"size:500"`
	Description      string         `json:"description" gorm:"type:text"`
	Address          string         `json:"address" gorm:"size:200"`
	ContactName      string         `json:"contact_name" gorm:"size:50"`
	ContactPhone     string         `json:"contact_phone" gorm:"size:20"`
	EstablishedYear  int            `json:"established_year"`
	ClubSize         string         `json:"club_size" gorm:"size:20"` // small/medium/large
	MemberLevel      MemberLevel    `json:"member_level" gorm:"size:20;default:'free'"`
	MemberExpireDate time.Time      `json:"member_expire_date"`
	FreeTestQuota    int            `json:"free_physical_test_quota" gorm:"default:10"`
	Province         string         `json:"province" gorm:"size:50"`
	City             string         `json:"city" gorm:"size:50"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 表名
func (Club) TableName() string {
	return "clubs"
}

// ClubPlayer 俱乐部-球员关联模型
type ClubPlayer struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	ClubID    uint           `json:"club_id" gorm:"index;not null"`
	UserID    uint           `json:"user_id" gorm:"index;not null"`
	JoinDate  time.Time      `json:"join_date" gorm:"not null"`
	AgeGroup  string         `json:"age_group" gorm:"size:10;index"` // U8/U10/U12/U14/U16
	Position  string         `json:"position" gorm:"size:20;index"`
	Tags      string         `json:"tags" gorm:"type:text"`                        // JSON数组
	Status    string         `json:"status" gorm:"size:20;default:'active';index"` // active/inactive/left
	Notes     string         `json:"notes" gorm:"type:text"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Club *Club `json:"club,omitempty" gorm:"foreignKey:ClubID"`
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName 表名
func (ClubPlayer) TableName() string {
	return "club_players"
}

// ClubOrder 俱乐部订单模型
type ClubOrder struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	ClubID      uint           `json:"club_id" gorm:"index;not null"`
	UserID      uint           `json:"user_id" gorm:"index"` // 下单人
	OrderNo     string         `json:"order_no" gorm:"size:50;uniqueIndex"`
	PlayerID    uint           `json:"player_id" gorm:"index;not null"`
	AnalystID   uint           `json:"analyst_id" gorm:"index"`
	ServiceType string         `json:"service_type" gorm:"size:50;not null"` // quick_report/full_report/video_analysis
	Price       float64        `json:"price" gorm:"not null"`
	Discount    float64        `json:"discount" gorm:"default:1.0"` // 折扣率 0.95 = 95折
	FinalPrice  float64        `json:"final_price" gorm:"not null"`
	Status      string         `json:"status" gorm:"size:20;default:'pending'"` // pending/paid/processing/completed/cancelled
	Remark      string         `json:"remark" gorm:"type:text"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	PaidAt      *time.Time     `json:"paid_at"`
	CompletedAt *time.Time     `json:"completed_at"`

	// 关联
	Club    *Club    `json:"club,omitempty" gorm:"foreignKey:ClubID"`
	Player  *User    `json:"player,omitempty" gorm:"foreignKey:PlayerID"`
	Analyst *Analyst `json:"analyst,omitempty" gorm:"foreignKey:AnalystID"`
}

// TableName 表名
func (ClubOrder) TableName() string {
	return "club_orders"
}
