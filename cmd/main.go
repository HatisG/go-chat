package main

import (
	"fmt"
	"go-chat/internal/cache"
	"go-chat/internal/chat"
	"go-chat/internal/config"
	"go-chat/internal/friend"
	"go-chat/internal/group"
	"go-chat/internal/logger"
	"go-chat/internal/message"
	"go-chat/internal/upload"
	"go-chat/internal/user"
	"go-chat/pkg/response"

	"github.com/gin-gonic/gin"
)

func main() {
	//读取配置
	config.Load()
	cfg := config.AppConfig

	//初始化日志
	logger.Init(cfg.Server.Mode)
	defer logger.Sync()

	//初始化数据库
	config.InitDB()
	config.DB.AutoMigrate(
		&user.User{},
		&friend.Friendship{},
		&friend.FriendRequest{},
		&chat.Message{},
		&group.Group{},
		&group.GroupMember{},
		&group.GroupMessage{},
		&group.UnreadCount{},
		&chat.ReadCursor{},
	)

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

	//启动服务
	gin.SetMode(cfg.Server.Mode)
	r := gin.Default()

	userRepo := user.NewRepository(config.DB)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	friendRepo := friend.NewRepository(config.DB)
	friendService := friend.NewService(friendRepo)
	friendHandler := friend.NewHandler(friendService)

	chatRepo := chat.NewRepository(config.DB)
	groupRepo := group.NewRepository(config.DB)
	groupService := group.NewService(groupRepo, chat.GetHub())
	groupHandler := group.NewHandler(groupService)

	chatHandler := chat.Init(friendRepo, chatRepo, groupService, groupRepo)

	chat.SetGroupService(groupService)

	uploadHandler := upload.NewHandler()

	r.Static("/uploads", "./uploads")

	//消费者并发
	for i := 0; i < 3; i++ {
		go message.StartConsumer(func(msg *message.ChatMessage) error {
			if msg.GroupID > 0 {
				//群消息
				groupMsg := &group.GroupMessage{
					GroupID:    msg.GroupID,
					FromUserID: msg.FromUserID,
					Content:    msg.Content,
					MsgType:    msg.MsgType,
				}
				return groupRepo.SaveMessage(groupMsg)
			} else {
				//单聊消息
				dbMsg := &chat.Message{
					FromUserID: msg.FromUserID,
					ToUserID:   msg.ToUserID,
					Content:    msg.Content,
					MsgType:    msg.MsgType,
					IsRead:     false,
				}
				return chatRepo.Create(dbMsg)

			}

		})
	}

	api := r.Group("/api/v1")
	user.RegisterRouts(api, userHandler)
	friend.RegisterRountes(api, friendHandler)
	chat.RegisterRoutes(api, chatHandler)
	upload.RegisterRoutes(api, uploadHandler)
	group.RegisterRoutes(api, groupHandler)

	r.GET("/ping", func(ctx *gin.Context) {
		response.Success(ctx, gin.H{"message": "pong"})
	})

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	r.Run(addr)

}
