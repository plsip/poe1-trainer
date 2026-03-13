# Shared Context

Use this shared context before any task-specific prompt.

```text
I am building a new web application that acts as a Path of Exile 1 campaign trainer for a specific build.

Tech stack:
- frontend: React + TypeScript
- backend: Go
- database: PostgreSQL
- local environment: Docker Compose

Product goals:
- the user selects a build and tracks one or more campaign runs for one or more characters,
- the build guide already exists as Markdown and should be converted into a structured application model,
- the application should help the user by showing:
  1. when a required gem can already be received or purchased,
  2. when gear or gem setup should be updated,
  3. how far the current run progressed through the guide,
  4. local timing, split comparison, and a local speedrun ranking,
- progress tracking should be hybrid:
  - the markdown guide is the source of guide content,
  - local PoE1 Client.txt log reading is used to detect area transitions and other safe-to-read signals,
  - official GGG OAuth API may be used for partial character state,
  - manual confirmation is required for milestones that cannot be inferred reliably.

Important technical constraints:
- nothing may interfere with the game client,
- reading the client log is allowed only as an external application and with user awareness,
- official GGG API does not provide full quest progression,
- the Go backend must be the source of truth for business logic, state transitions, and integrations,
- the React frontend must represent application state rather than raw guide text,
- the system should be extensible to multiple builds, multiple guide versions, and multiple runs.

Architecture rules:
- keep business logic, transport, integrations, and persistence separated,
- treat log parsing and API integrations as unreliable external inputs,
- all important state changes should be explicit and testable,
- if something cannot be tracked safely or deterministically, propose a manual fallback,
- prefer deterministic backend rules over prompt-only behavior.

Implementation style rules:
- work iteratively,
- do not implement the whole application in one step,
- propose the architecture and contracts first when appropriate,
- use English for code, identifiers, schema names, endpoint names, and technical artifacts,
- respond in Polish.
```
