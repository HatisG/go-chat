package group

import (
	"errors"
	"go-chat/internal/cache"
	"go-chat/internal/message"
	"time"
)

type HubInterface interface {
	IsOnline(userID uint) bool
	SendToUser(msg interface{}) // 广播消息
}

type GroupWSMessage struct {
	Type     string `json:"type"`
	ToUserID uint   `json:"to_user_id"`
	GroupID  uint   `json:"group_id"`
	MsgType  string `json:"msg_type"`
	Content  string `json:"content"`
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

func (s *Service) JoinGroup(groupID, userID uint) error {

	_, err := s.repo.FindByID(groupID)
	if err != nil {
		return errors.New("群不存在")
	}

	_, err = s.repo.FindMember(groupID, userID)
	if err != nil {
		return errors.New("已在该群成员")
	}

	member := &GroupMember{
		GroupID: groupID,
		UserID:  userID,
		Role:    RoleMember,
	}

	return s.repo.AddMember(member)

}

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

func (s *Service) GetGroupMembers(groupID uint) ([]MemberInfo, error) {
	members, err := s.repo.FindMembersByGroupID(groupID)
	if err != nil {
		return nil, errors.New("群成员获取失败")
	}

	result := make([]MemberInfo, 0, len(members))

	for _, m := range members {

		result = append(result, MemberInfo{
			UserID: m.UserID,
			Role:   m.Role,
		})

	}

	return result, nil
}

func (s *Service) SendGroupMessage(groupID, fromUserID uint, msgType, content string) error {

	_, err := s.repo.FindMember(groupID, fromUserID)
	if err != nil {
		return errors.New("不是该群成员")
	}

	//模拟io延迟，测试用
	time.Sleep(5 * time.Millisecond)

	//投递消息队列
	chatMsg := &message.ChatMessage{
		FromUserID: fromUserID,
		ToUserID:   0,
		Content:    content,
		MsgType:    msgType,
	}
	message.PublishMessage(chatMsg)

	//保存消息
	s.repo.SaveMessage(&GroupMessage{
		GroupID:    groupID,
		FromUserID: fromUserID,
		Content:    content,
		MsgType:    msgType,
	})

	//投递消息
	members, err := s.repo.FindMembersByGroupID(groupID)
	if err != nil {
		return err
	}
	for _, m := range members {
		if m.ID == fromUserID {
			continue
		}
		//在线发送，离线投到redis
		if s.hub.IsOnline(m.UserID) {
			s.hub.SendToUser(&GroupWSMessage{
				Type:     "group_chat",
				ToUserID: m.UserID,
				GroupID:  groupID,
				MsgType:  msgType,
				Content:  content,
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

	}
	return nil
}
