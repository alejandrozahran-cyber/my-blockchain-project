package node

import (
	"fmt"
	"log"
	"sync"
	"time"

	"nusa-chain/internal/blockchain"
	"nusa-chain/internal/consensus"
	"nusa-chain/internal/wallet"
)

type NUSANode struct {
	Config      *Config
	Blockchain  *blockchain.Blockchain
	Wallet      *wallet.Wallet
	PoVC        *consensus.PoVConsensus
	IsMining    bool
	mu          sync.RWMutex
	stopMining  chan bool
}

func NewNode(cfg *Config) (*NUSANode, error) {
	// Initialize blockchain
	bc := blockchain.NewBlockchain(4) // Difficulty 4

	// Generate or load wallet
	var w *wallet.Wallet
	var err error
	
	// For demo, always generate new wallet
	w, err = wallet.NewWallet()
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %v", err)
	}

	// Initialize PoVC consensus
	povc := consensus.NewPoVConsensus(
		25000000,  // Total supply
		100000,    // Monthly reward
		cfg.AIEngine.URL,
	)

	node := &NUSANode{
		Config:     cfg,
		Blockchain: bc,
		Wallet:     w,
		PoVC:       povc,
		IsMining:   false,
		stopMining: make(chan bool),
	}

	log.Printf("✅ Node initialized")
	log.Printf("   Address: %s", w.Address.Hex())
	log.Printf("   Chain: %s (ID: %d)", cfg.Network.Name, cfg.Network.ChainID)
	log.Printf("   AI Engine: %s", cfg.AIEngine.URL)

	return node, nil
}

func (n *NUSANode) Start() error {
	log.Println("🚀 Starting NUSA Node...")

	// Start API server
	go n.startAPIServer()

	// Start mining if enabled
	if n.Config.Consensus.Type == "PoVC" {
		go n.startMining()
	}

	log.Println("✅ NUSA Node started successfully")
	log.Printf("📡 API: http://%s:%d", n.Config.API.Host, n.Config.API.Port)
	log.Printf("🔗 P2P: %s", n.Config.P2P.ListenAddr)

	// Keep node running
	select {}
}

func (n *NUSANode) startMining() {
	n.mu.Lock()
	n.IsMining = true
	n.mu.Unlock()

	log.Println("⛏️  Starting PoVC mining...")

	ticker := time.NewTicker(time.Duration(n.Config.Consensus.BlockTime) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			n.mineBlock()
		case <-n.stopMining:
			n.mu.Lock()
			n.IsMining = false
			n.mu.Unlock()
			log.Println("⛏️  Mining stopped")
			return
		}
	}
}

func (n *NUSANode) mineBlock() {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Check if there are pending transactions
	if n.Blockchain.GetPendingTXCount() == 0 {
		return
	}

	// Mine pending transactions
	block := n.Blockchain.MinePendingTransactions(n.Wallet.Address.Hex())
	if block != nil {
		log.Printf("✅ Mined block #%d with %d transactions",
			block.Index, len(block.Transactions))
	}
}

func (n *NUSANode) startAPIServer() {
	// API server implementation
	// This would start the HTTP server on Config.API.Port
	log.Printf("🌐 API server starting on :%d", n.Config.API.Port)
	// Implementation would go here
}

func (n *NUSANode) Stop() {
	log.Println("🛑 Stopping NUSA Node...")

	// Stop mining
	if n.IsMining {
		n.stopMining <- true
	}

	log.Println("✅ NUSA Node stopped")
}

func (n *NUSANode) GetStatus() map[string]interface{} {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return map[string]interface{}{
		"status":        "running",
		"node_address":  n.Wallet.Address.Hex(),
		"block_height":  n.Blockchain.GetBlockCount(),
		"pending_txs":   n.Blockchain.GetPendingTXCount(),
		"is_mining":     n.IsMining,
		"chain_id":      n.Config.Network.ChainID,
		"network":       n.Config.Network.Name,
		"ai_engine":     n.Config.AIEngine.Enabled,
		"consensus":     n.Config.Consensus.Type,
		"timestamp":     time.Now().Unix(),
	}
}

func (n *NUSANode) CreateTransaction(to string, value float64, data string) (string, error) {
	// Create and sign transaction
	// This is simplified - actual implementation would handle nonce, gas, etc.
	
	tx := blockchain.Transaction{
		Hash:      fmt.Sprintf("0x%x", time.Now().UnixNano()),
		From:      n.Wallet.Address.Hex(),
		To:        to,
		Value:     value,
		Timestamp: time.Now().Unix(),
		Data:      data,
	}

	// Sign transaction
	signature, err := n.Wallet.Sign([]byte(tx.Hash))
	if err != nil {
		return "", err
	}
	tx.Signature = signature

	// Add to pending transactions
	n.Blockchain.AddTransaction(tx)

	return tx.Hash, nil
}