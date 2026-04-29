package utils

import (
	"testing"
	"time"

	"github.com/shaonianqiutan/backend/models"
)

func TestCalculatePlayerMapScoreUsesPhysicalAndScoutEvidence(t *testing.T) {
	sprint := 4.5
	jump := 210.0
	push := 30
	reportAverage := 86.0

	score := CalculatePlayerMapScore(PlayerScoreInput{
		User: &models.User{Age: 12, Position: "边锋"},
		PhysicalRecord: &models.PhysicalTestRecord{
			TestDate:         time.Now(),
			Sprint30m:        &sprint,
			StandingLongJump: &jump,
			PushUp:           &push,
		},
		ScoutReportAverage: &reportAverage,
		ScoutReportCount:   2,
	})

	if !score.HasScore {
		t.Fatal("expected score to be present")
	}
	if score.Score <= 0 {
		t.Fatalf("expected positive score, got %.1f", score.Score)
	}
	if score.FormulaVersion != PlayerScoreFormulaVersion {
		t.Fatalf("unexpected formula version: %s", score.FormulaVersion)
	}
	if score.BenchmarkGroup != "U11-U12/FWD" {
		t.Fatalf("unexpected benchmark group: %s", score.BenchmarkGroup)
	}
	if score.Confidence <= 0 || score.Confidence > 1 {
		t.Fatalf("unexpected confidence: %.2f", score.Confidence)
	}
	if len(score.Components) != 3 {
		t.Fatalf("expected physical, scout, and completeness components, got %+v", score.Components)
	}
	if len(score.Metrics) != 3 {
		t.Fatalf("expected three physical metrics, got %+v", score.Metrics)
	}
	if score.Metrics[0].Benchmark == "" {
		t.Fatalf("expected benchmark explanation on metrics, got %+v", score.Metrics[0])
	}
}

func TestCalculatePlayerMapScoreDoesNotInventScoreWithoutSources(t *testing.T) {
	score := CalculatePlayerMapScore(PlayerScoreInput{User: &models.User{Name: "未评估球员"}})

	if score.HasScore {
		t.Fatalf("expected no score, got %+v", score)
	}
	if score.Score != 0 {
		t.Fatalf("expected zero score, got %.1f", score.Score)
	}
	if score.Potential != "待评估" {
		t.Fatalf("expected pending potential, got %s", score.Potential)
	}
}

func TestCalculatePlayerMapScoreV2UsesAgeBenchmarks(t *testing.T) {
	sprint := 4.7

	u12 := CalculatePlayerMapScore(PlayerScoreInput{
		User: &models.User{Age: 12, Position: "边锋"},
		PhysicalRecord: &models.PhysicalTestRecord{
			TestDate:  time.Now(),
			Sprint30m: &sprint,
		},
	})
	u16 := CalculatePlayerMapScore(PlayerScoreInput{
		User: &models.User{Age: 16, Position: "边锋"},
		PhysicalRecord: &models.PhysicalTestRecord{
			TestDate:  time.Now(),
			Sprint30m: &sprint,
		},
	})

	if u12.BenchmarkGroup != "U11-U12/FWD" {
		t.Fatalf("unexpected U12 benchmark group: %s", u12.BenchmarkGroup)
	}
	if u16.BenchmarkGroup != "U15-U16/FWD" {
		t.Fatalf("unexpected U16 benchmark group: %s", u16.BenchmarkGroup)
	}
	if u12.Metrics[0].Score <= u16.Metrics[0].Score {
		t.Fatalf("expected same sprint to score higher in U12 than U16, got U12=%.1f U16=%.1f", u12.Metrics[0].Score, u16.Metrics[0].Score)
	}
}

func TestCalculatePlayerMapScoreIsDeterministic(t *testing.T) {
	sprint := 4.45
	jump := 218.0
	push := 27

	input := PlayerScoreInput{
		User: &models.User{Age: 12, Position: "门将"},
		PhysicalRecord: &models.PhysicalTestRecord{
			TestDate:         time.Now(),
			Sprint30m:        &sprint,
			StandingLongJump: &jump,
			PushUp:           &push,
		},
	}

	first := CalculatePlayerMapScore(input)
	second := CalculatePlayerMapScore(input)
	if first.Score != second.Score || first.BenchmarkGroup != second.BenchmarkGroup || first.Confidence != second.Confidence {
		t.Fatalf("expected deterministic score, got first=%+v second=%+v", first, second)
	}
}
