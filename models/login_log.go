package models

import (
	"time"

	"gorm.io/gorm"
)

// LoginLog 登录日志模型
type LoginLog struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"index"`
	Phone     string         `json:"phone" gorm:"size:20;index"`
	Nickname  string         `json:"nickname" gorm:"size:64"`
	Role      string         `json:"role" gorm:"size:20"`
	IP        string         `json:"ip" gorm:"size:64"`
	Device    string         `json:"device" gorm:"size:200"`
	Browser   string         `json:"browser" gorm:"size:200"`
	OS        string         `json:"os" gorm:"size:100"`
	Location  string         `json:"location" gorm:"size:100"`
	Status    string         `json:"status" gorm:"size:20;default:'success'"` // success/failed
	FailReason string        `json:"fail_reason" gorm:"size:200"`
	CreatedAt time.Time      `json:"created_at"`
}

// TableName 表名
func (LoginLog) TableName() string {
	return "login_logs"
}

// LoginLogRepository 登录日志数据访问层
type LoginLogRepository struct {
	db *gorm.DB
}

func NewLoginLogRepository(db *gorm.DB) *LoginLogRepository {
	return &LoginLogRepository{db: db}
}

// Create 创建登录日志
func (r *LoginLogRepository) Create(log *LoginLog) error {
	return r.db.Create(log).Error
}

// FindAll 获取登录日志列表
func (r *LoginLogRepository) FindAll(page, pageSize int, userID uint, status string, startDate, endDate string) ([]LoginLog, int64, error) {
	var list []LoginLog
	var total int64

	query := r.db.Model(&LoginLog{}).Order("created_at DESC")
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate+" 23:59:59")
	}
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// GetStatistics 获取登录统计
func (r *LoginLogRepository) GetStatistics(days int) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 总登录次数
	var totalCount int64
	r.db.Model(&LoginLog{}).Count(&totalCount)
	result["total_count"] = totalCount

	// 今日登录次数
	today := time.Now().Format("2006-01-02")
	var todayCount int64
	r.db.Model(&LoginLog{}).Where("DATE(created_at) = ?", today).Count(&todayCount)
	result["today_count"] = todayCount

	// 失败次数
	var failCount int64
	r.db.Model(&LoginLog{}).Where("status = ?", "failed").Count(&failCount)
	result["fail_count"] = failCount

	// 独立用户
	var uniqueUsers int64
	r.db.Model(&LoginLog{}).Select("COUNT(DISTINCT user_id)").Scan(&uniqueUsers)
	result["unique_users"] = uniqueUsers

	// 最近7天趋势
	type dailyStat struct {
		Date  string `json:"date"`
		Count int64  `json:"count"`
	}
	var trend []dailyStat
	r.db.Model(&LoginLog{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at >= ?", time.Now().AddDate(0, 0, -days).Format("2006-01-02")).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&trend)
	result["trend"] = trend

	return result, nil
}
