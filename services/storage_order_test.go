package services

import (
	"strings"
	"testing"

	"github.com/shaonianqiutan/backend/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupStorageOrderTest(t *testing.T) (*gorm.DB, *OrderService) {
	t.Helper()
	t.Setenv("STORAGE_DRIVER", "local")
	t.Setenv("BASE_URL", "http://localhost:8080")
	t.Setenv("COS_OBJECT_PREFIX", "prod")

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Analyst{}, &models.Report{}, &models.Order{}, &models.StorageObject{}); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	orderRepo := models.NewOrderRepository(db)
	storageRepo := models.NewStorageObjectRepository(db)
	storageService := NewStorageService(storageRepo)
	orderService := NewOrderService(orderRepo, models.NewAnalystRepository(db), models.NewReportRepository(db), models.NewUserRepository(db), storageService)
	return db, orderService
}

func TestCreateOrderSourceVideoUploadIntentUsesOrderScopedObjectKey(t *testing.T) {
	db, orderService := setupStorageOrderTest(t)
	user := models.User{Nickname: "player-storage", Phone: "18800000001", Password: "x", Role: models.RoleUser}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	order := models.Order{
		UserID:        user.ID,
		OrderNo:       "STORAGE-UPLOAD-001",
		Amount:        299,
		Status:        models.OrderStatusPaid,
		PaymentMethod: models.PaymentMethodWechat,
		OrderType:     "basic",
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	intent, err := orderService.CreateOrderSourceVideoUploadIntent(user.ID, order.ID, &CreateOrderSourceVideoUploadRequest{
		Filename:    "match.mp4",
		ContentType: "video/mp4",
		Size:        1024,
	})
	if err != nil {
		t.Fatalf("CreateOrderSourceVideoUploadIntent error: %v", err)
	}
	if intent.Driver != "local" {
		t.Fatalf("driver = %q, want local", intent.Driver)
	}
	wantPrefix := "prod/orders/"
	if !strings.HasPrefix(intent.ObjectKey, wantPrefix) || !strings.Contains(intent.ObjectKey, "/source/original_") {
		t.Fatalf("object key = %q, want order-scoped source path", intent.ObjectKey)
	}
	if !strings.HasSuffix(intent.ObjectKey, ".mp4") {
		t.Fatalf("object key = %q, want .mp4 suffix", intent.ObjectKey)
	}

	var object models.StorageObject
	if err := db.First(&object, intent.StorageObjectID).Error; err != nil {
		t.Fatalf("find storage object: %v", err)
	}
	if object.OwnerType != models.StorageOwnerOrder || object.OwnerID != order.ID || object.BusinessType != models.StorageBusinessOrderSourceVideo {
		t.Fatalf("object owner/business = %s/%d/%s", object.OwnerType, object.OwnerID, object.BusinessType)
	}
	if object.Status != models.StorageObjectStatusPendingUpload {
		t.Fatalf("object status = %q, want pending_upload", object.Status)
	}
}

func TestConfirmOrderSourceVideoBindsStorageObjectToPaidOrder(t *testing.T) {
	db, orderService := setupStorageOrderTest(t)
	user := models.User{Nickname: "player-confirm", Phone: "18800000002", Password: "x", Role: models.RoleUser}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	order := models.Order{
		UserID:        user.ID,
		OrderNo:       "STORAGE-CONFIRM-001",
		Amount:        799,
		Status:        models.OrderStatusPaid,
		PaymentMethod: models.PaymentMethodWechat,
		OrderType:     "pro",
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}
	intent, err := orderService.CreateOrderSourceVideoUploadIntent(user.ID, order.ID, &CreateOrderSourceVideoUploadRequest{
		Filename:    "confirm.mp4",
		ContentType: "video/mp4",
		Size:        2048,
	})
	if err != nil {
		t.Fatalf("CreateOrderSourceVideoUploadIntent error: %v", err)
	}

	updated, err := orderService.ConfirmOrderSourceVideo(user.ID, order.ID, &ConfirmOrderSourceVideoRequest{
		StorageObjectID: intent.StorageObjectID,
		ObjectKey:       intent.ObjectKey,
		ETag:            "test-etag",
		Size:            2048,
		VideoDuration:   3600,
	})
	if err != nil {
		t.Fatalf("ConfirmOrderSourceVideo error: %v", err)
	}
	if updated.Status != models.OrderStatusUploaded {
		t.Fatalf("order status = %q, want uploaded", updated.Status)
	}
	if updated.VideoStorageObjectID == nil || *updated.VideoStorageObjectID != intent.StorageObjectID {
		t.Fatalf("video_storage_object_id = %v, want %d", updated.VideoStorageObjectID, intent.StorageObjectID)
	}
	if !strings.Contains(updated.VideoURL, intent.ObjectKey) {
		t.Fatalf("video_url = %q, want object key %q", updated.VideoURL, intent.ObjectKey)
	}
	if updated.VideoDuration != 3600 {
		t.Fatalf("video_duration = %d, want 3600", updated.VideoDuration)
	}

	var object models.StorageObject
	if err := db.First(&object, intent.StorageObjectID).Error; err != nil {
		t.Fatalf("find storage object: %v", err)
	}
	if object.Status != models.StorageObjectStatusActive {
		t.Fatalf("object status = %q, want active", object.Status)
	}
	if object.ETag != "test-etag" {
		t.Fatalf("object etag = %q, want test-etag", object.ETag)
	}
}
