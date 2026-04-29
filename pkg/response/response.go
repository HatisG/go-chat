package response

import (
	"go-chat/pkg/errcode"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: errcode.Success,
		Msg:  errcode.GetMsg(errcode.Success),
		Data: data,
	})
}

func Error(c *gin.Context, code int) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  errcode.GetMsg(code),
	})
}

func ErrorWithMsg(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
	})
}
