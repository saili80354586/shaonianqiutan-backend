package wshub

import (
	"log"
	"sync"
)

// NotifyService 通知服务
type NotifyService struct {
	hub *Hub
	mu  sync.RWMutex
}

var notifyService *NotifyService
var once sync.Once

// GetNotifyService 获取通知服务单例
func GetNotifyService() *NotifyService {
	once.Do(func() {
		notifyService = &NotifyService{}
	})
	return notifyService
}

// SetHub 设置Hub
func (s *NotifyService) SetHub(hub *Hub) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hub = hub
}

// SendNotification 发送通知给用户
func (s *NotifyService) SendNotification(userID uint, notificationType string, payload NotificationPayload) {
	s.mu.RLock()
	hub := s.hub
	s.mu.RUnlock()

	if hub == nil {
		log.Printf("WebSocket Hub 未初始化，跳过推送")
		return
	}

	// 标记通知类型
	payload.Type = notificationType

	hub.SendToUser(userID, "notification", payload)
	log.Printf("WebSocket: 发送通知给用户 %d, 类型: %s", userID, notificationType)
}

// SendToMultiple 发送通知给多个用户
func (s *NotifyService) SendToMultiple(userIDs []uint, notificationType string, payload NotificationPayload) {
	s.mu.RLock()
	hub := s.hub
	s.mu.RUnlock()

	if hub == nil {
		log.Printf("WebSocket Hub 未初始化，跳过推送")
		return
	}

	payload.Type = notificationType

	for _, userID := range userIDs {
		hub.SendToUser(userID, "notification", payload)
	}
	log.Printf("WebSocket: 发送通知给 %d 个用户, 类型: %s", len(userIDs), notificationType)
}

// BroadcastSystem 广播系统消息
func (s *NotifyService) BroadcastSystem(message string) {
	s.mu.RLock()
	hub := s.hub
	s.mu.RUnlock()

	if hub == nil {
		return
	}

	payload := NotificationPayload{
		Title:   "系统通知",
		Content: message,
	}
	hub.Broadcast("system", payload)
}
