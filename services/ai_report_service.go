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
// 使用 Kimi K2.5 模型生成视频分析报告

// AIConfig LLM API 配置（OpenAI 兼容格式）
type AIConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

// DefaultAIConfig 默认配置 - 智谱AI (ZhipuAI)，敏感密钥必须来自环境变量
var DefaultAIConfig = AIConfig{
	BaseURL: "https://open.bigmodel.cn/api/paas/v4/",
	Model:   "glm-4-flash",
}

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

// GenerateReport 生成视频分析报告
func (s *AIService) GenerateReport(prompt string) (string, error) {
	systemPrompt := `你是一位拥有15年青训经验的专业足球球探，曾发掘过多名职业球员，为多家俱乐部提供球探报告。
你的分析报告以专业、客观、鼓励为基调，既要指出不足，更要肯定优点。
报告语言要专业但易懂，适合家长和教练阅读。
字数要求：约5000字，结构完整，分析深入。`

	messages := []ChatMessage{
		{Role: "system", Content: systemPrompt},
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

	sb.WriteString("请根据以下球员信息和分析师评分，生成一份专业的视频分析报告：\n\n")

	// 球员基本信息
	sb.WriteString("【球员基本信息】\n")
	sb.WriteString(fmt.Sprintf("- 姓名：%s\n", analysis.PlayerName))
	sb.WriteString(fmt.Sprintf("- 年龄：%d岁\n", analysis.PlayerAge))
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
	if analysis.PlayerTeam != "" {
		sb.WriteString(fmt.Sprintf("- 所属球队：%s\n", analysis.PlayerTeam))
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
	sb.WriteString(fmt.Sprintf("- 潜力等级：%s级\n", analysis.PotentialLevel))
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
	sb.WriteString("【报告要求】\n")
	sb.WriteString("1. 总字数约5000字\n")
	sb.WriteString("2. 语言专业但易懂，适合家长阅读\n")
	sb.WriteString("3. 每个评分维度要有深度分析，不只是复述分数\n")
	sb.WriteString("4. 结合具体场景举例说明\n")
	sb.WriteString("5. 成长建议要具体可执行\n")
	sb.WriteString("6. 全文语气积极正面，以鼓励为主\n")
	sb.WriteString("7. 结构清晰，使用多级标题\n\n")
	sb.WriteString("请开始生成报告：\n")

	return sb.String()
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
	Improvements string
	AnalystNotes string
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
