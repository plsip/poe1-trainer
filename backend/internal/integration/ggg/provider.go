package ggg

import "context"

// CharacterInfo przechowuje podstawowe metadane postaci zwrócone przez dostawcę.
type CharacterInfo struct {
	Name   string `json:"name"`
	Class  string `json:"class"`
	Level  int    `json:"level"`
	League string `json:"league"`
	Realm  string `json:"realm"`
}

// SnapshotData to niezależny od dostawcy opis stanu postaci.
// Mapuje się na run.CharacterSnapshot do zapisu w bazie danych.
//
// Pola skalarne (Level) są gotowe do użycia przez reguły alertów.
// EquippedItems, Skills, RawResponse to cache z GGG API — pola pochodne,
// nie są wymagane do działania aplikacji.
type SnapshotData struct {
	Level         int
	EquippedItems map[string]any // klucz: slot (np. "BodyArmour"), wartość: item
	Skills        map[string]any // klucz: slot, wartość: lista osadzonych gemów
	RawResponse   map[string]any // pełna odpowiedź API — cache do wglądu
}

// CharacterProvider to kontrakt aplikacyjny dla pobierania danych postaci
// z zewnętrznego źródła.
//
// Implementują go:
//   - *Client — prawdziwy klient GGG HTTP (gdy OAuth skonfigurowany)
//   - NoopProvider — gdy OAuth nie skonfigurowany (graceful degradation)
type CharacterProvider interface {
	// IsAvailable zwraca true, gdy dostawca jest skonfigurowany i ma ważny token.
	IsAvailable() bool
	// ListCharacters zwraca wszystkie postacie dla zalogowanego konta.
	ListCharacters(ctx context.Context) ([]CharacterInfo, error)
	// FetchSnapshot pobiera aktualny stan postaci o podanej nazwie.
	// realm: "pc" (domyślnie), "xbox", "sony".
	FetchSnapshot(ctx context.Context, characterName, realm string) (*SnapshotData, error)
}
