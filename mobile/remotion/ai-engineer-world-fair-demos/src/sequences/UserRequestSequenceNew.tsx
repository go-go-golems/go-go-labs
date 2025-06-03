import React from 'react';
import { AbsoluteFill } from 'remotion';
import { InteractionRenderer } from '../components/InteractionRenderer';
import { userRequestStepSequence } from './configs/UserRequestStepConfig';

export const UserRequestSequenceNew: React.FC = () => {
	return (
		<AbsoluteFill>
			<InteractionRenderer
				sequence={userRequestStepSequence}
				background="linear-gradient(135deg, #667eea 0%, #764ba2 100%)"
			/>
		</AbsoluteFill>
	);
}; 