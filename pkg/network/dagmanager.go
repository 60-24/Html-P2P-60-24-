package network

import (
	"sync"
	"time"

	"github.com/60-24/p2p-60-24/pkg/core"
	"github.com/60-24/p2p-60-24/pkg/data"
)

// Edge reprezentuje relację zaufania (RealBond)
type Edge struct {
	From         string    // NodeID of sender
	To           string    // NodeID of receiver
	Weight       float64   // TrustMetric [0, 1]
	ProofHash    string    // Proof of Meeting hash
	LastUpdated  int64     // Unix timestamp
	EventCount   int       // Liczba zdarzeń between nodes
	Confidence   float64   // Pewność [0, 1]
}

// DunbarCircle reprezentuje kręgi Dunbar
type DunbarCircle struct {
	Radius   int
	MaxPeers int
	Peers    map[string]float64 // NodeID -> TrustScore
}

// DAGManager zarządza RealBond DAG (Directed Acyclic Graph)
type DAGManager struct {
	mu          sync.RWMutex
	edges       map[string]map[string]*Edge // From -> To -> Edge
	nodes       map[string]*DunbarCircle    // NodeID -> Dunbar circles
	eventStore  *data.EventStore
	nodeID      string
	metrics     map[string]interface{}
}

// NewDAGManager tworzy nowy DAGManager
func NewDAGManager(nodeID string, eventStore *data.EventStore) *DAGManager {
	return &DAGManager{
		edges:      make(map[string]map[string]*Edge),
		nodes:      make(map[string]*DunbarCircle),
		eventStore: eventStore,
		nodeID:     nodeID,
		metrics:    make(map[string]interface{}),
	}
}

// AddOrUpdateEdge dodaje/aktualizuje relację zaufania
func (dm *DAGManager) AddOrUpdateEdge(from, to string, proofHash string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// Znormalizuj key
	if _, exists := dm.edges[from]; !exists {
		dm.edges[from] = make(map[string]*Edge)
	}

	// Oblicz TrustMetric
	trustScore := dm.calculateTrustMetric(from, to)

	edge := &Edge{
		From:        from,
		To:          to,
		Weight:      trustScore,
		ProofHash:   proofHash,
		LastUpdated: time.Now().UnixNano(),
		EventCount:  1,
		Confidence:  0.95, // Proof of Meeting = wysoka pewność
	}

	// Jeśli edge już istnieje, increment EventCount
	if existing, exists := dm.edges[from][to]; exists {
		edge.EventCount = existing.EventCount + 1
		edge.Weight = dm.calculateTrustMetric(from, to) // Recalculate
	}

	dm.edges[from][to] = edge

	// Update Dunbar circle
	dm.updateDunbarCircle(from, to, trustScore)

	return nil
}

// GetTrustScore zwraca TrustMetric między dwoma węzłami
func (dm *DAGManager) GetTrustScore(from, to string) float64 {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	if fromEdges, exists := dm.edges[from]; exists {
		if edge, exists := fromEdges[to]; exists {
			return edge.Weight
		}
	}

	return 0.0 // Brak relacji = zero trust
}

// CalculateTrustMetric oblicza TrustMetric (formula z OmniKernel)
func (dm *DAGManager) calculateTrustMetric(from, to string) float64 {
	// Formula: TrustMetric = Σ(Events) × Decay(time) × ProfileWeight
	// Uproszczone: base 0.5 + (eventCount * 0.1) + (recency bonus)

	eventCount := 0
	if fromEdges, exists := dm.edges[from]; exists {
		if edge, exists := fromEdges[to]; exists {
			eventCount = edge.EventCount
		}
	}

	base := 0.5
	eventBonus := float64(eventCount) * 0.1
	recencyBonus := 0.1 // Freshness bonus

	score := base + eventBonus + recencyBonus

	if score > 1.0 {
		score = 1.0
	}

	return score
}

// BuildQuorum assembluje Quorum dla mediacji
// Formula: max(3, min(5, floor(N/2)))
func (dm *DAGManager) BuildQuorum(requester string) []string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	// Get Dunbar circle (Kręgi 1+2 = najbliższe osoby)
	circle1 := 5
	circle2 := 15

	// Collect trusted peers
	var candidates []string
	if circle, exists := dm.nodes[requester]; exists {
		count := 0
		for peer := range circle.Peers {
			candidates = append(candidates, peer)
			count++
			if count >= circle1+circle2 {
				break
			}
		}
	}

	// Quorum size: max(3, min(5, floor(N/2)))
	quorumSize := 3
	if len(candidates) > 6 {
		quorumSize = 5
	} else if len(candidates) > 2 {
		quorumSize = (len(candidates) / 2)
		if quorumSize > 5 {
			quorumSize = 5
		}
		if quorumSize < 3 {
			quorumSize = 3
		}
	}

	// Select top trusted
	if len(candidates) > quorumSize {
		candidates = candidates[:quorumSize]
	}

	return candidates
}

// VerifyProofOfMeeting sprawdza Proof of Meeting
func (dm *DAGManager) VerifyProofOfMeeting(event *core.Event) bool {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	// Verify event hash
	if !event.Verify() {
		return false
	}

	// Check signatures (TODO: when we have crypto)
	if event.Signature == nil || len(event.Signature) == 0 {
		return false
	}

	return true
}

// GetReputationScore oblicza reputację (ILOCZYN, nie suma!)
// Rep = Consistency × Trustworthiness × Knowledge × Experience
func (dm *DAGManager) GetReputationScore(nodeID string) float64 {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	// Count outgoing edges (relationships)
	outgoing := len(dm.edges[nodeID])
	if outgoing == 0 {
		return 0.0
	}

	// Collect trust scores from all relationships
	var scores []float64
	if edges, exists := dm.edges[nodeID]; exists {
		for _, edge := range edges {
			scores = append(scores, edge.Weight)
		}
	}

	// Reputation = multiplicative model
	// If ANY score is low, whole reputation tanks
	rep := 1.0
	for _, score := range scores {
		rep *= score
	}

	return rep
}

// GetDAGStats zwraca statystyki DAG
func (dm *DAGManager) GetDAGStats() map[string]interface{} {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	edgeCount := 0
	totalWeight := 0.0

	for _, fromEdges := range dm.edges {
		for _, edge := range fromEdges {
			edgeCount++
			totalWeight += edge.Weight
		}
	}

	avgWeight := 0.0
	if edgeCount > 0 {
		avgWeight = totalWeight / float64(edgeCount)
	}

	return map[string]interface{}{
		"nodes":       len(dm.nodes),
		"edges":       edgeCount,
		"avg_weight":  avgWeight,
		"dunbar_1":    5,
		"dunbar_2":    15,
		"quorum_min":  3,
		"quorum_max":  5,
	}
}

// updateDunbarCircle aktualizuje kręgi Dunbar
func (dm *DAGManager) updateDunbarCircle(from, to string, trustScore float64) {
	if _, exists := dm.nodes[from]; !exists {
		dm.nodes[from] = &DunbarCircle{
			Radius:   1,
			MaxPeers: 150,
			Peers:    make(map[string]float64),
		}
	}

	circle := dm.nodes[from]

	// Add to peers if not at capacity
	if len(circle.Peers) < circle.MaxPeers {
		circle.Peers[to] = trustScore
	} else {
		// Replace if new score is higher
		if existing, exists := circle.Peers[to]; !exists || trustScore > existing {
			circle.Peers[to] = trustScore
		}
	}
}

// DetectCycles sprawdza czy DAG zawiera cykle (do debugowania)
func (dm *DAGManager) DetectCycles() [][]string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	// TODO: DFS cycle detection
	// Dla teraz: empty (DAG assumption)
	return make([][]string, 0)
}
