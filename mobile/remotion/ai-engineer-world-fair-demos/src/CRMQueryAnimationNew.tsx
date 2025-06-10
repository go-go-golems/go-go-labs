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

	return (
		<AbsoluteFill
			style={{
				background: 'linear-gradient(135deg, #e74c3c 0%, #c0392b 100%)',
				fontFamily: 'Arial, sans-serif',
			}}
		>
			{/* InteractionRenderer handles the rest */}
			<InteractionRenderer
				sequence={crmQuerySequence}
				background="transparent" // Use transparent since we have the gradient background
				containerStyle={{
					top: '5%', // Start below the title and subtitle
					height: '90%', // Adjust height to fit below title
				}}
			/>
		</AbsoluteFill>
	);
}; 