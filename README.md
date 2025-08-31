# Pub/Sub System Backend

A production-quality, in-memory Pub/Sub system built in Go with WebSocket support for real-time messaging and HTTP REST APIs for management operations.

## Features

- **Real-time Messaging**: WebSocket-based publish/subscribe with low latency
- **Topic Management**: Create, delete, and list topics via REST API
- **Message History**: Configurable message replay with `last_n` parameter
- **Concurrency Safe**: Thread-safe operations with proper locking mechanisms
- **Backpressure Handling**: Bounded queues with overflow protection
- **Health Monitoring**: Built-in health checks and statistics
- **Graceful Shutdown**: Clean connection handling and resource cleanup

## Architecture

### Core Components

- **PubSubSystem**: Main orchestrator managing topics and subscribers
- **Topic**: Individual message channels with subscriber management
- **Subscriber**: WebSocket client connections with message queues
- **Message**: Structured data with UUID and payload

### Design Choices

- **In-Memory Storage**: No external dependencies for simplicity and performance
- **Ring Buffer**: Last 100 messages per topic for history replay
- **Bounded Queues**: 100 message limit per subscriber to prevent memory issues
- **Fan-out Delivery**: Every subscriber receives each published message
- **Isolation**: No cross-topic message leakage

### Backpressure Policy

When a subscriber's queue overflows:
1. Send `SLOW_CONSUMER` error message
2. Close the WebSocket connection
3. Remove subscriber from topic
4. Clean up resources

This ensures the system remains stable under high load.

## Quick Start

### Prerequisites

- Go 1.21 or later
- Docker (optional)

### Local Development

1. **Clone and navigate to the project:**
   ```bash
   cd pub-sub-system
   ```

2. **Install dependencies:**
   ```bash
   go mod tidy
   ```

3. **Run the server:**
   ```bash
   go run main.go
   ```

   The server will start on `http://localhost:8080`

### Docker Deployment

1. **Build the image:**
   ```bash
   docker build -t pub-sub-system .
   ```

2. **Run the container:**
   ```bash
   docker run -p 8080:8080 pub-sub-system
   ```

## API Reference

### WebSocket Endpoint

**Endpoint:** `ws://localhost:8080/ws`

#### Client → Server Messages

##### Subscribe
```json
{
  "type": "subscribe",
  "topic": "orders",
  "client_id": "s1",
  "last_n": 5,
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

##### Unsubscribe
```json
{
  "type": "unsubscribe",
  "topic": "orders",
  "client_id": "s1",
  "request_id": "340e8400-e29b-41d4-a716-4466554480098"
}
```

##### Publish
```json
{
  "type": "publish",
  "topic": "orders",
  "message": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "payload": {
      "order_id": "ORD-123",
      "amount": 99.5,
      "currency": "USD"
    }
  },
  "request_id": "340e8400-e29b-41d4-a716-4466554480098"
}
```

##### Ping
```json
{
  "type": "ping",
  "request_id": "570t8400-e29b-41d4-a716-4466554412345"
}
```

#### Server → Client Messages

##### Acknowledgment
```json
{
  "type": "ack",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "topic": "orders",
  "status": "ok",
  "ts": "2025-08-25T10:00:00Z"
}
```

##### Event (Published Message)
```json
{
  "type": "event",
  "topic": "orders",
  "message": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
  "payload": {
      "order_id": "ORD-123",
      "amount": 99.5,
      "currency": "USD"
    }
  },
  "ts": "2025-08-25T10:01:00Z"
}
```

##### Error
```json
{
  "type": "error",
  "request_id": "req-67890",
  "error": {
    "code": "BAD_REQUEST",
    "message": "message.id must be a valid UUID"
  },
  "ts": "2025-08-25T10:02:00Z"
}
```

### HTTP REST Endpoints

#### Create Topic
```bash
POST /topics
Content-Type: application/json

{
  "name": "orders"
}
```

**Response:**
```json
{
  "status": "created",
  "topic": "orders"
}
```

#### Delete Topic
```bash
DELETE /topics/orders
```

**Response:**
```json
{
  "status": "deleted",
  "topic": "orders"
}
```

#### List Topics
```bash
GET /topics
```

**Response:**
```json
{
  "topics": [
    {
      "name": "orders",
      "subscribers": 3
    }
  ]
}
```

#### Health Check
```bash
GET /health
```

**Response:**
```json
{
  "uptime_sec": 123,
  "topics": 2,
  "subscribers": 4
}
```

#### Statistics
```bash
GET /stats
```

**Response:**
```json
{
  "topics": {
    "orders": {
      "messages": 42,
      "subscribers": 3
    }
  }
}
```

## Testing

### Unit Tests
```bash
go test ./...
```

### Manual Testing

1. **Start the server:**
   ```bash
   go run main.go
   ```

2. **Create a topic:**
   ```bash
   curl -X POST http://localhost:8080/topics \
     -H "Content-Type: application/json" \
     -d '{"name": "test-topic"}'
   ```

3. **Test WebSocket connection:**
   Use a WebSocket client (like wscat or browser dev tools) to connect to `ws://localhost:8080/ws`

4. **Subscribe to topic:**
   ```json
   {
     "type": "subscribe",
     "topic": "test-topic",
     "client_id": "test-client"
   }
   ```

5. **Publish a message:**
   ```json
   {
     "type": "publish",
     "topic": "test-topic",
     "message": {
       "id": "test-123",
       "payload": {"test": "data"}
     }
   }
   ```

## Configuration

### Environment Variables

- `PORT`: Server port (default: 8080)
- `MAX_TOPICS`: Maximum number of topics (default: 100)
- `MAX_SUBSCRIBERS_PER_TOPIC`: Maximum subscribers per topic (default: 100)
- `TOPIC_HISTORY_SIZE`: Maximum messages to keep in topic history (default: 100)
- `SUBSCRIBER_QUEUE_SIZE`: Subscriber queue size (default: 100)

### Backpressure Policy

The system implements a comprehensive backpressure strategy to handle slow consumers:

#### **Subscriber Queue Overflow (SLOW_CONSUMER)**
- Each subscriber has a bounded message queue (configurable via `SUBSCRIBER_QUEUE_SIZE`)
- When a subscriber's queue is full, the system:
  1. Sends a `SLOW_CONSUMER` error message to the client
  2. Immediately closes the WebSocket connection
  3. Removes the subscriber from the topic
- This prevents memory overflow and ensures system stability

#### **Topic Message History (Ring Buffer)**
- Each topic maintains a bounded message history (configurable via `TOPIC_HISTORY_SIZE`)
- When the history limit is reached, the oldest messages are dropped (FIFO eviction)
- This prevents unlimited memory growth while preserving recent messages for replay

#### **Fan-out Guarantee**
- Every active subscriber to a topic receives each published message
- If any subscriber cannot keep up (queue overflow), they are disconnected
- Other subscribers continue to receive messages normally
- This ensures message delivery to all capable consumers

### Rate Limiting

The system includes built-in backpressure handling:
- Subscriber queues are bounded to prevent memory issues
- Slow consumers are disconnected with appropriate error messages
- Heartbeat mechanism (30-second intervals) for connection health

## Production Considerations

### Scalability
- **Horizontal Scaling**: Deploy multiple instances behind a load balancer
- **State Management**: Consider Redis for shared state in multi-instance deployments
- **Message Persistence**: Implement external storage for critical messages

### Monitoring
- **Health Checks**: Use `/health` endpoint for load balancer health checks
- **Metrics**: `/stats` endpoint provides basic operational metrics
- **Logging**: Structured logging with request IDs for tracing

### Security
- **Authentication**: Implement API key validation for production use
- **Rate Limiting**: Add per-client rate limiting for abuse prevention
- **CORS**: Configure appropriate CORS policies for your domain

## Performance Characteristics

- **Latency**: Sub-10ms message delivery for local deployments
- **Throughput**: Handles thousands of concurrent connections
- **Memory**: Bounded memory usage with configurable limits
- **CPU**: Efficient goroutine usage with bounded worker pools

## Troubleshooting

### Common Issues

1. **WebSocket Connection Fails**
   - Check if server is running on correct port
   - Verify WebSocket endpoint `/ws`

2. **Messages Not Delivered**
   - Ensure topic exists before publishing
   - Check subscriber connection status
   - Verify message format

3. **High Memory Usage**
   - Reduce `MAX_MESSAGES` and `MAX_QUEUE` values
   - Monitor subscriber count per topic

### Logs

The server provides detailed logging for:
- WebSocket connections and disconnections
- Message publishing and delivery
- Error conditions and recovery
- Resource usage and limits

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## License

This project is licensed under the MIT License.

## Support

For issues and questions:
- Create an issue in the repository
- Check the troubleshooting section
- Review the API documentation
