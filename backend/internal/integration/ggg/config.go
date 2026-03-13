package ggg

import (
	"os"
	"path/filepath"
)

const (
	// oauthBaseURL to punkt wejścia dla OAuth 2.0 GGG.
	oauthBaseURL = "https://www.pathofexile.com"
	// apiBaseURL to baza REST API GGG (OAuth-gated).
	apiBaseURL = "https://api.pathofexile.com"
)

// Config przechowuje poświadczenia OAuth potrzebne do integracji z GGG API.
// Wszystkie pola są opcjonalne — gdy ClientID jest puste, integracja jest wyłączona.
type Config struct {
	ClientID      string
	ClientSecret  string
	RedirectURI   string
	// TokenFilePath to ścieżka do pliku JSON z tokenem OAuth.
	// Domyślnie: $XDG_CACHE_HOME/poe1-trainer/ggg_token.json (lub odpowiednik OS).
	TokenFilePath string
}

// IsConfigured zwraca true, jeśli poświadczenia OAuth są ustawione.
func (c Config) IsConfigured() bool {
	return c.ClientID != "" && c.ClientSecret != ""
}

// ConfigFromEnv odczytuje konfigurację GGG z zmiennych środowiskowych.
//
// Obsługiwane zmienne:
//   - GGG_CLIENT_ID      — wymagany do włączenia integracji
//   - GGG_CLIENT_SECRET  — wymagany do włączenia integracji
//   - GGG_REDIRECT_URI   — domyślnie http://localhost:8080/ggg/callback
//   - GGG_TOKEN_FILE     — ścieżka do pliku z tokenem; domyślnie katalog cache OS
func ConfigFromEnv() Config {
	return Config{
		ClientID:      os.Getenv("GGG_CLIENT_ID"),
		ClientSecret:  os.Getenv("GGG_CLIENT_SECRET"),
		RedirectURI:   envOrDefault("GGG_REDIRECT_URI", "http://localhost:8080/ggg/callback"),
		TokenFilePath: envOrDefault("GGG_TOKEN_FILE", defaultTokenFilePath()),
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func defaultTokenFilePath() string {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = os.TempDir()
	}
	return filepath.Join(cacheDir, "poe1-trainer", "ggg_token.json")
}
