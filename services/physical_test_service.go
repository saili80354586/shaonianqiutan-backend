package services

import (
	"encoding/json"
	"errors"
	"math"

	"github.com/shaonianqiutan/backend/models"
	"gorm.io/gorm"
)

// PhysicalTestService 体测服务
type PhysicalTestService struct {
	db *gorm.DB
}

// NewPhysicalTestService 创建体测服务
func NewPhysicalTestService(db *gorm.DB) *PhysicalTestService {
	return &PhysicalTestService{db: db}
}

// GetDB 获取数据库连接
func (s *PhysicalTestService) GetDB() *gorm.DB {
	return s.db
}

// GetClubByUserID 根据用户ID获取俱乐部
// 支持俱乐部管理员直接查询，也支持球队教练通过关联球队间接查询
func (s *PhysicalTestService) GetClubByUserID(userID uint) (*models.Club, error) {
	var club models.Club

	// 1. 先尝试直接查俱乐部管理员
	err := s.db.Where("user_id = ?", userID).First(&club).Error
	if err == nil {
		return &club, nil
	}

	// 2. 如果不是俱乐部管理员，尝试查是否是球队教练
	var teamCoach models.TeamCoach
	if err := s.db.Where("user_id = ? AND status = ?", userID, "active").First(&teamCoach).Error; err == nil {
		var team models.Team
		if err := s.db.First(&team, teamCoach.TeamID).Error; err == nil {
			if err := s.db.First(&club, team.ClubID).Error; err == nil {
				return &club, nil
			}
		}
	}

	return nil, errors.New("未找到关联俱乐部")
}

// GetPhysicalTests 获取体测活动列表
func (s *PhysicalTestService) GetPhysicalTests(clubID uint, teamID uint, page, pageSize int, status string) ([]models.PhysicalTestActivity, int64, error) {
	var allTests []models.PhysicalTestActivity

	query := s.db.Model(&models.PhysicalTestActivity{}).Where("club_id = ?", clubID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&allTests).Error; err != nil {
		return nil, 0, err
	}

	filteredTests := allTests
	if teamID > 0 {
		var teamPlayers []models.TeamPlayer
		if err := s.db.Where("team_id = ? AND status = ?", teamID, "active").Find(&teamPlayers).Error; err != nil {
			return nil, 0, err
		}

		teamPlayerSet := make(map[uint]struct{}, len(teamPlayers))
		for _, player := range teamPlayers {
			teamPlayerSet[player.UserID] = struct{}{}
		}

		filteredTests = make([]models.PhysicalTestActivity, 0, len(allTests))
		for _, test := range allTests {
			for _, playerID := range test.GetPlayerIDs() {
				if _, ok := teamPlayerSet[playerID]; ok {
					filteredTests = append(filteredTests, test)
					break
				}
			}
		}
	}

	total := int64(len(filteredTests))
	offset := (page - 1) * pageSize
	if offset >= len(filteredTests) {
		return []models.PhysicalTestActivity{}, total, nil
	}

	end := offset + pageSize
	if end > len(filteredTests) {
		end = len(filteredTests)
	}

	return filteredTests[offset:end], total, nil
}

// CreatePhysicalTest 创建体测活动
func (s *PhysicalTestService) CreatePhysicalTest(test *models.PhysicalTestActivity) error {
	return s.db.Create(test).Error
}

// GetPhysicalTestByID 根据ID获取体测活动
func (s *PhysicalTestService) GetPhysicalTestByID(id uint) (*models.PhysicalTestActivity, error) {
	var test models.PhysicalTestActivity
	err := s.db.First(&test, id).Error
	if err != nil {
		return nil, err
	}
	return &test, nil
}

// UpdatePhysicalTest 更新体测活动
func (s *PhysicalTestService) UpdatePhysicalTest(id uint, updates map[string]interface{}) error {
	return s.db.Model(&models.PhysicalTestActivity{}).Where("id = ?", id).Updates(updates).Error
}

// DeletePhysicalTest 删除体测活动
func (s *PhysicalTestService) DeletePhysicalTest(id uint) error {
	return s.db.Where("id = ?", id).Delete(&models.PhysicalTestActivity{}).Error
}

// GetCompletedRecordCount 获取已完成记录数
func (s *PhysicalTestService) GetCompletedRecordCount(activityID uint) (int, error) {
	var count int64
	err := s.db.Model(&models.PhysicalTestRecord{}).
		Where("activity_id = ? AND (height IS NOT NULL OR weight IS NOT NULL OR sprint_30m IS NOT NULL)", activityID).
		Count(&count).Error
	return int(count), err
}

// GetReportsCount 获取报告数
func (s *PhysicalTestService) GetReportsCount(activityID uint) (int, error) {
	var count int64
	err := s.db.Model(&models.PhysicalTestReport{}).
		Where("activity_id = ?", activityID).
		Count(&count).Error
	return int(count), err
}

// GetPhysicalTestRecords 获取体测记录列表
func (s *PhysicalTestService) GetPhysicalTestRecords(activityID uint, playerID *uint) ([]models.PhysicalTestRecord, error) {
	var records []models.PhysicalTestRecord
	query := s.db.Where("activity_id = ?", activityID)

	if playerID != nil {
		query = query.Where("player_id = ?", *playerID)
	}

	err := query.Preload("Player").Find(&records).Error
	return records, err
}

// GetLatestPhysicalTestRecordByPlayer 获取球员最新体测记录
func (s *PhysicalTestService) GetLatestPhysicalTestRecordByPlayer(playerID uint) (*models.PhysicalTestRecord, error) {
	var record models.PhysicalTestRecord
	err := s.db.Where("player_id = ?", playerID).
		Order("test_date DESC, created_at DESC").
		First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// CreatePhysicalTestRecord 创建体测记录
func (s *PhysicalTestService) CreatePhysicalTestRecord(record *models.PhysicalTestRecord) error {
	return s.db.Create(record).Error
}

// UpdateRecord 更新体测记录
func (s *PhysicalTestService) UpdateRecord(record *models.PhysicalTestRecord) error {
	return s.db.Save(record).Error
}

// SetRecordData 设置体测数据
func (s *PhysicalTestService) SetRecordData(record *models.PhysicalTestRecord, data map[string]float64) {
	if v, ok := data["height"]; ok {
		record.Height = &v
	}
	if v, ok := data["weight"]; ok {
		record.Weight = &v
	}
	if v, ok := data["sprint_30m"]; ok {
		record.Sprint30m = &v
	}
	if v, ok := data["sprint_50m"]; ok {
		record.Sprint50m = &v
	}
	if v, ok := data["sprint_100m"]; ok {
		record.Sprint100m = &v
	}
	if v, ok := data["agility_ladder"]; ok {
		record.AgilityLadder = &v
	}
	if v, ok := data["t_test"]; ok {
		record.TTest = &v
	}
	if v, ok := data["shuttle_run"]; ok {
		record.ShuttleRun = &v
	}
	if v, ok := data["standing_long_jump"]; ok {
		record.StandingLongJump = &v
	}
	if v, ok := data["vertical_jump"]; ok {
		record.VerticalJump = &v
	}
	if v, ok := data["sit_and_reach"]; ok {
		record.SitAndReach = &v
	}
	if v, ok := data["push_up"]; ok {
		i := int(v)
		record.PushUp = &i
	}
	if v, ok := data["sit_up"]; ok {
		i := int(v)
		record.SitUp = &i
	}
	if v, ok := data["plank"]; ok {
		i := int(v)
		record.Plank = &i
	}

	// 自动计算BMI
	if record.Height != nil && record.Weight != nil && *record.Height > 0 {
		heightM := *record.Height / 100.0
		bmi := *record.Weight / (heightM * heightM)
		record.BMI = &bmi
	}
}

// GetRecordProgress 获取记录进度
func (s *PhysicalTestService) GetRecordProgress(record *models.PhysicalTestRecord) map[string]int {
	total := 7 // 基础指标数量
	completed := 0

	if record.Height != nil {
		completed++
	}
	if record.Weight != nil {
		completed++
	}
	if record.Sprint30m != nil {
		completed++
	}
	if record.ShuttleRun != nil {
		completed++
	}
	if record.StandingLongJump != nil {
		completed++
	}
	if record.SitAndReach != nil {
		completed++
	}
	if record.BMI != nil {
		completed++
	}

	return map[string]int{
		"total":     total,
		"completed": completed,
	}
}

// GetCalculatedData 获取自动计算的数据
func (s *PhysicalTestService) GetCalculatedData(record *models.PhysicalTestRecord) map[string]interface{} {
	data := make(map[string]interface{})

	if record.BMI != nil {
		data["bmi"] = math.Round(*record.BMI*10) / 10
	}

	return data
}

// ========== 自定义模板相关方法 ==========

// GetCustomTemplates 获取俱乐部的自定义模板列表
func (s *PhysicalTestService) GetCustomTemplates(clubID uint) ([]models.PhysicalTestTemplateCustom, error) {
	var templates []models.PhysicalTestTemplateCustom
	err := s.db.Where("club_id = ?", clubID).Order("created_at DESC").Find(&templates).Error
	return templates, err
}

// CreateCustomTemplate 创建自定义模板
func (s *PhysicalTestService) CreateCustomTemplate(template *models.PhysicalTestTemplateCustom) error {
	return s.db.Create(template).Error
}

// GetCustomTemplateByID 根据ID获取自定义模板
func (s *PhysicalTestService) GetCustomTemplateByID(id uint) (*models.PhysicalTestTemplateCustom, error) {
	var template models.PhysicalTestTemplateCustom
	err := s.db.First(&template, id).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// DeleteCustomTemplate 删除自定义模板
func (s *PhysicalTestService) DeleteCustomTemplate(id, clubID uint) error {
	return s.db.Where("id = ? AND club_id = ?", id, clubID).Delete(&models.PhysicalTestTemplateCustom{}).Error
}

// GetTemplateItems 获取模板对应的项目key列表
func (s *PhysicalTestService) GetTemplateItems(template models.PhysicalTestTemplate, customItems string, customTemplateID uint) []string {
	// 内置模板项目定义
	basicItems := []string{"height", "weight", "bmi", "sprint_30m"}
	advancedItems := []string{"height", "weight", "bmi", "sprint_30m", "shuttle_run", "standing_long_jump", "sit_and_reach"}
	professionalItems := []string{"height", "weight", "bmi", "sprint_30m", "sprint_50m", "agility_ladder", "t_test", "shuttle_run", "standing_long_jump", "vertical_jump", "sit_and_reach", "push_up", "sit_up", "plank"}

	switch template {
	case models.PTTemplateBasic:
		return basicItems
	case models.PTTemplateAdvanced:
		return advancedItems
	case models.PTTemplateProfessional:
		return professionalItems
	case models.PTTemplateCustom:
		// 优先使用自定义模板ID查询
		if customTemplateID > 0 {
			t, err := s.GetCustomTemplateByID(customTemplateID)
			if err == nil {
				return t.GetItems()
			}
		}
		// 回退到 customItems 字段
		if customItems != "" {
			var items []string
			json.Unmarshal([]byte(customItems), &items)
			return items
		}
		return []string{}
	default:
		return advancedItems
	}
}

// GetTestDataMap 获取体测数据map (作为函数而非方法)
func GetTestDataMapFromRecord(r *models.PhysicalTestRecord) map[string]interface{} {
	data := make(map[string]interface{})

	if r.Height != nil {
		data["height"] = *r.Height
	}
	if r.Weight != nil {
		data["weight"] = *r.Weight
	}
	if r.BMI != nil {
		data["bmi"] = math.Round(*r.BMI*10) / 10
	}
	if r.Sprint30m != nil {
		data["sprint_30m"] = math.Round(*r.Sprint30m*100) / 100
	}
	if r.Sprint50m != nil {
		data["sprint_50m"] = math.Round(*r.Sprint50m*100) / 100
	}
	if r.Sprint100m != nil {
		data["sprint_100m"] = math.Round(*r.Sprint100m*100) / 100
	}
	if r.AgilityLadder != nil {
		data["agility_ladder"] = math.Round(*r.AgilityLadder*100) / 100
	}
	if r.TTest != nil {
		data["t_test"] = math.Round(*r.TTest*100) / 100
	}
	if r.ShuttleRun != nil {
		data["shuttle_run"] = math.Round(*r.ShuttleRun*100) / 100
	}
	if r.StandingLongJump != nil {
		data["standing_long_jump"] = *r.StandingLongJump
	}
	if r.VerticalJump != nil {
		data["vertical_jump"] = *r.VerticalJump
	}
	if r.SitAndReach != nil {
		data["sit_and_reach"] = math.Round(*r.SitAndReach*10) / 10
	}
	if r.PushUp != nil {
		data["push_up"] = *r.PushUp
	}
	if r.SitUp != nil {
		data["sit_up"] = *r.SitUp
	}
	if r.Plank != nil {
		data["plank"] = *r.Plank
	}

	return data
}
