package user

import (
	"go-chat/pkg/errcode"
	"go-chat/pkg/response"

	"github.com/gin-gonic/gin"
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
