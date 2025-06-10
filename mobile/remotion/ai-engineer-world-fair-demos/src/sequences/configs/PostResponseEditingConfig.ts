import {
	InteractionSequence,
	createState,
	createMessage,
	DEFAULT_MESSAGE_TYPES,
	InteractionState,
	FONT_SIZES,
} from '../../types/InteractionDSL';

export const postResponseEditingSequence: InteractionSequence = {
	title: (state: InteractionState) => {
		if (state.activeStates.includes('editingResponse')) {
			return 'Post-Response Editing: Fixing Code Issues';
		}
		return 'Post-Response Editing: Code → Test → Fix → Success';
	},
	
	subtitle: (state: InteractionState) => {
		if (state.activeStates.includes('editingResponse')) {
			return 'User edits AI response to fix failing tests';
		}
		return 'How users can edit AI responses when tests reveal issues';
	},

	messageTypes: {
		...DEFAULT_MESSAGE_TYPES,
		// Enhanced message types with dynamic content
		assistant_editing: {
			...DEFAULT_MESSAGE_TYPES.assistant,
			icon: (state: InteractionState) => 
				state.activeStates.includes('editingResponse') ? '✏️' : '🧠',
			label: (state: InteractionState) => 
				state.activeStates.includes('editingResponse') ? 'Assistant (Editing)' : 'Assistant',
		},
		tool_use_failed: {
			...DEFAULT_MESSAGE_TYPES.tool_use,
			bg: '#e74c3c',
			icon: '❌',
			label: 'Tool Use (Failed)',
		},
		tool_result_failed: {
			...DEFAULT_MESSAGE_TYPES.tool_result,
			bg: '#c0392b',
			icon: '💥',
			label: 'Test Failed',
		},
	},
	
	states: [
		createState('container', 0, 30),
		createState('userRequest', 30, 40),
		createState('aiResponse', 70, 50),
		createState('firstToolCall', 120, 40),
		createState('testFailure', 160, 50),
		createState('editingResponse', 210, 60),
		createState('toolsDisappear', 230, 20),
		createState('secondToolCall', 270, 40),
		createState('testSuccess', 310, 50),
		createState('nextUserMessage', 360, 40),
	],

	messages: [
		createMessage(
			'user-code-request',
			'user',
			'Write a function to calculate the factorial of a number',
			['container', 'userRequest', 'aiResponse', 'firstToolCall', 'testFailure', 'editingResponse', 'secondToolCall', 'testSuccess', 'nextUserMessage'],
			{ column: 'left' }
		),

		createMessage(
			'ai-code-response',
			'assistant_editing',
			(state: InteractionState) => {
				if (state.activeStates.includes('editingResponse') || 
					state.activeStates.includes('secondToolCall') || 
					state.activeStates.includes('testSuccess') ||
					state.activeStates.includes('nextUserMessage')) {
					return 'def factorial(n):\n    if n < 0:\n        raise ValueError("Negative numbers not allowed")\n    if n <= 1:\n        return 1\n    return n * factorial(n - 1)';
				}
				return 'def factorial(n):\n    if n <= 1:\n        return 1\n    return n * factorial(n - 1)';
			},
			['aiResponse', 'firstToolCall', 'testFailure', 'editingResponse', 'secondToolCall', 'testSuccess', 'nextUserMessage'],
			{ column: 'left' }
		),

		createMessage(
			'first-tool-call',
			'tool_use_failed',
			'run_tests("test_factorial.py")',
			['firstToolCall', 'testFailure'],
			{ 
				column: 'right',
				fadeOutStates: ['toolsDisappear']
			}
		),

		createMessage(
			'first-test-result',
			'tool_result_failed',
			'FAILED: factorial(-1) should raise ValueError\nExpected exception but got: 1',
			['testFailure'],
			{ 
				column: 'right',
				fadeOutStates: ['toolsDisappear']
			}
		),

		createMessage(
			'second-tool-call',
			'tool_use',
			'run_tests("test_factorial.py")',
			['secondToolCall', 'testSuccess', 'nextUserMessage'],
			{ column: 'right' }
		),

		createMessage(
			'second-test-result',
			'tool_result',
			'PASSED: All tests passed!\n✓ factorial(5) = 120\n✓ factorial(0) = 1\n✓ factorial(-1) raises ValueError',
			['testSuccess', 'nextUserMessage'],
			{ column: 'right' }
		),

		createMessage(
			'next-user-message',
			'user',
			'Great! Now can you add type hints to the function?',
			['nextUserMessage'],
			{ column: 'left' }
		),
	],

	overlays: [
		{
			id: 'editing-explanation',
			content: (state: InteractionState) => {
				if (state.activeStates.includes('testFailure')) {
					return `
						<div style="
							background-color: #e74c3c;
							color: white;
							padding: 12px 25px;
							border-radius: 10px;
							font-size: 14px;
							text-align: center;
							box-shadow: 0 4px 15px rgba(231, 76, 60, 0.4);
							font-family: Arial, sans-serif;
						">
							<div style="margin-bottom: 5px; font-weight: bold;">
								❌ Tests Failed
							</div>
							<div style="font-size: 12px; opacity: 0.9;">
								Code doesn't handle negative numbers properly
							</div>
						</div>
					`;
				} else if (state.activeStates.includes('editingResponse')) {
					return `
						<div style="
							background-color: #f39c12;
							color: white;
							padding: 12px 25px;
							border-radius: 10px;
							font-size: 14px;
							text-align: center;
							box-shadow: 0 4px 15px rgba(243, 156, 18, 0.4);
							font-family: Arial, sans-serif;
						">
							<div style="margin-bottom: 5px; font-weight: bold;">
								✏️ User Editing AI Response
							</div>
							<div style="font-size: 12px; opacity: 0.9;">
								Adding error handling for negative numbers
							</div>
						</div>
					`;
				} else if (state.activeStates.includes('toolsDisappear')) {
					return `
						<div style="
							background-color: #95a5a6;
							color: white;
							padding: 12px 25px;
							border-radius: 10px;
							font-size: 14px;
							text-align: center;
							box-shadow: 0 4px 15px rgba(149, 165, 166, 0.4);
							font-family: Arial, sans-serif;
						">
							<div style="margin-bottom: 5px; font-weight: bold;">
								🔄 Tool Calls Cleared
							</div>
							<div style="font-size: 12px; opacity: 0.9;">
								Previous test results invalidated by code changes
							</div>
						</div>
					`;
				} else if (state.activeStates.includes('testSuccess')) {
					return `
						<div style="
							background-color: #27ae60;
							color: white;
							padding: 12px 25px;
							border-radius: 10px;
							font-size: 14px;
							text-align: center;
							box-shadow: 0 4px 15px rgba(39, 174, 96, 0.4);
							font-family: Arial, sans-serif;
						">
							<div style="margin-bottom: 5px; font-weight: bold;">
								✅ Tests Passed
							</div>
							<div style="font-size: 12px; opacity: 0.9;">
								Fixed code now handles all test cases correctly
							</div>
						</div>
					`;
				}
				return `
					<div style="
						background-color: #34495e;
						color: white;
						padding: 12px 25px;
						border-radius: 10px;
						font-size: 14px;
						text-align: center;
						box-shadow: 0 4px 15px rgba(52, 73, 94, 0.4);
						font-family: Arial, sans-serif;
					">
						<div style="margin-bottom: 5px; font-weight: bold;">
							🔄 Edit-Test-Fix Workflow
						</div>
						<div style="font-size: 12px; opacity: 0.9;">
							Iterative improvement through testing and editing
						</div>
					</div>
				`;
			},
			position: {
				bottom: '20%',
				left: '50%',
			},
			style: {
				transform: 'translateX(-50%)',
			},
			visibleStates: ['firstToolCall', 'testFailure', 'editingResponse', 'toolsDisappear', 'secondToolCall', 'testSuccess'],
		},
	],

	layout: {
		columns: 2,
		autoFill: false,
		maxMessagesPerColumn: 6,
	},

	tokenCounter: {
		enabled: true,
		initialTokens: 1200,
		maxTokens: 128000,
		stateTokenCounts: {
			'userRequest': 1200,
			'aiResponse': 1800,
			'firstToolCall': 2100,
			'testFailure': 2400,
			'editingResponse': 2600,
			'secondToolCall': 2900,
			'testSuccess': 3200,
			'nextUserMessage': 3400,
		},
		optimizedStates: [],
	},
}; 