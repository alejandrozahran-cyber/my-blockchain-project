package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
	Address    common.Address
}

func NewWallet() (*Wallet, error) {
	// Generate private key
	privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	// Derive public key
	publicKey := &privateKey.PublicKey

	// Generate Ethereum-style address
	address := crypto.PubkeyToAddress(*publicKey)

	return &Wallet{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Address:    address,
	}, nil
}

func (w *Wallet) Sign(data []byte) (string, error) {
	// Hash the data
	hash := sha256.Sum256(data)

	// Sign the hash
	r, s, err := ecdsa.Sign(rand.Reader, w.PrivateKey, hash[:])
	if err != nil {
		return "", err
	}

	// Encode signature
	signature := append(r.Bytes(), s.Bytes()...)
	return hex.EncodeToString(signature), nil
}

func VerifySignature(publicKey *ecdsa.PublicKey, data []byte, signature string) bool {
	// Decode signature
	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}

	if len(sigBytes) != 64 {
		return false
	}

	r := new(big.Int).SetBytes(sigBytes[:32])
	s := new(big.Int).SetBytes(sigBytes[32:])

	// Hash the data
	hash := sha256.Sum256(data)

	// Verify signature
	return ecdsa.Verify(publicKey, hash[:], r, s)
}

func (w *Wallet) GetPrivateKeyHex() string {
	return hex.EncodeToString(crypto.FromECDSA(w.PrivateKey))
}

func (w *Wallet) GetPublicKeyHex() string {
	return hex.EncodeToString(crypto.FromECDSAPub(w.PublicKey))
}

func (w *Wallet) String() string {
	return fmt.Sprintf("Address: %s\nPrivate Key: %s...",
		w.Address.Hex(),
		w.GetPrivateKeyHex()[:16],
	)
}

// BIP39 Mnemonic support (simplified)
func GenerateMnemonic() string {
	// In production, use proper BIP39 implementation
	// This is simplified for demo
	words := []string{
		"nusa", "chain", "block", "crypto", "value", "creation",
		"ai", "neutrality", "wealth", "distribution", "fair", "economy",
		"proof", "contribution", "reward", "system", "anti", "whale",
		"decentralized", "autonomous", "organization", "consensus",
	}

	// Generate random 12-word mnemonic
	indices := make([]int, 12)
	for i := range indices {
		// In production, use proper random generation
		indices[i] = i % len(words)
	}

	mnemonic := ""
	for i, idx := range indices {
		if i > 0 {
			mnemonic += " "
		}
		mnemonic += words[idx]
	}

	return mnemonic
}