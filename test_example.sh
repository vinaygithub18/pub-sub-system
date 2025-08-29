#!/bin/bash

echo "🚀 Testing Pub/Sub System Backend"
echo "=================================="

# Wait for server to start
echo "⏳ Waiting for server to start..."
sleep 2

# Test health endpoint
echo "🏥 Testing health endpoint..."
curl -s http://localhost:8080/health | jq '.'

# Test creating a topic
echo "📝 Creating test topic..."
curl -s -X POST http://localhost:8080/topics \
  -H "Content-Type: application/json" \
  -d '{"name": "test-topic"}' | jq '.'

# Test listing topics
echo "📋 Listing topics..."
curl -s http://localhost:8080/topics | jq '.'

# Test getting stats
echo "📊 Getting statistics..."
curl -s http://localhost:8080/stats | jq '.'

echo ""
echo "✅ Basic HTTP API tests completed!"
echo ""
echo "🔌 To test WebSocket functionality:"
echo "1. Connect to ws://localhost:8080/ws"
echo "2. Subscribe to 'test-topic'"
echo "3. Publish messages"
echo "4. Check the README.md for detailed examples"
echo ""
echo "🧪 Run tests with: go test ./..."
