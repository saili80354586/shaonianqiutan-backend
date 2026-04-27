package repositories

import (
	"github.com/shaonianqiutan/backend/models"
	"gorm.io/gorm"
)

// TeamHomeRepository 球队主页仓库
type TeamHomeRepository struct {
	db *gorm.DB
}

// NewTeamHomeRepository 创建球队主页仓库
func NewTeamHomeRepository(db *gorm.DB) *TeamHomeRepository {
	return &TeamHomeRepository{db: db}
}

// AutoMigrate 自动迁移
func (r *TeamHomeRepository) AutoMigrate() error {
	return r.db.AutoMigrate(&models.TeamHome{})
}

// FindByTeamID 根据球队ID获取主页配置
func (r *TeamHomeRepository) FindByTeamID(teamID uint) (*models.TeamHome, error) {
	var home models.TeamHome
	err := r.db.Where("team_id = ?", teamID).First(&home).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &home, nil
}

// FindOrCreate 查找或创建
func (r *TeamHomeRepository) FindOrCreate(teamID uint) (*models.TeamHome, error) {
	home, err := r.FindByTeamID(teamID)
	if err != nil {
		return nil, err
	}
	if home != nil {
		return home, nil
	}

	// 创建默认配置
	home = models.DefaultTeamHome(teamID)
	if err := r.db.Create(home).Error; err != nil {
		return nil, err
	}
	return home, nil
}

// Save 保存
func (r *TeamHomeRepository) Save(home *models.TeamHome) error {
	if home.ID == 0 {
		return r.db.Create(home).Error
	}
	return r.db.Save(home).Error
}

// UpdateHero 更新 Hero 配置
func (r *TeamHomeRepository) UpdateHero(teamID uint, hero *models.TeamHomeHero) error {
	home, err := r.FindOrCreate(teamID)
	if err != nil {
		return err
	}
	home.Hero = *hero
	return r.Save(home)
}

// UpdateAbout 更新 About 配置
func (r *TeamHomeRepository) UpdateAbout(teamID uint, about *models.TeamHomeAbout) error {
	home, err := r.FindOrCreate(teamID)
	if err != nil {
		return err
	}
	home.About = *about
	return r.Save(home)
}

// UpdateContact 更新联系方式
func (r *TeamHomeRepository) UpdateContact(teamID uint, contact *models.TeamHomeContact) error {
	home, err := r.FindOrCreate(teamID)
	if err != nil {
		return err
	}
	home.Contact = *contact
	return r.Save(home)
}

// SaveHonors 保存荣誉列表
func (r *TeamHomeRepository) SaveHonors(teamID uint, honors []models.TeamHonor) error {
	home, err := r.FindOrCreate(teamID)
	if err != nil {
		return err
	}
	home.Honors = honors
	return r.Save(home)
}

// GetHonors 获取荣誉列表
func (r *TeamHomeRepository) GetHonors(teamID uint) ([]models.TeamHonor, error) {
	home, err := r.FindByTeamID(teamID)
	if err != nil || home == nil {
		return []models.TeamHonor{}, err
	}
	return home.Honors, nil
}

// GetDynamics 获取动态列表
func (r *TeamHomeRepository) GetDynamics(teamID uint) ([]models.TeamDynamic, error) {
	home, err := r.FindByTeamID(teamID)
	if err != nil || home == nil {
		return []models.TeamDynamic{}, err
	}
	return home.Dynamics, nil
}

// AddDynamic 添加动态
func (r *TeamHomeRepository) AddDynamic(teamID uint, dynamic *models.TeamDynamic) error {
	home, err := r.FindOrCreate(teamID)
	if err != nil {
		return err
	}
	dynamic.ID = uint(len(home.Dynamics) + 1)
	home.Dynamics = append(home.Dynamics, *dynamic)
	return r.Save(home)
}

// DeleteDynamic 删除动态
func (r *TeamHomeRepository) DeleteDynamic(teamID uint, dynamicID uint) error {
	home, err := r.FindByTeamID(teamID)
	if err != nil || home == nil {
		return err
	}
	var newDynamics []models.TeamDynamic
	for _, d := range home.Dynamics {
		if d.ID != dynamicID {
			newDynamics = append(newDynamics, d)
		}
	}
	home.Dynamics = newDynamics
	return r.Save(home)
}

// UpdateDynamics 更新动态列表
func (r *TeamHomeRepository) UpdateDynamics(teamID uint, dynamics []models.TeamDynamic) error {
	home, err := r.FindOrCreate(teamID)
	if err != nil {
		return err
	}
	home.Dynamics = dynamics
	return r.Save(home)
}
