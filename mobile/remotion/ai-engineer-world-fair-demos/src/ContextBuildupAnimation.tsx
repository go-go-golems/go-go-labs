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
import { InteractionRenderer } from './components/InteractionRenderer';
import { contextSummarizationSequence } from './sequences/configs/ContextSummarizationConfig';

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


			<Sequence from={0} durationInFrames={300}>
			<InteractionRenderer
				sequence={contextSummarizationSequence}
				background="transparent"
				containerStyle={{
					top: '15%',
					height: '60%',
				}}
			/>
			</Sequence>
		</AbsoluteFill>
	);
};
