import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	spring,
	useCurrentFrame,
	useVideoConfig,
	Sequence,
} from 'remotion';
import {TokenEfficiencyComparisonSequence} from './sequences/TokenEfficiencyComparisonSequence';
import {ViewPersistenceSequence} from './sequences/ViewPersistenceSequence';
import {ToolDiscoverySequence} from './sequences/ToolDiscoverySequence';
import {FutureEfficiencySequence} from './sequences/FutureEfficiencySequence';

export const ComprehensiveComparisonAnimation: React.FC = () => {
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
				background: 'linear-gradient(135deg, #27ae60 0%, #2ecc71 100%)',
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
					fontSize: '44px',
					fontWeight: 'bold',
					textAlign: 'center',
					opacity: titleOpacity,
					textShadow: '2px 2px 4px rgba(0,0,0,0.3)',
				}}
			>
				The Evolution of Tool Intelligence
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
				From inefficient calls to intelligent infrastructure
			</div>

			{/* Sequence 1: Token Efficiency Comparison (frames 90-390) */}
			<Sequence from={90} durationInFrames={300}>
				<TokenEfficiencyComparisonSequence />
			</Sequence>

			{/* Sequence 2: View Persistence (frames 390-660) */}
			<Sequence from={390} durationInFrames={270}>
				<ViewPersistenceSequence />
			</Sequence>

			{/* Sequence 3: Tool Discovery (frames 660-930) */}
			<Sequence from={660} durationInFrames={270}>
				<ToolDiscoverySequence />
			</Sequence>

			{/* Sequence 4: Future Efficiency (frames 930-1200) */}
			<Sequence from={930} durationInFrames={270}>
				<FutureEfficiencySequence />
			</Sequence>
		</AbsoluteFill>
	);
};
