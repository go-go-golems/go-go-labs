import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	spring,
	useCurrentFrame,
	useVideoConfig,
} from 'remotion';
import { InteractionRenderer } from './components/InteractionRenderer';
import { crmQuerySequence } from './sequences/configs/CRMQueryConfig';

export const CRMQueryAnimationNew: React.FC = () => {
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
					zIndex: 1,
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
					zIndex: 1,
				}}
			>
				When simple queries return massive datasets
			</div>

			{/* InteractionRenderer handles the rest */}
			<InteractionRenderer
				sequence={crmQuerySequence}
				background="transparent" // Use transparent since we have the gradient background
				containerStyle={{
					top: '22%', // Start below the title and subtitle
					height: '73%', // Adjust height to fit below title
				}}
			/>
		</AbsoluteFill>
	);
}; 