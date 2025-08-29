# 🚀 Quick Start Guide

## Immediate Testing

### 1. Start the Server
```bash
# Option 1: Direct run
go run main.go

# Option 2: Build and run
go build -o pub-sub-system .
./pub-sub-system
```

The server will start on `http://localhost:8080`

### 2. Test HTTP APIs
```bash
# Health check
curl http://localhost:8080/health

# Create a topic
curl -X POST http://localhost:8080/topics \
  -H "Content-Type: application/json" \
  -d '{"name": "orders"}'

# List topics
curl http://localhost:8080/topics

# Get stats
curl http://localhost:8080/stats

# Delete topic
curl -X DELETE http://localhost:8080/topics/orders
```

### 3. Test WebSocket
Open `websocket_test.html` in your browser and:
1. Connect to `ws://localhost:8080/ws`
2. Create a topic via HTTP API
3. Subscribe to the topic
4. Publish messages
5. See real-time delivery

## Docker Deployment

### 1. Build Image
```bash
docker build -t pub-sub-system .
```

### 2. Run Container
```bash
docker run -p 8080:8080 pub-sub-system
```

### 3. Docker Compose
```bash
docker-compose up --build
```

## Production Deployment

### 1. Build Binary
```bash
GOOS=linux GOARCH=amd64 go build -o pub-sub-system-linux .
```

### 2. Deploy to Server
```bash
scp pub-sub-system-linux user@server:/opt/pub-sub/
scp docker-compose.yml user@server:/opt/pub-sub/
```

### 3. Run on Server
```bash
cd /opt/pub-sub
docker-compose up -d
```

## Key Features Tested

✅ **HTTP REST APIs**: Create, delete, list topics  
✅ **Health & Stats**: System monitoring endpoints  
✅ **WebSocket Pub/Sub**: Real-time messaging  
✅ **Message History**: Configurable replay with `last_n`  
✅ **Concurrency**: Thread-safe operations  
✅ **Error Handling**: Proper HTTP status codes  
✅ **Backpressure**: Bounded queues with overflow protection  

## Next Steps

1. **Load Testing**: Test with multiple concurrent clients
2. **Authentication**: Add API key validation
3. **Persistence**: Integrate with Redis/PostgreSQL
4. **Monitoring**: Add Prometheus metrics
5. **Scaling**: Deploy behind load balancer

## Support

- Check `README.md` for detailed documentation
- Run `go test ./...` to verify functionality
- Use `websocket_test.html` for WebSocket testing
- Review logs for debugging information
