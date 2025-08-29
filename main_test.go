package main

import (
	"fmt"
	"testing"
)

func TestNewPubSubSystem(t *testing.T) {
	ps := NewPubSubSystem()
	
	if ps == nil {
		t.Fatal("NewPubSubSystem returned nil")
	}
	
	if ps.Topics == nil {
		t.Fatal("Topics map is nil")
	}
	
	if ps.StartTime.IsZero() {
		t.Fatal("StartTime is not set")
	}
}

func TestNewTopic(t *testing.T) {
	ps := NewPubSubSystem()
	
	topic, err := ps.NewTopic("test-topic")
	if err != nil {
		t.Fatalf("Failed to create topic: %v", err)
	}
	
	if topic.Name != "test-topic" {
		t.Errorf("Expected topic name 'test-topic', got '%s'", topic.Name)
	}
	
	if len(topic.Subscribers) != 0 {
		t.Errorf("Expected 0 subscribers, got %d", len(topic.Subscribers))
	}
	
	// Test duplicate topic creation
	_, err = ps.NewTopic("test-topic")
	if err == nil {
		t.Fatal("Expected error when creating duplicate topic")
	}
}

func TestTopicAddMessage(t *testing.T) {
	ps := NewPubSubSystem()
	topic, _ := ps.NewTopic("test-topic")
	
	msg := &Message{
		ID:      "test-1",
		Payload: "test payload",
	}
	
	topic.AddMessage(msg)
	
	if len(topic.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(topic.Messages))
	}
	
	if topic.Messages[0].ID != "test-1" {
		t.Errorf("Expected message ID 'test-1', got '%s'", topic.Messages[0].ID)
	}
}

func TestTopicGetLastMessages(t *testing.T) {
	ps := NewPubSubSystem()
	topic, _ := ps.NewTopic("test-topic")
	
	// Add 5 messages
	for i := 0; i < 5; i++ {
		msg := &Message{
			ID:      fmt.Sprintf("test-%d", i),
			Payload: fmt.Sprintf("payload-%d", i),
		}
		topic.AddMessage(msg)
	}
	
	// Test getting last 3 messages
	lastMessages := topic.GetLastMessages(3)
	if len(lastMessages) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(lastMessages))
	}
	
	// Test getting more messages than available
	allMessages := topic.GetLastMessages(10)
	if len(allMessages) != 5 {
		t.Errorf("Expected 5 messages, got %d", len(allMessages))
	}
}

func TestPubSubSystemGetHealth(t *testing.T) {
	ps := NewPubSubSystem()
	
	// Create a topic
	ps.NewTopic("test-topic")
	
	health := ps.GetHealth()
	
	if health["topics"] != 1 {
		t.Errorf("Expected 1 topic, got %v", health["topics"])
	}
	
	if health["subscribers"] != 0 {
		t.Errorf("Expected 0 subscribers, got %v", health["subscribers"])
	}
	
	uptime := health["uptime_sec"].(int)
	if uptime < 0 {
		t.Errorf("Expected positive uptime, got %d", uptime)
	}
}

func TestPubSubSystemGetStats(t *testing.T) {
	ps := NewPubSubSystem()
	
	// Create a topic and add a message
	topic, _ := ps.NewTopic("test-topic")
	topic.AddMessage(&Message{ID: "test-1", Payload: "test"})
	
	stats := ps.GetStats()
	
	topicStats := stats["test-topic"]
	if topicStats == nil {
		t.Fatal("Expected topic stats to exist")
	}
	
	if topicStats["messages"] != 1 {
		t.Errorf("Expected 1 message, got %v", topicStats["messages"])
	}
	
	if topicStats["subscribers"] != 0 {
		t.Errorf("Expected 0 subscribers, got %v", topicStats["subscribers"])
	}
}

func TestPubSubSystemDeleteTopic(t *testing.T) {
	ps := NewPubSubSystem()
	
	// Create a topic
	ps.NewTopic("test-topic")
	
	// Verify topic exists
	if len(ps.Topics) != 1 {
		t.Fatal("Expected 1 topic before deletion")
	}
	
	// Delete the topic
	err := ps.DeleteTopic("test-topic")
	if err != nil {
		t.Fatalf("Failed to delete topic: %v", err)
	}
	
	// Verify topic is deleted
	if len(ps.Topics) != 0 {
		t.Fatal("Expected 0 topics after deletion")
	}
	
	// Test deleting non-existent topic
	err = ps.DeleteTopic("non-existent")
	if err == nil {
		t.Fatal("Expected error when deleting non-existent topic")
	}
}

func TestMessageValidation(t *testing.T) {
	// Test message with empty ID
	msg := &Message{
		ID:      "",
		Payload: "test",
	}
	
	if msg.ID != "" {
		t.Error("Message ID should be empty")
	}
	
	// Test message with payload
	if msg.Payload != "test" {
		t.Error("Message payload should be 'test'")
	}
}

func TestConcurrentTopicOperations(t *testing.T) {
	ps := NewPubSubSystem()
	
	// Test concurrent topic creation
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			topicName := fmt.Sprintf("concurrent-topic-%d", id)
			_, err := ps.NewTopic(topicName)
			if err != nil {
				t.Errorf("Failed to create topic %s: %v", topicName, err)
			}
			done <- true
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Verify all topics were created
	if len(ps.Topics) != 10 {
		t.Errorf("Expected 10 topics, got %d", len(ps.Topics))
	}
}
