import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	spring,
	useCurrentFrame,
	useVideoConfig,
} from 'remotion';

export const SQLiteSchemaDiscoverySequence: React.FC = () => {
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

	const queryOpacity = interpolate(frame, [70, 100], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const arrowOpacity = interpolate(frame, [100, 130], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const arrowLength = interpolate(frame, [100, 130], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const dbOpacity = interpolate(frame, [120, 150], [0, 1], {
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

	const queryScale = spring({
		frame: frame - 70,
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
					top: '20%',
					left: '50%',
					transform: 'translateX(-50%)',
					color: 'white',
					fontSize: '28px',
					fontWeight: 'bold',
					opacity: stepOpacity,
				}}
			>
				Step 2: LLM explores database schema
			</div>

			{/* LLM */}
			<div
				style={{
					position: 'absolute',
					top: '30%',
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
						transform: frame > 60 ? `scale(${1 + 0.03 * Math.sin(frame * 0.2)})` : 'scale(1)',
					}}
				>
					🧠
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
					top: '25%',
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
						maxWidth: '280px',
						lineHeight: 1.4,
					}}
				>
					"I need to find customer orders, but I don't know the database structure. Let me explore the schema first."
				</div>
			</div>

			{/* SQL Query */}
			<div
				style={{
					position: 'absolute',
					top: '55%',
					left: '20%',
					opacity: queryOpacity,
					transform: `scale(${queryScale})`,
				}}
			>
				<div
					style={{
						backgroundColor: '#2c3e50',
						borderRadius: '15px',
						padding: '20px',
						boxShadow: '0 6px 20px rgba(0,0,0,0.3)',
						fontSize: '16px',
						color: '#ecf0f1',
						fontFamily: 'monospace',
						maxWidth: '350px',
						border: '2px solid #3498db',
					}}
				>
					<div style={{color: '#3498db', fontWeight: 'bold', marginBottom: '10px'}}>
						🔍 Tool Call #1:
					</div>
					<div style={{color: '#e74c3c', marginBottom: '5px'}}>sqlite_query(</div>
					<div style={{paddingLeft: '20px', color: '#f39c12'}}>
						"SELECT name FROM sqlite_master WHERE type='table';"
					</div>
					<div style={{color: '#e74c3c'}}>)</div>
				</div>
			</div>

			{/* Arrow */}
			<div
				style={{
					position: 'absolute',
					top: '65%',
					left: '55%',
					opacity: arrowOpacity,
				}}
			>
				<svg width="120" height="40" viewBox="0 0 120 40">
					<defs>
						<linearGradient id="schemaArrowGradient" x1="0%" y1="0%" x2="100%" y2="0%">
							<stop offset="0%" stopColor="#3498db" />
							<stop offset="100%" stopColor="#2980b9" />
						</linearGradient>
					</defs>
					<path
						d={`M 10 20 L ${10 + 80 * arrowLength} 20`}
						stroke="url(#schemaArrowGradient)"
						strokeWidth="4"
						strokeLinecap="round"
						fill="none"
					/>
					<polygon
						points={`${10 + 80 * arrowLength},20 ${10 + 80 * arrowLength - 12},15 ${10 + 80 * arrowLength - 12},25`}
						fill="url(#schemaArrowGradient)"
						opacity={arrowLength}
					/>
				</svg>
			</div>

			{/* Database */}
			<div
				style={{
					position: 'absolute',
					top: '60%',
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

			{/* Database Response */}
			{dbOpacity > 0.5 && (
				<div
					style={{
						position: 'absolute',
						bottom: '15%',
						left: '50%',
						transform: 'translateX(-50%)',
						opacity: interpolate(frame, [135, 150], [0, 1], {extrapolateRight: 'clamp'}),
					}}
				>
					<div
						style={{
							backgroundColor: 'rgba(255,255,255,0.95)',
							borderRadius: '15px',
							padding: '20px',
							boxShadow: '0 6px 20px rgba(0,0,0,0.15)',
							fontSize: '16px',
							color: '#2c3e50',
							fontFamily: 'monospace',
							textAlign: 'center',
						}}
					>
						<div style={{fontWeight: 'bold', marginBottom: '10px', color: '#16a085'}}>
							📋 Schema Discovery Result:
						</div>
						<div style={{display: 'flex', gap: '20px', justifyContent: 'center'}}>
							<div style={{padding: '10px', backgroundColor: '#ecf0f1', borderRadius: '8px'}}>
								customers
							</div>
							<div style={{padding: '10px', backgroundColor: '#ecf0f1', borderRadius: '8px'}}>
								orders
							</div>
							<div style={{padding: '10px', backgroundColor: '#ecf0f1', borderRadius: '8px'}}>
								products
							</div>
						</div>
					</div>
				</div>
			)}
		</AbsoluteFill>
	);
};
