package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	buildpkg "github.com/poe1-trainer/internal/build"
	"github.com/poe1-trainer/internal/guide"
	"github.com/poe1-trainer/internal/integration/ggg"
	"github.com/poe1-trainer/internal/recommendation"
	"github.com/poe1-trainer/internal/rule"
	runpkg "github.com/poe1-trainer/internal/run"
)

// Handlers holds all HTTP handler dependencies.
type Handlers struct {
	builds     *buildpkg.Repository
	guides     *guide.Repository
	runs       *runpkg.Service
	runRepo    *runpkg.Repository
	engine     *recommendation.Engine
	ruleEngine *rule.Engine
	// gggProvider is always non-nil: either *ggg.Client or ggg.NoopProvider.
	gggProvider ggg.CharacterProvider
	// gggClient is non-nil only when GGG OAuth is configured.
	// Provides AuthorizeURL and HandleCallback beyond the CharacterProvider interface.
	gggClient *ggg.Client
	// oauthStates stores in-flight OAuth state nonces for CSRF protection.
	// map[state string]expiresAt time.Time
	oauthStates sync.Map
	// watcherStatus returns the current logtail status string.
	// Nil when LOG_PATH is not configured.
	watcherStatus func() string

	// logLines broadcasts raw Client.txt lines to logtail SSE subscribers.
	logLines lineBroadcaster
	// runNotify broadcasts state-change signals to per-run SSE subscribers.
	runNotify signalBroadcaster
}

// SetWatcherStatusFunc registers a callback that returns the current logtail status.
// Called from main after the watcher is created.
func (h *Handlers) SetWatcherStatusFunc(fn func() string) {
	h.watcherStatus = fn
}

// EmitLogLine broadcasts a raw log line to all logtail SSE subscribers.
func (h *Handlers) EmitLogLine(line string) { h.logLines.emit(line) }

// NotifyRunUpdate signals all run-state SSE subscribers for the given run.
func (h *Handlers) NotifyRunUpdate(runID int64) { h.runNotify.emit(runID) }

// NewHandlers creates a new Handlers instance.
func NewHandlers(
	builds *buildpkg.Repository,
	guides *guide.Repository,
	runs *runpkg.Service,
	runRepo *runpkg.Repository,
	engine *recommendation.Engine,
	ruleEngine *rule.Engine,
	gggProvider ggg.CharacterProvider,
	gggClient *ggg.Client,
) *Handlers {
	return &Handlers{
		builds:      builds,
		guides:      guides,
		runs:        runs,
		runRepo:     runRepo,
		engine:      engine,
		ruleEngine:  ruleEngine,
		gggProvider: gggProvider,
		gggClient:   gggClient,
	}
}

// ─── Integration status ───────────────────────────────────────────────────

// GetIntegrationStatus handles GET /integration/status
func (h *Handlers) GetIntegrationStatus(w http.ResponseWriter, r *http.Request) {
	status := "disabled"
	if h.watcherStatus != nil {
		status = h.watcherStatus()
	}
	writeJSON(w, http.StatusOK, IntegrationStatusResponse{LogWatcher: status})
}

// StreamLogTail handles GET /integration/logtail/stream
// Streams raw Client.txt lines as Server-Sent Events.
func (h *Handlers) StreamLogTail(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}
	sseHeaders(w)

	status := "disabled"
	if h.watcherStatus != nil {
		status = h.watcherStatus()
	}
	sseEvent(w, flusher, "status", fmt.Sprintf(`{"status":%q}`, status))

	ch := h.logLines.subscribe()
	defer h.logLines.unsubscribe(ch)

	for {
		select {
		case <-r.Context().Done():
			return
		case line, ok := <-ch:
			if !ok {
				return
			}
			b, _ := json.Marshal(line)
			sseEvent(w, flusher, "log_line", fmt.Sprintf(`{"line":%s}`, b))
		}
	}
}

// StreamRunState handles GET /runs/{id}/stream
// Streams run state as Server-Sent Events; pushes on every state change
// and sends a heartbeat comment every 30 s to keep the connection alive.
func (h *Handlers) StreamRunState(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}
	id, ok := intPathParam(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid run id")
		return
	}
	sseHeaders(w)

	// Send the current state immediately so the client doesn't wait.
	if err := h.writeRunStateSSE(r.Context(), w, flusher, id); err != nil {
		return
	}

	notify := h.runNotify.subscribe(int64(id))
	defer h.runNotify.unsubscribe(int64(id), notify)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-notify:
			if err := h.writeRunStateSSE(r.Context(), w, flusher, id); err != nil {
				return
			}
		case <-ticker.C:
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}

// writeRunStateSSE fetches the current run state and sends it as an SSE event.
func (h *Handlers) writeRunStateSSE(ctx context.Context, w http.ResponseWriter, flusher http.Flusher, id int) error {
	state, err := h.runs.GetCurrentState(ctx, id)
	if err != nil {
		return err
	}
	b, err := json.Marshal(state)
	if err != nil {
		return err
	}
	sseEvent(w, flusher, "state", string(b))
	return nil
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
	var req CreateRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.GuideID == 0 {
		writeError(w, http.StatusBadRequest, "guide_id is required")
		return
	}
	run, err := h.runs.CreateRun(r.Context(), req.GuideID, req.CharacterName, req.League, req.AutoStart)
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
	g, err := h.loadGuideForRun(r.Context(), &state.Run)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	recs := h.engine.Produce(g, state)
	writeJSON(w, http.StatusOK, recs)
}

// GetRunGuide handles GET /runs/{id}/guide
func (h *Handlers) GetRunGuide(w http.ResponseWriter, r *http.Request) {
	id, ok := intPathParam(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid run id")
		return
	}
	run, err := h.runRepo.GetRun(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	g, err := h.loadGuideForRun(r.Context(), run)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, g)
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

// GetRanking handles GET /guides/{slug}/ranking — returns top 20 entries from
// local_rankings with act_splits and PB flag.
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
	entries, err := h.runRepo.GetDetailedRanking(r.Context(), g.ID, 20)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

// GetRankingStats handles GET /guides/{slug}/ranking/stats.
func (h *Handlers) GetRankingStats(w http.ResponseWriter, r *http.Request) {
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
	stats, err := h.runRepo.GetRankingStats(r.Context(), g.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

// GetSplitDeltas handles GET /runs/{id}/split-deltas.
func (h *Handlers) GetSplitDeltas(w http.ResponseWriter, r *http.Request) {
	id, ok := intPathParam(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid run id")
		return
	}
	deltas, err := h.runs.GetRunDeltas(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, deltas)
}

// PauseRun handles POST /runs/{id}/pause.
func (h *Handlers) PauseRun(w http.ResponseWriter, r *http.Request) {
	id, ok := intPathParam(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid run id")
		return
	}
	if err := h.runs.PauseRun(r.Context(), id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ResumeRun handles POST /runs/{id}/resume.
func (h *Handlers) ResumeRun(w http.ResponseWriter, r *http.Request) {
	id, ok := intPathParam(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid run id")
		return
	}
	if err := h.runs.ResumeRun(r.Context(), id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
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

func (h *Handlers) loadGuideForRun(ctx context.Context, run *runpkg.RunSession) (*guide.Guide, error) {
	if run == nil {
		return nil, fmt.Errorf("run is required")
	}
	if run.GuideRevision > 0 {
		return h.guides.GetByIDRevision(ctx, run.GuideID, run.GuideRevision)
	}
	return h.guides.GetByID(ctx, run.GuideID)
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

// ─── Builds ────────────────────────────────────────────────────────────────

// ListBuilds handles GET /builds
func (h *Handlers) ListBuilds(w http.ResponseWriter, r *http.Request) {
	builds, err := h.builds.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, builds)
}

// GetBuild handles GET /builds/{id}
func (h *Handlers) GetBuild(w http.ResponseWriter, r *http.Request) {
	id, ok := intPathParam(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid build id")
		return
	}
	b, err := h.builds.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	versions, err := h.builds.ListVersions(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"build": b, "versions": versions})
}

// CreateBuild handles POST /builds
func (h *Handlers) CreateBuild(w http.ResponseWriter, r *http.Request) {
	var req CreateBuildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Slug == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "slug and name are required")
		return
	}
	b := &buildpkg.Build{
		Slug:        req.Slug,
		Name:        req.Name,
		Class:       req.Class,
		Description: req.Description,
	}
	if err := h.builds.Create(r.Context(), b); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, b)
}

// ListBuildVersions handles GET /builds/{id}/versions
func (h *Handlers) ListBuildVersions(w http.ResponseWriter, r *http.Request) {
	id, ok := intPathParam(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid build id")
		return
	}
	versions, err := h.builds.ListVersions(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, versions)
}

// CreateBuildVersion handles POST /builds/{id}/versions
func (h *Handlers) CreateBuildVersion(w http.ResponseWriter, r *http.Request) {
	id, ok := intPathParam(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid build id")
		return
	}
	var req CreateVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Version == "" {
		writeError(w, http.StatusBadRequest, "version is required")
		return
	}
	v := &buildpkg.Version{
		BuildID:   id,
		Version:   req.Version,
		PatchTag:  req.PatchTag,
		Notes:     req.Notes,
		IsCurrent: req.IsCurrent,
	}
	if err := h.builds.CreateVersion(r.Context(), v); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, v)
}

// ─── Guide import ──────────────────────────────────────────────────────────

// ImportGuide handles POST /guides/import
func (h *Handlers) ImportGuide(w http.ResponseWriter, r *http.Request) {
	var req ImportGuideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Slug == "" || req.Title == "" || req.Content == "" {
		writeError(w, http.StatusBadRequest, "slug, title, and content are required")
		return
	}
	version, err := guide.ResolveVersion(req.Version)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	g, err := guide.ParseMarkdown(req.Slug, req.Title, req.BuildName, version, req.Content)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	if err := h.guides.Save(r.Context(), g); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, g)
}

// ─── Run extended actions ──────────────────────────────────────────────────

// ListActiveRuns handles GET /runs/active — returns currently active runs (0 or 1).
func (h *Handlers) ListActiveRuns(w http.ResponseWriter, r *http.Request) {
	run, err := h.runRepo.GetActiveRun(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if run == nil {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	writeJSON(w, http.StatusOK, []any{run})
}

// GetRun handles GET /runs/{id}
func (h *Handlers) GetRun(w http.ResponseWriter, r *http.Request) {
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

// AbandonRun handles POST /runs/{id}/abandon
func (h *Handlers) AbandonRun(w http.ResponseWriter, r *http.Request) {
	id, ok := intPathParam(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid run id")
		return
	}
	if err := h.runs.AbandonRun(r.Context(), id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// SkipStep handles POST /runs/{id}/steps/{step_id}/skip
func (h *Handlers) SkipStep(w http.ResponseWriter, r *http.Request) {
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
	if err := h.runs.SkipStep(r.Context(), runID, stepID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UndoStep handles POST /runs/{id}/steps/{step_id}/undo
func (h *Handlers) UndoStep(w http.ResponseWriter, r *http.Request) {
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
	if err := h.runs.UndoStep(r.Context(), runID, stepID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ─── Characters & snapshots ────────────────────────────────────────────────

// GetCharacter handles GET /runs/{id}/character
func (h *Handlers) GetCharacter(w http.ResponseWriter, r *http.Request) {
	id, ok := intPathParam(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid run id")
		return
	}
	c, err := h.runs.GetCharacter(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, c)
}

// UpsertCharacter handles PUT /runs/{id}/character
func (h *Handlers) UpsertCharacter(w http.ResponseWriter, r *http.Request) {
	id, ok := intPathParam(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid run id")
		return
	}
	var req UpsertCharacterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.CharacterName == "" {
		writeError(w, http.StatusBadRequest, "character_name is required")
		return
	}
	c := &runpkg.Character{
		RunID:          id,
		CharacterName:  req.CharacterName,
		CharacterClass: req.CharacterClass,
		League:         req.League,
		LevelAtStart:   req.LevelAtStart,
		LevelCurrent:   req.LevelAtStart,
	}
	if err := h.runs.UpsertCharacter(r.Context(), c); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, c)
}

// ListSnapshots handles GET /runs/{id}/snapshots
func (h *Handlers) ListSnapshots(w http.ResponseWriter, r *http.Request) {
	id, ok := intPathParam(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid run id")
		return
	}
	snaps, err := h.runs.ListSnapshots(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, snaps)
}

// CreateSnapshot handles POST /runs/{id}/snapshots
func (h *Handlers) CreateSnapshot(w http.ResponseWriter, r *http.Request) {
	id, ok := intPathParam(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid run id")
		return
	}
	var req CreateSnapshotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Level <= 0 {
		writeError(w, http.StatusBadRequest, "level must be > 0")
		return
	}
	snap := &runpkg.CharacterSnapshot{
		RunID:        id,
		Level:        req.Level,
		LifeMax:      req.LifeMax,
		ManaMax:      req.ManaMax,
		ResFire:      req.ResFire,
		ResCold:      req.ResCold,
		ResLightning: req.ResLightning,
		ResChaos:     req.ResChaos,
	}
	if err := h.runs.CreateSnapshot(r.Context(), snap); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, snap)
}

// ─── Alerts ────────────────────────────────────────────────────────────────

// GetAlerts handles GET /runs/{id}/alerts
// Returns step-specific gem requirements, gear hints from DB, and campaign-phase
// rule alerts evaluated against the current act, character level, and step type.
func (h *Handlers) GetAlerts(w http.ResponseWriter, r *http.Request) {
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
	if state.CurrentStepID == 0 {
		writeJSON(w, http.StatusOK, AlertsResponse{Alerts: []Alert{}})
		return
	}

	g, err := h.loadGuideForRun(r.Context(), &state.Run)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Locate the current step.
	var currentStep *guide.Step
	for i := range g.Steps {
		if g.Steps[i].ID == state.CurrentStepID {
			currentStep = &g.Steps[i]
			break
		}
	}

	var alerts []Alert

	// 1. Step-specific gem requirement alerts ("do this right now").
	if currentStep != nil {
		for _, gem := range currentStep.GemRequirements {
			verb := "Kup gem"
			actionType := "vendor"
			if currentStep.StepType == guide.StepTypeQuestReward {
				verb = "Odbierz gem z nagrody questa"
				actionType = "quest_reward"
			}
			alerts = append(alerts, Alert{
				Kind:        "gem",
				Priority:    "high",
				Description: verb + ": " + gem.GemName,
				GemName:     gem.GemName,
				ActionType:  actionType,
				Notes:       gem.Note,
				Reason:      gem.Note,
				StepID:      currentStep.ID,
				Source:      "step",
			})
		}
	}

	// 2. Gear hints from gear_hint_rules table (step-specific + global).
	hints, err := h.guides.GetGearHints(r.Context(), g.ID, state.CurrentStepID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for _, hint := range hints {
		stepID := 0
		if hint.StepID != nil {
			stepID = *hint.StepID
		}
		alerts = append(alerts, Alert{
			Kind:        "gear",
			Priority:    string(hint.Priority),
			Slot:        hint.Slot,
			Description: hint.Description,
			Notes:       hint.Notes,
			StepID:      stepID,
			Source:      "step",
		})
	}

	// 3. Campaign-phase rule alerts from embedded rule engine.
	if h.ruleEngine != nil && currentStep != nil {
		// Resolve character level from the cached run_characters row.
		charLevel := 0
		if char, err := h.runs.GetCharacter(r.Context(), id); err == nil && char != nil {
			charLevel = char.LevelCurrent
		}
		ctx := rule.EvalContext{
			Act:      currentStep.Act,
			Level:    charLevel,
			StepType: string(currentStep.StepType),
		}
		for _, ra := range h.ruleEngine.Evaluate(g.Slug, ctx) {
			alerts = append(alerts, Alert{
				Kind:        ra.Kind.Category(),
				Priority:    ra.Priority,
				Slot:        ra.Slot,
				Description: ra.Description,
				GemName:     ra.GemName,
				ActionType:  string(ra.Kind),
				Reason:      ra.Reason,
				Source:      "rule",
			})
		}
	}

	if alerts == nil {
		alerts = []Alert{}
	}
	writeJSON(w, http.StatusOK, AlertsResponse{
		StepID: state.CurrentStepID,
		Alerts: alerts,
	})
}

// ─── Events ────────────────────────────────────────────────────────────────

// ListEvents handles GET /runs/{id}/events
func (h *Handlers) ListEvents(w http.ResponseWriter, r *http.Request) {
	id, ok := intPathParam(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid run id")
		return
	}
	events, err := h.runs.ListEvents(r.Context(), id, 100)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, events)
}

// RecordEvent handles POST /runs/{id}/events
func (h *Handlers) RecordEvent(w http.ResponseWriter, r *http.Request) {
	id, ok := intPathParam(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid run id")
		return
	}
	var req RecordEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.EventType == "" {
		writeError(w, http.StatusBadRequest, "event_type is required")
		return
	}
	if req.Payload == nil {
		req.Payload = map[string]string{}
	}
	if err := h.runs.HandleAreaEvent(r.Context(), id, runpkg.AreaEvent{AreaName: req.Payload["area"]}); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ─── Splits ────────────────────────────────────────────────────────────────

// ListSplits handles GET /runs/{id}/splits
func (h *Handlers) ListSplits(w http.ResponseWriter, r *http.Request) {
	id, ok := intPathParam(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid run id")
		return
	}
	splits, err := h.runs.ListSplits(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, splits)
}

// RecordSplit handles POST /runs/{id}/steps/{step_id}/split
func (h *Handlers) RecordSplit(w http.ResponseWriter, r *http.Request) {
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
	var req RecordSplitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.SplitMs <= 0 {
		writeError(w, http.StatusBadRequest, "split_ms must be > 0")
		return
	}
	if err := h.runs.RecordSplit(r.Context(), runID, stepID, req.SplitMs); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ─── Manual checks ─────────────────────────────────────────────────────────

// ListPendingChecks handles GET /runs/{id}/checks
func (h *Handlers) ListPendingChecks(w http.ResponseWriter, r *http.Request) {
	id, ok := intPathParam(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid run id")
		return
	}
	checks, err := h.runs.ListPendingChecks(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, checks)
}

// AnswerCheck handles POST /runs/{id}/checks/{check_id}/answer
func (h *Handlers) AnswerCheck(w http.ResponseWriter, r *http.Request) {
	_, ok := intPathParam(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid run id")
		return
	}
	checkID, ok := intPathParam(r, "check_id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid check_id")
		return
	}
	var req AnswerCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	mc, err := h.runs.AnswerCheck(r.Context(), checkID, req.ResponseValue)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, mc)
}

// ─── GGG Integration ───────────────────────────────────────────────────────

// GGGStatus handles GET /ggg/status
// Zwraca czy integracja jest skonfigurowana i czy jest ważny token.
func (h *Handlers) GGGStatus(w http.ResponseWriter, r *http.Request) {
	resp := GGGStatusResponse{
		Configured: h.gggClient != nil,
		Available:  h.gggProvider.IsAvailable(),
	}
	writeJSON(w, http.StatusOK, resp)
}

// GGGAuth handles GET /ggg/auth
// Przekierowuje przeglądarkę do strony autoryzacji GGG OAuth.
// Wymaga skonfigurowanego GGG_CLIENT_ID i GGG_CLIENT_SECRET.
func (h *Handlers) GGGAuth(w http.ResponseWriter, r *http.Request) {
	if h.gggClient == nil {
		writeError(w, http.StatusServiceUnavailable, "GGG OAuth nie skonfigurowany — ustaw GGG_CLIENT_ID i GGG_CLIENT_SECRET")
		return
	}
	state, err := generateOAuthState()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "błąd generowania state")
		return
	}
	h.oauthStates.Store(state, time.Now().Add(10*time.Minute))
	http.Redirect(w, r, h.gggClient.AuthorizeURL(state), http.StatusFound)
}

// GGGCallback handles GET /ggg/callback
// Odbiera kod autoryzacji od GGG, weryfikuje state (CSRF), wymienia kod na token.
func (h *Handlers) GGGCallback(w http.ResponseWriter, r *http.Request) {
	if h.gggClient == nil {
		writeError(w, http.StatusServiceUnavailable, "GGG OAuth nie skonfigurowany")
		return
	}
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	if state == "" || code == "" {
		writeError(w, http.StatusBadRequest, "wymagane parametry: state, code")
		return
	}

	// Weryfikacja CSRF state.
	val, ok := h.oauthStates.LoadAndDelete(state)
	if !ok {
		writeError(w, http.StatusBadRequest, "nieprawidłowy lub wygasły state")
		return
	}
	if expiresAt, ok := val.(time.Time); !ok || time.Now().After(expiresAt) {
		writeError(w, http.StatusBadRequest, "wygasły state — powtórz autoryzację")
		return
	}

	tok, err := h.gggClient.HandleCallback(r.Context(), code)
	if err != nil {
		writeError(w, http.StatusBadGateway, "nie udało się wymienić kodu: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"username": tok.Username,
		"scope":    tok.Scope,
	})
}

// GGGSyncSnapshot handles POST /runs/{id}/snapshots/ggg
// Pobiera aktualny stan postaci z GGG API i zapisuje jako snapshot.
func (h *Handlers) GGGSyncSnapshot(w http.ResponseWriter, r *http.Request) {
	id, ok := intPathParam(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid run id")
		return
	}
	if !h.gggProvider.IsAvailable() {
		writeError(w, http.StatusServiceUnavailable, "GGG API niedostępne — skonfiguruj OAuth lub użyj GET /ggg/auth")
		return
	}
	var req GGGSyncSnapshotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.CharacterName == "" {
		writeError(w, http.StatusBadRequest, "character_name is required")
		return
	}
	realm := req.Realm
	if realm == "" {
		realm = "pc"
	}

	data, err := h.gggProvider.FetchSnapshot(r.Context(), req.CharacterName, realm)
	if err != nil {
		writeError(w, http.StatusBadGateway, "GGG API error: "+err.Error())
		return
	}

	snap := &runpkg.CharacterSnapshot{
		RunID:         id,
		Level:         data.Level,
		EquippedItems: data.EquippedItems,
		Skills:        data.Skills,
		RawResponse:   data.RawResponse,
	}
	if err := h.runs.CreateGGGSnapshot(r.Context(), snap); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, snap)
}

// generateOAuthState tworzy losową wartość state dla CSRF protection.
func generateOAuthState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
