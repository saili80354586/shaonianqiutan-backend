package controllers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/config"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/routes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTrialInviteTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "trial-invite-test-secret")
	dbName := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(sqlite.Open("file:"+dbName+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.TrialInvite{}, &models.Notification{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	oldDB := config.DB
	config.DB = db
	t.Cleanup(func() { config.DB = oldDB })

	router := gin.New()
	api := router.Group("/api")
	routes.SetupTrialInviteRoutes(api, controllers.NewTrialInviteController())
	return router, db
}

func authHeader(t *testing.T, user models.User) string {
	t.Helper()
	token, err := middleware.GenerateToken(user.ID, user.Phone)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	return "Bearer " + token
}

func TestTrialInviteListAndRespond(t *testing.T) {
	router, db := setupTrialInviteTestRouter(t)
	now := time.Date(2026, 4, 20, 9, 0, 0, 0, time.UTC)
	sender := models.User{ID: 7101, Phone: "13900007101", Password: "hashed", Role: models.RoleScout, CurrentRole: models.RoleScout, Status: models.StatusActive, Name: "测试球探", CreatedAt: now, UpdatedAt: now}
	player := models.User{ID: 7102, Phone: "13900007102", Password: "hashed", Role: models.RoleUser, CurrentRole: models.RoleUser, Status: models.StatusActive, Name: "测试球员", CreatedAt: now, UpdatedAt: now}
	if err := db.Create(&sender).Error; err != nil {
		t.Fatalf("create sender: %v", err)
	}
	if err := db.Create(&player).Error; err != nil {
		t.Fatalf("create player: %v", err)
	}

	createBody := `{"player_id":7102,"trial_date":"2026-05-02","trial_time":"15:00","location":"测试基地","contact_name":"王老师","contact_phone":"13900000000","note":"带球鞋"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/trial-invites", strings.NewReader(createBody))
	createReq.Header.Set("Authorization", authHeader(t, sender))
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusOK {
		t.Fatalf("create status = %d, body = %s", createRec.Code, createRec.Body.String())
	}

	var invite models.TrialInvite
	if err := db.Where("player_id = ? AND sender_id = ?", player.ID, sender.ID).First(&invite).Error; err != nil {
		t.Fatalf("find invite: %v", err)
	}
	var playerNotificationCount int64
	db.Model(&models.Notification{}).Where("user_id = ? AND type = ?", player.ID, models.NotificationTypeTrialInvite).Count(&playerNotificationCount)
	if playerNotificationCount != 1 {
		t.Fatalf("player trial invite notifications = %d, want 1", playerNotificationCount)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/trial-invites/my", nil)
	listReq.Header.Set("Authorization", authHeader(t, player))
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d, body = %s", listRec.Code, listRec.Body.String())
	}
	var listBody struct {
		Success bool `json:"success"`
		Data    struct {
			Total int `json:"total"`
			List  []struct {
				ID         uint   `json:"id"`
				SenderName string `json:"sender_name"`
				Status     string `json:"status"`
			} `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(listRec.Body.Bytes(), &listBody); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if !listBody.Success || listBody.Data.Total != 1 || listBody.Data.List[0].Status != "pending" {
		t.Fatalf("unexpected list body: %+v", listBody)
	}

	respondReq := httptest.NewRequest(http.MethodPut, "/api/trial-invites/"+itoaTest(int(invite.ID))+"/respond", strings.NewReader(`{"status":"accepted","response_note":"准时参加"}`))
	respondReq.Header.Set("Authorization", authHeader(t, player))
	respondReq.Header.Set("Content-Type", "application/json")
	respondRec := httptest.NewRecorder()
	router.ServeHTTP(respondRec, respondReq)
	if respondRec.Code != http.StatusOK {
		t.Fatalf("respond status = %d, body = %s", respondRec.Code, respondRec.Body.String())
	}

	var updated models.TrialInvite
	if err := db.First(&updated, invite.ID).Error; err != nil {
		t.Fatalf("find updated invite: %v", err)
	}
	if updated.Status != models.TrialInviteAccepted || updated.RespondedAt == nil {
		t.Fatalf("unexpected updated invite: %+v", updated)
	}
	var senderNotificationCount int64
	db.Model(&models.Notification{}).Where("user_id = ? AND type = ?", sender.ID, models.NotificationTypeTrialInvite).Count(&senderNotificationCount)
	if senderNotificationCount != 1 {
		t.Fatalf("sender trial invite notifications = %d, want 1", senderNotificationCount)
	}
}

func itoaTest(n int) string {
	if n == 0 {
		return "0"
	}
	out := ""
	for n > 0 {
		out = string(rune('0'+n%10)) + out
		n /= 10
	}
	return out
}
