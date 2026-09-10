# P2P 60-24 CLI Guide

The `p2p` command-line tool is the operator interface for running and
managing a P2P 60-24 OneClick Evo node. See [ADR-008](../ADRs/ADR-008-cli-strategy.md)
for the design rationale.

## Installation

```bash
go build -o p2p ./cmd/p2p
```

Or via Docker:

```bash
docker build -t p2p-60-24 .
docker run -p 6024:6024/udp -p 9090:9090 p2p-60-24
```

## Quick Start

```bash
# Run a node with the default config
p2p run --config config/default.yaml

# Check config is valid without starting the node
p2p status --config config/default.yaml

# See the fully-resolved (env-overridden) configuration
p2p config show --config config/default.yaml
```

## Commands

### `p2p run`

Starts a node: loads config, initializes the persistence store (ADR-005),
starts the Prometheus metrics HTTP server (ADR-006), and prints connection
info.

```bash
p2p run --config config.yaml
p2p run -c config.yaml --verbose
```

### `p2p config show`

Displays the fully-resolved configuration (file + environment overrides),
after validation. Useful for confirming what a container will actually run
with.

```bash
p2p config show --config config.yaml
```

### `p2p metrics`

Prints a local Prometheus-format metrics snapshot. For continuous scraping
in production, point Prometheus at the running node's `/metrics` HTTP
endpoint instead (`monitoring.prometheus_port` in config).

```bash
p2p metrics
curl http://localhost:9090/metrics   # while a node is running via `p2p run`
```

### `p2p peer list`

Shows the peers configured under `network.peers` in the config file.

```bash
p2p peer list --config config.yaml
```

### `p2p event export`

Exports persistence store metadata (event count, snapshot count,
checkpoint count) as JSON.

```bash
p2p event export --output events.json
p2p event export   # prints to stdout if --output is omitted
```

### `p2p status`

Health check: exit code `0` if config loads and validates, `1` otherwise.
Designed for scripts and container healthchecks:

```bash
p2p status && echo "Node is healthy" || exit 1
```

### `p2p version`

Prints version and a summary of the implemented engine set.

```bash
p2p version
```

## Environment Variable Overrides

Every config value can be overridden without editing the YAML file
(12-factor app pattern — see [ADR-007](../ADRs/ADR-007-configuration.md)):

| Variable | Overrides |
|---|---|
| `P2P_NETWORK_PORT` | `network.port` |
| `P2P_PERSISTENCE_ENGINE` | `persistence.engine` |
| `P2P_PERSISTENCE_PATH` | `persistence.path` |
| `P2P_MONITORING_PROMETHEUS_PORT` | `monitoring.prometheus_port` |
| `P2P_CONSENSUS_TIMEOUT_SEC` | `consensus.timeout_sec` |
| `P2P_CONSENSUS_BYZANTINE_THRESHOLD` | `consensus.byzantine_threshold` |

## Local 3-Node Network (Docker Compose)

```bash
docker-compose up --build
```

This starts three containerized nodes (`alice`, `bob`, `charlie`) with
metrics exposed on `localhost:9091`, `9092`, `9093` respectively.

## Error Messages

Per [ADR-008](../ADRs/ADR-008-cli-strategy.md), errors are always
accompanied by context and next steps — never a bare stack trace or a
single cryptic line:

```
Error: failed to load config from config.yaml: config validation failed:
invalid configuration:
  - network.port must be in [1024, 65535], got 80

Run 'p2p --help' for usage.
```
