package controllers

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/models"
)

func TestCreateHighlightSupportsRangeMarkerMetadata(t *testing.T) {
	db, ctrl, owner, other := setupVideoAnalysisControllerTest(t)

	analysis := models.VideoAnalysis{
		OrderID:        10,
		AnalystID:      owner.ID,
		UserID:         100,
		PlayerName:     "Marker Player",
		PlayerPosition: "winger",
		Status:         models.AnalysisStatusScoring,
	}
	if err := db.Create(&analysis).Error; err != nil {
		t.Fatalf("create analysis: %v", err)
	}

	includeInReport := true
	endTimeMs := 75000
	req := CreateHighlightRequest{
		AnalysisID:      analysis.ID,
		Timestamp:       "01:00-01:15",
		MarkerType:      models.HighlightMarkerIssue,
		Mode:            models.HighlightModeRange,
		StartTimeMs:     60000,
		EndTimeMs:       &endTimeMs,
		TagType:         models.HighlightPositioningError,
		Description:     "回防阶段站位偏慢，未及时保护中路空间。",
		IncludeInReport: &includeInReport,
	}

	forbidden := performVideoAnalysisRequest(t, other.ID, http.MethodPost, "/video-analysis/highlights", req, nil, ctrl.CreateHighlight)
	if forbidden.Code != http.StatusForbidden {
		t.Fatalf("other analyst status = %d, want %d", forbidden.Code, http.StatusForbidden)
	}

	allowed := performVideoAnalysisRequest(t, owner.ID, http.MethodPost, "/video-analysis/highlights", req, nil, ctrl.CreateHighlight)
	if allowed.Code != http.StatusOK {
		t.Fatalf("owner analyst status = %d, want %d, body=%s", allowed.Code, http.StatusOK, allowed.Body.String())
	}

	var marker models.AnalysisHighlight
	if err := db.Where("analysis_id = ?", analysis.ID).First(&marker).Error; err != nil {
		t.Fatalf("find marker: %v", err)
	}
	if marker.MarkerType != models.HighlightMarkerIssue {
		t.Fatalf("marker_type = %s, want %s", marker.MarkerType, models.HighlightMarkerIssue)
	}
	if marker.Mode != models.HighlightModeRange {
		t.Fatalf("mode = %s, want %s", marker.Mode, models.HighlightModeRange)
	}
	if marker.StartTimeMs != 60000 || marker.EndTimeMs == nil || *marker.EndTimeMs != endTimeMs {
		t.Fatalf("time range = %d-%v, want 60000-%d", marker.StartTimeMs, marker.EndTimeMs, endTimeMs)
	}
	if marker.TagType != models.HighlightPositioningError {
		t.Fatalf("tag_type = %s, want %s", marker.TagType, models.HighlightPositioningError)
	}

	params := gin.Params{{Key: "id", Value: strconv.Itoa(int(marker.ID))}}
	invalidEndTimeMs := 55000
	invalidReq := req
	invalidReq.EndTimeMs = &invalidEndTimeMs
	invalid := performVideoAnalysisRequest(t, owner.ID, http.MethodPut, "/video-analysis/highlights/"+params[0].Value, invalidReq, params, ctrl.UpdateHighlight)
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("invalid range status = %d, want %d", invalid.Code, http.StatusBadRequest)
	}

	updatedEndTimeMs := 80000
	updateReq := req
	updateReq.Timestamp = "01:05-01:20"
	updateReq.MarkerType = models.HighlightMarkerObservation
	updateReq.Mode = models.HighlightModeRange
	updateReq.StartTimeMs = 65000
	updateReq.EndTimeMs = &updatedEndTimeMs
	updateReq.TagType = models.HighlightTacticalNote
	updateReq.Description = "更新后的战术观察。"

	updated := performVideoAnalysisRequest(t, owner.ID, http.MethodPut, "/video-analysis/highlights/"+params[0].Value, updateReq, params, ctrl.UpdateHighlight)
	if updated.Code != http.StatusOK {
		t.Fatalf("update range status = %d, want %d, body=%s", updated.Code, http.StatusOK, updated.Body.String())
	}
	var updateBody struct {
		Success bool                     `json:"success"`
		Data    models.AnalysisHighlight `json:"data"`
	}
	if err := json.Unmarshal(updated.Body.Bytes(), &updateBody); err != nil {
		t.Fatalf("decode update body: %v", err)
	}
	if !updateBody.Success || updateBody.Data.ID != marker.ID {
		t.Fatalf("update body = %#v, want updated marker id %d", updateBody, marker.ID)
	}
	if updateBody.Data.MarkerType != models.HighlightMarkerObservation || updateBody.Data.TagType != models.HighlightTacticalNote {
		t.Fatalf("updated marker metadata = %s/%s", updateBody.Data.MarkerType, updateBody.Data.TagType)
	}
}

func TestRangeHighlightClipFailsWhenSourceVideoMissing(t *testing.T) {
	db, ctrl, owner, _ := setupVideoAnalysisControllerTest(t)

	analysis := models.VideoAnalysis{
		OrderID:        11,
		AnalystID:      owner.ID,
		UserID:         101,
		PlayerName:     "Missing Source Player",
		PlayerPosition: "midfielder",
		VideoURL:       "/uploads/videos/missing-source.mp4",
		Status:         models.AnalysisStatusScoring,
	}
	if err := db.Create(&analysis).Error; err != nil {
		t.Fatalf("create analysis: %v", err)
	}

	endTimeMs := 12000
	req := CreateHighlightRequest{
		AnalysisID:  analysis.ID,
		Mode:        models.HighlightModeRange,
		StartTimeMs: 5000,
		EndTimeMs:   &endTimeMs,
		TagType:     models.HighlightPass,
		Description: "源视频缺失的时间段。",
	}

	res := performVideoAnalysisRequest(t, owner.ID, http.MethodPost, "/video-analysis/highlights", req, nil, ctrl.CreateHighlight)
	if res.Code != http.StatusOK {
		t.Fatalf("create highlight status = %d, want %d, body=%s", res.Code, http.StatusOK, res.Body.String())
	}

	var marker models.AnalysisHighlight
	if err := db.Where("analysis_id = ?", analysis.ID).First(&marker).Error; err != nil {
		t.Fatalf("find marker: %v", err)
	}
	if marker.ClipStatus != models.HighlightClipFailed {
		t.Fatalf("clip_status = %s, want %s", marker.ClipStatus, models.HighlightClipFailed)
	}
	if marker.ClipError == "" {
		t.Fatal("clip_error should explain source video failure")
	}
}

func TestRangeHighlightQueuesAndGeneratesClip(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "source.mp4")
	if err := os.WriteFile(sourcePath, []byte("fake-source"), 0644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	ffmpegPath := filepath.Join(tempDir, "ffmpeg")
	ffmpegScript := "#!/bin/sh\nfor last do :; done\nmkdir -p \"$(dirname \"$last\")\"\nprintf 'fake-clip' > \"$last\"\n"
	if err := os.WriteFile(ffmpegPath, []byte(ffmpegScript), 0755); err != nil {
		t.Fatalf("write fake ffmpeg: %v", err)
	}

	t.Setenv("FFMPEG_PATH", ffmpegPath)
	t.Setenv("VIDEO_CLIP_OUTPUT_DIR", filepath.Join(tempDir, "clips"))
	t.Setenv("BASE_URL", "http://localhost:8080")

	db, ctrl, owner, _ := setupVideoAnalysisControllerTest(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)

	analysis := models.VideoAnalysis{
		OrderID:        12,
		AnalystID:      owner.ID,
		UserID:         102,
		PlayerName:     "Clip Player",
		PlayerPosition: "forward",
		VideoURL:       sourcePath,
		Status:         models.AnalysisStatusScoring,
	}
	if err := db.Create(&analysis).Error; err != nil {
		t.Fatalf("create analysis: %v", err)
	}

	endTimeMs := 15000
	req := CreateHighlightRequest{
		AnalysisID:  analysis.ID,
		Mode:        models.HighlightModeRange,
		StartTimeMs: 3000,
		EndTimeMs:   &endTimeMs,
		TagType:     models.HighlightGoal,
		Description: "可生成片段的时间段。",
	}
	res := performVideoAnalysisRequest(t, owner.ID, http.MethodPost, "/video-analysis/highlights", req, nil, ctrl.CreateHighlight)
	if res.Code != http.StatusOK {
		t.Fatalf("create highlight status = %d, want %d, body=%s", res.Code, http.StatusOK, res.Body.String())
	}

	var marker models.AnalysisHighlight
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if err := db.Where("analysis_id = ?", analysis.ID).First(&marker).Error; err != nil {
			t.Fatalf("find marker: %v", err)
		}
		if marker.ClipStatus == models.HighlightClipReady {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if marker.ClipStatus != models.HighlightClipReady {
		t.Fatalf("clip_status = %s, want %s, error=%s", marker.ClipStatus, models.HighlightClipReady, marker.ClipError)
	}
	if marker.VideoClipURL == "" {
		t.Fatal("video_clip_url should be set")
	}
	if _, err := os.Stat(filepath.Join(tempDir, "clips", "analysis_"+strconv.Itoa(int(analysis.ID))+"_marker_"+strconv.Itoa(int(marker.ID))+"_v1.mp4")); err != nil {
		t.Fatalf("generated clip missing: %v", err)
	}
}

func TestExportHighlightClipsIncludesReadyClipsAndManifest(t *testing.T) {
	tempDir := t.TempDir()
	clipDir := filepath.Join(tempDir, "clips")
	if err := os.MkdirAll(clipDir, 0755); err != nil {
		t.Fatalf("make clip dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(clipDir, "highlight.mp4"), []byte("highlight-clip"), 0644); err != nil {
		t.Fatalf("write highlight clip: %v", err)
	}
	if err := os.WriteFile(filepath.Join(clipDir, "issue.mp4"), []byte("issue-clip"), 0644); err != nil {
		t.Fatalf("write issue clip: %v", err)
	}

	t.Setenv("VIDEO_CLIP_OUTPUT_DIR", clipDir)
	t.Setenv("BASE_URL", "http://localhost:8080")

	db, ctrl, owner, other := setupVideoAnalysisControllerTest(t)
	analysis := models.VideoAnalysis{
		OrderID:        13,
		AnalystID:      owner.ID,
		UserID:         103,
		PlayerName:     "Export Player",
		PlayerPosition: "midfielder",
		Status:         models.AnalysisStatusScoring,
	}
	if err := db.Create(&analysis).Error; err != nil {
		t.Fatalf("create analysis: %v", err)
	}

	highlightEnd := 15000
	issueEnd := 26000
	pendingEnd := 36000
	readyHighlight := models.AnalysisHighlight{
		AnalysisID:      analysis.ID,
		Timestamp:       "00:05-00:15",
		MarkerType:      models.HighlightMarkerHighlight,
		Mode:            models.HighlightModeRange,
		StartTimeMs:     5000,
		EndTimeMs:       &highlightEnd,
		TagType:         models.HighlightGoal,
		Description:     "连续突破后完成射门。",
		VideoClipURL:    "http://localhost:8080/uploads/video-clips/highlight.mp4",
		ClipStatus:      models.HighlightClipReady,
		IncludeInReport: true,
		ClipGeneratedAt: ptrTime(time.Now()),
	}
	readyIssue := models.AnalysisHighlight{
		AnalysisID:      analysis.ID,
		Timestamp:       "00:20-00:26",
		MarkerType:      models.HighlightMarkerIssue,
		Mode:            models.HighlightModeRange,
		StartTimeMs:     20000,
		EndTimeMs:       &issueEnd,
		TagType:         models.HighlightPositioningError,
		Description:     "防守站位偏慢。",
		VideoClipURL:    "http://localhost:8080/uploads/video-clips/issue.mp4",
		ClipStatus:      models.HighlightClipReady,
		IncludeInReport: false,
	}
	pending := models.AnalysisHighlight{
		AnalysisID:  analysis.ID,
		Timestamp:   "00:30-00:36",
		MarkerType:  models.HighlightMarkerObservation,
		Mode:        models.HighlightModeRange,
		StartTimeMs: 30000,
		EndTimeMs:   &pendingEnd,
		TagType:     models.HighlightTacticalNote,
		Description: "仍在生成的片段。",
		ClipStatus:  models.HighlightClipQueued,
	}
	if err := db.Create(&[]models.AnalysisHighlight{readyHighlight, readyIssue, pending}).Error; err != nil {
		t.Fatalf("create highlights: %v", err)
	}

	params := gin.Params{{Key: "id", Value: strconv.Itoa(int(analysis.ID))}}
	forbidden := performVideoAnalysisRequest(t, other.ID, http.MethodPost, "/video-analysis/"+params[0].Value+"/clips/export", ExportHighlightClipsRequest{}, params, ctrl.ExportHighlightClips)
	if forbidden.Code != http.StatusForbidden {
		t.Fatalf("other analyst export status = %d, want %d", forbidden.Code, http.StatusForbidden)
	}

	res := performVideoAnalysisRequest(t, owner.ID, http.MethodPost, "/video-analysis/"+params[0].Value+"/clips/export", ExportHighlightClipsRequest{}, params, ctrl.ExportHighlightClips)
	if res.Code != http.StatusOK {
		t.Fatalf("export status = %d, want %d, body=%s", res.Code, http.StatusOK, res.Body.String())
	}
	if got := res.Header().Get("Content-Type"); !strings.Contains(got, "application/zip") {
		t.Fatalf("content-type = %q, want application/zip", got)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(res.Body.Bytes()), int64(res.Body.Len()))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	entries := make(map[string]string)
	for _, file := range zipReader.File {
		rc, err := file.Open()
		if err != nil {
			t.Fatalf("open zip entry %s: %v", file.Name, err)
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("read zip entry %s: %v", file.Name, err)
		}
		entries[file.Name] = string(data)
	}
	if len(entries) != 3 {
		t.Fatalf("zip entries = %#v, want 2 clips + manifest", entries)
	}
	if _, ok := entries["markers_manifest.csv"]; !ok {
		t.Fatalf("manifest missing from zip entries: %#v", entries)
	}
	if !strings.Contains(entries["markers_manifest.csv"], "精彩表现") || !strings.Contains(entries["markers_manifest.csv"], "待改进问题") {
		t.Fatalf("manifest content missing marker labels: %q", entries["markers_manifest.csv"])
	}
	if !hasZipEntryPrefix(entries, "01_精彩表现_进球_0m05s-0m15s") {
		t.Fatalf("highlight clip filename missing from entries: %#v", entries)
	}
	if !hasZipEntryPrefix(entries, "02_待改进问题_站位问题_0m20s-0m26s") {
		t.Fatalf("issue clip filename missing from entries: %#v", entries)
	}
}

func TestExportHighlightClipsSupportsMarkerTypeAndSelectedIDs(t *testing.T) {
	tempDir := t.TempDir()
	clipDir := filepath.Join(tempDir, "clips")
	if err := os.MkdirAll(clipDir, 0755); err != nil {
		t.Fatalf("make clip dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(clipDir, "highlight.mp4"), []byte("highlight-clip"), 0644); err != nil {
		t.Fatalf("write highlight clip: %v", err)
	}
	if err := os.WriteFile(filepath.Join(clipDir, "issue.mp4"), []byte("issue-clip"), 0644); err != nil {
		t.Fatalf("write issue clip: %v", err)
	}

	t.Setenv("VIDEO_CLIP_OUTPUT_DIR", clipDir)
	t.Setenv("BASE_URL", "http://localhost:8080")

	db, ctrl, owner, _ := setupVideoAnalysisControllerTest(t)
	analysis := models.VideoAnalysis{
		OrderID:        14,
		AnalystID:      owner.ID,
		UserID:         104,
		PlayerName:     "Selected Export Player",
		PlayerPosition: "forward",
		Status:         models.AnalysisStatusScoring,
	}
	otherAnalysis := models.VideoAnalysis{
		OrderID:        15,
		AnalystID:      owner.ID,
		UserID:         105,
		PlayerName:     "Other Analysis Player",
		PlayerPosition: "forward",
		Status:         models.AnalysisStatusScoring,
	}
	if err := db.Create(&analysis).Error; err != nil {
		t.Fatalf("create analysis: %v", err)
	}
	if err := db.Create(&otherAnalysis).Error; err != nil {
		t.Fatalf("create other analysis: %v", err)
	}

	highlightEnd := 11000
	issueEnd := 24000
	foreignEnd := 34000
	readyHighlight := models.AnalysisHighlight{
		AnalysisID:      analysis.ID,
		Timestamp:       "00:03-00:11",
		MarkerType:      models.HighlightMarkerHighlight,
		Mode:            models.HighlightModeRange,
		StartTimeMs:     3000,
		EndTimeMs:       &highlightEnd,
		TagType:         models.HighlightGoal,
		Description:     "正向片段。",
		VideoClipURL:    "http://localhost:8080/uploads/video-clips/highlight.mp4",
		ClipStatus:      models.HighlightClipReady,
		IncludeInReport: true,
	}
	readyIssue := models.AnalysisHighlight{
		AnalysisID:      analysis.ID,
		Timestamp:       "00:20-00:24",
		MarkerType:      models.HighlightMarkerIssue,
		Mode:            models.HighlightModeRange,
		StartTimeMs:     20000,
		EndTimeMs:       &issueEnd,
		TagType:         models.HighlightTurnover,
		Description:     "问题片段。",
		VideoClipURL:    "http://localhost:8080/uploads/video-clips/issue.mp4",
		ClipStatus:      models.HighlightClipReady,
		IncludeInReport: true,
	}
	foreignHighlight := models.AnalysisHighlight{
		AnalysisID:      otherAnalysis.ID,
		Timestamp:       "00:30-00:34",
		MarkerType:      models.HighlightMarkerHighlight,
		Mode:            models.HighlightModeRange,
		StartTimeMs:     30000,
		EndTimeMs:       &foreignEnd,
		TagType:         models.HighlightGoal,
		Description:     "其他分析片段。",
		VideoClipURL:    "http://localhost:8080/uploads/video-clips/highlight.mp4",
		ClipStatus:      models.HighlightClipReady,
		IncludeInReport: true,
	}
	if err := db.Create(&readyHighlight).Error; err != nil {
		t.Fatalf("create ready highlight: %v", err)
	}
	if err := db.Create(&readyIssue).Error; err != nil {
		t.Fatalf("create ready issue: %v", err)
	}
	if err := db.Create(&foreignHighlight).Error; err != nil {
		t.Fatalf("create foreign highlight: %v", err)
	}

	params := gin.Params{{Key: "id", Value: strconv.Itoa(int(analysis.ID))}}
	filtered := performVideoAnalysisRequest(t, owner.ID, http.MethodPost, "/video-analysis/"+params[0].Value+"/clips/export", ExportHighlightClipsRequest{
		MarkerType: models.HighlightMarkerIssue,
	}, params, ctrl.ExportHighlightClips)
	if filtered.Code != http.StatusOK {
		t.Fatalf("filtered export status = %d, want %d, body=%s", filtered.Code, http.StatusOK, filtered.Body.String())
	}
	filteredZip, err := zip.NewReader(bytes.NewReader(filtered.Body.Bytes()), int64(filtered.Body.Len()))
	if err != nil {
		t.Fatalf("open filtered zip: %v", err)
	}
	filteredEntries := map[string]string{}
	for _, file := range filteredZip.File {
		filteredEntries[file.Name] = file.Name
	}
	if !hasZipEntryPrefix(filteredEntries, "01_待改进问题_失误_0m20s-0m24s") {
		t.Fatalf("filtered issue clip missing: %#v", filteredEntries)
	}
	if hasZipEntryPrefix(filteredEntries, "01_精彩表现") {
		t.Fatalf("filtered zip unexpectedly contains highlight clip: %#v", filteredEntries)
	}

	invalidSelection := performVideoAnalysisRequest(t, owner.ID, http.MethodPost, "/video-analysis/"+params[0].Value+"/clips/export", ExportHighlightClipsRequest{
		MarkerIDs: []uint{foreignHighlight.ID},
	}, params, ctrl.ExportHighlightClips)
	if invalidSelection.Code != http.StatusBadRequest {
		t.Fatalf("foreign marker selection status = %d, want %d", invalidSelection.Code, http.StatusBadRequest)
	}
}

func TestHighlightClipsExportJobTracksProgressAndDownloadsZip(t *testing.T) {
	tempDir := t.TempDir()
	clipDir := filepath.Join(tempDir, "clips")
	if err := os.MkdirAll(clipDir, 0755); err != nil {
		t.Fatalf("make clip dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(clipDir, "job-ready.mp4"), []byte("job-ready-clip"), 0644); err != nil {
		t.Fatalf("write ready clip: %v", err)
	}

	t.Setenv("VIDEO_CLIP_OUTPUT_DIR", clipDir)
	t.Setenv("BASE_URL", "http://localhost:8080")

	db, ctrl, owner, _ := setupVideoAnalysisControllerTest(t)
	analysis := models.VideoAnalysis{
		OrderID:        16,
		AnalystID:      owner.ID,
		UserID:         106,
		PlayerName:     "Async Export Player",
		PlayerPosition: "midfielder",
		Status:         models.AnalysisStatusScoring,
	}
	if err := db.Create(&analysis).Error; err != nil {
		t.Fatalf("create analysis: %v", err)
	}

	endTime := 13000
	highlight := models.AnalysisHighlight{
		AnalysisID:      analysis.ID,
		Timestamp:       "00:04-00:13",
		MarkerType:      models.HighlightMarkerHighlight,
		Mode:            models.HighlightModeRange,
		StartTimeMs:     4000,
		EndTimeMs:       &endTime,
		TagType:         models.HighlightGoal,
		Description:     "后台任务导出片段。",
		VideoClipURL:    "http://localhost:8080/uploads/video-clips/job-ready.mp4",
		ClipStatus:      models.HighlightClipReady,
		IncludeInReport: true,
	}
	if err := db.Create(&highlight).Error; err != nil {
		t.Fatalf("create highlight: %v", err)
	}

	params := gin.Params{{Key: "id", Value: strconv.Itoa(int(analysis.ID))}}
	created := performVideoAnalysisRequest(t, owner.ID, http.MethodPost, "/video-analysis/"+params[0].Value+"/clips/export/jobs", ExportHighlightClipsRequest{}, params, ctrl.CreateHighlightClipsExportJob)
	if created.Code != http.StatusOK {
		t.Fatalf("create export job status = %d, want %d, body=%s", created.Code, http.StatusOK, created.Body.String())
	}

	var createdBody struct {
		Data HighlightClipExportJobResponse `json:"data"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &createdBody); err != nil {
		t.Fatalf("decode created job: %v", err)
	}
	if createdBody.Data.ID == "" {
		t.Fatalf("created job id is empty")
	}

	job := waitForClipExportJobStatus(t, ctrl, owner.ID, analysis.ID, createdBody.Data.ID, highlightClipExportReady)
	if job.Progress != 100 || job.DownloadURL == "" {
		t.Fatalf("ready job = %#v, want progress 100 and download URL", job)
	}
	var persisted models.VideoClipExportJob
	if err := db.Where("job_id = ?", job.ID).First(&persisted).Error; err != nil {
		t.Fatalf("find persisted export job: %v", err)
	}
	if persisted.Status != models.VideoClipExportReady || persisted.ZipPath == "" || persisted.RequestJSON == "" {
		t.Fatalf("persisted job = %#v, want ready status, zip path and request", persisted)
	}

	downloadParams := gin.Params{
		{Key: "id", Value: strconv.Itoa(int(analysis.ID))},
		{Key: "job_id", Value: createdBody.Data.ID},
	}
	download := performVideoAnalysisRequest(t, owner.ID, http.MethodGet, "/video-analysis/"+params[0].Value+"/clips/export/jobs/"+createdBody.Data.ID+"/download", nil, downloadParams, ctrl.DownloadHighlightClipsExportJob)
	if download.Code != http.StatusOK {
		t.Fatalf("download job status = %d, want %d, body=%s", download.Code, http.StatusOK, download.Body.String())
	}
	zipReader, err := zip.NewReader(bytes.NewReader(download.Body.Bytes()), int64(download.Body.Len()))
	if err != nil {
		t.Fatalf("open job zip: %v", err)
	}
	foundManifest := false
	foundClip := false
	for _, file := range zipReader.File {
		if file.Name == "markers_manifest.csv" {
			foundManifest = true
		}
		if strings.HasPrefix(file.Name, "01_精彩表现_进球_0m04s-0m13s") {
			foundClip = true
		}
	}
	if !foundManifest || !foundClip {
		t.Fatalf("job zip missing manifest=%v clip=%v", foundManifest, foundClip)
	}
}

func TestHighlightClipsExportJobRetryRequeuesFailedJob(t *testing.T) {
	clipDir := t.TempDir()
	t.Setenv("VIDEO_CLIP_OUTPUT_DIR", clipDir)
	t.Setenv("BASE_URL", "http://localhost:8080")

	db, ctrl, owner, _ := setupVideoAnalysisControllerTest(t)
	analysis := models.VideoAnalysis{
		OrderID:        17,
		AnalystID:      owner.ID,
		UserID:         107,
		PlayerName:     "Retry Export Player",
		PlayerPosition: "midfielder",
		Status:         models.AnalysisStatusScoring,
	}
	if err := db.Create(&analysis).Error; err != nil {
		t.Fatalf("create analysis: %v", err)
	}

	endTime := 9000
	highlight := models.AnalysisHighlight{
		AnalysisID:      analysis.ID,
		Timestamp:       "00:02-00:09",
		MarkerType:      models.HighlightMarkerHighlight,
		Mode:            models.HighlightModeRange,
		StartTimeMs:     2000,
		EndTimeMs:       &endTime,
		TagType:         models.HighlightGoal,
		Description:     "缺失文件片段。",
		VideoClipURL:    "http://localhost:8080/uploads/video-clips/retry-ready.mp4",
		ClipStatus:      models.HighlightClipReady,
		IncludeInReport: true,
	}
	if err := db.Create(&highlight).Error; err != nil {
		t.Fatalf("create retry highlight: %v", err)
	}
	job, err := ctrl.clipExportJobs.start(&analysis, owner.ID, ExportHighlightClipsRequest{}, []highlightClipExportItem{{
		Highlight: highlight,
		LocalPath: filepath.Join(t.TempDir(), "missing.mp4"),
		FileName:  "missing.mp4",
	}})
	if err != nil {
		t.Fatalf("start failed export job: %v", err)
	}
	waitForClipExportJobStatus(t, ctrl, owner.ID, analysis.ID, job.ID, highlightClipExportFailed)
	if err := os.WriteFile(filepath.Join(clipDir, "retry-ready.mp4"), []byte("retry-ready-clip"), 0644); err != nil {
		t.Fatalf("write retry clip: %v", err)
	}

	params := gin.Params{
		{Key: "id", Value: strconv.Itoa(int(analysis.ID))},
		{Key: "job_id", Value: job.ID},
	}
	retried := performVideoAnalysisRequest(t, owner.ID, http.MethodPost, "/video-analysis/"+params[0].Value+"/clips/export/jobs/"+job.ID+"/retry", nil, params, ctrl.RetryHighlightClipsExportJob)
	if retried.Code != http.StatusOK {
		t.Fatalf("retry job status = %d, want %d, body=%s", retried.Code, http.StatusOK, retried.Body.String())
	}
	var retryBody struct {
		Data HighlightClipExportJobResponse `json:"data"`
	}
	if err := json.Unmarshal(retried.Body.Bytes(), &retryBody); err != nil {
		t.Fatalf("decode retried job: %v", err)
	}
	if retryBody.Data.Status != highlightClipExportQueued {
		t.Fatalf("retried job status = %s, want queued", retryBody.Data.Status)
	}
}

func TestHighlightClipsExportJobMarksInterruptedPersistedJobFailed(t *testing.T) {
	db, ctrl, owner, _ := setupVideoAnalysisControllerTest(t)
	analysis := models.VideoAnalysis{
		OrderID:        18,
		AnalystID:      owner.ID,
		UserID:         108,
		PlayerName:     "Interrupted Export Player",
		PlayerPosition: "midfielder",
		Status:         models.AnalysisStatusScoring,
	}
	if err := db.Create(&analysis).Error; err != nil {
		t.Fatalf("create analysis: %v", err)
	}
	record := models.VideoClipExportJob{
		JobID:       "interrupted-job",
		AnalysisID:  analysis.ID,
		AnalystID:   owner.ID,
		Status:      models.VideoClipExportProcessing,
		Progress:    42,
		Processed:   1,
		Total:       3,
		FileName:    "interrupted.zip",
		RequestJSON: "{}",
	}
	if err := db.Create(&record).Error; err != nil {
		t.Fatalf("create interrupted job: %v", err)
	}

	params := gin.Params{
		{Key: "id", Value: strconv.Itoa(int(analysis.ID))},
		{Key: "job_id", Value: record.JobID},
	}
	res := performVideoAnalysisRequest(t, owner.ID, http.MethodGet, "/video-analysis/"+params[0].Value+"/clips/export/jobs/"+record.JobID, nil, params, ctrl.GetHighlightClipsExportJob)
	if res.Code != http.StatusOK {
		t.Fatalf("get interrupted job status = %d, want %d, body=%s", res.Code, http.StatusOK, res.Body.String())
	}
	var body struct {
		Data HighlightClipExportJobResponse `json:"data"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode interrupted job: %v", err)
	}
	if body.Data.Status != highlightClipExportFailed || !strings.Contains(body.Data.Error, "中断") {
		t.Fatalf("interrupted job = %#v, want failed interrupted state", body.Data)
	}
}

func waitForClipExportJobStatus(t *testing.T, ctrl *VideoAnalysisController, analystID, analysisID uint, jobID string, want highlightClipExportJobStatus) HighlightClipExportJobResponse {
	t.Helper()
	params := gin.Params{
		{Key: "id", Value: strconv.Itoa(int(analysisID))},
		{Key: "job_id", Value: jobID},
	}
	var last HighlightClipExportJobResponse
	for i := 0; i < 50; i++ {
		res := performVideoAnalysisRequest(t, analystID, http.MethodGet, "/video-analysis/"+params[0].Value+"/clips/export/jobs/"+jobID, nil, params, ctrl.GetHighlightClipsExportJob)
		if res.Code != http.StatusOK {
			t.Fatalf("get job status = %d, want %d, body=%s", res.Code, http.StatusOK, res.Body.String())
		}
		var body struct {
			Data HighlightClipExportJobResponse `json:"data"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode job: %v", err)
		}
		last = body.Data
		if body.Data.Status == want {
			return body.Data
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("job %s did not reach %s, last=%#v", jobID, want, last)
	return last
}

func ptrTime(value time.Time) *time.Time {
	return &value
}

func hasZipEntryPrefix(entries map[string]string, prefix string) bool {
	for name := range entries {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}
