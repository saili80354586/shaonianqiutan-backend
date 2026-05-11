package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/shaonianqiutan/backend/config"
	"github.com/shaonianqiutan/backend/models"
	"gorm.io/gorm"
)

const maxAnalysisClipDurationMs = 90 * 1000

type VideoClipService struct {
	db             *gorm.DB
	outputDir      string
	publicPrefix   string
	baseURL        string
	ffmpegPath     string
	storageService *StorageService
}

func NewVideoClipService(db *gorm.DB, storageService ...*StorageService) *VideoClipService {
	outputDir := strings.TrimSpace(os.Getenv("VIDEO_CLIP_OUTPUT_DIR"))
	if outputDir == "" {
		outputDir = "./uploads/video-clips"
	}
	_ = os.MkdirAll(outputDir, 0755)

	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("BASE_URL")), "/")
	if baseURL == "" && config.IsDevMode() {
		baseURL = "http://localhost" + config.GetPort()
	}
	if baseURL == "" {
		baseURL = strings.TrimRight(config.GetBaseUrl(), "/")
	}

	ffmpegPath := strings.TrimSpace(os.Getenv("FFMPEG_PATH"))
	if ffmpegPath == "" {
		if found, err := exec.LookPath("ffmpeg"); err == nil {
			ffmpegPath = found
		}
	}

	var storage *StorageService
	if len(storageService) > 0 {
		storage = storageService[0]
	}

	return &VideoClipService{
		db:             db,
		outputDir:      outputDir,
		publicPrefix:   "/uploads/video-clips",
		baseURL:        baseURL,
		ffmpegPath:     ffmpegPath,
		storageService: storage,
	}
}

func (s *VideoClipService) QueueHighlightClip(highlightID uint) (*models.AnalysisHighlight, error) {
	highlight, analysis, err := s.findHighlightWithAnalysis(highlightID)
	if err != nil {
		return nil, err
	}
	if highlight.Mode != models.HighlightModeRange {
		return s.ClearHighlightClip(highlightID)
	}
	if err := validateClipRange(highlight); err != nil {
		return s.markClipFailed(highlightID, err.Error())
	}
	if err := s.sourceInputAvailable(analysis.VideoURL); err != nil {
		return s.markClipFailed(highlightID, err.Error())
	}

	if err := s.db.Model(&models.AnalysisHighlight{}).Where("id = ?", highlightID).Updates(map[string]interface{}{
		"clip_status":       models.HighlightClipQueued,
		"clip_error":        "",
		"video_clip_url":    "",
		"clip_generated_at": nil,
	}).Error; err != nil {
		return nil, err
	}
	s.recordClipOperationEvent(analysis, highlight, "clip_generation_started", "片段生成开始", "高光片段已加入生成队列", map[string]interface{}{
		"highlight_id":  highlight.ID,
		"start_time_ms": highlight.StartTimeMs,
		"end_time_ms":   highlight.EndTimeMs,
	})

	go s.ProcessHighlightClip(highlightID)
	return s.FindHighlight(highlightID)
}

func (s *VideoClipService) ProcessHighlightClip(highlightID uint) {
	highlight, analysis, err := s.findHighlightWithAnalysis(highlightID)
	if err != nil {
		return
	}
	if err := validateClipRange(highlight); err != nil {
		_, _ = s.markClipFailed(highlightID, err.Error())
		return
	}
	source, cleanupSource, err := s.resolveSourceInput(analysis.VideoURL)
	if err != nil {
		_, _ = s.markClipFailed(highlightID, err.Error())
		return
	}
	defer cleanupSource()
	if s.ffmpegPath == "" {
		_, _ = s.markClipFailed(highlightID, "视频处理工具 ffmpeg 未安装或未配置")
		return
	}

	version := highlight.ClipVersion + 1
	filename := fmt.Sprintf("analysis_%d_marker_%d_v%d.mp4", analysis.ID, highlight.ID, version)
	outputPath := filepath.Join(s.outputDir, filename)
	durationMs := *highlight.EndTimeMs - highlight.StartTimeMs

	if err := s.db.Model(&models.AnalysisHighlight{}).Where("id = ?", highlightID).Updates(map[string]interface{}{
		"clip_status":    models.HighlightClipProcessing,
		"clip_error":     "",
		"clip_version":   version,
		"video_clip_url": "",
	}).Error; err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	args := []string{
		"-y",
		"-ss", formatFFmpegSeconds(highlight.StartTimeMs),
		"-i", source,
		"-t", formatFFmpegSeconds(durationMs),
		"-c:v", "libx264",
		"-c:a", "aac",
		"-movflags", "+faststart",
		outputPath,
	}
	output, err := exec.CommandContext(ctx, s.ffmpegPath, args...).CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			message = err.Error()
		}
		_, _ = s.markClipFailed(highlightID, truncateClipError(message))
		return
	}
	if _, err := os.Stat(outputPath); err != nil {
		_, _ = s.markClipFailed(highlightID, "剪辑输出文件不存在")
		return
	}

	now := time.Now()
	clipURL := strings.TrimRight(s.baseURL, "/") + s.publicPrefix + "/" + filename
	if s.storageService != nil {
		if uploadedURL, _, err := s.storageService.UploadAnalysisClip(context.Background(), UploadAnalysisClipInput{
			OrderID:     analysis.OrderID,
			AnalysisID:  analysis.ID,
			HighlightID: highlight.ID,
			Version:     version,
			LocalPath:   outputPath,
			ContentType: "video/mp4",
		}); err == nil && strings.TrimSpace(uploadedURL) != "" {
			clipURL = uploadedURL
		} else if err != nil {
			_, _ = s.markClipFailed(highlightID, truncateClipError("上传剪辑到对象存储失败: "+err.Error()))
			return
		}
	}
	if err := s.db.Model(&models.AnalysisHighlight{}).Where("id = ?", highlightID).Updates(map[string]interface{}{
		"clip_status":       models.HighlightClipReady,
		"clip_error":        "",
		"video_clip_url":    clipURL,
		"clip_generated_at": &now,
	}).Error; err != nil {
		_, _ = s.markClipFailed(highlightID, truncateClipError("保存剪辑结果失败: "+err.Error()))
		return
	}
	s.recordClipOperationEvent(analysis, highlight, "clip_generation_completed", "片段生成完成", "高光片段生成完成", map[string]interface{}{
		"highlight_id":      highlight.ID,
		"clip_url":          clipURL,
		"clip_version":      version,
		"clip_generated_at": now.Format(time.RFC3339),
	})
}

func (s *VideoClipService) ClearHighlightClip(highlightID uint) (*models.AnalysisHighlight, error) {
	if err := s.db.Model(&models.AnalysisHighlight{}).Where("id = ?", highlightID).Updates(map[string]interface{}{
		"clip_status":       models.HighlightClipNone,
		"clip_error":        "",
		"video_clip_url":    "",
		"clip_generated_at": nil,
	}).Error; err != nil {
		return nil, err
	}
	return s.FindHighlight(highlightID)
}

func (s *VideoClipService) FindHighlight(highlightID uint) (*models.AnalysisHighlight, error) {
	var highlight models.AnalysisHighlight
	if err := s.db.First(&highlight, highlightID).Error; err != nil {
		return nil, err
	}
	return &highlight, nil
}

func (s *VideoClipService) ResolveClipFilePath(clipURL string) (string, error) {
	if strings.TrimSpace(clipURL) == "" {
		return "", errors.New("片段文件不存在")
	}
	if s.storageService != nil {
		if objectKey, ok := s.storageService.ObjectKeyFromURL(clipURL); ok {
			return s.storageService.DownloadObjectToTemp(context.Background(), objectKey)
		}
	}
	parsed, err := url.Parse(clipURL)
	pathValue := clipURL
	if err == nil && parsed.Path != "" {
		pathValue = parsed.Path
	}
	if !strings.HasPrefix(pathValue, s.publicPrefix+"/") {
		return "", errors.New("片段文件路径无效")
	}
	localPath := filepath.Join(s.outputDir, filepath.Base(pathValue))
	if _, err := os.Stat(localPath); err != nil {
		return "", errors.New("片段文件不存在")
	}
	return localPath, nil
}

func (s *VideoClipService) findHighlightWithAnalysis(highlightID uint) (*models.AnalysisHighlight, *models.VideoAnalysis, error) {
	var highlight models.AnalysisHighlight
	if err := s.db.First(&highlight, highlightID).Error; err != nil {
		return nil, nil, err
	}
	var analysis models.VideoAnalysis
	if err := s.db.First(&analysis, highlight.AnalysisID).Error; err != nil {
		return nil, nil, err
	}
	return &highlight, &analysis, nil
}

func (s *VideoClipService) markClipFailed(highlightID uint, message string) (*models.AnalysisHighlight, error) {
	highlight, analysis, _ := s.findHighlightWithAnalysis(highlightID)
	if err := s.db.Model(&models.AnalysisHighlight{}).Where("id = ?", highlightID).Updates(map[string]interface{}{
		"clip_status":       models.HighlightClipFailed,
		"clip_error":        truncateClipError(message),
		"video_clip_url":    "",
		"clip_generated_at": nil,
	}).Error; err != nil {
		return nil, err
	}
	if highlight != nil && analysis != nil {
		s.recordClipOperationEvent(analysis, highlight, "clip_generation_failed", "片段生成失败", truncateClipError(message), map[string]interface{}{
			"highlight_id": highlight.ID,
			"error":        truncateClipError(message),
		})
	}
	return s.FindHighlight(highlightID)
}

func (s *VideoClipService) recordClipOperationEvent(analysis *models.VideoAnalysis, highlight *models.AnalysisHighlight, eventType, label, summary string, metadata map[string]interface{}) {
	if s == nil || s.db == nil || analysis == nil || highlight == nil {
		return
	}
	if metadata == nil {
		metadata = map[string]interface{}{}
	}
	metadata["timestamp"] = highlight.Timestamp
	metadata["marker_type"] = highlight.MarkerType
	metadata["tag_type"] = highlight.TagType
	data, _ := json.Marshal(metadata)
	if err := models.NewAnalysisOperationEventRepository(s.db).Create(&models.AnalysisOperationEvent{
		OrderID:      analysis.OrderID,
		AnalysisID:   analysis.ID,
		AnalystID:    analysis.AnalystID,
		EventType:    eventType,
		Section:      "clip",
		FieldKey:     fmt.Sprintf("highlight_%d", highlight.ID),
		FieldLabel:   label,
		AfterSummary: summary,
		Metadata:     string(data),
		CreatedAt:    time.Now(),
	}); err != nil {
		log.Printf("[AnalysisOperationEvent] clip event create failed: %v", err)
	}
}

func (s *VideoClipService) sourceInputAvailable(rawURL string) error {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return errors.New("源视频不存在")
	}
	if s.storageService != nil {
		if _, ok := s.storageService.ObjectKeyFromURL(rawURL); ok {
			return nil
		}
	}
	_, cleanup, err := s.resolveSourceInput(rawURL)
	if cleanup != nil {
		cleanup()
	}
	return err
}

func (s *VideoClipService) resolveSourceInput(rawURL string) (string, func(), error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", func() {}, errors.New("源视频不存在")
	}
	if s.storageService != nil {
		if objectKey, ok := s.storageService.ObjectKeyFromURL(rawURL); ok {
			localPath, err := s.storageService.DownloadObjectToTemp(context.Background(), objectKey)
			if err != nil {
				return "", func() {}, errors.New("源视频文件不存在")
			}
			return localPath, func() { _ = os.Remove(localPath) }, nil
		}
	}

	parsed, err := url.Parse(rawURL)
	if err == nil && parsed.Scheme != "" {
		switch parsed.Scheme {
		case "http", "https":
			if strings.HasPrefix(parsed.Path, "/uploads/") {
				localPath := filepath.Clean(strings.TrimPrefix(parsed.Path, "/"))
				if _, err := os.Stat(localPath); err != nil {
					return "", func() {}, errors.New("源视频文件不存在")
				}
				return localPath, func() {}, nil
			}
			return rawURL, func() {}, nil
		case "file":
			if _, err := os.Stat(parsed.Path); err != nil {
				return "", func() {}, errors.New("源视频文件不存在")
			}
			return parsed.Path, func() {}, nil
		default:
			return "", func() {}, errors.New("源视频地址格式不支持")
		}
	}

	localPath := rawURL
	if strings.HasPrefix(rawURL, "/uploads/") {
		localPath = filepath.Clean(strings.TrimPrefix(rawURL, "/"))
	}
	if _, err := os.Stat(localPath); err != nil {
		return "", func() {}, errors.New("源视频文件不存在")
	}
	return localPath, func() {}, nil
}

func validateClipRange(highlight *models.AnalysisHighlight) error {
	if highlight.EndTimeMs == nil || *highlight.EndTimeMs <= highlight.StartTimeMs {
		return errors.New("剪辑时间段无效")
	}
	duration := *highlight.EndTimeMs - highlight.StartTimeMs
	if duration > maxAnalysisClipDurationMs {
		return errors.New("剪辑时间段不能超过 90 秒")
	}
	return nil
}

func formatFFmpegSeconds(ms int) string {
	return fmt.Sprintf("%.3f", float64(ms)/1000)
}

func truncateClipError(message string) string {
	message = strings.TrimSpace(message)
	if len(message) > 500 {
		return message[:500]
	}
	return message
}
