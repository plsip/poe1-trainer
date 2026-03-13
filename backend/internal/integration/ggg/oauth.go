package ggg

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Token reprezentuje parę access + refresh token OAuth 2.0.
type Token struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scope        string    `json:"scope"`
	// Username to nazwa konta GGG zwrócona przez serwer tokenów.
	Username string `json:"username,omitempty"`
}

// Valid zwraca true, jeśli token jest niepusty i nie wygasł.
func (t *Token) Valid() bool {
	return t != nil && t.AccessToken != "" && time.Now().Before(t.ExpiresAt)
}

// tokenStore persystuje Token jako JSON na dysku.
type tokenStore struct {
	path string
}

func newTokenStore(path string) *tokenStore { return &tokenStore{path: path} }

// Load odczytuje zapisany token. Zwraca (nil, nil) gdy plik nie istnieje.
func (s *tokenStore) Load() (*Token, error) {
	data, err := os.ReadFile(s.path) // #nosec G304 — ścieżka z zaufanej konfiguracji
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("ggg: load token: %w", err)
	}
	var tok Token
	if err := json.Unmarshal(data, &tok); err != nil {
		return nil, fmt.Errorf("ggg: parse token: %w", err)
	}
	return &tok, nil
}

// Save zapisuje token na dysk z uprawnieniami 0600.
func (s *tokenStore) Save(tok *Token) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0700); err != nil {
		return fmt.Errorf("ggg: mkdir token dir: %w", err)
	}
	data, err := json.Marshal(tok)
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0600) // #nosec G306 — celowo 0600
}

// ExchangeCode wymienia kod autoryzacji na token i go persystuje.
// Wywołuj dokładnie raz po przekierowaniu z GGG OAuth callback.
func (s *tokenStore) ExchangeCode(ctx context.Context, cfg Config, code string) (*Token, error) {
	vals := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {cfg.ClientID},
		"client_secret": {cfg.ClientSecret},
		"redirect_uri":  {cfg.RedirectURI},
		"code":          {code},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		oauthBaseURL+"/oauth/token", strings.NewReader(vals.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "poe1-trainer/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ggg: exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ggg: exchange code: HTTP %d", resp.StatusCode)
	}

	var raw struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
		Username     string `json:"username"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("ggg: decode token response: %w", err)
	}
	tok := &Token{
		AccessToken:  raw.AccessToken,
		TokenType:    raw.TokenType,
		RefreshToken: raw.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(raw.ExpiresIn) * time.Second),
		Scope:        raw.Scope,
		Username:     raw.Username,
	}
	return tok, s.Save(tok)
}
