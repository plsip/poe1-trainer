package main

import (
	"context"
	"fmt"
	"log"
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
	"github.com/poe1-trainer/internal/recommendation"
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

	h := api.NewHandlers(buildRepo, guideRepo, runService, runRepo, engine, gggProvider, gggClient)
	router := api.NewRouter(h)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
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

