package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/config"
	"github.com/shaonianqiutan/backend/middleware"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/utils"
)

// PaymentController 支付控制器
type PaymentController struct {
	orderRepo *models.OrderRepository
}

// NewPaymentController 创建支付控制器
func NewPaymentController(orderRepo *models.OrderRepository) *PaymentController {
	return &PaymentController{orderRepo: orderRepo}
}

func paymentNotConfigured(c *gin.Context) {
	utils.Error(c, http.StatusServiceUnavailable, "生产支付服务尚未接入")
}

// CreatePaymentRequest 创建支付请求
type CreatePaymentRequest struct {
	OrderID uint    `json:"order_id" binding:"required"`
	Amount  float64 `json:"amount" binding:"required"`
	Type    string  `json:"type" binding:"required,oneof=wechat alipay balance"`
}

// CreatePayment 创建支付
func (ctrl *PaymentController) CreatePayment(c *gin.Context) {
	if !config.IsDevMode() {
		paymentNotConfigured(c)
		return
	}

	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 检查订单是否存在且属于当前用户
	order, err := ctrl.orderRepo.FindByID(req.OrderID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询订单失败")
		return
	}
	if order == nil {
		utils.Error(c, http.StatusNotFound, "订单不存在")
		return
	}
	if order.UserID != userID {
		utils.Error(c, http.StatusForbidden, "无权操作该订单")
		return
	}
	if order.Status != models.OrderStatusPending {
		utils.Error(c, http.StatusBadRequest, "订单状态不允许支付")
		return
	}

	// 过渡方案：返回模拟支付数据
	paymentID := fmt.Sprintf("PAY%d%d", req.OrderID, time.Now().Unix())
	utils.Success(c, "支付创建成功", gin.H{
		"payment_id": paymentID,
		"order_id":   req.OrderID,
		"amount":     req.Amount,
		"type":       req.Type,
		"status":     "pending",
		"pay_url":    "",
		"expired_at": time.Now().Add(15 * time.Minute).Format("2006-01-02 15:04:05"),
	})
}

// SimulatePayRequest 模拟支付请求
type SimulatePayRequest struct {
	OrderID       uint   `json:"order_id" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required,oneof=wechat alipay balance"`
}

// SimulatePay 模拟支付（过渡方案）
func (ctrl *PaymentController) SimulatePay(c *gin.Context) {
	if !config.IsDevMode() {
		utils.Error(c, http.StatusForbidden, "模拟支付仅允许在开发环境使用")
		return
	}

	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	var req SimulatePayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	order, err := ctrl.orderRepo.FindByID(req.OrderID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询订单失败")
		return
	}
	if order == nil {
		utils.Error(c, http.StatusNotFound, "订单不存在")
		return
	}
	if order.UserID != userID {
		utils.Error(c, http.StatusForbidden, "无权操作该订单")
		return
	}
	if order.Status != models.OrderStatusPending {
		utils.Error(c, http.StatusBadRequest, "订单状态不允许支付")
		return
	}

	// 模拟支付处理
	if req.PaymentMethod == "balance" {
		utils.Error(c, http.StatusBadRequest, "余额不足")
		return
	}

	// 更新订单状态为已支付
	updates := map[string]interface{}{
		"status":         models.OrderStatusPaid,
		"payment_method": req.PaymentMethod,
		"paid_at":        time.Now(),
	}
	if err := ctrl.orderRepo.Update(order.ID, updates); err != nil {
		utils.Error(c, http.StatusInternalServerError, "支付处理失败")
		return
	}

	utils.Success(c, "支付成功", gin.H{
		"order_id":       order.ID,
		"status":         "paid",
		"paid_at":        time.Now().Format("2006-01-02 15:04:05"),
		"amount":         order.Amount,
		"payment_method": req.PaymentMethod,
	})
}

// GetPaymentStatus 获取支付状态
func (ctrl *PaymentController) GetPaymentStatus(c *gin.Context) {
	if !config.IsDevMode() {
		paymentNotConfigured(c)
		return
	}

	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	paymentID := c.Param("id")
	if paymentID == "" {
		utils.Error(c, http.StatusBadRequest, "支付ID不能为空")
		return
	}

	utils.Success(c, "", gin.H{
		"payment_id": paymentID,
		"status":     "paid",
		"paid_at":    "2024-03-20 10:30:00",
	})
}

// GetOrderPaymentStatus 获取订单支付状态
func (ctrl *PaymentController) GetOrderPaymentStatus(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	orderIDStr := c.Param("orderId")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的订单ID")
		return
	}

	order, err := ctrl.orderRepo.FindByID(uint(orderID))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询订单失败")
		return
	}
	if order == nil {
		utils.Error(c, http.StatusNotFound, "订单不存在")
		return
	}
	if order.UserID != userID {
		utils.Error(c, http.StatusForbidden, "无权查看该订单")
		return
	}

	isPaid := order.Status == models.OrderStatusPaid || order.Status == models.OrderStatusAssigned || order.Status == models.OrderStatusProcessing || order.Status == models.OrderStatusCompleted
	utils.Success(c, "", gin.H{
		"order_id":       order.ID,
		"is_paid":        isPaid,
		"status":         string(order.Status),
		"amount":         order.Amount,
		"payment_method": order.PaymentMethod,
		"paid_at":        order.PaidAt,
	})
}

// PaymentCallback 支付回调
func (ctrl *PaymentController) PaymentCallback(c *gin.Context) {
	if !config.IsDevMode() {
		paymentNotConfigured(c)
		return
	}

	// TODO: 处理支付平台回调
	utils.Success(c, "回调处理成功", nil)
}

// Refund 申请退款
func (ctrl *PaymentController) Refund(c *gin.Context) {
	if !config.IsDevMode() {
		paymentNotConfigured(c)
		return
	}

	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	paymentID := c.Param("id")
	if paymentID == "" {
		utils.Error(c, http.StatusBadRequest, "支付ID不能为空")
		return
	}

	utils.Success(c, "退款申请已提交", gin.H{
		"refund_id": "R" + paymentID,
		"status":    "processing",
	})
}
