package pubsub

import (
	"context"
	"log"
	"time"

	"pub-sub-system/models"
)

// SubscriberManager provides methods to work with Subscriber models
type SubscriberManager struct{}

// NewSubscriberManager creates a new subscriber manager
func NewSubscriberManager() *SubscriberManager {
	return &SubscriberManager{}
}

// NewSubscriber creates a new subscriber
func (sm *SubscriberManager) NewSubscriber(id, topic string, conn models.WebSocketConn) *models.Subscriber {
	queueSize := getEnvInt("SUBSCRIBER_QUEUE_SIZE", 100)
	return &models.Subscriber{
		ID:       id,
		Conn:     conn,
		Topic:    topic,
		Queue:    make(chan *models.Message, queueSize), // Bounded queue
		MaxQueue: queueSize,
	}
}

// StartMessageProcessor starts processing messages for a subscriber
func (sm *SubscriberManager) StartMessageProcessor(sub *models.Subscriber, ctx context.Context) {
	go func() {
		defer func() {
			sub.Conn.Close()
			// Remove from topic
			// This will be handled by the topic manager
		}()

		for {
			select {
			case msg := <-sub.Queue:
				if msg == nil {
					return // Channel closed
				}

				serverMsg := &models.ServerMessage{
					Type:    "event",
					Topic:   sub.Topic,
					Message: msg,
					TS:      time.Now().UTC().Format(time.RFC3339),
				}

				if err := sub.Conn.WriteJSON(serverMsg); err != nil {
					log.Printf("Error sending message to subscriber %s: %v", sub.ID, err)
					return
				}

			case <-ctx.Done():
				return
			}
		}
	}()
}
