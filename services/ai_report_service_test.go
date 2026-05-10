package services

import (
	"strings"
	"testing"
)

func TestDefaultAIConfigUsesAnheproGPT55(t *testing.T) {
	if DefaultAIConfig.BaseURL != "https://api.anhepro.com/v1" {
		t.Fatalf("default AI base url = %q", DefaultAIConfig.BaseURL)
	}
	if DefaultAIConfig.Model != "gpt-5.5" {
		t.Fatalf("default AI model = %q", DefaultAIConfig.Model)
	}
}

func TestBuildReportPromptIncludesProfessionalScoutTemplateAndInputs(t *testing.T) {
	prompt := BuildReportPrompt(&VideoAnalysisReportInput{
		PlayerName:     "林子昂",
		PlayerAge:      13,
		PlayerPosition: "边锋",
		PlayerFoot:     "右脚",
		PlayerProfileFacts: []ReportFactInput{
			{Label: "技术标签", Value: `["盘带","速度"]`},
			{Label: "足球经历", Value: "2024 / 校队 / 边锋 / 市级比赛四强"},
		},
		PhysicalTestFacts: []ReportFactInput{
			{Label: "30米冲刺", Value: "4.8秒"},
		},
		MatchName:      "U13联赛",
		Opponent:       "蓝鹰U13",
		OverallScore:   82.5,
		PotentialLevel: "A",
		Scores: ScoresInput{
			BallControl:         ScoreInput{Score: 8.2, Weight: 0.05, Comment: "一停一带衔接稳定。"},
			OneVOne:             ScoreInput{Score: 8.5, Weight: 0.05, Comment: "边路一对一敢于主动突破。"},
			DefensiveCommitment: ScoreInput{Score: 6.8, Weight: 0.05, Comment: "丢球后第一反应略慢。"},
		},
		Highlights: []HighlightInput{
			{Timestamp: "12:20", MarkerType: "highlight", TagType: "dribble", Description: "边路连续突破后形成传中。"},
			{Mode: "range", StartTime: "20:10", EndTime: "20:28", MarkerType: "issue", TagType: "positioning_error", Description: "回防线路过直，未保护中路空间。"},
		},
		Summary:      "具备边路突破能力，但防守转换还需加强。",
		Strengths:    "启动速度快；一对一处理积极",
		Weaknesses:   "回防选位需要提升",
		Improvements: "加强防守转换和弱侧观察。",
		AnalystNotes: "本场对手压迫强度较高。",
	})

	mustContain := []string{
		"少年球探青少年足球视频分析球探报告",
		"【数据使用优先级】",
		"【重点参考信息】",
		"总字数建议不少于 5000 字",
		"相对同龄球员常见水平",
		"不得虚构数据库、排名或百分位",
		"技术标签：盘带、速度",
		"30米冲刺：4.8秒",
		"【分析师记录的核心优势】",
		"启动速度快",
		"【分析师记录的待加强点】",
		"回防选位需要提升",
		"[12:20][精彩表现][过人]",
		"[20:10-20:28][待改进问题][站位问题]",
		"## 11. 核心优势",
		"## 12. 待提升问题",
		"## 13. 4 周训练建议",
		"## 15. 数据边界说明",
		"不要输出手机号、微信、家庭联系方式等隐私信息",
	}

	for _, expected := range mustContain {
		if !strings.Contains(prompt, expected) {
			t.Fatalf("prompt missing %q\nprompt:\n%s", expected, prompt)
		}
	}
}

func TestProfessionalScoutReportSystemPromptDefinesSafetyBoundaries(t *testing.T) {
	for _, expected := range []string{
		"不编造未提供的事实",
		"分析师的评分和文字判断优先级最高",
		"不做医学诊断",
		"不输出联系方式",
	} {
		if !strings.Contains(ProfessionalScoutReportSystemPrompt, expected) {
			t.Fatalf("system prompt missing %q", expected)
		}
	}
}
