package logtail

import (
	"bufio"
	"context"
	"io"
	"os"
	"regexp"
	"strings"
	"time"
)

// AreaEnteredFunc is called each time the watcher detects a new area entry.
type AreaEnteredFunc func(areaName string)

// Watcher tails Client.txt and emits area-entered events.
// It reads the file in a background goroutine and does NOT modify any state
// directly — it only calls the provided callback.
type Watcher struct {
	path     string
	callback AreaEnteredFunc
	done     chan struct{}
}

// reAreaEntered matches lines like:
// 2024/01/01 12:00:00 625140781 cffb97d2 [INFO Client 12345] : You have entered Twilight Strand.
var reAreaEntered = regexp.MustCompile(`\[INFO Client \d+\] : You have entered (.+?)\.?\s*$`)

// New creates a new Watcher for the given Client.txt path.
func New(path string, callback AreaEnteredFunc) *Watcher {
	return &Watcher{
		path:     path,
		callback: callback,
		done:     make(chan struct{}),
	}
}

// Start begins tailing Client.txt. It seeks to the end of the file first so
// only new entries are processed. Call Stop() to shut down the watcher.
func (w *Watcher) Start(ctx context.Context) error {
	f, err := os.Open(w.path) // #nosec G304 — user-provided path, local file only
	if err != nil {
		return err
	}

	// Seek to end so we don't replay old log entries from previous sessions.
	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		f.Close()
		return err
	}

	go w.tail(ctx, f)
	return nil
}

// Stop signals the watcher goroutine to exit.
func (w *Watcher) Stop() {
	close(w.done)
}

func (w *Watcher) tail(ctx context.Context, f *os.File) {
	defer f.Close()
	reader := bufio.NewReader(f)

	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	var partial string
	for {
		select {
		case <-ctx.Done():
			return
		case <-w.done:
			return
		case <-ticker.C:
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					// Partial line: buffer it.
					partial += line
					break
				}
				full := partial + line
				partial = ""
				full = strings.TrimRight(full, "\r\n")
				w.processLine(full)
			}
		}
	}
}

func (w *Watcher) processLine(line string) {
	m := reAreaEntered.FindStringSubmatch(line)
	if m == nil {
		return
	}
	areaName := strings.TrimSpace(m[1])
	if areaName != "" {
		w.callback(areaName)
	}
}
