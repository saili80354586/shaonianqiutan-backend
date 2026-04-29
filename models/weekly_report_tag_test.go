package models

import (
	"reflect"
	"testing"
)

func TestWeeklyReportRequestTagsUseBindingValidation(t *testing.T) {
	tests := []struct {
		model       any
		field       string
		bindingRule string
	}{
		{WeeklyReportSubmit{}, "TechnicalContent", "max=500"},
		{WeeklyReportSubmit{}, "TacticalContent", "max=500"},
		{WeeklyReportSubmit{}, "PhysicalCondition", "max=300"},
		{WeeklyReportSubmit{}, "ImprovementsDetail", "max=300"},
		{WeeklyReportSubmit{}, "Weaknesses", "max=300"},
		{WeeklyReportSubmit{}, "Injuries", "max=200"},
		{WeeklyReportSubmit{}, "DietCondition", "max=200"},
		{WeeklyReportSubmit{}, "MessageToCoach", "max=300"},
		{WeeklyReportReview{}, "KnowledgeFeedback", "max=300"},
		{WeeklyReportReview{}, "NextWeekFocus", "max=300"},
	}

	for _, tt := range tests {
		field, ok := reflect.TypeOf(tt.model).FieldByName(tt.field)
		if !ok {
			t.Fatalf("field %s not found", tt.field)
		}

		if got := field.Tag.Get("binding"); got != tt.bindingRule {
			t.Fatalf("%s binding tag = %q, want %q", tt.field, got, tt.bindingRule)
		}
	}
}
