import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	spring,
	useCurrentFrame,
	useVideoConfig,
	Sequence,
} from 'remotion';
import {InitialMessageSequence} from './sequences/InitialMessageSequence';
import {FirstToolCallSequence} from './sequences/FirstToolCallSequence';
import {ContextSummarizationSequence} from './sequences/ContextSummarizationSequence';

export const ContextBuildupAnimation: React.FC = () => {
	const frame = useCurrentFrame();
	const {fps} = useVideoConfig();

	const titleOpacity = interpolate(frame, [0, 30], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const titleScale = spring({
		frame,
		fps,
		config: {
			damping: 10,
			stiffness: 100,
		},
	});

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
					top: '8%',
					left: '50%',
					transform: `translate(-50%, -50%) scale(${titleScale})`,
					color: 'white',
					fontSize: '56px',
					fontWeight: 'bold',
					textAlign: 'center',
					opacity: titleOpacity,
					textShadow: '2px 2px 4px rgba(0,0,0,0.3)',
				}}
			>
				LLM Context Buildup
			</div>

			{/* Subtitle */}
			<div
				style={{
					position: 'absolute',
					top: '14%',
					left: '50%',
					transform: 'translate(-50%, -50%)',
					color: 'rgba(255, 255, 255, 0.8)',
					fontSize: '24px',
					textAlign: 'center',
					opacity: titleOpacity,
				}}
			>
				LLM internal processing with chain of thought
			</div>

			{/* Sequence 1: Initial Messages (frames 60-300) */}
			<Sequence from={60} durationInFrames={240}>
				<InitialMessageSequence />
			</Sequence>

			{/* Sequence 2: First Tool Call (frames 300-540) */}
			<Sequence from={300} durationInFrames={240}>
				<FirstToolCallSequence />
			</Sequence>

			{/* Sequence 3: Context Summarization (frames 540-840) */}
			<Sequence from={540} durationInFrames={300}>
				<ContextSummarizationSequence />
			</Sequence>
		</AbsoluteFill>
	);
};
