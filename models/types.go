package models

import (
	"sync"
	"time"
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
	Mu          sync.RWMutex
}

// Subscriber represents a WebSocket client subscription
type Subscriber struct {
	ID       string
	Conn     WebSocketConn
	Topic    string
	Queue    chan *Message
	MaxQueue int
}

// WebSocketConn is an interface for WebSocket connections
type WebSocketConn interface {
	WriteJSON(v interface{}) error
	Close() error
}

// PubSubSystem manages the entire pub/sub system
type PubSubSystem struct {
	Topics         map[string]*Topic
	StartTime      time.Time
	Mu             sync.RWMutex
	MaxTopics      int
	MaxSubscribers int
}
