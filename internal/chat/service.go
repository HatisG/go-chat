package chat

import (
	"errors"
	"go-chat/internal/cache"
	"go-chat/internal/friend"
	"go-chat/internal/group"

	"go-chat/internal/message"
	"time"
)

type Service struct {
	hub          *Hub
	friendRepo   friend.Repository
	messageRepo  Repository
	groupService *group.Service
}

func NewService(hub *Hub, friendRepo friend.Repository, messageRepo Repository, groupService *group.Service) *Service {
	return &Service{
		hub:          hub,
		friendRepo:   friendRepo,
		messageRepo:  messageRepo,
		groupService: groupService,
	}
}

func (s *Service) SendMessage(fromUserID, toUserID uint, msgType, content string) error {
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
		MsgType:    msgType,
	}
	if err := message.PublishMessage(chatMsg); err != nil {
		return errors.New("消息投递失败")
	}

	//处理离线消息
	s.hub.mu.RLock()
	_, online := s.hub.Clients[toUserID]
	s.hub.mu.RUnlock()

	if !online {
		offlineMsg := &cache.OfflineMessage{
			FromUserID: fromUserID,
			Content:    content,
			MsgType:    "text",
			CreatedAt:  time.Now().Unix(),
		}
		cache.SaveOfflineMessage(toUserID, offlineMsg)
	}

	//发送给在线用户
	s.hub.Broadcast <- &WSMessage{
		Type:     "chat",
		ToUserID: toUserID,
		MsgType:  msgType,
		Content:  content,
	}
	return nil

}

func (s *Service) SendGroupMessage(groupID, fromUserID uint, msgType, content string) error {
	return s.groupService.SendGroupMessage(groupID, fromUserID, msgType, content)
}

func (h *Hub) IsOnline(userID uint) bool {
	hub.mu.RLock()
	defer hub.mu.RUnlock()
	_, ok := hub.Clients[userID]
	return ok
}
