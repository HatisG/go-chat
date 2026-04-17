package group

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
		log.Println(err)
		return
	}

	response.Success(c, gin.H{
		"id":   group.ID,
		"name": group.Name,
	})

}

func (h *Handler) JoinGroup(c *gin.Context) {
	groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errcode.InvalidParams)
		return
	}

	userID := c.GetUint("user_id")
	if err = h.service.JoinGroup(uint(groupID), userID); err != nil {
		response.Error(c, errcode.ServerError)
		log.Println(err)
		return
	}

	response.Success(c, gin.H{"message": "加入成功"})
}

func (h *Handler) LeaveGroup(c *gin.Context) {
	groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errcode.InvalidParams)
		return
	}

	userID := c.GetUint("user_id")
	if err = h.service.LeaveGroup(uint(groupID), userID); err != nil {
		response.Error(c, errcode.ServerError)
		log.Println(err)
		return
	}

	response.Success(c, gin.H{"message": "退出成功"})
}

func (h *Handler) GetMyGroups(c *gin.Context) {
	userID := c.GetUint("user_id")

	groups, err := h.service.GetMyGroups(userID)
	if err != nil {
		response.Error(c, errcode.ServerError)
		log.Println(err)
		return
	}

	response.Success(c, groups)

}

func (h *Handler) GetGroupMembers(c *gin.Context) {
	groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, errcode.InvalidParams)
		return
	}

	members, err := h.service.GetGroupMembers(uint(groupID))
	if err != nil {
		response.Error(c, errcode.ServerError)
		log.Println(err)
		return
	}

	response.Success(c, members)
}
