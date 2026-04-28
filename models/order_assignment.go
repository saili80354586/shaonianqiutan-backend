package models

import (
	"time"

	"gorm.io/gorm"
)

// OrderAssignmentStatus 订单派发响应状态
type OrderAssignmentStatus string

const (
	OrderAssignmentStatusPending  OrderAssignmentStatus = "pending"
	OrderAssignmentStatusAccepted OrderAssignmentStatus = "accepted"
	OrderAssignmentStatusRejected OrderAssignmentStatus = "rejected"
	OrderAssignmentStatusExpired  OrderAssignmentStatus = "expired"
)

// OrderAssignment 订单派发记录
type OrderAssignment struct {
	ID             uint                  `json:"id" gorm:"primaryKey"`
	OrderID        uint                  `json:"order_id" gorm:"not null;index"`
	Order          *Order                `json:"order,omitempty" gorm:"foreignKey:OrderID"`
	AnalystID      uint                  `json:"analyst_id" gorm:"not null;index"`
	Analyst        *Analyst              `json:"analyst,omitempty" gorm:"foreignKey:AnalystID"`
	AssignedBy     *uint                 `json:"assigned_by,omitempty" gorm:"index"`
	AssignedByUser *User                 `json:"assigned_by_user,omitempty" gorm:"foreignKey:AssignedBy"`
	AssignedAt     time.Time             `json:"assigned_at" gorm:"not null;index"`
	Status         OrderAssignmentStatus `json:"status" gorm:"size:20;not null;default:'pending';index"`
	RejectedReason string                `json:"rejected_reason" gorm:"size:500"`
	RespondedAt    *time.Time            `json:"responded_at"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
}

// OrderAssignmentRepository 订单派发记录仓库
type OrderAssignmentRepository struct {
	db *gorm.DB
}

// NewOrderAssignmentRepository 创建订单派发记录仓库
func NewOrderAssignmentRepository(db *gorm.DB) *OrderAssignmentRepository {
	return &OrderAssignmentRepository{db: db}
}

// CreateWithTx 在事务内创建派发记录
func (r *OrderAssignmentRepository) CreateWithTx(tx *gorm.DB, assignment *OrderAssignment) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Create(assignment).Error
}

// FindAll 获取派发记录列表
func (r *OrderAssignmentRepository) FindAll(page, pageSize int, status string) ([]OrderAssignment, int64, error) {
	var assignments []OrderAssignment
	var total int64

	query := r.db.Model(&OrderAssignment{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.
		Preload("Order.User").
		Preload("Analyst.User").
		Preload("AssignedByUser").
		Order("assigned_at DESC, created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&assignments).Error
	return assignments, total, err
}

// MarkLatestPendingWithTx 标记最近一次待响应派发记录
func (r *OrderAssignmentRepository) MarkLatestPendingWithTx(tx *gorm.DB, orderID, analystID uint, status OrderAssignmentStatus, reason string, respondedAt time.Time) error {
	db := r.db
	if tx != nil {
		db = tx
	}

	var assignment OrderAssignment
	err := db.
		Where("order_id = ? AND analyst_id = ? AND status = ?", orderID, analystID, OrderAssignmentStatusPending).
		Order("assigned_at DESC, id DESC").
		First(&assignment).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}

	return db.Model(&OrderAssignment{}).Where("id = ?", assignment.ID).Updates(map[string]interface{}{
		"status":          status,
		"rejected_reason": reason,
		"responded_at":    &respondedAt,
	}).Error
}

// IsValidOrderAssignmentStatus 校验派发状态筛选值
func IsValidOrderAssignmentStatus(status string) bool {
	switch OrderAssignmentStatus(status) {
	case OrderAssignmentStatusPending, OrderAssignmentStatusAccepted, OrderAssignmentStatusRejected, OrderAssignmentStatusExpired:
		return true
	default:
		return false
	}
}

// BackfillOrderAssignmentsFromOrders 从现有订单字段补一次历史记录，避免老演示订单在新页面空白。
func BackfillOrderAssignmentsFromOrders(db *gorm.DB) error {
	var orders []Order
	if err := db.Where("analyst_id IS NOT NULL AND assigned_at IS NOT NULL").Find(&orders).Error; err != nil {
		return err
	}

	for _, order := range orders {
		if order.AnalystID == nil || order.AssignedAt == nil {
			continue
		}

		var count int64
		if err := db.Model(&OrderAssignment{}).Where("order_id = ?", order.ID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}

		status := OrderAssignmentStatusPending
		var respondedAt *time.Time
		rejectedReason := ""
		switch order.Status {
		case OrderStatusProcessing, OrderStatusCompleted:
			status = OrderAssignmentStatusAccepted
			if order.AcceptedAt != nil {
				respondedAt = order.AcceptedAt
			} else {
				t := order.UpdatedAt
				respondedAt = &t
			}
		case OrderStatusCancelled:
			if order.CancelReason != "" {
				status = OrderAssignmentStatusRejected
				rejectedReason = order.CancelReason
			} else {
				status = OrderAssignmentStatusExpired
			}
			t := order.UpdatedAt
			respondedAt = &t
		}

		assignment := &OrderAssignment{
			OrderID:        order.ID,
			AnalystID:      *order.AnalystID,
			AssignedAt:     *order.AssignedAt,
			Status:         status,
			RejectedReason: rejectedReason,
			RespondedAt:    respondedAt,
		}
		if err := db.Create(assignment).Error; err != nil {
			return err
		}
	}

	return nil
}
