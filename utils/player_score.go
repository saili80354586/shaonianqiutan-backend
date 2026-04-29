package utils

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/shaonianqiutan/backend/models"
)

const (
	PlayerScoreFormulaVersionV1 = "map-score-v1"
	PlayerScoreFormulaVersionV2 = "map-score-v2"
	PlayerScoreFormulaVersion   = PlayerScoreFormulaVersionV2
)

// PlayerScoreInput collects the real data sources used by the map score.
type PlayerScoreInput struct {
	User               *models.User
	PhysicalRecord     *models.PhysicalTestRecord
	ScoutReportAverage *float64
	ScoutReportCount   int64
}

type PlayerScoreMetric struct {
	Key            string  `json:"key,omitempty"`
	Label          string  `json:"label"`
	Value          string  `json:"value"`
	Score          float64 `json:"score"`
	Direction      string  `json:"direction,omitempty"`
	Percentile     int     `json:"percentile,omitempty"`
	Benchmark      string  `json:"benchmark,omitempty"`
	BenchmarkGroup string  `json:"benchmarkGroup,omitempty"`
}

type PlayerScoreComponent struct {
	Label       string  `json:"label"`
	Source      string  `json:"source"`
	Score       float64 `json:"score"`
	Weight      float64 `json:"weight"`
	Count       int64   `json:"count,omitempty"`
	Description string  `json:"description,omitempty"`
}

type PlayerScoreResult struct {
	Score               float64                `json:"score"`
	Potential           string                 `json:"potential"`
	HasScore            bool                   `json:"hasScore"`
	DataCoverage        int                    `json:"dataCoverage"`
	ExpectedMetricCount int                    `json:"expectedMetricCount,omitempty"`
	MetricCoverage      float64                `json:"metricCoverage,omitempty"`
	Confidence          float64                `json:"confidence,omitempty"`
	BenchmarkGroup      string                 `json:"benchmarkGroup,omitempty"`
	FormulaVersion      string                 `json:"formulaVersion"`
	Sources             []string               `json:"sources"`
	Components          []PlayerScoreComponent `json:"components"`
	Metrics             []PlayerScoreMetric    `json:"metrics"`
	MissingMetrics      []string               `json:"missingMetrics,omitempty"`
	Notes               []string               `json:"notes,omitempty"`
}

// CalculatePlayerMapScore returns the default deterministic, explainable player map score.
func CalculatePlayerMapScore(input PlayerScoreInput) PlayerScoreResult {
	return CalculatePlayerMapScoreV2(input, nil)
}

// CalculatePlayerMapScoreV1 returns the original deterministic map score.
// It never invents a display score when there is no physical test or scout report data.
func CalculatePlayerMapScoreV1(input PlayerScoreInput) PlayerScoreResult {
	metrics := collectPhysicalScoreMetrics(input)
	components := make([]PlayerScoreComponent, 0, 2)

	if len(metrics) > 0 {
		components = append(components, PlayerScoreComponent{
			Label:  "体测综合",
			Source: "latest_physical_test",
			Score:  averageMetricScore(metrics),
			Weight: 0.7,
			Count:  int64(len(metrics)),
		})
	}

	if input.ScoutReportAverage != nil && *input.ScoutReportAverage > 0 {
		components = append(components, PlayerScoreComponent{
			Label:  "球探报告均分",
			Source: "published_scout_reports",
			Score:  clampScore(*input.ScoutReportAverage),
			Weight: 0.3,
			Count:  input.ScoutReportCount,
		})
	}

	if len(components) == 0 {
		return PlayerScoreResult{
			Score:          0,
			Potential:      "待评估",
			HasScore:       false,
			DataCoverage:   0,
			FormulaVersion: PlayerScoreFormulaVersionV1,
			Sources:        []string{},
			Components:     []PlayerScoreComponent{},
			Metrics:        []PlayerScoreMetric{},
		}
	}

	totalWeight := 0.0
	total := 0.0
	sources := make([]string, 0, len(components))
	for _, component := range components {
		totalWeight += component.Weight
		total += component.Score * component.Weight
		sources = append(sources, component.Source)
	}
	score := roundScore(total / totalWeight)
	return PlayerScoreResult{
		Score:          score,
		Potential:      PlayerPotentialFromScore(score),
		HasScore:       true,
		DataCoverage:   len(metrics) + int(input.ScoutReportCount),
		FormulaVersion: PlayerScoreFormulaVersionV1,
		Sources:        sources,
		Components:     components,
		Metrics:        metrics,
	}
}

func PlayerPotentialFromScore(score float64) string {
	switch {
	case score >= 90:
		return "S"
	case score >= 85:
		return "A"
	case score >= 75:
		return "B"
	case score >= 65:
		return "C"
	case score > 0:
		return "D"
	default:
		return "待评估"
	}
}

func BuildPlayerTags(user *models.User, limit int) []string {
	if user == nil {
		return []string{}
	}
	tags := make([]string, 0, limit)
	appendTags := func(values []string) {
		for _, value := range values {
			if value == "" || containsString(tags, value) {
				continue
			}
			tags = append(tags, value)
			if limit > 0 && len(tags) >= limit {
				return
			}
		}
	}
	appendTags(parseJSONStrings(user.TechnicalTags))
	if limit <= 0 || len(tags) < limit {
		appendTags(parseJSONStrings(user.MentalTags))
	}
	if user.FARegistered && (limit <= 0 || len(tags) < limit) {
		appendTags([]string{"足协注册"})
	}
	return tags
}

func collectPhysicalScoreMetrics(input PlayerScoreInput) []PlayerScoreMetric {
	metrics := make([]PlayerScoreMetric, 0, 8)
	if input.PhysicalRecord != nil {
		r := input.PhysicalRecord
		addLowerMetric(&metrics, "30m冲刺", r.Sprint30m, 4.2, 6.5, "s")
		addLowerMetric(&metrics, "50m冲刺", r.Sprint50m, 7.2, 10.5, "s")
		addLowerMetric(&metrics, "100m冲刺", r.Sprint100m, 13.0, 19.0, "s")
		addLowerMetric(&metrics, "敏捷梯", r.AgilityLadder, 7.5, 13.0, "s")
		addLowerMetric(&metrics, "T型跑", r.TTest, 8.5, 14.0, "s")
		addLowerMetric(&metrics, "折返跑", r.ShuttleRun, 9.5, 14.5, "s")
		addHigherMetric(&metrics, "立定跳远", r.StandingLongJump, 120, 260, "cm")
		addHigherMetric(&metrics, "纵跳", r.VerticalJump, 20, 70, "cm")
		addHigherMetric(&metrics, "坐位体前屈", r.SitAndReach, -5, 25, "cm")
		addHigherIntMetric(&metrics, "俯卧撑", r.PushUp, 8, 45, "个")
		addHigherIntMetric(&metrics, "仰卧起坐", r.SitUp, 15, 65, "个")
		addHigherIntMetric(&metrics, "平板支撑", r.Plank, 20, 180, "s")
		return metrics
	}

	if input.User == nil {
		return metrics
	}
	user := input.User
	addLowerValueMetric(&metrics, "30m冲刺", user.Sprint30m, 4.2, 6.5, "s")
	addLowerValueMetric(&metrics, "5x25折返跑", user.FiveMeterShuttle, 22, 34, "s")
	addLowerValueMetric(&metrics, "协调性", user.Coordination, 7.5, 13, "s")
	addHigherValueMetric(&metrics, "立定跳远", user.StandingLongJump, 120, 260, "cm")
	addHigherValueMetric(&metrics, "柔韧", firstPositiveFloat(user.SitAndReach, user.Flexibility), -5, 25, "cm")
	addHigherValueMetric(&metrics, "引体向上", float64(user.PullUps), 0, 12, "个")
	addHigherValueMetric(&metrics, "俯卧撑", float64(user.PushUp), 8, 45, "个")
	addHigherValueMetric(&metrics, "仰卧起坐", float64(user.SitUps), 15, 65, "个")
	return metrics
}

func addLowerMetric(metrics *[]PlayerScoreMetric, label string, value *float64, best, worst float64, unit string) {
	if value == nil {
		return
	}
	addLowerValueMetric(metrics, label, *value, best, worst, unit)
}

func addHigherMetric(metrics *[]PlayerScoreMetric, label string, value *float64, worst, best float64, unit string) {
	if value == nil {
		return
	}
	*metrics = append(*metrics, PlayerScoreMetric{
		Label: label,
		Value: formatScoreValue(*value, unit),
		Score: roundScore(((*value - worst) / (best - worst)) * 100),
	})
}

func addHigherIntMetric(metrics *[]PlayerScoreMetric, label string, value *int, worst, best float64, unit string) {
	if value == nil {
		return
	}
	addHigherValueMetric(metrics, label, float64(*value), worst, best, unit)
}

func addLowerValueMetric(metrics *[]PlayerScoreMetric, label string, value, best, worst float64, unit string) {
	if value <= 0 {
		return
	}
	*metrics = append(*metrics, PlayerScoreMetric{
		Label: label,
		Value: formatScoreValue(value, unit),
		Score: roundScore(((worst - value) / (worst - best)) * 100),
	})
}

func addHigherValueMetric(metrics *[]PlayerScoreMetric, label string, value, worst, best float64, unit string) {
	if value <= 0 {
		return
	}
	*metrics = append(*metrics, PlayerScoreMetric{
		Label: label,
		Value: formatScoreValue(value, unit),
		Score: roundScore(((value - worst) / (best - worst)) * 100),
	})
}

func averageMetricScore(metrics []PlayerScoreMetric) float64 {
	if len(metrics) == 0 {
		return 0
	}
	sum := 0.0
	for _, metric := range metrics {
		sum += metric.Score
	}
	return roundScore(sum / float64(len(metrics)))
}

func formatScoreValue(value float64, unit string) string {
	if unit == "个" {
		return fmt.Sprintf("%.0f%s", value, unit)
	}
	return fmt.Sprintf("%.1f%s", value, unit)
}

func firstPositiveFloat(values ...float64) float64 {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

func clampScore(score float64) float64 {
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return score
}

func roundScore(score float64) float64 {
	return math.Round(clampScore(score)*10) / 10
}

func parseJSONStrings(raw string) []string {
	if raw == "" {
		return []string{}
	}
	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return []string{}
	}
	return values
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
