package config

import (
	"reflect"
	"testing"
)

func TestGetCORSOriginsProductionSupportsMultipleOrigins(t *testing.T) {
	t.Setenv("NODE_ENV", "production")
	t.Setenv("FRONTEND_URL", "https://app.example.com/")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://admin.example.com, https://app.example.com/")

	got := GetCORSOrigins()
	want := []string{"https://admin.example.com", "https://app.example.com"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetCORSOrigins() = %#v, want %#v", got, want)
	}
}

func TestGetCORSOriginsDevelopmentIncludesLocalDefaults(t *testing.T) {
	t.Setenv("NODE_ENV", "development")
	t.Setenv("FRONTEND_URL", "http://127.0.0.1:5173/")
	t.Setenv("CORS_ALLOWED_ORIGINS", "")

	got := GetCORSOrigins()
	want := []string{"http://127.0.0.1:5173", "http://localhost:5173", "http://localhost:3000"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetCORSOrigins() = %#v, want %#v", got, want)
	}
}
