package concurrency

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/AnishMulay/Transaction-Conflict-Visualization/pkg/models"
)

func BenchmarkConcurrentMoves(b *testing.B) {
	gridSize := models.Position{X: 20, Y: 20}
	gameState := models.NewGameState(gridSize)
	controller := NewConcurrencyController(gameState)

	directions := []string{"up", "down", "left", "right"}

	b.RunParallel(func(pb *testing.PB) {
		playerID := fmt.Sprintf("player-%d", time.Now().UnixNano())
		i := 0

		for pb.Next() {
			requestID := fmt.Sprintf("req-%d-%d", time.Now().UnixNano(), i)
			direction := directions[i%len(directions)]

			tx, err := controller.BeginTransaction(playerID, requestID)
			if err != nil {
				b.Errorf("Failed to begin transaction: %v", err)
				continue
			}

			err = controller.ProposeMove(tx.ID, direction)
			if err != nil {
				controller.AbortTransaction(tx.ID)
				continue
			}

			_, err = controller.CommitTransaction(tx.ID)
			if err != nil {
				// Expected in high concurrency scenarios
			}

			i++
		}
	})
}

func TestHighConcurrencyScenario(t *testing.T) {
	gridSize := models.Position{X: 10, Y: 10}
	gameState := models.NewGameState(gridSize)
	controller := NewConcurrencyController(gameState)

	numGoroutines := 50
	movesPerGoroutine := 100

	var wg sync.WaitGroup
	successCount := int64(0)
	conflictCount := int64(0)

	var successMutex sync.Mutex

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			playerID := fmt.Sprintf("player-%d", goroutineID)

			for j := 0; j < movesPerGoroutine; j++ {
				requestID := fmt.Sprintf("req-%d-%d", goroutineID, j)
				direction := []string{"up", "down", "left", "right"}[j%4]

				tx, err := controller.BeginTransaction(playerID, requestID)
				if err != nil {
					t.Errorf("Failed to begin transaction: %v", err)
					continue
				}

				err = controller.ProposeMove(tx.ID, direction)
				if err != nil {
					controller.AbortTransaction(tx.ID)
					continue
				}

				_, err = controller.CommitTransaction(tx.ID)

				successMutex.Lock()
				if err != nil {
					conflictCount++
				} else {
					successCount++
				}
				successMutex.Unlock()
			}
		}(i)
	}

	wg.Wait()

	totalAttempts := successCount + conflictCount
	conflictRate := float64(conflictCount) / float64(totalAttempts) * 100

	t.Logf("High concurrency test results:")
	t.Logf("Total attempts: %d", totalAttempts)
	t.Logf("Successful moves: %d", successCount)
	t.Logf("Conflicts: %d", conflictCount)
	t.Logf("Conflict rate: %.2f%%", conflictRate)

	if successCount == 0 {
		t.Error("No successful moves in high concurrency scenario")
	}

	// Verify final game state consistency
	finalState := gameState.GetState()
	if finalState.Object.Version != successCount+1 {
		t.Errorf("Version mismatch: expected %d, got %d",
			successCount+1, finalState.Object.Version)
	}
}

func TestTransactionTimeout(t *testing.T) {
	gridSize := models.Position{X: 10, Y: 10}
	gameState := models.NewGameState(gridSize)
	controller := NewConcurrencyController(gameState)

	// Start a transaction but don't commit it immediately
	tx, err := controller.BeginTransaction("player1", "req1")
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// Start another transaction that will conflict
	tx2, err := controller.BeginTransaction("player2", "req2")
	if err != nil {
		t.Fatalf("Failed to begin second transaction: %v", err)
	}

	// Commit first transaction
	controller.ProposeMove(tx.ID, "right")
	_, err = controller.CommitTransaction(tx.ID)
	if err != nil {
		t.Fatalf("First commit failed: %v", err)
	}

	// Second transaction should fail due to version mismatch
	controller.ProposeMove(tx2.ID, "left")
	_, err = controller.CommitTransaction(tx2.ID)
	if err == nil {
		t.Error("Second transaction should have failed")
	}
}
