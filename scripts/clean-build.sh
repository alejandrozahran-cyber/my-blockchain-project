#!/bin/bash

echo "🧹 Cleaning previous builds..."
docker-compose down -v
docker system prune -f

echo "🔨 Building fresh images..."
docker-compose build --no-cache --pull

echo "🚀 Starting services..."
docker-compose up -d

echo "⏳ Waiting for services to be ready..."
sleep 15

echo "📊 Checking service status..."
docker-compose ps

echo ""
echo "✅ Done! Services:"
echo "   L1 Node:    http://localhost:8545"
echo "   L3 AI API:  http://localhost:8000"
echo "   Redis:      localhost:6379"
echo "   PostgreSQL: localhost:5432"