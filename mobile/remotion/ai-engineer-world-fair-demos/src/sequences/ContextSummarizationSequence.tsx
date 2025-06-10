import React from 'react';
import { InteractionRenderer } from '../components/InteractionRenderer';
import { contextSummarizationSequence } from './configs/ContextSummarizationConfig';

export const ContextSummarizationSequence: React.FC = () => {
	return <InteractionRenderer sequence={contextSummarizationSequence} />;
};
