package message

import (
	"encoding/json"
	"log"
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
		log.Fatalf("RabbitMQ 消费者启动失败: %v", err)
	}

	log.Println("RabbitMQ 消费者已启动")

	go func() {
		for msg := range msgs {
			var chatMsg ChatMessage
			if err := json.Unmarshal(msg.Body, &chatMsg); err != nil {
				log.Printf("consumer 消息解析失败: %v", err)
				msg.Nack(false, true) //重新入队
				continue
			}

			if err := handler(&chatMsg); err != nil {
				log.Printf("consumer 处理失败: %v", err)
				msg.Nack(false, true)
			} else {
				msg.Ack(false)
				log.Printf("consumer 消息已处理: from=%d to=%d", chatMsg.FromUserID, chatMsg.ToUserID)
			}
		}

	}()

}
