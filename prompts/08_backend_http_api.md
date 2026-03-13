# Prompt 08: Backend HTTP API

```text
Based on the shared context below, design and implement the Go HTTP API for the Path of Exile 1 campaign trainer backend.

[Paste shared context here]

Design endpoints for:
- builds and guide versions,
- guide import,
- runs and run characters,
- current progress state,
- gem and gear alerts,
- log-driven events,
- optional GGG snapshot synchronization,
- split timing and local ranking,
- manual confirm / undo / skip actions for steps.

Requirements:
- explicit request/response DTOs,
- consistent error format,
- validation at the API boundary,
- backend remains the source of truth for business rules,
- endpoints should be designed around application state, not raw tables,
- long-running work should use a clear async/task model if needed.

Deliver:
1. endpoint list,
2. DTO contracts,
3. backend layer structure,
4. HTTP transport implementation,
5. tests for handlers and use cases.

Do not build the full frontend yet.

Respond in Polish.
Use English for code, identifiers, API names, database schema names, and technical labels unless explicitly asked otherwise.
Keep user-facing explanations in Polish.
```