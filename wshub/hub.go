package wshub

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Client 连接的客户端
type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan []byte
	UserID   uint
	Username string
}

// Hub 中心管理器
type Hub struct {
	// 注册的客户端
	clients map[uint]*Client

	// 用户ID到客户端的映射（支持多设备）
	userClients map[uint][]*Client

	// 注册请求（导出）
	Register chan *Client

	// 注销请求（导出）
	Unregister chan *Client

	// 广播消息
	broadcast chan []byte

	// 发送给指定用户
	sendToUser chan *UserMessage

	// 锁
	mu sync.RWMutex
}

// UserMessage 发送给用户的消息
type UserMessage struct {
	UserID uint
	Data   []byte
}

// Message 消息结构
type Message struct {
	Type    string          `json:"type"`    // notification, chat, system
	Content json.RawMessage `json:"content"`
}

// NotificationPayload 通知消息内容
type NotificationPayload struct {
	ID        uint   `json:"id"`
	Type      string `json:"type"` // weekly_report, match_summary, order, system
	Title     string `json:"title"`
	Content   string `json:"content"`
	Data      any    `json:"data,omitempty"`
	CreatedAt string `json:"created_at"`
}

// NewHub 创建新的Hub
func NewHub() *Hub {
	return &Hub{
		clients:     make(map[uint]*Client),
		userClients: make(map[uint][]*Client),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		broadcast:   make(chan []byte),
		sendToUser:  make(chan *UserMessage, 100),
	}
}

// Run 启动Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client.UserID] = client
			h.userClients[client.UserID] = append(h.userClients[client.UserID], client)
			h.mu.Unlock()
			log.Printf("WebSocket: 用户 %d 已连接 (共 %d 个客户端)", client.UserID, len(h.userClients[client.UserID]))

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				// 从用户客户端列表中移除
				clients := h.userClients[client.UserID]
				for i, c := range clients {
					if c == client {
						h.userClients[client.UserID] = append(clients[:i], clients[i+1:]...)
						break
					}
				}
				if len(h.userClients[client.UserID]) == 0 {
					delete(h.userClients, client.UserID)
				}
			}
			h.mu.Unlock()
			log.Printf("WebSocket: 用户 %d 已断开连接", client.UserID)

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client.UserID)
				}
			}
			h.mu.RUnlock()

		case um := <-h.sendToUser:
			h.mu.RLock()
			clients := h.userClients[um.UserID]
			for _, client := range clients {
				select {
				case client.Send <- um.Data:
				default:
					// 客户端缓冲区满，跳过
				}
			}
			h.mu.RUnlock()
		}
	}
}

// SendToUser 发送消息给指定用户
func (h *Hub) SendToUser(userID uint, msgType string, payload any) {
	data := Message{
		Type:    msgType,
		Content: mustMarshal(payload),
	}
	msg := UserMessage{
		UserID: userID,
		Data:   mustMarshal(data),
	}
	h.sendToUser <- &msg
}

// Broadcast 广播消息给所有用户
func (h *Hub) Broadcast(msgType string, payload any) {
	data := Message{
		Type:    msgType,
		Content: mustMarshal(payload),
	}
	h.broadcast <- mustMarshal(data)
}

// GetOnlineUsers 获取在线用户数
func (h *Hub) GetOnlineUsers() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetUserClientCount 获取用户连接数
func (h *Hub) GetUserClientCount(userID uint) int {
	h.mu.RLock()
	defer h.mu.Unlock()
	return len(h.userClients[userID])
}

func mustMarshal(v any) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("WebSocket marshal error: %v", err)
		return []byte(`{"type":"error","content":{}}`)
	}
	return data
}
