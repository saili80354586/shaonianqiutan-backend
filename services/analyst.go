package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/shaonianqiutan/backend/models"
	"gorm.io/gorm"
)

// AnalystService 分析师服务
type AnalystService struct {
	analystRepo       *models.AnalystRepository
	orderRepo         *models.OrderRepository
	userRepo          *models.UserRepository
	assignmentRepo    *models.OrderAssignmentRepository
	statusHistoryRepo *models.OrderStatusHistoryRepository
}

// NewAnalystService 创建分析师服务
func NewAnalystService(
	analystRepo *models.AnalystRepository,
	orderRepo *models.OrderRepository,
	userRepo *models.UserRepository,
	assignmentRepo *models.OrderAssignmentRepository,
	statusHistoryRepo *models.OrderStatusHistoryRepository,
) *AnalystService {
	return &AnalystService{
		analystRepo:       analystRepo,
		orderRepo:         orderRepo,
		userRepo:          userRepo,
		assignmentRepo:    assignmentRepo,
		statusHistoryRepo: statusHistoryRepo,
	}
}

// GetAnalystByID 获取分析师详情
func (s *AnalystService) GetAnalystByID(id uint) (*models.Analyst, error) {
	analyst, err := s.analystRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取分析师失败: %w", err)
	}
	if analyst == nil {
		return nil, errors.New("分析师不存在")
	}
	return analyst, nil
}

// GetAnalystByUserID 根据用户ID获取分析师信息
func (s *AnalystService) GetAnalystByUserID(userID uint) (*models.Analyst, error) {
	analyst, err := s.analystRepo.FindByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取分析师失败: %w", err)
	}
	if analyst == nil {
		return nil, errors.New("该用户不是分析师")
	}
	return analyst, nil
}

// GetAnalystList 获取分析师列表
func (s *AnalystService) GetAnalystList(page, pageSize int) ([]models.Analyst, int64, error) {
	analysts, total, err := s.analystRepo.FindAll(page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("获取分析师列表失败: %w", err)
	}
	return analysts, total, nil
}

// GetAnalystOrders 获取分析师订单列表
func (s *AnalystService) GetAnalystOrders(analystID uint, page, pageSize int) ([]models.Order, int64, error) {
	orders, total, err := s.orderRepo.FindByAnalystID(analystID, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("获取订单列表失败: %w", err)
	}
	return orders, total, nil
}

// GetOrderByID 获取订单详情
func (s *AnalystService) GetOrderByID(orderID uint) (*models.Order, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, errors.New("订单不存在")
	}
	return order, nil
}

// GetAnalystRevenue 获取分析师收益统计
type RevenueStats struct {
	TotalRevenue  float64 `json:"total_revenue"`  // 总收益
	OrderCount    int     `json:"order_count"`    // 订单数
	AverageRating float64 `json:"average_rating"` // 平均评分
	ThisMonthRev  float64 `json:"this_month_rev"` // 本月收益
	LastMonthRev  float64 `json:"last_month_rev"` // 上月收益
	WeekRev       float64 `json:"week_rev"`       // 本周收益
	LastWeekRev   float64 `json:"last_week_rev"`  // 上周收益
	TodayRev      float64 `json:"today_rev"`      // 今日收益
}

func (s *AnalystService) GetAnalystRevenue(analystID uint) (*RevenueStats, error) {
	// 获取所有已完成订单的收益
	orders, _, err := s.orderRepo.FindByAnalystID(analystID, 1, 1000)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("获取订单失败: %w", err)
	}

	stats := &RevenueStats{
		AverageRating: 0,
	}

	// 获取分析师信息
	analyst, err := s.analystRepo.FindByID(analystID)
	if err != nil {
		return nil, fmt.Errorf("获取分析师信息失败: %w", err)
	}
	if analyst != nil {
		stats.AverageRating = analyst.Rating
	}

	now := time.Now()
	thisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastMonth := thisMonth.AddDate(0, -1, 0)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := time.Date(now.Year(), now.Month(), now.Day()-int(now.Weekday()), 0, 0, 0, 0, now.Location())
	lastWeekStart := weekStart.AddDate(0, 0, -7)
	lastWeekEnd := weekStart

	for _, order := range orders {
		// 只统计已完成的订单
		if order.Status == models.OrderStatusCompleted {
			stats.TotalRevenue += order.Amount
			stats.OrderCount++

			// 本月收益
			if order.UpdatedAt.After(thisMonth) {
				stats.ThisMonthRev += order.Amount
			}

			// 上月收益
			if order.UpdatedAt.After(lastMonth) && order.UpdatedAt.Before(thisMonth) {
				stats.LastMonthRev += order.Amount
			}

			// 本周收益
			if order.UpdatedAt.After(weekStart) {
				stats.WeekRev += order.Amount
			}

			// 上周收益
			if order.UpdatedAt.After(lastWeekStart) && order.UpdatedAt.Before(lastWeekEnd) {
				stats.LastWeekRev += order.Amount
			}

			// 今日收益
			if order.UpdatedAt.After(today) {
				stats.TodayRev += order.Amount
			}
		}
	}

	return stats, nil
}

// UpdateAnalystProfile 更新分析师资料
type UpdateProfileRequest struct {
	Name         string `json:"name" binding:"required,min=2,max=50"`
	Bio          string `json:"bio"`
	Specialty    string `json:"specialty" binding:"max=255"`
	Experience   int    `json:"experience"`
	Profession   string `json:"profession"`
	IsProPlayer  bool   `json:"isProPlayer"`
	HasCase      bool   `json:"hasCase"`
	CaseDetail   string `json:"caseDetail"`
	ContactPhone string `json:"contactPhone"`
	ContactEmail string `json:"contactEmail"`
}

func (s *AnalystService) UpdateAnalystProfile(analystID uint, req *UpdateProfileRequest) error {
	updates := map[string]interface{}{
		"name":          req.Name,
		"bio":           req.Bio,
		"specialty":     req.Specialty,
		"experience":    req.Experience,
		"profession":    req.Profession,
		"is_pro_player": req.IsProPlayer,
		"has_case":      req.HasCase,
		"case_detail":   req.CaseDetail,
		"contact_phone": req.ContactPhone,
		"contact_email": req.ContactEmail,
	}

	err := s.analystRepo.Update(analystID, updates)
	if err != nil {
		return fmt.Errorf("更新分析师资料失败: %w", err)
	}

	return nil
}

// GetAnalystPublicProfile 获取分析师公开主页数据
type AnalystPublicProfile struct {
	Analyst       *AnalystPublicInfo `json:"analyst"`
	User          *models.User       `json:"user,omitempty"`
	Stats         PublicStats        `json:"stats"`
	SampleReports []PublicReport     `json:"sample_reports"`
	// Reviews     []Review     `json:"reviews"`  // 后续扩展
}

// AnalystPublicInfo API响应专用的分析师信息结构
type AnalystPublicInfo struct {
	ID          uint     `json:"id"`
	UserID      uint     `json:"user_id"`
	Name        string   `json:"name"`
	Bio         string   `json:"bio"`
	Specialty   []string `json:"specialty"`
	Experience  int      `json:"experience"`
	Profession  string   `json:"profession"`
	IsProPlayer bool     `json:"is_pro_player"`
	HasCase     bool     `json:"has_case"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

type PublicStats struct {
	TotalReports     int64   `json:"total_reports"`
	CompletedReports int64   `json:"completed_reports"`
	AverageRating    float64 `json:"average_rating"`
	ReviewCount      int64   `json:"review_count"`
}

type PublicReport struct {
	ID              uint   `json:"id"`
	PlayerName      string `json:"player_name"`
	PlayerPosition  string `json:"player_position"`
	Title           string `json:"title"`
	Summary         string `json:"summary"`
	OverallRating   int    `json:"overall_rating"`
	PotentialRating string `json:"potential_rating"`
	CreatedAt       string `json:"created_at"`
}

func (s *AnalystService) GetAnalystPublicProfile(analystID uint) (*AnalystPublicProfile, error) {
	analyst, err := s.analystRepo.FindByID(analystID)
	if err != nil {
		return nil, fmt.Errorf("获取分析师失败: %w", err)
	}
	if analyst == nil {
		return nil, errors.New("分析师不存在")
	}

	// 获取统计
	var totalReports int64
	var completedReports int64
	s.orderRepo.GetDB().Model(&models.Order{}).Where("analyst_id = ?", analystID).Count(&totalReports)
	s.orderRepo.GetDB().Model(&models.Order{}).Where("analyst_id = ? AND status = ?", analystID, models.OrderStatusCompleted).Count(&completedReports)

	// 获取示例报告（已完成且有评分的3-5份）
	var reports []models.Order
	s.orderRepo.GetDB().Preload("Report").Preload("User").
		Where("analyst_id = ? AND status = ?", analystID, models.OrderStatusCompleted).
		Order("updated_at DESC").
		Limit(5).
		Find(&reports)

	sampleReports := make([]PublicReport, 0, len(reports))
	for _, r := range reports {
		playerName := ""
		playerPosition := ""
		if r.User.Name != "" {
			playerName = r.User.Nickname
			if playerName == "" {
				playerName = r.User.Name
			}
			playerPosition = r.User.Position
		}
		reportTitle := ""
		reportSummary := ""
		overallRating := 0
		potentialRating := ""
		if r.Report != nil {
			reportTitle = r.Report.PlayerName + " 分析报告"
			reportSummary = r.Report.Content
			if len(reportSummary) > 100 {
				reportSummary = reportSummary[:100] + "..."
			}
		}
		sampleReports = append(sampleReports, PublicReport{
			ID:              r.ID,
			PlayerName:      playerName,
			PlayerPosition:  playerPosition,
			Title:           reportTitle,
			Summary:         reportSummary,
			OverallRating:   overallRating,
			PotentialRating: potentialRating,
			CreatedAt:       r.CreatedAt.Format("2006-01-02"),
		})
	}

	// 解析专长数组
	specialty := parseJSONArray(analyst.Specialty)

	return &AnalystPublicProfile{
		Analyst: &AnalystPublicInfo{
			ID:          analyst.ID,
			UserID:      analyst.UserID,
			Name:        analyst.Name,
			Bio:         analyst.Bio,
			Specialty:   specialty,
			Experience:  analyst.Experience,
			Profession:  analyst.Profession,
			IsProPlayer: analyst.IsProPlayer,
			HasCase:     analyst.HasCase,
			CreatedAt:   analyst.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   analyst.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		},
		User: &analyst.User,
		Stats: PublicStats{
			TotalReports:     totalReports,
			CompletedReports: completedReports,
			AverageRating:    analyst.Rating,
			ReviewCount:      int64(analyst.ReviewCount),
		},
		SampleReports: sampleReports,
	}, nil
}

// CreateAnalystFromApplication 从申请创建分析师(管理员用)
func (s *AnalystService) CreateAnalystFromApplication(userID uint, name string) (*models.Analyst, error) {
	// 检查是否已存在
	existing, _ := s.analystRepo.FindByUserID(userID)
	if existing != nil {
		return nil, errors.New("该用户已经是分析师")
	}

	// 创建分析师
	analyst := &models.Analyst{
		UserID:      userID,
		Name:        name,
		Status:      models.AnalystStatusActive,
		Rating:      0,
		ReviewCount: 0,
	}

	err := s.analystRepo.Create(analyst)
	if err != nil {
		return nil, fmt.Errorf("创建分析师失败: %w", err)
	}

	return analyst, nil
}

// GetDashboardStats 获取分析师工作台统计
func getFloat64(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case float32:
			return float64(val)
		case int:
			return float64(val)
		case int64:
			return float64(val)
		}
	}
	return 0
}

type AnalystDashboardStats struct {
	PendingCount       int64   `json:"pendingCount"`
	ActiveCount        int64   `json:"activeCount"`
	TodayDeadlineCount int64   `json:"todayDeadlineCount"`
	MonthlyIncome      float64 `json:"monthlyIncome"`
	TotalCompleted     int64   `json:"totalCompleted"`
	AvgRating          float64 `json:"avgRating"`
	TodayIncome        float64 `json:"todayIncome"`
	WeekIncome         float64 `json:"weekIncome"`
	CompletionRate     float64 `json:"completionRate"`
}

func (s *AnalystService) GetDashboardStats(analystID uint) (*AnalystDashboardStats, error) {
	rawStats, err := s.orderRepo.GetAnalystDashboardStats(analystID)
	if err != nil {
		return nil, fmt.Errorf("获取统计数据失败: %w", err)
	}

	analyst, err := s.analystRepo.FindByID(analystID)
	if err != nil {
		return nil, fmt.Errorf("获取分析师信息失败: %w", err)
	}

	avgRating := 0.0
	if analyst != nil {
		avgRating = analyst.Rating
	}

	return &AnalystDashboardStats{
		PendingCount:       rawStats["pendingCount"].(int64),
		ActiveCount:        rawStats["activeCount"].(int64),
		TodayDeadlineCount: rawStats["todayDeadlineCount"].(int64),
		MonthlyIncome:      rawStats["monthlyIncome"].(float64),
		TotalCompleted:     rawStats["totalCompleted"].(int64),
		AvgRating:          avgRating,
		TodayIncome:        getFloat64(rawStats, "todayIncome"),
		WeekIncome:         getFloat64(rawStats, "weekIncome"),
		CompletionRate:     getFloat64(rawStats, "completionRate"),
	}, nil
}

// GetPendingOrders 获取待处理订单
func (s *AnalystService) GetPendingOrders(analystID uint) ([]models.Order, error) {
	orders, err := s.orderRepo.FindPendingByAnalystID(analystID)
	if err != nil {
		return nil, fmt.Errorf("获取待处理订单失败: %w", err)
	}
	return orders, nil
}

// GetActiveOrders 获取进行中订单
func (s *AnalystService) GetActiveOrders(analystID uint) ([]models.Order, error) {
	orders, err := s.orderRepo.FindActiveByAnalystID(analystID)
	if err != nil {
		return nil, fmt.Errorf("获取进行中订单失败: %w", err)
	}
	return orders, nil
}

// GetHistoryOrders 获取历史订单
func (s *AnalystService) GetHistoryOrders(analystID uint, status, orderType, startDate, endDate, keyword string, page, pageSize int) ([]models.Order, int64, error) {
	orders, total, err := s.orderRepo.FindHistoryByAnalystID(analystID, status, orderType, startDate, endDate, keyword, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("获取历史订单失败: %w", err)
	}
	return orders, total, nil
}

// AcceptOrder 分析师接单
func (s *AnalystService) AcceptOrder(analystID, orderID uint) error {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return fmt.Errorf("查询订单失败: %w", err)
	}
	if order == nil {
		return errors.New("订单不存在")
	}
	if order.AnalystID == nil || *order.AnalystID != analystID {
		return errors.New("无权操作该订单")
	}
	if order.Status != models.OrderStatusAssigned {
		return errors.New("订单状态不正确，无法接单")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":      models.OrderStatusProcessing,
		"accepted_at": &now,
	}
	return s.orderRepo.GetDB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Order{}).Where("id = ?", orderID).Updates(updates).Error; err != nil {
			return err
		}
		if s.assignmentRepo != nil {
			if err := s.assignmentRepo.MarkLatestPendingWithTx(tx, orderID, analystID, models.OrderAssignmentStatusAccepted, "", now); err != nil {
				return err
			}
		}
		return s.createStatusHistory(tx, orderID, order.Status, models.OrderStatusProcessing, analystID, "analyst", "分析师接单")
	})
}

// RejectOrder 分析师拒绝订单
func (s *AnalystService) RejectOrder(analystID, orderID uint, reason string) error {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return fmt.Errorf("查询订单失败: %w", err)
	}
	if order == nil {
		return errors.New("订单不存在")
	}
	if order.AnalystID == nil || *order.AnalystID != analystID {
		return errors.New("无权操作该订单")
	}
	if order.Status != models.OrderStatusAssigned {
		return errors.New("订单状态不正确，无法拒绝")
	}

	updates := map[string]interface{}{
		"status":        models.OrderStatusUploaded,
		"analyst_id":    nil,
		"assigned_at":   nil,
		"accepted_at":   nil,
		"deadline":      nil,
		"cancel_reason": reason,
	}
	now := time.Now()
	return s.orderRepo.GetDB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Order{}).Where("id = ?", orderID).Updates(updates).Error; err != nil {
			return err
		}
		if s.assignmentRepo != nil {
			if err := s.assignmentRepo.MarkLatestPendingWithTx(tx, orderID, analystID, models.OrderAssignmentStatusRejected, reason, now); err != nil {
				return err
			}
		}
		return s.createStatusHistory(tx, orderID, order.Status, models.OrderStatusUploaded, analystID, "analyst", reason)
	})
}

func (s *AnalystService) createStatusHistory(tx *gorm.DB, orderID uint, fromStatus, toStatus models.OrderStatus, actorID uint, actorRole, reason string) error {
	if s.statusHistoryRepo == nil || fromStatus == toStatus {
		return nil
	}
	return s.statusHistoryRepo.CreateWithTx(tx, &models.OrderStatusHistory{
		OrderID:    orderID,
		FromStatus: fromStatus,
		ToStatus:   toStatus,
		ActorID:    optionalActorID(actorID),
		ActorRole:  actorRole,
		Reason:     reason,
	})
}

// IncomeDetailItem 收益明细项
type IncomeDetailItem struct {
	OrderID     uint    `json:"order_id"`
	OrderNo     string  `json:"order_no"`
	Amount      float64 `json:"amount"`
	Commission  float64 `json:"commission"`
	NetIncome   float64 `json:"net_income"`
	CompletedAt string  `json:"completed_at"`
	OrderType   string  `json:"order_type"`
	PlayerName  string  `json:"player_name"`
}

// IncomeDetailsResult 收益明细结果
type IncomeDetailsResult struct {
	List     []IncomeDetailItem `json:"list"`
	Summary  IncomeSummary      `json:"summary"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"pageSize"`
}

// IncomeSummary 收益摘要
type IncomeSummary struct {
	TotalIncome       float64 `json:"totalIncome"`
	TotalOrders       int64   `json:"totalOrders"`
	AvgIncome         float64 `json:"avgIncome"`
	TextCount         int64   `json:"textCount"`
	VideoCount        int64   `json:"videoCount"`
	TextIncome        float64 `json:"textIncome"`
	VideoIncome       float64 `json:"videoIncome"`
	PendingSettlement float64 `json:"pendingSettlement"`
}

func (s *AnalystService) GetIncomeDetails(analystID uint, startDate, endDate string, page, pageSize int) (*IncomeDetailsResult, error) {
	orders, total, err := s.orderRepo.GetAnalystIncomeDetails(analystID, startDate, endDate, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("获取收益明细失败: %w", err)
	}

	list := make([]IncomeDetailItem, 0, len(orders))
	var textCount, videoCount int64
	var totalIncome, textIncome, videoIncome float64

	for _, order := range orders {
		commission := 0.7
		netIncome := order.Amount * commission
		list = append(list, IncomeDetailItem{
			OrderID:     order.ID,
			OrderNo:     order.OrderNo,
			Amount:      order.Amount,
			Commission:  commission,
			NetIncome:   netIncome,
			CompletedAt: order.CompletedAt.Format("2006-01-02 15:04:05"),
			OrderType:   order.OrderType,
			PlayerName:  order.PlayerName,
		})
		totalIncome += netIncome
		if order.OrderType == "basic" || order.OrderType == "text" {
			textCount++
			textIncome += netIncome
		} else {
			videoCount++
			videoIncome += netIncome
		}
	}

	avgIncome := 0.0
	if total > 0 {
		avgIncome = totalIncome / float64(total)
	}

	// 获取待结算金额
	pendingSettlement, _, _ := s.orderRepo.GetPendingSettlementByAnalystID(analystID)
	// 按佣金比例计算净收益
	pendingSettlement = pendingSettlement * 0.7

	return &IncomeDetailsResult{
		List: list,
		Summary: IncomeSummary{
			TotalIncome:       totalIncome,
			TotalOrders:       total,
			AvgIncome:         avgIncome,
			TextCount:         textCount,
			VideoCount:        videoCount,
			TextIncome:        textIncome,
			VideoIncome:       videoIncome,
			PendingSettlement: pendingSettlement,
		},
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetIncomeTrend 获取收益趋势
func (s *AnalystService) GetIncomeTrend(analystID uint, startDate, endDate string) ([]map[string]interface{}, error) {
	return s.orderRepo.GetAnalystIncomeTrend(analystID, startDate, endDate)
}

// SubmitReportRequest 提交报告请求
type SubmitReportRequest struct {
	Ratings      map[string]interface{} `json:"ratings"`
	Summary      string                 `json:"summary"`
	Suggestions  string                 `json:"suggestions"`
	Potential    string                 `json:"potential"`
	Strengths    []string               `json:"strengths"`
	Weaknesses   []string               `json:"weaknesses"`
	ClipVideoURL string                 `json:"clip_video_url,omitempty"`
}

// SubmitReport 提交报告（带事务和文档生成）
func (s *AnalystService) SubmitReport(analystID, orderID uint, req *SubmitReportRequest) error {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return fmt.Errorf("查询订单失败: %w", err)
	}
	if order == nil {
		return errors.New("订单不存在")
	}
	if order.AnalystID == nil || *order.AnalystID != analystID {
		return errors.New("无权操作该订单")
	}
	if order.Status != models.OrderStatusProcessing {
		return errors.New("订单状态不正确，无法提交报告")
	}

	// 计算综合评分
	overallRating := 0.0
	offenseRating := 0.0
	defenseRating := 0.0
	if ratings, ok := req.Ratings["overall"].(map[string]interface{}); ok {
		count := 0
		sum := 0.0
		for _, v := range ratings {
			if detail, ok := v.(map[string]interface{}); ok {
				if score, ok := detail["score"].(float64); ok {
					sum += score
					count++
				}
			}
		}
		if count > 0 {
			overallRating = sum / float64(count)
		}
	}
	if ratings, ok := req.Ratings["offense"].(map[string]interface{}); ok {
		count := 0
		sum := 0.0
		for _, v := range ratings {
			if detail, ok := v.(map[string]interface{}); ok {
				if score, ok := detail["score"].(float64); ok {
					sum += score
					count++
				}
			}
		}
		if count > 0 {
			offenseRating = sum / float64(count)
		}
	}
	if ratings, ok := req.Ratings["defense"].(map[string]interface{}); ok {
		count := 0
		sum := 0.0
		for _, v := range ratings {
			if detail, ok := v.(map[string]interface{}); ok {
				if score, ok := detail["score"].(float64); ok {
					sum += score
					count++
				}
			}
		}
		if count > 0 {
			defenseRating = sum / float64(count)
		}
	}

	// JSON 序列化
	ratingDetailsJSON := stringifyJSON(req.Ratings)
	strengthsJSON := stringifyJSON(req.Strengths)
	weaknessesJSON := stringifyJSON(req.Weaknesses)

	report := &models.Report{
		OrderID:        orderID,
		UserID:         order.UserID,
		AnalystID:      analystID,
		PlayerName:     order.PlayerName,
		PlayerPosition: order.PlayerPosition,
		Content:        req.Summary,
		OverallRating:  overallRating,
		OffenseRating:  offenseRating,
		DefenseRating:  defenseRating,
		Summary:        req.Summary,
		Strengths:      strengthsJSON,
		Weaknesses:     weaknessesJSON,
		Suggestions:    req.Suggestions,
		Potential:      req.Potential,
		ClipVideoURL:   req.ClipVideoURL,
		RatingDetails:  ratingDetailsJSON,
		Status:         models.ReportStatusCompleted,
	}

	// 生成 MD 文档（失败不阻塞主流程）
	var ratingMDPath, playerInfoMDPath string
	analyst, _ := s.analystRepo.FindByID(analystID)
	user, _ := s.userRepo.FindByID(order.UserID)
	if analyst != nil && user != nil {
		generator := NewReportGenerator("./uploads/reports")
		ratingMDPath, playerInfoMDPath, _ = generator.GenerateReportDocs(
			order, user, analyst, req.Ratings,
			req.Summary, req.Suggestions, req.Potential,
			req.Strengths, req.Weaknesses,
		)
	}

	// 如果文档生成成功则写入路径
	if ratingMDPath != "" {
		report.RatingReportMD = ratingMDPath
	}
	if playerInfoMDPath != "" {
		report.PlayerInfoMD = playerInfoMDPath
	}

	// 生成 Word 报告（异步，不阻塞主流程）
	var aiReportURL string
	if ratingMDPath != "" && playerInfoMDPath != "" {
		aiReportURL = generateWordReport(ratingMDPath, playerInfoMDPath, order.PlayerName)
		if aiReportURL != "" {
			report.AIReportURL = aiReportURL
		}
	}

	// 使用事务
	err = s.orderRepo.GetDB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(report).Error; err != nil {
			return err
		}
		now := time.Now()
		if err := tx.Model(&models.Order{}).Where("id = ?", orderID).Updates(map[string]interface{}{
			"status":       models.OrderStatusCompleted,
			"report_id":    report.ID,
			"completed_at": &now,
		}).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("提交报告失败: %w", err)
	}
	return nil
}

// stringifyJSON 将对象序列化为JSON字符串
func stringifyJSON(v interface{}) string {
	if v == nil {
		return ""
	}
	b, _ := json.Marshal(v)
	return string(b)
}

// generateWordReport 调用 Python 脚本生成 Word 报告
// 返回生成的 Word 文件路径，失败返回空字符串
func generateWordReport(ratingMDPath, playerInfoMDPath, playerName string) string {
	// 获取脚本所在目录
	scriptPath, err := filepath.Abs("./scripts/generate_word_report.py")
	if err != nil {
		fmt.Printf("获取脚本路径失败: %v\n", err)
		return ""
	}

	// 输出目录
	outputDir := filepath.Dir(ratingMDPath)

	// 调用 Python 脚本
	cmd := exec.Command("python3", scriptPath, ratingMDPath, playerInfoMDPath, outputDir, playerName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("生成 Word 报告失败: %v, output: %s\n", err, string(output))
		return ""
	}

	result := string(output)
	if len(result) > 8 && result[:8] == "SUCCESS:" {
		wordPath := result[8:]
		wordPath = filepath.Clean(wordPath)
		// 转换为相对路径
		relPath, err := filepath.Rel("./", wordPath)
		if err != nil {
			return wordPath
		}
		return relPath
	}

	fmt.Printf("生成 Word 报告失败: %s\n", result)
	return ""
}
