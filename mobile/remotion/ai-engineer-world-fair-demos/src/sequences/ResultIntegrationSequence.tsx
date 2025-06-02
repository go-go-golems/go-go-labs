import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	spring,
	useCurrentFrame,
	useVideoConfig,
} from 'remotion';

export const ResultIntegrationSequence: React.FC = () => {
	const frame = useCurrentFrame();
	const {fps} = useVideoConfig();

	const stepOpacity = interpolate(frame, [0, 30], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const llmOpacity = interpolate(frame, [0, 30], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const processingOpacity = interpolate(frame, [60, 90], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const responseOpacity = interpolate(frame, [120, 150], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const arrowProgress = interpolate(frame, [180, 240], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const userOpacity = interpolate(frame, [210, 240], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const finalMessageOpacity = interpolate(frame, [270, 300], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const processingScale = spring({
		frame: frame - 60,
		fps,
		config: {
			damping: 8,
			stiffness: 80,
		},
	});

	const responseScale = spring({
		frame: frame - 120,
		fps,
		config: {
			damping: 8,
			stiffness: 80,
		},
	});

	const finalMessageScale = spring({
		frame: frame - 270,
		fps,
		config: {
			damping: 8,
			stiffness: 80,
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
					top: '10%',
					left: '50%',
					transform: 'translateX(-50%)',
					color: 'white',
					fontSize: '32px',
					fontWeight: 'bold',
					opacity: stepOpacity,
				}}
			>
				Step 4: Result integration & response
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
						backgroundColor: '#7B68EE',
						display: 'flex',
						alignItems: 'center',
						justifyContent: 'center',
						fontSize: '60px',
						color: 'white',
						boxShadow: '0 4px 12px rgba(0,0,0,0.3)',
						animation: frame > 60 && frame < 150 ? 'pulse 1s infinite' : 'none',
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

			{/* Processing thought */}
			<div
				style={{
					position: 'absolute',
					top: '15%',
					left: '35%',
					opacity: processingOpacity,
					transform: `scale(${processingScale})`,
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
						maxWidth: '300px',
					}}
				>
					"Let me integrate this weather data into a helpful response..."
				</div>
			</div>

			{/* Generated Response */}
			<div
				style={{
					position: 'absolute',
					top: '45%',
					left: '25%',
					opacity: responseOpacity,
					transform: `scale(${responseScale})`,
				}}
			>
				<div
					style={{
						backgroundColor: '#7B68EE',
						borderRadius: '20px',
						padding: '20px',
						boxShadow: '0 4px 12px rgba(0,0,0,0.3)',
						fontSize: '18px',
						color: 'white',
						maxWidth: '500px',
						lineHeight: 1.4,
					}}
				>
					<div style={{fontWeight: 'bold', marginBottom: '10px'}}>
						🤖 LLM Response:
					</div>
					"The weather in San Francisco today is sunny with a temperature of 72°F. 
					It's quite pleasant with 65% humidity and a gentle 8 mph wind. 
					Perfect weather for outdoor activities!"
				</div>
			</div>

			{/* Arrow to user */}
			<div
				style={{
					position: 'absolute',
					top: '75%',
					left: '15%',
					width: '200px',
					height: '4px',
					backgroundColor: '#32CD32',
					borderRadius: '2px',
					transform: `scaleX(${arrowProgress})`,
					transformOrigin: 'left center',
				}}
			>
				<div
					style={{
						position: 'absolute',
						right: '-10px',
						top: '-8px',
						width: 0,
						height: 0,
						borderTop: '10px solid transparent',
						borderBottom: '10px solid transparent',
						borderLeft: '20px solid #32CD32',
						opacity: arrowProgress,
					}}
				/>
			</div>

			{/* User */}
			<div
				style={{
					position: 'absolute',
					top: '70%',
					right: '20%',
					opacity: userOpacity,
				}}
			>
				<div
					style={{
						width: '100px',
						height: '100px',
						borderRadius: '50%',
						backgroundColor: '#4A90E2',
						display: 'flex',
						alignItems: 'center',
						justifyContent: 'center',
						fontSize: '50px',
						color: 'white',
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

			{/* Final completion message */}
			<div
				style={{
					position: 'absolute',
					bottom: '10%',
					left: '50%',
					transform: `translateX(-50%) scale(${finalMessageScale})`,
					opacity: finalMessageOpacity,
				}}
			>
				<div
					style={{
						backgroundColor: 'rgba(50, 205, 50, 0.9)',
						borderRadius: '25px',
						padding: '20px 40px',
						color: 'white',
						fontSize: '24px',
						fontWeight: 'bold',
						textAlign: 'center',
						boxShadow: '0 4px 12px rgba(0,0,0,0.3)',
					}}
				>
					✅ Tool calling complete!
				</div>
			</div>

			<style jsx>{`
				@keyframes pulse {
					0%, 100% { transform: scale(1); }
					50% { transform: scale(1.05); }
				}
			`}</style>
		</AbsoluteFill>
	);
};
