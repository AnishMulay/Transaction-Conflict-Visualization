import React from 'react';
// import './PlayerList.css';

const PlayerList = ({ players }) => {
  const playerArray = Object.values(players || {});

  return (
    <div className="player-list">
      <h3>Players ({playerArray.length}/4)</h3>
      <div className="players">
        {playerArray.map(player => (
          <div key={player.id} className="player-item">
            <div 
              className="player-color" 
              style={{ backgroundColor: player.color }}
            />
            <div className="player-info">
              <div className="player-name">{player.name}</div>
              <div className="player-status">
                {player.connected ? (
                  <span className="status-online">Online</span>
                ) : (
                  <span className="status-offline">Offline</span>
                )}
              </div>
              <div className="last-seen">
                Last seen: {new Date(player.lastSeen).toLocaleTimeString()}
              </div>
            </div>
          </div>
        ))}
      </div>
      
      {playerArray.length === 0 && (
        <div className="no-players">
          No players connected
        </div>
      )}
    </div>
  );
};

export default PlayerList;