package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/shaonianqiutan/backend/models"
)

const (
	scoreSourcePhysical     = "latest_physical_test"
	scoreSourceScoutReports = "published_scout_reports"
	scoreSourceCompleteness = "data_completeness"
)

type ScoringBenchmarkConfig struct {
	Version string                                                  `json:"version"`
	Groups  map[string]map[string]map[string]ScoringMetricBenchmark `json:"groups"`
}

type ScoringMetricBenchmark struct {
	Direction string  `json:"direction"`
	P10       float64 `json:"p10"`
	P25       float64 `json:"p25"`
	P50       float64 `json:"p50"`
	P75       float64 `json:"p75"`
	P90       float64 `json:"p90"`
}

type physicalMetricSpec struct {
	Key   string
	Label string
	Unit  string
}

type physicalMetricValue struct {
	Value float64
	Label string
	Unit  string
}

type benchmarkPoint struct {
	Value      float64
	Score      float64
	Percentile int
}

var (
	scoringBenchmarksOnce sync.Once
	scoringBenchmarks     *ScoringBenchmarkConfig
)

var physicalMetricSpecs = []physicalMetricSpec{
	{Key: "sprint_30m", Label: "30m冲刺", Unit: "s"},
	{Key: "sprint_50m", Label: "50m冲刺", Unit: "s"},
	{Key: "sprint_100m", Label: "100m冲刺", Unit: "s"},
	{Key: "agility_ladder", Label: "敏捷梯", Unit: "s"},
	{Key: "t_test", Label: "T型跑", Unit: "s"},
	{Key: "shuttle_run", Label: "折返跑", Unit: "s"},
	{Key: "standing_long_jump", Label: "立定跳远", Unit: "cm"},
	{Key: "vertical_jump", Label: "纵跳", Unit: "cm"},
	{Key: "sit_and_reach", Label: "坐位体前屈", Unit: "cm"},
	{Key: "push_up", Label: "俯卧撑", Unit: "个"},
	{Key: "sit_up", Label: "仰卧起坐", Unit: "个"},
	{Key: "plank", Label: "平板支撑", Unit: "s"},
}

// CalculatePlayerMapScoreV2 scores players against age-group and position-aware benchmarks.
func CalculatePlayerMapScoreV2(input PlayerScoreInput, cfg *ScoringBenchmarkConfig) PlayerScoreResult {
	if cfg == nil {
		cfg = LoadScoringBenchmarks()
	}
	if cfg == nil || len(cfg.Groups) == 0 {
		score := CalculatePlayerMapScoreV1(input)
		score.Notes = append(score.Notes, "未加载评分基准配置，已回退到 map-score-v1")
		return score
	}
	ageGroup := ageBenchmarkGroup(input.User)
	positionGroup := positionBenchmarkGroup(input.User)
	benchmarkGroup := fmt.Sprintf("%s/%s", ageGroup, positionGroup)

	metrics, missingMetrics, expectedMetricCount := collectBenchmarkedPhysicalMetrics(input, cfg, ageGroup, positionGroup)
	components := make([]PlayerScoreComponent, 0, 3)
	hasPhysical := len(metrics) > 0
	hasScout := input.ScoutReportAverage != nil && *input.ScoutReportAverage > 0
	metricCoverage := 0.0
	if expectedMetricCount > 0 {
		metricCoverage = roundScore(float64(len(metrics)) / float64(expectedMetricCount) * 100)
	}

	if hasPhysical {
		weight := 0.85
		if hasScout {
			weight = 0.55
		}
		components = append(components, PlayerScoreComponent{
			Label:       "年龄/位置体测分",
			Source:      scoreSourcePhysical,
			Score:       averageMetricScore(metrics),
			Weight:      weight,
			Count:       int64(len(metrics)),
			Description: fmt.Sprintf("按 %s 基准换算", benchmarkGroup),
		})
	}

	if hasScout {
		weight := 0.85
		if hasPhysical {
			weight = 0.35
		}
		components = append(components, PlayerScoreComponent{
			Label:       "球探报告均分",
			Source:      scoreSourceScoutReports,
			Score:       clampScore(*input.ScoutReportAverage),
			Weight:      weight,
			Count:       input.ScoutReportCount,
			Description: "仅统计已发布或已采纳报告",
		})
	}

	if hasPhysical || hasScout {
		completenessWeight := 0.15
		if hasPhysical && hasScout {
			completenessWeight = 0.10
		}
		components = append(components, PlayerScoreComponent{
			Label:       "数据完整度",
			Source:      scoreSourceCompleteness,
			Score:       completenessScore(metricCoverage, input.ScoutReportCount),
			Weight:      completenessWeight,
			Count:       int64(len(metrics)),
			Description: fmt.Sprintf("体测覆盖 %.0f%%，报告 %d 份", metricCoverage, input.ScoutReportCount),
		})
	}

	if len(components) == 0 {
		return PlayerScoreResult{
			Score:               0,
			Potential:           "待评估",
			HasScore:            false,
			DataCoverage:        0,
			ExpectedMetricCount: expectedMetricCount,
			MetricCoverage:      0,
			Confidence:          0,
			BenchmarkGroup:      benchmarkGroup,
			FormulaVersion:      PlayerScoreFormulaVersionV2,
			Sources:             []string{},
			Components:          []PlayerScoreComponent{},
			Metrics:             []PlayerScoreMetric{},
			MissingMetrics:      missingMetrics,
			Notes:               []string{"缺少体测和已发布球探报告，未生成展示评分"},
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
		Score:               score,
		Potential:           PlayerPotentialFromScore(score),
		HasScore:            true,
		DataCoverage:        len(metrics) + int(input.ScoutReportCount),
		ExpectedMetricCount: expectedMetricCount,
		MetricCoverage:      metricCoverage,
		Confidence:          scoreConfidence(metricCoverage, input.ScoutReportCount, hasPhysical, hasScout),
		BenchmarkGroup:      benchmarkGroup,
		FormulaVersion:      PlayerScoreFormulaVersionV2,
		Sources:             sources,
		Components:          components,
		Metrics:             metrics,
		MissingMetrics:      missingMetrics,
		Notes:               scoreNotes(hasPhysical, hasScout, benchmarkGroup),
	}
}

func LoadScoringBenchmarks() *ScoringBenchmarkConfig {
	scoringBenchmarksOnce.Do(func() {
		cfg := loadScoringBenchmarksFromDisk()
		scoringBenchmarks = &cfg
	})
	return scoringBenchmarks
}

func loadScoringBenchmarksFromDisk() ScoringBenchmarkConfig {
	candidates := scoringBenchmarkPaths()
	for _, path := range candidates {
		raw, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var cfg ScoringBenchmarkConfig
		if err := json.Unmarshal(raw, &cfg); err == nil && cfg.Version != "" && len(cfg.Groups) > 0 {
			return cfg
		}
	}
	return ScoringBenchmarkConfig{Version: PlayerScoreFormulaVersionV2, Groups: map[string]map[string]map[string]ScoringMetricBenchmark{}}
}

func scoringBenchmarkPaths() []string {
	paths := []string{}
	if cwd, err := os.Getwd(); err == nil {
		paths = append(paths,
			filepath.Join(cwd, "config", "scoring_benchmarks.json"),
			filepath.Join(cwd, "..", "config", "scoring_benchmarks.json"),
		)
	}
	if _, file, _, ok := runtime.Caller(0); ok {
		root := filepath.Dir(filepath.Dir(file))
		paths = append(paths, filepath.Join(root, "config", "scoring_benchmarks.json"))
	}
	return paths
}

func collectBenchmarkedPhysicalMetrics(input PlayerScoreInput, cfg *ScoringBenchmarkConfig, ageGroup string, positionGroup string) ([]PlayerScoreMetric, []string, int) {
	values := physicalMetricValues(input)
	expectedKeys := benchmarkMetricKeys(cfg, ageGroup, positionGroup)
	metrics := make([]PlayerScoreMetric, 0, len(values))
	missing := make([]string, 0)

	for _, key := range expectedKeys {
		spec := physicalMetricSpecByKey(key)
		value, hasValue := values[key]
		if !hasValue {
			missing = append(missing, spec.Label)
			continue
		}
		benchmark, ok := lookupMetricBenchmark(cfg, ageGroup, positionGroup, key)
		if !ok {
			continue
		}
		score, percentile := scoreAgainstBenchmark(value.Value, benchmark)
		metrics = append(metrics, PlayerScoreMetric{
			Key:            key,
			Label:          value.Label,
			Value:          formatScoreValue(value.Value, value.Unit),
			Score:          score,
			Direction:      benchmark.Direction,
			Percentile:     percentile,
			Benchmark:      benchmarkSummary(benchmark, value.Unit),
			BenchmarkGroup: fmt.Sprintf("%s/%s", ageGroup, positionGroup),
		})
	}
	return metrics, missing, len(expectedKeys)
}

func physicalMetricValues(input PlayerScoreInput) map[string]physicalMetricValue {
	values := make(map[string]physicalMetricValue)
	put := func(key string, value float64) {
		if value <= 0 {
			return
		}
		spec := physicalMetricSpecByKey(key)
		values[key] = physicalMetricValue{Value: value, Label: spec.Label, Unit: spec.Unit}
	}
	putInt := func(key string, value *int) {
		if value == nil {
			return
		}
		put(key, float64(*value))
	}
	putFloat := func(key string, value *float64) {
		if value == nil {
			return
		}
		put(key, *value)
	}

	if input.PhysicalRecord != nil {
		r := input.PhysicalRecord
		putFloat("sprint_30m", r.Sprint30m)
		putFloat("sprint_50m", r.Sprint50m)
		putFloat("sprint_100m", r.Sprint100m)
		putFloat("agility_ladder", r.AgilityLadder)
		putFloat("t_test", r.TTest)
		putFloat("shuttle_run", r.ShuttleRun)
		putFloat("standing_long_jump", r.StandingLongJump)
		putFloat("vertical_jump", r.VerticalJump)
		putFloat("sit_and_reach", r.SitAndReach)
		putInt("push_up", r.PushUp)
		putInt("sit_up", r.SitUp)
		putInt("plank", r.Plank)
		return values
	}

	if input.User == nil {
		return values
	}
	user := input.User
	put("sprint_30m", user.Sprint30m)
	put("shuttle_run", user.FiveMeterShuttle)
	put("agility_ladder", user.Coordination)
	put("standing_long_jump", user.StandingLongJump)
	put("sit_and_reach", firstPositiveFloat(user.SitAndReach, user.Flexibility))
	put("push_up", float64(user.PushUp))
	put("sit_up", float64(user.SitUps))
	put("plank", 0)
	return values
}

func benchmarkMetricKeys(cfg *ScoringBenchmarkConfig, ageGroup string, positionGroup string) []string {
	keySet := make(map[string]struct{})
	if group, ok := cfg.Groups[ageGroup]; ok {
		for key := range group["GENERAL"] {
			keySet[key] = struct{}{}
		}
		for key := range group[positionGroup] {
			keySet[key] = struct{}{}
		}
	}
	if len(keySet) == 0 {
		for _, spec := range physicalMetricSpecs {
			keySet[spec.Key] = struct{}{}
		}
	}
	keys := make([]string, 0, len(keySet))
	for key := range keySet {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func lookupMetricBenchmark(cfg *ScoringBenchmarkConfig, ageGroup string, positionGroup string, key string) (ScoringMetricBenchmark, bool) {
	if group, ok := cfg.Groups[ageGroup]; ok {
		if benchmarks, ok := group[positionGroup]; ok {
			if benchmark, ok := benchmarks[key]; ok {
				return benchmark, true
			}
		}
		if benchmarks, ok := group["GENERAL"]; ok {
			if benchmark, ok := benchmarks[key]; ok {
				return benchmark, true
			}
		}
	}
	if group, ok := cfg.Groups["U13-U14"]; ok {
		if benchmarks, ok := group["GENERAL"]; ok {
			benchmark, ok := benchmarks[key]
			return benchmark, ok
		}
	}
	return ScoringMetricBenchmark{}, false
}

func scoreAgainstBenchmark(value float64, benchmark ScoringMetricBenchmark) (float64, int) {
	points := []benchmarkPoint{
		{Value: benchmark.P10, Score: 45, Percentile: 10},
		{Value: benchmark.P25, Score: 60, Percentile: 25},
		{Value: benchmark.P50, Score: 75, Percentile: 50},
		{Value: benchmark.P75, Score: 88, Percentile: 75},
		{Value: benchmark.P90, Score: 96, Percentile: 90},
	}
	if benchmark.Direction == "lower_better" {
		for i := range points {
			points[i].Value = -points[i].Value
		}
		value = -value
	}
	sort.Slice(points, func(i, j int) bool {
		return points[i].Value < points[j].Value
	})
	if value <= points[0].Value {
		score := points[0].Score - outOfRangePenalty(points[0], points[1], points[0].Value-value)
		return roundScore(score), maxInt(1, points[0].Percentile-5)
	}
	last := points[len(points)-1]
	beforeLast := points[len(points)-2]
	if value >= last.Value {
		score := last.Score + outOfRangeBonus(beforeLast, last, value-last.Value)
		return roundScore(score), minInt(99, last.Percentile+5)
	}
	for i := 0; i < len(points)-1; i++ {
		left := points[i]
		right := points[i+1]
		if value < left.Value || value > right.Value {
			continue
		}
		ratio := (value - left.Value) / (right.Value - left.Value)
		score := left.Score + ratio*(right.Score-left.Score)
		percentile := int(float64(left.Percentile) + ratio*float64(right.Percentile-left.Percentile))
		return roundScore(score), percentile
	}
	return 0, 0
}

func outOfRangePenalty(left benchmarkPoint, right benchmarkPoint, delta float64) float64 {
	span := right.Value - left.Value
	if span <= 0 {
		return 0
	}
	return minFloat(25, delta/span*15)
}

func outOfRangeBonus(left benchmarkPoint, right benchmarkPoint, delta float64) float64 {
	span := right.Value - left.Value
	if span <= 0 {
		return 0
	}
	return minFloat(4, delta/span*4)
}

func ageBenchmarkGroup(user *models.User) string {
	if user == nil || user.Age <= 0 {
		return "U13-U14"
	}
	switch {
	case user.Age <= 10:
		return "U8-U10"
	case user.Age <= 12:
		return "U11-U12"
	case user.Age <= 14:
		return "U13-U14"
	case user.Age <= 16:
		return "U15-U16"
	case user.Age <= 18:
		return "U17-U18"
	default:
		return "18+"
	}
}

func positionBenchmarkGroup(user *models.User) string {
	if user == nil {
		return "GENERAL"
	}
	position := strings.TrimSpace(user.Position)
	switch {
	case strings.Contains(position, "门将"):
		return "GK"
	case strings.Contains(position, "后卫") || strings.Contains(position, "中后卫") || strings.Contains(position, "边后卫"):
		return "DEF"
	case strings.Contains(position, "中场") || strings.Contains(position, "前腰") || strings.Contains(position, "后腰"):
		return "MID"
	case strings.Contains(position, "前锋") || strings.Contains(position, "边锋"):
		return "FWD"
	default:
		return "GENERAL"
	}
}

func physicalMetricSpecByKey(key string) physicalMetricSpec {
	for _, spec := range physicalMetricSpecs {
		if spec.Key == key {
			return spec
		}
	}
	return physicalMetricSpec{Key: key, Label: key, Unit: ""}
}

func benchmarkSummary(benchmark ScoringMetricBenchmark, unit string) string {
	return fmt.Sprintf("P10=%s, P50=%s, P90=%s", formatScoreValue(benchmark.P10, unit), formatScoreValue(benchmark.P50, unit), formatScoreValue(benchmark.P90, unit))
}

func completenessScore(metricCoverage float64, reportCount int64) float64 {
	reportScore := 0.0
	if reportCount > 0 {
		reportScore = minFloat(100, 55+float64(reportCount)*15)
	}
	if reportScore == 0 {
		return roundScore(metricCoverage)
	}
	return roundScore(metricCoverage*0.7 + reportScore*0.3)
}

func scoreConfidence(metricCoverage float64, reportCount int64, hasPhysical bool, hasScout bool) float64 {
	confidence := 0.10
	if hasPhysical {
		confidence += metricCoverage / 100 * 0.55
	}
	if hasScout {
		confidence += minFloat(1, float64(reportCount)/3) * 0.30
	}
	if hasPhysical && hasScout {
		confidence += 0.05
	}
	return roundConfidence(minFloat(1, confidence))
}

func roundConfidence(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}

func scoreNotes(hasPhysical bool, hasScout bool, benchmarkGroup string) []string {
	notes := []string{fmt.Sprintf("评分基于 %s 基准", benchmarkGroup)}
	if !hasPhysical {
		notes = append(notes, "缺少体测记录，体测分未参与计算")
	}
	if !hasScout {
		notes = append(notes, "缺少已发布或已采纳球探报告")
	}
	return notes
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
