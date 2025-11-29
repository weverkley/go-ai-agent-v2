import React, { useEffect, useState, useRef } from 'react';
import { wsService } from './services/socket';
import { ChatMessage, ToolCallState, WebSocketMessage } from './types';
import MessageBubble from './components/MessageBubble';
import { Send, Zap, Wifi, WifiOff, Settings } from 'lucide-react';

const App: React.FC = () => {
  const [isConnected, setIsConnected] = useState(false);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [input, setInput] = useState('');
  const [isThinking, setIsThinking] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to bottom
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages, isThinking]);

  // WebSocket Connection & Event Handling
  useEffect(() => {
    wsService.connect();

    const cleanupConnection = wsService.addConnectionHandler((status) => {
      setIsConnected(status);
    });

    const cleanupMessages = wsService.addMessageHandler((msg: WebSocketMessage) => {
      handleServerMessage(msg);
    });

    return () => {
      cleanupConnection();
      cleanupMessages();
      wsService.disconnect();
    };
  }, []);

  const handleServerMessage = (msg: WebSocketMessage) => {
    switch (msg.type) {
      case 'streaming_started':
        setIsThinking(true);
        // Create a placeholder AI message if one doesn't exist at the tail
        // or just rely on chunk arrival to create it.
        break;

      case 'thinking':
        setIsThinking(true);
        break;

      case 'chunk':
        setIsThinking(false); // We are generating text, no longer just "thinking"
        setMessages((prev) => {
          const newMessages = [...prev];
          const lastMsg = newMessages[newMessages.length - 1];

          // If last message is AI and is streaming, append
          if (lastMsg && lastMsg.role === 'ai' && lastMsg.isStreaming) {
            return [
              ...newMessages.slice(0, -1),
              { ...lastMsg, content: lastMsg.content + msg.payload.text }
            ];
          } 
          // Otherwise create new AI message
          else {
            return [
              ...newMessages,
              {
                id: crypto.randomUUID(),
                role: 'ai',
                content: msg.payload.text,
                timestamp: Date.now(),
                isStreaming: true,
                toolCalls: []
              }
            ];
          }
        });
        break;

      case 'tool_call_start':
        setIsThinking(true); // Technically executing tool is a form of thinking
        setMessages((prev) => {
          const newMessages = [...prev];
          let lastMsg = newMessages[newMessages.length - 1];

          // Ensure we have an active AI message to attach the tool call to
          if (!lastMsg || lastMsg.role !== 'ai' || !lastMsg.isStreaming) {
             const newMsg: ChatMessage = {
                id: crypto.randomUUID(),
                role: 'ai',
                content: '',
                timestamp: Date.now(),
                isStreaming: true,
                toolCalls: []
             };
             newMessages.push(newMsg);
             lastMsg = newMsg;
          }

          const toolCall: ToolCallState = {
            id: msg.payload.ToolCallID,
            name: msg.payload.ToolName,
            args: JSON.stringify(msg.payload.Args, null, 2),
            status: 'running'
          };

          const updatedMsg = {
            ...lastMsg,
            toolCalls: [...(lastMsg.toolCalls || []), toolCall]
          };

          return [...newMessages.slice(0, -1), updatedMsg];
        });
        break;

      case 'tool_call_end':
        setMessages((prev) => {
          const newMessages = [...prev];
          const lastMsg = newMessages[newMessages.length - 1];

          if (lastMsg && lastMsg.role === 'ai') {
            const updatedToolCalls = (lastMsg.toolCalls || []).map(t => 
              t.id === msg.payload.ToolCallID 
                ? { ...t, status: 'completed' as const, result: msg.payload.Result || msg.payload.Err || '' }
                : t
            );
            return [
              ...newMessages.slice(0, -1),
              { ...lastMsg, toolCalls: updatedToolCalls }
            ];
          }
          return prev;
        });
        break;

      case 'final_response':
        setIsThinking(false);
        setMessages((prev) => {
          const newMessages = [...prev];
          const lastMsg = newMessages[newMessages.length - 1];
          
          if (lastMsg && lastMsg.role === 'ai') {
            // Replace full content with final response to ensure integrity
            // or just unset isStreaming.
            // Based on logs, final_response contains full text usually, or just the end. 
            // The guide says "The final complete response".
            return [
              ...newMessages.slice(0, -1),
              { 
                ...lastMsg, 
                content: msg.payload.Content, 
                isStreaming: false 
              }
            ];
          }
          return prev;
        });
        break;

      case 'error':
        setIsThinking(false);
        setMessages((prev) => [
          ...prev,
          {
            id: crypto.randomUUID(),
            role: 'system',
            content: `Error: ${msg.payload.message}`,
            timestamp: Date.now()
          }
        ]);
        break;
    }
  };

  const handleSendMessage = (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim() || !isConnected) return;

    // Add user message immediately
    const userMsg: ChatMessage = {
      id: crypto.randomUUID(),
      role: 'user',
      content: input,
      timestamp: Date.now()
    };
    setMessages(prev => [...prev, userMsg]);
    
    // Send to WS
    wsService.send(input);
    setInput('');
  };

  return (
    <div className="flex flex-col h-screen bg-gray-950 text-gray-100 font-sans">
      
      {/* Header */}
      <header className="h-16 border-b border-gray-800 bg-gray-900/50 backdrop-blur flex items-center justify-between px-6 sticky top-0 z-10">
        <div className="flex items-center gap-3">
          <div className="bg-indigo-600 p-2 rounded-lg">
            <Zap className="w-5 h-5 text-white fill-current" />
          </div>
          <div>
            <h1 className="font-bold text-lg tracking-tight">GO Ai Agent V2</h1>
            <p className="text-xs text-gray-500">Real-time WebSocket Interface</p>
          </div>
        </div>
        
        <div className="flex items-center gap-4">
          <div className={`flex items-center gap-2 px-3 py-1.5 rounded-full text-xs font-medium border ${
            isConnected 
              ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' 
              : 'bg-red-500/10 text-red-400 border-red-500/20'
          }`}>
            {isConnected ? <Wifi className="w-3.5 h-3.5" /> : <WifiOff className="w-3.5 h-3.5" />}
            {isConnected ? 'Connected' : 'Disconnected'}
          </div>
          <button className="text-gray-400 hover:text-white transition-colors">
            <Settings className="w-5 h-5" />
          </button>
        </div>
      </header>

      {/* Chat Area */}
      <main className="flex-1 overflow-y-auto p-4 md:p-6 scroll-smooth">
        <div className="max-w-4xl mx-auto flex flex-col justify-end min-h-full">
          
          {messages.length === 0 && (
            <div className="flex-1 flex flex-col items-center justify-center text-gray-600 space-y-4 opacity-50">
              <Zap className="w-16 h-16 opacity-20" />
              <p>Ready to connect to Agent V2...</p>
            </div>
          )}

          {messages.map((msg) => (
            <MessageBubble key={msg.id} message={msg} />
          ))}

          {isThinking && (
             <div className="flex justify-start mb-6">
               <div className="bg-gray-800/50 border border-gray-700/50 rounded-full px-4 py-2 flex items-center gap-2 text-xs text-gray-400">
                  <span className="relative flex h-2 w-2">
                    <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-indigo-400 opacity-75"></span>
                    <span className="relative inline-flex rounded-full h-2 w-2 bg-indigo-500"></span>
                  </span>
                  Agent is thinking...
               </div>
             </div>
          )}
          
          <div ref={messagesEndRef} />
        </div>
      </main>

      {/* Input Area */}
      <footer className="p-4 bg-gray-900 border-t border-gray-800">
        <div className="max-w-4xl mx-auto">
          <form onSubmit={handleSendMessage} className="relative flex items-center gap-2">
            <input
              type="text"
              value={input}
              onChange={(e) => setInput(e.target.value)}
              placeholder={isConnected ? "Message the agent..." : "Connecting..."}
              disabled={!isConnected}
              className="w-full bg-gray-950 text-gray-100 border border-gray-700 rounded-xl py-3.5 pl-4 pr-12 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-all placeholder:text-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
            />
            <button
              type="submit"
              disabled={!input.trim() || !isConnected}
              className="absolute right-2 p-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-500 disabled:bg-gray-700 disabled:text-gray-500 transition-all shadow-lg shadow-indigo-500/20"
            >
              <Send className="w-4 h-4" />
            </button>
          </form>
          <div className="text-center mt-2">
             <p className="text-[10px] text-gray-600">
               Agent can execute tools and generate real-time responses.
             </p>
          </div>
        </div>
      </footer>
    </div>
  );
};

export default App;