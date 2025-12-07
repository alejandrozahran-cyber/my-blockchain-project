#!/bin/bash

echo "🚀 Starting NUSA Chain deployment..."

# Run setup first
chmod +x scripts/setup.sh
./scripts/setup.sh

echo ""
echo "📦 Starting Docker network..."
docker-compose down 2>/dev/null
docker-compose build --pull
docker-compose up -d

echo ""
echo "⏳ Waiting for services to start..."
sleep 10

# Check services
echo ""
echo "🔍 Checking services..."
if curl -s http://localhost:8000/health > /dev/null; then
    echo "✅ L3 AI API is running"
else
    echo "⚠️  L3 AI API not responding"
fi

if curl -s http://localhost:8545 > /dev/null; then
    echo "✅ L1 JSON-RPC is running"
else
    echo "⚠️  L1 JSON-RPC not responding"
fi

echo ""
echo "🌐 NUSA Chain deployment completed!"
echo ""
echo "📊 API Endpoints:"
echo "   L1 JSON-RPC:    http://localhost:8545"
echo "   L3 AI API:      http://localhost:8000"
echo "   L3 API Docs:    http://localhost:8000/docs"
echo "   Redis:          localhost:6379"
echo "   PostgreSQL:     localhost:5432"
echo ""
echo "📝 Logs:"
echo "   docker-compose logs -f l1-node"
echo "   docker-compose logs -f l3-ai"
echo ""
echo "🛑 To stop: docker-compose down"
echo ""