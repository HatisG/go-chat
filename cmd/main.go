package main

import (
	"fmt"
	"go-chat/internal/config"
	"go-chat/internal/middleware"
	"go-chat/internal/user"
	"go-chat/pkg/response"

	"github.com/gin-gonic/gin"
)

func main() {

	config.Load()
	cfg := config.AppConfig

	config.InitDB()
	config.DB.AutoMigrate(&user.User{})

	gin.SetMode(cfg.Server.Mode)
	r := gin.Default()

	userRepo := user.NewRepository(config.DB)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	api := r.Group("/api/v1")
	user.RegisterRouts(api, userHandler)

	authApi := r.Group("/api/v1")
	authApi.Use(middleware.Auth())
	{
		authApi.GET("/profile", func(c *gin.Context) {
			userID, _ := c.Get("user_id")
			username, _ := c.Get("username")
			response.Success(c, gin.H{
				"user_id":  userID,
				"username": username,
			})
		})
	}

	r.GET("/ping", func(ctx *gin.Context) {
		response.Success(ctx, gin.H{"message": "pong"})
	})

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	r.Run(addr)

}
