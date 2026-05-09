package services

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shaonianqiutan/backend/models"
)

func TestGenerateVideoAnalysisWordReportCreatesDocx(t *testing.T) {
	reportsDir := t.TempDir()
	generator := NewReportGenerator(reportsDir)

	analysis := &models.VideoAnalysis{
		OrderID:         42,
		PlayerName:      `测试/球员`,
		PlayerAge:       13,
		PlayerPosition:  "边锋",
		PlayerFoot:      "右脚",
		PlayerTeam:      "少年一队",
		MatchName:       "春季联赛",
		Opponent:        "城南队",
		OverallScore:    82.5,
		Summary:         "整体推进积极。",
		Strengths:       "边路推进积极\n连续动作稳定",
		Weaknesses:      "对抗后出球选择需要提升",
		Improvements:    "加强对抗后的传球选择。",
		AnalystNotes:    "正式交付模板验证。",
		AIReport:        "## 技术表现\n- 控球稳定\n- 传中质量较好",
		AIReportVersion: 2,
	}
	user := &models.User{
		Name:     "测试球员",
		Phone:    "13900000000",
		Age:      13,
		Position: "边锋",
		Club:     "少年俱乐部",
	}

	reportURL, err := generator.GenerateVideoAnalysisWordReport(analysis, "测试分析师", user)
	if err != nil {
		t.Fatalf("generate word report: %v", err)
	}
	if !strings.HasPrefix(reportURL, "/uploads/reports/少年球探_视频分析报告_测试_球员_订单42_v2") {
		t.Fatalf("report url = %q, want sanitized formal report path", reportURL)
	}
	if !strings.HasSuffix(reportURL, ".docx") {
		t.Fatalf("report url = %q, want docx suffix", reportURL)
	}

	docxPath := filepath.Join(reportsDir, filepath.Base(reportURL))
	documentXML := readDocxEntry(t, docxPath, "word/document.xml")
	for _, want := range []string{
		"少年球探视频分析报告",
		VideoAnalysisReportTemplateVersion,
		"测试/球员",
		"综合评分：82.5 / 100",
		"边路推进积极",
		"技术表现",
	} {
		if !strings.Contains(documentXML, want) {
			t.Fatalf("document.xml missing %q", want)
		}
	}
	if strings.Contains(documentXML, user.Phone) {
		t.Fatalf("document.xml should not contain player phone")
	}
}

func TestGenerateVideoAnalysisPDFReportCreatesPdf(t *testing.T) {
	reportsDir := t.TempDir()
	generator := NewReportGenerator(reportsDir)

	analysis := &models.VideoAnalysis{
		OrderID:         88,
		PlayerName:      "PDF球员",
		PlayerAge:       15,
		PlayerPosition:  "中场",
		MatchName:       "秋季联赛",
		Opponent:        "北区队",
		OverallScore:    76.0,
		Summary:         "组织调度稳定。",
		Strengths:       "控球稳定",
		Weaknesses:      "转身后推进速度还需加强",
		Improvements:    "加强中路接应。",
		AIReport:        "## 结构化结论\n- 传控稳定\n- 决策清晰",
		AIReportVersion: 1,
	}

	reportURL, err := generator.GenerateVideoAnalysisPDFReport(analysis, "测试分析师", nil)
	if err != nil {
		t.Fatalf("generate pdf report: %v", err)
	}
	if !strings.HasPrefix(reportURL, "/uploads/reports/少年球探_视频分析报告_PDF球员_订单88_v1") {
		t.Fatalf("pdf url = %q, want formal pdf path", reportURL)
	}
	if !strings.HasSuffix(reportURL, ".pdf") {
		t.Fatalf("pdf url = %q, want pdf suffix", reportURL)
	}

	pdfPath := filepath.Join(reportsDir, filepath.Base(reportURL))
	raw, err := os.ReadFile(pdfPath)
	if err != nil {
		t.Fatalf("read generated pdf: %v", err)
	}
	content := string(raw)
	for _, want := range []string{"%PDF-1.4", "/Type /Catalog", "/UniGB-UCS2-H", "STSong-Light", "%%EOF"} {
		if !strings.Contains(content, want) {
			t.Fatalf("pdf missing %q", want)
		}
	}
}

func readDocxEntry(t *testing.T, path string, entryName string) string {
	t.Helper()
	reader, err := zip.OpenReader(path)
	if err != nil {
		t.Fatalf("open docx zip %s: %v", path, err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		if file.Name != entryName {
			continue
		}
		entry, err := file.Open()
		if err != nil {
			t.Fatalf("open docx entry %s: %v", entryName, err)
		}
		defer entry.Close()
		data, err := io.ReadAll(entry)
		if err != nil {
			t.Fatalf("read docx entry %s: %v", entryName, err)
		}
		return string(data)
	}
	t.Fatalf("docx entry %s not found", entryName)
	return ""
}
