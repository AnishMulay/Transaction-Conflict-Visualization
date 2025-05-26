import React, { useEffect, useCallback } from 'react';
import './GameBoard.css';

const GameBoard = ({ gameState, onMove, playerName }) => {
  const { object, players, gridSize } = gameState;

  const handleKeyPress = useCallback((event) => {
    const keyMap = {
      'ArrowUp': 'up',
      'ArrowDown': 'down',
      'ArrowLeft': 'left',
      'ArrowRight': 'right',
      'w': 'up',
      's': 'down',
      'a': 'left',
      'd': 'right'
    };

    const direction = keyMap[event.key];
    if (direction) {
      event.preventDefault();
      onMove(direction);
    }
  }, [onMove]);

  useEffect(() => {
    document.addEventListener('keydown', handleKeyPress);
    return () => {
      document.removeEventListener('keydown', handleKeyPress);
    };
  }, [handleKeyPress]);

  const renderGrid = () => {
    const cells = [];
    
    for (let y = 0; y < gridSize.y; y++) {
      for (let x = 0; x < gridSize.x; x++) {
        const isObjectHere = object.position.x === x && object.position.y === y;
        
        cells.push(
          <div
            key={`${x}-${y}`}
            className={`grid-cell ${isObjectHere ? 'has-object' : ''}`}
            style={{
              gridColumn: x + 1,
              gridRow: y + 1,
            }}
          >
            {isObjectHere && (
              <div className="game-object">
                <div className="object-version">v{object.version}</div>
              </div>
            )}
          </div>
        );
      }
    }
    
    return cells;
  };

  return (
    <div className="game-board">
      <div className="game-info">
        <h3>Welcome, {playerName}!</h3>
        <p>Use arrow keys or WASD to move the shared object</p>
        <div className="object-info">
          <span>Object Position: ({object.position.x}, {object.position.y})</span>
          <span>Version: {object.version}</span>
          <span>Last Updated: {new Date(object.lastUpdated).toLocaleTimeString()}</span>
        </div>
      </div>
      
      <div 
        className="grid-container"
        style={{
          gridTemplateColumns: `repeat(${gridSize.x}, 1fr)`,
          gridTemplateRows: `repeat(${gridSize.y}, 1fr)`
        }}
        tabIndex={0}
      >
        {renderGrid()}
      </div>
      
      <div className="controls">
        <div className="control-row">
          <button onClick={() => onMove('up')}>↑</button>
        </div>
        <div className="control-row">
          <button onClick={() => onMove('left')}>←</button>
          <button onClick={() => onMove('down')}>↓</button>
          <button onClick={() => onMove('right')}>→</button>
        </div>
      </div>
    </div>
  );
};

export default GameBoard;