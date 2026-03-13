package guide_test

import (
	"testing"

	"github.com/poe1-trainer/internal/guide"
)

const sampleMarkdown = `
# Test Guide

## Zasady ogólne

1. This should be skipped.

## Akt 1

Kolejność lokacji:

<span style="color:white;">Twilight Strand</span>

1. Idź prosto do <span style="color:red;">Hillocka</span> bez farmienia plaży.
2. Zabij <span style="color:red;">Hillocka</span>.
3. Weź <span style="color:#7f7fff;">Rolling Magmę</span> od Tarkleigha.

## Akt 2

1. Wyjdź do The Old Fields.
2. Zabij <span style="color:red;">Fidelitasa</span>.
`

func TestParseMarkdown_BasicSteps(t *testing.T) {
	g, err := guide.ParseMarkdown("test_guide", "Test Guide", "test_build", "1", sampleMarkdown)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(g.Steps) != 5 {
		t.Errorf("expected 5 steps, got %d", len(g.Steps))
	}
}

func TestParseMarkdown_ActAssignment(t *testing.T) {
	g, err := guide.ParseMarkdown("test_guide", "Test Guide", "test_build", "1", sampleMarkdown)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, s := range g.Steps[:3] {
		if s.Act != 1 {
			t.Errorf("expected act 1 for step %d, got %d", s.StepNumber, s.Act)
		}
	}
	for _, s := range g.Steps[3:] {
		if s.Act != 2 {
			t.Errorf("expected act 2 for step %d, got %d", s.StepNumber, s.Act)
		}
	}
}

func TestParseMarkdown_GemExtraction(t *testing.T) {
	g, err := guide.ParseMarkdown("test_guide", "Test Guide", "test_build", "1", sampleMarkdown)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Step 3 (index 2) has Rolling Magma as a blue gem.
	step := g.Steps[2]
	if len(step.GemRequirements) != 1 {
		t.Errorf("expected 1 gem requirement in step 3, got %d", len(step.GemRequirements))
	} else if step.GemRequirements[0].Color != "blue" {
		t.Errorf("expected blue gem, got %s", step.GemRequirements[0].Color)
	}
}

func TestParseMarkdown_CheckpointDetection(t *testing.T) {
	g, err := guide.ParseMarkdown("test_guide", "Test Guide", "test_build", "1", sampleMarkdown)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Step 2: "Zabij Hillocka" should be a checkpoint.
	if !g.Steps[1].IsCheckpoint {
		t.Error("expected step 2 ('Zabij Hillocka') to be a checkpoint")
	}
	// Step 1: "Idź prosto..." is not a checkpoint.
	if g.Steps[0].IsCheckpoint {
		t.Error("expected step 1 to NOT be a checkpoint")
	}
}

func TestParseMarkdown_NoPreambleSteps(t *testing.T) {
	g, err := guide.ParseMarkdown("test_guide", "Test Guide", "test_build", "1", sampleMarkdown)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Steps from "## Zasady ogólne" must not be included.
	for _, s := range g.Steps {
		if s.Title == "This should be skipped." {
			t.Error("preamble step should not be included")
		}
	}
}

func TestParseMarkdown_EmptyGuide(t *testing.T) {
	_, err := guide.ParseMarkdown("empty", "Empty", "build", "1", "## Akt 1\n\nno steps here\n")
	if err == nil {
		t.Error("expected error for guide with no steps")
	}
}
