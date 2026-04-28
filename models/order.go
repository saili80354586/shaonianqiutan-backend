package models

import (
	"time"

	"gorm.io/gorm"
)

// OrderStatus 订单状态
type OrderStatus string

const (
	// OrderStatusPending 待支付
	OrderStatusPending OrderStatus = "pending"
	// OrderStatusPaid 已支付（待上传视频）
	OrderStatusPaid OrderStatus = "paid"
	// OrderStatusUploaded 已上传视频（待分配分析师）
	OrderStatusUploaded OrderStatus = "uploaded"
	// OrderStatusAssigned 已分配
	OrderStatusAssigned OrderStatus = "assigned"
	// OrderStatusProcessing 处理中
	OrderStatusProcessing OrderStatus = "processing"
	// OrderStatusCompleted 已完成
	OrderStatusCompleted OrderStatus = "completed"
	// OrderStatusCancelled 已取消
	OrderStatusCancelled OrderStatus = "cancelled"
	// OrderStatusRefunded 已退款
	OrderStatusRefunded OrderStatus = "refunded"
)

// PaymentMethod 支付方式
type PaymentMethod string

const (
	PaymentMethodWechat  PaymentMethod = "wechat"
	PaymentMethodAlipay  PaymentMethod = "alipay"
	PaymentMethodBalance PaymentMethod = "balance"
)

// Order 订单模型
type Order struct {
	ID            uint          `json:"id" gorm:"primaryKey"`
	UserID        uint          `json:"user_id" gorm:"not null;index"`
	User          User          `json:"user" gorm:"foreignKey:UserID"`
	AnalystID     *uint         `json:"analyst_id" gorm:"index;default:null"`
	Analyst       *Analyst      `json:"analyst" gorm:"foreignKey:AnalystID"`
	OrderNo       string        `json:"order_no" gorm:"uniqueIndex;not null;size:32"`
	Amount        float64       `json:"amount" gorm:"not null;type:decimal(10,2)"`
	Status        OrderStatus   `json:"status" gorm:"size:20;default:'pending'"`
	PaymentMethod PaymentMethod `json:"payment_method" gorm:"size:20"`
	PaymentTime   *time.Time    `json:"payment_time"`

	// 视频相关字段
	VideoURL      string `json:"video_url" gorm:"size:500"`
	VideoFilename string `json:"video_filename" gorm:"size:255"`

	// 报告关联
	ReportID *uint   `json:"report_id" gorm:"index"`
	Report   *Report `json:"report" gorm:"foreignKey:ReportID"`

	// 时间戳
	PaidAt      *time.Time `json:"paid_at"`
	CompletedAt *time.Time `json:"completed_at"`

	// 备注
	Remark string `json:"remark" gorm:"size:500"`

	// ===== 新增：订单业务字段 =====
	OrderType      string     `json:"order_type" gorm:"size:20;default:'basic'"` // basic | pro
	PlayerName     string     `json:"player_name" gorm:"size:50"`
	PlayerAge      int        `json:"player_age"`
	PlayerPosition string     `json:"player_position" gorm:"size:50"`
	JerseyColor    string     `json:"jersey_color" gorm:"size:50"`
	JerseyNumber   string     `json:"jersey_number" gorm:"size:10"`
	MatchName      string     `json:"match_name" gorm:"size:100"`
	Opponent       string     `json:"opponent" gorm:"size:100"`
	VideoDuration  int        `json:"video_duration"`                // 秒
	Deadline       *time.Time `json:"deadline"`                      // 分析师截止提交时间
	AssignedAt     *time.Time `json:"assigned_at"`                   // 管理员派发时间
	AcceptedAt     *time.Time `json:"accepted_at"`                   // 分析师接单时间
	CancelReason   string     `json:"cancel_reason" gorm:"size:500"` // 拒绝/取消原因

	// ===== 结算相关字段 =====
	SettledAt     *time.Time `json:"settled_at"`                                         // 结算时间
	SettledBy     uint       `json:"settled_by"`                                         // 结算人ID
	SettledAmount float64    `json:"settled_amount" gorm:"type:decimal(10,2);default:0"` // 实际结算金额

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// OrderRepository 订单仓库
type OrderRepository struct {
	db *gorm.DB
}

// NewOrderRepository 创建订单仓库
func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// GetDB 获取数据库连接（用于复杂查询）
func (r *OrderRepository) GetDB() *gorm.DB {
	return r.db
}

// Create 创建订单
func (r *OrderRepository) Create(order *Order) error {
	return r.db.Create(order).Error
}

// FindByID 根据ID查找订单
func (r *OrderRepository) FindByID(id uint) (*Order, error) {
	var order Order
	if err := r.db.Preload("User").Preload("Analyst").First(&order, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// FindByOrderNo 根据订单号查找订单
func (r *OrderRepository) FindByOrderNo(orderNo string) (*Order, error) {
	var order Order
	if err := r.db.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// FindByUserID 根据用户ID查找订单列表
func (r *OrderRepository) FindByUserID(userID uint, page, pageSize int, keyword string) ([]Order, int64, error) {
	var orders []Order
	var total int64

	query := r.db.Model(&Order{}).Where("user_id = ?", userID).Order("created_at DESC")

	if keyword != "" {
		query = query.Where("order_no LIKE ? OR player_name LIKE ? OR match_name LIKE ? OR opponent LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Find(&orders).Error
	return orders, total, err
}

// FindByAnalystID 根据分析师ID查找订单列表
func (r *OrderRepository) FindByAnalystID(analystID uint, page, pageSize int) ([]Order, int64, error) {
	var orders []Order
	var total int64

	query := r.db.Model(&Order{}).Where("analyst_id = ?", analystID).Order("created_at DESC")

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("User").Offset(offset).Limit(pageSize).Find(&orders).Error
	return orders, total, err
}

// FindAll 查找所有订单（管理员）
func (r *OrderRepository) FindAll(page, pageSize int, status string) ([]Order, int64, error) {
	var orders []Order
	var total int64

	query := r.db.Model(&Order{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("User").Preload("Analyst").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&orders).Error
	return orders, total, err
}

// Update 更新订单
func (r *OrderRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&Order{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateStatus 更新订单状态
func (r *OrderRepository) UpdateStatus(id uint, status OrderStatus) error {
	updates := map[string]interface{}{
		"status": status,
	}

	now := time.Now()

	switch status {
	case OrderStatusPaid:
		updates["paid_at"] = &now
	case OrderStatusCompleted:
		updates["completed_at"] = &now
	}

	return r.db.Model(&Order{}).Where("id = ?", id).Updates(updates).Error
}

// OrderStatistics 订单统计
type OrderStatistics struct {
	TotalCount   int64
	TotalRevenue float64
	TodayCount   int64
	TodayRevenue float64
}

// GetStatistics 获取订单统计
func (r *OrderRepository) GetStatistics() (*OrderStatistics, error) {
	stats := &OrderStatistics{}

	// 总订单数和总金额
	if err := r.db.Model(&Order{}).Select("COUNT(*), COALESCE(SUM(amount), 0)").Row().Scan(&stats.TotalCount, &stats.TotalRevenue); err != nil {
		return nil, err
	}

	// 今日订单数和金额
	today := time.Now().Format("2006-01-02")
	if err := r.db.Model(&Order{}).Where("DATE(created_at) = ?", today).Select("COUNT(*), COALESCE(SUM(amount), 0)").Row().Scan(&stats.TodayCount, &stats.TodayRevenue); err != nil {
		return nil, err
	}

	return stats, nil
}

// GetStatisticsByDate 按日期获取订单统计
func (r *OrderRepository) GetStatisticsByDate(date string) (int64, float64, error) {
	var count int64
	var revenue float64

	err := r.db.Model(&Order{}).Where("DATE(created_at) = ?", date).Select("COUNT(*), COALESCE(SUM(amount), 0)").Row().Scan(&count, &revenue)

	return count, revenue, err
}

// CancelOrder 取消订单
func (r *OrderRepository) CancelOrder(id uint) error {
	return r.db.Model(&Order{}).Where("id = ?", id).Update("status", OrderStatusCancelled).Error
}

// GetUserOrderStatistics 获取用户订单统计
type UserOrderStatistics struct {
	TotalOrders      int64   `json:"total_orders"`
	PendingOrders    int64   `json:"pending_orders"`
	ProcessingOrders int64   `json:"processing_orders"`
	CompletedOrders  int64   `json:"completed_orders"`
	CancelledOrders  int64   `json:"cancelled_orders"`
	TotalSpent       float64 `json:"total_spent"`
}

func (r *OrderRepository) GetUserOrderStatistics(userID uint) (*UserOrderStatistics, error) {
	stats := &UserOrderStatistics{}

	// 总订单数
	if err := r.db.Model(&Order{}).Where("user_id = ?", userID).Count(&stats.TotalOrders).Error; err != nil {
		return nil, err
	}

	// 各状态订单数
	r.db.Model(&Order{}).Where("user_id = ? AND status = ?", userID, OrderStatusPending).Count(&stats.PendingOrders)
	r.db.Model(&Order{}).Where("user_id = ? AND status = ?", userID, OrderStatusProcessing).Count(&stats.ProcessingOrders)
	r.db.Model(&Order{}).Where("user_id = ? AND status = ?", userID, OrderStatusCompleted).Count(&stats.CompletedOrders)
	r.db.Model(&Order{}).Where("user_id = ? AND status = ?", userID, OrderStatusCancelled).Count(&stats.CancelledOrders)

	// 总消费金额
	if err := r.db.Model(&Order{}).Where("user_id = ? AND status = ?", userID, OrderStatusCompleted).
		Select("COALESCE(SUM(amount), 0)").Scan(&stats.TotalSpent).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

// CanTransition 检查订单状态是否可以流转
func CanTransition(currentStatus, newStatus OrderStatus) bool {
	// 定义允许的状态流转
	transitions := map[OrderStatus][]OrderStatus{
		OrderStatusPending:    {OrderStatusPaid, OrderStatusCancelled},
		OrderStatusPaid:       {OrderStatusUploaded, OrderStatusCancelled, OrderStatusRefunded},
		OrderStatusUploaded:   {OrderStatusAssigned, OrderStatusCancelled, OrderStatusRefunded},
		OrderStatusAssigned:   {OrderStatusProcessing, OrderStatusCancelled},
		OrderStatusProcessing: {OrderStatusCompleted},
		OrderStatusCompleted:  {OrderStatusRefunded}, // 仅管理员可以退款
		OrderStatusCancelled:  {},
		OrderStatusRefunded:   {},
	}

	allowedStatuses, exists := transitions[currentStatus]
	if !exists {
		return false
	}

	for _, allowed := range allowedStatuses {
		if allowed == newStatus {
			return true
		}
	}

	return false
}

// FindPendingByAnalystID 获取分析师待处理订单（status = assigned）
func (r *OrderRepository) FindPendingByAnalystID(analystID uint) ([]Order, error) {
	var orders []Order
	err := r.db.Where("analyst_id = ? AND status = ?", analystID, OrderStatusAssigned).
		Order("assigned_at ASC").
		Preload("User").
		Find(&orders).Error
	return orders, err
}

// FindActiveByAnalystID 获取分析师进行中订单（status = processing）
func (r *OrderRepository) FindActiveByAnalystID(analystID uint) ([]Order, error) {
	var orders []Order
	err := r.db.Where("analyst_id = ? AND status = ?", analystID, OrderStatusProcessing).
		Order("deadline ASC").
		Preload("User").
		Find(&orders).Error
	return orders, err
}

// FindHistoryByAnalystID 获取分析师历史订单（completed/cancelled）支持筛选和分页
func (r *OrderRepository) FindHistoryByAnalystID(analystID uint, status string, orderType string, startDate, endDate, keyword string, page, pageSize int) ([]Order, int64, error) {
	var orders []Order
	var total int64

	query := r.db.Model(&Order{}).Where("analyst_id = ?", analystID).
		Where("status IN ?", []OrderStatus{OrderStatusCompleted, OrderStatusCancelled})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if orderType != "" {
		query = query.Where("order_type = ?", orderType)
	}
	if startDate != "" && endDate != "" {
		query = query.Where("DATE(updated_at) BETWEEN ? AND ?", startDate, endDate)
	}
	if keyword != "" {
		query = query.Where("order_no LIKE ? OR player_name LIKE ? OR match_name LIKE ? OR opponent LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("updated_at DESC").Preload("Report").Offset(offset).Limit(pageSize).Find(&orders).Error
	return orders, total, err
}

// GetAnalystDashboardStats 获取分析师工作台统计数据
func (r *OrderRepository) GetAnalystDashboardStats(analystID uint) (map[string]interface{}, error) {
	var pendingCount, activeCount, todayDeadlineCount, totalCompleted, totalAssigned int64
	var monthlyIncome, todayIncome, weekIncome float64

	now := time.Now()
	today := now.Format("2006-01-02")
	weekStart := now.AddDate(0, 0, -int(now.Weekday())).Format("2006-01-02")
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// 注意：每个查询必须用独立的 Session，否则 GORM 会累加 Where 条件！
	r.db.Model(&Order{}).Where("analyst_id = ? AND status = ?", analystID, OrderStatusAssigned).Count(&pendingCount)
	r.db.Model(&Order{}).Where("analyst_id = ? AND status = ?", analystID, OrderStatusProcessing).Count(&activeCount)
	r.db.Model(&Order{}).Where("analyst_id = ? AND status = ? AND DATE(deadline) = ?", analystID, OrderStatusProcessing, today).Count(&todayDeadlineCount)
	r.db.Model(&Order{}).Where("analyst_id = ? AND status = ?", analystID, OrderStatusCompleted).Count(&totalCompleted)
	r.db.Model(&Order{}).Where("analyst_id = ? AND status IN ?", analystID, []string{string(OrderStatusAssigned), string(OrderStatusProcessing), string(OrderStatusCompleted)}).Count(&totalAssigned)

	var monthlyRevenue float64
	r.db.Model(&Order{}).Where("analyst_id = ? AND status = ? AND completed_at >= ?", analystID, OrderStatusCompleted, monthStart).
		Select("COALESCE(SUM(amount), 0)").Scan(&monthlyRevenue)
	monthlyIncome = monthlyRevenue

	var todayRevenue float64
	r.db.Model(&Order{}).Where("analyst_id = ? AND status = ? AND DATE(completed_at) = ?", analystID, OrderStatusCompleted, today).
		Select("COALESCE(SUM(amount), 0)").Scan(&todayRevenue)
	todayIncome = todayRevenue

	var weekRevenue float64
	r.db.Model(&Order{}).Where("analyst_id = ? AND status = ? AND DATE(completed_at) >= ?", analystID, OrderStatusCompleted, weekStart).
		Select("COALESCE(SUM(amount), 0)").Scan(&weekRevenue)
	weekIncome = weekRevenue

	// 完成率 = 已完成 / 总接单数（已分配+进行中+已完成）
	completionRate := 0.0
	if totalAssigned > 0 {
		completionRate = float64(totalCompleted) / float64(totalAssigned) * 100
	}

	return map[string]interface{}{
		"pendingCount":       pendingCount,
		"activeCount":        activeCount,
		"todayDeadlineCount": todayDeadlineCount,
		"monthlyIncome":      monthlyIncome,
		"totalCompleted":     totalCompleted,
		"todayIncome":        todayIncome,
		"weekIncome":         weekIncome,
		"completionRate":     completionRate,
	}, nil
}

// GetAnalystIncomeDetails 获取分析师收益明细
func (r *OrderRepository) GetAnalystIncomeDetails(analystID uint, startDate, endDate string, page, pageSize int) ([]Order, int64, error) {
	var orders []Order
	var total int64

	query := r.db.Model(&Order{}).Where("analyst_id = ? AND status = ?", analystID, OrderStatusCompleted)
	if startDate != "" && endDate != "" {
		query = query.Where("DATE(completed_at) BETWEEN ? AND ?", startDate, endDate)
	}

	query.Count(&total)
	offset := (page - 1) * pageSize
	err := query.Order("completed_at DESC").Offset(offset).Limit(pageSize).Find(&orders).Error
	return orders, total, err
}

// GetAnalystIncomeTrend 获取分析师收益趋势数据
func (r *OrderRepository) GetAnalystIncomeTrend(analystID uint, startDate, endDate string) ([]map[string]interface{}, error) {
	type TrendRow struct {
		Date   string  `json:"date"`
		Income float64 `json:"income"`
		Orders int64   `json:"orders"`
	}
	var rows []TrendRow

	err := r.db.Model(&Order{}).
		Select("DATE(completed_at) as date, COALESCE(SUM(amount), 0) as income, COUNT(*) as orders").
		Where("analyst_id = ? AND status = ?", analystID, OrderStatusCompleted).
		Where("DATE(completed_at) BETWEEN ? AND ?", startDate, endDate).
		Group("DATE(completed_at)").
		Order("date ASC").
		Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(rows))
	for i, row := range rows {
		result[i] = map[string]interface{}{
			"date":   row.Date[5:], // MM-DD
			"income": row.Income,
			"orders": row.Orders,
		}
	}
	return result, nil
}

// GetPendingSettlementByAnalystID 获取分析师待结算金额（processing状态订单）
func (r *OrderRepository) GetPendingSettlementByAnalystID(analystID uint) (float64, int64, error) {
	var amount float64
	var count int64
	err := r.db.Model(&Order{}).Where("analyst_id = ? AND status = ?", analystID, OrderStatusProcessing).
		Select("COALESCE(SUM(amount), 0), COUNT(*)").Row().Scan(&amount, &count)
	return amount, count, err
}

// GetTotalCount 获取总订单数
func (r *OrderRepository) GetTotalCount(count *int64) error {
	return r.db.Model(&Order{}).Count(count).Error
}

// GetPaidCount 获取已支付订单数
func (r *OrderRepository) GetPaidCount(count *int64) error {
	return r.db.Model(&Order{}).Where("status IN ?", []OrderStatus{OrderStatusPaid, OrderStatusUploaded, OrderStatusAssigned, OrderStatusProcessing, OrderStatusCompleted}).Count(count).Error
}

// GetCompletedCount 获取已完成订单数
func (r *OrderRepository) GetCompletedCount(count *int64) error {
	return r.db.Model(&Order{}).Where("status = ?", OrderStatusCompleted).Count(count).Error
}

// GetStatusCounts 按状态统计订单数
func (r *OrderRepository) GetStatusCounts(counts interface{}) error {
	return r.db.Model(&Order{}).Select("status as status, COUNT(*) as count").Group("status").Scan(counts).Error
}

// FindCompletedUnsettled 查找已完成未结算的订单
func (r *OrderRepository) FindCompletedUnsettled(page, pageSize int) ([]Order, int64, error) {
	var orders []Order
	var total int64

	query := r.db.Model(&Order{}).Where("status = ? AND settled_at IS NULL", OrderStatusCompleted).Order("completed_at DESC")
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Find(&orders).Error
	return orders, total, err
}

// UpdateSettlement 更新订单结算状态
func (r *OrderRepository) UpdateSettlement(orderID uint, settledBy uint, settledAt time.Time) error {
	return r.db.Model(&Order{}).Where("id = ?", orderID).Updates(map[string]interface{}{
		"settled_at":     &settledAt,
		"settled_by":     settledBy,
		"settled_amount": gorm.Expr("amount * 0.7"), // 分析师分成 70%
	}).Error
}

// GetAnalystTotalIncome 获取分析师总收入
func (r *OrderRepository) GetAnalystTotalIncome(analystID uint) (float64, error) {
	var amount float64
	err := r.db.Model(&Order{}).Where("analyst_id = ? AND status = ? AND settled_at IS NOT NULL", analystID, OrderStatusCompleted).
		Select("COALESCE(SUM(settled_amount), 0)").Scan(&amount).Error
	return amount, err
}

// GetAnalystMonthIncome 获取分析师本月收入
func (r *OrderRepository) GetAnalystMonthIncome(analystID uint) (float64, error) {
	var amount float64
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	err := r.db.Model(&Order{}).Where("analyst_id = ? AND status = ? AND settled_at >= ?", analystID, OrderStatusCompleted, startOfMonth).
		Select("COALESCE(SUM(settled_amount), 0)").Scan(&amount).Error
	return amount, err
}

// GetAnalystOrderCount 获取分析师订单数
func (r *OrderRepository) GetAnalystOrderCount(analystID uint) (int64, error) {
	var count int64
	err := r.db.Model(&Order{}).Where("analyst_id = ? AND status = ?", analystID, OrderStatusCompleted).Count(&count).Error
	return count, err
}
