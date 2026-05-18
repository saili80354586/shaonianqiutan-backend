package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/shaonianqiutan/backend/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type candidate struct {
	HomeID           uint
	ClubID           uint
	ClubName         string
	ClubDescription  string
	ClubContactPhone string
	PublishStatus    string
	TemplateType     string
	CompletionScore  int
	ShareSlug        string
}

func main() {
	var dbPathFlag string
	var execute bool
	flag.StringVar(&dbPathFlag, "db", "", "SQLite database path. Defaults to DB_PATH or ./shaonianqiutan.db")
	flag.BoolVar(&execute, "execute", false, "update eligible existing club homes. Defaults to dry-run")
	flag.Parse()

	config.LoadEnv()
	dbPath := strings.TrimSpace(dbPathFlag)
	if dbPath == "" {
		dbPath = strings.TrimSpace(os.Getenv("DB_PATH"))
	}
	if dbPath == "" {
		dbPath = "./shaonianqiutan.db"
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("打开数据库失败: %v", err)
	}

	rows, err := loadCandidates(db)
	if err != nil {
		log.Fatalf("读取俱乐部主页失败: %v", err)
	}

	eligible := make([]candidate, 0, len(rows))
	skipped := make([]candidate, 0)
	for _, row := range rows {
		if isEligible(row) {
			eligible = append(eligible, row)
		} else {
			skipped = append(skipped, row)
		}
	}

	mode := "dry-run"
	if execute {
		mode = "execute"
	}
	log.Printf("mode=%s db=%s draft_candidates=%d eligible=%d skipped=%d", mode, dbPath, len(rows), len(eligible), len(skipped))
	for _, row := range eligible {
		log.Printf("eligible home_id=%d club_id=%d club=%q current_status=%q target_status=published", row.HomeID, row.ClubID, row.ClubName, row.PublishStatus)
	}
	for _, row := range skipped {
		log.Printf("skip home_id=%d club_id=%d club=%q reason=missing_club_intro_or_contact", row.HomeID, row.ClubID, row.ClubName)
	}
	if !execute {
		log.Println("dry-run only. Re-run with -execute to update eligible rows.")
		return
	}

	now := time.Now()
	for _, row := range eligible {
		updates := map[string]any{
			"publish_status": "published",
			"published_at":   now,
		}
		if strings.TrimSpace(row.TemplateType) == "" {
			updates["template_type"] = "professional"
		}
		if row.CompletionScore <= 0 {
			updates["completion_score"] = 80
		}
		if strings.TrimSpace(row.ShareSlug) == "" {
			updates["share_slug"] = fmt.Sprintf("club-%d", row.ClubID)
		}
		if err := db.Table("club_homes").Where("id = ?", row.HomeID).Updates(updates).Error; err != nil {
			log.Fatalf("更新主页 %d 失败: %v", row.HomeID, err)
		}
	}
	log.Printf("updated=%d", len(eligible))
}

func loadCandidates(db *gorm.DB) ([]candidate, error) {
	var rows []candidate
	err := db.Raw(`
		SELECT
			h.id AS home_id,
			h.club_id AS club_id,
			COALESCE(c.name, '') AS club_name,
			COALESCE(c.description, '') AS club_description,
			COALESCE(c.contact_phone, '') AS club_contact_phone,
			COALESCE(h.publish_status, '') AS publish_status,
			COALESCE(h.template_type, '') AS template_type,
			COALESCE(h.completion_score, 0) AS completion_score,
			COALESCE(h.share_slug, '') AS share_slug
		FROM club_homes h
		JOIN clubs c ON c.id = h.club_id
		WHERE COALESCE(h.publish_status, '') IN ('', 'draft')
		ORDER BY h.id
	`).Scan(&rows).Error
	return rows, err
}

func isEligible(row candidate) bool {
	if strings.TrimSpace(row.ClubName) == "" {
		return false
	}
	return strings.TrimSpace(row.ClubContactPhone) != "" || strings.TrimSpace(row.ClubDescription) != ""
}
