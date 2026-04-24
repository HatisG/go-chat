package group

import "gorm.io/gorm"

type Group struct {
	gorm.Model
	Name      string `gorm:"size:64;not null"`
	Avatar    string `gorm:"size:255"`
	CreatorID uint   `gorm:"not null;index"`
}

type GroupMember struct {
	gorm.Model
	GroupID uint `gorm:"not null;uniqueIndex:uk_group_user;index"`
	UserID  uint `gorm:"not null;uniqueIndex:uk_group_user;index"`
	Role    int  `gorm:"default:0"` // 0:成员 1:管理员 2:群主
}

type GroupMessage struct {
	gorm.Model
	GroupID    uint   `gorm:"not null;index"`
	FromUserID uint   `gorm:"not null;index"`
	Content    string `gorm:"type:text;not null"`
	MsgType    string `gorm:"size:20;default:'text'"`
}

type UnreadCount struct {
	gorm.Model
	UserID      uint `gorm:"not null;uniqueIndex:uk_user_group"`
	GroupID     uint `gorm:"not null;uniqueIndex:uk_user_group"`
	UnreadCount int  `gorm:"default:0"`
}

const (
	RoleMember = 0
	RoleAdmin  = 1
	RoleOwner  = 2
)
