import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	useCurrentFrame,
	useVideoConfig,
} from 'remotion';

interface MessageProps {
	type: 'user' | 'assistant' | 'tool_use' | 'tool_result';
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
			case 'tool_use':
				return {bg: '#e67e22', icon: '⚡', label: 'Tool Use'};
			case 'tool_result':
				return {bg: '#27ae60', icon: '📊', label: 'Tool Result'};
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
				padding: '12px 15px',
				color: 'white',
				fontSize: '14px',
				margin: '6px 0',
				display: 'flex',
				alignItems: 'center',
				gap: '12px',
				boxShadow: '0 3px 10px rgba(0,0,0,0.2)',
			}}
		>
			<span style={{fontSize: '16px'}}>{config.icon}</span>
			<div>
				<div style={{fontSize: '10px', opacity: 0.8, marginBottom: '4px'}}>
					{config.label}
				</div>
				<div style={{fontSize: '13px', lineHeight: 1.3}}>{content}</div>
			</div>
		</div>
	);
};

export const FirstToolCallSequence: React.FC = () => {
	const frame = useCurrentFrame();

	// Container
	const containerOpacity = interpolate(frame, [0, 30], [0, 1], {
		extrapolateRight: 'clamp',
	});

	// Previous messages (already there)
	const prevMessagesOpacity = interpolate(frame, [30, 40], [0, 1], {
		extrapolateRight: 'clamp',
	});

	// New messages appear
	const toolUseOpacity = interpolate(frame, [60, 90], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const toolResultOpacity = interpolate(frame, [120, 150], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const assistantOpacity = interpolate(frame, [180, 210], [0, 1], {
		extrapolateRight: 'clamp',
	});

	// Token counter
	const tokenCount = Math.floor(
		interpolate(frame, [30, 210], [850, 3200], {
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
				Step 2: First Tool Call
			</div>

			{/* Context Container */}
			<div
				style={{
					position: 'absolute',
					top: '25%',
					left: '50%',
					transform: 'translate(-50%, 0)',
					width: '900px',
					height: '450px',
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
						marginBottom: '15px',
						textAlign: 'center',
					}}
				>
					Context Window
				</div>

				{/* Messages */}
				<div style={{height: '380px', overflow: 'hidden'}}>
					{/* Previous conversation */}
					<Message
						type="user"
						content="What's the weather like in San Francisco?"
						opacity={prevMessagesOpacity}
					/>
					<Message
						type="assistant"
						content="I'll check the weather for you. Let me use the weather API..."
						opacity={prevMessagesOpacity}
					/>
					<Message
						type="user"
						content="Also, what about the forecast for tomorrow?"
						opacity={prevMessagesOpacity}
					/>

					{/* New tool interaction */}
					<Message
						type="tool_use"
						content="get_weather(location='San Francisco', days=2)"
						opacity={toolUseOpacity}
					/>
					<Message
						type="tool_result"
						content="Today: 72°F, sunny. Tomorrow: 68°F, partly cloudy. Wind: 8mph..."
						opacity={toolResultOpacity}
					/>
					<Message
						type="assistant"
						content="The weather in San Francisco today is 72°F and sunny. Tomorrow will be 68°F and partly cloudy..."
						opacity={assistantOpacity}
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
