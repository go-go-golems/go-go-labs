import React, { useState } from 'react';
import {
	InteractionSequence,
	createState,
	createMessage,
	DEFAULT_MESSAGE_TYPES,
	InteractionState,
} from '../../types/InteractionDSL';

// Dynamic UI Components based on LLM-generated DSL
const DynamicToolUI: React.FC<{ 
	uiDsl: any;
	toolCall: string;
	result?: string;
	isExecuted: boolean;
}> = ({ uiDsl, toolCall, result, isExecuted }) => {
	const [formData, setFormData] = useState(uiDsl.defaultValues || {});
	const [showAdvanced, setShowAdvanced] = useState(false);

	const handleInputChange = (field: string, value: any) => {
		setFormData(prev => ({ ...prev, [field]: value }));
	};

	return (
		<div style={{
			backgroundColor: 'rgba(255, 255, 255, 0.1)',
			borderRadius: '12px',
			padding: '16px',
			border: '2px solid rgba(74, 144, 226, 0.3)',
			fontSize: '11px',
			maxWidth: '400px',
		}}>
			{/* Header */}
			<div style={{ 
				marginBottom: '12px', 
				fontWeight: 'bold',
				color: '#4a90e2',
				display: 'flex',
				alignItems: 'center',
				gap: '8px'
			}}>
				🎨 {uiDsl.title}
				<div style={{
					fontSize: '9px',
					backgroundColor: 'rgba(74, 144, 226, 0.2)',
					padding: '2px 6px',
					borderRadius: '10px',
					color: '#4a90e2'
				}}>
					LLM-Generated UI
				</div>
			</div>

			{/* Description */}
			<div style={{ 
				marginBottom: '12px', 
				fontSize: '10px',
				color: '#bdc3c7',
				fontStyle: 'italic'
			}}>
				{uiDsl.description}
			</div>

			{/* Form Fields */}
			<div style={{ marginBottom: '12px' }}>
				{uiDsl.fields.map((field: any, idx: number) => (
					<div key={idx} style={{ marginBottom: '8px' }}>
						<label style={{ 
							display: 'block', 
							marginBottom: '4px',
							fontSize: '10px',
							fontWeight: 'bold'
						}}>
							{field.label}
							{field.required && <span style={{ color: '#e74c3c' }}>*</span>}
						</label>
						
						{field.type === 'select' ? (
							<select
								style={{
									width: '100%',
									padding: '4px',
									backgroundColor: 'rgba(0, 0, 0, 0.3)',
									color: 'white',
									border: '1px solid rgba(255, 255, 255, 0.2)',
									borderRadius: '4px',
									fontSize: '10px'
								}}
								value={formData[field.name] || field.defaultValue}
								onChange={(e) => handleInputChange(field.name, e.target.value)}
							>
								{field.options.map((opt: any, optIdx: number) => (
									<option key={optIdx} value={opt.value}>{opt.label}</option>
								))}
							</select>
						) : field.type === 'range' ? (
							<div>
								<input
									type="range"
									min={field.min}
									max={field.max}
									step={field.step}
									value={formData[field.name] || field.defaultValue}
									onChange={(e) => handleInputChange(field.name, parseInt(e.target.value))}
									style={{ width: '100%' }}
								/>
								<div style={{ fontSize: '9px', color: '#95a5a6', textAlign: 'center' }}>
									{formData[field.name] || field.defaultValue} {field.unit}
								</div>
							</div>
						) : field.type === 'checkbox' ? (
							<label style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
								<input
									type="checkbox"
									checked={formData[field.name] || field.defaultValue}
									onChange={(e) => handleInputChange(field.name, e.target.checked)}
								/>
								<span style={{ fontSize: '9px' }}>{field.description}</span>
							</label>
						) : (
							<input
								type={field.type}
								placeholder={field.placeholder}
								value={formData[field.name] || field.defaultValue || ''}
								onChange={(e) => handleInputChange(field.name, e.target.value)}
								style={{
									width: '100%',
									padding: '4px',
									backgroundColor: 'rgba(0, 0, 0, 0.3)',
									color: 'white',
									border: '1px solid rgba(255, 255, 255, 0.2)',
									borderRadius: '4px',
									fontSize: '10px'
								}}
							/>
						)}
						
						{field.help && (
							<div style={{ fontSize: '8px', color: '#95a5a6', marginTop: '2px' }}>
								💡 {field.help}
							</div>
						)}
					</div>
				))}
			</div>

			{/* Advanced Options Toggle */}
			{uiDsl.advancedFields && (
				<div style={{ marginBottom: '12px' }}>
					<button
						onClick={() => setShowAdvanced(!showAdvanced)}
						style={{
							backgroundColor: 'transparent',
							color: '#4a90e2',
							border: '1px solid #4a90e2',
							borderRadius: '4px',
							padding: '4px 8px',
							fontSize: '9px',
							cursor: 'pointer',
							width: '100%'
						}}
					>
						{showAdvanced ? '▼' : '▶'} Advanced Options
					</button>
					
					{showAdvanced && (
						<div style={{ marginTop: '8px', paddingLeft: '8px', borderLeft: '2px solid #4a90e2' }}>
							{uiDsl.advancedFields.map((field: any, idx: number) => (
								<div key={idx} style={{ marginBottom: '6px' }}>
									<label style={{ fontSize: '9px', fontWeight: 'bold' }}>
										{field.label}
									</label>
									<input
										type={field.type}
										placeholder={field.placeholder}
										value={formData[field.name] || field.defaultValue || ''}
										onChange={(e) => handleInputChange(field.name, e.target.value)}
										style={{
											width: '100%',
											padding: '3px',
											backgroundColor: 'rgba(0, 0, 0, 0.2)',
											color: 'white',
											border: '1px solid rgba(74, 144, 226, 0.3)',
											borderRadius: '3px',
											fontSize: '9px'
										}}
									/>
								</div>
							))}
						</div>
					)}
				</div>
			)}

			{/* Generated Tool Call Preview */}
			<div style={{ 
				marginBottom: '12px',
				backgroundColor: 'rgba(0, 0, 0, 0.3)',
				padding: '8px',
				borderRadius: '4px',
				border: '1px solid rgba(255, 255, 255, 0.1)'
			}}>
				<div style={{ fontSize: '9px', marginBottom: '4px', color: '#f39c12' }}>
					🔧 Generated Tool Call:
				</div>
				<div style={{ 
					fontFamily: 'monospace', 
					fontSize: '9px',
					color: '#2ecc71',
					wordBreak: 'break-all'
				}}>
					{uiDsl.generateToolCall ? uiDsl.generateToolCall(formData) : toolCall}
				</div>
			</div>

			{/* Result Display */}
			{isExecuted && result && (
				<div style={{ 
					marginBottom: '12px',
					backgroundColor: 'rgba(46, 204, 113, 0.1)',
					padding: '8px',
					borderRadius: '4px',
					border: '1px solid rgba(46, 204, 113, 0.3)'
				}}>
					<div style={{ fontSize: '9px', marginBottom: '4px', color: '#2ecc71' }}>
						📊 Results:
					</div>
					<div style={{ fontSize: '10px', fontFamily: 'monospace' }}>
						{result}
					</div>
					
					{/* Result Actions */}
					<div style={{ marginTop: '8px', display: 'flex', gap: '4px', flexWrap: 'wrap' }}>
						{uiDsl.resultActions?.map((action: any, idx: number) => (
							<button
								key={idx}
								style={{
									backgroundColor: action.color || '#3498db',
									color: 'white',
									border: 'none',
									borderRadius: '3px',
									padding: '3px 6px',
									fontSize: '8px',
									cursor: 'pointer'
								}}
							>
								{action.icon} {action.label}
							</button>
						))}
					</div>
				</div>
			)}

			{/* Action Buttons */}
			<div style={{ display: 'flex', gap: '6px', flexWrap: 'wrap' }}>
				{!isExecuted ? (
					<>
						<button style={{
							backgroundColor: '#27ae60',
							color: 'white',
							border: 'none',
							borderRadius: '4px',
							padding: '6px 12px',
							fontSize: '10px',
							cursor: 'pointer',
							flex: 1
						}}>
							🚀 Execute Query
						</button>
						<button style={{
							backgroundColor: '#95a5a6',
							color: 'white',
							border: 'none',
							borderRadius: '4px',
							padding: '6px 8px',
							fontSize: '10px',
							cursor: 'pointer'
						}}>
							💾
						</button>
					</>
				) : (
					<>
						<button style={{
							backgroundColor: '#3498db',
							color: 'white',
							border: 'none',
							borderRadius: '4px',
							padding: '6px 12px',
							fontSize: '10px',
							cursor: 'pointer',
							flex: 1
						}}>
							🔄 Re-run with Changes
						</button>
						<button style={{
							backgroundColor: '#e67e22',
							color: 'white',
							border: 'none',
							borderRadius: '4px',
							padding: '6px 8px',
							fontSize: '10px',
							cursor: 'pointer'
						}}>
							📤
						</button>
					</>
				)}
			</div>
		</div>
	);
};

// Sample UI DSL generated by LLM
const sampleUIDSL = {
	title: "Customer Analytics Query Builder",
	description: "Dynamically query customer data with filters and aggregations",
	fields: [
		{
			name: "timeRange",
			label: "Time Range",
			type: "select",
			required: true,
			defaultValue: "30d",
			options: [
				{ value: "7d", label: "Last 7 days" },
				{ value: "30d", label: "Last 30 days" },
				{ value: "90d", label: "Last 90 days" },
				{ value: "1y", label: "Last year" }
			],
			help: "Select the time period for analysis"
		},
		{
			name: "minAmount",
			label: "Minimum Amount",
			type: "range",
			min: 0,
			max: 100000,
			step: 1000,
			defaultValue: 5000,
			unit: "USD",
			help: "Filter customers by minimum transaction amount"
		},
		{
			name: "status",
			label: "Customer Status",
			type: "select",
			defaultValue: "all",
			options: [
				{ value: "all", label: "All Customers" },
				{ value: "active", label: "Active Only" },
				{ value: "overdue", label: "Overdue Only" },
				{ value: "at_risk", label: "At Risk" }
			]
		},
		{
			name: "includePersonalData",
			label: "Include Personal Data",
			type: "checkbox",
			defaultValue: false,
			description: "Include phone numbers and addresses (requires approval)"
		}
	],
	advancedFields: [
		{
			name: "customFilter",
			label: "Custom SQL Filter",
			type: "text",
			placeholder: "e.g., region = 'US' AND tier = 'premium'",
			defaultValue: ""
		},
		{
			name: "groupBy",
			label: "Group By",
			type: "text",
			placeholder: "e.g., region, tier, signup_month",
			defaultValue: ""
		}
	],
	generateToolCall: (formData: any) => {
		const filters = [];
		if (formData.minAmount) filters.push(`amount >= ${formData.minAmount}`);
		if (formData.status !== 'all') filters.push(`status = '${formData.status}'`);
		if (formData.customFilter) filters.push(formData.customFilter);
		
		const fields = formData.includePersonalData 
			? "name, email, phone, address, amount, status, created_at"
			: "name, email, amount, status, created_at";
			
		return `query_customers(
  fields="${fields}",
  time_range="${formData.timeRange}",
  filters="${filters.join(' AND ')}",
  group_by="${formData.groupBy || ''}"
)`;
	},
	resultActions: [
		{ label: "Export CSV", icon: "📊", color: "#27ae60" },
		{ label: "Create Report", icon: "📋", color: "#3498db" },
		{ label: "Schedule Alert", icon: "🔔", color: "#f39c12" },
		{ label: "Filter Sensitive", icon: "🔒", color: "#e74c3c" }
	],
	defaultValues: {
		timeRange: "30d",
		minAmount: 5000,
		status: "overdue",
		includePersonalData: false
	}
};

export const llmGeneratedUISequence: InteractionSequence = {
	title: 'LLM-Generated Dynamic UI',
	
	subtitle: (state: InteractionState) => {
		if (state.activeStates.includes('uiGeneration')) {
			return 'LLM generates custom UI DSL for tool interaction';
		} else if (state.activeStates.includes('userInteraction')) {
			return 'User manipulates LLM-generated interface';
		} else if (state.activeStates.includes('resultManipulation')) {
			return 'User processes results through generated UI';
		}
		return 'Dynamic UI generation for tool calls and result processing';
	},

	messageTypes: {
		...DEFAULT_MESSAGE_TYPES,
		// UI DSL related
		ui_dsl: {
			bg: '#4a90e2',
			icon: '🎨',
			label: 'UI DSL',
			fontSize: '11px',
			padding: '8px 12px',
			boxShadow: '0 3px 10px rgba(74, 144, 226, 0.3)',
			border: '2px solid rgba(255, 255, 255, 0.3)',
		},
		dynamic_ui: {
			bg: '#9b59b6',
			icon: '🖥️',
			label: 'Dynamic UI',
			fontSize: '11px',
			padding: '8px 12px',
			boxShadow: '0 3px 10px rgba(155, 89, 182, 0.3)',
			border: '2px solid rgba(255, 255, 255, 0.3)',
		},
		ui_interaction: {
			bg: '#e67e22',
			icon: '👆',
			label: 'User Input',
			fontSize: '11px',
			padding: '8px 12px',
			boxShadow: '0 3px 10px rgba(230, 126, 34, 0.3)',
		},
	},
	
	states: [
		createState('container', 0, 30),
		createState('userRequest', 30, 40),
		createState('llmAnalysis', 70, 50),
		createState('uiGeneration', 120, 60),
		createState('uiRendering', 180, 40),
		createState('userInteraction', 220, 60),
		createState('toolExecution', 280, 40),
		createState('resultDisplay', 320, 40),
		createState('resultManipulation', 360, 60),
		createState('finalOutput', 420, 40),
	],

	messages: [
		// User request
		createMessage(
			'user-request',
			'user',
			'I need to analyze customer data with flexible filtering options',
			['userRequest', 'llmAnalysis', 'uiGeneration', 'uiRendering', 'userInteraction', 'toolExecution', 'resultDisplay', 'resultManipulation', 'finalOutput'],
			{ column: 'left' }
		),

		// LLM analysis
		createMessage(
			'llm-analysis',
			'assistant',
			'I\'ll create a dynamic interface for customer analytics. Let me generate a custom UI that allows you to configure the query parameters interactively.',
			['llmAnalysis', 'uiGeneration', 'uiRendering', 'userInteraction', 'toolExecution', 'resultDisplay', 'resultManipulation', 'finalOutput'],
			{ column: 'left' }
		),

		// UI DSL generation
		createMessage(
			'ui-dsl-generation',
			'ui_dsl',
			`Generated UI DSL:
{
  "title": "Customer Analytics Query Builder",
  "fields": [
    {"name": "timeRange", "type": "select", "options": ["7d", "30d", "90d"]},
    {"name": "minAmount", "type": "range", "min": 0, "max": 100000},
    {"name": "status", "type": "select", "options": ["all", "active", "overdue"]},
    {"name": "includePersonalData", "type": "checkbox"}
  ],
  "generateToolCall": "query_customers(...)",
  "resultActions": ["export", "filter", "schedule"]
}`,
			['uiGeneration', 'uiRendering', 'userInteraction', 'toolExecution', 'resultDisplay', 'resultManipulation', 'finalOutput'],
			{ column: 'left' }
		),

		// Rendered UI
		createMessage(
			'rendered-ui',
			'dynamic_ui',
			(state: InteractionState) => (
				<DynamicToolUI 
					uiDsl={sampleUIDSL}
					toolCall="query_customers(...)"
					result={state.activeStates.includes('resultDisplay') || state.activeStates.includes('resultManipulation') || state.activeStates.includes('finalOutput') ? 
						'Found 8 customers:\n• Acme Corp: $12,500 (45 days overdue)\n• TechStart Inc: $18,750 (62 days overdue)\n• Global Systems: $8,200 (38 days overdue)\n• Innovation Labs: $15,300 (51 days overdue)\n• Future Tech: $6,800 (29 days overdue)\n• Smart Solutions: $22,100 (67 days overdue)\n• Digital Dynamics: $9,400 (42 days overdue)\n• Cloud Corp: $11,900 (55 days overdue)' : 
						undefined}
					isExecuted={state.activeStates.includes('resultDisplay') || state.activeStates.includes('resultManipulation') || state.activeStates.includes('finalOutput')}
				/>
			),
			['uiRendering', 'userInteraction', 'toolExecution', 'resultDisplay', 'resultManipulation', 'finalOutput'],
			{ column: 'right', isReactContent: true }
		),

		// User interaction
		createMessage(
			'user-interaction',
			'ui_interaction',
			'User adjusts: Time Range → 30d, Min Amount → $5,000, Status → Overdue, Personal Data → No',
			['userInteraction', 'toolExecution', 'resultDisplay', 'resultManipulation', 'finalOutput'],
			{ column: 'right' }
		),

		// Tool execution
		createMessage(
			'tool-execution',
			'tool_use',
			'query_customers(fields="name,email,amount,status,created_at", time_range="30d", filters="amount >= 5000 AND status = \'overdue\'")',
			['toolExecution', 'resultDisplay', 'resultManipulation', 'finalOutput'],
			{ column: 'left' }
		),

		// Result display
		createMessage(
			'result-display',
			'tool_result',
			'Found 8 customers with overdue invoices over $5,000 in the last 30 days. Total outstanding: $104,950. Average overdue period: 47 days.',
			['resultDisplay', 'resultManipulation', 'finalOutput'],
			{ column: 'left' }
		),

		// Result manipulation
		createMessage(
			'result-manipulation',
			'ui_interaction',
			'User clicks "Filter Sensitive" → removes personal identifiers, then "Create Report" → generates summary',
			['resultManipulation', 'finalOutput'],
			{ column: 'right' }
		),

		// Final output
		createMessage(
			'final-output',
			'assistant',
			'Generated filtered report: 8 overdue customers, $104,950 total outstanding, 47-day average overdue period. Report saved and scheduled for weekly updates.',
			['finalOutput'],
			{ column: 'left' }
		),
	],

	overlays: [
		{
			id: 'ui-generation-indicator',
			content: (state: InteractionState) => {
				if (state.activeStates.includes('uiGeneration')) {
					return `
						<div style="
							background-color: rgba(74, 144, 226, 0.9);
							color: white;
							padding: 8px 20px;
							border-radius: 20px;
							font-size: 12px;
							font-weight: bold;
							box-shadow: 0 2px 10px rgba(74, 144, 226, 0.4);
						">
							🎨 GENERATING DYNAMIC UI
						</div>
					`;
				} else if (state.activeStates.includes('userInteraction')) {
					return `
						<div style="
							background-color: rgba(230, 126, 34, 0.9);
							color: white;
							padding: 8px 20px;
							border-radius: 20px;
							font-size: 12px;
							font-weight: bold;
							box-shadow: 0 2px 10px rgba(230, 126, 34, 0.4);
						">
							👆 USER CONFIGURING QUERY
						</div>
					`;
				} else if (state.activeStates.includes('resultManipulation')) {
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
							🔧 PROCESSING RESULTS
						</div>
					`;
				}
				return '';
			},
			position: {
				top: '15%',
				right: '5%',
			},
			visibleStates: ['uiGeneration', 'userInteraction', 'resultManipulation'],
		},
		{
			id: 'ui-dsl-flow',
			content: () => `
				<div style="
					background-color: rgba(52, 73, 94, 0.8);
					color: white;
					padding: 12px;
					border-radius: 8px;
					font-size: 10px;
					max-width: 200px;
					border: 1px solid rgba(255, 255, 255, 0.2);
				">
					<div style="font-weight: bold; margin-bottom: 8px;">🔄 UI DSL Flow</div>
					<div>1. LLM analyzes request</div>
					<div>2. Generates UI schema</div>
					<div>3. Renders interactive form</div>
					<div>4. User configures parameters</div>
					<div>5. Executes dynamic tool call</div>
					<div>6. Processes results via UI</div>
				</div>
			`,
			position: {
				bottom: '10%',
				left: '5%',
			},
			visibleStates: ['uiGeneration', 'uiRendering', 'userInteraction'],
		},
	],

	layout: {
		columns: 2,
		autoFill: false,
		maxMessagesPerColumn: 6,
	},
}; 