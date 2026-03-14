package logtail

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/poe1-trainer/internal/progress"
)

// Status opisuje bieżący stan operacyjny Watchera.
type Status string

const (
	// StatusWaitingForFile — plik logu jeszcze nie istnieje na dysku.
	StatusWaitingForFile Status = "waiting_for_file"
	// StatusActive — Watcher aktywnie odczytuje nowe linie.
	StatusActive Status = "active"
	// StatusWaitingForNewLines — plik otwarty, brak nowych danych od krótkiego czasu.
	StatusWaitingForNewLines Status = "waiting_for_new_lines"
	// StatusGameNotRunning — brak nowych danych przez dłuższy czas; gra prawdopodobnie wyłączona.
	StatusGameNotRunning Status = "game_not_running"
	// StatusParserError — wystąpił nieoczekiwany błąd parsowania linii strukturalnej.
	StatusParserError Status = "parser_error"
)

// EventSink odbiera znormalizowane zdarzenia domenowe produkowane przez Watchera.
// Implementacja musi być nieblokująca lub używać buforowanego kanału.
type EventSink interface {
	Emit(event progress.DomainEvent)
}

// EventSinkFunc jest adapterem funkcyjnym dla EventSink.
type EventSinkFunc func(progress.DomainEvent)

// Emit implementuje EventSink.
func (f EventSinkFunc) Emit(e progress.DomainEvent) { f(e) }

// ChannelSink dostarcza zdarzenia do buforowanego kanału.
// Jeśli kanał jest pełny, zdarzenie jest odrzucane z ostrzeżeniem w logu.
type ChannelSink struct {
	ch chan<- progress.DomainEvent
}

// NewChannelSink tworzy sink oparty na kanale.
func NewChannelSink(ch chan<- progress.DomainEvent) *ChannelSink {
	return &ChannelSink{ch: ch}
}

// Emit implementuje EventSink.
func (s *ChannelSink) Emit(e progress.DomainEvent) {
	select {
	case s.ch <- e:
	default:
		slog.Warn("logtail: sink full, dropping event", "kind", string(e.Kind))
	}
}

// StatusObserver jest wywoływany przy każdej zmianie stanu Watchera.
// err jest non-nil tylko dla StatusParserError.
type StatusObserver func(s Status, err error)

// checkpoint to utrwalony stan Watchera między uruchomieniami.
type checkpoint struct {
	Offset int64 `json:"offset"`
}

// Watcher monitoruje plik Client.txt PoE1 i emituje zdarzenia domenowe.
//
// Zasady projektowe:
//   - Tylko odczyt: Watcher nigdy nie pisze do pliku logu gry.
//   - Odporność: przetrwa rotację pliku, restart gry i restart procesu.
//   - Obserwowalność: zmiany stanu raportowane przez StatusObserver.
//   - Separacja: produkuje progress.DomainEvent konsumowane przez EventSink.
type Watcher struct {
	cfg             Config
	sink            EventSink
	observer        StatusObserver
	rawLineObserver func(string)
	status          atomic.Value // stores Status
	done            chan struct{}
}

// New tworzy nowy Watcher.
//
//   cfg      — konfiguracja; użyj DefaultConfig() dla wartości domyślnych.
//   sink     — cel zdarzeń domenowych; nie może być nil.
//   observer — wywoływany przy każdej zmianie stanu; może być nil.
func New(cfg Config, sink EventSink, observer StatusObserver) *Watcher {
	w := &Watcher{
		cfg:      cfg,
		sink:     sink,
		observer: observer,
		done:     make(chan struct{}),
	}
	w.status.Store(StatusWaitingForFile)
	return w
}

// Status zwraca bieżący stan Watchera. Bezpieczne wielowątkowo.
func (w *Watcher) Status() Status {
	return w.status.Load().(Status)
}

// SetRawLineObserver rejestruje callback wywoływany dla każdej surowej linii odczytanej z pliku.
// Musi być ustawiony przed wywołaniem Start. Nil wyłącza obserwację.
func (w *Watcher) SetRawLineObserver(fn func(line string)) {
	w.rawLineObserver = fn
}

// Start uruchamia goroutine Watchera. Wraca natychmiast.
// Zatrzymaj przez Stop() lub anulując ctx.
func (w *Watcher) Start(ctx context.Context) {
	go w.run(ctx)
}

// Stop sygnalizuje goroutine Watchera do zakończenia. Bezpieczne do wywołania wielokrotnie.
func (w *Watcher) Stop() {
	select {
	case <-w.done:
	default:
		close(w.done)
	}
}

// run to główna pętla Watchera.
func (w *Watcher) run(ctx context.Context) {
	ticker := time.NewTicker(w.cfg.PollInterval)
	defer ticker.Stop()

	var (
		f        *os.File
		reader   *bufio.Reader
		partial  string
		lastLine time.Time
	)

	setStatus := func(s Status, err error) {
		old := w.status.Swap(s).(Status)
		if old == s {
			return
		}
		if w.observer != nil {
			w.observer(s, err)
		}
	}

	closeFile := func() {
		if f != nil {
			_ = f.Close()
			f = nil
			reader = nil
			partial = ""
		}
	}
	defer closeFile()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.done:
			return
		case <-ticker.C:
		}

		// Faza 1: upewnij się, że plik jest otwarty.
		if f == nil {
			opened, err := w.openFile()
			if err != nil {
				setStatus(StatusWaitingForFile, err)
				continue
			}
			f = opened
			reader = bufio.NewReader(f)
			lastLine = time.Now()
			setStatus(StatusActive, nil)
		}

		if reopen, err := w.shouldReopenFile(f); reopen {
			if err != nil {
				slog.Warn("logtail: file changed, reopening", "path", w.cfg.LogPath, "err", err)
			} else {
				slog.Info("logtail: file changed, reopening", "path", w.cfg.LogPath)
			}
			closeFile()
			continue
		}

		// Faza 2: odczytaj dostępne linie.
		gotData := false
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				partial += line
				if err == io.EOF {
					break
				}
				// Nieoczekiwany błąd odczytu — plik mógł zostać zastąpiony.
				slog.Warn("logtail: read error, reopening file", "path", w.cfg.LogPath, "err", err)
				closeFile()
				break
			}
			full := partial + line
			partial = ""
			full = strings.TrimRight(full, "\r\n")
			gotData = true
			lastLine = time.Now()
			if w.rawLineObserver != nil {
				w.rawLineObserver(full)
			}

			parsed, parseErr := ParseLine(full)
			if parseErr != nil {
				slog.Warn("logtail: parser error", "line", full, "err", parseErr)
				setStatus(StatusParserError, parseErr)
				continue
			}
			if parsed == nil {
				continue
			}

			setStatus(StatusActive, nil)
			w.emitEvent(parsed)
		}

		// Faza 3: zaktualizuj status na podstawie czasu bezczynności.
		if f != nil && !gotData {
			idle := time.Since(lastLine)
			switch {
			case idle >= w.cfg.GameNotRunningAfter:
				setStatus(StatusGameNotRunning, nil)
			case idle >= w.cfg.IdleAfter:
				setStatus(StatusWaitingForNewLines, nil)
			}
		}

		// Faza 4: zapisz checkpoint.
		if f != nil {
			if off, err := f.Seek(0, io.SeekCurrent); err == nil {
				if err := w.saveCheckpoint(off); err != nil {
					slog.Warn("logtail: checkpoint save failed", "err", err)
				}
			}
		}
	}
}

// openFile otwiera plik logu i ustawia pozycję odczytu na ostatnim
// checkpointowanym offsecie, lub na końcu pliku jeśli brak checkpointu.
func (w *Watcher) openFile() (*os.File, error) {
	f, err := os.Open(w.cfg.LogPath) // #nosec G304 — lokalna ścieżka konfigurowana przez użytkownika
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("stat log file: %w", err)
	}
	offset := w.loadCheckpoint()
	if offset > 0 {
		if info.Size() < offset {
			slog.Info("logtail: stale checkpoint beyond EOF, resetting to end", "offset", offset, "size", info.Size())
			offset = info.Size()
			if err := w.saveCheckpoint(offset); err != nil {
				slog.Warn("logtail: checkpoint reset failed", "err", err)
			}
		}
		if _, err := f.Seek(offset, io.SeekStart); err != nil {
			// Checkpoint nieaktualny (plik mógł być obcięty) — start od końca.
			slog.Warn("logtail: checkpoint seek failed, starting from end", "offset", offset, "err", err)
			if _, err2 := f.Seek(0, io.SeekEnd); err2 != nil {
				_ = f.Close()
				return nil, fmt.Errorf("seek to end: %w", err2)
			}
		}
	} else {
		// Brak checkpointu: start od końca pliku, żeby nie odtwarzać starej historii.
		if _, err := f.Seek(0, io.SeekEnd); err != nil {
			_ = f.Close()
			return nil, fmt.Errorf("seek to end: %w", err)
		}
	}
	return f, nil
}

func (w *Watcher) shouldReopenFile(f *os.File) (bool, error) {
	openInfo, err := f.Stat()
	if err != nil {
		return true, fmt.Errorf("stat open file: %w", err)
	}
	pathInfo, err := os.Stat(w.cfg.LogPath) // #nosec G304 — lokalna ścieżka konfigurowana przez użytkownika
	if err != nil {
		return true, fmt.Errorf("stat log path: %w", err)
	}
	if !os.SameFile(openInfo, pathInfo) {
		return true, nil
	}
	offset, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		return true, fmt.Errorf("read current offset: %w", err)
	}
	if pathInfo.Size() < offset {
		return true, nil
	}
	return false, nil
}

// emitEvent konwertuje ParsedLine na DomainEvent i wysyła go do sink.
// RunID jest ustawiony na 0 — integracja z run.Service uzupełnia RunID przed przekazaniem do silnika.
func (w *Watcher) emitEvent(p *ParsedLine) {
	switch p.Kind {
	case ParsedKindAreaEntered:
		w.sink.Emit(progress.NewAreaEnteredEvent(0, p.AreaName, p.Timestamp))
	case ParsedKindAreaGenerated:
		w.sink.Emit(progress.NewAreaGeneratedEvent(0, p.AreaLevel, p.AreaCode, p.AreaSeed, p.Timestamp))
	case ParsedKindLevelUp:
		w.sink.Emit(progress.NewLevelUpEvent(0, p.Level, p.Timestamp))
	case ParsedKindPassiveAllocated:
		w.sink.Emit(progress.NewPassiveAllocatedEvent(0, p.PassiveID, p.PassiveName, p.Timestamp))
	case ParsedKindTradeAccepted:
		w.sink.Emit(progress.NewTradeAcceptedEvent(0, p.Timestamp))
	}
}

// ─── Checkpoint ──────────────────────────────────────────────────────────────

func (w *Watcher) loadCheckpoint() int64 {
	if w.cfg.CheckpointPath == "" {
		return 0
	}
	data, err := os.ReadFile(w.cfg.CheckpointPath) // #nosec G304 — lokalna ścieżka
	if err != nil {
		return 0
	}
	var cp checkpoint
	if err := json.Unmarshal(data, &cp); err != nil {
		return 0
	}
	return cp.Offset
}

func (w *Watcher) saveCheckpoint(offset int64) error {
	if w.cfg.CheckpointPath == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(w.cfg.CheckpointPath), 0o700); err != nil {
		return fmt.Errorf("create checkpoint dir: %w", err)
	}
	data, err := json.Marshal(checkpoint{Offset: offset})
	if err != nil {
		return err
	}
	// Zapis atomowy: najpierw plik tymczasowy, potem rename.
	tmp := w.cfg.CheckpointPath + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("write checkpoint tmp: %w", err)
	}
	if err := os.Rename(tmp, w.cfg.CheckpointPath); err != nil {
		return fmt.Errorf("rename checkpoint: %w", err)
	}
	return nil
}
