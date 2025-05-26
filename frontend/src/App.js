import React, { useState, useEffect, useCallback } from 'react';
import './App.css';
import GameBoard from './components/GameBoard';
import PlayerList from './components/PlayerList';
import ConnectionStatus from './components/ConnectionStatus';
import JoinForm from './components/JoinForm';
import ConflictNotification from './components/ConflictNotification';
import useWebSocket from './hooks/useWebSocket';

function App() {
  const [gameState, setGameState] = useState(null);
  const [playerName, setPlayerName] = useState('');
  const [isJoined, setIsJoined] = useState(false);
  const [conflicts, setConflicts] = useState([]);
  
  const {
    isConnected,
    connectionError,
    sendMessage,
    lastMessage,
    conflictStats
  } = useWebSocket('ws://localhost:8080/ws');

  // Handle incoming WebSocket messages
  useEffect(() => {
    if (!lastMessage) return;

    switch (lastMessage.type) {
      case 'gameState':
        setGameState(lastMessage.data);
        break;
      
      case 'error':
        console.error('Game error:', lastMessage.data);
        if (lastMessage.data.code === 'GAME_FULL') {
          alert('Game is full! Please try again later.');
        }
        break;
      
      case 'conflict':
        console.warn('Move conflict:', lastMessage.data);
        setConflicts(prev => [...prev, {
          id: Date.now(),
          message: lastMessage.data.message,
          timestamp: new Date(lastMessage.timestamp)
        }]);
        break;
      
      default:
        console.log('Unknown message type:', lastMessage.type);
    }
  }, [lastMessage]);

  const handleJoinGame = useCallback((name) => {
    if (!isConnected) {
      alert('Not connected to server');
      return;
    }

    setPlayerName(name);
    sendMessage({
      type: 'join',
      data: { playerName: name },
      timestamp: new Date().toISOString()
    });
    setIsJoined(true);
  }, [isConnected, sendMessage]);

  const handleMove = useCallback((direction) => {
    if (!isJoined || !gameState) return;

    const requestId = `${Date.now()}-${Math.random()}`;
    sendMessage({
      type: 'move',
      data: {
        direction,
        objectVersion: gameState.object.version,
        requestId
      },
      timestamp: new Date().toISOString()
    });
  }, [isJoined, gameState, sendMessage]);

  const removeConflict = useCallback((conflictId) => {
    setConflicts(prev => prev.filter(c => c.id !== conflictId));
  }, []);

  if (!isJoined) {
    return (
      <div className="app">
        <header className="app-header">
          <h1>Real-Time Multiplayer Game</h1>
          <p>Demonstrating Optimistic Concurrency Control</p>
        </header>
        
        <ConnectionStatus 
          isConnected={isConnected} 
          error={connectionError} 
        />
        
        {isConnected && (
          <JoinForm onJoin={handleJoinGame} />
        )}
      </div>
    );
  }

  return (
    <div className="app">
      <header className="app-header">
        <h1>Real-Time Multiplayer Game</h1>
        <ConnectionStatus 
          isConnected={isConnected} 
          error={connectionError} 
        />
        {conflictStats && (
          <div className="conflict-stats">
            <span>Conflicts: {conflictStats.conflicts}</span>
            <span>Success Rate: {conflictStats.successRate}%</span>
          </div>
        )}
      </header>

      <main className="game-container">
        {gameState && (
          <>
            <div className="game-area">
              <GameBoard 
                gameState={gameState}
                onMove={handleMove}
                playerName={playerName}
              />
            </div>
            
            <div className="sidebar">
              <PlayerList players={gameState.players} />
            </div>
          </>
        )}
      </main>

      {conflicts.map(conflict => (
        <ConflictNotification
          key={conflict.id}
          conflict={conflict}
          onClose={() => removeConflict(conflict.id)}
        />
      ))}
    </div>
  );
}

export default App;