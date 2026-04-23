package user

import (
	"errors"
	"go-chat/internal/config"
	"go-chat/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// 注册
func (s *Service) Register(username, password string) (*User, error) {
	//查用户名
	_, err := s.repo.FindByUsername(username)
	if err == nil {
		return nil, errors.New("用户名已存在")
	}

	//哈希加密密码
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	user := &User{
		Username: username,
		Password: string(hashPassword),
		Nickname: username,
	}

	//创建
	if err := s.repo.Create(user); err != nil {
		return nil, errors.New("用户创建失败")
	}
	return user, nil
}

// 登录
func (s *Service) Login(username, password string) (*User, string, error) {
	//查用户名
	user, err := s.repo.FindByUsername(username)
	if err != nil {
		return nil, "", errors.New("用户名或密码错误")
	}

	//哈希密码比较
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, "", errors.New("用户名或密码错误")
	}

	cfg := config.AppConfig

	//生成token
	token, err := jwt.GenerateToken(user.ID, user.Username, cfg.JWT.Secret, cfg.JWT.ExpireHours)
	if err != nil {
		return nil, "", errors.New("生成token失败")
	}

	return user, token, nil
}

// 更新用户资料
func (s *Service) UpdateProfile(userID uint, nickname string) error {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}
	user.Nickname = nickname
	return s.repo.Update(user)
}

// 更新用户头像
func (s *Service) UpdateAvatar(userID uint, avatar string) error {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}
	user.Avatar = avatar
	return s.repo.Update(user)
}

// 修改密码
func (s *Service) ChangePassword(userID uint, oldPassword, newPassword string) error {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	//验证旧密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
	if err != nil {
		return errors.New("旧密码错误")
	}
	//加密新密码
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}
	return s.repo.UpdatePassword(userID, string(hashedPwd))
}
