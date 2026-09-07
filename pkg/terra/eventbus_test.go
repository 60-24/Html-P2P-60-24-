package terra

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"github.com/60-24/p2p-60-24/pkg/core"
)

func TestEventBusPublish(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	eb := NewEventBus("Alice", priv, pub)
	defer eb.Close()
	
	event, err := eb.Publish(core.EventTypeProofOfMeeting, []byte("test"), core.PriorityNormal)
	
	if err != nil {
		t.Errorf("Publish failed: %v", err)
	}
	if event == nil {
		t.Errorf("Event should not be nil")
	}
	if event.Signature == nil || len(event.Signature) == 0 {
		t.Errorf("Event should be signed")
	}
}

func TestEventBusSubscribe(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	eb := NewEventBus("Alice", priv, pub)
	defer eb.Close()
	
	subscriber := eb.Subscribe(core.EventTypeProofOfMeeting)
	if subscriber == nil {
		t.Errorf("Subscriber should not be nil")
	}
}

func TestEventBusEventLog(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	eb := NewEventBus("Alice", priv, pub)
	defer eb.Close()
	
	eb.Publish(core.EventTypeProofOfMeeting, []byte("test1"), core.PriorityNormal)
	eb.Publish(core.EventTypeCommitment, []byte("test2"), core.PriorityHigh)
	
	log := eb.GetEventLog()
	if len(log) != 2 {
		t.Errorf("Expected 2 events in log, got %d", len(log))
	}
}

func TestEventBusVerifySignature(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	eb := NewEventBus("Alice", priv, pub)
	defer eb.Close()
	
	event, _ := eb.Publish(core.EventTypeProofOfMeeting, []byte("test"), core.PriorityNormal)
	
	if !eb.VerifySignature(event, pub) {
		t.Errorf("Signature verification failed")
	}
}

func TestEventBusMultipleSubscribers(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	eb := NewEventBus("Alice", priv, pub)
	defer eb.Close()
	
	sub1 := eb.Subscribe(core.EventTypeProofOfMeeting)
	sub2 := eb.Subscribe(core.EventTypeProofOfMeeting)
	
	eb.Publish(core.EventTypeProofOfMeeting, []byte("test"), core.PriorityNormal)
	
	// Both subscribers should receive event
	select {
	case e := <-sub1:
		if e == nil {
			t.Errorf("Subscriber 1 should receive event")
		}
	default:
		t.Errorf("Subscriber 1 did not receive event")
	}
	
	select {
	case e := <-sub2:
		if e == nil {
			t.Errorf("Subscriber 2 should receive event")
		}
	default:
		t.Errorf("Subscriber 2 did not receive event")
	}
}
