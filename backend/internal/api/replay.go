package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/poe1-trainer/internal/integration/logtail"
	"github.com/poe1-trainer/internal/progress"
	runpkg "github.com/poe1-trainer/internal/run"
)

// ReplayLog handles POST /runs/{id}/replay-log.
//
// Odczytuje plik Client.txt od początku, przetwarza każdą linię przez parser
// logtail i przekierowuje zdarzenia domenowe do wskazanego runa. Umożliwia
// weryfikację czy splity pojawiają się przy właściwych krokach bez potrzeby
// przejaścia kampanii na żywo.
func (h *Handlers) ReplayLog(w http.ResponseWriter, r *http.Request) {
	runID, ok := intPathParam(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid run id")
		return
	}

	var req ReplayLogRequest
	if r.ContentLength != 0 {
		if decErr := json.NewDecoder(r.Body).Decode(&req); decErr != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
	}

	// Resolve log path: request overrides server config overrides logtail default.
	logPath := h.logPath
	if req.LogPath != "" {
		logPath = req.LogPath
	}
	if logPath == "" {
		defaultCfg := logtail.DefaultConfig()
		logPath = defaultCfg.LogPath
	}
	if logPath == "" {
		writeError(w, http.StatusBadRequest, "brak ścieżki do pliku logu; ustaw LOG_PATH lub podaj log_path w żądaniu")
		return
	}

	// Resolve timezone.
	logLoc := h.logLocation
	if req.LogTZ != "" {
		loc, locErr := time.LoadLocation(req.LogTZ)
		if locErr != nil {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid log_tz: %v", locErr))
			return
		}
		logLoc = loc
	}

	// Verify run exists (GetRun fails for unknown IDs).
	if _, getErr := h.runRepo.GetRun(r.Context(), runID); getErr != nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("run %d not found", runID))
		return
	}

	f, openErr := os.Open(logPath) // #nosec G304 — lokalna ścieżka konfigurowana przez operatora
	if openErr != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("cannot open log file: %v", openErr))
		return
	}
	defer f.Close()

	ctx := r.Context()
	scanner := bufio.NewScanner(f)
	// Increase buffer for exceptionally long lines.
	scanner.Buffer(make([]byte, 512*1024), 512*1024)

	var linesRead, eventsDispatched, parseErrors int
	started := time.Now()

	for scanner.Scan() {
		line := scanner.Text()
		linesRead++

		parsed, parseErr := logtail.ParseLine(line, logLoc)
		if parseErr != nil {
			parseErrors++
			continue
		}
		if parsed == nil {
			continue
		}

		ev := replayParsedToEvent(parsed)
		if ev == nil {
			continue
		}

		if dispatchErr := h.dispatchEventToRun(ctx, runID, *ev); dispatchErr != nil {
			slog.Warn("replay-log: dispatch error", "run_id", runID, "err", dispatchErr)
			continue
		}
		eventsDispatched++
	}

	if scanErr := scanner.Err(); scanErr != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("read error: %v", scanErr))
		return
	}

	writeJSON(w, http.StatusOK, ReplayLogResponse{
		LinesRead:        linesRead,
		EventsDispatched: eventsDispatched,
		ParseErrors:      parseErrors,
		DurationMs:       time.Since(started).Milliseconds(),
	})
	h.NotifyRunUpdate(int64(runID))
}

// replayParsedToEvent converts a logtail.ParsedLine to a progress.DomainEvent.
// Returns nil for line kinds that have no corresponding domain event.
func replayParsedToEvent(p *logtail.ParsedLine) *progress.DomainEvent {
	switch p.Kind {
	case logtail.ParsedKindAreaEntered:
		ev := progress.NewAreaEnteredEvent(0, p.AreaName, p.Timestamp)
		return &ev
	case logtail.ParsedKindAreaGenerated:
		ev := progress.NewAreaGeneratedEvent(0, p.AreaLevel, p.AreaCode, p.AreaSeed, p.Timestamp)
		return &ev
	case logtail.ParsedKindLevelUp:
		ev := progress.NewLevelUpEvent(0, p.Level, p.Timestamp)
		return &ev
	case logtail.ParsedKindPassiveAllocated:
		ev := progress.NewPassiveAllocatedEvent(0, p.PassiveID, p.PassiveName, p.Timestamp)
		return &ev
	case logtail.ParsedKindTradeAccepted:
		ev := progress.NewTradeAcceptedEvent(0, p.Timestamp)
		return &ev
	}
	return nil
}

// dispatchEventToRun applies a single domain event to the run with the given ID.
// Mirrors the logic in cmd/server/main.go:dispatchLogtailEvent, but targets a
// specific run rather than the currently active run.
func (h *Handlers) dispatchEventToRun(ctx context.Context, runID int, ev progress.DomainEvent) error {
	switch ev.Kind {
	case progress.KindAreaEntered:
		if ev.Area == nil {
			return nil
		}
		return h.runs.HandleAreaEvent(ctx, runID, runpkg.AreaEvent{AreaName: ev.Area.AreaName, OccurredAt: ev.OccurredAt})
	case progress.KindAreaGenerated:
		if ev.AreaGenerated == nil {
			return nil
		}
		return h.runs.HandleAreaGenerated(ctx, runID, ev.AreaGenerated.AreaCode, ev.AreaGenerated.AreaLevel, ev.OccurredAt)
	case progress.KindLevelUp:
		if ev.Level == nil {
			return nil
		}
		return h.runs.RecordLogEvent(ctx, runID, runpkg.EventLevelUp, map[string]string{
			"level": fmt.Sprint(ev.Level.Level),
		})
	case progress.KindPassiveAllocated:
		if ev.Passive == nil {
			return nil
		}
		return h.runs.RecordLogEvent(ctx, runID, runpkg.EventPassiveAllocated, map[string]string{
			"passive_id":   ev.Passive.PassiveID,
			"passive_name": ev.Passive.PassiveName,
		})
	case progress.KindTradeAccepted:
		return h.runs.RecordLogEvent(ctx, runID, runpkg.EventTradeAccepted, map[string]string{
			"outcome": "accepted",
		})
	}
	return nil
}
