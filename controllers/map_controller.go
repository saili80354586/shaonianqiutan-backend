package controllers

import (
	"encoding/json"
	"fmt"
	"math"
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
	"gorm.io/gorm"
)

// MapController 地图数据控制器
type MapController struct{}

// NewMapController 创建地图控制器
func NewMapController() *MapController {
	return &MapController{}
}

const (
	dataVChinaGeoJSONURL    = "https://geo.datav.aliyun.com/areas_v3/bound/100000_full.json"
	dataVProvinceGeoJSONURL = "https://geo.datav.aliyun.com/areas_v3/bound/%s_full.json"
)

func (ctrl *MapController) proxyDataVGeoJSON(c *gin.Context, geoURL string) {
	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, geoURL, nil)
	if err != nil {
		utils.Error(c, http.StatusBadGateway, "地图底图请求失败")
		return
	}
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		utils.Error(c, http.StatusBadGateway, "地图底图请求失败")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		utils.Error(c, http.StatusBadGateway, "地图底图数据加载失败")
		return
	}

	c.DataFromReader(http.StatusOK, resp.ContentLength, "application/json; charset=utf-8", resp.Body, map[string]string{
		"Cache-Control": "public, max-age=86400",
	})
}

// ProxyChinaGeoJSON 代理全国地图 GeoJSON，避免浏览器直连 DataV 资源被拦截
func (ctrl *MapController) ProxyChinaGeoJSON(c *gin.Context) {
	ctrl.proxyDataVGeoJSON(c, dataVChinaGeoJSONURL)
}

// ProxyProvinceGeoJSON 代理省份地图 GeoJSON，避免浏览器直连 DataV 资源被拦截
func (ctrl *MapController) ProxyProvinceGeoJSON(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	if len(code) != 6 {
		utils.Error(c, http.StatusBadRequest, "无效的地图编码")
		return
	}
	if _, err := strconv.Atoi(code); err != nil {
		utils.Error(c, http.StatusBadRequest, "无效的地图编码")
		return
	}

	ctrl.proxyDataVGeoJSON(c, fmt.Sprintf(dataVProvinceGeoJSONURL, code))
}

// NationalMapResponse 全国地图响应
type NationalMapResponse struct {
	Provinces []ProvinceAggregate `json:"provinces"`
}

// ProvinceAggregate 省份聚合数据
type ProvinceAggregate struct {
	ProvinceCode       string  `json:"provinceCode"`
	ProvinceName       string  `json:"provinceName"`
	PlayerCount        int64   `json:"playerCount"`
	ClubCount          int64   `json:"clubCount"`
	ScoutReportCount   int64   `json:"scoutReportCount"`
	AvgScore           float64 `json:"avgScore"`
	NewPlayerCount30d  int64   `json:"newPlayerCount30d"`
	ReportCoverageRate float64 `json:"reportCoverageRate"`
	HeatLevel          int     `json:"heatLevel"`
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
	ID          uint     `json:"id"`
	Name        string   `json:"name"`
	Avatar      string   `json:"avatar"`
	Position    string   `json:"position"`
	Age         int      `json:"age"`
	Score       float64  `json:"score"`
	Potential   string   `json:"potential"`
	Tags        []string `json:"tags"`
	NormalizedX float64  `json:"normalizedX"`
	NormalizedY float64  `json:"normalizedY"`
	HasReport   bool     `json:"hasReport"`
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

func calculatePotential(score float64) string {
	return utils.PlayerPotentialFromScore(score)
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

	scoreDetail := buildTraceablePlayerScore(db, &user)
	radar := buildMapRadarFromScore(scoreDetail)
	physical := buildMapPhysicalItemsFromScore(scoreDetail)
	reports := buildMapProfileReports(db, user.ID)

	var followers int64
	db.Model(&models.ScoutFollowPlayer{}).Where("user_id = ?", user.ID).Count(&followers)

	profile := gin.H{
		"id":              user.ID,
		"name":            user.Name,
		"nickname":        user.Nickname,
		"avatar":          user.Avatar,
		"city":            user.City,
		"province":        user.Province,
		"age":             user.Age,
		"position":        user.Position,
		"second_position": user.SecondPosition,
		"height":          user.Height,
		"weight":          user.Weight,
		"foot":            user.Foot,
		"club":            user.Club,
		"current_team":    user.CurrentTeam,
		"start_year":      user.StartYear,
		"jersey_number":   user.JerseyNumber,
		"jersey_color":    user.JerseyColor,
		"fa_registered":   user.FARegistered,
		"association":     user.Association,
		"school":          user.School,
		"tags":            technicalTags,
		"playing_style":   playingStyles,
		"technical_tags":  technicalTags,
		"mental_tags":     mentalTags,
		"experiences":     experiences,
		"score":           scoreDetail.Score,
		"potential":       scoreDetail.Potential,
		"scoreBreakdown":  scoreDetail,
		"heat": gin.H{
			"views7d":   0,
			"followers": followers,
		},
		"radar":    radar,
		"physical": physical,
		"timeline": buildMapTimeline(experiences),
		"reports":  reports,
		"permissions": gin.H{
			"canViewRadar":    scoreDetail.HasScore && len(radar["dimensions"].([]string)) > 0,
			"canViewPhysical": scoreDetail.HasScore && len(physical["items"].([]gin.H)) > 0,
			"canViewReports":  false,
			"canContact":      false,
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
		Period      string `json:"period"`
		Team        string `json:"team"`
		Position    string `json:"position"`
		Achievement string `json:"achievement,omitempty"`
	}
	if err := json.Unmarshal([]byte(s), &items); err != nil {
		return []gin.H{}
	}
	result := make([]gin.H, len(items))
	for i, item := range items {
		result[i] = gin.H{
			"date":        item.Period,
			"type":        "experience",
			"title":       item.Team,
			"summary":     item.Position,
			"achievement": item.Achievement,
		}
	}
	return result
}

type scoutReportScoreAverage struct {
	Average float64
	Count   int64
}

func buildTraceablePlayerScore(db *gorm.DB, user *models.User) utils.PlayerScoreResult {
	if user == nil {
		return utils.CalculatePlayerMapScore(utils.PlayerScoreInput{})
	}
	scores := buildTraceablePlayerScoreIndex(db, []models.User{*user})
	if score, ok := scores[user.ID]; ok {
		return score
	}
	return utils.CalculatePlayerMapScore(utils.PlayerScoreInput{User: user})
}

func buildTraceablePlayerScoreIndex(db *gorm.DB, users []models.User) map[uint]utils.PlayerScoreResult {
	result := make(map[uint]utils.PlayerScoreResult, len(users))
	if len(users) == 0 {
		return result
	}
	if db == nil {
		for i := range users {
			user := users[i]
			result[user.ID] = utils.CalculatePlayerMapScore(utils.PlayerScoreInput{User: &user})
		}
		return result
	}

	playerIDs := make([]uint, 0, len(users))
	for _, user := range users {
		playerIDs = append(playerIDs, user.ID)
	}
	physicalRecords := latestPhysicalRecordsByPlayer(db, playerIDs)
	reportAverages := scoutReportScoresByPlayer(db, playerIDs)

	for i := range users {
		user := users[i]
		reportScore := reportAverages[user.ID]
		var scoutAverage *float64
		if reportScore.Count > 0 {
			avg := reportScore.Average
			scoutAverage = &avg
		}
		result[user.ID] = utils.CalculatePlayerMapScore(utils.PlayerScoreInput{
			User:               &user,
			PhysicalRecord:     physicalRecords[user.ID],
			ScoutReportAverage: scoutAverage,
			ScoutReportCount:   reportScore.Count,
		})
	}
	return result
}

func latestPhysicalRecordsByPlayer(db *gorm.DB, playerIDs []uint) map[uint]*models.PhysicalTestRecord {
	result := make(map[uint]*models.PhysicalTestRecord, len(playerIDs))
	if len(playerIDs) == 0 {
		return result
	}
	var records []models.PhysicalTestRecord
	if err := db.Where("player_id IN ?", playerIDs).
		Order("player_id ASC, test_date DESC, created_at DESC, id DESC").
		Find(&records).Error; err != nil {
		return result
	}
	for _, record := range records {
		if _, ok := result[record.PlayerID]; ok {
			continue
		}
		recordCopy := record
		result[record.PlayerID] = &recordCopy
	}
	return result
}

func scoutReportScoresByPlayer(db *gorm.DB, playerIDs []uint) map[uint]scoutReportScoreAverage {
	result := make(map[uint]scoutReportScoreAverage, len(playerIDs))
	if len(playerIDs) == 0 {
		return result
	}
	reportPlayerIDsByUser := reportPlayerIDsByUserID(db, playerIDs)
	reportPlayerToUser := make(map[uint]uint, len(playerIDs)*2)
	reportPlayerIDs := make([]uint, 0, len(playerIDs)*2)
	for _, userID := range playerIDs {
		for _, reportPlayerID := range reportPlayerIDsByUser[userID] {
			reportPlayerToUser[reportPlayerID] = userID
			reportPlayerIDs = append(reportPlayerIDs, reportPlayerID)
		}
	}
	var rows []struct {
		PlayerID uint
		Average  float64
		Count    int64
	}
	db.Model(&models.ScoutReport{}).
		Select("player_id, AVG(overall_rating) AS average, COUNT(*) AS count").
		Where("player_id IN ? AND status IN ? AND overall_rating > ?", reportPlayerIDs, []string{"published", "adopted"}, 0).
		Group("player_id").
		Scan(&rows)
	for _, row := range rows {
		userID := reportPlayerToUser[row.PlayerID]
		if userID == 0 {
			continue
		}
		current := result[userID]
		total := current.Average*float64(current.Count) + row.Average*float64(row.Count)
		current.Count += row.Count
		current.Average = total / float64(current.Count)
		result[userID] = current
	}
	return result
}

func reportPlayerIDsByUserID(db *gorm.DB, userIDs []uint) map[uint][]uint {
	result := make(map[uint][]uint, len(userIDs))
	for _, userID := range userIDs {
		result[userID] = []uint{userID}
	}
	if len(userIDs) == 0 {
		return result
	}
	var rows []struct {
		ID     uint
		UserID uint
	}
	if err := db.Model(&models.Player{}).
		Select("id, user_id").
		Where("user_id IN ?", userIDs).
		Find(&rows).Error; err != nil {
		return result
	}
	for _, row := range rows {
		if row.ID == 0 || row.UserID == 0 {
			continue
		}
		if !uintSliceContains(result[row.UserID], row.ID) {
			result[row.UserID] = append(result[row.UserID], row.ID)
		}
	}
	return result
}

func uintSliceContains(values []uint, target uint) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func calculateScoreFromUser(user *models.User) float64 {
	return utils.CalculatePlayerMapScore(utils.PlayerScoreInput{User: user}).Score
}

func calculateMapPercentile(value float64, min, max float64) int {
	if value <= 0 {
		return 0
	}
	if value <= min {
		return 30
	}
	if value >= max {
		return 95
	}
	return int(30 + (value-min)/(max-min)*65)
}

func calculateLowerBetterMapPercentile(value float64, best, worst float64) int {
	if value <= 0 {
		return 0
	}
	if value <= best {
		return 95
	}
	if value >= worst {
		return 30
	}
	return int(95 - (value-best)/(worst-best)*65)
}

func buildMapRadarFromUser(user *models.User) gin.H {
	dimensions := []string{}
	values := []float64{}

	if user.Sprint30m > 0 {
		dimensions = append(dimensions, "速度")
		values = append(values, mapScoreLowerIsBetter(user.Sprint30m, 4.2, 6.5))
	}
	if user.StandingLongJump > 0 {
		dimensions = append(dimensions, "爆发")
		values = append(values, mapScoreHigherIsBetter(user.StandingLongJump, 120, 260))
	}
	if user.PushUp > 0 {
		dimensions = append(dimensions, "力量")
		values = append(values, mapScoreHigherIsBetter(float64(user.PushUp), 8, 45))
	}

	return gin.H{
		"visible":    len(dimensions) > 0,
		"dimensions": dimensions,
		"values":     values,
	}
}

func buildMapRadarFromScore(score utils.PlayerScoreResult) gin.H {
	dimensions := make([]string, 0, len(score.Metrics))
	values := make([]float64, 0, len(score.Metrics))
	for _, metric := range score.Metrics {
		dimensions = append(dimensions, metric.Label)
		values = append(values, metric.Score)
		if len(dimensions) >= 6 {
			break
		}
	}
	return gin.H{
		"visible":    len(dimensions) > 0,
		"dimensions": dimensions,
		"values":     values,
	}
}

func buildMapPhysicalItemsFromScore(score utils.PlayerScoreResult) gin.H {
	items := make([]gin.H, 0, len(score.Metrics))
	for _, metric := range score.Metrics {
		items = append(items, gin.H{
			"name":       metric.Label,
			"value":      metric.Value,
			"percentile": int(math.Round(metric.Score)),
		})
	}
	return gin.H{
		"visible": len(items) > 0,
		"items":   items,
	}
}

func buildMapPhysicalItemsFromUser(user *models.User) gin.H {
	items := []gin.H{}
	if user.Sprint30m > 0 {
		items = append(items, gin.H{
			"name":       "30m冲刺",
			"value":      fmt.Sprintf("%.1fs", user.Sprint30m),
			"percentile": calculateLowerBetterMapPercentile(user.Sprint30m, 4.2, 6.5),
		})
	}
	if user.StandingLongJump > 0 {
		items = append(items, gin.H{
			"name":       "立定跳远",
			"value":      fmt.Sprintf("%.0fcm", user.StandingLongJump),
			"percentile": calculateMapPercentile(user.StandingLongJump, 120, 260),
		})
	}
	if user.PushUp > 0 {
		items = append(items, gin.H{
			"name":       "俯卧撑",
			"value":      fmt.Sprintf("%d个", user.PushUp),
			"percentile": calculateMapPercentile(float64(user.PushUp), 8, 45),
		})
	}

	return gin.H{
		"visible": len(items) > 0,
		"items":   items,
	}
}

func mapScoreLowerIsBetter(value, best, worst float64) float64 {
	return roundMapScore(((worst - value) / (worst - best)) * 100)
}

func mapScoreHigherIsBetter(value, worst, best float64) float64 {
	return roundMapScore(((value - worst) / (best - worst)) * 100)
}

func roundMapScore(score float64) float64 {
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return math.Round(score*10) / 10
}

func buildMapProfileReports(db *gorm.DB, playerID uint) []gin.H {
	reportPlayerIDs := reportPlayerIDsByUserID(db, []uint{playerID})[playerID]
	if len(reportPlayerIDs) == 0 {
		reportPlayerIDs = []uint{playerID}
	}
	var reports []models.ScoutReport
	db.Preload("Scout.User").
		Where("player_id IN ? AND status IN ?", reportPlayerIDs, []string{"published", "adopted"}).
		Order("published_at DESC, created_at DESC").
		Limit(3).
		Find(&reports)

	items := make([]gin.H, 0, len(reports))
	for _, report := range reports {
		author := "球探报告"
		if report.Scout != nil && report.Scout.User != nil {
			author = report.Scout.User.Name
			if author == "" {
				author = report.Scout.User.Nickname
			}
		}
		if author == "" {
			author = "球探报告"
		}
		items = append(items, gin.H{
			"id":      report.ID,
			"type":    "scout",
			"author":  author,
			"score":   report.OverallRating,
			"summary": report.Summary,
		})
	}
	return items
}

func buildMapTimeline(experiences []gin.H) []gin.H {
	if len(experiences) == 0 {
		return []gin.H{}
	}
	return experiences
}

// DashboardStats 数据看板统计响应
type DashboardStats struct {
	TotalPlayers         int64   `json:"totalPlayers"`
	ScoredPlayerCount    int64   `json:"scoredPlayerCount"`
	TotalProvinces       int64   `json:"totalProvinces"`
	AvgAge               float64 `json:"avgAge"`
	AvgScore             float64 `json:"avgScore"`
	MonthlyNew           int64   `json:"monthlyNew"`
	RegionDistribution   []gin.H `json:"regionDistribution"`
	AgeDistribution      []gin.H `json:"ageDistribution"`
	PositionDistribution []gin.H `json:"positionDistribution"`
	ScoreRanking         []gin.H `json:"scoreRanking"`
	GrowthTrend          []gin.H `json:"growthTrend"`
}

type CityHotlistPlayer struct {
	ID        uint    `json:"id"`
	Name      string  `json:"name"`
	Avatar    string  `json:"avatar"`
	Position  string  `json:"position"`
	Age       int     `json:"age"`
	Score     float64 `json:"score"`
	Potential string  `json:"potential"`
	Province  string  `json:"province"`
	City      string  `json:"city"`
}

type CityHotlistClub struct {
	ID               uint   `json:"id"`
	Name             string `json:"name"`
	Logo             string `json:"logo"`
	ActivityCount    int64  `json:"activityCount"`
	ParticipantCount int64  `json:"participantCount"`
}

type CityHotlistActivity struct {
	ID                  uint      `json:"id"`
	Title               string    `json:"title"`
	Type                string    `json:"type"`
	ClubID              uint      `json:"clubId"`
	ClubName            string    `json:"clubName"`
	ClubLogo            string    `json:"clubLogo"`
	Province            string    `json:"province"`
	City                string    `json:"city"`
	StartTime           string    `json:"startTime"`
	StartTimeRaw        time.Time `json:"-"`
	CurrentParticipants int64     `json:"currentParticipants"`
}

type CityHotlistItem struct {
	Province           string               `json:"province"`
	City               string               `json:"city"`
	PlayerCount        int64                `json:"playerCount"`
	ClubCount          int64                `json:"clubCount"`
	AvgScore           float64              `json:"avgScore"`
	RisingStar         *CityHotlistPlayer   `json:"risingStar,omitempty"`
	ActiveClub         *CityHotlistClub     `json:"activeClub,omitempty"`
	HotActivity        *CityHotlistActivity `json:"hotActivity,omitempty"`
	ReportCoverageRate *float64             `json:"reportCoverageRate,omitempty"`
	CoverageGap        *float64             `json:"coverageGap,omitempty"`
	ActivityCount      int64                `json:"activityCount"`
	HotScore           float64              `json:"hotScore"`
}

type cityHotlistGroup struct {
	Province         string
	City             string
	PlayerCount      int64
	ClubCount        int64
	TotalScore       float64
	ScoredCount      int64
	ActivityCount    int64
	ParticipantCount int64
	ReportCovered    int64
	RisingStar       *CityHotlistPlayer
	ActiveClub       *CityHotlistClub
	ClubStats        map[uint]*CityHotlistClub
	HotActivity      *CityHotlistActivity
}

func normalizeMapScopeName(value string) string {
	value = strings.TrimSpace(value)
	for _, suffix := range []string{"特别行政区", "自治区", "省", "市"} {
		value = strings.TrimSuffix(value, suffix)
	}
	return value
}

func cityHotlistKey(province, city string) string {
	return normalizeMapScopeName(province) + "|" + normalizeMapScopeName(city)
}

func mapScopeLikeQuery(value string) string {
	value = normalizeMapScopeName(value)
	if value == "" {
		return ""
	}
	return "%" + value + "%"
}

func mapCoverageRateLabel(rate float64) float64 {
	if rate <= 0 {
		return 0
	}
	if rate > 1 {
		rate = rate / 100
	}
	return rate
}

func coveredMapUsersByScoutReports(db *gorm.DB, users []models.User) map[uint]bool {
	covered := make(map[uint]bool, len(users))
	if len(users) == 0 {
		return covered
	}

	userIDs := make([]uint, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}
	reportPlayerIDsByUser := reportPlayerIDsByUserID(db, userIDs)
	reportPlayerToUser := make(map[uint]uint, len(users)*2)
	reportPlayerIDs := make([]uint, 0, len(users)*2)
	for _, userID := range userIDs {
		for _, reportPlayerID := range reportPlayerIDsByUser[userID] {
			reportPlayerToUser[reportPlayerID] = userID
			reportPlayerIDs = append(reportPlayerIDs, reportPlayerID)
		}
	}

	var rows []struct {
		PlayerID uint
	}
	if err := db.Model(&models.ScoutReport{}).
		Select("DISTINCT player_id").
		Where("player_id IN ? AND status IN ?", reportPlayerIDs, []string{"published", "adopted"}).
		Find(&rows).Error; err != nil {
		return covered
	}

	for _, row := range rows {
		if userID, ok := reportPlayerToUser[row.PlayerID]; ok && userID > 0 {
			covered[userID] = true
		}
	}
	return covered
}

func parseCityHotlistScope(queryValue string, fallback string) string {
	if strings.TrimSpace(queryValue) != "" {
		return normalizeMapScopeName(queryValue)
	}
	return normalizeMapScopeName(fallback)
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
	var scoredPlayers int64
	var monthlyNew int64
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// 分布统计
	regionMap := make(map[string]int64)
	ageMap := make(map[string]int64)
	positionMap := make(map[string]int64)
	scoreLevelMap := make(map[string]int64)
	growthMap := make(map[string]int64)
	scoreIndex := buildTraceablePlayerScoreIndex(db, users)

	for _, u := range users {
		totalPlayers++
		if u.Province != "" {
			provinceMap[u.Province] = true
			regionMap[u.Province]++
		}
		totalAge += int64(u.Age)
		scoreDetail := scoreIndex[u.ID]
		if scoreDetail.HasScore {
			totalScore += scoreDetail.Score
			scoredPlayers++
		}

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
		level := scoreDetail.Potential
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
	for _, level := range []string{"S", "A", "B", "C", "D", "待评估"} {
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
	}
	if scoredPlayers > 0 {
		avgScore = math.Round(totalScore/float64(scoredPlayers)*10) / 10
	}

	stats := DashboardStats{
		TotalPlayers:         totalPlayers,
		ScoredPlayerCount:    scoredPlayers,
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

// GetCityHotlist 获取城市足球热榜
func (ctrl *MapController) GetCityHotlist(c *gin.Context) {
	db := config.GetDB()

	limit := 3
	if limitStr := strings.TrimSpace(c.Query("limit")); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			if parsed > 8 {
				parsed = 8
			}
			limit = parsed
		}
	}

	scopeProvince := parseCityHotlistScope(c.Query("province"), "")
	scopeCity := parseCityHotlistScope(c.Query("city"), "")

	var users []models.User
	userQuery := db.Where("role = ? AND status = ? AND city <> ''", "user", "active")
	if scopeProvince != "" {
		userQuery = userQuery.Where("province LIKE ?", mapScopeLikeQuery(scopeProvince))
	}
	if scopeCity != "" {
		userQuery = userQuery.Where("city LIKE ?", mapScopeLikeQuery(scopeCity))
	}
	if err := userQuery.Find(&users).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询球员数据失败")
		return
	}

	var clubs []models.Club
	clubQuery := db.Where("city <> '' AND deleted_at IS NULL")
	if scopeProvince != "" {
		clubQuery = clubQuery.Where("province LIKE ?", mapScopeLikeQuery(scopeProvince))
	}
	if scopeCity != "" {
		clubQuery = clubQuery.Where("city LIKE ?", mapScopeLikeQuery(scopeCity))
	}
	if err := clubQuery.Find(&clubs).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询俱乐部数据失败")
		return
	}

	var activities []models.ClubActivity
	activityQuery := db.Where("publish_status = ? AND status IN ?", "published", []string{"upcoming", "ongoing"})
	activityQuery = applyActivityTimeRange(c, activityQuery, 30)
	if err := activityQuery.Order("start_time ASC").Find(&activities).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询活动数据失败")
		return
	}

	type activityMeta struct {
		activity            models.ClubActivity
		province            string
		city                string
		clubName            string
		clubLogo            string
		currentParticipants int64
	}

	clubMetaMap := make(map[uint]struct{ name, logo string }, len(clubs))
	for _, club := range clubs {
		clubMetaMap[club.ID] = struct{ name, logo string }{name: club.Name, logo: club.Logo}
	}

	registrationCounts := make(map[uint]int64, len(activities))
	activityIDs := make([]uint, 0, len(activities))
	for _, activity := range activities {
		activityIDs = append(activityIDs, activity.ID)
	}
	if len(activityIDs) > 0 {
		var rows []struct {
			ActivityID uint
			Count      int64
		}
		if err := db.Model(&models.ClubActivityRegistration{}).
			Select("activity_id, COUNT(*) as count").
			Where("activity_id IN ? AND status IN ?", activityIDs, activeActivityRegistrationStatuses).
			Group("activity_id").
			Scan(&rows).Error; err == nil {
			for _, row := range rows {
				registrationCounts[row.ActivityID] = row.Count
			}
		}
	}

	groups := make(map[string]*cityHotlistGroup)
	ensureGroup := func(province, city string) *cityHotlistGroup {
		key := cityHotlistKey(province, city)
		if groups[key] == nil {
			groups[key] = &cityHotlistGroup{
				Province:  strings.TrimSpace(province),
				City:      strings.TrimSpace(city),
				ClubStats: make(map[uint]*CityHotlistClub),
			}
		}
		return groups[key]
	}

	scoreIndex := buildTraceablePlayerScoreIndex(db, users)
	coveredUsers := coveredMapUsersByScoutReports(db, users)
	starCandidates := make([]models.User, len(users))
	copy(starCandidates, users)
	sort.SliceStable(starCandidates, func(i, j int) bool {
		left := scoreIndex[starCandidates[i].ID]
		right := scoreIndex[starCandidates[j].ID]
		if left.HasScore != right.HasScore {
			return left.HasScore
		}
		if left.Score != right.Score {
			return left.Score > right.Score
		}
		if starCandidates[i].UpdatedAt.Equal(starCandidates[j].UpdatedAt) {
			return starCandidates[i].ID < starCandidates[j].ID
		}
		return starCandidates[i].UpdatedAt.After(starCandidates[j].UpdatedAt)
	})

	for _, user := range users {
		if user.City == "" {
			continue
		}
		group := ensureGroup(user.Province, user.City)
		group.PlayerCount++
		if scoreDetail, ok := scoreIndex[user.ID]; ok {
			group.TotalScore += scoreDetail.Score
			if scoreDetail.HasScore {
				group.ScoredCount++
			}
		}
		if coveredUsers[user.ID] {
			group.ReportCovered++
		}
	}

	for _, user := range starCandidates {
		if user.City == "" {
			continue
		}
		group := ensureGroup(user.Province, user.City)
		if group.RisingStar != nil {
			continue
		}
		scoreDetail := scoreIndex[user.ID]
		group.RisingStar = &CityHotlistPlayer{
			ID:        user.ID,
			Name:      user.Name,
			Avatar:    user.Avatar,
			Position:  user.Position,
			Age:       user.Age,
			Score:     scoreDetail.Score,
			Potential: scoreDetail.Potential,
			Province:  user.Province,
			City:      user.City,
		}
	}

	for _, club := range clubs {
		if club.City == "" {
			continue
		}
		group := ensureGroup(club.Province, club.City)
		group.ClubCount++
	}

	for _, a := range activities {
		province, city, _ := parseLocation(a.Location)
		if city == "" {
			continue
		}
		if scopeProvince != "" && !strings.Contains(normalizeMapScopeName(province), scopeProvince) {
			continue
		}
		if scopeCity != "" && !strings.Contains(normalizeMapScopeName(city), scopeCity) {
			continue
		}
		group := ensureGroup(province, city)
		regCount := registrationCounts[a.ID]
		clubMeta, ok := clubMetaMap[a.ClubID]
		clubName := ""
		clubLogo := ""
		if ok {
			clubName = clubMeta.name
			clubLogo = clubMeta.logo
		}
		activity := &CityHotlistActivity{
			ID:                  a.ID,
			Title:               a.Title,
			Type:                mapActivityType(a.Type),
			ClubID:              a.ClubID,
			ClubName:            clubName,
			ClubLogo:            clubLogo,
			Province:            province,
			City:                city,
			StartTime:           a.StartTime.Format("2006-01-02 15:04"),
			StartTimeRaw:        a.StartTime,
			CurrentParticipants: regCount,
		}
		group.ActivityCount++
		group.ParticipantCount += regCount
		clubStat := group.ClubStats[a.ClubID]
		if clubStat == nil {
			clubStat = &CityHotlistClub{
				ID:   a.ClubID,
				Name: clubName,
				Logo: clubLogo,
			}
			group.ClubStats[a.ClubID] = clubStat
		}
		clubStat.ActivityCount++
		clubStat.ParticipantCount += regCount
		if group.HotActivity == nil || regCount > group.HotActivity.CurrentParticipants || (regCount == group.HotActivity.CurrentParticipants && a.StartTime.Before(group.HotActivity.StartTimeRaw)) {
			group.HotActivity = activity
		}
	}

	items := make([]CityHotlistItem, 0, len(groups))
	for _, group := range groups {
		if group.PlayerCount == 0 && group.ActivityCount == 0 && group.ClubCount == 0 {
			continue
		}
		var avgScore float64
		if group.ScoredCount > 0 {
			avgScore = math.Round((group.TotalScore/float64(group.ScoredCount))*10) / 10
		}
		var activeClub *CityHotlistClub
		for _, club := range group.ClubStats {
			if club == nil {
				continue
			}
			if activeClub == nil || club.ActivityCount > activeClub.ActivityCount || (club.ActivityCount == activeClub.ActivityCount && club.ParticipantCount > activeClub.ParticipantCount) || (club.ActivityCount == activeClub.ActivityCount && club.ParticipantCount == activeClub.ParticipantCount && club.ID < activeClub.ID) {
				cloned := *club
				activeClub = &cloned
			}
		}
		var coverageRate *float64
		var coverageGap *float64
		if group.PlayerCount > 0 {
			rate := math.Round((float64(group.ReportCovered)/float64(group.PlayerCount))*1000) / 10
			gap := math.Round((100-rate)*10) / 10
			coverageRate = &rate
			coverageGap = &gap
		}
		hotScore := float64(group.PlayerCount)*1.5 + float64(group.ActivityCount)*12 + float64(group.ParticipantCount)*0.8
		if coverageGap != nil {
			hotScore += *coverageGap
		}
		items = append(items, CityHotlistItem{
			Province:           group.Province,
			City:               group.City,
			PlayerCount:        group.PlayerCount,
			ClubCount:          group.ClubCount,
			AvgScore:           avgScore,
			RisingStar:         group.RisingStar,
			ActiveClub:         activeClub,
			HotActivity:        group.HotActivity,
			ReportCoverageRate: coverageRate,
			CoverageGap:        coverageGap,
			ActivityCount:      group.ActivityCount,
			HotScore:           math.Round(hotScore*10) / 10,
		})
	}

	sort.SliceStable(items, func(i, j int) bool {
		if scopeCity != "" {
			iSelected := normalizeMapScopeName(items[i].City) == scopeCity && (scopeProvince == "" || normalizeMapScopeName(items[i].Province) == scopeProvince)
			jSelected := normalizeMapScopeName(items[j].City) == scopeCity && (scopeProvince == "" || normalizeMapScopeName(items[j].Province) == scopeProvince)
			if iSelected != jSelected {
				return iSelected
			}
		}
		if items[i].HotScore != items[j].HotScore {
			return items[i].HotScore > items[j].HotScore
		}
		if items[i].ActivityCount != items[j].ActivityCount {
			return items[i].ActivityCount > items[j].ActivityCount
		}
		if items[i].PlayerCount != items[j].PlayerCount {
			return items[i].PlayerCount > items[j].PlayerCount
		}
		if items[i].City == items[j].City {
			return items[i].Province < items[j].Province
		}
		return items[i].City < items[j].City
	})

	if len(items) > limit {
		items = items[:limit]
	}

	utils.Success(c, "", gin.H{
		"items":     items,
		"updatedAt": time.Now().Format(time.RFC3339),
		"province":  scopeProvince,
		"city":      scopeCity,
		"limit":     limit,
	})
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

	players := make([]gin.H, 0, len(users))
	scoreIndex := buildTraceablePlayerScoreIndex(db, users)
	for _, u := range users {
		scoreDetail := scoreIndex[u.ID]
		country := u.Country
		if country == "" {
			country = "海外"
		}
		players = append(players, gin.H{
			"id":             u.ID,
			"name":           u.Name,
			"avatar":         u.Avatar,
			"country":        country,
			"city":           u.City,
			"position":       u.Position,
			"age":            u.Age,
			"score":          scoreDetail.Score,
			"potential":      scoreDetail.Potential,
			"tags":           utils.BuildPlayerTags(&u, 4),
			"scoreBreakdown": scoreDetail,
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

	scoreDetail := buildTraceablePlayerScore(db, &me)

	// 全省排名（同年龄组，按可追溯评分排序）
	var provinceRank int64 = 1
	if me.Province != "" {
		var peers []models.User
		db.Where("role = ? AND status = ? AND province = ? AND age = ?", "user", "active", me.Province, me.Age).Find(&peers)
		provinceRank = rankPlayerInScope(db, peers, me.ID)
	}

	// 全市排名
	var cityRank int64 = 1
	if me.City != "" {
		var peers []models.User
		db.Where("role = ? AND status = ? AND province = ? AND city = ? AND age = ?", "user", "active", me.Province, me.City, me.Age).Find(&peers)
		cityRank = rankPlayerInScope(db, peers, me.ID)
	}

	// 全国同位置排名
	var positionRank int64 = 1
	if me.Position != "" {
		var peers []models.User
		db.Where("role = ? AND status = ? AND position = ? AND age = ?", "user", "active", me.Position, me.Age).Find(&peers)
		positionRank = rankPlayerInScope(db, peers, me.ID)
	}

	utils.Success(c, "", gin.H{
		"player": gin.H{
			"id":             me.ID,
			"name":           me.Name,
			"province":       me.Province,
			"city":           me.City,
			"position":       me.Position,
			"age":            me.Age,
			"score":          scoreDetail.Score,
			"potential":      scoreDetail.Potential,
			"scoreBreakdown": scoreDetail,
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
	ID             uint                    `json:"id"`
	Name           string                  `json:"name"`
	Avatar         string                  `json:"avatar"`
	Position       string                  `json:"position"`
	Age            int                     `json:"age"`
	City           string                  `json:"city"`
	Province       string                  `json:"province"`
	Score          float64                 `json:"score"`
	Potential      string                  `json:"potential"`
	Tags           []string                `json:"tags"`
	Reason         string                  `json:"reason"`
	ScoreBreakdown utils.PlayerScoreResult `json:"scoreBreakdown"`
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
		user      models.User
		score     utils.PlayerScoreResult
		rankScore float64
	}
	scoredList := make([]scored, 0, len(candidates))
	scoreIndex := buildTraceablePlayerScoreIndex(db, candidates)
	for _, u := range candidates {
		scoreDetail := scoreIndex[u.ID]
		scoredList = append(scoredList, scored{
			user:      u,
			score:     scoreDetail,
			rankScore: recommendationRankScore(u, scoreDetail, me, isPlayer, userId > 0),
		})
	}

	// 推荐排序逻辑
	var recommendations []RecommendPlayer
	maxResults := 6

	sort.SliceStable(scoredList, func(i, j int) bool {
		if scoredList[i].rankScore != scoredList[j].rankScore {
			return scoredList[i].rankScore > scoredList[j].rankScore
		}
		if scoredList[i].score.Score != scoredList[j].score.Score {
			return scoredList[i].score.Score > scoredList[j].score.Score
		}
		if scoredList[i].user.UpdatedAt.Equal(scoredList[j].user.UpdatedAt) {
			return scoredList[i].user.ID < scoredList[j].user.ID
		}
		return scoredList[i].user.UpdatedAt.After(scoredList[j].user.UpdatedAt)
	})

	for i := 0; i < len(scoredList) && len(recommendations) < maxResults; i++ {
		u := scoredList[i].user
		scoreDetail := scoredList[i].score
		reason := scoreReason(scoreDetail)
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
			if scoreDetail.Score >= 85 {
				reason = scoreReason(scoreDetail)
			} else if scoreDetail.Potential == "A" || scoreDetail.Potential == "S" {
				reason = "高潜力球员"
			}
		} else {
			if scoreDetail.Score >= 85 {
				reason = scoreReason(scoreDetail)
			} else if scoreDetail.Potential == "A" || scoreDetail.Potential == "S" {
				reason = "高潜力新星"
			}
		}
		recommendations = append(recommendations, RecommendPlayer{
			ID:             u.ID,
			Name:           u.Name,
			Avatar:         u.Avatar,
			Position:       u.Position,
			Age:            u.Age,
			City:           u.City,
			Province:       u.Province,
			Score:          scoreDetail.Score,
			Potential:      scoreDetail.Potential,
			Tags:           utils.BuildPlayerTags(&u, 4),
			Reason:         reason,
			ScoreBreakdown: scoreDetail,
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

func rankPlayerInScope(db *gorm.DB, players []models.User, targetID uint) int64 {
	if len(players) == 0 {
		return 1
	}
	scoreIndex := buildTraceablePlayerScoreIndex(db, players)
	sort.SliceStable(players, func(i, j int) bool {
		left := scoreIndex[players[i].ID]
		right := scoreIndex[players[j].ID]
		if left.HasScore != right.HasScore {
			return left.HasScore
		}
		if left.Score != right.Score {
			return left.Score > right.Score
		}
		if players[i].UpdatedAt.Equal(players[j].UpdatedAt) {
			return players[i].ID < players[j].ID
		}
		return players[i].UpdatedAt.After(players[j].UpdatedAt)
	})
	for i, player := range players {
		if player.ID == targetID {
			return int64(i + 1)
		}
	}
	return 1
}

func recommendationRankScore(candidate models.User, score utils.PlayerScoreResult, me models.User, isPlayer bool, isLoggedIn bool) float64 {
	rank := score.Score
	if !score.HasScore {
		rank = 0
	}
	if isPlayer && me.ID > 0 {
		if candidate.Position == me.Position && me.Position != "" {
			rank += 25
		}
		if candidate.City == me.City && me.City != "" {
			rank += 18
		}
		if candidate.Province == me.Province && me.Province != "" {
			rank += 10
		}
		rank -= float64(abs(candidate.Age-me.Age)) * 2
		return rank
	}
	if isLoggedIn {
		if score.HasScore {
			return rank + float64(score.DataCoverage)*0.1
		}
		return rank
	}
	if score.HasScore {
		return rank
	}
	return float64(candidate.UpdatedAt.Unix()%1000) / 1000
}

func scoreReason(score utils.PlayerScoreResult) string {
	if !score.HasScore {
		return "资料待补充"
	}
	hasPhysical := false
	hasScout := false
	for _, source := range score.Sources {
		if source == "latest_physical_test" {
			hasPhysical = true
		}
		if source == "published_scout_reports" {
			hasScout = true
		}
	}
	switch {
	case hasPhysical && hasScout:
		return "体测与球探报告综合靠前"
	case hasPhysical:
		return "近期体测表现靠前"
	case hasScout:
		return "球探报告评分靠前"
	default:
		return "真实资料评分靠前"
	}
}

// GetRisingStars 获取本周新星
func (ctrl *MapController) GetRisingStars(c *gin.Context) {
	db := config.GetDB()

	// 取最近7天注册或更新的活跃球员，按真实评分与最近活跃时间排序
	var users []models.User
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	if err := db.Where("role = ? AND status = ? AND (created_at >= ? OR updated_at >= ?)", "user", "active", sevenDaysAgo, sevenDaysAgo).
		Order("created_at DESC").
		Limit(20).
		Find(&users).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}

	// 如果数量不足，补充历史球员，仍按真实评分排序
	if len(users) < 5 {
		var extraUsers []models.User
		db.Where("role = ? AND status = ?", "user", "active").
			Order("updated_at DESC, created_at DESC").
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

	scoreIndex := buildTraceablePlayerScoreIndex(db, uniqueUsers)
	sort.SliceStable(uniqueUsers, func(i, j int) bool {
		left := scoreIndex[uniqueUsers[i].ID]
		right := scoreIndex[uniqueUsers[j].ID]
		if left.HasScore != right.HasScore {
			return left.HasScore
		}
		if left.Score != right.Score {
			return left.Score > right.Score
		}
		if uniqueUsers[i].UpdatedAt.Equal(uniqueUsers[j].UpdatedAt) {
			return uniqueUsers[i].ID < uniqueUsers[j].ID
		}
		return uniqueUsers[i].UpdatedAt.After(uniqueUsers[j].UpdatedAt)
	})

	// 取前8位
	if len(uniqueUsers) > 8 {
		uniqueUsers = uniqueUsers[:8]
	}

	var players []gin.H
	for _, u := range uniqueUsers {
		scoreDetail := scoreIndex[u.ID]
		players = append(players, gin.H{
			"id":             u.ID,
			"name":           u.Name,
			"avatar":         u.Avatar,
			"province":       u.Province,
			"city":           u.City,
			"position":       u.Position,
			"age":            u.Age,
			"score":          scoreDetail.Score,
			"potential":      scoreDetail.Potential,
			"tags":           utils.BuildPlayerTags(&u, 4),
			"reason":         scoreReason(scoreDetail),
			"scoreBreakdown": scoreDetail,
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
	csvBuilder.WriteString("\uFEFF姓名,年龄,位置,城市,俱乐部,综合评分,潜力,评分来源\n")
	scoreIndex := buildTraceablePlayerScoreIndex(db, users)
	for _, u := range users {
		scoreDetail := scoreIndex[u.ID]
		csvBuilder.WriteString(fmt.Sprintf("%s,%d,%s,%s,%s,%.1f,%s,%s\n", u.Name, u.Age, u.Position, u.City, u.Club, scoreDetail.Score, scoreDetail.Potential, strings.Join(scoreDetail.Sources, "+")))
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=compare.csv")
	c.String(http.StatusOK, csvBuilder.String())
}
