package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shaonianqiutan/backend/config"
	"github.com/shaonianqiutan/backend/models"
	"github.com/tencentyun/cos-go-sdk-v5"
)

const maxOrderSourceVideoSize = 500 * 1024 * 1024

type StorageService struct {
	cfg  config.StorageConfig
	repo *models.StorageObjectRepository
	cos  *cos.Client
}

type CreateOrderSourceUploadInput struct {
	UserID        uint
	OrderID       uint
	Filename      string
	ContentType   string
	Size          int64
	VideoDuration int
}

type UploadIntent struct {
	StorageObjectID uint              `json:"storage_object_id"`
	Driver          string            `json:"driver"`
	Bucket          string            `json:"bucket"`
	Region          string            `json:"region"`
	ObjectKey       string            `json:"object_key"`
	UploadURL       string            `json:"upload_url"`
	Method          string            `json:"method"`
	Headers         map[string]string `json:"headers"`
	ExpiresAt       time.Time         `json:"expires_at"`
	MaxSize         int64             `json:"max_size"`
}

type ConfirmStorageObjectInput struct {
	StorageObjectID uint
	ObjectKey       string
	ETag            string
	Size            int64
}

type UploadAnalysisClipInput struct {
	OrderID     uint
	AnalysisID  uint
	HighlightID uint
	Version     int
	LocalPath   string
	ContentType string
}

func NewStorageService(repo *models.StorageObjectRepository) *StorageService {
	cfg := config.GetStorageConfig()
	service := &StorageService{cfg: cfg, repo: repo}
	if cfg.Driver == config.StorageDriverCOS && cfg.BucketURL != "" {
		if parsed, err := url.Parse(cfg.BucketURL); err == nil {
			service.cos = cos.NewClient(&cos.BaseURL{BucketURL: parsed}, &http.Client{
				Timeout: 30 * time.Second,
				Transport: &cos.AuthorizationTransport{
					SecretID:  cfg.SecretID,
					SecretKey: cfg.SecretKey,
				},
			})
		}
	}
	return service
}

func (s *StorageService) CreateOrderSourceUploadIntent(ctx context.Context, input CreateOrderSourceUploadInput) (*UploadIntent, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if input.OrderID == 0 {
		return nil, errors.New("订单ID不能为空")
	}
	filename := strings.TrimSpace(input.Filename)
	if filename == "" {
		return nil, errors.New("文件名不能为空")
	}
	if input.Size <= 0 || input.Size > maxOrderSourceVideoSize {
		return nil, fmt.Errorf("视频大小超过限制（最大 %dMB）", maxOrderSourceVideoSize/1024/1024)
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if !isAllowedOrderSourceVideoExt(ext) {
		return nil, errors.New("仅支持 MP4/MOV/AVI/MKV/WebM 视频格式")
	}
	contentType := strings.TrimSpace(input.ContentType)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	objectKey := s.orderSourceObjectKey(input.OrderID, ext)
	object := &models.StorageObject{
		Bucket:       s.cfg.Bucket,
		Region:       s.cfg.Region,
		ObjectKey:    objectKey,
		OriginalName: filename,
		ContentType:  contentType,
		Size:         input.Size,
		OwnerType:    models.StorageOwnerOrder,
		OwnerID:      input.OrderID,
		BusinessType: models.StorageBusinessOrderSourceVideo,
		Status:       models.StorageObjectStatusPendingUpload,
	}
	if err := s.repo.Create(object); err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(s.cfg.UploadExpire)
	intent := &UploadIntent{
		StorageObjectID: object.ID,
		Driver:          s.cfg.Driver,
		Bucket:          s.cfg.Bucket,
		Region:          s.cfg.Region,
		ObjectKey:       objectKey,
		Method:          http.MethodPut,
		Headers:         map[string]string{"Content-Type": contentType},
		ExpiresAt:       expiresAt,
		MaxSize:         maxOrderSourceVideoSize,
	}

	switch s.cfg.Driver {
	case config.StorageDriverCOS:
		uploadURL, err := s.cosPresignedURL(ctx, http.MethodPut, objectKey, s.cfg.UploadExpire)
		if err != nil {
			return nil, err
		}
		intent.UploadURL = uploadURL
	case config.StorageDriverLocal:
		intent.UploadURL = config.GetBaseUrl() + "/api/upload/file"
	default:
		return nil, fmt.Errorf("不支持的存储驱动: %s", s.cfg.Driver)
	}

	return intent, nil
}

func (s *StorageService) ConfirmStorageObject(input ConfirmStorageObjectInput) (*models.StorageObject, error) {
	object, err := s.repo.FindByID(input.StorageObjectID)
	if err != nil {
		return nil, err
	}
	if object == nil {
		return nil, errors.New("存储对象不存在")
	}
	if input.ObjectKey != "" && input.ObjectKey != object.ObjectKey {
		return nil, errors.New("存储对象路径不匹配")
	}
	updates := map[string]interface{}{
		"status": models.StorageObjectStatusActive,
	}
	if strings.TrimSpace(input.ETag) != "" {
		updates["e_tag"] = strings.TrimSpace(input.ETag)
	}
	if input.Size > 0 {
		updates["size"] = input.Size
	}
	if err := s.repo.Update(object.ID, updates); err != nil {
		return nil, err
	}
	return s.repo.FindByID(object.ID)
}

func (s *StorageService) UploadAnalysisClip(ctx context.Context, input UploadAnalysisClipInput) (string, *models.StorageObject, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if input.OrderID == 0 || input.AnalysisID == 0 || input.HighlightID == 0 {
		return "", nil, errors.New("剪辑归属信息不完整")
	}
	if strings.TrimSpace(input.LocalPath) == "" {
		return "", nil, errors.New("剪辑文件路径不能为空")
	}
	info, err := os.Stat(input.LocalPath)
	if err != nil {
		return "", nil, err
	}
	ext := strings.ToLower(filepath.Ext(input.LocalPath))
	if ext == "" {
		ext = ".mp4"
	}
	contentType := strings.TrimSpace(input.ContentType)
	if contentType == "" {
		contentType = "video/mp4"
	}
	objectKey := s.analysisClipObjectKey(input.OrderID, input.AnalysisID, input.HighlightID, input.Version, ext)

	if s.cfg.Driver == config.StorageDriverCOS {
		if s.cos == nil {
			return "", nil, errors.New("COS 客户端未初始化")
		}
		resp, err := s.cos.Object.PutFromFile(ctx, objectKey, input.LocalPath, &cos.ObjectPutOptions{
			ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{ContentType: contentType},
		})
		if err != nil {
			return "", nil, err
		}
		etag := ""
		if resp != nil {
			etag = resp.Header.Get("ETag")
		}
		object := &models.StorageObject{
			Bucket:       s.cfg.Bucket,
			Region:       s.cfg.Region,
			ObjectKey:    objectKey,
			OriginalName: filepath.Base(input.LocalPath),
			ContentType:  contentType,
			Size:         info.Size(),
			ETag:         strings.Trim(etag, "\""),
			OwnerType:    models.StorageOwnerVideoAnalysis,
			OwnerID:      input.AnalysisID,
			BusinessType: models.StorageBusinessAnalysisClip,
			Status:       models.StorageObjectStatusActive,
		}
		if err := s.repo.Create(object); err != nil {
			return "", nil, err
		}
		return s.PublicObjectURL(object), object, nil
	}

	return "", nil, nil
}

func (s *StorageService) ObjectKeyFromURL(rawURL string) (string, bool) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", false
	}
	if strings.HasPrefix(rawURL, strings.Trim(s.cfg.ObjectPrefix, "/")+"/") {
		return strings.TrimLeft(rawURL, "/"), true
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", false
	}
	if parsed.Scheme == "" && strings.HasPrefix(strings.TrimLeft(parsed.Path, "/"), strings.Trim(s.cfg.ObjectPrefix, "/")+"/") {
		return strings.TrimLeft(parsed.Path, "/"), true
	}
	bucketURL, err := url.Parse(s.cfg.BucketURL)
	if err != nil || bucketURL.Host == "" || parsed.Host != bucketURL.Host {
		return "", false
	}
	objectKey := strings.TrimLeft(parsed.Path, "/")
	return objectKey, objectKey != ""
}

func (s *StorageService) DownloadObjectToTemp(ctx context.Context, objectKey string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if s.cfg.Driver != config.StorageDriverCOS {
		return "", errors.New("当前存储驱动不支持对象下载")
	}
	if s.cos == nil {
		return "", errors.New("COS 客户端未初始化")
	}
	objectKey = strings.TrimLeft(strings.TrimSpace(objectKey), "/")
	if objectKey == "" {
		return "", errors.New("对象路径不能为空")
	}
	resp, err := s.cos.Object.Get(ctx, objectKey, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ext := filepath.Ext(objectKey)
	if ext == "" {
		ext = ".tmp"
	}
	tmpFile, err := os.CreateTemp("", "cos-object-*"+ext)
	if err != nil {
		return "", err
	}
	tmpPath := tmpFile.Name()
	cleanup := true
	defer func() {
		_ = tmpFile.Close()
		if cleanup {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return "", err
	}
	cleanup = false
	return tmpPath, nil
}

func (s *StorageService) PublicObjectURL(object *models.StorageObject) string {
	if object == nil {
		return ""
	}
	if s.cfg.Driver == config.StorageDriverCOS {
		return strings.TrimRight(s.cfg.BucketURL, "/") + "/" + strings.TrimLeft(object.ObjectKey, "/")
	}
	return strings.TrimRight(config.GetBaseUrl(), "/") + "/" + strings.TrimLeft(object.ObjectKey, "/")
}

func (s *StorageService) cosPresignedURL(ctx context.Context, method, objectKey string, expire time.Duration) (string, error) {
	if s.cos == nil {
		return "", errors.New("COS 客户端未初始化")
	}
	if strings.TrimSpace(s.cfg.SecretID) == "" || strings.TrimSpace(s.cfg.SecretKey) == "" {
		return "", errors.New("COS_SECRET_ID/COS_SECRET_KEY 未配置")
	}
	u, err := s.cos.Object.GetPresignedURL(ctx, method, objectKey, s.cfg.SecretID, s.cfg.SecretKey, expire, nil, true)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (s *StorageService) orderSourceObjectKey(orderID uint, ext string) string {
	prefix := s.cfg.ObjectPrefix
	if prefix == "" {
		prefix = "prod"
	}
	return fmt.Sprintf("%s/orders/%d/source/original_%s%s", prefix, orderID, randomHex(12), ext)
}

func (s *StorageService) analysisClipObjectKey(orderID, analysisID, highlightID uint, version int, ext string) string {
	prefix := s.cfg.ObjectPrefix
	if prefix == "" {
		prefix = "prod"
	}
	if version <= 0 {
		version = 1
	}
	return fmt.Sprintf("%s/orders/%d/analysis/%d/clips/marker_%d_v%d%s", prefix, orderID, analysisID, highlightID, version, ext)
}

func isAllowedOrderSourceVideoExt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".mp4", ".mov", ".avi", ".mkv", ".webm":
		return true
	default:
		return false
	}
}

func randomHex(byteLen int) string {
	if byteLen <= 0 {
		byteLen = 12
	}
	buf := make([]byte, byteLen)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}
