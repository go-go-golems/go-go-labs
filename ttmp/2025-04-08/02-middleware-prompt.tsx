import React, { useState, useCallback, useEffect } from 'react';

// Define our types
interface Context {
  [key: string]: any;
}

interface PromptFragmentMetadata {
  id?: string;
  type?: string;
  position?: 'start' | 'middle' | 'end';
  priority?: number;
  tags?: string[];
}

interface PromptFragment {
  content: string;
  metadata: PromptFragmentMetadata;
}

interface Middleware {
  id: string;
  name: string;
  description: string;
  prompt: (context: Context, fragments: PromptFragment[]) => [Context, PromptFragment[]];
  parse: (context: Context, response: string) => [Context, string];
  enabled: boolean;
}

// Context viewer component for structured display
const ContextViewer = ({ context }) => {
  if (!context || Object.keys(context).length === 0) {
    return <div className="text-gray-500 italic">No context data available</div>;
  }
  
  return (
    <div className="bg-gray-100 p-3 rounded overflow-auto max-h-64">
      {Object.entries(context).map(([key, value]) => (
        <div key={key} className="mb-2">
          <span className="font-medium text-blue-600">{key}: </span>
          <span className="font-mono">
            {typeof value === 'object' 
              ? JSON.stringify(value, null, 2) 
              : String(value)}
          </span>
        </div>
      ))}
    </div>
  );
};

  // Sample middlewares
const sampleMiddlewares: Middleware[] = [
  {
    id: 'system-instruction',
    name: 'System Instruction',
    description: 'Adds system instructions at the beginning of the prompt',
    enabled: true,
    prompt: (context, fragments) => {
      const systemInstruction = {
        content: 'You are a helpful AI assistant. Answer clearly and concisely.',
        metadata: {
          id: 'system-instruction',
          type: 'system',
          position: 'start',
          priority: 100
        }
      };
      
      // This middleware always replaces the fragments array with a new array
      // where the system instruction is first
      return [context, [systemInstruction, ...fragments]];
    },
    parse: (context, response) => {
      // This middleware doesn't modify the response
      return [context, response];
    }
  },
  {
    id: 'thinking-mode',
    name: 'Thinking Mode',
    description: 'Adds thinking mode instructions and extracts thinking from response',
    enabled: true,
    prompt: (context, fragments) => {
      const thinkingMode = context.thinkingMode || false;
      
      const thinkingInstruction = {
        content: thinkingMode 
          ? 'Please think step by step and show your reasoning in <thinking>...</thinking> tags.'
          : '',
        metadata: {
          id: 'thinking-instruction',
          type: 'instruction',
          position: 'middle',
          priority: 50
        }
      };
      
      // Only add if thinking mode is enabled
      const newFragments = thinkingMode 
        ? [...fragments, thinkingInstruction]
        : fragments;
      
      return [context, newFragments];
    },
    parse: (context, response) => {
      // Check if there's thinking tags in the response
      const thinkingRegex = /<thinking>([\s\S]*?)<\/thinking>/;
      const match = response.match(thinkingRegex);
      
      if (match) {
        const thinking = match[1];
        const newContext = { ...context, extractedThinking: thinking };
        const cleanedResponse = response.replace(thinkingRegex, '').trim();
        return [newContext, cleanedResponse];
      }
      
      return [context, response];
    }
  },
  {
    id: 'json-response',
    name: 'JSON Response',
    description: 'Formats prompt to request JSON response and validates response format',
    enabled: false,
    prompt: (context, fragments) => {
      const jsonInstruction = {
        content: 'Return your response in valid JSON format.',
        metadata: {
          id: 'json-format',
          type: 'format',
          position: 'end',
          priority: 80
        }
      };
      
      return [context, [...fragments, jsonInstruction]];
    },
    parse: (context, response) => {
      try {
        // Try to parse as JSON
        JSON.parse(response);
        return [{ ...context, isValidJson: true }, response];
      } catch (e) {
        // Not valid JSON
        return [{ ...context, isValidJson: false }, response];
      }
    }
  },
  {
    id: 'token-counter',
    name: 'Token Counter',
    description: 'Estimates token usage in prompt and response',
    enabled: true,
    prompt: (context, fragments) => {
      // Simple token estimation (in real app would use proper tokenizer)
      const combinedText = fragments.map(f => f.content).join(' ');
      const estimatedTokens = Math.ceil(combinedText.length / 4);
      
      return [{ ...context, promptTokens: estimatedTokens }, fragments];
    },
    parse: (context, response) => {
      // Simple token estimation (in real app would use proper tokenizer)
      const estimatedTokens = Math.ceil(response.length / 4);
      
      return [{ ...context, responseTokens: estimatedTokens }, response];
    }
  },
  {
    id: 'output-format',
    name: 'Output Format',
    description: 'Adds instructions for specific output format (markdown)',
    enabled: false,
    prompt: (context, fragments) => {
      const formatInstruction = {
        content: 'Format your response using Markdown.',
        metadata: {
          id: 'markdown-format',
          type: 'format',
          position: 'end',
          priority: 70
        }
      };
      
      return [context, [...fragments, formatInstruction]];
    },
    parse: (context, response) => {
      // This middleware doesn't modify the response in the demo
      return [{ ...context, formattingApplied: true }, response];
    }
  }
];

// Generate mock LLM response based on enabled middlewares
const generateMockResponse = (enabledMiddlewares, query) => {
  // Base response
  let response = `I'll help answer your question about "${query}".`;
  
  // Check if thinking mode is enabled
  if (enabledMiddlewares.some(m => m.id === 'thinking-mode')) {
    response += `\n\n<thinking>
Let me analyze this step by step:
1. The user is asking about ${query}
2. This topic requires some background explanation
3. I should provide a concise answer with key points
4. I'll make sure to address the core concepts
</thinking>\n\n`;
  }
  
  // Main content of the response
  if (enabledMiddlewares.some(m => m.id === 'json-response')) {
    // JSON format
    response += `{\n  "answer": "Quantum computing uses qubits that can exist in multiple states simultaneously due to quantum superposition, enabling faster processing of certain problems compared to classical computers.",\n  "keyPoints": ["Superposition", "Qubits", "Entanglement"],\n  "references": ["Nielsen & Chuang, 2010"]\n}`;
  } else if (enabledMiddlewares.some(m => m.id === 'output-format')) {
    // Markdown format
    response += `# Quantum Computing Basics

Quantum computing uses **qubits** that can exist in multiple states simultaneously due to quantum superposition.

## Key concepts:
* Superposition
* Entanglement
* Quantum gates

This allows quantum computers to solve *certain* problems much faster than classical computers.`;
  } else {
    // Plain text
    response += `Quantum computing uses qubits that can exist in multiple states simultaneously due to quantum superposition, enabling faster processing of certain problems compared to classical computers. Key concepts include superposition, entanglement, and quantum gates.`;
  }
  
  return response;
};

// Middleware item component with manual reordering
const MiddlewareItem = ({ middleware, index, moveUp, moveDown, isFirst, isLast, toggleMiddleware }) => {
  return (
    <div className={`p-4 mb-2 border rounded shadow-sm ${middleware.enabled ? 'bg-white' : 'bg-gray-100'}`}>
      <div className="flex items-center justify-between">
        <h3 className="font-semibold">{middleware.name}</h3>
        <div className="flex items-center">
          <div className="flex mr-4">
            <button 
              onClick={() => moveUp(index)} 
              disabled={isFirst}
              className={`px-2 py-1 mr-1 rounded ${isFirst ? 'text-gray-400 cursor-not-allowed' : 'text-blue-500 hover:bg-blue-100'}`}
            >
              ↑
            </button>
            <button 
              onClick={() => moveDown(index)} 
              disabled={isLast}
              className={`px-2 py-1 rounded ${isLast ? 'text-gray-400 cursor-not-allowed' : 'text-blue-500 hover:bg-blue-100'}`}
            >
              ↓
            </button>
          </div>
          <span className="text-sm text-gray-500 mr-2">
            {middleware.enabled ? 'Enabled' : 'Disabled'}
          </span>
          <label className="switch relative inline-block w-10 h-5">
            <input
              type="checkbox"
              checked={middleware.enabled}
              onChange={() => toggleMiddleware(middleware.id)}
              className="opacity-0 w-0 h-0"
            />
            <span className={`slider absolute cursor-pointer inset-0 rounded-full ${
              middleware.enabled ? 'bg-blue-500' : 'bg-gray-300'
            } transition-colors duration-200`}></span>
          </label>
        </div>
      </div>
      <p className="text-sm text-gray-600 mt-1">{middleware.description}</p>
    </div>
  );
};

// Fragment visual component
const FragmentVisualizer = ({ fragmentStages }) => {
  if (!fragmentStages || fragmentStages.length === 0) {
    return null;
  }
  
  return (
    <div className="mb-4">
      <h3 className="font-medium mb-2">Prompt Fragment Pipeline</h3>
      <div className="overflow-x-auto">
        {fragmentStages.map((stage, stageIndex) => (
          <div key={stageIndex} className="mb-6">
            <div className="flex items-center mb-2">
              <div className="w-8 h-8 rounded-full bg-blue-500 flex items-center justify-center text-white font-bold mr-2">
                {stageIndex + 1}
              </div>
              <h4 className="font-medium">{stage.stage}</h4>
            </div>
            
            <div className="border rounded overflow-hidden">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="p-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">ID/Type</th>
                    <th className="p-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Position</th>
                    <th className="p-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Priority</th>
                    <th className="p-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Content</th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {stage.fragments.length > 0 ? (
                    stage.fragments.map((fragment, fragmentIndex) => (
                      <tr key={fragmentIndex} className="hover:bg-gray-50">
                        <td className="p-2 whitespace-nowrap text-sm font-medium text-gray-900">
                          {fragment.metadata.id || 'N/A'}<br/>
                          <span className="text-xs text-gray-500">{fragment.metadata.type || 'N/A'}</span>
                        </td>
                        <td className="p-2 whitespace-nowrap text-sm text-gray-700">
                          {fragment.metadata.position || 'middle'}
                        </td>
                        <td className="p-2 whitespace-nowrap text-sm text-gray-700">
                          {fragment.metadata.priority || 0}
                        </td>
                        <td className="p-2 text-sm text-gray-700">
                          <div className="max-h-16 overflow-y-auto">
                            {fragment.content.length > 100 
                              ? `${fragment.content.substring(0, 100)}...` 
                              : fragment.content}
                          </div>
                        </td>
                      </tr>
                    ))
                  ) : (
                    <tr>
                      <td colSpan={4} className="p-2 text-center text-sm text-gray-500">
                        No fragments at this stage
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

// Main app component
const MiddlewareDemo = () => {
  const [middlewares, setMiddlewares] = useState(sampleMiddlewares);
  const [userQuery, setUserQuery] = useState('Explain quantum computing in simple terms');
  const [context, setContext] = useState<Context>({ thinkingMode: true });
  const [finalPrompt, setFinalPrompt] = useState('');
  const [llmResponse, setLlmResponse] = useState('');
  const [editorContent, setEditorContent] = useState('');
  const [processedResponse, setProcessedResponse] = useState('');
  const [finalContext, setFinalContext] = useState<Context>({});
  const [processingActive, setProcessingActive] = useState(false);
  const [fragmentStages, setFragmentStages] = useState([]);
  
  // Function to move middleware up
  const moveUp = useCallback((index: number) => {
    if (index > 0) {
      setMiddlewares(prevMiddlewares => {
        const newMiddlewares = [...prevMiddlewares];
        const temp = newMiddlewares[index];
        newMiddlewares[index] = newMiddlewares[index - 1];
        newMiddlewares[index - 1] = temp;
        return newMiddlewares;
      });
    }
  }, []);
  
  // Function to move middleware down
  const moveDown = useCallback((index: number) => {
    setMiddlewares(prevMiddlewares => {
      if (index < prevMiddlewares.length - 1) {
        const newMiddlewares = [...prevMiddlewares];
        const temp = newMiddlewares[index];
        newMiddlewares[index] = newMiddlewares[index + 1];
        newMiddlewares[index + 1] = temp;
        return newMiddlewares;
      }
      return prevMiddlewares;
    });
  }, []);
  
  // Function to toggle middleware enabled/disabled
  const toggleMiddleware = useCallback((id: string) => {
    setMiddlewares(prevMiddlewares =>
      prevMiddlewares.map(middleware =>
        middleware.id === id
          ? { ...middleware, enabled: !middleware.enabled }
          : middleware
      )
    );
  }, []);
  
  // Process pipeline whenever something changes
  useEffect(() => {
    if (processingActive) {
      processPrompt();
    }
  }, [middlewares, userQuery, context, editorContent, processingActive]);
  
  // Execute middleware pipeline
  const processPrompt = useCallback(() => {
    // Initial fragments with user query
    let currentFragments: PromptFragment[] = [
      {
        content: userQuery,
        metadata: {
          id: 'user-query',
          type: 'query',
          position: 'middle',
          priority: 50
        }
      }
    ];
    
    // Initial context
    let currentContext = { ...context };
    
    // Track fragments at each stage for visualization
    let fragmentsAtEachStage = [];
    fragmentsAtEachStage.push({
      stage: 'Initial',
      fragments: [...currentFragments],
      context: {...currentContext}
    });
    
    // Process through middlewares (prompt phase)
    const enabledMiddlewares = middlewares.filter(m => m.enabled);
    
    for (const middleware of enabledMiddlewares) {
      [currentContext, currentFragments] = middleware.prompt(
        currentContext,
        currentFragments
      );
      
      // Store fragments after this middleware
      fragmentsAtEachStage.push({
        stage: middleware.name,
        fragments: [...currentFragments],
        context: {...currentContext}
      });
    }
    
    // Sort fragments by position and priority
    const positionOrder = { start: 0, middle: 1, end: 2 };
    const sortedFragments = [...currentFragments].sort((a, b) => {
      const aPos = a.metadata.position || 'middle';
      const bPos = b.metadata.position || 'middle';
      
      if (positionOrder[aPos] !== positionOrder[bPos]) {
        return positionOrder[aPos] - positionOrder[bPos];
      }
      
      return (b.metadata.priority || 0) - (a.metadata.priority || 0);
    });
    
    // Store the final sorted fragments
    fragmentsAtEachStage.push({
      stage: 'Final Sorted',
      fragments: sortedFragments,
      context: currentContext
    });
    
    // Combine into final prompt
    const prompt = sortedFragments.map(f => f.content).join('\n\n');
    setFinalPrompt(prompt);
    
    // Save fragment stages for visualization
    setFragmentStages(fragmentsAtEachStage);
    
    // Generate mock LLM response based on enabled middlewares
    const mockResponse = generateMockResponse(enabledMiddlewares, userQuery);
    setLlmResponse(mockResponse);
    
    // Set editor content if it's empty or processing was just activated
    if (!editorContent || editorContent === '') {
      setEditorContent(mockResponse);
    }
    
    // Use editorContent instead of mockResponse for further processing
    // This allows manual edits to be processed
    let processedResp = editorContent;
    let finalCtx = currentContext;
    
    // Process in reverse order for the parsing phase
    for (const middleware of [...enabledMiddlewares].reverse()) {
      [finalCtx, processedResp] = middleware.parse(finalCtx, processedResp);
    }
    
    setProcessedResponse(processedResp);
    setFinalContext(finalCtx);
  }, [context, middlewares, userQuery, editorContent]);
  
  // Start or stop real-time processing
  const toggleProcessing = () => {
    if (!processingActive) {
      // When starting, generate a fresh response
      setEditorContent('');
      setProcessingActive(true);
    } else {
      setProcessingActive(false);
    }
  };
  
  return (
    <div className="container mx-auto p-4">
      <h1 className="text-2xl font-bold mb-6">LLM Middleware Demo</h1>
      
      <div className="bg-yellow-50 p-4 rounded border border-yellow-200 mb-4">
        <h2 className="text-lg font-semibold text-yellow-800">How Middleware Pipeline Works</h2>
        <ul className="list-disc ml-5 mt-2 text-sm text-yellow-700">
          <li>Each middleware adds, modifies, or removes prompt fragments</li>
          <li>The <strong>order of middlewares</strong> determines how they transform fragments</li>
          <li>After all middlewares run, fragments are sorted by position and priority</li>
          <li>This sorting is what determines the final order in the prompt, not middleware order</li>
          <li>Position options are: 'start', 'middle', 'end' (processed in that order)</li>
          <li>Within each position, higher priority fragments come first</li>
        </ul>
      </div>
      
      <div className="mb-4 bg-blue-50 p-4 rounded border border-blue-200">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-blue-800">Real-time Processing</h2>
          <label className="switch relative inline-block w-12 h-6">
            <input
              type="checkbox"
              checked={processingActive}
              onChange={toggleProcessing}
              className="opacity-0 w-0 h-0"
            />
            <span className={`slider absolute cursor-pointer inset-0 rounded-full ${
              processingActive ? 'bg-blue-600' : 'bg-gray-300'
            } transition-colors duration-200`}></span>
          </label>
        </div>
        <p className="text-sm text-blue-700 mt-1">
          {processingActive 
            ? "Real-time processing is ON. Changes will be processed immediately." 
            : "Real-time processing is OFF. Click the button below to process manually."}
        </p>
      </div>
      
      <div className="flex flex-col lg:flex-row gap-6">
        {/* Left Pane - Middleware Configuration */}
        <div className="lg:w-1/2">
          <div className="bg-white rounded-lg shadow p-4">
            <h2 className="text-xl font-semibold mb-4">Middleware Pipeline</h2>
            <p className="text-gray-600 mb-4">Use the up/down arrows to reorder, toggle to enable/disable</p>
            
            <div>
              {middlewares.map((middleware, index) => (
                <MiddlewareItem
                  key={middleware.id}
                  middleware={middleware}
                  index={index}
                  moveUp={moveUp}
                  moveDown={moveDown}
                  isFirst={index === 0}
                  isLast={index === middlewares.length - 1}
                  toggleMiddleware={toggleMiddleware}
                />
              ))}
            </div>
            
            <div className="mt-4">
              <h3 className="text-lg font-medium mb-2">Initial Context</h3>
              <ContextViewer context={context} />
              <div className="flex items-center mb-4 mt-2">
                <input
                  type="checkbox"
                  id="thinking-mode"
                  checked={context.thinkingMode || false}
                  onChange={(e) => setContext({...context, thinkingMode: e.target.checked})}
                  className="mr-2"
                />
                <label htmlFor="thinking-mode" className="text-sm">
                  Enable Thinking Mode
                </label>
              </div>
            </div>
            
            <div className="mt-4">
              <h3 className="text-lg font-medium mb-2">User Query</h3>
              <textarea
                value={userQuery}
                onChange={(e) => setUserQuery(e.target.value)}
                className="w-full p-2 border rounded"
                rows={2}
              />
            </div>
            
            {!processingActive && (
              <button
                onClick={processPrompt}
                className="mt-4 bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
              >
                Process with Middlewares
              </button>
            )}
          </div>
        </div>
        
        {/* Right Pane - Output Display */}
        <div className="lg:w-1/2">
          <div className="bg-white rounded-lg shadow p-4 h-full overflow-auto">
            <h2 className="text-xl font-semibold mb-4">Pipeline Results</h2>
            
            {finalPrompt && (
              <div className="mb-6">
                <h3 className="font-medium mb-1">Final Prompt</h3>
                <div className="bg-gray-100 p-3 rounded overflow-auto max-h-48 text-sm">
                  {finalPrompt.split('\n\n').map((paragraph, i) => (
                    <p key={i} className="mb-2">{paragraph}</p>
                  ))}
                </div>
              </div>
            )}
            
            {fragmentStages.length > 0 && (
              <div className="mb-6">
                <FragmentVisualizer fragmentStages={fragmentStages} />
              </div>
            )}
            
            {llmResponse && (
              <div className="mb-6">
                <h3 className="font-medium mb-1">LLM Response (Editable)</h3>
                <textarea
                  value={editorContent}
                  onChange={(e) => setEditorContent(e.target.value)}
                  className="w-full p-2 border rounded font-mono text-sm"
                  rows={8}
                />
                <p className="text-xs text-gray-500 mt-1">
                  Edit the response above to see how middlewares process different content
                </p>
              </div>
            )}
            
            {processedResponse && (
              <div className="mb-6">
                <h3 className="font-medium mb-1">Processed Response</h3>
                <div className="bg-gray-100 p-3 rounded overflow-auto max-h-48 text-sm">
                  {processedResponse.split('\n\n').map((paragraph, i) => (
                    <p key={i} className="mb-2">{paragraph}</p>
                  ))}
                </div>
              </div>
            )}
            
            {Object.keys(finalContext).length > 0 && (
              <div>
                <h3 className="font-medium mb-1">Final Context</h3>
                <ContextViewer context={finalContext} />
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default MiddlewareDemo;