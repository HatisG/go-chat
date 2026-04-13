package friend

import (
	"go-chat/pkg/errcode"
	"go-chat/pkg/response"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// 发送好友申请
func (h *Handler) SendRequest(c *gin.Context) {
	var req struct {
		ToUserID uint   `json:"to_user_id" binding:"required"`
		Msg      string `json:"msg" binding:"max=255"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.InvalidParams)
		return
	}

	currentUserID := c.GetUint("user_id")

	err := h.service.SendRequest(currentUserID, req.ToUserID, req.Msg)
	if err != nil {
		response.Error(c, errcode.ServerError)
		log.Println(err)
		return
	}

	response.Success(c, gin.H{"message": "申请已发送"})
}

// 接受申请
func (h *Handler) AcceptRequest(c *gin.Context) {
	requestID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errcode.InvalidParams)
		return
	}

	currentUserID := c.GetUint("user_id")

	err = h.service.AcceptRequest(uint(requestID), currentUserID)
	if err != nil {
		response.Error(c, errcode.ServerError)
		log.Println(err)
		return
	}

	response.Success(c, gin.H{"message": "已同意好友申请"})

}

// 拒绝申请
func (h *Handler) RejectRequest(c *gin.Context) {
	requestID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errcode.InvalidParams)
		return
	}

	currentUserID := c.GetUint("user_id")

	err = h.service.RejectRequest(uint(requestID), currentUserID)
	if err != nil {
		response.Error(c, errcode.ServerError)
		log.Println(err)
		return
	}

	response.Success(c, gin.H{"message": "已拒绝好友申请"})

}

// 获取好友申请列表
func (h *Handler) GetPendingRequests(c *gin.Context) {
	currentUserID := c.GetUint("user_id")

	reqs, err := h.service.GetPendingRequests(currentUserID)
	if err != nil {
		response.Error(c, errcode.ServerError)
		log.Println(err)
		return
	}

	response.Success(c, reqs)

}

// 获取好友列表
func (h *Handler) GetFriendList(c *gin.Context) {
	currentUserID := c.GetUint("user_id")

	friends, err := h.service.GetFriendListWithInfo(currentUserID)
	if err != nil {
		response.Error(c, errcode.ServerError)
		log.Println(err)
		return
	}

	response.Success(c, friends)
}

// 删除好友
func (h *Handler) DeleteFriend(c *gin.Context) {
	friendID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errcode.InvalidParams)
		return
	}

	currentUserID := c.GetUint("user_id")

	err = h.service.DeleteFriend(currentUserID, uint(friendID))
	if err != nil {
		response.Error(c, errcode.ServerError)
		log.Println(err)
		return
	}

	response.Success(c, gin.H{"message": "好友已删除"})

}
