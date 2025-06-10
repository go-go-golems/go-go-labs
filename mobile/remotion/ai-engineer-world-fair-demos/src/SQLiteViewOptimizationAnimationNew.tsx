import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	spring,
	useCurrentFrame,
	useVideoConfig,
} from 'remotion';
import { InteractionRenderer } from './components/InteractionRenderer';
import { sqliteViewOptimizationSequence } from './sequences/configs/SQLiteViewOptimizationConfig';

export const SQLiteViewOptimizationAnimationNew: React.FC = () => {
	const frame = useCurrentFrame();
	const { fps } = useVideoConfig();

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

			{/* InteractionRenderer handles the conversation flow */}
			<InteractionRenderer
				sequence={sqliteViewOptimizationSequence}
				background="transparent"
				containerStyle={{
					top: '10%',
					height: '70%',
				}}
			/>
		</AbsoluteFill>
	);
}; 