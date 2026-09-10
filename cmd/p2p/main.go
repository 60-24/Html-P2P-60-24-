// Command p2p is the entry point for the P2P 60-24 OneClick Evo CLI,
// implementing the strategy from ADR-008.
package main

import (
	"github.com/60-24/p2p-60-24/cmd/p2p/commands"
)

func main() {
	commands.Execute()
}
