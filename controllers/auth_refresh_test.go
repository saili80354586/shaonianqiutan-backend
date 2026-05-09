package controllers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/routes"
	"github.com/shaonianqiutan/backend/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const authRefreshTestSecret = "auth-refresh-test-secret"

func setupAuthRefreshTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()

	t.Setenv("JWT_SECRET", authRefreshTestSecret)
	t.Setenv("JWT_EXPIRES_IN", "168h")
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "auth-refresh.db")), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Analyst{},
		&models.Scout{},
		&models.Club{},
		&models.ClubCoach{},
		&models.Team{},
		&models.TeamCoach{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	authService := services.NewAuthService(models.NewUserRepository(db), nil, nil, nil, nil, nil, db)
	authController := controllers.NewAuthController(authService, nil)

	router := gin.New()
	api := router.Group("/api")
	routes.SetupAuthRoutes(api, authController)
	return router, db
}

func signAuthRefreshTestToken(t *testing.T, userID uint, phone string, expiresAt time.Time) string {
	t.Helper()

	claims := middleware.Claims{
		UserID: userID,
		Phone:  phone,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(expiresAt.Add(-time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(authRefreshTestSecret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

func seedAuthRefreshUser(t *testing.T, db *gorm.DB) models.User {
	t.Helper()

	user := models.User{
		Phone:       "13900009999",
		Password:    "hashed-password",
		Role:        models.RoleUser,
		CurrentRole: models.RoleUser,
		Status:      models.StatusActive,
		Name:        "刷新测试用户",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}

func TestRefreshTokenAllowsRecentlyExpiredBearerToken(t *testing.T) {
	router, db := setupAuthRefreshTestRouter(t)
	user := seedAuthRefreshUser(t, db)
	expiredToken := signAuthRefreshTestToken(t, user.ID, user.Phone, time.Now().Add(-time.Hour))

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh-token", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Success bool `json:"success"`
		Data    struct {
			Token string `json:"token"`
			User  struct {
				ID uint `json:"id"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !body.Success || body.Data.Token == "" {
		t.Fatalf("expected refresh success with token, got %+v", body)
	}
	if body.Data.User.ID != user.ID {
		t.Fatalf("expected user %d, got %d", user.ID, body.Data.User.ID)
	}
}

func TestRefreshTokenRejectsExpiredTokenOutsideGraceWindow(t *testing.T) {
	router, db := setupAuthRefreshTestRouter(t)
	user := seedAuthRefreshUser(t, db)
	expiredToken := signAuthRefreshTestToken(t, user.ID, user.Phone, time.Now().Add(-25*time.Hour))

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh-token", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestRefreshTokenRequiresBearerToken(t *testing.T) {
	router, _ := setupAuthRefreshTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh-token", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}
