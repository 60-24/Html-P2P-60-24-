# P2P 60-24 OneClick Evo — Architecture (Post-SESJA #5)

Status snapshot after SESJA #5 (Production Readiness). For full history
see individual SESSION_N_README.md files and ADRs/.

## Layer Stack (Complete)

```
┌──────────────────────────────────────────────┐
│ CLI / Operator Layer                          │  cmd/p2p (ADR-008)
├──────────────────────────────────────────────┤
│ Configuration Layer                           │  pkg/config (ADR-007)
├──────────────────────────────────────────────┤
│ Observability Layer                           │  pkg/monitoring (ADR-006)
├──────────────────────────────────────────────┤
│ Persistence Layer                             │  pkg/storage (ADR-005)
├──────────────────────────────────────────────┤
│ Testing / Simulation Layer                    │  pkg/testing (SESJA #4)
├──────────────────────────────────────────────┤
│ Consensus Layer (Byzantine BFT, Resonance)    │  pkg/consensus (SESJA #3)
├──────────────────────────────────────────────┤
│ Network Layer (Gossip, RealBond DAG)          │  pkg/network (SESJA #3)
├──────────────────────────────────────────────┤
│ TERRA OS Core (EventBus, EventStore,          │  pkg/terra, pkg/data
│ StateEngine)                                  │  (SESJA #2)
├──────────────────────────────────────────────┤
│ Security (Ed25519 KeyManager)                 │  pkg/sec (SESJA #2)
├──────────────────────────────────────────────┤
│ Core Primitives (Event)                       │  pkg/core (SESJA #2)
└──────────────────────────────────────────────┘
```

## What's New in SESJA #5

| Module | Package | ADR | LOC | Tests |
|---|---|---|---|---|
| PersistenceStore | pkg/storage | ADR-005 | ~340 | 9 |
| MetricsCollector | pkg/monitoring | ADR-006 | ~230 | 7 |
| ConfigManager | pkg/config | ADR-007 | ~185 | 9 |
| CLI Tool | cmd/p2p | ADR-008 | ~250 | 8 |
| **Total SESJA #5** | | | **~1005** | **33** |

## Cumulative Project Status

| Session | Focus | LOC | Tests |
|---|---|---|---|
| SESJA #2 | Event System (TERRA OS engines 1-4) | 750 | 24 |
| SESJA #3 | Network + Consensus (engines 5-6) | 560 | 24 |
| SESJA #4 | Integration Testing & Stress | 590 | 26 |
| SESJA #5 | Production Readiness | ~1005 | 33 |
| **Total** | | **~2905** | **107** |

## Data Flow (End-to-End, Post-SESJA #5)

```
p2p run --config config.yaml
    │
    ▼
config.LoadFromFile()          [ADR-007: parse YAML, apply ENV, validate]
    │
    ▼
storage.NewInMemoryStore()     [ADR-005: init persistence]
    │
    ▼
monitoring.NewMetricsCollector()  [ADR-006: start /metrics HTTP server]
    │
    ▼
(future: wire GossipNode + ConsensusLayer + DAGManager to
 collector.RecordLatency/IncrementCounter calls, and to
 store.StoreEvent/CreateSnapshot on every finalized event)
```

**Integration note (tracked follow-up):** `pkg/network`, `pkg/consensus`,
and `pkg/data` do not yet call into `pkg/monitoring` or `pkg/storage`
directly — SESJA #5 delivered these as independently-tested, ADR-approved
modules with a documented integration pattern (see
`pkg/monitoring/integration_test.go`). Wiring them into the live event
path is the first task of SESJA #6.

## Known Tech Debt (Explicitly Tracked)

1. **RocksDBStore** (ADR-005): `InMemoryStore` is the only implementation.
   RocksDB backend deferred to SESJA #6 (requires cgo).
2. **Live metrics wiring**: `GossipNode`/`ConsensusLayer` don't yet call
   `MetricsCollector` methods at runtime (see integration note above).
3. **Live persistence wiring**: `ConsensusLayer.finalizeEvent()` doesn't
   yet call `PersistenceStore.StoreEvent()` — currently only
   `EventStore` (in-memory, SESJA #2) is updated on finality.
4. **CLI `run` command** blocks with `select {}` — no graceful shutdown
   (SIGINT/SIGTERM handling) yet.

## ADR Index

- [ADR-005: Persistence Strategy](../ADRs/ADR-005-persistence-strategy.md)
- [ADR-006: Observability Strategy](../ADRs/ADR-006-observability.md)
- [ADR-007: Configuration Management](../ADRs/ADR-007-configuration.md)
- [ADR-008: CLI Strategy](../ADRs/ADR-008-cli-strategy.md)
