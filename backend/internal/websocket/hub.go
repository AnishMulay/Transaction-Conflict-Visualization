package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/AnishMulay/Transaction-Conflict-Visualization/internal/concurrency"
	"github.com/AnishMulay/Transaction-Conflict-Visualization/pkg/models"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Hub maintains active WebSocket connections and coordinates message distribution
type Hub struct {
	clients               map[*Client]bool
	playerClients         map[string]*Client
	broadcast             chan []byte
	register              chan *Client
	unregister            chan *Client
	gameState             *models.GameState
	concurrencyController *concurrency.ConcurrencyController
	mu                    sync.RWMutex
}

// Client represents a WebSocket client connection
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	playerID string
	player   *models.Player
	mu       sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub(gameState *models.GameState, controller *concurrency.ConcurrencyController) *Hub {
	return &Hub{
		clients:               make(map[*Client]bool),
		playerClients:         make(map[string]*Client),
		broadcast:             make(chan []byte, 256),
		register:              make(chan *Client),
		unregister:            make(chan *Client),
		gameState:             gameState,
		concurrencyController: controller,
	}
}

// Run starts the hub's main event loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// ServeWS handles WebSocket upgrade requests
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &Client{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 256),
	}

	h.register <- client

	// Start client goroutines
	go client.writePump()
	go client.readPump()
}

func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client] = true
	log.Printf("Client connected. Total clients: %d", len(h.clients))

	// Send current game state to new client
	snapshot := h.gameState.GetState()
	h.sendToClient(client, models.WebSocketMessage{
		Type:      models.MessageTypeGameState,
		Data:      snapshot,
		Timestamp: time.Now(),
	})
}

func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.send)

		if client.playerID != "" {
			delete(h.playerClients, client.playerID)
			h.removePlayer(client.playerID)
		}

		log.Printf("Client disconnected. Total clients: %d", len(h.clients))
	}
}

func (h *Hub) broadcastMessage(message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

func (h *Hub) sendToClient(client *Client, message models.WebSocketMessage) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	select {
	case client.send <- data:
	default:
		close(client.send)
		delete(h.clients, client)
	}
}

func (h *Hub) broadcastGameState() {
	snapshot := h.gameState.GetState()
	message := models.WebSocketMessage{
		Type:      models.MessageTypeGameState,
		Data:      snapshot,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal game state: %v", err)
		return
	}

	h.broadcast <- data
}

func (h *Hub) removePlayer(playerID string) {
	h.gameState.Mu.Lock()
	defer h.gameState.Mu.Unlock()

	if player, exists := h.gameState.Players[playerID]; exists {
		player.Connected = false
		player.LastSeen = time.Now()

		// Remove after grace period
		go func() {
			time.Sleep(30 * time.Second)
			h.gameState.Mu.Lock()
			delete(h.gameState.Players, playerID)
			h.gameState.Mu.Unlock()
			h.broadcastGameState()
		}()
	}
}

// Constants for WebSocket configuration
const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// readPump handles incoming WebSocket messages
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, messageData, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var message models.WebSocketMessage
		if err := json.Unmarshal(messageData, &message); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}

		c.handleMessage(message)
	}
}

// writePump handles outgoing WebSocket messages
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (c *Client) handleMessage(message models.WebSocketMessage) {
	switch message.Type {
	case models.MessageTypeJoin:
		c.handleJoin(message)
	case models.MessageTypeMove:
		c.handleMove(message)
	case models.MessageTypeLeave:
		c.handleLeave()
	default:
		log.Printf("Unknown message type: %s", message.Type)
	}
}

func (c *Client) handleJoin(message models.WebSocketMessage) {
	var joinRequest models.JoinRequest
	data, _ := json.Marshal(message.Data)
	if err := json.Unmarshal(data, &joinRequest); err != nil {
		c.sendError("Invalid join request", "INVALID_JOIN")
		return
	}

	c.hub.mu.Lock()
	defer c.hub.mu.Unlock()

	// Check if game is full
	if len(c.hub.gameState.Players) >= c.hub.gameState.MaxPlayers {
		c.sendError("Game is full", "GAME_FULL")
		return
	}

	// Create new player
	playerID := uuid.New().String()
	colors := []string{"#FF0000", "#00FF00", "#0000FF", "#FFFF00"}
	playerColor := colors[len(c.hub.gameState.Players)%len(colors)]

	player := &models.Player{
		ID:        playerID,
		Name:      joinRequest.PlayerName,
		Color:     playerColor,
		Connected: true,
		LastSeen:  time.Now(),
	}

	c.playerID = playerID
	c.player = player
	c.hub.gameState.Players[playerID] = player
	c.hub.playerClients[playerID] = c

	log.Printf("Player %s (%s) joined the game", player.Name, playerID)

	// Broadcast updated game state
	c.hub.broadcastGameState()
}

func (c *Client) handleMove(message models.WebSocketMessage) {
	if c.playerID == "" {
		c.sendError("Player not registered", "NOT_REGISTERED")
		return
	}

	var moveRequest models.MoveRequest
	data, _ := json.Marshal(message.Data)
	if err := json.Unmarshal(data, &moveRequest); err != nil {
		c.sendError("Invalid move request", "INVALID_MOVE")
		return
	}

	// Begin optimistic transaction
	transaction, err := c.hub.concurrencyController.BeginTransaction(c.playerID, moveRequest.RequestID)
	if err != nil {
		c.sendError("Failed to begin transaction", "TRANSACTION_ERROR")
		return
	}

	// Propose the move
	if err := c.hub.concurrencyController.ProposeMove(transaction.ID, moveRequest.Direction); err != nil {
		c.hub.concurrencyController.AbortTransaction(transaction.ID)
		c.sendError(err.Error(), "INVALID_MOVE")
		return
	}

	// Attempt to commit
	snapshot, err := c.hub.concurrencyController.CommitTransaction(transaction.ID)
	if err != nil {
		// Handle concurrency conflict
		c.sendConflict(moveRequest.RequestID, err.Error())
		c.hub.broadcastGameState() // Send current state to all clients
		return
	}

	log.Printf("Snapshot after commit: %+v", snapshot)

	// Update last seen
	c.player.LastSeen = time.Now()

	// Broadcast successful move
	c.hub.broadcastGameState()
}

func (c *Client) handleLeave() {
	c.hub.unregister <- c
}

func (c *Client) sendError(message, code string) {
	errorMsg := models.WebSocketMessage{
		Type: models.MessageTypeError,
		Data: models.ErrorResponse{
			Message: message,
			Code:    code,
		},
		Timestamp: time.Now(),
	}
	c.hub.sendToClient(c, errorMsg)
}

func (c *Client) sendConflict(requestID, message string) {
	snapshot := c.hub.gameState.GetState()
	conflictMsg := models.WebSocketMessage{
		Type: models.MessageTypeConflict,
		Data: models.ConflictResponse{
			Message:         message,
			ExpectedVersion: snapshot.Object.Version,
			ActualVersion:   snapshot.Object.Version,
			RequestID:       requestID,
			Timestamp:       time.Now(),
		},
		Timestamp: time.Now(),
	}
	c.hub.sendToClient(c, conflictMsg)
}
