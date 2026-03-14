# Copilot Instructions For poe1-trainer

## Product direction

This repository is for a Path of Exile 1 training application, not a full autonomous agent.

The system should help the player learn the game over repeated runs.
It should suggest actions, explain why they matter, and highlight timing used by experienced players.
It should avoid taking full control of decision-making when an educational design would work better.

## Architecture rules

- Keep backend business logic in Go as the source of truth.
- Keep frontend state in React focused on structured application state, not raw markdown.
- Treat markdown guides as source content that must be parsed into explicit application models.
- Treat Client.txt parsing, trade data, and external APIs as unreliable inputs.
- Require manual confirmation for milestones that cannot be inferred safely.

## Product rules

- Prefer guidance over automation.
- Prefer explicit reasoning over opaque recommendations.
- Show why a suggestion matters at the current stage of progression.
- When in doubt, propose a smaller and more teachable workflow first.
- Do not design the system around full autonomy unless explicitly requested.

## Implementation rules

- Work iteratively.
- Define contracts before building broad features.
- Keep domain logic, integrations, persistence, and transport separated.
- At the current stage, do not create unit tests or fix existing tests unless the user explicitly asks for tests.
- Prefer runtime verification, logs, and manual checks over unit tests during early iteration.
- Use English for code, identifiers, API names, schema names, and technical labels.
- Use Polish for explanations and documentation unless asked otherwise.

## Early priorities

1. Guide ingestion and normalization.
2. Campaign progression model.
3. Run tracking and manual confirmations.
4. Actionable suggestions with explanations.
5. Safe local integrations.

## Avoid in the first iteration

- building a generic autonomous agent,
- relying on LLM prompts as the main control mechanism,
- mixing UI rendering, domain rules, and integration code,
- implementing full trade automation,
- attempting complete quest-state inference from unreliable signals.
