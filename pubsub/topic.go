package pubsub

import (
	"fmt"
	"time"

	"pub-sub-system/models"
)

// TopicManager provides methods to work with Topic models
type TopicManager struct{}

// NewTopicManager creates a new topic manager
func NewTopicManager() *TopicManager {
	return &TopicManager{}
}

// AddMessage adds a message to a topic's history
func (tm *TopicManager) AddMessage(topic *models.Topic, msg *models.Message) {
	topic.Mu.Lock()
	defer topic.Mu.Unlock()

	maxMessages := getEnvInt("TOPIC_HISTORY_SIZE", 100)
	topic.Messages = append(topic.Messages, msg)
	if len(topic.Messages) > maxMessages {
		topic.Messages = topic.Messages[1:] // Remove oldest message
	}
}

// GetLastMessages returns the last N messages
func (tm *TopicManager) GetLastMessages(topic *models.Topic, n int) []*models.Message {
	topic.Mu.RLock()
	defer topic.Mu.RUnlock()

	if n <= 0 || n > len(topic.Messages) {
		n = len(topic.Messages)
	}

	result := make([]*models.Message, n)
	copy(result, topic.Messages[len(topic.Messages)-n:])
	return result
}

// AddSubscriber adds a subscriber to a topic
func (tm *TopicManager) AddSubscriber(topic *models.Topic, sub *models.Subscriber) error {
	topic.Mu.Lock()
	defer topic.Mu.Unlock()

	maxSubs := getEnvInt("MAX_SUBSCRIBERS_PER_TOPIC", 100)
	if len(topic.Subscribers) >= maxSubs {
		return fmt.Errorf("maximum subscribers reached for topic")
	}

	topic.Subscribers[sub.ID] = sub
	return nil
}

// RemoveSubscriber removes a subscriber from a topic
func (tm *TopicManager) RemoveSubscriber(topic *models.Topic, subID string) {
	topic.Mu.Lock()
	defer topic.Mu.Unlock()

	if sub, exists := topic.Subscribers[subID]; exists {
		close(sub.Queue)
		delete(topic.Subscribers, subID)
	}
}

// Broadcast sends a message to all subscribers of a topic
func (tm *TopicManager) Broadcast(topic *models.Topic, msg *models.Message) {
	topic.Mu.RLock()
	subscribers := make([]*models.Subscriber, 0, len(topic.Subscribers))
	for _, sub := range topic.Subscribers {
		subscribers = append(subscribers, sub)
	}
	topic.Mu.RUnlock()

	// Send message to each subscriber
	for _, sub := range subscribers {
		select {
		case sub.Queue <- msg:
			// Message sent successfully
		default:
			// Queue is full, send SLOW_CONSUMER error
			errorMsg := &models.ServerMessage{
				Type: "error",
				Error: &models.Error{
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
