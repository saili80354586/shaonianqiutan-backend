package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	StorageDriverLocal = "local"
	StorageDriverCOS   = "cos"

	defaultCOSBucket       = "shaonianqiutan1-1411539107"
	defaultCOSRegion       = "ap-shanghai"
	defaultUploadExpire    = 15 * time.Minute
	defaultDownloadExpire  = 30 * time.Minute
	defaultDeliveredRetain = 7 * 24 * time.Hour
	defaultCancelledRetain = 24 * time.Hour
)

type StorageConfig struct {
	Driver                string
	Bucket                string
	Region                string
	BucketURL             string
	ObjectPrefix          string
	SecretID              string
	SecretKey             string
	UploadExpire          time.Duration
	DownloadExpire        time.Duration
	DeliveredVideoRetain  time.Duration
	CancelledObjectRetain time.Duration
}

func GetStorageConfig() StorageConfig {
	driver := strings.ToLower(strings.TrimSpace(os.Getenv("STORAGE_DRIVER")))
	if driver == "" {
		driver = StorageDriverLocal
	}

	bucket := firstNonEmptyEnv("COS_BUCKET", defaultCOSBucket)
	region := firstNonEmptyEnv("COS_REGION", defaultCOSRegion)
	bucketURL := strings.TrimRight(strings.TrimSpace(os.Getenv("COS_BUCKET_URL")), "/")
	if bucketURL == "" && bucket != "" && region != "" {
		bucketURL = "https://" + bucket + ".cos." + region + ".myqcloud.com"
	}

	return StorageConfig{
		Driver:                driver,
		Bucket:                bucket,
		Region:                region,
		BucketURL:             bucketURL,
		ObjectPrefix:          strings.Trim(strings.TrimSpace(firstNonEmptyEnv("COS_OBJECT_PREFIX", "prod")), "/"),
		SecretID:              firstNonEmptyEnv("COS_SECRET_ID", os.Getenv("COS_SECRETID")),
		SecretKey:             firstNonEmptyEnv("COS_SECRET_KEY", os.Getenv("COS_SECRETKEY")),
		UploadExpire:          durationFromEnv("COS_UPLOAD_EXPIRE_SECONDS", defaultUploadExpire),
		DownloadExpire:        durationFromEnv("COS_DOWNLOAD_EXPIRE_SECONDS", defaultDownloadExpire),
		DeliveredVideoRetain:  durationFromEnv("ORDER_SOURCE_VIDEO_DELETE_AFTER_DELIVERED_SECONDS", defaultDeliveredRetain),
		CancelledObjectRetain: durationFromEnv("ORDER_CANCELLED_DELETE_AFTER_SECONDS", defaultCancelledRetain),
	}
}

func firstNonEmptyEnv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value != "" {
		return value
	}
	return strings.TrimSpace(fallback)
}

func durationFromEnv(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds <= 0 {
		return fallback
	}
	return time.Duration(seconds) * time.Second
}
