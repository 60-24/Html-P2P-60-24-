package network

import (
	"net"
	"sync"
	"time"

	"github.com/60-24/p2p-60-24/pkg/core"
	"github.com/60-24/p2p-60-24/pkg/terra"
)

// Peer reprezentuje połączenie z innym węzłem
type Peer struct {
	NodeID      string
	Address     string
	PubKey      []byte
	Connected   bool
	LastSeen    time.Time
	TrustScore  float64
	EventCount  int
}

// GossipNode implementuje Gossip Protocol + First Handshake
type GossipNode struct {
	mu          sync.RWMutex
	nodeID      string
	listenAddr  string
	listenPort  int
	peers       map[string]*Peer
	eventBus    *terra.EventBus
	conn        net.PacketConn
	stopChan    chan bool
	inFlight    map[string]*core.Event // events being gossiped
}

// NewGossipNode tworzy nowy GossipNode
func NewGossipNode(nodeID string, port int, eventBus *terra.EventBus) *GossipNode {
	return &GossipNode{
		nodeID:     nodeID,
		listenAddr: "0.0.0.0",
		listenPort: port,
		peers:      make(map[string]*Peer),
		eventBus:   eventBus,
		stopChan:   make(chan bool),
		inFlight:   make(map[string]*core.Event),
	}
}

// Start uruchamia UDP listener (port 6024)
func (gn *GossipNode) Start() error {
	gn.mu.Lock()
	defer gn.mu.Unlock()

	addr := net.UDPAddr{
		Port: gn.listenPort,
		IP:   net.ParseIP(gn.listenAddr),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return err
	}

	gn.conn = conn

	// Start background gossip routine
	go gn.gossipLoop()
	go gn.listenLoop()

	return nil
}

// ConnectToPeer implementuje First Handshake Protocol
func (gn *GossipNode) ConnectToPeer(peerAddr string) error {
	gn.mu.Lock()
	defer gn.mu.Unlock()

	// 1. UDP Discover
	udpAddr, err := net.ResolveUDPAddr("udp", peerAddr)
	if err != nil {
		return err
	}

	// 2. PING/PONG handshake
	msg := []byte("PING:" + gn.nodeID)
	_, err = gn.conn.(*net.UDPConn).WriteToUDP(msg, udpAddr)
	if err != nil {
		return err
	}

	// 3. Create peer entry
	peer := &Peer{
		NodeID:    extractNodeID(peerAddr),
		Address:   peerAddr,
		Connected: true,
		LastSeen:  time.Now(),
	}

	gn.peers[peer.NodeID] = peer

	// 4. Begin gossip
	go gn.syncWithPeer(peer)

	return nil
}

// BroadcastEvent wysyła event do wszystkich peer'ów
func (gn *GossipNode) BroadcastEvent(event *core.Event) error {
	gn.mu.Lock()
	defer gn.mu.Unlock()

	gn.inFlight[event.ID] = event

	// Roześlij do każdego connected peer'a
	for _, peer := range gn.peers {
		if peer.Connected {
			go gn.sendEventToPeer(event, peer)
		}
	}

	return nil
}

// GetPeerStats zwraca statystyki peer'ów
func (gn *GossipNode) GetPeerStats() map[string]interface{} {
	gn.mu.RLock()
	defer gn.mu.RUnlock()

	connectedCount := 0
	totalEvents := 0

	for _, peer := range gn.peers {
		if peer.Connected {
			connectedCount++
		}
		totalEvents += peer.EventCount
	}

	return map[string]interface{}{
		"node_id":          gn.nodeID,
		"total_peers":      len(gn.peers),
		"connected_peers":  connectedCount,
		"total_events_seen": totalEvents,
		"port":             gn.listenPort,
	}
}

// listenLoop słucha na przychodzące UDP pakiety
func (gn *GossipNode) listenLoop() {
	buffer := make([]byte, 4096)

	for {
		select {
		case <-gn.stopChan:
			return
		default:
		}

		gn.conn.(*net.UDPConn).SetReadDeadline(time.Now().Add(5 * time.Second))
		n, remoteAddr, err := gn.conn.(*net.UDPConn).ReadFromUDP(buffer)

		if err != nil {
			continue
		}

		// Parse message (simplified)
		msg := string(buffer[:n])
		if msg == "PING:"+gn.nodeID {
			// Send PONG
			pongMsg := []byte("PONG:" + gn.nodeID)
			gn.conn.(*net.UDPConn).WriteToUDP(pongMsg, remoteAddr)
		}
	}
}

// gossipLoop periodically broadcasts events to peers
func (gn *GossipNode) gossipLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-gn.stopChan:
			return
		case <-ticker.C:
			gn.mu.RLock()
			for _, event := range gn.inFlight {
				for _, peer := range gn.peers {
					if peer.Connected {
						go gn.sendEventToPeer(event, peer)
					}
				}
			}
			gn.mu.RUnlock()
		}
	}
}

// syncWithPeer synchronizuje state z peer'em
func (gn *GossipNode) syncWithPeer(peer *Peer) {
	// TODO: Send Event Log snapshot
	// TODO: Receive missing events
	// TODO: Update TrustScore based on consistency
}

// sendEventToPeer wysyła event do konkretnego peer'a
func (gn *GossipNode) sendEventToPeer(event *core.Event, peer *Peer) {
	// TODO: Serialize event
	// TODO: Send via UDP
	// TODO: Track delivery
}

// DisconnectPeer odłącza peer'a
func (gn *GossipNode) DisconnectPeer(nodeID string) error {
	gn.mu.Lock()
	defer gn.mu.Unlock()

	peer, exists := gn.peers[nodeID]
	if !exists {
		return nil
	}

	peer.Connected = false
	peer.LastSeen = time.Now()

	return nil
}

// Stop zamyka GossipNode
func (gn *GossipNode) Stop() error {
	close(gn.stopChan)
	if gn.conn != nil {
		gn.conn.Close()
	}
	return nil
}

// Helper: extractNodeID z adresu
func extractNodeID(addr string) string {
	// Extract hostname/IP part
	parts := make([]byte, 0)
	for i := 0; i < len(addr); i++ {
		if addr[i] == ':' {
			break
		}
		parts = append(parts, addr[i])
	}
	return string(parts)
}
