package handlers

import (
	"encoding/json"
	"net/http"

	"pub-sub-system/pubsub"
)

// HTTPHandler manages HTTP REST API endpoints
type HTTPHandler struct {
	pubSubSystem *pubsub.PubSubSystem
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(pubSubSystem *pubsub.PubSubSystem) *HTTPHandler {
	return &HTTPHandler{
		pubSubSystem: pubSubSystem,
	}
}

// HandleTopics handles all topic operations based on HTTP method
func (h *HTTPHandler) HandleTopics(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.handleCreateTopic(w, r)
	case http.MethodGet:
		h.handleListTopics(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleCreateTopic handles topic creation
func (h *HTTPHandler) handleCreateTopic(w http.ResponseWriter, r *http.Request) {
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

	topic, err := h.pubSubSystem.NewTopic(req.Name)
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

// HandleDeleteTopic handles topic deletion
func (h *HTTPHandler) HandleDeleteTopic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	topicName := r.URL.Path[len("/topics/"):]
	if topicName == "" {
		http.Error(w, "Topic name is required", http.StatusBadRequest)
		return
	}

	if err := h.pubSubSystem.DeleteTopic(topicName); err != nil {
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

// handleListTopics handles topic listing
func (h *HTTPHandler) handleListTopics(w http.ResponseWriter, r *http.Request) {
	topics := h.pubSubSystem.GetTopics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(topics)
}

// HandleHealth handles health check
func (h *HTTPHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	health := h.pubSubSystem.GetHealth()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// HandleStats handles statistics
func (h *HTTPHandler) HandleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := h.pubSubSystem.GetStats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
