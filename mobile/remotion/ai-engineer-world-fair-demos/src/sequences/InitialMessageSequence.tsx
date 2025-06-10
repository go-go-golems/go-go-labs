import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	useCurrentFrame,
	useVideoConfig,
} from 'remotion';

interface MessageProps {
	type: 'user' | 'assistant';
	content: string;
	opacity: number;
}

const Message: React.FC<MessageProps> = ({type, content, opacity}) => {
	const getTypeConfig = (type: string) => {
		switch (type) {
			case 'user':
				return {bg: '#3498db', icon: '👤', label: 'User'};
			case 'assistant':
				return {bg: '#9b59b6', icon: '🧠', label: 'Assistant'};
			default:
				return {bg: '#7f8c8d', icon: '?', label: 'Unknown'};
		}
	};

	const config = getTypeConfig(type);

	return (
		<div
			style={{
				opacity,
				backgroundColor: config.bg,
				borderRadius: '12px',
				padding: '12px 16px',
				color: 'white',
				fontSize: '16px',
				margin: '8px 0',
				display: 'flex',
				alignItems: 'center',
				gap: '12px',
				boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
			}}
		>
			<span style={{fontSize: '20px'}}>{config.icon}</span>
			<div>
				<div style={{fontSize: '12px', opacity: 0.8, marginBottom: '4px'}}>
					{config.label}
				</div>
				<div>{content}</div>
			</div>
		</div>
	);
};

export const InitialMessageSequence: React.FC = () => {
	const frame = useCurrentFrame();

	// Context window
	const containerOpacity = interpolate(frame, [0, 30], [0, 1], {
		extrapolateRight: 'clamp',
	});

	// Messages appear one by one
	const message1Opacity = interpolate(frame, [60, 90], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const message2Opacity = interpolate(frame, [120, 150], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const message3Opacity = interpolate(frame, [180, 210], [0, 1], {
		extrapolateRight: 'clamp',
	});

	// Token counter
	const tokenCount = Math.floor(
		interpolate(frame, [60, 210], [0, 850], {
			extrapolateRight: 'clamp',
		})
	);

	return (
		<AbsoluteFill
			style={{
				background: 'linear-gradient(135deg, #2c3e50 0%, #34495e 100%)',
				fontFamily: 'Arial, sans-serif',
			}}
		>
			{/* Title */}
			<div
				style={{
					position: 'absolute',
					top: '15%',
					left: '50%',
					transform: 'translate(-50%, -50%)',
					color: 'white',
					fontSize: '28px',
					fontWeight: 'bold',
					textAlign: 'center',
					opacity: containerOpacity,
				}}
			>
				Step 1: Initial Messages
			</div>

			{/* Context Container */}
			<div
				style={{
					position: 'absolute',
					top: '25%',
					left: '50%',
					transform: 'translate(-50%, 0)',
					width: '800px',
					height: '400px',
					border: '2px solid rgba(255, 255, 255, 0.3)',
					borderRadius: '16px',
					backgroundColor: 'rgba(255, 255, 255, 0.05)',
					padding: '20px',
					opacity: containerOpacity,
				}}
			>
				<div
					style={{
						color: 'white',
						fontSize: '16px',
						fontWeight: 'bold',
						marginBottom: '20px',
						textAlign: 'center',
					}}
				>
					Context Window
				</div>

				{/* Messages */}
				<div style={{height: '320px', overflow: 'hidden'}}>
					<Message
						type="user"
						content="What's the weather like in San Francisco?"
						opacity={message1Opacity}
					/>
					<Message
						type="assistant"
						content="I'll check the weather for you. Let me use the weather API..."
						opacity={message2Opacity}
					/>
					<Message
						type="user"
						content="Also, what about the forecast for tomorrow?"
						opacity={message3Opacity}
					/>
				</div>
			</div>

			{/* Token Counter */}
			<div
				style={{
					position: 'absolute',
					bottom: '15%',
					left: '50%',
					transform: 'translateX(-50%)',
					color: 'white',
					fontSize: '18px',
					fontWeight: 'bold',
					opacity: containerOpacity,
				}}
			>
				Tokens: {tokenCount.toLocaleString()} / 128,000
			</div>
		</AbsoluteFill>
	);
};
