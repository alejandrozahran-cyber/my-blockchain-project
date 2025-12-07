package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

// Real Block Structure
type Block struct {
	Header       BlockHeader   `json:"header"`
	Transactions []Transaction `json:"transactions"`
	Signature    string        `json:"signature,omitempty"`
}

type BlockHeader struct {
	Version        uint64    `json:"version"`
	Height         uint64    `json:"height"`
	Timestamp      int64     `json:"timestamp"`
	PrevHash       string    `json:"prev_hash"`
	MerkleRoot     string    `json:"merkle_root"`
	StateRoot      string    `json:"state_root"`
	Validator      string    `json:"validator"`
	Nonce          uint64    `json:"nonce"`
	Difficulty     uint64    `json:"difficulty"`
	GasLimit       uint64    `json:"gas_limit"`
	GasUsed        uint64    `json:"gas_used"`
	Reward         uint64    `json:"reward"`
	ExtraData      string    `json:"extra_data,omitempty"`
}

type Transaction struct {
	Hash        string          `json:"hash"`
	Nonce       uint64          `json:"nonce"`
	From        string          `json:"from"`
	To          string          `json:"to"`
	Value       uint64          `json:"value"`
	GasPrice    uint64          `json:"gas_price"`
	GasLimit    uint64          `json:"gas_limit"`
	Data        []byte          `json:"data,omitempty"`
	Signature   TransactionSig  `json:"signature"`
	Timestamp   int64           `json:"timestamp"`
}

type TransactionSig struct {
	R string `json:"r"`
	S string `json:"s"`
	V uint8  `json:"v"`
}

// Create new block
func NewBlock(height uint64, prevHash string, txs []Transaction, validator string) *Block {
	block := &Block{
		Header: BlockHeader{
			Version:    1,
			Height:     height,
			Timestamp:  time.Now().Unix(),
			PrevHash:   prevHash,
			Validator:  validator,
			Difficulty: 1000000,
			GasLimit:   8000000,
			Reward:     2000000000000000000, // 2 NUSA
		},
		Transactions: txs,
	}
	
	// Calculate merkle root
	block.Header.MerkleRoot = block.CalculateMerkleRoot()
	block.Header.StateRoot = block.CalculateStateRoot()
	
	return block
}

// Calculate block hash
func (b *Block) Hash() string {
	headerBytes, _ := json.Marshal(b.Header)
	hash := sha256.Sum256(headerBytes)
	return hex.EncodeToString(hash[:])
}

// Calculate merkle root
func (b *Block) CalculateMerkleRoot() string {
	if len(b.Transactions) == 0 {
		return hex.EncodeToString(sha256.New().Sum(nil))
	}
	
	var hashes []string
	for _, tx := range b.Transactions {
		hashes = append(hashes, tx.Hash)
	}
	
	return buildMerkleTree(hashes)
}

func buildMerkleTree(hashes []string) string {
	if len(hashes) == 1 {
		return hashes[0]
	}
	
	var newLevel []string
	for i := 0; i < len(hashes); i += 2 {
		if i+1 < len(hashes) {
			combined := hashes[i] + hashes[i+1]
			hash := sha256.Sum256([]byte(combined))
			newLevel = append(newLevel, hex.EncodeToString(hash[:]))
		} else {
			// Duplicate last hash if odd number
			combined := hashes[i] + hashes[i]
			hash := sha256.Sum256([]byte(combined))
			newLevel = append(newLevel, hex.EncodeToString(hash[:]))
		}
	}
	
	return buildMerkleTree(newLevel)
}

// Validate block
func (b *Block) Validate(prevBlock *Block) bool {
	// Check block hash
	expectedHash := b.Hash()
	if b.Header.Height > 0 && b.Header.PrevHash != prevBlock.Hash() {
		return false
	}
	
	// Check timestamp
	if b.Header.Timestamp > time.Now().Unix()+10 {
		return false // Future block
	}
	
	if prevBlock != nil && b.Header.Timestamp <= prevBlock.Header.Timestamp {
		return false // Older than previous block
	}
	
	// Check merkle root
	if b.Header.MerkleRoot != b.CalculateMerkleRoot() {
		return false
	}
	
	// Validate all transactions
	for _, tx := range b.Transactions {
		if !tx.Validate() {
			return false
		}
	}
	
	return true
}

// Validate transaction
func (tx *Transaction) Validate() bool {
	// Basic validation
	if tx.From == "" || tx.Value == 0 {
		return false
	}
	
	// Check hash
	if tx.Hash != tx.CalculateHash() {
		return false
	}
	
	// TODO: Verify signature
	// if !tx.VerifySignature() { return false }
	
	return true
}

func (tx *Transaction) CalculateHash() string {
	data := struct {
		Nonce    uint64 `json:"nonce"`
		From     string `json:"from"`
		To       string `json:"to"`
		Value    uint64 `json:"value"`
		GasPrice uint64 `json:"gas_price"`
		GasLimit uint64 `json:"gas_limit"`
		Data     []byte `json:"data,omitempty"`
	}{
		Nonce:    tx.Nonce,
		From:     tx.From,
		To:       tx.To,
		Value:    tx.Value,
		GasPrice: tx.GasPrice,
		GasLimit: tx.GasLimit,
		Data:     tx.Data,
	}
	
	bytes, _ := json.Marshal(data)
	hash := sha256.Sum256(bytes)
	return hex.EncodeToString(hash[:])
}

func (b *Block) CalculateStateRoot() string {
	// Simplified state root calculation
	// In production, use Patricia Merkle Tree
	stateData := struct {
		Height    uint64 `json:"height"`
		Timestamp int64  `json:"timestamp"`
		TxCount   int    `json:"tx_count"`
	}{
		Height:    b.Header.Height,
		Timestamp: b.Header.Timestamp,
		TxCount:   len(b.Transactions),
	}
	
	bytes, _ := json.Marshal(stateData)
	hash := sha256.Sum256(bytes)
	return hex.EncodeToString(hash[:])
}