package sec

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
)

// KeyManager zarządza kluczami Ed25519
type KeyManager struct {
	PrivKey ed25519.PrivateKey
	PubKey  ed25519.PublicKey
	NodeID  string
}

// NewKeyManager generuje nową parę kluczy
func NewKeyManager(nodeID string) (*KeyManager, error) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	return &KeyManager{
		PrivKey: privKey,
		PubKey:  pubKey,
		NodeID:  nodeID,
	}, nil
}

// Sign podpisuje dane
func (km *KeyManager) Sign(data []byte) []byte {
	return ed25519.Sign(km.PrivKey, data)
}

// Verify sprawdza podpis
func (km *KeyManager) Verify(data []byte, sig []byte) bool {
	return ed25519.Verify(km.PubKey, data, sig)
}

// GetPubKeyHex zwraca klucz publiczny jako hex
func (km *KeyManager) GetPubKeyHex() string {
	return hex.EncodeToString(km.PubKey)
}

// GetPrivKeyHex zwraca klucz prywatny jako hex
func (km *KeyManager) GetPrivKeyHex() string {
	return hex.EncodeToString(km.PrivKey)
}

// SaveToFile zapisuje klucze do pliku
func (km *KeyManager) SaveToFile(filepath string) error {
	data := map[string]string{
		"node_id":   km.NodeID,
		"pub_key":   km.GetPubKeyHex(),
		"priv_key":  km.GetPrivKeyHex(),
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, jsonData, 0600)
}

// LoadFromFile wczytuje klucze z pliku
func LoadFromFile(filepath string) (*KeyManager, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var keyData map[string]string
	if err := json.Unmarshal(data, &keyData); err != nil {
		return nil, err
	}

	pubKeyHex := keyData["pub_key"]
	privKeyHex := keyData["priv_key"]
	nodeID := keyData["node_id"]

	pubKeyBytes, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode pub key: %w", err)
	}

	privKeyBytes, err := hex.DecodeString(privKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode priv key: %w", err)
	}

	pubKey := ed25519.PublicKey(pubKeyBytes)
	privKey := ed25519.PrivateKey(privKeyBytes)

	return &KeyManager{
		PubKey:  pubKey,
		PrivKey: privKey,
		NodeID:  nodeID,
	}, nil
}
