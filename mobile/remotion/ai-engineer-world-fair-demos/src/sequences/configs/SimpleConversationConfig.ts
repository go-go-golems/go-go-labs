import {
	InteractionSequence,
	createState,
	createMessage,
	createMessageType,
	DEFAULT_MESSAGE_TYPES,
} from '../../types/InteractionDSL';

// Custom message types for this sequence
const customMessageTypes = {
	...DEFAULT_MESSAGE_TYPES,
	system: createMessageType('#34495e', '⚙️', 'System', {
		fontSize: '12px',
		padding: '10px 14px',
		fontStyle: 'italic',
	}),
	error: createMessageType('#e74c3c', '❌', 'Error', {
		fontSize: '12px',
		padding: '10px 14px',
		border: '1px solid rgba(255, 0, 0, 0.3)',
	}),
};

export const simpleConversationSequence: InteractionSequence = {
	title: 'Simple API Request Flow',
	subtitle: 'User → Assistant → Tool → Response',
	messageTypes: customMessageTypes,
	
	states: [
		createState('container', 0, 20),
		createState('userQuestion', 20, 15),
		createState('thinking', 35, 15),
		createState('toolCall', 50, 15),
		createState('toolResponse', 65, 15),
		createState('finalAnswer', 80, 15),
		createState('systemNote', 95, 15),
	],

	messages: [
		createMessage(
			'user-query',
			'user',
			'What is the current temperature in New York?',
			['userQuestion', 'thinking', 'toolCall', 'toolResponse', 'finalAnswer', 'systemNote']
		),

		createMessage(
			'assistant-thinking',
			'assistant_cot',
			'User wants weather data for NYC. I need to call the weather API.',
			['thinking', 'toolCall', 'toolResponse', 'finalAnswer', 'systemNote']
		),

		createMessage(
			'weather-api-call',
			'tool_use',
			'get_weather(location="New York, NY")',
			['toolCall', 'toolResponse', 'finalAnswer', 'systemNote']
		),

		createMessage(
			'api-response',
			'tool_result',
			'{"temperature": 72, "condition": "sunny", "humidity": 45}',
			['toolResponse', 'finalAnswer', 'systemNote']
		),

		createMessage(
			'assistant-answer',
			'assistant',
			'The current temperature in New York is 72°F with sunny conditions.',
			['finalAnswer', 'systemNote']
		),

		createMessage(
			'system-log',
			'system',
			'Request completed successfully in 1.2s',
			['systemNote']
		),
	],

	layout: {
		columns: 1,
		autoFill: true,
	},

	tokenCounter: {
		enabled: true,
		initialTokens: 150,
		maxTokens: 128000,
		stateTokenCounts: {
			'userQuestion': 150,
			'thinking': 180,
			'toolCall': 220,
			'toolResponse': 280,
			'finalAnswer': 320,
			'systemNote': 340,
		},
	},
}; 