package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// migrateMultiLayerMap 为多图层地图功能执行数据库迁移
// 功能：
//   1. 为 coaches 表添加 city 字段（如缺失）
//   2. 为 clubs 表添加 province 和 city 字段（如缺失）
//   3. 从 users 表回填 coaches/clubs 的城市数据
//   4. 验证所有角色地理位置数据完整性
//
// 用法：
//   cd cmd/migrate_multi_layer_map && go run main.go
//   或：go run cmd/migrate_multi_layer_map/main.go
func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		// 默认从项目根目录运行
		dbPath = "./shaonianqiutan.db"
	}

	fmt.Printf("连接到数据库: %s\n", dbPath)
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("获取原生 DB 失败:", err)
	}
	defer sqlDB.Close()

	fmt.Println("\n=== Phase 1: Schema Repair ===")

	// 1. coaches 表添加 city 字段
	fmt.Println("\n[1/4] 检查 coaches 表 city 字段...")
	if err := db.Exec("ALTER TABLE coaches ADD COLUMN city TEXT DEFAULT ''").Error; err != nil {
		fmt.Printf("  ⚠️  coaches.city 可能已存在: %v\n", err)
	} else {
		fmt.Println("  ✅ coaches.city 添加成功")
	}

	// 2. clubs 表添加 province/city 字段
	fmt.Println("\n[2/4] 检查 clubs 表 province/city 字段...")
	if err := db.Exec("ALTER TABLE clubs ADD COLUMN province TEXT DEFAULT ''").Error; err != nil {
		fmt.Printf("  ⚠️  clubs.province 可能已存在: %v\n", err)
	} else {
		fmt.Println("  ✅ clubs.province 添加成功")
	}
	if err := db.Exec("ALTER TABLE clubs ADD COLUMN city TEXT DEFAULT ''").Error; err != nil {
		fmt.Printf("  ⚠️  clubs.city 可能已存在: %v\n", err)
	} else {
		fmt.Println("  ✅ clubs.city 添加成功")
	}

	// 3. 从 users 表回填 coaches.city
	fmt.Println("\n[3/4] 回填 coaches 城市数据...")
	result := db.Exec(`
		UPDATE coaches 
		SET city = COALESCE((SELECT city FROM users WHERE users.id = coaches.user_id), '')
		WHERE city = '' OR city IS NULL
	`)
	if result.Error != nil {
		log.Printf("  ❌ 回填 coaches 失败: %v", result.Error)
	} else {
		fmt.Printf("  ✅ 更新 %d 条 coaches 记录\n", result.RowsAffected)
	}

	// 4. 从 users 表回填 clubs.province/city
	fmt.Println("\n[4/4] 回填 clubs 城市数据...")
	result = db.Exec(`
		UPDATE clubs 
		SET 
			province = COALESCE((SELECT province FROM users WHERE users.id = clubs.user_id), ''),
			city = COALESCE((SELECT city FROM users WHERE users.id = clubs.user_id), '')
		WHERE (province = '' OR province IS NULL) AND (city = '' OR city IS NULL)
	`)
	if result.Error != nil {
		log.Printf("  ❌ 回填 clubs 失败: %v", result.Error)
	} else {
		fmt.Printf("  ✅ 更新 %d 条 clubs 记录\n", result.RowsAffected)
	}

	// 验证所有角色数据完整性
	fmt.Println("\n=== Phase 2: Data Validation ===")
	validateData(db)

	fmt.Println("\n✅ 迁移完成！")
}

func validateData(db *gorm.DB) {
	queries := []struct {
		name  string
		query string
	}{
		{
			name: "球员 (users)",
			query: `
				SELECT COUNT(*) as total, 
					COUNT(CASE WHEN province IS NOT NULL AND province != '' AND city IS NOT NULL AND city != '' THEN 1 END) as has_loc 
				FROM users WHERE role = 'user'
			`,
		},
		{
			name: "分析师 (analysts → users)",
			query: `
				SELECT COUNT(*) as total, 
					COUNT(CASE WHEN u.province IS NOT NULL AND u.province != '' AND u.city IS NOT NULL AND u.city != '' THEN 1 END) as has_loc 
				FROM analysts a JOIN users u ON a.user_id = u.id
			`,
		},
		{
			name: "球探 (scouts → users)",
			query: `
				SELECT COUNT(*) as total, 
					COUNT(CASE WHEN u.province IS NOT NULL AND u.province != '' AND u.city IS NOT NULL AND u.city != '' THEN 1 END) as has_loc 
				FROM scouts s JOIN users u ON s.user_id = u.id
			`,
		},
		{
			name: "教练 (coaches)",
			query: `
				SELECT COUNT(*) as total, 
					COUNT(CASE WHEN city IS NOT NULL AND city != '' THEN 1 END) as has_loc 
				FROM coaches
			`,
		},
		{
			name: "俱乐部 (clubs)",
			query: `
				SELECT COUNT(*) as total, 
					COUNT(CASE WHEN province IS NOT NULL AND province != '' AND city IS NOT NULL AND city != '' THEN 1 END) as has_loc 
				FROM clubs
			`,
		},
	}

	for _, q := range queries {
		var total, hasLoc int64
		db.Raw(q.query).Row().Scan(&total, &hasLoc)
		status := "✅"
		if total > 0 && hasLoc < total {
			status = "⚠️"
		}
		fmt.Printf("  %s %s: %d/%d (%.0f%%)\n", status, q.name, hasLoc, total, float64(hasLoc)/float64(total)*100)
	}
}
