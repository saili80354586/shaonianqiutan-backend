package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/config"
	"github.com/shaonianqiutan/backend/middleware"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/utils"
)

// 上传文件大小限制（按类型）
const (
	maxAvatarSize = 10 * 1024 * 1024  // 头像 10MB
	maxImageSize  = 20 * 1024 * 1024  // 图片 20MB
	maxVideoSize  = 500 * 1024 * 1024 // 视频 500MB
	maxFileSize   = 50 * 1024 * 1024  // 其他文件 50MB
)

// getMaxSizeByType 根据上传类型返回最大文件大小
func getMaxSizeByType(fileType string) int64 {
	switch fileType {
	case "avatars":
		return maxAvatarSize
	case "videos":
		return maxVideoSize
	case "images":
		return maxImageSize
	default:
		return maxFileSize
	}
}

// UploadController 文件上传控制器
type UploadController struct {
	uploadDir string
	baseURL   string
}

// NewUploadController 创建上传控制器
func NewUploadController() *UploadController {
	uploadDir := "./uploads"
	if config.IsDevMode() {
		uploadDir = "./uploads"
	}
	// 确保上传目录存在
	_ = os.MkdirAll(uploadDir, 0755)
	_ = os.MkdirAll(filepath.Join(uploadDir, "avatars"), 0755)
	_ = os.MkdirAll(filepath.Join(uploadDir, "videos"), 0755)
	_ = os.MkdirAll(filepath.Join(uploadDir, "images"), 0755)
	_ = os.MkdirAll(filepath.Join(uploadDir, "files"), 0755)

	return &UploadController{
		uploadDir: uploadDir,
		baseURL:   config.GetBaseUrl(),
	}
}

// UploadFileRequest 上传文件请求
func (ctrl *UploadController) UploadFile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	fileType := c.PostForm("type")
	if fileType == "" {
		fileType = "files"
	}

	// 只允许特定的上传类型
	allowedTypes := map[string]bool{
		"avatars": true,
		"videos":  true,
		"images":  true,
		"files":   true,
	}
	if !allowedTypes[fileType] {
		utils.Error(c, http.StatusBadRequest, "无效的文件类型")
		return
	}

	// 限制请求体大小，防止内存耗尽
	maxSize := getMaxSizeByType(fileType)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize+1024*1024) // 额外留1MB缓冲区给multipart开销

	file, err := c.FormFile("file")
	if err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			utils.Error(c, http.StatusBadRequest, fmt.Sprintf("文件大小超过限制（最大 %dMB）", maxSize/1024/1024))
			return
		}
		utils.Error(c, http.StatusBadRequest, "获取文件失败: "+err.Error())
		return
	}

	// 校验文件大小
	if file.Size > maxSize {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("文件大小超过限制（最大 %dMB）", maxSize/1024/1024))
		return
	}

	// 检查文件后缀
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string][]string{
		"avatars": {".jpg", ".jpeg", ".png", ".gif", ".webp"},
		"videos":  {".mp4", ".mov", ".avi", ".mkv", ".webm"},
		"images":  {".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg"},
		"files":   {".jpg", ".jpeg", ".png", ".gif", ".pdf", ".doc", ".docx", ".xls", ".xlsx", ".csv", ".mp4", ".mov"},
	}
	validExt := false
	for _, ae := range allowedExts[fileType] {
		if ext == ae {
			validExt = true
			break
		}
	}
	if !validExt {
		utils.Error(c, http.StatusBadRequest, "不支持的文件格式")
		return
	}

	// 生成唯一文件名: {timestamp}_{userid}_{random}{ext}
	timestamp := time.Now().UnixNano()
	random := fmt.Sprintf("%04d", userID)
	newFilename := fmt.Sprintf("%d_%s%s", timestamp, random, ext)
	savePath := filepath.Join(ctrl.uploadDir, fileType, newFilename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		utils.Error(c, http.StatusInternalServerError, "保存文件失败: "+err.Error())
		return
	}

	// 返回可访问的 URL
	fileURL := fmt.Sprintf("%s/uploads/%s/%s", ctrl.baseURL, fileType, newFilename)

	utils.Success(c, "上传成功", gin.H{
		"url":      fileURL,
		"filename": newFilename,
		"original": file.Filename,
		"size":     file.Size,
	})
}

// UploadAvatar 上传头像（兼容旧接口）
func (ctrl *UploadController) UploadAvatar(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		utils.Error(c, http.StatusUnauthorized, "未认证")
		return
	}

	// 限制头像大小
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxAvatarSize+1024*1024)

	file, err := c.FormFile("file")
	if err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			utils.Error(c, http.StatusBadRequest, fmt.Sprintf("头像大小超过限制（最大 %dMB）", maxAvatarSize/1024/1024))
			return
		}
		utils.Error(c, http.StatusBadRequest, "获取文件失败: "+err.Error())
		return
	}

	if file.Size > maxAvatarSize {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("头像大小超过限制（最大 %dMB）", maxAvatarSize/1024/1024))
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	valid := false
	for _, ae := range allowed {
		if ext == ae {
			valid = true
			break
		}
	}
	if !valid {
		utils.Error(c, http.StatusBadRequest, "仅支持 JPG/PNG/GIF/WebP 格式")
		return
	}

	timestamp := time.Now().UnixNano()
	newFilename := fmt.Sprintf("%d_avatar_%d%s", userID, timestamp, ext)
	savePath := filepath.Join(ctrl.uploadDir, "avatars", newFilename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		utils.Error(c, http.StatusInternalServerError, "保存文件失败: "+err.Error())
		return
	}

	fileURL := fmt.Sprintf("%s/uploads/avatars/%s", ctrl.baseURL, newFilename)

	if err := config.GetDB().Model(&models.User{}).Where("id = ?", userID).Update("avatar", fileURL).Error; err != nil {
		_ = os.Remove(savePath)
		utils.Error(c, http.StatusInternalServerError, "更新头像失败: "+err.Error())
		return
	}

	utils.Success(c, "头像上传成功", gin.H{
		"avatar":   fileURL,
		"filename": newFilename,
	})
}
