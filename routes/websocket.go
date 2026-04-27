package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/wshub"
)

// SetupWebSocketRoutes 设置 WebSocket 路由
func SetupWebSocketRoutes(r *gin.Engine, hub *wshub.Hub) {
	// WebSocket 客户端处理器
	clientHandler := wshub.NewClientHandler(hub)

	// WebSocket 连接端点（需要认证）
	r.GET("/ws", clientHandler.HandleWebSocket)
}
