package repositories

import (
	"gorm.io/gorm"
	"github.com/shaonianqiutan/backend/utils"
)

// EntityLayer 实体图层类型
type EntityLayer string

const (
	LayerPlayers  EntityLayer = "players"
	LayerClubs    EntityLayer = "clubs"
	LayerCoaches  EntityLayer = "coaches"
	LayerAnalysts EntityLayer = "analysts"
	LayerScouts   EntityLayer = "scouts"
	LayerAll      EntityLayer = "all"
)

// ParseEntityLayer 解析图层参数，默认返回 players
func ParseEntityLayer(s string) EntityLayer {
	switch s {
	case "clubs":
		return LayerClubs
	case "coaches":
		return LayerCoaches
	case "analysts":
		return LayerAnalysts
	case "scouts":
		return LayerScouts
	case "all":
		return LayerAll
	default:
		return LayerPlayers
	}
}

// IsValid 验证图层类型
func (l EntityLayer) IsValid() bool {
	switch l {
	case LayerPlayers, LayerClubs, LayerCoaches, LayerAnalysts, LayerScouts, LayerAll:
		return true
	}
	return false
}

// NationalAggregate 全国聚合（统一格式）
type NationalAggregate struct {
	ProvinceCode          string            `json:"provinceCode"`
	ProvinceName          string            `json:"provinceName"`
	Count                 int64             `json:"count"`
	PlayerCount           int64             `json:"playerCount,omitempty"`
	ClubCount             int64             `json:"clubCount,omitempty"`
	CoachCount            int64             `json:"coachCount,omitempty"`
	AnalystCount          int64             `json:"analystCount,omitempty"`
	ScoutCount            int64             `json:"scoutCount,omitempty"`
	HeatLevel             int               `json:"heatLevel"`
	SizeDistribution      map[string]int64  `json:"sizeDistribution,omitempty"`      // P2-14: 俱乐部规模分布
	LicenseDistribution   map[string]int64  `json:"licenseDistribution,omitempty"`   // P2-15: 教练执照分布
	SpecialtyDistribution map[string]int64  `json:"specialtyDistribution,omitempty"` // P2-16: 分析师擅长领域
	AdoptionRate          float64           `json:"adoptionRate,omitempty"`          // P2-17: 球探采纳率
}

// ProvincialAggregate 省市聚合
type ProvincialAggregate struct {
	CityCode              string            `json:"cityCode"`
	CityName              string            `json:"cityName"`
	Count                 int64             `json:"count"`
	PlayerCount           int64             `json:"playerCount,omitempty"`
	ClubCount             int64             `json:"clubCount,omitempty"`
	CoachCount            int64             `json:"coachCount,omitempty"`
	AnalystCount          int64             `json:"analystCount,omitempty"`
	ScoutCount            int64             `json:"scoutCount,omitempty"`
	HeatLevel             int               `json:"heatLevel"`
	SizeDistribution      map[string]int64  `json:"sizeDistribution,omitempty"`      // P2-14
	LicenseDistribution   map[string]int64  `json:"licenseDistribution,omitempty"`   // P2-15
	SpecialtyDistribution map[string]int64  `json:"specialtyDistribution,omitempty"` // P2-16
	AdoptionRate          float64           `json:"adoptionRate,omitempty"`          // P2-17
}

// CityEntityItem 城市散点项
type CityEntityItem struct {
	ID          uint                   `json:"id"`
	Name        string                 `json:"name"`
	Avatar      string                 `json:"avatar"`
	Type        string                 `json:"type"` // player/club/coach/analyst/scout
	Score       float64                `json:"score"`
	Tags        []string               `json:"tags"`
	NormalizedX float64                `json:"normalizedX"`
	NormalizedY float64                `json:"normalizedY"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
}

// MultiLayerMapRepository 多图层仓库
type MultiLayerMapRepository struct {
	db *gorm.DB
}

// NewMultiLayerMapRepository 创建仓库
func NewMultiLayerMapRepository(db *gorm.DB) *MultiLayerMapRepository {
	return &MultiLayerMapRepository{db: db}
}

// ===== National =====

// GetNationalAggregates 全国聚合分发
func (r *MultiLayerMapRepository) GetNationalAggregates(layer EntityLayer) ([]NationalAggregate, error) {
	if layer == LayerAll {
		return r.queryNationalAll()
	}
	sqlMap := map[EntityLayer]string{
		LayerPlayers: `SELECT province as p, COUNT(*) as c FROM users WHERE role='user' AND status='active' AND province!='' GROUP BY province`,
		LayerClubs:   `SELECT province as p, COUNT(*) as c FROM clubs WHERE province!='' AND deleted_at IS NULL GROUP BY province`,
		LayerCoaches: `SELECT u.province as p, COUNT(*) as c FROM coaches c JOIN users u ON c.user_id=u.id WHERE u.province!='' AND c.deleted_at IS NULL AND u.deleted_at IS NULL GROUP BY u.province`,
		LayerAnalysts:`SELECT u.province as p, COUNT(*) as c FROM analysts a JOIN users u ON a.user_id=u.id WHERE u.province!='' AND a.deleted_at IS NULL AND u.deleted_at IS NULL GROUP BY u.province`,
		LayerScouts:  `SELECT u.province as p, COUNT(*) as c FROM scouts s JOIN users u ON s.user_id=u.id WHERE u.province!='' AND s.deleted_at IS NULL AND u.deleted_at IS NULL GROUP BY u.province`,
	}
	aggregates, err := r.queryNational(sqlMap[layer])
	if err != nil {
		return nil, err
	}
	// 补充分布数据（P2-14~P2-17）
	switch layer {
	case LayerClubs:
		r.fillSizeDistributionForNational(aggregates)
	case LayerCoaches:
		r.fillLicenseDistributionForNational(aggregates)
	case LayerAnalysts:
		r.fillSpecialtyDistributionForNational(aggregates)
	case LayerScouts:
		r.fillAdoptionRateForNational(aggregates)
	}
	return aggregates, nil
}

func (r *MultiLayerMapRepository) queryNational(sql string) ([]NationalAggregate, error) {
	var rows []struct{ P string; C int64 }
	if err := r.db.Raw(sql).Scan(&rows).Error; err != nil {
		return nil, err
	}
	res := make([]NationalAggregate, 0, len(rows))
	for _, v := range rows {
		res = append(res, NationalAggregate{
			ProvinceCode: utils.GetProvinceCode(v.P),
			ProvinceName: v.P,
			Count:        v.C,
			HeatLevel:    utils.CalculateHeatLevel(v.C),
		})
	}
	return res, nil
}

// queryNationalAll 全国全部实体拆分查询（ stacked bar / mixed scatter 数据源）
func (r *MultiLayerMapRepository) queryNationalAll() ([]NationalAggregate, error) {
	sql := `SELECT 
		p.province as p,
		COALESCE(u.c, 0) as player_count,
		COALESCE(cl.c, 0) as club_count,
		COALESCE(co.c, 0) as coach_count,
		COALESCE(a.c, 0) as analyst_count,
		COALESCE(s.c, 0) as scout_count
	FROM (
		SELECT DISTINCT province FROM users WHERE province!='' UNION
		SELECT DISTINCT province FROM clubs WHERE province!='' AND deleted_at IS NULL UNION
		SELECT DISTINCT u.province FROM coaches c JOIN users u ON c.user_id=u.id WHERE u.province!='' AND c.deleted_at IS NULL AND u.deleted_at IS NULL UNION
		SELECT DISTINCT u.province FROM analysts a JOIN users u ON a.user_id=u.id WHERE u.province!='' AND a.deleted_at IS NULL AND u.deleted_at IS NULL UNION
		SELECT DISTINCT u.province FROM scouts s JOIN users u ON s.user_id=u.id WHERE u.province!='' AND s.deleted_at IS NULL AND u.deleted_at IS NULL
	) p
	LEFT JOIN (SELECT province, COUNT(*) as c FROM users WHERE role='user' AND status='active' AND province!='' GROUP BY province) u ON p.province=u.province
	LEFT JOIN (SELECT province, COUNT(*) as c FROM clubs WHERE province!='' AND deleted_at IS NULL GROUP BY province) cl ON p.province=cl.province
	LEFT JOIN (SELECT u.province, COUNT(*) as c FROM coaches c JOIN users u ON c.user_id=u.id WHERE u.province!='' AND c.deleted_at IS NULL AND u.deleted_at IS NULL GROUP BY u.province) co ON p.province=co.province
	LEFT JOIN (SELECT u.province, COUNT(*) as c FROM analysts a JOIN users u ON a.user_id=u.id WHERE u.province!='' AND a.deleted_at IS NULL AND u.deleted_at IS NULL GROUP BY u.province) a ON p.province=a.province
	LEFT JOIN (SELECT u.province, COUNT(*) as c FROM scouts s JOIN users u ON s.user_id=u.id WHERE u.province!='' AND s.deleted_at IS NULL AND u.deleted_at IS NULL GROUP BY u.province) s ON p.province=s.province`

	var rows []struct {
		P            string
		PlayerCount  int64
		ClubCount    int64
		CoachCount   int64
		AnalystCount int64
		ScoutCount   int64
	}
	if err := r.db.Raw(sql).Scan(&rows).Error; err != nil {
		return nil, err
	}
	res := make([]NationalAggregate, 0, len(rows))
	for _, v := range rows {
		total := v.PlayerCount + v.ClubCount + v.CoachCount + v.AnalystCount + v.ScoutCount
		res = append(res, NationalAggregate{
			ProvinceCode: utils.GetProvinceCode(v.P),
			ProvinceName: v.P,
			Count:        total,
			PlayerCount:  v.PlayerCount,
			ClubCount:    v.ClubCount,
			CoachCount:   v.CoachCount,
			AnalystCount: v.AnalystCount,
			ScoutCount:   v.ScoutCount,
			HeatLevel:    utils.CalculateHeatLevel(total),
		})
	}
	return res, nil
}

// ===== Provincial =====

// GetProvincialAggregates 省市聚合分发
func (r *MultiLayerMapRepository) GetProvincialAggregates(province string, layer EntityLayer) ([]ProvincialAggregate, error) {
	if layer == LayerAll {
		return r.queryProvincialAll(province)
	}
	sqlMap := map[EntityLayer]string{
		LayerPlayers:  `SELECT city as c, COUNT(*) as n FROM users WHERE role='user' AND status='active' AND province=? AND city!='' GROUP BY city`,
		LayerClubs:    `SELECT city as c, COUNT(*) as n FROM clubs WHERE province=? AND city!='' AND deleted_at IS NULL GROUP BY city`,
		LayerCoaches:  `SELECT u.city as c, COUNT(*) as n FROM coaches c JOIN users u ON c.user_id=u.id WHERE u.province=? AND u.city!='' AND c.deleted_at IS NULL AND u.deleted_at IS NULL GROUP BY u.city`,
		LayerAnalysts: `SELECT u.city as c, COUNT(*) as n FROM analysts a JOIN users u ON a.user_id=u.id WHERE u.province=? AND u.city!='' AND a.deleted_at IS NULL AND u.deleted_at IS NULL GROUP BY u.city`,
		LayerScouts:   `SELECT u.city as c, COUNT(*) as n FROM scouts s JOIN users u ON s.user_id=u.id WHERE u.province=? AND u.city!='' AND s.deleted_at IS NULL AND u.deleted_at IS NULL GROUP BY u.city`,
	}
	aggregates, err := r.queryProvincial(sqlMap[layer], province)
	if err != nil {
		return nil, err
	}
	// 补充分布数据（P2-14~P2-17）
	switch layer {
	case LayerClubs:
		r.fillSizeDistributionForProvincial(aggregates, province)
	case LayerCoaches:
		r.fillLicenseDistributionForProvincial(aggregates, province)
	case LayerAnalysts:
		r.fillSpecialtyDistributionForProvincial(aggregates, province)
	case LayerScouts:
		r.fillAdoptionRateForProvincial(aggregates, province)
	}
	return aggregates, nil
}

func (r *MultiLayerMapRepository) queryProvincial(sql, province string) ([]ProvincialAggregate, error) {
	var rows []struct{ C string; N int64 }
	if sql == "" {
		return []ProvincialAggregate{}, nil
	}
	if err := r.db.Raw(sql, province).Scan(&rows).Error; err != nil {
		return nil, err
	}
	res := make([]ProvincialAggregate, 0, len(rows))
	for _, v := range rows {
		res = append(res, ProvincialAggregate{
			CityCode:  utils.GetCityCode(v.C),
			CityName:  v.C,
			Count:     v.N,
			HeatLevel: utils.CalculateHeatLevel(v.N),
		})
	}
	return res, nil
}

// queryProvincialAll 省市全部实体拆分查询
func (r *MultiLayerMapRepository) queryProvincialAll(province string) ([]ProvincialAggregate, error) {
	sql := `SELECT 
		p.city as c,
		COALESCE(u.n, 0) as player_count,
		COALESCE(cl.n, 0) as club_count,
		COALESCE(co.n, 0) as coach_count,
		COALESCE(a.n, 0) as analyst_count,
		COALESCE(s.n, 0) as scout_count
	FROM (
		SELECT DISTINCT city FROM users WHERE province=? AND city!='' UNION
		SELECT DISTINCT city FROM clubs WHERE province=? AND city!='' AND deleted_at IS NULL UNION
		SELECT DISTINCT u.city FROM coaches c JOIN users u ON c.user_id=u.id WHERE u.province=? AND u.city!='' AND c.deleted_at IS NULL AND u.deleted_at IS NULL UNION
		SELECT DISTINCT u.city FROM analysts a JOIN users u ON a.user_id=u.id WHERE u.province=? AND u.city!='' AND a.deleted_at IS NULL AND u.deleted_at IS NULL UNION
		SELECT DISTINCT u.city FROM scouts s JOIN users u ON s.user_id=u.id WHERE u.province=? AND u.city!='' AND s.deleted_at IS NULL AND u.deleted_at IS NULL
	) p
	LEFT JOIN (SELECT city, COUNT(*) as n FROM users WHERE role='user' AND status='active' AND province=? AND city!='' GROUP BY city) u ON p.city=u.city
	LEFT JOIN (SELECT city, COUNT(*) as n FROM clubs WHERE province=? AND city!='' AND deleted_at IS NULL GROUP BY city) cl ON p.city=cl.city
	LEFT JOIN (SELECT u.city, COUNT(*) as n FROM coaches c JOIN users u ON c.user_id=u.id WHERE u.province=? AND u.city!='' AND c.deleted_at IS NULL AND u.deleted_at IS NULL GROUP BY u.city) co ON p.city=co.city
	LEFT JOIN (SELECT u.city, COUNT(*) as n FROM analysts a JOIN users u ON a.user_id=u.id WHERE u.province=? AND u.city!='' AND a.deleted_at IS NULL AND u.deleted_at IS NULL GROUP BY u.city) a ON p.city=a.city
	LEFT JOIN (SELECT u.city, COUNT(*) as n FROM scouts s JOIN users u ON s.user_id=u.id WHERE u.province=? AND u.city!='' AND s.deleted_at IS NULL AND u.deleted_at IS NULL GROUP BY u.city) s ON p.city=s.city`

	var rows []struct {
		C            string
		PlayerCount  int64
		ClubCount    int64
		CoachCount   int64
		AnalystCount int64
		ScoutCount   int64
	}
	args := []interface{}{province, province, province, province, province, province, province, province, province, province}
	if err := r.db.Raw(sql, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	res := make([]ProvincialAggregate, 0, len(rows))
	for _, v := range rows {
		total := v.PlayerCount + v.ClubCount + v.CoachCount + v.AnalystCount + v.ScoutCount
		res = append(res, ProvincialAggregate{
			CityCode:     utils.GetCityCode(v.C),
			CityName:     v.C,
			Count:        total,
			PlayerCount:  v.PlayerCount,
			ClubCount:    v.ClubCount,
			CoachCount:   v.CoachCount,
			AnalystCount: v.AnalystCount,
			ScoutCount:   v.ScoutCount,
			HeatLevel:    utils.CalculateHeatLevel(total),
		})
	}
	return res, nil
}

// ===== City =====

// GetCityEntities 城市散点分发
func (r *MultiLayerMapRepository) GetCityEntities(province, city string, layer EntityLayer) ([]CityEntityItem, error) {
	switch layer {
	case LayerClubs:
		return r.getCityClubs(province, city)
	case LayerCoaches:
		return r.getCityCoaches(province, city)
	case LayerAnalysts:
		return r.getCityAnalysts(province, city)
	case LayerScouts:
		return r.getCityScouts(province, city)
	case LayerAll:
		return r.getCityAll(province, city)
	default:
		return r.getCityPlayers(province, city)
	}
}

func (r *MultiLayerMapRepository) getCityPlayers(prov, city string) ([]CityEntityItem, error) {
	var rows []struct {
		ID uint; Name string; Avatar string; Position string; Age int; Club string
	}
	err := r.db.Raw(`SELECT id, name, avatar, position, age, club FROM users WHERE role='user' AND status='active' AND province=? AND city=? LIMIT 200`, prov, city).Scan(&rows).Error
	if err != nil { return nil, err }
	items := make([]CityEntityItem, 0, len(rows))
	for _, v := range rows {
		nx, ny := utils.GenerateNormalizedCoordinates(v.ID)
		items = append(items, CityEntityItem{ID: v.ID, Name: v.Name, Avatar: v.Avatar, Type: "player", Tags: utils.GenerateTags(v.Position), NormalizedX: nx, NormalizedY: ny, Extra: map[string]interface{}{"position": v.Position, "age": v.Age, "club": v.Club}})
	}
	return items, nil
}

func (r *MultiLayerMapRepository) getCityClubs(prov, city string) ([]CityEntityItem, error) {
	var rows []struct {
		ID uint; Name string; Logo string; Address string; ClubSize string
	}
	err := r.db.Raw(`SELECT id, name, logo, address, club_size FROM clubs WHERE province=? AND city=? AND deleted_at IS NULL LIMIT 200`, prov, city).Scan(&rows).Error
	if err != nil { return nil, err }
	items := make([]CityEntityItem, 0, len(rows))
	for _, v := range rows {
		nx, ny := utils.GenerateNormalizedCoordinates(v.ID + 10000)
		items = append(items, CityEntityItem{ID: v.ID, Name: v.Name, Avatar: v.Logo, Type: "club", Tags: []string{"青训俱乐部"}, NormalizedX: nx, NormalizedY: ny, Extra: map[string]interface{}{"address": v.Address, "clubSize": v.ClubSize}})
	}
	return items, nil
}

func (r *MultiLayerMapRepository) getCityCoaches(prov, city string) ([]CityEntityItem, error) {
	var rows []struct {
		ID uint; UserID uint; Name string; Avatar string; Position string; LicenseType string; CoachingYears int; CurrentClub string
	}
	err := r.db.Raw(`SELECT c.id, c.user_id, u.name, u.avatar, u.position, c.license_type, c.coaching_years, c.current_club FROM coaches c JOIN users u ON c.user_id=u.id WHERE u.province=? AND u.city=? AND c.deleted_at IS NULL AND u.deleted_at IS NULL LIMIT 200`, prov, city).Scan(&rows).Error
	if err != nil { return nil, err }
	items := make([]CityEntityItem, 0, len(rows))
	for _, v := range rows {
		nx, ny := utils.GenerateNormalizedCoordinates(v.ID + 20000)
		tags := []string{v.LicenseType + "级教练"}
		if v.Position != "" {
			tags = append(tags, v.Position)
		}
		items = append(items, CityEntityItem{ID: v.UserID, Name: v.Name, Avatar: v.Avatar, Type: "coach", Tags: tags, NormalizedX: nx, NormalizedY: ny, Extra: map[string]interface{}{"position": v.Position, "licenseType": v.LicenseType, "coachingYears": v.CoachingYears, "currentClub": v.CurrentClub}})
	}
	return items, nil
}

func (r *MultiLayerMapRepository) getCityAnalysts(prov, city string) ([]CityEntityItem, error) {
	var rows []struct {
		ID uint; UserID uint; Name string; Avatar string; Specialty string; Experience int; Rating float64
	}
	err := r.db.Raw(`SELECT a.id, a.user_id, u.name, u.avatar, a.specialty, a.experience, a.rating FROM analysts a JOIN users u ON a.user_id=u.id WHERE u.province=? AND u.city=? AND a.deleted_at IS NULL AND u.deleted_at IS NULL LIMIT 200`, prov, city).Scan(&rows).Error
	if err != nil { return nil, err }
	items := make([]CityEntityItem, 0, len(rows))
	for _, v := range rows {
		nx, ny := utils.GenerateNormalizedCoordinates(v.ID + 30000)
		items = append(items, CityEntityItem{ID: v.UserID, Name: v.Name, Avatar: v.Avatar, Type: "analyst", Tags: []string{v.Specialty}, Score: v.Rating, NormalizedX: nx, NormalizedY: ny, Extra: map[string]interface{}{"specialty": v.Specialty, "experience": v.Experience, "rating": v.Rating}})
	}
	return items, nil
}

func (r *MultiLayerMapRepository) getCityScouts(prov, city string) ([]CityEntityItem, error) {
	var rows []struct {
		ID uint; UserID uint; Name string; Avatar string; ScoutingExperience string; CurrentOrganization string; TotalDiscovered int
	}
	err := r.db.Raw(`SELECT s.id, s.user_id, u.name, u.avatar, s.scouting_experience, s.current_organization, s.total_discovered FROM scouts s JOIN users u ON s.user_id=u.id WHERE u.province=? AND u.city=? AND s.deleted_at IS NULL AND u.deleted_at IS NULL LIMIT 200`, prov, city).Scan(&rows).Error
	if err != nil { return nil, err }
	items := make([]CityEntityItem, 0, len(rows))
	for _, v := range rows {
		nx, ny := utils.GenerateNormalizedCoordinates(v.ID + 40000)
		items = append(items, CityEntityItem{ID: v.UserID, Name: v.Name, Avatar: v.Avatar, Type: "scout", Tags: []string{v.ScoutingExperience + "经验"}, NormalizedX: nx, NormalizedY: ny, Extra: map[string]interface{}{"scoutingExperience": v.ScoutingExperience, "currentOrganization": v.CurrentOrganization, "totalDiscovered": v.TotalDiscovered}})
	}
	return items, nil
}

func (r *MultiLayerMapRepository) getCityAll(prov, city string) ([]CityEntityItem, error) {
	var all []CityEntityItem
	if a, _ := r.getCityPlayers(prov, city); len(a) > 0 { all = append(all, a...) }
	if a, _ := r.getCityClubs(prov, city); len(a) > 0 { all = append(all, a...) }
	if a, _ := r.getCityCoaches(prov, city); len(a) > 0 { all = append(all, a...) }
	if a, _ := r.getCityAnalysts(prov, city); len(a) > 0 { all = append(all, a...) }
	if a, _ := r.getCityScouts(prov, city); len(a) > 0 { all = append(all, a...) }
	return all, nil
}

// ===== P2-14~P2-17 Fill Methods =====

// ----- National -----

func (r *MultiLayerMapRepository) fillSizeDistributionForNational(aggregates []NationalAggregate) {
	if len(aggregates) == 0 {
		return
	}
	var rows []struct {
		Province string
		Size     string
		N        int64
	}
	sql := `SELECT province, club_size as size, COUNT(*) as n FROM clubs WHERE province!='' AND deleted_at IS NULL GROUP BY province, club_size`
	if err := r.db.Raw(sql).Scan(&rows).Error; err != nil {
		return
	}
	idx := make(map[string]int, len(aggregates))
	for i, a := range aggregates {
		idx[a.ProvinceName] = i
	}
	for _, v := range rows {
		if i, ok := idx[v.Province]; ok {
			if aggregates[i].SizeDistribution == nil {
				aggregates[i].SizeDistribution = make(map[string]int64)
			}
			aggregates[i].SizeDistribution[v.Size] = v.N
		}
	}
}

func (r *MultiLayerMapRepository) fillLicenseDistributionForNational(aggregates []NationalAggregate) {
	if len(aggregates) == 0 {
		return
	}
	var rows []struct {
		Province string
		License  string
		N        int64
	}
	sql := `SELECT u.province, c.license_type as license, COUNT(*) as n FROM coaches c JOIN users u ON c.user_id=u.id WHERE u.province!='' AND c.deleted_at IS NULL AND u.deleted_at IS NULL GROUP BY u.province, c.license_type`
	if err := r.db.Raw(sql).Scan(&rows).Error; err != nil {
		return
	}
	idx := make(map[string]int, len(aggregates))
	for i, a := range aggregates {
		idx[a.ProvinceName] = i
	}
	for _, v := range rows {
		if i, ok := idx[v.Province]; ok {
			if aggregates[i].LicenseDistribution == nil {
				aggregates[i].LicenseDistribution = make(map[string]int64)
			}
			aggregates[i].LicenseDistribution[v.License] = v.N
		}
	}
}

func (r *MultiLayerMapRepository) fillSpecialtyDistributionForNational(aggregates []NationalAggregate) {
	if len(aggregates) == 0 {
		return
	}
	var rows []struct {
		Province  string
		Specialty string
		N         int64
	}
	sql := `SELECT u.province, a.specialty, COUNT(*) as n FROM analysts a JOIN users u ON a.user_id=u.id WHERE u.province!='' AND a.deleted_at IS NULL AND u.deleted_at IS NULL GROUP BY u.province, a.specialty`
	if err := r.db.Raw(sql).Scan(&rows).Error; err != nil {
		return
	}
	idx := make(map[string]int, len(aggregates))
	for i, a := range aggregates {
		idx[a.ProvinceName] = i
	}
	for _, v := range rows {
		if i, ok := idx[v.Province]; ok {
			if aggregates[i].SpecialtyDistribution == nil {
				aggregates[i].SpecialtyDistribution = make(map[string]int64)
			}
			aggregates[i].SpecialtyDistribution[v.Specialty] = v.N
		}
	}
}

func (r *MultiLayerMapRepository) fillAdoptionRateForNational(aggregates []NationalAggregate) {
	if len(aggregates) == 0 {
		return
	}
	var rows []struct {
		Province string
		Rate     float64
	}
	sql := `SELECT u.province, COALESCE(AVG(CASE WHEN s.total_reports > 0 THEN CAST(s.total_adopted AS REAL) / s.total_reports ELSE 0 END), 0) as rate FROM scouts s JOIN users u ON s.user_id=u.id WHERE u.province!='' AND s.deleted_at IS NULL AND u.deleted_at IS NULL GROUP BY u.province`
	if err := r.db.Raw(sql).Scan(&rows).Error; err != nil {
		return
	}
	idx := make(map[string]int, len(aggregates))
	for i, a := range aggregates {
		idx[a.ProvinceName] = i
	}
	for _, v := range rows {
		if i, ok := idx[v.Province]; ok {
			aggregates[i].AdoptionRate = v.Rate
		}
	}
}

// ----- Provincial -----

func (r *MultiLayerMapRepository) fillSizeDistributionForProvincial(aggregates []ProvincialAggregate, province string) {
	if len(aggregates) == 0 {
		return
	}
	var rows []struct {
		City string
		Size string
		N    int64
	}
	sql := `SELECT city, club_size as size, COUNT(*) as n FROM clubs WHERE province=? AND city!='' AND deleted_at IS NULL GROUP BY city, club_size`
	if err := r.db.Raw(sql, province).Scan(&rows).Error; err != nil {
		return
	}
	idx := make(map[string]int, len(aggregates))
	for i, a := range aggregates {
		idx[a.CityName] = i
	}
	for _, v := range rows {
		if i, ok := idx[v.City]; ok {
			if aggregates[i].SizeDistribution == nil {
				aggregates[i].SizeDistribution = make(map[string]int64)
			}
			aggregates[i].SizeDistribution[v.Size] = v.N
		}
	}
}

func (r *MultiLayerMapRepository) fillLicenseDistributionForProvincial(aggregates []ProvincialAggregate, province string) {
	if len(aggregates) == 0 {
		return
	}
	var rows []struct {
		City    string
		License string
		N       int64
	}
	sql := `SELECT u.city, c.license_type as license, COUNT(*) as n FROM coaches c JOIN users u ON c.user_id=u.id WHERE u.province=? AND u.city!='' AND c.deleted_at IS NULL AND u.deleted_at IS NULL GROUP BY u.city, c.license_type`
	if err := r.db.Raw(sql, province).Scan(&rows).Error; err != nil {
		return
	}
	idx := make(map[string]int, len(aggregates))
	for i, a := range aggregates {
		idx[a.CityName] = i
	}
	for _, v := range rows {
		if i, ok := idx[v.City]; ok {
			if aggregates[i].LicenseDistribution == nil {
				aggregates[i].LicenseDistribution = make(map[string]int64)
			}
			aggregates[i].LicenseDistribution[v.License] = v.N
		}
	}
}

func (r *MultiLayerMapRepository) fillSpecialtyDistributionForProvincial(aggregates []ProvincialAggregate, province string) {
	if len(aggregates) == 0 {
		return
	}
	var rows []struct {
		City      string
		Specialty string
		N         int64
	}
	sql := `SELECT u.city, a.specialty, COUNT(*) as n FROM analysts a JOIN users u ON a.user_id=u.id WHERE u.province=? AND u.city!='' AND a.deleted_at IS NULL AND u.deleted_at IS NULL GROUP BY u.city, a.specialty`
	if err := r.db.Raw(sql, province).Scan(&rows).Error; err != nil {
		return
	}
	idx := make(map[string]int, len(aggregates))
	for i, a := range aggregates {
		idx[a.CityName] = i
	}
	for _, v := range rows {
		if i, ok := idx[v.City]; ok {
			if aggregates[i].SpecialtyDistribution == nil {
				aggregates[i].SpecialtyDistribution = make(map[string]int64)
			}
			aggregates[i].SpecialtyDistribution[v.Specialty] = v.N
		}
	}
}

func (r *MultiLayerMapRepository) fillAdoptionRateForProvincial(aggregates []ProvincialAggregate, province string) {
	if len(aggregates) == 0 {
		return
	}
	var rows []struct {
		City string
		Rate float64
	}
	sql := `SELECT u.city, COALESCE(AVG(CASE WHEN s.total_reports > 0 THEN CAST(s.total_adopted AS REAL) / s.total_reports ELSE 0 END), 0) as rate FROM scouts s JOIN users u ON s.user_id=u.id WHERE u.province=? AND u.city!='' AND s.deleted_at IS NULL AND u.deleted_at IS NULL GROUP BY u.city`
	if err := r.db.Raw(sql, province).Scan(&rows).Error; err != nil {
		return
	}
	idx := make(map[string]int, len(aggregates))
	for i, a := range aggregates {
		idx[a.CityName] = i
	}
	for _, v := range rows {
		if i, ok := idx[v.City]; ok {
			aggregates[i].AdoptionRate = v.Rate
		}
	}
}
