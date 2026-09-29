#!/usr/bin/env python3
"""
Evidence Chain Logger
Track all coding decisions with full provenance and outcomes
Used by autonomous Claude Code for learning and auditing
"""

import json
import datetime
from pathlib import Path
from typing import Dict, List, Any, Optional
from enum import Enum

class Confidence(Enum):
    """Confidence level in decision"""
    HIGH = "HIGH"
    MEDIUM = "MEDIUM"
    LOW = "LOW"

class TestResult(Enum):
    """Test outcome"""
    PASS = "PASS"
    FAIL = "FAIL"
    PARTIAL = "PARTIAL"
    SKIP = "SKIP"

class EvidenceEntry:
    """Single evidence entry - one decision/action"""
    
    def __init__(self,
                 node_id: str,
                 decision: str,
                 rationale: str,
                 timestamp: Optional[str] = None):
        
        self.node_id = node_id
        self.decision = decision
        self.rationale = rationale
        self.timestamp = timestamp or datetime.datetime.utcnow().isoformat()
        
        self.test_result: Optional[TestResult] = None
        self.coverage_percent: float = 0.0
        self.impact: List[str] = []
        self.confidence: Confidence = Confidence.MEDIUM
        self.invariants_checked: List[str] = []
        self.contracts_satisfied: List[str] = []
        self.warnings: List[str] = []
        self.next_node: Optional[str] = None
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary"""
        return {
            "timestamp": self.timestamp,
            "node_id": self.node_id,
            "decision": self.decision,
            "rationale": self.rationale,
            "test_result": self.test_result.value if self.test_result else None,
            "coverage_percent": self.coverage_percent,
            "impact": self.impact,
            "confidence": self.confidence.value,
            "invariants_checked": self.invariants_checked,
            "contracts_satisfied": self.contracts_satisfied,
            "warnings": self.warnings,
            "next_node": self.next_node
        }
    
    def to_markdown(self) -> str:
        """Convert to markdown for human reading"""
        lines = [
            f"## {self.timestamp}",
            "",
            f"**Node**: `{self.node_id}`",
            f"**Decision**: {self.decision}",
            f"**Rationale**: {self.rationale}",
            ""
        ]
        
        if self.test_result:
            lines.append(f"**Test Result**: {self.test_result.value} ({self.coverage_percent:.1f}% coverage)")
        
        if self.invariants_checked:
            lines.append(f"**Invariants Checked**: {', '.join(self.invariants_checked)}")
        
        if self.contracts_satisfied:
            lines.append(f"**Contracts Satisfied**: {', '.join(self.contracts_satisfied)}")
        
        if self.impact:
            lines.append(f"**Impact**: {', '.join(self.impact)}")
        
        if self.warnings:
            lines.append("**Warnings**:")
            for w in self.warnings:
                lines.append(f"  - {w}")
        
        lines.append(f"**Confidence**: {self.confidence.value}")
        
        if self.next_node:
            lines.append(f"**Next Node**: {self.next_node}")
        
        lines.extend(["", "---", ""])
        
        return "\n".join(lines)


class EvidenceChain:
    """Maintain ordered chain of evidence entries"""
    
    def __init__(self, session_dir: Path):
        self.session_dir = Path(session_dir)
        self.entries: List[EvidenceEntry] = []
        self.load_existing()
    
    def load_existing(self) -> None:
        """Load existing evidence chain from file"""
        chain_file = self.session_dir / "EVIDENCE_CHAIN.json"
        
        if chain_file.exists():
            with open(chain_file, 'r') as f:
                data = json.load(f)
                self.entries = data.get("entries", [])
    
    def add_entry(self, entry: EvidenceEntry) -> None:
        """Add entry to chain"""
        self.entries.append(entry)
        self.save()
    
    def save(self) -> None:
        """Save chain to files"""
        # Save as JSON (machine-readable)
        json_file = self.session_dir / "EVIDENCE_CHAIN.json"
        with open(json_file, 'w') as f:
            json.dump({
                "session": self.session_dir.name,
                "count": len(self.entries),
                "entries": [e.to_dict() if isinstance(e, EvidenceEntry) else e for e in self.entries]
            }, f, indent=2)
        
        # Save as Markdown (human-readable)
        md_file = self.session_dir / "EVIDENCE_CHAIN.md"
        with open(md_file, 'w') as f:
            f.write("# Evidence Chain\n\n")
            f.write(f"Session: {self.session_dir.name}\n")
            f.write(f"Entries: {len(self.entries)}\n\n")
            
            for entry in self.entries:
                if isinstance(entry, EvidenceEntry):
                    f.write(entry.to_markdown())
                else:
                    # Backward compat with dict entries
                    f.write(f"## {entry.get('timestamp', 'N/A')}\n")
                    f.write(f"- Node: {entry.get('node_id')}\n")
                    f.write(f"- Decision: {entry.get('decision')}\n\n")
    
    def statistics(self) -> Dict[str, Any]:
        """Generate statistics over chain"""
        if not self.entries:
            return {"count": 0}
        
        passed = sum(1 for e in self.entries 
                    if isinstance(e, EvidenceEntry) and e.test_result == TestResult.PASS)
        
        failed = sum(1 for e in self.entries 
                    if isinstance(e, EvidenceEntry) and e.test_result == TestResult.FAIL)
        
        avg_coverage = sum(e.coverage_percent for e in self.entries 
                          if isinstance(e, EvidenceEntry)) / len([e for e in self.entries if isinstance(e, EvidenceEntry)]) if self.entries else 0
        
        high_confidence = sum(1 for e in self.entries 
                             if isinstance(e, EvidenceEntry) and e.confidence == Confidence.HIGH)
        
        unique_nodes = set(e.node_id if isinstance(e, EvidenceEntry) else e.get('node_id') 
                          for e in self.entries)
        
        total_impact = len(set(
            imp for e in self.entries if isinstance(e, EvidenceEntry)
            for imp in e.impact
        ))
        
        return {
            "count": len(self.entries),
            "passed": passed,
            "failed": failed,
            "pass_rate": f"{(passed / len(self.entries) * 100):.1f}%" if self.entries else "0%",
            "avg_coverage": f"{avg_coverage:.1f}%",
            "high_confidence_decisions": high_confidence,
            "unique_nodes_modified": len(unique_nodes),
            "total_modules_impacted": total_impact
        }
    
    def audit_trail(self) -> str:
        """Generate human-readable audit trail"""
        lines = [
            "# AUDIT TRAIL",
            "",
            f"Session: {self.session_dir.name}",
            f"Total Decisions: {len(self.entries)}",
            ""
        ]
        
        # Group by node
        by_node: Dict[str, List] = {}
        for entry in self.entries:
            node = entry.node_id if isinstance(entry, EvidenceEntry) else entry.get('node_id', 'unknown')
            if node not in by_node:
                by_node[node] = []
            by_node[node].append(entry)
        
        for node in sorted(by_node.keys()):
            decisions = by_node[node]
            lines.append(f"## Node: {node}")
            lines.append(f"Decisions: {len(decisions)}")
            
            for entry in decisions:
                if isinstance(entry, EvidenceEntry):
                    lines.append(f"  - {entry.timestamp}: {entry.decision}")
                    if entry.test_result:
                        lines.append(f"    Result: {entry.test_result.value}")
            
            lines.append("")
        
        return "\n".join(lines)


class DecisionBuilder:
    """Fluent builder for creating evidence entries"""
    
    def __init__(self, node_id: str, decision: str):
        self.entry = EvidenceEntry(node_id, decision, "")
    
    def rationale(self, text: str) -> 'DecisionBuilder':
        self.entry.rationale = text
        return self
    
    def test_passed(self, coverage: float) -> 'DecisionBuilder':
        self.entry.test_result = TestResult.PASS
        self.entry.coverage_percent = coverage
        return self
    
    def test_failed(self) -> 'DecisionBuilder':
        self.entry.test_result = TestResult.FAIL
        return self
    
    def impacts(self, *modules: str) -> 'DecisionBuilder':
        self.entry.impact = list(modules)
        return self
    
    def invariants(self, *invariants: str) -> 'DecisionBuilder':
        self.entry.invariants_checked = list(invariants)
        return self
    
    def contracts(self, *contracts: str) -> 'DecisionBuilder':
        self.entry.contracts_satisfied = list(contracts)
        return self
    
    def confidence(self, level: Confidence) -> 'DecisionBuilder':
        self.entry.confidence = level
        return self
    
    def warn(self, message: str) -> 'DecisionBuilder':
        self.entry.warnings.append(message)
        return self
    
    def next(self, node_id: str) -> 'DecisionBuilder':
        self.entry.next_node = node_id
        return self
    
    def build(self) -> EvidenceEntry:
        return self.entry


# Example usage
if __name__ == "__main__":
    from pathlib import Path
    
    # Create example session
    session_dir = Path("/tmp/SES-001")
    session_dir.mkdir(exist_ok=True)
    
    chain = EvidenceChain(session_dir)
    
    # Example: First decision
    entry1 = (DecisionBuilder("omnikernel:EventBus", "Implement Event publishing")
        .rationale("Required for TERRA OS Engine #1, specification from P2P_GRAPH.jsonld")
        .test_passed(87.5)
        .impacts("terraos:Broadcasting", "terraos:Listening")
        .invariants("INVARIANT: All events must be signed")
        .contracts("EventBusContract")
        .confidence(Confidence.HIGH)
        .next("terraos:Broadcasting")
        .build())
    
    chain.add_entry(entry1)
    
    # Print statistics
    print("Evidence Chain Statistics:")
    print(json.dumps(chain.statistics(), indent=2))
    print("\nAudit Trail:")
    print(chain.audit_trail())
