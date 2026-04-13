package user

import "gorm.io/gorm"

type Repository interface {
	Create(user *User) error
	FindByUsername(username string) (*User, error)
	FindByID(id uint) (*User, error)
	Update(user *User) error
	Delete(id uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

//创建用户
func (r *repository) Create(user *User) error {
	return r.db.Create(user).Error
}

//用户名查找
func (r *repository) FindByUsername(username string) (*User, error) {
	var user User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

//id查找
func (r *repository) FindByID(id uint) (*User, error) {
	var user User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

//更新
func (r *repository) Update(user *User) error {
	return r.db.Save(user).Error
}

//删除
func (r *repository) Delete(id uint) error {
	return r.db.Delete(&User{}, id).Error
}
