import {
	InteractionSequence,
	createState,
	createMessage,
	DEFAULT_MESSAGE_TYPES,
} from '../../types/InteractionDSL';

export const contextSummarizationSequence: InteractionSequence = {
	title: 'Step 3: Long Coding Session → Summary',
	messageTypes: DEFAULT_MESSAGE_TYPES,
	
	states: [
		createState('container', 0, 30),
		createState('prevConversation', 30, 10),
		createState('userRequest', 40, 10),
		createState('cot1', 60, 10),
		createState('diary1', 80, 10),
		createState('toolUse', 100, 10),
		createState('toolResult', 120, 10),
		createState('cot2', 140, 10),
		createState('diary2', 160, 10),
		createState('response', 180, 10),
		createState('fadeOut', 220, 30),
		createState('summary', 270, 30),
		createState('cleanUp', 300, 30),
	],

	messages: [
		createMessage(
			'debug-request',
			'user',
			'Help me debug this Python sorting function: def sort_list(arr): arr.sort() return arr[0]',
			['container', 'userRequest', 'cot1', 'diary1', 'toolUse', 'toolResult', 'cot2', 'diary2', 'response', 'summary'],
			{ fadeOutStates: ['cleanUp'] }
		),

		createMessage(
			'cot-analysis',
			'assistant_cot',
			"User has a Python function issue. They're calling sort() but returning arr[0]. This returns just the minimum element, not the sorted list.",
			['cot1', 'diary1', 'toolUse', 'toolResult', 'cot2', 'diary2', 'response'],
			{ fadeOutStates: ['fadeOut', 'cleanUp'] }
		),

		createMessage(
			'diary-task',
			'assistant_diary',
			'Task: Debug Python sort function. Issue identified: returns single element instead of full sorted array.',
			['diary1', 'toolUse', 'toolResult', 'cot2', 'diary2', 'response'],
			{ fadeOutStates: ['fadeOut', 'cleanUp'] }
		),

		createMessage(
			'tool-execution',
			'tool_use',
			"run_python_code('def sort_list(arr): arr.sort(); return arr[0]; print(sort_list([3,1,4]))')",
			['toolUse', 'toolResult', 'cot2', 'diary2', 'response'],
			{ fadeOutStates: ['fadeOut', 'cleanUp'] }
		),

		createMessage(
			'tool-output',
			'tool_result',
			'Output: 1\nFunction returns minimum value, not sorted list [1, 3, 4]',
			['toolResult', 'cot2', 'diary2', 'response'],
			{ fadeOutStates: ['fadeOut', 'cleanUp'] }
		),

		createMessage(
			'cot-solution',
			'assistant_cot',
			"Confirmed the bug. Need to explain that they should return 'arr' not 'arr[0]'. Also suggest sorted() as alternative to avoid mutation.",
			['cot2', 'diary2', 'response'],
			{ fadeOutStates: ['fadeOut', 'cleanUp'] }
		),

		createMessage(
			'diary-solution',
			'assistant_diary',
			'Solution: Change return arr[0] to return arr. Also mention sorted() function as non-mutating alternative.',
			['diary2', 'response'],
			{ fadeOutStates: ['fadeOut', 'cleanUp'] }
		),

		createMessage(
			'final-response',
			'assistant',
			"I found the issue! Your function returns arr[0] (just the minimum) instead of the sorted list. Fix: return arr",
			['response'],
			{ fadeOutStates: ['fadeOut', 'cleanUp'] }
		),

		createMessage(
			'summary-block',
			'summary',
			'Debugging Session: Fixed Python sort function - changed \'return arr[0]\' to \'return arr\' to return full sorted list instead of minimum element.',
			['summary', 'cleanUp']
		),
	],

	overlays: [
		{
			id: 'summarization-explanation',
			content: `
				<div style="
					background-color: #6c3483;
					color: white;
					padding: 12px 25px;
					border-radius: 10px;
					font-size: 14px;
					text-align: center;
					box-shadow: 0 4px 15px rgba(108, 52, 131, 0.4);
					font-family: Arial, sans-serif;
				">
					<div style="margin-bottom: 5px; font-weight: bold;">
						✨ 7 message blocks condensed into 1 summary
					</div>
					<div style="font-size: 12px; opacity: 0.9;">
						Full COT + diary + tool sequence → single summary block
					</div>
				</div>
			`,
			position: {
				bottom: '25%',
				left: '50%',
			},
			style: {
				transform: 'translateX(-50%)',
			},
			visibleStates: ['summary', 'cleanUp'],
		},
	],

	layout: {
		columns: 2,
		autoFill: true,
		maxMessagesPerColumn: 5,
	},

	tokenCounter: {
		enabled: true,
		initialTokens: 3200,
		maxTokens: 128000,
		stateTokenCounts: {
			'prevConversation': 3200,
			'userRequest': 3800,
			'cot1': 4400,
			'diary1': 5000,
			'toolUse': 5600,
			'toolResult': 6800,
			'cot2': 7600,
			'diary2': 8400,
			'response': 9200,
			'summary': 4800,
		},
		optimizedStates: ['summary'],
	},
}; 