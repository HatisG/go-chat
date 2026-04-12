package friend

import "errors"

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

//发送好友申请
func (s *Service) SendRequest(fromUserID, toUserID uint, msg string) error {
	if fromUserID == toUserID {
		return errors.New("不能添加自己为好友")
	}

	_, err := s.repo.FindFriendship(fromUserID, toUserID)
	if err == nil {
		return errors.New("已经是好友")
	}

	_, err = s.repo.FindPendingRequest(fromUserID, toUserID)
	if err == nil {
		return errors.New("已有待处理的好友申请")
	}

	req := &FriendRequest{
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		RequestMsg: msg,
		Status:     0,
	}

	return s.repo.CreateRequest(req)

}

//接受好友申请
func (s *Service) AcceptRequest(requestID, currentUserID uint) error {

	req, err := s.repo.FindRequestByID(requestID)
	if err != nil {
		return errors.New("申请不存在")
	}

	if req.ToUserID != currentUserID {
		return errors.New("无权处理该申请")
	}

	if req.Status != 0 {
		return errors.New("申请已处理")
	}

	if err := s.repo.UpdateRequestStatus(requestID, 1); err != nil {
		return errors.New("更新申请状态失败")
	}

	friendship := &Friendship{
		UserID:   req.FromUserID,
		FriendID: req.ToUserID,
		Status:   1,
	}

	return s.repo.CreateFriendship(friendship)
}

//拒绝好友申请
func (s *Service) RejectRequest(requestID, currentUserID uint) error {

	req, err := s.repo.FindRequestByID(requestID)
	if err != nil {
		return errors.New("申请不存在")
	}

	if req.ToUserID != currentUserID {
		return errors.New("无权处理该申请")
	}

	if req.Status != 0 {
		return errors.New("申请已处理")
	}

	return s.repo.UpdateRequestStatus(requestID, 2)
}

//删除好友
func (s *Service) DeleteFriend(userID, friendID uint) error {

	_, err := s.repo.FindFriendship(userID, friendID)
	if err != nil {
		return errors.New("不是好友关系")
	}

	return s.repo.DeleteFriendship(userID, friendID)
}

//获取好友申请列表
func (s *Service) GetPendingRequests(userID uint) ([]FriendRequest, error) {
	return s.repo.FindPendingRequestsByToUser(userID)
}

//获取好友列表
func (s *Service) GetFriendList(userID uint) ([]Friendship, error) {
	return s.repo.FindFriendsByUserID(userID)
}

//获取好友列表（详细）
func (s *Service) GetFriendListWithInfo(userID uint) ([]FriendInfo, error) {
	return s.repo.FindFriendInfoByUserID(userID)
}
