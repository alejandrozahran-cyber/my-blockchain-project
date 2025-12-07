package blockchain

import (
	"encoding/json"
	"log"
	"sync"
)

type Blockchain struct {
	Chain        []*Block
	Difficulty   int
	PendingTXs   []Transaction
	mu           sync.RWMutex
}

func NewBlockchain(difficulty int) *Blockchain {
	// Create genesis block
	genesisBlock := NewBlock(0, []Transaction{}, "0")
	genesisBlock.MineBlock(difficulty)

	return &Blockchain{
		Chain:      []*Block{genesisBlock},
		Difficulty: difficulty,
		PendingTXs: []Transaction{},
	}
}

func (bc *Blockchain) GetLatestBlock() *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.Chain[len(bc.Chain)-1]
}

func (bc *Blockchain) AddBlock(transactions []Transaction) *Block {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	previousBlock := bc.Chain[len(bc.Chain)-1]
	newBlock := NewBlock(previousBlock.Index+1, transactions, previousBlock.Hash)
	newBlock.MineBlock(bc.Difficulty)

	bc.Chain = append(bc.Chain, newBlock)
	return newBlock
}

func (bc *Blockchain) AddTransaction(tx Transaction) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.PendingTXs = append(bc.PendingTXs, tx)
}

func (bc *Blockchain) MinePendingTransactions(validator string) *Block {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if len(bc.PendingTXs) == 0 {
		return nil
	}

	// Get pending transactions
	pendingTXs := make([]Transaction, len(bc.PendingTXs))
	copy(pendingTXs, bc.PendingTXs)

	// Create new block
	previousBlock := bc.Chain[len(bc.Chain)-1]
	newBlock := NewBlock(previousBlock.Index+1, pendingTXs, previousBlock.Hash)
	newBlock.Validator = validator
	newBlock.MineBlock(bc.Difficulty)

	// Add to chain and clear pending
	bc.Chain = append(bc.Chain, newBlock)
	bc.PendingTXs = []Transaction{}

	log.Printf("✅ Block #%d mined by %s", newBlock.Index, validator)
	return newBlock
}

func (bc *Blockchain) IsChainValid() bool {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	for i := 1; i < len(bc.Chain); i++ {
		currentBlock := bc.Chain[i]
		previousBlock := bc.Chain[i-1]

		// Check current block hash
		if currentBlock.Hash != currentBlock.CalculateHash() {
			return false
		}

		// Check previous block hash
		if currentBlock.PreviousHash != previousBlock.Hash {
			return false
		}

		// Check if block is properly mined
		if !currentBlock.IsValid() {
			return false
		}
	}

	return true
}

func (bc *Blockchain) GetChainJSON() (string, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	bytes, err := json.MarshalIndent(bc.Chain, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (bc *Blockchain) GetBlockCount() int {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return len(bc.Chain)
}

func (bc *Blockchain) GetPendingTXCount() int {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return len(bc.PendingTXs)
}