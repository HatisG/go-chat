package group

import (
	"errors"
	"go-chat/internal/cache"
	"go-chat/internal/config"
	"go-chat/internal/message"
	"go-chat/internal/user"
	"time"
)

type HubInterface interface {
	IsOnline(userID uint) bool
	SendToUser(msg interface{}) // 广播消息
}

type GroupWSMessage struct {
	Type         string `json:"type"`
	ToUserID     uint   `json:"to_user_id"`
	GroupID      uint   `json:"group_id"`
	FromUserID   uint   `json:"from_user_id"`
	FromUserName string `json:"from_user_name"`
	MsgType      string `json:"msg_type"`
	Content      string `json:"content"`
}

type Service struct {
	repo        Repository
	hub         HubInterface      // 依赖接口，不依赖具体实现
	messageRepo MessageRepository // 也需要定义接口
}

type MessageRepository interface {
	SaveGroupMessage(msg *GroupMessage) error
}

func NewService(repo Repository, hub HubInterface) *Service {
	return &Service{
		repo: repo,
		hub:  hub,
	}
}

// 创建群组
func (s *Service) CreateGroup(creatorID uint, name string) (*Group, error) {
	group := &Group{
		Name:      name,
		CreatorID: creatorID,
	}

	if err := s.repo.Create(group); err != nil {
		return nil, errors.New("创建群聊失败")
	}

	member := &GroupMember{
		GroupID: group.ID,
		UserID:  creatorID,
		Role:    RoleOwner,
	}

	if err := s.repo.AddMember(member); err != nil {
		return nil, errors.New("添加群主失败")
	}

	return group, nil

}

// 加入群
func (s *Service) JoinGroup(groupID, userID uint) error {

	_, err := s.repo.FindByID(groupID)
	if err != nil {
		return errors.New("群不存在")
	}

	_, err = s.repo.FindMember(groupID, userID)
	if err == nil {
		return errors.New("已在该群成员")
	}

	member := &GroupMember{
		GroupID: groupID,
		UserID:  userID,
		Role:    RoleMember,
	}

	return s.repo.AddMember(member)

}

// 离开群
func (s *Service) LeaveGroup(groupID, userID uint) error {

	member, err := s.repo.FindMember(groupID, userID)
	if err != nil {
		return errors.New("不是群成员")
	}

	//群主退群为解散群
	if member.Role == RoleOwner {
		return s.repo.Delete(groupID)
	}

	return s.repo.RemoveMember(groupID, userID)
}

// 获取群组
func (s *Service) GetMyGroups(userID uint) ([]GroupInfo, error) {
	groups, err := s.repo.FindGroupsByUserID(userID)
	if err != nil {
		return nil, errors.New("群列表获取失败")
	}

	result := make([]GroupInfo, 0, len(groups))
	for _, g := range groups {
		count, _ := s.repo.CountMembers(g.ID)
		member, _ := s.repo.FindMember(g.ID, userID)
		result = append(result, GroupInfo{
			ID:          g.ID,
			Name:        g.Name,
			Avatar:      g.Avatar,
			CreatorID:   g.CreatorID,
			MemberCount: count,
			Role:        member.Role,
		})

	}
	return result, nil
}

// 获取群成员
func (s *Service) GetGroupMembers(groupID uint) ([]MemberInfo, error) {
	members, err := s.repo.FindMembersByGroupID(groupID)
	if err != nil {
		return nil, errors.New("群成员获取失败")
	}

	result := make([]MemberInfo, 0, len(members))

	for _, m := range members {

		memberInfo, _ := s.getMemberInfo(m.UserID) // 需要新增辅助方法
		result = append(result, MemberInfo{
			UserID:   m.UserID,
			Username: memberInfo.Username,
			Nickname: memberInfo.Nickname,
			Avatar:   memberInfo.Avatar,
			Role:     m.Role,
		})

	}

	return result, nil
}

// 发送群消息
func (s *Service) SendGroupMessage(groupID, fromUserID uint, msgType, content string) error {

	_, err := s.repo.FindMember(groupID, fromUserID)
	if err != nil {
		return errors.New("不是该群成员")
	}

	//模拟io延迟，测试用
	//time.Sleep(5 * time.Millisecond)

	//投递消息队列
	chatMsg := &message.ChatMessage{
		FromUserID: fromUserID,
		ToUserID:   0,
		GroupID:    groupID,
		Content:    content,
		MsgType:    msgType,
	}
	message.PublishMessage(chatMsg)

	//投递消息
	members, err := s.repo.FindMembersByGroupID(groupID)
	if err != nil {
		return err
	}
	for _, m := range members {
		if m.UserID == fromUserID {
			continue
		}
		//在线发送，离线投到redis
		if s.hub.IsOnline(m.UserID) {
			s.hub.SendToUser(&GroupWSMessage{
				Type:         "group_chat",
				ToUserID:     m.UserID,
				GroupID:      groupID,
				FromUserID:   fromUserID,
				FromUserName: s.getSenderNickname(fromUserID), // 需要新增辅助方法
				MsgType:      msgType,
				Content:      content,
			})
		} else {
			offlineMsg := cache.OfflineMessage{
				FromUserID: fromUserID,
				GroupID:    groupID,
				Content:    content,
				MsgType:    msgType,
				CreatedAt:  time.Now().Unix(),
			}
			cache.SaveOfflineMessage(m.UserID, &offlineMsg)
		}
		//给每个群成员增加未读数
		s.repo.IncrGroupUnread(m.UserID, groupID)

	}
	return nil
}

// MarkGroupChatRead 标记群聊已读
func (s *Service) MarkGroupChatRead(userID, groupID uint) error {
	return s.repo.ClearGroupUnread(userID, groupID)
}

// GetAllGroupUnread 获取所有群未读数
func (s *Service) GetAllGroupUnread(userID uint) ([]UnreadCount, error) {
	return s.repo.GetAllUnreadCounts(userID)
}

// GetGroupMessages 获取群聊历史消息
func (s *Service) GetGroupMessages(groupID, cursor uint, limit int) ([]GroupMessageResp, error) {
	messages, err := s.repo.FindMessagesByGroupID(groupID, limit+1, 0)
	if err != nil {
		return nil, err
	}

	// 获取所有发送者的用户信息
	userIDs := make([]uint, len(messages))
	for i, msg := range messages {
		userIDs[i] = msg.FromUserID
	}

	var users []user.User
	if len(userIDs) > 0 {
		config.DB.Where("id IN ?", userIDs).Find(&users)
	}
	userMap := make(map[uint]user.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	result := make([]GroupMessageResp, 0, len(messages))
	for _, msg := range messages {
		senderName := ""
		if u, ok := userMap[msg.FromUserID]; ok {
			senderName = u.Nickname
		}
		result = append(result, GroupMessageResp{
			ID:           msg.ID,
			GroupID:      msg.GroupID,
			FromUserID:   msg.FromUserID,
			FromUserName: senderName,
			Content:      msg.Content,
			MsgType:      msg.MsgType,
			CreatedAt:    uint(msg.CreatedAt.Unix()),
		})
	}
	return result, nil
}

func (s *Service) getMemberInfo(userID uint) (MemberInfo, error) {
	var user user.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return MemberInfo{}, err
	}
	return MemberInfo{
		UserID:   user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
	}, nil
}

func (s *Service) getSenderNickname(userID uint) string {
	var user user.User
	err := config.DB.First(&user, userID).Error
	if err != nil {
		return ""
	}
	return user.Nickname
}
