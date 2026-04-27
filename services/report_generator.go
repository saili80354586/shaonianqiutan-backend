package services

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// buildVideoAnalysisReport 构建视频分析评分报告内容
func (g *ReportGenerator) buildVideoAnalysisReport(analysis *models.VideoAnalysis, analystName string) string {
	var buf bytes.Buffer

	// 解析 scores JSON
	var scores map[string]map[string]map[string]interface{}
	_ = json.Unmarshal([]byte(analysis.Scores), &scores)

	// 评分标签映射（snake_case key -> 中文标签）
	overallLabels := map[string]string{
		"ball_control":      "控球能力",
		"off_ball_movement": "无球跑动",
		"pressing_awareness": "逼抢意识",
		"positioning":        "站位选择",
	}
	offenseLabels := map[string]string{
		"width_participation":  "拉开宽度参与",
		"off_ball_support":    "无球支援",
		"one_v_one":           "1v1过人能力",
		"crossing_assist":      "传中/助攻",
		"combat_ability":       "对抗能力",
		"pace_rhythm":         "节奏把控",
		"pass_vision":         "传球视野",
		"body_posture":        "身体姿态",
	}
	defenseLabels := map[string]string{
		"defensive_commitment":    "防守投入度",
		"loss_recovery":           "丢球回追",
		"teammate_coordination":    "队友协防配合",
		"second_ball":             "二点球争抢",
		"aerial_duel":            "空中争顶",
		"defensive_shape":        "防守阵型保持",
		"role_adjustment":          "角色调整能力",
		"defensive_rhythm":        "防守节奏",
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
