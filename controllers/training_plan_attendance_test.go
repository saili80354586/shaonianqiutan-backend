package controllers

import "testing"

func TestIsValidTrainingAttendanceStatus(t *testing.T) {
	validStatuses := []string{"present", "leave", "absent", "late"}
	for _, status := range validStatuses {
		if !isValidTrainingAttendanceStatus(status) {
			t.Fatalf("expected %s to be valid", status)
		}
	}

	if isValidTrainingAttendanceStatus("unknown") {
		t.Fatalf("expected unknown status to be invalid")
	}
}

func TestEmptyAttendanceSummary(t *testing.T) {
	summary := emptyAttendanceSummary()
	for _, key := range []string{"total", "present", "leave", "absent", "late", "unmarked"} {
		if _, ok := summary[key]; !ok {
			t.Fatalf("expected key %s in summary", key)
		}
	}
}
