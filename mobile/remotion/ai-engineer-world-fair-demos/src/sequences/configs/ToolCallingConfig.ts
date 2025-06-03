import {
	InteractionSequence,
	createState,
	createMessage,
	createMessageType,
	DEFAULT_MESSAGE_TYPES,
	InteractionState,
} from '../../types/InteractionDSL';

// Custom message types for tool calling sequence
const toolCallingMessageTypes = {
	...DEFAULT_MESSAGE_TYPES,
	weather_api: createMessageType('#3498db', '🌤️', 'Weather API', {
		fontSize: '12px',
		padding: '10px 14px',
		boxShadow: '0 3px 10px rgba(52, 152, 219, 0.3)',
		border: '1px solid rgba(52, 152, 219, 0.3)',
	}),
	step_indicator: createMessageType('#95a5a6', '📍', 'Step', {
		fontSize: '14px',
		padding: '8px 16px',
		fontWeight: 'bold',
		boxShadow: '0 2px 8px rgba(149, 165, 166, 0.3)',
	}),
	thinking: createMessageType('#f39c12', '🤔', 'Analyzing', {
		fontSize: '11px',
		padding: '8px 12px',
		fontStyle: 'italic',
		border: '1px solid rgba(243, 156, 18, 0.3)',
	}),
	tool_selection: createMessageType('#e74c3c', '🎯', 'Tool Selection', {
		fontSize: '12px',
		padding: '10px 14px',
		border: '1px solid rgba(231, 76, 60, 0.3)',
	}),
};

export const toolCallingSequence: InteractionSequence = {
	title: 'How LLMs Use Tools',
	
	subtitle: (state: InteractionState) => {
		if (state.activeStates.includes('userRequest')) {
			return 'User sends a weather query';
		} else if (state.activeStates.includes('llmReceives')) {
			return 'LLM receives and understands the request';
		} else if (state.activeStates.includes('toolAnalysis')) {
			return 'LLM analyzes available tools';
		} else if (state.activeStates.includes('toolSelection')) {
			return 'LLM selects the best tool for the task';
		} else if (state.activeStates.includes('toolExecution')) {
			return 'LLM executes the weather API call';
		} else if (state.activeStates.includes('apiProcessing')) {
			return 'Weather service processes the request';
		} else if (state.activeStates.includes('resultIntegration')) {
			return 'LLM processes and integrates the results';
		} else if (state.activeStates.includes('finalResponse')) {
			return 'LLM provides natural language response';
		} else if (state.activeStates.includes('workflowComplete')) {
			return 'Complete tool calling workflow demonstrated';
		}
		return 'A seamless demonstration of LLM tool calling';
	},

	messageTypes: toolCallingMessageTypes,
	
	states: [
		// Seamless flow with overlapping states for natural conversation
		createState('container', 0, 60),
		createState('userRequest', 60, 120),
		createState('llmReceives', 120, 90),
		createState('toolAnalysis', 180, 150),
		createState('toolSelection', 270, 120),
		createState('toolExecution', 360, 180),
		createState('apiProcessing', 450, 150),
		createState('resultIntegration', 540, 180),
		createState('finalResponse', 660, 120),
		createState('workflowComplete', 720, 60),
	],

	messages: [
		// User Request
		createMessage(
			'user-weather-request',
			'user',
			'"What\'s the weather like in San Francisco today?"',
			['userRequest', 'llmReceives', 'toolAnalysis', 'toolSelection', 'toolExecution', 'apiProcessing', 'resultIntegration', 'finalResponse', 'workflowComplete']
		),

		createMessage(
			'llm-receives-request',
			'assistant',
			'I need to get current weather information for San Francisco. Let me check what tools are available.',
			['llmReceives', 'toolAnalysis', 'toolSelection', 'toolExecution', 'apiProcessing', 'resultIntegration', 'finalResponse', 'workflowComplete']
		),

		// Tool Analysis
		createMessage(
			'llm-thinking',
			'thinking',
			'Analyzing request: User wants weather data → Need location-based weather service → Check available tools...',
			['toolAnalysis', 'toolSelection', 'toolExecution', 'apiProcessing', 'resultIntegration', 'finalResponse', 'workflowComplete']
		),

		createMessage(
			'available-tools',
			'tool_selection',
			`Available Tools:
• get_weather(location) - ✅ Perfect match!
• calculate(expression) - Not needed
• search_web(query) - Could work but weather API is better
• send_email(to, subject, body) - Not relevant`,
			['toolAnalysis', 'toolSelection', 'toolExecution', 'apiProcessing', 'resultIntegration', 'finalResponse', 'workflowComplete']
		),

		createMessage(
			'tool-selected',
			'assistant_cot',
			'Selected: get_weather() - This tool provides real-time weather data for any location.',
			['toolSelection', 'toolExecution', 'apiProcessing', 'resultIntegration', 'finalResponse', 'workflowComplete']
		),

		// Tool Execution
		createMessage(
			'weather-api-call',
			'tool_use',
			'get_weather(location="San Francisco, CA")',
			['toolExecution', 'apiProcessing', 'resultIntegration', 'finalResponse', 'workflowComplete']
		),

		createMessage(
			'api-processing',
			'weather_api',
			'🌐 Connecting to weather service...\n📍 Locating San Francisco, CA...\n🌤️ Fetching current conditions...',
			['apiProcessing', 'resultIntegration', 'finalResponse', 'workflowComplete']
		),

		createMessage(
			'weather-result',
			'tool_result',
			`{
  "location": "San Francisco, CA",
  "temperature": 68,
  "condition": "Partly Cloudy",
  "humidity": 72,
  "wind_speed": 12,
  "timestamp": "2024-01-15T14:30:00Z"
}`,
			['apiProcessing', 'resultIntegration', 'finalResponse', 'workflowComplete']
		),

		// Result Integration
		createMessage(
			'llm-processing-result',
			'assistant_cot',
			'Processing weather data: 68°F, partly cloudy, 72% humidity, 12mph wind. Converting to natural language response...',
			['resultIntegration', 'finalResponse', 'workflowComplete']
		),

		createMessage(
			'final-response',
			'assistant',
			'The current weather in San Francisco is 68°F with partly cloudy skies. The humidity is at 72% and there\'s a light breeze at 12 mph. It\'s a pleasant day!',
			['finalResponse', 'workflowComplete']
		),

		createMessage(
			'workflow-summary',
			'summary',
			'✅ Complete workflow: User request → Tool analysis → API call → Result integration → Natural response',
			['workflowComplete']
		),
	],

	overlays: [
	],

	layout: {
		columns: 1,
		autoFill: true,
		maxMessagesPerColumn: 15,
	},

}; 