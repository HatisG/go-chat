package group

import (
	"go-chat/internal/logger"
	"go-chat/pkg/errcode"
	"go-chat/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// 创建群聊
func (h *Handler) CreateGroup(c *gin.Context) {
	var req CreateGroupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.InvalidParams)
		return
	}

	userID := c.GetUint("user_id")
	group, err := h.service.CreateGroup(userID, req.Name)
	if err != nil {
		response.Error(c, errcode.ServerError)
		logger.Logger.Info("创建群聊失败", zap.Error(err))
		return
	}

	response.Success(c, gin.H{
		"id":   group.ID,
		"name": group.Name,
	})

}

// 加入群聊
func (h *Handler) JoinGroup(c *gin.Context) {
	groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errcode.InvalidParams)
		return
	}

	userID := c.GetUint("user_id")
	if err = h.service.JoinGroup(uint(groupID), userID); err != nil {
		response.Error(c, errcode.ServerError)
		logger.Logger.Info("加入群聊失败", zap.Error(err))
		return
	}

	response.Success(c, gin.H{"message": "加入成功"})
}

// 离开群聊
func (h *Handler) LeaveGroup(c *gin.Context) {
	groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errcode.InvalidParams)
		return
	}

	userID := c.GetUint("user_id")
	if err = h.service.LeaveGroup(uint(groupID), userID); err != nil {
		response.Error(c, errcode.ServerError)
		logger.Logger.Info("离开群聊失败", zap.Error(err))
		return
	}

	response.Success(c, gin.H{"message": "退出成功"})
}

// 获取用户群组
func (h *Handler) GetMyGroups(c *gin.Context) {
	userID := c.GetUint("user_id")

	groups, err := h.service.GetMyGroups(userID)
	if err != nil {
		response.Error(c, errcode.ServerError)
		logger.Logger.Info("获取群聊失败", zap.Error(err))
		return
	}

	response.Success(c, groups)

}

// 获取群成员
func (h *Handler) GetGroupMembers(c *gin.Context) {
	groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errcode.InvalidParams)
		return
	}

	members, err := h.service.GetGroupMembers(uint(groupID))
	if err != nil {
		response.Error(c, errcode.ServerError)
		logger.Logger.Info("获取群成员失败", zap.Error(err))
		return
	}

	response.Success(c, members)
}

// 标记群聊已读
func (h *Handler) MarkGroupRead(c *gin.Context) {
	var req struct {
		GroupID uint `json:"group_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.InvalidParams)
		return
	}

	userID := c.GetUint("user_id")

	if err := h.service.MarkGroupChatRead(userID, req.GroupID); err != nil {
		response.Error(c, errcode.ServerError)
		return
	}

	response.Success(c, gin.H{"message": "已标记已读"})
}

// 获取所有群未读数
func (h *Handler) GetAllGroupUnread(c *gin.Context) {
	userID := c.GetUint("user_id")

	counts, err := h.service.GetAllGroupUnread(userID)
	if err != nil {
		response.Error(c, errcode.ServerError)
		return
	}

	response.Success(c, counts)
}
