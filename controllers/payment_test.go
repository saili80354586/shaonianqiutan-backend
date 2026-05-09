package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupPaymentControllerTest(t *testing.T) (*gorm.DB, *PaymentController, models.User, models.Order) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Order{}); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	user := models.User{
		Phone:    "13930009001",
		Password: "test",
		Role:     models.RoleUser,
		Status:   models.StatusActive,
		Name:     "Payment Player",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	order := models.Order{
		UserID:        user.ID,
		OrderNo:       "PAY-TEST-001",
		Amount:        799,
		Status:        models.OrderStatusPending,
		PaymentMethod: models.PaymentMethodWechat,
		OrderType:     "pro",
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	return db, NewPaymentController(models.NewOrderRepository(db)), user, order
}

func TestSimulatePayMarksOrderPaid(t *testing.T) {
	t.Setenv("PAYMENT_MODE", "mock")

	db, ctrl, user, order := setupPaymentControllerTest(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/payment/simulate", strings.NewReader(`{"order_id":`+strconv.FormatUint(uint64(order.ID), 10)+`,"payment_method":"wechat"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("userId", user.ID)

	ctrl.SimulatePay(c)

	if w.Code != http.StatusOK {
		t.Fatalf("simulate pay status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var body struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !body.Success {
		t.Fatalf("simulate pay response = %s, want success", w.Body.String())
	}

	var refreshed models.Order
	if err := db.First(&refreshed, order.ID).Error; err != nil {
		t.Fatalf("reload order: %v", err)
	}
	if refreshed.Status != models.OrderStatusPaid {
		t.Fatalf("order status = %s, want %s", refreshed.Status, models.OrderStatusPaid)
	}
	if refreshed.PaidAt == nil {
		t.Fatalf("paid_at = nil, want timestamp")
	}
}

func TestSimulatePayBlockedWhenPaymentModeIsReal(t *testing.T) {
	t.Setenv("PAYMENT_MODE", "real")

	db, ctrl, user, order := setupPaymentControllerTest(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/payment/simulate", strings.NewReader(`{"order_id":`+strconv.FormatUint(uint64(order.ID), 10)+`,"payment_method":"wechat"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("userId", user.ID)

	ctrl.SimulatePay(c)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("simulate pay status = %d, want %d, body=%s", w.Code, http.StatusServiceUnavailable, w.Body.String())
	}

	var refreshed models.Order
	if err := db.First(&refreshed, order.ID).Error; err != nil {
		t.Fatalf("reload order: %v", err)
	}
	if refreshed.Status != models.OrderStatusPending {
		t.Fatalf("order status = %s, want %s", refreshed.Status, models.OrderStatusPending)
	}
	if refreshed.PaidAt != nil {
		t.Fatalf("paid_at = %v, want nil", refreshed.PaidAt)
	}
}
