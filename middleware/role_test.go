package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/config"
	"github.com/shaonianqiutan/backend/middleware"
	"github.com/shaonianqiutan/backend/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRoleMiddlewareTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()

	t.Setenv("JWT_SECRET", "role-middleware-test-secret")
	t.Setenv("JWT_EXPIRES_IN", "168h")
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "role-middleware.db")), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Scout{},
		&models.Coach{},
		&models.ClubCoach{},
		&models.TeamCoach{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	oldDB := config.DB
	config.DB = db
	t.Cleanup(func() {
		config.DB = oldDB
	})

	router := gin.New()
	router.GET("/scout/probe", middleware.AuthMiddleware(), middleware.ScoutRoleMiddleware(), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	router.GET("/coach/probe", middleware.AuthMiddleware(), middleware.CoachRoleMiddleware(), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	return router, db
}

func createRoleMiddlewareTestUser(t *testing.T, db *gorm.DB, phone string, role models.UserRole) models.User {
	t.Helper()

	user := models.User{
		Phone:       phone,
		Password:    "hashed-password",
		Role:        role,
		CurrentRole: role,
		Status:      models.StatusActive,
		Name:        "角色权限测试用户",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}

func performRoleMiddlewareRequest(t *testing.T, router *gin.Engine, method, path string, user models.User) *httptest.ResponseRecorder {
	t.Helper()

	token, err := middleware.GenerateToken(user.ID, user.Phone)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestScoutRoleMiddlewareRejectsAuthenticatedPlayerWithoutCreatingScout(t *testing.T) {
	router, db := setupRoleMiddlewareTestRouter(t)
	user := createRoleMiddlewareTestUser(t, db, "13900007001", models.RoleUser)

	rec := performRoleMiddlewareRequest(t, router, http.MethodGet, "/scout/probe", user)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var scoutCount int64
	if err := db.Model(&models.Scout{}).Where("user_id = ?", user.ID).Count(&scoutCount).Error; err != nil {
		t.Fatalf("count scouts: %v", err)
	}
	if scoutCount != 0 {
		t.Fatalf("expected no scout profile to be created, got %d", scoutCount)
	}
}

func TestScoutRoleMiddlewareAllowsScoutPrimaryRole(t *testing.T) {
	router, db := setupRoleMiddlewareTestRouter(t)
	user := createRoleMiddlewareTestUser(t, db, "13900007002", models.RoleScout)

	rec := performRoleMiddlewareRequest(t, router, http.MethodGet, "/scout/probe", user)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestCoachRoleMiddlewareRejectsCoachProfileWithoutRoleMembership(t *testing.T) {
	router, db := setupRoleMiddlewareTestRouter(t)
	user := createRoleMiddlewareTestUser(t, db, "13900007003", models.RoleUser)
	if err := db.Create(&models.Coach{UserID: user.ID}).Error; err != nil {
		t.Fatalf("create coach profile: %v", err)
	}

	rec := performRoleMiddlewareRequest(t, router, http.MethodGet, "/coach/probe", user)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestCoachRoleMiddlewareAllowsTeamCoachMembership(t *testing.T) {
	router, db := setupRoleMiddlewareTestRouter(t)
	user := createRoleMiddlewareTestUser(t, db, "13900007004", models.RoleUser)
	if err := db.Create(&models.TeamCoach{
		TeamID:   1,
		UserID:   user.ID,
		Role:     models.CoachRoleAssistant,
		Status:   "active",
		JoinedAt: time.Now(),
	}).Error; err != nil {
		t.Fatalf("create team coach: %v", err)
	}

	rec := performRoleMiddlewareRequest(t, router, http.MethodGet, "/coach/probe", user)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}
