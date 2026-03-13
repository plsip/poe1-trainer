# Prompt 06: Official GGG API Integration

```text
Design an optional Go backend integration with the official Path of Exile API for partial character state retrieval.

[Paste shared context here]

Important facts:
- GGG API is an official HTTP/JSON API secured by OAuth 2.1,
- a web app with backend should be designed around a confidential client,
- the application must still work without OAuth because API access may be unavailable,
- the API does not provide full quest progression,
- the API can still provide useful character data such as character info, passives, bandit choice, pantheon, equipment, and inventory.

I want you to:
1. design the integration adapter,
2. identify the minimum useful endpoints and scopes for MVP,
3. explain which data materially helps this trainer application,
4. propose a character snapshot model in the database,
5. design graceful degradation when OAuth is not configured,
6. provide internal contracts and implementation stubs,
7. clearly separate provider-specific details from application logic.

Do not implement the full OAuth registration process with GGG.
Focus on integration architecture and application contracts.

Respond in Polish.
Use English for code, identifiers, API names, database schema names, and technical labels unless explicitly asked otherwise.
Keep user-facing explanations in Polish.
```