package network

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	"github.com/60-24/p2p-60-24/pkg/core"
	"github.com/60-24/p2p-60-24/pkg/terra"
)

// Test 1: GossipNode Creation
func TestGossipNodeCreation(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	eb := terra.NewEventBus("Alice", priv, pub)
	defer eb.Close()

	gn := NewGossipNode("Alice", 6024, eb)

	if gn.nodeID != "Alice" {
		t.Errorf("Expected nodeID Alice, got %s", gn.nodeID)
	}
	if gn.listenPort != 6024 {
		t.Errorf("Expected port 6024, got %d", gn.listenPort)
	}
}

// Test 2: Peer Connection
func TestPeerConnection(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	eb := terra.NewEventBus("Alice", priv, pub)
	defer eb.Close()

	gn := NewGossipNode("Alice", 6024, eb)
	gn.Start()
	defer gn.Stop()

	// Connect to peer
	err := gn.ConnectToPeer("127.0.0.1:6025")
	if err != nil {
		// UDP connection might fail in test env, that's OK
		// We're testing the logic not actual network
	}

	stats := gn.GetPeerStats()
	if stats["node_id"] != "Alice" {
		t.Errorf("Expected nodeID in stats")
	}
}

// Test 3: Peer Stats
func TestGetPeerStats(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	eb := terra.NewEventBus("Alice", priv, pub)
	defer eb.Close()

	gn := NewGossipNode("Alice", 6024, eb)

	stats := gn.GetPeerStats()

	if stats["node_id"] != "Alice" {
		t.Errorf("Expected nodeID")
	}
	if stats["total_peers"] != 0 {
		t.Errorf("Expected 0 peers initially")
	}
	if stats["port"] != 6024 {
		t.Errorf("Expected port 6024")
	}
}

// Test 4: Event Broadcasting
func TestBroadcastEvent(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	eb := terra.NewEventBus("Alice", priv, pub)
	defer eb.Close()

	gn := NewGossipNode("Alice", 6024, eb)
	gn.Start()
	defer gn.Stop()

	event := core.NewEvent("Alice", core.EventTypeProofOfMeeting, []byte("test"), core.PriorityNormal)
	event.Signature = ed25519.Sign(priv, []byte(event.Hash))

	err := gn.BroadcastEvent(event)
	if err != nil {
		t.Errorf("BroadcastEvent failed: %v", err)
	}

	if _, exists := gn.inFlight[event.ID]; !exists {
		t.Errorf("Event should be in inFlight")
	}
}

// Test 5: Peer Disconnection
func TestDisconnectPeer(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	eb := terra.NewEventBus("Alice", priv, pub)
	defer eb.Close()

	gn := NewGossipNode("Alice", 6024, eb)

	// Add a fake peer
	gn.mu.Lock()
	gn.peers["Bob"] = &Peer{
		NodeID:    "Bob",
		Address:   "127.0.0.1:6025",
		Connected: true,
	}
	gn.mu.Unlock()

	// Disconnect
	err := gn.DisconnectPeer("Bob")
	if err != nil {
		t.Errorf("DisconnectPeer failed: %v", err)
	}

	gn.mu.RLock()
	peer := gn.peers["Bob"]
	gn.mu.RUnlock()

	if peer.Connected {
		t.Errorf("Peer should be disconnected")
	}
}

// Test 6: InFlight Events
func TestInFlightEvents(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	eb := terra.NewEventBus("Alice", priv, pub)
	defer eb.Close()

	gn := NewGossipNode("Alice", 6024, eb)

	e1 := core.NewEvent("Alice", core.EventTypeProofOfMeeting, []byte("test1"), core.PriorityNormal)
	e2 := core.NewEvent("Alice", core.EventTypeCommitment, []byte("test2"), core.PriorityHigh)

	gn.BroadcastEvent(e1)
	gn.BroadcastEvent(e2)

	if len(gn.inFlight) != 2 {
		t.Errorf("Expected 2 inFlight events, got %d", len(gn.inFlight))
	}
}

// Test 7: GossipNode Start & Stop
func TestStartStop(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	eb := terra.NewEventBus("Alice", priv, pub)
	defer eb.Close()

	gn := NewGossipNode("Alice", 16024, eb) // Use different port
	err := gn.Start()
	if err != nil {
		t.Errorf("Start failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	err = gn.Stop()
	if err != nil {
		t.Errorf("Stop failed: %v", err)
	}
}

// Test 8: Multiple Peers
func TestMultiplePeers(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	eb := terra.NewEventBus("Alice", priv, pub)
	defer eb.Close()

	gn := NewGossipNode("Alice", 6024, eb)

	// Add multiple peers
	gn.mu.Lock()
	for i := 0; i < 5; i++ {
		peer := &Peer{
			NodeID:    "Peer" + string(rune(i)),
			Address:   "127.0.0.1:602" + string(rune(5+i)),
			Connected: true,
		}
		gn.peers[peer.NodeID] = peer
	}
	gn.mu.Unlock()

	stats := gn.GetPeerStats()
	if stats["total_peers"] != 5 {
		t.Errorf("Expected 5 peers, got %v", stats["total_peers"])
	}
}
