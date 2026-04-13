package chat

import (
	"go-chat/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	chatGroup := r.Group("/chat")
	chatGroup.Use(middleware.Auth())
	{
		chatGroup.GET("/ws", WebSocketHandler)
	}
}
