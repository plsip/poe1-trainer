# Prompt 05: PoE1 Client Log Watcher

```text
Design and implement a Go module for monitoring the local Path of Exile 1 Client.txt log file.

[Paste shared context here]

Assumptions:
- the application runs locally,
- it only reads the game client log,
- it primarily detects area transitions and other stable signals that can be safely parsed from the log,
- it publishes normalized events into the progress engine,
- it must be safe, simple, and resilient to restarts.

Requirements:
- configurable log file path,
- support for default Windows path configuration,
- file tailing behavior,
- checkpointing or offset persistence,
- parser for log lines with tests,
- explicit statuses such as:
  - waiting_for_file,
  - waiting_for_new_lines,
  - game_not_running,
  - parser_error,
  - active.

Deliver:
1. watcher design,
2. interfaces and contracts,
3. watcher implementation,
4. log parser,
5. mapping from raw lines to domain events,
6. tests and sample inputs.

Do not connect it to the frontend yet.

Respond in Polish.
Use English for code, identifiers, API names, database schema names, and technical labels unless explicitly asked otherwise.
Keep user-facing explanations in Polish.
```