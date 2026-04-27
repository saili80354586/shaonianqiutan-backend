package services

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/shaonianqiutan/backend/config"
	"github.com/shaonianqiutan/backend/models"
)

// SmsService 短信服务
type SmsService struct {
	smsCodeRepo *models.SmsCodeRepository
	rng         *rand.Rand
}

func NewSmsService(smsCodeRepo *models.SmsCodeRepository) *SmsService {
	source := rand.NewSource(time.Now().UnixNano())
	return &SmsService{
		smsCodeRepo: smsCodeRepo,
		rng:         rand.New(source),
	}
}

// SendCodeResult 发送验证码结果
type SendCodeResult struct {
	DevMode bool   `json:"devMode"`
	Code    string `json:"code,omitempty"`
}

// GenerateCode 生成6位随机验证码
func (s *SmsService) GenerateCode() string {
	code := 100000 + s.rng.Intn(900000)
	return fmt.Sprintf("%06d", code)
}

// SendCode 发送短信验证码
func (s *SmsService) SendCode(phone, code string) (*SendCodeResult, error) {
	// 开发模式：直接返回验证码，不实际发送短信
	if config.IsDevMode() {
		return &SendCodeResult{
			DevMode: true,
			Code:    code,
		}, nil
	}

	// TODO: 生产环境对接实际短信服务商
	// 这里需要根据实际使用的短信服务商进行对接
	// 阿里云、腾讯云、华为云等

	return &SendCodeResult{
		DevMode: false,
	}, nil
}

// VerifyCode 验证验证码
func (s *SmsService) VerifyCode(phone, code string, codeType models.SmsCodeType) (*models.SmsCode, error) {
	return s.smsCodeRepo.Verify(phone, code, codeType)
}

// MarkAsUsed 标记验证码已使用
func (s *SmsService) MarkAsUsed(id uint) error {
	return s.smsCodeRepo.MarkAsUsed(id)
}

// CleanExpired 清理过期验证码
func (s *SmsService) CleanExpired() error {
	return s.smsCodeRepo.CleanExpired()
}

// CreateCode 创建验证码记录
func (s *SmsService) CreateCode(phone, code string, codeType models.SmsCodeType) (*models.SmsCode, error) {
	expiresAt := time.Now().Add(10 * time.Minute)
	return s.smsCodeRepo.Create(phone, code, codeType, expiresAt)
}
