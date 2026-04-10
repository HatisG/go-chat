package main

import (
	"fmt"
	"go-chat/internal/config"
	"go-chat/pkg/response"
	"os/user"

	"github.com/gin-gonic/gin"
)

func main() {

	config.Load()
	cfg := config.AppConfig

	config.InitDB()
	config.DB.AutoMigrate(&user.User{})

	gin.SetMode(cfg.Server.Mode)
	r := gin.Default()

	r.GET("/ping", func(ctx *gin.Context) {
		response.Success(ctx, gin.H{"message": "pong"})
	})

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	r.Run(addr)

}
