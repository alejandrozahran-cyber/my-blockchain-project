package consensus

import (
	"time"
	"math/rand"
	"sync"
	"fmt"
	"encoding/json"
	"net/http"
	
	"nusa-chain/internal/blockchain"
)

type PoVCReal struct {
	chainManager *blockchain.ChainManager
	aiEngineURL  string
	validators   map[string]Validator
	mutex        sync.RWMutex
	isProducing  bool
	stopChan     chan bool
}

type Validator struct {
	Address     string `json:"address"`
	Stake       uint64 `json:"stake"`
	NVSScore    float64 `json:"nvs_score"`
	LastActive  int64  `json:"last_active"`
	IsActive    bool   `json:"is_active"`
}

func NewPoVCReal(chainManager *blockchain.ChainManager, aiEngineURL string) *PoVCReal {
	return &PoVCReal{
		chainManager: chainManager,
		aiEngineURL:  aiEngineURL,
		validators:   make(map[string]Validator),
		stopChan:     make(chan bool),
	}
}

// Start block production
func (p *PoVCReal) Start() {
	p.isProducing = true
	go p.blockProductionLoop()
}

// Stop block production
func (p *PoVCReal) Stop() {
	p.isProducing = false
	p.stopChan <- true
}

func (p *PoVCReal) blockProductionLoop() {
	ticker := time.NewTicker(5 * time.Second) // 5 second block time
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			if p.shouldProduceBlock() {
				p.produceBlock()
			}
		case <-p.stopChan:
			return
		}
	}
}

func (p *PoVCReal) shouldProduceBlock() bool {
	// Check if we're a validator and it's our turn
	latestBlock := p.chainManager.GetLatestBlock()
	if latestBlock == nil {
		return true // Genesis block
	}
	
	// Simple round-robin validator selection
	validators := p.getActiveValidators()
	if len(validators) == 0 {
		return true // No validators, produce anyway
	}
	
	// Get our validator index
	myAddress := "validator_address" // TODO: Get from wallet
	for i, validator := range validators {
		if validator.Address == myAddress {
			// Check if it's our turn based on block height
			return (latestBlock.Header.Height+1)%uint64(len(validators)) == uint64(i)
		}
	}
	
	return false
}

func (p *PoVCReal) produceBlock() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	latestBlock := p.chainManager.GetLatestBlock()
	height := uint64(0)
	prevHash := "0"
	
	if latestBlock != nil {
		height = latestBlock.Header.Height + 1
		prevHash = latestBlock.Hash()
	}
	
	// Get pending transactions
	pendingTXs := p.chainManager.GetPendingTXs()
	
	// Limit transactions per block
	maxTXs := 100
	if len(pendingTXs) > maxTXs {
		pendingTXs = pendingTXs[:maxTXs]
	}
	
	// Create new block
	newBlock := blockchain.NewBlock(
		height,
		prevHash,
		pendingTXs,
		"validator_address", // TODO: Get from wallet
	)
	
	// Apply PoVC adjustments based on AI Engine
	p.applyPoVCRewards(newBlock)
	
	// Add block to chain
	if err := p.chainManager.AddBlock(newBlock); err != nil {
		fmt.Printf("❌ Failed to add block: %v\n", err)
		return
	}
	
	fmt.Printf("✅ Produced block #%d with %d transactions\n", height, len(pendingTXs))
}

func (p *PoVCReal) applyPoVCRewards(block *blockchain.Block) {
	// Get NVS scores from AI Engine
	nvsScores := p.getNVSScoresFromAI(block)
	
	// Adjust block reward based on validator's NVS score
	if len(nvsScores) > 0 {
		validatorScore := nvsScores[block.Header.Validator]
		if validatorScore > 0 {
			// Reward multiplier based on NVS score
			multiplier := validatorScore // 0-1 scale
			adjustedReward := uint64(float64(block.Header.Reward) * multiplier)
			block.Header.Reward = adjustedReward
		}
	}
	
	// Apply anti-whale penalties
	p.applyAntiWhalePenalties(block)
}

func (p *PoVCReal) getNVSScoresFromAI(block *blockchain.Block) map[string]float64 {
	// Call AI Engine to get NVS scores
	// This is simplified - in production, batch the request
	
	client := &http.Client{Timeout: 10 * time.Second}
	
	// Prepare request data
	requestData := map[string]interface{}{
		"block_height": block.Header.Height,
		"validators":   p.getActiveValidators(),
	}
	
	jsonData, _ := json.Marshal(requestData)
	
	resp, err := client.Post(p.aiEngineURL+"/povc/batch", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("⚠️  AI Engine not available: %v\n", err)
		return make(map[string]float64)
	}
	defer resp.Body.Close()
	
	var result struct {
		Success bool `json:"success"`
		Data    struct {
			Scores map[string]float64 `json:"scores"`
		} `json:"data"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return make(map[string]float64)
	}
	
	return result.Data.Scores
}

func (p *PoVCReal) applyAntiWhalePenalties(block *blockchain.Block) {
	// Check if validator is a whale
	for _, tx := range block.Transactions {
		senderBalance := p.chainManager.GetBalance(tx.From)
		totalSupply := uint64(25000000 * 1000000000000000000) // 25M NUSA in wei
		
		percentage := float64(senderBalance) / float64(totalSupply) * 100
		
		// Apply penalties for whales
		if percentage > 2 {
			// Whale detected - reduce reward
			block.Header.Reward = 0
			fmt.Printf("⚠️  Whale detected: %s (%f%% of supply)\n", tx.From, percentage)
		} else if percentage > 0.5 {
			// Warning zone - partial penalty
			penalty := uint64(float64(block.Header.Reward) * 0.5)
			block.Header.Reward -= penalty
		}
	}
}

func (p *PoVCReal) getActiveValidators() []Validator {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	var active []Validator
	for _, validator := range p.validators {
		if validator.IsActive && validator.Stake > 0 {
			active = append(active, validator)
		}
	}
	return active
}

// Register validator
func (p *PoVCReal) RegisterValidator(address string, stake uint64) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	// Get NVS score from AI Engine
	nvsScore, err := p.getValidatorScore(address)
	if err != nil {
		return err
	}
	
	p.validators[address] = Validator{
		Address:    address,
		Stake:      stake,
		NVSScore:   nvsScore,
		LastActive: time.Now().Unix(),
		IsActive:   true,
	}
	
	return nil
}

func (p *PoVCReal) getValidatorScore(address string) (float64, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	
	requestData := map[string]interface{}{
		"wallet_address": address,
		"wallet_balance": p.chainManager.GetBalance(address),
	}
	
	jsonData, _ := json.Marshal(requestData)
	
	resp, err := client.Post(p.aiEngineURL+"/povc/calculate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0.5, err // Default score
	}
	defer resp.Body.Close()
	
	var result struct {
		Success bool `json:"success"`
		Data    struct {
			NVSScore float64 `json:"nvs_score"`
		} `json:"data"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0.5, err
	}
	
	return result.Data.NVSScore, nil
}

// Select validator for next block
func (p *PoVCReal) SelectValidator() string {
	validators := p.getActiveValidators()
	if len(validators) == 0 {
		return ""
	}
	
	// Weighted selection based on NVS score
	var totalWeight float64
	for _, v := range validators {
		totalWeight += v.NVSScore
	}
	
	if totalWeight == 0 {
		// Random selection if no scores
		return validators[rand.Intn(len(validators))].Address
	}
	
	// Weighted random selection
	r := rand.Float64() * totalWeight
	var cumulative float64
	for _, v := range validators {
		cumulative += v.NVSScore
		if r <= cumulative {
			return v.Address
		}
	}
	
	return validators[len(validators)-1].Address
}