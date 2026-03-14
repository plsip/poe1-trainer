package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/poe1-trainer/internal/api"
	buildpkg "github.com/poe1-trainer/internal/build"
	"github.com/poe1-trainer/internal/db"
	"github.com/poe1-trainer/internal/guide"
	"github.com/poe1-trainer/internal/integration/ggg"
	"github.com/poe1-trainer/internal/integration/logtail"
	"github.com/poe1-trainer/internal/progress"
	"github.com/poe1-trainer/internal/recommendation"
	"github.com/poe1-trainer/internal/rule"
	runpkg "github.com/poe1-trainer/internal/run"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	dsn := envOrDefault("DATABASE_URL", "postgres://poe:poe@localhost:5432/poetrainer?sslmode=disable")
	port := envOrDefault("PORT", "8080")

	store, err := db.New(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer store.Close()
	log.Println("database connected and migrations applied")

	guideRepo := guide.NewRepository(store.Pool)
	buildRepo := buildpkg.NewRepository(store.Pool)
	runRepo := runpkg.NewRepository(store.Pool)
	runService := runpkg.NewService(runRepo, guideRepo)
	engine := recommendation.NewEngine()
	ruleEngine := rule.NewEngine()

	// Konfiguracja integracji GGG API (opcjonalna).
	// Bez ustawienia GGG_CLIENT_ID i GGG_CLIENT_SECRET aplikacja działa normalnie.
	gggCfg := ggg.ConfigFromEnv()
	var gggProvider ggg.CharacterProvider
	var gggClient *ggg.Client

	if gggCfg.IsConfigured() {
		c, err := ggg.NewClient(gggCfg)
		if err != nil {
			log.Printf("ggg: błąd inicjalizacji klienta: %v — integracja wyłączona", err)
			gggProvider = ggg.NoopProvider{}
		} else {
			gggClient = c
			gggProvider = c
			log.Println("ggg: OAuth skonfigurowany — integracja aktywna")
		}
	} else {
		log.Println("ggg: brak konfiguracji OAuth — integracja wyłączona (aplikacja działa normalnie)")
		gggProvider = ggg.NoopProvider{}
	}

	h := api.NewHandlers(buildRepo, guideRepo, runService, runRepo, engine, ruleEngine, gggProvider, gggClient)
	router := api.NewRouter(h)

	// Konfiguracja logtail watchera (opcjonalna).
	// LOG_PATH nadpisuje ścieżkę domyślną do LatestClient.txt, jeśli środowisko tego wymaga.
	ltCfg := logtail.DefaultConfig()
	if logPath := os.Getenv("LOG_PATH"); logPath != "" {
		ltCfg.LogPath = logPath
	}
	if ltCfg.LogPath != "" {

		ch := make(chan progress.DomainEvent, 64)
		watcher := logtail.New(ltCfg, logtail.NewChannelSink(ch), func(s logtail.Status, err error) {
			if err != nil {
				slog.Warn("logtail: zmiana stanu", "status", string(s), "err", err)
			} else {
				slog.Info("logtail: zmiana stanu", "status", string(s))
			}
		})
		watcher.SetRawLineObserver(h.EmitLogLine)
		watcher.Start(ctx)
		h.SetWatcherStatusFunc(func() string { return string(watcher.Status()) })
		log.Printf("logtail: nasłuchiwanie pliku %s", ltCfg.LogPath)

		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case ev := <-ch:
					if runID := dispatchLogtailEvent(ctx, runRepo, runService, ev); runID != 0 {
						h.NotifyRunUpdate(int64(runID))
					}
				}
			}
		}()
	} else {
		log.Println("logtail: brak ścieżki do logu gry — nasłuchiwanie wyłączone")
	}

	srv := &http.Server{
		Addr:        fmt.Sprintf(":%s", port),
		Handler:     router,
		ReadTimeout: 10 * time.Second,
		// WriteTimeout is intentionally omitted: SSE connections stream
		// indefinitely and would be killed by a hard write deadline.
		IdleTimeout: 60 * time.Second,
	}

	go func() {
		log.Printf("poe1-trainer backend listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// dispatchLogtailEvent odnajduje aktywny run i przekazuje zdarzenie do run.Service.
// Eventy bez aktywnego runa są po cichu odrzucane — gracz nie ma otwartego przebiegu.
// Zwraca ID przetworzonego runa lub 0 gdy nie było aktywnego runa / błąd.
func dispatchLogtailEvent(ctx context.Context, repo *runpkg.Repository, svc *runpkg.Service, ev progress.DomainEvent) int {
	active, err := repo.GetActiveRun(ctx)
	if err != nil {
		slog.Warn("logtail: błąd pobierania aktywnego runa", "err", err)
		return 0
	}
	if active == nil {
		return 0
	}

	switch ev.Kind {
	case progress.KindAreaEntered:
		if ev.Area == nil {
			return active.ID
		}
		if err := svc.HandleAreaEvent(ctx, active.ID, runpkg.AreaEvent{AreaName: ev.Area.AreaName, OccurredAt: ev.OccurredAt}); err != nil {
			slog.Warn("logtail: HandleAreaEvent error", "run_id", active.ID, "area", ev.Area.AreaName, "err", err)
		}
	case progress.KindAreaGenerated:
		if ev.AreaGenerated == nil {
			return active.ID
		}
		if err := svc.RecordLogEvent(ctx, active.ID, runpkg.EventAreaGenerated, map[string]string{
			"area_code":  ev.AreaGenerated.AreaCode,
			"area_level": fmt.Sprint(ev.AreaGenerated.AreaLevel),
			"seed":       fmt.Sprint(ev.AreaGenerated.Seed),
		}); err != nil {
			slog.Warn("logtail: RecordLogEvent area_generated error", "run_id", active.ID, "area_code", ev.AreaGenerated.AreaCode, "err", err)
		}
	case progress.KindLevelUp:
		if ev.Level == nil {
			return active.ID
		}
		slog.Info("logtail: level up", "run_id", active.ID, "level", ev.Level.Level)
		if err := svc.RecordLogEvent(ctx, active.ID, runpkg.EventLevelUp, map[string]string{
			"level": fmt.Sprint(ev.Level.Level),
		}); err != nil {
			slog.Warn("logtail: RecordLogEvent level_up error", "run_id", active.ID, "level", ev.Level.Level, "err", err)
		}
	case progress.KindPassiveAllocated:
		if ev.Passive == nil {
			return active.ID
		}
		if err := svc.RecordLogEvent(ctx, active.ID, runpkg.EventPassiveAllocated, map[string]string{
			"passive_id":   ev.Passive.PassiveID,
			"passive_name": ev.Passive.PassiveName,
		}); err != nil {
			slog.Warn("logtail: RecordLogEvent passive_allocated error", "run_id", active.ID, "passive_id", ev.Passive.PassiveID, "err", err)
		}
	case progress.KindTradeAccepted:
		if err := svc.RecordLogEvent(ctx, active.ID, runpkg.EventTradeAccepted, map[string]string{
			"outcome": "accepted",
		}); err != nil {
			slog.Warn("logtail: RecordLogEvent trade_accepted error", "run_id", active.ID, "err", err)
		}
	}
	return active.ID
}
