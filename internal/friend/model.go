package friend

import (
	"time"

	"gorm.io/gorm"
)

type Friendship struct {
	gorm.Model
	UserID   uint `gorm:"not null;index:idx_user_friend;uniqueIndex:uk_user_friend"`
	FriendID uint `gorm:"not null;index:idx_user_friend;uniqueIndex:uk_user_friend"`
	Status   int  `gorm:"default:1"` //1正常 2拉黑
}

type FriendRequest struct {
	gorm.Model
	FromUserID uint   `gorm:"not null"`
	ToUserID   uint   `gorm:"not null;index:idx_to_user_status"`
	Status     int    `gorm:"default:0;index:idx_to_user_status"` //0待处理 1同意 2拒绝
	RequestMsg string `gorm:"size:255"`
	HandledAt  *time.Time
}
