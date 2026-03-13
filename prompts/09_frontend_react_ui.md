# Prompt 09: Frontend React UI

```text
Design and implement the React + TypeScript frontend for the Path of Exile 1 campaign trainer.

[Paste shared context here]

The frontend must represent structured application state. It should not be just a markdown viewer.

MVP views:
- build and guide selection screen,
- active run dashboard,
- guide step list with statuses:
  - confirmed,
  - inferred,
  - requires manual confirmation,
  - skipped,
- alert panel:
  - gems that can now be bought or received,
  - gems that should be leveled,
  - gear warnings,
- timing panel:
  - current run timer,
  - split times per stage,
  - comparison to previous runs and PB,
- integration status panel:
  - log watcher status,
  - GGG API status,
  - last synchronization or event time.

Requirements:
- small, composable components,
- separate client state, API transport, and presentational components,
- clear loading, partial-data, retry, and error states,
- support manual step confirmation and correction,
- support filtering by act, status, and step type,
- use a clean modern UI, not generic boilerplate.

Deliver:
1. frontend architecture,
2. view model design,
3. component structure,
4. API integration layer,
5. MVP implementation.

Do not build account management.

Respond in Polish.
Use English for code, identifiers, API names, database schema names, and technical labels unless explicitly asked otherwise.
Keep user-facing explanations in Polish.
```