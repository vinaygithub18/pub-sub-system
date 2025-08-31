package pubsub

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"pub-sub-system/models"
)

// PubSubSystem manages the entire pub/sub system
type PubSubSystem struct {
	Topics         map[string]*models.Topic
	StartTime      time.Time
	Mu             sync.RWMutex
	MaxTopics      int
	MaxSubscribers int
}

// NewPubSubSystem creates a new pub/sub system
func NewPubSubSystem() *PubSubSystem {
	maxTopics := getEnvInt("MAX_TOPICS", 100)
	maxSubscribers := getEnvInt("MAX_SUBSCRIBERS_PER_TOPIC", 100)

	return &PubSubSystem{
		Topics:         make(map[string]*models.Topic),
		StartTime:      time.Now(),
		MaxTopics:      maxTopics,
		MaxSubscribers: maxSubscribers,
	}
}

// getEnvInt gets an environment variable as an integer with a default value
func getEnvInt(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// NewTopic creates a new topic
func (ps *PubSubSystem) NewTopic(name string) (*models.Topic, error) {
	ps.Mu.Lock()
	defer ps.Mu.Unlock()

	if len(ps.Topics) >= ps.MaxTopics {
		return nil, fmt.Errorf("maximum topics reached")
	}

	if _, exists := ps.Topics[name]; exists {
		return nil, fmt.Errorf("topic already exists")
	}

	topic := &models.Topic{
		Name:        name,
		Subscribers: make(map[string]*models.Subscriber),
		Messages:    make([]*models.Message, 0),
		MaxMessages: 100,
	}

	ps.Topics[name] = topic
	return topic, nil
}

// DeleteTopic deletes a topic and unsubscribes all subscribers
func (ps *PubSubSystem) DeleteTopic(name string) error {
	ps.Mu.Lock()
	defer ps.Mu.Unlock()

	topic, exists := ps.Topics[name]
	if !exists {
		return fmt.Errorf("topic not found")
	}

	// Close all subscriber connections
	for _, sub := range topic.Subscribers {
		close(sub.Queue)
		sub.Conn.Close()
	}

	delete(ps.Topics, name)
	return nil
}

// GetTopic returns a topic by name
func (ps *PubSubSystem) GetTopic(name string) (*models.Topic, bool) {
	ps.Mu.RLock()
	defer ps.Mu.RUnlock()

	topic, exists := ps.Topics[name]
	return topic, exists
}

// GetStats returns system statistics
func (ps *PubSubSystem) GetStats() map[string]interface{} {
	ps.Mu.RLock()
	defer ps.Mu.RUnlock()

	stats := make(map[string]interface{})
	topicStats := make(map[string]interface{})

	totalSubscribers := 0
	for name, topic := range ps.Topics {
		subscriberCount := len(topic.Subscribers)
		totalSubscribers += subscriberCount

		topicStats[name] = map[string]interface{}{
			"messages":    len(topic.Messages),
			"subscribers": subscriberCount,
		}
	}

	stats["topics"] = topicStats
	return stats
}

// GetHealth returns system health information
func (ps *PubSubSystem) GetHealth() map[string]interface{} {
	ps.Mu.RLock()
	defer ps.Mu.RUnlock()

	totalSubscribers := 0
	for _, topic := range ps.Topics {
		totalSubscribers += len(topic.Subscribers)
	}

	return map[string]interface{}{
		"uptime_sec":  int(time.Since(ps.StartTime).Seconds()),
		"topics":      len(ps.Topics),
		"subscribers": totalSubscribers,
	}
}

// GetTopics returns a list of all topics with subscriber counts
func (ps *PubSubSystem) GetTopics() map[string]interface{} {
	ps.Mu.RLock()
	defer ps.Mu.RUnlock()

	topics := make(map[string]interface{})
	for name, topic := range ps.Topics {
		topics[name] = map[string]interface{}{
			"subscribers": len(topic.Subscribers),
		}
	}

	return map[string]interface{}{
		"topics": topics,
	}
}
