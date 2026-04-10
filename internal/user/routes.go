package user

import "github.com/gin-gonic/gin"

func RegisterRouts(r *gin.RouterGroup, handler *Handler) {
	userGroup := r.Group("/user")
	{
		userGroup.POST("/register", handler.Register)
		userGroup.POST("/login", handler.Login)
	}
}
