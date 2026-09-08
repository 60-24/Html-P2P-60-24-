# ADR-007: Configuration Management

## Status
ACCEPTED

## Context
Do tej pory wszystkie parametry (port 6024, timeout 5s, próg Byzantine 0.66)
były zaszyte w kodzie (hard-coded) jako stałe lub argumenty konstruktorów.
To uniemożliwia operatorowi zmianę zachowania węzła bez rekompilacji —
niedopuszczalne dla produkcyjnego deploymentu (różne środowiska: dev/staging/prod).

## Decision
Implementujemy `ConfigManager` oparty o pliki YAML z możliwością nadpisania
przez zmienne środowiskowe (wzorzec 12-factor app).

**Struktura konfiguracji:**
```yaml
network:
  port: 6024
  peers: []
persistence:
  engine: inmemory      # "inmemory" | "rocksdb" (patrz ADR-005)
  path: /var/p2p/data
monitoring:
  prometheus_port: 9090
consensus:
  timeout_sec: 5
  byzantine_threshold: 0.66
```

**Nadpisania środowiskowe (dla kontenerów):**
```
P2P_NETWORK_PORT=6025
P2P_PERSISTENCE_PATH=/custom/path
```

**Walidacja (wymagana przed użyciem):**
- Pola wymagane: `network.port`, `persistence.engine`
- Sprawdzanie typów: int, string, duration
- Walidacja zakresów: `port` ∈ [1024, 65535], `byzantine_threshold` ∈ [0, 1]

## Consequences

**Positive:**
- Czytelność dla człowieka (operator może edytować bez znajomości Go)
- Natywne dla Kubernetes (ConfigMap montowany jako plik YAML)
- Wsparcie dla komentarzy w pliku (dokumentacja "in-place")
- Env override pozwala na Docker/K8s bez przebudowy obrazu

**Negative:**
- Dodatkowa zależność (biblioteka YAML parsing)
- Błędy składni YAML bywają nieintuicyjne (mitygacja: jasne komunikaty walidacji)

## Alternatives Considered
1. **TOML** — odrzucone: mniej rozpowszechnione w ekosystemie Kubernetes/Docker
   niż YAML, mimo że czytelniejsze składniowo.
2. **HCL (HashiCorp)** — odrzucone: dodatkowa krzywa uczenia się, kojarzone
   głównie z Terraform, nie jest standardem dla App config w Go.
3. **JSON** — odrzucone: brak komentarzy, mniej czytelne dla człowieka.
4. **YAML + ENV override (wybrane)** — najlepszy kompromis czytelność/ekosystem.

## References
- Code: `pkg/config/config.go`, `config/default.yaml`
- Consumed by: End-to-End Application (`cmd/p2p/main.go`), CLI Tool (ADR-008)
