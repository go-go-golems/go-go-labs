import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	spring,
	useCurrentFrame,
	useVideoConfig,
} from 'remotion';

export const SQLiteTableExplorationSequence: React.FC = () => {
	const frame = useCurrentFrame();
	const {fps} = useVideoConfig();

	const stepOpacity = interpolate(frame, [0, 30], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const llmOpacity = interpolate(frame, [0, 30], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const thoughtOpacity = interpolate(frame, [30, 60], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const query1Opacity = interpolate(frame, [70, 100], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const response1Opacity = interpolate(frame, [110, 140], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const query2Opacity = interpolate(frame, [150, 180], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const thoughtScale = spring({
		frame: frame - 30,
		fps,
		config: {
			damping: 8,
			stiffness: 80,
		},
	});

	const query1Scale = spring({
		frame: frame - 70,
		fps,
		config: {
			damping: 8,
			stiffness: 80,
		},
	});

	const query2Scale = spring({
		frame: frame - 150,
		fps,
		config: {
			damping: 8,
			stiffness: 80,
		},
	});

	return (
		<AbsoluteFill
			style={{
				background: 'linear-gradient(135deg, #2c3e50 0%, #34495e 100%)',
				fontFamily: 'Arial, sans-serif',
			}}
		>
			{/* Step indicator */}
			<div
				style={{
					position: 'absolute',
					top: '15%',
					left: '50%',
					transform: 'translateX(-50%)',
					color: 'white',
					fontSize: '28px',
					fontWeight: 'bold',
					opacity: stepOpacity,
				}}
			>
				Step 3: Exploring table structures
			</div>

			{/* LLM */}
			<div
				style={{
					position: 'absolute',
					top: '25%',
					left: '10%',
					opacity: llmOpacity,
				}}
			>
				<div
					style={{
						width: '100px',
						height: '100px',
						borderRadius: '20px',
						backgroundColor: '#9b59b6',
						display: 'flex',
						alignItems: 'center',
						justifyContent: 'center',
						fontSize: '50px',
						color: 'white',
						boxShadow: '0 6px 20px rgba(0,0,0,0.2)',
						transform: frame > 60 ? `scale(${1 + 0.02 * Math.sin(frame * 0.15)})` : 'scale(1)',
					}}
				>
					🧠
				</div>
				<div
					style={{
						textAlign: 'center',
						color: 'white',
						marginTop: '5px',
						fontSize: '16px',
						fontWeight: 'bold',
					}}
				>
					LLM
				</div>
			</div>

			{/* Thought bubble */}
			<div
				style={{
					position: 'absolute',
					top: '20%',
					left: '25%',
					opacity: thoughtOpacity,
					transform: `scale(${thoughtScale})`,
				}}
			>
				<div
					style={{
						backgroundColor: 'rgba(255,255,255,0.95)',
						borderRadius: '15px',
						padding: '15px',
						boxShadow: '0 6px 20px rgba(0,0,0,0.15)',
						fontSize: '14px',
						color: '#2c3e50',
						maxWidth: '250px',
						lineHeight: 1.3,
					}}
				>
					"I found customers & orders tables. Let me check their structure to understand how to join them."
				</div>
			</div>

			{/* First Query - Customers table structure */}
			<div
				style={{
					position: 'absolute',
					top: '40%',
					left: '5%',
					opacity: query1Opacity,
					transform: `scale(${query1Scale})`,
				}}
			>
				<div
					style={{
						backgroundColor: '#2c3e50',
						borderRadius: '12px',
						padding: '15px',
						boxShadow: '0 6px 20px rgba(0,0,0,0.3)',
						fontSize: '14px',
						color: '#ecf0f1',
						fontFamily: 'monospace',
						maxWidth: '280px',
						border: '2px solid #3498db',
					}}
				>
					<div style={{color: '#3498db', fontWeight: 'bold', marginBottom: '8px'}}>
						🔍 Tool Call #2:
					</div>
					<div style={{color: '#e74c3c'}}>sqlite_query(</div>
					<div style={{paddingLeft: '10px', color: '#f39c12', fontSize: '12px'}}>
						"PRAGMA table_info(customers);"
					</div>
					<div style={{color: '#e74c3c'}}>)</div>
				</div>
			</div>

			{/* First Response */}
			{response1Opacity > 0 && (
				<div
					style={{
						position: 'absolute',
						top: '40%',
						left: '35%',
						opacity: response1Opacity,
					}}
				>
					<div
						style={{
							backgroundColor: 'rgba(255,255,255,0.95)',
							borderRadius: '10px',
							padding: '12px',
							boxShadow: '0 4px 15px rgba(0,0,0,0.15)',
							fontSize: '12px',
							color: '#2c3e50',
							fontFamily: 'monospace',
							maxWidth: '200px',
						}}
					>
						<div style={{fontWeight: 'bold', marginBottom: '8px', color: '#16a085'}}>
							customers:
						</div>
						<div>• id (INTEGER)</div>
						<div>• name (TEXT)</div>
						<div>• email (TEXT)</div>
						<div>• created_at (TEXT)</div>
					</div>
				</div>
			)}

			{/* Second Query - Orders table structure */}
			<div
				style={{
					position: 'absolute',
					top: '65%',
					left: '5%',
					opacity: query2Opacity,
					transform: `scale(${query2Scale})`,
				}}
			>
				<div
					style={{
						backgroundColor: '#2c3e50',
						borderRadius: '12px',
						padding: '15px',
						boxShadow: '0 6px 20px rgba(0,0,0,0.3)',
						fontSize: '14px',
						color: '#ecf0f1',
						fontFamily: 'monospace',
						maxWidth: '280px',
						border: '2px solid #3498db',
					}}
				>
					<div style={{color: '#3498db', fontWeight: 'bold', marginBottom: '8px'}}>
						🔍 Tool Call #3:
					</div>
					<div style={{color: '#e74c3c'}}>sqlite_query(</div>
					<div style={{paddingLeft: '10px', color: '#f39c12', fontSize: '12px'}}>
						"PRAGMA table_info(orders);"
					</div>
					<div style={{color: '#e74c3c'}}>)</div>
				</div>
			</div>

			{/* Second Response */}
			{query2Opacity > 0.5 && (
				<div
					style={{
						position: 'absolute',
						top: '65%',
						left: '35%',
						opacity: interpolate(frame, [165, 180], [0, 1], {extrapolateRight: 'clamp'}),
					}}
				>
					<div
						style={{
							backgroundColor: 'rgba(255,255,255,0.95)',
							borderRadius: '10px',
							padding: '12px',
							boxShadow: '0 4px 15px rgba(0,0,0,0.15)',
							fontSize: '12px',
							color: '#2c3e50',
							fontFamily: 'monospace',
							maxWidth: '200px',
						}}
					>
						<div style={{fontWeight: 'bold', marginBottom: '8px', color: '#16a085'}}>
							orders:
						</div>
						<div>• id (INTEGER)</div>
						<div style={{color: '#e74c3c', fontWeight: 'bold'}}>• customer_id (INTEGER)</div>
						<div>• amount (REAL)</div>
						<div>• order_date (TEXT)</div>
					</div>
				</div>
			)}

			{/* Analysis insight */}
			{query2Opacity > 0.5 && (
				<div
					style={{
						position: 'absolute',
						top: '40%',
						right: '10%',
						opacity: interpolate(frame, [170, 180], [0, 1], {extrapolateRight: 'clamp'}),
					}}
				>
					<div
						style={{
							backgroundColor: 'rgba(155, 89, 182, 0.9)',
							borderRadius: '15px',
							padding: '20px',
							color: 'white',
							fontSize: '16px',
							maxWidth: '250px',
							textAlign: 'center',
						}}
					>
						<div style={{fontSize: '24px', marginBottom: '10px'}}>💡</div>
						<div style={{fontWeight: 'bold', marginBottom: '10px'}}>
							Schema Analysis Complete!
						</div>
						<div style={{fontSize: '14px', lineHeight: 1.4}}>
							Now I understand:<br/>
							• customers.id links to orders.customer_id<br/>
							• I can filter by customer name and date
						</div>
					</div>
				</div>
			)}
		</AbsoluteFill>
	);
};
