import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	useCurrentFrame,
	spring,
	useVideoConfig,
} from 'remotion';
import { InteractionRenderer } from './components/InteractionRenderer';
import { postResponseEditingSequence } from './sequences/configs/PostResponseEditingConfig';

export const PostResponseEditingAnimation: React.FC = () => {
	const frame = useCurrentFrame();
	const { fps } = useVideoConfig();

	// Title animations
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

	// Subtitle animation
	const subtitleOpacity = interpolate(frame, [20, 50], [0, 1], {
		extrapolateRight: 'clamp',
	});

	// Main content animation
	const contentOpacity = interpolate(frame, [40, 70], [0, 1], {
		extrapolateRight: 'clamp',
	});

	return (
		<AbsoluteFill
			style={{
				background: 'linear-gradient(135deg, #2c3e50 0%, #34495e 100%)',
				fontFamily: 'Arial, sans-serif',
			}}
		>
			{/* Main Title */}
			<div
				style={{
					position: 'absolute',
					top: '8%',
					left: '50%',
					transform: `translate(-50%, -50%) scale(${titleScale})`,
					opacity: titleOpacity,
					color: 'white',
					fontSize: '42px',
					fontWeight: 'bold',
					textAlign: 'center',
					textShadow: '0 4px 8px rgba(0,0,0,0.3)',
				}}
			>
				Code Testing & Editing Workflow
			</div>

			{/* Subtitle */}
			<div
				style={{
					position: 'absolute',
					top: '13%',
					left: '50%',
					transform: 'translate(-50%, -50%)',
					opacity: subtitleOpacity,
					color: 'rgba(255, 255, 255, 0.9)',
					fontSize: '20px',
					textAlign: 'center',
					maxWidth: '800px',
					lineHeight: 1.4,
				}}
			>
				How users can edit AI responses when tests reveal code issues
			</div>

			{/* Main Content */}
			<div
				style={{
					opacity: contentOpacity,
					position: 'absolute',
					top: '18%',
					left: '0',
					right: '0',
					bottom: '0',
				}}
			>
				<InteractionRenderer
					sequence={postResponseEditingSequence}
					background="transparent"
				/>
			</div>

			{/* Feature highlights */}
			<div
				style={{
					position: 'absolute',
					bottom: '5%',
					left: '50%',
					transform: 'translateX(-50%)',
					opacity: interpolate(frame, [350, 380], [0, 1], {
						extrapolateRight: 'clamp',
					}),
					color: 'rgba(255, 255, 255, 0.8)',
					fontSize: '14px',
					textAlign: 'center',
					display: 'flex',
					gap: '40px',
				}}
			>
				<div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
					<span style={{ fontSize: '18px' }}>🧪</span>
					<span>Test Code</span>
				</div>
				<div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
					<span style={{ fontSize: '18px' }}>✏️</span>
					<span>Edit Response</span>
				</div>
				<div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
					<span style={{ fontSize: '18px' }}>✅</span>
					<span>Fix Issues</span>
				</div>
			</div>
		</AbsoluteFill>
	);
}; 