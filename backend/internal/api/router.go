package api

import (
	"net/http"
)

// NewRouter builds and returns the HTTP router for all API endpoints.
// It uses the stdlib ServeMux with Go 1.22+ path parameter syntax.
func NewRouter(h *Handlers) http.Handler {
	mux := http.NewServeMux()

	// --- guides ---
	mux.HandleFunc("GET /guides", h.ListGuides)
	mux.HandleFunc("GET /guides/{slug}", h.GetGuide)
	mux.HandleFunc("GET /guides/{slug}/runs", h.ListRuns)
	mux.HandleFunc("GET /guides/{slug}/ranking", h.GetRanking)

	// --- runs ---
	mux.HandleFunc("POST /runs", h.CreateRun)
	mux.HandleFunc("GET /runs/{id}/state", h.GetRunState)
	mux.HandleFunc("GET /runs/{id}/recommendations", h.GetRunRecommendations)
	mux.HandleFunc("POST /runs/{id}/steps/{step_id}/confirm", h.ConfirmStep)
	mux.HandleFunc("POST /runs/{id}/finish", h.FinishRun)

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
