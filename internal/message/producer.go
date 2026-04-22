package message

import (
	"encoding/json"
	"go-chat/internal/logger"

	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// ChatMessage 消息结构体（用于 MQ 传输）
type ChatMessage struct {
	FromUserID uint   `json:"from_user_id"`
	ToUserID   uint   `json:"to_user_id"`
	Content    string `json:"content"`
	MsgType    string `json:"msg_type"`
}

// 投递消息到队列
func PublishMessage(msg *ChatMessage) error {
	//序列化消息
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	//发布消息
	err = Channel.Publish(
		"",
		QueueName,
		false,
		false,
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp091.Persistent,
		},
	)
	if err != nil {
		return err
	}

	logger.Logger.Info("producer 消息已投递", zap.Uint("from_user_id", msg.FromUserID), zap.Uint("to_user_id", msg.ToUserID))

	return nil
}
