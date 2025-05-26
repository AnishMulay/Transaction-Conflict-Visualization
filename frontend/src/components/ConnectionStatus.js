import React from 'react';
// import './ConnectionStatus.css';

const ConnectionStatus = ({ isConnected, error }) => {
  return (
    <div className={`connection-status ${isConnected ? 'connected' : 'disconnected'}`}>
      <div className="status-indicator" />
      <span className="status-text">
        {isConnected ? 'Connected' : 'Disconnected'}
      </span>
      {error && (
        <span className="error-text">{error}</span>
      )}
    </div>
  );
};

export default ConnectionStatus;