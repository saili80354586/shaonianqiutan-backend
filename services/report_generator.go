package services

import (
	"archive/zip"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
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
func (g *ReportGenerator) GenerateVideoAnalysisWordReport(analysis *models.VideoAnalysis, analystName string, user *models.User) (string, error) {
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

	paragraphs := buildVideoAnalysisWordParagraphs(analysis, analystName, user, playerName, version)
	if err := writeSimpleDocx(fullPath, paragraphs); err != nil {
		return "", fmt.Errorf("写入视频分析 Word 报告失败: %w", err)
	}
	return "/uploads/reports/" + fileName, nil
}

// GenerateVideoAnalysisPDFReport 生成视频分析正式 PDF 报告。
// 返回值为可写入 reports.pdf_url 的 Web 路径，文件实体写入 reportsDir。
func (g *ReportGenerator) GenerateVideoAnalysisPDFReport(analysis *models.VideoAnalysis, analystName string, user *models.User) (string, error) {
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

	paragraphs := buildVideoAnalysisWordParagraphs(analysis, analystName, user, playerName, version)
	if err := writeSimplePDF(fullPath, paragraphs); err != nil {
		return "", fmt.Errorf("写入视频分析 PDF 报告失败: %w", err)
	}
	return "/uploads/reports/" + fileName, nil
}

type docxParagraph struct {
	Text    string
	Bold    bool
	Center  bool
	Size    int
	Spacing int
}

func buildVideoAnalysisWordParagraphs(analysis *models.VideoAnalysis, analystName string, user *models.User, playerName string, version int) []docxParagraph {
	paragraphs := []docxParagraph{
		{Text: "少年球探视频分析报告", Bold: true, Center: true, Size: 36, Spacing: 240},
		{Text: fmt.Sprintf("球员：%s", playerName), Center: true, Size: 24, Spacing: 120},
		{Text: fmt.Sprintf("订单：%d  版本：v%d", analysis.OrderID, version), Center: true, Size: 20, Spacing: 80},
		{Text: fmt.Sprintf("模板：%s", VideoAnalysisReportTemplateVersion), Center: true, Size: 18, Spacing: 240},
		{},
		{Text: "一、球员与比赛信息", Bold: true, Size: 26, Spacing: 160},
	}

	appendNonEmptyDocxLine := func(label string, value any) {
		text := strings.TrimSpace(fmt.Sprint(value))
		if text == "" || text == "0" || text == "0.0" {
			return
		}
		paragraphs = append(paragraphs, docxParagraph{Text: fmt.Sprintf("%s：%s", label, text), Size: 21, Spacing: 80})
	}

	appendNonEmptyDocxLine("姓名", playerName)
	if analysis.PlayerAge > 0 {
		appendNonEmptyDocxLine("年龄", fmt.Sprintf("%d岁", analysis.PlayerAge))
	} else if user != nil && user.Age > 0 {
		appendNonEmptyDocxLine("年龄", fmt.Sprintf("%d岁", user.Age))
	}
	appendNonEmptyDocxLine("位置", firstNonEmptyReportText(analysis.PlayerPosition, userPosition(user)))
	appendNonEmptyDocxLine("惯用脚", firstNonEmptyReportText(analysis.PlayerFoot, userFoot(user)))
	appendNonEmptyDocxLine("当前球队", firstNonEmptyReportText(analysis.PlayerTeam, userClub(user)))
	appendNonEmptyDocxLine("比赛名称", analysis.MatchName)
	appendNonEmptyDocxLine("比赛日期", analysis.MatchDate)
	appendNonEmptyDocxLine("对手", analysis.Opponent)
	if analysis.PlayTime > 0 {
		appendNonEmptyDocxLine("出场时间", fmt.Sprintf("%d分钟", analysis.PlayTime))
	}
	appendNonEmptyDocxLine("分析师", firstNonEmptyReportText(analystName, "未知分析师"))

	paragraphs = append(paragraphs,
		docxParagraph{},
		docxParagraph{Text: "二、评分概览", Bold: true, Size: 26, Spacing: 160},
		docxParagraph{Text: fmt.Sprintf("综合评分：%.1f / 100", analysis.OverallScore), Size: 21, Spacing: 80},
		docxParagraph{Text: fmt.Sprintf("潜力等级：%s", models.GetPotentialLevel(analysis.OverallScore)), Size: 21, Spacing: 80},
	)
	if strings.TrimSpace(analysis.Summary) != "" {
		paragraphs = append(paragraphs, docxParagraph{Text: "综合评价", Bold: true, Size: 23, Spacing: 120})
		paragraphs = append(paragraphs, markdownTextToDocxParagraphs(analysis.Summary)...)
	}

	appendDocxListSection := func(title, text string) {
		items := splitReportTextItems(text)
		if len(items) == 0 {
			return
		}
		paragraphs = append(paragraphs, docxParagraph{Text: title, Bold: true, Size: 23, Spacing: 120})
		for _, item := range items {
			paragraphs = append(paragraphs, docxParagraph{Text: "• " + item, Size: 21, Spacing: 60})
		}
	}
	appendDocxListSection("核心优势", analysis.Strengths)
	appendDocxListSection("待提升领域", analysis.Weaknesses)
	appendDocxListSection("重点改进建议", analysis.Improvements)
	appendDocxListSection("分析师补充说明", analysis.AnalystNotes)

	if strings.TrimSpace(analysis.AIReport) != "" {
		paragraphs = append(paragraphs,
			docxParagraph{},
			docxParagraph{Text: "三、AI报告正文", Bold: true, Size: 26, Spacing: 160},
		)
		paragraphs = append(paragraphs, markdownTextToDocxParagraphs(analysis.AIReport)...)
	}

	paragraphs = append(paragraphs,
		docxParagraph{},
		docxParagraph{Text: "声明：本报告基于分析师录入的比赛信息、评分、高光片段与AI辅助生成内容形成，供青少年足球训练与成长规划参考，不作为医学、升学或职业签约承诺。", Size: 18, Spacing: 120},
		docxParagraph{Text: fmt.Sprintf("生成时间：%s", time.Now().Format("2006-01-02 15:04:05")), Center: true, Size: 18, Spacing: 80},
	)

	return paragraphs
}

func writeSimpleDocx(path string, paragraphs []docxParagraph) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	zw := zip.NewWriter(file)
	if err := writeDocxZipEntry(zw, "[Content_Types].xml", docxContentTypesXML); err != nil {
		_ = zw.Close()
		return err
	}
	if err := writeDocxZipEntry(zw, "_rels/.rels", docxRelsXML); err != nil {
		_ = zw.Close()
		return err
	}
	if err := writeDocxZipEntry(zw, "word/document.xml", buildDocxDocumentXML(paragraphs)); err != nil {
		_ = zw.Close()
		return err
	}
	return zw.Close()
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

	objects := make([]pdfObject, 0, 6+len(pages)*2)
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

	pageObjectBase := 6
	for i, page := range pages {
		content := page.render()
		contentObjectNumber := pageObjectBase + i*2 + 1
		pageObjectNumber := pageObjectBase + i*2
		pageBody := fmt.Sprintf(
			"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Contents %d 0 R /Resources << /Font << /F1 3 0 R >> >> >>",
			contentObjectNumber,
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
}

type pdfPage struct {
	Lines []pdfLine
}

func (p pdfPage) render() string {
	var buf strings.Builder
	buf.WriteString("<< /Length ")
	content := p.contentBytes()
	buf.WriteString(strconv.Itoa(len(content)))
	buf.WriteString(" >>\nstream\n")
	buf.Write(content)
	buf.WriteString("\nendstream")
	return buf.String()
}

func (p pdfPage) contentBytes() []byte {
	var buf strings.Builder
	y := 800.0
	leftMargin := 54.0
	rightMargin := 54.0
	pageWidth := 595.0
	usableWidth := pageWidth - leftMargin - rightMargin
	for _, line := range p.Lines {
		if line.Text == "" {
			y -= lineHeightForPDF(line.Size) * 0.7
			continue
		}
		wrapped := wrapPDFText(line.Text, line.Size, usableWidth)
		for _, item := range wrapped {
			if y < 60 {
				break
			}
			x := leftMargin
			if line.Center {
				width := estimatePDFTextWidth(item, line.Size)
				if width < usableWidth {
					x = leftMargin + (usableWidth-width)/2
				}
			}
			buf.WriteString("BT ")
			buf.WriteString(fmt.Sprintf("/F1 %.1f Tf ", line.sizeOrDefault()))
			buf.WriteString(fmt.Sprintf("1 0 0 1 %.2f %.2f Tm ", x, y))
			buf.WriteString("<")
			buf.WriteString(encodeUTF16BEHex(item))
			buf.WriteString("> Tj ET\n")
			y -= lineHeightForPDF(line.Size)
		}
		y -= lineHeightForPDF(line.Size) * 0.25
	}
	return []byte(buf.String())
}

func (l pdfLine) sizeOrDefault() float64 {
	if l.Size <= 0 {
		return 18
	}
	return l.Size
}

func buildPDFPages(paragraphs []docxParagraph) []pdfPage {
	const (
		pageWidth    = 595.0
		leftMargin   = 54.0
		rightMargin  = 54.0
		topMargin    = 42.0
		bottomMargin = 60.0
	)
	usableWidth := pageWidth - leftMargin - rightMargin
	pages := make([]pdfPage, 0, 2)
	current := pdfPage{Lines: []pdfLine{}}
	currentY := 800.0
	appendPage := func() {
		if len(current.Lines) > 0 {
			pages = append(pages, current)
		}
		current = pdfPage{Lines: []pdfLine{}}
		currentY = 800.0
	}

	_ = topMargin
	for _, paragraph := range paragraphs {
		if strings.TrimSpace(paragraph.Text) == "" {
			current.Lines = append(current.Lines, pdfLine{Text: "", Size: float64(paragraph.Size)})
			currentY -= lineHeightForPDF(float64(paragraph.Size)) * 0.5
			continue
		}
		size := float64(paragraph.Size)
		if size <= 0 {
			size = 18
		}
		wrapped := wrapPDFText(paragraph.Text, size, usableWidth)
		if len(wrapped) == 0 {
			wrapped = []string{paragraph.Text}
		}
		for _, line := range wrapped {
			if currentY < bottomMargin+lineHeightForPDF(size) {
				appendPage()
			}
			current.Lines = append(current.Lines, pdfLine{Text: line, Size: size, Center: paragraph.Center})
			currentY -= lineHeightForPDF(size)
		}
		currentY -= lineHeightForPDF(size) * 0.25
	}
	if len(current.Lines) > 0 {
		pages = append(pages, current)
	}
	return pages
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
		return 0.35
	case r < 128:
		return 0.55
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

func buildDocxDocumentXML(paragraphs []docxParagraph) string {
	var buf strings.Builder
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	buf.WriteString(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">`)
	buf.WriteString(`<w:body>`)
	for _, paragraph := range paragraphs {
		buf.WriteString(renderDocxParagraph(paragraph))
	}
	buf.WriteString(`<w:sectPr><w:pgSz w:w="11906" w:h="16838"/><w:pgMar w:top="1440" w:right="1440" w:bottom="1440" w:left="1440" w:header="708" w:footer="708" w:gutter="0"/></w:sectPr>`)
	buf.WriteString(`</w:body></w:document>`)
	return buf.String()
}

func renderDocxParagraph(p docxParagraph) string {
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
	var buf strings.Builder
	buf.WriteString(`<w:p>`)
	buf.WriteString(`<w:pPr>`)
	if p.Center {
		buf.WriteString(`<w:jc w:val="center"/>`)
	}
	buf.WriteString(fmt.Sprintf(`<w:spacing w:after="%d"/>`, spacing))
	buf.WriteString(`</w:pPr>`)
	buf.WriteString(`<w:r><w:rPr>`)
	buf.WriteString(`<w:rFonts w:ascii="Microsoft YaHei" w:eastAsia="Microsoft YaHei" w:hAnsi="Microsoft YaHei"/>`)
	if p.Bold {
		buf.WriteString(`<w:b/>`)
	}
	buf.WriteString(fmt.Sprintf(`<w:sz w:val="%d"/>`, size))
	buf.WriteString(`</w:rPr>`)
	buf.WriteString(`<w:t xml:space="preserve">`)
	buf.WriteString(html.EscapeString(strings.TrimSpace(p.Text)))
	buf.WriteString(`</w:t></w:r></w:p>`)
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
			paragraphs = append(paragraphs, docxParagraph{Text: strings.TrimSpace(strings.TrimPrefix(trimmed, "### ")), Bold: true, Size: 22, Spacing: 100})
		case strings.HasPrefix(trimmed, "## "):
			paragraphs = append(paragraphs, docxParagraph{Text: strings.TrimSpace(strings.TrimPrefix(trimmed, "## ")), Bold: true, Size: 24, Spacing: 120})
		case strings.HasPrefix(trimmed, "# "):
			paragraphs = append(paragraphs, docxParagraph{Text: strings.TrimSpace(strings.TrimPrefix(trimmed, "# ")), Bold: true, Size: 26, Spacing: 140})
		case strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* "):
			paragraphs = append(paragraphs, docxParagraph{Text: "• " + strings.TrimSpace(trimmed[2:]), Size: 21, Spacing: 60})
		default:
			paragraphs = append(paragraphs, docxParagraph{Text: strings.Trim(trimmed, "*_"), Size: 21, Spacing: 80})
		}
	}
	return paragraphs
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
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`

const docxRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`

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

	// AI 生成报告（如果有）
	if analysis.AIReport != "" {
		buf.WriteString("---\n\n")
		buf.WriteString("## 附：AI 生成报告参考\n\n")
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
