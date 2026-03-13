# Prompt 02: Domain Model and Database Schema

```text
Based on the shared context below, design the domain model and PostgreSQL schema for a React + Go + PostgreSQL Path of Exile 1 campaign trainer.

[Paste shared context here]

Design entities and relations for at least:
- build
- build_version
- guide
- guide_step
- guide_step_condition
- gem
- gem_availability_rule
- gem_upgrade_rule
- gear_hint_rule
- run
- run_character
- run_step_progress
- run_event
- run_split
- character_snapshot
- manual_check
- local_ranking

Requirements:
- one build can have multiple guide versions,
- one guide consists of acts, sections, steps, and milestones,
- a step must support:
  - manual completion,
  - inferred completion from log events,
  - inferred completion from API data,
  - status that requires manual confirmation,
- the schema must support gem alerts, gear hints, and timing splits,
- the schema must support multiple historical runs per build and per character.

Deliver:
1. entity descriptions,
2. relations,
3. proposed columns and types,
4. indexes,
5. integrity constraints,
6. a v1 SQL migration proposal,
7. notes about which columns are sources of truth and which are cached/derived.

Do not implement HTTP endpoints yet.
First produce a solid domain and data model.

Respond in Polish.
Use English for code, identifiers, API names, database schema names, and technical labels unless explicitly asked otherwise.
Keep user-facing explanations in Polish.
```