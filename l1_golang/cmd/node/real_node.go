package main

import (
	"log"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	
	"nusa-chain/internal/blockchain"
	"nusa-chain/internal/consensus"
	"nusa-chain/internal/p2p"
	"nusa-chain/internal/wallet"
)

func main() {
	fmt.Println("🚀 NUSA CHAIN - REAL BLOCKCHAIN NODE")
	fmt.Println("======================================")
	
	// Load configuration
	config := loadConfig()
	
	// Initialize wallet
	w, err := wallet.NewWallet()
	if err != nil {
		log.Fatal("Failed to create wallet:", err)
	}
	fmt.Printf("👛 Node Wallet: %s\n", w.Address.Hex())
	
	// Initialize blockchain
	chainConfig := blockchain.ChainConfig{
		ChainID:    2024,
		BlockTime:  5,
		Difficulty: 1000000,
		MaxGasLimit: 8000000,
		MinGasPrice: 1000000000, // 1 gwei
		BlockReward: 2000000000000000000, // 2 NUSA
		GenesisAccounts: []blockchain.GenesisAccount{
			{
				Address: w.Address.Hex(),
				Balance: 100000000000000000000, // 100 NUSA
				Stake:   10000000000000000000,  // 10 NUSA
			},
		},
	}
	
	chainManager, err := blockchain.NewChainManager(chainConfig)
	if err != nil {
		log.Fatal("Failed to initialize blockchain:", err)
	}
	
	// Initialize PoVC consensus
	povc := consensus.NewPoVCReal(chainManager, "http://localhost:8000")
	
	// Register as validator
	if err := povc.RegisterValidator(w.Address.Hex(), 10000000000000000000); err != nil {
		fmt.Printf("⚠️  Failed to register as validator: %v\n", err)
	}
	
	// Initialize P2P network
	p2pNetwork, err := p2p.NewP2PNetwork(30303, []string{
		"/ip4/127.0.0.1/tcp/30303/p2p/12D3KooWTest", //