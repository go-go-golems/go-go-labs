import React from 'react';
import {AbsoluteFill} from 'remotion';
import {InteractionRenderer} from './components/InteractionRenderer';
import {userControlledToolsSequence} from './sequences/configs/UserControlledToolsConfig';

export const UserControlledToolsAnimation: React.FC = () => {
	return (
		<AbsoluteFill>
			<InteractionRenderer
				sequence={userControlledToolsSequence}
				background="linear-gradient(135deg, #34495e 0%, #2c3e50 100%)"
			/>
		</AbsoluteFill>
	);
}; 