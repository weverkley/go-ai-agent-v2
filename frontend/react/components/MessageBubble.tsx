import React from 'react';
import { ChatMessage } from '../types';
import ToolEvent from './ToolEvent';
import { Bot, User, Cpu } from 'lucide-react';

interface MessageBubbleProps {
  message: ChatMessage;
}

const MessageBubble: React.FC<MessageBubbleProps> = ({ message }) => {
  const isUser = message.role === 'user';
  const isSystem = message.role === 'system';

  if (isSystem) {
    return (
      <div className="flex justify-center my-4 opacity-70">
        <div className="flex items-center gap-2 text-xs text-gray-400 bg-gray-800/50 px-3 py-1 rounded-full border border-gray-700">
            <Cpu className="w-3 h-3" />
            {message.content}
        </div>
      </div>
    );
  }

  return (
    <div className={`flex w-full mb-6 ${isUser ? 'justify-end' : 'justify-start'}`}>
      <div className={`flex max-w-[85%] md:max-w-[75%] gap-3 ${isUser ? 'flex-row-reverse' : 'flex-row'}`}>
        
        {/* Avatar */}
        <div className={`shrink-0 w-8 h-8 rounded-lg flex items-center justify-center ${
          isUser ? 'bg-indigo-600' : 'bg-gray-700'
        }`}>
          {isUser ? <User className="w-5 h-5 text-white" /> : <Bot className="w-5 h-5 text-emerald-400" />}
        </div>

        {/* Content Box */}
        <div className={`flex flex-col min-w-0 ${isUser ? 'items-end' : 'items-start'}`}>
          
          <div className={`relative px-4 py-3 rounded-2xl shadow-md text-sm leading-relaxed overflow-hidden ${
            isUser 
              ? 'bg-indigo-600 text-white rounded-tr-none' 
              : 'bg-gray-800 text-gray-100 rounded-tl-none border border-gray-700'
          }`}>
             {/* Text Content */}
            <div className="whitespace-pre-wrap break-words font-sans">
              {message.content}
              {message.isStreaming && (
                <span className="inline-block w-2 h-4 ml-1 align-middle bg-emerald-400 animate-pulse"></span>
              )}
            </div>
          </div>

          {/* Tools Display (Only for AI) */}
          {!isUser && message.toolCalls && message.toolCalls.length > 0 && (
            <div className="w-full mt-2 space-y-1">
              {message.toolCalls.map((tool) => (
                <ToolEvent key={tool.id} tool={tool} />
              ))}
            </div>
          )}

          {/* Timestamp */}
          <span className="text-[10px] text-gray-500 mt-1 px-1">
            {new Date(message.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
          </span>

        </div>
      </div>
    </div>
  );
};

export default MessageBubble;