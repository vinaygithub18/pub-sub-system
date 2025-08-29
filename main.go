package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Message represents a published message
type Message struct {
	ID      string      `json:"id"`
	Payload interface{} `json:"payload"`
}

// ClientMessage represents incoming WebSocket messages from clients
type ClientMessage struct {
	Type      string   `json:"type"`
	Topic     string   `json:"topic"`
	Message   *Message `json:"message,omitempty"`
	ClientID  string   `json:"client_id"`
	LastN     int      `json:"last_n,omitempty"`
	RequestID string   `json:"request_id,omitempty"`
}

// ServerMessage represents outgoing WebSocket messages to clients
type ServerMessage struct {
	Type      string   `json:"type"`
	RequestID string   `json:"request_id,omitempty"`
	Topic     string   `json:"topic,omitempty"`
	Message   *Message `json:"message,omitempty"`
	Error     *Error   `json:"error,omitempty"`
	Status    string   `json:"status,omitempty"`
	Msg       string   `json:"msg,omitempty"`
	TS        string   `json:"ts,omitempty"`
}

// Error represents error details
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Topic represents a pub/sub topic
type Topic struct {
	Name        string
	Subscribers map[string]*Subscriber
	Messages    []*Message
	MaxMessages int
	mu          sync.RWMutex
}

// Subscriber represents a WebSocket client subscription
type Subscriber struct {
	ID       string
	Conn     *websocket.Conn
	Topic    string
	Queue    chan *Message
	MaxQueue int
}

// PubSubSystem manages the entire pub/sub system
type PubSubSystem struct {
	Topics         map[string]*Topic
	StartTime      time.Time
	mu             sync.RWMutex
	MaxTopics      int
	MaxSubscribers int
}

// Global pub/sub system instance
var pubSub = NewPubSubSystem()

// NewPubSubSystem creates a new pub/sub system
func NewPubSubSystem() *PubSubSystem {
	return &PubSubSystem{
		Topics:         make(map[string]*Topic),
		StartTime:      time.Now(),
		MaxTopics:      100,
		MaxSubscribers: 1000,
	}
}

// NewTopic creates a new topic
func (ps *PubSubSystem) NewTopic(name string) (*Topic, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if len(ps.Topics) >= ps.MaxTopics {
		return nil, fmt.Errorf("maximum topics reached")
	}

	if _, exists := ps.Topics[name]; exists {
		return nil, fmt.Errorf("topic already exists")
	}

	topic := &Topic{
		Name:        name,
		Subscribers: make(map[string]*Subscriber),
		Messages:    make([]*Message, 0),
		MaxMessages: 100, // Ring buffer size
	}

	ps.Topics[name] = topic
	return topic, nil
}

// DeleteTopic deletes a topic and disconnects all subscribers
func (ps *PubSubSystem) DeleteTopic(name string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	topic, exists := ps.Topics[name]
	if !exists {
		return fmt.Errorf("topic not found")
	}

	// Disconnect all subscribers
	topic.mu.Lock()
	for _, sub := range topic.Subscribers {
		close(sub.Queue)
		sub.Conn.Close()
	}
	topic.mu.Unlock()

	delete(ps.Topics, name)
	return nil
}

// GetTopics returns all topics with subscriber counts
func (ps *PubSubSystem) GetTopics() map[string]int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	result := make(map[string]int)
	for name, topic := range ps.Topics {
		topic.mu.RLock()
		result[name] = len(topic.Subscribers)
		topic.mu.RUnlock()
	}
	return result
}

// GetStats returns detailed statistics
func (ps *PubSubSystem) GetStats() map[string]map[string]interface{} {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	result := make(map[string]map[string]interface{})
	for name, topic := range ps.Topics {
		topic.mu.RLock()
		result[name] = map[string]interface{}{
			"messages":    len(topic.Messages),
			"subscribers": len(topic.Subscribers),
		}
		topic.mu.RUnlock()
	}
	return result
}

// GetHealth returns health status
func (ps *PubSubSystem) GetHealth() map[string]interface{} {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	totalSubscribers := 0
	for _, topic := range ps.Topics {
		topic.mu.RLock()
		totalSubscribers += len(topic.Subscribers)
		topic.mu.RUnlock()
	}

	return map[string]interface{}{
		"uptime_sec":  int(time.Since(pubSub.StartTime).Seconds()),
		"topics":      len(ps.Topics),
		"subscribers": totalSubscribers,
	}
}

// AddMessage adds a message to a topic's history
func (t *Topic) AddMessage(msg *Message) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Messages = append(t.Messages, msg)
	if len(t.Messages) > t.MaxMessages {
		t.Messages = t.Messages[1:] // Remove oldest message
	}
}

// GetLastMessages returns the last N messages
func (t *Topic) GetLastMessages(n int) []*Message {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if n <= 0 || n > len(t.Messages) {
		n = len(t.Messages)
	}

	result := make([]*Message, n)
	copy(result, t.Messages[len(t.Messages)-n:])
	return result
}

// AddSubscriber adds a subscriber to a topic
func (t *Topic) AddSubscriber(sub *Subscriber) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(t.Subscribers) >= 100 { // Max subscribers per topic
		return fmt.Errorf("maximum subscribers reached for topic")
	}

	t.Subscribers[sub.ID] = sub
	return nil
}

// RemoveSubscriber removes a subscriber from a topic
func (t *Topic) RemoveSubscriber(subID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if sub, exists := t.Subscribers[subID]; exists {
		close(sub.Queue)
		delete(t.Subscribers, subID)
	}
}

// Broadcast sends a message to all subscribers of a topic
func (t *Topic) Broadcast(msg *Message) {
	t.mu.RLock()
	subscribers := make([]*Subscriber, 0, len(t.Subscribers))
	for _, sub := range t.Subscribers {
		subscribers = append(subscribers, sub)
	}
	t.mu.RUnlock()

	// Send message to each subscriber
	for _, sub := range subscribers {
		select {
		case sub.Queue <- msg:
			// Message sent successfully
		default:
			// Queue is full, send SLOW_CONSUMER error
			errorMsg := &ServerMessage{
				Type: "error",
				Error: &Error{
					Code:    "SLOW_CONSUMER",
					Message: "Subscriber queue overflow",
				},
				TS: time.Now().UTC().Format(time.RFC3339),
			}
			sub.Conn.WriteJSON(errorMsg)
			// Close connection for slow consumer
			sub.Conn.Close()
		}
	}
}

// NewSubscriber creates a new subscriber
func NewSubscriber(id, topic string, conn *websocket.Conn) *Subscriber {
	return &Subscriber{
		ID:       id,
		Conn:     conn,
		Topic:    topic,
		Queue:    make(chan *Message, 100), // Bounded queue
		MaxQueue: 100,
	}
}

// StartMessageProcessor starts processing messages for a subscriber
func (s *Subscriber) StartMessageProcessor(ctx context.Context) {
	go func() {
		defer func() {
			s.Conn.Close()
			// Remove from topic
			if topic, exists := pubSub.Topics[s.Topic]; exists {
				topic.RemoveSubscriber(s.ID)
			}
		}()

		for {
			select {
			case msg := <-s.Queue:
				if msg == nil {
					return // Channel closed
				}

				serverMsg := &ServerMessage{
					Type:    "event",
					Topic:   s.Topic,
					Message: msg,
					TS:      time.Now().UTC().Format(time.RFC3339),
				}

				if err := s.Conn.WriteJSON(serverMsg); err != nil {
					log.Printf("Error sending message to subscriber %s: %v", s.ID, err)
					return
				}

			case <-ctx.Done():
				return
			}
		}
	}()
}

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// WebSocket handler
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start heartbeat goroutine
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				infoMsg := &ServerMessage{
					Type: "info",
					Msg:  "ping",
					TS:   time.Now().UTC().Format(time.RFC3339),
				}
				conn.WriteJSON(infoMsg)
			case <-ctx.Done():
				return
			}
		}
	}()

	for {
		var clientMsg ClientMessage
		if err := conn.ReadJSON(&clientMsg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		handleClientMessage(conn, &clientMsg, ctx)
	}
}

// handleClientMessage processes incoming client messages
func handleClientMessage(conn *websocket.Conn, msg *ClientMessage, ctx context.Context) {
	switch msg.Type {
	case "subscribe":
		handleSubscribe(conn, msg, ctx)
	case "unsubscribe":
		handleUnsubscribe(conn, msg)
	case "publish":
		handlePublish(conn, msg)
	case "ping":
		handlePing(conn, msg)
	default:
		sendError(conn, "BAD_REQUEST", "Invalid message type", msg.RequestID)
	}
}

// handleSubscribe handles subscription requests
func handleSubscribe(conn *websocket.Conn, msg *ClientMessage, ctx context.Context) {
	if msg.Topic == "" || msg.ClientID == "" {
		sendError(conn, "BAD_REQUEST", "Topic and client_id are required", msg.RequestID)
		return
	}

	pubSub.mu.RLock()
	topic, exists := pubSub.Topics[msg.Topic]
	pubSub.mu.RUnlock()

	if !exists {
		sendError(conn, "TOPIC_NOT_FOUND", "Topic does not exist", msg.RequestID)
		return
	}

	sub := NewSubscriber(msg.ClientID, msg.Topic, conn)
	if err := topic.AddSubscriber(sub); err != nil {
		sendError(conn, "INTERNAL", err.Error(), msg.RequestID)
		return
	}

	// Start message processor
	sub.StartMessageProcessor(ctx)

	// Send acknowledgment
	ack := &ServerMessage{
		Type:      "ack",
		RequestID: msg.RequestID,
		Topic:     msg.Topic,
		Status:    "ok",
		TS:        time.Now().UTC().Format(time.RFC3339),
	}
	conn.WriteJSON(ack)

	// Send historical messages if requested
	if msg.LastN > 0 {
		historicalMessages := topic.GetLastMessages(msg.LastN)
		for _, histMsg := range historicalMessages {
			eventMsg := &ServerMessage{
				Type:    "event",
				Topic:   msg.Topic,
				Message: histMsg,
				TS:      time.Now().UTC().Format(time.RFC3339),
			}
			conn.WriteJSON(eventMsg)
		}
	}
}

// handleUnsubscribe handles unsubscription requests
func handleUnsubscribe(conn *websocket.Conn, msg *ClientMessage) {
	if msg.Topic == "" || msg.ClientID == "" {
		sendError(conn, "BAD_REQUEST", "Topic and client_id are required", msg.RequestID)
		return
	}

	pubSub.mu.RLock()
	topic, exists := pubSub.Topics[msg.Topic]
	pubSub.mu.RUnlock()

	if !exists {
		sendError(conn, "TOPIC_NOT_FOUND", "Topic does not exist", msg.RequestID)
		return
	}

	topic.RemoveSubscriber(msg.ClientID)

	ack := &ServerMessage{
		Type:      "ack",
		RequestID: msg.RequestID,
		Topic:     msg.Topic,
		Status:    "ok",
		TS:        time.Now().UTC().Format(time.RFC3339),
	}
	conn.WriteJSON(ack)
}

// handlePublish handles publish requests
func handlePublish(conn *websocket.Conn, msg *ClientMessage) {
	if msg.Topic == "" || msg.Message == nil {
		sendError(conn, "BAD_REQUEST", "Topic and message are required", msg.RequestID)
		return
	}

	// Validate message ID
	if msg.Message.ID == "" {
		msg.Message.ID = uuid.New().String()
	}

	pubSub.mu.RLock()
	topic, exists := pubSub.Topics[msg.Topic]
	pubSub.mu.RUnlock()

	if !exists {
		sendError(conn, "TOPIC_NOT_FOUND", "Topic does not exist", msg.RequestID)
		return
	}

	// Add message to topic history
	topic.AddMessage(msg.Message)

	// Broadcast to all subscribers
	topic.Broadcast(msg.Message)

	// Send acknowledgment
	ack := &ServerMessage{
		Type:      "ack",
		RequestID: msg.RequestID,
		Topic:     msg.Topic,
		Status:    "ok",
		TS:        time.Now().UTC().Format(time.RFC3339),
	}
	conn.WriteJSON(ack)
}

// handlePing handles ping requests
func handlePing(conn *websocket.Conn, msg *ClientMessage) {
	pong := &ServerMessage{
		Type:      "pong",
		RequestID: msg.RequestID,
		TS:        time.Now().UTC().Format(time.RFC3339),
	}
	conn.WriteJSON(pong)
}

// sendError sends an error message to the client
func sendError(conn *websocket.Conn, code, message, requestID string) {
	errorMsg := &ServerMessage{
		Type:      "error",
		RequestID: requestID,
		Error: &Error{
			Code:    code,
			Message: message,
		},
		TS: time.Now().UTC().Format(time.RFC3339),
	}
	conn.WriteJSON(errorMsg)
}

// HTTP handlers
func handleCreateTopic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Topic name is required", http.StatusBadRequest)
		return
	}

	topic, err := pubSub.NewTopic(req.Name)
	if err != nil {
		if err.Error() == "topic already exists" {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "created",
		"topic":  topic.Name,
	})
}

func handleDeleteTopic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	topicName := r.URL.Path[len("/topics/"):]
	if topicName == "" {
		http.Error(w, "Topic name is required", http.StatusBadRequest)
		return
	}

	if err := pubSub.DeleteTopic(topicName); err != nil {
		if err.Error() == "topic not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "deleted",
		"topic":  topicName,
	})
}

func handleListTopics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	topics := pubSub.GetTopics()
	topicsList := make([]map[string]interface{}, 0, len(topics))

	for name, subscribers := range topics {
		topicsList = append(topicsList, map[string]interface{}{
			"name":        name,
			"subscribers": subscribers,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"topics": topicsList,
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	health := pubSub.GetHealth()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := pubSub.GetStats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"topics": stats,
	})
}

// Main function
func main() {
	// Create a new mux for better routing
	mux := http.NewServeMux()

	// Set up HTTP routes
	mux.HandleFunc("/ws", handleWebSocket)
	mux.HandleFunc("/topics", handleTopics)          // Combined handler for /topics
	mux.HandleFunc("/topics/", handleTopicsWithPath) // Handle /topics/name patterns
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/stats", handleStats)

	// Start server
	port := ":8080"
	log.Printf("Starting Pub/Sub server on port %s", port)
	log.Fatal(http.ListenAndServe(port, mux))
}

// Combined handler for /topics endpoint
func handleTopics(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handleCreateTopic(w, r)
	case http.MethodGet:
		handleListTopics(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Handler for /topics/name patterns
func handleTopicsWithPath(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract topic name from path
	path := r.URL.Path
	if len(path) <= 8 || path[:8] != "/topics/" {
		http.Error(w, "Invalid topic path", http.StatusBadRequest)
		return
	}

	topicName := path[8:] // Remove "/topics/" prefix
	// Remove trailing slash if present
	if len(topicName) > 0 && topicName[len(topicName)-1] == '/' {
		topicName = topicName[:len(topicName)-1]
	}

	if topicName == "" {
		http.Error(w, "Topic name required for deletion", http.StatusBadRequest)
		return
	}

	handleDeleteTopicByName(w, topicName)
}

// Helper function to handle delete by topic name
func handleDeleteTopicByName(w http.ResponseWriter, topicName string) {
	if err := pubSub.DeleteTopic(topicName); err != nil {
		if err.Error() == "topic not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "deleted",
		"topic":  topicName,
	})
}
