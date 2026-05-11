package services

import (
	"archive/zip"
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/shaonianqiutan/backend/models"
)

// ReportGenerator 报告文档生成器
type ReportGenerator struct {
	reportsDir string
}

const (
	defaultBrandWideLightLogo = "assets/report/brand-wide-light.png"
	defaultBrandIconLogo      = "assets/report/brand-icon.png"
	legacyBrandWideLightLogo  = "/Users/saili/Desktop/少年球探/少年球探logo 白色背景.png"
	legacyBrandIconLogo       = "/Users/saili/Desktop/少年球探/官网logo2.png"
)

// VideoAnalysisDocumentTemplateVersion marks the Word/PDF rendering template.
const VideoAnalysisDocumentTemplateVersion = "video-analysis-document-v1.5-2026-05-10"

type reportBrandAssets struct {
	WideLightLogo string
	IconLogo      string
}

type reportDocumentTemplateConfig struct {
	Version           string
	Title             string
	Subtitle          string
	FooterText        string
	Disclaimer        string
	DeliveryNote      string
	ShowHighlights    bool
	ShowScoreOverview bool
}

// NewReportGenerator 创建文档生成器
func NewReportGenerator(reportsDir string) *ReportGenerator {
	return &ReportGenerator{reportsDir: reportsDir}
}

// EnsureDir 确保目录存在
func (g *ReportGenerator) EnsureDir() error {
	if _, err := os.Stat(g.reportsDir); os.IsNotExist(err) {
		return os.MkdirAll(g.reportsDir, 0755)
	}
	return nil
}

func (g *ReportGenerator) brandAssets() reportBrandAssets {
	return reportBrandAssets{
		WideLightLogo: firstExistingReportAsset(os.Getenv("REPORT_BRAND_LOGO_LIGHT"), defaultBrandWideLightLogo, legacyBrandWideLightLogo),
		IconLogo:      firstExistingReportAsset(os.Getenv("REPORT_BRAND_LOGO_ICON"), defaultBrandIconLogo, legacyBrandIconLogo),
	}
}

func (g *ReportGenerator) documentTemplateConfig() reportDocumentTemplateConfig {
	return reportDocumentTemplateConfig{
		Version:           firstNonEmptyReportText(os.Getenv("REPORT_DOCUMENT_TEMPLATE_VERSION"), VideoAnalysisDocumentTemplateVersion),
		Title:             firstNonEmptyReportText(os.Getenv("REPORT_DOCUMENT_TITLE"), "青少年足球视频分析报告"),
		Subtitle:          firstNonEmptyReportText(os.Getenv("REPORT_DOCUMENT_SUBTITLE"), "Youth Football Video Analysis Report"),
		FooterText:        firstNonEmptyReportText(os.Getenv("REPORT_DOCUMENT_FOOTER"), "少年球探 · uscout.cn"),
		Disclaimer:        firstNonEmptyReportText(os.Getenv("REPORT_DOCUMENT_DISCLAIMER"), "声明：本报告基于分析师录入的比赛信息、评分、文字评价与关键片段形成，供青少年足球训练与成长规划参考，不作为医学、升学或职业签约承诺。"),
		DeliveryNote:      firstNonEmptyReportText(os.Getenv("REPORT_DOCUMENT_DELIVERY_NOTE"), "本报告由系统模板根据球员基础信息、分析师评分评语、比赛关键标记和 AI 辅助正文渲染生成。Word 版本用于分析师检查与补充，PDF 版本用于正式阅读和打印。"),
		ShowHighlights:    envBoolDefault("REPORT_DOCUMENT_SHOW_HIGHLIGHTS", true),
		ShowScoreOverview: envBoolDefault("REPORT_DOCUMENT_SHOW_SCORE_OVERVIEW", true),
	}
}

func envBoolDefault(name string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(name)))
	if value == "" {
		return fallback
	}
	return value == "1" || value == "true" || value == "yes" || value == "on"
}

func firstExistingReportAsset(paths ...string) string {
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		for _, candidate := range reportAssetCandidates(path) {
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
				return candidate
			}
		}
	}
	return ""
}

func reportAssetCandidates(path string) []string {
	if filepath.IsAbs(path) {
		return []string{path}
	}
	return []string{path, filepath.Join("..", path)}
}

// GenerateReportDocs 生成两份 MD 文档，返回 (评分报告路径, 球员基础信息路径, error)
func (g *ReportGenerator) GenerateReportDocs(order *models.Order, user *models.User, analyst *models.Analyst, ratings map[string]interface{}, summary, suggestions, potential string, strengths, weaknesses []string) (ratingMDPath string, playerInfoMDPath string, err error) {
	if err := g.EnsureDir(); err != nil {
		return "", "", fmt.Errorf("创建目录失败: %w", err)
	}

	timestamp := time.Now().Format("20060102150405")

	// 1. 生成球员评分报告
	ratingContent := g.buildRatingReport(order, analyst, ratings, summary, suggestions, potential, strengths, weaknesses)
	ratingFileName := fmt.Sprintf("评分报告_%s_%s.md", order.OrderNo, timestamp)
	ratingMDPath = filepath.Join(g.reportsDir, ratingFileName)
	if err := os.WriteFile(ratingMDPath, []byte(ratingContent), 0644); err != nil {
		return "", "", fmt.Errorf("写入评分报告失败: %w", err)
	}

	// 2. 生成球员基础信息文档
	playerContent := g.buildPlayerInfoDoc(order, user)
	playerFileName := fmt.Sprintf("球员基础信息_%d_%s_%s.md", user.ID, strings.TrimSpace(user.Name), timestamp)
	playerInfoMDPath = filepath.Join(g.reportsDir, playerFileName)
	if err := os.WriteFile(playerInfoMDPath, []byte(playerContent), 0644); err != nil {
		return "", "", fmt.Errorf("写入球员基础信息失败: %w", err)
	}

	return ratingMDPath, playerInfoMDPath, nil
}

// buildRatingReport 构建球员评分报告内容
func (g *ReportGenerator) buildRatingReport(order *models.Order, analyst *models.Analyst, ratings map[string]interface{}, summary, suggestions, potential string, strengths, weaknesses []string) string {
	var buf bytes.Buffer

	// 计算评分
	overallRating, offenseRating, defenseRating := calcRatings(ratings)
	level := getRatingLevel(overallRating)

	buf.WriteString("# 球员评分报告\n\n")
	buf.WriteString(fmt.Sprintf("**报告编号：** REP-%s<br>\n", order.OrderNo))
	buf.WriteString(fmt.Sprintf("**球员姓名：** %s<br>\n", order.PlayerName))
	buf.WriteString(fmt.Sprintf("**球员位置：** %s<br>\n", order.PlayerPosition))
	buf.WriteString(fmt.Sprintf("**分析日期：** %s<br>\n", time.Now().Format("2006年1月2日")))
	buf.WriteString(fmt.Sprintf("**分析师：** %s<br>\n", analyst.Name))
	buf.WriteString(fmt.Sprintf("**订单类型：** %s<br>\n", getOrderTypeName(order.OrderType)))
	buf.WriteString(fmt.Sprintf("**综合评分：** %.1f / 10.0（%s）<br>\n", overallRating, level))
	buf.WriteString(fmt.Sprintf("**潜力评估：** %s<br>\n", getPotentialLabel(potential)))
	buf.WriteString("\n---\n\n")

	// 整体维度
	buf.WriteString("## 一、整体技术能力评估\n\n")
	if overall, ok := ratings["overall"].(map[string]interface{}); ok {
		for _, item := range getOverallItems() {
			if detail, ok := overall[item.key].(map[string]interface{}); ok {
				score, _ := detail["score"].(float64)
				comment, _ := detail["comment"].(string)
				timestamps := extractTimestamps(detail)
				buf.WriteString(fmt.Sprintf("### %s（评分：%.1f）\n\n", item.label, score))
				buf.WriteString(fmt.Sprintf("%s\n\n", comment))
				if len(timestamps) > 0 {
					buf.WriteString(fmt.Sprintf("**高光时刻：** %s\n\n", formatTimestamps(timestamps)))
				}
			}
		}
	}

	// 进攻维度
	buf.WriteString("## 二、进攻能力分析\n\n")
	if offense, ok := ratings["offense"].(map[string]interface{}); ok {
		for _, item := range getOffenseItems() {
			if detail, ok := offense[item.key].(map[string]interface{}); ok {
				score, _ := detail["score"].(float64)
				comment, _ := detail["comment"].(string)
				timestamps := extractTimestamps(detail)
				buf.WriteString(fmt.Sprintf("### %s（评分：%.1f）\n\n", item.label, score))
				buf.WriteString(fmt.Sprintf("%s\n\n", comment))
				if len(timestamps) > 0 {
					buf.WriteString(fmt.Sprintf("**高光时刻：** %s\n\n", formatTimestamps(timestamps)))
				}
			}
		}
	}

	// 防守维度
	buf.WriteString("## 三、防守能力分析\n\n")
	if defense, ok := ratings["defense"].(map[string]interface{}); ok {
		for _, item := range getDefenseItems() {
			if detail, ok := defense[item.key].(map[string]interface{}); ok {
				score, _ := detail["score"].(float64)
				comment, _ := detail["comment"].(string)
				timestamps := extractTimestamps(detail)
				buf.WriteString(fmt.Sprintf("### %s（评分：%.1f）\n\n", item.label, score))
				buf.WriteString(fmt.Sprintf("%s\n\n", comment))
				if len(timestamps) > 0 {
					buf.WriteString(fmt.Sprintf("**高光时刻：** %s\n\n", formatTimestamps(timestamps)))
				}
			}
		}
	}

	// 核心优势
	if len(strengths) > 0 {
		buf.WriteString("## 四、核心优势\n\n")
		for _, s := range strengths {
			buf.WriteString(fmt.Sprintf("- %s\n", s))
		}
		buf.WriteString("\n")
	}

	// 待提升领域
	if len(weaknesses) > 0 {
		buf.WriteString("## 五、待提升领域\n\n")
		for _, w := range weaknesses {
			buf.WriteString(fmt.Sprintf("- %s\n", w))
		}
		buf.WriteString("\n")
	}

	// 综合评价
	if summary != "" {
		buf.WriteString("## 六、综合评价\n\n")
		buf.WriteString(fmt.Sprintf("%s\n\n", summary))
	}

	// 发展建议
	if suggestions != "" {
		buf.WriteString("## 七、发展建议\n\n")
		buf.WriteString(fmt.Sprintf("%s\n\n", suggestions))
	}

	// 评分总览表
	buf.WriteString("## 八、评分总览\n\n")
	buf.WriteString("| 维度 | 评分 |\n")
	buf.WriteString("|------|------|\n")
	buf.WriteString(fmt.Sprintf("| 整体技术 | %.1f |\n", overallRating))
	buf.WriteString(fmt.Sprintf("| 进攻能力 | %.1f |\n", offenseRating))
	buf.WriteString(fmt.Sprintf("| 防守能力 | %.1f |\n", defenseRating))
	buf.WriteString(fmt.Sprintf("| **综合评分** | **%.1f** |\n", overallRating))
	buf.WriteString("\n---\n\n")
	buf.WriteString(fmt.Sprintf("*报告生成时间：%s*\n", time.Now().Format("2006-01-02 15:04:05")))

	return buf.String()
}

// buildPlayerInfoDoc 构建球员基础信息文档
func (g *ReportGenerator) buildPlayerInfoDoc(order *models.Order, user *models.User) string {
	var buf bytes.Buffer

	buf.WriteString("# 球员基础信息\n\n")
	buf.WriteString(fmt.Sprintf("**球员 ID：** %d<br>\n", user.ID))
	buf.WriteString(fmt.Sprintf("**姓名：** %s<br>\n", user.Name))
	buf.WriteString(fmt.Sprintf("**昵称：** %s<br>\n", user.Nickname))
	buf.WriteString(fmt.Sprintf("**位置：** %s<br>\n", user.Position))
	buf.WriteString(fmt.Sprintf("**年龄：** %d 岁<br>\n", user.Age))
	buf.WriteString(fmt.Sprintf("**生日：** %s<br>\n", user.BirthDate))
	buf.WriteString(fmt.Sprintf("**性别：** %s<br>\n", user.Gender))
	buf.WriteString(fmt.Sprintf("**身高：** %.1f cm<br>\n", user.Height))
	buf.WriteString(fmt.Sprintf("**体重：** %.1f kg<br>\n", user.Weight))
	buf.WriteString(fmt.Sprintf("**惯用脚：** %s<br>\n", user.Foot))
	buf.WriteString(fmt.Sprintf("**地区：** %s %s<br>\n", user.Province, user.City))
	buf.WriteString(fmt.Sprintf("**俱乐部：** %s<br>\n", user.Club))
	buf.WriteString(fmt.Sprintf("**学校：** %s<br>\n", user.School))

	// 足球经历
	if user.Experiences != "" {
		var experiences []map[string]interface{}
		if err := json.Unmarshal([]byte(user.Experiences), &experiences); err == nil && len(experiences) > 0 {
			buf.WriteString("\n## 足球经历\n\n")
			for _, exp := range experiences {
				period, _ := exp["period"].(string)
				team, _ := exp["team"].(string)
				position, _ := exp["position"].(string)
				achievement, _ := exp["achievement"].(string)
				buf.WriteString(fmt.Sprintf("### %s | %s\n\n", period, team))
				buf.WriteString(fmt.Sprintf("- **位置：** %s\n", position))
				if achievement != "" {
					buf.WriteString(fmt.Sprintf("- **成就：** %s\n", achievement))
				}
				buf.WriteString("\n")
			}
		}
	}

	// 技术标签
	if user.TechnicalTags != "" {
		var tags []string
		if err := json.Unmarshal([]byte(user.TechnicalTags), &tags); err == nil && len(tags) > 0 {
			buf.WriteString("## 技术标签\n\n")
			buf.WriteString("| ")
			for _, tag := range tags {
				buf.WriteString(fmt.Sprintf("%s | ", tag))
			}
			buf.WriteString("\n\n")
		}
	}

	// mental_tags 字段
	userMentalTags := g.getFieldFromJSON(user.MentalTags)
	if userMentalTags != "" {
		buf.WriteString("## 心理素质标签\n\n")
		buf.WriteString(fmt.Sprintf("%s\n\n", userMentalTags))
	}

	// 比赛风格
	if user.PlayingStyle != "" {
		buf.WriteString("## 比赛风格\n\n")
		buf.WriteString(fmt.Sprintf("%s\n\n", user.PlayingStyle))
	}

	// 体测数据（从 physical_test_records 表查询，这里只记录字段占位）
	buf.WriteString("## 体测数据\n\n")
	buf.WriteString("*（请在体测模块补充最新体测数据）*\n\n")

	// 订单关联信息
	buf.WriteString("---\n\n")
	buf.WriteString("## 订单信息\n\n")
	buf.WriteString(fmt.Sprintf("- **订单号：** %s\n", order.OrderNo))
	buf.WriteString(fmt.Sprintf("- **订单类型：** %s\n", getOrderTypeName(order.OrderType)))
	buf.WriteString(fmt.Sprintf("- **比赛名称：** %s\n", order.MatchName))
	buf.WriteString(fmt.Sprintf("- **对手：** %s\n", order.Opponent))
	buf.WriteString(fmt.Sprintf("- **球衣颜色：** %s %s\n", order.JerseyColor, order.JerseyNumber))
	if order.VideoDuration > 0 {
		buf.WriteString(fmt.Sprintf("- **视频时长：** %d 秒（约%.1f分钟）\n", order.VideoDuration, float64(order.VideoDuration)/60))
	}

	buf.WriteString("\n---\n\n")
	buf.WriteString(fmt.Sprintf("*文档生成时间：%s*\n", time.Now().Format("2006-01-02 15:04:05")))

	return buf.String()
}

func (g *ReportGenerator) getFieldFromJSON(jsonStr string) string {
	if jsonStr == "" || jsonStr == "[]" {
		return ""
	}
	var items []string
	if err := json.Unmarshal([]byte(jsonStr), &items); err == nil && len(items) > 0 {
		return strings.Join(items, " / ")
	}
	return jsonStr
}

// ===== 辅助函数 =====

type ratingItem struct {
	key   string
	label string
}

func getOverallItems() []ratingItem {
	return []ratingItem{
		{key: "ballControl", label: "控球能力"},
		{key: "pressing", label: "逼抢能力"},
		{key: "positioning", label: "站位意识"},
	}
}

func getOffenseItems() []ratingItem {
	return []ratingItem{
		{key: "widthAndAttack", label: "拉开宽度并参与进攻组织"},
		{key: "offTheBallMovement", label: "跑位支援灵活"},
		{key: "duelVariety", label: "对抗中表现多变"},
		{key: "oneOnOne", label: "擅长一对一突破"},
		{key: "crossing", label: "传中与助攻能力"},
		{key: "speed", label: "速度与节奏变化"},
		{key: "passingRisk", label: "传球风险判断"},
		{key: "firstTouch", label: "身体姿态与一脚传球"},
	}
}

func getDefenseItems() []ratingItem {
	return []ratingItem{
		{key: "defensiveEffort", label: "防守阶段投入"},
		{key: "reactionSpeed", label: "失球后反应迅速"},
		{key: "teamCoordination", label: "与队友配合默契"},
		{key: "secondBall", label: "注重第二落点争夺"},
		{key: "aerialDuel", label: "空中球争夺"},
		{key: "positioning", label: "向中路收缩"},
		{key: "roleAdaptation", label: "快速调整防守角色"},
		{key: "tackling", label: "防守节奏把控"},
	}
}

func calcRatings(ratings map[string]interface{}) (overall, offense, defense float64) {
	calcAvg := func(category string) float64 {
		if cat, ok := ratings[category].(map[string]interface{}); ok {
			count := 0
			sum := 0.0
			for _, v := range cat {
				if detail, ok := v.(map[string]interface{}); ok {
					if score, ok := detail["score"].(float64); ok {
						sum += score
						count++
					}
				}
			}
			if count > 0 {
				return sum / float64(count)
			}
		}
		return 0
	}
	return calcAvg("overall"), calcAvg("offense"), calcAvg("defense")
}

func getRatingLevel(score float64) string {
	switch {
	case score >= 9.0:
		return "世界级"
	case score >= 8.0:
		return "优秀"
	case score >= 7.0:
		return "良好"
	case score >= 6.0:
		return "合格"
	case score >= 5.0:
		return "待提高"
	default:
		return "薄弱"
	}
}

func getPotentialLabel(p string) string {
	switch p {
	case "top":
		return "顶级"
	case "high":
		return "优秀"
	case "medium":
		return "良好"
	case "low":
		return "一般"
	default:
		return p
	}
}

func getOrderTypeName(t string) string {
	switch t {
	case "video":
		return "视频版"
	case "pro":
		return "文字+视频版"
	default:
		return "文字版"
	}
}

func extractTimestamps(detail map[string]interface{}) []float64 {
	if ts, ok := detail["timestamps"].([]interface{}); ok {
		result := make([]float64, 0, len(ts))
		for _, t := range ts {
			if f, ok := t.(float64); ok {
				result = append(result, f)
			} else if i, ok := t.(int); ok {
				result = append(result, float64(i))
			}
		}
		return result
	}
	return nil
}

func formatTimestamps(ts []float64) string {
	var parts []string
	for _, t := range ts {
		parts = append(parts, formatSecond(t))
	}
	return strings.Join(parts, " / ")
}

func formatSecond(s float64) string {
	m := int(s) / 60
	sec := int(s) % 60
	return fmt.Sprintf("%d'%02d\"", m, sec)
}

// ParseFloat64 安全解析 float64
func ParseFloat64(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return 0
}

// GenerateFromVideoAnalysis 从视频分析记录生成两份 MD 文档
// 返回 (评分报告路径, 球员基础信息路径, error)
func (g *ReportGenerator) GenerateFromVideoAnalysis(analysis *models.VideoAnalysis, analystName string, user *models.User) (ratingMDPath string, playerInfoMDPath string, err error) {
	if err := g.EnsureDir(); err != nil {
		return "", "", fmt.Errorf("创建目录失败: %w", err)
	}

	timestamp := time.Now().Format("20060102150405")

	// 1. 生成球员评分报告
	ratingContent := g.buildVideoAnalysisReport(analysis, analystName)
	ratingFileName := fmt.Sprintf("评分报告_VA%d_%s.md", analysis.ID, timestamp)
	ratingMDPath = filepath.Join(g.reportsDir, ratingFileName)
	if err := os.WriteFile(ratingMDPath, []byte(ratingContent), 0644); err != nil {
		return "", "", fmt.Errorf("写入评分报告失败: %w", err)
	}

	// 2. 生成球员基础信息文档
	playerContent := g.buildPlayerInfoFromVideoAnalysis(analysis, user)
	playerFileName := fmt.Sprintf("球员基础信息_%d_%s_%s.md", analysis.UserID, strings.TrimSpace(analysis.PlayerName), timestamp)
	playerInfoMDPath = filepath.Join(g.reportsDir, playerFileName)
	if err := os.WriteFile(playerInfoMDPath, []byte(playerContent), 0644); err != nil {
		return "", "", fmt.Errorf("写入球员基础信息失败: %w", err)
	}

	return ratingMDPath, playerInfoMDPath, nil
}

// GenerateVideoAnalysisWordReport 生成视频分析正式 Word 报告。
// 返回值为可写入 reports.ai_report_url 的 Web 路径，文件实体写入 reportsDir。
func (g *ReportGenerator) GenerateVideoAnalysisWordReport(analysis *models.VideoAnalysis, analystName string, user *models.User, highlights ...models.AnalysisHighlight) (string, error) {
	if analysis == nil {
		return "", fmt.Errorf("分析记录不能为空")
	}
	if err := g.EnsureDir(); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	playerName := firstNonEmptyReportText(analysis.PlayerName, userName(user), "未知球员")
	version := analysis.AIReportVersion
	if version <= 0 {
		version = 1
	}
	fileName := fmt.Sprintf(
		"少年球探_视频分析报告_%s_订单%d_v%d.docx",
		sanitizeReportFileNamePart(playerName),
		analysis.OrderID,
		version,
	)
	fullPath := filepath.Join(g.reportsDir, fileName)

	paragraphs := g.buildVideoAnalysisDocumentParagraphs(analysis, analystName, user, playerName, version, highlights)
	if err := writeSimpleDocx(fullPath, paragraphs); err != nil {
		return "", fmt.Errorf("写入视频分析 Word 报告失败: %w", err)
	}
	return "/uploads/reports/" + fileName, nil
}

// GenerateVideoAnalysisPDFReport 生成视频分析正式 PDF 报告。
// 返回值为可写入 reports.pdf_url 的 Web 路径，文件实体写入 reportsDir。
func (g *ReportGenerator) GenerateVideoAnalysisPDFReport(analysis *models.VideoAnalysis, analystName string, user *models.User, highlights ...models.AnalysisHighlight) (string, error) {
	if analysis == nil {
		return "", fmt.Errorf("分析记录不能为空")
	}
	if err := g.EnsureDir(); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	playerName := firstNonEmptyReportText(analysis.PlayerName, userName(user), "未知球员")
	version := analysis.AIReportVersion
	if version <= 0 {
		version = 1
	}
	fileName := fmt.Sprintf(
		"少年球探_视频分析报告_%s_订单%d_v%d.pdf",
		sanitizeReportFileNamePart(playerName),
		analysis.OrderID,
		version,
	)
	fullPath := filepath.Join(g.reportsDir, fileName)

	paragraphs := g.buildVideoAnalysisDocumentParagraphs(analysis, analystName, user, playerName, version, highlights)
	if err := writeSimplePDF(fullPath, paragraphs); err != nil {
		return "", fmt.Errorf("写入视频分析 PDF 报告失败: %w", err)
	}
	return "/uploads/reports/" + fileName, nil
}

type docxParagraph struct {
	Text            string
	Bold            bool
	Center          bool
	Size            int
	Spacing         int
	Before          int
	Color           string
	Shading         string
	BorderBottom    string
	IndentLeft      int
	Bullet          bool
	ImagePath       string
	ImageWidthPt    float64
	PageBreak       bool
	PageBreakBefore bool
	Table           *docxTable
}

type docxTable struct {
	ColWidths []int
	Rows      []docxTableRow
}

type docxTableRow struct {
	Header bool
	Cells  []docxTableCell
}

type docxTableCell struct {
	Text    string
	Bold    bool
	Center  bool
	Size    int
	Color   string
	Shading string
}

func (g *ReportGenerator) buildVideoAnalysisDocumentParagraphs(analysis *models.VideoAnalysis, analystName string, user *models.User, playerName string, version int, highlights []models.AnalysisHighlight) []docxParagraph {
	assets := g.brandAssets()
	config := g.documentTemplateConfig()
	reportNo := videoAnalysisReportNo(analysis, version)
	paragraphs := []docxParagraph{
		{ImagePath: assets.WideLightLogo, ImageWidthPt: 330, Center: true, Spacing: 260},
		{Text: config.Title, Bold: true, Center: true, Size: 42, Color: "0B1726", Spacing: 90},
		{Text: config.Subtitle, Center: true, Size: 18, Color: "64748B", Spacing: 220},
		{Table: buildDocxCoverMetaTable(playerName, analysis, user, analystName, reportNo, version)},
		{Text: "本报告基于球员基础资料、分析师评分评语、关键片段标记与 AI 生成正文排版生成，供训练复盘与成长规划使用。", Center: true, Size: 18, Color: "475569", Shading: "F8FAFC", Spacing: 220},
		{Text: "一、球员与比赛信息", Bold: true, Size: 28, Color: "0B1726", BorderBottom: "00A6D6", PageBreakBefore: true, Spacing: 160},
	}

	infoPairs := []docxInfoPair{
		{Label: "姓名", Value: playerName},
		{Label: "报告编号", Value: reportNo},
	}
	if analysis.PlayerAge > 0 {
		infoPairs = append(infoPairs, docxInfoPair{Label: "年龄", Value: fmt.Sprintf("%d岁", analysis.PlayerAge)})
	} else if user != nil && user.Age > 0 {
		infoPairs = append(infoPairs, docxInfoPair{Label: "年龄", Value: fmt.Sprintf("%d岁", user.Age)})
	}
	infoPairs = append(infoPairs,
		docxInfoPair{Label: "位置", Value: firstNonEmptyReportText(analysis.PlayerPosition, userPosition(user))},
		docxInfoPair{Label: "惯用脚", Value: firstNonEmptyReportText(analysis.PlayerFoot, userFoot(user))},
	)
	if height := firstPositiveFloat(analysis.PlayerHeight, userHeight(user)); height > 0 {
		infoPairs = append(infoPairs, docxInfoPair{Label: "身高", Value: fmt.Sprintf("%.0fcm", height)})
	}
	if weight := firstPositiveFloat(analysis.PlayerWeight, userWeight(user)); weight > 0 {
		infoPairs = append(infoPairs, docxInfoPair{Label: "体重", Value: fmt.Sprintf("%.0fkg", weight)})
	}
	infoPairs = append(infoPairs,
		docxInfoPair{Label: "当前球队", Value: firstNonEmptyReportText(analysis.PlayerTeam, userClub(user))},
		docxInfoPair{Label: "比赛名称", Value: analysis.MatchName},
		docxInfoPair{Label: "比赛日期", Value: analysis.MatchDate},
		docxInfoPair{Label: "对手", Value: analysis.Opponent},
	)
	if analysis.PlayTime > 0 {
		infoPairs = append(infoPairs, docxInfoPair{Label: "出场时间", Value: fmt.Sprintf("%d分钟", analysis.PlayTime)})
	}
	infoPairs = append(infoPairs, docxInfoPair{Label: "分析师", Value: firstNonEmptyReportText(analystName, "未知分析师")})
	paragraphs = append(paragraphs, docxParagraph{Table: buildDocxInfoTable(infoPairs)})

	paragraphs = append(paragraphs,
		docxParagraph{},
		docxParagraph{Text: "二、评分概览", Bold: true, Size: 28, Color: "0B1726", BorderBottom: "00A6D6", Spacing: 160},
		docxParagraph{Table: buildDocxScoreSummaryTable(analysis)},
	)
	if config.ShowScoreOverview {
		if scoreTable := buildDocxScoreTable(analysis); scoreTable != nil {
			paragraphs = append(paragraphs, docxParagraph{Table: scoreTable})
		}
	}
	if strings.TrimSpace(analysis.Summary) != "" {
		paragraphs = append(paragraphs, docxParagraph{Text: "综合评价", Bold: true, Size: 24, Color: "0B1726", Spacing: 120})
		paragraphs = append(paragraphs, markdownTextToDocxParagraphs(analysis.Summary)...)
	}

	appendDocxListSection := func(title, text string) {
		items := splitReportTextItems(text)
		if len(items) == 0 {
			return
		}
		paragraphs = append(paragraphs, docxParagraph{Text: title, Bold: true, Size: 24, Color: "0B1726", Spacing: 120})
		for _, item := range items {
			paragraphs = append(paragraphs, docxParagraph{Text: item, Size: 21, Spacing: 60, Bullet: true})
		}
	}
	appendDocxListSection("核心优势", analysis.Strengths)
	appendDocxListSection("待提升领域", analysis.Weaknesses)
	appendDocxListSection("重点改进建议", analysis.Improvements)
	appendDocxListSection("分析师补充说明", analysis.AnalystNotes)

	hasHighlights := config.ShowHighlights && len(highlights) > 0
	if hasHighlights {
		paragraphs = append(paragraphs, buildVideoAnalysisHighlightTimelineParagraphs(highlights)...)
	}

	if strings.TrimSpace(analysis.AIReport) != "" {
		reportBodyTitle := "三、报告正文"
		if hasHighlights {
			reportBodyTitle = "四、报告正文"
		}
		paragraphs = append(paragraphs,
			docxParagraph{},
			docxParagraph{Text: reportBodyTitle, Bold: true, Size: 28, Color: "0B1726", BorderBottom: "00A6D6", Spacing: 160},
		)
		paragraphs = append(paragraphs, markdownTextToDocxParagraphs(analysis.AIReport)...)
	}

	paragraphs = append(paragraphs,
		docxParagraph{},
		docxParagraph{ImagePath: assets.IconLogo, ImageWidthPt: 70, Center: true, Spacing: 90},
		docxParagraph{Text: "交付说明与免责声明", Bold: true, Size: 22, Color: "0B1726", Spacing: 90},
		docxParagraph{Text: config.DeliveryNote, Size: 18, Color: "475569", Shading: "F8FAFC", Spacing: 90},
		docxParagraph{Text: config.Disclaimer, Size: 18, Color: "475569", Shading: "F8FAFC", Spacing: 120},
		docxParagraph{Text: fmt.Sprintf("模板追踪：内容模板 %s；文档模板 %s", VideoAnalysisReportTemplateVersion, VideoAnalysisDocumentTemplateVersion), Center: true, Size: 16, Color: "64748B", Spacing: 80},
		docxParagraph{Text: fmt.Sprintf("生成时间：%s", time.Now().Format("2006-01-02 15:04:05")), Center: true, Size: 18, Spacing: 80},
		docxParagraph{Text: config.FooterText, Center: true, Size: 18, Spacing: 80},
	)

	return paragraphs
}

func videoAnalysisReportNo(analysis *models.VideoAnalysis, version int) string {
	if analysis == nil {
		return fmt.Sprintf("VA-0000-v%d", version)
	}
	return fmt.Sprintf("VA-%06d-v%d", analysis.OrderID, version)
}

func compactReportTemplateVersion(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "video-analysis-report-")
	value = strings.TrimPrefix(value, "video-analysis-document-")
	return value
}

type docxInfoPair struct {
	Label string
	Value string
}

type scoreLineInput struct {
	Group   string
	Label   string
	Score   float64
	Comment string
}

func buildDocxCoverMetaTable(playerName string, analysis *models.VideoAnalysis, user *models.User, analystName string, reportNo string, version int) *docxTable {
	position := ""
	if analysis != nil {
		position = analysis.PlayerPosition
	}
	position = firstNonEmptyReportText(position, userPosition(user), "未记录")
	score := 0.0
	orderID := uint(0)
	if analysis != nil {
		score = analysis.OverallScore
		orderID = analysis.OrderID
	}
	return &docxTable{
		ColWidths: []int{1500, 3013, 1500, 3013},
		Rows: []docxTableRow{
			{Cells: []docxTableCell{
				docxLabelCell("球员"),
				docxValueCell(playerName),
				docxLabelCell("位置"),
				docxValueCell(position),
			}},
			{Cells: []docxTableCell{
				docxLabelCell("综合评分"),
				docxValueCell(fmt.Sprintf("%.1f / 100", score)),
				docxLabelCell("潜力等级"),
				docxValueCell(string(models.GetPotentialLevel(score))),
			}},
			{Cells: []docxTableCell{
				docxLabelCell("报告编号"),
				docxValueCell(reportNo),
				docxLabelCell("订单编号"),
				docxValueCell(fmt.Sprintf("#%d", orderID)),
			}},
			{Cells: []docxTableCell{
				docxLabelCell("分析师"),
				docxValueCell(firstNonEmptyReportText(analystName, "未知分析师")),
				docxLabelCell("报告版本"),
				docxValueCell(fmt.Sprintf("v%d", version)),
			}},
		},
	}
}

func buildDocxInfoTable(pairs []docxInfoPair) *docxTable {
	rows := make([]docxTableRow, 0, (len(pairs)+1)/2)
	cleanPairs := make([]docxInfoPair, 0, len(pairs))
	for _, pair := range pairs {
		value := strings.TrimSpace(pair.Value)
		if value == "" || value == "0" || value == "0.0" {
			continue
		}
		cleanPairs = append(cleanPairs, docxInfoPair{Label: strings.TrimSpace(pair.Label), Value: value})
	}
	for i := 0; i < len(cleanPairs); i += 2 {
		cells := []docxTableCell{
			docxLabelCell(cleanPairs[i].Label),
			docxValueCell(cleanPairs[i].Value),
		}
		if i+1 < len(cleanPairs) {
			cells = append(cells, docxLabelCell(cleanPairs[i+1].Label), docxValueCell(cleanPairs[i+1].Value))
		} else {
			cells = append(cells, docxLabelCell(""), docxValueCell(""))
		}
		rows = append(rows, docxTableRow{Cells: cells})
	}
	return &docxTable{ColWidths: []int{1450, 3063, 1450, 3063}, Rows: rows}
}

func buildDocxScoreSummaryTable(analysis *models.VideoAnalysis) *docxTable {
	if analysis == nil {
		return nil
	}
	scores, err := models.ParseScoresFromJSON(analysis.Scores)
	overallAvg, offenseAvg, defenseAvg := 0.0, 0.0, 0.0
	if err == nil && scores != nil {
		overallAvg = averageScoreValues(scores.BallControl.Score, scores.OffBallMovement.Score, scores.PressingAwareness.Score, scores.Positioning.Score)
		offenseAvg = averageScoreValues(scores.WidthParticipation.Score, scores.OffBallSupport.Score, scores.OneVOne.Score, scores.CrossingAssist.Score, scores.CombatAbility.Score, scores.PaceRhythm.Score, scores.PassVision.Score, scores.BodyPosture.Score)
		defenseAvg = averageScoreValues(scores.DefensiveCommitment.Score, scores.LossRecovery.Score, scores.TeammateCoordination.Score, scores.SecondBall.Score, scores.AerialDuel.Score, scores.DefensiveShape.Score, scores.RoleAdjustment.Score, scores.DefensiveRhythm.Score)
	}
	return &docxTable{
		ColWidths: []int{2256, 2256, 2257, 2257},
		Rows: []docxTableRow{
			{Cells: []docxTableCell{
				docxMetricCell("综合评分", fmt.Sprintf("%.1f", analysis.OverallScore), "00A6D6"),
				docxMetricCell("潜力等级", string(models.GetPotentialLevel(analysis.OverallScore)), "0B1726"),
				docxMetricCell("整体维度", fmt.Sprintf("%.1f / 10", overallAvg), "1D4ED8"),
				docxMetricCell("进攻/防守", fmt.Sprintf("%.1f / %.1f", offenseAvg, defenseAvg), "0F766E"),
			}},
		},
	}
}

func buildDocxScoreTable(analysis *models.VideoAnalysis) *docxTable {
	if analysis == nil {
		return nil
	}
	scores, err := models.ParseScoresFromJSON(analysis.Scores)
	if err != nil || scores == nil {
		return nil
	}
	lines := []scoreLineInput{
		{"整体", "控球能力", scores.BallControl.Score, scores.BallControl.Comment},
		{"整体", "无球跑动", scores.OffBallMovement.Score, scores.OffBallMovement.Comment},
		{"整体", "逼抢意识", scores.PressingAwareness.Score, scores.PressingAwareness.Comment},
		{"整体", "站位/选位", scores.Positioning.Score, scores.Positioning.Comment},
		{"进攻", "拉开宽度参与", scores.WidthParticipation.Score, scores.WidthParticipation.Comment},
		{"进攻", "无球支援", scores.OffBallSupport.Score, scores.OffBallSupport.Comment},
		{"进攻", "1v1过人能力", scores.OneVOne.Score, scores.OneVOne.Comment},
		{"进攻", "传中/助攻", scores.CrossingAssist.Score, scores.CrossingAssist.Comment},
		{"进攻", "对抗能力", scores.CombatAbility.Score, scores.CombatAbility.Comment},
		{"进攻", "节奏把控", scores.PaceRhythm.Score, scores.PaceRhythm.Comment},
		{"进攻", "传球视野", scores.PassVision.Score, scores.PassVision.Comment},
		{"进攻", "身体姿态", scores.BodyPosture.Score, scores.BodyPosture.Comment},
		{"防守", "防守投入度", scores.DefensiveCommitment.Score, scores.DefensiveCommitment.Comment},
		{"防守", "丢球回追", scores.LossRecovery.Score, scores.LossRecovery.Comment},
		{"防守", "队友协防配合", scores.TeammateCoordination.Score, scores.TeammateCoordination.Comment},
		{"防守", "二点球争抢", scores.SecondBall.Score, scores.SecondBall.Comment},
		{"防守", "空中争顶", scores.AerialDuel.Score, scores.AerialDuel.Comment},
		{"防守", "防守阵型保持", scores.DefensiveShape.Score, scores.DefensiveShape.Comment},
		{"防守", "角色调整能力", scores.RoleAdjustment.Score, scores.RoleAdjustment.Comment},
		{"防守", "防守节奏", scores.DefensiveRhythm.Score, scores.DefensiveRhythm.Comment},
	}
	rows := []docxTableRow{{
		Header: true,
		Cells: []docxTableCell{
			docxHeaderCell("分组"),
			docxHeaderCell("评分项"),
			docxHeaderCell("得分"),
			docxHeaderCell("分析师评语"),
		},
	}}
	for _, line := range lines {
		rows = append(rows, docxTableRow{Cells: []docxTableCell{
			docxValueCell(line.Group),
			docxValueCell(line.Label),
			{Text: fmt.Sprintf("%.1f", line.Score), Bold: true, Center: true, Size: 20, Color: scoreColor(line.Score), Shading: "FFFFFF"},
			docxValueCell(firstNonEmptyReportText(line.Comment, "暂无评语")),
		}})
	}
	return &docxTable{ColWidths: []int{1100, 2100, 1100, 4726}, Rows: rows}
}

func docxLabelCell(text string) docxTableCell {
	return docxTableCell{Text: text, Bold: true, Size: 18, Color: "475569", Shading: "F1F5F9"}
}

func docxValueCell(text string) docxTableCell {
	return docxTableCell{Text: text, Size: 19, Color: "0F172A", Shading: "FFFFFF"}
}

func docxHeaderCell(text string) docxTableCell {
	return docxTableCell{Text: text, Bold: true, Center: true, Size: 18, Color: "0B1726", Shading: "E0F2FE"}
}

func docxMetricCell(label string, value string, color string) docxTableCell {
	return docxTableCell{Text: label + "\n" + value, Bold: true, Center: true, Size: 20, Color: color, Shading: "F8FAFC"}
}

func scoreColor(score float64) string {
	switch {
	case score >= 8.5:
		return "047857"
	case score >= 7:
		return "0369A1"
	default:
		return "B45309"
	}
}

func buildVideoAnalysisScoreBarParagraphs(analysis *models.VideoAnalysis) []docxParagraph {
	if analysis == nil {
		return nil
	}
	lines := []docxParagraph{
		{Text: "评分可视化", Bold: true, Size: 23, Spacing: 100},
		{Text: scoreBarLine("综合评分", analysis.OverallScore/10, 10), Size: 20, Spacing: 60},
	}
	scores, err := models.ParseScoresFromJSON(analysis.Scores)
	if err != nil || scores == nil {
		return lines
	}
	overallAvg := averageScoreValues(
		scores.BallControl.Score,
		scores.OffBallMovement.Score,
		scores.PressingAwareness.Score,
		scores.Positioning.Score,
	)
	offenseAvg := averageScoreValues(
		scores.WidthParticipation.Score,
		scores.OffBallSupport.Score,
		scores.OneVOne.Score,
		scores.CrossingAssist.Score,
		scores.CombatAbility.Score,
		scores.PaceRhythm.Score,
		scores.PassVision.Score,
		scores.BodyPosture.Score,
	)
	defenseAvg := averageScoreValues(
		scores.DefensiveCommitment.Score,
		scores.LossRecovery.Score,
		scores.TeammateCoordination.Score,
		scores.SecondBall.Score,
		scores.AerialDuel.Score,
		scores.DefensiveShape.Score,
		scores.RoleAdjustment.Score,
		scores.DefensiveRhythm.Score,
	)
	lines = append(lines,
		docxParagraph{Text: scoreBarLine("整体维度", overallAvg, 10), Size: 20, Spacing: 60},
		docxParagraph{Text: scoreBarLine("进攻维度", offenseAvg, 10), Size: 20, Spacing: 60},
		docxParagraph{Text: scoreBarLine("防守维度", defenseAvg, 10), Size: 20, Spacing: 90},
	)
	return lines
}

func buildVideoAnalysisHighlightTimelineParagraphs(highlights []models.AnalysisHighlight) []docxParagraph {
	if len(highlights) == 0 {
		return nil
	}
	lines := []docxParagraph{
		{},
		{Text: "三、关键片段时间轴", Bold: true, Size: 28, Color: "0B1726", BorderBottom: "00A6D6", Spacing: 160},
	}
	limit := len(highlights)
	if limit > 15 {
		limit = 15
	}
	rows := []docxTableRow{{
		Header: true,
		Cells: []docxTableCell{
			docxHeaderCell("序号"),
			docxHeaderCell("时间"),
			docxHeaderCell("类型"),
			docxHeaderCell("片段说明"),
		},
	}}
	for i := 0; i < limit; i++ {
		highlight := highlights[i]
		timeText := firstNonEmptyReportText(highlight.Timestamp, formatReportHighlightMs(highlight.StartTimeMs))
		if highlight.Mode == models.HighlightModeRange && highlight.EndTimeMs != nil {
			timeText = fmt.Sprintf("%s-%s", formatReportHighlightMs(highlight.StartTimeMs), formatReportHighlightMs(*highlight.EndTimeMs))
		}
		description := strings.TrimSpace(highlight.Description)
		if description == "" {
			description = "暂无描述"
		}
		rows = append(rows, docxTableRow{Cells: []docxTableCell{
			{Text: fmt.Sprintf("%02d", i+1), Center: true, Size: 18, Color: "475569", Shading: "FFFFFF"},
			docxValueCell(timeText),
			docxValueCell(fmt.Sprintf("%s / %s", reportHighlightMarkerLabel(highlight.MarkerType), reportHighlightTagLabel(highlight.TagType))),
			docxValueCell(description),
		}})
	}
	lines = append(lines, docxParagraph{Table: &docxTable{ColWidths: []int{800, 1500, 2100, 4626}, Rows: rows}})
	if len(highlights) > limit {
		lines = append(lines, docxParagraph{Text: fmt.Sprintf("其余 %d 条关键标记已省略，可在订单详情中查看完整列表。", len(highlights)-limit), Size: 18, Spacing: 80})
	}
	return lines
}

func formatReportHighlightMs(ms int) string {
	if ms < 0 {
		ms = 0
	}
	totalSeconds := ms / 1000
	return fmt.Sprintf("%02d:%02d", totalSeconds/60, totalSeconds%60)
}

func reportHighlightMarkerLabel(markerType models.HighlightMarkerType) string {
	switch markerType {
	case models.HighlightMarkerIssue:
		return "待改进问题"
	case models.HighlightMarkerObservation:
		return "战术观察"
	default:
		return "精彩表现"
	}
}

func reportHighlightTagLabel(tagType models.HighlightTagType) string {
	labels := map[models.HighlightTagType]string{
		models.HighlightGoal:             "进球",
		models.HighlightAssist:           "助攻",
		models.HighlightSteal:            "抢断",
		models.HighlightSave:             "扑救",
		models.HighlightDribble:          "过人",
		models.HighlightPass:             "关键传球",
		models.HighlightDefense:          "防守关键",
		models.HighlightPositioningError: "站位问题",
		models.HighlightDecisionError:    "决策问题",
		models.HighlightTurnover:         "失误",
		models.HighlightRecoverySlow:     "回防不及时",
		models.HighlightTacticalNote:     "战术观察",
		models.HighlightOffBallRun:       "无球跑动",
	}
	if label, ok := labels[tagType]; ok {
		return label
	}
	return string(tagType)
}

func averageScoreValues(values ...float64) float64 {
	sum := 0.0
	count := 0
	for _, value := range values {
		if value <= 0 {
			continue
		}
		sum += value
		count++
	}
	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

func scoreBarLine(label string, score float64, max float64) string {
	if max <= 0 {
		max = 10
	}
	if score < 0 {
		score = 0
	}
	if score > max {
		score = max
	}
	filled := int(score/max*10 + 0.5)
	if filled > 10 {
		filled = 10
	}
	return fmt.Sprintf("%s：%s%s %.1f / %.0f", label, strings.Repeat("█", filled), strings.Repeat("░", 10-filled), score, max)
}

func writeSimpleDocx(path string, paragraphs []docxParagraph) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	zw := zip.NewWriter(file)
	images := collectDocxImages(paragraphs)
	if err := writeDocxZipEntry(zw, "[Content_Types].xml", docxContentTypesXML); err != nil {
		_ = zw.Close()
		return err
	}
	if err := writeDocxZipEntry(zw, "_rels/.rels", docxRelsXML); err != nil {
		_ = zw.Close()
		return err
	}
	if err := writeDocxZipEntry(zw, "word/_rels/document.xml.rels", buildDocxDocumentRelsXML(images)); err != nil {
		_ = zw.Close()
		return err
	}
	if err := writeDocxZipEntry(zw, "word/footer1.xml", docxFooterXML); err != nil {
		_ = zw.Close()
		return err
	}
	if err := writeDocxZipEntry(zw, "word/numbering.xml", docxNumberingXML); err != nil {
		_ = zw.Close()
		return err
	}
	for _, image := range images {
		if err := writeDocxBinaryEntry(zw, "word/media/"+image.FileName, image.Data); err != nil {
			_ = zw.Close()
			return err
		}
	}
	if err := writeDocxZipEntry(zw, "word/document.xml", buildDocxDocumentXML(paragraphs, images)); err != nil {
		_ = zw.Close()
		return err
	}
	return zw.Close()
}

type docxEmbeddedImage struct {
	Path      string
	RelID     string
	FileName  string
	Data      []byte
	WidthEMU  int64
	HeightEMU int64
}

func collectDocxImages(paragraphs []docxParagraph) map[string]docxEmbeddedImage {
	images := map[string]docxEmbeddedImage{}
	next := 1
	for _, paragraph := range paragraphs {
		path := strings.TrimSpace(paragraph.ImagePath)
		if path == "" {
			continue
		}
		if _, ok := images[path]; ok {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		widthEMU, heightEMU := docxImageSizeEMU(path, paragraph.ImageWidthPt)
		fileName := fmt.Sprintf("image%d%s", next, strings.ToLower(filepath.Ext(path)))
		if filepath.Ext(fileName) == "" {
			fileName += ".png"
		}
		images[path] = docxEmbeddedImage{
			Path:      path,
			RelID:     fmt.Sprintf("rIdImage%d", next),
			FileName:  fileName,
			Data:      data,
			WidthEMU:  widthEMU,
			HeightEMU: heightEMU,
		}
		next++
	}
	return images
}

func docxImageSizeEMU(path string, requestedWidthPt float64) (int64, int64) {
	if requestedWidthPt <= 0 {
		requestedWidthPt = 320
	}
	const emuPerPoint = 12700.0
	widthEMU := int64(requestedWidthPt * emuPerPoint)
	heightEMU := int64(requestedWidthPt * 0.28 * emuPerPoint)
	file, err := os.Open(path)
	if err != nil {
		return widthEMU, heightEMU
	}
	defer file.Close()
	config, _, err := image.DecodeConfig(file)
	if err != nil || config.Width <= 0 || config.Height <= 0 {
		return widthEMU, heightEMU
	}
	heightPt := requestedWidthPt * float64(config.Height) / float64(config.Width)
	return widthEMU, int64(heightPt * emuPerPoint)
}

func writeDocxZipEntry(zw *zip.Writer, name string, content string) error {
	header := &zip.FileHeader{Name: name, Method: zip.Deflate}
	w, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(content))
	return err
}

func writeDocxBinaryEntry(zw *zip.Writer, name string, content []byte) error {
	header := &zip.FileHeader{Name: name, Method: zip.Deflate}
	w, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = w.Write(content)
	return err
}

func writeSimplePDF(path string, paragraphs []docxParagraph) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	pages := buildPDFPages(paragraphs)
	if len(pages) == 0 {
		pages = []pdfPage{{Lines: []pdfLine{{Text: "少年球探视频分析报告", Size: 20, Center: true}}}}
	}

	type pdfObject struct {
		Body []byte
	}

	pdfImages := collectPDFImages(pages)
	objects := make([]pdfObject, 0, 6+len(pdfImages)+len(pages)*2)
	objects = append(objects,
		pdfObject{Body: []byte("<< /Type /Catalog /Pages 2 0 R >>")},
	)

	pageRefs := make([]int, len(pages))

	objects = append(objects,
		pdfObject{Body: []byte(fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>", buildPDFKidRefs(len(pages)), len(pages)))},
		pdfObject{Body: []byte("<< /Type /Font /Subtype /Type0 /BaseFont /STSong-Light /Encoding /UniGB-UCS2-H /DescendantFonts [4 0 R] >>")},
		pdfObject{Body: []byte("<< /Type /Font /Subtype /CIDFontType0 /BaseFont /STSong-Light /CIDSystemInfo << /Registry (Adobe) /Ordering (GB1) /Supplement 0 >> /DW 1000 /FontDescriptor 5 0 R >>")},
		pdfObject{Body: []byte("<< /Type /FontDescriptor /FontName /STSong-Light /Flags 4 /FontBBox [0 -200 1000 900] /ItalicAngle 0 /Ascent 880 /Descent -120 /CapHeight 700 /StemV 80 >>")},
	)

	for idx := range pdfImages {
		imageObj := &pdfImages[idx]
		imageObj.ObjectNo = len(objects) + 1
		objects = append(objects, pdfObject{Body: buildPDFImageObject(imageObj)})
	}

	imageNamesByPath := map[string]string{}
	imageResourceRefs := make([]string, 0, len(pdfImages))
	for _, imageObj := range pdfImages {
		imageNamesByPath[imageObj.Path] = imageObj.Name
		imageResourceRefs = append(imageResourceRefs, fmt.Sprintf("/%s %d 0 R", imageObj.Name, imageObj.ObjectNo))
	}
	xObjectResource := ""
	if len(imageResourceRefs) > 0 {
		xObjectResource = fmt.Sprintf(" /XObject << %s >>", strings.Join(imageResourceRefs, " "))
	}

	pageObjectBase := len(objects) + 1
	for i, page := range pages {
		content := page.render(imageNamesByPath, i+1, len(pages))
		contentObjectNumber := pageObjectBase + i*2 + 1
		pageObjectNumber := pageObjectBase + i*2
		pageBody := fmt.Sprintf(
			"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Contents %d 0 R /Resources << /Font << /F1 3 0 R >>%s >> >>",
			contentObjectNumber,
			xObjectResource,
		)
		objects = append(objects,
			pdfObject{Body: []byte(pageBody)},
			pdfObject{Body: []byte(content)},
		)
		pageRefs[i] = pageObjectNumber
	}

	// 重建 Pages object，确保引用号正确
	objects[1].Body = []byte(fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>", buildPDFRefs(pageRefs), len(pages)))

	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n%\xE2\xE3\xCF\xD3\n")

	offsets := make([]int, 0, len(objects)+1)
	offsets = append(offsets, 0)
	for idx, object := range objects {
		offsets = append(offsets, buf.Len())
		fmt.Fprintf(&buf, "%d 0 obj\n", idx+1)
		buf.Write(object.Body)
		buf.WriteString("\nendobj\n")
	}

	xrefStart := buf.Len()
	fmt.Fprintf(&buf, "xref\n0 %d\n", len(objects)+1)
	buf.WriteString("0000000000 65535 f \n")
	for _, off := range offsets[1:] {
		fmt.Fprintf(&buf, "%010d 00000 n \n", off)
	}
	fmt.Fprintf(&buf, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n", len(objects)+1, xrefStart)
	buf.WriteString("%%EOF\n")

	_, err = file.Write(buf.Bytes())
	return err
}

type pdfLine struct {
	Text   string
	Size   float64
	Center bool
	X      float64
	Y      float64
	Color  pdfColor
}

type pdfColor struct {
	R float64
	G float64
	B float64
}

type pdfRect struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
	Fill   pdfColor
}

type pdfImagePlacement struct {
	Path   string
	X      float64
	Y      float64
	Width  float64
	Height float64
}

type pdfPage struct {
	Lines  []pdfLine
	Images []pdfImagePlacement
	Rects  []pdfRect
}

func (p pdfPage) render(imageNamesByPath map[string]string, pageNumber int, pageCount int) string {
	var buf strings.Builder
	buf.WriteString("<< /Length ")
	content := p.contentBytes(imageNamesByPath, pageNumber, pageCount)
	buf.WriteString(strconv.Itoa(len(content)))
	buf.WriteString(" >>\nstream\n")
	buf.Write(content)
	buf.WriteString("\nendstream")
	return buf.String()
}

func (p pdfPage) contentBytes(imageNamesByPath map[string]string, pageNumber int, pageCount int) []byte {
	var buf strings.Builder
	for _, rect := range p.Rects {
		buf.WriteString(fmt.Sprintf(
			"q %.3f %.3f %.3f rg %.2f %.2f %.2f %.2f re f Q\n",
			clampPDFColor(rect.Fill.R),
			clampPDFColor(rect.Fill.G),
			clampPDFColor(rect.Fill.B),
			rect.X,
			rect.Y,
			rect.Width,
			rect.Height,
		))
	}
	for _, image := range p.Images {
		name := imageNamesByPath[image.Path]
		if name == "" {
			continue
		}
		buf.WriteString(fmt.Sprintf("q %.2f 0 0 %.2f %.2f %.2f cm /%s Do Q\n", image.Width, image.Height, image.X, image.Y, name))
	}
	leftMargin := 54.0
	pageWidth := 595.0
	for _, line := range p.Lines {
		if line.Text == "" {
			continue
		}
		y := line.Y
		if y <= 0 {
			y = 800
		}
		x := leftMargin
		if line.X > 0 {
			x = line.X
		} else if line.Center {
			width := estimatePDFTextWidth(line.Text, line.Size)
			usableWidth := pageWidth - leftMargin*2
			if width < usableWidth {
				x = leftMargin + (usableWidth-width)/2
			}
		}
		buf.WriteString("BT ")
		color := line.Color
		if color == (pdfColor{}) {
			color = pdfColor{R: 0.10, G: 0.14, B: 0.22}
		}
		buf.WriteString(fmt.Sprintf("%.3f %.3f %.3f rg ", clampPDFColor(color.R), clampPDFColor(color.G), clampPDFColor(color.B)))
		buf.WriteString(fmt.Sprintf("/F1 %.1f Tf ", line.sizeOrDefault()))
		buf.WriteString(fmt.Sprintf("1 0 0 1 %.2f %.2f Tm ", x, y))
		buf.WriteString("<")
		buf.WriteString(encodeUTF16BEHex(line.Text))
		buf.WriteString("> Tj ET\n")
	}
	footer := fmt.Sprintf("少年球探 uscout.cn  |  第 %d / %d 页", pageNumber, pageCount)
	buf.WriteString("BT 0.430 0.480 0.560 rg /F1 9.0 Tf ")
	buf.WriteString("1 0 0 1 210.00 28.00 Tm <")
	buf.WriteString(encodeUTF16BEHex(footer))
	buf.WriteString("> Tj ET\n")
	return []byte(buf.String())
}

func (l pdfLine) sizeOrDefault() float64 {
	if l.Size <= 0 {
		return 18
	}
	return l.Size
}

func clampPDFColor(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func pdfColorFromHex(value string, fallback pdfColor) pdfColor {
	value = normalizeDocxHexColor(value)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseUint(value, 16, 32)
	if err != nil {
		return fallback
	}
	return pdfColor{
		R: float64((parsed>>16)&0xFF) / 255,
		G: float64((parsed>>8)&0xFF) / 255,
		B: float64(parsed&0xFF) / 255,
	}
}

func buildPDFPages(paragraphs []docxParagraph) []pdfPage {
	const (
		pageWidth    = 595.0
		pageHeight   = 842.0
		leftMargin   = 54.0
		rightMargin  = 54.0
		topMargin    = 42.0
		bottomMargin = 60.0
	)
	usableWidth := pageWidth - leftMargin - rightMargin
	ink := pdfColor{R: 0.10, G: 0.14, B: 0.22}
	mutedInk := pdfColor{R: 0.36, G: 0.42, B: 0.52}
	brandBlue := pdfColor{R: 0.00, G: 0.57, B: 0.80}
	deepNavy := pdfColor{R: 0.04, G: 0.08, B: 0.14}
	lightCanvas := pdfColor{R: 0.96, G: 0.98, B: 1.00}
	white := pdfColor{R: 1, G: 1, B: 1}

	pageNumber := 0
	newContentPage := func() pdfPage {
		pageNumber++
		page := pdfPage{
			Lines: []pdfLine{},
			Rects: []pdfRect{
				{X: 0, Y: 0, Width: pageWidth, Height: pageHeight, Fill: lightCanvas},
				{X: 0, Y: pageHeight - 30, Width: pageWidth, Height: 30, Fill: deepNavy},
				{X: leftMargin, Y: pageHeight - 52, Width: pageWidth - leftMargin - rightMargin, Height: 1.2, Fill: pdfColor{R: 0.82, G: 0.88, B: 0.94}},
			},
		}
		if pageNumber == 1 {
			page.Rects = []pdfRect{
				{X: 0, Y: 0, Width: pageWidth, Height: pageHeight, Fill: deepNavy},
				{X: 0, Y: 0, Width: 18, Height: pageHeight, Fill: brandBlue},
				{X: 18, Y: 0, Width: 5, Height: pageHeight, Fill: pdfColor{R: 0.20, G: 0.85, B: 0.95}},
				{X: 54, Y: 178, Width: 487, Height: 1.2, Fill: pdfColor{R: 0.22, G: 0.34, B: 0.46}},
			}
		}
		return page
	}
	pages := make([]pdfPage, 0, 2)
	current := newContentPage()
	currentY := 800.0
	appendPage := func() {
		if len(current.Lines) > 0 || len(current.Images) > 0 {
			pages = append(pages, current)
		}
		current = newContentPage()
		currentY = 800.0
	}
	appendPDFTable := func(table *docxTable) {
		if table == nil || len(table.Rows) == 0 {
			return
		}
		colWidths := table.ColWidths
		if len(colWidths) == 0 {
			colWidths = []int{9026}
		}
		totalWidth := 0
		for _, width := range colWidths {
			if width > 0 {
				totalWidth += width
			}
		}
		if totalWidth <= 0 {
			totalWidth = 9026
		}
		pdfColWidths := make([]float64, len(colWidths))
		for i, width := range colWidths {
			pdfColWidths[i] = usableWidth * float64(width) / float64(totalWidth)
		}
		border := pdfColor{R: 0.84, G: 0.88, B: 0.93}
		type preparedPDFCell struct {
			Cell  docxTableCell
			Lines []string
			Width float64
			Size  float64
		}
		type preparedPDFRow struct {
			Header bool
			Cells  []preparedPDFCell
			Height float64
		}
		prepareRow := func(row docxTableRow) preparedPDFRow {
			prepared := preparedPDFRow{Header: row.Header, Cells: make([]preparedPDFCell, 0, len(row.Cells))}
			for i, cell := range row.Cells {
				colWidth := pdfColWidths[len(pdfColWidths)-1]
				if i < len(pdfColWidths) {
					colWidth = pdfColWidths[i]
				}
				size := float64(cell.Size)
				if size <= 0 {
					size = 9.5
				} else {
					size = size / 2
				}
				lines := wrapPDFText(strings.ReplaceAll(cell.Text, "\n", " "), size, colWidth-10)
				if len(lines) == 0 {
					lines = []string{""}
				}
				prepared.Cells = append(prepared.Cells, preparedPDFCell{Cell: cell, Lines: lines, Width: colWidth, Size: size})
				height := float64(len(lines))*lineHeightForPDF(size) + 8
				if height > prepared.Height {
					prepared.Height = height
				}
			}
			if prepared.Height < 24 {
				prepared.Height = 24
			}
			return prepared
		}
		drawRow := func(row preparedPDFRow) {
			x := leftMargin
			for _, preparedCell := range row.Cells {
				cell := preparedCell.Cell
				colWidth := preparedCell.Width
				shading := pdfColorFromHex(firstNonEmptyReportText(cell.Shading, "FFFFFF"), pdfColor{R: 1, G: 1, B: 1})
				if row.Header {
					shading = pdfColorFromHex(firstNonEmptyReportText(cell.Shading, "E0F2FE"), pdfColor{R: 0.88, G: 0.97, B: 1.00})
				}
				current.Rects = append(current.Rects,
					pdfRect{X: x, Y: currentY - row.Height, Width: colWidth, Height: row.Height, Fill: shading},
					pdfRect{X: x, Y: currentY - 0.6, Width: colWidth, Height: 0.6, Fill: border},
					pdfRect{X: x, Y: currentY - row.Height, Width: colWidth, Height: 0.6, Fill: border},
					pdfRect{X: x, Y: currentY - row.Height, Width: 0.6, Height: row.Height, Fill: border},
					pdfRect{X: x + colWidth - 0.6, Y: currentY - row.Height, Width: 0.6, Height: row.Height, Fill: border},
				)
				lineColor := pdfColorFromHex(firstNonEmptyReportText(cell.Color, "0F172A"), ink)
				textY := currentY - 7 - preparedCell.Size
				for _, line := range preparedCell.Lines {
					textX := x + 5
					if cell.Center {
						textWidth := estimatePDFTextWidth(line, preparedCell.Size)
						if textWidth < colWidth-10 {
							textX = x + (colWidth-textWidth)/2
						}
					}
					current.Lines = append(current.Lines, pdfLine{Text: line, Size: preparedCell.Size, X: textX, Y: textY, Color: lineColor})
					textY -= lineHeightForPDF(preparedCell.Size)
				}
				x += colWidth
			}
			currentY -= row.Height
		}
		var headerRow *preparedPDFRow
		for _, row := range table.Rows {
			prepared := prepareRow(row)
			if row.Header {
				headerCopy := prepared
				headerRow = &headerCopy
			}
			if currentY-prepared.Height < bottomMargin {
				appendPage()
				if headerRow != nil && !prepared.Header {
					if currentY-headerRow.Height >= bottomMargin {
						drawRow(*headerRow)
					}
				}
			}
			drawRow(prepared)
		}
		currentY -= 10
	}

	_ = topMargin
	isCoverPage := func() bool {
		return pageNumber == 1 && len(pages) == 0
	}
	for _, paragraph := range paragraphs {
		if paragraph.PageBreak {
			appendPage()
			continue
		}
		if paragraph.PageBreakBefore && (len(current.Lines) > 0 || len(current.Images) > 0) {
			appendPage()
		}
		if paragraph.Table != nil {
			appendPDFTable(paragraph.Table)
			continue
		}
		if strings.TrimSpace(paragraph.ImagePath) != "" {
			width, height := pdfImageSizePoints(paragraph.ImagePath, paragraph.ImageWidthPt)
			if currentY < bottomMargin+height {
				appendPage()
			}
			x := leftMargin
			if paragraph.Center && width < usableWidth {
				x = leftMargin + (usableWidth-width)/2
			}
			current.Images = append(current.Images, pdfImagePlacement{
				Path:   paragraph.ImagePath,
				X:      x,
				Y:      currentY - height,
				Width:  width,
				Height: height,
			})
			extraSpacing := 22.0
			if isCoverPage() && paragraph.Spacing > 0 {
				extraSpacing += float64(paragraph.Spacing) / 5
			}
			currentY -= height + extraSpacing
			continue
		}
		if strings.TrimSpace(paragraph.Text) == "" {
			current.Lines = append(current.Lines, pdfLine{Text: "", Size: float64(paragraph.Size), Y: currentY})
			currentY -= lineHeightForPDF(float64(paragraph.Size)) * 0.5
			continue
		}
		size := float64(paragraph.Size)
		if size <= 0 {
			size = 18
		}
		lineColor := ink
		if isCoverPage() {
			switch {
			case paragraph.Size >= 34:
				size = 30
			case paragraph.Size >= 24:
				size = 20
			case paragraph.Size <= 18:
				size = 12
			}
			lineColor = white
			if paragraph.Size <= 18 {
				lineColor = pdfColor{R: 0.77, G: 0.84, B: 0.91}
			}
			if paragraph.Size >= 34 {
				lineColor = pdfColor{R: 0.30, G: 0.92, B: 1.00}
			}
		} else if paragraph.Bold && paragraph.Size >= 23 {
			lineColor = deepNavy
		} else if paragraph.Size <= 18 {
			lineColor = mutedInk
		}
		if !isCoverPage() && paragraph.Color != "" {
			lineColor = pdfColorFromHex(paragraph.Color, lineColor)
		}
		text := paragraph.Text
		wrapWidth := usableWidth
		if isCoverPage() {
			wrapWidth = usableWidth - 70
		} else if paragraph.Bullet {
			text = "• " + text
			wrapWidth = usableWidth - 18
		}
		wrapped := wrapPDFText(text, size, wrapWidth)
		if len(wrapped) == 0 {
			wrapped = []string{text}
		}
		afterSpacing := lineHeightForPDF(size) * 0.25
		if paragraph.Spacing > 0 {
			afterSpacing = float64(paragraph.Spacing) / 20
		}
		if isCoverPage() && paragraph.Spacing > 0 {
			afterSpacing += float64(paragraph.Spacing) / 14
		}
		blockHeight := float64(len(wrapped))*lineHeightForPDF(size) + afterSpacing
		if !isCoverPage() && currentY-blockHeight < bottomMargin {
			appendPage()
		}
		if !isCoverPage() {
			if shading := pdfColorFromHex(paragraph.Shading, pdfColor{}); shading != (pdfColor{}) {
				current.Rects = append(current.Rects, pdfRect{
					X:      leftMargin - 8,
					Y:      currentY - blockHeight + afterSpacing*0.4,
					Width:  usableWidth + 16,
					Height: blockHeight,
					Fill:   shading,
				})
			}
		}
		for lineIndex, line := range wrapped {
			if currentY < bottomMargin+lineHeightForPDF(size) {
				appendPage()
			}
			if lineIndex == 0 && !isCoverPage() && paragraph.Bold && paragraph.Size >= 23 {
				current.Rects = append(current.Rects, pdfRect{
					X:      leftMargin - 13,
					Y:      currentY - 4,
					Width:  4,
					Height: lineHeightForPDF(size),
					Fill:   brandBlue,
				})
			}
			lineX := 0.0
			if !isCoverPage() && paragraph.Bullet {
				lineX = leftMargin + 16
			}
			current.Lines = append(current.Lines, pdfLine{Text: line, Size: size, Center: paragraph.Center, X: lineX, Y: currentY, Color: lineColor})
			currentY -= lineHeightForPDF(size)
		}
		currentY -= afterSpacing
	}
	if len(current.Lines) > 0 || len(current.Images) > 0 {
		pages = append(pages, current)
	}
	return pages
}

func pdfImageSizePoints(path string, requestedWidth float64) (float64, float64) {
	if requestedWidth <= 0 {
		requestedWidth = 320
	}
	height := requestedWidth * 0.28
	file, err := os.Open(path)
	if err != nil {
		return requestedWidth, height
	}
	defer file.Close()
	config, _, err := image.DecodeConfig(file)
	if err != nil || config.Width <= 0 || config.Height <= 0 {
		return requestedWidth, height
	}
	return requestedWidth, requestedWidth * float64(config.Height) / float64(config.Width)
}

type pdfEmbeddedImage struct {
	Path     string
	Name     string
	ObjectNo int
	Width    int
	Height   int
	Data     []byte
}

func collectPDFImages(pages []pdfPage) []pdfEmbeddedImage {
	images := make([]pdfEmbeddedImage, 0)
	seen := map[string]bool{}
	for _, page := range pages {
		for _, placement := range page.Images {
			path := strings.TrimSpace(placement.Path)
			if path == "" || seen[path] {
				continue
			}
			imageObj, err := loadPDFImage(path, len(images)+1)
			if err != nil {
				continue
			}
			images = append(images, imageObj)
			seen[path] = true
		}
	}
	return images
}

func loadPDFImage(path string, index int) (pdfEmbeddedImage, error) {
	file, err := os.Open(path)
	if err != nil {
		return pdfEmbeddedImage{}, err
	}
	defer file.Close()
	src, _, err := image.Decode(file)
	if err != nil {
		return pdfEmbeddedImage{}, err
	}
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= 0 || height <= 0 {
		return pdfEmbeddedImage{}, fmt.Errorf("invalid image size")
	}
	maxDim := 1200
	step := 1
	if width > maxDim || height > maxDim {
		widthStep := (width + maxDim - 1) / maxDim
		heightStep := (height + maxDim - 1) / maxDim
		if widthStep > heightStep {
			step = widthStep
		} else {
			step = heightStep
		}
	}
	outWidth := (width + step - 1) / step
	outHeight := (height + step - 1) / step
	raw := make([]byte, 0, outWidth*outHeight*3)
	for y := 0; y < outHeight; y++ {
		srcY := bounds.Min.Y + y*step
		if srcY >= bounds.Max.Y {
			srcY = bounds.Max.Y - 1
		}
		for x := 0; x < outWidth; x++ {
			srcX := bounds.Min.X + x*step
			if srcX >= bounds.Max.X {
				srcX = bounds.Max.X - 1
			}
			r, g, b, a := src.At(srcX, srcY).RGBA()
			alpha := float64(a) / 65535.0
			rr := uint8(float64(r>>8)*alpha + 255*(1-alpha))
			gg := uint8(float64(g>>8)*alpha + 255*(1-alpha))
			bb := uint8(float64(b>>8)*alpha + 255*(1-alpha))
			raw = append(raw, rr, gg, bb)
		}
	}
	var compressed bytes.Buffer
	zw := zlib.NewWriter(&compressed)
	if _, err := zw.Write(raw); err != nil {
		_ = zw.Close()
		return pdfEmbeddedImage{}, err
	}
	if err := zw.Close(); err != nil {
		return pdfEmbeddedImage{}, err
	}
	return pdfEmbeddedImage{
		Path:   path,
		Name:   fmt.Sprintf("Im%d", index),
		Width:  outWidth,
		Height: outHeight,
		Data:   compressed.Bytes(),
	}, nil
}

func buildPDFImageObject(imageObj *pdfEmbeddedImage) []byte {
	if imageObj == nil {
		return nil
	}
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "<< /Type /XObject /Subtype /Image /Width %d /Height %d /ColorSpace /DeviceRGB /BitsPerComponent 8 /Filter /FlateDecode /Length %d >>\nstream\n", imageObj.Width, imageObj.Height, len(imageObj.Data))
	buf.Write(imageObj.Data)
	buf.WriteString("\nendstream")
	return buf.Bytes()
}

func buildPDFKidRefs(pageCount int) string {
	refs := make([]string, 0, pageCount)
	for i := 0; i < pageCount; i++ {
		refs = append(refs, fmt.Sprintf("%d 0 R", 6+i*2))
	}
	return strings.Join(refs, " ")
}

func buildPDFRefs(pageRefs []int) string {
	refs := make([]string, 0, len(pageRefs))
	for _, ref := range pageRefs {
		refs = append(refs, fmt.Sprintf("%d 0 R", ref))
	}
	return strings.Join(refs, " ")
}

func wrapPDFText(text string, size float64, usableWidth float64) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	maxUnits := usableWidth / size
	if maxUnits < 8 {
		maxUnits = 8
	}
	paragraphs := strings.Split(text, "\n")
	lines := make([]string, 0, len(paragraphs))
	for _, paragraph := range paragraphs {
		trimmed := strings.TrimSpace(paragraph)
		if trimmed == "" {
			lines = append(lines, "")
			continue
		}
		runes := []rune(trimmed)
		current := strings.Builder{}
		currentWidth := 0.0
		for _, r := range runes {
			w := pdfRuneWidth(r)
			if currentWidth+w > maxUnits && current.Len() > 0 {
				lines = append(lines, current.String())
				current.Reset()
				currentWidth = 0
			}
			current.WriteRune(r)
			currentWidth += w
		}
		if current.Len() > 0 {
			lines = append(lines, current.String())
		}
	}
	return lines
}

func lineHeightForPDF(size float64) float64 {
	if size <= 0 {
		size = 18
	}
	return size * 1.45
}

func estimatePDFTextWidth(text string, size float64) float64 {
	width := 0.0
	for _, r := range text {
		width += pdfRuneWidth(r)
	}
	return width * size
}

func pdfRuneWidth(r rune) float64 {
	switch {
	case r == ' ':
		return 0.60
	case r < 128:
		return 0.85
	default:
		return 1.0
	}
}

func encodeUTF16BEHex(text string) string {
	data := []byte{0xFE, 0xFF}
	for _, r := range text {
		data = append(data, byte(r>>8), byte(r))
	}
	return strings.ToUpper(hex.EncodeToString(data))
}

func buildDocxDocumentXML(paragraphs []docxParagraph, images map[string]docxEmbeddedImage) string {
	var buf strings.Builder
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	buf.WriteString(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture">`)
	buf.WriteString(`<w:body>`)
	for _, paragraph := range paragraphs {
		buf.WriteString(renderDocxParagraph(paragraph, images))
	}
	buf.WriteString(`<w:sectPr><w:footerReference w:type="default" r:id="rIdFooter1"/><w:pgSz w:w="11906" w:h="16838"/><w:pgMar w:top="1440" w:right="1440" w:bottom="1440" w:left="1440" w:header="708" w:footer="708" w:gutter="0"/></w:sectPr>`)
	buf.WriteString(`</w:body></w:document>`)
	return buf.String()
}

func renderDocxParagraph(p docxParagraph, images map[string]docxEmbeddedImage) string {
	if p.Table != nil {
		return renderDocxTable(p.Table)
	}
	if p.PageBreak {
		return `<w:p><w:r><w:br w:type="page"/></w:r></w:p>`
	}
	if strings.TrimSpace(p.ImagePath) != "" {
		if imageObj, ok := images[p.ImagePath]; ok {
			return renderDocxImageParagraph(p, imageObj)
		}
	}
	if strings.TrimSpace(p.Text) == "" {
		return `<w:p/>`
	}
	size := p.Size
	if size <= 0 {
		size = 21
	}
	spacing := p.Spacing
	if spacing <= 0 {
		spacing = 80
	}
	before := p.Before
	color := normalizeDocxHexColor(firstNonEmptyReportText(p.Color, "111827"))
	var buf strings.Builder
	buf.WriteString(`<w:p>`)
	buf.WriteString(`<w:pPr>`)
	if p.Bullet {
		buf.WriteString(`<w:numPr><w:ilvl w:val="0"/><w:numId w:val="1"/></w:numPr>`)
		if p.IndentLeft <= 0 {
			p.IndentLeft = 720
		}
	}
	if p.PageBreakBefore {
		buf.WriteString(`<w:pageBreakBefore/>`)
	}
	if p.IndentLeft > 0 {
		if p.Bullet {
			buf.WriteString(fmt.Sprintf(`<w:ind w:left="%d" w:hanging="360"/>`, p.IndentLeft))
		} else {
			buf.WriteString(fmt.Sprintf(`<w:ind w:left="%d"/>`, p.IndentLeft))
		}
	}
	if p.Center {
		buf.WriteString(`<w:jc w:val="center"/>`)
	}
	buf.WriteString(fmt.Sprintf(`<w:spacing w:before="%d" w:after="%d"/>`, before, spacing))
	if borderColor := normalizeDocxHexColor(p.BorderBottom); borderColor != "" {
		buf.WriteString(fmt.Sprintf(`<w:pBdr><w:bottom w:val="single" w:sz="8" w:space="4" w:color="%s"/></w:pBdr>`, borderColor))
	}
	if shading := normalizeDocxHexColor(p.Shading); shading != "" {
		buf.WriteString(fmt.Sprintf(`<w:shd w:val="clear" w:color="auto" w:fill="%s"/>`, shading))
	}
	buf.WriteString(`</w:pPr>`)
	renderDocxTextRuns(&buf, strings.TrimSpace(p.Text), size, p.Bold, color)
	buf.WriteString(`</w:p>`)
	return buf.String()
}

func renderDocxTable(table *docxTable) string {
	if table == nil || len(table.Rows) == 0 {
		return `<w:p/>`
	}
	colWidths := table.ColWidths
	if len(colWidths) == 0 {
		colWidths = []int{9026}
	}
	var buf strings.Builder
	buf.WriteString(`<w:tbl>`)
	buf.WriteString(`<w:tblPr><w:tblW w:w="9026" w:type="dxa"/><w:tblLayout w:type="fixed"/><w:tblCellMar><w:top w:w="90" w:type="dxa"/><w:left w:w="120" w:type="dxa"/><w:bottom w:w="90" w:type="dxa"/><w:right w:w="120" w:type="dxa"/></w:tblCellMar><w:tblBorders><w:top w:val="single" w:sz="4" w:color="D7DEE8"/><w:left w:val="single" w:sz="4" w:color="D7DEE8"/><w:bottom w:val="single" w:sz="4" w:color="D7DEE8"/><w:right w:val="single" w:sz="4" w:color="D7DEE8"/><w:insideH w:val="single" w:sz="4" w:color="E2E8F0"/><w:insideV w:val="single" w:sz="4" w:color="E2E8F0"/></w:tblBorders></w:tblPr>`)
	buf.WriteString(`<w:tblGrid>`)
	for _, width := range colWidths {
		buf.WriteString(fmt.Sprintf(`<w:gridCol w:w="%d"/>`, width))
	}
	buf.WriteString(`</w:tblGrid>`)
	for _, row := range table.Rows {
		buf.WriteString(`<w:tr>`)
		if row.Header {
			buf.WriteString(`<w:trPr><w:tblHeader/></w:trPr>`)
		}
		for index, cell := range row.Cells {
			width := colWidths[len(colWidths)-1]
			if index < len(colWidths) {
				width = colWidths[index]
			}
			buf.WriteString(renderDocxTableCell(cell, width))
		}
		buf.WriteString(`</w:tr>`)
	}
	buf.WriteString(`</w:tbl>`)
	return buf.String()
}

func renderDocxTableCell(cell docxTableCell, width int) string {
	size := cell.Size
	if size <= 0 {
		size = 18
	}
	color := normalizeDocxHexColor(firstNonEmptyReportText(cell.Color, "0F172A"))
	shading := normalizeDocxHexColor(firstNonEmptyReportText(cell.Shading, "FFFFFF"))
	var buf strings.Builder
	buf.WriteString(`<w:tc>`)
	buf.WriteString(fmt.Sprintf(`<w:tcPr><w:tcW w:w="%d" w:type="dxa"/><w:shd w:val="clear" w:color="auto" w:fill="%s"/><w:vAlign w:val="center"/></w:tcPr>`, width, shading))
	buf.WriteString(`<w:p><w:pPr>`)
	if cell.Center {
		buf.WriteString(`<w:jc w:val="center"/>`)
	}
	buf.WriteString(`<w:spacing w:after="20"/>`)
	buf.WriteString(`</w:pPr>`)
	renderDocxTextRuns(&buf, strings.TrimSpace(cell.Text), size, cell.Bold, color)
	buf.WriteString(`</w:p></w:tc>`)
	return buf.String()
}

func renderDocxTextRuns(buf *strings.Builder, text string, size int, bold bool, color string) {
	parts := strings.Split(text, "\n")
	for index, part := range parts {
		if index > 0 {
			buf.WriteString(`<w:r><w:br/></w:r>`)
		}
		buf.WriteString(`<w:r><w:rPr>`)
		buf.WriteString(`<w:rFonts w:ascii="Microsoft YaHei" w:eastAsia="Microsoft YaHei" w:hAnsi="Microsoft YaHei"/>`)
		if bold {
			buf.WriteString(`<w:b/>`)
		}
		if color != "" {
			buf.WriteString(fmt.Sprintf(`<w:color w:val="%s"/>`, color))
		}
		buf.WriteString(fmt.Sprintf(`<w:sz w:val="%d"/>`, size))
		buf.WriteString(`</w:rPr><w:t xml:space="preserve">`)
		buf.WriteString(html.EscapeString(part))
		buf.WriteString(`</w:t></w:r>`)
	}
}

func renderDocxImageParagraph(p docxParagraph, imageObj docxEmbeddedImage) string {
	var buf strings.Builder
	buf.WriteString(`<w:p><w:pPr>`)
	if p.Center {
		buf.WriteString(`<w:jc w:val="center"/>`)
	}
	if p.Spacing > 0 {
		buf.WriteString(fmt.Sprintf(`<w:spacing w:after="%d"/>`, p.Spacing))
	}
	buf.WriteString(`</w:pPr><w:r><w:drawing>`)
	buf.WriteString(fmt.Sprintf(`<wp:inline distT="0" distB="0" distL="0" distR="0"><wp:extent cx="%d" cy="%d"/><wp:docPr id="1" name="%s"/><a:graphic><a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/picture"><pic:pic><pic:nvPicPr><pic:cNvPr id="0" name="%s"/><pic:cNvPicPr/></pic:nvPicPr><pic:blipFill><a:blip r:embed="%s"/><a:stretch><a:fillRect/></a:stretch></pic:blipFill><pic:spPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="%d" cy="%d"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom></pic:spPr></pic:pic></a:graphicData></a:graphic></wp:inline>`,
		imageObj.WidthEMU,
		imageObj.HeightEMU,
		html.EscapeString(firstNonEmptyReportText(p.Text, "少年球探品牌标识")),
		html.EscapeString(imageObj.FileName),
		imageObj.RelID,
		imageObj.WidthEMU,
		imageObj.HeightEMU,
	))
	buf.WriteString(`</w:drawing></w:r></w:p>`)
	return buf.String()
}

func markdownTextToDocxParagraphs(text string) []docxParagraph {
	lines := strings.Split(text, "\n")
	paragraphs := make([]docxParagraph, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case trimmed == "":
			paragraphs = append(paragraphs, docxParagraph{})
		case strings.HasPrefix(trimmed, "### "):
			paragraphs = append(paragraphs, docxParagraph{Text: strings.TrimSpace(strings.TrimPrefix(trimmed, "### ")), Bold: true, Size: 22, Color: "0B1726", Spacing: 100})
		case strings.HasPrefix(trimmed, "## "):
			paragraphs = append(paragraphs, docxParagraph{Text: strings.TrimSpace(strings.TrimPrefix(trimmed, "## ")), Bold: true, Size: 24, Color: "0B1726", Spacing: 120})
		case strings.HasPrefix(trimmed, "# "):
			paragraphs = append(paragraphs, docxParagraph{Text: strings.TrimSpace(strings.TrimPrefix(trimmed, "# ")), Bold: true, Size: 26, Color: "0B1726", BorderBottom: "00A6D6", Spacing: 140})
		case strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* "):
			paragraphs = append(paragraphs, docxParagraph{Text: strings.TrimSpace(trimmed[2:]), Size: 21, Color: "1F2937", Bullet: true, Spacing: 60})
		default:
			paragraphs = append(paragraphs, docxParagraph{Text: strings.Trim(trimmed, "*_"), Size: 21, Color: "1F2937", Spacing: 80})
		}
	}
	return paragraphs
}

func normalizeDocxHexColor(value string) string {
	value = strings.TrimSpace(strings.TrimPrefix(value, "#"))
	if len(value) != 6 {
		return ""
	}
	for _, r := range value {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return ""
		}
	}
	return strings.ToUpper(value)
}

func splitReportTextItems(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	items := make([]string, 0)
	for _, part := range strings.FieldsFunc(text, func(r rune) bool {
		return r == '\n' || r == ';' || r == '；'
	}) {
		item := strings.TrimSpace(strings.Trim(part, "-* "))
		if item != "" {
			items = append(items, item)
		}
	}
	if len(items) == 0 {
		return []string{text}
	}
	return items
}

func sanitizeReportFileNamePart(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		value = "未知球员"
	}
	replacer := strings.NewReplacer("/", "_", "\\", "_", ":", "_", "*", "_", "?", "_", `"`, "_", "<", "_", ">", "_", "|", "_")
	value = replacer.Replace(value)
	value = strings.Join(strings.Fields(value), " ")
	if value == "" {
		return "未知球员"
	}
	return value
}

func firstNonEmptyReportText(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func firstPositiveFloat(values ...float64) float64 {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

func userName(user *models.User) string {
	if user == nil {
		return ""
	}
	return user.Name
}

func userPosition(user *models.User) string {
	if user == nil {
		return ""
	}
	return user.Position
}

func userFoot(user *models.User) string {
	if user == nil {
		return ""
	}
	return firstNonEmptyReportText(user.Foot, user.DominantFoot)
}

func userHeight(user *models.User) float64 {
	if user == nil {
		return 0
	}
	return user.Height
}

func userWeight(user *models.User) float64 {
	if user == nil {
		return 0
	}
	return user.Weight
}

func userClub(user *models.User) string {
	if user == nil {
		return ""
	}
	return firstNonEmptyReportText(user.CurrentTeam, user.Club, user.School)
}

const docxContentTypesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Default Extension="png" ContentType="image/png"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
  <Override PartName="/word/footer1.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.footer+xml"/>
  <Override PartName="/word/numbering.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.numbering+xml"/>
</Types>`

const docxRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`

func buildDocxDocumentRelsXML(images map[string]docxEmbeddedImage) string {
	var buf strings.Builder
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	buf.WriteString(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	buf.WriteString(`<Relationship Id="rIdFooter1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/footer" Target="footer1.xml"/>`)
	buf.WriteString(`<Relationship Id="rIdNumbering" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/numbering" Target="numbering.xml"/>`)
	for _, imageObj := range images {
		buf.WriteString(fmt.Sprintf(`<Relationship Id="%s" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="media/%s"/>`, imageObj.RelID, html.EscapeString(imageObj.FileName)))
	}
	buf.WriteString(`</Relationships>`)
	return buf.String()
}

const docxNumberingXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:numbering xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:abstractNum w:abstractNumId="1">
    <w:lvl w:ilvl="0">
      <w:start w:val="1"/>
      <w:numFmt w:val="bullet"/>
      <w:lvlText w:val="•"/>
      <w:lvlJc w:val="left"/>
      <w:pPr><w:ind w:left="720" w:hanging="360"/></w:pPr>
      <w:rPr><w:rFonts w:ascii="Symbol" w:hAnsi="Symbol"/></w:rPr>
    </w:lvl>
  </w:abstractNum>
  <w:num w:numId="1"><w:abstractNumId w:val="1"/></w:num>
</w:numbering>`

const docxFooterXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:ftr xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:p>
    <w:pPr><w:jc w:val="center"/></w:pPr>
    <w:r>
      <w:rPr><w:rFonts w:ascii="Microsoft YaHei" w:eastAsia="Microsoft YaHei" w:hAnsi="Microsoft YaHei"/><w:sz w:val="16"/></w:rPr>
      <w:t>少年球探 uscout.cn ｜ 第 </w:t>
    </w:r>
    <w:r><w:fldChar w:fldCharType="begin"/></w:r>
    <w:r><w:instrText xml:space="preserve"> PAGE </w:instrText></w:r>
    <w:r><w:fldChar w:fldCharType="end"/></w:r>
    <w:r>
      <w:rPr><w:rFonts w:ascii="Microsoft YaHei" w:eastAsia="Microsoft YaHei" w:hAnsi="Microsoft YaHei"/><w:sz w:val="16"/></w:rPr>
      <w:t> 页</w:t>
    </w:r>
  </w:p>
</w:ftr>`

// buildVideoAnalysisReport 构建视频分析评分报告内容
func (g *ReportGenerator) buildVideoAnalysisReport(analysis *models.VideoAnalysis, analystName string) string {
	var buf bytes.Buffer

	// 解析 scores JSON
	var scores map[string]map[string]map[string]interface{}
	_ = json.Unmarshal([]byte(analysis.Scores), &scores)

	// 评分标签映射（snake_case key -> 中文标签）
	overallLabels := map[string]string{
		"ball_control":       "控球能力",
		"off_ball_movement":  "无球跑动",
		"pressing_awareness": "逼抢意识",
		"positioning":        "站位选择",
	}
	offenseLabels := map[string]string{
		"width_participation": "拉开宽度参与",
		"off_ball_support":    "无球支援",
		"one_v_one":           "1v1过人能力",
		"crossing_assist":     "传中/助攻",
		"combat_ability":      "对抗能力",
		"pace_rhythm":         "节奏把控",
		"pass_vision":         "传球视野",
		"body_posture":        "身体姿态",
	}
	defenseLabels := map[string]string{
		"defensive_commitment":  "防守投入度",
		"loss_recovery":         "丢球回追",
		"teammate_coordination": "队友协防配合",
		"second_ball":           "二点球争抢",
		"aerial_duel":           "空中争顶",
		"defensive_shape":       "防守阵型保持",
		"role_adjustment":       "角色调整能力",
		"defensive_rhythm":      "防守节奏",
	}

	buf.WriteString("# 球员视频分析评分报告\n\n")
	buf.WriteString(fmt.Sprintf("**球员姓名：** %s<br>\n", analysis.PlayerName))
	buf.WriteString(fmt.Sprintf("**球员位置：** %s<br>\n", analysis.PlayerPosition))
	buf.WriteString(fmt.Sprintf("**分析日期：** %s<br>\n", time.Now().Format("2006年1月2日")))
	buf.WriteString(fmt.Sprintf("**分析师：** %s<br>\n", analystName))
	buf.WriteString(fmt.Sprintf("**比赛名称：** %s<br>\n", analysis.MatchName))
	buf.WriteString(fmt.Sprintf("**对手：** %s<br>\n", analysis.Opponent))
	buf.WriteString(fmt.Sprintf("**综合评分：** %.1f / 100（%s）<br>\n", analysis.OverallScore, models.GetPotentialLevel(analysis.OverallScore)))
	buf.WriteString(fmt.Sprintf("**潜力等级：** %s<br>\n", models.GetPotentialLevel(analysis.OverallScore)))
	buf.WriteString("\n---\n\n")

	// 整体表现
	if overall, ok := scores["overall"]; ok {
		buf.WriteString("## 一、整体表现\n\n")
		for key, label := range overallLabels {
			if item, ok := overall[key]; ok {
				score, _ := item["score"].(float64)
				comment, _ := item["comment"].(string)
				buf.WriteString(fmt.Sprintf("### %s（评分：%.1f）\n\n", label, score/10))
				if comment != "" {
					buf.WriteString(fmt.Sprintf("%s\n\n", comment))
				}
			}
		}
	}

	// 进攻能力
	if offense, ok := scores["offense"]; ok {
		buf.WriteString("## 二、进攻能力\n\n")
		for key, label := range offenseLabels {
			if item, ok := offense[key]; ok {
				score, _ := item["score"].(float64)
				comment, _ := item["comment"].(string)
				buf.WriteString(fmt.Sprintf("### %s（评分：%.1f）\n\n", label, score/10))
				if comment != "" {
					buf.WriteString(fmt.Sprintf("%s\n\n", comment))
				}
			}
		}
	}

	// 防守能力
	if defense, ok := scores["defense"]; ok {
		buf.WriteString("## 三、防守能力\n\n")
		for key, label := range defenseLabels {
			if item, ok := defense[key]; ok {
				score, _ := item["score"].(float64)
				comment, _ := item["comment"].(string)
				buf.WriteString(fmt.Sprintf("### %s（评分：%.1f）\n\n", label, score/10))
				if comment != "" {
					buf.WriteString(fmt.Sprintf("%s\n\n", comment))
				}
			}
		}
	}

	// 综合评价摘要
	if analysis.Summary != "" {
		buf.WriteString("## 四、综合评价摘要\n\n")
		buf.WriteString(fmt.Sprintf("%s\n\n", analysis.Summary))
	}

	// 成长建议
	if analysis.Improvements != "" {
		buf.WriteString("## 五、成长建议\n\n")
		buf.WriteString(fmt.Sprintf("%s\n\n", analysis.Improvements))
	}

	// 分析师备注
	if analysis.AnalystNotes != "" {
		buf.WriteString("## 六、分析师备注\n\n")
		buf.WriteString(fmt.Sprintf("%s\n\n", analysis.AnalystNotes))
	}

	// 报告正文补充（如果有）
	if analysis.AIReport != "" {
		buf.WriteString("---\n\n")
		buf.WriteString("## 附：报告正文参考\n\n")
		buf.WriteString(analysis.AIReport)
		buf.WriteString("\n\n")
	}

	buf.WriteString(fmt.Sprintf("*报告生成时间：%s*\n", time.Now().Format("2006-01-02 15:04:05")))

	return buf.String()
}

// buildPlayerInfoFromVideoAnalysis 构建球员基础信息文档（从视频分析记录）
func (g *ReportGenerator) buildPlayerInfoFromVideoAnalysis(analysis *models.VideoAnalysis, user *models.User) string {
	var buf bytes.Buffer

	buf.WriteString("# 球员基础信息\n\n")
	buf.WriteString(fmt.Sprintf("**球员 ID：** %d<br>\n", analysis.UserID))
	buf.WriteString(fmt.Sprintf("**姓名：** %s<br>\n", analysis.PlayerName))
	buf.WriteString(fmt.Sprintf("**位置：** %s<br>\n", analysis.PlayerPosition))
	if user != nil {
		if user.Nickname != "" {
			buf.WriteString(fmt.Sprintf("**昵称：** %s<br>\n", user.Nickname))
		}
		if user.BirthDate != "" {
			buf.WriteString(fmt.Sprintf("**生日：** %s<br>\n", user.BirthDate))
		}
		if user.Gender != "" {
			buf.WriteString(fmt.Sprintf("**性别：** %s<br>\n", user.Gender))
		}
		buf.WriteString(fmt.Sprintf("**身高：** %.1f cm<br>\n", user.Height))
		buf.WriteString(fmt.Sprintf("**体重：** %.1f kg<br>\n", user.Weight))
		buf.WriteString(fmt.Sprintf("**惯用脚：** %s<br>\n", user.Foot))
		buf.WriteString(fmt.Sprintf("**地区：** %s %s<br>\n", user.Province, user.City))
		buf.WriteString(fmt.Sprintf("**俱乐部/球队：** %s<br>\n", user.Club))
		buf.WriteString(fmt.Sprintf("**学校：** %s<br>\n", user.School))
	}
	if analysis.PlayerTeam != "" {
		buf.WriteString(fmt.Sprintf("**当前球队：** %s<br>\n", analysis.PlayerTeam))
	}
	if analysis.PlayerFoot != "" {
		buf.WriteString(fmt.Sprintf("**惯用脚：** %s<br>\n", analysis.PlayerFoot))
	}
	buf.WriteString(fmt.Sprintf("**身高：** %.1f cm<br>\n", analysis.PlayerHeight))
	buf.WriteString(fmt.Sprintf("**体重：** %.1f kg<br>\n", analysis.PlayerWeight))

	// 比赛信息
	buf.WriteString("\n---\n\n## 比赛信息\n\n")
	buf.WriteString(fmt.Sprintf("- **比赛名称：** %s<br>\n", analysis.MatchName))
	buf.WriteString(fmt.Sprintf("- **比赛日期：** %s<br>\n", analysis.MatchDate))
	buf.WriteString(fmt.Sprintf("- **对手：** %s<br>\n", analysis.Opponent))
	if analysis.PlayTime > 0 {
		buf.WriteString(fmt.Sprintf("- **出场时间：** %d 分钟<br>\n", analysis.PlayTime))
	}
	if analysis.Goals > 0 || analysis.Assists > 0 {
		buf.WriteString(fmt.Sprintf("- **进球/助攻：** %d / %d<br>\n", analysis.Goals, analysis.Assists))
	}

	buf.WriteString("\n---\n\n")
	buf.WriteString(fmt.Sprintf("*文档生成时间：%s*\n", time.Now().Format("2006-01-02 15:04:05")))

	return buf.String()
}
