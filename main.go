package main

import (
	"log"
	"net/http"
	"os"

	"pub-sub-system/handlers"
	"pub-sub-system/middleware"
	"pub-sub-system/pubsub"
)

func main() {
	// Initialize the pub/sub system
	pubSubSystem := pubsub.NewPubSubSystem()

	// Initialize handlers
	wsHandler := handlers.NewWebSocketHandler(pubSubSystem)
	httpHandler := handlers.NewHTTPHandler(pubSubSystem)

	// Create a new mux for better routing
	mux := http.NewServeMux()

	// Set up HTTP routes
	mux.HandleFunc("/ws", wsHandler.HandleWebSocket)
	mux.HandleFunc("/topics", httpHandler.HandleTopics)       // Combined handler for POST/GET
	mux.HandleFunc("/topics/", httpHandler.HandleDeleteTopic) // DELETE for delete
	mux.HandleFunc("/health", httpHandler.HandleHealth)
	mux.HandleFunc("/stats", httpHandler.HandleStats)

	// Start server - use Railway's PORT environment variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	port = ":" + port
	log.Printf("Starting Pub/Sub server on port %s", port)

	// Apply CORS middleware to all routes
	handler := middleware.CORS(mux)
	log.Fatal(http.ListenAndServe(port, handler))
}
