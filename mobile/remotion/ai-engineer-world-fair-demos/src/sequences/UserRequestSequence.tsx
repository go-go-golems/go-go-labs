import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	spring,
	useCurrentFrame,
	useVideoConfig,
} from 'remotion';

export const UserRequestSequence: React.FC = () => {
	const frame = useCurrentFrame();
	const {fps} = useVideoConfig();

	const userIconOpacity = interpolate(frame, [0, 30], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const messageOpacity = interpolate(frame, [30, 60], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const arrowProgress = interpolate(frame, [90, 150], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const llmOpacity = interpolate(frame, [120, 150], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const messageScale = spring({
		frame: frame - 30,
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
					top: '20%',
					left: '50%',
					transform: 'translateX(-50%)',
					color: 'white',
					fontSize: '32px',
					fontWeight: 'bold',
					opacity: 0.8,
				}}
			>
				Step 1: User sends a request
			</div>

			{/* User Icon */}
			<div
				style={{
					position: 'absolute',
					top: '40%',
					left: '15%',
					opacity: userIconOpacity,
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

			{/* User Message */}
			<div
				style={{
					position: 'absolute',
					top: '30%',
					left: '30%',
					opacity: messageOpacity,
					transform: `scale(${messageScale})`,
				}}
			>
				<div
					style={{
						backgroundColor: 'white',
						borderRadius: '20px',
						padding: '20px',
						boxShadow: '0 4px 12px rgba(0,0,0,0.3)',
						maxWidth: '400px',
						fontSize: '18px',
						color: '#333',
						position: 'relative',
					}}
				>
					"What's the weather like in San Francisco today?"
					<div
						style={{
							position: 'absolute',
							left: '-10px',
							top: '20px',
							width: 0,
							height: 0,
							borderTop: '10px solid transparent',
							borderBottom: '10px solid transparent',
							borderRight: '10px solid white',
						}}
					/>
				</div>
			</div>

			{/* Arrow */}
			<div
				style={{
					position: 'absolute',
					top: '45%',
					left: '70%',
					width: '150px',
					height: '4px',
					backgroundColor: '#FFD700',
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
						borderLeft: '20px solid #FFD700',
						opacity: arrowProgress,
					}}
				/>
			</div>

			{/* LLM */}
			<div
				style={{
					position: 'absolute',
					top: '40%',
					right: '15%',
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
		</AbsoluteFill>
	);
};
