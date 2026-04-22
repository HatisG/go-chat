package upload

import (
	"fmt"
	"go-chat/internal/logger"
	"go-chat/pkg/errcode"
	"go-chat/pkg/response"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-uuid"
	"go.uber.org/zap"
)

const (
	UploadDir   = "./uploads"
	MaxFileSize = 10 << 20
)

// Handler 文件上传处理器
type Handler struct{}

func NewHandler() *Handler {
	if err := os.MkdirAll(UploadDir, 0755); err != nil {
		logger.Logger.Fatal("创建上传目录失败", zap.Error(err))
	}
	return &Handler{}
}

func (h *Handler) Upload(c *gin.Context) {

	//获取上传文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Error(c, errcode.InvalidParams)
		return
	}
	defer file.Close()

	//检测大小
	if header.Size > MaxFileSize {
		response.Error(c, errcode.InvalidParams)
		return
	}

	//生成文件名
	ext := filepath.Ext(header.Filename)
	uUid, err := uuid.GenerateUUID()
	if err != nil {
		response.Error(c, errcode.ServerError)
		return
	}
	newFileName := fmt.Sprintf("%s%s", uUid, ext)

	//按日期创建子目录
	dataDir := time.Now().Format("2006/01/02")
	fullDir := filepath.Join(UploadDir, dataDir)
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		response.Error(c, errcode.ServerError)
		return
	}

	//保存
	fullPath := filepath.Join(fullDir, newFileName)
	dst, err := os.Create(fullPath)
	if err != nil {
		response.Error(c, errcode.ServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		response.Error(c, errcode.ServerError)
		return
	}

	//返回url
	fileURL := fmt.Sprintf("/uploads/%s/%s", dataDir, newFileName)

	//判断类型
	msgType := getMsgType(ext)

	response.Success(c, gin.H{
		"url":      fileURL,
		"msg_type": msgType,
		"filename": header.Filename,
		"size":     header.Size,
	})
}

func getMsgType(ext string) string {
	ext = strings.ToLower(ext)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp":
		return "image"
	case ".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv":
		return "video"
	default:
		return "file"
	}
}
