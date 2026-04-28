package models

import (
	"time"

	"gorm.io/gorm"
)

// OrderStatusHistory 订单状态流转记录
type OrderStatusHistory struct {
	ID         uint        `json:"id" gorm:"primaryKey"`
	OrderID    uint        `json:"order_id" gorm:"not null;index"`
	Order      *Order      `json:"order,omitempty" gorm:"foreignKey:OrderID"`
	FromStatus OrderStatus `json:"from_status" gorm:"size:20;not null"`
	ToStatus   OrderStatus `json:"to_status" gorm:"size:20;not null"`
	ActorID    *uint       `json:"actor_id,omitempty" gorm:"index"`
	ActorRole  string      `json:"actor_role" gorm:"size:20"`
	Reason     string      `json:"reason" gorm:"size:500"`
	CreatedAt  time.Time   `json:"created_at"`
}

// OrderStatusHistoryRepository 订单状态历史仓库
type OrderStatusHistoryRepository struct {
	db *gorm.DB
}

// NewOrderStatusHistoryRepository 创建订单状态历史仓库
func NewOrderStatusHistoryRepository(db *gorm.DB) *OrderStatusHistoryRepository {
	return &OrderStatusHistoryRepository{db: db}
}

// CreateWithTx 在事务内创建状态历史
func (r *OrderStatusHistoryRepository) CreateWithTx(tx *gorm.DB, history *OrderStatusHistory) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Create(history).Error
}

// FindByOrderID 获取某订单状态流转历史
func (r *OrderStatusHistoryRepository) FindByOrderID(orderID uint) ([]OrderStatusHistory, error) {
	var histories []OrderStatusHistory
	err := r.db.
		Where("order_id = ?", orderID).
		Order("created_at ASC, id ASC").
		Find(&histories).Error
	return histories, err
}
