package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

type Block struct {
	Index        int64         `json:"index"`
	Timestamp    int64         `json:"timestamp"`
	Transactions []Transaction `json:"transactions"`
	PreviousHash string        `json:"previous_hash"`
	Hash         string        `json:"hash"`
	Validator    string        `json:"validator"`
	Nonce        int64         `json:"nonce"`
	Difficulty   int           `json:"difficulty"`
}

type Transaction struct {
	Hash      string  `json:"hash"`
	From      string  `json:"from"`
	To        string  `json:"to"`
	Value     float64 `json:"value"`
	Gas       int64   `json:"gas"`
	GasPrice  float64 `json:"gas_price"`
	Nonce     int64   `json:"nonce"`
	Timestamp int64   `json:"timestamp"`
	Data      string  `json:"data"`
	Signature string  `json:"signature"`
}

func NewBlock(index int64, transactions []Transaction, previousHash string) *Block {
	block := &Block{
		Index:        index,
		Timestamp:    time.Now().Unix(),
		Transactions: transactions,
		PreviousHash: previousHash,
		Nonce:        0,
		Difficulty:   4,
	}
	block.Hash = block.CalculateHash()
	return block
}

func (b *Block) CalculateHash() string {
	data := struct {
		Index        int64         `json:"index"`
		Timestamp    int64         `json:"timestamp"`
		Transactions []Transaction `json:"transactions"`
		PreviousHash string        `json:"previous_hash"`
		Nonce        int64         `json:"nonce"`
	}{
		Index:        b.Index,
		Timestamp:    b.Timestamp,
		Transactions: b.Transactions,
		PreviousHash: b.PreviousHash,
		Nonce:        b.Nonce,
	}

	bytes, _ := json.Marshal(data)
	hash := sha256.Sum256(bytes)
	return hex.EncodeToString(hash[:])
}

func (b *Block) MineBlock(difficulty int) {
	target := ""
	for i := 0; i < difficulty; i++ {
		target += "0"
	}

	for b.Hash[:difficulty] != target {
		b.Nonce++
		b.Hash = b.CalculateHash()
	}
}

func (b *Block) IsValid() bool {
	if b.Hash != b.CalculateHash() {
		return false
	}

	// Check if hash meets difficulty requirement
	target := ""
	for i := 0; i < b.Difficulty; i++ {
		target += "0"
	}

	if b.Hash[:b.Difficulty] != target {
		return false
	}

	return true
}