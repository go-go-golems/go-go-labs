import React from 'react';
import {
	InteractionSequence,
	createState,
	createMessage,
	createMessageType,
	DEFAULT_MESSAGE_TYPES,
	InteractionState,
	FONT_SIZES,
} from '../../types/InteractionDSL';

// Scrolling Data Widget Component
const ScrollingDataWidget: React.FC<{ 
	scrollProgress: number;
	isVisible: boolean;
	shouldCollapse?: boolean;
}> = ({ scrollProgress, isVisible, shouldCollapse = false }) => {
	// Generate massive company data
	const generateCompanyData = () => {
		const companies = [
			'OpenAI', 'Microsoft', 'Google', 'Apple', 'Amazon', 'Meta', 'Tesla', 'Netflix',
			'Salesforce', 'Oracle', 'IBM', 'Intel', 'NVIDIA', 'Adobe', 'Uber', 'Airbnb',
			'Spotify', 'Zoom', 'Slack', 'Dropbox', 'Twitter', 'LinkedIn', 'PayPal', 'Square',
			'Stripe', 'Shopify', 'ServiceNow', 'Snowflake', 'Palantir', 'Datadog',
			'MongoDB', 'Atlassian', 'Twilio', 'Okta', 'CrowdStrike', 'Zscaler'
		];
		
		return companies.map((company, index) => ({
			id: 1000 + index,
			name: company,
			email: `contact@${company.toLowerCase().replace(/\s+/g, '')}.com`,
			phone: `+1-555-${String(Math.floor(Math.random() * 9000) + 1000)}`,
			address: `${Math.floor(Math.random() * 9999) + 1} Tech St, Silicon Valley, CA`,
			employees: Math.floor(Math.random() * 50000) + 100,
			revenue: `$${Math.floor(Math.random() * 100)}B`,
			founded: Math.floor(Math.random() * 30) + 1990,
		}));
	};

	const allCompanies = React.useMemo(() => generateCompanyData(), []);
	const scrollOffset = scrollProgress * (allCompanies.length - 6) * 80;
	
	// Calculate collapsed height - show only the target OpenAI record when collapsed
	const collapsedHeight = shouldCollapse ? '160px' : '300px';
	const isCollapsed = shouldCollapse;

	if (!isVisible) return null;

	return (
		<div style={{
			backgroundColor: 'rgba(255, 255, 255, 0.95)',
			borderRadius: '12px',
			padding: '16px',
			boxShadow: '0 4px 15px rgba(0,0,0,0.2)',
			fontSize: '10px',
			color: '#2c3e50',
			fontFamily: 'monospace',
			height: collapsedHeight,
			overflow: 'hidden',
			position: 'relative',
			border: '2px solid #e74c3c',
			transition: 'height 0.5s ease-in-out',
		}}>
			<div style={{
				fontWeight: 'bold', 
				marginBottom: '12px', 
				fontSize: '12px', 
				color: isCollapsed ? '#27ae60' : '#e74c3c',
				textAlign: 'center'
			}}>
				{isCollapsed ? 
					'✅ FOUND TARGET: OpenAI (1 of 36 companies)' : 
					'🚨 MASSIVE RESPONSE: ' + allCompanies.length + ' companies (Only need 1!)'
				}
			</div>
			
			{isCollapsed ? (
				// Show only the OpenAI record when collapsed
				<div style={{
					padding: '8px',
					backgroundColor: 'rgba(39, 174, 96, 0.2)',
					borderRadius: '6px',
					borderLeft: '3px solid #27ae60',
				}}>
					<div style={{
						fontWeight: 'bold', 
						color: '#27ae60',
						fontSize: '11px'
					}}>
						OpenAI ← TARGET FOUND!
					</div>
					<div>📧 contact@openai.com</div>
					<div>📞 +1-555-OPENAI</div>
					<div>🏢 2,500 employees</div>
					<div>💰 $29B</div>
				</div>
			) : (
				<div style={{
					transform: `translateY(-${scrollOffset}px)`,
					lineHeight: 1.6,
				}}>
					{allCompanies.map((company, index) => (
						<div
							key={index}
							style={{
								marginBottom: '12px',
								padding: '8px',
								backgroundColor: company.name === 'OpenAI' ? 'rgba(39, 174, 96, 0.2)' : 'rgba(0,0,0,0.05)',
								borderRadius: '6px',
								borderLeft: company.name === 'OpenAI' ? '3px solid #27ae60' : '3px solid transparent',
							}}
						>
							<div style={{
								fontWeight: 'bold', 
								color: company.name === 'OpenAI' ? '#27ae60' : '#2c3e50',
								fontSize: '11px'
							}}>
								{company.name} {company.name === 'OpenAI' ? '← TARGET!' : ''}
							</div>
							<div>📧 {company.email}</div>
							<div>📞 {company.phone}</div>
							<div>🏢 {company.employees} employees</div>
							<div>💰 {company.revenue}</div>
						</div>
					))}
				</div>
			)}
			
			{/* Scroll indicator - only show when not collapsed */}
			{!isCollapsed && (
				<div style={{
					position: 'absolute',
					right: '8px',
					top: '40px',
					bottom: '8px',
					width: '6px',
					backgroundColor: 'rgba(0,0,0,0.1)',
					borderRadius: '3px',
				}}>
					<div style={{
						position: 'absolute',
						top: `${(scrollOffset / ((allCompanies.length - 6) * 80)) * 100}%`,
						width: '6px',
						height: '30px',
						backgroundColor: '#e74c3c',
						borderRadius: '3px',
					}} />
				</div>
			)}
		</div>
	);
};

// Custom message types for CRM query sequence
const crmQueryMessageTypes = {
	...DEFAULT_MESSAGE_TYPES,
	inefficient_warning: createMessageType('#e74c3c', '⚠️', 'Inefficiency Warning', {
		fontSize: '12px',
		padding: '10px 14px',
		boxShadow: '0 3px 10px rgba(231, 76, 60, 0.3)',
		border: '2px solid rgba(231, 76, 60, 0.5)',
		fontWeight: 'bold',
	}),
	crm_database: createMessageType('#27ae60', '🗄️', 'CRM Database', {
		fontSize: '11px',
		padding: '8px 12px',
		border: '1px solid rgba(39, 174, 96, 0.3)',
	}),
	data_flood: createMessageType('#f39c12', '🌊', 'Data Flood', {
		fontSize: '11px',
		padding: '8px 12px',
		border: '1px solid rgba(243, 156, 18, 0.3)',
		fontWeight: 'bold',
	}),
	token_counter: createMessageType('#8e44ad', '🔢', 'Token Usage', {
		fontSize: '10px',
		padding: '6px 10px',
		border: '1px solid rgba(142, 68, 173, 0.3)',
	}),
};

export const crmQuerySequence: InteractionSequence = {
	title: 'The Token Inefficiency Problem',
	
	subtitle: (state: InteractionState) => {
		if (state.activeStates.includes('userRequest')) {
			return 'User makes a simple, specific request';
		} else if (state.activeStates.includes('llmAnalysis')) {
			return 'LLM analyzes the request';
		} else if (state.activeStates.includes('toolExecution')) {
			return 'Tool returns MASSIVE unfiltered dataset';
		} else if (state.activeStates.includes('dataFlood')) {
			return 'Scrolling through thousands of irrelevant records';
		} else if (state.activeStates.includes('tokenOverload')) {
			return 'Token count explodes for a simple answer';
		} else if (state.activeStates.includes('inefficientResult')) {
			return 'Finally finds the one needed record';
		}
		return 'When simple queries return massive datasets';
	},

	messageTypes: crmQueryMessageTypes,
	
	states: [
		// Natural flow showing the inefficiency problem
		createState('container', 0, 60),
		createState('userRequest', 60, 90),
		createState('llmAnalysis', 120, 90),
		createState('toolExecution', 180, 120),
		createState('dataFlood', 270, 180),
		createState('scrollingData', 330, 150),
		createState('tokenOverload', 420, 120),
		createState('inefficientResult', 510, 90),
		createState('problemSummary', 570, 60),
	],

	messages: [
		// User Request
		createMessage(
			'user-simple-request',
			'user',
			'"Give me the contact information for OpenAI"',
			['container', 'userRequest', 'llmAnalysis', 'toolExecution', 'dataFlood', 'scrollingData', 'tokenOverload', 'inefficientResult', 'problemSummary']
		),

		createMessage(
			'request-analysis',
			'assistant',
			'Simple request: User wants contact info for one specific company. Let me query the CRM database.',
			['llmAnalysis', 'toolExecution', 'dataFlood', 'scrollingData', 'tokenOverload', 'inefficientResult', 'problemSummary']
		),

		// Tool Analysis & Execution
		createMessage(
			'tool-call-unfiltered',
			'tool_use',
			'get_crm_companies()',
			['toolExecution', 'dataFlood', 'scrollingData', 'tokenOverload', 'inefficientResult', 'problemSummary']
		),

		createMessage(
			'database-response',
			'crm_database',
			'Executing query... No filters applied! Returning ALL company records from database.',
			['toolExecution', 'dataFlood', 'scrollingData', 'tokenOverload', 'inefficientResult', 'problemSummary']
		),

		// Data Flood with Scrolling Widget
		createMessage(
			'massive-dataset',
			'data_flood',
			(state: InteractionState) => (
				<ScrollingDataWidget 
					scrollProgress={state.activeStates.includes('scrollingData') ? 
						Math.min((state.currentFrame - 330) / 150, 1) : 0}
					isVisible={state.activeStates.includes('dataFlood') || 
						state.activeStates.includes('scrollingData') || 
						state.activeStates.includes('tokenOverload') ||
						state.activeStates.includes('inefficientResult') ||
						state.activeStates.includes('problemSummary')}
					shouldCollapse={state.activeStates.includes('inefficientResult') || 
						state.activeStates.includes('problemSummary')
					}
				/>
			),
			['dataFlood', 'scrollingData', 'tokenOverload', 'inefficientResult', 'problemSummary'],
			{ isReactContent: true }
		),


		// Inefficient Result
		createMessage(
			'finally-found',
			'assistant',
			'Found it! OpenAI contact: contact@openai.com, +1-555-OPENAI. But I had to process 36 companies to find this one result.',
			['inefficientResult', 'problemSummary']
		),

	],

	layout: {
		columns: 1,
		autoFill: true,
		maxMessagesPerColumn: 12,
	},

}; 