import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	useCurrentFrame,
} from 'remotion';
import {
	InteractionSequence,
	MessageDefinition,
	MessageTypeConfig,
	StateTransition,
	OverlayElement,
	InteractionState,
	resolveContent,
	FONT_SIZES,
} from '../types/InteractionDSL';

interface MessageProps {
	message: MessageDefinition;
	config: MessageTypeConfig;
	opacity: number;
	fadeOut?: boolean;
	state: InteractionState;
}

const Message: React.FC<MessageProps> = ({message, config, opacity, fadeOut = false, state}) => {
	const finalOpacity = message.customOpacity ?? opacity;
	
	// Resolve dynamic content
	const content = resolveContent(message.content, state);
	const icon = resolveContent(config.icon, state);
	const label = resolveContent(config.label, state);
	
	// Check if content is React component
	const isReactContent = message.isReactContent || React.isValidElement(content);
	
	return (
		<div
			style={{
				opacity: finalOpacity,
				backgroundColor: config.bg,
				borderRadius: '10px',
				padding: config.padding,
				color: 'white',
				fontSize: config.fontSize,
				margin: '12px 0',
				display: 'flex',
				alignItems: isReactContent ? 'flex-start' : 'center',
				gap: '10px',
				boxShadow: config.boxShadow,
				border: config.border,
				minHeight: 'fit-content',
				height: 'auto',
			}}
		>
			<span style={{
				fontSize: FONT_SIZES.icon,
				marginTop: isReactContent ? '2px' : '0'
			}}>{icon}</span>
			<div style={{ flex: 1 }}>
				<div style={{
					fontSize: FONT_SIZES.label,
					opacity: 0.8, 
					marginBottom: '3px',
					fontWeight: config.fontWeight === 'bold' || config.fontWeight === '500' ? 'bold' : 'normal'
				}}>
					{label}
				</div>
				{isReactContent ? (
					<div style={{ width: '100%' }}>
						{content}
					</div>
				) : (
					<div style={{
						fontSize: config.fontSize,
						lineHeight: 1.2,
						fontWeight: config.fontWeight,
						fontStyle: config.fontStyle,
						whiteSpace: 'pre-wrap'
					}}>
						{content}
					</div>
				)}
			</div>
		</div>
	);
};

interface InteractionRendererProps {
	sequence: InteractionSequence;
	background?: string;
	containerStyle?: React.CSSProperties;
}

export const InteractionRenderer: React.FC<InteractionRendererProps> = ({
	sequence,
	background = 'linear-gradient(135deg, #2c3e50 0%, #34495e 100%)',
	containerStyle = {},
}) => {
	const frame = useCurrentFrame();

	// Calculate state opacities
	const getStateOpacity = (state: StateTransition): number => {
		return interpolate(frame, [state.startFrame, state.endFrame], [0, 1], {
			extrapolateRight: 'clamp',
		});
	};

	// Get current active states
	const getActiveStates = (): string[] => {
		return sequence.states
			.filter(state => frame >= state.startFrame && frame <= state.endFrame)
			.map(state => state.name);
	};

	// Get fade out states
	const getFadeOutStates = (): string[] => {
		return sequence.states
			.filter(state => frame > state.endFrame)
			.map(state => state.name);
	};

	// Token counter calculation
	const calculateTokens = (): { current: number; isOptimized: boolean } => {
		if (!sequence.tokenCounter?.enabled) return { current: 0, isOptimized: false };

		const activeStates = getActiveStates();
		const { tokenCounter } = sequence;
		
		// Find the most recent state with token count
		let currentTokens = tokenCounter.initialTokens;
		for (const state of sequence.states) {
			if (frame >= state.startFrame && tokenCounter.stateTokenCounts[state.name]) {
				currentTokens = tokenCounter.stateTokenCounts[state.name];
			}
		}

		const isOptimized = tokenCounter.optimizedStates?.some(state => activeStates.includes(state)) ?? false;
		return { current: currentTokens, isOptimized };
	};

	const { current: currentTokens, isOptimized } = calculateTokens();

	// Create interaction state for dynamic content
	const interactionState: InteractionState = {
		currentFrame: frame,
		activeStates: getActiveStates(),
		fadeOutStates: getFadeOutStates(),
		tokenCount: currentTokens,
		isOptimized,
		customData: {},
	};

	// Calculate message visibility and opacity
	const getMessageOpacity = (message: MessageDefinition): { opacity: number; fadeOut: boolean } => {
		const activeStates = getActiveStates();
		const fadeOutStates = getFadeOutStates();
		
		// Check if message should fade out
		const shouldFadeOut = message.fadeOutStates?.some(state => fadeOutStates.includes(state)) ?? false;
		
		// Check if any of the message's visible states have started
		const hasStarted = message.visibleStates.some(stateName => {
			const state = sequence.states.find(s => s.name === stateName);
			return state && frame >= state.startFrame;
		});
		
		// Binary opacity: either fully visible (1) or completely hidden (0)
		const opacity = hasStarted ? 1 : 0;
		
		return { opacity, fadeOut: shouldFadeOut };
	};

	// Distribute messages across columns
	const distributeMessages = (): { left: MessageDefinition[]; right: MessageDefinition[] } => {
		if (sequence.layout.columns === 1) {
			return { left: sequence.messages, right: [] };
		}

		if (!sequence.layout.autoFill) {
			// Manual column assignment
			return {
				left: sequence.messages.filter(msg => msg.column === 'left'),
				right: sequence.messages.filter(msg => msg.column === 'right'),
			};
		}

		// Auto-fill from left to right
		const maxPerColumn = sequence.layout.maxMessagesPerColumn ?? Math.ceil(sequence.messages.length / 2);
		return {
			left: sequence.messages.slice(0, maxPerColumn),
			right: sequence.messages.slice(maxPerColumn),
		};
	};

	const { left: leftMessages, right: rightMessages } = distributeMessages();

	// Container opacity
	const containerOpacity = interpolate(frame, [0, 30], [0, 1], {
		extrapolateRight: 'clamp',
	});

	// Calculate overlay visibility
	const getOverlayOpacity = (overlay: OverlayElement): number => {
		const hasStarted = overlay.visibleStates.some(stateName => {
			const state = sequence.states.find(s => s.name === stateName);
			return state && frame >= state.startFrame;
		});
		
		if (!hasStarted) return 0;

		// Find the earliest state that has started for opacity calculation
		const startedStates = overlay.visibleStates
			.map(stateName => sequence.states.find(s => s.name === stateName))
			.filter(state => state && frame >= state.startFrame)
			.sort((a, b) => a!.startFrame - b!.startFrame);

		const relevantState = startedStates[0];
		return relevantState ? getStateOpacity(relevantState) : 1;
	};

	// Resolve dynamic title and subtitle
	const title = resolveContent(sequence.title, interactionState);
	const subtitle = sequence.subtitle ? resolveContent(sequence.subtitle, interactionState) : undefined;

	return (
		<AbsoluteFill
			style={{
				background,
				fontFamily: 'Arial, sans-serif',
				...containerStyle,
			}}
		>

			{/* Context Container */}
			<div
				style={{
					position: 'absolute',
					left: '50%',
					transform: 'translate(-50%, 0)',
					width: '1280px',
					border: '2px solid rgba(255, 255, 255, 0.3)',
					borderRadius: '16px',
					backgroundColor: 'rgba(255, 255, 255, 0.05)',
					padding: '20px',
					opacity: containerOpacity,
				}}
			>

				{/* Messages Layout */}
				{sequence.layout.columns === 1 ? (
					<div style={{overflowY: 'auto', paddingRight: '10px', minHeight: 'fit-content'}}>
						{leftMessages.map((message) => {
							const { opacity, fadeOut } = getMessageOpacity(message);
							const config = sequence.messageTypes[message.type];
							if (!config || opacity === 0) return null;

							return (
								<Message
									key={message.id}
									message={message}
									config={config}
									opacity={opacity}
									fadeOut={fadeOut}
									state={interactionState}
								/>
							);
						})}
					</div>
				) : (
					<div style={{display: 'flex', gap: '20px',  overflowY: 'auto', minHeight: 'fit-content', overflowX: 'hidden'}}>
						{/* Left Column */}
						<div style={{flex: 1}}>
							{leftMessages.map((message) => {
								const { opacity, fadeOut } = getMessageOpacity(message);
								const config = sequence.messageTypes[message.type];
								if (!config || opacity === 0) return null;

								return (
									<Message
										key={message.id}
										message={message}
										config={config}
										opacity={opacity}
										fadeOut={fadeOut}
										state={interactionState}
									/>
								);
							})}
						</div>

						{/* Right Column */}
						<div style={{flex: 1}}>
							{rightMessages.map((message) => {
								const { opacity, fadeOut } = getMessageOpacity(message);
								const config = sequence.messageTypes[message.type];
								if (!config || opacity === 0) return null;

								return (
									<Message
										key={message.id}
										message={message}
										config={config}
										opacity={opacity}
										fadeOut={fadeOut}
										state={interactionState}
									/>
								);
							})}
						</div>
					</div>
				)}
			</div>

			{/* Overlays */}
		</AbsoluteFill>
	);
}; 