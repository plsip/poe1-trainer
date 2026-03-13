package api

import (
	"net/http"
)

// NewRouter builds and returns the HTTP router for all API endpoints.
// It uses the stdlib ServeMux with Go 1.22+ path parameter syntax.
func NewRouter(h *Handlers) http.Handler {
	mux := http.NewServeMux()

	// --- builds ---
	mux.HandleFunc("GET /builds", h.ListBuilds)
	mux.HandleFunc("POST /builds", h.CreateBuild)
	mux.HandleFunc("GET /builds/{id}", h.GetBuild)
	mux.HandleFunc("GET /builds/{id}/versions", h.ListBuildVersions)
	mux.HandleFunc("POST /builds/{id}/versions", h.CreateBuildVersion)

	// --- guides ---
	mux.HandleFunc("GET /guides", h.ListGuides)
	mux.HandleFunc("POST /guides/import", h.ImportGuide)
	mux.HandleFunc("GET /guides/{slug}", h.GetGuide)
	mux.HandleFunc("GET /guides/{slug}/runs", h.ListRuns)
	mux.HandleFunc("GET /guides/{slug}/ranking", h.GetRanking)
	mux.HandleFunc("GET /guides/{slug}/ranking/stats", h.GetRankingStats)

	// --- runs ---
	mux.HandleFunc("POST /runs", h.CreateRun)
	mux.HandleFunc("GET /runs/{id}", h.GetRun)
	mux.HandleFunc("GET /runs/{id}/state", h.GetRunState)
	mux.HandleFunc("POST /runs/{id}/finish", h.FinishRun)
	mux.HandleFunc("POST /runs/{id}/abandon", h.AbandonRun)
	mux.HandleFunc("POST /runs/{id}/pause", h.PauseRun)
	mux.HandleFunc("POST /runs/{id}/resume", h.ResumeRun)

	// --- step actions ---
	mux.HandleFunc("GET /runs/{id}/recommendations", h.GetRunRecommendations)
	mux.HandleFunc("POST /runs/{id}/steps/{step_id}/confirm", h.ConfirmStep)
	mux.HandleFunc("POST /runs/{id}/steps/{step_id}/skip", h.SkipStep)
	mux.HandleFunc("POST /runs/{id}/steps/{step_id}/undo", h.UndoStep)
	mux.HandleFunc("POST /runs/{id}/steps/{step_id}/split", h.RecordSplit)

	// --- character & snapshots ---
	mux.HandleFunc("GET /runs/{id}/character", h.GetCharacter)
	mux.HandleFunc("PUT /runs/{id}/character", h.UpsertCharacter)
	mux.HandleFunc("GET /runs/{id}/snapshots", h.ListSnapshots)
	mux.HandleFunc("POST /runs/{id}/snapshots", h.CreateSnapshot)

	// --- alerts ---
	mux.HandleFunc("GET /runs/{id}/alerts", h.GetAlerts)

	// --- events (logtail-driven) ---
	mux.HandleFunc("GET /runs/{id}/events", h.ListEvents)
	mux.HandleFunc("POST /runs/{id}/events", h.RecordEvent)

	// --- splits, deltas & ranking ---
	mux.HandleFunc("GET /runs/{id}/splits", h.ListSplits)
	mux.HandleFunc("GET /runs/{id}/split-deltas", h.GetSplitDeltas)

	// --- manual checks ---
	mux.HandleFunc("GET /runs/{id}/checks", h.ListPendingChecks)
	mux.HandleFunc("POST /runs/{id}/checks/{check_id}/answer", h.AnswerCheck)

	// --- integration status ---
	mux.HandleFunc("GET /integration/status", h.GetIntegrationStatus)

	// --- GGG integration (optional, gracefully disabled when not configured) ---
	mux.HandleFunc("GET /ggg/status", h.GGGStatus)
	mux.HandleFunc("GET /ggg/auth", h.GGGAuth)
	mux.HandleFunc("GET /ggg/callback", h.GGGCallback)
	mux.HandleFunc("POST /runs/{id}/snapshots/ggg", h.GGGSyncSnapshot)

	return corsMiddleware(mux)
}

// corsMiddleware adds permissive CORS headers for local development.
// In production this should be tightened to the actual frontend origin.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
