package models

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// Position represents coordinates on the game grid
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// GameObject represents the shared object that players manipulate
type GameObject struct {
	ID          string    `json:"id"`
	Position    Position  `json:"position"`
	Version     int64     `json:"version"`
	LastUpdated time.Time `json:"lastUpdated"`
}

// Player represents a connected player
type Player struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	Connected bool      `json:"connected"`
	LastSeen  time.Time `json:"lastSeen"`
}

// GameState represents the complete state of the game
type GameState struct {
	Mu         sync.RWMutex
	Object     *GameObject        `json:"object"`
	Players    map[string]*Player `json:"players"`
	Version    int64              `json:"version"`
	MaxPlayers int                `json:"maxPlayers"`
	GridSize   Position           `json:"gridSize"`
}

// NewGameState creates a new game state with initial values
func NewGameState(gridSize Position) *GameState {
	return &GameState{
		Object: &GameObject{
			ID:          uuid.New().String(),
			Position:    Position{X: gridSize.X / 2, Y: gridSize.Y / 2},
			Version:     1,
			LastUpdated: time.Now(),
		},
		Players:    make(map[string]*Player),
		Version:    1,
		MaxPlayers: 4,
		GridSize:   gridSize,
	}
}

// GetState returns a thread-safe copy of the current state
func (gs *GameState) GetState() GameStateSnapshot {
	gs.Mu.RLock()
	defer gs.Mu.RUnlock()

	playersSnapshot := make(map[string]*Player)
	for id, player := range gs.Players {
		playersSnapshot[id] = &Player{
			ID:        player.ID,
			Name:      player.Name,
			Color:     player.Color,
			Connected: player.Connected,
			LastSeen:  player.LastSeen,
		}
	}

	return GameStateSnapshot{
		Object: &GameObject{
			ID:          gs.Object.ID,
			Position:    gs.Object.Position,
			Version:     gs.Object.Version,
			LastUpdated: gs.Object.LastUpdated,
		},
		Players:    playersSnapshot,
		Version:    gs.Version,
		MaxPlayers: gs.MaxPlayers,
		GridSize:   gs.GridSize,
	}
}

// GameStateSnapshot represents a read-only snapshot of game state
type GameStateSnapshot struct {
	Object     *GameObject        `json:"object"`
	Players    map[string]*Player `json:"players"`
	Version    int64              `json:"version"`
	MaxPlayers int                `json:"maxPlayers"`
	GridSize   Position           `json:"gridSize"`
}
