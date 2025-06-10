import React from 'react';
import {
	InteractionSequence,
	createState,
	createMessage,
	DEFAULT_MESSAGE_TYPES,
	InteractionState,
	FONT_SIZES,
} from '../../types/InteractionDSL';

// Interactive UI Components
const ToolCallApproval: React.FC<{ 
	toolName: string; 
	params: string; 
	isEditing: boolean;
	editedParams?: string;
}> = ({ toolName, params, isEditing, editedParams }) => (
	<div style={{
		backgroundColor: 'rgba(255, 255, 255, 0.1)',
		borderRadius: '8px',
		padding: '12px',
		border: '1px solid rgba(255, 255, 255, 0.2)',
		fontSize: '11px',
	}}>
		<div style={{ marginBottom: '8px', fontWeight: 'bold' }}>
			🔧 Tool Call: {toolName}
		</div>
		{isEditing ? (
			<div style={{ marginBottom: '8px' }}>
				<div style={{ fontSize: '9px', marginBottom: '4px', color: '#f39c12' }}>
					✏️ EDITING MODE
				</div>
				<textarea
					style={{
						width: '100%',
						height: '40px',
						backgroundColor: 'rgba(0, 0, 0, 0.3)',
						color: 'white',
						border: '1px solid #f39c12',
						borderRadius: '4px',
						padding: '4px',
						fontSize: '10px',
						fontFamily: 'monospace',
						resize: 'none'
					}}
					value={editedParams || params}
					readOnly
				/>
			</div>
		) : (
			<div style={{ marginBottom: '8px', fontFamily: 'monospace', fontSize: '10px' }}>
				{params}
			</div>
		)}
		<div style={{ display: 'flex', gap: '8px' }}>
			{isEditing ? (
				<>
					<button style={{
						backgroundColor: '#27ae60',
						color: 'white',
						border: 'none',
						borderRadius: '4px',
						padding: '4px 8px',
						fontSize: '10px',
						cursor: 'pointer'
					}}>
						✓ Save & Approve
					</button>
					<button style={{
						backgroundColor: '#95a5a6',
						color: 'white',
						border: 'none',
						borderRadius: '4px',
						padding: '4px 8px',
						fontSize: '10px',
						cursor: 'pointer'
					}}>
						↶ Cancel
					</button>
				</>
			) : (
				<>
					<button style={{
						backgroundColor: '#27ae60',
						color: 'white',
						border: 'none',
						borderRadius: '4px',
						padding: '4px 8px',
						fontSize: '10px',
						cursor: 'pointer'
					}}>
						✓ Approve
					</button>
					<button style={{
						backgroundColor: '#e74c3c',
						color: 'white',
						border: 'none',
						borderRadius: '4px',
						padding: '4px 8px',
						fontSize: '10px',
						cursor: 'pointer'
					}}>
						✗ Reject
					</button>
					<button style={{
						backgroundColor: '#f39c12',
						color: 'white',
						border: 'none',
						borderRadius: '4px',
						padding: '4px 8px',
						fontSize: '10px',
						cursor: 'pointer'
					}}>
						✏️ Edit
					</button>
				</>
			)}
		</div>
	</div>
);

const ResultValidation: React.FC<{ 
	result: string; 
	isFiltering: boolean;
	filteredResult?: string;
}> = ({ result, isFiltering, filteredResult }) => (
	<div style={{
		backgroundColor: 'rgba(255, 255, 255, 0.1)',
		borderRadius: '8px',
		padding: '12px',
		border: '1px solid rgba(255, 255, 255, 0.2)',
		fontSize: '11px',
	}}>
		<div style={{ marginBottom: '8px', fontWeight: 'bold' }}>
			📊 Tool Result Validation
		</div>
		{isFiltering && (
			<div style={{ fontSize: '9px', marginBottom: '4px', color: '#f39c12' }}>
				🔍 FILTERING SENSITIVE DATA
			</div>
		)}
		<div style={{ 
			marginBottom: '8px', 
			fontFamily: 'monospace', 
			fontSize: '10px',
			maxHeight: '80px',
			overflow: 'auto',
			backgroundColor: 'rgba(0, 0, 0, 0.2)',
			padding: '6px',
			borderRadius: '4px',
			border: isFiltering ? '1px solid #f39c12' : 'none'
		}}>
			{isFiltering ? filteredResult : result}
		</div>
		<div style={{ display: 'flex', gap: '8px' }}>
			{isFiltering ? (
				<>
					<button style={{
						backgroundColor: '#27ae60',
						color: 'white',
						border: 'none',
						borderRadius: '4px',
						padding: '4px 8px',
						fontSize: '10px',
						cursor: 'pointer'
					}}>
						✓ Pass Filtered
					</button>
					<button style={{
						backgroundColor: '#95a5a6',
						color: 'white',
						border: 'none',
						borderRadius: '4px',
						padding: '4px 8px',
						fontSize: '10px',
						cursor: 'pointer'
					}}>
						↶ Cancel
					</button>
				</>
			) : (
				<>
					<button style={{
						backgroundColor: '#27ae60',
						color: 'white',
						border: 'none',
						borderRadius: '4px',
						padding: '4px 8px',
						fontSize: '10px',
						cursor: 'pointer'
					}}>
						✓ Pass to LLM
					</button>
					<button style={{
						backgroundColor: '#e74c3c',
						color: 'white',
						border: 'none',
						borderRadius: '4px',
						padding: '4px 8px',
						fontSize: '10px',
						cursor: 'pointer'
					}}>
						✗ Block
					</button>
					<button style={{
						backgroundColor: '#f39c12',
						color: 'white',
						border: 'none',
						borderRadius: '4px',
						padding: '4px 8px',
						fontSize: '10px',
						cursor: 'pointer'
					}}>
						🔍 Filter
					</button>
				</>
			)}
		</div>
	</div>
);

export const userControlledToolsSequence: InteractionSequence = {
	title: 'User-Controlled Tool Execution',
	
	subtitle: (state: InteractionState) => {
		if (state.activeStates.includes('toolApproval')) {
			return 'User validates tool call before execution';
		} else if (state.activeStates.includes('resultValidation')) {
			return 'User validates result before passing to LLM';
		}
		return 'Human-in-the-loop tool calling with approval gates';
	},

	messageTypes: {
		...DEFAULT_MESSAGE_TYPES,
		// User control UI
		user_control: {
			bg: '#34495e',
			icon: '🎛️',
			label: 'User Control',
			fontSize: '12px',
			padding: '8px 12px',
			boxShadow: '0 3px 10px rgba(52, 73, 94, 0.3)',
			border: '2px solid rgba(255, 255, 255, 0.3)',
		},
		// Status indicators
		approved: {
			bg: '#27ae60',
			icon: '✅',
			label: 'Approved',
			fontSize: '11px',
			padding: '8px 12px',
			boxShadow: '0 2px 8px rgba(39, 174, 96, 0.3)',
		},
		blocked: {
			bg: '#e74c3c',
			icon: '🚫',
			label: 'Blocked',
			fontSize: '11px',
			padding: '8px 12px',
			boxShadow: '0 2px 8px rgba(231, 76, 60, 0.3)',
		},
	},
	
	states: [
		createState('container', 0, 30),
		createState('userRequest', 30, 40),
		createState('llmPlanning', 70, 50),
		createState('toolCall', 120, 40),
		createState('toolApproval', 160, 40),
		createState('editing', 200, 60),
		createState('toolExecution', 260, 40),
		createState('toolResult', 300, 40),
		createState('resultValidation', 340, 40),
		createState('filtering', 380, 60),
		createState('llmResponse', 440, 40),
	],

	messages: [
		// User request
		createMessage(
			'user-question',
			'user',
			'Find all customers with overdue invoices over $5000',
			['container', 'userRequest', 'llmPlanning', 'toolCall', 'toolApproval', 'editing', 'toolExecution', 'toolResult', 'resultValidation', 'filtering', 'llmResponse'],
			{ column: 'left' }
		),

		// LLM planning
		createMessage(
			'llm-planning',
			'assistant',
			'I\'ll query the CRM database for customers with overdue invoices.',
			['llmPlanning', 'toolCall', 'toolApproval', 'editing', 'toolExecution', 'toolResult', 'resultValidation', 'filtering', 'llmResponse'],
			{ column: 'left' }
		),

		// Tool call
		createMessage(
			'tool-call',
			'tool_use',
			'query_crm(table="customers", filter="invoice_amount > 5000 AND status = \'overdue\'")',
			['toolCall', 'toolApproval', 'editing', 'toolExecution', 'toolResult', 'resultValidation', 'filtering', 'llmResponse'],
			{ column: 'left' }
		),

		// User approval UI (initial)
		createMessage(
			'tool-approval-ui',
			'user_control',
			(state: InteractionState) => (
				<ToolCallApproval 
					toolName="query_crm" 
					params={'table="customers", filter="invoice_amount > 5000 AND status = \'overdue\'"'}
					isEditing={state.activeStates.includes('editing')}
					editedParams={'table="customers", filter="invoice_amount > 10000 AND status = \'overdue\' AND days_overdue > 30"'}
				/>
			),
			['toolApproval', 'editing', 'toolExecution', 'toolResult', 'resultValidation', 'filtering', 'llmResponse'],
			{ column: 'right', isReactContent: true }
		),

		// Edit status
		createMessage(
			'edit-status',
			'edit_indicator',
			'User modified query: increased threshold to $10k, added 30+ days overdue',
			['editing', 'toolExecution', 'toolResult', 'resultValidation', 'filtering', 'llmResponse'],
			{ column: 'right' }
		),

		// Approval status
		createMessage(
			'approval-status',
			'approved',
			'Modified tool call approved',
			['toolExecution', 'toolResult', 'resultValidation', 'filtering', 'llmResponse'],
			{ column: 'right' }
		),

		// Tool result
		createMessage(
			'tool-result',
			'tool_result',
			'Found 2 customers:\n• Acme Corp: $12,500 (45 days overdue)\n• TechStart Inc: $18,750 (62 days overdue)\n\nIncludes: names, emails, phones, SSNs',
			['toolResult', 'resultValidation', 'filtering', 'llmResponse'],
			{ column: 'left' }
		),

		// Result validation UI
		createMessage(
			'result-validation-ui',
			'user_control',
			(state: InteractionState) => (
				<ResultValidation 
					result={'Found 2 customers:\n• Acme Corp: $12,500 (45 days overdue)\n• TechStart Inc: $18,750 (62 days overdue)\n\nIncludes: names, emails, phones, SSNs'}
					isFiltering={state.activeStates.includes('filtering')}
					filteredResult={'Found 2 customers:\n• Acme Corp: $12,500 (45 days overdue)\n• TechStart Inc: $18,750 (62 days overdue)\n\nFiltered: phones and SSNs removed'}
				/>
			),
			['resultValidation', 'filtering', 'llmResponse'],
			{ column: 'right', isReactContent: true }
		),

		// Filter status
		createMessage(
			'filter-status',
			'edit_indicator',
			'User filtered out phone numbers and SSNs for privacy',
			['filtering', 'llmResponse'],
			{ column: 'right' }
		),

		// Validation status
		createMessage(
			'validation-status',
			'approved',
			'Filtered result approved - passing to LLM',
			['llmResponse'],
			{ column: 'right' }
		),

		// Final LLM response
		createMessage(
			'llm-response',
			'assistant',
			'Found 2 customers with overdue invoices over $10k: Acme Corp ($12,500, 45 days) and TechStart Inc ($18,750, 62 days).',
			['llmResponse'],
			{ column: 'left' }
		),
	],

	overlays: [
		{
			id: 'control-indicator',
			content: (state: InteractionState) => {
				if (state.activeStates.includes('toolApproval')) {
					return `
						<div style="
							background-color: rgba(241, 196, 15, 0.9);
							color: black;
							padding: 8px 20px;
							border-radius: 20px;
							font-size: 12px;
							font-weight: bold;
							box-shadow: 0 2px 10px rgba(241, 196, 15, 0.4);
						">
							⏸️ AWAITING TOOL APPROVAL
						</div>
					`;
				} else if (state.activeStates.includes('editing')) {
					return `
						<div style="
							background-color: rgba(243, 156, 18, 0.9);
							color: white;
							padding: 8px 20px;
							border-radius: 20px;
							font-size: 12px;
							font-weight: bold;
							box-shadow: 0 2px 10px rgba(243, 156, 18, 0.4);
						">
							✏️ USER EDITING QUERY
						</div>
					`;
				} else if (state.activeStates.includes('resultValidation')) {
					return `
						<div style="
							background-color: rgba(52, 152, 219, 0.9);
							color: white;
							padding: 8px 20px;
							border-radius: 20px;
							font-size: 12px;
							font-weight: bold;
							box-shadow: 0 2px 10px rgba(52, 152, 219, 0.4);
						">
							🔍 VALIDATING RESULT
						</div>
					`;
				} else if (state.activeStates.includes('filtering')) {
					return `
						<div style="
							background-color: rgba(155, 89, 182, 0.9);
							color: white;
							padding: 8px 20px;
							border-radius: 20px;
							font-size: 12px;
							font-weight: bold;
							box-shadow: 0 2px 10px rgba(155, 89, 182, 0.4);
						">
							🔒 FILTERING SENSITIVE DATA
						</div>
					`;
				}
				return '';
			},
			position: {
				top: '15%',
				right: '5%',
			},
			visibleStates: ['toolApproval', 'editing', 'resultValidation', 'filtering'],
		},
	],

	layout: {
		columns: 2,
		autoFill: false,
		maxMessagesPerColumn: 6,
	},

	tokenCounter: {
		enabled: true,
		initialTokens: 500,
		maxTokens: 128000,
		stateTokenCounts: {
			'userRequest': 600,
			'llmPlanning': 750,
			'toolCall': 800,
			'toolApproval': 800, // No tokens consumed during approval
			'editing': 800, // No tokens consumed during editing
			'toolExecution': 850,
			'toolResult': 1200, // More data returned
			'resultValidation': 1200, // No tokens consumed during validation
			'filtering': 1200, // No tokens consumed during filtering
			'llmResponse': 1350,
		},
		optimizedStates: [],
	},
}; 