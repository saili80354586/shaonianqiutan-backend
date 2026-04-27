package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/services"
	"github.com/shaonianqiutan/backend/utils"
)

// ClubOrderController 俱乐部订单控制器
type ClubOrderController struct {
	orderService *services.ClubOrderService
}

// NewClubOrderController 创建俱乐部订单控制器
func NewClubOrderController(orderService *services.ClubOrderService) *ClubOrderController {
	return &ClubOrderController{orderService: orderService}
}

// GetOrders 获取俱乐部订单列表
func (c *ClubOrderController) GetOrders(ctx *gin.Context) {
	userID := ctx.GetUint("userId")

	pagination := utils.ParsePagination(ctx)
	page := pagination.Page
	pageSize := pagination.PageSize
	status := ctx.Query("status")

	club, err := c.orderService.GetClubByUserID(userID)
	if err != nil || club == nil {
		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"orders": []interface{}{},
				"pagination": gin.H{
					"page":       1,
					"pageSize":   pageSize,
					"total":      0,
					"totalPages": 0,
				},
			},
		})
		return
	}

	orders, total, err := c.orderService.GetOrders(club.ID, page, pageSize, status)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "SERVER_ERROR", "message": "获取订单列表失败"}})
		return
	}

	list := make([]interface{}, 0, len(orders))
	for _, o := range orders {
		list = append(list, gin.H{
			"id":          o.ID,
			"orderNo":     o.OrderNo,
			"playerId":    o.PlayerID,
			"playerName":  getPlayerName(o.Player),
			"analystId":   o.AnalystID,
			"analystName": getAnalystName(o.Analyst),
			"serviceType": o.ServiceType,
			"serviceName": getServiceDisplayName(o.ServiceType),
			"price":       o.Price,
			"discount":    o.Discount,
			"finalPrice":  o.FinalPrice,
			"status":      o.Status,
			"statusName":  getStatusDisplayName(o.Status),
			"remark":      o.Remark,
			"createdAt":   o.CreatedAt.Format("2006-01-02"),
			"paidAt":      formatTime(o.PaidAt),
			"completedAt": formatTime(o.CompletedAt),
		})
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"orders": list,
			"pagination": gin.H{
				"page":       page,
				"pageSize":   pageSize,
				"total":      total,
				"totalPages": totalPages,
			},
		},
	})
}

// GetStats 获取订单统计
func (c *ClubOrderController) GetStats(ctx *gin.Context) {
	userID := ctx.GetUint("userId")

	club, _ := c.orderService.GetClubByUserID(userID)
	if club == nil {
		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"totalOrders":     0,
				"totalAmount":     0,
				"pendingOrders":   0,
				"completedOrders": 0,
				"avgOrderValue":   0,
			},
		})
		return
	}

	stats, _ := c.orderService.GetOrderStats(club.ID)
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetOrder 获取订单详情
func (c *ClubOrderController) GetOrder(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	orderID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "VALIDATION_ERROR", "message": "无效的订单ID"}})
		return
	}

	club, _ := c.orderService.GetClubByUserID(userID)
	if club == nil {
		ctx.JSON(http.StatusForbidden, gin.H{"success": false, "error": gin.H{"code": "FORBIDDEN", "message": "无权限"}})
		return
	}

	order, err := c.orderService.GetOrderByID(club.ID, uint(orderID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"success": false, "error": gin.H{"code": "NOT_FOUND", "message": "订单不存在"}})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":           order.ID,
			"orderNo":      order.OrderNo,
			"playerId":     order.PlayerID,
			"playerName":   getPlayerName(order.Player),
			"analystId":    order.AnalystID,
			"analystName":  getAnalystName(order.Analyst),
			"serviceType":  order.ServiceType,
			"serviceName":  getServiceDisplayName(order.ServiceType),
			"price":        order.Price,
			"discount":     order.Discount,
			"finalPrice":   order.FinalPrice,
			"status":       order.Status,
			"statusName":   getStatusDisplayName(order.Status),
			"remark":       order.Remark,
			"createdAt":    order.CreatedAt.Format("2006-01-02 15:04:05"),
			"paidAt":       formatTime(order.PaidAt),
			"completedAt":  formatTime(order.CompletedAt),
		},
	})
}

// CancelOrder 取消订单
func (c *ClubOrderController) CancelOrder(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	orderID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "VALIDATION_ERROR", "message": "无效的订单ID"}})
		return
	}

	club, _ := c.orderService.GetClubByUserID(userID)
	if club == nil {
		ctx.JSON(http.StatusForbidden, gin.H{"success": false, "error": gin.H{"code": "FORBIDDEN", "message": "无权限"}})
		return
	}

	err = c.orderService.CancelOrder(club.ID, uint(orderID))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "CANCEL_FAILED", "message": err.Error()}})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "订单已取消",
	})
}

// CreateBatchOrders 批量创建订单
func (c *ClubOrderController) CreateBatchOrders(ctx *gin.Context) {
	userID := ctx.GetUint("userId")

	var req struct {
		PlayerIDs   []uint `json:"playerIds" binding:"required,min=1"`
		ServiceType string `json:"serviceType" binding:"required"`
		AnalystID   *uint  `json:"analystId"`
		Remark      string `json:"remark"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "VALIDATION_ERROR", "message": "参数错误"},
		})
		return
	}

	club, _ := c.orderService.GetClubByUserID(userID)
	if club == nil {
		ctx.JSON(http.StatusForbidden, gin.H{"success": false, "error": gin.H{"code": "FORBIDDEN", "message": "无权限"}})
		return
	}

	discount := services.CalculateDiscount(len(req.PlayerIDs))

	orders, err := c.orderService.CreateOrders(club.ID, userID, req.PlayerIDs, req.ServiceType, req.AnalystID, req.Remark, discount)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "SERVER_ERROR", "message": "创建订单失败"}})
		return
	}

	orderList := make([]gin.H, 0, len(orders))
	originalPrice := getServicePriceByType(req.ServiceType) * float64(len(req.PlayerIDs))
	finalPrice := originalPrice * discount

	for _, o := range orders {
		orderList = append(orderList, gin.H{
			"playerId":    o.PlayerID,
			"orderId":     o.ID,
			"price":       o.FinalPrice,
			"serviceType": o.ServiceType,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"orders": orderList,
			"summary": gin.H{
				"totalPlayers":   len(req.PlayerIDs),
				"originalPrice":  originalPrice,
				"discount":       discount,
				"discountAmount": originalPrice - finalPrice,
				"finalPrice":     finalPrice,
			},
		},
		"message": "订单创建成功",
	})
}

func getPlayerName(player *models.User) string {
	if player == nil {
		return "未知球员"
	}
	if player.Name != "" {
		return player.Name
	}
	return player.Nickname
}

func getAnalystName(analyst *models.Analyst) string {
	if analyst == nil {
		return "待分配"
	}
	if analyst.Name != "" {
		return analyst.Name
	}
	return analyst.User.Nickname
}

func getServiceDisplayName(serviceType string) string {
	names := map[string]string{
		"quick_report":   "快速分析报告",
		"full_report":    "全方位技术分析报告",
		"video_analysis": "视频分析报告",
	}
	if name, ok := names[serviceType]; ok {
		return name
	}
	return serviceType
}

func getStatusDisplayName(status string) string {
	names := map[string]string{
		"pending":    "待支付",
		"paid":       "已支付",
		"processing": "分析中",
		"completed":  "已完成",
		"cancelled":  "已取消",
	}
	if name, ok := names[status]; ok {
		return name
	}
	return status
}

func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

func getServicePriceByType(serviceType string) float64 {
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
