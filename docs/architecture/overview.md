# Architektura startowa

## Główne moduły

### 1. Guide domain

Odpowiada za import markdown guide, normalizację kroków, checkpointów, zaleceń i warunków.

### 2. Run tracking domain

Odpowiada za stan konkretnego przejścia postaci, checkpointy, progres, potwierdzenia ręczne i historię runu.

### 3. Recommendation domain

Odpowiada za generowanie sugestii na podstawie etapu runu, wymagań buildu i wykrytych sygnałów.

### 4. Integration layer

Odpowiada za wejścia zewnętrzne:

- Client.txt log watcher,
- opcjonalne GGG API,
- później opcjonalne integracje z trade.

### 5. Frontend application

Odpowiada za prezentację aktualnego stanu runu, następnych kroków i uzasadnień rekomendacji.

## Zasady architektoniczne

- Backend w Go jest źródłem prawdy.
- Frontend nie implementuje logiki domenowej poza lekką logiką prezentacyjną.
- Integracje zewnętrzne nie mogą bezpośrednio zmieniać stanu bez walidacji backendowej.
- Każda ważna zmiana stanu powinna być jawna i testowalna.

## MVP data flow

1. Markdown guide jest importowany do struktury domenowej.
2. Użytkownik tworzy run dla konkretnego buildu.
3. Backend śledzi checkpointy i rekomendacje dla runu.
4. Frontend pokazuje aktualny etap, następne kroki i powody sugestii.
5. Jeśli sygnał z logów jest niejednoznaczny, użytkownik ręcznie potwierdza progres.

## Czego nie robić od razu

- pełnego systemu agentowego,
- automatycznego decydowania o wszystkim,
- szerokiej wielobuildowości bez stabilnego modelu guide,
- rozbudowanych integracji zanim będzie gotowy model progresu.
