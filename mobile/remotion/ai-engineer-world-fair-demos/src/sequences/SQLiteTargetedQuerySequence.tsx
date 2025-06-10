import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	spring,
	useCurrentFrame,
	useVideoConfig,
} from 'remotion';

export const SQLiteTargetedQuerySequence: React.FC = () => {
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

	const responseOpacity = interpolate(frame, [200, 230], [0, 1], {
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

	const responseScale = spring({
		frame: frame - 200,
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
					top: '12%',
					left: '50%',
					transform: 'translateX(-50%)',
					color: 'white',
					fontSize: '28px',
					fontWeight: 'bold',
					opacity: stepOpacity,
				}}
			>
				Step 4: Crafting precise, targeted query
			</div>

			{/* LLM */}
			<div
				style={{
					position: 'absolute',
					top: '22%',
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
						transform: frame > 70 && frame < 180 ? `scale(${1 + 0.05 * Math.sin(frame * 0.3)})` : 'scale(1)',
					}}
				>
					{frame > 70 && frame < 180 ? '🤔' : '🧠'}
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
					top: '18%',
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
					"Perfect! Now I know the schema. I can write a precise query that joins customers and orders, filters by name and date range."
				</div>
			</div>

			{/* Targeted SQL Query */}
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
						maxWidth: '600px',
						border: '3px solid #27ae60',
					}}
				>
					<div style={{color: '#27ae60', fontWeight: 'bold', marginBottom: '15px', fontSize: '18px'}}>
						🎯 Tool Call #4 - Targeted Query:
					</div>
					<div style={{color: '#e74c3c'}}>sqlite_query(</div>
					<div style={{paddingLeft: '20px', color: '#f39c12', lineHeight: 1.6}}>
						"SELECT COUNT(*) as order_count<br/>
						FROM orders o<br/>
						JOIN customers c ON o.customer_id = c.id<br/>
						WHERE c.name = 'John Smith'<br/>
						AND o.order_date LIKE '2024-11%';"
					</div>
					<div style={{color: '#e74c3c'}}>)</div>
					
					<div style={{marginTop: '15px', padding: '10px', backgroundColor: 'rgba(39, 174, 96, 0.2)', borderRadius: '8px', fontSize: '14px'}}>
						<div style={{color: '#27ae60', fontWeight: 'bold'}}>✨ Smart query features:</div>
						<div style={{color: '#ecf0f1'}}>• Joins only needed tables</div>
						<div style={{color: '#ecf0f1'}}>• Filters by exact customer name</div>
						<div style={{color: '#ecf0f1'}}>• Includes date range filter</div>
						<div style={{color: '#ecf0f1'}}>• Returns only the count (no extra data)</div>
					</div>
				</div>
			</div>

			{/* Arrow */}
			<div
				style={{
					position: 'absolute',
					top: '75%',
					left: '65%',
					opacity: arrowOpacity,
				}}
			>
				<svg width="150" height="40" viewBox="0 0 150 40">
					<defs>
						<linearGradient id="targetedArrowGradient" x1="0%" y1="0%" x2="100%" y2="0%">
							<stop offset="0%" stopColor="#27ae60" />
							<stop offset="100%" stopColor="#2ecc71" />
						</linearGradient>
					</defs>
					<path
						d={`M 10 20 L ${10 + 110 * arrowLength} 20`}
						stroke="url(#targetedArrowGradient)"
						strokeWidth="5"
						strokeLinecap="round"
						fill="none"
					/>
					<polygon
						points={`${10 + 110 * arrowLength},20 ${10 + 110 * arrowLength - 15},12 ${10 + 110 * arrowLength - 15},28`}
						fill="url(#targetedArrowGradient)"
						opacity={arrowLength}
					/>
				</svg>
			</div>

			{/* Database */}
			<div
				style={{
					position: 'absolute',
					top: '70%',
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

			{/* Precise Response */}
			<div
				style={{
					position: 'absolute',
					bottom: '8%',
					left: '50%',
					transform: `translateX(-50%) scale(${responseScale})`,
					opacity: responseOpacity,
				}}
			>
				<div
					style={{
						backgroundColor: '#27ae60',
						borderRadius: '20px',
						padding: '25px 40px',
						boxShadow: '0 8px 25px rgba(0,0,0,0.3)',
						color: 'white',
						fontSize: '20px',
						textAlign: 'center',
						border: '3px solid #2ecc71',
					}}
				>
					<div style={{fontSize: '28px', marginBottom: '10px'}}>🎯</div>
					<div style={{fontWeight: 'bold', marginBottom: '10px'}}>
						Precise Result:
					</div>
					<div style={{fontSize: '32px', fontWeight: 'bold', color: '#ecf0f1'}}>
						order_count: 7
					</div>
					<div style={{fontSize: '14px', marginTop: '10px', opacity: 0.9}}>
						Exactly what was asked - no unnecessary data!
					</div>
				</div>
			</div>
		</AbsoluteFill>
	);
};
