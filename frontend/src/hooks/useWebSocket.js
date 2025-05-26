import { useState, useEffect, useRef, useCallback } from 'react';

const useWebSocket = (url) => {
  const [isConnected, setIsConnected] = useState(false);
  const [connectionError, setConnectionError] = useState(null);
  const [lastMessage, setLastMessage] = useState(null);
  const [conflictStats, setConflictStats] = useState({
    conflicts: 0,
    total: 0,
    successRate: 100
  });
  
  const ws = useRef(null);
  const reconnectTimeoutRef = useRef(null);
  const reconnectAttempts = useRef(0);
  const maxReconnectAttempts = 10;
  const reconnectInterval = 3000;

  const connect = useCallback(() => {
    try {
      ws.current = new WebSocket(url);
      
      ws.current.onopen = () => {
        console.log('WebSocket connected');
        setIsConnected(true);
        setConnectionError(null);
        reconnectAttempts.current = 0;
      };
      
      ws.current.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data);
          setLastMessage(message);
          
          // Update conflict statistics
          if (message.type === 'conflict') {
            setConflictStats(prev => ({
              conflicts: prev.conflicts + 1,
              total: prev.total + 1,
              successRate: Math.round(((prev.total - prev.conflicts) / (prev.total + 1)) * 100)
            }));
          } else if (message.type === 'gameState') {
            setConflictStats(prev => ({
              ...prev,
              total: prev.total + 1,
              successRate: Math.round(((prev.total - prev.conflicts) / prev.total) * 100)
            }));
          }
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error);
        }
      };
      
      ws.current.onclose = (event) => {
        console.log('WebSocket disconnected:', event.code, event.reason);
        setIsConnected(false);
        
        if (!event.wasClean && reconnectAttempts.current < maxReconnectAttempts) {
          reconnectAttempts.current++;
          setConnectionError(`Connection lost. Reconnecting... (${reconnectAttempts.current}/${maxReconnectAttempts})`);
          
          reconnectTimeoutRef.current = setTimeout(() => {
            connect();
          }, reconnectInterval);
        } else if (reconnectAttempts.current >= maxReconnectAttempts) {
          setConnectionError('Failed to reconnect after multiple attempts');
        }
      };
      
      ws.current.onerror = (error) => {
        console.error('WebSocket error:', error);
        setConnectionError('Connection error occurred');
      };
      
    } catch (error) {
      console.error('Failed to create WebSocket connection:', error);
      setConnectionError('Failed to connect to server');
    }
  }, [url]);

  const sendMessage = useCallback((message) => {
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      try {
        ws.current.send(JSON.stringify(message));
      } catch (error) {
        console.error('Failed to send message:', error);
      }
    } else {
      console.warn('WebSocket is not connected');
    }
  }, []);

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
    }
    
    if (ws.current) {
      ws.current.close(1000, 'User disconnect');
    }
  }, []);

  useEffect(() => {
    connect();
    
    return () => {
      disconnect();
    };
  }, [connect, disconnect]);

  return {
    isConnected,
    connectionError,
    sendMessage,
    lastMessage,
    conflictStats,
    disconnect
  };
};

export default useWebSocket;