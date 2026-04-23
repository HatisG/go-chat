package user

import (
	"go-chat/internal/logger"
	"go-chat/pkg/errcode"
	"go-chat/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	service *Service
}

type Request struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// 注册
func (h *Handler) Register(c *gin.Context) {
	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.InvalidParams)
		return
	}

	user, err := h.service.Register(req.Username, req.Password)
	if err != nil {
		response.Error(c, errcode.UserAlreadyExists)
		return
	}

	response.Success(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"nickname": user.Nickname,
	})

}

// 登录
func (h *Handler) Login(c *gin.Context) {
	var req Request

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.InvalidParams)
		return
	}

	user, token, err := h.service.Login(req.Username, req.Password)
	if err != nil {
		response.Error(c, errcode.Unauthorized)
		return
	}

	response.Success(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"nickname": user.Nickname,
		"token":    token,
	})

}

// 获取当前用户信息
func (h *Handler) GetProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	user, err := h.service.repo.FindByID(userID)
	if err != nil {
		response.Error(c, errcode.UserNotFound)
		return
	}

	response.Success(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"nickname": user.Nickname,
		"avatar":   user.Avatar,
	})

}

// 修改个人资料
func (h *Handler) UpdateProfile(c *gin.Context) {
	var req struct {
		Nickname string `json:"nickname" binding:"required,max=32"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.InvalidParams)
		return
	}

	userID := c.GetUint("user_id")

	if err := h.service.UpdateProfile(userID, req.Nickname); err != nil {
		response.Error(c, errcode.ServerError)
		return
	}

	response.Success(c, gin.H{"message": "修改成功"})
}

// 修改头像
func (h *Handler) UpdateAvatar(c *gin.Context) {
	var req struct {
		Avatar string `json:"avatar" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.InvalidParams)
		return
	}

	userID := c.GetUint("user_id")

	if err := h.service.UpdateAvatar(userID, req.Avatar); err != nil {
		response.Error(c, errcode.ServerError)
		return
	}

	response.Success(c, gin.H{"message": "头像更新成功"})
}

// 修改密码
func (h *Handler) ChangePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.InvalidParams)
		return
	}

	userID := c.GetUint("user_id")

	if err := h.service.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		response.Error(c, errcode.ServerError)
		logger.Logger.Info("用户修改密码失败", zap.Uint("user_id", userID), zap.Error(err))
		return
	}

	response.Success(c, gin.H{"message": "密码修改成功"})
}
