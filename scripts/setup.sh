#!/bin/bash

echo "🔧 Setting up NUSA Chain Development Environment..."

# Install prerequisites (for Windows, skip if using Docker only)
echo "1. Checking prerequisites..."
if ! command -v go &> /dev/null; then
    echo "⚠️  Go not found. Please install Go 1.21+"
    exit 1
fi

if ! command -v docker &> /dev/null; then
    echo "⚠️  Docker not found. Please install Docker Desktop"
    exit 1
fi

# Create necessary directories
echo "2. Creating directories..."
mkdir -p {bin,data,logs,config}

# Build binaries
echo "3. Building binaries..."
cd l1_golang
go mod tidy
go build -o ../bin/nusa-node ./cmd/node || echo "⚠️  Go build failed, continuing..."

cd ../l2_rust
cargo build --release 2>/dev/null || echo "⚠️  Rust build failed, continuing..."

# Create default configs
echo "4. Creating default configurations..."
cd ..
cat > config/node.yaml << 'EOF'
network:
  name: "nusa-testnet"
  chain_id: 2024
  port: 30303
  rpc_port: 8545

consensus:
  type: "PoVC"
  block_time: 5
  validator_count: 1

database:
  path: "./data/chaindata"
  type: "leveldb"
EOF

cat > config/genesis.json << 'EOF'
{
  "chainId": 2024,
  "alloc": {
    "0x0000000000000000000000000000000000000000": {
      "balance": "25000000000000000000000000"
    }
  },
  "config": {
    "homesteadBlock": 0,
    "eip150Block": 0,
    "eip155Block": 0,
    "eip158Block": 0,
    "povcBlock": 0
  }
}
EOF

echo "✅ Setup completed!"
echo ""
echo "📋 Next steps:"
echo "1. Start with Docker: docker-compose up"
echo "2. Or run manually: ./bin/nusa-node --config config/node.yaml"
echo "3. Access API: http://localhost:8000/docs"