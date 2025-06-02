import {
	InteractionSequence,
	createState,
	createMessage,
	DEFAULT_MESSAGE_TYPES,
	InteractionState,
} from '../../types/InteractionDSL';

export const assistantDiscussionSequence: InteractionSequence = {
	title: 'Assistant Mode Discussion: Internal Collaboration',
	
	subtitle: (state: InteractionState) => {
		if (state.activeStates.includes('discussion')) {
			return 'Two assistant modes collaborating internally';
		} else if (state.activeStates.includes('actionBlocks')) {
			return 'Generating coordinated action blocks';
		}
		return 'How LLMs simulate internal discussions between specialized modes';
	},

	messageTypes: {
		...DEFAULT_MESSAGE_TYPES,
		// User request
		user_request: {
			bg: '#3498db',
			icon: '👤',
			label: 'User Request',
			fontSize: '13px',
			padding: '14px 18px',
			boxShadow: '0 4px 15px rgba(52, 152, 219, 0.4)',
		},
		// Internal discussion modes
		research_mode: {
			bg: '#e67e22',
			icon: '🔍',
			label: 'Research Assistant',
			fontSize: '12px',
			padding: '12px 15px',
			boxShadow: '0 3px 10px rgba(230, 126, 34, 0.3)',
			fontStyle: 'italic',
		},
		code_mode: {
			bg: '#27ae60',
			icon: '💻',
			label: 'Code Assistant',
			fontSize: '12px',
			padding: '12px 15px',
			boxShadow: '0 3px 10px rgba(39, 174, 96, 0.3)',
			fontStyle: 'italic',
		},
		// Coordinator
		coordinator: {
			bg: '#8e44ad',
			icon: '🧠',
			label: 'LLM Coordinator',
			fontSize: '13px',
			padding: '14px 18px',
			boxShadow: '0 4px 15px rgba(142, 68, 173, 0.4)',
			fontWeight: '500',
		},
		// Action blocks
		action_research: {
			bg: '#d35400',
			icon: '📋',
			label: 'Research Action',
			fontSize: '11px',
			padding: '10px 12px',
			boxShadow: '0 2px 8px rgba(211, 84, 0, 0.3)',
			border: '2px solid rgba(255, 255, 255, 0.3)',
		},
		action_code: {
			bg: '#16a085',
			icon: '⚡',
			label: 'Code Action',
			fontSize: '11px',
			padding: '10px 12px',
			boxShadow: '0 2px 8px rgba(22, 160, 133, 0.3)',
			border: '2px solid rgba(255, 255, 255, 0.3)',
		},
	},
	
	states: [
		createState('container', 0, 30),
		createState('userRequest', 30, 40),
		createState('coordination', 70, 50),
		createState('discussion', 120, 120),
		createState('actionBlocks', 240, 80),
		createState('finalResponse', 320, 40),
	],

	messages: [
		// User request
		createMessage(
			'user-question',
			'user_request',
			'Help me build an ML model for stock prediction - research and implement?',
			['userRequest', 'coordination', 'discussion', 'actionBlocks', 'finalResponse'],
			{ column: 'left' }
		),

		// LLM Coordinator initiates internal discussion
		createMessage(
			'coordinator-init',
			'coordinator',
			'INTERNAL MODE: Simulating Research + Code assistant discussion...',
			['coordination', 'discussion', 'actionBlocks', 'finalResponse'],
			{ column: 'left' }
		),

		// Research Assistant perspective
		createMessage(
			'research-assistant-input',
			'research_mode',
			'RESEARCH: "Need to compare LSTM, ARIMA, and transformers for financial data."',
			['discussion', 'actionBlocks', 'finalResponse'],
			{ column: 'left' }
		),

		// Code Assistant response
		createMessage(
			'code-assistant-input',
			'code_mode',
			'CODE: "I\'ll build LSTM with TensorFlow. Need preprocessing and metrics."',
			['discussion', 'actionBlocks', 'finalResponse'],
			{ column: 'right' }
		),

		// Research Assistant follow-up
		createMessage(
			'research-followup',
			'research_mode',
			'RESEARCH: "I\'ll gather recent papers and preprocessing techniques."',
			['discussion', 'actionBlocks', 'finalResponse'],
			{ column: 'left' }
		),

		// Code Assistant agreement
		createMessage(
			'code-agreement',
			'code_mode',
			'CODE: "I\'ll handle implementation and validation. Let\'s coordinate."',
			['discussion', 'actionBlocks', 'finalResponse'],
			{ column: 'right' }
		),

		// Research Action Block
		createMessage(
			'research-action',
			'action_research',
			'ACTION_1: Research LSTM vs Transformer performance',
			['actionBlocks', 'finalResponse'],
			{ column: 'left' }
		),

		// Code Action Block
		createMessage(
			'code-action',
			'action_code',
			'ACTION_2: Build LSTM model with TensorFlow + metrics',
			['actionBlocks', 'finalResponse'],
			{ column: 'right' }
		),

		// Final coordinated response
		createMessage(
			'final-response',
			'coordinator',
			'RESPONSE: I\'ll research approaches then build an LSTM model.',
			['finalResponse'],
			{ column: 'left' }
		),
	],

	overlays: [
		{
			id: 'discussion-indicator',
			content: (state: InteractionState) => {
				if (state.activeStates.includes('discussion')) {
					return `
						<div style="
							background-color: rgba(142, 68, 173, 0.9);
							color: white;
							padding: 8px 20px;
							border-radius: 20px;
							font-size: 12px;
							font-weight: bold;
							box-shadow: 0 2px 10px rgba(142, 68, 173, 0.4);
						">
							🧠 INTERNAL DISCUSSION ACTIVE
						</div>
					`;
				} else if (state.activeStates.includes('actionBlocks')) {
					return `
						<div style="
							background-color: rgba(52, 73, 94, 0.9);
							color: white;
							padding: 8px 20px;
							border-radius: 20px;
							font-size: 12px;
							font-weight: bold;
							box-shadow: 0 2px 10px rgba(52, 73, 94, 0.4);
						">
							📋 GENERATING ACTION BLOCKS
						</div>
					`;
				}
				return '';
			},
			position: {
				top: '15%',
				right: '5%',
			},
			visibleStates: ['discussion', 'actionBlocks'],
		},
	],

	layout: {
		columns: 2,
		autoFill: true,
		maxMessagesPerColumn: 6,
	},

	tokenCounter: {
		enabled: true,
		initialTokens: 800,
		maxTokens: 128000,
		stateTokenCounts: {
			'userRequest': 900,
			'coordination': 1100,
			'discussion': 1800,
			'actionBlocks': 2200,
			'finalResponse': 2400,
		},
		optimizedStates: [],
	},
}; 