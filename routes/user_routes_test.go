package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/controllers"
)

func TestLegacyVideoAnalyzeRouteIsNotRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	api := router.Group("/api")
	SetupUserRoutes(api, &controllers.UserController{})

	req := httptest.NewRequest(http.MethodPost, "/api/video/analyze", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("POST /api/video/analyze status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
