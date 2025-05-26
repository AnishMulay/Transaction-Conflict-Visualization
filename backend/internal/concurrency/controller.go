package concurrency

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/AnishMulay/Transaction-Conflict-Visualization/pkg/models"
)

var (
	ErrVersionMismatch = errors.New("version mismatch: concurrent modification detected")
	ErrInvalidMove     = errors.New("invalid move: out of bounds")
	ErrNoTransaction   = errors.New("no active transaction")
)

// ConcurrencyController manages optimistic concurrency control
type ConcurrencyController struct {
	mu                 sync.RWMutex
	gameState          *models.GameState
	activeTransactions map[string]*Transaction
	conflictStats      ConflictStats
}

// Transaction represents an optimistic transaction
type Transaction struct {
	ID              string
	PlayerID        string
	StartTime       time.Time
	InitialVersion  int64
	ProposedChanges *models.GameObject
	RequestID       string
}

// ConflictStats tracks concurrency conflicts for analysis
type ConflictStats struct {
	TotalTransactions int64
	ConflictCount     int64
	SuccessfulMoves   int64
	AverageLatency    time.Duration
}

// NewConcurrencyController creates a new concurrency controller
func NewConcurrencyController(gameState *models.GameState) *ConcurrencyController {
	return &ConcurrencyController{
		gameState:          gameState,
		activeTransactions: make(map[string]*Transaction),
		conflictStats:      ConflictStats{},
	}
}

// BeginTransaction starts an optimistic transaction for a move
func (cc *ConcurrencyController) BeginTransaction(playerID, requestID string) (*Transaction, error) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	snapshot := cc.gameState.GetState()

	transaction := &Transaction{
		ID:             fmt.Sprintf("%s-%s-%d", playerID, requestID, time.Now().UnixNano()),
		PlayerID:       playerID,
		StartTime:      time.Now(),
		InitialVersion: snapshot.Object.Version,
		RequestID:      requestID,
	}

	cc.activeTransactions[transaction.ID] = transaction
	cc.conflictStats.TotalTransactions++

	return transaction, nil
}

// ProposeMove validates and prepares a move within a transaction
func (cc *ConcurrencyController) ProposeMove(transactionID, direction string) error {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	transaction, exists := cc.activeTransactions[transactionID]
	if !exists {
		return ErrNoTransaction
	}

	snapshot := cc.gameState.GetState()
	newPosition := calculateNewPosition(snapshot.Object.Position, direction, snapshot.GridSize)

	if !isValidPosition(newPosition, snapshot.GridSize) {
		return ErrInvalidMove
	}

	transaction.ProposedChanges = &models.GameObject{
		ID:          snapshot.Object.ID,
		Position:    newPosition,
		Version:     snapshot.Object.Version + 1,
		LastUpdated: time.Now(),
	}

	return nil
}

// CommitTransaction attempts to commit the transaction using optimistic concurrency
func (cc *ConcurrencyController) CommitTransaction(transactionID string) (*models.GameStateSnapshot, error) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	transaction, exists := cc.activeTransactions[transactionID]
	if !exists {
		return nil, ErrNoTransaction
	}

	defer delete(cc.activeTransactions, transactionID)

	// Critical section: check version and commit atomically
	cc.gameState.Mu.Lock()
	defer cc.gameState.Mu.Unlock()

	currentVersion := cc.gameState.Object.Version

	// Optimistic concurrency check
	if transaction.InitialVersion != currentVersion {
		cc.conflictStats.ConflictCount++
		return nil, fmt.Errorf("%w: expected version %d, got %d",
			ErrVersionMismatch, transaction.InitialVersion, currentVersion)
	}

	// Commit the changes
	cc.gameState.Object.Position = transaction.ProposedChanges.Position
	cc.gameState.Object.Version = transaction.ProposedChanges.Version
	cc.gameState.Object.LastUpdated = transaction.ProposedChanges.LastUpdated
	cc.gameState.Version++

	cc.conflictStats.SuccessfulMoves++
	cc.conflictStats.AverageLatency = updateAverageLatency(
		cc.conflictStats.AverageLatency,
		time.Since(transaction.StartTime),
		cc.conflictStats.SuccessfulMoves,
	)

	return &models.GameStateSnapshot{
		Object: &models.GameObject{
			ID:          cc.gameState.Object.ID,
			Position:    cc.gameState.Object.Position,
			Version:     cc.gameState.Object.Version,
			LastUpdated: cc.gameState.Object.LastUpdated,
		},
		Players:    make(map[string]*models.Player), // Copy current players
		Version:    cc.gameState.Version,
		MaxPlayers: cc.gameState.MaxPlayers,
		GridSize:   cc.gameState.GridSize,
	}, nil
}

// AbortTransaction cancels a transaction
func (cc *ConcurrencyController) AbortTransaction(transactionID string) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	delete(cc.activeTransactions, transactionID)
}

// GetConflictStats returns current concurrency statistics
func (cc *ConcurrencyController) GetConflictStats() ConflictStats {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.conflictStats
}

// Helper functions
func calculateNewPosition(current models.Position, direction string, gridSize models.Position) models.Position {
	newPos := current

	switch direction {
	case "up":
		newPos.Y = max(0, current.Y-1)
	case "down":
		newPos.Y = min(gridSize.Y-1, current.Y+1)
	case "left":
		newPos.X = max(0, current.X-1)
	case "right":
		newPos.X = min(gridSize.X-1, current.X+1)
	}

	return newPos
}

func isValidPosition(pos models.Position, gridSize models.Position) bool {
	return pos.X >= 0 && pos.X < gridSize.X && pos.Y >= 0 && pos.Y < gridSize.Y
}

func updateAverageLatency(currentAvg time.Duration, newLatency time.Duration, count int64) time.Duration {
	if count == 1 {
		return newLatency
	}
	return time.Duration((int64(currentAvg)*(count-1) + int64(newLatency)) / count)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
