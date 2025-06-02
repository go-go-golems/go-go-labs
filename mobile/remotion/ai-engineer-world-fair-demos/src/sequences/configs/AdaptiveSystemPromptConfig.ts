import {
	InteractionSequence,
	createState,
	createMessage,
	DEFAULT_MESSAGE_TYPES,
	InteractionState,
} from '../../types/InteractionDSL';

export const adaptiveSystemPromptSequence: InteractionSequence = {
	title: 'Adaptive System Prompts: Dynamic Assistant Modes',
	
	subtitle: (state: InteractionState) => {
		if (state.activeStates.includes('researcherMode')) {
			return 'RESEARCHER MODE: Gathering and analyzing information';
		} else if (state.activeStates.includes('coderMode')) {
			return 'CODER MODE: Writing and debugging code';
		} else if (state.activeStates.includes('coachMode')) {
			return 'COACH MODE: Teaching and guiding learning';
		}
		return 'How LLMs adapt their behavior by changing system prompts';
	},

	messageTypes: {
		...DEFAULT_MESSAGE_TYPES,
		// Mode selection block
		mode_selection: {
			bg: '#8e44ad',
			icon: '🧠',
			label: 'Mode Selection',
			fontSize: '13px',
			padding: '14px 18px',
			boxShadow: '0 4px 15px rgba(142, 68, 173, 0.4)',
			border: '2px solid rgba(255, 255, 255, 0.3)',
			fontWeight: '500',
		},
		// System prompt display
		system_prompt: {
			bg: '#2c3e50',
			icon: '⚙️',
			label: 'System Prompt',
			fontSize: '11px',
			padding: '12px 15px',
			boxShadow: '0 3px 10px rgba(44, 62, 80, 0.3)',
			border: '1px solid rgba(255, 255, 255, 0.2)',
			fontStyle: 'italic',
		},
		// Mode-specific message types
		researcher_response: {
			bg: '#3498db',
			icon: '🔍',
			label: 'Researcher Assistant',
			fontSize: '13px',
			padding: '12px 15px',
			boxShadow: '0 2px 8px rgba(52, 152, 219, 0.3)',
		},
		coder_response: {
			bg: '#e67e22',
			icon: '💻',
			label: 'Coding Assistant',
			fontSize: '13px',
			padding: '12px 15px',
			boxShadow: '0 2px 8px rgba(230, 126, 34, 0.3)',
		},
		coach_response: {
			bg: '#27ae60',
			icon: '🎓',
			label: 'Learning Coach',
			fontSize: '13px',
			padding: '12px 15px',
			boxShadow: '0 2px 8px rgba(39, 174, 96, 0.3)',
		},
	},
	
	states: [
		createState('container', 0, 30),
		createState('userQuestion1', 30, 40),
		createState('modeSelection1', 70, 60),
		createState('systemPrompt1', 130, 40),
		createState('researcherMode', 170, 60),
		createState('userQuestion2', 230, 40),
		createState('modeSelection2', 270, 60),
		createState('systemPrompt2', 330, 40),
		createState('coderMode', 370, 60),
		createState('userQuestion3', 430, 40),
		createState('modeSelection3', 470, 60),
		createState('systemPrompt3', 530, 40),
		createState('coachMode', 570, 60),
	],

	messages: [
		// Persistent system prompt at the top that changes content
		createMessage(
			'persistent-system-prompt',
			'system_prompt',
			(state: InteractionState) => {
				if (state.activeStates.includes('researcherMode') || 
					state.activeStates.includes('systemPrompt1')) {
					return `SYSTEM PROMPT:

You are a research assistant specializing in cutting-edge technology analysis. Your role is to:
- Gather and synthesize current information from multiple sources
- Provide comprehensive overviews of complex technical topics
- Analyze implications and future trends`;
				} else if (state.activeStates.includes('coderMode') || 
						   state.activeStates.includes('systemPrompt2')) {
					return `SYSTEM PROMPT:

You are a senior software engineer specializing in cryptography and security. Your role is to:
- Write clean, efficient, and secure code
- Provide detailed implementation explanations
- Follow best practices and coding standards`;
				} else if (state.activeStates.includes('coachMode') || 
						   state.activeStates.includes('systemPrompt3')) {
					return `SYSTEM PROMPT:

You are an expert educator and learning coach specializing in complex technical concepts. Your role is to:
- Break down complex topics into digestible steps
- Use analogies and real-world examples
- Encourage questions and active learning`;
				}
				return `SYSTEM PROMPT:

You are a helpful AI assistant. Analyze the user's request and select the most appropriate mode to respond effectively.`;
			},
			['container', 'userQuestion1', 'modeSelection1', 'systemPrompt1', 'researcherMode', 'userQuestion2', 'modeSelection2', 'systemPrompt2', 'coderMode', 'userQuestion3', 'modeSelection3', 'systemPrompt3', 'coachMode'],
			{ column: 'left' }
		),

		// First interaction - Research question
		createMessage(
			'user-research-question',
			'user',
			'What are the latest developments in quantum computing and their potential impact on cryptography?',
			['userQuestion1', 'modeSelection1', 'systemPrompt1', 'researcherMode', 'userQuestion2', 'modeSelection2', 'systemPrompt2', 'coderMode', 'userQuestion3', 'modeSelection3', 'systemPrompt3', 'coachMode'],
			{ column: 'left' }
		),

		createMessage(
			'mode-selection-1',
			'mode_selection',
			(state: InteractionState) => {
				return `ASSISTANT_MODE: RESEARCHER

Chain of Thought:
- User asking about "latest developments" → research needed
- Complex technical domain → comprehensive analysis required
- Best served by research-focused approach

Selected Mode: RESEARCHER`;
			},
			['modeSelection1', 'systemPrompt1', 'researcherMode'],
			{ column: 'right' }
		),

		createMessage(
			'researcher-response',
			'researcher_response',
			'Based on recent research, quantum computing has made significant strides in 2024 with IBM\'s 1000+ qubit processors and Google\'s error correction breakthroughs. This progress makes RSA encryption increasingly vulnerable, with post-quantum cryptography standards emerging to address the threat within 10-15 years.',
			['researcherMode'],
			{ column: 'left' }
		),

		// Second interaction - Coding question
		createMessage(
			'user-coding-question',
			'user',
			'Can you implement a post-quantum cryptography algorithm in Python?',
			['userQuestion2', 'modeSelection2', 'systemPrompt2', 'coderMode', 'userQuestion3', 'modeSelection3', 'systemPrompt3', 'coachMode'],
			{ column: 'left' }
		),

		createMessage(
			'mode-selection-2',
			'mode_selection',
			(state: InteractionState) => {
				return `ASSISTANT_MODE: CODER

Chain of Thought:
- User wants implementation → coding expertise needed
- Specific request for Python code
- Technical implementation required
- Best served by development-focused approach

Selected Mode: CODER`;
			},
			['modeSelection2', 'systemPrompt2', 'coderMode'],
			{ column: 'right' }
		),

		createMessage(
			'coder-response',
			'coder_response',
			'Here\'s a basic Kyber implementation:\n\n```python\nclass KyberKEM:\n    def __init__(self, n=256, q=3329):\n        self.n, self.q = n, q\n    \n    def keygen(self):\n        s = np.random.randint(0, self.q, self.n)\n        a = np.random.randint(0, self.q, (self.n, self.n))\n        e = np.random.normal(0, 1, self.n).astype(int)\n        return (s, (a, (a @ s + e) % self.q))\n```',
			['coderMode'],
			{ column: 'left' }
		),

		// Third interaction - Learning question
		createMessage(
			'user-learning-question',
			'user',
			'I\'m struggling to understand how this algorithm works. Can you help me learn step by step?',
			['userQuestion3', 'modeSelection3', 'systemPrompt3', 'coachMode'],
			{ column: 'left' }
		),

		createMessage(
			'mode-selection-3',
			'mode_selection',
			(state: InteractionState) => {
				return `ASSISTANT_MODE: COACH

Chain of Thought:
- User expresses learning difficulty → educational support needed
- Requests "step by step" → pedagogical approach required
- Focus on understanding over implementation

Selected Mode: COACH`;
			},
			['modeSelection3', 'systemPrompt3', 'coachMode'],
			{ column: 'right' }
		),

		createMessage(
			'coach-response',
			'coach_response',
			'Let\'s break this down! 🎓 Current encryption is like a jigsaw puzzle - classical computers try pieces randomly, but quantum computers can "see" the solution. Post-quantum algorithms like Kyber use different math problems that even quantum computers find hard by adding "mathematical noise" that only the intended recipient can filter out. What part would you like me to explain further?',
			['coachMode'],
			{ column: 'left' }
		),
	],

	overlays: [
		{
			id: 'mode-indicator',
			content: (state: InteractionState) => {
				if (state.activeStates.includes('researcherMode')) {
					return `
						<div style="
							background-color: #3498db;
							color: white;
							padding: 8px 20px;
							border-radius: 20px;
							font-size: 12px;
							font-weight: bold;
							box-shadow: 0 2px 10px rgba(52, 152, 219, 0.4);
						">
							🔍 RESEARCHER MODE ACTIVE
						</div>
					`;
				} else if (state.activeStates.includes('coderMode')) {
					return `
						<div style="
							background-color: #e67e22;
							color: white;
							padding: 8px 20px;
							border-radius: 20px;
							font-size: 12px;
							font-weight: bold;
							box-shadow: 0 2px 10px rgba(230, 126, 34, 0.4);
						">
							💻 CODER MODE ACTIVE
						</div>
					`;
				} else if (state.activeStates.includes('coachMode')) {
					return `
						<div style="
							background-color: #27ae60;
							color: white;
							padding: 8px 20px;
							border-radius: 20px;
							font-size: 12px;
							font-weight: bold;
							box-shadow: 0 2px 10px rgba(39, 174, 96, 0.4);
						">
							🎓 COACH MODE ACTIVE
						</div>
					`;
				}
				return '';
			},
			position: {
				top: '15%',
				right: '5%',
			},
			visibleStates: ['researcherMode', 'coderMode', 'coachMode'],
		},
	],

	layout: {
		columns: 2,
		autoFill: true,
		maxMessagesPerColumn: 6,
	},

	tokenCounter: {
		enabled: true,
		initialTokens: 600,
		maxTokens: 128000,
		stateTokenCounts: {
			'userQuestion1': 700,
			'modeSelection1': 1000,
			'systemPrompt1': 1100,
			'researcherMode': 1600,
			'userQuestion2': 1700,
			'modeSelection2': 2000,
			'systemPrompt2': 2100,
			'coderMode': 2800,
			'userQuestion3': 2900,
			'modeSelection3': 3200,
			'systemPrompt3': 3300,
			'coachMode': 3900,
		},
		optimizedStates: [],
	},
}; 