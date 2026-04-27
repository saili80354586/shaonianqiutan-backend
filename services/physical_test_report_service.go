package services

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/shaonianqiutan/backend/models"
	"gorm.io/gorm"
)

// PhysicalTestReportService 体测报告服务
type PhysicalTestReportService struct {
	db *gorm.DB
}

// NewPhysicalTestReportService 创建体测报告服务
func NewPhysicalTestReportService(db *gorm.DB) *PhysicalTestReportService {
	return &PhysicalTestReportService{db: db}
}

// PercentileResult 百分位结果
type PercentileResult struct {
	Value      float64 `json:"value"`
	Unit       string  `json:"unit"`
	Percentile int     `json:"percentile"`
	Rating     string  `json:"rating"`
	Change     string  `json:"change"`
	ChangeType string  `json:"change_type"`
}

// GenerateReport 生成体测报告
func (s *PhysicalTestReportService) GenerateReport(record *models.PhysicalTestRecord) (*models.PhysicalTestReportData, error) {
	report := &models.PhysicalTestReportData{}

	// 获取球员信息
	var player models.User
	s.db.First(&player, record.PlayerID)
	report.PlayerName = player.Name
	report.PlayerAge = player.Age
	report.TestDate = record.TestDate.Format("2006-01-02")

	// 计算百分位和评级
	report.TestData = s.calculateTestData(record, player.Age, player.Gender)

	// 获取历史数据计算趋势
	report.GrowthTrend = s.getGrowthTrend(record.PlayerID, record.ActivityID)

	// 计算综合评级 (简化实现)
	total := 0
	count := 0
	for _, v := range report.TestData {
		total += v.Percentile
		count++
	}
	if count > 0 {
		avg := total / count
		if avg >= 85 {
			report.OverallRating = "优秀"
		} else if avg >= 70 {
			report.OverallRating = "良好"
		} else if avg >= 50 {
			report.OverallRating = "中等"
		} else {
			report.OverallRating = "需加强"
		}
		report.Percentile = avg
	}

	// 优势与待提升
	report.Strengths = []string{"身体素质良好"}
	report.Improvements = []string{"继续努力"}

	// 训练建议
	report.TrainingSuggestions = []string{"建议加强体能训练"}

	// 营养建议
	report.NutritionSuggestions = []string{"保持均衡饮食"}

	// 休息建议
	report.RestSuggestions = []string{"保证充足睡眠"}

	// 下次建议体测时间
	report.NextTestSuggestion = time.Now().AddDate(0, 3, 0).Format("2006-01-02")

	return report, nil
}

func (s *PhysicalTestReportService) calculateTestData(record *models.PhysicalTestRecord, age int, gender string) map[string]models.TestItemData {
	data := make(map[string]models.TestItemData)

	// 身高
	if record.Height != nil {
		data["height"] = models.TestItemData{
			Value:      *record.Height,
			Unit:       "cm",
			Percentile: s.calculatePercentile(*record.Height, "height", age, gender),
			Rating:     s.getRatingFromPercentile(s.calculatePercentile(*record.Height, "height", age, gender)),
		}
	}

	// 体重
	if record.Weight != nil {
		data["weight"] = models.TestItemData{
			Value:      *record.Weight,
			Unit:       "kg",
			Percentile: s.calculatePercentile(*record.Weight, "weight", age, gender),
			Rating:     s.getRatingFromPercentile(s.calculatePercentile(*record.Weight, "weight", age, gender)),
		}
	}

	// BMI
	if record.BMI != nil {
		data["bmi"] = models.TestItemData{
			Value:      math.Round(*record.BMI*10) / 10,
			Unit:       "",
			Percentile: s.calculateBMIPercentile(*record.BMI, age),
			Rating:     s.getRatingFromPercentile(s.calculateBMIPercentile(*record.BMI, age)),
		}
	}

	// 30米跑（越小越好）
	if record.Sprint30m != nil {
		percentile := s.calculateRunningPercentile(*record.Sprint30m, "30m", age)
		data["sprint_30m"] = models.TestItemData{
			Value:      math.Round(*record.Sprint30m*100) / 100,
			Unit:       "秒",
			Percentile: percentile,
			Rating:     s.getRatingFromPercentile(percentile),
		}
	}

	// 立定跳远（越大越好）
	if record.StandingLongJump != nil {
		percentile := s.calculateJumpPercentile(*record.StandingLongJump, age)
		data["standing_long_jump"] = models.TestItemData{
			Value:      *record.StandingLongJump,
			Unit:       "cm",
			Percentile: percentile,
			Rating:     s.getRatingFromPercentile(percentile),
		}
	}

	return data
}

func (s *PhysicalTestReportService) calculatePercentile(value float64, itemType string, age int, gender string) int {
	// 简化：使用预设的年龄段数据分布
	basePercentile := 50

	// 根据数值偏移
	if value > 0 {
		basePercentile = 50 + int(value/10)*5
		if basePercentile > 95 {
			basePercentile = 95
		}
		if basePercentile < 5 {
			basePercentile = 5
		}
	}

	return basePercentile
}

func (s *PhysicalTestReportService) calculateRunningPercentile(value float64, distance string, age int) int {
	// 跑步类项目：数值越小排名越高
	basePercentile := 50
	if value < 6.0 {
		basePercentile = 85
	} else if value < 6.5 {
		basePercentile = 75
	} else if value < 7.0 {
		basePercentile = 60
	} else if value < 8.0 {
		basePercentile = 45
	} else {
		basePercentile = 25
	}
	return basePercentile
}

func (s *PhysicalTestReportService) calculateJumpPercentile(value float64, age int) int {
	if value > 200 {
		return 85
	} else if value > 180 {
		return 75
	} else if value > 160 {
		return 60
	} else if value > 140 {
		return 45
	}
	return 30
}

func (s *PhysicalTestReportService) calculateBMIPercentile(bmi float64, age int) int {
	if bmi < 16.0 {
		return 30
	} else if bmi < 18.5 {
		return 60
	} else if bmi < 22.0 {
		return 75
	} else if bmi < 25.0 {
		return 55
	}
	return 35
}

func (s *PhysicalTestReportService) getRatingFromPercentile(percentile int) string {
	switch {
	case percentile >= 90:
		return "优秀"
	case percentile >= 75:
		return "良好"
	case percentile >= 50:
		return "平均"
	case percentile >= 25:
		return "中下"
	default:
		return "需加强"
	}
}

func (s *PhysicalTestReportService) getGrowthTrend(playerID uint, excludeActivityID uint) map[string][]models.TrendPoint {
	// 获取该球员的历史体测记录
	var records []models.PhysicalTestRecord
	s.db.Where("player_id = ? AND id != ?", playerID, excludeActivityID).
		Order("test_date DESC").Limit(6).Find(&records)

	trend := make(map[string][]models.TrendPoint)
	for _, r := range records {
		point := models.TrendPoint{
			Date:  r.TestDate.Format("2006-01-02"),
			Value: 0,
		}
		if r.Sprint30m != nil {
			point.Value = *r.Sprint30m
			trend["sprint_30m"] = append(trend["sprint_30m"], point)
		}
		if r.Height != nil {
			point.Value = *r.Height
			trend["height"] = append(trend["height"], point)
		}
	}

	// 排序
	for k := range trend {
		sort.Slice(trend[k], func(i, j int) bool {
			return trend[k][i].Date < trend[k][j].Date
		})
	}

	return trend
}


func (s *PhysicalTestReportService) generateTrainingSuggestions(data map[string]models.TestItemData) []string {
	suggestions := []string{
		"继续保持当前训练节奏，注意循序渐进",
		"建议每周至少2-3次专项体能训练",
	}

	for item, itemData := range data {
		if itemData.Percentile >= 75 {
			switch item {
			case "sprint_30m":
				suggestions = append([]string{"你的速度素质出色，建议保持速度训练量，每周2次速度专项训练"}, suggestions...)
			case "standing_long_jump":
				suggestions = append([]string{"爆发力优秀，可增加力量训练进一步提升"}, suggestions...)
			}
		} else if itemData.Percentile < 50 {
			switch item {
			case "sprint_30m":
				suggestions = append([]string{"建议加强速度训练，每周3次冲刺练习"}, suggestions...)
			case "standing_long_jump":
				suggestions = append([]string{"建议增加爆发力训练，如深蹲跳、跳箱等"}, suggestions...)
			}
		}
	}

	return suggestions
}

func (s *PhysicalTestReportService) generateNutritionSuggestions(age int) []string {
	return []string{
		"正处于生长发育期，注意补充优质蛋白质和钙质",
		"每天500ml牛奶 + 鸡蛋2个 + 瘦肉/豆制品",
		"训练后及时补充碳水化合物和蛋白质",
		"保持充足水分摄入，训练前中后都要喝水",
	}
}

func (s *PhysicalTestReportService) generateRestSuggestions() []string {
	return []string{
		"保证每天8-10小时睡眠",
		"训练后充分休息48小时再进行下一次高强度训练",
		"出现疲劳或不适及时休息，不要带伤训练",
	}
}

// analyzeStrengthsAndImprovements 分析球员优劣势
func (s *PhysicalTestReportService) analyzeStrengthsAndImprovements(data map[string]models.TestItemData) (strengths []string, improvements []string) {
	names := map[string]string{
		"sprint_30m":          "30米跑",
		"standing_long_jump": "立定跳远",
		"height":            "身高",
		"weight":            "体重",
	}

	for item, itemData := range data {
		name := names[item]
		if name == "" {
			name = item
		}
		if itemData.Percentile >= 75 {
			strengths = append(strengths, name+"表现突出，处于同龄前"+fmt.Sprintf("%d", itemData.Percentile)+"%，是重要竞争优势")
		} else if itemData.Percentile < 50 {
			improvements = append(improvements, name+"处于同龄平均水平，建议加强训练")
		}
	}

	if len(strengths) == 0 {
		strengths = append(strengths, "各项素质发展均衡，继续保持")
	}
	if len(improvements) == 0 {
		improvements = append(improvements, "暂无明显短板，继续保持全面训练")
	}

	return
}
