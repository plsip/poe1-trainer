package guide

import (
	"regexp"
	"strconv"
	"strings"
)

// ─── Heuristics — jawna, przeglądalna tabela reguł ───────────────────────────
//
// Każda reguła klasyfikuje krok na podstawie obecności kolorowych spanów
// (HTML z poradnika) oraz słów kluczowych w tekście bez tagów.
//
// Reguły sprawdzane są w kolejności (priorytet malejący); pierwsza pasująca
// wygrywa. Zasada "explicit over clever" — każda reguła to własna sekcja
// kodu, bez magicznych pętli po tablicy regułek.

// HTML-color flags extracted from raw line before classification.
type spanFlags struct {
	HasRed   bool // color:red — bossowie, wrogowie questowi
	HasGreen bool // color:#66ff66 — zielone gemy
	HasBlue  bool // color:#7f7fff — niebieskie gemy
	HasRedGem bool // color:#ff6a2f — czerwone gemy (aktywne skille)
	HasWhite bool // color:white — nazwy lokacji
	HasTeal  bool // color:teal — Labyrinth Trial
}

// extractSpanFlags walks the raw HTML line and returns a flags struct.
func extractSpanFlags(raw string) spanFlags {
	return spanFlags{
		HasRed:    strings.Contains(raw, "color:red"),
		HasGreen:  strings.Contains(raw, "color:#66ff66"),
		HasBlue:   strings.Contains(raw, "color:#7f7fff"),
		HasRedGem: strings.Contains(raw, "color:#ff6a2f"),
		HasWhite:  strings.Contains(raw, "color:white"),
		HasTeal:   strings.Contains(raw, "color:teal"),
	}
}

func (f spanFlags) HasAnyGem() bool {
	return f.HasGreen || f.HasBlue || f.HasRedGem
}

// ClassifyStepType determines the primary action type of a guide step.
// plain — tekst bez tagów HTML (do porównań słów kluczowych)
// raw   — surowy tekst z HTML (do analizy spanów kolorów)
//
// Kolejność sprawdzania reguł jest celowo jawna i sekwencyjna —
// każda reguła powinna być zrozumiała bez kontekstu pozostałych.
func ClassifyStepType(plain, raw string) StepType {
	lower := strings.ToLower(plain)
	f := extractSpanFlags(raw)

	// ─── 1. Labyrinth Trial ───────────────────────────────────────────────────
	// Teal span = "Labyrinth Trial" lub tekst zawiera frazę kluczową.
	if f.HasTeal || containsAny(lower, "labyrinth trial", "lab trial") {
		return StepTypeLabyrinth
	}

	// ─── 2. Boss kill ─────────────────────────────────────────────────────────
	// Czerwony span (boss/wróg) + czasownik zabijania.
	// Priorytet wyższy niż quest_reward, bo krok "Zabij X i odbierz nagrodę"
	// powinien być zaklasyfikowany jako boss_kill.
	if f.HasRed && containsAny(lower, "zabij ", "kill ", "pokonaj ", "zabić ") {
		return StepTypeBossKill
	}

	// ─── 3. Quest reward ──────────────────────────────────────────────────────
	// Zbieranie nagrody po ukończeniu questa.
	if containsAny(lower,
		"odbierz reward", "reward gemowy", "book of skill",
		"weź nagrodę", "weź book", "odbierz book",
	) {
		return StepTypeQuestReward
	}

	// ─── 4. Gem acquire ───────────────────────────────────────────────────────
	// Krop zawiera gem (dowolnego koloru) + czasownik nabycia.
	if f.HasAnyGem() && containsAny(lower,
		"weź ", "kup ", "odbierz ", "dostaniesz ", "wrzuć go", "wrzuć gem",
		"wrzuć na weapon", "kup go",
	) {
		return StepTypeGemAcquire
	}

	// ─── 5. Vendor recipe ─────────────────────────────────────────────────────
	if containsAny(lower, "vendor recipe") {
		return StepTypeVendorRecipe
	}

	// ─── 6. Craft ─────────────────────────────────────────────────────────────
	if containsAny(lower, "crafting bench", "crafting bencha", "zrób craft", "użyj crafting") {
		return StepTypeCraft
	}

	// ─── 7. Portal / logout ───────────────────────────────────────────────────
	if containsAny(lower, "postaw portal", "zrób logout", "wróć portalem", "użyj portalu", "wróć logoutem") {
		return StepTypePortal
	}

	// ─── 8. Navigation ────────────────────────────────────────────────────────
	// Krok zawiera obszar (biały span) i czasownik ruchu LUB "złap waypoint".
	if f.HasWhite && containsAny(lower,
		"idź", "biegnij", "wejdź", "przejdź", "złap waypoint",
		"wróć waypointem", "wróć do", "leć do", "dostań się do",
	) {
		return StepTypeNavigation
	}

	// ─── 9. Gear check ───────────────────────────────────────────────────────
	// Krok zawiera słowa kluczowe związane z ekwipunkiem lub flaskami.
	if containsAny(lower,
		"flask", "flasz", "flaski", "flaskę",
		"broń", "broni", "wand", "różdżk", "sceptre",
		"tarczy", "tarcza", "shield",
		"hełm", "helm",
		"buty", "boots",
		"chest", "zbroj", "pierś",
		"ring", "rękawic", "amulet", "pasek", "belt",
		"resist", "resyst", "odporności",
		"armor", "armour", "evasion",
	) {
		return StepTypeGearCheck
	}

	// ─── 10. Fallback ─────────────────────────────────────────────────────────
	return StepTypeGeneral
}

// ─── Area extraction ─────────────────────────────────────────────────────────

var reWhiteSpan = regexp.MustCompile(`color:white[^>]*>([^<]+)<`)

// ExtractArea returns the first white-coloured area name found in the raw HTML
// line, or "" if none is present.
func ExtractArea(raw string) string {
	if m := reWhiteSpan.FindStringSubmatch(raw); m != nil {
		return strings.TrimSpace(m[1])
	}
	return ""
}

// ─── Quest name extraction ───────────────────────────────────────────────────

// reQuestPhrase matches phrases like "questa Mercy Mission" or "quest The Caged Brute".
// Quest names in this guide are in English and start with a capital letter.
var reQuestPhrase = regexp.MustCompile(`(?i)questa?\s+([A-Z][A-Za-z' \-]+?)(?:[:.!?]|$|\s+i\s|\s+:)`)

// ExtractQuestName returns the quest name found in the plain text,
// or "" if no quest reference is detected.
func ExtractQuestName(plain string) string {
	if m := reQuestPhrase.FindStringSubmatch(plain); m != nil {
		return strings.TrimSpace(m[1])
	}
	return ""
}

// ─── Completion mode inference ───────────────────────────────────────────────

// InferCompletionMode decides how a step will be verified as complete.
//
// Reguła:
//   - navigation + znana lokacja → logtail (Client.txt loguje wejście do strefy)
//   - wszystkie pozostałe        → manual (gracz klika "Potwierdź")
//
// GGG API nie jest obsługiwana w MVP.
func InferCompletionMode(stepType StepType, area string) CompletionMode {
	if stepType == StepTypeNavigation && area != "" {
		return CompletionLogtail
	}
	return CompletionManual
}

// ─── Step conditions generation ──────────────────────────────────────────────

// BuildConditions generates the inference rule set for a step.
// For navigation steps with a known area, a logtail_area condition is added.
// All other steps get a manual_confirm condition as a fallback.
func BuildConditions(stepType StepType, area string) []StepCondition {
	var conds []StepCondition

	if stepType == StepTypeNavigation && area != "" {
		conds = append(conds, StepCondition{
			ConditionType: ConditionLogtailArea,
			Payload:       map[string]string{"area": area},
			Priority:      0,
			Notes:         "auto-detected: player enters the area",
		})
		return conds
	}

	// Manual confirm is the safe fallback for everything else.
	conds = append(conds, StepCondition{
		ConditionType: ConditionManualConfirm,
		Priority:      0,
	})
	return conds
}

// ─── Section name ─────────────────────────────────────────────────────────────

// ActSection returns the canonical section label for an act number.
// Returned string matches the Polish headings used in the guide.
func ActSection(act int) string {
	if act <= 0 {
		return ""
	}
	return "Akt " + strconv.Itoa(act)
}

// ─── helpers ──────────────────────────────────────────────────────────────────

// containsAny returns true if s contains at least one of the substrings.
func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
