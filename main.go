package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// Set up HTTP routes - order matters in Go ServeMux (most specific first)
	mux.HandleFunc("/ws", wsHandler.HandleWebSocket)
	mux.HandleFunc("/topics/", httpHandler.HandleDeleteTopic) // DELETE for delete (more specific first)
	mux.HandleFunc("/topics", httpHandler.HandleTopics)       // Combined handler for POST/GET
	mux.HandleFunc("/health", httpHandler.HandleHealth)
	mux.HandleFunc("/stats", httpHandler.HandleStats)
	mux.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "API endpoint working!", "status": "success"}`))
	}) // Test API endpoint

	// Start server - use Railway's PORT environment variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	port = ":" + port
	log.Printf("Starting Pub/Sub server on port %s", port)

	// Create HTTP server with graceful shutdown
	server := &http.Server{
		Addr:    port,
		Handler: middleware.CORS(mux),
	}

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited gracefully")
}
