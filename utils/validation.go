package utils

import (
	"regexp"
)

// ValidatePhone 验证手机号格式（中国大陆）
func ValidatePhone(phone string) bool {
	pattern := `^1[3-9]\d{9}$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(phone)
}

// GenerateCode 生成6位随机验证码（这里预留接口，实际生成在service层）
func GenerateCode() string {
	// 实际生成逻辑在service层
	return ""
}
