package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/AnishMulay/Transaction-Conflict-Visualization/internal/concurrency"
	"github.com/AnishMulay/Transaction-Conflict-Visualization/internal/websocket"
	"github.com/AnishMulay/Transaction-Conflict-Visualization/pkg/models"
)

func main() {
	fmt.Println("Real-time Multiplayer Game Server")
	fmt.Println("Starting server on :8080...")

	// Initialize game state
	gridSize := models.Position{X: 20, Y: 20}
	gameState := models.NewGameState(gridSize)

	// Initialize concurrency controller
	controller := concurrency.NewConcurrencyController(gameState)

	// Initialize WebSocket hub
	hub := websocket.NewHub(gameState, controller)
	go hub.Run()

	// Routes
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Server is running"))
	})

	http.HandleFunc("/ws", hub.ServeWS)

	// Serve static files for frontend
	http.Handle("/", http.FileServer(http.Dir("../../frontend/build/")))

	log.Printf("Server starting on :8080")
	log.Printf("WebSocket endpoint: ws://localhost:8080/ws")
	log.Printf("Health check: http://localhost:8080/health")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
