# Zasady redakcji poradnika kampanii

Ten dokument zbiera zasady wypracowane przy redakcji Aktu 1 dla `guides/stormburst_campaign_v1.md`.
Jego celem jest ujednolicenie dalszej pracy nad kolejnymi aktami tak, żeby poradnik był jednocześnie:

- czytelny dla gracza,
- możliwy do sparsowania przez backend,
- sensowny pod przyszłe dopasowywanie kroków do wejść do lokacji z logów.

## 1. Źródło prawdy

- Kanonicznym źródłem treści jest plik w `guides/`, a nie kopia w `docs/` ani seed JSON.
- Redagujemy bezpośrednio `guides/stormburst_campaign_v1.md`.
- Dokumenty w `docs/` mogą opisywać zasady i decyzje, ale nie powinny być traktowane jako runtime source of truth.

## 2. Twarde ograniczenia formatu

- Tylko numerowane punkty `1.`, `2.`, `3.` są importowane jako kroki guide.
- Nagłówki `## Akt N` wyznaczają akt.
- Sekcja `## Zasady ogólne` jest preambułą i nie należy do kroków konkretnego aktu.
- Jeśli informacja nie ma być osobnym krokiem w trainerze, nie zapisuj jej jako osobnego numerowanego punktu.
- Zachowuj obecny format kolorów i znaczników HTML, bo parser oraz heurystyki już na nim pracują.

## 3. Główna zasada redakcyjna

- Każdy punkt ma opisywać jedną realną akcję do wykonania.
- Punkt ma być proceduralny: gracz ma wiedzieć, co zrobić teraz, a nie co wiedzieć ogólnie.
- Jeśli krok istnieje w poradniku, powinien oznaczać, że da się go wykonać w tym momencie trasy.

Dobra forma:

- `Wejdź do lokacji.`
- `Zabij bossa i wróć do miasta portalem albo logoutem.`
- `Kup gem u konkretnego NPC i wrzuć go na weapon swap.`

Zła forma:

- `Po odblokowaniu kup X.`
- `Odbierz reward gemowy.`
- `Przygotuj się do późniejszego switchu.`
- `Jeśli chcesz / jeśli wolisz / jeśli może się przydać`, jeśli to nie jest naprawdę decyzja opcjonalna.

## 4. Kiedy łączyć kroki

- Łączymy kroki, gdy w praktyce są jedną ciągłą akcją.
- Nie rozbijamy sztucznie sekwencji tylko dlatego, że da się ją opisać na dwa zdania.

Łączyć warto zwłaszcza wtedy, gdy:

- kupujesz rzecz i od razu wykonujesz oczywiste następstwo,
- zabijasz bossa i naturalnym następnym ruchem jest powrót do miasta,
- wchodzisz do lokacji i od razu bierzesz waypoint przy wejściu,
- wchodzisz do pobocznej lokacji, bierzesz quest item i wracasz skrótem do głównej trasy.

Przykłady dobrych scaleń:

- `Kup Spell Totem Support u Nessy i wrzuć go na weapon swap do levelowania.`
- `Zabij Brutusa i wróć do miasta portalem albo logoutem.`
- `Wejdź do Ship Graveyard Cave, zabierz Allflame i wyjdź skrótem z powrotem do Ship Graveyard.`

## 5. Kiedy rozdzielać kroki

- Rozdzielamy kroki, gdy osobna akcja ma własną wartość dla nawigacji, trackingu albo decyzji gracza.
- Najczęstszy przypadek: wejście do nowej lokacji powinno być osobnym krokiem, bo to najlepiej nadaje się do przyszłego matchowania z logami.

Rozdzielaj osobno, gdy:

- wejście do lokacji jest osobnym kamieniem milowym,
- po wejściu do lokacji dopiero później wykonujesz inną ważną czynność,
- krok zmienia cel trasy albo wymaga powrotu do miasta,
- osobny krok ułatwia późniejsze mapowanie `area entered -> guide step`.

## 6. Reguły dla lokacji i trasy

- Krok wejścia do lokacji zapisujemy prostym czasownikiem: `Wejdź do ...`.
- Jeśli waypoint jest przy wejściu albo jest oczywistą częścią wejścia, można go dopisać w tym samym kroku.
- Jeśli powrót do miasta jest istotny dla trasy, trzeba napisać czym wracamy: `portalem`, `logoutem` albo `waypointem`.
- Nie twórz osobnych kroków typu `Złap waypoint`, jeśli to jest tylko drobna część bieżącego ruchu i nie wnosi osobnego znaczenia.
- Nie zostawiaj niejasnych sformułowań typu `wróć do wejścia`, jeśli nie wiadomo skąd i w jaki sposób.

Preferowane sformułowania:

- `Wejdź do Cavern of Wrath i złap waypoint.`
- `Wróć waypointem do Lioneye's Watch.`
- `Zabij Fairgravesa i wróć do miasta portalem albo logoutem.`

## 7. Reguły dla rewardów, vendorów i gemów

- Każdy reward odbierany w mieście powinien mieć podanego konkretnego NPC.
- Każdy zakup gema powinien mieć podane:
  - nazwę gema,
  - NPC,
  - co zrobić z gemem od razu, jeśli to oczywiste i ważne.
- Nie używamy ogólników typu `reward gemowy`, jeśli można podać konkretną nazwę albo przynajmniej NPC.
- Jeśli dokładna nazwa gema nie jest pewna, lepiej usunąć taki krok niż zostawić nieprecyzyjną informację.
- Nie piszemy `po odblokowaniu`, jeśli sam fakt istnienia kroku ma znaczyć, że odblokowanie już nastąpiło.

Preferowane sformułowania:

- `Odbierz Book of Skill od Bestela.`
- `Kup Shield Charge u Nessy i wrzuć go do tarczy.`
- `Wróć do Lioneye's Watch, kup u Nessy Storm Burst i wrzuć go na weapon swap do levelowania.`

## 8. Reguły dla warunków i opcjonalności

- Warunki zostawiamy tylko wtedy, gdy są naprawdę potrzebne i dotyczą realnej decyzji gracza.
- Jeśli coś jest częścią standardowej trasy builda, zapisujemy to bez warunku.
- Jeśli coś jest opcjonalne, zaznaczamy to jasno i krótko.

Dopuszczalne:

- `Jeśli nie masz drugiego Quicksilvera, zrób The Great White Beast.`
- `Jeśli masz już drugi quicksilver, pomiń ten objazd.`

Niepożądane:

- `Jeśli masz sensowną tarczę i wymagany level...`
- `Jeśli wolisz, możesz może kupić...` w kroku, który ma być częścią głównej ścieżki.

## 9. Czego usuwać z głównej listy kroków

Z głównej listy usuwamy albo scalamy rzeczy, które są zbyt ogólne, oczywiste albo nieczytelne jako osobny krok:

- luźne przypomnienia o gearze,
- ogólne porady typu `sprawdź vendor`,
- mikrokroki bez własnej wartości trasy,
- kroki zależne od niepewnego stanu questa,
- kroki, które nie mówią u kogo coś odebrać albo kupić,
- kroki opisane zbyt mgliście, np. `zrób reward`, `wróć do wejścia`, `po odblokowaniu kup`.

Takie rzeczy powinny trafić albo do `Zasad ogólnych`, albo zostać dopisane do sąsiedniego kroku, albo zostać usunięte.

## 10. Styl językowy

- Piszemy krótko, rozkazująco i proceduralnie.
- Jedno zdanie na krok zwykle wystarcza.
- Unikamy dygresji i tłumaczenia teorii builda w środku trasy.
- Nazwy techniczne, lokacje, NPC i gemy zostają po angielsku zgodnie z grą.
- Opis czynności i komentarze redakcyjne są po polsku.

## 11. Checklista redakcyjna dla każdego aktu

Przed uznaniem aktu za gotowy sprawdź:

- Czy każdy numerowany punkt jest realnym krokiem do wykonania teraz.
- Czy wejścia do ważnych lokacji są widoczne jako osobne kroki.
- Czy rewardy i zakupy mają wskazanego NPC.
- Czy nie ma ogólników typu `reward gemowy`, `po odblokowaniu`, `wróć do wejścia`.
- Czy waypoint, portal i logout są wpisane tam, gdzie wpływają na trasę.
- Czy nie ma osobnych kroków, które powinny być scalone z sąsiednim.
- Czy nie ma osobnych kroków, które są tylko luźną poradą gearową.
- Czy krok nie zawiera zbędnego warunku, jeśli jest częścią głównej ścieżki.
- Czy po przeczytaniu kroku gracz wie dokładnie: gdzie iść, co zrobić, do kogo wrócić.

## 12. Praktyczna zasada rozstrzygania sporów

Jeśli nie wiadomo, czy coś zostawić jako osobny krok, zadaj dwa pytania:

1. Czy gracz wykona to jako osobną decyzję albo osobny ruch na trasie?
2. Czy osobny krok pomoże w późniejszym dopasowaniu progresu do logów albo checkpointów?

Jeśli odpowiedź na oba pytania brzmi `nie`, taki krok zwykle należy scalić albo usunąć.