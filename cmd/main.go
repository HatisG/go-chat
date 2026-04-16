package main

import (
	"fmt"
	"go-chat/internal/cache"
	"go-chat/internal/chat"
	"go-chat/internal/config"
	"go-chat/internal/friend"
	"go-chat/internal/message"
	"go-chat/internal/middleware"
	"go-chat/internal/upload"
	"go-chat/internal/user"
	"go-chat/pkg/response"

	"github.com/gin-gonic/gin"
)

func main() {
	//读取配置
	config.Load()
	cfg := config.AppConfig

	//初始化数据库
	config.InitDB()
	config.DB.AutoMigrate(&user.User{},
		&friend.Friendship{},
		&friend.FriendRequest{},
		&chat.Message{})

	//初始化消息队列
	rabbitmqURL := fmt.Sprintf("amqp://%s:%s@%s:%d/",
		config.AppConfig.RabbitMQ.User,
		config.AppConfig.RabbitMQ.Password,
		config.AppConfig.RabbitMQ.Host,
		config.AppConfig.RabbitMQ.Port,
	)
	message.InitMQ(rabbitmqURL)
	defer message.CloseMQ()

	//初始化redis
	cache.InitRedis(
		config.AppConfig.Redis.Host,
		config.AppConfig.Redis.Port,
		config.AppConfig.Redis.Password,
		config.AppConfig.Redis.DB,
	)
	defer cache.CloseRedis()

	gin.SetMode(cfg.Server.Mode)
	r := gin.Default()

	userRepo := user.NewRepository(config.DB)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	friendRepo := friend.NewRepository(config.DB)
	friendService := friend.NewService(friendRepo)
	friendHandler := friend.NewHandler(friendService)

	chatRepo := chat.NewRepository(config.DB)
	chat.Init(friendRepo, chatRepo)

	uploadHandler := upload.NewHandler()

	r.Static("/uploads", "./uploads")

	message.StartConsumer(func(msg *message.ChatMessage) error {
		dbMsg := &chat.Message{
			FromUserID: msg.FromUserID,
			ToUserID:   msg.ToUserID,
			Content:    msg.Content,
			MsgType:    msg.MsgType,
			IsRead:     false,
		}
		return chatRepo.Create(dbMsg)
	})

	api := r.Group("/api/v1")
	user.RegisterRouts(api, userHandler)
	friend.RegisterRountes(api, friendHandler)
	chat.RegisterRoutes(api)
	upload.RegisterRoutes(api, uploadHandler)

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
