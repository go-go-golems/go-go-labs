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
	compressed?: boolean;
}

const ContextBox: React.FC<ContextBoxProps> = ({
	type,
	content,
	opacity,
	position,
	scale = 1,
	highlight = false,
	compressed = false,
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
				opacity,
				transform: `scale(${scale})`,
			}}
		>
			<div
				style={{
					backgroundColor: config.bg,
					borderRadius: compressed ? '8px' : '12px',
					padding: compressed ? '8px 10px' : '12px 15px',
					color: 'white',
					fontSize: compressed ? '12px' : '14px',
					minWidth: compressed ? '120px' : '180px',
					maxWidth: compressed ? '180px' : '250px',
					boxShadow: highlight
						? '0 0 20px rgba(255, 215, 0, 0.8)'
						: '0 3px 10px rgba(0,0,0,0.2)',
					border: highlight ? '2px solid #ffd700' : 'none',
					display: 'flex',
					flexDirection: 'column',
					gap: compressed ? '3px' : '5px',
				}}
			>
				<div style={{display: 'flex', alignItems: 'center', gap: '6px'}}>
					<span style={{fontSize: compressed ? '14px' : '16px'}}>
						{config.icon}
					</span>
					<span
						style={{
							fontWeight: 'bold',
							fontSize: compressed ? '10px' : '12px',
						}}
					>
						{config.label}
					</span>
				</div>
				<div
					style={{
						fontSize: compressed ? '11px' : '13px',
						lineHeight: 1.2,
						overflow: 'hidden',
						textOverflow: 'ellipsis',
						whiteSpace: compressed ? 'nowrap' : 'normal',
					}}
				>
					{content}
				</div>
			</div>
		</div>
	);
};

export const SecondToolCallSequence: React.FC = () => {
	const frame = useCurrentFrame();
	const {fps} = useVideoConfig();

	// Show compressed previous context
	const previousContextOpacity = interpolate(frame, [0, 20], [0, 1], {
		extrapolateRight: 'clamp',
	});

	// New user request
	const newUserRequestOpacity = interpolate(frame, [30, 60], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const newUserRequestScale = spring({
		frame: frame - 30,
		fps,
		config: {
			damping: 8,
			stiffness: 80,
		},
	});

	// Second tool use
	const secondToolUseOpacity = interpolate(frame, [90, 120], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const secondToolUseScale = spring({
		frame: frame - 90,
		fps,
		config: {
			damping: 8,
			stiffness: 80,
		},
	});

	// Second tool result
	const secondToolResultOpacity = interpolate(frame, [150, 180], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const secondToolResultScale = spring({
		frame: frame - 150,
		fps,
		config: {
			damping: 8,
			stiffness: 80,
		},
	});

	// Final assistant response
	const finalResponseOpacity = interpolate(frame, [210, 240], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const finalResponseScale = spring({
		frame: frame - 210,
		fps,
		config: {
			damping: 8,
			stiffness: 80,
		},
	});

	// Token counter
	const tokenCount = Math.floor(
		interpolate(frame, [0, 240], [3200, 8500], {
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
			{/* Step indicator */}
			<div
				style={{
					position: 'absolute',
					top: '12%',
					left: '50%',
					transform: 'translate(-50%, -50%)',
					color: 'white',
					fontSize: '32px',
					fontWeight: 'bold',
					textAlign: 'center',
					opacity: previousContextOpacity,
				}}
			>
				Step 3: Multiple Tool Calls
			</div>

			{/* Context window */}
			<div
				style={{
					position: 'absolute',
					top: '20%',
					left: '50%',
					transform: 'translate(-50%, 0)',
					width: '900px',
					height: '500px',
					border: '3px solid rgba(255, 255, 255, 0.3)',
					borderRadius: '20px',
					opacity: previousContextOpacity,
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
						color: 'white',
						fontSize: '16px',
						fontWeight: 'bold',
					}}
				>
					Context Window - Accumulating History
				</div>
			</div>

			{/* Previous conversation (heavily compressed) */}
			{previousContextOpacity > 0 && (
				<>
					{/* Row 1 */}
					<ContextBox
						type="user"
						content="Weather in SF?"
						opacity={previousContextOpacity}
						position={{top: '25%', left: '8%'}}
						scale={0.6}
						compressed
					/>
					<ContextBox
						type="assistant"
						content="I'll check..."
						opacity={previousContextOpacity}
						position={{top: '25%', left: '20%'}}
						scale={0.6}
						compressed
					/>
					<ContextBox
						type="user"
						content="Tomorrow too?"
						opacity={previousContextOpacity}
						position={{top: '25%', left: '32%'}}
						scale={0.6}
						compressed
					/>
					<ContextBox
						type="tool_use"
						content="get_weather(SF,2)"
						opacity={previousContextOpacity}
						position={{top: '25%', left: '44%'}}
						scale={0.6}
						compressed
					/>
					<ContextBox
						type="tool_result"
						content="72°F sunny, 68°F cloudy"
						opacity={previousContextOpacity}
						position={{top: '25%', left: '56%'}}
						scale={0.6}
						compressed
					/>
					<ContextBox
						type="assistant"
						content="Today 72°F sunny..."
						opacity={previousContextOpacity}
						position={{top: '25%', left: '68%'}}
						scale={0.6}
						compressed
					/>
				</>
			)}

			{/* New user request */}
			{newUserRequestOpacity > 0 && (
				<ContextBox
					type="user"
					content="Can you also check the weather in New York City and compare it?"
					opacity={newUserRequestOpacity}
					position={{top: '40%', left: '15%'}}
					scale={newUserRequestScale}
					highlight={frame >= 30 && frame <= 90}
				/>
			)}

			{/* Second tool use */}
			{secondToolUseOpacity > 0 && (
				<ContextBox
					type="tool_use"
					content="get_weather(location='New York City', days=1)"
					opacity={secondToolUseOpacity}
					position={{top: '55%', left: '15%'}}
					scale={secondToolUseScale}
					highlight={frame >= 90 && frame <= 150}
				/>
			)}

			{/* Second tool result */}
			{secondToolResultOpacity > 0 && (
				<ContextBox
					type="tool_result"
					content="NYC: 65°F, overcast, humidity 78%, wind 12mph NE"
					opacity={secondToolResultOpacity}
					position={{top: '55%', left: '45%'}}
					scale={secondToolResultScale}
					highlight={frame >= 150 && frame <= 210}
				/>
			)}

			{/* Final assistant response */}
			{finalResponseOpacity > 0 && (
				<ContextBox
					type="assistant"
					content="Comparing SF and NYC: SF is warmer (72°F vs 65°F) and sunnier. NYC is more humid and windier..."
					opacity={finalResponseOpacity}
					position={{top: '70%', left: '30%'}}
					scale={finalResponseScale}
					highlight={frame >= 210}
				/>
			)}

			{/* Context accumulation notice */}
			{frame > 180 && (
				<div
					style={{
						position: 'absolute',
						top: '85%',
						left: '50%',
						transform: 'translate(-50%, 0)',
						color: 'white',
						fontSize: '16px',
						textAlign: 'center',
						opacity: interpolate(frame, [180, 200], [0, 1], {
							extrapolateRight: 'clamp',
						}),
					}}
				>
					<div style={{marginBottom: '5px'}}>
						Each message stays in context for future reference
					</div>
					<div style={{fontSize: '14px', color: 'rgba(255, 255, 255, 0.7)'}}>
						Context grows with every interaction
					</div>
				</div>
			)}

			{/* Token counter */}
			<div
				style={{
					position: 'absolute',
					top: '92%',
					right: '8%',
					color: 'white',
					fontSize: '18px',
					fontWeight: 'bold',
					opacity: previousContextOpacity,
				}}
			>
				Tokens: {tokenCount.toLocaleString()} / 128,000
			</div>

			{/* Progress bar */}
			<div
				style={{
					position: 'absolute',
					top: '95%',
					right: '8%',
					width: '200px',
					height: '8px',
					backgroundColor: 'rgba(255, 255, 255, 0.2)',
					borderRadius: '4px',
					overflow: 'hidden',
					opacity: previousContextOpacity,
				}}
			>
				<div
					style={{
						width: `${(tokenCount / 128000) * 200}px`,
						height: '100%',
						backgroundColor: tokenCount > 100000 ? '#e74c3c' : '#27ae60',
						transition: 'width 0.3s ease',
					}}
				/>
			</div>
		</AbsoluteFill>
	);
};
