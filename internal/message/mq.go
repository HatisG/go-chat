package message

import (
	"log"

	"github.com/rabbitmq/amqp091-go"
)

const (
	QueueName = "chat_queue"
)

var Conn *amqp091.Connection
var Channel *amqp091.Channel

// 初始化连接rabbitmq
func InitMQ(url string) {
	var err error

	//连接rabbitmq
	Conn, err = amqp091.Dial(url)
	if err != nil {
		log.Fatalf("RabbitMQ连接失败: %v", err)
	}

	//创建channel
	Channel, err = Conn.Channel()
	if err != nil {
		log.Fatalf("RabbitMQ Channel 创建失败: %v", err)
	}

	//声明队列
	_, err = Channel.QueueDeclare(
		QueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("RabbitMQ 队列声明失败: %v", err)
	}

	log.Println("RabbitMQ 初始化成功")
}

// 关闭连接
func CloseMQ() {
	if Channel != nil {
		Channel.Close()
	}
	if Conn != nil {
		Conn.Close()
	}
}
