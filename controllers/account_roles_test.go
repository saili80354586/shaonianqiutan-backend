package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAccountRoleControllerTest(t *testing.T) (*gorm.DB, *AccountRoleController, models.User, models.User) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.UserRoleRecord{},
		&models.RoleApplication{},
		&models.Analyst{},
		&models.Scout{},
		&models.Club{},
		&models.ClubCoach{},
		&models.TeamCoach{},
	); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	user := models.User{
		Phone:    "13931009001",
		Password: "test",
		Role:     models.RoleUser,
		Status:   models.StatusActive,
		Name:     "多身份测试用户",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	admin := models.User{
		Phone:    "13931009002",
		Password: "test",
		Role:     models.RoleAdmin,
		Status:   models.StatusActive,
		Name:     "审核管理员",
	}
	if err := db.Create(&admin).Error; err != nil {
		t.Fatalf("create admin: %v", err)
	}

	return db, NewAccountRoleController(db), user, admin
}

func TestApplyRoleCreatesPendingRoleRecord(t *testing.T) {
	db, ctrl, user, _ := setupAccountRoleControllerTest(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/account/roles/apply", strings.NewReader(`{"role":"analyst","profile":{"summary":"视频分析经验"},"source":"self_apply"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user", &user)
	c.Set("userId", user.ID)

	ctrl.ApplyRole(c)

	if w.Code != http.StatusOK {
		t.Fatalf("apply role status = %d, body = %s", w.Code, w.Body.String())
	}

	var application models.RoleApplication
	if err := db.Where("user_id = ? AND role = ?", user.ID, models.RoleAnalyst).First(&application).Error; err != nil {
		t.Fatalf("find role application: %v", err)
	}
	if application.Status != models.RoleApplicationStatusPending {
		t.Fatalf("application status = %s, want pending", application.Status)
	}

	var record models.UserRoleRecord
	if err := db.Where("user_id = ? AND role = ?", user.ID, models.RoleAnalyst).First(&record).Error; err != nil {
		t.Fatalf("find role record: %v", err)
	}
	if record.Status != string(models.RoleApplicationStatusPending) {
		t.Fatalf("role record status = %s, want pending", record.Status)
	}
}

func TestReviewRoleApplicationApprovesAnalystRole(t *testing.T) {
	db, ctrl, user, admin := setupAccountRoleControllerTest(t)

	application := models.RoleApplication{
		UserID:      user.ID,
		Role:        models.RoleAnalyst,
		Status:      models.RoleApplicationStatusPending,
		Source:      "self_apply",
		ProfileJSON: `{"summary":"长期负责青训视频分析","experience":3,"contact_email":"analyst@example.com"}`,
	}
	if err := db.Create(&application).Error; err != nil {
		t.Fatalf("create application: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: uintToString(application.ID)}}
	c.Request = httptest.NewRequest(http.MethodPost, "/admin/role-applications/"+uintToString(application.ID)+"/review", strings.NewReader(`{"status":"approved"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user", &admin)
	c.Set("userId", admin.ID)

	ctrl.ReviewRoleApplication(c)

	if w.Code != http.StatusOK {
		t.Fatalf("review status = %d, body = %s", w.Code, w.Body.String())
	}

	var record models.UserRoleRecord
	if err := db.Where("user_id = ? AND role = ?", user.ID, models.RoleAnalyst).First(&record).Error; err != nil {
		t.Fatalf("find role record: %v", err)
	}
	if record.Status != "active" || record.ApprovedBy != admin.ID {
		t.Fatalf("role record = %+v, want active approved by admin", record)
	}

	var analyst models.Analyst
	if err := db.Where("user_id = ? AND status = ?", user.ID, models.AnalystStatusActive).First(&analyst).Error; err != nil {
		t.Fatalf("find active analyst profile: %v", err)
	}
	if analyst.Name == "" || analyst.ContactEmail != "analyst@example.com" {
		t.Fatalf("analyst profile not hydrated: %+v", analyst)
	}
}
