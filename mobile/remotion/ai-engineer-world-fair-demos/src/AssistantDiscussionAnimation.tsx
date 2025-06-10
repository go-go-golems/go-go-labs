import React from 'react';
import {AbsoluteFill} from 'remotion';
import {InteractionRenderer} from './components/InteractionRenderer';
import {assistantDiscussionSequence} from './sequences/configs/AssistantDiscussionConfig';

export const AssistantDiscussionAnimation: React.FC = () => {
	return (
		<AbsoluteFill>
			<InteractionRenderer
				sequence={assistantDiscussionSequence}
				background="linear-gradient(135deg, #2c3e50 0%, #34495e 100%)"
			/>
		</AbsoluteFill>
	);
}; 