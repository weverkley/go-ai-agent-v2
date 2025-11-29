import { WebSocketMessage } from '../types';

type MessageHandler = (message: WebSocketMessage) => void;
type ConnectionHandler = (isConnected: boolean) => void;

class WebSocketService {
  private socket: WebSocket | null = null;
  private url: string = 'ws://localhost:8080/ws'; // Default, configurable
  private messageHandlers: Set<MessageHandler> = new Set();
  private connectionHandlers: Set<ConnectionHandler> = new Set();
  private reconnectTimeout: ReturnType<typeof setTimeout> | null = null;
  private isExplicitlyDisconnected: boolean = false;

  constructor() {
    // Bind methods
    this.connect = this.connect.bind(this);
    this.disconnect = this.disconnect.bind(this);
    this.send = this.send.bind(this);
  }

  public setUrl(url: string) {
    this.url = url;
  }

  public connect() {
    if (this.socket?.readyState === WebSocket.OPEN || this.socket?.readyState === WebSocket.CONNECTING) {
      return;
    }

    this.isExplicitlyDisconnected = false;

    try {
      this.socket = new WebSocket(this.url);

      this.socket.onopen = () => {
        console.log('WS Connected');
        this.notifyConnection(true);
      };

      this.socket.onclose = () => {
        console.log('WS Closed');
        this.notifyConnection(false);
        this.socket = null;
        if (!this.isExplicitlyDisconnected) {
          this.scheduleReconnect();
        }
      };

      this.socket.onerror = (error) => {
        console.error('WS Error', error);
        // Error will trigger close, which handles reconnect
      };

      this.socket.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data) as WebSocketMessage;
          this.notifyMessage(data);
        } catch (e) {
          console.error('Failed to parse WS message', e);
        }
      };
    } catch (e) {
      console.error('WS Connection failed', e);
      this.scheduleReconnect();
    }
  }

  public disconnect() {
    this.isExplicitlyDisconnected = true;
    if (this.socket) {
      this.socket.close();
    }
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
    }
  }

  public send(prompt: string) {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify({ prompt }));
    } else {
      console.warn('Cannot send message, socket not open');
    }
  }

  public addMessageHandler(handler: MessageHandler) {
    this.messageHandlers.add(handler);
    return () => this.messageHandlers.delete(handler);
  }

  public addConnectionHandler(handler: ConnectionHandler) {
    this.connectionHandlers.add(handler);
    // Immediately notify current state
    handler(this.socket?.readyState === WebSocket.OPEN);
    return () => this.connectionHandlers.delete(handler);
  }

  private notifyMessage(message: WebSocketMessage) {
    this.messageHandlers.forEach(h => h(message));
  }

  private notifyConnection(isConnected: boolean) {
    this.connectionHandlers.forEach(h => h(isConnected));
  }

  private scheduleReconnect() {
    if (this.reconnectTimeout) return;
    
    console.log('Scheduling reconnect in 3s...');
    this.reconnectTimeout = setTimeout(() => {
      this.reconnectTimeout = null;
      this.connect();
    }, 3000);
  }
}

export const wsService = new WebSocketService();