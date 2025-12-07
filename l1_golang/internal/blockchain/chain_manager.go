package blockchain

import (
	"sync"
	"time"
	"fmt"
	"os"
	"encoding/json"
)

type ChainManager struct {
	Chain         []*Block
	PendingTXs    []Transaction
	State         map[string]AccountState
	mutex         sync.RWMutex
	blockchainDB  *os.File
	mempoolDB     *os.File
	config        ChainConfig
}

type AccountState struct {
	Balance    uint64 `json:"balance"`
	Nonce      uint64 `json:"nonce"`
	Stake      uint64 `json:"stake,omitempty"`
	LastActive int64  `json:"last_active"`
}

type ChainConfig struct {
	ChainID         uint64 `json:"chain_id"`
	BlockTime       uint64 `json:"block_time"` // in seconds
	Difficulty      uint64 `json:"difficulty"`
	MaxGasLimit     uint64 `json:"max_gas_limit"`
	MinGasPrice     uint64 `json:"min_gas_price"`
	BlockReward     uint64 `json:"block_reward"`
	GenesisAccounts []GenesisAccount `json:"genesis_accounts"`
}

type GenesisAccount struct {
	Address string `json:"address"`
	Balance uint64 `json:"balance"`
	Stake   uint64 `json:"stake,omitempty"`
}

func NewChainManager(config ChainConfig) (*ChainManager, error) {
	cm := &ChainManager{
		Chain:   []*Block{},
		State:   make(map[string]AccountState),
		PendingTXs: []Transaction{},
		config:  config,
	}
	
	// Initialize genesis block
	genesisBlock := createGenesisBlock(config)
	cm.Chain = append(cm.Chain, genesisBlock)
	
	// Initialize genesis accounts
	for _, acc := range config.GenesisAccounts {
		cm.State[acc.Address] = AccountState{
			Balance:    acc.Balance,
			Nonce:      0,
			Stake:      acc.Stake,
			LastActive: time.Now().Unix(),
		}
	}
	
	return cm, nil
}

func createGenesisBlock(config ChainConfig) *Block {
	// Create genesis transactions from genesis accounts
	var genesisTXs []Transaction
	for _, acc := range config.GenesisAccounts {
		tx := Transaction{
			Hash:      "genesis_" + acc.Address,
			Nonce:     0,
			From:      "0x0000000000000000000000000000000000000000",
			To:        acc.Address,
			Value:     acc.Balance,
			GasPrice:  0,
			GasLimit:  0,
			Timestamp: time.Now().Unix(),
		}
		genesisTXs = append(genesisTXs, tx)
	}
	
	genesisBlock := NewBlock(0, "0", genesisTXs, "genesis")
	genesisBlock.Header.Timestamp = 1701388800 // Fixed timestamp
	genesisBlock.Header.Difficulty = 1
	
	return genesisBlock
}

// Add new block to chain
func (cm *ChainManager) AddBlock(block *Block) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	// Validate block
	if !block.Validate(cm.GetLatestBlock()) {
		return fmt.Errorf("invalid block")
	}
	
	// Apply transactions
	for _, tx := range block.Transactions {
		if err := cm.applyTransaction(tx); err != nil {
			return fmt.Errorf("failed to apply transaction: %v", err)
		}
	}
	
	// Update validator stake (PoVC reward)
	cm.updateValidatorReward(block.Header.Validator, block.Header.Reward)
	
	// Add block to chain
	cm.Chain = append(cm.Chain, block)
	
	// Remove processed transactions from mempool
	cm.removeFromMempool(block.Transactions)
	
	return nil
}

func (cm *ChainManager) applyTransaction(tx Transaction) error {
	// Check sender balance
	senderState, exists := cm.State[tx.From]
	if !exists {
		senderState = AccountState{Nonce: 0, Balance: 0}
	}
	
	// Check nonce
	if tx.Nonce != senderState.Nonce {
		return fmt.Errorf("invalid nonce: expected %d, got %d", senderState.Nonce, tx.Nonce)
	}
	
	// Calculate total cost
	totalCost := tx.Value + (tx.GasPrice * tx.GasLimit)
	if senderState.Balance < totalCost {
		return fmt.Errorf("insufficient balance")
	}
	
	// Update sender
	senderState.Balance -= totalCost
	senderState.Nonce++
	senderState.LastActive = time.Now().Unix()
	cm.State[tx.From] = senderState
	
	// Update receiver
	receiverState := cm.State[tx.To]
	receiverState.Balance += tx.Value
	receiverState.LastActive = time.Now().Unix()
	cm.State[tx.To] = receiverState
	
	return nil
}

func (cm *ChainManager) updateValidatorReward(validator string, reward uint64) {
	state := cm.State[validator]
	state.Balance += reward
	state.LastActive = time.Now().Unix()
	cm.State[validator] = state
}

func (cm *ChainManager) removeFromMempool(txs []Transaction) {
	newMempool := []Transaction{}
	for _, pending := range cm.PendingTXs {
		found := false
		for _, processed := range txs {
			if pending.Hash == processed.Hash {
				found = true
				break
			}
		}
		if !found {
			newMempool = append(newMempool, pending)
		}
	}
	cm.PendingTXs = newMempool
}

// Add transaction to mempool
func (cm *ChainManager) AddTransaction(tx Transaction) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	// Basic validation
	if !tx.Validate() {
		return fmt.Errorf("invalid transaction")
	}
	
	// Check if already in mempool
	for _, pending := range cm.PendingTXs {
		if pending.Hash == tx.Hash {
			return fmt.Errorf("transaction already in mempool")
		}
	}
	
	cm.PendingTXs = append(cm.PendingTXs, tx)
	return nil
}

// Get latest block
func (cm *ChainManager) GetLatestBlock() *Block {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	if len(cm.Chain) == 0 {
		return nil
	}
	return cm.Chain[len(cm.Chain)-1]
}

// Get chain height
func (cm *ChainManager) GetHeight() uint64 {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return uint64(len(cm.Chain))
}

// Get account balance
func (cm *ChainManager) GetBalance(address string) uint64 {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	state, exists := cm.State[address]
	if !exists {
		return 0
	}
	return state.Balance
}

// Get pending transactions
func (cm *ChainManager) GetPendingTXs() []Transaction {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.PendingTXs
}

// Create new transaction
func (cm *ChainManager) CreateTransaction(from, to string, value uint64, data []byte) (*Transaction, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	state, exists := cm.State[from]
	if !exists {
		state = AccountState{Nonce: 0, Balance: 0}
	}
	
	tx := Transaction{
		Nonce:     state.Nonce,
		From:      from,
		To:        to,
		Value:     value,
		GasPrice:  cm.config.MinGasPrice,
		GasLimit:  21000, // Basic transfer
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
	
	tx.Hash = tx.CalculateHash()
	
	return &tx, nil
}

// Export chain to JSON
func (cm *ChainManager) ExportChain() (string, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	data := struct {
		Chain   []*Block               `json:"chain"`
		State   map[string]AccountState `json:"state"`
		Pending []Transaction          `json:"pending_txs"`
	}{
		Chain:   cm.Chain,
		State:   cm.State,
		Pending: cm.PendingTXs,
	}
	
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	
	return string(bytes), nil
}