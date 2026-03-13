#!/usr/bin/env bash
# scripts/seed-dev.sh
#
# Importuje przewodnik Stormburst Totemy do lokalnej bazy danych.
# Uruchamia się poza Dockerem — wymaga działającej bazy (docker compose dev up -d db).
#
# Użycie:
#   bash scripts/seed-dev.sh
#   DATABASE_URL=postgres://... bash scripts/seed-dev.sh

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

DB_URL="${DATABASE_URL:-postgres://poe:poe@localhost:5432/poetrainer?sslmode=disable}"
GUIDE_FILE="$REPO_ROOT/guides/stormburst_campaign_v1.md"

if [[ ! -f "$GUIDE_FILE" ]]; then
  echo "Błąd: nie znaleziono pliku przewodnika: $GUIDE_FILE" >&2
  exit 1
fi

echo "Importowanie przewodnika: $GUIDE_FILE"
cd "$REPO_ROOT/backend"

go run ./cmd/import \
  -db "$DB_URL" \
  -file "$GUIDE_FILE" \
  -slug stormburst_campaign_v1 \
  -title "Stormburst Totemy — Kampania" \
  -build "Storm Burst Totemy"

echo "Gotowe. Przewodnik zaimportowany do bazy danych."
