package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// Event reprezentuje atomowe zdarzenie w systemie
// = Najważniejsza struktura — wszystko opiera się na Events
type Event struct {
	ID        string    `json:"id"`        // Unikalny ID zdarzenia
	Timestamp int64     `json:"timestamp"` // Unix timestamp (nanosekund)
	Sender    string    `json:"sender"`    // Kto to zdarzenie opublikował
	EventType string    `json:"type"`      // meeting, commitment, resolution, etc
	Payload   []byte    `json:"payload"`   // Dane zdarzenia (JSON bytes)
	Signature []byte    `json:"signature"` // Ed25519 signature
	Priority  int       `json:"priority"`  // 0=Low, 1=Normal, 2=High, 3=Critical
	TTL       int64     `json:"ttl"`       // Time to live (sekundy, 0 = forever)
	Hash      string    `json:"hash"`      // SHA256 hash zdarzenia
}

// EventMetadata zawiera meta-informacje o zdarzeniu
type EventMetadata struct {
	CreatedAt    time.Time
	ProcessedAt  time.Time
	Verified     bool
	VerifiedBy   []string // Które węzły potwierdziły
	Replicators  int      // Ile węzłów ma kopię
}

// Priority levels
const (
	PriorityLow      = 0
	PriorityNormal   = 1
	PriorityHigh     = 2
	PriorityCritical = 3
)

// EventType constants
const (
	EventTypeProofOfMeeting = "proof_of_meeting"
	EventTypeCommitment     = "commitment"
	EventTypeResolution     = "resolution"
	EventTypeReputationUpdate = "reputation_update"
	EventTypeNodeJoined     = "node_joined"
	EventTypeNodeLeft       = "node_left"
)

// NewEvent tworzy nowe zdarzenie
func NewEvent(sender, eventType string, payload []byte, priority int) *Event {
	now := time.Now()
	
	e := &Event{
		ID:        generateEventID(sender, now),
		Timestamp: now.UnixNano(),
		Sender:    sender,
		EventType: eventType,
		Payload:   payload,
		Priority:  priority,
		TTL:       0, // forever
	}
	
	// Oblicz hash (przed podpisaniem)
	e.Hash = e.ComputeHash()
	
	return e
}

// ComputeHash oblicza SHA256 hash zdarzenia
func (e *Event) ComputeHash() string {
	// Hash obejmuje: ID + Timestamp + Sender + Type + Payload
	data := fmt.Sprintf("%s|%d|%s|%s|%s",
		e.ID, e.Timestamp, e.Sender, e.EventType,
		hex.EncodeToString(e.Payload))
	
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// Verify sprawdza czy hash się zgadza
func (e *Event) Verify() bool {
	return e.Hash == e.ComputeHash()
}

// ToJSON serializuje event do JSON
func (e *Event) ToJSON() ([]byte, error) {
	return json.MarshalIndent(e, "", "  ")
}

// FromJSON deserializuje event z JSON
func FromJSON(data []byte) (*Event, error) {
	var e Event
	err := json.Unmarshal(data, &e)
	return &e, err
}

// generateEventID tworzy unikalny ID
func generateEventID(sender string, now time.Time) string {
	ts := fmt.Sprintf("%d", now.UnixNano())
	hash := sha256.Sum256([]byte(sender + ts))
	return hex.EncodeToString(hash[:8]) // 16 chars
}

// String implementuje Stringer interface
func (e *Event) String() string {
	return fmt.Sprintf("Event{ID:%s, Type:%s, Sender:%s, Priority:%d, TS:%d}",
		e.ID, e.EventType, e.Sender, e.Priority, e.Timestamp)
}
