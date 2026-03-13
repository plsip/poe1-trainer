package recommendation_test

import (
	"testing"

	"github.com/poe1-trainer/internal/guide"
	"github.com/poe1-trainer/internal/recommendation"
	runpkg "github.com/poe1-trainer/internal/run"
)

func makeGuide() *guide.Guide {
	return &guide.Guide{
		ID:    1,
		Slug:  "test",
		Steps: []guide.Step{
			{ID: 1, GuideID: 1, StepNumber: 1, Act: 1, Title: "Idź do Hillocka", IsCheckpoint: false},
			{ID: 2, GuideID: 1, StepNumber: 2, Act: 1, Title: "Zabij Hillocka", IsCheckpoint: true},
			{
				ID: 3, GuideID: 1, StepNumber: 3, Act: 1, Title: "Weź Rolling Magmę",
				IsCheckpoint: false,
				GemRequirements: []guide.GemRequirement{
					{ID: 1, StepID: 3, GemName: "Rolling Magma", Color: "blue"},
				},
			},
		},
	}
}

func makeState(currentStepID int, confirmedIDs []int, elapsedMs int64) *runpkg.CurrentState {
	return &runpkg.CurrentState{
		Run:              runpkg.RunSession{ID: 1, GuideID: 1},
		CurrentStepID:    currentStepID,
		ConfirmedStepIDs: confirmedIDs,
		ElapsedMs:        elapsedMs,
	}
}

func TestEngine_ProducesCurrentStepRec(t *testing.T) {
	engine := recommendation.NewEngine()
	g := makeGuide()
	state := makeState(1, nil, 0)

	recs := engine.Produce(g, state)
	if len(recs) == 0 {
		t.Fatal("expected at least one recommendation")
	}
	if recs[0].Priority != recommendation.PriorityHigh {
		t.Errorf("expected high priority for current step rec, got %s", recs[0].Priority)
	}
	if recs[0].StepID != 1 {
		t.Errorf("expected step_id 1, got %d", recs[0].StepID)
	}
}

func TestEngine_CheckpointRecAdded(t *testing.T) {
	engine := recommendation.NewEngine()
	g := makeGuide()
	// Current step is step 2 (checkpoint).
	state := makeState(2, []int{1}, 0)

	recs := engine.Produce(g, state)
	hasConfirmRec := false
	for _, r := range recs {
		if r.ID == "confirm_2" {
			hasConfirmRec = true
		}
	}
	if !hasConfirmRec {
		t.Error("expected confirm recommendation for checkpoint step")
	}
}

func TestEngine_GemRecAdded(t *testing.T) {
	engine := recommendation.NewEngine()
	g := makeGuide()
	state := makeState(3, []int{1, 2}, 0)

	recs := engine.Produce(g, state)
	hasGemRec := false
	for _, r := range recs {
		if r.ID == "gem_3_rolling_magma" {
			hasGemRec = true
		}
	}
	if !hasGemRec {
		t.Error("expected gem recommendation for step 3")
	}
}

func TestEngine_PaceRecAfter30Min(t *testing.T) {
	engine := recommendation.NewEngine()
	g := makeGuide()
	// 35 minutes elapsed, still in Act 1.
	state := makeState(1, nil, 35*60*1000)

	recs := engine.Produce(g, state)
	hasPaceRec := false
	for _, r := range recs {
		if r.ID == "pace_act1" {
			hasPaceRec = true
		}
	}
	if !hasPaceRec {
		t.Error("expected pace recommendation after 30 minutes in Act 1")
	}
}

func TestEngine_AllDoneWhenNoCurrentStep(t *testing.T) {
	engine := recommendation.NewEngine()
	g := makeGuide()
	state := makeState(0, []int{1, 2, 3}, 0)

	recs := engine.Produce(g, state)
	if len(recs) == 0 || recs[0].ID != "all_done" {
		t.Error("expected 'all_done' recommendation when no current step")
	}
}

func TestEngine_NilGuide(t *testing.T) {
	engine := recommendation.NewEngine()
	recs := engine.Produce(nil, makeState(1, nil, 0))
	if len(recs) != 0 {
		t.Error("expected no recommendations for nil guide")
	}
}
