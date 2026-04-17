package group

import (
	"go-chat/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, handler *Handler) {
	groupRoute := r.Group("/group")
	groupRoute.Use(middleware.Auth())
	{
		groupRoute.POST("", handler.CreateGroup)                // 创建群
		groupRoute.POST("/:id/join", handler.JoinGroup)         // 加入群
		groupRoute.POST("/:id/leave", handler.LeaveGroup)       // 退出群
		groupRoute.GET("/my", handler.GetMyGroups)              // 我的群列表
		groupRoute.GET("/:id/members", handler.GetGroupMembers) // 群成员
	}

}
