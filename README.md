# poe1-trainer

PoE1 Trainer is a web application designed to help players learn Path of Exile 1 through deliberate campaign progression and later character advancement.

The system should suggest decisions, highlight important moments, and explain why recommendations matter, but it should not replace the player's own thinking.

## Product Goal

- help the player understand what to do next and why,
- teach patterns used by experienced players,
- track run progress against the guide,
- highlight when to update setup, gems, gear, and flasks,
- stay extensible toward maps, atlas, and endgame features.

## Technical Assumptions

- frontend: React + TypeScript
- backend: Go
- database: PostgreSQL
- local environment: Docker Compose

## Product Principles

- the trainer suggests, but does not hand-hold without explanation,
- the backend is the source of truth for business logic and integrations,
- the markdown guide is the source of build content,
- Client.txt parsing and other external integrations are treated as unreliable inputs,
- when something cannot be detected safely, the system requires manual confirmation.

## MVP

- import one markdown guide for one build,
- model campaign steps and checkpoints,
- manual or semi-automatic run progress tracking,
- suggestions for gems, gear, flasks, and key decisions,
- a simple view of the current stage and next recommendations.

## Out Of Scope For MVP

- full endgame support,
- advanced trade purchase recommendations,
- broad multi-build support,
- agentic decision workflows.

## Starter Documents

- [docs/product/vision.md](docs/product/vision.md)
- [docs/architecture/overview.md](docs/architecture/overview.md)
- [docs/decisions/0001-trainer-not-agent.md](docs/decisions/0001-trainer-not-agent.md)
- [docs/next-steps.md](docs/next-steps.md)

## Working Prompts

The `prompts/` directory contains the working prompt pack for continuing development with Copilot in the new repository.

Recommended starting order:

1. `prompts/00_shared_context.md`
2. `prompts/01_architecture_and_plan.md`
3. `prompts/02_domain_model_and_database.md`
4. `prompts/03_markdown_guide_importer.md`
5. `prompts/08_backend_http_api.md`
6. `prompts/09_frontend_react_ui.md`
