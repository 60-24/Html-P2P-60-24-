#!/usr/bin/env python3
"""
P2P 60-24 Session Checkpoint System
Automated state tracking, context compression, evidence chain logging
"""

import json
import yaml
import hashlib
import datetime
import subprocess
from pathlib import Path
from typing import Dict, List, Any

class CheckpointManager:
    """Manage session checkpoints with compression and evidence tracking"""
    
    def __init__(self, repo_root: str = "."):
        self.repo_root = Path(repo_root)
        self.sessions_dir = self.repo_root / "sessions"
        self.sessions_dir.mkdir(exist_ok=True)
    
    def get_current_session_number(self) -> int:
        """Get next session number"""
        existing = list(self.sessions_dir.glob("SES-*"))
        if not existing:
            return 1
        numbers = [int(d.name.split("-")[1]) for d in existing]
        return max(numbers) + 1
    
    def get_git_info(self) -> Dict[str, Any]:
        """Extract git commit info"""
        try:
            commit = subprocess.run(
                ["git", "rev-parse", "HEAD"],
                capture_output=True, text=True, cwd=self.repo_root
            ).stdout.strip()
            
            branch = subprocess.run(
                ["git", "rev-parse", "--abbrev-ref", "HEAD"],
                capture_output=True, text=True, cwd=self.repo_root
            ).stdout.strip()
            
            shortlog = subprocess.run(
                ["git", "log", "-1", "--pretty=%B"],
                capture_output=True, text=True, cwd=self.repo_root
            ).stdout.strip()
            
            return {
                "commit": commit[:12],
                "branch": branch,
                "message": shortlog.split("\n")[0]
            }
        except:
            return {"commit": "unknown", "branch": "unknown", "message": ""}
    
    def get_test_results(self) -> Dict[str, Any]:
        """Extract latest test results"""
        try:
            result = subprocess.run(
                ["go", "test", "./...", "-v", "-json"],
                capture_output=True, text=True, cwd=self.repo_root,
                timeout=30
            )
            
            lines = result.stdout.strip().split("\n")
            passed = failed = 0
            
            for line in lines:
                if '"Action":"pass"' in line:
                    passed += 1
                elif '"Action":"fail"' in line:
                    failed += 1
            
            return {
                "passed": passed,
                "failed": failed,
                "status": "PASS" if failed == 0 else "FAIL"
            }
        except:
            return {"passed": 0, "failed": 0, "status": "UNKNOWN"}
    
    def get_graph_snapshot(self) -> Dict[str, Any]:
        """Extract P2P_GRAPH.jsonld summary"""
        graph_file = self.repo_root / "P2P_GRAPH.jsonld"
        
        if not graph_file.exists():
            return {"nodes": 0, "edges": 0}
        
        try:
            with open(graph_file, 'r') as f:
                graph = json.load(f)
            
            nodes = graph.get("@graph", [])
            edge_count = sum(1 for n in nodes if any(k.startswith("incoming") or k.startswith("outgoing") for k in n.keys()))
            
            return {
                "nodes": len(nodes),
                "edges": edge_count,
                "primitives": [n.get("@id") for n in nodes if n.get("level") == "OmniKernel"][:5]
            }
        except:
            return {"nodes": 0, "edges": 0}
    
    def compress_context(self, full_context: str, max_tokens: int = 500) -> str:
        """Compress session context (simple word-based compression)"""
        words = full_context.split()
        if len(words) <= max_tokens:
            return full_context
        
        # Simple: keep first and last parts
        keep = max_tokens // 2
        return " ".join(words[:keep]) + " [... compressed ...] " + " ".join(words[-keep:])
    
    def create_checkpoint(self, 
                         summary: str,
                         decisions: List[str],
                         nodes_implemented: List[str],
                         next_steps: str,
                         context: str = "") -> Path:
        """Create session checkpoint"""
        
        ses_num = self.get_current_session_number()
        ses_dir = self.sessions_dir / f"SES-{ses_num:03d}"
        ses_dir.mkdir(exist_ok=True)
        
        timestamp = datetime.datetime.utcnow().isoformat()
        
        checkpoint = {
            "session": {
                "number": ses_num,
                "timestamp": timestamp,
                "status": "ACTIVE"
            },
            
            "summary": summary,
            
            "work": {
                "nodes_implemented": nodes_implemented,
                "decisions_made": decisions,
                "next_steps": next_steps
            },
            
            "state": {
                "git": self.get_git_info(),
                "tests": self.get_test_results(),
                "graph": self.get_graph_snapshot(),
                "context_compressed": self.compress_context(context)
            },
            
            "metadata": {
                "ontology_version": "SFO v1.1",
                "kernel_version": "KERNEL-DNA v3.0",
                "microkernel_version": "v0.1"
            }
        }
        
        # Save as YAML
        checkpoint_file = ses_dir / "checkpoint.yaml"
        with open(checkpoint_file, 'w') as f:
            yaml.dump(checkpoint, f, default_flow_style=False)
        
        # Save as JSON for machines
        checkpoint_json = ses_dir / "checkpoint.json"
        with open(checkpoint_json, 'w') as f:
            json.dump(checkpoint, f, indent=2)
        
        print(f"✓ Checkpoint created: {checkpoint_file}")
        return ses_dir
    
    def load_checkpoint(self, session_num: int) -> Dict[str, Any]:
        """Load prior checkpoint"""
        checkpoint_file = self.sessions_dir / f"SES-{session_num:03d}" / "checkpoint.yaml"
        
        if not checkpoint_file.exists():
            raise FileNotFoundError(f"Checkpoint {session_num} not found")
        
        with open(checkpoint_file, 'r') as f:
            return yaml.safe_load(f)
    
    def log_evidence(self, session_dir: Path, entry: Dict[str, Any]) -> None:
        """Log decision to evidence chain"""
        evidence_file = session_dir / "EVIDENCE_CHAIN.md"
        
        line = f"""
### Evidence Entry {entry['timestamp']}

**Node**: {entry.get('node_id', 'N/A')}  
**Decision**: {entry.get('decision', 'N/A')}  
**Rationale**: {entry.get('rationale', 'N/A')}  
**Test Result**: {entry.get('test_result', 'N/A')}  
**Impact**: {entry.get('impact', 'N/A')}  
**Confidence**: {entry.get('confidence', 'UNKNOWN')}  

---
"""
        
        with open(evidence_file, 'a') as f:
            f.write(line)
        
        print(f"✓ Evidence logged: {evidence_file}")


class CheckpointCLI:
    """CLI interface for checkpoint management"""
    
    def __init__(self):
        self.manager = CheckpointManager()
    
    def cmd_create(self, summary: str, nodes: List[str], decisions: List[str], next_steps: str):
        """Create new checkpoint"""
        print(f"\n📸 Creating checkpoint...")
        session_dir = self.manager.create_checkpoint(
            summary=summary,
            nodes_implemented=nodes,
            decisions_made=decisions,
            next_steps=next_steps
        )
        print(f"✓ Session {session_dir.name} ready")
    
    def cmd_load(self, session_num: int):
        """Load and display checkpoint"""
        print(f"\n📖 Loading checkpoint SES-{session_num:03d}...")
        checkpoint = self.manager.load_checkpoint(session_num)
        print(json.dumps(checkpoint, indent=2, default=str))
    
    def cmd_status(self):
        """Show current status"""
        print("\n📊 Session Status")
        print("=" * 50)
        
        # List all sessions
        sessions = sorted(self.manager.sessions_dir.glob("SES-*"))
        
        if not sessions:
            print("No sessions yet")
            return
        
        for ses_dir in sessions[-5:]:  # Last 5
            checkpoint_file = ses_dir / "checkpoint.yaml"
            if checkpoint_file.exists():
                with open(checkpoint_file, 'r') as f:
                    cp = yaml.safe_load(f)
                
                print(f"\n{ses_dir.name}")
                print(f"  Timestamp: {cp['session']['timestamp']}")
                print(f"  Nodes: {len(cp['work']['nodes_implemented'])}")
                print(f"  Tests: {cp['state']['tests']['passed']} passed, {cp['state']['tests']['failed']} failed")
    
    def cmd_summary(self, text: str):
        """Add summary line to current session"""
        sessions = sorted(self.manager.sessions_dir.glob("SES-*"))
        if not sessions:
            print("No active session")
            return
        
        latest = sessions[-1]
        summary_file = latest / "SUMMARY.txt"
        
        with open(summary_file, 'a') as f:
            f.write(f"{datetime.datetime.utcnow().isoformat()} - {text}\n")
        
        print(f"✓ Summary updated: {summary_file}")


if __name__ == "__main__":
    import sys
    
    cli = CheckpointCLI()
    
    if len(sys.argv) < 2:
        print("Usage:")
        print("  python checkpoint-system.py create <summary> [node1,node2] [decision1;decision2] <next_steps>")
        print("  python checkpoint-system.py load <session_num>")
        print("  python checkpoint-system.py status")
        sys.exit(1)
    
    cmd = sys.argv[1]
    
    if cmd == "create" and len(sys.argv) >= 5:
        summary = sys.argv[2]
        nodes = sys.argv[3].split(",") if sys.argv[3] != "[]" else []
        decisions = sys.argv[4].split(";") if sys.argv[4] != "[]" else []
        next_steps = sys.argv[5] if len(sys.argv) > 5 else "TBD"
        cli.cmd_create(summary, nodes, decisions, next_steps)
    
    elif cmd == "load" and len(sys.argv) > 2:
        cli.cmd_load(int(sys.argv[2]))
    
    elif cmd == "status":
        cli.cmd_status()
    
    else:
        print(f"Unknown command: {cmd}")
        sys.exit(1)
