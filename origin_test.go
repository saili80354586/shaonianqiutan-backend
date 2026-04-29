package main

import "testing"

func TestIsAllowedOriginRequiresExactMatch(t *testing.T) {
	allowedOrigins := []string{
		"http://localhost:5173",
		"http://admin.localhost:5173",
	}

	if !isAllowedOrigin("http://localhost:5173", allowedOrigins) {
		t.Fatal("expected exact localhost origin to be allowed")
	}

	if !isAllowedOrigin("http://admin.localhost:5173", allowedOrigins) {
		t.Fatal("expected configured admin origin to be allowed")
	}

	if isAllowedOrigin("http://evil.localhost:5173", allowedOrigins) {
		t.Fatal("expected unconfigured localhost subdomain to be rejected")
	}

	if isAllowedOrigin("", allowedOrigins) {
		t.Fatal("expected empty origin to be rejected by CORS helper")
	}
}
