package utils

import (
	"math"
	"math/rand"
	"time"
)

// CalculateHeatLevel 根据数量计算热度等级 1-5
func CalculateHeatLevel(count int64) int {
	switch {
	case count >= 200:
		return 5
	case count >= 100:
		return 4
	case count >= 50:
		return 3
	case count >= 20:
		return 2
	default:
		return 1
	}
}

// GenerateNormalizedCoordinates 基于种子生成归一化坐标 (0.05~0.95)
func GenerateNormalizedCoordinates(seed uint) (float64, float64) {
	r := rand.New(rand.NewSource(int64(seed) + time.Now().Unix()/86400))
	nx := 0.5 + r.NormFloat64()*0.15
	ny := 0.5 + r.NormFloat64()*0.15
	nx = math.Max(0.05, math.Min(0.95, nx))
	ny = math.Max(0.05, math.Min(0.95, ny))
	return nx, ny
}

// GetProvinceCode 省份名称转编码（简化版）
func GetProvinceCode(name string) string {
	codeMap := map[string]string{
		"北京市": "110000", "天津市": "120000", "上海市": "310000", "重庆市": "500000",
		"河北省": "130000", "山西省": "140000", "辽宁省": "210000", "吉林省": "220000",
		"黑龙江省": "230000", "江苏省": "320000", "浙江省": "330000", "安徽省": "340000",
		"福建省": "350000", "江西省": "360000", "山东省": "370000", "河南省": "410000",
		"湖北省": "420000", "湖南省": "430000", "广东省": "440000", "海南省": "460000",
		"四川省": "510000", "贵州省": "520000", "云南省": "530000", "陕西省": "610000",
		"甘肃省": "620000", "青海省": "630000", "台湾省": "710000", "内蒙古自治区": "150000",
		"广西壮族自治区": "450000", "西藏自治区": "540000", "宁夏回族自治区": "640000",
		"新疆维吾尔自治区": "650000", "香港特别行政区": "810000", "澳门特别行政区": "820000",
	}
	if code, ok := codeMap[name]; ok {
		return code
	}
	return ""
}

// GetCityCode 城市名称转编码（简化版，暂返回原名）
func GetCityCode(name string) string {
	if name == "" {
		return ""
	}
	return name
}

// GenerateTags 根据位置生成标签
func GenerateTags(position string) []string {
	tagMap := map[string][]string{
		"前锋": {"速度型", "突破强", "射门准"},
		"中场": {"传球好", "视野广", "控球稳"},
		"后卫": {"防守硬", "对抗强", "头球好"},
		"门将": {"反应快", "门线技术好", "指挥能力强"},
	}
	if tags, ok := tagMap[position]; ok {
		return tags
	}
	return []string{"潜力新星"}
}
