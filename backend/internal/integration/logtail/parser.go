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
	// ParsedKindAreaGenerated — klient wygenerował nową instancję obszaru.
	ParsedKindAreaGenerated ParsedLineKind = "area_generated"
	// ParsedKindLevelUp — postać awansowała na wyższy poziom.
	ParsedKindLevelUp ParsedLineKind = "level_up"
	// ParsedKindPassiveAllocated — gracz przydzielił punkt pasywki.
	ParsedKindPassiveAllocated ParsedLineKind = "passive_allocated"
	// ParsedKindTradeAccepted — gracz zaakceptował wymianę lub transakcję NPC.
	ParsedKindTradeAccepted ParsedLineKind = "trade_accepted"
)

// ParsedLine reprezentuje zdekodowaną linię logu niosącą zdarzenie.
// Dokładnie jedno pole zdarzenia jest niezerowe na instancję.
type ParsedLine struct {
	Timestamp time.Time
	Kind      ParsedLineKind

	AreaName    string // ustawione gdy Kind == ParsedKindAreaEntered
	AreaCode    string // ustawione gdy Kind == ParsedKindAreaGenerated
	AreaLevel   int    // ustawione gdy Kind == ParsedKindAreaGenerated
	AreaSeed    int64  // ustawione gdy Kind == ParsedKindAreaGenerated
	Level       int    // ustawione gdy Kind == ParsedKindLevelUp
	PassiveID   string // ustawione gdy Kind == ParsedKindPassiveAllocated
	PassiveName string // ustawione gdy Kind == ParsedKindPassiveAllocated
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

	// reAreaGenerated dopasowuje techniczną linię generowania instancji:
	// "Generating level 4 area \"1_1_3\" with seed 2734685965"
	reAreaGenerated = regexp.MustCompile(`^Generating level (\d+) area "([^"]+)" with seed (\d+)\s*$`)

	// reLevelUpLegacy dopasowuje starszy format linii awansu postaci:
	// ": Level up! You are now level 10."
	reLevelUpLegacy = regexp.MustCompile(`^: Level up! You are now level (\d+)\.$`)

	// reLevelUpCurrent dopasowuje aktualny format linii awansu postaci:
	// ": fdogsb (Templar) is now level 3"
	reLevelUpCurrent = regexp.MustCompile(`^: .+? \([^)]+\) is now level (\d+)\s*$`)

	// rePassiveAllocated dopasowuje przydzielenie punktu pasywki:
	// "Successfully allocated passive skill id: elemental_damage722, name: Damage and Mana"
	rePassiveAllocated = regexp.MustCompile(`^Successfully allocated passive skill id: ([^,]+), name: (.+?)\s*$`)

	// reTradeAccepted dopasowuje zatwierdzoną wymianę.
	reTradeAccepted = regexp.MustCompile(`^: Trade accepted\.\s*$`)
)

// ParseLine dekoduje pojedynczą surową linię Client.txt.
//
// loc określa strefę czasową znaczników czasu w pliku logu — PoE zapisuje czas
// lokalny maszyny gracza. Przekaż nil aby użyć time.Local.
//
// Zwraca (nil, nil) dla linii strukturalnie poprawnych, ale nieinteresujących
// (czat, upadek przedmiotów, aktywacja umiejętności itp.).
//
// Zwraca (nil, error) gdy linia pasuje do struktury logu, ale jej treść
// nie może być zdekodowana — wywołujący powinien traktować to jako StatusParserError.
//
// Zwraca (*ParsedLine, nil) dla zdekodowanych linii ze zdarzeniem.
func ParseLine(raw string, loc *time.Location) (*ParsedLine, error) {
	raw = strings.TrimRight(raw, "\r\n ")
	if raw == "" {
		return nil, nil
	}

	m := reLogLine.FindStringSubmatch(raw)
	if m == nil {
		// Linia niestrukturalna (np. nagłówek startowy) — ignorowana.
		return nil, nil
	}

	if loc == nil {
		loc = time.Local
	}
	ts, err := time.ParseInLocation(logTimestampLayout, m[1], loc)
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

	if gm := reAreaGenerated.FindStringSubmatch(body); gm != nil {
		areaLevel, err := strconv.Atoi(gm[1])
		if err != nil {
			return nil, fmt.Errorf("parse area level %q: %w", gm[1], err)
		}
		areaSeed, err := strconv.ParseInt(gm[3], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse area seed %q: %w", gm[3], err)
		}
		return &ParsedLine{
			Timestamp: ts,
			Kind:      ParsedKindAreaGenerated,
			AreaCode:  strings.TrimSpace(gm[2]),
			AreaLevel: areaLevel,
			AreaSeed:  areaSeed,
		}, nil
	}

	if lm := reLevelUpCurrent.FindStringSubmatch(body); lm != nil {
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

	if lm := reLevelUpLegacy.FindStringSubmatch(body); lm != nil {
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

	if pm := rePassiveAllocated.FindStringSubmatch(body); pm != nil {
		return &ParsedLine{
			Timestamp:   ts,
			Kind:        ParsedKindPassiveAllocated,
			PassiveID:   strings.TrimSpace(pm[1]),
			PassiveName: strings.TrimSpace(pm[2]),
		}, nil
	}

	if reTradeAccepted.MatchString(body) {
		return &ParsedLine{
			Timestamp: ts,
			Kind:      ParsedKindTradeAccepted,
		}, nil
	}

	return nil, nil
}
