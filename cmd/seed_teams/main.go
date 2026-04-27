package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Team struct {
	ID           uint      `gorm:"primaryKey"`
	Name         string    `gorm:"not null"`
	ClubID       uint      `gorm:"not null"`
	AgeGroup     string    `gorm:"not null"`
	Description  string
	BirthYearStart int
	BirthYearEnd   int
	Status       string    `gorm:"default:'active'"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type TeamCoach struct {
	ID       uint `gorm:"primaryKey"`
	TeamID   uint `gorm:"not null"`
	CoachID  uint `gorm:"not null"`
	Role     string `gorm:"not null"` // head_coach, assistant, goalkeeper_coach, fitness_coach
	JoinedAt time.Time
}

type TeamPlayer struct {
	ID       uint `gorm:"primaryKey"`
	TeamID   uint `gorm:"not null"`
	PlayerID uint `gorm:"not null"`
	JerseyNumber int
	JoinedAt time.Time
}

func main() {
	db, err := gorm.Open(sqlite.Open("../../shaonianqiutan.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}

	// 自动迁移表结构
	db.AutoMigrate(&Team{}, &TeamCoach{}, &TeamPlayer{})

	// 为俱乐部 777 创建球队
	clubID := uint(777)
	
	teams := []Team{
		{Name: "U12一队", ClubID: clubID, AgeGroup: "U12", Description: "U12年龄段主力球队", BirthYearStart: 2013, BirthYearEnd: 2014},
		{Name: "U12二队", ClubID: clubID, AgeGroup: "U12", Description: "U12年龄段预备队", BirthYearStart: 2013, BirthYearEnd: 2014},
		{Name: "U14精英队", ClubID: clubID, AgeGroup: "U14", Description: "U14精英培养队", BirthYearStart: 2011, BirthYearEnd: 2012},
	}

	fmt.Println("开始创建测试球队...")
	
	for _, team := range teams {
		var existing Team
		result := db.Where("name = ? AND club_id = ?", team.Name, team.ClubID).First(&existing)
		
		if result.Error != nil {
			// 创建新球队
			team.Status = "active"
			team.CreatedAt = time.Now()
			team.UpdatedAt = time.Now()
			db.Create(&team)
			fmt.Printf("✅ 创建球队: %s (ID: %d)\n", team.Name, team.ID)
		} else {
			fmt.Printf("ℹ️ 球队已存在: %s (ID: %d)\n", existing.Name, existing.ID)
		}
	}

	// 获取第一个球队 ID 并关联教练
	var firstTeam Team
	db.Where("club_id = ?", clubID).First(&firstTeam)
	
	if firstTeam.ID > 0 {
		// 关联教练 (ID: 666)
		var coach TeamCoach
		result := db.Where("team_id = ? AND coach_id = ?", firstTeam.ID, 666).First(&coach)
		if result.Error != nil {
			db.Create(&TeamCoach{
				TeamID:   firstTeam.ID,
				CoachID:  666,
				Role:     "head_coach",
				JoinedAt: time.Now(),
			})
			fmt.Printf("✅ 关联教练 666 到球队 %s\n", firstTeam.Name)
		}

		// 关联一些球员 (ID: 5, 6, 7)
		playerIDs := []uint{5, 6, 7}
		for i, pid := range playerIDs {
			var tp TeamPlayer
			result := db.Where("team_id = ? AND player_id = ?", firstTeam.ID, pid).First(&tp)
			if result.Error != nil {
				db.Create(&TeamPlayer{
					TeamID:       firstTeam.ID,
					PlayerID:     pid,
					JerseyNumber: i + 1,
					JoinedAt:     time.Now(),
				})
				fmt.Printf("✅ 关联球员 %d 到球队 %s (球衣号: %d)\n", pid, firstTeam.Name, i+1)
			}
		}
	}

	// 显示所有球队
	fmt.Println("\n=== 当前球队列表 ===")
	var allTeams []Team
	db.Find(&allTeams)
	for _, t := range allTeams {
		fmt.Printf("ID: %d | 名称: %s | 俱乐部: %d | 年龄段: %s\n", t.ID, t.Name, t.ClubID, t.AgeGroup)
	}

	fmt.Println("\n测试数据创建完成！")
}
