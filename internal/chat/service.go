package chat

import (
	"errors"
	"go-chat/internal/cache"
	"go-chat/internal/config"
	"go-chat/internal/friend"
	"go-chat/internal/group"
	"sort"

	"go-chat/internal/message"
	"time"
)

type Service struct {
	hub          *Hub
	friendRepo   friend.Repository
	messageRepo  Repository
	groupService *group.Service
	groupRepo    group.Repository
}

type Conversation struct {
	PeerID      uint      `json:"peer_id"`
	PeerName    string    `json:"peer_name"`
	Avatar      string    `json:"avatar"`
	ConvType    int       `json:"conv_type"` // 0:单聊 1:群聊
	LastMsg     string    `json:"last_msg"`
	LastTime    time.Time `json:"last_time"`
	UnreadCount int       `json:"unread_count"`
}

func NewService(hub *Hub, friendRepo friend.Repository, messageRepo Repository, groupService *group.Service, groupRepo group.Repository) *Service {
	return &Service{
		hub:          hub,
		friendRepo:   friendRepo,
		messageRepo:  messageRepo,
		groupService: groupService,
		groupRepo:    groupRepo,
	}
}

func (s *Service) SendMessage(fromUserID, toUserID uint, msgType, content string) error {
	//校验好友关系
	//压测期间临时注释，测试全链路性能
	_, err := s.friendRepo.FindFriendship(fromUserID, toUserID)
	if err != nil {
		return errors.New("不是好友,无法发送消息")
	}

	//模拟io延迟5ms，仅测试
	//time.Sleep(5 * time.Millisecond)

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
		Type:       "chat",
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		MsgType:    msgType,
		Content:    content,
	}
	return nil

}

// 发送群消息
func (s *Service) SendGroupMessage(groupID, fromUserID uint, msgType, content string) error {
	return s.groupService.SendGroupMessage(groupID, fromUserID, msgType, content)
}

// 是否在线
func (h *Hub) IsOnline(userID uint) bool {
	hub.mu.RLock()
	defer hub.mu.RUnlock()
	_, ok := hub.Clients[userID]
	return ok
}

// MarkSingleChatRead 标记单聊已读
func (s *Service) MarkSingleChatRead(userID, peerID, lastMsgID uint) error {
	return s.messageRepo.UpdateReadCursor(userID, peerID, lastMsgID)
}

// GetSingleChatUnread 获取单聊未读数
func (s *Service) GetSingleChatUnread(userID, peerID uint) (int64, error) {
	return s.messageRepo.GetUnreadCount(userID, peerID)
}

func (s *Service) GetConversations(userID uint) ([]Conversation, error) {
	var conversations []Conversation

	// 1. 单聊会话
	friendships, err := s.friendRepo.FindFriendsByUserID(userID)
	if err != nil {
		return nil, err
	}
	for _, f := range friendships {
		peerID := f.FriendID
		if f.UserID != userID {
			peerID = f.UserID
		}

		// 查最后一条消息
		msgs, _ := s.messageRepo.FindConversation(userID, peerID, 1, 0)
		lastMsg := ""
		lastTime := time.Time{}
		if len(msgs) > 0 {
			lastMsg = msgs[0].Content
			lastTime = msgs[0].CreatedAt
		}

		// 查未读数
		count, _ := s.messageRepo.GetUnreadCount(userID, peerID)
		// 查对方昵称
		var peerUser struct {
			Nickname string
			Avatar   string
		}
		config.DB.Table("users").Where("id = ?", peerID).Select("nickname", "avatar").Scan(&peerUser)

		conversations = append(conversations, Conversation{
			PeerID:      peerID,
			PeerName:    peerUser.Nickname,
			Avatar:      peerUser.Avatar,
			ConvType:    0,
			LastMsg:     lastMsg,
			LastTime:    lastTime,
			UnreadCount: int(count),
		})
	}

	// 2. 群聊会话
	groups, err := s.groupRepo.FindGroupsByUserID(userID)
	if err != nil {
		return nil, err
	}
	for _, g := range groups {
		// 查最后一条消息
		msgs, _ := s.groupRepo.FindMessagesByGroupID(g.ID, 1, 0)
		lastMsg := ""
		lastTime := time.Time{}
		if len(msgs) > 0 {
			lastMsg = msgs[0].Content
			lastTime = msgs[0].CreatedAt
		}

		// 查未读数
		count, _ := s.groupRepo.GetGroupUnreadCount(userID, g.ID)

		conversations = append(conversations, Conversation{
			PeerID:      g.ID,
			PeerName:    g.Name,
			ConvType:    1,
			LastMsg:     lastMsg,
			LastTime:    lastTime,
			UnreadCount: count,
		})
	}

	// 3. 按最后消息时间降序排序
	sort.Slice(conversations, func(i, j int) bool {
		return conversations[i].LastTime.After(conversations[j].LastTime)
	})

	return conversations, nil
}

func (s *Service) GetConversationMessages(userID, peerID, cursor uint, limit int) ([]Message, error) {
	return s.messageRepo.FindConversation(userID, peerID, limit+1, 0)
}
