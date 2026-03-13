package ggg

import (
	"context"
	"fmt"
)

// NoopProvider jest zwracany gdy GGG OAuth nie jest skonfigurowany.
// Wszystkie wywołania zwracają błąd; IsAvailable zwraca false.
// Dzięki temu reszta aplikacji działa normalnie bez integracji z GGG.
type NoopProvider struct{}

// IsAvailable zawsze zwraca false.
func (NoopProvider) IsAvailable() bool { return false }

// ListCharacters zawsze zwraca błąd — integracja nie jest skonfigurowana.
func (NoopProvider) ListCharacters(_ context.Context) ([]CharacterInfo, error) {
	return nil, fmt.Errorf("ggg: OAuth nie skonfigurowany — ustaw GGG_CLIENT_ID i GGG_CLIENT_SECRET")
}

// FetchSnapshot zawsze zwraca błąd — integracja nie jest skonfigurowana.
func (NoopProvider) FetchSnapshot(_ context.Context, _, _ string) (*SnapshotData, error) {
	return nil, fmt.Errorf("ggg: OAuth nie skonfigurowany — ustaw GGG_CLIENT_ID i GGG_CLIENT_SECRET")
}
