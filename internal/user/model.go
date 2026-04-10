package user

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex;not null;size:32"`
	Password string `gorm:"not null;size:128"`
	Nickname string `gorm:"size:32"`
	Avatar   string `gorm:"size:255"`
}
