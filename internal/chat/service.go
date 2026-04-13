package chat

import (
	"errors"
	"go-chat/internal/friend"
)

type Service struct {
	hub         *Hub
	friendRepo  friend.Repository
	messageRepo Repository
}

func NewService(hub *Hub, friendRepo friend.Repository) *Service {
	return &Service{
		hub:        hub,
		friendRepo: friendRepo,
	}
}

func (s *Service) SendMessage(fromUserID, toUserID uint, content string) error {
	//校验好友关系
	_, err := s.friendRepo.FindFriendship(fromUserID, toUserID)
	if err != nil {
		return errors.New("不是好友,无法发送消息")
	}

	//存储信息

	//处理离线消息

	//发送给在线用户
	s.hub.Broadcast <- &WSMessage{
		Type:     "chat",
		ToUserID: toUserID,
		Content:  content,
	}
	return nil

}
