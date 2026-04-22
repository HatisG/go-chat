package message

import (
	"encoding/json"
	"go-chat/internal/logger"

	"go.uber.org/zap"
)

type MessageHandler func(msg *ChatMessage) error

// 启动消费者，异步写入数据库
func StartConsumer(handler MessageHandler) {

	msgs, err := Channel.Consume(
		QueueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logger.Logger.Fatal("RabbitMQ 消费者启动失败", zap.Error(err))
	}

	logger.Logger.Info("RabbitMQ 消费者已启动")

	go func() {
		for msg := range msgs {
			var chatMsg ChatMessage
			if err := json.Unmarshal(msg.Body, &chatMsg); err != nil {
				logger.Logger.Info("consumer 消息解析失败", zap.Error(err))
				msg.Nack(false, true) //重新入队
				continue
			}

			if err := handler(&chatMsg); err != nil {
				logger.Logger.Info("consumer 处理失败", zap.Error(err))
				msg.Nack(false, true)
			} else {
				msg.Ack(false)
				logger.Logger.Info("consumer 消息已处理", zap.Uint("from_user_id", chatMsg.FromUserID), zap.Uint("to_user_id", chatMsg.ToUserID))
			}
		}

	}()

}
