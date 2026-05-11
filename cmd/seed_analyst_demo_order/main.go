package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/shaonianqiutan/backend/config"
	"github.com/shaonianqiutan/backend/models"
)

func main() {
	config.LoadEnv()
	config.InitDB()
	db := config.GetDB()

	playerPhone := strings.TrimSpace(os.Getenv("DEMO_PLAYER_PHONE"))
	playerName := strings.TrimSpace(os.Getenv("DEMO_PLAYER_NAME"))
	playerBirthDate := strings.TrimSpace(os.Getenv("DEMO_PLAYER_BIRTH_DATE"))
	matchName := strings.TrimSpace(os.Getenv("DEMO_MATCH_NAME"))
	matchDate := strings.TrimSpace(os.Getenv("DEMO_MATCH_DATE"))
	opponent := strings.TrimSpace(os.Getenv("DEMO_OPPONENT"))
	matchResult := strings.TrimSpace(os.Getenv("DEMO_MATCH_RESULT"))
	playerPosition := strings.TrimSpace(os.Getenv("DEMO_PLAYER_POSITION"))
	jerseyColor := strings.TrimSpace(os.Getenv("DEMO_JERSEY_COLOR"))
	jerseyNumber := strings.TrimSpace(os.Getenv("DEMO_JERSEY_NUMBER"))
	videoURL := strings.TrimSpace(os.Getenv("DEMO_VIDEO_URL"))
	videoSource := strings.TrimSpace(os.Getenv("DEMO_VIDEO_SOURCE_FILE"))
	videoFilename := strings.TrimSpace(os.Getenv("DEMO_VIDEO_FILENAME"))
	orderType := strings.TrimSpace(os.Getenv("DEMO_ORDER_TYPE"))
	if orderType == "" {
		orderType = "pro"
	}

	if playerPhone == "" || playerName == "" || matchName == "" || playerPosition == "" || jerseyColor == "" || jerseyNumber == "" {
		log.Fatal("DEMO_PLAYER_PHONE, DEMO_PLAYER_NAME, DEMO_MATCH_NAME, DEMO_PLAYER_POSITION, DEMO_JERSEY_COLOR, DEMO_JERSEY_NUMBER are required")
	}

	if videoURL == "" && videoSource == "" {
		log.Fatal("DEMO_VIDEO_URL or DEMO_VIDEO_SOURCE_FILE is required")
	}

	if videoURL == "" {
		var err error
		videoURL, videoFilename, err = copyVideoToUploads(videoSource, videoFilename)
		if err != nil {
			log.Fatalf("copy video: %v", err)
		}
	}
	if videoFilename == "" {
		videoFilename = filepath.Base(videoURL)
	}

	var player models.User
	if err := db.Where("phone = ? AND role = ? AND status = ?", playerPhone, models.RoleUser, models.StatusActive).First(&player).Error; err != nil {
		log.Fatalf("find player %s: %v", playerPhone, err)
	}
	if playerBirthDate != "" {
		if player.BirthDate != playerBirthDate || player.Age == 0 {
			player.BirthDate = playerBirthDate
			player.Age = calculateAgeFromBirthDate(playerBirthDate)
			if err := db.Model(&player).Updates(map[string]any{
				"birth_date": playerBirthDate,
				"age":        player.Age,
			}).Error; err != nil {
				log.Fatalf("update player birth date: %v", err)
			}
		}
	}

	amount := parseAmount(os.Getenv("DEMO_ORDER_AMOUNT"), 0)
	duration := parseIntEnv("DEMO_VIDEO_DURATION", 0)
	remark := strings.TrimSpace(os.Getenv("DEMO_ORDER_REMARK"))
	if remark == "" {
		remark = "系统样例模板订单"
	}

	now := time.Now()
	order := &models.Order{
		UserID:         player.ID,
		OrderNo:        fmt.Sprintf("TPL%d%04d", now.UnixNano(), player.ID%10000),
		Amount:         amount,
		Status:         models.OrderStatusUploaded,
		PaymentMethod:  models.PaymentMethodWechat,
		PaidAt:         &now,
		VideoURL:       videoURL,
		VideoFilename:  videoFilename,
		Remark:         remark,
		OrderType:      orderType,
		PlayerName:     playerName,
		PlayerAge:      player.Age,
		PlayerPosition: playerPosition,
		JerseyColor:    jerseyColor,
		JerseyNumber:   jerseyNumber,
		MatchName:      matchName,
		MatchDate:      matchDate,
		Opponent:       opponent,
		MatchResult:    matchResult,
		VideoDuration:  duration,
	}

	if err := db.Create(order).Error; err != nil {
		log.Fatalf("create template order: %v", err)
	}

	fmt.Printf("created template order id=%d\n", order.ID)
	fmt.Printf("set ANALYST_DEFAULT_DEMO_ORDER_ENABLED=true\n")
	fmt.Printf("set ANALYST_DEFAULT_DEMO_ORDER_TEMPLATE_ORDER_ID=%d\n", order.ID)
}

func calculateAgeFromBirthDate(birthDate string) int {
	if birthDate == "" {
		return 0
	}
	layouts := []string{"2006-01-02", "2006/01/02", "01-02-2006"}
	var t time.Time
	var err error
	for _, layout := range layouts {
		t, err = time.Parse(layout, birthDate)
		if err == nil {
			break
		}
	}
	if err != nil {
		return 0
	}
	now := time.Now()
	age := now.Year() - t.Year()
	if now.YearDay() < t.YearDay() {
		age--
	}
	return age
}

func copyVideoToUploads(sourcePath, videoFilename string) (string, string, error) {
	src, err := os.Open(sourcePath)
	if err != nil {
		return "", "", err
	}
	defer src.Close()

	if videoFilename == "" {
		videoFilename = filepath.Base(sourcePath)
	}

	ext := filepath.Ext(videoFilename)
	if ext == "" {
		ext = filepath.Ext(sourcePath)
	}
	if ext == "" {
		ext = ".mp4"
	}
	if !strings.HasSuffix(strings.ToLower(videoFilename), strings.ToLower(ext)) {
		videoFilename += ext
	}

	dstName := fmt.Sprintf("analyst-demo-%d-%s", time.Now().UnixNano(), videoFilename)
	uploadDir := filepath.Join(".", "uploads", "videos")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", "", err
	}

	dstPath := filepath.Join(uploadDir, dstName)
	dst, err := os.Create(dstPath)
	if err != nil {
		return "", "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", "", err
	}

	baseURL := config.GetBaseUrl()
	return fmt.Sprintf("%s/uploads/videos/%s", baseURL, dstName), dstName, nil
}

func parseAmount(raw string, fallback float64) float64 {
	if strings.TrimSpace(raw) == "" {
		return fallback
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return fallback
	}
	return value
}

func parseIntEnv(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}
