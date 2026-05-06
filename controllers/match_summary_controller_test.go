package controllers

import (
	"testing"

	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/repositories"
	"github.com/shaonianqiutan/backend/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newMatchSummaryControllerTestDB(t *testing.T) (*MatchSummaryController, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Club{},
		&models.Team{},
		&models.TeamPlayer{},
		&models.MatchSummary{},
		&models.PlayerReview{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	service := services.NewMatchSummaryService(
		db,
		repositories.NewMatchSummaryRepository(db),
		repositories.NewPlayerReviewRepository(db),
		repositories.NewMatchVideoRepository(db),
		repositories.NewTeamRepository(db),
		models.NewUserRepository(db),
	)

	return NewMatchSummaryController(service, db), db
}

func TestBuildDetailResponseIncludesPlayersFromUserIDs(t *testing.T) {
	controller, db := newMatchSummaryControllerTestDB(t)

	coach := models.User{ID: 900, Phone: "coach", Password: "x", Name: "王教练", Role: models.RoleCoach, Status: models.StatusActive}
	players := []models.User{
		{ID: 1001, Phone: "player-1", Password: "x", Name: "林子墨", Avatar: "/avatars/lin.png", Role: models.RoleUser, Status: models.StatusActive},
		{ID: 1002, Phone: "player-2", Password: "x", Name: "周宇航", Avatar: "/avatars/zhou.png", Role: models.RoleUser, Status: models.StatusActive},
	}
	if err := db.Create(&coach).Error; err != nil {
		t.Fatalf("seed coach: %v", err)
	}
	if err := db.Create(&players).Error; err != nil {
		t.Fatalf("seed players: %v", err)
	}

	team := models.Team{ID: 300, ClubID: 1, Name: "U12 精英队", AgeGroup: "U12", Status: models.TeamStatusActive}
	if err := db.Create(&team).Error; err != nil {
		t.Fatalf("seed team: %v", err)
	}
	if err := db.Create(&[]models.TeamPlayer{
		{ID: 11, TeamID: team.ID, UserID: players[0].ID, JerseyNumber: "7", Position: "边锋", Status: "active"},
		{ID: 12, TeamID: team.ID, UserID: players[1].ID, JerseyNumber: "10", Position: "中场", Status: "active"},
	}).Error; err != nil {
		t.Fatalf("seed team players: %v", err)
	}

	summary := &models.MatchSummary{
		ID:          500,
		TeamID:      team.ID,
		Team:        &team,
		CoachID:     coach.ID,
		Coach:       &coach,
		MatchName:   "测试比赛",
		MatchDate:   "2026-04-30",
		Opponent:    "测试对手",
		Location:    "home",
		MatchFormat: "11人制",
		PlayerIDs:   "[1001,1002]",
		PlayerCount: 2,
		Status:      "pending",
	}

	resp := controller.buildDetailResponse(summary)

	if resp.PlayerCount != 2 {
		t.Fatalf("expected playerCount=2, got %d", resp.PlayerCount)
	}
	if len(resp.PlayerIDs) != 2 || resp.PlayerIDs[0] != 1001 || resp.PlayerIDs[1] != 1002 {
		t.Fatalf("expected playerIds to keep user IDs [1001 1002], got %#v", resp.PlayerIDs)
	}
	if len(resp.Players) != 2 {
		t.Fatalf("expected two player detail rows, got %#v", resp.Players)
	}

	assertPlayer := func(index int, id uint, name string, number int, position string) {
		t.Helper()
		player := resp.Players[index]
		if player.ID != id || player.Name != name || player.Number != number || player.Position != position {
			t.Fatalf("unexpected player at index %d: %#v", index, player)
		}
	}
	assertPlayer(0, 1001, "林子墨", 7, "边锋")
	assertPlayer(1, 1002, "周宇航", 10, "中场")
}
