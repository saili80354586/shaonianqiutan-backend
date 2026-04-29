package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/controllers"
)

func TestLegacyOrderPaymentRouteIsNotRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	api := router.Group("/api")
	SetupOrderRoutes(api, controllers.NewOrderController(nil))

	req := httptest.NewRequest(http.MethodPost, "/api/orders/1/payment", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("POST /api/orders/1/payment status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
