package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// ========== AI 配置 ==========
// 使用 OpenAI 兼容接口生成视频分析报告

// AIConfig LLM API 配置（OpenAI 兼容格式）
type AIConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

// DefaultAIConfig 默认配置，敏感密钥必须来自环境变量
var DefaultAIConfig = AIConfig{
	BaseURL: "https://api.anhepro.com/v1",
	Model:   "gpt-5.5",
}

// ProfessionalScoutReportSystemPrompt is the fixed system prompt for youth scouting reports.
const ProfessionalScoutReportSystemPrompt = `你是“少年球探”平台的青少年足球球探报告专家，具备长期青训观察、比赛视频分析和球员发展规划经验。
你的任务是基于用户提供的结构化数据生成专业、客观、可执行的青少年足球视频分析球探报告。

写作原则：
1. 只依据已提供的球员档案、比赛信息、分析师评分、分析师文字评价和关键片段作判断，不编造未提供的事实。
2. 分析师的评分和文字判断优先级最高，AI 只能进行结构化表达、扩写、归纳和建议，不得推翻分析师结论。
3. 报告面向青训球员、家长、教练和俱乐部，语言专业但易懂，语气客观、鼓励、不过度承诺。
4. 每个重要结论都要能对应到评分、评语、关键片段或球员档案；证据不足时必须说明“数据有限”。
5. 不做医学诊断，不承诺职业化结果，不输出联系方式、家庭隐私或与足球评估无关的信息。
6. 输出使用 Markdown，多级标题清晰，不要输出 JSON、代码块或解释提示词本身。`

// VideoAnalysisReportTemplateVersion 标记视频分析报告模板版本，便于输入快照追踪。
const VideoAnalysisReportTemplateVersion = "video-analysis-report-v1.3-2026-05-10"

// ========== OpenAI兼容请求/响应结构 ==========

// ChatRequest 聊天补全请求（OpenAI兼容格式）
type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
}

// ChatMessage 聊天消息
type ChatMessage struct {
	Role             string `json:"role"`
	Content          string `json:"content"`
	ReasoningContent string `json:"reasoning_content,omitempty"`
}

// ChatResponse 聊天补全响应（OpenAI兼容格式）
type ChatResponse struct {
	ID      string   `json:"id"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice 选择
type Choice struct {
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// Usage 使用量
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// AIService AI服务
type AIService struct {
	config AIConfig
	client *http.Client
}

// NewAIService 创建AI服务
func NewAIService(config AIConfig) *AIService {
	if config.APIKey == "" {
		config.APIKey = os.Getenv("AI_API_KEY")
	}
	if config.BaseURL == "" {
		config.BaseURL = os.Getenv("AI_BASE_URL")
		if config.BaseURL == "" {
			config.BaseURL = DefaultAIConfig.BaseURL
		}
	}
	if config.Model == "" {
		config.Model = os.Getenv("AI_MODEL")
		if config.Model == "" {
			config.Model = DefaultAIConfig.Model
		}
	}
	return &AIService{
		config: config,
		client: &http.Client{
			Timeout: 180 * time.Second,
		},
	}
}

// IsConfigured reports whether the service can make external model calls.
func (s *AIService) IsConfigured() bool {
	return s != nil && strings.TrimSpace(s.config.APIKey) != ""
}

// GenerateReport 生成视频分析报告
func (s *AIService) GenerateReport(prompt string) (string, error) {
	messages := []ChatMessage{
		{Role: "system", Content: ProfessionalScoutReportSystemPrompt},
		{Role: "user", Content: prompt},
	}

	return s.chat(messages)
}

// chat 调用 LLM API（OpenAI兼容格式）
func (s *AIService) chat(messages []ChatMessage) (string, error) {
	if s.config.APIKey == "" {
		return "", fmt.Errorf("AI_API_KEY 未配置")
	}

	reqBody := ChatRequest{
		Model:       s.config.Model,
		Messages:    messages,
		MaxTokens:   8192,
		Temperature: 0.7,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("请求序列化失败: %v", err)
	}

	baseURL := strings.TrimRight(s.config.BaseURL, "/")
	url := baseURL + "/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)

	log.Printf("[AIService] sending request to %s, model=%s, messages=%d", url, reqBody.Model, len(reqBody.Messages))
	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	log.Printf("[AIService] response status=%s, bodyLen=%d", resp.Status, len(body))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("AI API错误: %s - %s", resp.Status, string(body))
	}

	var result ChatResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("响应解析失败: %v", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("未获取到有效回复")
	}

	msg := result.Choices[0].Message
	content := msg.Content
	if content == "" && msg.ReasoningContent != "" {
		content = msg.ReasoningContent
	}
	if content == "" {
		return "", fmt.Errorf("AI 返回内容为空")
	}

	return content, nil
}

// BuildReportPrompt 构建报告生成提示词
func BuildReportPrompt(analysis *VideoAnalysisReportInput) string {
	var sb strings.Builder

	if analysis == nil {
		return ""
	}

	sb.WriteString("【任务】\n")
	sb.WriteString("请基于以下结构化输入，生成一份“少年球探青少年足球视频分析球探报告”。报告应贴合青训球员评估场景，覆盖技术特点、比赛表现、优缺点、潜力判断和发展建议。\n\n")

	sb.WriteString("【数据使用优先级】\n")
	sb.WriteString("1. 分析师评分、单项评语、综合评价、优势、待加强点、改进建议为最高优先级依据。\n")
	sb.WriteString("2. 关键片段是报告结论的重要证据：精彩表现写入优势与关键片段分析；待改进问题写入问题诊断和训练建议；战术观察用于补充比赛理解。\n")
	sb.WriteString("3. 球员档案仅作为背景信息，不得用未提供的数据推断身体、心理或职业前景。\n")
	sb.WriteString("4. 对球员能力的判断应结合其年龄段和同龄球员常见发展水平进行客观表述；没有平台同龄样本数据时，只能使用“相对同龄青训球员的常见表现”这类定性表达，不得编造排名、百分位或权威结论。\n")
	sb.WriteString("5. 若某项数据缺失，请写“暂无记录”或在数据边界中说明，不要编造。\n\n")

	sb.WriteString("【重点参考信息】\n")
	sb.WriteString("请优先整合以下四类信息，再进行扩写与判断：\n")
	sb.WriteString("1. 球员基础信息：年龄、位置、惯用脚、身高体重、球队/学校、训练背景、体测与画像标签。\n")
	sb.WriteString("2. 评分信息：20项评分的分值、权重与单项评语，不要机械逐条罗列，要归纳成技术、战术、进攻、防守、身体/心理几个维度。\n")
	sb.WriteString("3. 评语信息：综合评价、优势、待提升点、训练建议、分析师备注。\n")
	sb.WriteString("4. 比赛信息：对手、比赛类型、出场时间、进球助攻、潜力等级、关键片段与时间点。\n")
	sb.WriteString("5. 同龄参考：基于球员年龄、位置和本场评分，客观分析其相对同龄球员常见水平的优势、短板与成长空间；只做定性判断，不输出虚构的样本规模或排名。\n")
	sb.WriteString("6. 身体条件参考：如果提供了身高、体重，请结合其年龄段青少年常见身高体重水平做谨慎的定性比较，并与本场对抗、速度、护球、争顶等表现交叉验证后，再判断是否具备身体优势；不得编造医学结论、成长预测或精确百分位。\n\n")

	// 球员基本信息
	sb.WriteString("【球员基本信息】\n")
	writePromptField(&sb, "姓名", analysis.PlayerName)
	if analysis.PlayerAge > 0 {
		sb.WriteString(fmt.Sprintf("- 年龄：%d岁\n", analysis.PlayerAge))
	}
	if analysis.PlayerPosition != "" {
		sb.WriteString(fmt.Sprintf("- 位置：%s\n", analysis.PlayerPosition))
	}
	if analysis.PlayerFoot != "" {
		sb.WriteString(fmt.Sprintf("- 惯用脚：%s\n", analysis.PlayerFoot))
	}
	if analysis.PlayerHeight > 0 {
		sb.WriteString(fmt.Sprintf("- 身高：%.0fcm\n", analysis.PlayerHeight))
	}
	if analysis.PlayerWeight > 0 {
		sb.WriteString(fmt.Sprintf("- 体重：%.0fkg\n", analysis.PlayerWeight))
	}
	if analysis.PlayerHeight > 0 || analysis.PlayerWeight > 0 {
		sb.WriteString("- 身体条件分析要求：结合年龄段常见身高体重水平和本场身体对抗表现，判断身体条件是否形成比赛优势或限制。\n")
	}
	if analysis.PlayerTeam != "" {
		sb.WriteString(fmt.Sprintf("- 所属球队：%s\n", analysis.PlayerTeam))
	}
	for _, fact := range analysis.PlayerProfileFacts {
		writePromptField(&sb, fact.Label, fact.Value)
	}
	if len(analysis.PhysicalTestFacts) > 0 {
		sb.WriteString("\n【体测与身体素质数据】\n")
		for _, fact := range analysis.PhysicalTestFacts {
			writePromptField(&sb, fact.Label, fact.Value)
		}
	}
	sb.WriteString("\n")

	// 比赛信息
	sb.WriteString("【比赛信息】\n")
	if analysis.MatchName != "" {
		sb.WriteString(fmt.Sprintf("- 比赛名称：%s\n", analysis.MatchName))
	}
	if analysis.MatchDate != "" {
		sb.WriteString(fmt.Sprintf("- 比赛日期：%s\n", analysis.MatchDate))
	}
	if analysis.MatchType != "" {
		sb.WriteString(fmt.Sprintf("- 比赛性质：%s\n", analysis.MatchType))
	}
	if analysis.OpponentLevel != "" {
		sb.WriteString(fmt.Sprintf("- 对手实力：%s\n", analysis.OpponentLevel))
	}
	if analysis.Opponent != "" {
		sb.WriteString(fmt.Sprintf("- 对手：%s\n", analysis.Opponent))
	}
	if analysis.PlayTime > 0 {
		sb.WriteString(fmt.Sprintf("- 出场时间：%d分钟\n", analysis.PlayTime))
	}
	if analysis.Goals > 0 {
		sb.WriteString(fmt.Sprintf("- 进球：%d个\n", analysis.Goals))
	}
	if analysis.Assists > 0 {
		sb.WriteString(fmt.Sprintf("- 助攻：%d次\n", analysis.Assists))
	}
	sb.WriteString("\n")

	// 综合评分
	sb.WriteString("【分析师评分】\n")
	sb.WriteString(fmt.Sprintf("- 综合评分：%.1f分\n", analysis.OverallScore))
	if analysis.PotentialLevel != "" {
		sb.WriteString(fmt.Sprintf("- 潜力等级：%s级\n", analysis.PotentialLevel))
	}
	sb.WriteString("\n")

	// 各项评分 - 20维度
	sb.WriteString("【详细评分与评价】\n")
	dimensions := []struct {
		Name     string
		Score    float64
		Weight   float64
		Comment  string
		Category string
	}{
		// 整体表现 (4项)
		{"控球能力", analysis.Scores.BallControl.Score, analysis.Scores.BallControl.Weight * 100, analysis.Scores.BallControl.Comment, "整体表现"},
		{"无球跑动", analysis.Scores.OffBallMovement.Score, analysis.Scores.OffBallMovement.Weight * 100, analysis.Scores.OffBallMovement.Comment, "整体表现"},
		{"逼抢意识", analysis.Scores.PressingAwareness.Score, analysis.Scores.PressingAwareness.Weight * 100, analysis.Scores.PressingAwareness.Comment, "整体表现"},
		{"站位选位", analysis.Scores.Positioning.Score, analysis.Scores.Positioning.Weight * 100, analysis.Scores.Positioning.Comment, "整体表现"},
		// 进攻能力 (8项)
		{"拉开宽度参与", analysis.Scores.WidthParticipation.Score, analysis.Scores.WidthParticipation.Weight * 100, analysis.Scores.WidthParticipation.Comment, "进攻能力"},
		{"无球支援", analysis.Scores.OffBallSupport.Score, analysis.Scores.OffBallSupport.Weight * 100, analysis.Scores.OffBallSupport.Comment, "进攻能力"},
		{"1v1过人能力", analysis.Scores.OneVOne.Score, analysis.Scores.OneVOne.Weight * 100, analysis.Scores.OneVOne.Comment, "进攻能力"},
		{"传中/助攻", analysis.Scores.CrossingAssist.Score, analysis.Scores.CrossingAssist.Weight * 100, analysis.Scores.CrossingAssist.Comment, "进攻能力"},
		{"对抗能力", analysis.Scores.CombatAbility.Score, analysis.Scores.CombatAbility.Weight * 100, analysis.Scores.CombatAbility.Comment, "进攻能力"},
		{"节奏把控", analysis.Scores.PaceRhythm.Score, analysis.Scores.PaceRhythm.Weight * 100, analysis.Scores.PaceRhythm.Comment, "进攻能力"},
		{"传球视野", analysis.Scores.PassVision.Score, analysis.Scores.PassVision.Weight * 100, analysis.Scores.PassVision.Comment, "进攻能力"},
		{"身体姿态", analysis.Scores.BodyPosture.Score, analysis.Scores.BodyPosture.Weight * 100, analysis.Scores.BodyPosture.Comment, "进攻能力"},
		// 防守能力 (8项)
		{"防守投入度", analysis.Scores.DefensiveCommitment.Score, analysis.Scores.DefensiveCommitment.Weight * 100, analysis.Scores.DefensiveCommitment.Comment, "防守能力"},
		{"丢球回追", analysis.Scores.LossRecovery.Score, analysis.Scores.LossRecovery.Weight * 100, analysis.Scores.LossRecovery.Comment, "防守能力"},
		{"队友协防配合", analysis.Scores.TeammateCoordination.Score, analysis.Scores.TeammateCoordination.Weight * 100, analysis.Scores.TeammateCoordination.Comment, "防守能力"},
		{"二点球争抢", analysis.Scores.SecondBall.Score, analysis.Scores.SecondBall.Weight * 100, analysis.Scores.SecondBall.Comment, "防守能力"},
		{"空中争顶", analysis.Scores.AerialDuel.Score, analysis.Scores.AerialDuel.Weight * 100, analysis.Scores.AerialDuel.Comment, "防守能力"},
		{"防守阵型保持", analysis.Scores.DefensiveShape.Score, analysis.Scores.DefensiveShape.Weight * 100, analysis.Scores.DefensiveShape.Comment, "防守能力"},
		{"角色调整能力", analysis.Scores.RoleAdjustment.Score, analysis.Scores.RoleAdjustment.Weight * 100, analysis.Scores.RoleAdjustment.Comment, "防守能力"},
		{"防守节奏", analysis.Scores.DefensiveRhythm.Score, analysis.Scores.DefensiveRhythm.Weight * 100, analysis.Scores.DefensiveRhythm.Comment, "防守能力"},
	}

	for _, dim := range dimensions {
		sb.WriteString(fmt.Sprintf("\n%s (%.1f分 × %.0f%%权重) [%s]\n", dim.Name, dim.Score, dim.Weight, dim.Category))
		if dim.Comment != "" {
			sb.WriteString(fmt.Sprintf("评价：%s\n", dim.Comment))
		}
	}
	sb.WriteString("\n")

	if analysis.Strengths != "" {
		sb.WriteString("【分析师记录的核心优势】\n")
		sb.WriteString(normalizePromptText(analysis.Strengths) + "\n\n")
	}

	if analysis.Weaknesses != "" {
		sb.WriteString("【分析师记录的待加强点】\n")
		sb.WriteString(normalizePromptText(analysis.Weaknesses) + "\n\n")
	}

	// 高光时刻
	if len(analysis.Highlights) > 0 {
		sb.WriteString("【关键片段标记】\n")
		for i, h := range analysis.Highlights {
			timeText := h.Timestamp
			if h.Mode == "range" && h.StartTime != "" && h.EndTime != "" {
				timeText = h.StartTime + "-" + h.EndTime
			}
			sb.WriteString(fmt.Sprintf("%d. [%s][%s][%s] %s\n", i+1, timeText, markerTypeLabel(h.MarkerType), tagTypeLabel(h.TagType), h.Description))
		}
		sb.WriteString("\n")
	}

	// 分析师补充
	if analysis.AnalystNotes != "" {
		sb.WriteString("【分析师补充说明】\n")
		sb.WriteString(analysis.AnalystNotes + "\n\n")
	}

	// 综合评价摘要
	if analysis.Summary != "" {
		sb.WriteString("【综合评价摘要】\n")
		sb.WriteString(analysis.Summary + "\n\n")
	}

	// 改进建议
	if analysis.Improvements != "" {
		sb.WriteString("【重点改进建议】\n")
		sb.WriteString(analysis.Improvements + "\n\n")
	}

	// 报告要求
	sb.WriteString("【固定输出结构】\n")
	sb.WriteString("请严格按以下 Markdown 章节输出，章节标题不得缺失：\n")
	sb.WriteString("# 青少年足球视频分析球探报告\n")
	sb.WriteString("## 1. 报告摘要\n")
	sb.WriteString("用 3-5 段概括球员当前画像、主要优势、主要问题、综合评分含义和后续发展方向。\n")
	sb.WriteString("## 2. 球员基础画像\n")
	sb.WriteString("结合年龄、位置、身体数据、踢球风格、球队/学校等信息说明评估背景；缺失信息不要编造。\n")
	sb.WriteString("## 3. 本场比赛背景\n")
	sb.WriteString("说明比赛、对手、出场时间、进球助攻等背景，并解释这些背景对评价的影响。\n")
	sb.WriteString("## 4. 综合评分与潜力解读\n")
	sb.WriteString("解释综合评分和潜力等级，结合球员年龄段说明其相对同龄球员常见水平的含义、优势和限制。\n")
	sb.WriteString("## 5. 技术能力分析\n")
	sb.WriteString("围绕控球、传球视野、1v1、传中/助攻、节奏和身体姿态做专业分析。\n")
	sb.WriteString("## 6. 战术与无球表现\n")
	sb.WriteString("围绕无球跑动、站位选位、无球支援、角色调整和比赛阅读进行分析。\n")
	sb.WriteString("## 7. 进攻表现\n")
	sb.WriteString("解释进攻维度高低分的原因，并结合关键片段或评语说明。\n")
	sb.WriteString("## 8. 防守表现\n")
	sb.WriteString("解释防守投入、回追、协防、阵型保持、二点球和空中争顶表现。\n")
	sb.WriteString("## 9. 身体与心理特点\n")
	sb.WriteString("只基于已提供身体数据、对抗表现、心智标签和比赛行为分析，不做医学判断。\n")
	sb.WriteString("## 10. 关键片段分析\n")
	sb.WriteString("按时间点/时间段引用关键片段，分别说明精彩表现、待改进问题和战术观察。\n")
	sb.WriteString("## 11. 核心优势\n")
	sb.WriteString("列出 3-5 条优势，每条包含证据和对未来比赛的价值。\n")
	sb.WriteString("## 12. 待提升问题\n")
	sb.WriteString("列出 3-5 条问题，每条说明表现、原因和可能影响，避免否定式评价。\n")
	sb.WriteString("## 13. 4 周训练建议\n")
	sb.WriteString("给出具体、可执行、可追踪的训练计划，至少包含技术、战术、身体或心理中的三类。\n")
	sb.WriteString("## 14. 给家长和教练的建议\n")
	sb.WriteString("分别给家长和教练写建议，强调长期成长、比赛复盘和训练反馈方式。\n")
	sb.WriteString("## 15. 数据边界说明\n")
	sb.WriteString("说明本报告基于单场/当前视频、分析师评分和已有档案生成；同龄水平判断为基于年龄段和青训常见表现的定性分析，不代表平台排名或权威分级。\n\n")
	sb.WriteString("【写作约束】\n")
	sb.WriteString("1. 总字数建议不少于 5000 字；如果输入信息较少，也要尽可能扩写到 4000 字以上，并优先保证结构完整和建议可执行。\n")
	sb.WriteString("2. 不要逐项机械复述 20 个分数；应归纳为技术、战术、进攻、防守、身体心理几个维度。\n")
	sb.WriteString("3. 低分项必须给出具体训练方法；高分项必须说明可迁移到比赛中的价值。\n")
	sb.WriteString("4. 不使用“必进职业队”“天才确定无疑”等绝对化表达。\n")
	sb.WriteString("5. 不要输出手机号、微信、家庭联系方式等隐私信息。\n")
	sb.WriteString("6. 涉及同龄对比时，必须使用客观、谨慎表述，例如“在同龄球员中属于较突出/稳定/仍需积累的表现”，不得虚构数据库、排名或百分位。\n\n")
	sb.WriteString("请开始生成正式报告正文：\n")

	return sb.String()
}

func writePromptField(sb *strings.Builder, label string, value string) {
	label = strings.TrimSpace(label)
	value = normalizePromptText(value)
	if label == "" || value == "" {
		return
	}
	sb.WriteString(fmt.Sprintf("- %s：%s\n", label, value))
}

func normalizePromptText(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	var stringItems []string
	if err := json.Unmarshal([]byte(trimmed), &stringItems); err == nil && len(stringItems) > 0 {
		cleanItems := make([]string, 0, len(stringItems))
		for _, item := range stringItems {
			item = strings.TrimSpace(item)
			if item != "" {
				cleanItems = append(cleanItems, item)
			}
		}
		if len(cleanItems) > 0 {
			return strings.Join(cleanItems, "、")
		}
	}

	return trimmed
}

// VideoAnalysisReportInput 报告生成输入
type VideoAnalysisReportInput struct {
	PlayerName     string
	PlayerAge      int
	PlayerPosition string
	PlayerFoot     string
	PlayerHeight   float64
	PlayerWeight   float64
	PlayerTeam     string
	// Whitelisted profile facts from user profile. Sensitive contact/family fields must not be included.
	PlayerProfileFacts []ReportFactInput
	PhysicalTestFacts  []ReportFactInput

	MatchName     string
	MatchDate     string
	MatchType     string
	OpponentLevel string
	Opponent      string
	PlayTime      int
	Goals         int
	Assists       int

	OverallScore   float64
	PotentialLevel string
	Scores         ScoresInput

	Highlights   []HighlightInput
	Summary      string
	Strengths    string
	Weaknesses   string
	Improvements string
	AnalystNotes string
}

// ReportFactInput is a label-value fact included in the AI prompt.
type ReportFactInput struct {
	Label string
	Value string
}

// ScoresInput 评分输入（20项对齐前端）
type ScoresInput struct {
	// 整体表现
	BallControl       ScoreInput `json:"ball_control"`
	OffBallMovement   ScoreInput `json:"off_ball_movement"`
	PressingAwareness ScoreInput `json:"pressing_awareness"`
	Positioning       ScoreInput `json:"positioning"`
	// 进攻能力
	WidthParticipation ScoreInput `json:"width_participation"`
	OffBallSupport     ScoreInput `json:"off_ball_support"`
	OneVOne            ScoreInput `json:"one_v_one"`
	CrossingAssist     ScoreInput `json:"crossing_assist"`
	CombatAbility      ScoreInput `json:"combat_ability"`
	PaceRhythm         ScoreInput `json:"pace_rhythm"`
	PassVision         ScoreInput `json:"pass_vision"`
	BodyPosture        ScoreInput `json:"body_posture"`
	// 防守能力
	DefensiveCommitment  ScoreInput `json:"defensive_commitment"`
	LossRecovery         ScoreInput `json:"loss_recovery"`
	TeammateCoordination ScoreInput `json:"teammate_coordination"`
	SecondBall           ScoreInput `json:"second_ball"`
	AerialDuel           ScoreInput `json:"aerial_duel"`
	DefensiveShape       ScoreInput `json:"defensive_shape"`
	RoleAdjustment       ScoreInput `json:"role_adjustment"`
	DefensiveRhythm      ScoreInput `json:"defensive_rhythm"`
}

// ScoreInput 单项评分输入
type ScoreInput struct {
	Score   float64 `json:"score"`
	Weight  float64 `json:"weight"`
	Comment string  `json:"comment"`
}

// HighlightInput 高光时刻输入
type HighlightInput struct {
	Timestamp   string `json:"timestamp"`
	MarkerType  string `json:"marker_type"`
	Mode        string `json:"mode"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	TagType     string `json:"tag_type"`
	Description string `json:"description"`
}

func markerTypeLabel(markerType string) string {
	switch markerType {
	case "issue":
		return "待改进问题"
	case "observation":
		return "战术观察"
	default:
		return "精彩表现"
	}
}

func tagTypeLabel(tagType string) string {
	labels := map[string]string{
		"goal":              "进球",
		"assist":            "助攻",
		"steal":             "抢断",
		"save":              "扑救",
		"dribble":           "过人",
		"pass":              "关键传球",
		"defense":           "防守关键",
		"positioning_error": "站位问题",
		"decision_error":    "决策问题",
		"turnover":          "失误",
		"recovery_slow":     "回防不及时",
		"tactical_note":     "战术观察",
		"off_ball_run":      "无球跑动",
	}
	if label, ok := labels[tagType]; ok {
		return label
	}
	return tagType
}
