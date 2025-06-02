import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	spring,
	useCurrentFrame,
	useVideoConfig,
} from 'remotion';

export const PerformanceComparisonSequence: React.FC = () => {
	const frame = useCurrentFrame();
	const {fps} = useVideoConfig();

	const stepOpacity = interpolate(frame, [0, 30], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const beforeOpacity = interpolate(frame, [50, 80], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const afterOpacity = interpolate(frame, [120, 150], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const comparisonOpacity = interpolate(frame, [200, 230], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const benefitsOpacity = interpolate(frame, [280, 320], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const finalMessageOpacity = interpolate(frame, [360, 400], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const beforeScale = spring({
		frame: frame - 50,
		fps,
		config: {
			damping: 8,
			stiffness: 80,
		},
	});

	const afterScale = spring({
		frame: frame - 120,
		fps,
		config: {
			damping: 8,
			stiffness: 80,
		},
	});

	const comparisonScale = spring({
		frame: frame - 200,
		fps,
		config: {
			damping: 10,
			stiffness: 100,
		},
	});

	const finalMessageScale = spring({
		frame: frame - 360,
		fps,
		config: {
			damping: 10,
			stiffness: 100,
		},
	});

	return (
		<AbsoluteFill
			style={{
				background: 'linear-gradient(135deg, #8e44ad 0%, #9b59b6 100%)',
				fontFamily: 'Arial, sans-serif',
			}}
		>
			{/* Step indicator */}
			<div
				style={{
					position: 'absolute',
					top: '8%',
					left: '50%',
					transform: 'translateX(-50%)',
					color: 'white',
					fontSize: '28px',
					fontWeight: 'bold',
					opacity: stepOpacity,
				}}
			>
				Step 3: Performance comparison
			</div>

			{/* Before: Without Views */}
			<div
				style={{
					position: 'absolute',
					top: '20%',
					left: '5%',
					opacity: beforeOpacity,
					transform: `scale(${beforeScale})`,
				}}
			>
				<div
					style={{
						backgroundColor: '#e74c3c',
						borderRadius: '15px',
						padding: '20px',
						boxShadow: '0 8px 25px rgba(0,0,0,0.3)',
						color: 'white',
						maxWidth: '400px',
						border: '3px solid #c0392b',
					}}
				>
					<div style={{fontSize: '20px', fontWeight: 'bold', marginBottom: '15px', textAlign: 'center'}}>
						❌ Before: Without Views
					</div>
					<div style={{fontSize: '14px', lineHeight: 1.6, fontFamily: 'monospace'}}>
						<div style={{marginBottom: '10px'}}>
							<strong>Query 1:</strong> SELECT COUNT(*) FROM orders o<br/>
							JOIN customers c ON o.customer_id = c.id<br/>
							WHERE c.name = 'John Smith';
						</div>
						<div style={{marginBottom: '10px'}}>
							<strong>Query 2:</strong> SELECT SUM(amount) FROM orders o<br/>
							JOIN customers c ON o.customer_id = c.id<br/>
							WHERE c.name = 'John Smith';
						</div>
						<div style={{marginBottom: '10px'}}>
							<strong>Query 3:</strong> SELECT AVG(amount) FROM orders o<br/>
							JOIN customers c ON o.customer_id = c.id<br/>
							WHERE c.name = 'John Smith';
						</div>
						<div>
							<strong>Query 4:</strong> SELECT MAX(order_date) FROM orders o<br/>
							JOIN customers c ON o.customer_id = c.id<br/>
							WHERE c.name = 'John Smith';
						</div>
					</div>
					<div style={{marginTop: '15px', padding: '10px', backgroundColor: 'rgba(192, 57, 43, 0.3)', borderRadius: '8px', textAlign: 'center'}}>
						<div style={{fontWeight: 'bold', color: '#ecf0f1'}}>4 queries × 4 JOINs = 16 JOIN operations</div>
					</div>
				</div>
			</div>

			{/* After: With Views */}
			<div
				style={{
					position: 'absolute',
					top: '20%',
					right: '5%',
					opacity: afterOpacity,
					transform: `scale(${afterScale})`,
				}}
			>
				<div
					style={{
						backgroundColor: '#27ae60',
						borderRadius: '15px',
						padding: '20px',
						boxShadow: '0 8px 25px rgba(0,0,0,0.3)',
						color: 'white',
						maxWidth: '400px',
						border: '3px solid #2ecc71',
					}}
				>
					<div style={{fontSize: '20px', fontWeight: 'bold', marginBottom: '15px', textAlign: 'center'}}>
						✅ After: With Views
					</div>
					<div style={{fontSize: '14px', lineHeight: 1.6, fontFamily: 'monospace'}}>
						<div style={{marginBottom: '10px', backgroundColor: 'rgba(46, 204, 113, 0.2)', padding: '8px', borderRadius: '5px'}}>
							<strong>Setup (once):</strong> CREATE VIEW customer_orders_view...
						</div>
						<div style={{marginBottom: '8px'}}>
							<strong>Query 1:</strong> SELECT COUNT(*) FROM customer_orders_view<br/>
							WHERE customer_name = 'John Smith';
						</div>
						<div style={{marginBottom: '8px'}}>
							<strong>Query 2:</strong> SELECT SUM(amount) FROM customer_orders_view<br/>
							WHERE customer_name = 'John Smith';
						</div>
						<div style={{marginBottom: '8px'}}>
							<strong>Query 3:</strong> SELECT AVG(amount) FROM customer_orders_view<br/>
							WHERE customer_name = 'John Smith';
						</div>
						<div>
							<strong>Query 4:</strong> SELECT MAX(order_date) FROM customer_orders_view<br/>
							WHERE customer_name = 'John Smith';
						</div>
					</div>
					<div style={{marginTop: '15px', padding: '10px', backgroundColor: 'rgba(46, 204, 113, 0.3)', borderRadius: '8px', textAlign: 'center'}}>
						<div style={{fontWeight: 'bold', color: '#ecf0f1'}}>1 view setup + 4 simple queries = 1 JOIN total</div>
					</div>
				</div>
			</div>

			{/* Comparison metrics */}
			<div
				style={{
					position: 'absolute',
					bottom: '35%',
					left: '50%',
					transform: `translateX(-50%) scale(${comparisonScale})`,
					opacity: comparisonOpacity,
				}}
			>
				<div
					style={{
						backgroundColor: 'rgba(44, 62, 80, 0.95)',
						borderRadius: '20px',
						padding: '25px',
						color: 'white',
						fontSize: '16px',
						textAlign: 'center',
						boxShadow: '0 8px 25px rgba(0,0,0,0.3)',
						border: '2px solid #9b59b6',
						minWidth: '700px',
					}}
				>
					<div style={{fontSize: '24px', marginBottom: '20px', fontWeight: 'bold'}}>📊 Performance Impact</div>
					<div style={{display: 'grid', gridTemplateColumns: '1fr 1fr 1fr 1fr', gap: '25px', fontSize: '14px'}}>
						<div>
							<div style={{fontSize: '20px', fontWeight: 'bold', color: '#e74c3c', marginBottom: '5px'}}>16x</div>
							<div style={{color: '#ecf0f1'}}>JOIN Operations</div>
							<div style={{fontSize: '12px', color: '#bdc3c7'}}>Without views</div>
						</div>
						<div>
							<div style={{fontSize: '20px', fontWeight: 'bold', color: '#27ae60', marginBottom: '5px'}}>1x</div>
							<div style={{color: '#ecf0f1'}}>JOIN Operation</div>
							<div style={{fontSize: '12px', color: '#bdc3c7'}}>With views</div>
						</div>
						<div>
							<div style={{fontSize: '20px', fontWeight: 'bold', color: '#f39c12', marginBottom: '5px'}}>~500</div>
							<div style={{color: '#ecf0f1'}}>Tokens Saved</div>
							<div style={{fontSize: '12px', color: '#bdc3c7'}}>Per query set</div>
						</div>
						<div>
							<div style={{fontSize: '20px', fontWeight: 'bold', color: '#9b59b6', marginBottom: '5px'}}>75%</div>
							<div style={{color: '#ecf0f1'}}>Code Reduction</div>
							<div style={{fontSize: '12px', color: '#bdc3c7'}}>Cleaner queries</div>
						</div>
					</div>
				</div>
			</div>

			{/* Benefits list */}
			{benefitsOpacity > 0 && (
				<div
					style={{
						position: 'absolute',
						bottom: '18%',
						left: '50%',
						transform: 'translateX(-50%)',
						opacity: benefitsOpacity,
					}}
				>
					<div
						style={{
							backgroundColor: 'rgba(155, 89, 182, 0.1)',
							borderRadius: '15px',
							padding: '20px',
							color: 'white',
							fontSize: '16px',
							textAlign: 'center',
							border: '2px solid rgba(155, 89, 182, 0.5)',
							display: 'flex',
							gap: '40px',
							alignItems: 'center',
						}}
					>
						<div>
							<div style={{fontSize: '20px', marginBottom: '5px'}}>🏗️</div>
							<div style={{fontWeight: 'bold'}}>Infrastructure</div>
							<div style={{fontSize: '14px'}}>Reusable views</div>
						</div>
						<div>
							<div style={{fontSize: '20px', marginBottom: '5px'}}>⚡</div>
							<div style={{fontWeight: 'bold'}}>Performance</div>
							<div style={{fontSize: '14px'}}>Faster execution</div>
						</div>
						<div>
							<div style={{fontSize: '20px', marginBottom: '5px'}}>🧹</div>
							<div style={{fontWeight: 'bold'}}>Cleaner Code</div>
							<div style={{fontSize: '14px'}}>Simpler queries</div>
						</div>
						<div>
							<div style={{fontSize: '20px', marginBottom: '5px'}}>💰</div>
							<div style={{fontWeight: 'bold'}}>Token Savings</div>
							<div style={{fontSize: '14px'}}>Reduced costs</div>
						</div>
					</div>
				</div>
			)}

			{/* Final message */}
			<div
				style={{
					position: 'absolute',
					bottom: '5%',
					left: '50%',
					transform: `translateX(-50%) scale(${finalMessageScale})`,
					opacity: finalMessageOpacity,
				}}
			>
				<div
					style={{
						backgroundColor: 'rgba(155, 89, 182, 0.9)',
						borderRadius: '25px',
						padding: '20px 40px',
						color: 'white',
						fontSize: '22px',
						fontWeight: 'bold',
						textAlign: 'center',
						boxShadow: '0 8px 25px rgba(0,0,0,0.3)',
						border: '3px solid #8e44ad',
					}}
				>
					🚀 Smart infrastructure = Multiple efficient queries!
				</div>
			</div>
		</AbsoluteFill>
	);
};
