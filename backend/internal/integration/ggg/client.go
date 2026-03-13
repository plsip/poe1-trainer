package ggg

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client wywołuje GGG API używając tokenu OAuth.
// Implementuje CharacterProvider.
type Client struct {
	cfg   Config
	store *tokenStore
	http  *http.Client
}

// NewClient tworzy Client. Zwraca błąd jeśli konfiguracja jest niekompletna.
func NewClient(cfg Config) (*Client, error) {
	if !cfg.IsConfigured() {
		return nil, fmt.Errorf("ggg: OAuth nie skonfigurowany (brak GGG_CLIENT_ID / GGG_CLIENT_SECRET)")
	}
	return &Client{
		cfg:   cfg,
		store: newTokenStore(cfg.TokenFilePath),
		http:  &http.Client{Timeout: 15 * time.Second},
	}, nil
}

// AuthorizeURL zwraca URL, pod który przeglądarka użytkownika musi przejść
// aby rozpocząć przepływ OAuth.
// state powinno być nieprzewidywalną losową wartością (CSRF protection).
func (c *Client) AuthorizeURL(state string) string {
	v := url.Values{
		"client_id":     {c.cfg.ClientID},
		"response_type": {"code"},
		"scope":         {"account:profile account:characters"},
		"redirect_uri":  {c.cfg.RedirectURI},
		"state":         {state},
	}
	return oauthBaseURL + "/oauth/authorize?" + v.Encode()
}

// HandleCallback wymienia kod z callbacku OAuth na token i go persystuje.
// Wywołuj po otrzymaniu code + state na endpoincie GET /ggg/callback.
func (c *Client) HandleCallback(ctx context.Context, code string) (*Token, error) {
	return c.store.ExchangeCode(ctx, c.cfg, code)
}

// IsAvailable implementuje CharacterProvider.
// Zwraca true gdy jest zapisany ważny token.
func (c *Client) IsAvailable() bool {
	tok, err := c.store.Load()
	return err == nil && tok.Valid()
}

// ListCharacters implementuje CharacterProvider.
// Wymaga scope: account:characters.
func (c *Client) ListCharacters(ctx context.Context) ([]CharacterInfo, error) {
	tok, err := c.validToken()
	if err != nil {
		return nil, err
	}
	var chars []apiCharacter
	if err := c.get(ctx, tok, "/character", &chars); err != nil {
		return nil, err
	}
	out := make([]CharacterInfo, 0, len(chars))
	for _, ch := range chars {
		if !ch.Expired {
			out = append(out, CharacterInfo{
				Name:   ch.Name,
				Class:  ch.Class,
				Level:  ch.Level,
				League: ch.League,
				Realm:  ch.Realm,
			})
		}
	}
	return out, nil
}

// FetchSnapshot implementuje CharacterProvider.
// Pobiera ekwipunek i osadzone gemy dla podanej postaci.
// Wymaga scope: account:characters.
func (c *Client) FetchSnapshot(ctx context.Context, characterName, _ string) (*SnapshotData, error) {
	tok, err := c.validToken()
	if err != nil {
		return nil, err
	}

	// Rozwiązuje ID postaci na podstawie listy.
	var chars []apiCharacter
	if err := c.get(ctx, tok, "/character", &chars); err != nil {
		return nil, err
	}
	var charID string
	var charLevel int
	for _, ch := range chars {
		if strings.EqualFold(ch.Name, characterName) {
			charID = ch.ID
			charLevel = ch.Level
			break
		}
	}
	if charID == "" {
		return nil, fmt.Errorf("ggg: postać %q nie znaleziona w koncie", characterName)
	}

	var detail apiCharacterDetailResponse
	if err := c.get(ctx, tok, "/character/"+charID, &detail); err != nil {
		return nil, err
	}

	equippedItems := make(map[string]any, len(detail.Items))
	skills := make(map[string]any)
	for _, item := range detail.Items {
		equippedItems[item.InventoryID] = item
		if len(item.SocketedItems) > 0 {
			skills[item.InventoryID] = item.SocketedItems
		}
	}

	raw, _ := json.Marshal(detail)
	rawMap := make(map[string]any)
	_ = json.Unmarshal(raw, &rawMap)

	return &SnapshotData{
		Level:         charLevel,
		EquippedItems: equippedItems,
		Skills:        skills,
		RawResponse:   rawMap,
	}, nil
}

// validToken ładuje token i weryfikuje jego ważność.
func (c *Client) validToken() (*Token, error) {
	tok, err := c.store.Load()
	if err != nil {
		return nil, err
	}
	if !tok.Valid() {
		return nil, fmt.Errorf("ggg: brak ważnego tokenu — przejdź przez OAuth na GET /ggg/auth")
	}
	return tok, nil
}

// get wykonuje uwierzytelnione żądanie GET do GGG API i dekoduje JSON.
func (c *Client) get(ctx context.Context, tok *Token, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiBaseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	req.Header.Set("User-Agent", "poe1-trainer/1.0") // wymagane przez GGG ToS

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("ggg: GET %s: %w", path, err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return json.NewDecoder(resp.Body).Decode(out)
	case http.StatusUnauthorized:
		return fmt.Errorf("ggg: brak autoryzacji — token mógł wygasnąć, powtórz OAuth na GET /ggg/auth")
	default:
		return fmt.Errorf("ggg: GET %s: HTTP %d", path, resp.StatusCode)
	}
}
