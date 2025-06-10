import React from 'react';
import {AbsoluteFill} from 'remotion';
import {InteractionRenderer} from './components/InteractionRenderer';
import {llmGeneratedUISequence} from './sequences/configs/LLMGeneratedUIConfig';

export const LLMGeneratedUIAnimation: React.FC = () => {
	return (
		<AbsoluteFill
			style={{
				background: 'linear-gradient(135deg, #4a90e2 0%, #9b59b6 50%, #e67e22 100%)',
			}}
		>
			<InteractionRenderer
				containerStyle={{
					top: '10%',
					height: '70%',
				}}
				sequence={llmGeneratedUISequence}
				background="transparent"
			/>
		</AbsoluteFill>
	);
}; 