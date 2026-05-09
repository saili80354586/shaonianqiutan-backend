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

func TestGetPaymentModeDefaultsToMock(t *testing.T) {
	t.Setenv("PAYMENT_MODE", "")

	if got := GetPaymentMode(); got != PaymentModeMock {
		t.Fatalf("GetPaymentMode() = %q, want %q", got, PaymentModeMock)
	}
}

func TestGetPaymentModeHonorsReal(t *testing.T) {
	t.Setenv("PAYMENT_MODE", "real")

	if got := GetPaymentMode(); got != PaymentModeReal {
		t.Fatalf("GetPaymentMode() = %q, want %q", got, PaymentModeReal)
	}
}

func TestGetSmsModeDefaultsToMockInDevelopment(t *testing.T) {
	t.Setenv("NODE_ENV", "development")
	t.Setenv("SMS_MODE", "")

	if got := GetSmsMode(); got != SmsModeMock {
		t.Fatalf("GetSmsMode() = %q, want %q", got, SmsModeMock)
	}
}

func TestGetSmsModeDefaultsToRealOutsideDevelopment(t *testing.T) {
	t.Setenv("NODE_ENV", "production")
	t.Setenv("SMS_MODE", "")

	if got := GetSmsMode(); got != SmsModeReal {
		t.Fatalf("GetSmsMode() = %q, want %q", got, SmsModeReal)
	}
}

func TestGetSmsModeHonorsExplicitMockOutsideDevelopment(t *testing.T) {
	t.Setenv("NODE_ENV", "production")
	t.Setenv("SMS_MODE", "mock")

	if got := GetSmsMode(); got != SmsModeMock {
		t.Fatalf("GetSmsMode() = %q, want %q", got, SmsModeMock)
	}
}

func TestIsAnalystRegistrationAutoApprovedDefaultsToTrueInDevelopment(t *testing.T) {
	t.Setenv("NODE_ENV", "development")
	t.Setenv("ANALYST_REGISTRATION_AUTO_APPROVE", "")

	if !IsAnalystRegistrationAutoApproved() {
		t.Fatal("IsAnalystRegistrationAutoApproved() = false, want true")
	}
}

func TestIsAnalystRegistrationAutoApprovedDefaultsToFalseOutsideDevelopment(t *testing.T) {
	t.Setenv("NODE_ENV", "production")
	t.Setenv("ANALYST_REGISTRATION_AUTO_APPROVE", "")

	if IsAnalystRegistrationAutoApproved() {
		t.Fatal("IsAnalystRegistrationAutoApproved() = true, want false")
	}
}

func TestIsAnalystRegistrationAutoApprovedHonorsExplicitTrueOutsideDevelopment(t *testing.T) {
	t.Setenv("NODE_ENV", "production")
	t.Setenv("ANALYST_REGISTRATION_AUTO_APPROVE", "true")

	if !IsAnalystRegistrationAutoApproved() {
		t.Fatal("IsAnalystRegistrationAutoApproved() = false, want true")
	}
}
