package friend

import (
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	// 好友申请相关
	CreateRequest(req *FriendRequest) error
	FindRequestByID(id uint) (*FriendRequest, error)
	FindPendingRequest(fromUserID, toUserID uint) (*FriendRequest, error)
	FindPendingRequestsByToUser(toUserID uint) ([]FriendRequest, error)
	UpdateRequestStatus(id uint, status int) error

	// 好友关系相关
	CreateFriendship(friendship *Friendship) error
	FindFriendship(userID, friendID uint) (*Friendship, error)
	FindFriendsByUserID(userID uint) ([]Friendship, error)
	DeleteFriendship(userID, friendID uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateRequest(req *FriendRequest) error {
	return r.db.Create(req).Error
}

func (r *repository) FindRequestByID(id uint) (*FriendRequest, error) {
	var req FriendRequest

	err := r.db.First(&req, id).Error
	if err != nil {
		return nil, err
	}

	return &req, nil

}

func (r *repository) FindPendingRequest(fromUserID, toUserID uint) (*FriendRequest, error) {
	var req FriendRequest

	err := r.db.Where("from_user_id = ? and to_user_id = ? and status = ?",
		fromUserID, toUserID, 0).First(&req).Error
	if err != nil {
		return nil, err
	}
	return &req, nil

}

func (r *repository) FindPendingRequestsByToUser(toUserID uint) ([]FriendRequest, error) {
	var reqs []FriendRequest

	err := r.db.Where("to_user_id = ? and status = ?", toUserID, 0).Find(&reqs).Error

	return reqs, err

}

func (r *repository) UpdateRequestStatus(id uint, status int) error {

	now := time.Now()
	return r.db.Model(&FriendRequest{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     status,
		"handled_at": &now,
	}).Error

}

func (r *repository) CreateFriendship(friendship *Friendship) error {
	return r.db.Create(friendship).Error
}

func (r *repository) FindFriendship(userID, friendID uint) (*Friendship, error) {
	var friendship Friendship

	err := r.db.Where(
		"(user_id = ? and friend_id = ?) or (user_id = ? and friend_id = ?)",
		userID, friendID, friendID, userID,
	).First(&friendship).Error

	if err != nil {
		return nil, err
	}

	return &friendship, nil

}

func (r *repository) FindFriendsByUserID(userID uint) ([]Friendship, error) {
	var friendships []Friendship

	err := r.db.Where("user_id = ? or friend_id = ?", userID, userID).Find(&friendships).Error
	return friendships, err
}

func (r *repository) DeleteFriendship(userID, friendID uint) error {
	return r.db.Where(
		"(user_id = ? and friend_id = ?) or (user_id = ? and friend_id = ?)",
		userID, friendID, friendID, userID,
	).Delete(&Friendship{}).Error

}
