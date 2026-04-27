package repositories

import (
	"log"

	"github.com/shaonianqiutan/backend/models"
	"gorm.io/gorm"
)

// ClubHomeRepository 俱乐部主页数据访问层
type ClubHomeRepository struct {
	db *gorm.DB
}

// NewClubHomeRepository 创建俱乐部主页仓储
func NewClubHomeRepository(db *gorm.DB) *ClubHomeRepository {
	return &ClubHomeRepository{db: db}
}

// Create 创建俱乐部主页配置
func (r *ClubHomeRepository) Create(home *models.ClubHome) error {
	if err := r.db.Create(home).Error; err != nil {
		log.Printf("[ClubHomeRepository] Create failed: %v", err)
		return err
	}
	return nil
}

// Save 保存俱乐部主页配置
func (r *ClubHomeRepository) Save(home *models.ClubHome) error {
	if err := r.db.Save(home).Error; err != nil {
		log.Printf("[ClubHomeRepository] Save failed: %v", err)
		return err
	}
	return nil
}

// FindByClubID 根据俱乐部ID获取主页配置
func (r *ClubHomeRepository) FindByClubID(clubID uint) (*models.ClubHome, error) {
	var home models.ClubHome
	if err := r.db.Where("club_id = ?", clubID).First(&home).Error; err != nil {
		log.Printf("[ClubHomeRepository] FindByClubID failed for clubID=%d: %v", clubID, err)
		return nil, err
	}
	return &home, nil
}

// UpdateHero 更新 Hero 配置
func (r *ClubHomeRepository) UpdateHero(clubID uint, hero *models.ClubHomeHero) error {
	home, err := r.FindByClubID(clubID)
	if err != nil {
		home = models.DefaultClubHome(clubID)
		home.Hero = *hero
		return r.Create(home)
	}
	home.Hero = *hero
	return r.Save(home)
}

// UpdateAbout 更新 About 配置
func (r *ClubHomeRepository) UpdateAbout(clubID uint, about *models.ClubHomeAbout) error {
	home, err := r.FindByClubID(clubID)
	if err != nil {
		home = models.DefaultClubHome(clubID)
		home.About = *about
		return r.Create(home)
	}
	home.About = *about
	return r.Save(home)
}

// UpdateContact 更新联系方式
func (r *ClubHomeRepository) UpdateContact(clubID uint, contact *models.ClubHomeContact) error {
	home, err := r.FindByClubID(clubID)
	if err != nil {
		home = models.DefaultClubHome(clubID)
		home.Contact = *contact
		return r.Create(home)
	}
	home.Contact = *contact
	return r.Save(home)
}

// UpdateFacilities 更新训练环境配置
func (r *ClubHomeRepository) UpdateFacilities(clubID uint, facilities *models.ClubHomeFacilities) error {
	home, err := r.FindByClubID(clubID)
	if err != nil {
		home = models.DefaultClubHome(clubID)
		home.Facilities = *facilities
		return r.Create(home)
	}
	home.Facilities = *facilities
	return r.Save(home)
}

// UpdateRecruitment 更新招生信息配置
func (r *ClubHomeRepository) UpdateRecruitment(clubID uint, recruitment *models.ClubHomeRecruitment) error {
	home, err := r.FindByClubID(clubID)
	if err != nil {
		home = models.DefaultClubHome(clubID)
		home.Recruitment = *recruitment
		return r.Create(home)
	}
	home.Recruitment = *recruitment
	return r.Save(home)
}

// UpdateSocialLinks 更新社交媒体链接
func (r *ClubHomeRepository) UpdateSocialLinks(clubID uint, links *models.ClubHomeSocialLinks) error {
	home, err := r.FindByClubID(clubID)
	if err != nil {
		home = models.DefaultClubHome(clubID)
		home.SocialLinks = *links
		return r.Create(home)
	}
	home.SocialLinks = *links
	return r.Save(home)
}

// UpdateNews 更新手工置顶公告
func (r *ClubHomeRepository) UpdateNews(clubID uint, items []models.ClubHomeNewsItem) error {
	home, err := r.FindByClubID(clubID)
	if err != nil {
		home = models.DefaultClubHome(clubID)
		home.NewsItems = items
		return r.Create(home)
	}
	home.NewsItems = items
	return r.Save(home)
}

// UpdateModules 更新模块排序和可见性
func (r *ClubHomeRepository) UpdateModules(clubID uint, order string, visibility map[string]bool) error {
	home, err := r.FindByClubID(clubID)
	if err != nil {
		home = models.DefaultClubHome(clubID)
		home.ModuleOrder = order
		home.ModuleVisibility = visibility
		return r.Create(home)
	}
	home.ModuleOrder = order
	home.ModuleVisibility = visibility
	return r.Save(home)
}

// AchievementResult 成就查询结果
type AchievementResult struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Count       string `json:"count"`
}

// GetAchievements 获取成就列表
func (r *ClubHomeRepository) GetAchievements(clubID uint) ([]AchievementResult, error) {
	var achievements []AchievementResult
	if err := r.db.Table("achievements").
		Select("id, title, description, count").
		Where("club_id = ?", clubID).
		Order("sort ASC").
		Scan(&achievements).Error; err != nil {
		log.Printf("[ClubHomeRepository] GetAchievements failed for clubID=%d: %v", clubID, err)
		return nil, err
	}
	return achievements, nil
}

// SaveAchievements 保存成就列表
func (r *ClubHomeRepository) SaveAchievements(clubID uint, achievements []models.Achievement) error {
	if err := r.db.Where("club_id = ?", clubID).Delete(&models.Achievement{}).Error; err != nil {
		log.Printf("[ClubHomeRepository] SaveAchievements (delete old) failed for clubID=%d: %v", clubID, err)
		return err
	}

	for i := range achievements {
		achievements[i].ClubID = clubID
		achievements[i].Sort = i
	}

	if len(achievements) == 0 {
		return nil
	}

	if err := r.db.Create(&achievements).Error; err != nil {
		log.Printf("[ClubHomeRepository] SaveAchievements (create) failed for clubID=%d: %v", clubID, err)
		return err
	}
	return nil
}

// CoachResult 教练查询结果
type CoachResult struct {
	ID       uint   `json:"id"`
	UserID   uint   `json:"user_id"`
	Name     string `json:"name"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Role     string `json:"role"`
}

// GetCoaches 获取教练列表
func (r *ClubHomeRepository) GetCoaches(clubID uint) ([]CoachResult, error) {
	var coaches []CoachResult
	if err := r.db.Table("team_coaches tc").
		Select("tc.id, tc.user_id, u.name, u.nickname, u.avatar, tc.role").
		Joins("JOIN teams t ON t.id = tc.team_id").
		Joins("JOIN users u ON u.id = tc.user_id").
		Where("t.club_id = ? AND tc.status = ?", clubID, "active").
		Limit(10).
		Scan(&coaches).Error; err != nil {
		log.Printf("[ClubHomeRepository] GetCoaches failed for clubID=%d: %v", clubID, err)
		return nil, err
	}
	return coaches, nil
}

// GetClubHomeTeams 获取主页展示的球队配置
func (r *ClubHomeRepository) GetClubHomeTeams(clubID uint) ([]models.ClubHomeTeam, error) {
	var items []models.ClubHomeTeam
	if err := r.db.Where("club_id = ?", clubID).Order("sort ASC").Find(&items).Error; err != nil {
		log.Printf("[ClubHomeRepository] GetClubHomeTeams failed for clubID=%d: %v", clubID, err)
		return nil, err
	}
	return items, nil
}

// SaveClubHomeTeams 保存主页展示的球队配置
func (r *ClubHomeRepository) SaveClubHomeTeams(clubID uint, items []models.ClubHomeTeam) error {
	if err := r.db.Where("club_id = ?", clubID).Delete(&models.ClubHomeTeam{}).Error; err != nil {
		log.Printf("[ClubHomeRepository] SaveClubHomeTeams (delete old) failed for clubID=%d: %v", clubID, err)
		return err
	}
	for i := range items {
		items[i].ClubID = clubID
		items[i].Sort = i
	}
	if len(items) == 0 {
		return nil
	}
	if err := r.db.Create(&items).Error; err != nil {
		log.Printf("[ClubHomeRepository] SaveClubHomeTeams (create) failed for clubID=%d: %v", clubID, err)
		return err
	}
	return nil
}

// GetClubHomeCoaches 获取主页展示的教练配置
func (r *ClubHomeRepository) GetClubHomeCoaches(clubID uint) ([]models.ClubHomeCoach, error) {
	var items []models.ClubHomeCoach
	if err := r.db.Where("club_id = ?", clubID).Order("sort ASC").Find(&items).Error; err != nil {
		log.Printf("[ClubHomeRepository] GetClubHomeCoaches failed for clubID=%d: %v", clubID, err)
		return nil, err
	}
	return items, nil
}

// SaveClubHomeCoaches 保存主页展示的教练配置
func (r *ClubHomeRepository) SaveClubHomeCoaches(clubID uint, items []models.ClubHomeCoach) error {
	if err := r.db.Where("club_id = ?", clubID).Delete(&models.ClubHomeCoach{}).Error; err != nil {
		log.Printf("[ClubHomeRepository] SaveClubHomeCoaches (delete old) failed for clubID=%d: %v", clubID, err)
		return err
	}
	for i := range items {
		items[i].ClubID = clubID
		items[i].Sort = i
	}
	if len(items) == 0 {
		return nil
	}
	if err := r.db.Create(&items).Error; err != nil {
		log.Printf("[ClubHomeRepository] SaveClubHomeCoaches (create) failed for clubID=%d: %v", clubID, err)
		return err
	}
	return nil
}

// GetClubHomePlayers 获取主页展示的球员配置
func (r *ClubHomeRepository) GetClubHomePlayers(clubID uint) ([]models.ClubHomePlayer, error) {
	var items []models.ClubHomePlayer
	if err := r.db.Where("club_id = ?", clubID).Order("sort ASC").Find(&items).Error; err != nil {
		log.Printf("[ClubHomeRepository] GetClubHomePlayers failed for clubID=%d: %v", clubID, err)
		return nil, err
	}
	return items, nil
}

// SaveClubHomePlayers 保存主页展示的球员配置
func (r *ClubHomeRepository) SaveClubHomePlayers(clubID uint, items []models.ClubHomePlayer) error {
	if err := r.db.Where("club_id = ?", clubID).Delete(&models.ClubHomePlayer{}).Error; err != nil {
		log.Printf("[ClubHomeRepository] SaveClubHomePlayers (delete old) failed for clubID=%d: %v", clubID, err)
		return err
	}
	for i := range items {
		items[i].ClubID = clubID
		items[i].Sort = i
	}
	if len(items) == 0 {
		return nil
	}
	if err := r.db.Create(&items).Error; err != nil {
		log.Printf("[ClubHomeRepository] SaveClubHomePlayers (create) failed for clubID=%d: %v", clubID, err)
		return err
	}
	return nil
}
