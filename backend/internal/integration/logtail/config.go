package logtail

import (
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// Config parametryzuje działanie Watchera.
type Config struct {
	// LogPath to absolutna ścieżka do aktywnego pliku logu gry.
	// Użyj DefaultLogPath() lub DefaultConfig() dla wartości domyślnej.
	LogPath string

	// CheckpointPath to ścieżka do pliku, w którym Watcher zapisuje ostatnio
	// odczytany offset. Puste pole wyłącza checkpointowanie (start od końca pliku).
	CheckpointPath string

	// PollInterval określa jak często Watcher sprawdza nowe linie w pliku.
	// Domyślnie 250ms.
	PollInterval time.Duration

	// IdleAfter określa czas bez nowych danych po którym status zmienia się
	// na StatusWaitingForNewLines. Domyślnie 30s.
	IdleAfter time.Duration

	// GameNotRunningAfter określa czas bez nowych danych po którym status
	// zmienia się na StatusGameNotRunning. Musi być > IdleAfter. Domyślnie 5m.
	GameNotRunningAfter time.Duration

	// LogLocation określa strefę czasową używaną do interpretacji znaczników
	// czasu w Client.txt. PoE zapisuje czas lokalny maszyny gracza.
	// W środowisku Docker (UTC) należy ustawić LOG_TZ na właściwą strefę,
	// np. "Europe/Warsaw". Nil oznacza time.Local.
	LogLocation *time.Location
}

// DefaultConfig zwraca Config wypełniony sensownymi wartościami domyślnymi.
// Na Windows ścieżka logu wskazuje pierwszy istniejący kandydat Client.txt/LatestClient.txt.
func DefaultConfig() Config {
	return Config{
		LogPath:             DefaultLogPath(),
		CheckpointPath:      defaultCheckpointPath(),
		PollInterval:        250 * time.Millisecond,
		IdleAfter:           30 * time.Second,
		GameNotRunningAfter: 5 * time.Minute,
		LogLocation:         time.Local,
	}
}

// DefaultLogPath zwraca domyślną ścieżkę do logu klienta dla bieżącej platformy.
// Na Windows preferuje pierwszy istniejący kandydat z typowych lokalizacji PoE1.
// Na innych platformach zwraca pusty string — ścieżkę trzeba skonfigurować ręcznie.
func DefaultLogPath() string {
	if runtime.GOOS != "windows" {
		return ""
	}
	candidates := windowsLogPathCandidates()
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	if len(candidates) == 0 {
		return ""
	}
	return candidates[0]
}

func windowsLogPathCandidates() []string {
	seen := make(map[string]struct{})
	add := func(paths []string, value string) []string {
		if value == "" {
			return paths
		}
		if _, ok := seen[value]; ok {
			return paths
		}
		seen[value] = struct{}{}
		return append(paths, value)
	}

	var candidates []string
	steamRoots := []string{
		`C:\Program Files (x86)\Steam\steamapps\common\Path of Exile\logs`,
		`C:\Program Files\Steam\steamapps\common\Path of Exile\logs`,
	}
	for _, root := range steamRoots {
		candidates = add(candidates, filepath.Join(root, "Client.txt"))
		candidates = add(candidates, filepath.Join(root, "LatestClient.txt"))
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		docRoot := filepath.Join(home, "Documents", "My Games", "Path of Exile")
		candidates = add(candidates, filepath.Join(docRoot, "Client.txt"))
		candidates = add(candidates, filepath.Join(docRoot, "LatestClient.txt"))
		candidates = add(candidates, filepath.Join(docRoot, "logs", "Client.txt"))
		candidates = add(candidates, filepath.Join(docRoot, "logs", "LatestClient.txt"))
	}
	return candidates
}

func defaultCheckpointPath() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		dir = os.TempDir()
	}
	return filepath.Join(dir, "poe1-trainer", "logtail_checkpoint.json")
}
