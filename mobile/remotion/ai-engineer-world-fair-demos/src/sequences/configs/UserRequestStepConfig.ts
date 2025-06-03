import {
	InteractionSequence,
	createState,
	createMessage,
	createMessageType,
	DEFAULT_MESSAGE_TYPES,
} from '../../types/InteractionDSL';

// Custom message types for user request step
const userRequestMessageTypes = {
	...DEFAULT_MESSAGE_TYPES,
	step_indicator: createMessageType('#95a5a6', '📍', 'Step', {
		fontSize: '16px',
		padding: '10px 18px',
		fontWeight: 'bold',
		boxShadow: '0 3px 12px rgba(149, 165, 166, 0.4)',
	}),
	user_thinking: createMessageType('#3498db', '💭', 'User Thinking', {
		fontSize: '12px',
		padding: '8px 14px',
		fontStyle: 'italic',
		border: '1px solid rgba(52, 152, 219, 0.3)',
	}),
};

export const userRequestStepSequence: InteractionSequence = {
	title: 'Step 1: User Request',
	subtitle: 'User sends a weather query to the LLM',
	messageTypes: userRequestMessageTypes,
	
	states: [
		createState('container', 0, 30),
		createState('userAppears', 30, 40),
		createState('userThinking', 70, 50),
		createState('userSpeaks', 120, 60),
		createState('messageTravel', 180, 30),
	],

	messages: [
		createMessage(
			'step-indicator',
			'step_indicator',
			'Step 1: User Request',
			['userAppears', 'userThinking', 'userSpeaks', 'messageTravel'],
			{ column: 'left' }
		),

		createMessage(
			'user-thinking',
			'user_thinking',
			'I wonder what the weather is like in San Francisco today...',
			['userThinking', 'userSpeaks', 'messageTravel'],
			{ column: 'left' }
		),

		createMessage(
			'user-question',
			'user',
			'"What\'s the weather like in San Francisco today?"',
			['userSpeaks', 'messageTravel'],
			{ column: 'left' }
		),

		createMessage(
			'message-received',
			'assistant',
			'Message received. Let me help you with the weather information for San Francisco.',
			['messageTravel'],
			{ column: 'right' }
		),
	],

	overlays: [
		{
			id: 'workflow-step',
			content: () => `
				<div style="
					background-color: rgba(52, 152, 219, 0.9);
					color: white;
					padding: 10px 16px;
					border-radius: 20px;
					font-size: 12px;
					font-weight: bold;
					box-shadow: 0 3px 10px rgba(52, 152, 219, 0.4);
				">
					👤 USER REQUEST (1/4)
				</div>
			`,
			position: {
				top: '10%',
				right: '5%',
			},
			visibleStates: ['userAppears', 'userThinking', 'userSpeaks', 'messageTravel'],
		},
	],

	layout: {
		columns: 2,
		autoFill: false,
		maxMessagesPerColumn: 4,
	},

	tokenCounter: {
		enabled: true,
		initialTokens: 150,
		maxTokens: 128000,
		stateTokenCounts: {
			'userAppears': 150,
			'userThinking': 150,
			'userSpeaks': 180,
			'messageTravel': 200,
		},
	},
}; 