package logtail

import (
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// Config parametryzuje działanie Watchera.
type Config struct {
	// LogPath to absolutna ścieżka do pliku Client.txt gry.
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
}

// DefaultConfig zwraca Config wypełniony sensownymi wartościami domyślnymi.
// Na Windows ścieżka logu wskazuje na standardową instalację PoE1 przez Steam.
func DefaultConfig() Config {
	return Config{
		LogPath:             DefaultLogPath(),
		CheckpointPath:      defaultCheckpointPath(),
		PollInterval:        250 * time.Millisecond,
		IdleAfter:           30 * time.Second,
		GameNotRunningAfter: 5 * time.Minute,
	}
}

// DefaultLogPath zwraca domyślną ścieżkę do Client.txt dla bieżącej platformy.
// Na Windows zwraca standardową lokalizację instalacji Steam.
// Na innych platformach zwraca pusty string — ścieżkę trzeba skonfigurować ręcznie.
func DefaultLogPath() string {
	if runtime.GOOS == "windows" {
		return `C:\Program Files (x86)\Steam\steamapps\common\Path of Exile\logs\Client.txt`
	}
	return ""
}

func defaultCheckpointPath() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		dir = os.TempDir()
	}
	return filepath.Join(dir, "poe1-trainer", "logtail_checkpoint.json")
}
