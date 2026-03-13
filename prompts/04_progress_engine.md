# Prompt 04: Progress Engine

```text
Design the backend progress engine for a Path of Exile 1 campaign trainer.

[Paste shared context here]

The engine must combine three sources of truth:
- structured guide steps from the database,
- domain events inferred from the local PoE1 Client.txt log,
- optional character snapshots from the official GGG API.

I want a model that:
- transforms raw inputs into domain events,
- maps domain events to potentially completed guide steps,
- handles uncertainty with statuses such as:
  - confirmed,
  - inferred,
  - requires_manual_confirmation,
  - skipped,
- can determine the next relevant step,
- can generate alerts such as:
  - you can now buy a gem,
  - you can now receive a gem,
  - you should already level a specific gem,
  - check gear now,
- keeps a clear audit trail of why a step was marked.

Deliver:
1. event-driven model,
2. matching rules between events and guide steps,
3. confidence/scoring or evidence model,
4. internal backend contracts,
5. state model for step progression,
6. testable Go core logic.

Do not implement the file watcher or OAuth flow yet.
First implement the core engine and its contracts.

Respond in Polish.
Use English for code, identifiers, API names, database schema names, and technical labels unless explicitly asked otherwise.
Keep user-facing explanations in Polish.
```