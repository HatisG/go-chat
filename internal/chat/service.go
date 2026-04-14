package chat

import (
	"errors"
	"go-chat/internal/friend"
	"go-chat/internal/message"
	"time"
)

type Service struct {
	hub         *Hub
	friendRepo  friend.Repository
	messageRepo Repository
}

func NewService(hub *Hub, friendRepo friend.Repository, messageRepo Repository) *Service {
	return &Service{
		hub:         hub,
		friendRepo:  friendRepo,
		messageRepo: messageRepo,
	}
}

func (s *Service) SendMessage(fromUserID, toUserID uint, content string) error {
	//校验好友关系
	//压测期间临时注释，测试全链路性能
	// _, err := s.friendRepo.FindFriendship(fromUserID, toUserID)
	// if err != nil {
	// 	return errors.New("不是好友,无法发送消息")
	// }

	//模拟io延迟5ms，仅测试
	time.Sleep(5 * time.Millisecond)

	//投递到rabbitmq
	chatMsg := &message.ChatMessage{
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		Content:    content,
		MsgType:    "text",
	}
	if err := message.PublishMessage(chatMsg); err != nil {
		return errors.New("消息投递失败")
	}

	//处理离线消息

	//发送给在线用户
	s.hub.Broadcast <- &WSMessage{
		Type:     "chat",
		ToUserID: toUserID,
		Content:  content,
	}
	return nil

}
