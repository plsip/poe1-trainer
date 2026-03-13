# Prompt 10: Timer and Local Rankings

```text
Design the run timing and local speedrun ranking module for the Path of Exile 1 campaign trainer.

[Paste shared context here]

Requirements:
- manual run start,
- optional auto-start after the first valid gameplay event from the log,
- split points may be triggered by:
  - entering an area,
  - completing a guide step,
  - manual checkpoint,
- local ranking should compare:
  - full run time,
  - act times,
  - selected milestones,
- it should support PB, median, last run, and recent run history,
- the UI should show delta vs PB and delta vs previous run,
- canceled or incomplete runs should be stored distinctly from finished runs.

Deliver:
1. time and split data model,
2. split closing rules,
3. local ranking logic,
4. backend API contracts,
5. frontend view design,
6. edge case handling for reset, abandoned run, AFK pause policy, and missing log events.

Do not design global leaderboards.
Only local per-user or per-instance ranking.

Respond in Polish.
Use English for code, identifiers, API names, database schema names, and technical labels unless explicitly asked otherwise.
Keep user-facing explanations in Polish.
```