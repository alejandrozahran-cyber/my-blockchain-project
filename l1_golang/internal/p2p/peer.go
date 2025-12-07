package p2p

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type Peer struct {
	ID        string
	Address   string
	Port      int
	Connected bool
	LastSeen  time.Time
	mu        sync.RWMutex
}

type PeerManager struct {
	peers    map[string]*Peer
	bootnodes []string
	mu       sync.RWMutex
}

func NewPeerManager(bootnodes []string) *PeerManager {
	return &PeerManager{
		peers:     make(map[string]*Peer),
		bootnodes: bootnodes,
	}
}

func (pm *PeerManager) AddPeer(peer *Peer) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	pm.peers[peer.ID] = peer
}

func (pm *PeerManager) RemovePeer(peerID string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	delete(pm.peers, peerID)
}

func (pm *PeerManager) GetPeer(peerID string) (*Peer, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	peer, exists := pm.peers[peerID]
	return peer, exists
}

func (pm *PeerManager) GetAllPeers() []*Peer {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	peers := make([]*Peer, 0, len(pm.peers))
	for _, peer := range pm.peers {
		peers = append(peers, peer)
	}
	return peers
}

func (pm *PeerManager) ConnectToBootnodes() {
	for _, addr := range pm.bootnodes {
		go pm.connectToPeer(addr)
	}
}

func (pm *PeerManager) connectToPeer(address string) {
	// Try to connect to peer
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		fmt.Printf("❌ Failed to connect to %s: %v\n", address, err)
		return
	}
	defer conn.Close()

	// Create peer object
	peer := &Peer{
		ID:        generatePeerID(),
		Address:   address,
		Connected: true,
		LastSeen:  time.Now(),
	}

	pm.AddPeer(peer)
	fmt.Printf("✅ Connected to peer %s at %s\n", peer.ID, address)
}

func (pm *PeerManager) BroadcastMessage(message []byte) {
	pm.mu.RLock()
	peers := pm.GetAllPeers()
	pm.mu.RUnlock()

	for _, peer := range peers {
		if peer.Connected {
			go pm.sendToPeer(peer, message)
		}
	}
}

func (pm *PeerManager) sendToPeer(peer *Peer, message []byte) {
	// Send message to peer
	// Implementation depends on your protocol
	fmt.Printf("📤 Sending message to peer %s\n", peer.ID)
}

func generatePeerID() string {
	// Generate random peer ID
	// In production, use proper cryptographic random
	return fmt.Sprintf("peer-%d", time.Now().UnixNano())
}