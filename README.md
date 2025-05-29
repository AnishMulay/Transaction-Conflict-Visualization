# Transaction Conflict Visualization

> A live demonstration of **optimistic concurrency control** in action. Watch multiple players compete to move the same object simultaneously, and see how conflicts are resolved in real-time.

 ![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=flat&logo=go&logoColor=white) ![React](https://img.shields.io/badge/react-%2320232a.svg?style=flat&logo=react&logoColor=%2361DAFB) ![WebSocket](https://img.shields.io/badge/websocket-realtime-blue)

### What this project is about

Ever wondered how databases handle multiple users trying to update the same record simultaneously? This game brings that concept to life! Multiple players can move a shared object on a grid, and you'll see **optimistic concurrency control** resolve conflicts in real-time.

**Perfect for:**
- Understanding concurrency patterns
- Understanding real-time WebSocket communication
- Seeing conflict resolution in action
- Having fun with distributed systems concepts

### Quick Start

**Prerequisites:** Go 1.21+ and Node.js 16+

```bash
# 1. Clone and setup
git clone <your-repo-url>
cd realtime-multiplayer-game

# 2. Build backend
cd backend
go mod download
go build cmd/server/main.go

# 3. Build frontend
cd ../frontend
npm install && npm run build

# 4. Start the game!
cd ../backend
go run cmd/server/main.go
```

**Open http://localhost:8080 and start playing!**

### How to Play

1. **Join the Game** - Enter your name (up to 4 players)
2. **Move the Object** - Use arrow keys, WASD, or on-screen buttons
3. **Watch the Magic** - See how concurrent moves are handled
4. **Observe Conflicts** - Notice when your move conflicts with others

Open multiple browser tabs to simulate multiple players!

### What's Under the Hood

#### Backend (Go)
- **WebSocket Hub** - Manages real-time connections
- **Optimistic Concurrency** - Version-based conflict detection
- **Conflict Resolution** - First-wins strategy with client notifications

#### Frontend (React)
- **Live Updates** - Real-time game state synchronization  
- **Conflict Visualization** - User-friendly conflict notifications

### Testing It Out

```bash
# Run tests
cd backend
go test ./... -v

# Test the frontend
cd frontend
npm test
```

Want to see conflicts in action? Open multiple browser tabs and try moving the object at the same time!

### API Endpoints

- **Game Interface:** `http://localhost:8080`
- **Health Check:** `http://localhost:8080/health`  
- **WebSocket:** `ws://localhost:8080/ws`

### Why This Matters

This project demonstrates key concepts used in:
- **Database systems** (MVCC, transaction isolation)
- **Distributed systems** (conflict resolution, eventual consistency)  
- **Real-time applications** (WebSocket communication, state synchronization)
- **Modern web development** (Go backends, React frontends)

###  Project Structure

```
realtime-multiplayer-game/
├── backend/           # Go server with WebSocket support
│   ├── cmd/server/    # Main server entry point
│   ├── internal/      # Core game logic and concurrency control
│   └── pkg/           # Shared models and utilities
├── frontend/          # React application
│   └── src/           # Components, hooks, and utilities
└── docs/              # Additional documentation
```

### Contributing

Found a bug or have an idea? PRs welcome! Just make sure to:
1. Add tests for new features
2. Keep the existing style
3. Test with multiple concurrent players

### License

MIT License - feel free to use this for learning or building upon!

---
