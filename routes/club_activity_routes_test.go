package routes_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/config"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/routes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupClubActivityRouteTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()

	t.Setenv("JWT_SECRET", "club-activity-route-test-secret")
	t.Setenv("JWT_EXPIRES_IN", "168h")
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "club-activity-route.db")), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.Club{},
		&models.ClubActivity{},
		&models.ClubActivityRegistration{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	oldDB := config.DB
	config.DB = db
	t.Cleanup(func() {
		config.DB = oldDB
	})

	router := gin.New()
	api := router.Group("/api")
	routes.SetupClubActivityRoutes(api, controllers.NewClubActivityController(db, nil))
	return router, db
}

func createClubActivityRouteTestUser(t *testing.T, db *gorm.DB, phone string, role models.UserRole) models.User {
	t.Helper()

	user := models.User{
		Phone:       phone,
		Password:    "hashed-password",
		Role:        role,
		CurrentRole: role,
		Status:      models.StatusActive,
		Name:        "活动权限测试用户",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}

func createClubActivityRouteTestClub(t *testing.T, db *gorm.DB, user models.User) models.Club {
	t.Helper()

	club := models.Club{
		UserID: user.ID,
		Name:   "活动权限测试俱乐部",
	}
	if err := db.Create(&club).Error; err != nil {
		t.Fatalf("create club: %v", err)
	}
	return club
}

func createClubActivityRouteTestActivity(t *testing.T, db *gorm.DB, clubID uint, title, publishStatus string) models.ClubActivity {
	t.Helper()

	activity := models.ClubActivity{
		ClubID:        clubID,
		Title:         title,
		Type:          "trial",
		Status:        "upcoming",
		Description:   "测试活动",
		StartTime:     time.Now().Add(24 * time.Hour),
		EndTime:       time.Now().Add(26 * time.Hour),
		Location:      "上海市浦东新区",
		PublishStatus: publishStatus,
	}
	if err := db.Create(&activity).Error; err != nil {
		t.Fatalf("create activity: %v", err)
	}
	return activity
}

func performClubActivityRouteRequest(t *testing.T, router *gin.Engine, method, path string, user *models.User) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, nil)
	if user != nil {
		token, err := middleware.GenerateToken(user.ID, user.Phone)
		if err != nil {
			t.Fatalf("generate token: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestClubActivityPublicListRejectsAllPublishStatusWithoutOwner(t *testing.T) {
	router, db := setupClubActivityRouteTestRouter(t)
	clubOwner := createClubActivityRouteTestUser(t, db, "13900007201", models.RoleClub)
	club := createClubActivityRouteTestClub(t, db, clubOwner)
	createClubActivityRouteTestActivity(t, db, club.ID, "公开活动", "published")
	createClubActivityRouteTestActivity(t, db, club.ID, "草稿活动", "draft")

	rec := performClubActivityRouteRequest(t, router, http.MethodGet, "/api/clubs/1/activities?publishStatus=all", nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestClubActivityOwnerCanListAllPublishStatuses(t *testing.T) {
	router, db := setupClubActivityRouteTestRouter(t)
	clubOwner := createClubActivityRouteTestUser(t, db, "13900007202", models.RoleClub)
	club := createClubActivityRouteTestClub(t, db, clubOwner)
	createClubActivityRouteTestActivity(t, db, club.ID, "公开活动", "published")
	createClubActivityRouteTestActivity(t, db, club.ID, "草稿活动", "draft")

	rec := performClubActivityRouteRequest(t, router, http.MethodGet, "/api/clubs/1/activities?publishStatus=all", &clubOwner)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Success bool                     `json:"success"`
		Data    []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.Success || len(resp.Data) != 2 {
		t.Fatalf("unexpected response: success=%v count=%d body=%s", resp.Success, len(resp.Data), rec.Body.String())
	}
}

func TestClubActivityRegistrationsRejectAuthenticatedPlayer(t *testing.T) {
	router, db := setupClubActivityRouteTestRouter(t)
	clubOwner := createClubActivityRouteTestUser(t, db, "13900007203", models.RoleClub)
	player := createClubActivityRouteTestUser(t, db, "13900007204", models.RoleUser)
	club := createClubActivityRouteTestClub(t, db, clubOwner)
	activity := createClubActivityRouteTestActivity(t, db, club.ID, "公开活动", "published")
	if err := db.Create(&models.ClubActivityRegistration{
		ActivityID: activity.ID,
		Name:       "报名用户",
		Phone:      "13800000000",
		Status:     "pending",
	}).Error; err != nil {
		t.Fatalf("create registration: %v", err)
	}

	rec := performClubActivityRouteRequest(t, router, http.MethodGet, "/api/clubs/1/activities/1/registrations", &player)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}
