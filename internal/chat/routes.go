package chat

import (
	"go-chat/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, handler *Handler) {
	chatGroup := r.Group("/chat")
	chatGroup.Use(middleware.Auth())
	{
		chatGroup.GET("/ws", WebSocketHandler)
		chatGroup.POST("/read", handler.MarkSingleRead)
		chatGroup.GET("/unread", handler.GetSingleUnread)
		chatGroup.GET("/conversations", handler.GetConversations)
		chatGroup.GET("/messages", handler.GetMessages)
	}
}
