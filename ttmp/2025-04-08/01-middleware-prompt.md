I'll write a detailed technical specification for your middleware pattern, focusing on how it works from a senior developer's perspective.

# LLM Middleware Pattern Specification

## Introduction

The LLM Middleware Pattern provides a powerful abstraction layer for managing interactions with Large Language Models. It creates a pipeline of processing units that transform prompts before they're sent to an LLM and parse responses after they're received. This pattern is particularly valuable when working with LLMs because it:

1. Decouples concerns in LLM interaction logic
2. Enables composable, reusable components
3. Provides a consistent interface for prompt manipulation and response processing
4. Facilitates context sharing between processing stages

## Core Architecture

At its heart, the middleware pattern revolves around three key concepts:

### 1. Context

Context is a flexible, extensible state object that flows through the middleware pipeline. It serves as a shared memory space where middlewares can:
- Store metadata about the current request/response cycle
- Share information between different middlewares
- Accumulate state that persists across the entire pipeline

```typescript
// Simple, extensible context interface
interface Context {
  [key: string]: any;
}
```

The beauty of this approach is that context can be extended dynamically during runtime. Middlewares can add arbitrary properties, allowing for emergent behaviors when middlewares are combined.

### 2. Prompt Fragments

Rather than treating prompts as monolithic strings, we break them down into structured "fragments" with metadata. This enables fine-grained control over how prompts are constructed.

```typescript
interface PromptFragmentMetadata {
  id?: string;                                  // Unique identifier for the fragment
  type?: string;                                // Classification (e.g., "system", "user", "example")
  position?: 'start' | 'middle' | 'end';        // Desired position in final prompt
  priority?: number;                            // Ordering weight within position
  tags?: string[];                              // Arbitrary labels for filtering/selection
}

interface PromptFragment {
  content: string;                              // The actual text content
  metadata: PromptFragmentMetadata;             // Associated metadata
}
```

Prompt fragments give us several advantages:
- We can add, remove, or modify specific parts of the prompt
- We can declaratively control the final ordering through metadata
- We maintain semantic meaning alongside the text content

### 3. Middleware Pipeline

The pipeline is a sequence of middleware components that each provide two key functions:

```typescript
interface Middleware {
  // Transforms context and prompt fragments before sending to LLM
  prompt(context: Context, fragments: PromptFragment[]): [Context, PromptFragment[]];
  
  // Processes LLM response and updates context
  parse(context: Context, response: string): [Context, string];
}
```

Each middleware can:
- Transform the context by adding, modifying, or removing properties
- Transform prompt fragments by adding, modifying, or removing fragments
- Process LLM responses by extracting information or modifying the response text

## Pipeline Execution Flow

The execution flow has two distinct phases:

### Prompt Phase

1. Start with initial context and prompt fragments
2. Pass through each middleware's `prompt()` function in registration order
3. After all middlewares, sort fragments by position and priority
4. Concatenate sorted fragments to create the final prompt
5. Send final prompt to the LLM

```typescript
function executePromptPhase(
  initialContext: Context, 
  initialFragments: PromptFragment[]
): [Context, string] {
  let currentContext = { ...initialContext };
  let currentFragments = [...initialFragments];
  
  // Pass through each middleware's prompt function
  for (const middleware of middlewares) {
    [currentContext, currentFragments] = middleware.prompt(
      currentContext, 
      currentFragments
    );
  }
  
  // Sort fragments by position and priority
  const sortedFragments = sortFragments(currentFragments);
  
  // Combine into final prompt
  const finalPrompt = combineFragments(sortedFragments);
  
  return [currentContext, finalPrompt];
}
```

### Parse Phase

1. Receive response from LLM
2. Pass through each middleware's `parse()` function in reverse registration order
3. Return final context and processed response

```typescript
function executeParsePhase(
  context: Context, 
  llmResponse: string
): [Context, string] {
  let currentContext = { ...context };
  let currentResponse = llmResponse;
  
  // Process in reverse order for symmetry
  for (const middleware of [...middlewares].reverse()) {
    [currentContext, currentResponse] = middleware.parse(
      currentContext, 
      currentResponse
    );
  }
  
  return [currentContext, currentResponse];
}
```

Note that we process middlewares in reverse order during the parse phase. This creates a symmetrical pattern where the last middleware to modify the prompt is the first to see the response.

## Fragment Composition

A key aspect of the pattern is how fragments are ultimately composed into the final prompt. We recommend a structured approach:

```typescript
function sortFragments(fragments: PromptFragment[]): PromptFragment[] {
  const positionOrder = { start: 0, middle: 1, end: 2 };
  
  return [...fragments].sort((a, b) => {
    const aPos = a.metadata.position || 'middle';
    const bPos = b.metadata.position || 'middle';
    
    // First sort by position
    if (positionOrder[aPos] !== positionOrder[bPos]) {
      return positionOrder[aPos] - positionOrder[bPos];
    }
    
    // Then by priority (higher numbers come first)
    return (b.metadata.priority || 0) - (a.metadata.priority || 0);
  });
}

function combineFragments(fragments: PromptFragment[]): string {
  return fragments
    .filter(f => f.content.trim() !== '')  // Remove empty fragments
    .map(f => f.content)
    .join('\n\n');  // Join with double newlines
}
```

This approach ensures that fragments appear in a logical order in the final prompt, regardless of when or where they were added in the middleware pipeline.

## Pipeline Configuration

A middleware pipeline can be configured through a simple builder pattern:

```typescript
class MiddlewarePipeline {
  private middlewares: Middleware[] = [];
  
  use(middleware: Middleware): MiddlewarePipeline {
    this.middlewares.push(middleware);
    return this;
  }
  
  async execute(initialContext: Context, initialFragments: PromptFragment[] = []): Promise<[Context, string]> {
    // Execute prompt phase
    const [contextAfterPrompt, finalPrompt] = this.executePromptPhase(initialContext, initialFragments);
    
    // Call LLM
    const llmResponse = await this.callLLM(finalPrompt);
    
    // Execute parse phase
    return this.executeParsePhase(contextAfterPrompt, llmResponse);
  }
  
  // Implementation details omitted for brevity
}
```

This allows for clean, chainable configuration:

```typescript
const pipeline = new MiddlewarePipeline()
  .use(new SystemInstructionMiddleware())
  .use(new ThinkingModeMiddleware())
  .use(new OutputFormatMiddleware());
```

## Example Middleware Implementations

Let's look at some concrete examples of middleware implementations:

### System Instruction Middleware

```typescript
class SystemInstructionMiddleware implements Middleware {
  private instructions: string;
  
  constructor(instructions: string = "You are a helpful assistant.") {
    this.instructions = instructions;
  }
  
  prompt(context: Context, fragments: PromptFragment[]): [Context, PromptFragment[]] {
    const systemFragment: PromptFragment = {
      content: this.instructions,
      metadata: {
        id: 'system-instruction',
        type: 'system',
        position: 'start',
        priority: 100
      }
    };
    
    return [context, [systemFragment, ...fragments]];
  }
  
  parse(context: Context, response: string): [Context, string] {
    // This middleware doesn't modify the response
    return [context, response];
  }
}
```

### Thinking Mode Middleware

```typescript
class ThinkingModeMiddleware implements Middleware {
  prompt(context: Context, fragments: PromptFragment[]): [Context, PromptFragment[]] {
    // Check if thinking mode is enabled in context
    const thinkingEnabled = context.thinkingMode || false;
    
    if (thinkingEnabled) {
      const thinkingFragment: PromptFragment = {
        content: "Show your reasoning step by step in <thinking>...</thinking> tags.",
        metadata: {
          id: 'thinking-instruction',
          type: 'instruction',
          position: 'middle',
          priority: 60
        }
      };
      
      return [context, [...fragments, thinkingFragment]];
    }
    
    return [context, fragments];
  }
  
  parse(context: Context, response: string): [Context, string] {
    // Extract thinking section from response
    const thinkingRegex = /<thinking>([\s\S]*?)<\/thinking>/;
    const match = response.match(thinkingRegex);
    
    if (match) {
      const thinking = match[1].trim();
      const newContext = { ...context, extractedThinking: thinking };
      
      // Remove thinking section from response
      const cleanedResponse = response.replace(thinkingRegex, '').trim();
      
      return [newContext, cleanedResponse];
    }
    
    return [context, response];
  }
}
```

### Token Counter Middleware

```typescript
class TokenCounterMiddleware implements Middleware {
  prompt(context: Context, fragments: PromptFragment[]): [Context, PromptFragment[]] {
    // Estimate tokens in all fragments (simplified)
    const combinedText = fragments.map(f => f.content).join(' ');
    const estimatedTokens = this.estimateTokens(combinedText);
    
    return [{ ...context, promptTokens: estimatedTokens }, fragments];
  }
  
  parse(context: Context, response: string): [Context, string] {
    // Estimate tokens in response
    const responseTokens = this.estimateTokens(response);
    const totalTokens = (context.promptTokens || 0) + responseTokens;
    
    return [
      { 
        ...context, 
        responseTokens, 
        totalTokens 
      }, 
      response
    ];
  }
  
  private estimateTokens(text: string): number {
    // Simplified token estimation (in production, use a proper tokenizer)
    return Math.ceil(text.length / 4);
  }
}
```

## Advanced Pattern Features

### Middleware Ordering and Dependencies

It's important to understand that the order of middleware registration can significantly impact behavior. Middlewares can:

1. Build on context established by earlier middlewares
2. Modify or remove fragments added by earlier middlewares
3. Add fragments that will be processed by later middlewares

For this reason, some middlewares may have implicit dependencies on others. It's good practice to document these dependencies:

```typescript
class OutputFormatMiddleware implements Middleware {
  // This middleware should run after SystemInstructionMiddleware
  // and before TokenCounterMiddleware
  // ...
}
```

### Error Handling

The pattern should account for middleware errors:

```typescript
async execute(initialContext: Context, initialFragments: PromptFragment[]): Promise<[Context, string]> {
  try {
    // Execute middleware pipeline...
  } catch (error) {
    // Log error
    console.error("Middleware pipeline error:", error);
    
    // Add error to context
    const errorContext = { 
      ...initialContext, 
      error, 
      errorPhase: 'prompt' 
    };
    
    // Optional: Run error recovery middleware
    if (this.errorMiddleware) {
      return this.errorMiddleware.handle(errorContext, initialFragments);
    }
    
    throw error;
  }
}
```

### Conditional Middleware

Some middleware might only need to run in specific circumstances:

```typescript
class ConditionalMiddleware implements Middleware {
  private condition: (context: Context) => boolean;
  private innerMiddleware: Middleware;
  
  constructor(condition: (context: Context) => boolean, middleware: Middleware) {
    this.condition = condition;
    this.innerMiddleware = middleware;
  }
  
  prompt(context: Context, fragments: PromptFragment[]): [Context, PromptFragment[]] {
    if (this.condition(context)) {
      return this.innerMiddleware.prompt(context, fragments);
    }
    return [context, fragments];
  }
  
  parse(context: Context, response: string): [Context, string] {
    if (this.condition(context)) {
      return this.innerMiddleware.parse(context, response);
    }
    return [context, response];
  }
}
```

## Porting to Go

When porting this pattern to Go, consider these adaptations:

1. Use interfaces for the core Middleware type:

```go
type Context map[string]interface{}

type PromptFragmentMetadata struct {
  ID       string   `json:"id,omitempty"`
  Type     string   `json:"type,omitempty"`
  Position string   `json:"position,omitempty"`
  Priority int      `json:"priority,omitempty"`
  Tags     []string `json:"tags,omitempty"`
}

type PromptFragment struct {
  Content  string                `json:"content"`
  Metadata PromptFragmentMetadata `json:"metadata"`
}

type Middleware interface {
  Prompt(context Context, fragments []PromptFragment) (Context, []PromptFragment)
  Parse(context Context, response string) (Context, string)
}
```

2. Use slices and maps instead of arrays and objects:

```go
func SortFragments(fragments []PromptFragment) []PromptFragment {
  positionOrder := map[string]int{
    "start":  0,
    "middle": 1,
    "end":    2,
  }
  
  // Create a copy to avoid modifying the original
  sortedFragments := make([]PromptFragment, len(fragments))
  copy(sortedFragments, fragments)
  
  sort.SliceStable(sortedFragments, func(i, j int) bool {
    // Position comparison logic
    posI := "middle"
    if sortedFragments[i].Metadata.Position != "" {
      posI = sortedFragments[i].Metadata.Position
    }
    
    posJ := "middle"
    if sortedFragments[j].Metadata.Position != "" {
      posJ = sortedFragments[j].Metadata.Position
    }
    
    if positionOrder[posI] != positionOrder[posJ] {
      return positionOrder[posI] < positionOrder[posJ]
    }
    
    // Priority comparison (higher first)
    return sortedFragments[i].Metadata.Priority > sortedFragments[j].Metadata.Priority
  })
  
  return sortedFragments
}
```

3. Use method receivers for middleware implementations:

```go
type SystemInstructionMiddleware struct {
  Instructions string
}

func (m *SystemInstructionMiddleware) Prompt(context Context, fragments []PromptFragment) (Context, []PromptFragment) {
  systemFragment := PromptFragment{
    Content: m.Instructions,
    Metadata: PromptFragmentMetadata{
      ID:       "system-instruction",
      Type:     "system",
      Position: "start",
      Priority: 100,
    },
  }
  
  return context, append([]PromptFragment{systemFragment}, fragments...)
}

func (m *SystemInstructionMiddleware) Parse(context Context, response string) (Context, string) {
  return context, response
}
```

## Conclusion

The LLM Middleware Pattern offers a powerful, flexible approach to managing LLM interactions. By breaking prompt creation and response parsing into composable units, it enables:

1. Cleaner separation of concerns
2. More maintainable, testable code
3. Reusable components that can be mixed and matched
4. Dynamic behavior based on context

This pattern can be extended with additional features like middleware groups, conditional execution, and error recovery mechanisms. The core concepts—context, fragments, and middleware functions—provide a solid foundation for building sophisticated LLM interaction systems that can evolve with your application's needs.