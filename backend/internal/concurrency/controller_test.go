package concurrency

import (
	"errors"
	"testing"

	"github.com/AnishMulay/Transaction-Conflict-Visualization/pkg/models"
)

func TestConcurrencyController(t *testing.T) {
	gridSize := models.Position{X: 10, Y: 10}
	gameState := models.NewGameState(gridSize)
	controller := NewConcurrencyController(gameState)

	// Test successful transaction
	transaction, err := controller.BeginTransaction("player1", "req1")
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	err = controller.ProposeMove(transaction.ID, "right")
	if err != nil {
		t.Fatalf("Failed to propose move: %v", err)
	}

	snapshot, err := controller.CommitTransaction(transaction.ID)
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	expectedPos := models.Position{X: 6, Y: 5} // moved right from center (5,5)
	if snapshot.Object.Position != expectedPos {
		t.Errorf("Expected position %v, got %v", expectedPos, snapshot.Object.Position)
	}
}

func TestConcurrencyConflict(t *testing.T) {
	gridSize := models.Position{X: 10, Y: 10}
	gameState := models.NewGameState(gridSize)
	controller := NewConcurrencyController(gameState)

	// Start two concurrent transactions
	tx1, _ := controller.BeginTransaction("player1", "req1")
	tx2, _ := controller.BeginTransaction("player2", "req2")

	// Both propose moves
	controller.ProposeMove(tx1.ID, "right")
	controller.ProposeMove(tx2.ID, "left")

	// First commit should succeed
	_, err1 := controller.CommitTransaction(tx1.ID)
	if err1 != nil {
		t.Fatalf("First commit should succeed: %v", err1)
	}

	// Second commit should fail due to version mismatch
	_, err2 := controller.CommitTransaction(tx2.ID)
	if err2 == nil {
		t.Fatal("Second commit should fail due to version conflict")
	}

	if err2 != ErrVersionMismatch && !errors.Is(err2, ErrVersionMismatch) {
		t.Errorf("Expected version mismatch error, got: %v", err2)
	}
}
