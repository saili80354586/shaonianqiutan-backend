package models

import (
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// UserRole 用户角色
type UserRole string

const (
	RoleUser    UserRole = "user"
	RoleAnalyst UserRole = "analyst"
	RoleAdmin   UserRole = "admin"
	RoleClub    UserRole = "club"
	RoleCoach   UserRole = "coach"
	RoleScout   UserRole = "scout"
)

// UserStatus 用户状态
type UserStatus string

const (
	StatusActive   UserStatus = "active"
	StatusInactive UserStatus = "inactive"
	StatusBanned   UserStatus = "banned"
	StatusPending  UserStatus = "pending"
)

// User 用户模型
type User struct {
	ID       uint       `json:"id" gorm:"primaryKey"`
	Phone    string     `json:"phone" gorm:"uniqueIndex;size:20;not null"`
	Password string     `json:"-" gorm:"size:255;not null"`
	Nickname string     `json:"nickname" gorm:"size:50"`
	Avatar   string     `json:"avatar" gorm:"size:255"`
	Role     UserRole   `json:"role" gorm:"size:20;default:'user'"`
	Status   UserStatus `json:"status" gorm:"size:20;default:'active'"`

	// 球员资料信息
	Name           string  `json:"name" gorm:"size:50"`
	BirthDate      string  `json:"birth_date" gorm:"size:10"`
	Age            int     `json:"age"`
	Gender         string  `json:"gender" gorm:"size:10"`
	Height         float64 `json:"height"`
	Weight         float64 `json:"weight"`
	Foot           string  `json:"foot" gorm:"size:10"`
	Position       string  `json:"position" gorm:"size:50"`
	SecondPosition string  `json:"second_position" gorm:"size:50"`
	Province       string  `json:"province" gorm:"size:50"`
	City           string  `json:"city" gorm:"size:50"`
	Country        string  `json:"country" gorm:"size:50"`
	Club           string  `json:"club" gorm:"size:100"`
	StartYear      int     `json:"start_year"`
	FARegistered   bool    `json:"fa_registered" gorm:"default:false"`
	Association    string  `json:"association" gorm:"size:100"`
	JerseyColor    string  `json:"jersey_color" gorm:"size:20"`
	JerseyNumber   int     `json:"jersey_number"`

	// 家庭信息
	FatherHeight     float64 `json:"father_height"`
	FatherPhone      string  `json:"father_phone" gorm:"size:20"`
	FatherOccupation string  `json:"father_occupation" gorm:"size:100"`
	FatherEdu        string  `json:"father_edu" gorm:"size:50"`
	FatherJob        string  `json:"father_job" gorm:"size:100"`
	FatherAthlete    bool    `json:"father_athlete" gorm:"default:false"`
	MotherHeight     float64 `json:"mother_height"`
	MotherPhone      string  `json:"mother_phone" gorm:"size:20"`
	MotherOccupation string  `json:"mother_occupation" gorm:"size:100"`
	MotherEdu        string  `json:"mother_edu" gorm:"size:50"`
	MotherJob        string  `json:"mother_job" gorm:"size:100"`
	MotherAthlete    bool    `json:"mother_athlete" gorm:"default:false"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 当前激活角色（用于多角色切换状态同步）
	CurrentRole UserRole `json:"current_role" gorm:"size:20;default:''"`

	// 球员扩展资料
	CurrentTeam   string `json:"current_team" gorm:"size:100"`    // 当前球队/学校
	PlayingStyle  string `json:"playing_style" gorm:"type:text"`  // 踢球风格 JSON: ["tech","speed"]
	Wechat        string `json:"wechat" gorm:"size:50"`           // 微信号
	School        string `json:"school" gorm:"size:100"`          // 学校
	TechnicalTags string `json:"technical_tags" gorm:"type:text"` // 技术特点标签 JSON: ["盘带","射门"]
	MentalTags    string `json:"mental_tags" gorm:"type:text"`    // 心智性格标签 JSON: ["领导力","抗压"]
	Experiences   string `json:"experiences" gorm:"type:text"`    // 足球经历 JSON: [{period,team,position,achievement}]
	DominantFoot  string `json:"dominant_foot" gorm:"size:10"`    // 惯用脚：left/right/both
	VideoUrl      string `json:"video_url" gorm:"type:text"`      // 视频链接

	// 体测数据（简化存储在 User 表）
	Sprint30m        float64 `json:"sprint_30m"`         // 30米冲刺(秒)
	StandingLongJump float64 `json:"standing_long_jump"` // 立定跳远(cm)
	Flexibility      float64 `json:"flexibility"`        // 坐位体前屈(cm)
	PullUps          int     `json:"pull_ups"`           // 引体向上(个)
	PushUp           int     `json:"push_up"`            // 俯卧撑(个)
	SitUps           int     `json:"sit_ups"`            // 仰卧起坐(个/分钟)
	FiveMeterShuttle float64 `json:"five_meter_shuttle"` // 5×25米折返跑(秒)
	Coordination     float64 `json:"coordination"`       // 协调性测试(秒)
	SitAndReach      float64 `json:"sit_and_reach"`      // 坐位体前屈(cm)

	// 俱乐部扩展资料（注册时填写的球队/球员/教练数量、主要成绩）
	TeamCount    int    `json:"team_count" gorm:"default:0"`   // 球队数量
	PlayerCount  int    `json:"player_count" gorm:"default:0"` // 球员数量
	CoachCount   int    `json:"coach_count" gorm:"default:0"`  // 教练数量
	Achievements string `json:"achievements" gorm:"type:text"` // 主要成绩/荣誉

	// 前端多角色支持（登录时动态填充，不存储）
	Roles []UserRoleInfo `json:"roles,omitempty" gorm:"-"`

	// 通知设置（JSON字符串）
	NotificationSettings string `json:"notification_settings" gorm:"type:text;default:''"`
	// 隐私设置（JSON字符串）
	PrivacySettings string `json:"privacy_settings" gorm:"type:text;default:''"`
}

// UserRoleInfo 用户角色信息
type UserRoleInfo struct {
	Type   UserRole `json:"type"`
	Status string   `json:"status"`
}

// UserRepository 用户仓库
type UserRepository struct {
	db *gorm.DB
}

// AdminUserListFilters 管理员用户列表筛选条件
type AdminUserListFilters struct {
	Keyword string
	Role    string
	Status  string
	City    string
	AgeMin  *int
	AgeMax  *int
}

// AdminUserListItem 管理员用户列表项
type AdminUserListItem struct {
	ID           uint       `json:"id" gorm:"column:id"`
	Phone        string     `json:"phone" gorm:"column:phone"`
	Nickname     string     `json:"nickname" gorm:"column:nickname"`
	Name         string     `json:"name" gorm:"column:name"`
	DisplayName  string     `json:"display_name" gorm:"column:display_name"`
	Avatar       string     `json:"avatar" gorm:"column:avatar"`
	Role         UserRole   `json:"role" gorm:"column:role"`
	Status       UserStatus `json:"status" gorm:"column:status"`
	Province     string     `json:"province" gorm:"column:province"`
	City         string     `json:"city" gorm:"column:city"`
	Age          int        `json:"age" gorm:"column:age"`
	CurrentTeam  string     `json:"current_team" gorm:"column:current_team"`
	Club         string     `json:"club" gorm:"column:club"`
	JerseyColor  string     `json:"jersey_color" gorm:"column:jersey_color"`
	JerseyNumber int        `json:"jersey_number" gorm:"column:jersey_number"`
	Position     string     `json:"position" gorm:"column:position"`
	CreatedAt    time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"column:updated_at"`
	LastLoginAt  *string    `json:"last_login_at,omitempty" gorm:"column:last_login_at"`
}

// NewUserRepository 创建用户仓库
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByPhone 根据手机号查找用户
func (r *UserRepository) FindByPhone(phone string) (*User, error) {
	var user User
	if err := r.db.Where("phone = ?", phone).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByID 根据ID查找用户
func (r *UserRepository) FindByID(id uint) (*User, error) {
	var user User
	if err := r.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// Create 创建用户
func (r *UserRepository) Create(user *User) error {
	return r.db.Create(user).Error
}

// Update 更新用户信息
func (r *UserRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&User{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateStatus 更新用户状态
func (r *UserRepository) UpdateStatus(userID uint, status string) error {
	return r.db.Model(&User{}).Where("id = ?", userID).Update("status", status).Error
}

// UpdateAge 更新用户年龄
func (r *UserRepository) UpdateAge(userID uint, age int) error {
	return r.db.Model(&User{}).Where("id = ?", userID).Update("age", age).Error
}

// FindAdminUsers 获取管理员后台用户列表，支持筛选和最后登录时间
func (r *UserRepository) FindAdminUsers(page, pageSize int, filters AdminUserListFilters) ([]AdminUserListItem, int64, error) {
	var users []AdminUserListItem
	var total int64

	lastLoginSubQuery := r.db.
		Table("login_logs").
		Select("user_id, MAX(created_at) AS last_login_at").
		Group("user_id")

	query := r.db.
		Model(&User{}).
		Select(`
			users.id,
			users.phone,
			users.nickname,
			users.name,
			COALESCE(NULLIF(TRIM(users.nickname), ''), NULLIF(TRIM(users.name), ''), users.phone) AS display_name,
			users.avatar,
			users.role,
			users.status,
			users.province,
			users.city,
			users.age,
			users.current_team,
			users.club,
			users.jersey_color,
			users.jersey_number,
			users.position,
			users.created_at,
			users.updated_at,
			ll.last_login_at
		`).
		Joins("LEFT JOIN (?) AS ll ON ll.user_id = users.id", lastLoginSubQuery)

	if keyword := strings.TrimSpace(filters.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		if id, err := strconv.ParseUint(keyword, 10, 64); err == nil && id > 0 {
			query = query.Where("(users.nickname LIKE ? OR users.name LIKE ? OR users.phone LIKE ? OR users.id = ?)", like, like, like, id)
		} else {
			query = query.Where("(users.nickname LIKE ? OR users.name LIKE ? OR users.phone LIKE ?)", like, like, like)
		}
	}
	if role := strings.TrimSpace(filters.Role); role != "" {
		query = query.Where("users.role = ?", role)
	}
	if status := strings.TrimSpace(filters.Status); status != "" {
		query = query.Where("users.status = ?", status)
	}
	if city := strings.TrimSpace(filters.City); city != "" {
		like := "%" + city + "%"
		query = query.Where("users.city LIKE ?", like)
	}
	if filters.AgeMin != nil {
		query = query.Where("users.age >= ?", *filters.AgeMin)
	}
	if filters.AgeMax != nil {
		query = query.Where("users.age <= ?", *filters.AgeMax)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("users.created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Scan(&users).Error
	return users, total, err
}

// Count 统计总用户数
func (r *UserRepository) Count() (int64, error) {
	var count int64
	if err := r.db.Model(&User{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// CountByDate 按日期统计新增用户数
func (r *UserRepository) CountByDate(date string) (int64, error) {
	var count int64
	if err := r.db.Model(&User{}).Where("DATE(created_at) = ?", date).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// FindAll 获取所有用户列表
func (r *UserRepository) FindAll(page, pageSize int) ([]User, int64, error) {
	var users []User
	var total int64

	query := r.db.Model(&User{}).Order("created_at DESC")
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Find(&users).Error
	return users, total, err
}

// FindByUsername 根据用户名查找用户
func (r *UserRepository) FindByUsername(username string) (*User, error) {
	var user User
	// 注意: 当前User模型没有username字段,使用phone代替
	if err := r.db.Where("phone = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// Delete 删除用户
func (r *UserRepository) Delete(id uint) error {
	return r.db.Delete(&User{}, id).Error
}

// FindByRole 根据角色查找用户列表
func (r *UserRepository) FindByRole(role string, page, pageSize int, status string) ([]User, int64, error) {
	var users []User
	var total int64

	query := r.db.Model(&User{}).Where("role = ?", role)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	query = query.Order("created_at DESC")
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Find(&users).Error
	return users, total, err
}

// FindClubByUserID 根据用户ID查找俱乐部资料
func (r *UserRepository) FindClubByUserID(userID uint) (*Club, error) {
	var club Club
	err := r.db.Where("user_id = ?", userID).First(&club).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &club, err
}

// FindCoachByUserID 根据用户ID查找教练资料
func (r *UserRepository) FindCoachByUserID(userID uint) (*Coach, error) {
	var coach Coach
	err := r.db.Where("user_id = ?", userID).First(&coach).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &coach, err
}

// FindAnalystByUserID 根据用户ID查找分析师资料
func (r *UserRepository) FindAnalystByUserID(userID uint) (*Analyst, error) {
	var analyst Analyst
	err := r.db.Where("user_id = ?", userID).First(&analyst).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &analyst, err
}

// FindScoutByUserID 根据用户ID查找球探资料
func (r *UserRepository) FindScoutByUserID(userID uint) (*Scout, error) {
	var scout Scout
	err := r.db.Where("user_id = ?", userID).First(&scout).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &scout, err
}

// FindPlayerByUserID 根据用户ID查找球员资料
func (r *UserRepository) FindPlayerByUserID(userID uint) (*Player, error) {
	var player Player
	err := r.db.Where("user_id = ?", userID).First(&player).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &player, err
}

// CountByRole 按角色统计用户数
func (r *UserRepository) CountByRole(role string, count *int64) error {
	query := r.db.Model(&User{})
	if role != "" {
		query = query.Where("role = ?", role)
	}
	return query.Count(count).Error
}

// GetTopByRole 按角色获取Top用户（按创建时间）
func (r *UserRepository) GetTopByRole(role string, limit int) ([]User, error) {
	var users []User
	err := r.db.Where("role = ?", role).Order("created_at DESC").Limit(limit).Find(&users).Error
	return users, err
}

// CountActiveByDate 按日期统计活跃用户（有登录行为的用户，这里简化用创建时间）
func (r *UserRepository) CountActiveByDate(date string) (int64, error) {
	var count int64
	err := r.db.Model(&User{}).Where("DATE(created_at) = ?", date).Count(&count).Error
	return count, err
}
