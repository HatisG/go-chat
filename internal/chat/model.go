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
