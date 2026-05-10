package controllers

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/config"
	"github.com/shaonianqiutan/backend/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUploadAvatarPersistsUserAvatar(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("migrate user: %v", err)
	}

	user := models.User{
		Phone:    "13930009009",
		Password: "test",
		Role:     models.RoleUser,
		Status:   models.StatusActive,
		Name:     "Avatar Player",
		Avatar:   "/old-avatar.png",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	oldDB := config.DB
	config.DB = db
	t.Cleanup(func() { config.DB = oldDB })

	uploadDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(uploadDir, "avatars"), 0755); err != nil {
		t.Fatalf("create avatar upload dir: %v", err)
	}
	if err := db.Exec("SELECT 1").Error; err != nil {
		t.Fatalf("db not ready: %v", err)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "avatar.jpg")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write([]byte("fake image bytes")); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	ctrl := NewUploadController()
	ctrl.uploadDir = uploadDir
	ctrl.baseURL = "http://example.test"

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/upload/avatar", &body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())
	c.Set("userId", user.ID)

	ctrl.UploadAvatar(c)

	if w.Code != http.StatusOK {
		t.Fatalf("upload status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			Avatar   string `json:"avatar"`
			Filename string `json:"filename"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.Success || resp.Data.Avatar == "" {
		t.Fatalf("unexpected response: %s", w.Body.String())
	}
	if !strings.HasPrefix(resp.Data.Avatar, "http://example.test/uploads/avatars/") {
		t.Fatalf("avatar url = %q, want uploaded avatar URL", resp.Data.Avatar)
	}

	var updated models.User
	if err := db.First(&updated, user.ID).Error; err != nil {
		t.Fatalf("load updated user: %v", err)
	}
	if updated.Avatar != resp.Data.Avatar {
		t.Fatalf("user avatar = %q, want %q", updated.Avatar, resp.Data.Avatar)
	}
	if resp.Data.Filename == "" {
		t.Fatal("response filename is empty")
	}
	if !strings.HasSuffix(resp.Data.Filename, ".jpg") {
		t.Fatalf("filename = %q, want .jpg suffix", resp.Data.Filename)
	}
	if _, err := os.Stat(filepath.Join(uploadDir, "avatars", resp.Data.Filename)); err != nil {
		t.Fatalf("uploaded file was not saved: %v", err)
	}
}
