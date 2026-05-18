package controllers

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/shaonianqiutan/backend/models"
)

func TestTrainingPlanIncludesPlayer(t *testing.T) {
	playerIDsJSON, _ := json.Marshal([]uint{101, 102})
	plan := models.TrainingPlan{PlayerIDs: string(playerIDsJSON)}

	if !trainingPlanIncludesPlayer(plan, 101) {
		t.Fatalf("expected player 101 to be included")
	}
	if trainingPlanIncludesPlayer(plan, 103) {
		t.Fatalf("expected player 103 not to be included")
	}

	teamWidePlan := models.TrainingPlan{StartTime: time.Now()}
	if !trainingPlanIncludesPlayer(teamWidePlan, 103) {
		t.Fatalf("empty player_ids should include active team players")
	}
}
