package logtail

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// logTimestampLayout jest formatem znacznika czasu używanym w Client.txt.
const logTimestampLayout = "2006/01/02 15:04:05"

// ParsedLineKind identyfikuje typ zdarzenia zdekodowanego z linii logu.
type ParsedLineKind string

const (
	// ParsedKindAreaEntered — gracz wszedł do nowej strefy.
	ParsedKindAreaEntered ParsedLineKind = "area_entered"
	// ParsedKindLevelUp — postać awansowała na wyższy poziom.
	ParsedKindLevelUp ParsedLineKind = "level_up"
)

// ParsedLine reprezentuje zdekodowaną linię logu niosącą zdarzenie.
// Dokładnie jedno pole zdarzenia jest niezerowe na instancję.
type ParsedLine struct {
	Timestamp time.Time
	Kind      ParsedLineKind

	AreaName string // ustawione gdy Kind == ParsedKindAreaEntered
	Level    int    // ustawione gdy Kind == ParsedKindLevelUp
}

var (
	// reLogLine dopasowuje obowiązkowy prefiks każdej strukturalnej linii Client.txt:
	// "2024/01/01 12:34:56 625140781 cffb97d2 [SEVERITY Client 12345] TREŚĆ"
	reLogLine = regexp.MustCompile(
		`^(\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}) \S+ \S+ \[\w+ Client \d+\] (.*)$`,
	)

	// reAreaEntered dopasowuje treść linii przejścia do strefy:
	// ": You have entered Twilight Strand."
	reAreaEntered = regexp.MustCompile(`^: You have entered (.+?)\.?\s*$`)

	// reLevelUp dopasowuje treść linii awansu postaci:
	// ": Level up! You are now level 10."
	reLevelUp = regexp.MustCompile(`^: Level up! You are now level (\d+)\.$`)
)

// ParseLine dekoduje pojedynczą surową linię Client.txt.
//
// Zwraca (nil, nil) dla linii strukturalnie poprawnych, ale nieinteresujących
// (czat, upadek przedmiotów, aktywacja umiejętności itp.).
//
// Zwraca (nil, error) gdy linia pasuje do struktury logu, ale jej treść
// nie może być zdekodowana — wywołujący powinien traktować to jako StatusParserError.
//
// Zwraca (*ParsedLine, nil) dla zdekodowanych linii ze zdarzeniem.
func ParseLine(raw string) (*ParsedLine, error) {
	raw = strings.TrimRight(raw, "\r\n ")
	if raw == "" {
		return nil, nil
	}

	m := reLogLine.FindStringSubmatch(raw)
	if m == nil {
		// Linia niestrukturalna (np. nagłówek startowy) — ignorowana.
		return nil, nil
	}

	ts, err := time.ParseInLocation(logTimestampLayout, m[1], time.Local)
	if err != nil {
		// Zniekształcony znacznik czasu w linii strukturalnej = błąd parsera.
		return nil, fmt.Errorf("parse timestamp %q: %w", m[1], err)
	}

	body := m[2]

	if am := reAreaEntered.FindStringSubmatch(body); am != nil {
		return &ParsedLine{
			Timestamp: ts,
			Kind:      ParsedKindAreaEntered,
			AreaName:  strings.TrimSpace(am[1]),
		}, nil
	}

	if lm := reLevelUp.FindStringSubmatch(body); lm != nil {
		lvl, err := strconv.Atoi(lm[1])
		if err != nil {
			return nil, fmt.Errorf("parse level %q: %w", lm[1], err)
		}
		return &ParsedLine{
			Timestamp: ts,
			Kind:      ParsedKindLevelUp,
			Level:     lvl,
		}, nil
	}

	return nil, nil
}
