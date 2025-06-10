import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	spring,
	useCurrentFrame,
	useVideoConfig,
} from 'remotion';

interface ContextBoxProps {
	type: 'user' | 'assistant' | 'tool_use' | 'tool_result';
	content: string;
	opacity: number;
	position: {top: string; left: string};
	scale?: number;
	highlight?: boolean;
	fading?: boolean;
}

const ContextBox: React.FC<ContextBoxProps> = ({
	type,
	content,
	opacity,
	position,
	scale = 1,
	highlight = false,
	fading = false,
}) => {
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
				position: 'absolute',
				...position,
				opacity: fading ? opacity * 0.3 : opacity,
				transform: `scale(${scale})`,
				transition: 'opacity 0.5s ease',
			}}
		>
			<div
				style={{
					backgroundColor: fading ? '#7f8c8d' : config.bg,
					borderRadius: '8px',
					padding: '6px 8px',
					color: 'white',
					fontSize: '10px',
					minWidth: '80px',
					maxWidth: '120px',
					boxShadow: highlight
						? '0 0 20px rgba(255, 215, 0, 0.8)'
						: '0 2px 6px rgba(0,0,0,0.2)',
					border: highlight ? '2px solid #ffd700' : 'none',
					display: 'flex',
					flexDirection: 'column',
					gap: '2px',
				}}
			>
				<div style={{display: 'flex', alignItems: 'center', gap: '4px'}}>
					<span style={{fontSize: '12px'}}>{config.icon}</span>
					<span style={{fontWeight: 'bold', fontSize: '8px'}}>
						{config.label}
					</span>
				</div>
				<div
					style={{
						fontSize: '9px',
						lineHeight: 1.1,
						overflow: 'hidden',
						textOverflow: 'ellipsis',
						whiteSpace: 'nowrap',
					}}
				>
					{content}
				</div>
			</div>
		</div>
	);
};

export const ContextOverflowSequence: React.FC = () => {
	const frame = useCurrentFrame();
	const {fps} = useVideoConfig();

	// Initial context state
	const initialOpacity = interpolate(frame, [0, 20], [0, 1], {
		extrapolateRight: 'clamp',
	});

	// More messages coming in
	const newMessagesOpacity = interpolate(frame, [30, 60], [0, 1], {
		extrapolateRight: 'clamp',
	});

	// Context window overflow warning
	const warningOpacity = interpolate(frame, [90, 120], [0, 1], {
		extrapolateRight: 'clamp',
	});

	// Context sliding/truncation
	const truncationProgress = interpolate(frame, [150, 210], [0, 1], {
		extrapolateRight: 'clamp',
	});

	// Summary/optimization
	const optimizationOpacity = interpolate(frame, [240, 270], [0, 1], {
		extrapolateRight: 'clamp',
	});

	// Token counter
	const tokenCount = Math.floor(
		interpolate(frame, [0, 150], [8500, 127000], {
			extrapolateRight: 'clamp',
		})
	);

	const optimizedTokenCount = Math.floor(
		interpolate(frame, [240, 300], [127000, 15000], {
			extrapolateRight: 'clamp',
		})
	);

	const currentTokens = frame > 240 ? optimizedTokenCount : tokenCount;

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
					top: '8%',
					left: '50%',
					transform: 'translate(-50%, -50%)',
					color: 'white',
					fontSize: '32px',
					fontWeight: 'bold',
					textAlign: 'center',
					opacity: initialOpacity,
				}}
			>
				Step 4: Context Management
			</div>

			{/* Context window */}
			<div
				style={{
					position: 'absolute',
					top: '15%',
					left: '50%',
					transform: 'translate(-50%, 0)',
					width: '950px',
					height: '550px',
					border: `3px solid ${currentTokens > 120000 ? '#e74c3c' : 'rgba(255, 255, 255, 0.3)'}`,
					borderRadius: '20px',
					opacity: initialOpacity,
					backgroundColor: 'rgba(255, 255, 255, 0.05)',
					overflow: 'hidden',
				}}
			>
				<div
					style={{
						position: 'absolute',
						top: '-35px',
						left: '50%',
						transform: 'translateX(-50%)',
						color: currentTokens > 120000 ? '#e74c3c' : 'white',
						fontSize: '16px',
						fontWeight: 'bold',
					}}
				>
					Context Window - {currentTokens > 120000 ? 'OVERFLOW!' : 'Managed'}
				</div>
			</div>

			{/* Massive conversation history - compressed view */}
			{initialOpacity > 0 && (
				<>
					{/* Row 1 - Original conversation */}
					{Array.from({length: 8}).map((_, i) => (
						<ContextBox
							key={`row1-${i}`}
							type={['user', 'assistant', 'tool_use', 'tool_result'][i % 4] as any}
							content={
								['Weather SF?', 'Checking...', 'get_weather()', '72°F sunny'][
									i % 4
								]
							}
							opacity={initialOpacity}
							position={{
								top: '20%',
								left: `${8 + i * 11}%`,
							}}
							scale={0.4}
							fading={truncationProgress > 0.5}
						/>
					))}

					{/* Row 2 - More conversation */}
					{Array.from({length: 8}).map((_, i) => (
						<ContextBox
							key={`row2-${i}`}
							type={['user', 'assistant', 'tool_use', 'tool_result'][i % 4] as any}
							content={
								['NYC too?', 'Sure!', 'get_weather(NYC)', '65°F cloudy'][i % 4]
							}
							opacity={initialOpacity}
							position={{
								top: '30%',
								left: `${8 + i * 11}%`,
							}}
							scale={0.4}
							fading={truncationProgress > 0.3}
						/>
					))}

					{/* Row 3 - Even more */}
					{Array.from({length: 8}).map((_, i) => (
						<ContextBox
							key={`row3-${i}`}
							type={['user', 'assistant', 'tool_use', 'tool_result'][i % 4] as any}
							content={
								['London?', 'Checking...', 'get_weather()', '58°F rainy'][i % 4]
							}
							opacity={initialOpacity}
							position={{
								top: '40%',
								left: `${8 + i * 11}%`,
							}}
							scale={0.4}
							fading={truncationProgress > 0.1}
						/>
					))}
				</>
			)}

			{/* New messages flooding in */}
			{newMessagesOpacity > 0 && (
				<>
					<ContextBox
						type="user"
						content="What about Tokyo weather?"
						opacity={newMessagesOpacity}
						position={{top: '50%', left: '15%'}}
						scale={0.7}
						highlight={frame >= 30 && frame <= 90}
					/>
					<ContextBox
						type="user"
						content="And can you help me plan a trip?"
						opacity={newMessagesOpacity}
						position={{top: '55%', left: '35%'}}
						scale={0.7}
						highlight={frame >= 45 && frame <= 90}
					/>
					<ContextBox
						type="user"
						content="Also translate this text..."
						opacity={newMessagesOpacity}
						position={{top: '50%', left: '55%'}}
						scale={0.7}
						highlight={frame >= 60 && frame <= 90}
					/>
				</>
			)}

			{/* Warning message */}
			{warningOpacity > 0 && (
				<div
					style={{
						position: 'absolute',
						top: '72%',
						left: '50%',
						transform: 'translate(-50%, 0)',
						backgroundColor: '#e74c3c',
						color: 'white',
						padding: '15px 25px',
						borderRadius: '10px',
						fontSize: '18px',
						fontWeight: 'bold',
						textAlign: 'center',
						opacity: warningOpacity,
						boxShadow: '0 5px 15px rgba(231, 76, 60, 0.3)',
					}}
				>
					⚠️ Context window approaching limit!
				</div>
			)}

			{/* Context management in action */}
			{optimizationOpacity > 0 && (
				<div
					style={{
						position: 'absolute',
						top: '75%',
						left: '50%',
						transform: 'translate(-50%, 0)',
						backgroundColor: '#27ae60',
						color: 'white',
						padding: '15px 25px',
						borderRadius: '10px',
						fontSize: '16px',
						fontWeight: 'bold',
						textAlign: 'center',
						opacity: optimizationOpacity,
						boxShadow: '0 5px 15px rgba(39, 174, 96, 0.3)',
					}}
				>
					<div style={{marginBottom: '5px'}}>
						✅ Context optimized: Old messages summarized
					</div>
					<div style={{fontSize: '14px', fontWeight: 'normal'}}>
						Key information preserved, redundant details removed
					</div>
				</div>
			)}

			{/* Token counter */}
			<div
				style={{
					position: 'absolute',
					top: '88%',
					right: '5%',
					color: currentTokens > 120000 ? '#e74c3c' : 'white',
					fontSize: '20px',
					fontWeight: 'bold',
					opacity: initialOpacity,
				}}
			>
				Tokens: {currentTokens.toLocaleString()} / 128,000
			</div>

			{/* Progress bar */}
			<div
				style={{
					position: 'absolute',
					top: '93%',
					right: '5%',
					width: '250px',
					height: '12px',
					backgroundColor: 'rgba(255, 255, 255, 0.2)',
					borderRadius: '6px',
					overflow: 'hidden',
					opacity: initialOpacity,
				}}
			>
				<div
					style={{
						width: `${Math.min((currentTokens / 128000) * 250, 250)}px`,
						height: '100%',
						backgroundColor:
							currentTokens > 120000
								? '#e74c3c'
								: currentTokens > 100000
									? '#f39c12'
									: '#27ae60',
						transition: 'all 0.5s ease',
					}}
				/>
			</div>

			{/* Context lifecycle explanation */}
			{frame > 210 && (
				<div
					style={{
						position: 'absolute',
						bottom: '2%',
						left: '50%',
						transform: 'translateX(-50%)',
						color: 'white',
						fontSize: '14px',
						textAlign: 'center',
						opacity: interpolate(frame, [210, 240], [0, 1], {
							extrapolateRight: 'clamp',
						}),
					}}
				>
					LLMs must manage context size through summarization, truncation,
					or sliding windows
				</div>
			)}
		</AbsoluteFill>
	);
};
