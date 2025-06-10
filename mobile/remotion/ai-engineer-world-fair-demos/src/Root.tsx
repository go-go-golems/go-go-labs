import {Composition} from 'remotion';
import {ToolCallingAnimation} from './ToolCallingAnimation';
import {ToolCallingAnimationNew} from './ToolCallingAnimationNew';
import {CRMQueryAnimation} from './CRMQueryAnimation';
import {CRMQueryAnimationNew} from './CRMQueryAnimationNew';
import {SQLiteQueryAnimation} from './SQLiteQueryAnimation';
import {SQLiteQueryAnimationNew} from './SQLiteQueryAnimationNew';
import {SQLiteViewOptimizationAnimation} from './SQLiteViewOptimizationAnimation';
import {SQLiteViewOptimizationAnimationNew} from './SQLiteViewOptimizationAnimationNew';
import {ComprehensiveComparisonAnimation} from './ComprehensiveComparisonAnimation';
import {ContextBuildupAnimation} from './ContextBuildupAnimation';
import {PostResponseEditingAnimation} from './PostResponseEditingAnimation';
import {AdaptiveSystemPromptAnimation} from './AdaptiveSystemPromptAnimation';
import {AssistantDiscussionAnimation} from './AssistantDiscussionAnimation';
import {UserControlledToolsAnimation} from './UserControlledToolsAnimation';
import {LLMGeneratedUIAnimation} from './LLMGeneratedUIAnimation';

// Individual step components
import {UserRequestSequence} from './sequences/UserRequestSequence';
import {UserRequestSequenceNew} from './sequences/UserRequestSequenceNew';
import {ToolAnalysisSequence} from './sequences/ToolAnalysisSequence';
import {ToolExecutionSequence} from './sequences/ToolExecutionSequence';
import {ResultIntegrationSequence} from './sequences/ResultIntegrationSequence';

import {CRMUserRequestSequence} from './sequences/CRMUserRequestSequence';
import {CRMToolAnalysisSequence} from './sequences/CRMToolAnalysisSequence';
import {CRMToolExecutionSequence} from './sequences/CRMToolExecutionSequence';
import {CRMResultProcessingSequence} from './sequences/CRMResultProcessingSequence';

import {SQLiteUserRequestSequence} from './sequences/SQLiteUserRequestSequence';
import {SQLiteSchemaDiscoverySequence} from './sequences/SQLiteSchemaDiscoverySequence';
import {SQLiteTableExplorationSequence} from './sequences/SQLiteTableExplorationSequence';
import {SQLiteTargetedQuerySequence} from './sequences/SQLiteTargetedQuerySequence';
import {SQLiteFinalResponseSequence} from './sequences/SQLiteFinalResponseSequence';

import {ViewCreationSequence} from './sequences/ViewCreationSequence';
import {MultipleQueriesSequence} from './sequences/MultipleQueriesSequence';
import {PerformanceComparisonSequence} from './sequences/PerformanceComparisonSequence';

import {TokenEfficiencyComparisonSequence} from './sequences/TokenEfficiencyComparisonSequence';
import {ViewPersistenceSequence} from './sequences/ViewPersistenceSequence';
import {ToolDiscoverySequence} from './sequences/ToolDiscoverySequence';
import {FutureEfficiencySequence} from './sequences/FutureEfficiencySequence';

// Context Buildup Animation sequences
import {InitialMessageSequence} from './sequences/InitialMessageSequence';
import {FirstToolCallSequence} from './sequences/FirstToolCallSequence';
import {ContextSummarizationSequence} from './sequences/ContextSummarizationSequence';

export const RemotionRoot: React.FC = () => {
	return (
		<>
			{/* Full Animations */}
			<Composition
				id="ToolCallingAnimationNew"
				component={ToolCallingAnimationNew}
				durationInFrames={780}
				fps={30}
				width={1920}
				height={1080}
			/>
			<Composition
				id="CRMQueryAnimationNew"
				component={CRMQueryAnimationNew}
				durationInFrames={630}
				fps={30}
				width={1920}
				height={1080}
			/>
			<Composition
				id="SQLiteQueryAnimationNew"
				component={SQLiteQueryAnimationNew}
				durationInFrames={1440}
				fps={30}
				width={1920}
				height={1080}
			/>
			<Composition
				id="SQLiteViewOptimizationAnimationNew"
				component={SQLiteViewOptimizationAnimationNew}
				durationInFrames={1280}
				fps={30}
				width={1920}
				height={1080}
			/>
			<Composition
				id="ContextBuildupAnimation"
				component={ContextBuildupAnimation}
				durationInFrames={300}
				fps={30}
				width={1920}
				height={1080}
			/>
			<Composition
				id="PostResponseEditingAnimation"
				component={PostResponseEditingAnimation}
				durationInFrames={400}
				fps={30}
				width={1920}
				height={1080}
			/>
			<Composition
				id="AdaptiveSystemPromptAnimation"
				component={AdaptiveSystemPromptAnimation}
				durationInFrames={630}
				fps={30}
				width={1920}
				height={1080}
			/>
			<Composition
				id="AssistantDiscussionAnimation"
				component={AssistantDiscussionAnimation}
				durationInFrames={360}
				fps={30}
				width={1920}
				height={1080}
			/>
			<Composition
				id="UserControlledToolsAnimation"
				component={UserControlledToolsAnimation}
				durationInFrames={480}
				fps={30}
				width={1920}
				height={1080}
			/>
			<Composition
				id="LLMGeneratedUIAnimation"
				component={LLMGeneratedUIAnimation}
				durationInFrames={460}
				fps={30}
				width={1920}
				height={1080}
			/>

		</>
	);
};
