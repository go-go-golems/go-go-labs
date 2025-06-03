import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	spring,
	useCurrentFrame,
	useVideoConfig,
} from 'remotion';
import { InteractionRenderer } from './components/InteractionRenderer';
import { toolCallingSequence } from './sequences/configs/ToolCallingConfig';

export const ToolCallingAnimationNew: React.FC = () => {
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
			{/* Main title - appears before the sequence starts
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
					zIndex: 1,
				}}
			>
				How LLMs Use Tools
			</div>

			{/* Subtitle */}
			{/* <div
				style={{
					position: 'absolute',
					top: '15%',
					left: '50%',
					transform: 'translate(-50%, -50%)',
					color: 'rgba(255, 255, 255, 0.8)',
					fontSize: '24px',
					textAlign: 'center',
					opacity: titleOpacity,
					textShadow: '1px 1px 2px rgba(0,0,0,0.3)',
					zIndex: 1,
				}}
			>
				A complete workflow demonstration
			</div> */}

			{/* InteractionRenderer handles the rest */}
			<InteractionRenderer
				sequence={toolCallingSequence}
				background="transparent" // Use transparent since we have the gradient background
				containerStyle={{
					top: '20%', // Start below the title
					height: '75%', // Adjust height to fit below title
				}}
			/>
		</AbsoluteFill>
	);
}; 