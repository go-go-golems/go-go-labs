import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	spring,
	useCurrentFrame,
	useVideoConfig,
} from 'remotion';

export const ToolAnalysisSequence: React.FC = () => {
	const frame = useCurrentFrame();
	const {fps} = useVideoConfig();

	const stepOpacity = interpolate(frame, [0, 30], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const llmOpacity = interpolate(frame, [0, 30], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const thoughtBubbleOpacity = interpolate(frame, [30, 60], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const tool1Opacity = interpolate(frame, [90, 120], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const tool2Opacity = interpolate(frame, [120, 150], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const tool3Opacity = interpolate(frame, [150, 180], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const selectionGlow = interpolate(frame, [210, 240], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const thoughtScale = spring({
		frame: frame - 30,
		fps,
		config: {
			damping: 10,
			stiffness: 100,
		},
	});

	return (
		<AbsoluteFill
			style={{
				background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
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
					fontSize: '32px',
					fontWeight: 'bold',
					opacity: stepOpacity,
				}}
			>
				Step 2: LLM analyzes available tools
			</div>

			{/* LLM with thinking animation */}
			<div
				style={{
					position: 'absolute',
					top: '30%',
					left: '20%',
					opacity: llmOpacity,
				}}
			>
				<div
					style={{
						width: '120px',
						height: '120px',
						borderRadius: '20px',
						backgroundColor: '#7B68EE',
						display: 'flex',
						alignItems: 'center',
						justifyContent: 'center',
						fontSize: '60px',
						color: 'white',
						boxShadow: '0 4px 12px rgba(0,0,0,0.3)',
						animation: frame > 60 ? 'pulse 1s infinite' : 'none',
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
					opacity: thoughtBubbleOpacity,
					transform: `scale(${thoughtScale})`,
				}}
			>
				<div
					style={{
						backgroundColor: 'rgba(255,255,255,0.95)',
						borderRadius: '20px',
						padding: '15px',
						boxShadow: '0 4px 12px rgba(0,0,0,0.3)',
						fontSize: '16px',
						color: '#333',
						maxWidth: '250px',
					}}
				>
					"I need weather data... Let me check what tools I have available."
				</div>
			</div>

			{/* Available Tools */}
			<div
				style={{
					position: 'absolute',
					top: '55%',
					left: '50%',
					transform: 'translateX(-50%)',
					color: 'white',
					fontSize: '24px',
					fontWeight: 'bold',
					opacity: tool1Opacity,
				}}
			>
				Available Tools:
			</div>

			{/* Tool 1: Weather API */}
			<div
				style={{
					position: 'absolute',
					top: '65%',
					left: '15%',
					opacity: tool1Opacity,
				}}
			>
				<div
					style={{
						backgroundColor: selectionGlow > 0.5 ? '#32CD32' : '#FF6B6B',
						borderRadius: '15px',
						padding: '20px',
						boxShadow: `0 4px 12px rgba(0,0,0,0.3) ${selectionGlow > 0.5 ? ', 0 0 20px #32CD32' : ''}`,
						textAlign: 'center',
						color: 'white',
						fontSize: '16px',
						fontWeight: 'bold',
						minWidth: '150px',
					}}
				>
					<div style={{fontSize: '30px', marginBottom: '10px'}}>🌤️</div>
					Weather API
				</div>
			</div>

			{/* Tool 2: Calculator */}
			<div
				style={{
					position: 'absolute',
					top: '65%',
					left: '42.5%',
					opacity: tool2Opacity,
				}}
			>
				<div
					style={{
						backgroundColor: '#4A90E2',
						borderRadius: '15px',
						padding: '20px',
						boxShadow: '0 4px 12px rgba(0,0,0,0.3)',
						textAlign: 'center',
						color: 'white',
						fontSize: '16px',
						fontWeight: 'bold',
						minWidth: '150px',
					}}
				>
					<div style={{fontSize: '30px', marginBottom: '10px'}}>🔢</div>
					Calculator
				</div>
			</div>

			{/* Tool 3: File Reader */}
			<div
				style={{
					position: 'absolute',
					top: '65%',
					right: '15%',
					opacity: tool3Opacity,
				}}
			>
				<div
					style={{
						backgroundColor: '#9B59B6',
						borderRadius: '15px',
						padding: '20px',
						boxShadow: '0 4px 12px rgba(0,0,0,0.3)',
						textAlign: 'center',
						color: 'white',
						fontSize: '16px',
						fontWeight: 'bold',
						minWidth: '150px',
					}}
				>
					<div style={{fontSize: '30px', marginBottom: '10px'}}>📁</div>
					File Reader
				</div>
			</div>

			{/* Selection indicator */}
			{selectionGlow > 0 && (
				<div
					style={{
						position: 'absolute',
						top: '85%',
						left: '50%',
						transform: 'translateX(-50%)',
						color: '#32CD32',
						fontSize: '20px',
						fontWeight: 'bold',
						opacity: selectionGlow,
					}}
				>
					✓ Weather API selected!
				</div>
			)}
		</AbsoluteFill>
	);
};
