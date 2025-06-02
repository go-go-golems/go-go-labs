import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	spring,
	useCurrentFrame,
	useVideoConfig,
} from 'remotion';

export const ViewCreationSequence: React.FC = () => {
	const frame = useCurrentFrame();
	const {fps} = useVideoConfig();

	const stepOpacity = interpolate(frame, [0, 30], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const llmOpacity = interpolate(frame, [0, 30], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const thoughtOpacity = interpolate(frame, [40, 70], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const queryOpacity = interpolate(frame, [90, 120], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const arrowOpacity = interpolate(frame, [140, 170], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const arrowLength = interpolate(frame, [140, 170], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const dbOpacity = interpolate(frame, [160, 190], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const successOpacity = interpolate(frame, [210, 240], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const benefitsOpacity = interpolate(frame, [250, 270], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const thoughtScale = spring({
		frame: frame - 40,
		fps,
		config: {
			damping: 8,
			stiffness: 80,
		},
	});

	const queryScale = spring({
		frame: frame - 90,
		fps,
		config: {
			damping: 8,
			stiffness: 80,
		},
	});

	const successScale = spring({
		frame: frame - 210,
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
					top: '15%',
					left: '50%',
					transform: 'translateX(-50%)',
					color: 'white',
					fontSize: '28px',
					fontWeight: 'bold',
					opacity: stepOpacity,
				}}
			>
				Step 1: Creating a reusable SQL view
			</div>

			{/* LLM */}
			<div
				style={{
					position: 'absolute',
					top: '25%',
					left: '15%',
					opacity: llmOpacity,
				}}
			>
				<div
					style={{
						width: '120px',
						height: '120px',
						borderRadius: '20px',
						backgroundColor: '#9b59b6',
						display: 'flex',
						alignItems: 'center',
						justifyContent: 'center',
						fontSize: '60px',
						color: 'white',
						boxShadow: '0 6px 20px rgba(0,0,0,0.2)',
						transform: frame > 70 && frame < 180 ? `scale(${1 + 0.03 * Math.sin(frame * 0.2)})` : 'scale(1)',
					}}
				>
					{frame > 70 && frame < 180 ? '💡' : '🧠'}
				</div>
				<div
					style={{
						textAlign: 'center',
						color: 'white',
						marginTop: '10px',
						fontSize: '18px',
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
					left: '35%',
					opacity: thoughtOpacity,
					transform: `scale(${thoughtScale})`,
				}}
			>
				<div
					style={{
						backgroundColor: 'rgba(255,255,255,0.95)',
						borderRadius: '20px',
						padding: '20px',
						boxShadow: '0 6px 20px rgba(0,0,0,0.15)',
						fontSize: '16px',
						color: '#2c3e50',
						maxWidth: '300px',
						lineHeight: 1.4,
					}}
				>
					"I'll be making multiple customer queries. Let me create a view to pre-join the tables and optimize future queries."
				</div>
			</div>

			{/* CREATE VIEW SQL Query */}
			<div
				style={{
					position: 'absolute',
					top: '50%',
					left: '10%',
					opacity: queryOpacity,
					transform: `scale(${queryScale})`,
				}}
			>
				<div
					style={{
						backgroundColor: '#2c3e50',
						borderRadius: '15px',
						padding: '25px',
						boxShadow: '0 8px 25px rgba(0,0,0,0.4)',
						fontSize: '16px',
						color: '#ecf0f1',
						fontFamily: 'monospace',
						maxWidth: '700px',
						border: '3px solid #9b59b6',
					}}
				>
					<div style={{color: '#9b59b6', fontWeight: 'bold', marginBottom: '15px', fontSize: '18px'}}>
						🏗️ Tool Call: Create Infrastructure
					</div>
					<div style={{color: '#e74c3c'}}>sqlite_query(</div>
					<div style={{paddingLeft: '20px', color: '#f39c12', lineHeight: 1.8}}>
						"CREATE VIEW customer_orders_view AS<br/>
						SELECT <br/>
						&nbsp;&nbsp;c.id as customer_id,<br/>
						&nbsp;&nbsp;c.name as customer_name,<br/>
						&nbsp;&nbsp;c.email,<br/>
						&nbsp;&nbsp;o.id as order_id,<br/>
						&nbsp;&nbsp;o.amount,<br/>
						&nbsp;&nbsp;o.order_date<br/>
						FROM customers c<br/>
						JOIN orders o ON c.id = o.customer_id;"
					</div>
					<div style={{color: '#e74c3c'}}>)</div>
					
					<div style={{marginTop: '15px', padding: '10px', backgroundColor: 'rgba(155, 89, 182, 0.2)', borderRadius: '8px', fontSize: '14px'}}>
						<div style={{color: '#9b59b6', fontWeight: 'bold'}}>💎 Smart infrastructure:</div>
						<div style={{color: '#ecf0f1'}}>• Pre-joins customers & orders</div>
						<div style={{color: '#ecf0f1'}}>• Meaningful column names</div>
						<div style={{color: '#ecf0f1'}}>• Reusable for multiple queries</div>
						<div style={{color: '#ecf0f1'}}>• No repeated JOIN logic needed</div>
					</div>
				</div>
			</div>

			{/* Arrow */}
			<div
				style={{
					position: 'absolute',
					top: '78%',
					left: '65%',
					opacity: arrowOpacity,
				}}
			>
				<svg width="150" height="40" viewBox="0 0 150 40">
					<defs>
						<linearGradient id="viewArrowGradient" x1="0%" y1="0%" x2="100%" y2="0%">
							<stop offset="0%" stopColor="#9b59b6" />
							<stop offset="100%" stopColor="#8e44ad" />
						</linearGradient>
					</defs>
					<path
						d={`M 10 20 L ${10 + 110 * arrowLength} 20`}
						stroke="url(#viewArrowGradient)"
						strokeWidth="5"
						strokeLinecap="round"
						fill="none"
					/>
					<polygon
						points={`${10 + 110 * arrowLength},20 ${10 + 110 * arrowLength - 15},12 ${10 + 110 * arrowLength - 15},28`}
						fill="url(#viewArrowGradient)"
						opacity={arrowLength}
					/>
				</svg>
			</div>

			{/* Database */}
			<div
				style={{
					position: 'absolute',
					top: '73%',
					right: '15%',
					opacity: dbOpacity,
				}}
			>
				<div
					style={{
						width: '120px',
						height: '120px',
						borderRadius: '20px',
						backgroundColor: '#16a085',
						display: 'flex',
						alignItems: 'center',
						justifyContent: 'center',
						fontSize: '60px',
						color: 'white',
						boxShadow: '0 6px 20px rgba(0,0,0,0.2)',
					}}
				>
					🗃️
				</div>
				<div
					style={{
						textAlign: 'center',
						color: 'white',
						marginTop: '10px',
						fontSize: '18px',
						fontWeight: 'bold',
					}}
				>
					SQLite DB
				</div>
			</div>

			{/* Success Message */}
			<div
				style={{
					position: 'absolute',
					bottom: '20%',
					left: '50%',
					transform: `translateX(-50%) scale(${successScale})`,
					opacity: successOpacity,
				}}
			>
				<div
					style={{
						backgroundColor: '#27ae60',
						borderRadius: '20px',
						padding: '20px 30px',
						boxShadow: '0 8px 25px rgba(0,0,0,0.3)',
						color: 'white',
						fontSize: '18px',
						textAlign: 'center',
						border: '3px solid #2ecc71',
					}}
				>
					<div style={{fontSize: '24px', marginBottom: '10px'}}>✅</div>
					<div style={{fontWeight: 'bold'}}>
						View "customer_orders_view" created successfully!
					</div>
				</div>
			</div>

			{/* Benefits callout */}
			{benefitsOpacity > 0 && (
				<div
					style={{
						position: 'absolute',
						bottom: '8%',
						left: '50%',
						transform: 'translateX(-50%)',
						opacity: benefitsOpacity,
					}}
				>
					<div
						style={{
							backgroundColor: 'rgba(255,255,255,0.1)',
							borderRadius: '15px',
							padding: '15px 25px',
							color: 'white',
							fontSize: '16px',
							textAlign: 'center',
							border: '2px solid rgba(255,255,255,0.3)',
							display: 'flex',
							gap: '30px',
							alignItems: 'center',
						}}
					>
						<div>
							<div style={{fontWeight: 'bold'}}>🚀 One-time setup</div>
							<div style={{fontSize: '14px'}}>CREATE VIEW once</div>
						</div>
						<div>
							<div style={{fontWeight: 'bold'}}>⚡ Fast queries</div>
							<div style={{fontSize: '14px'}}>No repeated JOINs</div>
						</div>
						<div>
							<div style={{fontWeight: 'bold'}}>🎯 Clean syntax</div>
							<div style={{fontSize: '14px'}}>Simple SELECT statements</div>
						</div>
					</div>
				</div>
			)}
		</AbsoluteFill>
	);
};
