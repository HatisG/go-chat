package chat

import (
	"time"

	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	FromUserID uint   `gorm:"not null;index"`
	ToUserID   uint   `gorm:"not null;index"`
	Content    string `gorm:"type:text;not null"`
	MsgType    string `gorm:"size:20;default:'text'"` //text,image,file
	IsRead     bool   `gorm:"default:false"`
	ReadAt     *time.Time
}

type ReadCursor struct {
	gorm.Model
	UserID        uint `gorm:"not null;uniqueIndex:uk_user_peer"`
	PeerID        uint `gorm:"not null;uniqueIndex:uk_user_peer"`
	LastReadMsgID uint `gorm:"default:0"`
}
