package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Phone    string `gorm:"uniqueIndex"`
	Password string
	Role     string
	Status   string
}

func main() {
	// 打开数据库（从项目根目录运行）
	db, err := gorm.Open(sqlite.Open("../../shaonianqiutan.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}

	// 定义要重置的账号 (phone 字段)
	accounts := []struct {
		phone    string
		password string
		role     string
	}{
		{"club", "club123456", "club"},         // ID=777
		{"coach", "coach123456", "coach"},      // ID=666
		{"analyst", "analyst123456", "analyst"}, // ID=888
		{"admin", "admin123456", "admin"},      // ID=1
		{"13800138005", "scout123456", "scout"}, // ID=889
		{"13800138000", "player123456", "player"}, // ID=25
	}

	fmt.Println("开始重置测试账号密码...")
	
	for _, acc := range accounts {
		// 生成 bcrypt hash
		hash, err := bcrypt.GenerateFromPassword([]byte(acc.password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("密码哈希失败 %s: %v", acc.phone, err)
			continue
		}

		var user User
		result := db.Where("phone = ?", acc.phone).First(&user)
		
		if result.Error != nil {
			log.Printf("用户 %s 不存在: %v", acc.phone, result.Error)
			continue
		}

		// 更新密码
		db.Model(&user).Updates(map[string]interface{}{
			"password": string(hash),
			"status":   "active",
		})
		
		fmt.Printf("✅ 已重置 %s (%s) 的密码为: %s\n", acc.phone, acc.role, acc.password)
	}

	fmt.Println("\n密码重置完成！")
	fmt.Println("\n测试账号:")
	fmt.Println("  俱乐部: 777 / club123456")
	fmt.Println("  教练: 666 / coach123456")
	fmt.Println("  分析师: 888 / analyst123456")
	fmt.Println("  管理员: admin / admin123456")
	fmt.Println("  球探: 889 / scout123456")
	fmt.Println("  球员: 13800138000 / player123456")
}
