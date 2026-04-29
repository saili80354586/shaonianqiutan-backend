package services

import (
	"fmt"
	"time"

	"github.com/shaonianqiutan/backend/models"
	"gorm.io/gorm"
)

// ClubOrderService 俱乐部订单服务
type ClubOrderService struct {
	db *gorm.DB
}

// NewClubOrderService 创建俱乐部订单服务
func NewClubOrderService(db *gorm.DB) *ClubOrderService {
	return &ClubOrderService{db: db}
}

// GetClubByUserID 根据用户ID获取俱乐部
func (s *ClubOrderService) GetClubByUserID(userID uint) (*models.Club, error) {
	var club models.Club
	err := s.db.Where("user_id = ?", userID).First(&club).Error
	if err != nil {
		return nil, err
	}
	return &club, nil
}

// GetOrders 获取俱乐部订单
func (s *ClubOrderService) GetOrders(clubID uint, page, pageSize int, status string) ([]models.ClubOrder, int64, error) {
	var orders []models.ClubOrder
	var total int64

	query := s.db.Model(&models.ClubOrder{}).Where("club_id = ?", clubID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).
		Preload("Player").Preload("Analyst").
		Find(&orders).Error

	return orders, total, err
}

// GetOrderStats 获取俱乐部订单统计
func (s *ClubOrderService) GetOrderStats(clubID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var totalOrders int64
	s.db.Model(&models.ClubOrder{}).Where("club_id = ?", clubID).Count(&totalOrders)
	stats["totalOrders"] = totalOrders

	var totalAmount float64
	s.db.Model(&models.ClubOrder{}).Where("club_id = ?", clubID).Select("COALESCE(SUM(final_price), 0)").Scan(&totalAmount)
	stats["totalAmount"] = totalAmount

	var pendingOrders int64
	s.db.Model(&models.ClubOrder{}).Where("club_id = ? AND status IN ?", clubID, []string{"pending", "paid"}).Count(&pendingOrders)
	stats["pendingOrders"] = pendingOrders

	var completedOrders int64
	s.db.Model(&models.ClubOrder{}).Where("club_id = ? AND status = ?", clubID, "completed").Count(&completedOrders)
	stats["completedOrders"] = completedOrders

	avgOrderValue := 0.0
	if totalOrders > 0 {
		avgOrderValue = totalAmount / float64(totalOrders)
	}
	stats["avgOrderValue"] = avgOrderValue

	// 月度趋势（最近6个月）
	monthlyTrend := make([]map[string]interface{}, 0, 6)
	now := time.Now()
	for i := 5; i >= 0; i-- {
		t := now.AddDate(0, -i, 0)
		monthStr := t.Format("2006-01")
		monthLabel := t.Format("1月")

		var monthOrders int64
		var monthAmount float64
		s.db.Model(&models.ClubOrder{}).
			Where("club_id = ? AND strftime('%Y-%m', created_at) = ?", clubID, monthStr).
			Count(&monthOrders)
		s.db.Model(&models.ClubOrder{}).
			Where("club_id = ? AND strftime('%Y-%m', created_at) = ?", clubID, monthStr).
			Select("COALESCE(SUM(final_price), 0)").Scan(&monthAmount)

		monthlyTrend = append(monthlyTrend, map[string]interface{}{
			"name":   monthLabel,
			"orders": monthOrders,
			"amount": monthAmount,
		})
	}
	stats["monthlyTrend"] = monthlyTrend

	// 报告类型分布
	type reportTypeStat struct {
		ServiceType string  `gorm:"column:service_type"`
		Count       int64   `gorm:"column:count"`
		Amount      float64 `gorm:"column:amount"`
	}
	var reportTypeStats []reportTypeStat
	s.db.Raw(`
		SELECT service_type, COUNT(*) as count, COALESCE(SUM(final_price), 0) as amount
		FROM club_orders
		WHERE club_id = ?
		GROUP BY service_type
	`, clubID).Scan(&reportTypeStats)

	serviceTypeNames := map[string]string{
		"quick_report":   "快速报告",
		"full_report":    "完整报告",
		"video_analysis": "视频分析",
	}
	reportTypeDistribution := make([]map[string]interface{}, 0, len(reportTypeStats))
	for _, rt := range reportTypeStats {
		name := serviceTypeNames[rt.ServiceType]
		if name == "" {
			name = rt.ServiceType
		}
		reportTypeDistribution = append(reportTypeDistribution, map[string]interface{}{
			"name":   name,
			"value":  rt.Count,
			"amount": rt.Amount,
		})
	}
	stats["reportTypeDistribution"] = reportTypeDistribution

	// TOP消费球员
	type playerSpending struct {
		PlayerID   uint    `gorm:"column:player_id"`
		Name       string  `gorm:"column:name"`
		Orders     int64   `gorm:"column:orders"`
		TotalSpent float64 `gorm:"column:total_spent"`
		LastReport string  `gorm:"column:last_report"`
	}
	var topSpending []playerSpending
	s.db.Raw(`
		SELECT 
			co.player_id,
			u.name as name,
			COUNT(co.id) as orders,
			COALESCE(SUM(co.final_price), 0) as total_spent,
			MAX(co.created_at) as last_report
		FROM club_orders co
		JOIN users u ON co.player_id = u.id
		WHERE co.club_id = ?
		GROUP BY co.player_id, u.name
		ORDER BY total_spent DESC
		LIMIT 5
	`, clubID).Scan(&topSpending)

	topPlayers := make([]map[string]interface{}, 0, len(topSpending))
	for _, p := range topSpending {
		lastReport := ""
		if p.LastReport != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", p.LastReport); err == nil {
				lastReport = t.Format("2006-01-02")
			} else if t, err := time.Parse(time.RFC3339, p.LastReport); err == nil {
				lastReport = t.Format("2006-01-02")
			}
		}
		topPlayers = append(topPlayers, map[string]interface{}{
			"name":       p.Name,
			"orders":     p.Orders,
			"totalSpent": p.TotalSpent,
			"lastReport": lastReport,
		})
	}
	stats["topPlayers"] = topPlayers

	return stats, nil
}

// CreateOrders 批量创建订单
func (s *ClubOrderService) CreateOrders(clubID uint, userID uint, playerIDs []uint, serviceType string, analystID *uint, remark string, discount float64) ([]models.ClubOrder, error) {
	price := getServicePrice(serviceType)
	if discount <= 0 {
		discount = 1
	}

	orders := make([]models.ClubOrder, 0, len(playerIDs))
	resolvedAnalystID := uint(0)
	if analystID != nil {
		resolvedAnalystID = *analystID
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	for _, playerID := range playerIDs {
		if playerID == 0 {
			tx.Rollback()
			return nil, fmt.Errorf("无效的球员ID")
		}

		var memberCount int64
		if err := tx.Model(&models.ClubPlayer{}).
			Where("club_id = ? AND user_id = ? AND status = ?", clubID, playerID, "active").
			Count(&memberCount).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		if memberCount == 0 {
			tx.Rollback()
			return nil, fmt.Errorf("球员不属于该俱乐部")
		}

		order := models.ClubOrder{
			ClubID:      clubID,
			UserID:      userID,
			OrderNo:     generateOrderNo(),
			PlayerID:    playerID,
			AnalystID:   resolvedAnalystID,
			ServiceType: serviceType,
			Price:       price,
			Discount:    discount,
			FinalPrice:  price * discount,
			Status:      "pending",
			Remark:      remark,
		}

		if err := tx.Create(&order).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		orders = append(orders, order)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return orders, nil
}

func getServicePrice(serviceType string) float64 {
	prices := map[string]float64{
		"quick_report":   99,
		"full_report":    299,
		"video_analysis": 499,
	}
	if price, ok := prices[serviceType]; ok {
		return price
	}
	return 299
}

func generateOrderNo() string {
	now := time.Now()
	return fmt.Sprintf("CLUB%s%d", now.Format("20060102150405"), now.UnixNano()%10000)
}

// CalculateDiscount 计算团队折扣
func CalculateDiscount(playerCount int) float64 {
	switch {
	case playerCount >= 50:
		return 0.80
	case playerCount >= 20:
		return 0.85
	case playerCount >= 10:
		return 0.90
	case playerCount >= 5:
		return 0.95
	default:
		return 1.0
	}
}

// GetDiscountLabel 获取折扣标签
func GetDiscountLabel(count int) string {
	discount := CalculateDiscount(count)
	switch discount {
	case 0.80:
		return "50人以上8折"
	case 0.85:
		return "20-49人85折"
	case 0.90:
		return "10-19人9折"
	case 0.95:
		return "5-9人95折"
	default:
		return "无折扣"
	}
}

// GetOrderByID 根据ID获取订单详情
func (s *ClubOrderService) GetOrderByID(clubID uint, orderID uint) (*models.ClubOrder, error) {
	var order models.ClubOrder
	err := s.db.Where("id = ? AND club_id = ?", orderID, clubID).
		Preload("Player").Preload("Analyst").
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// CancelOrder 取消订单（仅限待支付状态）
func (s *ClubOrderService) CancelOrder(clubID uint, orderID uint) error {
	var order models.ClubOrder
	err := s.db.Where("id = ? AND club_id = ?", orderID, clubID).First(&order).Error
	if err != nil {
		return fmt.Errorf("订单不存在")
	}

	// 只有待支付状态才能取消
	if order.Status != "pending" {
		return fmt.Errorf("只有待支付状态的订单才能取消")
	}

	return s.db.Model(&order).Update("status", "cancelled").Error
}
