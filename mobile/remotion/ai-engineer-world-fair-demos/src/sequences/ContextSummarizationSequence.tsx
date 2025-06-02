import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	useCurrentFrame,
	useVideoConfig,
} from 'remotion';

interface MessageProps {
	type: 'user' | 'assistant' | 'assistant_cot' | 'assistant_diary' | 'tool_use' | 'tool_result' | 'summary';
	content: string;
	opacity: number;
	fadeOut?: boolean;
}

const Message: React.FC<MessageProps> = ({type, content, opacity, fadeOut = false}) => {
	const getTypeConfig = (type: string) => {
		switch (type) {
			case 'user':
				return {bg: '#3498db', icon: '👤', label: 'User'};
			case 'assistant':
				return {bg: '#9b59b6', icon: '🧠', label: 'Assistant'};
			case 'assistant_cot':
				return {bg: '#e74c3c', icon: '🤔', label: 'Chain of Thought'};
			case 'assistant_diary':
				return {bg: '#8e44ad', icon: '📔', label: 'Diary'};
			case 'tool_use':
				return {bg: '#e67e22', icon: '⚡', label: 'Tool Use'};
			case 'tool_result':
				return {bg: '#27ae60', icon: '📊', label: 'Tool Result'};
			case 'summary':
				return {bg: '#6c3483', icon: '📝', label: 'Summary'};
			default:
				return {bg: '#7f8c8d', icon: '?', label: 'Unknown'};
		}
	};

	const config = getTypeConfig(type);

	return (
		<div
			style={{
				opacity: fadeOut ? opacity * 0.3 : opacity,
				backgroundColor: config.bg,
				borderRadius: '10px',
				padding: (type === 'summary') ? '14px 18px' : 
					(type === 'assistant_cot' || type === 'assistant_diary') ? '8px 12px' : '12px 15px',
				color: 'white',
				fontSize: (type === 'summary') ? '13px' : 
					(type === 'assistant_cot' || type === 'assistant_diary') ? '11px' : '13px',
				margin: '3px 0',
				display: 'flex',
				alignItems: 'center',
				gap: '10px',
				boxShadow: (type === 'summary') ? 
					'0 4px 15px rgba(108, 52, 131, 0.4)' : 
					(type === 'assistant_cot' || type === 'assistant_diary') ? 
					'0 3px 10px rgba(0,0,0,0.2)' : 
					'0 2px 8px rgba(0,0,0,0.1)',
				border: (type === 'summary' || type === 'assistant_cot' || type === 'assistant_diary') ? 
					'1px solid rgba(255, 255, 255, 0.2)' : 'none',
			}}
		>
			<span style={{
				fontSize: (type === 'summary') ? '18px' : 
					(type === 'assistant_cot' || type === 'assistant_diary') ? '14px' : '16px'
			}}>{config.icon}</span>
			<div>
				<div style={{
					fontSize: (type === 'summary') ? '10px' : 
						(type === 'assistant_cot' || type === 'assistant_diary') ? '8px' : '9px', 
					opacity: 0.8, 
					marginBottom: '3px',
					fontWeight: (type === 'summary' || type === 'assistant_cot' || type === 'assistant_diary') ? 'bold' : 'normal'
				}}>
					{config.label}
				</div>
				<div style={{
					fontSize: (type === 'summary') ? '12px' : 
						(type === 'assistant_cot' || type === 'assistant_diary') ? '10px' : '12px', 
					lineHeight: 1.2,
					fontWeight: type === 'summary' ? '500' : 'normal',
					fontStyle: (type === 'assistant_cot' || type === 'assistant_diary') ? 'italic' : 'normal'
				}}>
					{content}
				</div>
			</div>
		</div>
	);
};

export const ContextSummarizationSequence: React.FC = () => {
	const frame = useCurrentFrame();

	// Container
	const containerOpacity = interpolate(frame, [0, 30], [0, 1], {
		extrapolateRight: 'clamp',
	});

	// Previous conversation 
	const prevOpacity = interpolate(frame, [30, 35], [0, 1], {
		extrapolateRight: 'clamp',
	});

	// Start coding session - user request
	const userOpacity = interpolate(frame, [40, 50], [0, 1], {
		extrapolateRight: 'clamp',
	});

	// Assistant COT + Diary + Tool sequence (the long part)
	const cot1Opacity = interpolate(frame, [60, 70], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const diary1Opacity = interpolate(frame, [80, 90], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const tool1Opacity = interpolate(frame, [100, 110], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const result1Opacity = interpolate(frame, [120, 130], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const cot2Opacity = interpolate(frame, [140, 150], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const diary2Opacity = interpolate(frame, [160, 170], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const responseOpacity = interpolate(frame, [180, 190], [0, 1], {
		extrapolateRight: 'clamp',
	});

	// Fade out the long sequence (8 blocks total)
	const fadeOutProgress = interpolate(frame, [220, 250], [0, 1], {
		extrapolateRight: 'clamp',
	});

	// Summary appears
	const summaryOpacity = interpolate(frame, [270, 300], [0, 1], {
		extrapolateRight: 'clamp',
	});

	// Token counter
	const tokenCount = Math.floor(
		interpolate(frame, [30, 190], [3200, 9200], {
			extrapolateRight: 'clamp',
		})
	);

	const optimizedTokens = Math.floor(
		interpolate(frame, [270, 300], [9200, 4800], {
			extrapolateRight: 'clamp',
		})
	);

	const currentTokens = frame > 270 ? optimizedTokens : tokenCount;

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
					top: '12%',
					left: '50%',
					transform: 'translate(-50%, -50%)',
					color: 'white',
					fontSize: '28px',
					fontWeight: 'bold',
					textAlign: 'center',
					opacity: containerOpacity,
				}}
			>
				Step 3: Long Coding Session → Summary
			</div>

			{/* Context Container */}
			<div
				style={{
					position: 'absolute',
					top: '20%',
					left: '50%',
					transform: 'translate(-50%, 0)',
					width: '950px',
					height: '500px',
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

				{/* Two Column Layout */}
				<div style={{display: 'flex', gap: '20px', height: '420px', overflow: 'hidden'}}>
					{/* Left Column - Main Conversation */}
					<div style={{flex: 1}}>
						{/* Previous conversation */}
						<Message
							type="user"
							content="What's the weather in SF?"
							opacity={prevOpacity}
						/>

						{/* New coding request */}
						<Message
							type="user"
							content="Help me debug this Python sorting function: def sort_list(arr): arr.sort() return arr[0]"
							opacity={userOpacity}
						/>

						{/* Tool interaction */}
						<Message
							type="tool_use"
							content="run_python_code('def sort_list(arr): arr.sort(); return arr[0]; print(sort_list([3,1,4]))')"
							opacity={tool1Opacity}
							fadeOut={fadeOutProgress > 0}
						/>
						<Message
							type="tool_result"
							content="Output: 1\nFunction returns minimum value, not sorted list [1, 3, 4]"
							opacity={result1Opacity}
							fadeOut={fadeOutProgress > 0}
						/>

						{/* Assistant response */}
						<Message
							type="assistant"
							content="I found the issue! Your function returns arr[0] (just the minimum) instead of the sorted list. Fix: return arr"
							opacity={responseOpacity}
							fadeOut={fadeOutProgress > 0}
						/>

						{/* Summary replaces the sequences */}
						{summaryOpacity > 0 && (
							<Message
								type="summary"
								content="Debugging Session: Fixed Python sort function - changed 'return arr[0]' to 'return arr' to return full sorted list instead of minimum element."
								opacity={summaryOpacity}
							/>
						)}
					</div>

					{/* Right Column - Internal Processing */}
					<div style={{flex: 1}}>
						{/* Chain of thought and diary entries */}
						<Message
							type="assistant_cot"
							content="User has a Python function issue. They're calling sort() but returning arr[0]. This returns just the minimum element, not the sorted list."
							opacity={cot1Opacity}
							fadeOut={fadeOutProgress > 0}
						/>
						<Message
							type="assistant_diary"
							content="Task: Debug Python sort function. Issue identified: returns single element instead of full sorted array."
							opacity={diary1Opacity}
							fadeOut={fadeOutProgress > 0}
						/>
						<Message
							type="assistant_cot"
							content="Confirmed the bug. Need to explain that they should return 'arr' not 'arr[0]'. Also suggest sorted() as alternative to avoid mutation."
							opacity={cot2Opacity}
							fadeOut={fadeOutProgress > 0}
						/>
						<Message
							type="assistant_diary"
							content="Solution: Change return arr[0] to return arr. Also mention sorted() function as non-mutating alternative."
							opacity={diary2Opacity}
							fadeOut={fadeOutProgress > 0}
						/>
					</div>
				</div>
			</div>



			{/* Summarization explanation */}
			{summaryOpacity > 0 && (
				<div
					style={{
						position: 'absolute',
						bottom: '25%',
						left: '50%',
						transform: 'translateX(-50%)',
						backgroundColor: '#6c3483',
						color: 'white',
						padding: '12px 25px',
						borderRadius: '10px',
						fontSize: '14px',
						textAlign: 'center',
						opacity: summaryOpacity,
						boxShadow: '0 4px 15px rgba(108, 52, 131, 0.4)',
					}}
				>
					<div style={{marginBottom: '5px', fontWeight: 'bold'}}>
						✨ 7 message blocks condensed into 1 summary
					</div>
					<div style={{fontSize: '12px', opacity: 0.9}}>
						Full COT + diary + tool sequence → single summary block
					</div>
				</div>
			)}

			{/* Token Counter */}
			<div
				style={{
					position: 'absolute',
					bottom: '10%',
					left: '50%',
					transform: 'translateX(-50%)',
					color: currentTokens < 5000 ? '#27ae60' : 'white',
					fontSize: '18px',
					fontWeight: 'bold',
					opacity: containerOpacity,
				}}
			>
				Tokens: {currentTokens.toLocaleString()} / 128,000
				{frame > 270 && (
					<span style={{ color: '#27ae60', marginLeft: '10px' }}>
						↓ Optimized!
					</span>
				)}
			</div>
		</AbsoluteFill>
	);
};
