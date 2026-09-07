package main

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/60-24/p2p-60-24/pkg/core"
	"github.com/60-24/p2p-60-24/pkg/data"
	"github.com/60-24/p2p-60-24/pkg/sec"
	"github.com/60-24/p2p-60-24/pkg/terra"
)

func main() {
	separator := strings.Repeat("=", 50)
	
	fmt.Println("🚀 P2P 60-24 Event System — Demo")
	fmt.Println(separator)

	// 1. KeyManager — Generuj klucze
	fmt.Println("\n📋 KROK 1: KeyManager — Generowanie kluczy")
	km, err := sec.NewKeyManager("Alice")
	if err != nil {
		log.Fatalf("Failed to create KeyManager: %v", err)
	}
	fmt.Printf("✓ Klucz publiczny: %s...\n", km.GetPubKeyHex()[:16])
	fmt.Printf("✓ NodeID: Alice\n")

	// 2. EventBus — Pub/Sub system
	fmt.Println("\n📋 KROK 2: EventBus — Publikacja zdarzeń")
	eb := terra.NewEventBus("Alice", km.PrivKey, km.PubKey)
	defer eb.Close()

	// Publikuj Event #1: Proof of Meeting
	payload1 := map[string]interface{}{
		"participant_1": "Alice",
		"participant_2": "Bob",
		"duration":      45,
		"topic":         "pożyczka 5000 PLN",
	}
	p1, _ := json.Marshal(payload1)

	e1, _ := eb.Publish(core.EventTypeProofOfMeeting, p1, core.PriorityHigh)
	fmt.Printf("✓ Event #1 opublikowany: %s\n", e1.ID)

	// Publikuj Event #2: Commitment
	payload2 := map[string]interface{}{
		"type":     "loan",
		"amount":   5000,
		"currency": "PLN",
		"duration": "6 months",
	}
	p2, _ := json.Marshal(payload2)

	e2, _ := eb.Publish(core.EventTypeCommitment, p2, core.PriorityNormal)
	fmt.Printf("✓ Event #2 opublikowany: %s\n", e2.ID)

	// 3. EventStore — Immutable log
	fmt.Println("\n📋 KROK 3: EventStore — Zapis na stałe")
	store := data.NewEventStore()

	for _, evt := range eb.GetEventLog() {
		store.Append(evt)
	}
	fmt.Printf("✓ %d zdarzeń zapisanych w EventStore\n", store.Size())

	// 4. StateEngine — Obliczenie stanu
	fmt.Println("\n📋 KROK 4: StateEngine — Aplikacja eventów → stan")
	engine := data.NewStateEngine(store)

	for _, evt := range store.GetAll() {
		engine.ApplyEvent(evt)
	}

	state := engine.GetState()
	fmt.Printf("✓ Stan obliczony z %d eventów\n", len(state))

	// 5. Weryfikacja — Proof of Meeting
	fmt.Println("\n📋 KROK 5: Weryfikacja — Proof of Meeting")
	allEvents := store.GetByType(core.EventTypeProofOfMeeting)
	fmt.Printf("✓ Znaleziono %d spotkań w EventStore\n", len(allEvents))

	if len(allEvents) > 0 {
		meeting := allEvents[0]
		fmt.Printf("✓ Spotkanie: %s ↔ %s\n",
			state["participant_1"], state["participant_2"])
	}

	// 6. TrustMetric — Symulacja obliczenia
	fmt.Println("\n📋 KROK 6: TrustMetric — Obliczenie zaufania")
	trustScore := calculateTrustMetric(len(allEvents), store)
	fmt.Printf("✓ Trust Score (Alice → Bob): %.0f%%\n", trustScore*100)

	// Podsumowanie
	fmt.Println("\n" + separator)
	fmt.Println("✅ SESJA #2 — EVENT SYSTEM WORKS!")
	fmt.Println(separator)
	fmt.Printf("\nStatystyka:\n")
	fmt.Printf("  • Total Events: %d\n", store.Size())
	fmt.Printf("  • Proof of Meetings: %d\n", len(store.GetByType(core.EventTypeProofOfMeeting)))
	fmt.Printf("  • Commitments: %d\n", len(store.GetByType(core.EventTypeCommitment)))
	fmt.Printf("  • State Keys: %d\n", len(state))
	fmt.Printf("  • Trust Score: %.0f%%\n\n", trustScore*100)
}

// calculateTrustMetric uproszczone obliczenie zaufania
func calculateTrustMetric(eventCount int, store *data.EventStore) float64 {
	base := 0.5
	eventBonus := float64(eventCount) * 0.1

	trustScore := base + eventBonus
	if trustScore > 1.0 {
		trustScore = 1.0
	}

	return trustScore
}
