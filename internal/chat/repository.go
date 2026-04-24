package chat

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	Create(msg *Message) error
	FindByID(id uint) (*Message, error)
	FindByUserID(userID uint, limit, offset int) ([]Message, error)
	FindConversation(userID1, userID2 uint, limit, offset int) ([]Message, error)
	MarkAsRead(messageID uint) error
	MarkConversationAsRead(userID, toUserID uint) error

	//单聊已读游标
	UpdateReadCursor(userID, peerID, lastMsgID uint) error
	GetUnreadCount(userID, peerID uint) (int64, error)
	IsMessageRead(msgID, userID, peerID uint) (bool, error)

	GetConversations(userID uint) ([]Conversation, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// 创建消息
func (r *repository) Create(msg *Message) error {
	return r.db.Create(msg).Error
}

// id查消息
func (r *repository) FindByID(id uint) (*Message, error) {
	var msg Message

	err := r.db.First(&msg, id).Error
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// 查询某个用户的所有消息
func (r *repository) FindByUserID(userID uint, limit, offset int) ([]Message, error) {
	var messages []Message
	err := r.db.Where("from_user_id = ? or to_user_id = ?", userID, userID).
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error
	return messages, err
}

// 查询两个用户之间的聊天记录
func (r *repository) FindConversation(userID1, userID2 uint, limit, offset int) ([]Message, error) {
	var messages []Message
	err := r.db.Where(
		"(from_user_id = ? and to_user_id = ?) or (from_user_id = ? and to_user_id = ?)",
		userID1, userID2, userID2, userID1,
	).Order("created_at desc").Limit(limit).
		Offset(offset).Find(&messages).Error

	return messages, err

}

// 标记单条消息已读
func (r *repository) MarkAsRead(messageID uint) error {
	return r.db.Model(&Message{}).Where("id = ?", messageID).
		Update("is_read", true).Error
}

// 标记整个对话已读
func (r *repository) MarkConversationAsRead(userID, toUserID uint) error {
	return r.db.Model(&Message{}).
		Where("to_user_id = ? and from_user_id = ? and is_read = ?",
			userID, toUserID, false).Update("is_read", true).Error

}

// 更新单聊已读游标
func (r *repository) UpdateReadCursor(userID, peerID, lastMsgID uint) error {
	cursor := &ReadCursor{
		UserID:        userID,
		PeerID:        peerID,
		LastReadMsgID: lastMsgID,
	}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "peer_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"last_read_msg_id", "updated_at"}),
	}).Create(cursor).Error
}

// 获取单聊未读数
func (r *repository) GetUnreadCount(userID, peerID uint) (int64, error) {
	var cursor ReadCursor
	r.db.Where("user_id = ? AND peer_id = ?", userID, peerID).First(&cursor)

	var count int64
	r.db.Model(&Message{}).
		Where("from_user_id = ? AND to_user_id = ? AND id > ?", peerID, userID, cursor.LastReadMsgID).
		Count(&count)
	return count, nil
}

// 判断消息是否已被对方已读
func (r *repository) IsMessageRead(msgID, userID, peerID uint) (bool, error) {
	var cursor ReadCursor
	err := r.db.Where("user_id = ? AND peer_id = ?", peerID, userID).First(&cursor).Error
	if err != nil {
		return false, nil // 对方还没打开过会话
	}
	return msgID <= cursor.LastReadMsgID, nil
}

// 获取会话列表
func (r *repository) GetConversations(userID uint) ([]Conversation, error) {
	var conversations []Conversation

	// 单聊会话 - 查询所有聊过天的对方
	r.db.Raw(`
        SELECT 
            CASE WHEN m.from_user_id = ? THEN m.to_user_id ELSE m.from_user_id END AS peer_id,
            u.nickname AS peer_name,
            u.avatar,
            0 AS conv_type,
            m.content AS last_msg,
            m.created_at AS last_time
        FROM messages m
        JOIN users u ON u.id = CASE WHEN m.from_user_id = ? THEN m.to_user_id ELSE m.from_user_id END
        WHERE (m.from_user_id = ? OR m.to_user_id = ?)
        ORDER BY m.created_at DESC
        LIMIT 1
    `, userID, userID, userID, userID).Scan(&conversations)

	// 群聊会话
	r.db.Raw(`
        SELECT 
            gm.group_id AS peer_id,
            g.name AS peer_name,
            g.avatar,
            1 AS conv_type,
            gm.content AS last_msg,
            gm.created_at AS last_time
        FROM group_messages gm
        JOIN group_members gmbr ON gm.group_id = gmbr.group_id
        JOIN `+"`groups`"+` g ON g.id = gm.group_id
        WHERE gmbr.user_id = ?
        ORDER BY gm.created_at DESC
        LIMIT 1
    `, userID).Scan(&conversations)

	return conversations, nil
}
