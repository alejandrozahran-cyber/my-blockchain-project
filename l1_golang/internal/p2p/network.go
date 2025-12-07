package p2p

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
	
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/multiformats/go-multiaddr"
)

type P2PNetwork struct {
	host        host.Host
	peers       map[peer.ID]*PeerInfo
	peerMutex   sync.RWMutex
	bootnodes   []string
	port        int
	isRunning   bool
}

type PeerInfo struct {
	ID        peer.ID
	Address   multiaddr.Multiaddr
	Connected bool
	LastSeen  time.Time
	Height    uint64
}

func NewP2PNetwork(port int, bootnodes []string) (*P2PNetwork, error) {
	// Create libp2p host
	h, err := libp2p.New(
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port)),
		libp2p.NATPortMap(),
		libp2p.EnableNATService(),
	)
	if err != nil {
		return nil, err
	}
	
	network := &P2PNetwork{
		host:      h,
		peers:     make(map[peer.ID]*PeerInfo),
		bootnodes: bootnodes,
		port:      port,
	}
	
	// Set stream handlers
	h.SetStreamHandler("/nusa/1.0.0", network.handleStream)
	
	return network, nil
}

func (n *P2PNetwork) Start() error {
	n.isRunning = true
	
	// Print our addresses
	fmt.Printf("🚀 P2P Node ID: %s\n", n.host.ID())
	fmt.Printf("📡 Listening on:\n")
	for _, addr := range n.host.Addrs() {
		fmt.Printf("  %s/p2p/%s\n", addr, n.host.ID())
	}
	
	// Connect to bootnodes
	go n.connectToBootnodes()
	
	// Start peer discovery
	go n.discoverPeers()
	
	// Start peer maintenance
	go n.maintainPeers()
	
	return nil
}

func (n *P2PNetwork) connectToBootnodes() {
	for _, bootnode := range n.bootnodes {
		ma, err := multiaddr.NewMultiaddr(bootnode)
		if err != nil {
			log.Printf("❌ Invalid bootnode address %s: %v", bootnode, err)
			continue
		}
		
		info, err := peer.AddrInfoFromP2pAddr(ma)
		if err != nil {
			log.Printf("❌ Failed to parse bootnode info %s: %v", bootnode, err)
			continue
		}
		
		if err := n.connectToPeer(*info); err != nil {
			log.Printf("❌ Failed to connect to bootnode %s: %v", bootnode, err)
		}
	}
}

func (n *P2PNetwork) connectToPeer(info peer.AddrInfo) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	n.host.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)
	
	if err := n.host.Connect(ctx, info); err != nil {
		return err
	}
	
	n.peerMutex.Lock()
	n.peers[info.ID] = &PeerInfo{
		ID:        info.ID,
		Address:   info.Addrs[0],
		Connected: true,
		LastSeen:  time.Now(),
	}
	n.peerMutex.Unlock()
	
	fmt.Printf("✅ Connected to peer: %s\n", info.ID)
	return nil
}

func (n *P2PNetwork) discoverPeers() {
	// Simple peer discovery - in production use DHT or mDNS
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		if !n.isRunning {
			return
		}
		
		// Ask connected peers for their peers
		n.peerMutex.RLock()
		for _, peer := range n.peers {
			if peer.Connected {
				go n.requestPeersFromPeer(peer.ID)
			}
		}
		n.peerMutex.RUnlock()
	}
}

func (n *P2PNetwork) requestPeersFromPeer(peerID peer.ID) {
	// Implementation for requesting peer list
	// This would open a stream and request peer addresses
}

func (n *P2PNetwork) maintainPeers() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		if !n.isRunning {
			return
		}
		
		n.peerMutex.Lock()
		for id, peer := range n.peers {
			// Remove inactive peers
			if time.Since(peer.LastSeen) > 5*time.Minute {
				delete(n.peers, id)
				fmt.Printf("🗑️  Removed inactive peer: %s\n", id)
			}
		}
		n.peerMutex.Unlock()
	}
}

func (n *P2PNetwork) handleStream(s network.Stream) {
	defer s.Close()
	
	peerID := s.Conn().RemotePeer()
	
	n.peerMutex.Lock()
	if _, exists := n.peers[peerID]; !exists {
		n.peers[peerID] = &PeerInfo{
			ID:        peerID,
			Address:   s.Conn().RemoteMultiaddr(),
			Connected: true,
			LastSeen:  time.Now(),
		}
	} else {
		n.peers[peerID].LastSeen = time.Now()
		n.peers[peerID].Connected = true
	}
	n.peerMutex.Unlock()
	
	// Handle incoming messages
	// This would parse and process different message types
}

func (n *P2PNetwork) BroadcastBlock(blockData []byte) {
	n.peerMutex.RLock()
	defer n.peerMutex.RUnlock()
	
	for _, peer := range n.peers {
		if peer.Connected {
			go n.sendToPeer(peer.ID, "/nusa/block", blockData)
		}
	}
}

func (n *P2PNetwork) BroadcastTransaction(txData []byte) {
	n.peerMutex.RLock()
	defer n.peerMutex.RUnlock()
	
	for _, peer := range n.peers {
		if peer.Connected {
			go n.sendToPeer(peer.ID, "/nusa/tx", txData)
		}
	}
}

func (n *P2PNetwork) sendToPeer(peerID peer.ID, protocol string, data []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	s, err := n.host.NewStream(ctx, peerID, protocol)
	if err != nil {
		return err
	}
	defer s.Close()
	
	_, err = s.Write(data)
	return err
}

func (n *P2PNetwork) GetPeerCount() int {
	n.peerMutex.RLock()
	defer n.peerMutex.RUnlock()
	return len(n.peers)
}

func (n *P2PNetwork) Stop() {
	n.isRunning = false
	n.host.Close()
}