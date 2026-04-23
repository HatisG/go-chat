package user

import (
	"go-chat/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRouts(r *gin.RouterGroup, handler *Handler) {
	userGroup := r.Group("/user")
	{
		userGroup.POST("/register", handler.Register)
		userGroup.POST("/login", handler.Login)
	}

	authGroup := userGroup.Group("")
	authGroup.Use(middleware.Auth())
	{
		authGroup.GET("/profile", handler.GetProfile)
		authGroup.PUT("/profile", handler.UpdateProfile)
		authGroup.PUT("/avatar", handler.UpdateAvatar)
		authGroup.PUT("/password", handler.ChangePassword)
	}

}
