# Prompty Do Trenera Kampanii PoE1

Ten katalog zawiera prompty do nowego repozytorium dla aplikacji webowej, która ma działać jako trener kampanii Path of Exile 1 dla konkretnego buildu.

Zalecana kolejność użycia:

1. `00_shared_context.md`
2. `01_architecture_and_plan.md`
3. `02_domain_model_and_database.md`
4. `03_markdown_guide_importer.md`
5. `04_progress_engine.md`
6. `08_backend_http_api.md`
7. `09_frontend_react_ui.md`
8. `10_timer_and_local_rankings.md`
9. `11_docker_and_local_dev.md`
10. `05_client_log_watcher.md`
11. `06_ggg_api_integration.md`
12. `07_build_rules_and_alerts.md`

Jak tego używać:

- Na start wklej zawartość `00_shared_context.md` do rozmowy z agentem.
- Potem wklejaj po jednym pliku promptu na raz.
- Same prompty zostaw po angielsku.
- Każ agentowi odpowiadać po polsku, ale zostawić kod, identyfikatory, nazwy schematów bazy i etykiety API po angielsku.
- Nie proś agenta o implementację całej aplikacji w jednym kroku.

Zalecana reguła odpowiedzi, którą warto zostawiać na końcu promptu:

```text
Respond in Polish.
Use English for code, identifiers, API names, database schema names, and technical labels unless explicitly asked otherwise.
Keep user-facing explanations in Polish.
Work in small, verifiable steps.
```
