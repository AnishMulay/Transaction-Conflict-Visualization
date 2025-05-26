import React, { useState } from 'react';
import './JoinForm.css';

const JoinForm = ({ onJoin }) => {
  const [playerName, setPlayerName] = useState('');

  const handleSubmit = (e) => {
    e.preventDefault();
    if (playerName.trim()) {
      onJoin(playerName.trim());
    }
  };

  return (
    <div className="join-form">
      <h2>Join the Game</h2>
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="playerName">Your Name:</label>
          <input
            type="text"
            id="playerName"
            value={playerName}
            onChange={(e) => setPlayerName(e.target.value)}
            placeholder="Enter your name"
            maxLength={20}
            required
          />
        </div>
        <button type="submit" disabled={!playerName.trim()}>
          Join Game
        </button>
      </form>
    </div>
  );
};

export default JoinForm;