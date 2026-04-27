package models

import (
	"time"

	"gorm.io/gorm"
)

// TeamStatus 球队状态
type TeamStatus string

const (
	TeamStatusActive   TeamStatus = "active"
	TeamStatusInactive TeamStatus = "inactive"
)

// Team 球队模型
type Team struct {
	ID              uint         `json:"id" gorm:"primaryKey"`
	ClubID         uint         `json:"club_id" gorm:"index;not null"`
	Name           string       `json:"name" gorm:"size:100;not null"`
	AgeGroup       string       `json:"age_group" gorm:"size:10;not null;index"` // U6/U7/U8/.../U18
	BirthYearStart *int         `json:"birth_year_start"` // 出生年份范围开始，如 2014
	BirthYearEnd   *int         `json:"birth_year_end"`   // 出生年份范围结束，如 2014
	Description    string       `json:"description" gorm:"type:text"`
	Status         TeamStatus   `json:"status" gorm:"size:20;default:'active'"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Club     *Club          `json:"club,omitempty" gorm:"foreignKey:ClubID"`
	Players  []TeamPlayer   `json:"players,omitempty" gorm:"foreignKey:TeamID"`
	Coaches  []TeamCoach   `json:"coaches,omitempty" gorm:"foreignKey:TeamID"`
	Invitations []TeamInvitation `json:"invitations,omitempty" gorm:"foreignKey:TeamID"`
}

// TableName 表名
func (Team) TableName() string {
	return "teams"
}

// GetPlayerCount 获取球员数量
func (t *Team) GetPlayerCount(db *gorm.DB) int64 {
	var count int64
	db.Model(&TeamPlayer{}).Where("team_id = ? AND status = ?", t.ID, "active").Count(&count)
	return count
}

// GetCoachCount 获取教练数量
func (t *Team) GetCoachCount(db *gorm.DB) int64 {
	var count int64
	db.Model(&TeamCoach{}).Where("team_id = ? AND status = ?", t.ID, "active").Count(&count)
	return count
}

// TeamPlayer 球队-球员关联模型
type TeamPlayer struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	TeamID      uint           `json:"team_id" gorm:"index;not null;uniqueIndex:idx_team_player"`
	UserID      uint           `json:"user_id" gorm:"index;not null;uniqueIndex:idx_team_player"`
	JerseyNumber string        `json:"jersey_number" gorm:"size:10"`
	Position    string        `json:"position" gorm:"size:50"` // 前锋/中场/后卫/门将
	Status      string        `json:"status" gorm:"size:20;default:'active';index"` // active/inactive/transferred
	Source      string        `json:"source" gorm:"size:20;default:'invited'"` // invited / applied / trial
	JoinedAt    time.Time     `json:"joined_at"`
	LeftAt      *time.Time   `json:"left_at"`
	Notes       string       `json:"notes" gorm:"type:text"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Team *Team `json:"team,omitempty" gorm:"foreignKey:TeamID"`
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName 表名
func (TeamPlayer) TableName() string {
	return "team_players"
}

// CoachRole 教练角色
type CoachRole string

const (
	CoachRoleHead          CoachRole = "head_coach"        // 主教练
	CoachRoleAssistant    CoachRole = "assistant"         // 助理教练
	CoachRoleGoalkeeper  CoachRole = "goalkeeper_coach" // 守门员教练
	CoachRoleFitness     CoachRole = "fitness_coach"    // 体能教练
	CoachRoleManager     CoachRole = "team_manager"       // 领队
)

// CoachRoleLabels 教练角色标签（中文）
var CoachRoleLabels = map[CoachRole]string{
	CoachRoleHead:         "主教练",
	CoachRoleAssistant:   "助理教练",
	CoachRoleGoalkeeper:  "守门员教练",
	CoachRoleFitness:     "体能教练",
	CoachRoleManager:     "领队",
}

// IsValidCoachRole 检查角色是否合法
func IsValidCoachRole(role string) bool {
	switch CoachRole(role) {
	case CoachRoleHead, CoachRoleAssistant, CoachRoleGoalkeeper, CoachRoleFitness, CoachRoleManager:
		return true
	}
	return false
}

// GetCoachRoleLabel 获取角色中文标签
func GetCoachRoleLabel(role CoachRole) string {
	if label, ok := CoachRoleLabels[role]; ok {
		return label
	}
	return string(role)
}

// ClubCoachStatus 俱乐部教练状态
type ClubCoachStatus string

const (
	ClubCoachStatusActive   ClubCoachStatus = "active"   // 在职
	ClubCoachStatusInactive ClubCoachStatus = "inactive" // 离职
	ClubCoachStatusPending  ClubCoachStatus = "pending"  // 待确认
)

// ClubCoach 俱乐部-教练关联模型（教练归属俱乐部）
type ClubCoach struct {
	ID           uint            `json:"id" gorm:"primaryKey"`
	ClubID       uint            `json:"club_id" gorm:"index;not null;uniqueIndex:idx_club_coach"`
	UserID       uint            `json:"user_id" gorm:"index;not null;uniqueIndex:idx_club_coach"`
	PrimaryRole  CoachRole       `json:"primary_role" gorm:"size:30;default:'head_coach'"` // 主角色
	Status       ClubCoachStatus `json:"status" gorm:"size:20;default:'active';index"`
	JoinedAt     time.Time       `json:"joined_at"`
	LeftAt       *time.Time      `json:"left_at"`
	Notes        string          `json:"notes" gorm:"type:text"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	DeletedAt    gorm.DeletedAt  `json:"-" gorm:"index"`

	// 关联
	Club *Club `json:"club,omitempty" gorm:"foreignKey:ClubID"`
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName 表名
func (ClubCoach) TableName() string {
	return "club_coaches"
}

// TeamCoach 球队-教练关联模型（教练分配到球队）
type TeamCoach struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	TeamID    uint           `json:"team_id" gorm:"index;not null;uniqueIndex:idx_team_coach_role"`
	UserID    uint           `json:"user_id" gorm:"index;not null;uniqueIndex:idx_team_coach_role"`
	Role      CoachRole     `json:"role" gorm:"size:30;default:'assistant';uniqueIndex:idx_team_coach_role"`
	Status    string        `json:"status" gorm:"size:20;default:'active';index"`
	InvitedBy *uint        `json:"invited_by" gorm:"index"` // 邀请人ID
	JoinedAt  time.Time     `json:"joined_at"`
	LeftAt    *time.Time   `json:"left_at"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Team *Team `json:"team,omitempty" gorm:"foreignKey:TeamID"`
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName 表名
func (TeamCoach) TableName() string {
	return "team_coaches"
}

// InvitationType 邀请类型
type InvitationType string

const (
	InvitationTypePlayer InvitationType = "player"
	InvitationTypeCoach InvitationType = "coach"
)

// InvitationStatus 邀请状态
type InvitationStatus string

const (
	InvitationStatusPending  InvitationStatus = "pending"
	InvitationStatusAccepted InvitationStatus = "accepted"
	InvitationStatusRejected InvitationStatus = "rejected"
	InvitationStatusExpired InvitationStatus = "expired"
)

// TeamInvitation 邀请记录模型
type TeamInvitation struct {
	ID           uint             `json:"id" gorm:"primaryKey"`
	TeamID      uint             `json:"team_id" gorm:"index;not null"`
	ClubID      uint             `json:"club_id" gorm:"index;not null"` // 俱乐部ID
	Type        InvitationType   `json:"type" gorm:"size:20;not null"` // player/coach
	InviteCode  string           `json:"invite_code" gorm:"size:100;uniqueIndex;not null"`
	TargetUserID *uint           `json:"target_user_id" gorm:"index"` // 直接邀请的用户ID
	TargetPhone  string           `json:"target_phone" gorm:"size:20"` // 预留：目标手机号
	Status      InvitationStatus `json:"status" gorm:"size:20;default:'pending';index"`
	CreatedBy   uint             `json:"created_by" gorm:"not null"`
	CreatedAt   time.Time       `json:"created_at"`
	ExpiresAt   time.Time       `json:"expires_at"`
	AcceptedAt  *time.Time      `json:"accepted_at"`
	RejectedAt  *time.Time      `json:"rejected_at"`
	RejectedReason string        `json:"rejected_reason" gorm:"type:text"`
	DeletedAt   gorm.DeletedAt   `json:"-" gorm:"index"`

	// 关联
	Team      *Team      `json:"team,omitempty" gorm:"foreignKey:TeamID"`
	Club     *Club      `json:"club,omitempty" gorm:"foreignKey:ClubID"`
	TargetUser *User      `json:"target_user,omitempty" gorm:"foreignKey:TargetUserID"`
	Creator   *User      `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
}

// TableName 表名
func (TeamInvitation) TableName() string {
	return "team_invitations"
}

// IsExpired 检查邀请是否过期
func (ti *TeamInvitation) IsExpired() bool {
	return time.Now().After(ti.ExpiresAt)
}

// CanAccept 检查邀请是否可以接受
func (ti *TeamInvitation) CanAccept() bool {
	return ti.Status == InvitationStatusPending && !ti.IsExpired()
}

// ClubInvitation 俱乐部邀请记录模型
// 用于俱乐部级别的教练邀请（不关联具体球队）
type ClubInvitation struct {
	ID             uint             `json:"id" gorm:"primaryKey"`
	ClubID         uint             `json:"club_id" gorm:"index;not null"`
	Type           InvitationType   `json:"type" gorm:"size:20;not null"` // coach
	InviteCode     string           `json:"invite_code" gorm:"size:100;uniqueIndex;not null"`
	TargetUserID   *uint            `json:"target_user_id" gorm:"index"` // 直接邀请的用户ID
	TargetPhone    string           `json:"target_phone" gorm:"size:20"` // 目标手机号
	TargetRole     CoachRole        `json:"target_role" gorm:"size:30;default:'assistant'"` // 预设角色
	Status         InvitationStatus `json:"status" gorm:"size:20;default:'pending';index"`
	CreatedBy      uint             `json:"created_by" gorm:"not null"`
	CreatedAt      time.Time        `json:"created_at"`
	ExpiresAt      time.Time        `json:"expires_at"`
	AcceptedAt     *time.Time       `json:"accepted_at"`
	RejectedAt     *time.Time       `json:"rejected_at"`
	RejectedReason string           `json:"rejected_reason" gorm:"type:text"`
	DeletedAt      gorm.DeletedAt   `json:"-" gorm:"index"`

	// 关联
	Club       *Club `json:"club,omitempty" gorm:"foreignKey:ClubID"`
	TargetUser *User `json:"target_user,omitempty" gorm:"foreignKey:TargetUserID"`
	Creator    *User `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
}

// TableName 表名
func (ClubInvitation) TableName() string {
	return "club_invitations"
}

// IsExpired 检查邀请是否过期
func (ci *ClubInvitation) IsExpired() bool {
	return time.Now().After(ci.ExpiresAt)
}

// CanAccept 检查邀请是否可以接受
func (ci *ClubInvitation) CanAccept() bool {
	return ci.Status == InvitationStatusPending && !ci.IsExpired()
}
