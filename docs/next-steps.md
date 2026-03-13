# Następne kroki

## Co powinno wejść zaraz po inicjalnym commicie

1. Ustalenie struktury repo dla `frontend`, `backend`, `docker` i `docs`.
2. Implementacja modelu guide w backendzie.
3. Import jednego markdown guide do struktury domenowej.
4. Implementacja modelu runu i checkpointów.
5. Przygotowanie prostego HTTP API do odczytu guide i stanu runu.
6. Przygotowanie prostego UI z widokiem aktualnego etapu runu.

## Jak pracować z agentem Copilota

- Zaczynaj od małych kroków.
- Wklej najpierw kontekst z istniejącego pakietu promptów w `tmp/poe_trainer_prompts`.
- Potem prowadź implementację etapami, nie całościowo.
- Pilnuj, żeby agent nie przeskakiwał od razu do pełnej agentowości produktu.

## Sugestia dla pierwszych promptów

Najpierw użyj:

1. `00_shared_context.md`
2. `01_architecture_and_plan.md`
3. `02_domain_model_and_database.md`
4. `03_markdown_guide_importer.md`
5. `08_backend_http_api.md`
6. `09_frontend_react_ui.md`
