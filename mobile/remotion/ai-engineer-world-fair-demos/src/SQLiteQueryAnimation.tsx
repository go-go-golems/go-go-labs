import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	spring,
	useCurrentFrame,
	useVideoConfig,
	Sequence,
} from 'remotion';
import {SQLiteUserRequestSequence} from './sequences/SQLiteUserRequestSequence';
import {SQLiteSchemaDiscoverySequence} from './sequences/SQLiteSchemaDiscoverySequence';
import {SQLiteTableExplorationSequence} from './sequences/SQLiteTableExplorationSequence';
import {SQLiteTargetedQuerySequence} from './sequences/SQLiteTargetedQuerySequence';
import {SQLiteFinalResponseSequence} from './sequences/SQLiteFinalResponseSequence';

export const SQLiteQueryAnimation: React.FC = () => {
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
					fontSize: '48px',
					fontWeight: 'bold',
					textAlign: 'center',
					opacity: titleOpacity,
					textShadow: '2px 2px 4px rgba(0,0,0,0.3)',
				}}
			>
				Intelligent Multi-Step Tool Use
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
				Exploring database schema to craft precise queries
			</div>

			{/* Sequence 1: User Request (frames 90-210) */}
			<Sequence from={90} durationInFrames={120}>
				<SQLiteUserRequestSequence />
			</Sequence>

			{/* Sequence 2: Schema Discovery (frames 210-360) */}
			<Sequence from={210} durationInFrames={150}>
				<SQLiteSchemaDiscoverySequence />
			</Sequence>

			{/* Sequence 3: Table Exploration (frames 360-540) */}
			<Sequence from={360} durationInFrames={180}>
				<SQLiteTableExplorationSequence />
			</Sequence>

			{/* Sequence 4: Targeted Query (frames 540-780) */}
			<Sequence from={540} durationInFrames={240}>
				<SQLiteTargetedQuerySequence />
			</Sequence>

			{/* Sequence 5: Final Response (frames 780-1200) */}
			<Sequence from={780} durationInFrames={420}>
				<SQLiteFinalResponseSequence />
			</Sequence>
		</AbsoluteFill>
	);
};
