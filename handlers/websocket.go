package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"pub-sub-system/models"
	"pub-sub-system/pubsub"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// WebSocketHandler manages WebSocket connections
type WebSocketHandler struct {
	pubSubSystem *pubsub.PubSubSystem
	topicManager *pubsub.TopicManager
	subManager   *pubsub.SubscriberManager
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(pubSubSystem *pubsub.PubSubSystem) *WebSocketHandler {
	return &WebSocketHandler{
		pubSubSystem: pubSubSystem,
		topicManager: pubsub.NewTopicManager(),
		subManager:   pubsub.NewSubscriberManager(),
	}
}

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// HandleWebSocket handles WebSocket connections
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start heartbeat goroutine
	go h.startHeartbeat(conn, ctx)

	// Process incoming messages
	h.processMessages(conn, ctx)
}

// startHeartbeat sends periodic heartbeat messages
func (h *WebSocketHandler) startHeartbeat(conn *websocket.Conn, ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			infoMsg := &models.ServerMessage{
				Type: "info",
				Msg:  "ping",
				TS:   time.Now().UTC().Format(time.RFC3339),
			}
			conn.WriteJSON(infoMsg)
		case <-ctx.Done():
			return
		}
	}
}

// processMessages processes incoming WebSocket messages
func (h *WebSocketHandler) processMessages(conn *websocket.Conn, ctx context.Context) {
	for {
		var clientMsg models.ClientMessage
		if err := conn.ReadJSON(&clientMsg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		h.handleClientMessage(conn, &clientMsg, ctx)
	}
}

// handleClientMessage processes incoming client messages
func (h *WebSocketHandler) handleClientMessage(conn *websocket.Conn, msg *models.ClientMessage, ctx context.Context) {
	switch msg.Type {
	case "subscribe":
		h.handleSubscribe(conn, msg, ctx)
	case "unsubscribe":
		h.handleUnsubscribe(conn, msg)
	case "publish":
		h.handlePublish(conn, msg)
	case "ping":
		h.handlePing(conn, msg)
	default:
		h.sendError(conn, "BAD_REQUEST", "Invalid message type", msg.RequestID)
	}
}

// handleSubscribe handles subscription requests
func (h *WebSocketHandler) handleSubscribe(conn *websocket.Conn, msg *models.ClientMessage, ctx context.Context) {
	if msg.Topic == "" || msg.ClientID == "" {
		h.sendError(conn, "BAD_REQUEST", "Topic and client_id are required", msg.RequestID)
		return
	}

	topic, exists := h.pubSubSystem.GetTopic(msg.Topic)
	if !exists {
		h.sendError(conn, "TOPIC_NOT_FOUND", "Topic does not exist", msg.RequestID)
		return
	}

	sub := h.subManager.NewSubscriber(msg.ClientID, msg.Topic, conn)
	if err := h.topicManager.AddSubscriber(topic, sub); err != nil {
		h.sendError(conn, "INTERNAL", err.Error(), msg.RequestID)
		return
	}

	// Start message processor
	h.subManager.StartMessageProcessor(sub, ctx)

	// Send acknowledgment
	ack := &models.ServerMessage{
		Type:      "ack",
		RequestID: msg.RequestID,
		Topic:     msg.Topic,
		Status:    "ok",
		TS:        time.Now().UTC().Format(time.RFC3339),
	}
	conn.WriteJSON(ack)

	// Send historical messages if requested
	if msg.LastN > 0 {
		historicalMessages := h.topicManager.GetLastMessages(topic, msg.LastN)
		for _, histMsg := range historicalMessages {
			eventMsg := &models.ServerMessage{
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
func (h *WebSocketHandler) handleUnsubscribe(conn *websocket.Conn, msg *models.ClientMessage) {
	if msg.Topic == "" || msg.ClientID == "" {
		h.sendError(conn, "BAD_REQUEST", "Topic and client_id are required", msg.RequestID)
		return
	}

	topic, exists := h.pubSubSystem.GetTopic(msg.Topic)
	if !exists {
		h.sendError(conn, "TOPIC_NOT_FOUND", "Topic does not exist", msg.RequestID)
		return
	}

	h.topicManager.RemoveSubscriber(topic, msg.ClientID)

	ack := &models.ServerMessage{
		Type:      "ack",
		RequestID: msg.RequestID,
		Topic:     msg.Topic,
		Status:    "ok",
		TS:        time.Now().UTC().Format(time.RFC3339),
	}
	conn.WriteJSON(ack)
}

// handlePublish handles publish requests
func (h *WebSocketHandler) handlePublish(conn *websocket.Conn, msg *models.ClientMessage) {
	if msg.Topic == "" || msg.Message == nil {
		h.sendError(conn, "BAD_REQUEST", "Topic and message are required", msg.RequestID)
		return
	}

	// Validate message ID
	if msg.Message.ID == "" {
		msg.Message.ID = uuid.New().String()
	} else {
		// Validate UUID format if provided
		if _, err := uuid.Parse(msg.Message.ID); err != nil {
			h.sendError(conn, "BAD_REQUEST", "message.id must be a valid UUID", msg.RequestID)
			return
		}
	}

	topic, exists := h.pubSubSystem.GetTopic(msg.Topic)
	if !exists {
		h.sendError(conn, "TOPIC_NOT_FOUND", "Topic does not exist", msg.RequestID)
		return
	}

	// Add message to topic history
	h.topicManager.AddMessage(topic, msg.Message)

	// Broadcast to all subscribers
	h.topicManager.Broadcast(topic, msg.Message)

	// Send acknowledgment
	ack := &models.ServerMessage{
		Type:      "ack",
		RequestID: msg.RequestID,
		Topic:     msg.Topic,
		Status:    "ok",
		TS:        time.Now().UTC().Format(time.RFC3339),
	}
	conn.WriteJSON(ack)
}

// handlePing handles ping requests
func (h *WebSocketHandler) handlePing(conn *websocket.Conn, msg *models.ClientMessage) {
	pong := &models.ServerMessage{
		Type:      "pong",
		RequestID: msg.RequestID,
		TS:        time.Now().UTC().Format(time.RFC3339),
	}
	conn.WriteJSON(pong)
}

// sendError sends an error message to the client
func (h *WebSocketHandler) sendError(conn *websocket.Conn, code, message, requestID string) {
	errorMsg := &models.ServerMessage{
		Type:      "error",
		RequestID: requestID,
		Error: &models.Error{
			Code:    code,
			Message: message,
		},
		TS: time.Now().UTC().Format(time.RFC3339),
	}
	conn.WriteJSON(errorMsg)
}
