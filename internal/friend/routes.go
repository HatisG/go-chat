package friend

import (
	"go-chat/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRountes(r *gin.RouterGroup, handler *Handler) {
	friendGroup := r.Group("/friend")
	friendGroup.Use(middleware.Auth())
	{
		friendGroup.POST("/request", handler.SendRequest)
		friendGroup.POST("/request/:id/accept", handler.AcceptRequest)
		friendGroup.POST("/request/:id/reject", handler.RejectRequest)
		friendGroup.GET("/requests", handler.GetPendingRequests)
		friendGroup.GET("/friends", handler.GetFriendList)
		friendGroup.DELETE("/:id", handler.DeleteFriend)
	}

}
