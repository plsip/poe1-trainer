# Architektura i plan repozytorium — `poe1-trainer`

---

## 1. Struktura repozytorium

```
poe1-trainer/
├── backend/                        # Go — źródło prawdy logiki domenowej
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   ├── internal/
│   │   ├── guide/                  # domain: import i normalizacja guide
│   │   │   ├── model.go
│   │   │   ├── parser.go
│   │   │   └── repository.go
│   │   ├── run/                    # domain: przebieg konkretnego runu
│   │   │   ├── model.go
│   │   │   ├── service.go
│   │   │   └── repository.go
│   │   ├── recommendation/         # domain: sugestie i uzasadnienia
│   │   │   ├── model.go
│   │   │   └── engine.go
│   │   ├── integration/
│   │   │   ├── logtail/            # Client.txt watcher
│   │   │   └── ggg/                # opcjonalne GGG OAuth API
│   │   ├── api/                    # HTTP handlers (transport layer)
│   │   │   ├── handlers.go
│   │   │   └── router.go
│   │   └── db/                     # migracje, połączenie, queries
│   │       ├── migrations/
│   │       └── store.go
│   ├── go.mod
│   └── go.sum
│
├── frontend/                       # React + TypeScript
│   ├── src/
│   │   ├── api/                    # klienty HTTP do backendu
│   │   ├── store/                  # stan aplikacji (np. Zustand / Redux)
│   │   ├── pages/
│   │   │   ├── RunPage/
│   │   │   ├── GuidePage/
│   │   │   └── HistoryPage/
│   │   └── components/
│   ├── package.json
│   └── tsconfig.json
│
├── docs/
│   ├── architecture/
│   ├── decisions/                  # ADRy
│   └── product/
│
├── guides/                         # surowe markdown guide, źródło treści
│   └── stormburst_campaign.md
│
├── prompts/                        # zestawy promptów dla Copilota
│
├── docker-compose.yml
├── docker-compose.dev.yml
└── README.md
```

---

## 2. Domeny i bounded contexts

| Domena | Odpowiedzialność | Granice |
|---|---|---|
| `guide` | Import markdown, normalizacja kroków, checkpointów i wymagań buildu | Nie zna stanu konkretnego runu |
| `run` | Stan jednego przejścia postaci, historia progresu, potwierdzenia manualne | Nie generuje sugestii, nie parsuje guide |
| `recommendation` | Generuje sugestie na podstawie aktualnego etapu runu + wymagań z guide | Nie zapisuje stanu, tylko czyta |
| `integration/logtail` | Odczytuje `Client.txt`, emituje zdarzenia obszarowe | Nie zmienia stanu — tylko dostarcza sygnały |
| `integration/ggg` | Opcjonalnie odczytuje dane z GGG API (OAuth) | Traktowane jako nieautorytatywny sygnał |
| `api` | Transport HTTP, routing, walidacja wejścia | Nie zawiera logiki domenowej |
| `db` | Persystencja, migracje | Bez logiki poza zapytaniami |

**Zasada graniczna:** integracje zewnętrzne (`logtail`, `ggg`) nie mogą bezpośrednio modyfikować stanu `run`. Dostarczają zdarzenia, które `run service` weryfikuje i przetwarza.

---

## 3. Przepływ danych

```
guides/*.md
    │
    ▼
guide.Parser
    │  parsuje markdown, wyodrębnia kroki, gemy, checkpointy,
    │  sprzęt, wymagania czasowe
    ▼
guide.Repository (PostgreSQL: guides, guide_steps, guide_checkpoints)
    │
    ▼ (przy starcie runu)
run.Service
    │  tworzy RunSession powiązaną z guide,
    │  śledzi aktywny krok i stan checkpointów,
    │  wymaga manualnego potwierdzenia dla celów, których nie można
    │  ustalić deterministycznie
    ▼
run.Repository (PostgreSQL: runs, run_events, run_checkpoints)
    │
    ▼
recommendation.Engine
    │  na podstawie aktywnego kroku + niespełnionych wymagań buildu
    │  produkuje listę Recommendation{text, reason, priority}
    │
    ├──◄── logtail.Watcher
    │        odczytuje Client.txt, emituje AreaEvent
    │        → run.Service.HandleAreaEvent()
    │
    └──◄── ggg.Client (opcjonalnie)
             OAuth token user → GET /character → częściowe dane postaci
             → run.Service.HandleExternalHint() (niższy priorytet niż manual)
    │
    ▼
api (HTTP/JSON) ←──────────────────────────────────► React frontend
    GET /runs/:id/state                              store (Zustand/Redux)
    GET /runs/:id/recommendations                    RunPage, GuidePage
    POST /runs/:id/checkpoints/:step/confirm         HistoryPage
    POST /runs                                       TimerComponent
    GET /guides
    ...
```

---

## 4. MVP vs Phase 2 vs Phase 3

### MVP (faza 1) — rdzeń uczący

- Import jednego markdown guide do modelu domenowego
- Model `Run` z checkpointami i manualnym potwierdzaniem
- `recommendation.Engine` z prostymi regułami (etap → lista sugestii)
- REST API: odczyt guide, tworzenie runu, potwierdzanie kroków
- Prosty frontend: aktualny krok, następna sugestia z uzasadnieniem, przycisk potwierdzenia
- Docker Compose z Go + PostgreSQL + React

### Phase 2 — integracje i czas

- `logtail.Watcher`: automatyczne wykrywanie przejść między obszarami z `Client.txt`
- Timer per run, split comparison, lokalny ranking (tabela `run_splits`)
- Podstawowe GGG OAuth API (opcjonalne, z jawnym fallbackiem manualnym)
- Historia runów z widokiem progressu

### Phase 3 — wielobuildowość i rozszerzenia

- Obsługa wielu buildów i wielu wersji guide
- Porównywanie runów między buildami
- Zaawansowane alerty buildowe (optymalizacja sprzętu, progi dpsowe)
- Eksport danych, statystyki nauki

---

## 5. Ryzyka techniczne i twarde ograniczenia PoE1

| Ryzyko | Opis | Mitygacja |
|---|---|---|
| **Client.txt ograniczony** | Logi zawierają tylko przejścia między obszarami, nie questy, nie zabijanego bossa | Wymagaj manualnych potwierdzeń dla kamieni milowych quest |
| **GGG API niekompletny** | Brak pełnej historii questów, dane inventory nie są w czasie rzeczywistym | Traktuj GGG API jako opcjonalny hint, nie jako źródło prawdy |
| **Ingerencja w klienta gry** | Absolutny zakaz — narusza TOS | Tylko `Client.txt` readonly + GGG OAuth; żadnego memory reading / injection |
| **Markdown guide niestrukturalny** | Każdy guide ma własny format, brak schematu | Zdefiniuj normalizowany model `GuideStep` i parser mapujący konkretny format |
| **Timing guide zależy od gracza** | Realne czasy dzielenia runu zależą od skillów, itemów, lagu | Traktuj timing jako informacyjny, nigdy nie blokuj progresu opóźnieniem |
| **Niejednoznaczność stanu** | Obszar może być odwiedzany wielokrotnie | `AreaEvent` + kontekst kroku guide musi być filtrowany przez engine, nie zapis raw eventu |
| **Race condition logtail** | Watcher może zgubić logi przy szybkich przejściach | Buforuj eventy, przetwarzaj seryjnie w `run.Service` |

---

## 6. Plan implementacji (8 kroków)

**Krok 1 — Inicjalizacja repozytorium i infrastruktury**
Struktura katalogów, `go.mod`, `package.json`, `docker-compose.yml`, pusta baza PostgreSQL, migracje schematowe (tabele `guides`, `guide_steps`, `runs`, `run_events`).

**Krok 2 — Model domenowy `guide` i parser markdown**
Definicja `GuideStep`, `GuideCheckpoint`, `GemRequirement`. Parser jednego konkretnego guide. Testy jednostkowe parsera. Zapis do bazy.

**Krok 3 — Model domenowy `run` i service**
`RunSession`, `RunCheckpoint`, `RunEvent`. `run.Service` z operacjami: `CreateRun`, `ConfirmStep`, `GetCurrentState`. Testy jednostkowe. Persystencja.

**Krok 4 — `recommendation.Engine`**
Reguły: dla danego `CurrentStep` + stanu checkpointów → lista `Recommendation{text, reason, priority}`. Pokryte testami. Bez zależności od warstwy HTTP.

**Krok 5 — REST API (transport layer)**
Endpointy: `POST /runs`, `GET /runs/:id/state`, `GET /runs/:id/recommendations`, `POST /runs/:id/steps/:step/confirm`, `GET /guides`. Walidacja wejścia na poziomie handlerów.

**Krok 6 — Podstawowy frontend**
`RunPage`: aktualny krok, lista rekomendacji z uzasadnieniami, przycisk potwierdzenia. `GuidePage`: podgląd całego guide z oznaczonym postępem. Komunikacja przez `api/` klientów HTTP.

**Krok 7 — `logtail.Watcher`**
Odczyt `Client.txt` tail jako goroutine. Emitowanie `AreaEvent` do kanału. `run.Service.HandleAreaEvent()` przetwarzający event z uwzględnieniem aktualnego kroku. Testowalne przez interfejs.

**Krok 8 — Timer i lokalny ranking**
Stoper per run w backendzie (start/stop event), tabela `run_splits`, endpoint `GET /runs/ranking`. Komponent timer w frontendzie. Porównanie splitów do poprzednich runów.

---

## 7. Czego nie robić w pierwszej iteracji

- **Nie budować generycznego agenta AI** — sugestie muszą pochodzić z deterministycznych reguł, nie z promptowania LLM w pętli.
- **Nie implementować pełnej wielobuildowości** — zaczynamy od jednego buildu i jednej wersji guide, zanim model będzie stabilny.
- **Nie podpinać GGG API w MVP** — zbyt dużo niepewności OAuth + niekompletne dane; manual confirmation jest wystarczający.
- **Nie automatyzować decyzji o questach** — nie da się tego zrobić deterministycznie bez ingerencji w klienta (naruszenie TOS).
- **Nie mieszać logiki domenowej z transportem ani UI** — żadna reguła dotycząca progresu nie trafia do handlerów ani komponentów React.
- **Nie budować systemu trade automation** — poza zakresem produktu i zakresem bezpiecznych integracji.
- **Nie tworzyć ogólnego importera markdown** — zanim stworzymy abstrakcję, zaimportujemy jeden konkretny guide i wyciągniemy wspólny model po zobaczeniu rzeczywistych danych.
