package terra

import (
	"crypto/ed25519"
	"sync"

	"github.com/60-24/p2p-60-24/pkg/core"
)

// EventBus implementuje Pub/Sub pattern dla Events
type EventBus struct {
	mu          sync.RWMutex
	events      chan *core.Event
	subscribers map[string][]chan *core.Event
	privKey     ed25519.PrivateKey
	pubKey      ed25519.PublicKey
	nodeID      string
	eventLog    []*core.Event
}

// NewEventBus tworzy nowy EventBus
func NewEventBus(nodeID string, privKey ed25519.PrivateKey, pubKey ed25519.PublicKey) *EventBus {
	return &EventBus{
		events:      make(chan *core.Event, 1000),
		subscribers: make(map[string][]chan *core.Event),
		privKey:     privKey,
		pubKey:      pubKey,
		nodeID:      nodeID,
		eventLog:    make([]*core.Event, 0),
	}
}

// Publish publikuje event
func (eb *EventBus) Publish(eventType string, payload []byte, priority int) (*core.Event, error) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	event := core.NewEvent(eb.nodeID, eventType, payload, priority)
	event.Signature = ed25519.Sign(eb.privKey, []byte(event.Hash))
	eb.eventLog = append(eb.eventLog, event)
	eb.events <- event
	eb.broadcastToSubscribers(event)

	return event, nil
}

// Subscribe subskrybuje na zdarzenia
func (eb *EventBus) Subscribe(eventType string) <-chan *core.Event {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	subscriber := make(chan *core.Event, 100)
	eb.subscribers[eventType] = append(eb.subscribers[eventType], subscriber)
	return subscriber
}

// GetEventLog zwraca event log
func (eb *EventBus) GetEventLog() []*core.Event {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	log := make([]*core.Event, len(eb.eventLog))
	copy(log, eb.eventLog)
	return log
}

// VerifySignature sprawdza podpis
func (eb *EventBus) VerifySignature(event *core.Event, senderPubKey ed25519.PublicKey) bool {
	return ed25519.Verify(senderPubKey, []byte(event.Hash), event.Signature)
}

func (eb *EventBus) broadcastToSubscribers(event *core.Event) {
	if subs, exists := eb.subscribers[event.EventType]; exists {
		for _, sub := range subs {
			select {
			case sub <- event:
			default:
			}
		}
	}
}

func (eb *EventBus) Close() {
	close(eb.events)
}
