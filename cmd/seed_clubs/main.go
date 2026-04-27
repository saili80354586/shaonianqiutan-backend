package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Club struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint `gorm:"index"`
	Name      string
	Logo      string
	ClubType  string
	Address   string
	Status    string
	CreatedAt string
	UpdatedAt string
}

type User struct {
	ID        uint `gorm:"primaryKey"`
	Phone     string
	Password  string
	Nickname  string
	Role      string
	Status    string
	Name      string
	Province  string
	City      string
	CreatedAt string
	UpdatedAt string
}

func main() {
	db, err := gorm.Open(sqlite.Open("./shaonianqiutan.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// 俱乐部账号列表 - 与前端 Login.tsx 中的账号匹配
	clubs := []struct {
		id       uint
		phone    string
		password string
		nickname string
		name     string
		province string
		city     string
	}{
		{1001, "13800111001", "test", "北京国安青训", "北京国安青训俱乐部", "北京", "北京"},
		{1002, "13800111002", "test", "上海根宝足球基地", "上海根宝足球基地", "上海", "上海"},
		{1003, "13800111003", "test", "广州恒大足校", "广州恒大足球学校", "广东", "广州"},
		{1004, "13800111004", "test", "山东泰山青训", "山东泰山青训俱乐部", "山东", "济南"},
		{1005, "13800111005", "test", "江苏苏宁青训", "江苏苏宁青训俱乐部", "江苏", "南京"},
		{1006, "13800111006", "test", "成都蓉城青训", "成都蓉城青训俱乐部", "四川", "成都"},
		{1007, "13800111007", "test", "武汉三镇青训", "武汉三镇青训俱乐部", "湖北", "武汉"},
		{1008, "13800111008", "test", "浙江绿城足校", "浙江绿城足球学校", "浙江", "杭州"},
		{1009, "13800111009", "test", "河南嵩山青训", "河南嵩山青训俱乐部", "河南", "郑州"},
		{1010, "13800111010", "test", "天津津门虎青训", "天津津门虎青训俱乐部", "天津", "天津"},
	}

	fmt.Println("开始创建俱乐部账号...")

	for _, club := range clubs {
		hash, err := bcrypt.GenerateFromPassword([]byte(club.password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("密码哈希失败 %s: %v", club.phone, err)
			continue
		}

		var user User
		result := db.Where("phone = ?", club.phone).First(&user)
		if result.Error == gorm.ErrRecordNotFound {
			user = User{
				ID:        club.id,
				Phone:     club.phone,
				Password:  string(hash),
				Nickname:  club.nickname,
				Role:      "club",
				Status:    "active",
				Name:      club.name,
				Province:  club.province,
				City:      club.city,
			}
			if err := db.Create(&user).Error; err != nil {
				fmt.Printf("❌ 创建用户失败 %s: %v\n", club.phone, err)
			} else {
				fmt.Printf("✅ 创建俱乐部用户: %s (%s) - 密码: %s\n", club.phone, club.nickname, club.password)
			}
		} else {
			// 更新已有用户
			db.Model(&user).Updates(map[string]interface{}{
				"password":    string(hash),
				"role":        "club",
				"status":      "active",
				"nickname":    club.nickname,
				"name":        club.name,
				"province":    club.province,
				"city":        club.city,
			})
			fmt.Printf("🔄 更新俱乐部用户: %s (%s) - 密码: %s\n", club.phone, club.nickname, club.password)
		}
	}

	fmt.Println("\n✅ 俱乐部账号创建完成!")
	fmt.Println("\n可用的俱乐部测试账号:")
	for _, club := range clubs {
		fmt.Printf("  %s / %s (%s)\n", club.phone, club.password, club.nickname)
	}
}
