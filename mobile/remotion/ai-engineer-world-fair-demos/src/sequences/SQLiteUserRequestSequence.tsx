import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	spring,
	useCurrentFrame,
	useVideoConfig,
} from 'remotion';

export const SQLiteUserRequestSequence: React.FC = () => {
	const frame = useCurrentFrame();
	const {fps} = useVideoConfig();

	const userIconOpacity = interpolate(frame, [0, 20], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const messageOpacity = interpolate(frame, [20, 50], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const arrowOpacity = interpolate(frame, [70, 100], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const arrowLength = interpolate(frame, [70, 100], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const llmOpacity = interpolate(frame, [90, 120], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const messageScale = spring({
		frame: frame - 20,
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
					top: '22%',
					left: '50%',
					transform: 'translateX(-50%)',
					color: 'white',
					fontSize: '28px',
					fontWeight: 'bold',
					opacity: 0.9,
				}}
			>
				Step 1: User needs database information
			</div>

			{/* User Icon */}
			<div
				style={{
					position: 'absolute',
					top: '45%',
					left: '15%',
					opacity: userIconOpacity,
				}}
			>
				<div
					style={{
						width: '100px',
						height: '100px',
						borderRadius: '50%',
						backgroundColor: '#3498db',
						display: 'flex',
						alignItems: 'center',
						justifyContent: 'center',
						fontSize: '50px',
						color: 'white',
						boxShadow: '0 4px 15px rgba(0,0,0,0.2)',
					}}
				>
					👤
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
					User
				</div>
			</div>

			{/* User Message */}
			<div
				style={{
					position: 'absolute',
					top: '35%',
					left: '32%',
					opacity: messageOpacity,
					transform: `scale(${messageScale})`,
				}}
			>
				<div
					style={{
						backgroundColor: 'white',
						borderRadius: '20px',
						padding: '25px',
						boxShadow: '0 6px 20px rgba(0,0,0,0.15)',
						maxWidth: '400px',
						fontSize: '18px',
						color: '#2c3e50',
						position: 'relative',
						lineHeight: 1.4,
					}}
				>
					"How many orders did customer John Smith place last month?"
					<div
						style={{
							position: 'absolute',
							left: '-15px',
							top: '25px',
							width: 0,
							height: 0,
							borderTop: '15px solid transparent',
							borderBottom: '15px solid transparent',
							borderRight: '15px solid white',
						}}
					/>
				</div>
				<div
					style={{
						marginTop: '10px',
						padding: '8px 15px',
						backgroundColor: 'rgba(255,255,255,0.2)',
						borderRadius: '10px',
						color: 'white',
						fontSize: '14px',
						textAlign: 'center',
						fontStyle: 'italic',
					}}
				>
					Requires database knowledge
				</div>
			</div>

			{/* Smooth Arrow */}
			<div
				style={{
					position: 'absolute',
					top: '50%',
					left: '65%',
					opacity: arrowOpacity,
				}}
			>
				<svg width="120" height="40" viewBox="0 0 120 40">
					<defs>
						<linearGradient id="sqliteArrowGradient" x1="0%" y1="0%" x2="100%" y2="0%">
							<stop offset="0%" stopColor="#34495e" />
							<stop offset="100%" stopColor="#2c3e50" />
						</linearGradient>
					</defs>
					<path
						d={`M 10 20 L ${10 + 80 * arrowLength} 20`}
						stroke="url(#sqliteArrowGradient)"
						strokeWidth="4"
						strokeLinecap="round"
						fill="none"
					/>
					<polygon
						points={`${10 + 80 * arrowLength},20 ${10 + 80 * arrowLength - 12},15 ${10 + 80 * arrowLength - 12},25`}
						fill="url(#sqliteArrowGradient)"
						opacity={arrowLength}
					/>
				</svg>
			</div>

			{/* LLM */}
			<div
				style={{
					position: 'absolute',
					top: '45%',
					right: '15%',
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

			{/* Available Tool indicator */}
			<div
				style={{
					position: 'absolute',
					bottom: '15%',
					left: '50%',
					transform: 'translateX(-50%)',
					opacity: llmOpacity,
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
					}}
				>
					<div style={{fontSize: '24px', marginBottom: '5px'}}>🗃️</div>
					<div style={{fontWeight: 'bold'}}>Only tool available:</div>
					<div style={{fontFamily: 'monospace', fontSize: '18px'}}>sqlite_query(sql)</div>
				</div>
			</div>
		</AbsoluteFill>
	);
};
