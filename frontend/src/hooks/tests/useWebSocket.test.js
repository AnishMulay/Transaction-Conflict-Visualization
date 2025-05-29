import { renderHook, act } from '@testing-library/react';
import useWebSocket from '../useWebSocket';

// Mock WebSocket
class MockWebSocket {
  constructor(url) {
    this.url = url;
    this.readyState = WebSocket.CONNECTING;
    setTimeout(() => {
      this.readyState = WebSocket.OPEN;
      if (this.onopen) this.onopen();
    }, 100);
  }

  send(data) {
    if (this.onmessage) {
      setTimeout(() => {
        this.onmessage({
          data: JSON.stringify({
            type: 'echo',
            data: JSON.parse(data),
            timestamp: new Date().toISOString()
          })
        });
      }, 50);
    }
  }

  close() {
    this.readyState = WebSocket.CLOSED;
    if (this.onclose) this.onclose({ wasClean: true });
  }
}

global.WebSocket = MockWebSocket;

describe('useWebSocket', () => {
  test('establishes connection', async () => {
    const { result } = renderHook(() => useWebSocket('ws://localhost:8080/ws'));
    
    expect(result.current.isConnected).toBe(false);
    
    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 150));
    });
    
    expect(result.current.isConnected).toBe(true);
  });

  test('sends and receives messages', async () => {
    const { result } = renderHook(() => useWebSocket('ws://localhost:8080/ws'));
    
    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 150));
    });
    
    const testMessage = { type: 'test', data: 'hello' };
    
    act(() => {
      result.current.sendMessage(testMessage);
    });
    
    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 100));
    });
    
    expect(result.current.lastMessage).toBeDefined();
    expect(result.current.lastMessage.type).toBe('echo');
  });

  test('tracks conflict statistics', async () => {
    const { result } = renderHook(() => useWebSocket('ws://localhost:8080/ws'));
    
    expect(result.current.conflictStats).toEqual({
      conflicts: 0,
      total: 0,
      successRate: 100
    });
  });
});