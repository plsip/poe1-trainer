package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/poe1-trainer/internal/db"
	"github.com/poe1-trainer/internal/guide"
)

func main() {
	var (
		dsn       = flag.String("db", envOrDefault("DATABASE_URL", "postgres://poe:poe@localhost:5432/poetrainer?sslmode=disable"), "PostgreSQL DSN")
		file      = flag.String("file", "", "Path to markdown guide file (required)")
		slug      = flag.String("slug", "", "Guide slug, e.g. stormburst_campaign_v1 (required)")
		title     = flag.String("title", "", "Guide title (required)")
		buildName = flag.String("build", "", "Build name, e.g. Storm Burst Totemy (required)")
		version   = flag.String("version", "", "Guide version; defaults to current git commit hash or GUIDE_VERSION")
	)
	flag.Parse()

	if *file == "" || *slug == "" || *title == "" || *buildName == "" {
		flag.Usage()
		os.Exit(1)
	}

	data, err := os.ReadFile(*file) // #nosec G304 — CLI tool, path from trusted user input
	if err != nil {
		log.Fatalf("read file: %v", err)
	}

	resolvedVersion, err := guide.ResolveVersion(*version)
	if err != nil {
		log.Fatalf("resolve version: %v", err)
	}

	g, err := guide.ParseMarkdown(*slug, *title, *buildName, resolvedVersion, string(data))
	if err != nil {
		log.Fatalf("parse: %v", err)
	}
	fmt.Printf("parsed %d steps\n", len(g.Steps))

	ctx := context.Background()
	store, err := db.New(ctx, *dsn)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer store.Close()

	repo := guide.NewRepository(store.Pool)
	if err := repo.Save(ctx, g); err != nil {
		log.Fatalf("save: %v", err)
	}
	fmt.Printf("guide %q saved with ID %d\n", g.Slug, g.ID)
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
