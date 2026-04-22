package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"go-chat/internal/logger"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	OfflineMsgPrefix = "offline:msg:"
	OfflineMsgTTL    = 7 * 24 * time.Hour
)

var Client *redis.Client
var ctx = context.Background()

type OfflineMessage struct {
	FromUserID uint   `json:"from_user_id"`
	GroupID    uint   `json:"group_id,omitempty"`
	Content    string `json:"content"`
	MsgType    string `json:"msg_type"`
	CreatedAt  int64  `json:"created_at"`
}

// 初始化redis连接
func InitRedis(host string, port int, password string, db int) {
	Client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       db,
	})

	if err := Client.Ping(ctx).Err(); err != nil {
		logger.Logger.Fatal("Redis 连接失败", zap.Error(err))
	}

	logger.Logger.Info("Redis 连接成功")

}

// 关闭连接
func CloseRedis() {
	if Client != nil {
		Client.Close()
	}
}

// 保存离线消息
func SaveOfflineMessage(userID uint, msg *OfflineMessage) error {
	key := fmt.Sprintf("%s%d", OfflineMsgPrefix, userID)

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if err := Client.RPush(ctx, key, data).Err(); err != nil {
		return err
	}

	Client.Expire(ctx, key, OfflineMsgTTL)

	return nil
}

// 获取并清空离线消息
func GetOfflineMessage(userID uint) ([]OfflineMessage, error) {
	key := fmt.Sprintf("%s%d", OfflineMsgPrefix, userID)

	dataList, err := Client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	if len(dataList) == 0 {
		return nil, nil
	}

	messages := make([]OfflineMessage, 0, len(dataList))
	for _, data := range dataList {
		var msg OfflineMessage
		if err := json.Unmarshal([]byte(data), &msg); err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	Client.Del(ctx, key)

	return messages, nil

}
