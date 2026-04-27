package models

import (
	"time"

	"gorm.io/gorm"
)

// SmsCodeType 验证码类型
type SmsCodeType string

const (
	SmsCodeTypeRegister SmsCodeType = "register"
	SmsCodeTypeReset    SmsCodeType = "reset"
)

// SmsCode 短信验证码模型
type SmsCode struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Phone     string         `json:"phone" gorm:"size:20;not null;index"`
	Code      string         `json:"code" gorm:"size:6;not null"`
	Type      SmsCodeType    `json:"type" gorm:"size:20;not null"`
	IsUsed    bool           `json:"is_used" gorm:"default:false"`
	ExpiresAt time.Time      `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// SmsCodeRepository 短信验证码数据访问层
type SmsCodeRepository struct {
	db *gorm.DB
}

func NewSmsCodeRepository(db *gorm.DB) *SmsCodeRepository {
	return &SmsCodeRepository{db: db}
}

// Create 创建验证码
func (r *SmsCodeRepository) Create(phone, code string, codeType SmsCodeType, expiresAt time.Time) (*SmsCode, error) {
	smsCode := &SmsCode{
		Phone:     phone,
		Code:      code,
		Type:      codeType,
		IsUsed:    false,
		ExpiresAt: expiresAt,
	}
	err := r.db.Create(smsCode).Error
	if err != nil {
		return nil, err
	}
	return smsCode, nil
}

// Verify 验证验证码
func (r *SmsCodeRepository) Verify(phone, code string, codeType SmsCodeType) (*SmsCode, error) {
	var smsCode SmsCode
	result := r.db.Where(
		"phone = ? AND code = ? AND type = ? AND is_used = ? AND expires_at > ?",
		phone, code, codeType, false, time.Now(),
	).Order("id DESC").First(&smsCode)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &smsCode, nil
}

// MarkAsUsed 标记为已使用
func (r *SmsCodeRepository) MarkAsUsed(id uint) error {
	return r.db.Model(&SmsCode{}).Where("id = ?", id).Update("is_used", true).Error
}

// CleanExpired 清理过期验证码
func (r *SmsCodeRepository) CleanExpired() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&SmsCode{}).Error
}
