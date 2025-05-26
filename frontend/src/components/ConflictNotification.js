import React, { useEffect } from 'react';
// import './ConflictNotification.css';

const ConflictNotification = ({ conflict, onClose }) => {
  useEffect(() => {
    const timer = setTimeout(() => {
      onClose();
    }, 5000); // Auto-close after 5 seconds

    return () => clearTimeout(timer);
  }, [onClose]);

  return (
    <div className="conflict-notification">
      <div className="conflict-header">
        <span className="conflict-title">Move Conflict Detected</span>
        <button className="close-button" onClick={onClose}>×</button>
      </div>
      <div className="conflict-message">
        {conflict.message}
      </div>
      <div className="conflict-time">
        {conflict.timestamp.toLocaleTimeString()}
      </div>
    </div>
  );
};

export default ConflictNotification;