import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	spring,
	useCurrentFrame,
	useVideoConfig,
	Sequence,
} from 'remotion';
import {ViewCreationSequence} from './sequences/ViewCreationSequence';
import {MultipleQueriesSequence} from './sequences/MultipleQueriesSequence';
import {PerformanceComparisonSequence} from './sequences/PerformanceComparisonSequence';

export const SQLiteViewOptimizationAnimation: React.FC = () => {
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
				background: 'linear-gradient(135deg, #8e44ad 0%, #9b59b6 100%)',
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
					fontSize: '48px',
					fontWeight: 'bold',
					textAlign: 'center',
					opacity: titleOpacity,
					textShadow: '2px 2px 4px rgba(0,0,0,0.3)',
				}}
			>
				Optimizing with SQL Views
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
				Creating reusable infrastructure for multiple queries
			</div>

			{/* Sequence 1: View Creation (frames 90-360) */}
			<Sequence from={90} durationInFrames={270}>
				<ViewCreationSequence />
			</Sequence>

			{/* Sequence 2: Multiple Queries (frames 360-780) */}
			<Sequence from={360} durationInFrames={420}>
				<MultipleQueriesSequence />
			</Sequence>

			{/* Sequence 3: Performance Comparison (frames 780-1200) */}
			<Sequence from={780} durationInFrames={420}>
				<PerformanceComparisonSequence />
			</Sequence>
		</AbsoluteFill>
	);
};
