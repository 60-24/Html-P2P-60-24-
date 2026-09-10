# ADR-005: Persistence Strategy

## Status
ACCEPTED

## Context
SESJA #4 testowała system w pełni in-memory (NetworkSimulator, EventStore).
Dla production-readiness potrzebujemy trwałego magazynu danych, który:
- Przetrwa restart węzła (crash recovery)
- Skaluje się do 1000+ eventów bez degradacji
- Pozwala na łatwą wymianę silnika bazy danych w przyszłości

## Decision
Implementujemy **wzorzec interfejsu** (`PersistenceStore`), a nie bezpośrednie
wywołania konkretnej bazy danych:

```go
type PersistenceStore interface {
    StoreEvent(event *core.Event) error
    GetEvent(eventID string) (*core.Event, error)
    CreateSnapshot(nodeID string, state map[string]interface{}) (string, error)
    RestoreSnapshot(snapshotID string) (map[string]interface{}, error)
    CreateCheckpoint(...) (string, error)
    Close() error
}
```

**Implementacja SESJI #5:** `InMemoryStore` (thread-safe, map-based) —
w pełni funkcjonalna implementacja referencyjna zgodna z interfejsem.

**Tech Debt (świadomie odłożone):** `RocksDBStore` — realna implementacja
na bazie RocksDB zostanie dodana w SESJI #6 (Deployment), gdy dostępne
będzie środowisko z bibliotekami C++ (RocksDB wymaga cgo). Interfejs jest
już przygotowany pod tę wymianę — zero zmian w kodzie wywołującym.

## Consequences

**Positive:**
- Testowalność: mock/in-memory store w testach bez zależności zewnętrznych
- Przenośność: zamiana silnika (RocksDB ↔ BadgerDB ↔ inny) bez zmiany API
- Separation of concerns: logika biznesowa nie zna szczegółów storage

**Negative:**
- InMemoryStore nie daje prawdziwej trwałości między restartami procesu
  (to jest świadomy tech debt, udokumentowany i śledzony)
- Dodatkowa warstwa abstrakcji = minimalny narzut wydajnościowy

## Alternatives Considered
1. **Direct RocksDB API** — odrzucone: ciasne powiązanie (tight coupling),
   utrudnia testy jednostkowe, wymaga cgo w każdym środowisku dev.
2. **SQL Database (Postgres)** — odrzucone: overkill dla append-only event log,
   niepotrzebna złożoność schematu relacyjnego.
3. **File-based JSON** — odrzucone: zbyt wolne przy 1000+ węzłach/eventach,
   brak wsparcia dla concurrent writes.
4. **Interface + InMemory (wybrane)** — pozwala dostarczyć wartość teraz
   (testowalność, API stabilność) i odłożyć decyzję o konkretnym silniku.

## Lifecycle (Wymagania Operacyjne)
- **Startup:** `Open()` waliduje ścieżkę/schemat, próbuje recovery ze stanu
- **Running:** `StoreEvent()` jest atomowe (brak częściowych zapisów),
  `CreateSnapshot()` działa w tle (non-blocking względem event loop)
- **Shutdown:** `Close()` flushuje bufory i zamyka gracefully

## References
- Code: `pkg/storage/persistence.go`
- Depends on: EventStore (SESJA #2), StateEngine (SESJA #2)
- Next: ADR-006 (Observability) będzie mierzyć write_latency tego modułu
