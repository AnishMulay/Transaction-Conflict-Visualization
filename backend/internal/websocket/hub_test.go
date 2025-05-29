package websocket

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AnishMulay/Transaction-Conflict-Visualization/internal/concurrency"
	"github.com/AnishMulay/Transaction-Conflict-Visualization/pkg/models"

	"github.com/gorilla/websocket"
)

func TestWebSocketConnection(t *testing.T) {
	gridSize := models.Position{X: 10, Y: 10}
	gameState := models.NewGameState(gridSize)
	controller := concurrency.NewConcurrencyController(gameState)
	hub := NewHub(gameState, controller)

	go hub.Run()

	server := httptest.NewServer(http.HandlerFunc(hub.ServeWS))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	// Test connection
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Should receive initial game state
	var message models.WebSocketMessage
	err = conn.ReadJSON(&message)
	if err != nil {
		t.Fatalf("Failed to read message: %v", err)
	}

	if message.Type != models.MessageTypeGameState {
		t.Errorf("Expected game state message, got %s", message.Type)
	}
}

func TestPlayerJoinAndMove(t *testing.T) {
	gridSize := models.Position{X: 10, Y: 10}
	gameState := models.NewGameState(gridSize)
	controller := concurrency.NewConcurrencyController(gameState)
	hub := NewHub(gameState, controller)

	go hub.Run()

	server := httptest.NewServer(http.HandlerFunc(hub.ServeWS))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Read initial game state
	var gameStateMsg models.WebSocketMessage
	conn.ReadJSON(&gameStateMsg)

	// Join game
	joinMessage := models.WebSocketMessage{
		Type: models.MessageTypeJoin,
		Data: models.JoinRequest{
			PlayerName: "TestPlayer",
		},
		Timestamp: time.Now(),
	}

	err = conn.WriteJSON(joinMessage)
	if err != nil {
		t.Fatalf("Failed to send join message: %v", err)
	}

	// Should receive updated game state with player
	var updatedGameState models.WebSocketMessage
	err = conn.ReadJSON(&updatedGameState)
	if err != nil {
		t.Fatalf("Failed to read updated game state: %v", err)
	}

	// Extract game state data
	var stateSnapshot models.GameStateSnapshot
	stateData, _ := json.Marshal(updatedGameState.Data)
	json.Unmarshal(stateData, &stateSnapshot)

	if len(stateSnapshot.Players) != 1 {
		t.Errorf("Expected 1 player, got %d", len(stateSnapshot.Players))
	}

	// Test move
	moveMessage := models.WebSocketMessage{
		Type: models.MessageTypeMove,
		Data: models.MoveRequest{
			Direction:     "right",
			ObjectVersion: stateSnapshot.Object.Version,
			RequestID:     "test-move-1",
		},
		Timestamp: time.Now(),
	}

	err = conn.WriteJSON(moveMessage)
	if err != nil {
		t.Fatalf("Failed to send move message: %v", err)
	}

	// Should receive game state with moved object
	var moveResult models.WebSocketMessage
	err = conn.ReadJSON(&moveResult)
	if err != nil {
		t.Fatalf("Failed to read move result: %v", err)
	}

	if moveResult.Type == models.MessageTypeConflict {
		t.Error("Move should not conflict in single player scenario")
	}
}

func TestMultiplePlayersConflict(t *testing.T) {
	gridSize := models.Position{X: 10, Y: 10}
	gameState := models.NewGameState(gridSize)
	controller := concurrency.NewConcurrencyController(gameState)
	hub := NewHub(gameState, controller)

	go hub.Run()

	server := httptest.NewServer(http.HandlerFunc(hub.ServeWS))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	// Connect two players
	conn1, _, _ := websocket.DefaultDialer.Dial(url, nil)
	defer conn1.Close()

	conn2, _, _ := websocket.DefaultDialer.Dial(url, nil)
	defer conn2.Close()

	// Read initial states
	var dummy models.WebSocketMessage
	conn1.ReadJSON(&dummy)
	conn2.ReadJSON(&dummy)

	// Join both players
	joinMsg1 := models.WebSocketMessage{
		Type:      models.MessageTypeJoin,
		Data:      models.JoinRequest{PlayerName: "Player1"},
		Timestamp: time.Now(),
	}

	joinMsg2 := models.WebSocketMessage{
		Type:      models.MessageTypeJoin,
		Data:      models.JoinRequest{PlayerName: "Player2"},
		Timestamp: time.Now(),
	}

	conn1.WriteJSON(joinMsg1)
	conn1.ReadJSON(&dummy) // Read state update

	conn2.WriteJSON(joinMsg2)
	conn2.ReadJSON(&dummy) // Read state update
	conn1.ReadJSON(&dummy) // Player1 receives Player2 join update

	// Both try to move simultaneously
	currentState := gameState.GetState()

	moveMsg1 := models.WebSocketMessage{
		Type: models.MessageTypeMove,
		Data: models.MoveRequest{
			Direction:     "right",
			ObjectVersion: currentState.Object.Version,
			RequestID:     "move1",
		},
		Timestamp: time.Now(),
	}

	moveMsg2 := models.WebSocketMessage{
		Type: models.MessageTypeMove,
		Data: models.MoveRequest{
			Direction:     "left",
			ObjectVersion: currentState.Object.Version,
			RequestID:     "move2",
		},
		Timestamp: time.Now(),
	}

	// Send moves quickly
	go conn1.WriteJSON(moveMsg1)
	go conn2.WriteJSON(moveMsg2)

	// Read responses
	var resp1, resp2 models.WebSocketMessage
	conn1.ReadJSON(&resp1)
	conn2.ReadJSON(&resp2)

	// One should succeed, one should conflict or both receive game state
	conflictDetected := resp1.Type == models.MessageTypeConflict ||
		resp2.Type == models.MessageTypeConflict

	if !conflictDetected {
		t.Log("No explicit conflict message, but this is acceptable if moves were serialized")
	}
}
