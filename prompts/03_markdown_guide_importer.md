# Prompt 03: Markdown Guide Importer

```text
I already have a Path of Exile 1 campaign guide in Markdown. I want to convert it into a structured application model for a campaign trainer.

[Paste shared context here]

Design and implement a Go backend module that:
- imports a markdown guide,
- parses it into explicit structured data,
- splits the guide into acts, sections, steps, and milestones,
- enriches steps with metadata such as:
  - step type,
  - area name,
  - quest name,
  - related gems,
  - related gear alerts,
  - possible log-based detection,
  - possible API-based detection,
  - whether manual confirmation is required.

Requirements:
- the parser should be resilient to small formatting changes,
- markdown should not remain the only runtime source of truth after import,
- imported data should be stored in the application model and database,
- propose a clear intermediate format such as JSON schema or Go structs,
- the parser must have tests,
- step typing heuristics must be explicit and reviewable.

Deliver:
1. parser design,
2. intermediate data format,
3. mapping heuristics,
4. Go implementation,
5. tests,
6. one example import flow for a single guide.

Do not build frontend UI yet.

Respond in Polish.
Use English for code, identifiers, API names, database schema names, and technical labels unless explicitly asked otherwise.
Keep user-facing explanations in Polish.
```