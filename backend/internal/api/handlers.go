package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/poe1-trainer/internal/guide"
	"github.com/poe1-trainer/internal/recommendation"
	runpkg "github.com/poe1-trainer/internal/run"
)

// Handlers holds all HTTP handler dependencies.
type Handlers struct {
	guides  *guide.Repository
	runs    *runpkg.Service
	runRepo *runpkg.Repository
	engine  *recommendation.Engine
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(
	guides *guide.Repository,
	runs *runpkg.Service,
	runRepo *runpkg.Repository,
	engine *recommendation.Engine,
) *Handlers {
	return &Handlers{
		guides:  guides,
		runs:    runs,
		runRepo: runRepo,
		engine:  engine,
	}
}

// ─── Guides ────────────────────────────────────────────────────────────────

// ListGuides handles GET /guides
func (h *Handlers) ListGuides(w http.ResponseWriter, r *http.Request) {
	guides, err := h.guides.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, guides)
}

// GetGuide handles GET /guides/{slug}
func (h *Handlers) GetGuide(w http.ResponseWriter, r *http.Request) {
	slug := pathParam(r, "slug")
	if slug == "" {
		writeError(w, http.StatusBadRequest, "missing guide slug")
		return
	}
	g, err := h.guides.GetBySlug(r.Context(), slug)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, g)
}

// ─── Runs ──────────────────────────────────────────────────────────────────

// CreateRun handles POST /runs
func (h *Handlers) CreateRun(w http.ResponseWriter, r *http.Request) {
	var req struct {
		GuideID       int    `json:"guide_id"`
		CharacterName string `json:"character_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.GuideID == 0 {
		writeError(w, http.StatusBadRequest, "guide_id is required")
		return
	}
	run, err := h.runs.CreateRun(r.Context(), req.GuideID, req.CharacterName)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, run)
}

// GetRunState handles GET /runs/{id}/state
func (h *Handlers) GetRunState(w http.ResponseWriter, r *http.Request) {
	id, ok := intPathParam(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid run id")
		return
	}
	state, err := h.runs.GetCurrentState(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, state)
}

// GetRunRecommendations handles GET /runs/{id}/recommendations
func (h *Handlers) GetRunRecommendations(w http.ResponseWriter, r *http.Request) {
	id, ok := intPathParam(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid run id")
		return
	}
	state, err := h.runs.GetCurrentState(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	g, err := h.guides.GetByID(r.Context(), state.Run.GuideID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	recs := h.engine.Produce(g, state)
	writeJSON(w, http.StatusOK, recs)
}

// ConfirmStep handles POST /runs/{id}/steps/{step_id}/confirm
func (h *Handlers) ConfirmStep(w http.ResponseWriter, r *http.Request) {
	runID, ok := intPathParam(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid run id")
		return
	}
	stepID, ok := intPathParam(r, "step_id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid step_id")
		return
	}
	cp, err := h.runs.ConfirmStep(r.Context(), runID, stepID, runpkg.ConfirmedByManual)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cp)
}

// FinishRun handles POST /runs/{id}/finish
func (h *Handlers) FinishRun(w http.ResponseWriter, r *http.Request) {
	id, ok := intPathParam(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid run id")
		return
	}
	if err := h.runs.FinishRun(r.Context(), id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListRuns handles GET /guides/{slug}/runs
func (h *Handlers) ListRuns(w http.ResponseWriter, r *http.Request) {
	slug := pathParam(r, "slug")
	if slug == "" {
		writeError(w, http.StatusBadRequest, "missing guide slug")
		return
	}
	g, err := h.guides.GetBySlug(r.Context(), slug)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	runs, err := h.runs.ListRuns(r.Context(), g.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, runs)
}

// GetRanking handles GET /guides/{slug}/ranking
func (h *Handlers) GetRanking(w http.ResponseWriter, r *http.Request) {
	slug := pathParam(r, "slug")
	if slug == "" {
		writeError(w, http.StatusBadRequest, "missing guide slug")
		return
	}
	g, err := h.guides.GetBySlug(r.Context(), slug)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	entries, err := h.runRepo.GetRanking(r.Context(), g.ID, 20)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

// ─── helpers ───────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func pathParam(r *http.Request, key string) string {
	return r.PathValue(key)
}

func intPathParam(r *http.Request, key string) (int, bool) {
	s := pathParam(r, key)
	if s == "" {
		return 0, false
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}
