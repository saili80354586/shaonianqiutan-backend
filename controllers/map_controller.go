package controllers

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/cache"
	"github.com/shaonianqiutan/backend/config"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/repositories"
	"github.com/shaonianqiutan/backend/utils"
)

// MapController 地图数据控制器
type MapController struct{}

// NewMapController 创建地图控制器
func NewMapController() *MapController {
	return &MapController{}
}

// NationalMapResponse 全国地图响应
type NationalMapResponse struct {
	Provinces []ProvinceAggregate `json:"provinces"`
}

// ProvinceAggregate 省份聚合数据
type ProvinceAggregate struct {
	ProvinceCode          string  `json:"provinceCode"`
	ProvinceName          string  `json:"provinceName"`
	PlayerCount           int64   `json:"playerCount"`
	ClubCount             int64   `json:"clubCount"`
	ScoutReportCount      int64   `json:"scoutReportCount"`
	AvgScore              float64 `json:"avgScore"`
	NewPlayerCount30d     int64   `json:"newPlayerCount30d"`
	ReportCoverageRate    float64 `json:"reportCoverageRate"`
	HeatLevel             int     `json:"heatLevel"`
}

// ProvincialMapResponse 省市地图响应
type ProvincialMapResponse struct {
	Cities []CityAggregate `json:"cities"`
}

// CityAggregate 城市聚合数据
type CityAggregate struct {
	CityCode           string  `json:"cityCode"`
	CityName           string  `json:"cityName"`
	PlayerCount        int64   `json:"playerCount"`
	ClubCount          int64   `json:"clubCount"`
	AvgScore           float64 `json:"avgScore"`
	NewPlayerCount30d  int64   `json:"newPlayerCount30d"`
	ReportCoverageRate float64 `json:"reportCoverageRate"`
	HeatLevel          int     `json:"heatLevel"`
	CenterX            float64 `json:"centerX"`
	CenterY            float64 `json:"centerY"`
	TopPlayers         []gin.H `json:"topPlayers"`
}

// CityMapResponse 城市地图响应
type CityMapResponse struct {
	Players []CityPlayer `json:"players"`
}

// CityPlayer 城市级球员散点数据
type CityPlayer struct {
	ID           uint     `json:"id"`
	Name         string   `json:"name"`
	Avatar       string   `json:"avatar"`
	Position     string   `json:"position"`
	Age          int      `json:"age"`
	Score        float64  `json:"score"`
	Potential    string   `json:"potential"`
	Tags         []string `json:"tags"`
	NormalizedX  float64  `json:"normalizedX"`
	NormalizedY  float64  `json:"normalizedY"`
	HasReport    bool     `json:"hasReport"`
}

// GetNationalMapData 获取全国地图聚合数据（支持多图层 ?layer=players|clubs|coaches|analysts|scouts|all）
func (ctrl *MapController) GetNationalMapData(c *gin.Context) {
	db := config.GetDB()
	layer := repositories.ParseEntityLayer(c.Query("layer"))

	// P5-7: 缓存命中检查（TTL=5min）
	mapCache := cache.GetMapCache()
	if cached, ok := mapCache.Get("national", map[string]string{"layer": string(layer)}); ok {
		if data, valid := cached.(cache.NationalCacheData); valid {
			utils.Success(c, "", gin.H{"layer": data.Layer, "provinces": data.Provinces, "cached": true})
			return
		}
	}

	// 使用多图层仓库查询
	repo := repositories.NewMultiLayerMapRepository(db)
	aggregates, err := repo.GetNationalAggregates(layer)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询省份数据失败")
		return
	}

	// 写入缓存
	mapCache.Set("national", map[string]string{"layer": string(layer)}, cache.NationalCacheData{
		Layer:     string(layer),
		Provinces: aggregates,
	})

	utils.Success(c, "", gin.H{"layer": layer, "provinces": aggregates})
}

// GetProvincialMapData 获取省市地图聚合数据（支持多图层 ?layer=）
func (ctrl *MapController) GetProvincialMapData(c *gin.Context) {
	province := c.Query("province")
	if province == "" {
		utils.Error(c, http.StatusBadRequest, "省份参数不能为空")
		return
	}

	db := config.GetDB()
	layer := repositories.ParseEntityLayer(c.Query("layer"))

	// P5-7: 缓存命中检查（TTL=5min）
	mapCache := cache.GetMapCache()
	if cached, ok := mapCache.Get("provincial", map[string]string{"province": province, "layer": string(layer)}); ok {
		if data, valid := cached.(cache.ProvincialCacheData); valid {
			utils.Success(c, "", gin.H{"layer": data.Layer, "cities": data.Cities, "truncated": data.Truncated, "totalCities": data.TotalCities, "cached": true})
			return
		}
	}

	repo := repositories.NewMultiLayerMapRepository(db)
	aggregates, err := repo.GetProvincialAggregates(province, layer)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询城市数据失败")
		return
	}

	// P4-7: 性能保护 — Top 50 城市限制
	totalCities := len(aggregates)
	truncated := totalCities > 50
	if truncated {
		sort.Slice(aggregates, func(i, j int) bool {
			return aggregates[i].Count > aggregates[j].Count
		})
		aggregates = aggregates[:50]
	}

	// 写入缓存
	mapCache.Set("provincial", map[string]string{"province": province, "layer": string(layer)}, cache.ProvincialCacheData{
		Layer:       string(layer),
		Province:    province,
		Cities:      aggregates,
		Truncated:   truncated,
		TotalCities: totalCities,
	})

	utils.Success(c, "", gin.H{"layer": layer, "cities": aggregates, "truncated": truncated, "totalCities": totalCities})
}

// GetCityMapData 获取城市级散点数据（支持多图层 ?layer=）
func (ctrl *MapController) GetCityMapData(c *gin.Context) {
	province := c.Query("province")
	city := c.Query("city")
	if province == "" || city == "" {
		utils.Error(c, http.StatusBadRequest, "省份和城市参数不能为空")
		return
	}

	db := config.GetDB()
	layer := repositories.ParseEntityLayer(c.Query("layer"))

	// P5-7: 缓存命中检查（TTL=5min）
	mapCache := cache.GetMapCache()
	if cached, ok := mapCache.Get("city", map[string]string{"province": province, "city": city, "layer": string(layer)}); ok {
		if data, valid := cached.(cache.CityCacheData); valid {
			utils.Success(c, "", gin.H{"layer": data.Layer, "entities": data.Entities, "cached": true})
			return
		}
	}

	repo := repositories.NewMultiLayerMapRepository(db)
	entities, err := repo.GetCityEntities(province, city, layer)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询实体数据失败")
		return
	}

	// 写入缓存
	mapCache.Set("city", map[string]string{"province": province, "city": city, "layer": string(layer)}, cache.CityCacheData{
		Layer:    string(layer),
		Province: province,
		City:     city,
		Entities: entities,
	})

	utils.Success(c, "", gin.H{"layer": layer, "entities": entities})
}

// GetScoutMapData 兼容旧版 V2 接口
func (ctrl *MapController) GetScoutMapData(c *gin.Context) {
	db := config.GetDB()

	var users []models.User
	if err := db.Where("role = ? AND status = ?", "user", "active").Find(&users).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询球员数据失败")
		return
	}

	var playerList []gin.H
	provinceMap := make(map[string]bool)
	for _, user := range users {
		playerList = append(playerList, gin.H{
			"id":       user.ID,
			"name":     user.Name,
			"nickname": user.Nickname,
			"province": user.Province,
			"city":     user.City,
			"position": user.Position,
			"age":      user.Age,
			"height":   user.Height,
			"weight":   user.Weight,
			"foot":     user.Foot,
			"club":     user.Club,
			"avatar":   user.Avatar,
		})
		if user.Province != "" {
			provinceMap[user.Province] = true
		}
	}

	utils.Success(c, "", gin.H{
		"players":   playerList,
		"total":     len(playerList),
		"provinces": len(provinceMap),
	})
}

// GetPlayersByProvince 兼容旧版按省份获取球员
func (ctrl *MapController) GetPlayersByProvince(c *gin.Context) {
	province := c.Query("province")
	if province == "" {
		utils.Error(c, http.StatusBadRequest, "省份参数不能为空")
		return
	}

	db := config.GetDB()

	var users []models.User
	if err := db.Where("role = ? AND status = ? AND province = ?", "user", "active", province).Find(&users).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询球员数据失败")
		return
	}

	cityMap := make(map[string][]gin.H)
	for _, user := range users {
		playerData := gin.H{
			"id":       user.ID,
			"name":     user.Name,
			"nickname": user.Nickname,
			"city":     user.City,
			"position": user.Position,
			"age":      user.Age,
			"avatar":   user.Avatar,
		}
		cityMap[user.City] = append(cityMap[user.City], playerData)
	}

	utils.Success(c, "", gin.H{
		"province": province,
		"data":     cityMap,
		"total":    len(users),
	})
}

// ========== 辅助函数 ==========

func calculateHeatLevel(count int64) int {
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

func calculateRandomScore(min, max float64) float64 {
	rand.Seed(time.Now().UnixNano())
	score := min + rand.Float64()*(max-min)
	return math.Round(score*10) / 10
}

func calculatePotential(score float64) string {
	switch {
	case score >= 85:
		return "A"
	case score >= 75:
		return "B"
	case score >= 65:
		return "C"
	default:
		return "D"
	}
}

func generateTags(position string) []string {
	tagMap := map[string][]string{
		"前锋": {"速度型", "突破强", "射门准"},
		"中场": {"传球好", "视野广", "控球稳"},
		"后卫": {"防守硬", "对抗强", "头球好"},
		"门将": {"反应快", "门线技术好", "指挥能力强"},
	}
	tags, ok := tagMap[position]
	if !ok {
		return []string{"潜力新星"}
	}
	return tags
}

func generateNormalizedCoordinates(seed uint) (float64, float64) {
	rand.Seed(int64(seed) + time.Now().Unix()/86400)
	// 基于正态分布，中心在 0.5，标准差 0.2，确保大部分点落在 0.1-0.9 之间
	nx := 0.5 + rand.NormFloat64()*0.15
	ny := 0.5 + rand.NormFloat64()*0.15
	nx = math.Max(0.05, math.Min(0.95, nx))
	ny = math.Max(0.05, math.Min(0.95, ny))
	return nx, ny
}

func getProvinceCode(name string) string {
	// 简化版省份编码映射，后续可维护完整对照表
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

func getCityCode(name string) string {
	// 城市编码后续维护完整对照表，暂时用拼音简化
	if name == "" {
		return ""
	}
	return name
}

// GetPlayerMapProfile 获取球员地图详情页资料
func (ctrl *MapController) GetPlayerMapProfile(c *gin.Context) {
	userIDStr := c.Param("userId")

	db := config.GetDB()
	var user models.User
	if err := db.Where("id = ? AND role = ? AND status = ?", userIDStr, "user", "active").First(&user).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "球员不存在")
		return
	}

	// 解析 JSON 字段
	playingStyles := parseMapJSONArray(user.PlayingStyle)
	technicalTags := parseMapJSONArray(user.TechnicalTags)
	mentalTags := parseMapJSONArray(user.MentalTags)
	experiences := parseMapExperiences(user.Experiences)

	// 计算评分（基于体测数据或随机）
	score := calculateScoreFromUser(&user)
	potential := calculatePotential(score)
	
	// 如果没有技术标签，生成默认标签
	if len(technicalTags) == 0 {
		technicalTags = generateTags(user.Position)
	}

	profile := gin.H{
		"id":       user.ID,
		"name":     user.Name,
		"nickname": user.Nickname,
		"avatar":   user.Avatar,
		"city":     user.City,
		"province": user.Province,
		"age":      user.Age,
		"position": user.Position,
		"second_position": user.SecondPosition,
		"height":   user.Height,
		"weight":   user.Weight,
		"foot":     user.Foot,
		"club":     user.Club,
		"current_team": user.CurrentTeam,
		"start_year": user.StartYear,
		"jersey_number": user.JerseyNumber,
		"jersey_color": user.JerseyColor,
		"fa_registered": user.FARegistered,
		"association": user.Association,
		"school": user.School,
		"tags":     technicalTags,
		"playing_style": playingStyles,
		"technical_tags": technicalTags,
		"mental_tags": mentalTags,
		"experiences": experiences,
		"score":    score,
		"potential": potential,
		"heat": gin.H{
			"views7d":   int(user.ID) % 50,
			"followers": int(user.ID) % 10,
		},
		"radar": gin.H{
			"visible":   true,
			"dimensions": []string{"速度", "技术", "身体", "战术", "心理", "潜力"},
			"values":    []float64{85, 78, 80, 75, 82, 88},
		},
		"physical": gin.H{
			"visible": true,
			"items": []gin.H{
				{"name": "30m冲刺", "value": fmt.Sprintf("%.1fs", user.Sprint30m), "percentile": calculateMapPercentile(user.Sprint30m, 4.5, 5.5)},
				{"name": "立定跳远", "value": fmt.Sprintf("%.0fcm", user.StandingLongJump), "percentile": calculateMapPercentile(user.StandingLongJump, 180, 220)},
				{"name": "俯卧撑", "value": fmt.Sprintf("%d个", user.PushUp), "percentile": calculateMapPercentile(float64(user.PushUp), 5, 20)},
			},
		},
		"timeline": buildMapTimeline(experiences),
		"reports": []gin.H{
			{"id": 101, "type": "ai", "author": "AI分析师", "score": 82, "summary": "突破能力强，建议加强逆足训练"},
		},
		"permissions": gin.H{
			"canViewRadar":   true,
			"canViewPhysical": true,
			"canViewReports": false,
			"canContact":     false,
		},
	}

	utils.Success(c, "", profile)
}

// ============ 辅助函数 ============

func parseMapJSONArray(s string) []string {
	if s == "" {
		return []string{}
	}
	var arr []string
	json.Unmarshal([]byte(s), &arr)
	return arr
}

func parseMapExperiences(s string) []gin.H {
	if s == "" {
		return []gin.H{}
	}
	var items []struct {
		Period     string `json:"period"`
		Team       string `json:"team"`
		Position   string `json:"position"`
		Achievement string `json:"achievement,omitempty"`
	}
	if err := json.Unmarshal([]byte(s), &items); err != nil {
		return []gin.H{}
	}
	result := make([]gin.H, len(items))
	for i, item := range items {
		result[i] = gin.H{
			"date": item.Period,
			"type": "experience",
			"title": item.Team,
			"summary": item.Position,
			"achievement": item.Achievement,
		}
	}
	return result
}

func calculateScoreFromUser(user *models.User) float64 {
	// 基于体测数据计算评分
	score := 75.0
	if user.Sprint30m > 0 {
		// 30米冲刺越快分数越高
		score += (5.5 - user.Sprint30m) * 10
	}
	if user.StandingLongJump > 0 {
		// 立定跳远越远分数越高
		score += (user.StandingLongJump - 180) / 5
	}
	if user.PushUp > 0 {
		// 俯卧撑越多分数越高
		score += float64(user.PushUp - 10) * 0.5
	}
	// 限制在 60-95 范围
	if score < 60 {
		score = 60
	}
	if score > 95 {
		score = 95
	}
	return math.Round(score*10) / 10
}

func calculateMapPercentile(value float64, min, max float64) int {
	if value <= 0 {
		return 50
	}
	if value <= min {
		return 30
	}
	if value >= max {
		return 95
	}
	return int(30 + (value-min)/(max-min)*65)
}

func buildMapTimeline(experiences []gin.H) []gin.H {
	if len(experiences) == 0 {
		return []gin.H{
			{"date": fmt.Sprintf("%d", time.Now().Year()-1) + "-01", "type": "experience", "title": "开始踢球", "summary": "加入青训"},
		}
	}
	return experiences
}

// DashboardStats 数据看板统计响应
type DashboardStats struct {
	TotalPlayers      int64       `json:"totalPlayers"`
	TotalProvinces    int64       `json:"totalProvinces"`
	AvgAge            float64     `json:"avgAge"`
	AvgScore          float64     `json:"avgScore"`
	MonthlyNew        int64       `json:"monthlyNew"`
	RegionDistribution []gin.H    `json:"regionDistribution"`
	AgeDistribution    []gin.H    `json:"ageDistribution"`
	PositionDistribution []gin.H  `json:"positionDistribution"`
	ScoreRanking       []gin.H    `json:"scoreRanking"`
	GrowthTrend        []gin.H    `json:"growthTrend"`
}

// GetDashboardStats 获取数据看板统计
func (ctrl *MapController) GetDashboardStats(c *gin.Context) {
	db := config.GetDB()

	var users []models.User
	if err := db.Where("role = ? AND status = ?", "user", "active").Find(&users).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询球员数据失败")
		return
	}

	// 基础统计
	var totalPlayers int64
	provinceMap := make(map[string]bool)
	var totalAge int64
	var totalScore float64
	var monthlyNew int64
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// 分布统计
	regionMap := make(map[string]int64)
	ageMap := make(map[string]int64)
	positionMap := make(map[string]int64)
	scoreLevelMap := make(map[string]int64)
	growthMap := make(map[string]int64)

	for _, u := range users {
		totalPlayers++
		if u.Province != "" {
			provinceMap[u.Province] = true
			regionMap[u.Province]++
		}
		totalAge += int64(u.Age)
		score := calculateRandomScore(60, 95)
		totalScore += score

		if u.CreatedAt.After(monthStart) || u.CreatedAt.Equal(monthStart) {
			monthlyNew++
		}

		// 年龄分布（按段）
		ageGroup := fmt.Sprintf("%d岁", u.Age)
		ageMap[ageGroup]++

		// 位置分布
		pos := u.Position
		if pos == "" {
			pos = "未知"
		}
		positionMap[pos]++

		// 评分等级分布
		level := calculatePotential(score)
		scoreLevelMap[level]++

		// 成长趋势（按注册月份）
		monthKey := u.CreatedAt.Format("2006-01")
		growthMap[monthKey]++
	}

	// 构造地区分布（Top 10）
	var regionDist []gin.H
	for prov, count := range regionMap {
		regionDist = append(regionDist, gin.H{"name": prov, "value": count})
	}
	// 按数量降序取前10
	sort.Slice(regionDist, func(i, j int) bool {
		return regionDist[i]["value"].(int64) > regionDist[j]["value"].(int64)
	})
	if len(regionDist) > 10 {
		regionDist = regionDist[:10]
	}

	// 构造年龄分布（排序）
	var ageDist []gin.H
	var ageKeys []string
	for k := range ageMap {
		ageKeys = append(ageKeys, k)
	}
	sort.Slice(ageKeys, func(i, j int) bool {
		ai, _ := strconv.Atoi(strings.TrimSuffix(ageKeys[i], "岁"))
		aj, _ := strconv.Atoi(strings.TrimSuffix(ageKeys[j], "岁"))
		return ai < aj
	})
	for _, k := range ageKeys {
		ageDist = append(ageDist, gin.H{"name": k, "value": ageMap[k]})
	}

	// 构造位置分布
	var posDist []gin.H
	for _, pos := range []string{"前锋", "中场", "后卫", "门将", "未知"} {
		if count, ok := positionMap[pos]; ok {
			posDist = append(posDist, gin.H{"name": pos, "value": count})
		}
	}

	// 评分排名（按等级）
	var scoreRank []gin.H
	for _, level := range []string{"S", "A", "B", "C", "D"} {
		scoreRank = append(scoreRank, gin.H{"name": level + "级", "value": scoreLevelMap[level]})
	}

	// 成长趋势（最近12个月）
	var growthTrend []gin.H
	for i := 11; i >= 0; i-- {
		t := now.AddDate(0, -i, 0)
		key := t.Format("2006-01")
		label := fmt.Sprintf("%d月", t.Month())
		if t.Year() != now.Year() {
			label = fmt.Sprintf("%d年%d月", t.Year(), t.Month())
		}
		growthTrend = append(growthTrend, gin.H{"name": label, "value": growthMap[key]})
	}

	avgAge := 0.0
	avgScore := 0.0
	if totalPlayers > 0 {
		avgAge = math.Round(float64(totalAge)/float64(totalPlayers)*10) / 10
		avgScore = math.Round(totalScore/float64(totalPlayers)*10) / 10
	}

	stats := DashboardStats{
		TotalPlayers:         totalPlayers,
		TotalProvinces:       int64(len(provinceMap)),
		AvgAge:               avgAge,
		AvgScore:             avgScore,
		MonthlyNew:           monthlyNew,
		RegionDistribution:   regionDist,
		AgeDistribution:      ageDist,
		PositionDistribution: posDist,
		ScoreRanking:         scoreRank,
		GrowthTrend:          growthTrend,
	}

	utils.Success(c, "", stats)
}

// GetOverseasPlayers 获取海外球员列表
func (ctrl *MapController) GetOverseasPlayers(c *gin.Context) {
	db := config.GetDB()

	var users []models.User
	if err := db.Where("role = ? AND status = ? AND country != ? AND country != ?", "user", "active", "", "中国").
		Order("created_at DESC").
		Limit(50).
		Find(&users).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询海外球员失败")
		return
	}

	var players []gin.H
	for _, u := range users {
		score := calculateRandomScore(65, 98)
		potential := calculatePotential(score)
		tags := generateTags(u.Position)
		country := u.Country
		if country == "" {
			country = "海外"
		}
		players = append(players, gin.H{
			"id":        u.ID,
			"name":      u.Name,
			"avatar":    u.Avatar,
			"country":   country,
			"city":      u.City,
			"position":  u.Position,
			"age":       u.Age,
			"score":     score,
			"potential": potential,
			"tags":      tags,
		})
	}

	utils.Success(c, "", gin.H{"players": players, "total": len(players)})
}

// GetMyRank 获取当前登录球员的排名信息
func (ctrl *MapController) GetMyRank(c *gin.Context) {
	userId := c.GetUint("userId")
	if userId == 0 {
		utils.Error(c, http.StatusUnauthorized, "请先登录")
		return
	}

	db := config.GetDB()

	var me models.User
	if err := db.Where("id = ? AND role = ? AND status = ?", userId, "user", "active").First(&me).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "球员信息不存在")
		return
	}

	// 全省排名（按年龄分组内的评分）
	var provinceRank int64 = 1
	if me.Province != "" {
		var provinceCount int64
		db.Model(&models.User{}).
			Where("role = ? AND status = ? AND province = ? AND age = ? AND score > ?", "user", "active", me.Province, me.Age, 0).
			Count(&provinceCount)
		// 简单模拟：实际应基于真实评分字段排序
		if provinceCount > 0 {
			provinceRank = int64(me.ID)%provinceCount + 1
		}
	}

	// 全市排名
	var cityRank int64 = 1
	if me.City != "" {
		var cityCount int64
		db.Model(&models.User{}).
			Where("role = ? AND status = ? AND province = ? AND city = ? AND age = ?", "user", "active", me.Province, me.City, me.Age).
			Count(&cityCount)
		if cityCount > 0 {
			cityRank = int64(me.ID)%cityCount + 1
		}
	}

	// 全国同位置排名
	var positionRank int64 = 1
	if me.Position != "" {
		var positionCount int64
		db.Model(&models.User{}).
			Where("role = ? AND status = ? AND position = ? AND age = ?", "user", "active", me.Position, me.Age).
			Count(&positionCount)
		if positionCount > 0 {
			positionRank = int64(me.ID)%positionCount + 1
		}
	}

	score := calculateRandomScore(60, 95)
	utils.Success(c, "", gin.H{
		"player": gin.H{
			"id":       me.ID,
			"name":     me.Name,
			"province": me.Province,
			"city":     me.City,
			"position": me.Position,
			"age":      me.Age,
			"score":    score,
		},
		"ranks": gin.H{
			"provinceRank": provinceRank,
			"cityRank":     cityRank,
			"positionRank": positionRank,
		},
	})
}

// RecommendPlayer 推荐球员数据
type RecommendPlayer struct {
	ID           uint     `json:"id"`
	Name         string   `json:"name"`
	Avatar       string   `json:"avatar"`
	Position     string   `json:"position"`
	Age          int      `json:"age"`
	City         string   `json:"city"`
	Province     string   `json:"province"`
	Score        float64  `json:"score"`
	Potential    string   `json:"potential"`
	Tags         []string `json:"tags"`
	Reason       string   `json:"reason"`
}

// GetRecommendations 智能推荐「猜你感兴趣」（支持未登录匿名访问）
func (ctrl *MapController) GetRecommendations(c *gin.Context) {
	userId := c.GetUint("userId")
	db := config.GetDB()

	var me models.User
	var isPlayer bool
	if userId > 0 {
		db.Where("id = ? AND status = ?", userId, "active").First(&me)
		isPlayer = me.Role == "user"
	}

	var allUsers []models.User
	if err := db.Where("role = ? AND status = ?", "user", "active").Find(&allUsers).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询球员数据失败")
		return
	}

	// 如果有登录用户，排除自己
	candidates := make([]models.User, 0, len(allUsers))
	for _, u := range allUsers {
		if u.ID != me.ID {
			candidates = append(candidates, u)
		}
	}
	if len(candidates) == 0 {
		utils.Success(c, "", gin.H{"players": []gin.H{}, "total": 0})
		return
	}

	// 评分和标签预处理
	type scored struct {
		user   models.User
		score  float64
		reason string
	}
	scoredList := make([]scored, 0, len(candidates))
	for _, u := range candidates {
		s := calculateRandomScore(60, 95)
		scoredList = append(scoredList, scored{user: u, score: s})
	}

	// 推荐排序逻辑
	var recommendations []RecommendPlayer
	maxResults := 6

	if isPlayer && me.ID > 0 {
		// 球员视角：同位置 + 同城/同省 + 评分相近优先
		sort.Slice(scoredList, func(i, j int) bool {
			si := 0
			sj := 0
			if scoredList[i].user.Position == me.Position { si += 100 }
			if scoredList[j].user.Position == me.Position { sj += 100 }
			if scoredList[i].user.City == me.City && me.City != "" { si += 50 }
			if scoredList[j].user.City == me.City && me.City != "" { sj += 50 }
			if scoredList[i].user.Province == me.Province && me.Province != "" { si += 30 }
			if scoredList[j].user.Province == me.Province && me.Province != "" { sj += 30 }
			// 评分接近加分
			ageDiffI := abs(scoredList[i].user.Age - me.Age)
			ageDiffJ := abs(scoredList[j].user.Age - me.Age)
			si -= ageDiffI * 5
			sj -= ageDiffJ * 5
			si += int(scoredList[i].score)
			sj += int(scoredList[j].score)
			return si > sj
		})
	} else if userId > 0 {
		// B端视角：优先高评分、高潜力
		sort.Slice(scoredList, func(i, j int) bool {
			return scoredList[i].score > scoredList[j].score
		})
	} else {
		// 未登录：随机混排热门+潜力
		rand.Shuffle(len(scoredList), func(i, j int) {
			scoredList[i], scoredList[j] = scoredList[j], scoredList[i]
		})
	}

	for i := 0; i < len(scoredList) && len(recommendations) < maxResults; i++ {
		u := scoredList[i].user
		score := scoredList[i].score
		potential := calculatePotential(score)
		tags := generateTags(u.Position)
		reason := "潜力新星"
		if isPlayer && me.ID > 0 {
			if u.Position == me.Position && u.City == me.City && me.City != "" {
				reason = "同城同位置热门"
			} else if u.Position == me.Position && u.Province == me.Province && me.Province != "" {
				reason = "同省同位置推荐"
			} else if u.City == me.City && me.City != "" {
				reason = "同城潜力新星"
			} else if u.Province == me.Province && me.Province != "" {
				reason = "同省热门球员"
			} else if u.Position == me.Position {
				reason = "同位置推荐"
			}
		} else if userId > 0 {
			if score >= 85 {
				reason = "高评分热门"
			} else if potential == "A" || potential == "S" {
				reason = "高潜力球员"
			}
		} else {
			// 未登录匿名推荐
			if score >= 85 {
				reason = "平台热门"
			} else if potential == "A" || potential == "S" {
				reason = "高潜力新星"
			} else {
				reason = "值得关注"
			}
		}
		recommendations = append(recommendations, RecommendPlayer{
			ID:        u.ID,
			Name:      u.Name,
			Avatar:    u.Avatar,
			Position:  u.Position,
			Age:       u.Age,
			City:      u.City,
			Province:  u.Province,
			Score:     score,
			Potential: potential,
			Tags:      tags,
			Reason:    reason,
		})
	}

	utils.Success(c, "", gin.H{"players": recommendations, "total": len(recommendations)})
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

// GetRisingStars 获取本周新星
func (ctrl *MapController) GetRisingStars(c *gin.Context) {
	db := config.GetDB()

	// 取最近7天注册或更新的活跃球员，按ID随机模拟潜力排序
	var users []models.User
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	if err := db.Where("role = ? AND status = ? AND (created_at >= ? OR updated_at >= ?)", "user", "active", sevenDaysAgo, sevenDaysAgo).
		Order("created_at DESC").
		Limit(20).
		Find(&users).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}

	// 如果数量不足，补充一些历史高评分球员
	if len(users) < 5 {
		var extraUsers []models.User
		db.Where("role = ? AND status = ?", "user", "active").
			Order("age ASC").
			Limit(10).
			Find(&extraUsers)
		users = append(users, extraUsers...)
	}

	// 去重
	seen := make(map[uint]bool)
	var uniqueUsers []models.User
	for _, u := range users {
		if !seen[u.ID] {
			seen[u.ID] = true
			uniqueUsers = append(uniqueUsers, u)
		}
	}

	// 取前8位
	if len(uniqueUsers) > 8 {
		uniqueUsers = uniqueUsers[:8]
	}

	var players []gin.H
	for _, u := range uniqueUsers {
		score := calculateRandomScore(70, 95)
		potential := calculatePotential(score)
		tags := generateTags(u.Position)
		players = append(players, gin.H{
			"id":        u.ID,
			"name":      u.Name,
			"avatar":    u.Avatar,
			"province":  u.Province,
			"city":      u.City,
			"position":  u.Position,
			"age":       u.Age,
			"score":     score,
			"potential": potential,
			"tags":      tags,
			"reason":    "本周活跃新星",
		})
	}

	utils.Success(c, "", gin.H{"players": players, "total": len(players)})
}

// ExportCompare 导出对比数据为 CSV
func (ctrl *MapController) ExportCompare(c *gin.Context) {
	var req struct {
		PlayerIDs []uint `json:"player_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "参数错误：player_ids 必填")
		return
	}

	db := config.GetDB()
	var users []models.User
	if err := db.Where("id IN ? AND role = ? AND status = ?", req.PlayerIDs, "user", "active").Find(&users).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询球员数据失败")
		return
	}

	var csvBuilder strings.Builder
	csvBuilder.WriteString("\uFEFF姓名,年龄,位置,城市,俱乐部,综合评分,潜力\n")
	for _, u := range users {
		score := calculateRandomScore(60, 95)
		potential := calculatePotential(score)
		csvBuilder.WriteString(fmt.Sprintf("%s,%d,%s,%s,%s,%.1f,%s\n", u.Name, u.Age, u.Position, u.City, u.Club, score, potential))
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=compare.csv")
	c.String(http.StatusOK, csvBuilder.String())
}
