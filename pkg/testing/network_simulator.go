package testing

import (
	"fmt"
	"sync"
	"time"

	"github.com/60-24/p2p-60-24/pkg/consensus"
	"github.com/60-24/p2p-60-24/pkg/core"
	"github.com/60-24/p2p-60-24/pkg/data"
	"github.com/60-24/p2p-60-24/pkg/network"
	"github.com/60-24/p2p-60-24/pkg/terra"
)

// SimulatedNode reprezentuje węzeł w sieci testowej
type SimulatedNode struct {
	mu            sync.RWMutex
	NodeID        string
	GossipNode    *network.GossipNode
	DAGManager    *network.DAGManager
	Consensus     *consensus.ConsensusLayer
	EventStore    *data.EventStore
	StateEngine   *data.StateEngine
	EventBus      *terra.EventBus
	IsOnline      bool
	EventsSeeded  int
}

// NetworkSimulator symuluje sieć P2P z wieloma węzłami
type NetworkSimulator struct {
	mu           sync.RWMutex
	nodes        map[string]*SimulatedNode
	partitions   map[string]map[string]bool // nodeID -> set of partitioned nodes
	eventLog     []*core.Event
	eventChan    chan *core.Event
	stopChan     chan bool
	latencies    map[string]int // milliseconds
	metrics      map[string]interface{}
}

// NewNetworkSimulator tworzy nowy simulator
func NewNetworkSimulator() *NetworkSimulator {
	return &NetworkSimulator{
		nodes:       make(map[string]*SimulatedNode),
		partitions:  make(map[string]map[string]bool),
		eventLog:    make([]*core.Event, 0),
		eventChan:   make(chan *core.Event, 1000),
		stopChan:    make(chan bool),
		latencies:   make(map[string]int),
		metrics:     make(map[string]interface{}),
	}
}

// CreateNode tworzy nowy węzeł w sieci
func (ns *NetworkSimulator) CreateNode(nodeID string) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	// Initialize components
	pubKey := []byte("pubkey_" + nodeID)
	privKey := []byte("privkey_" + nodeID)

	eventStore := data.NewEventStore()
	eventBus := terra.NewEventBus(nodeID, privKey, pubKey)
	gossipNode := network.NewGossipNode(nodeID, 6024, eventBus)
	dagManager := network.NewDAGManager(nodeID, eventStore)
	consensusLayer := consensus.NewConsensusLayer(nodeID, dagManager, eventStore)
	stateEngine := data.NewStateEngine(eventStore)

	// Start gossip node
	err := gossipNode.Start()
	if err != nil {
		eventBus.Close()
		return fmt.Errorf("failed to start gossip node: %w", err)
	}

	node := &SimulatedNode{
		NodeID:       nodeID,
		GossipNode:   gossipNode,
		DAGManager:   dagManager,
		Consensus:    consensusLayer,
		EventStore:   eventStore,
		StateEngine:  stateEngine,
		EventBus:     eventBus,
		IsOnline:     true,
		EventsSeeded: 0,
	}

	ns.nodes[nodeID] = node
	ns.partitions[nodeID] = make(map[string]bool)

	return nil
}

// ConnectNodes łączy dwa węzły
func (ns *NetworkSimulator) ConnectNodes(nodeID1, nodeID2 string) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	node1, exists := ns.nodes[nodeID1]
	if !exists {
		return fmt.Errorf("node %s not found", nodeID1)
	}
	node2, exists := ns.nodes[nodeID2]
	if !exists {
		return fmt.Errorf("node %s not found", nodeID2)
	}

	// Add peer relationship
	node1.DAGManager.AddOrUpdateEdge(nodeID1, nodeID2, "connection_proof")
	node2.DAGManager.AddOrUpdateEdge(nodeID2, nodeID1, "connection_proof")

	// Clear partition if exists
	delete(node1.GossipNode.peers, nodeID2)
	delete(node2.GossipNode.peers, nodeID1)

	return nil
}

// SendEvent wysyła event od jednego węzła
func (ns *NetworkSimulator) SendEvent(nodeID string, eventType string, payload []byte) (*core.Event, error) {
	ns.mu.Lock()
	node, exists := ns.nodes[nodeID]
	ns.mu.Unlock()

	if !exists {
		return nil, fmt.Errorf("node %s not found", nodeID)
	}

	if !node.IsOnline {
		return nil, fmt.Errorf("node %s is offline", nodeID)
	}

	// Create event
	event := core.NewEvent(nodeID, eventType, payload, core.PriorityNormal)

	// Publish via EventBus
	node.EventBus.Publish(eventType, payload, core.PriorityNormal)

	// Add to log
	ns.mu.Lock()
	ns.eventLog = append(ns.eventLog, event)
	ns.mu.Unlock()

	// Broadcast to connected peers
	go ns.broadcastEvent(nodeID, event)

	return event, nil
}

// broadcastEvent propaguje event do peer'ów
func (ns *NetworkSimulator) broadcastEvent(fromNode string, event *core.Event) {
	ns.mu.Lock()
	node, exists := ns.nodes[fromNode]
	ns.mu.Unlock()

	if !exists || !node.IsOnline {
		return
	}

	ns.mu.RLock()
	peers := node.GossipNode.peers
	ns.mu.RUnlock()

	for _, peer := range peers {
		if peer.Connected {
			ns.mu.Lock()
			targetNode, exists := ns.nodes[peer.NodeID]
			ns.mu.Unlock()

			if exists && targetNode.IsOnline {
				// Simulate latency
				if latency, exists := ns.latencies[peer.NodeID]; exists {
					time.Sleep(time.Duration(latency) * time.Millisecond)
				}

				// Apply event at target node
				targetNode.EventStore.Append(event)
				targetNode.StateEngine.ApplyEvent(event)
				targetNode.EventsSeeded++

				// Propose for consensus
				targetNode.Consensus.ProposeEvent(event)
			}
		}
	}
}

// PartitionNetwork tworzy podział między węzłami
func (ns *NetworkSimulator) PartitionNetwork(nodeID1, nodeID2 string) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if _, exists := ns.nodes[nodeID1]; !exists {
		return fmt.Errorf("node %s not found", nodeID1)
	}
	if _, exists := ns.nodes[nodeID2]; !exists {
		return fmt.Errorf("node %s not found", nodeID2)
	}

	ns.partitions[nodeID1][nodeID2] = true
	ns.partitions[nodeID2][nodeID1] = true

	return nil
}

// HealPartition naprawia podział sieci
func (ns *NetworkSimulator) HealPartition(nodeID1, nodeID2 string) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	delete(ns.partitions[nodeID1], nodeID2)
	delete(ns.partitions[nodeID2], nodeID1)

	return nil
}

// AwaitConsensus czeka aż event osiągnie consensus
func (ns *NetworkSimulator) AwaitConsensus(eventID string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)

	for {
		if time.Now().After(deadline) {
			return false
		}

		ns.mu.RLock()
		finalized := 0
		for _, node := range ns.nodes {
			if node.IsOnline {
				state := node.Consensus.GetConsensusState(eventID)
				if state != nil && state.Status == "finalized" {
					finalized++
				}
			}
		}
		ns.mu.RUnlock()

		// Require consensus on >66% of online nodes
		onlineCount := ns.countOnlineNodes()
		if finalized > int(float64(onlineCount)*0.66) {
			return true
		}

		time.Sleep(100 * time.Millisecond)
	}
}

// GetNetworkState zwraca snapshot stanu sieci
func (ns *NetworkSimulator) GetNetworkState() map[string]interface{} {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	onlineCount := 0
	totalEvents := 0
	totalEdges := 0

	nodeStates := make(map[string]map[string]interface{})

	for nodeID, node := range ns.nodes {
		if node.IsOnline {
			onlineCount++
		}
		totalEvents += node.EventStore.Size()

		dagStats := node.DAGManager.GetDAGStats()
		totalEdges += dagStats["edges"].(int)

		nodeState := map[string]interface{}{
			"online":       node.IsOnline,
			"events":       node.EventStore.Size(),
			"state_keys":   len(node.StateEngine.GetState()),
			"consensus":    node.Consensus.GetConsensusStats(),
			"edges":        dagStats["edges"],
		}
		nodeStates[nodeID] = nodeState
	}

	return map[string]interface{}{
		"total_nodes":    len(ns.nodes),
		"online_nodes":   onlineCount,
		"total_events":   totalEvents,
		"total_edges":    totalEdges,
		"event_log_size": len(ns.eventLog),
		"nodes":          nodeStates,
	}
}

// CheckStateConvergence sprawdza czy wszystkie węzły mają taki sam stan
func (ns *NetworkSimulator) CheckStateConvergence() bool {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	if len(ns.nodes) == 0 {
		return true
	}

	var referenceState map[string]interface{}
	referenceNodeID := ""

	for nodeID, node := range ns.nodes {
		if !node.IsOnline {
			continue
		}

		if referenceState == nil {
			referenceState = node.StateEngine.GetState()
			referenceNodeID = nodeID
			continue
		}

		currentState := node.StateEngine.GetState()

		// Compare state sizes as proxy for convergence
		if len(currentState) != len(referenceState) {
			return false
		}
	}

	return true
}

// GetNodeState zwraca stan konkretnego węzła
func (ns *NetworkSimulator) GetNodeState(nodeID string) map[string]interface{} {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	node, exists := ns.nodes[nodeID]
	if !exists {
		return nil
	}

	return map[string]interface{}{
		"online":    node.IsOnline,
		"events":    node.EventStore.Size(),
		"state":     node.StateEngine.GetState(),
		"dag_stats": node.DAGManager.GetDAGStats(),
	}
}

// SetNodeOnline zmienia status online węzła
func (ns *NetworkSimulator) SetNodeOnline(nodeID string, online bool) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	node, exists := ns.nodes[nodeID]
	if !exists {
		return fmt.Errorf("node %s not found", nodeID)
	}

	node.IsOnline = online
	return nil
}

// Shutdown zamyka simulator
func (ns *NetworkSimulator) Shutdown() error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	for _, node := range ns.nodes {
		node.EventBus.Close()
		node.GossipNode.Stop()
	}

	return nil
}

// Helper: count online nodes
func (ns *NetworkSimulator) countOnlineNodes() int {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	count := 0
	for _, node := range ns.nodes {
		if node.IsOnline {
			count++
		}
	}
	return count
}
