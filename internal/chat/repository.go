package chat

import "gorm.io/gorm"

type Repository interface {
	Create(msg *Message) error
	FindByID(id uint) (*Message, error)
	FindByUserID(userID uint, limit, offset int) ([]Message, error)
	FindConversation(userID1, userID2 uint, limit, offset int) ([]Message, error)
	MarkAsRead(messageID uint) error
	MarkConversationAsRead(userID, toUserID uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

//创建消息
func (r *repository) Create(msg *Message) error {
	return r.db.Create(msg).Error
}

//id查消息
func (r *repository) FindByID(id uint) (*Message, error) {
	var msg Message

	err := r.db.First(&msg, id).Error
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

//查询某个用户的所有消息
func (r *repository) FindByUserID(userID uint, limit, offset int) ([]Message, error) {
	var messages []Message
	err := r.db.Where("from_user_id = ? or to_user_id = ?", userID, userID).
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error
	return messages, err
}

//查询两个用户之间的聊天记录
func (r *repository) FindConversation(userID1, userID2 uint, limit, offset int) ([]Message, error) {
	var messages []Message
	err := r.db.Where(
		"(from_user_id = ? and to_user_id = ?) or (from_user_id = ? and to_user_id = ?)",
		userID1, userID2, userID2, userID1,
	).Order("created_at desc").Limit(limit).
		Offset(offset).Find(&messages).Error

	return messages, err

}

//标记单条消息已读
func (r *repository) MarkAsRead(messageID uint) error {
	return r.db.Model(&Message{}).Where("id = ?", messageID).
		Update("is_read", true).Error
}

//标记整个对话已读
func (r *repository) MarkConversationAsRead(userID, toUserID uint) error {
	return r.db.Model(&Message{}).
		Where("to_user_id = ? and from_user_id = ? and is_read = ?",
			userID, toUserID, false).Update("is_read", true).Error

}
