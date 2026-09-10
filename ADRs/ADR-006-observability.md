# ADR-006: Observability Strategy

## Status
ACCEPTED

## Context
SESJA #4 dostarczyła `PerformanceBenchmark` do testów ad-hoc, ale nie mamy
ciągłego, produkcyjnego sposobu obserwacji działającego węzła: ile peerów
jest połączonych, jak szybko propagują się eventy, ile trwają rundy
głosowania konsensusu. Bez tego nie da się debugować systemu na produkcji.

## Decision
Implementujemy `MetricsCollector` zgodny z modelem Prometheus (pull-based,
counter/gauge/histogram), eksponowany przez endpoint HTTP `/metrics`.

**Metryki do zbierania (wg kategorii):**
- **Network:** `peer_count` (gauge), `event_broadcast_latency_ms` (histogram)
- **Consensus:** `voting_rounds_total` (counter), `finalized_events_total` (counter)
- **Persistence:** `write_latency_ms` (histogram), `compaction_duration_ms` (histogram)
- **Memory:** `heap_size_bytes` (gauge), `event_log_size` (gauge)

**Integration point (wzorzec użycia w każdym module):**
```go
collector.RecordLatency("event_broadcast", duration)
collector.IncrementCounter("consensus_rounds")
collector.GaugeMemory("event_log_size", bytes)
```

## Consequences

**Positive:**
- Pull model: Prometheus scrapuje `/metrics`, węzeł nie musi znać adresu
  systemu monitoringu (prostsze niż push)
- Natywne wsparcie w Kubernetes (ServiceMonitor, scrape annotations)
- Integracja z Grafana "za darmo" (gotowe dashboardy dla formatu Prometheus)

**Negative:**
- Wymaga dodatkowego portu HTTP (domyślnie 9090) — trzeba udokumentować
  w firewall/network policy przy deploymencie
- Pull model nie nadaje się do zdarzeń rzadkich/natychmiastowych (alerting
  opóźniony o interwał scrape) — akceptowalne dla naszego przypadku użycia

## Alternatives Considered
1. **InfluxDB (push model)** — odrzucone: wymaga że węzeł zna adres bazy,
   trudniejsze w środowiskach z NAT/firewall między węzłami P2P.
2. **Datadog / komercyjny SaaS** — odrzucone: koszt, zależność od
   zewnętrznego dostawcy dla infrastruktury decentralizowanej (sprzeczność
   filozoficzna z projektem P2P 60-24).
3. **Prometheus (wybrane)** — open-source, standard de facto dla Go +
   Kubernetes, zero kosztów licencyjnych, spójne z etosem projektu.

## References
- Code: `pkg/monitoring/metrics.go`
- Depends on: PersistenceStore (ADR-005) dla write_latency
- Consumed by: CLI Tool (ADR-008) — `p2p metrics` command
