package message

import (
	"encoding/json"
	"log"

	"github.com/rabbitmq/amqp091-go"
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

	log.Printf("producer 消息已投递: from=%d to=%d", msg.FromUserID, msg.ToUserID)

	return nil
}
