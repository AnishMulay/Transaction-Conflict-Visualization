import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import GameBoard from '../GameBoard';

const mockGameState = {
  object: {
    id: 'test-object',
    position: { x: 5, y: 5 },
    version: 1,
    lastUpdated: new Date().toISOString()
  },
  players: {
    'player1': {
      id: 'player1',
      name: 'Test Player',
      color: '#FF0000',
      connected: true,
      lastSeen: new Date().toISOString()
    }
  },
  gridSize: { x: 10, y: 10 },
  version: 1,
  maxPlayers: 4
};

const mockOnMove = jest.fn();

describe('GameBoard', () => {
  beforeEach(() => {
    mockOnMove.mockClear();
  });

  test('renders game board with correct grid size', () => {
    render(
      <GameBoard 
        gameState={mockGameState} 
        onMove={mockOnMove}
        playerName="Test Player"
      />
    );
    
    expect(screen.getByText('Welcome, Test Player!')).toBeInTheDocument();
    expect(screen.getByText(/Object Position: \(5, 5\)/)).toBeInTheDocument();
  });

  test('handles keyboard input for movement', () => {
    render(
      <GameBoard 
        gameState={mockGameState} 
        onMove={mockOnMove}
        playerName="Test Player"
      />
    );
    
    const gridContainer = document.querySelector('.grid-container');
    
    // Simulate arrow key presses
    fireEvent.keyDown(gridContainer, { key: 'ArrowUp' });
    expect(mockOnMove).toHaveBeenCalledWith('up');
    
    fireEvent.keyDown(gridContainer, { key: 'ArrowDown' });
    expect(mockOnMove).toHaveBeenCalledWith('down');
    
    fireEvent.keyDown(gridContainer, { key: 'ArrowLeft' });
    expect(mockOnMove).toHaveBeenCalledWith('left');
    
    fireEvent.keyDown(gridContainer, { key: 'ArrowRight' });
    expect(mockOnMove).toHaveBeenCalledWith('right');
  });

  test('handles WASD keys for movement', () => {
    render(
      <GameBoard 
        gameState={mockGameState} 
        onMove={mockOnMove}
        playerName="Test Player"
      />
    );
    
    const gridContainer = document.querySelector('.grid-container');
    
    fireEvent.keyDown(gridContainer, { key: 'w' });
    expect(mockOnMove).toHaveBeenCalledWith('up');
    
    fireEvent.keyDown(gridContainer, { key: 's' });
    expect(mockOnMove).toHaveBeenCalledWith('down');
    
    fireEvent.keyDown(gridContainer, { key: 'a' });
    expect(mockOnMove).toHaveBeenCalledWith('left');
    
    fireEvent.keyDown(gridContainer, { key: 'd' });
    expect(mockOnMove).toHaveBeenCalledWith('right');
  });

  test('handles button clicks for movement', () => {
    render(
      <GameBoard 
        gameState={mockGameState} 
        onMove={mockOnMove}
        playerName="Test Player"
      />
    );
    
    fireEvent.click(screen.getByText('↑'));
    expect(mockOnMove).toHaveBeenCalledWith('up');
    
    fireEvent.click(screen.getByText('↓'));
    expect(mockOnMove).toHaveBeenCalledWith('down');
    
    fireEvent.click(screen.getByText('←'));
    expect(mockOnMove).toHaveBeenCalledWith('left');
    
    fireEvent.click(screen.getByText('→'));
    expect(mockOnMove).toHaveBeenCalledWith('right');
  });

  test('displays object at correct position', () => {
    render(
      <GameBoard 
        gameState={mockGameState} 
        onMove={mockOnMove}
        playerName="Test Player"
      />
    );
    
    const objectElement = document.querySelector('.game-object');
    expect(objectElement).toBeInTheDocument();
    expect(screen.getByText('v1')).toBeInTheDocument(); // Version display
  });
});