package chat

import (
	"go-chat/internal/friend"

	"github.com/gin-gonic/gin"
)

var (
	hub     *Hub
	service *Service
)

func Init(friendRepo friend.Repository) {
	hub = NewHub()
	service = NewService(hub, friendRepo)
	go hub.Run()
}

func WebSocketHandler(c *gin.Context) {
	userID := c.GetUint("user_id")
	ServerWS(hub, service, userID, c.Writer, c.Request)
}
