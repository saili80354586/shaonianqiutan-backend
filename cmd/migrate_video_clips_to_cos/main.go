package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/shaonianqiutan/backend/config"
	"github.com/shaonianqiutan/backend/models"
	"github.com/tencentyun/cos-go-sdk-v5"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type clipRow struct {
	HighlightID     uint   `json:"highlight_id"`
	AnalysisID      uint   `json:"analysis_id"`
	OrderID         uint   `json:"order_id"`
	ClipStatus      string `json:"clip_status"`
	VideoClipURL    string `json:"old_url"`
	ClipGeneratedAt string `json:"clip_generated_at"`
	UpdatedAt       string `json:"updated_at"`
}

type manifestItem struct {
	HighlightID  uint   `json:"highlight_id"`
	AnalysisID   uint   `json:"analysis_id"`
	OrderID      uint   `json:"order_id"`
	MarkerID     uint   `json:"marker_id"`
	Version      int    `json:"version"`
	ClipStatus   string `json:"clip_status"`
	OldURL       string `json:"old_url"`
	NewURL       string `json:"new_url,omitempty"`
	LocalPath    string `json:"local_path"`
	Exists       bool   `json:"exists"`
	Size         int64  `json:"size"`
	TargetBucket string `json:"target_bucket"`
	TargetRegion string `json:"target_region"`
	TargetKey    string `json:"target_key"`
	Uploaded     bool   `json:"uploaded"`
	DBUpdated    bool   `json:"db_updated"`
	Error        string `json:"error,omitempty"`
	GeneratedAt  string `json:"clip_generated_at"`
	UpdatedAt    string `json:"updated_at"`
}

type manifest struct {
	GeneratedAt   string         `json:"generated_at"`
	Mode          string         `json:"mode"`
	DBPath        string         `json:"db_path"`
	ClipsDir      string         `json:"clips_dir"`
	TargetBucket  string         `json:"target_bucket"`
	TargetRegion  string         `json:"target_region"`
	TargetPrefix  string         `json:"target_prefix"`
	TotalRecords  int            `json:"total_records"`
	ExistingFiles int            `json:"existing_files"`
	MissingFiles  int            `json:"missing_files"`
	TotalBytes    int64          `json:"total_bytes"`
	Uploaded      int            `json:"uploaded"`
	DBUpdated     int            `json:"db_updated"`
	Failed        int            `json:"failed"`
	Items         []manifestItem `json:"items"`
}

var clipNamePattern = regexp.MustCompile(`^analysis_(\d+)_marker_(\d+)_v(\d+)\.mp4$`)

func main() {
	var dbPathFlag string
	var clipsDir string
	var execute bool
	var includeAll bool
	var output string
	flag.StringVar(&dbPathFlag, "db", "", "SQLite database path. Defaults to DB_PATH or ./shaonianqiutan.db")
	flag.StringVar(&clipsDir, "clips-dir", "./uploads/video-clips", "local video clips directory")
	flag.BoolVar(&execute, "execute", false, "upload files to COS and update database URLs")
	flag.BoolVar(&includeAll, "include-unreferenced", false, "include unreferenced local MP4 files in dry-run output only")
	flag.StringVar(&output, "output", "", "optional JSON manifest output path")
	flag.Parse()

	config.LoadEnv()
	storageCfg := config.GetStorageConfig()
	dbPath := strings.TrimSpace(dbPathFlag)
	if dbPath == "" {
		dbPath = strings.TrimSpace(os.Getenv("DB_PATH"))
	}
	if dbPath == "" {
		dbPath = "./shaonianqiutan.db"
	}

	if execute {
		if err := validateCOSConfig(storageCfg); err != nil {
			log.Fatalf("COS 配置不完整，停止执行: %v", err)
		}
		includeAll = false
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("打开数据库失败: %v", err)
	}
	if execute {
		if err := db.AutoMigrate(&models.StorageObject{}); err != nil {
			log.Fatalf("迁移 storage_objects 失败: %v", err)
		}
	}

	items, err := buildManifestItems(db, clipsDir, storageCfg, includeAll)
	if err != nil {
		log.Fatalf("生成迁移清单失败: %v", err)
	}

	result := manifest{
		GeneratedAt:  time.Now().Format(time.RFC3339),
		Mode:         "dry-run",
		DBPath:       dbPath,
		ClipsDir:     clipsDir,
		TargetBucket: storageCfg.Bucket,
		TargetRegion: storageCfg.Region,
		TargetPrefix: storageCfg.ObjectPrefix,
		TotalRecords: len(items),
		Items:        items,
	}
	if execute {
		result.Mode = "execute"
		cosClient := newCOSClient(storageCfg)
		for i := range result.Items {
			if err := migrateOne(context.Background(), db, cosClient, storageCfg, &result.Items[i]); err != nil {
				result.Items[i].Error = err.Error()
				result.Failed++
				continue
			}
			if result.Items[i].Uploaded {
				result.Uploaded++
			}
			if result.Items[i].DBUpdated {
				result.DBUpdated++
			}
		}
	}
	summarize(&result)

	payload, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatalf("生成 JSON 失败: %v", err)
	}
	fmt.Println(string(payload))
	if output != "" {
		if err := os.WriteFile(output, append(payload, '\n'), 0644); err != nil {
			log.Fatalf("写入清单失败: %v", err)
		}
	}
	if result.Failed > 0 {
		os.Exit(1)
	}
}

func buildManifestItems(db *gorm.DB, clipsDir string, storageCfg config.StorageConfig, includeAll bool) ([]manifestItem, error) {
	var rows []clipRow
	if err := db.Raw(`
select h.id as highlight_id,
       h.analysis_id as analysis_id,
       va.order_id as order_id,
       h.clip_status as clip_status,
       h.video_clip_url as video_clip_url,
       h.clip_generated_at as clip_generated_at,
       h.updated_at as updated_at
from analysis_highlights h
left join video_analyses va on va.id = h.analysis_id
where h.video_clip_url is not null and trim(h.video_clip_url) <> ''
  and h.video_clip_url like '%/uploads/video-clips/%'
order by datetime(h.updated_at) desc, h.id desc
`).Scan(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]manifestItem, 0, len(rows))
	referenced := make(map[string]bool, len(rows))
	for _, row := range rows {
		filename := filepath.Base(row.VideoClipURL)
		referenced[filename] = true
		item := buildItemFromRow(row, clipsDir, storageCfg, filename)
		items = append(items, item)
	}

	if includeAll {
		files, _ := filepath.Glob(filepath.Join(clipsDir, "*.mp4"))
		for _, path := range files {
			filename := filepath.Base(path)
			if referenced[filename] {
				continue
			}
			item := buildUnreferencedItem(path, storageCfg)
			items = append(items, item)
		}
	}

	return items, nil
}

func buildItemFromRow(row clipRow, clipsDir string, storageCfg config.StorageConfig, filename string) manifestItem {
	localPath := filepath.Join(clipsDir, filename)
	info, statErr := os.Stat(localPath)
	markerID, version := parseMarkerAndVersion(filename, row.HighlightID)
	targetKey := ""
	if row.OrderID > 0 {
		targetKey = targetClipKey(storageCfg.ObjectPrefix, row.OrderID, row.AnalysisID, markerID, version)
	}
	item := manifestItem{
		HighlightID:  row.HighlightID,
		AnalysisID:   row.AnalysisID,
		OrderID:      row.OrderID,
		MarkerID:     markerID,
		Version:      version,
		ClipStatus:   row.ClipStatus,
		OldURL:       row.VideoClipURL,
		LocalPath:    localPath,
		Exists:       statErr == nil,
		TargetBucket: storageCfg.Bucket,
		TargetRegion: storageCfg.Region,
		TargetKey:    targetKey,
		GeneratedAt:  row.ClipGeneratedAt,
		UpdatedAt:    row.UpdatedAt,
	}
	if statErr == nil {
		item.Size = info.Size()
	}
	if targetKey != "" && storageCfg.BucketURL != "" {
		item.NewURL = strings.TrimRight(storageCfg.BucketURL, "/") + "/" + targetKey
	}
	return item
}

func buildUnreferencedItem(localPath string, storageCfg config.StorageConfig) manifestItem {
	info, statErr := os.Stat(localPath)
	filename := filepath.Base(localPath)
	analysisID, markerID, version := parseClipFilename(filename)
	item := manifestItem{
		AnalysisID:   analysisID,
		MarkerID:     markerID,
		Version:      version,
		ClipStatus:   "unreferenced",
		LocalPath:    localPath,
		Exists:       statErr == nil,
		TargetBucket: storageCfg.Bucket,
		TargetRegion: storageCfg.Region,
	}
	if statErr == nil {
		item.Size = info.Size()
	}
	return item
}

func migrateOne(ctx context.Context, db *gorm.DB, cosClient *cos.Client, storageCfg config.StorageConfig, item *manifestItem) error {
	if item.HighlightID == 0 {
		return errors.New("未引用文件不会在执行模式迁移")
	}
	if !item.Exists {
		return errors.New("本地剪辑文件不存在")
	}
	if item.OrderID == 0 || item.AnalysisID == 0 || item.TargetKey == "" {
		return errors.New("缺少 order_id/analysis_id/target_key")
	}
	resp, err := cosClient.Object.PutFromFile(ctx, item.TargetKey, item.LocalPath, &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{ContentType: "video/mp4"},
	})
	if err != nil {
		return fmt.Errorf("上传 COS 失败: %w", err)
	}
	item.Uploaded = true

	head, err := cosClient.Object.Head(ctx, item.TargetKey, nil)
	if err != nil {
		return fmt.Errorf("校验 COS 对象失败: %w", err)
	}
	if head != nil && head.ContentLength > 0 && head.ContentLength != item.Size {
		return fmt.Errorf("COS 对象大小不一致: got %d want %d", head.ContentLength, item.Size)
	}

	etag := ""
	if resp != nil {
		etag = strings.Trim(resp.Header.Get("ETag"), "\"")
	}
	object := models.StorageObject{
		Bucket:       storageCfg.Bucket,
		Region:       storageCfg.Region,
		ObjectKey:    item.TargetKey,
		OriginalName: filepath.Base(item.LocalPath),
		ContentType:  "video/mp4",
		Size:         item.Size,
		ETag:         etag,
		OwnerType:    models.StorageOwnerVideoAnalysis,
		OwnerID:      item.AnalysisID,
		BusinessType: models.StorageBusinessAnalysisClip,
		Status:       models.StorageObjectStatusActive,
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "bucket"}, {Name: "object_key"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"size":          object.Size,
				"e_tag":         object.ETag,
				"status":        object.Status,
				"business_type": object.BusinessType,
				"owner_type":    object.OwnerType,
				"owner_id":      object.OwnerID,
				"updated_at":    time.Now(),
			}),
		}).Create(&object).Error; err != nil {
			return err
		}
		updateResult := tx.Model(&models.AnalysisHighlight{}).
			Where("id = ? AND video_clip_url = ?", item.HighlightID, item.OldURL).
			Update("video_clip_url", item.NewURL)
		if updateResult.Error != nil {
			return updateResult.Error
		}
		if updateResult.RowsAffected != 1 {
			return fmt.Errorf("更新剪辑链接失败: highlight_id=%d rows=%d", item.HighlightID, updateResult.RowsAffected)
		}
		item.DBUpdated = true
		return nil
	})
}

func summarize(result *manifest) {
	for _, item := range result.Items {
		if item.Exists {
			result.ExistingFiles++
			result.TotalBytes += item.Size
		} else {
			result.MissingFiles++
		}
		if item.Error != "" && result.Mode != "execute" {
			result.Failed++
		}
	}
}

func validateCOSConfig(storageCfg config.StorageConfig) error {
	if storageCfg.Driver != config.StorageDriverCOS {
		return fmt.Errorf("STORAGE_DRIVER must be %q", config.StorageDriverCOS)
	}
	missing := []string{}
	if storageCfg.Bucket == "" {
		missing = append(missing, "COS_BUCKET")
	}
	if storageCfg.Region == "" {
		missing = append(missing, "COS_REGION")
	}
	if storageCfg.BucketURL == "" {
		missing = append(missing, "COS_BUCKET_URL")
	}
	if storageCfg.SecretID == "" {
		missing = append(missing, "COS_SECRET_ID")
	}
	if storageCfg.SecretKey == "" {
		missing = append(missing, "COS_SECRET_KEY")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing %s", strings.Join(missing, ", "))
	}
	return nil
}

func newCOSClient(storageCfg config.StorageConfig) *cos.Client {
	parsed, err := url.Parse(storageCfg.BucketURL)
	if err != nil {
		log.Fatalf("COS_BUCKET_URL 无效: %v", err)
	}
	return cos.NewClient(&cos.BaseURL{BucketURL: parsed}, &http.Client{
		Timeout: 60 * time.Second,
		Transport: &cos.AuthorizationTransport{
			SecretID:  storageCfg.SecretID,
			SecretKey: storageCfg.SecretKey,
		},
	})
}

func parseMarkerAndVersion(filename string, fallbackHighlightID uint) (uint, int) {
	_, markerID, version := parseClipFilename(filename)
	if markerID == 0 {
		markerID = fallbackHighlightID
	}
	if version == 0 {
		version = 1
	}
	return markerID, version
}

func parseClipFilename(filename string) (uint, uint, int) {
	match := clipNamePattern.FindStringSubmatch(filename)
	if len(match) != 4 {
		return 0, 0, 0
	}
	return parseUint(match[1]), parseUint(match[2]), int(parseUint(match[3]))
}

func parseUint(value string) uint {
	var parsed uint64
	_, _ = fmt.Sscanf(value, "%d", &parsed)
	return uint(parsed)
}

func targetClipKey(prefix string, orderID, analysisID, markerID uint, version int) string {
	prefix = strings.Trim(strings.TrimSpace(prefix), "/")
	if prefix == "" {
		prefix = "prod"
	}
	if version <= 0 {
		version = 1
	}
	return fmt.Sprintf("%s/orders/%d/analysis/%d/clips/marker_%d_v%d.mp4", prefix, orderID, analysisID, markerID, version)
}
