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

// Custom message types for SQLite animation
const sqliteMessageTypes = {
	...DEFAULT_MESSAGE_TYPES,
	
	schema_discovery: createMessageType('#3498db', '🔍', 'Schema Discovery', {
		fontSize: FONT_SIZES.small,
		padding: '10px 14px',
		border: '2px solid rgba(52, 152, 219, 0.4)',
		boxShadow: '0 3px 10px rgba(52, 152, 219, 0.3)',
	}),
	
	table_exploration: createMessageType('#9b59b6', '🗂️', 'Table Structure', {
		fontSize: FONT_SIZES.small,
		padding: '10px 14px',
		border: '2px solid rgba(155, 89, 182, 0.4)',
		boxShadow: '0 3px 10px rgba(155, 89, 182, 0.3)',
	}),
	
	targeted_query: createMessageType('#27ae60', '🎯', 'Targeted Query', {
		fontSize: FONT_SIZES.small,
		padding: '10px 14px',
		border: '2px solid rgba(39, 174, 96, 0.4)',
		boxShadow: '0 3px 10px rgba(39, 174, 96, 0.3)',
		fontWeight: 'bold',
	}),
	
	database_response: createMessageType('#16a085', '🗃️', 'Database', {
		fontSize: FONT_SIZES.small,
		padding: '8px 12px',
		border: '1px solid rgba(22, 160, 133, 0.3)',
	}),
	
	insight: createMessageType('#f39c12', '💡', 'Insight', {
		fontSize: FONT_SIZES.small,
		padding: '10px 14px',
		border: '2px solid rgba(243, 156, 18, 0.4)',
		boxShadow: '0 3px 10px rgba(243, 156, 18, 0.3)',
	}),
	
	final_result: createMessageType('#27ae60', '✅', 'Result', {
		fontSize: FONT_SIZES.small,
		padding: '12px 16px',
		border: '3px solid rgba(39, 174, 96, 0.5)',
		boxShadow: '0 4px 15px rgba(39, 174, 96, 0.4)',
		fontWeight: 'bold',
	}),
};

// Schema visualization component
const SchemaVisualization: React.FC<{
	tables: string[];
	isVisible: boolean;
}> = ({ tables, isVisible }) => {
	if (!isVisible) return null;
	
	return (
		<div style={{
			display: 'flex',
			gap: '15px',
			justifyContent: 'center',
			flexWrap: 'wrap',
			padding: '15px',
			backgroundColor: 'rgba(255, 255, 255, 0.1)',
			borderRadius: '12px',
			border: '2px solid rgba(52, 152, 219, 0.3)',
		}}>
			{tables.map((table, idx) => (
				<div key={idx} style={{
					padding: '8px 16px',
					backgroundColor: 'rgba(52, 152, 219, 0.2)',
					borderRadius: '8px',
					color: 'white',
					fontSize: '12px',
					fontWeight: 'bold',
					border: '1px solid rgba(52, 152, 219, 0.4)',
				}}>
					📋 {table}
				</div>
			))}
		</div>
	);
};

// Table structure component
const TableStructure: React.FC<{
	tableName: string;
	columns: Array<{ name: string; type: string; isKey?: boolean }>;
	isVisible: boolean;
}> = ({ tableName, columns, isVisible }) => {
	if (!isVisible) return null;
	
	return (
		<div style={{
			backgroundColor: 'rgba(255, 255, 255, 0.95)',
			borderRadius: '10px',
			padding: '12px',
			boxShadow: '0 4px 15px rgba(0,0,0,0.15)',
			fontSize: '11px',
			color: '#2c3e50',
			fontFamily: 'monospace',
			maxWidth: '200px',
			border: '2px solid rgba(155, 89, 182, 0.3)',
		}}>
			<div style={{
				fontWeight: 'bold',
				marginBottom: '8px',
				color: '#9b59b6',
				fontSize: '12px',
			}}>
				{tableName}:
			</div>
			{columns.map((col, idx) => (
				<div key={idx} style={{
					color: col.isKey ? '#e74c3c' : '#2c3e50',
					fontWeight: col.isKey ? 'bold' : 'normal',
				}}>
					• {col.name} ({col.type})
				</div>
			))}
		</div>
	);
};

// SQL Query component
const SQLQuery: React.FC<{
	query: string;
	callNumber: number;
	isVisible: boolean;
	isTargeted?: boolean;
}> = ({ query, callNumber, isVisible, isTargeted }) => {
	if (!isVisible) return null;
	
	return (
		<div style={{
			backgroundColor: '#2c3e50',
			borderRadius: '8px',
			padding: '12px',
			boxShadow: '0 4px 15px rgba(0,0,0,0.3)',
			fontSize: '11px',
			color: '#ecf0f1',
			fontFamily: 'monospace',
			maxWidth: isTargeted ? '400px' : '280px',
			border: `2px solid ${isTargeted ? '#27ae60' : '#3498db'}`,
		}}>
			<div style={{
				paddingLeft: '12px',
				color: '#f39c12',
				lineHeight: 1.3,
				whiteSpace: 'pre-wrap',
				fontSize: '10px',
			}}>
				"{query}"
			</div>
			
			{isTargeted && (
				<div style={{
					marginTop: '8px',
					padding: '6px',
					backgroundColor: 'rgba(39, 174, 96, 0.2)',
					borderRadius: '4px',
					fontSize: '9px',
				}}>
					<div style={{ color: '#27ae60', fontWeight: 'bold' }}>✨ Smart query features:</div>
					<div style={{ color: '#ecf0f1' }}>• Joins only needed tables</div>
					<div style={{ color: '#ecf0f1' }}>• Filters by exact customer name</div>
					<div style={{ color: '#ecf0f1' }}>• Includes date range filter</div>
					<div style={{ color: '#ecf0f1' }}>• Returns only the count</div>
				</div>
			)}
		</div>
	);
};

export const sqliteQuerySequence: InteractionSequence = {
	title: 'Intelligent Multi-Step Tool Use',
	
	subtitle: (state: InteractionState) => {
		if (state.activeStates.includes('userRequest')) {
			return 'User needs database information';
		} else if (state.activeStates.includes('schemaDiscovery')) {
			return 'LLM explores database schema';
		} else if (state.activeStates.includes('tableExploration')) {
			return 'Exploring table structures';
		} else if (state.activeStates.includes('targetedQuery')) {
			return 'Crafting precise, targeted query';
		} else if (state.activeStates.includes('finalResponse')) {
			return 'Efficient response delivery';
		}
		return 'Exploring database schema to craft precise queries';
	},

	messageTypes: sqliteMessageTypes,
	
	states: [
		// User request phase (frames 90-210 → 60-180)
		createState('container', 0, 600),
		createState('userRequest', 60, 120),
		createState('userSpeaks', 90, 90),
		
		// Schema discovery phase (frames 210-360 → 180-330)
		createState('schemaDiscovery', 180, 150),
		createState('llmThinks', 210, 60),
		createState('schemaQuery', 270, 300), // Keep SQL visible longer
		createState('schemaResult', 330, 300),
		
		// Table exploration phase (frames 360-540 → 390-570)
		createState('tableExploration', 390, 180),
		createState('tableThinking', 420, 60),
		createState('customersQuery', 480, 300), // Keep SQL visible longer
		createState('customersResult', 540, 300),
		createState('ordersQuery', 600, 300), // Keep SQL visible longer
		createState('ordersResult', 660, 300),
		createState('schemaInsight', 720, 60),
		
		// Targeted query phase (frames 540-780 → 780-1020)
		createState('targetedQuery', 780, 240),
		createState('targetedThinking', 810, 60),
		createState('targetedExecution', 870, 300), // Keep SQL visible longer
		createState('targetedResult', 960, 300),
		
		// Final response phase (frames 780-1200 → 1020-1440)
		createState('finalResponse', 1020, 420),
		createState('finalThinking', 1050, 60),
		createState('llmResponse', 1110, 90),
		createState('userReceives', 1200, 60),
		createState('processSummary', 1260, 90),
		createState('finalMessage', 1350, 90),
	],

	messages: [
		// User request
		createMessage(
			'user-question',
			'user',
			'"How many orders did customer John Smith place last month?"',
			['container', 'userSpeaks', 'schemaDiscovery', 'tableExploration', 'targetedQuery', 'finalResponse']
		),

		createMessage(
			'user-context',
			'system',
			'Requires database knowledge - only tool available: sqlite_query(sql)',
			['userSpeaks', 'schemaDiscovery', 'tableExploration', 'targetedQuery', 'finalResponse']
		),

		// LLM initial thinking
		createMessage(
			'llm-initial-thought',
			'assistant',
			'"I need to find customer orders, but I don\'t know the database structure. Let me explore the schema first."',
			['llmThinks', 'schemaDiscovery', 'tableExploration', 'targetedQuery', 'finalResponse']
		),

		// Schema discovery
		createMessage(
			'schema-query',
			'schema_discovery',
			(state: InteractionState) => (
				<SQLQuery
					query="SELECT name FROM sqlite_master WHERE type='table';"
					callNumber={1}
					isVisible={true}
				/>
			),
			['schemaQuery', 'schemaResult', 'tableExploration', 'targetedQuery', 'finalResponse'],
			{ isReactContent: true }
		),

		// Table exploration thinking
		createMessage(
			'table-exploration-thought',
			'assistant',
			'"I found customers & orders tables. Let me check their structure to understand how to join them."',
			['tableThinking', 'tableExploration', 'targetedQuery', 'finalResponse']
		),

		// Customers table exploration
		createMessage(
			'customers-query',
			'table_exploration',
			(state: InteractionState) => (
				<SQLQuery
					query="PRAGMA table_info(customers);"
					callNumber={2}
					isVisible={true}
				/>
			),
			['customersQuery', 'customersResult', 'targetedQuery', 'finalResponse'],
			{ isReactContent: true }
		),

		createMessage(
			'customers-result',
			'database_response',
			(state: InteractionState) => (
				<TableStructure
					tableName="customers"
					columns={[
						{ name: 'id', type: 'INTEGER' },
						{ name: 'name', type: 'TEXT' },
						{ name: 'email', type: 'TEXT' },
						{ name: 'created_at', type: 'TEXT' },
					]}
					isVisible={true}
				/>
			),
			['customersResult', 'targetedQuery', 'finalResponse'],
			{ isReactContent: true }
		),

		// Orders table exploration
		createMessage(
			'orders-query',
			'table_exploration',
			(state: InteractionState) => (
				<SQLQuery
					query="PRAGMA table_info(orders);"
					callNumber={3}
					isVisible={true}
				/>
			),
			['ordersQuery', 'ordersResult', 'targetedQuery', 'finalResponse'],
			{ isReactContent: true }
		),

		createMessage(
			'orders-result',
			'database_response',
			(state: InteractionState) => (
				<TableStructure
					tableName="orders"
					columns={[
						{ name: 'id', type: 'INTEGER' },
						{ name: 'customer_id', type: 'INTEGER', isKey: true },
						{ name: 'amount', type: 'REAL' },
						{ name: 'order_date', type: 'TEXT' },
					]}
					isVisible={true}
				/>
			),
			['ordersResult', 'targetedQuery', 'finalResponse'],
			{ isReactContent: true }
		),

		// Targeted query thinking
		createMessage(
			'targeted-thinking',
			'assistant',
			'"Perfect! Now I know the schema. I can write a precise query that joins customers and orders, filters by name and date range."',
			['targetedThinking', 'targetedQuery', 'finalResponse']
		),

		// Targeted query execution
		createMessage(
			'targeted-execution',
			'targeted_query',
			(state: InteractionState) => (
				<SQLQuery
					query={`SELECT COUNT(*) as order_count
FROM orders o
JOIN customers c ON o.customer_id = c.id
WHERE c.name = 'John Smith'
AND o.order_date LIKE '2024-11%';`}
					callNumber={4}
					isVisible={true}
					isTargeted={true}
				/>
			),
			['targetedExecution', 'targetedResult', 'finalResponse'],
			{ isReactContent: true }
		),

		// Targeted result
		createMessage(
			'targeted-result',
			'final_result',
			'🎯 Precise Result: order_count: 7\n\nExactly what was asked - no unnecessary data!',
			['targetedResult', 'finalResponse']
		),
	],

	overlays: [
		{
			id: 'step-indicator',
			content: (state: InteractionState) => {
				let step = '';
				let color = '#3498db';
				
				if (state.activeStates.includes('userRequest')) {
					step = 'Step 1: User needs database information';
					color = '#3498db';
				} else if (state.activeStates.includes('schemaDiscovery')) {
					step = 'Step 2: LLM explores database schema';
					color = '#3498db';
				} else if (state.activeStates.includes('tableExploration')) {
					step = 'Step 3: Exploring table structures';
					color = '#9b59b6';
				} else if (state.activeStates.includes('targetedQuery')) {
					step = 'Step 4: Crafting precise, targeted query';
					color = '#27ae60';
				} else if (state.activeStates.includes('finalResponse')) {
					step = 'Step 5: Efficient response delivery';
					color = '#27ae60';
				}
				
				if (!step) return '';
				
				return `
					<div style="
						background-color: rgba(44, 62, 80, 0.9);
						color: white;
						padding: 12px 24px;
						border-radius: 20px;
						font-size: 16px;
						font-weight: bold;
						box-shadow: 0 4px 15px rgba(0,0,0,0.3);
						border: 2px solid ${color};
						text-align: center;
						margin: 0 auto;
					">
						${step}
					</div>
				`;
			},
			position: {
				top: '8%',
				left: '25%',
			},
			visibleStates: ['userRequest', 'schemaDiscovery', 'tableExploration', 'targetedQuery', 'finalResponse'],
		},
		{
			id: 'efficiency-stats',
			content: () => `
				<div style="
					background-color: rgba(39, 174, 96, 0.9);
					color: white;
					padding: 12px 20px;
					border-radius: 12px;
					font-size: 12px;
					text-align: center;
					display: flex;
					gap: 20px;
					align-items: center;
					box-shadow: 0 3px 10px rgba(0,0,0,0.3);
					justify-content: center;
				">
					<div>
						<div style="font-weight: bold;">Schema Discovery:</div>
						<div>2 exploration queries</div>
					</div>
					<div>
						<div style="font-weight: bold;">Final Query:</div>
						<div>1 targeted result</div>
					</div>
					<div>
						<div style="font-weight: bold;">Total Tokens:</div>
						<div>~300 (vs 3,600+ bulk)</div>
					</div>
				</div>
			`,
			position: {
				bottom: '8%',
				left: '10%',
			},
			visibleStates: ['processSummary', 'finalMessage'],
		},
	],

	layout: {
		columns: 2,
		autoFill: true,
		maxMessagesPerColumn: 7,
	},
	
}; 