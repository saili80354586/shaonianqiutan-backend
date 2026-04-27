package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/shaonianqiutan/backend/models"
)

// OrderService 订单服务
type OrderService struct {
	orderRepo   *models.OrderRepository
	analystRepo *models.AnalystRepository
	reportRepo  *models.ReportRepository
	userRepo    *models.UserRepository
}

func NewOrderService(orderRepo *models.OrderRepository, analystRepo *models.AnalystRepository, reportRepo *models.ReportRepository, userRepo *models.UserRepository) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		analystRepo: analystRepo,
		reportRepo:  reportRepo,
		userRepo:    userRepo,
	}
}

// calculateAgeFromBirthDate 根据出生日期计算年龄
func calculateAgeFromBirthDate(birthDate string) int {
	if birthDate == "" {
		return 0
	}
	// 尝试解析日期
	layouts := []string{"2006-01-02", "2006/01/02", "01-02-2006"}
	var t time.Time
	var err error
	for _, layout := range layouts {
		t, err = time.Parse(layout, birthDate)
		if err == nil {
			break
		}
	}
	if err != nil {
		return 0
	}
	now := time.Now()
	age := now.Year() - t.Year()
	if now.YearDay() < t.YearDay() {
		age--
	}
	return age
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	AnalystID      *uint                `json:"analyst_id"`
	Amount         float64              `json:"amount" binding:"required,gt=0"`
	VideoURL       string               `json:"video_url"`
	VideoFilename  string               `json:"video_filename"`
	PaymentMethod  models.PaymentMethod `json:"payment_method" binding:"required,oneof=wechat alipay balance"`
	Remark         string               `json:"remark"`
	OrderType      string               `json:"order_type" binding:"required,oneof=basic pro"`
	PlayerName     string               `json:"player_name"`
	PlayerAge      int                  `json:"player_age"`
	PlayerPosition string               `json:"player_position"`
	MatchName      string               `json:"match_name"`
	VideoDuration  int                  `json:"video_duration"`
}

// SupplementOrderRequest 支付后补充订单信息请求
type SupplementOrderRequest struct {
	VideoURL       string `json:"video_url" binding:"required"`
	VideoFilename  string `json:"video_filename"`
	PlayerName     string `json:"player_name"`
	PlayerAge      int    `json:"player_age"`
	PlayerPosition string `json:"player_position"`
	JerseyColor    string `json:"jersey_color"`
	JerseyNumber   string `json:"jersey_number"`
	MatchName      string `json:"match_name"`
	VideoDuration  int    `json:"video_duration"`
	Remark         string `json:"remark"`
}

// AssignOrderRequest 管理员派单请求
type AssignOrderRequest struct {
	AnalystID uint `json:"analyst_id" binding:"required"`
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(userID uint, req *CreateOrderRequest) (*models.Order, error) {
	// 若指定了分析师，验证其是否存在且活跃
	if req.AnalystID != nil && *req.AnalystID > 0 {
		analyst, err := s.analystRepo.FindByID(*req.AnalystID)
		if err != nil {
			return nil, err
		}
		if analyst == nil {
			return nil, errors.New("分析师不存在")
		}
		if analyst.Status != models.AnalystStatusActive {
			return nil, errors.New("分析师暂不可用")
		}
	}

	// 根据套餐类型自动设置是否有视频剪辑
	hasVideoEdit := req.OrderType == "pro"

	// 生成订单号：使用纳秒时间戳 + 用户ID + 4位随机数，确保唯一性
	orderNo := fmt.Sprintf("ORD%d%04d", time.Now().UnixNano(), userID%10000)

	order := &models.Order{
		UserID:         userID,
		AnalystID:      req.AnalystID,
		OrderNo:        orderNo,
		Amount:         req.Amount,
		Status:         models.OrderStatusPending,
		PaymentMethod:  req.PaymentMethod,
		VideoURL:       req.VideoURL,
		VideoFilename:  req.VideoFilename,
		Remark:         req.Remark,
		OrderType:      req.OrderType,
		PlayerName:     req.PlayerName,
		PlayerAge:      req.PlayerAge,
		PlayerPosition: req.PlayerPosition,
		MatchName:      req.MatchName,
		VideoDuration:  req.VideoDuration,
		// 新增字段需要模型同步支持，若模型尚未扩展，先通过 map 更新或后续迁移
	}

	err := s.orderRepo.Create(order)
	if err != nil {
		return nil, err
	}

	// 若模型已支持 has_video_edit 则同步更新（GORM AutoMigrate 后生效）
	if hasVideoEdit {
		_ = s.orderRepo.Update(order.ID, map[string]interface{}{"has_video_edit": true})
	}

	// 重新加载关联数据
	return s.orderRepo.FindByID(order.ID)
}

// SupplementOrder 支付后补充订单信息（上传视频和球员资料）
func (s *OrderService) SupplementOrder(userID, orderID uint, req *SupplementOrderRequest) (*models.Order, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, errors.New("订单不存在")
	}
	if order.UserID != userID {
		return nil, errors.New("无权限操作此订单")
	}
	if order.Status != models.OrderStatusPaid {
		return nil, errors.New("订单状态不允许补充资料")
	}

	// 计算年龄：如果请求的 player_age 为 0，从用户资料获取 birth_date 计算
	playerAge := req.PlayerAge
	if playerAge == 0 && s.userRepo != nil {
		if user, err := s.userRepo.FindByID(userID); err == nil && user != nil {
			playerAge = calculateAgeFromBirthDate(user.BirthDate)
			// 同时更新用户表中的 age 字段
			if playerAge > 0 {
				_ = s.userRepo.UpdateAge(userID, playerAge)
			}
		}
	}

	updates := map[string]interface{}{
		"video_url":       req.VideoURL,
		"video_filename":  req.VideoFilename,
		"player_name":     req.PlayerName,
		"player_age":      playerAge,
		"player_position": req.PlayerPosition,
		"jersey_color":    req.JerseyColor,
		"jersey_number":   req.JerseyNumber,
		"match_name":      req.MatchName,
		"video_duration":  req.VideoDuration,
		"remark":          req.Remark,
		"status":          models.OrderStatusUploaded,
	}

	if err := s.orderRepo.Update(orderID, updates); err != nil {
		return nil, err
	}

	return s.orderRepo.FindByID(orderID)
}

// AssignOrder 管理员派单给分析师
func (s *OrderService) AssignOrder(orderID, analystID uint) (*models.Order, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, errors.New("订单不存在")
	}
	if order.Status != models.OrderStatusUploaded {
		return nil, errors.New("订单状态不允许派单")
	}

	// 验证分析师
	analyst, err := s.analystRepo.FindByID(analystID)
	if err != nil {
		return nil, err
	}
	if analyst == nil {
		return nil, errors.New("分析师不存在")
	}
	if analyst.Status != models.AnalystStatusActive {
		return nil, errors.New("分析师暂不可用")
	}

	// 计算截止时间：basic 48h，pro 72h
	deadlineHours := 48
	if order.OrderType == "pro" {
		deadlineHours = 72
	}
	deadline := time.Now().Add(time.Duration(deadlineHours) * time.Hour)

	updates := map[string]interface{}{
		"analyst_id":   analystID,
		"status":       models.OrderStatusAssigned,
		"assigned_at":  time.Now(),
		"deadline":     deadline,
	}

	if err := s.orderRepo.Update(orderID, updates); err != nil {
		return nil, err
	}

	return s.orderRepo.FindByID(orderID)
}

// GetMyOrders 获取我的订单列表
func (s *OrderService) GetMyOrders(userID uint, page, pageSize int, keyword string) ([]models.Order, int64, error) {
	return s.orderRepo.FindByUserID(userID, page, pageSize, keyword)
}

// GetOrderByID 根据ID获取订单
func (s *OrderService) GetOrderByID(id uint) (*models.Order, error) {
	return s.orderRepo.FindByID(id)
}

// CancelOrder 取消订单
func (s *OrderService) CancelOrder(userID, orderID uint) error {
	// 获取订单
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("订单不存在")
	}

	// 检查订单是否属于当前用户
	if order.UserID != userID {
		return errors.New("无权限操作此订单")
	}

	// 检查订单状态是否可以取消
	if !models.CanTransition(order.Status, models.OrderStatusCancelled) {
		return errors.New("订单状态不允许取消")
	}

	return s.orderRepo.CancelOrder(orderID)
}

// UpdateOrderStatus 更新订单状态
func (s *OrderService) UpdateOrderStatus(id uint, status models.OrderStatus) error {
	// 获取订单
	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("订单不存在")
	}

	// 检查状态流转是否合法
	if !models.CanTransition(order.Status, status) {
		return fmt.Errorf("订单状态不能从 %s 变更为 %s", order.Status, status)
	}

	return s.orderRepo.UpdateStatus(id, status)
}

// GetOrderStatistics 获取用户订单统计
func (s *OrderService) GetOrderStatistics(userID uint) (*models.UserOrderStatistics, error) {
	return s.orderRepo.GetUserOrderStatistics(userID)
}

// CompleteOrder 完成订单
func (s *OrderService) CompleteOrder(orderID uint, reportID uint) error {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("订单不存在")
	}

	// 更新订单状态和关联的报告ID
	updates := map[string]interface{}{
		"status":    models.OrderStatusCompleted,
		"report_id": reportID,
	}

	return s.orderRepo.Update(orderID, updates)
}

// GetAnalystOrders 获取分析师的订单列表
func (s *OrderService) GetAnalystOrders(analystID uint, page, pageSize int) ([]models.Order, int64, error) {
	return s.orderRepo.FindByAnalystID(analystID, page, pageSize)
}
