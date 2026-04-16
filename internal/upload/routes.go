package upload

import (
	"go-chat/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, handler *Handler) {
	uploadGroup := r.Group("/upload")
	uploadGroup.Use(middleware.Auth())
	{
		uploadGroup.POST("", handler.Upload)
	}

}
