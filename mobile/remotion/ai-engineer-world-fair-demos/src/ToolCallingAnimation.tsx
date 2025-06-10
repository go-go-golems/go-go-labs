import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	spring,
	useCurrentFrame,
	useVideoConfig,
	Sequence,
} from 'remotion';
import {UserRequestSequence} from './sequences/UserRequestSequence';
import {ToolAnalysisSequence} from './sequences/ToolAnalysisSequence';
import {ToolExecutionSequence} from './sequences/ToolExecutionSequence';
import {ResultIntegrationSequence} from './sequences/ResultIntegrationSequence';

export const ToolCallingAnimation: React.FC = () => {
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
				background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
				fontFamily: 'Arial, sans-serif',
			}}
		>
			{/* Title */}
			<div
				style={{
					position: 'absolute',
					top: '10%',
					left: '50%',
					transform: `translate(-50%, -50%) scale(${titleScale})`,
					color: 'white',
					fontSize: '60px',
					fontWeight: 'bold',
					textAlign: 'center',
					opacity: titleOpacity,
					textShadow: '2px 2px 4px rgba(0,0,0,0.3)',
				}}
			>
				How LLMs Use Tools
			</div>

			{/* Sequence 1: User Request (frames 60-240) */}
			<Sequence from={60} durationInFrames={180}>
				<UserRequestSequence />
			</Sequence>

			{/* Sequence 2: Tool Analysis (frames 240-480) */}
			<Sequence from={240} durationInFrames={240}>
				<ToolAnalysisSequence />
			</Sequence>

			{/* Sequence 3: Tool Execution (frames 480-780) */}
			<Sequence from={480} durationInFrames={300}>
				<ToolExecutionSequence />
			</Sequence>

			{/* Sequence 4: Result Integration (frames 780-1080) */}
			<Sequence from={780} durationInFrames={300}>
				<ResultIntegrationSequence />
			</Sequence>
		</AbsoluteFill>
	);
};
