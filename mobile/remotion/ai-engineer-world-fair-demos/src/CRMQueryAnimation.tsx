import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	spring,
	useCurrentFrame,
	useVideoConfig,
	Sequence,
} from 'remotion';
import {CRMUserRequestSequence} from './sequences/CRMUserRequestSequence';
import {CRMToolAnalysisSequence} from './sequences/CRMToolAnalysisSequence';
import {CRMToolExecutionSequence} from './sequences/CRMToolExecutionSequence';
import {CRMResultProcessingSequence} from './sequences/CRMResultProcessingSequence';

export const CRMQueryAnimation: React.FC = () => {
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

	const subtitleOpacity = interpolate(frame, [30, 60], [0, 1], {
		extrapolateRight: 'clamp',
	});

	return (
		<AbsoluteFill
			style={{
				background: 'linear-gradient(135deg, #e74c3c 0%, #c0392b 100%)',
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
					fontSize: '52px',
					fontWeight: 'bold',
					textAlign: 'center',
					opacity: titleOpacity,
					textShadow: '2px 2px 4px rgba(0,0,0,0.3)',
				}}
			>
				The Token Inefficiency Problem
			</div>

			{/* Subtitle */}
			<div
				style={{
					position: 'absolute',
					top: '15%',
					left: '50%',
					transform: 'translateX(-50%)',
					color: 'rgba(255,255,255,0.9)',
					fontSize: '24px',
					textAlign: 'center',
					opacity: subtitleOpacity,
				}}
			>
				When simple queries return massive datasets
			</div>

			{/* Sequence 1: User Request (frames 90-240) */}
			<Sequence from={90} durationInFrames={150}>
				<CRMUserRequestSequence />
			</Sequence>

			{/* Sequence 2: Tool Analysis (frames 240-420) */}
			<Sequence from={240} durationInFrames={180}>
				<CRMToolAnalysisSequence />
			</Sequence>

			{/* Sequence 3: Tool Execution (frames 420-780) */}
			<Sequence from={420} durationInFrames={360}>
				<CRMToolExecutionSequence />
			</Sequence>

			{/* Sequence 4: Result Processing (frames 780-1200) */}
			<Sequence from={780} durationInFrames={420}>
				<CRMResultProcessingSequence />
			</Sequence>
		</AbsoluteFill>
	);
};
