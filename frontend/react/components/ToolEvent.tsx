import React, { useState } from 'react';
import { ToolCallState } from '../types';
import { Terminal, CheckCircle2, XCircle, ChevronDown, ChevronRight, Activity } from 'lucide-react';

interface ToolEventProps {
  tool: ToolCallState;
}

const ToolEvent: React.FC<ToolEventProps> = ({ tool }) => {
  const [isExpanded, setIsExpanded] = useState(false);

  const getStatusIcon = () => {
    switch (tool.status) {
      case 'running':
        return <Activity className="w-4 h-4 text-amber-400 animate-pulse" />;
      case 'completed':
        return <CheckCircle2 className="w-4 h-4 text-emerald-400" />;
      case 'failed':
        return <XCircle className="w-4 h-4 text-red-400" />;
    }
  };

  const getBorderColor = () => {
    switch (tool.status) {
      case 'running': return 'border-amber-500/30 bg-amber-500/5';
      case 'completed': return 'border-emerald-500/30 bg-emerald-500/5';
      case 'failed': return 'border-red-500/30 bg-red-500/5';
    }
  };

  return (
    <div className={`my-2 rounded-lg border ${getBorderColor()} overflow-hidden transition-all duration-200`}>
      <button 
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full flex items-center gap-3 px-3 py-2 text-xs font-mono text-left hover:bg-white/5 transition-colors"
      >
        <span className="shrink-0">{getStatusIcon()}</span>
        <span className="font-semibold text-gray-300">TOOL:</span>
        <span className="text-blue-300">{tool.name}</span>
        <span className="flex-1"></span>
        {isExpanded ? <ChevronDown className="w-3 h-3 text-gray-500" /> : <ChevronRight className="w-3 h-3 text-gray-500" />}
      </button>

      {isExpanded && (
        <div className="border-t border-gray-700/50 p-3 bg-black/20 space-y-2">
          <div>
            <div className="text-[10px] uppercase tracking-wider text-gray-500 mb-1">Arguments</div>
            <pre className="text-xs text-gray-300 whitespace-pre-wrap font-mono bg-black/30 p-2 rounded border border-gray-800">
              {tool.args}
            </pre>
          </div>
          
          {tool.result && (
            <div>
              <div className="text-[10px] uppercase tracking-wider text-gray-500 mb-1">Result</div>
              <pre className="text-xs text-emerald-300/80 whitespace-pre-wrap font-mono bg-black/30 p-2 rounded border border-gray-800 max-h-40 overflow-y-auto">
                {tool.result}
              </pre>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default ToolEvent;