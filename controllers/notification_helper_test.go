package controllers

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/wshub"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestNotificationHelperPushesWebSocketPayload(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:notification-helper?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Notification{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	hub := wshub.NewHub()
	go hub.Run()
	wshub.GetNotifyService().SetHub(hub)
	t.Cleanup(func() {
		wshub.GetNotifyService().SetHub(nil)
	})

	client := &wshub.Client{
		Hub:    hub,
		Send:   make(chan []byte, 1),
		UserID: 9201,
	}
	hub.Register <- client

	helper := NewNotificationHelper(db)
	if err := helper.NotifyWeeklyReportApproved(9201, "王教练", 5, 8101); err != nil {
		t.Fatalf("create notification: %v", err)
	}

	select {
	case raw := <-client.Send:
		var message struct {
			Type    string `json:"type"`
			Content struct {
				ID        uint                   `json:"id"`
				Type      string                 `json:"type"`
				Title     string                 `json:"title"`
				Content   string                 `json:"content"`
				Data      map[string]interface{} `json:"data"`
				CreatedAt string                 `json:"created_at"`
			} `json:"content"`
		}
		if err := json.Unmarshal(raw, &message); err != nil {
			t.Fatalf("decode websocket payload: %v, raw=%s", err, string(raw))
		}
		if message.Type != "notification" {
			t.Fatalf("message type = %q, want notification", message.Type)
		}
		if message.Content.Type != string(models.NotificationTypeWeeklyReportApproved) {
			t.Fatalf("notification type = %q", message.Content.Type)
		}
		if message.Content.ID == 0 {
			t.Fatalf("notification id was not included")
		}
		if message.Content.Data["link"] != "/weekly-reports/8101" {
			t.Fatalf("payload link = %#v", message.Content.Data["link"])
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for websocket notification")
	}
}
