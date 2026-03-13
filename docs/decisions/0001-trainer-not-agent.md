# ADR 0001: Trainer first, not agent first

## Status

Accepted

## Context

Produkt ma pomagać użytkownikowi uczyć się gry i podejmować coraz lepsze decyzje z każdym kolejnym runem.

Początkowo rozważany był kierunek bardziej agentowy, ale istnieje ryzyko, że użytkownik zacznie polegać na systemie bez budowania własnego rozumienia gry.

## Decision

Na tym etapie produkt jest projektowany jako trainer.

System ma:

- sugerować,
- ostrzegać,
- wskazywać dobry moment na decyzję,
- tłumaczyć znaczenie rekomendacji,
- budować wiedzę użytkownika.

System nie ma w pierwszej iteracji:

- przejmować pełnej decyzyjności,
- działać jako autonomiczny agent,
- optymalizować wszystkiego bez wyjaśnienia,
- ukrywać logiki wyłącznie w promptach.

## Consequences

- UX powinien premiować zrozumienie, nie delegowanie.
- Rekomendacje powinny zawierać krótkie uzasadnienia.
- System może kiedyś zyskać warstwę agentową, ale nie powinna ona definiować MVP.
