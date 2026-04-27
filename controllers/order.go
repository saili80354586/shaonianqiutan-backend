package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/middleware"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/services"
	"github.com/shaonianqiutan/backend/utils"
)

// OrderController 订单控制器
type OrderController struct {
	orderService *services.OrderService
}

func NewOrderController(orderService *services.OrderService) *OrderController {
	return &OrderController{orderService: orderService}
}

// CreateOrder 创建订单
func (ctrl *OrderController) CreateOrder(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	var req services.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	order, err := ctrl.orderService.CreateOrder(userID, &req)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "创建订单失败: "+err.Error())
		return
	}

	utils.Success(c, "订单创建成功", gin.H{"order": order})
}

// GetMyOrders 获取我的订单列表
func (ctrl *OrderController) GetMyOrders(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	pagination := utils.ParsePaginationWithSize(c, 10)
	page := pagination.Page
	pageSize := pagination.PageSize
	keyword := c.DefaultQuery("keyword", "")

	orders, total, err := ctrl.orderService.GetMyOrders(userID, page, pageSize, keyword)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取订单列表失败: "+err.Error())
		return
	}

	utils.Success(c, "", gin.H{
		"list":       orders,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// SupplementOrder 支付后补充订单信息
func (ctrl *OrderController) SupplementOrder(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的订单ID")
		return
	}

	var req services.SupplementOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	order, err := ctrl.orderService.SupplementOrder(userID, uint(id), &req)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "补充订单信息失败: "+err.Error())
		return
	}

	utils.Success(c, "资料提交成功", gin.H{"order": order})
}

// GetOrderDetail 获取订单详情
func (ctrl *OrderController) GetOrderDetail(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的订单ID")
		return
	}

	order, err := ctrl.orderService.GetOrderByID(uint(id))
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取订单详情失败")
		return
	}

	if order == nil {
		utils.Error(c, http.StatusNotFound, "订单不存在")
		return
	}

	// 检查订单是否属于当前用户
	if order.UserID != userID {
		utils.Error(c, http.StatusForbidden, "无权限查看此订单")
		return
	}

	utils.Success(c, "", gin.H{"order": order})
}

// PayOrder 支付订单
func (ctrl *OrderController) PayOrder(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的订单ID")
		return
	}

	// 获取支付方式
	var req struct {
		Method string `json:"method" binding:"required,oneof=wechat alipay"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 模拟支付返回二维码链接
	var qrCodeUrl string
	if req.Method == "wechat" {
		qrCodeUrl = "weixin://wxpay/bizpayurl?pr=mock123"
	} else {
		qrCodeUrl = "https://qr.alipay.com/mock123"
	}

	utils.Success(c, "支付创建成功", gin.H{
		"order_id":    id,
		"qr_code_url": qrCodeUrl,
		"method":      req.Method,
		"status":      "pending",
	})
}

// CancelOrder 取消订单
func (ctrl *OrderController) CancelOrder(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的订单ID")
		return
	}

	err = ctrl.orderService.CancelOrder(userID, uint(id))
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "取消订单失败: "+err.Error())
		return
	}

	utils.Success(c, "订单已取消", nil)
}

// UpdateOrderStatus 更新订单状态
func (ctrl *OrderController) UpdateOrderStatus(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的订单ID")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=pending paid processing completed cancelled refunded"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 将字符串转换为OrderStatus类型
	status := models.OrderStatus(req.Status)

	err = ctrl.orderService.UpdateOrderStatus(uint(id), status)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "更新订单状态失败: "+err.Error())
		return
	}

	utils.Success(c, "订单状态更新成功", nil)
}

// GetOrderStatistics 获取用户订单统计
func (ctrl *OrderController) GetOrderStatistics(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	stats, err := ctrl.orderService.GetOrderStatistics(userID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取订单统计失败")
		return
	}

	utils.Success(c, "", gin.H{"statistics": stats})
}

// DownloadAIReport 球员端下载 AI 报告
func (ctrl *OrderController) DownloadAIReport(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的订单ID")
		return
	}

	order, err := ctrl.orderService.GetOrderByID(uint(orderID))
	if err != nil || order == nil {
		utils.Error(c, http.StatusNotFound, "订单不存在")
		return
	}

	if order.UserID != userID {
		utils.Error(c, http.StatusForbidden, "无权限")
		return
	}

	if order.Report == nil {
		utils.Error(c, http.StatusNotFound, "报告不存在")
		return
	}

	reportType := c.DefaultQuery("type", "report") // report | video
	var filePath string
	if reportType == "video" {
		filePath = "./uploads/reports/" + strings.TrimPrefix(order.Report.AIVideoURL, "/uploads/reports/")
		if order.Report.AIVideoURL == "" {
			utils.Error(c, http.StatusNotFound, "AI 视频分析尚未上传")
			return
		}
	} else {
		filePath = "./uploads/reports/" + strings.TrimPrefix(order.Report.AIReportURL, "/uploads/reports/")
		if order.Report.AIReportURL == "" {
			utils.Error(c, http.StatusNotFound, "AI 报告尚未上传")
			return
		}
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		utils.Error(c, http.StatusNotFound, "文件已被删除")
		return
	}

	fileName := filepath.Base(filePath)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", fileName))
	c.File(filePath)
}
