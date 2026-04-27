package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(sqlite.Open("./shaonianqiutan.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	type User struct {
		ID    uint `gorm:"primaryKey"`
		Phone string
		Password string
		Nickname string
		Role string
		Status string
		Name string
	}

	accounts := []struct {
		id       uint
		phone    string
		password string
		nickname string
		role     string
		name     string
	}{
		{999, "admin", "admin123", "管理员", "admin", "系统管理员"},
		{888, "analyst", "analyst123", "分析师", "analyst", "专业分析师"},
		{777, "club", "club123", "俱乐部", "club", "上海绿地俱乐部"},
		{666, "coach", "coach123", "教练", "coach", "王教练"},
	}

	for _, acc := range accounts {
		hash, err := bcrypt.GenerateFromPassword([]byte(acc.password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal(err)
		}

		var user User
		result := db.Where("phone = ?", acc.phone).First(&user)
		if result.Error == gorm.ErrRecordNotFound {
			user = User{
				ID:       acc.id,
				Phone:    acc.phone,
				Password: string(hash),
				Nickname: acc.nickname,
				Role:     acc.role,
				Status:   "active",
				Name:     acc.name,
			}
			if err := db.Create(&user).Error; err != nil {
				fmt.Printf("Insert %s failed: %v\n", acc.phone, err)
			} else {
				fmt.Printf("Inserted %s (%s) - pwd: %s\n", acc.phone, acc.role, acc.password)
			}
		} else {
			// Update password
			db.Model(&user).Updates(map[string]interface{}{
				"password": string(hash),
				"role":     acc.role,
				"status":   "active",
			})
			fmt.Printf("Updated %s (%s) - pwd: %s\n", acc.phone, acc.role, acc.password)
		}
	}

	// Also fix user 13800110005 password to "test"
	var user5 User
	if err := db.Where("phone = ?", "13800110005").First(&user5).Error; err == nil {
		// Try "test" hash
		hash, _ := bcrypt.GenerateFromPassword([]byte("test"), bcrypt.DefaultCost)
		db.Model(&user5).Update("password", string(hash))
		fmt.Println("Fixed 13800110005 password to 'test'")
	}

	// Update 13800110006-08 to have consistent password
	for _, phone := range []string{"13800110006", "13800110007", "13800110008"} {
		var u User
		if err := db.Where("phone = ?", phone).First(&u).Error; err == nil {
			hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
			db.Model(&u).Update("password", string(hash))
			fmt.Printf("Updated %s password to 'password'\n", phone)
		}
	}

	fmt.Println("\nDone! Demo accounts created:")
	fmt.Println("  admin    / admin123    (管理员)")
	fmt.Println("  analyst  / analyst123  (分析师)")
	fmt.Println("  club     / club123     (俱乐部)")
	fmt.Println("  coach    / coach123     (教练)")
	fmt.Println("  13800110006 / password (球员-深圳闪电)")
	fmt.Println("  13800110007 / password (球员-北京小王子)")
	fmt.Println("  13800110008 / password (球员-上海闪电侠)")
}
