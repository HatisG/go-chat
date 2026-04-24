package chat

import (
	"go-chat/internal/friend"
	"go-chat/internal/group"
	"go-chat/pkg/errcode"
	"go-chat/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

var (
	hub     *Hub
	service *Service
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func Init(friendRepo friend.Repository, messageRepo Repository, groupService *group.Service, groupRepo group.Repository) *Handler {
	hub = NewHub()
	service = NewService(hub, friendRepo, messageRepo, groupService, groupRepo)
	go hub.Run()
	return NewHandler(service)
}

func WebSocketHandler(c *gin.Context) {
	userID := c.GetUint("user_id")
	ServerWS(hub, service, userID, c.Writer, c.Request)
}

func GetHub() *Hub {
	return hub
}

func SetGroupService(gs *group.Service) {
	service.groupService = gs
}

// 标记单聊已读
func (h *Handler) MarkSingleRead(c *gin.Context) {
	var req struct {
		PeerID    uint `json:"peer_id" binding:"required"`
		LastMsgID uint `json:"last_msg_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.InvalidParams)
		return
	}

	userID := c.GetUint("user_id")

	if err := h.service.MarkSingleChatRead(userID, req.PeerID, req.LastMsgID); err != nil {
		response.Error(c, errcode.ServerError)
		return
	}

	response.Success(c, gin.H{"message": "已标记已读"})
}

// 获取单聊未读数
func (h *Handler) GetSingleUnread(c *gin.Context) {
	peerID, err := strconv.ParseUint(c.Query("peer_id"), 10, 32)
	if err != nil {
		response.Error(c, errcode.InvalidParams)
		return
	}

	userID := c.GetUint("user_id")

	count, err := h.service.GetSingleChatUnread(userID, uint(peerID))
	if err != nil {
		response.Error(c, errcode.ServerError)
		return
	}

	response.Success(c, gin.H{"unread_count": count})
}

// 获取会话列表
func (h *Handler) GetConversations(c *gin.Context) {
	userID := c.GetUint("user_id")

	conversations, err := h.service.GetConversations(userID)
	if err != nil {
		response.Error(c, errcode.ServerError)
		return
	}

	response.Success(c, conversations)
}

// GetMessages 获取单聊历史消息（游标分页）
func (h *Handler) GetMessages(c *gin.Context) {
	peerID, _ := strconv.ParseUint(c.Query("peer_id"), 10, 32)
	cursor, _ := strconv.ParseUint(c.Query("cursor"), 10, 32)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	userID := c.GetUint("user_id")

	messages, err := h.service.GetConversationMessages(userID, uint(peerID), uint(cursor), limit)
	if err != nil {
		response.Error(c, errcode.ServerError)
		return
	}

	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}

	response.Success(c, gin.H{
		"messages": messages,
		"has_more": hasMore,
	})
}
