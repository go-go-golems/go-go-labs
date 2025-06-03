import React from 'react';
import {
	InteractionSequence,
	createState,
	createMessage,
	createMessageType,
	DEFAULT_MESSAGE_TYPES,
	InteractionState,
} from '../../types/InteractionDSL';

// Custom message types for SQLite View Optimization animation
const sqliteViewMessageTypes = {
	...DEFAULT_MESSAGE_TYPES,
	
	view_creation: createMessageType('#9b59b6', '🏗️', 'View Creation', {
		fontSize: '12px',
		padding: '10px 14px',
		border: '2px solid rgba(155, 89, 182, 0.4)',
		boxShadow: '0 3px 10px rgba(155, 89, 182, 0.3)',
		fontWeight: 'bold',
	}),
	
	efficient_query: createMessageType('#3498db', '⚡', 'Efficient Query', {
		fontSize: '11px',
		padding: '8px 12px',
		border: '2px solid rgba(52, 152, 219, 0.4)',
		boxShadow: '0 3px 10px rgba(52, 152, 219, 0.3)',
	}),
	
	query_result: createMessageType('#27ae60', '📊', 'Result', {
		fontSize: '12px',
		padding: '10px 14px',
		border: '2px solid rgba(39, 174, 96, 0.4)',
		boxShadow: '0 3px 10px rgba(39, 174, 96, 0.3)',
		fontWeight: 'bold',
	}),
	
	performance_comparison: createMessageType('#e74c3c', '📈', 'Performance', {
		fontSize: '12px',
		padding: '10px 14px',
		border: '2px solid rgba(231, 76, 60, 0.4)',
		boxShadow: '0 3px 10px rgba(231, 76, 60, 0.3)',
	}),
	
	infrastructure: createMessageType('#8e44ad', '🏛️', 'Infrastructure', {
		fontSize: '12px',
		padding: '10px 14px',
		border: '2px solid rgba(142, 68, 173, 0.4)',
		boxShadow: '0 3px 10px rgba(142, 68, 173, 0.3)',
		fontWeight: 'bold',
	}),
	
	optimization: createMessageType('#f39c12', '🚀', 'Optimization', {
		fontSize: '13px',
		padding: '12px 16px',
		border: '3px solid rgba(243, 156, 18, 0.5)',
		boxShadow: '0 4px 15px rgba(243, 156, 18, 0.4)',
		fontWeight: 'bold',
	}),
};

// CREATE VIEW SQL component
const CreateViewSQL: React.FC<{
	isVisible: boolean;
}> = ({ isVisible }) => {
	if (!isVisible) return null;
	
	return (
		<div style={{
			backgroundColor: '#2c3e50',
			borderRadius: '15px',
			padding: '20px',
			boxShadow: '0 8px 25px rgba(0,0,0,0.4)',
			fontSize: '14px',
			color: '#ecf0f1',
			fontFamily: 'monospace',
			maxWidth: '600px',
			border: '3px solid #9b59b6',
		}}>
			<div style={{
				color: '#9b59b6',
				fontWeight: 'bold',
				marginBottom: '12px',
				fontSize: '16px',
			}}>
				🏗️ Tool Call: Create Infrastructure
			</div>
			<div style={{ color: '#e74c3c' }}>sqlite_query(</div>
			<div style={{
				paddingLeft: '15px',
				color: '#f39c12',
				lineHeight: 1.6,
				whiteSpace: 'pre-wrap',
			}}>
				{`"CREATE VIEW customer_orders_view AS
SELECT 
  c.id as customer_id,
  c.name as customer_name,
  c.email,
  o.id as order_id,
  o.amount,
  o.order_date
FROM customers c
JOIN orders o ON c.id = o.customer_id;"`}
			</div>
			<div style={{ color: '#e74c3c' }}>)</div>
			
			<div style={{
				marginTop: '12px',
				padding: '10px',
				backgroundColor: 'rgba(155, 89, 182, 0.2)',
				borderRadius: '8px',
				fontSize: '12px',
			}}>
				<div style={{ color: '#9b59b6', fontWeight: 'bold' }}>💎 Smart infrastructure:</div>
				<div style={{ color: '#ecf0f1' }}>• Pre-joins customers & orders</div>
				<div style={{ color: '#ecf0f1' }}>• Meaningful column names</div>
				<div style={{ color: '#ecf0f1' }}>• Reusable for multiple queries</div>
				<div style={{ color: '#ecf0f1' }}>• No repeated JOIN logic needed</div>
			</div>
		</div>
	);
};

// Efficient Query component
const EfficientQuery: React.FC<{
	queryNumber: number;
	title: string;
	sql: string;
	result: string;
	color: string;
	isVisible: boolean;
}> = ({ queryNumber, title, sql, result, color, isVisible }) => {
	if (!isVisible) return null;
	
	return (
		<div style={{
			display: 'flex',
			alignItems: 'center',
			gap: '20px',
			marginBottom: '15px',
		}}>
			<div style={{
				backgroundColor: '#2c3e50',
				borderRadius: '12px',
				padding: '12px',
				boxShadow: '0 6px 20px rgba(0,0,0,0.3)',
				fontSize: '12px',
				color: '#ecf0f1',
				fontFamily: 'monospace',
				maxWidth: '300px',
				border: `2px solid ${color}`,
				flex: 1,
			}}>
				<div style={{
					color: color,
					fontWeight: 'bold',
					marginBottom: '6px',
					fontSize: '13px',
				}}>
					{title}
				</div>
				<div style={{
					color: '#f39c12',
					fontSize: '11px',
					lineHeight: 1.4,
				}}>
					{sql}
				</div>
			</div>
			
			<div style={{
				backgroundColor: '#27ae60',
				borderRadius: '10px',
				padding: '10px 16px',
				color: 'white',
				fontSize: '14px',
				fontWeight: 'bold',
				textAlign: 'center',
				minWidth: '100px',
			}}>
				{result}
			</div>
		</div>
	);
};

// Multiple Queries Widget
const MultipleQueriesWidget: React.FC<{
	activeQueries: number;
	isVisible: boolean;
}> = ({ activeQueries, isVisible }) => {
	if (!isVisible) return null;
	
	const queries = [
		{
			title: '🔍 Query #1:',
			sql: 'SELECT COUNT(*) FROM customer_orders_view\nWHERE customer_name = \'John Smith\';',
			result: '7 orders',
			color: '#3498db',
		},
		{
			title: '💰 Query #2:',
			sql: 'SELECT SUM(amount) FROM customer_orders_view\nWHERE customer_name = \'John Smith\';',
			result: '$2,847.50',
			color: '#e74c3c',
		},
		{
			title: '📊 Query #3:',
			sql: 'SELECT AVG(amount) FROM customer_orders_view\nWHERE customer_name = \'John Smith\';',
			result: '$406.79',
			color: '#f39c12',
		},
		{
			title: '📅 Query #4:',
			sql: 'SELECT MAX(order_date) FROM customer_orders_view\nWHERE customer_name = \'John Smith\';',
			result: '2024-11-28',
			color: '#9b59b6',
		},
	];
	
	return (
		<div style={{
			backgroundColor: 'rgba(255, 255, 255, 0.1)',
			borderRadius: '15px',
			padding: '20px',
			border: '2px solid rgba(155, 89, 182, 0.3)',
			maxWidth: '500px',
		}}>
			{queries.slice(0, activeQueries).map((query, idx) => (
				<EfficientQuery
					key={idx}
					queryNumber={idx + 1}
					title={query.title}
					sql={query.sql}
					result={query.result}
					color={query.color}
					isVisible={true}
				/>
			))}
		</div>
	);
};

// Performance Comparison component
const PerformanceComparison: React.FC<{
	isVisible: boolean;
}> = ({ isVisible }) => {
	if (!isVisible) return null;
	
	return (
		<div style={{
			display: 'flex',
			gap: '20px',
			maxWidth: '800px',
		}}>
			{/* Before: Without Views */}
			<div style={{
				backgroundColor: '#e74c3c',
				borderRadius: '15px',
				padding: '15px',
				boxShadow: '0 8px 25px rgba(0,0,0,0.3)',
				color: 'white',
				flex: 1,
				border: '3px solid #c0392b',
			}}>
				<div style={{
					fontSize: '16px',
					fontWeight: 'bold',
					marginBottom: '12px',
					textAlign: 'center',
				}}>
					❌ Before: Without Views
				</div>
				<div style={{
					fontSize: '11px',
					lineHeight: 1.4,
					fontFamily: 'monospace',
				}}>
					<div style={{ marginBottom: '8px' }}>
						<strong>Query 1:</strong> SELECT COUNT(*) FROM orders o<br/>
						JOIN customers c ON o.customer_id = c.id<br/>
						WHERE c.name = 'John Smith';
					</div>
					<div style={{ marginBottom: '8px' }}>
						<strong>Query 2:</strong> SELECT SUM(amount) FROM orders o<br/>
						JOIN customers c ON o.customer_id = c.id<br/>
						WHERE c.name = 'John Smith';
					</div>
					<div style={{ fontSize: '10px', opacity: 0.8 }}>
						... 2 more similar queries
					</div>
				</div>
				<div style={{
					marginTop: '10px',
					padding: '8px',
					backgroundColor: 'rgba(192, 57, 43, 0.3)',
					borderRadius: '6px',
					textAlign: 'center',
					fontSize: '12px',
				}}>
					<div style={{ fontWeight: 'bold' }}>4 queries × 4 JOINs = 16 JOIN operations</div>
				</div>
			</div>

			{/* After: With Views */}
			<div style={{
				backgroundColor: '#27ae60',
				borderRadius: '15px',
				padding: '15px',
				boxShadow: '0 8px 25px rgba(0,0,0,0.3)',
				color: 'white',
				flex: 1,
				border: '3px solid #2ecc71',
			}}>
				<div style={{
					fontSize: '16px',
					fontWeight: 'bold',
					marginBottom: '12px',
					textAlign: 'center',
				}}>
					✅ After: With Views
				</div>
				<div style={{
					fontSize: '11px',
					lineHeight: 1.4,
					fontFamily: 'monospace',
				}}>
					<div style={{
						marginBottom: '8px',
						backgroundColor: 'rgba(46, 204, 113, 0.2)',
						padding: '6px',
						borderRadius: '4px',
					}}>
						<strong>Setup (once):</strong> CREATE VIEW customer_orders_view...
					</div>
					<div style={{ marginBottom: '6px' }}>
						<strong>Query 1:</strong> SELECT COUNT(*) FROM customer_orders_view<br/>
						WHERE customer_name = 'John Smith';
					</div>
					<div style={{ marginBottom: '6px' }}>
						<strong>Query 2:</strong> SELECT SUM(amount) FROM customer_orders_view<br/>
						WHERE customer_name = 'John Smith';
					</div>
					<div style={{ fontSize: '10px', opacity: 0.8 }}>
						... 2 more simple queries
					</div>
				</div>
				<div style={{
					marginTop: '10px',
					padding: '8px',
					backgroundColor: 'rgba(46, 204, 113, 0.3)',
					borderRadius: '6px',
					textAlign: 'center',
					fontSize: '12px',
				}}>
					<div style={{ fontWeight: 'bold' }}>1 view setup + 4 simple queries = 1 JOIN total</div>
				</div>
			</div>
		</div>
	);
};

// Performance Metrics component
const PerformanceMetrics: React.FC<{
	isVisible: boolean;
}> = ({ isVisible }) => {
	if (!isVisible) return null;
	
	return (
		<div style={{
			backgroundColor: 'rgba(44, 62, 80, 0.95)',
			borderRadius: '15px',
			padding: '20px',
			color: 'white',
			fontSize: '14px',
			textAlign: 'center',
			boxShadow: '0 8px 25px rgba(0,0,0,0.3)',
			border: '2px solid #9b59b6',
			maxWidth: '600px',
		}}>
			<div style={{ fontSize: '20px', marginBottom: '15px', fontWeight: 'bold' }}>📊 Performance Impact</div>
			<div style={{
				display: 'grid',
				gridTemplateColumns: '1fr 1fr 1fr 1fr',
				gap: '20px',
				fontSize: '12px',
			}}>
				<div>
					<div style={{ fontSize: '18px', fontWeight: 'bold', color: '#e74c3c', marginBottom: '4px' }}>16x</div>
					<div style={{ color: '#ecf0f1' }}>JOIN Operations</div>
					<div style={{ fontSize: '10px', color: '#bdc3c7' }}>Without views</div>
				</div>
				<div>
					<div style={{ fontSize: '18px', fontWeight: 'bold', color: '#27ae60', marginBottom: '4px' }}>1x</div>
					<div style={{ color: '#ecf0f1' }}>JOIN Operation</div>
					<div style={{ fontSize: '10px', color: '#bdc3c7' }}>With views</div>
				</div>
				<div>
					<div style={{ fontSize: '18px', fontWeight: 'bold', color: '#f39c12', marginBottom: '4px' }}>~500</div>
					<div style={{ color: '#ecf0f1' }}>Tokens Saved</div>
					<div style={{ fontSize: '10px', color: '#bdc3c7' }}>Per query set</div>
				</div>
				<div>
					<div style={{ fontSize: '18px', fontWeight: 'bold', color: '#9b59b6', marginBottom: '4px' }}>75%</div>
					<div style={{ color: '#ecf0f1' }}>Code Reduction</div>
					<div style={{ fontSize: '10px', color: '#bdc3c7' }}>Cleaner queries</div>
				</div>
			</div>
		</div>
	);
};

export const sqliteViewOptimizationSequence: InteractionSequence = {
	title: 'Optimizing with SQL Views',
	
	subtitle: (state: InteractionState) => {
		if (state.activeStates.includes('viewCreation')) {
			return 'Creating reusable infrastructure';
		} else if (state.activeStates.includes('multipleQueries')) {
			return 'Running multiple efficient queries';
		} else if (state.activeStates.includes('performanceComparison')) {
			return 'Performance comparison';
		}
		return 'Creating reusable infrastructure for multiple queries';
	},

	messageTypes: sqliteViewMessageTypes,
	
	states: [
		// View creation phase (frames 90-360 → 60-330)
		createState('container', 0, 600),
		createState('viewCreation', 60, 270),
		createState('llmThinking', 90, 60),
		createState('viewQuery', 150, 90),
		createState('viewSuccess', 240, 90),
		createState('viewBenefits', 300, 60),
		
		// Multiple queries phase (frames 360-780 → 330-750)
		createState('multipleQueries', 330, 420),
		createState('query1', 360, 60),
		createState('result1', 420, 60),
		createState('query2', 480, 60),
		createState('result2', 540, 60),
		createState('query3', 600, 60),
		createState('result3', 660, 60),
		createState('query4', 720, 60),
		createState('result4', 780, 60),
		createState('querySummary', 840, 60),
		
		// Performance comparison phase (frames 780-1200 → 900-1320)
		createState('performanceComparison', 900, 420),
		createState('beforeComparison', 930, 90),
		createState('afterComparison', 1020, 90),
		createState('metricsComparison', 1110, 90),
		createState('benefitsList', 1200, 90),
		createState('finalMessage', 1290, 90),
	],

	messages: [
		// LLM thinking about optimization
		createMessage(
			'llm-optimization-thought',
			'assistant',
			'"I\'ll be making multiple customer queries. Let me create a view to pre-join the tables and optimize future queries."',
			['llmThinking', 'viewCreation', 'multipleQueries', 'performanceComparison'],
			{ column: 'left' }
		),

		// View creation
		createMessage(
			'create-view-query',
			'view_creation',
			(state: InteractionState) => (
				<CreateViewSQL
					isVisible={state.activeStates.includes('viewQuery')}
				/>
			),
			['viewQuery', 'viewSuccess', 'multipleQueries', 'performanceComparison'],
			{ column: 'left', isReactContent: true }
		),

		// View creation success
		createMessage(
			'view-success',
			'infrastructure',
			'✅ View "customer_orders_view" created successfully!\n\n🚀 One-time setup complete - ready for efficient queries',
			['viewSuccess', 'viewBenefits', 'multipleQueries', 'performanceComparison'],
			{ column: 'right' }
		),

		// Benefits explanation
		createMessage(
			'view-benefits',
			'optimization',
			'💎 Smart Infrastructure Benefits:\n• Pre-joins customers & orders\n• No repeated JOIN logic needed\n• Clean, simple query syntax\n• Reusable for multiple queries',
			['viewBenefits', 'multipleQueries', 'performanceComparison'],
			{ column: 'right' }
		),

		// Multiple queries execution
		createMessage(
			'multiple-queries-execution',
			'efficient_query',
			(state: InteractionState) => {
				let activeQueries = 0;
				if (state.activeStates.includes('result1')) activeQueries = 1;
				if (state.activeStates.includes('result2')) activeQueries = 2;
				if (state.activeStates.includes('result3')) activeQueries = 3;
				if (state.activeStates.includes('result4')) activeQueries = 4;
				
				return (
					<MultipleQueriesWidget
						activeQueries={activeQueries}
						isVisible={state.activeStates.includes('multipleQueries')}
					/>
				);
			},
			['query1', 'result1', 'query2', 'result2', 'query3', 'result3', 'query4', 'result4', 'querySummary', 'performanceComparison'],
			{ column: 'left', isReactContent: true }
		),

		// Query summary
		createMessage(
			'query-summary',
			'optimization',
			'⚡ 4 Queries, 0 JOINs Needed!\n\n🚀 Fast: Pre-joined data\n🎯 Simple: Clean syntax\n🔄 Reusable: One view, many uses',
			['querySummary', 'performanceComparison'],
			{ column: 'right' }
		),

		// Performance comparison
		createMessage(
			'performance-comparison-visual',
			'performance_comparison',
			(state: InteractionState) => (
				<PerformanceComparison
					isVisible={state.activeStates.includes('beforeComparison') || state.activeStates.includes('afterComparison')}
				/>
			),
			['beforeComparison', 'afterComparison', 'metricsComparison', 'benefitsList', 'finalMessage'],
			{ column: 'left', isReactContent: true }
		),

		// Performance metrics
		createMessage(
			'performance-metrics',
			'performance_comparison',
			(state: InteractionState) => (
				<PerformanceMetrics
					isVisible={state.activeStates.includes('metricsComparison')}
				/>
			),
			['metricsComparison', 'benefitsList', 'finalMessage'],
			{ column: 'right', isReactContent: true }
		),

		// Benefits summary
		createMessage(
			'benefits-summary',
			'infrastructure',
			'🏗️ Infrastructure: Reusable views\n⚡ Performance: Faster execution\n🧹 Cleaner Code: Simpler queries\n💰 Token Savings: Reduced costs',
			['benefitsList', 'finalMessage'],
			{ column: 'right' }
		),

		// Final message
		createMessage(
			'final-optimization-message',
			'optimization',
			'🚀 Smart infrastructure = Multiple efficient queries!\n\nViews provide the foundation for scalable, maintainable database operations.',
			['finalMessage'],
			{ column: 'right' }
		),
	],

	overlays: [
		{
			id: 'step-indicator',
			content: (state: InteractionState) => {
				let step = '';
				let color = '#9b59b6';
				
				if (state.activeStates.includes('viewCreation')) {
					step = 'Step 1: Creating a reusable SQL view';
					color = '#9b59b6';
				} else if (state.activeStates.includes('multipleQueries')) {
					step = 'Step 2: Running multiple efficient queries';
					color = '#3498db';
				} else if (state.activeStates.includes('performanceComparison')) {
					step = 'Step 3: Performance comparison';
					color = '#e74c3c';
				}
				
				if (!step) return '';
				
				return `
					<div style="
						background-color: rgba(142, 68, 173, 0.9);
						color: white;
						padding: 12px 24px;
						border-radius: 20px;
						font-size: 16px;
						font-weight: bold;
						box-shadow: 0 4px 15px rgba(0,0,0,0.3);
						border: 2px solid ${color};
					">
						${step}
					</div>
				`;
			},
			position: {
				top: '8%',
				left: '50%',
				transform: 'translateX(-50%)',
			},
			visibleStates: ['viewCreation', 'multipleQueries', 'performanceComparison'],
		},
		{
			id: 'progress-indicator',
			content: (state: InteractionState) => {
				const queries = ['Count', 'Sum', 'Average', 'Latest'];
				let completedQueries = 0;
				if (state.activeStates.includes('result1')) completedQueries = 1;
				if (state.activeStates.includes('result2')) completedQueries = 2;
				if (state.activeStates.includes('result3')) completedQueries = 3;
				if (state.activeStates.includes('result4')) completedQueries = 4;
				
				if (!state.activeStates.includes('multipleQueries')) return '';
				
				return `
					<div style="
						background-color: rgba(255,255,255,0.1);
						border-radius: 10px;
						padding: 12px;
						color: white;
						font-size: 12px;
						text-align: center;
						min-width: 80px;
					">
						<div style="font-weight: bold; margin-bottom: 8px;">Progress</div>
						${queries.map((query, idx) => `
							<div style="opacity: ${idx < completedQueries ? 1 : 0.3}; margin-bottom: 3px;">
								✅ ${query}
							</div>
						`).join('')}
					</div>
				`;
			},
			position: {
				top: '20%',
				right: '5%',
			},
			visibleStates: ['multipleQueries'],
		},
		{
			id: 'efficiency-stats',
			content: () => `
				<div style="
					background-color: rgba(155, 89, 182, 0.1);
					border-radius: 12px;
					padding: 15px;
					color: white;
					font-size: 12px;
					text-align: center;
					display: flex;
					gap: 25px;
					align-items: center;
					box-shadow: 0 3px 10px rgba(0,0,0,0.3);
					border: 2px solid rgba(155, 89, 182, 0.5);
				">
					<div>
						<div style="font-size: 16px; margin-bottom: 3px;">🏗️</div>
						<div style="font-weight: bold;">Infrastructure</div>
						<div style="font-size: 10px;">Reusable views</div>
					</div>
					<div>
						<div style="font-size: 16px; margin-bottom: 3px;">⚡</div>
						<div style="font-weight: bold;">Performance</div>
						<div style="font-size: 10px;">Faster execution</div>
					</div>
					<div>
						<div style="font-size: 16px; margin-bottom: 3px;">🧹</div>
						<div style="font-weight: bold;">Cleaner Code</div>
						<div style="font-size: 10px;">Simpler queries</div>
					</div>
					<div>
						<div style="font-size: 16px; margin-bottom: 3px;">💰</div>
						<div style="font-weight: bold;">Token Savings</div>
						<div style="font-size: 10px;">Reduced costs</div>
					</div>
				</div>
			`,
			position: {
				bottom: '8%',
				left: '50%',
				transform: 'translateX(-50%)',
			},
			visibleStates: ['benefitsList', 'finalMessage'],
		},
	],

	layout: {
		columns: 2,
		autoFill: false,
		maxMessagesPerColumn: 8,
	},
	
	tokenCounter: {
		enabled: true,
		initialTokens: 200,
		maxTokens: 128000,
		stateTokenCounts: {
			'llmThinking': 230,
			'viewQuery': 350,
			'viewSuccess': 370,
			'result1': 390,
			'result2': 410,
			'result3': 430,
			'result4': 450,
			'querySummary': 470,
			'metricsComparison': 490,
			'finalMessage': 520,
		},
		optimizedStates: ['querySummary', 'metricsComparison', 'finalMessage'],
	},
}; 