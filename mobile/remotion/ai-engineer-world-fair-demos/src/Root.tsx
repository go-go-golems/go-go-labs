import {Composition} from 'remotion';
import {ToolCallingAnimation} from './ToolCallingAnimation';
import {CRMQueryAnimation} from './CRMQueryAnimation';
import {SQLiteQueryAnimation} from './SQLiteQueryAnimation';
import {SQLiteViewOptimizationAnimation} from './SQLiteViewOptimizationAnimation';
import {ComprehensiveComparisonAnimation} from './ComprehensiveComparisonAnimation';
import {ContextBuildupAnimation} from './ContextBuildupAnimation';
import {PostResponseEditingAnimation} from './PostResponseEditingAnimation';
import {AdaptiveSystemPromptAnimation} from './AdaptiveSystemPromptAnimation';

// Individual step components
import {UserRequestSequence} from './sequences/UserRequestSequence';
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
				id="ToolCallingAnimation"
				component={ToolCallingAnimation}
				durationInFrames={1200}
				fps={30}
				width={1920}
				height={1080}
			/>
			<Composition
				id="CRMQueryAnimation"
				component={CRMQueryAnimation}
				durationInFrames={1200}
				fps={30}
				width={1920}
				height={1080}
			/>
			<Composition
				id="SQLiteQueryAnimation"
				component={SQLiteQueryAnimation}
				durationInFrames={1200}
				fps={30}
				width={1920}
				height={1080}
			/>
			<Composition
				id="SQLiteViewOptimizationAnimation"
				component={SQLiteViewOptimizationAnimation}
				durationInFrames={1200}
				fps={30}
				width={1920}
				height={1080}
			/>
			<Composition
				id="ComprehensiveComparisonAnimation"
				component={ComprehensiveComparisonAnimation}
				durationInFrames={1200}
				fps={30}
				width={1920}
				height={1080}
			/>
			<Composition
				id="ContextBuildupAnimation"
				component={ContextBuildupAnimation}
				durationInFrames={840}
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

			{/* Weather API Animation - Individual Steps */}
			<Composition
				id="Weather-Step1-UserRequest"
				component={UserRequestSequence}
				durationInFrames={210}
				fps={30}
				width={1920}
				height={1080}
			/>
			<Composition
				id="Weather-Step2-ToolAnalysis"
				component={ToolAnalysisSequence}
				durationInFrames={270}
				fps={30}
				width={1920}
				height={1080}
			/>
			<Composition
				id="Weather-Step3-ToolExecution"
				component={ToolExecutionSequence}
				durationInFrames={330}
				fps={30}
				width={1920}
				height={1080}
			/>
			<Composition
				id="Weather-Step4-ResultIntegration"
				component={ResultIntegrationSequence}
				durationInFrames={330}
				fps={30}
				width={1920}
				height={1080}
			/>

			{/* CRM Animation - Individual Steps */}
			<Composition
				id="CRM-Step1-UserRequest"
				component={CRMUserRequestSequence}
				durationInFrames={180}
				fps={30}
				width={1920}
				height={1080}
			/>
			<Composition
				id="CRM-Step2-ToolAnalysis"
				component={CRMToolAnalysisSequence}
				durationInFrames={210}
				fps={30}
				width={1920}
				height={1080}
			/>
			<Composition
				id="CRM-Step3-ToolExecution"
				component={CRMToolExecutionSequence}
				durationInFrames={390}
				fps={30}
				width={1920}
				height={1080}
			/>
			<Composition
				id="CRM-Step4-ResultProcessing"
				component={CRMResultProcessingSequence}
				durationInFrames={450}
				fps={30}
				width={1920}
				height={1080}
			/>

			{/* SQLite Animation - Individual Steps */}
			<Composition
				id="SQLite-Step1-UserRequest"
				component={SQLiteUserRequestSequence}
				durationInFrames={150}
				fps={30}
				width={1920}
				height={1080}
			/>
			<Composition
				id="SQLite-Step2-SchemaDiscovery"
				component={SQLiteSchemaDiscoverySequence}
				durationInFrames={180}
				fps={30}
				width={1920}
				height={1080}
			/>
			<Composition
				id="SQLite-Step3-TableExploration"
				component={SQLiteTableExplorationSequence}
				durationInFrames={210}
				fps={30}
				width={1920}
				height={1080}
			/>
			<Composition
				id="SQLite-Step4-TargetedQuery"
				component={SQLiteTargetedQuerySequence}
				durationInFrames={270}
				fps={30}
				width={1920}
				height={1080}
			/>
			<Composition
				id="SQLite-Step5-FinalResponse"
				component={SQLiteFinalResponseSequence}
				durationInFrames={450}
				fps={30}
				width={1920}
				height={1080}
			/>

			{/* SQLite View Optimization Animation - Individual Steps */}
			<Composition
				id="SQLiteView-Step1-ViewCreation"
				component={ViewCreationSequence}
				durationInFrames={300}
				fps={30}
				width={1920}
				height={1080}
			/>
			<Composition
				id="SQLiteView-Step2-MultipleQueries"
				component={MultipleQueriesSequence}
				durationInFrames={450}
				fps={30}
				width={1920}
				height={1080}
			/>
			<Composition
				id="SQLiteView-Step3-PerformanceComparison"
				component={PerformanceComparisonSequence}
				durationInFrames={450}
				fps={30}
				width={1920}
				height={1080}
			/>

			{/* Comprehensive Comparison Animation - Individual Steps */}
			<Composition
				id="Comparison-Step1-TokenEfficiency"
				component={TokenEfficiencyComparisonSequence}
				durationInFrames={330}
				fps={30}
				width={1920}
				height={1080}
			/>
			<Composition
				id="Comparison-Step2-ViewPersistence"
				component={ViewPersistenceSequence}
				durationInFrames={300}
				fps={30}
				width={1920}
				height={1080}
			/>
			<Composition
				id="Comparison-Step3-ToolDiscovery"
				component={ToolDiscoverySequence}
				durationInFrames={300}
				fps={30}
				width={1920}
				height={1080}
			/>
			<Composition
				id="Comparison-Step4-FutureEfficiency"
				component={FutureEfficiencySequence}
				durationInFrames={300}
				fps={30}
				width={1920}
				height={1080}
			/>

			{/* Context Buildup Animation - Individual Steps */}
			<Composition
				id="Context-Step1-InitialMessages"
				component={InitialMessageSequence}
				durationInFrames={240}
				fps={30}
				width={1920}
				height={1080}
			/>
			<Composition
				id="Context-Step2-FirstToolCall"
				component={FirstToolCallSequence}
				durationInFrames={240}
				fps={30}
				width={1920}
				height={1080}
			/>
			<Composition
				id="Context-Step3-Summarization"
				component={ContextSummarizationSequence}
				durationInFrames={300}
				fps={30}
				width={1920}
				height={1080}
			/>
		</>
	);
};
