# AI Engineer World Fair Demos - Tool Calling Animations

This Remotion project creates educational animations explaining how Large Language Models (LLMs) perform tool calling, including both successful scenarios and common inefficiencies.

## Animations

### 1. ToolCallingAnimation - Efficient Tool Use
Demonstrates the ideal 4-step process of LLM tool calling:

1. **User Request**: User sends a weather request to the LLM
2. **Tool Analysis**: LLM analyzes available tools and selects the appropriate Weather API
3. **Tool Execution**: LLM calls the selected tool and receives precise results
4. **Result Integration**: LLM integrates the tool results into a comprehensive response

### 2. CRMQueryAnimation - Token Inefficiency Problem
Shows how simple queries can lead to massive token waste:

1. **Simple Request**: User asks for "OpenAI contact information"
2. **Poor Tool Choice**: LLM selects `get_crm_companies()` which returns ALL companies
3. **Data Overload**: Tool returns 36+ companies (3,600+ tokens) when only 1 was needed
4. **Inefficient Processing**: LLM must scan through massive dataset to find the single answer

### 3. SQLiteQueryAnimation - Intelligent Multi-Step Tool Use
Demonstrates sophisticated tool calling with exploration and precision:

1. **Complex Request**: User asks "How many orders did John Smith place last month?"
2. **Schema Discovery**: LLM uses `sqlite_query()` to explore available tables
3. **Structure Analysis**: Multiple queries to understand table relationships and columns
4. **Targeted Execution**: Crafts precise JOIN query with filters for exact result
5. **Efficient Response**: Gets exactly what's needed with minimal token usage (~300 vs 3,600+)

### 4. SQLiteViewOptimizationAnimation - Infrastructure for Multiple Queries
Shows how to optimize for repeated queries by creating reusable database views:

1. **View Creation**: LLM creates `customer_orders_view` to pre-join tables with meaningful names
2. **Multiple Queries**: Runs 4 different analytics queries (count, sum, average, latest) using the view
3. **Performance Comparison**: Before (16 JOIN operations) vs After (1 JOIN total) with 75% code reduction

### 5. ComprehensiveComparisonAnimation - Evolution of Tool Intelligence
Demonstrates the complete journey from inefficient calls to intelligent infrastructure:

1. **Token Efficiency Comparison**: Shows that even schema exploration + view creation (400 tokens) is 90% more efficient than bulk CRM calls (3,650 tokens)
2. **View Persistence**: How SQL views are saved with metadata and become discoverable infrastructure
3. **Tool Discovery**: System startup automatically scans for views and registers them as callable tools
4. **Future Efficiency**: New queries use existing views as instant tools (50 tokens vs 3,650+ tokens)

## Getting Started

1. Install dependencies:
```bash
npm install
```

2. Start the preview server:
```bash
npm start
```

3. Build the video:
```bash
npm run build
```

## Rendering Individual Clips

For presentations where you want to stop and discuss each step, you can render individual clips:

### List all available clips:
```bash
node render-clips.js --list
```

### Render specific steps:
```bash
# Weather API steps
node render-clips.js weather-step1-user-request
node render-clips.js weather-step2-tool-analysis
node render-clips.js weather-step3-tool-execution
node render-clips.js weather-step4-result-integration

# CRM inefficiency steps
node render-clips.js crm-step1-user-request
node render-clips.js crm-step2-tool-analysis
node render-clips.js crm-step3-tool-execution
node render-clips.js crm-step4-result-processing

# SQLite multi-step steps
node render-clips.js sqlite-step1-user-request
node render-clips.js sqlite-step2-schema-discovery
node render-clips.js sqlite-step3-table-exploration
node render-clips.js sqlite-step4-targeted-query
node render-clips.js sqlite-step5-final-response

# SQLite view optimization steps
node render-clips.js sqlite-view-step1-view-creation
node render-clips.js sqlite-view-step2-multiple-queries
node render-clips.js sqlite-view-step3-performance-comparison

# Comprehensive comparison steps
node render-clips.js comparison-step1-token-efficiency
node render-clips.js comparison-step2-view-persistence
node render-clips.js comparison-step3-tool-discovery
node render-clips.js comparison-step4-future-efficiency
```

### Render all clips at once:
```bash
node render-clips.js --all
```

### Render full animations:
```bash
node render-clips.js weather-full
node render-clips.js crm-full
node render-clips.js sqlite-full
node render-clips.js sqlite-view-optimization-full
node render-clips.js comprehensive-comparison-full
```

## Animation Details

- **Duration**: 40 seconds (1200 frames at 30fps)
- **Resolution**: 1920x1080 (Full HD)
- **Format**: MP4

## Sequences

- **UserRequestSequence** (0-6s): Shows user sending weather request
- **ToolAnalysisSequence** (6-14s): LLM analyzing and selecting weather API tool
- **ToolExecutionSequence** (14-24s): API call execution and response
- **ResultIntegrationSequence** (24-36s): Final response generation and delivery

## Customization

Each sequence is in its own component under `src/sequences/`. You can modify timing, styling, or content by editing these individual components.

The main composition settings can be adjusted in `src/Root.tsx`.
