package message

import (
	"go-chat/internal/logger"

	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
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
		logger.Logger.Fatal("RabbitMQ连接失败", zap.Error(err))
	}

	//创建channel
	Channel, err = Conn.Channel()
	if err != nil {
		logger.Logger.Fatal("RabbitMQ Channel 创建失败: ", zap.Error(err))
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
		logger.Logger.Fatal("RabbitMQ 队列声明失败", zap.Error(err))
	}

	logger.Logger.Info("RabbitMQ 初始化成功")
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
