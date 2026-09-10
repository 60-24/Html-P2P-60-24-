package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTestConfig(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test-config.yaml")
	content := `
network:
  port: 7024
  peers:
    - "peer1:6024"
    - "peer2:6024"
persistence:
  engine: inmemory
  path: /tmp/p2p-cli-test
monitoring:
  prometheus_port: 9091
consensus:
  timeout_sec: 5
  byzantine_threshold: 0.66
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}
	return path
}

// Test 1: Root Command Has All Subcommands Registered
func TestRootCommandHasSubcommands(t *testing.T) {
	root := NewRootCommand()

	expected := []string{"run", "config", "metrics", "peer", "event", "status", "version"}
	found := make(map[string]bool)

	for _, cmd := range root.Commands() {
		found[cmd.Name()] = true
	}

	for _, name := range expected {
		if !found[name] {
			t.Errorf("Expected subcommand %q to be registered", name)
		}
	}
}

// Test 2: Version Command Output
func TestVersionCommand(t *testing.T) {
	root := NewRootCommand()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "p2p 60-24") {
		t.Errorf("Expected version output to mention p2p 60-24, got: %s", output)
	}
}

// Test 3: Config Show Command (reads real file)
func TestConfigShowCommand(t *testing.T) {
	path := writeTestConfig(t)

	root := NewRootCommand()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"config", "show", "--config", path})

	if err := root.Execute(); err != nil {
		t.Fatalf("config show command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "7024") {
		t.Errorf("Expected config output to contain port 7024, got: %s", output)
	}
}

// Test 4: Status Command with Valid Config (healthy path)
func TestStatusCommandHealthy(t *testing.T) {
	path := writeTestConfig(t)

	root := NewRootCommand()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"status", "--config", path})

	err := root.Execute()
	if err != nil {
		t.Errorf("status command should succeed for valid config: %v", err)
	}
}

// Test 5: Status Command with Missing Config (unhealthy path).
// Prior to review, this command called os.Exit(1) directly on failure,
// which would have killed the test binary if tested here. Fixed to return
// an error instead (see status.go) - root.Execute() already maps any
// command error to exit code 1, so this test can now safely exercise the
// actual unhealthy path in-process.
func TestStatusCommandUnhealthy(t *testing.T) {
	root := NewRootCommand()
	root.SetArgs([]string{"status", "--config", "/nonexistent/config.yaml"})

	err := root.Execute()
	if err == nil {
		t.Errorf("Expected error for status check against missing config")
	}
	if !strings.Contains(err.Error(), "UNHEALTHY") {
		t.Errorf("Expected UNHEALTHY marker in error, got: %v", err)
	}
}

// Test 6: Peer List Command
func TestPeerListCommand(t *testing.T) {
	path := writeTestConfig(t)

	root := NewRootCommand()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"peer", "list", "--config", path})

	if err := root.Execute(); err != nil {
		t.Fatalf("peer list command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "peer1:6024") {
		t.Errorf("Expected peer1:6024 in output, got: %s", output)
	}
}

// Bonus Test: Event Export to stdout
func TestEventExportCommand(t *testing.T) {
	root := NewRootCommand()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"event", "export"})

	if err := root.Execute(); err != nil {
		t.Fatalf("event export command failed: %v", err)
	}
}

// Bonus Test: Any command with missing config produces wrapped,
// human-friendly error (general contract, checked via `peer list`).
func TestMissingConfigProducesWrappedError(t *testing.T) {
	root := NewRootCommand()
	root.SetArgs([]string{"peer", "list", "--config", "/nonexistent/config.yaml"})

	err := root.Execute()
	if err == nil {
		t.Errorf("Expected error for missing config file")
	}
	if !strings.Contains(err.Error(), "failed to load config") {
		t.Errorf("Expected human-friendly error message, got: %v", err)
	}
}

// Bonus Test: Unknown Flag Produces Human-Friendly Usage (ADR-008 requirement)
func TestUnknownFlagShowsUsage(t *testing.T) {
	root := NewRootCommand()
	root.SetArgs([]string{"run", "--invalid-flag"})

	err := root.Execute()
	if err == nil {
		t.Errorf("Expected error for unknown flag")
	}
}
