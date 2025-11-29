// WebSocket Protocol Types based on provided logs

export type EventType = 
  | 'streaming_started'
  | 'thinking'
  | 'tool_call_start'
  | 'tool_call_end'
  | 'chunk'
  | 'final_response'
  | 'error'
  | 'token_count';

export interface BaseMessage {
  type: EventType;
  sessionId: string;
}

export interface StreamingStartedMessage extends BaseMessage {
  type: 'streaming_started';
  payload: Record<string, never>; // Empty object
}

export interface ThinkingMessage extends BaseMessage {
  type: 'thinking';
  payload: Record<string, never>;
}

export interface ToolCallStartPayload {
  ToolCallID: string;
  ToolName: string;
  Args: Record<string, any>;
}

export interface ToolCallStartMessage extends BaseMessage {
  type: 'tool_call_start';
  payload: ToolCallStartPayload;
}

export interface ToolCallEndPayload {
  ToolCallID: string;
  ToolName: string;
  Result: string;
  Err: string | null;
}

export interface ToolCallEndMessage extends BaseMessage {
  type: 'tool_call_end';
  payload: ToolCallEndPayload;
}

export interface ChunkPayload {
  text: string;
}

export interface ChunkMessage extends BaseMessage {
  type: 'chunk';
  payload: ChunkPayload;
}

export interface FinalResponsePayload {
  Content: string;
}

export interface FinalResponseMessage extends BaseMessage {
  type: 'final_response';
  payload: FinalResponsePayload;
}

export interface ErrorPayload {
  message: string;
}

export interface ErrorMessage extends BaseMessage {
  type: 'error';
  payload: ErrorPayload;
}

export type WebSocketMessage = 
  | StreamingStartedMessage
  | ThinkingMessage
  | ToolCallStartMessage
  | ToolCallEndMessage
  | ChunkMessage
  | FinalResponseMessage
  | ErrorMessage;

// Internal UI State Types
export interface ChatMessage {
  id: string;
  role: 'user' | 'ai' | 'system';
  content: string;
  timestamp: number;
  isStreaming?: boolean;
  toolCalls?: ToolCallState[];
}

export interface ToolCallState {
  id: string;
  name: string;
  args: string; // stringified json
  result?: string;
  status: 'running' | 'completed' | 'failed';
}