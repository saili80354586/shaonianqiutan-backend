package controllers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/config"
	"github.com/shaonianqiutan/backend/middleware"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/services"
	"github.com/shaonianqiutan/backend/utils"
)

// AuthController 认证控制器
type AuthController struct {
	authService *services.AuthService
	smsService  *services.SmsService
}

func NewAuthController(authService *services.AuthService, smsService *services.SmsService) *AuthController {
	return &AuthController{
		authService: authService,
		smsService:  smsService,
	}
}

func getBrowserOS(userAgent string) (string, string) {
	ua := strings.ToLower(userAgent)
	browser := "Unknown"
	switch {
	case strings.Contains(ua, "edg/"):
		browser = "Edge"
	case strings.Contains(ua, "chrome/"):
		browser = "Chrome"
	case strings.Contains(ua, "safari/"):
		browser = "Safari"
	case strings.Contains(ua, "firefox/"):
		browser = "Firefox"
	}

	osName := "Unknown"
	switch {
	case strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad"):
		osName = "iOS"
	case strings.Contains(ua, "android"):
		osName = "Android"
	case strings.Contains(ua, "mac os"):
		osName = "macOS"
	case strings.Contains(ua, "windows"):
		osName = "Windows"
	case strings.Contains(ua, "linux"):
		osName = "Linux"
	}

	return browser, osName
}

func deviceName(browser, osName string) string {
	if browser == "Unknown" && osName == "Unknown" {
		return "未知设备"
	}
	if browser == "Unknown" {
		return osName
	}
	if osName == "Unknown" {
		return browser
	}
	return browser + " / " + osName
}

func deviceLocation(ip string) string {
	if ip == "" || ip == "::1" || strings.HasPrefix(ip, "127.") || strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") {
		return "本地网络"
	}
	return "未知地区"
}

func (ctrl *AuthController) recordLoginLog(c *gin.Context, user *models.User, phone string, status string, failReason string) {
	if config.GetDB() == nil {
		return
	}
	userAgent := c.GetHeader("User-Agent")
	browser, osName := getBrowserOS(userAgent)
	log := models.LoginLog{
		Phone:      phone,
		IP:         c.ClientIP(),
		Device:     deviceName(browser, osName),
		Browser:    browser,
		OS:         osName,
		Location:   deviceLocation(c.ClientIP()),
		Status:     status,
		FailReason: failReason,
		CreatedAt:  time.Now(),
	}
	if user != nil {
		log.UserID = user.ID
		log.Nickname = user.Nickname
		log.Role = string(user.Role)
		if log.Phone == "" {
			log.Phone = user.Phone
		}
	}
	_ = config.GetDB().Create(&log).Error
}

// SendCodeRequest 发送验证码请求
type SendCodeRequest struct {
	Phone string             `json:"phone" binding:"required"`
	Type  models.SmsCodeType `json:"type" binding:"required,oneof=register reset"`
}

// SendCode 发送短信验证码
func (ctrl *AuthController) SendCode(c *gin.Context) {
	var req SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if req.Type == models.SmsCodeTypeRegister {
		settings := models.LoadAdminSystemSettings(config.GetDB())
		if !settings.AllowRegistration {
			utils.Error(c, http.StatusForbidden, "平台注册已关闭")
			return
		}
	}

	// 根据类型检查手机号
	if req.Type == models.SmsCodeTypeRegister {
		exists, err := ctrl.authService.CheckPhoneExists(req.Phone)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, "服务器错误")
			return
		}
		if exists {
			utils.Error(c, http.StatusBadRequest, "该手机号已注册")
			return
		}
	}

	if req.Type == models.SmsCodeTypeReset {
		exists, err := ctrl.authService.CheckPhoneExists(req.Phone)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, "服务器错误")
			return
		}
		if !exists {
			utils.Error(c, http.StatusBadRequest, "该手机号未注册")
			return
		}
	}

	// 生成验证码
	// 开发模式使用固定验证码，方便测试
	var code string
	if config.IsDevMode() {
		code = "123456"
	} else {
		code = ctrl.smsService.GenerateCode()
	}

	// 创建验证码记录
	_, err := ctrl.smsService.CreateCode(req.Phone, code, req.Type)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "验证码生成失败")
		return
	}

	// 发送验证码
	result, err := ctrl.smsService.SendCode(req.Phone, code)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "发送验证码失败")
		return
	}

	// 异步清理过期验证码
	go func() {
		_ = ctrl.smsService.CleanExpired()
	}()

	// 开发模式返回验证码
	if result.DevMode {
		c.JSON(http.StatusOK, gin.H{
			"message": "验证码已生成（开发模式）",
			"code":    result.Code,
		})
		return
	}

	utils.Success(c, "验证码已发送", nil)
}

// Register 用户注册
func (ctrl *AuthController) Register(c *gin.Context) {
	settings := models.LoadAdminSystemSettings(config.GetDB())
	if !settings.AllowRegistration {
		utils.Error(c, http.StatusForbidden, "平台注册已关闭")
		return
	}

	var req services.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := ctrl.authService.Register(&req)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if result == nil {
		utils.Error(c, http.StatusBadRequest, "验证码无效或已过期")
		return
	}

	ctrl.recordLoginLog(c, result.User, req.Phone, "success", "")

	utils.Success(c, result.Message, gin.H{
		"token": result.Token,
		"user":  result.User,
	})
}

// Login 用户登录
func (ctrl *AuthController) Login(c *gin.Context) {
	var req services.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := ctrl.authService.Login(&req)
	if err != nil {
		ctrl.recordLoginLog(c, nil, req.Phone, "failed", err.Error())
		if errors.Is(err, services.ErrAccountNotActive) {
			utils.Error(c, http.StatusForbidden, err.Error())
			return
		}
		utils.Error(c, http.StatusInternalServerError, "登录失败")
		return
	}

	if result == nil {
		ctrl.recordLoginLog(c, nil, req.Phone, "failed", "手机号或密码错误")
		utils.Error(c, http.StatusBadRequest, "手机号或密码错误")
		return
	}

	ctrl.recordLoginLog(c, result.User, req.Phone, "success", "")

	utils.Success(c, result.Message, gin.H{
		"token": result.Token,
		"user":  result.User,
	})
}

// ResetPassword 重置密码
func (ctrl *AuthController) ResetPassword(c *gin.Context) {
	var req services.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	success, err := ctrl.authService.ResetPassword(&req)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "重置密码失败")
		return
	}

	if !success {
		utils.Error(c, http.StatusBadRequest, "验证码无效或已过期，或者手机号未注册")
		return
	}

	utils.Success(c, "密码重置成功", nil)
}

// GetMe 获取当前用户信息
func (ctrl *AuthController) GetMe(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	user, err := ctrl.authService.GetUserByID(userID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取用户信息失败")
		return
	}

	if user == nil {
		utils.Error(c, http.StatusNotFound, "用户不存在")
		return
	}

	utils.Success(c, "", gin.H{"user": user})
}

// UpdateMe 更新当前用户信息
func (ctrl *AuthController) UpdateMe(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	var req services.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := ctrl.authService.UpdateUser(userID, &req)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "更新用户信息失败")
		return
	}

	utils.Success(c, "用户信息更新成功", gin.H{"user": user})
}

// VerifyCode 验证码校验
func (ctrl *AuthController) VerifyCode(c *gin.Context) {
	var req services.VerifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := ctrl.authService.VerifyCode(&req)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "验证码校验失败")
		return
	}

	if !result.Valid {
		utils.Error(c, http.StatusBadRequest, "验证码无效或已过期")
		return
	}

	utils.Success(c, "验证码有效", result)
}

// RefreshToken 刷新Token
func (ctrl *AuthController) RefreshToken(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		authHeader := c.GetHeader("Authorization")
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.Error(c, http.StatusUnauthorized, "未提供认证令牌")
			return
		}

		claims, err := middleware.ParseTokenAllowExpired(parts[1])
		if err != nil {
			utils.Error(c, http.StatusUnauthorized, "认证令牌无效")
			return
		}
		if claims.ExpiresAt == nil {
			utils.Error(c, http.StatusUnauthorized, "认证令牌缺少过期时间")
			return
		}
		if time.Since(claims.ExpiresAt.Time) > 24*time.Hour {
			utils.Error(c, http.StatusUnauthorized, "登录已过期，请重新登录")
			return
		}
		userID = claims.UserID
	}

	result, err := ctrl.authService.RefreshToken(userID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Token刷新失败")
		return
	}

	if result == nil {
		utils.Error(c, http.StatusNotFound, "用户不存在")
		return
	}
	if result.User.Status != models.StatusActive {
		utils.Error(c, http.StatusForbidden, "账号未激活或已被禁用")
		return
	}

	utils.Success(c, "Token刷新成功", gin.H{
		"token": result.Token,
		"user":  result.User,
	})
}
