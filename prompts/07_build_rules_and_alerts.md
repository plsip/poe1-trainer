# Prompt 07: Build Rules and Alert System

```text
Design the deterministic build rule system for a Path of Exile 1 campaign trainer.

[Paste shared context here]

I need rule support for:
- gem availability:
  - quest reward unlock,
  - vendor unlock,
  - fallback vendor later,
- setup transitions:
  - when to start leveling a gem,
  - when to place a gem on weapon swap,
  - when to do a full setup switch,
- gear alerts:
  - check boots for movement speed,
  - craft resistances,
  - look for 4-link,
  - weapon upgrade timing,
  - specific campaign checkpoints like Sapphire Ring before Merveil,
- conditional alerts based on act, step, level, route position, and available character state.

Deliver:
1. rules model,
2. storage strategy for rules (database seeds or versioned files),
3. evaluation strategy,
4. example rules for the Storm Burst Totems build,
5. backend response shape for the frontend,
6. tests for rule evaluation.

This is not an AI chat feature.
It should be a deterministic recommendation and alert system driven by application state.

Respond in Polish.
Use English for code, identifiers, API names, database schema names, and technical labels unless explicitly asked otherwise.
Keep user-facing explanations in Polish.
```