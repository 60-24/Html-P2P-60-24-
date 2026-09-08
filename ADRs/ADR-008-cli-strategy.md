# ADR-008: CLI Strategy

## Status
ACCEPTED

## Context
Po SESJI #4 system ma pełny stos: EventBus, GossipNode, ConsensusLayer,
NetworkSimulator. Ale nie istnieje żaden sposób na uruchomienie węzła jako
samodzielny proces produkcyjny ani interfejs operatorski do zarządzania nim
(start, status, eksport danych). Operator potrzebuje narzędzia wiersza poleceń.

## Decision
Implementujemy CLI w oparciu o wzorzec Cobra (subcommands + flags),
de facto standard w ekosystemie Go (używany przez `kubectl`, `docker`, `gh`).

**Zestaw komend:**
```
p2p run              Uruchom węzeł z podanym configiem
p2p config show       Wyświetl bieżącą konfigurację (po walidacji)
p2p metrics           Wypisz metryki Prometheus (lokalnie, bez serwera)
p2p peer list          Pokaż połączonych peerów
p2p event export      Eksportuj event log do pliku
p2p status            Health check (kod wyjścia 0/1 dla skryptów)
p2p version           Informacja o wersji
```

**Standard komunikatów błędów** — pełne, pomocne, z przykładami (nie unix
minimalizm w stylu pojedynczej linii):
```
Error: unknown flag: --invalid-flag

Usage:
  p2p run [flags]

Flags:
  -c, --config string   Path to config.yaml (required)
  -v, --verbose         Verbose logging
  -h, --help            Show this help

Examples:
  p2p run --config config.yaml
  p2p run -c config.yaml --verbose

See 'p2p help run' for more.
```

## Consequences

**Positive:**
- Znajomy wzorzec dla każdego operatora znającego `kubectl`/`docker`
- Automatyczne generowanie `--help` dla każdej komendy (Cobra built-in)
- `p2p status` z kodem wyjścia ułatwia integrację ze skryptami/healthcheckami
- Rozszerzalność: dodanie nowej komendy nie wymaga zmian w istniejących

**Negative:**
- Dodatkowa zależność (biblioteka Cobra) — akceptowalne, biblioteka stabilna
  i szeroko stosowana w produkcji

## Alternatives Considered
1. **flag (standard library)** — odrzucone: brak wsparcia dla subcommands
   bez ręcznego parsowania, brak auto-generowanego help.
2. **urfave/cli** — rozważane, ale Cobra ma szersze wsparcie w ekosystemie
   Kubernetes-adjacent (kubectl, helm używają Cobra) — spójność z ekosystemem
   docelowego środowiska deploymentu.
3. **Cobra (wybrane)** — standard branżowy, dobra dokumentacja, generuje
   czytelne komunikaty pomocy.

## References
- Code: `cmd/p2p/main.go`, `cmd/p2p/commands/*.go`
- Depends on: ConfigManager (ADR-007), MetricsCollector (ADR-006),
  PersistenceStore (ADR-005)
