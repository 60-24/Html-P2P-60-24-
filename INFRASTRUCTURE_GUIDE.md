# P2P 60-24 Autonomous Infrastructure

## Overview

This infrastructure enables **autonomous AI development** of P2P 60-24 while maintaining architectural integrity and auditability.

**Key Files Created:**
1. ✅ `P2P_GRAPH.jsonld` — Machine-readable ontology
2. ✅ `claude-code-config.yaml` — Operating procedures for Claude Code
3. ✅ `.github/workflows/test.yml` — CI/CD validation
4. ✅ `checkpoint-system.py` — Session tracking & compression
5. ✅ `evidence-chain.py` — Decision logging with provenance

---

## Architecture: How It Works

```
HTML Docs (ToJa)
    ↓ (human source)
P2P_GRAPH.jsonld (machine-readable)
    ↓ (automated extraction)
Claude Code (autonomous agent)
    ↓ (7-step procedure)
Code + Tests
    ↓ (automated)
CI/CD Workflows
    ├─ test.yml (unit + integration)
    ├─ graph-validate.yml (ontology check)
    ├─ contract-verify.yml (code vs contract)
    └─ security.yml (crypto + secrets)
    ↓ (result)
✅ PASS → Merge to main
❌ FAIL → Evidence logged, AI fixes
    ↓
Session Checkpoint
    ├─ State snapshot
    ├─ Evidence chain
    ├─ Decisions made
    └─ Next steps
```

---

## 1. P2P_GRAPH.jsonld — The Graph

**Purpose**: Centralized source of truth for ontology, contracts, invariants.

**Structure**:
```json
{
  "@context": {...},
  "@graph": [
    {
      "@id": "omnikernel:Identity",
      "@type": "owl:Class",
      "level": "OmniKernel",
      "axiom": "INVARIANT: Identity is immutable"
    }
  ]
}
```

**When to Update**:
- New concept added to ontology → add node
- New contract defined → add contract node
- Invariant discovered → add axiom
- **Never manually edit** — extract from HTML docs

**Usage**:
```bash
# Validate graph structure
python3 -c "import json; json.load(open('P2P_GRAPH.jsonld'))"

# Query nodes
grep -A5 '"@id": "omnikernel:' P2P_GRAPH.jsonld

# Use in Claude Code
# The agent loads graph before every coding session
```

---

## 2. claude-code-config.yaml — Agent Procedures

**Purpose**: Define how Claude Code should operate autonomously.

**Key Sections**:

### Before Coding (mandatory)
1. IDENTIFY_NODE — load from graph
2. READ_CONTRACT — what must be implemented
3. MAP_DEPENDENCIES — who uses this? who provides this?
4. ANALYZE_TESTS — coverage gaps
5. PROTOCOL_CHECK — verify compliance

### During Coding (methodology)
- **TDD_STRICT**: Spec → RED test → GREEN impl → REFACTOR
- **Rules**: No silent behavior changes, explicit errors, no global state
- **Stopping conditions**: If contract conflicts → STOP and report

### After Coding (validation)
```
TASK
WHAT CHANGED
FILES
CONTRACTS AFFECTED
GRAPH IMPACT
TESTS
RESULT
KNOWN LIMITATIONS
NEXT STEP
```

**Usage**:
```bash
# Agent loads this at start
claude-code --config claude-code-config.yaml

# Claude Code follows 7-step procedure for each node
# Agent can self-correct within GREEN_ZONE
# Agent escalates to YELLOW/RED zones for human review
```

---

## 3. CI/CD Workflows — Automated Validation

**File**: `.github/workflows/test.yml`

**Pipeline**:

| Stage | Action | Artifact |
|-------|--------|----------|
| **test** | Unit + integration tests | coverage.html |
| **lint** | Code quality (golangci, gofmt, vet) | lint report |
| **graph-validate** | P2P_GRAPH.jsonld syntax + invariants | validation log |
| **contract-verify** | Code implements contracts | test results |
| **security** | gosec, secret scan, crypto audit | security report |

**How It Works**:

1. Push code → GitHub
2. Actions trigger automatically
3. Each workflow validates one aspect
4. **Result**: ✅ PASS → merge OR ❌ FAIL → Claude Code fixes
5. Evidence logged regardless

**Usage**:
```bash
# View results
cd repo
git log --oneline -n 10  # Check workflow status
gh workflow view test  # GitHub CLI

# Manual trigger (if needed)
gh workflow run test.yml
```

---

## 4. checkpoint-system.py — Session Tracking

**Purpose**: Automatically save session state, compress context, track test results.

**What It Tracks**:

```yaml
session:
  number: 1
  timestamp: "2026-09-25T14:30:00Z"

work:
  nodes_implemented: [EventBus, Broadcasting, Listening]
  decisions_made: ["TDD workflow", "Ed25519 only for crypto"]
  next_steps: "Implement StateEngine"

state:
  git: {commit: abc123, branch: main}
  tests: {passed: 47, failed: 0, status: PASS}
  graph: {nodes: 24, edges: 18}
```

**Usage**:

```bash
# Create checkpoint (auto-runs after session)
python3 checkpoint-system.py create \
  "Implemented EventBus + Broadcasting" \
  "EventBus,Broadcasting" \
  "Used TDD; Ed25519 signing" \
  "Next: Implement StateEngine + Consensus"

# Load prior checkpoint to resume
python3 checkpoint-system.py load 1

# Check status
python3 checkpoint-system.py status

# Output: sessions/SES-001/
#   ├── checkpoint.yaml (human-readable)
#   ├── checkpoint.json (machine-readable)
#   ├── EVIDENCE_CHAIN.md (decisions)
#   └── SUMMARY.txt (working notes)
```

**How Compression Works**:
```
Full context: 10,000 tokens → Compress → 500 tokens
├─ Keep: key decisions, node names, test results
├─ Drop: intermediate outputs, verbose logs
└─ Result: Restore in next session with full continuity
```

---

## 5. evidence-chain.py — Decision Audit Trail

**Purpose**: Log every coding decision with provenance, test results, impact.

**Entry Format**:

```python
entry = (DecisionBuilder("omnikernel:EventBus", "Implement publishing")
    .rationale("Required for TERRA OS Engine #1")
    .test_passed(87.5)  # coverage %
    .impacts("terraos:Broadcasting", "terraos:Listening")
    .invariants("INVARIANT: All events must be signed")
    .contracts("EventBusContract")
    .confidence(Confidence.HIGH)
    .next("terraos:Broadcasting")
    .build())

chain.add_entry(entry)
```

**Output Files** (per session):

| File | Format | Audience |
|------|--------|----------|
| EVIDENCE_CHAIN.json | Structured JSON | Machines (AI learning) |
| EVIDENCE_CHAIN.md | Markdown | Humans (audit trail) |

**Statistics**:

```
Count: 47 decisions
Passed: 47/47 (100%)
Avg Coverage: 87.3%
High Confidence: 44/47 (93.6%)
Unique Nodes Modified: 8
Total Impact: 12 modules
```

**Usage**:

```bash
# AI automatically logs during coding:
# Each test result → evidence entry
# Each contract satisfied → logged
# Each impact detected → documented

# Human review after session:
cat sessions/SES-001/EVIDENCE_CHAIN.md
# Shows every decision made and why
```

---

## Integration Flow: A Concrete Example

### Scenario: Implement Node EventBus

**Step 1: Agent Loads Graph**
```
P2P_GRAPH.jsonld
→ Find node: omnikernel:EventBus
→ @id, @type, level, axioms, contracts
→ Incoming edges: (none)
→ Outgoing edges: PUBLISHES → Event, DELIVERS → Subscriber
```

**Step 2: Agent Reads Config**
```
claude-code-config.yaml
→ 5-step pre-coding procedure
→ TDD methodology
→ Stopping conditions
→ Test requirements
```

**Step 3: Agent Implements**
```
SPEC: Event Publishing (from EventBusContract)
RED:  Tests fail (new code doesn't exist)
GREEN: Implement Event struct + Emit() method
REFACTOR: Clean up, add error handling
TEST: All 12 tests pass (87.5% coverage)
```

**Step 4: GitHub Actions**
```
Push → Workflows trigger
├─ test.yml: ✅ 12/12 PASS
├─ lint.yml: ✅ No issues
├─ graph-validate.yml: ✅ Node valid
├─ contract-verify.yml: ✅ EventBusContract implemented
└─ security.yml: ✅ Ed25519 used, no secrets
→ Merge to main
```

**Step 5: Evidence Logged**
```
Evidence Entry:
├─ Node: omnikernel:EventBus
├─ Decision: TDD + Ed25519 signing
├─ Test Result: PASS (87.5%)
├─ Contracts: EventBusContract ✓
├─ Impact: Broadcasting, Listening (2 modules)
├─ Confidence: HIGH
└─ Next: terraos:Broadcasting
```

**Step 6: Checkpoint**
```
Session SES-001 saved:
├─ Nodes implemented: [EventBus]
├─ Tests: 12/12 passed
├─ Decisions: 1 (TDD + crypto)
├─ Next: Implement Broadcasting
└─ Evidence: 1 entry logged
```

---

## How to Use This Infrastructure

### For Human (ToJa):

1. **Review Checkpoints**
   ```bash
   cd repo
   cat sessions/SES-001/checkpoint.yaml  # Status snapshot
   cat sessions/SES-001/EVIDENCE_CHAIN.md  # Decisions
   ```

2. **Approve/Escalate**
   - GREEN zone: Agent proceeds (tests pass, no issues)
   - YELLOW zone: Agent requests approval (contract change, API change)
   - RED zone: Human decides (ontology, trust model, constitution)

3. **Update Ontology**
   - When new concept discovered → add to HTML docs
   - P2P_GRAPH.jsonld updates automatically (CI job)

### For Agent (Claude Code):

1. **Start Session**
   ```
   Load P2P_GRAPH.jsonld
   Load claude-code-config.yaml
   Load prior checkpoint (for context)
   → Ready to code
   ```

2. **Pick Next Node**
   ```
   Query graph: MATCH (n) WHERE n.status = "TODO"
   Select highest-priority node
   Apply 7-step procedure
   ```

3. **After Each Node**
   ```
   Log evidence entry
   Run CI/CD
   If ✅ PASS → checkpoint saved
   If ❌ FAIL → log error + fix + retry
   ```

---

## Autonomy Zones

| Zone | What Agent Can Do | When to Escalate |
|------|------------------|------------------|
| 🟢 **GREEN** | Tests, docs, refactoring, safe bugfixes | N/A |
| 🟡 **YELLOW** | Contract changes, API changes, performance | Need human approval |
| 🔴 **RED** | Ontology, trust model, constitution | Must ask human |
| ⚫ **BLACK** | Error: invariant violated, contradiction detected | STOP immediately |

---

## Maintenance & Evolution

### Weekly:
- Review checkpoint summaries
- Check evidence chains for patterns
- Validate graph against code

### Monthly:
- Update P2P_GRAPH.jsonld from HTML docs (if changes made)
- Audit evidence statistics (decision quality, test pass rates)
- Refactor tech debt logged in evidence

### Quarterly:
- Full consistency audit (code vs graph vs contracts)
- Archive old sessions
- Update documentation

---

## Troubleshooting

**"Graph validation failed"**
```bash
# Check JSON-LD syntax
python3 -c "import json; json.load(open('P2P_GRAPH.jsonld')); print('Valid')"

# Check for missing invariants
grep -c "axiom" P2P_GRAPH.jsonld
```

**"Test failed but no evidence?"**
```bash
# Evidence should auto-log. Check:
ls sessions/SES-XXX/EVIDENCE_CHAIN.*

# If missing, manually create:
python3 evidence-chain.py --node broken_node --decision "fix_attempt" --result FAIL
```

**"Agent stuck in loop"**
```bash
# Check stopping conditions in config
grep -A5 "stopping_conditions:" claude-code-config.yaml

# Check if invariant violation detected
cat sessions/SES-XXX/EVIDENCE_CHAIN.md | grep -i "violation"
```

---

## Next Steps

1. **Integrate into repo** (copy files to `https://github.com/60-24/Html-P2P-60-24-`)
2. **Set up GitHub Actions** (add workflows to `.github/`)
3. **Configure Claude Code** (use claude-code-config.yaml in agent startup)
4. **Create first checkpoint** (baseline state before first coding session)
5. **Start Session SES-001** (first autonomous implementation round)

---

## Questions?

- **Graph structure**: Check P2P_GRAPH.jsonld schema
- **Agent procedures**: Read claude-code-config.yaml + AGENT_PROTOCOL.md
- **CI/CD**: Check .github/workflows/*.yml
- **Sessions**: See sessions/SES-XXX/ directories
- **Decisions**: Review EVIDENCE_CHAIN.md for full audit trail

**Last Updated**: 2026-09-25  
**Status**: ✅ READY FOR DEPLOYMENT
