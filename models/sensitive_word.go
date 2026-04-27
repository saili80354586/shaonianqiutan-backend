package models

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

// SensitiveWord 敏感词模型
type SensitiveWord struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Word      string         `json:"word" gorm:"size:100;not null;uniqueIndex"`
	Category  string         `json:"category" gorm:"size:50;default:'general'"` // general/politics/porn/violence/advertising
	Level     int            `json:"level" gorm:"default:1"`                    // 1=警告 2=拦截 3=封号
	Enabled   bool           `json:"enabled" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 表名
func (SensitiveWord) TableName() string {
	return "sensitive_words"
}

// SensitiveWordRepository 敏感词数据访问层
type SensitiveWordRepository struct {
	db *gorm.DB
}

func NewSensitiveWordRepository(db *gorm.DB) *SensitiveWordRepository {
	return &SensitiveWordRepository{db: db}
}

// Create 创建敏感词
func (r *SensitiveWordRepository) Create(word *SensitiveWord) error {
	word.Word = strings.TrimSpace(word.Word)
	return r.db.Create(word).Error
}

// FindByID 根据ID查询
func (r *SensitiveWordRepository) FindByID(id uint) (*SensitiveWord, error) {
	var word SensitiveWord
	result := r.db.First(&word, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &word, nil
}

// FindAll 获取敏感词列表
func (r *SensitiveWordRepository) FindAll(page, pageSize int, category string, enabled *bool) ([]SensitiveWord, int64, error) {
	var words []SensitiveWord
	var total int64

	query := r.db.Model(&SensitiveWord{}).Order("created_at DESC")
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if enabled != nil {
		query = query.Where("enabled = ?", *enabled)
	}
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Find(&words).Error
	return words, total, err
}

// Update 更新敏感词
func (r *SensitiveWordRepository) Update(id uint, updates map[string]interface{}) error {
	if word, ok := updates["word"]; ok {
		updates["word"] = strings.TrimSpace(word.(string))
	}
	return r.db.Model(&SensitiveWord{}).Where("id = ?", id).Updates(updates).Error
}

// Delete 删除敏感词
func (r *SensitiveWordRepository) Delete(id uint) error {
	return r.db.Delete(&SensitiveWord{}, id).Error
}

// FindAllEnabled 获取所有启用的敏感词
func (r *SensitiveWordRepository) FindAllEnabled() ([]SensitiveWord, error) {
	var words []SensitiveWord
	err := r.db.Where("enabled = ?", true).Find(&words).Error
	return words, err
}

// CheckText 检查文本是否包含敏感词
func (r *SensitiveWordRepository) CheckText(text string) ([]string, error) {
	var words []SensitiveWord
	if err := r.db.Where("enabled = ?", true).Find(&words).Error; err != nil {
		return nil, err
	}

	var found []string
	for _, w := range words {
		if strings.Contains(text, w.Word) {
			found = append(found, w.Word)
		}
	}
	return found, nil
}
