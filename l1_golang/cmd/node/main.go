package main

import (
	"log"
	"nusa-chain/internal/node"
)

func main() {
	// Load configuration
	config := node.DefaultConfig()
	
	// Create node
	n, err := node.NewNode(config)
	if err != nil {
		log.Fatal("Failed to create node:", err)
	}
	
	// Start node
	if err := n.Start(); err != nil {
		log.Fatal("Failed to start node:", err)
	}
	
	// Keep running
	select {}
}