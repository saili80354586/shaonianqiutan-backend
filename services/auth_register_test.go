package services

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/shaonianqiutan/backend/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAuthRegisterTestService(t *testing.T) (*AuthService, *SmsService, *gorm.DB) {
	t.Helper()

	t.Setenv("JWT_SECRET", "auth-register-test-secret")
	t.Setenv("JWT_EXPIRES_IN", "168h")
	t.Setenv("SMS_MODE", "mock")
	t.Setenv("ANALYST_REGISTRATION_AUTO_APPROVE", "true")

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.SmsCode{},
		&models.UserRoleRecord{},
		&models.Analyst{},
		&models.Order{},
		&models.OrderAssignment{},
		&models.OrderStatusHistory{},
		&models.Scout{},
		&models.Club{},
		&models.ClubCoach{},
		&models.TeamCoach{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	smsService := NewSmsService(models.NewSmsCodeRepository(db))
	authService := NewAuthService(
		models.NewUserRepository(db),
		models.NewAnalystRepository(db),
		models.NewOrderRepository(db),
		models.NewOrderAssignmentRepository(db),
		models.NewOrderStatusHistoryRepository(db),
		smsService,
		db,
	)
	return authService, smsService, db
}

func TestAnalystRegisterCreatesActiveLoginReadyAnalystProfile(t *testing.T) {
	authService, smsService, db := setupAuthRegisterTestService(t)
	phone := "13900008888"
	code := "123456"
	if _, err := smsService.CreateCode(phone, code, models.SmsCodeTypeRegister); err != nil {
		t.Fatalf("create sms code: %v", err)
	}

	result, err := authService.Register(&RegisterRequest{
		Phone:        phone,
		Code:         code,
		Password:     "123456",
		Role:         "analyst",
		Name:         "测试分析师",
		Nickname:     "测试分析",
		Profession:   "coach",
		Experience:   "5年青少年比赛视频分析经验",
		IsProPlayer:  true,
		HasCase:      true,
		CaseDetail:   "完成过区域青训比赛分析。",
		ContactPhone: phone,
		ContactEmail: "analyst@example.com",
	})
	if err != nil {
		t.Fatalf("register analyst: %v", err)
	}
	if result == nil || result.Token == "" || result.User == nil {
		t.Fatalf("register result = %#v, want token and user", result)
	}
	if result.User.Status != models.StatusActive {
		t.Fatalf("registered user status = %q, want active", result.User.Status)
	}
	if result.User.CurrentRole != models.RoleAnalyst {
		t.Fatalf("current role = %q, want analyst", result.User.CurrentRole)
	}

	var analyst models.Analyst
	if err := db.Where("user_id = ?", result.User.ID).First(&analyst).Error; err != nil {
		t.Fatalf("find analyst profile: %v", err)
	}
	if analyst.Status != models.AnalystStatusActive {
		t.Fatalf("analyst status = %q, want active", analyst.Status)
	}
	if analyst.Name != "测试分析师" || analyst.ContactPhone != phone || analyst.ContactEmail != "analyst@example.com" {
		t.Fatalf("analyst profile not populated: %#v", analyst)
	}

	var roleRecord models.UserRoleRecord
	if err := db.Where("user_id = ? AND role = ?", result.User.ID, models.RoleAnalyst).First(&roleRecord).Error; err != nil {
		t.Fatalf("find analyst role record: %v", err)
	}
	if roleRecord.Status != "active" {
		t.Fatalf("role record status = %q, want active", roleRecord.Status)
	}

	login, err := authService.Login(&LoginRequest{Phone: phone, Password: "123456"})
	if err != nil {
		t.Fatalf("login registered analyst: %v", err)
	}
	if login == nil || login.User == nil || login.User.CurrentRole != models.RoleAnalyst {
		t.Fatalf("login result = %#v, want analyst current role", login)
	}
}

func TestAnalystRegisterCreatesAssignedDefaultDemoOrderWhenEnabled(t *testing.T) {
	authService, smsService, db := setupAuthRegisterTestService(t)

	player := &models.User{
		Phone:     "15638160405",
		Password:  "unused",
		Name:      "程奕",
		BirthDate: "2014-11-21",
		Role:      models.RoleUser,
		Status:    models.StatusActive,
	}
	if err := db.Create(player).Error; err != nil {
		t.Fatalf("create player: %v", err)
	}

	templateOrder := &models.Order{
		UserID:         player.ID,
		OrderNo:        "TPL-ANALYST-DEMO-001",
		Amount:         0,
		Status:         models.OrderStatusUploaded,
		PaymentMethod:  models.PaymentMethodWechat,
		VideoURL:       "https://cdn.example.com/demo.mp4",
		VideoFilename:  "demo.mp4",
		Remark:         "系统样例模板订单",
		OrderType:      "pro",
		PlayerName:     "程奕",
		PlayerPosition: "右前锋",
		JerseyColor:    "黄色",
		JerseyNumber:   "18号",
		MatchName:      "河南星途 vs 金桥小学",
	}
	if err := db.Create(templateOrder).Error; err != nil {
		t.Fatalf("create template order: %v", err)
	}

	t.Setenv("ANALYST_DEFAULT_DEMO_ORDER_ENABLED", "true")
	t.Setenv("ANALYST_DEFAULT_DEMO_ORDER_TEMPLATE_ORDER_ID", fmt.Sprintf("%d", templateOrder.ID))

	phone := "13900009999"
	code := "123456"
	if _, err := smsService.CreateCode(phone, code, models.SmsCodeTypeRegister); err != nil {
		t.Fatalf("create sms code: %v", err)
	}

	result, err := authService.Register(&RegisterRequest{
		Phone:        phone,
		Code:         code,
		Password:     "123456",
		Role:         "analyst",
		Name:         "默认订单分析师",
		ContactPhone: phone,
	})
	if err != nil {
		t.Fatalf("register analyst: %v", err)
	}
	if result == nil || result.User == nil {
		t.Fatalf("register result = %#v, want user", result)
	}

	var analyst models.Analyst
	if err := db.Where("user_id = ?", result.User.ID).First(&analyst).Error; err != nil {
		t.Fatalf("find analyst profile: %v", err)
	}

	var clonedOrder models.Order
	if err := db.Where("analyst_id = ?", analyst.ID).First(&clonedOrder).Error; err != nil {
		t.Fatalf("find cloned order: %v", err)
	}
	if clonedOrder.Status != models.OrderStatusAssigned {
		t.Fatalf("cloned order status = %q, want assigned", clonedOrder.Status)
	}
	if clonedOrder.UserID != player.ID || clonedOrder.VideoURL != templateOrder.VideoURL {
		t.Fatalf("cloned order mismatch: %#v", clonedOrder)
	}
	wantAge := calculateAgeFromBirthDate(player.BirthDate)
	if clonedOrder.PlayerAge != wantAge {
		t.Fatalf("cloned order player_age = %d, want %d", clonedOrder.PlayerAge, wantAge)
	}
	if !strings.Contains(clonedOrder.Remark, "系统样例订单") {
		t.Fatalf("cloned order remark = %q, want marker", clonedOrder.Remark)
	}

	var assignment models.OrderAssignment
	if err := db.Where("order_id = ? AND analyst_id = ?", clonedOrder.ID, analyst.ID).First(&assignment).Error; err != nil {
		t.Fatalf("find assignment: %v", err)
	}
	if assignment.Status != models.OrderAssignmentStatusPending {
		t.Fatalf("assignment status = %q, want pending", assignment.Status)
	}

	var history models.OrderStatusHistory
	if err := db.Where("order_id = ?", clonedOrder.ID).First(&history).Error; err != nil {
		t.Fatalf("find history: %v", err)
	}
	if history.FromStatus != models.OrderStatusUploaded || history.ToStatus != models.OrderStatusAssigned {
		t.Fatalf("history = %#v, want uploaded -> assigned", history)
	}
}

func TestMockSmsSendCodeReturnsFixedCodeOutsideDevelopment(t *testing.T) {
	_, smsService, _ := setupAuthRegisterTestService(t)
	t.Setenv("NODE_ENV", "production")
	t.Setenv("SMS_MODE", "mock")

	code := "123456"
	result, err := smsService.SendCode("13900007777", code)
	if err != nil {
		t.Fatalf("send mock sms: %v", err)
	}
	if !result.DevMode || result.Code != code {
		t.Fatalf("SendCode() = %#v, want mock result with fixed code", result)
	}
}

func TestFirstInt(t *testing.T) {
	cases := map[string]int{
		"5年青少年比赛视频分析经验": 5,
		"暂无":            0,
		"12 months":     12,
	}
	for input, want := range cases {
		if got := firstInt(input); got != want {
			t.Fatalf("firstInt(%q) = %d, want %d", input, got, want)
		}
	}
}

func TestSmsCodeExpiresAtFuture(t *testing.T) {
	_, smsService, _ := setupAuthRegisterTestService(t)
	code, err := smsService.CreateCode("13900006666", "123456", models.SmsCodeTypeRegister)
	if err != nil {
		t.Fatalf("create code: %v", err)
	}
	if !code.ExpiresAt.After(time.Now()) {
		t.Fatalf("expires_at = %v, want future time", code.ExpiresAt)
	}
}
