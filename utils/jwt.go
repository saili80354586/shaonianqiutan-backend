package utils

import (
	"github.com/shaonianqiutan/backend/middleware"
)

// GenerateToken 生成JWT令牌
// 这是一个包装函数，调用 middleware 中的 GenerateToken
func GenerateToken(userID uint, phone string, role string) (string, error) {
	return middleware.GenerateToken(userID, phone)
}
