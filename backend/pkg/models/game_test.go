package models

import (
	"testing"
	"time"
)

func TestNewGameState(t *testing.T) {
	gridSize := Position{X: 20, Y: 20}
	gameState := NewGameState(gridSize)

	if gameState.Object == nil {
		t.Fatal("Game object should not be nil")
	}

	if gameState.Object.Position.X != 10 || gameState.Object.Position.Y != 10 {
		t.Errorf("Expected object at center (10,10), got (%d,%d)",
			gameState.Object.Position.X, gameState.Object.Position.Y)
	}

	if gameState.Object.Version != 1 {
		t.Errorf("Expected version 1, got %d", gameState.Object.Version)
	}

	if gameState.MaxPlayers != 4 {
		t.Errorf("Expected max players 4, got %d", gameState.MaxPlayers)
	}
}

func TestGameStateSnapshot(t *testing.T) {
	gridSize := Position{X: 20, Y: 20}
	gameState := NewGameState(gridSize)

	// Add a player
	player := &Player{
		ID:        "test-player",
		Name:      "Test Player",
		Color:     "#FF0000",
		Connected: true,
		LastSeen:  time.Now(),
	}
	gameState.Players[player.ID] = player

	snapshot := gameState.GetState()

	if snapshot.Object.Version != gameState.Object.Version {
		t.Errorf("Snapshot version mismatch")
	}

	if len(snapshot.Players) != 1 {
		t.Errorf("Expected 1 player in snapshot, got %d", len(snapshot.Players))
	}

	// Verify snapshot independence
	snapshot.Object.Version = 999
	if gameState.Object.Version == 999 {
		t.Error("Snapshot should not affect original game state")
	}
}
