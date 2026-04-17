package chat

import (
	"go-chat/internal/friend"
	"go-chat/internal/group"

	"github.com/gin-gonic/gin"
)

var (
	hub     *Hub
	service *Service
)

func Init(friendRepo friend.Repository, messageRepo Repository, groupService *group.Service) {
	hub = NewHub()
	service = NewService(hub, friendRepo, messageRepo, groupService)
	go hub.Run()
}

func WebSocketHandler(c *gin.Context) {
	userID := c.GetUint("user_id")
	ServerWS(hub, service, userID, c.Writer, c.Request)
}

func GetHub() *Hub {
	return hub
}
