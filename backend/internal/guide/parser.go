package guide

import (
	"bufio"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ParseMarkdown parses a guide markdown file into a Guide struct.
// It supports the concrete format used in guides/stormburst_campaign_v1.md.
func ParseMarkdown(slug, title, buildName, version, content string) (*Guide, error) {
	g := &Guide{
		Slug:      slug,
		Title:     title,
		BuildName: buildName,
		Version:   version,
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	currentAct := 0
	stepNumber := 0
	sortOrder := 0

	// Detect "## Akt N" or "## Act N" headings.
	reAct := regexp.MustCompile(`(?i)^##\s+akt\s+(\d+)`)
	// Detect "## Zasady ogólne" — skip as preamble.
	rePreamble := regexp.MustCompile(`(?i)^##\s+(zasady|general|preamble)`)
	// Detect ordered list items: "1. text", "2. text".
	reStep := regexp.MustCompile(`^\s*(\d+)\.\s+(.+)`)
	// Detect location order lines.
	reLocOrder := regexp.MustCompile(`(?i)kolejno[śs][ćc] lokacji`)
	// Gem colors by span style.
	reGemRed := regexp.MustCompile(`color:#ff6a2f[^>]*>([^<]+)<`)
	reGemGreen := regexp.MustCompile(`color:#66ff66[^>]*>([^<]+)<`)
	reGemBlue := regexp.MustCompile(`color:#7f7fff[^>]*>([^<]+)<`)
	// Strip all HTML tags for plain text.
	reHTML := regexp.MustCompile(`<[^>]+>`)

	inPreamble := false
	var locationOrder string

	for scanner.Scan() {
		line := scanner.Text()

		// Detect act heading.
		if m := reAct.FindStringSubmatch(line); m != nil {
			n, _ := strconv.Atoi(m[1])
			currentAct = n
			inPreamble = false
			locationOrder = ""
			continue
		}
		// Skip preamble sections.
		if rePreamble.MatchString(line) {
			inPreamble = true
			currentAct = 0
			continue
		}
		if strings.HasPrefix(line, "#") {
			inPreamble = false
		}
		if inPreamble || currentAct == 0 {
			continue
		}

		// Capture location order line for the act.
		if reLocOrder.MatchString(line) {
			locationOrder = reHTML.ReplaceAllString(line, "")
			_ = locationOrder // stored for potential future use
			continue
		}

		// Match numbered steps.
		if m := reStep.FindStringSubmatch(line); m != nil {
			rawText := m[2]
			plain := reHTML.ReplaceAllString(rawText, "")
			plain = strings.TrimSpace(plain)

			stepNumber++
			sortOrder++

			area := ExtractArea(rawText)
			questName := ExtractQuestName(plain)
			stepType := ClassifyStepType(plain, rawText)
			completionMode := InferCompletionMode(stepType, area)
			conditions := BuildConditions(stepType, area)

			step := Step{
				StepNumber:     stepNumber,
				Act:            currentAct,
				Section:        ActSection(currentAct),
				Title:          truncate(plain, 100),
				Description:    rawText,
				Area:           area,
				QuestName:      questName,
				StepType:       stepType,
				IsCheckpoint:   isCheckpointLine(plain),
				RequiresManual: completionMode != CompletionLogtail,
				CompletionMode: completionMode,
				SortOrder:      sortOrder,
				Conditions:     conditions,
			}

			// Extract gem requirements from the raw HTML line.
			for _, m2 := range reGemRed.FindAllStringSubmatch(rawText, -1) {
				step.GemRequirements = append(step.GemRequirements, GemRequirement{
					GemName: m2[1], Color: "red",
				})
			}
			for _, m2 := range reGemGreen.FindAllStringSubmatch(rawText, -1) {
				step.GemRequirements = append(step.GemRequirements, GemRequirement{
					GemName: m2[1], Color: "green",
				})
			}
			for _, m2 := range reGemBlue.FindAllStringSubmatch(rawText, -1) {
				step.GemRequirements = append(step.GemRequirements, GemRequirement{
					GemName: m2[1], Color: "blue",
				})
			}

			g.Steps = append(g.Steps, step)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("guide: parse: %w", err)
	}
	if len(g.Steps) == 0 {
		return nil, fmt.Errorf("guide: no steps found in %q", slug)
	}
	return g, nil
}

// isCheckpointLine returns true when the plaintext step looks like a
// major milestone (boss kill, labyrinth, act transition).
func isCheckpointLine(text string) bool {
	keywords := []string{
		"zabij ", "kill ", "lab", "labyrinth", "przejdź do aktu",
		"przejdź do akt", "zabij bossa", "pokonaj",
	}
	lower := strings.ToLower(text)
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}
