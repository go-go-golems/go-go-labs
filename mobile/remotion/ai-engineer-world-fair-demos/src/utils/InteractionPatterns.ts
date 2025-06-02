import {
	InteractionSequence,
	MessageDefinition,
	StateTransition,
	createState,
	createMessage,
	DEFAULT_MESSAGE_TYPES,
} from '../types/InteractionDSL';

// Common interaction patterns

export const createLinearSequence = (
	title: string,
	messages: Array<{
		id: string;
		type: string;
		content: string;
		duration?: number;
	}>,
	options: {
		startFrame?: number;
		defaultDuration?: number;
		columns?: number;
		tokenCounter?: boolean;
	} = {}
): InteractionSequence => {
	const {
		startFrame = 30,
		defaultDuration = 15,
		columns = 1,
		tokenCounter = false,
	} = options;

	// Create states for each message
	const states: StateTransition[] = [
		createState('container', 0, 30),
	];

	let currentFrame = startFrame;
	const messageStates: string[] = [];

	messages.forEach((msg, index) => {
		const stateName = `message_${index}`;
		const duration = msg.duration ?? defaultDuration;
		states.push(createState(stateName, currentFrame, duration));
		messageStates.push(stateName);
		currentFrame += duration;
	});

	// Create message definitions
	const messageDefinitions: MessageDefinition[] = messages.map((msg, index) => {
		const visibleStates = messageStates.slice(index);
		return createMessage(msg.id, msg.type, msg.content, visibleStates);
	});

	return {
		title,
		messageTypes: DEFAULT_MESSAGE_TYPES,
		states,
		messages: messageDefinitions,
		layout: {
			columns,
			autoFill: true,
		},
		tokenCounter: tokenCounter ? {
			enabled: true,
			initialTokens: 100,
			maxTokens: 128000,
			stateTokenCounts: {},
		} : undefined,
	};
};

export const createConversationFlow = (
	title: string,
	exchanges: Array<{
		user: string;
		assistant: string;
		thinking?: string;
		tools?: Array<{ call: string; result: string }>;
	}>,
	options: {
		startFrame?: number;
		exchangeDuration?: number;
		columns?: number;
	} = {}
): InteractionSequence => {
	const {
		startFrame = 30,
		exchangeDuration = 60,
		columns = 2,
	} = options;

	const states: StateTransition[] = [
		createState('container', 0, 30),
	];

	const messages: MessageDefinition[] = [];
	let currentFrame = startFrame;
	let messageId = 0;

	exchanges.forEach((exchange, exchangeIndex) => {
		const exchangeStates: string[] = [];
		
		// User message
		const userState = `exchange_${exchangeIndex}_user`;
		states.push(createState(userState, currentFrame, 10));
		exchangeStates.push(userState);
		currentFrame += 10;

		messages.push(createMessage(
			`msg_${messageId++}`,
			'user',
			exchange.user,
			[userState, ...exchangeStates]
		));

		// Thinking (if provided)
		if (exchange.thinking) {
			const thinkingState = `exchange_${exchangeIndex}_thinking`;
			states.push(createState(thinkingState, currentFrame, 10));
			exchangeStates.push(thinkingState);
			currentFrame += 10;

			messages.push(createMessage(
				`msg_${messageId++}`,
				'assistant_cot',
				exchange.thinking,
				[thinkingState, ...exchangeStates]
			));
		}

		// Tools (if provided)
		if (exchange.tools) {
			exchange.tools.forEach((tool, toolIndex) => {
				const toolCallState = `exchange_${exchangeIndex}_tool_${toolIndex}_call`;
				const toolResultState = `exchange_${exchangeIndex}_tool_${toolIndex}_result`;
				
				states.push(createState(toolCallState, currentFrame, 8));
				exchangeStates.push(toolCallState);
				currentFrame += 8;

				states.push(createState(toolResultState, currentFrame, 8));
				exchangeStates.push(toolResultState);
				currentFrame += 8;

				messages.push(createMessage(
					`msg_${messageId++}`,
					'tool_use',
					tool.call,
					[toolCallState, ...exchangeStates]
				));

				messages.push(createMessage(
					`msg_${messageId++}`,
					'tool_result',
					tool.result,
					[toolResultState, ...exchangeStates]
				));
			});
		}

		// Assistant response
		const assistantState = `exchange_${exchangeIndex}_assistant`;
		states.push(createState(assistantState, currentFrame, 15));
		exchangeStates.push(assistantState);
		currentFrame += 15;

		messages.push(createMessage(
			`msg_${messageId++}`,
			'assistant',
			exchange.assistant,
			[assistantState, ...exchangeStates]
		));

		// Update all messages in this exchange to be visible in all subsequent states
		const exchangeMessageIds = messages.slice(-1 - (exchange.thinking ? 1 : 0) - (exchange.tools?.length || 0) * 2 - 1);
		exchangeMessageIds.forEach(msg => {
			msg.visibleStates.push(...exchangeStates);
		});
	});

	return {
		title,
		messageTypes: DEFAULT_MESSAGE_TYPES,
		states,
		messages,
		layout: {
			columns,
			autoFill: true,
		},
	};
};

export const createSummarizationFlow = (
	title: string,
	originalMessages: Array<{
		id: string;
		type: string;
		content: string;
	}>,
	summary: {
		content: string;
		type?: string;
	},
	options: {
		fadeOutFrame?: number;
		summaryFrame?: number;
		columns?: number;
	} = {}
): InteractionSequence => {
	const {
		fadeOutFrame = 200,
		summaryFrame = 250,
		columns = 2,
	} = options;

	const states: StateTransition[] = [
		createState('container', 0, 30),
		createState('showOriginal', 30, fadeOutFrame - 30),
		createState('fadeOut', fadeOutFrame, summaryFrame - fadeOutFrame),
		createState('showSummary', summaryFrame, 50),
	];

	const messages: MessageDefinition[] = [
		...originalMessages.map((msg, index) => 
			createMessage(
				msg.id,
				msg.type,
				msg.content,
				['showOriginal'],
				{ fadeOutStates: ['fadeOut'] }
			)
		),
		createMessage(
			'summary',
			summary.type || 'summary',
			summary.content,
			['showSummary']
		),
	];

	return {
		title,
		messageTypes: DEFAULT_MESSAGE_TYPES,
		states,
		messages,
		layout: {
			columns,
			autoFill: true,
		},
		tokenCounter: {
			enabled: true,
			initialTokens: originalMessages.length * 200,
			maxTokens: 128000,
			stateTokenCounts: {
				'showOriginal': originalMessages.length * 200,
				'showSummary': 400,
			},
			optimizedStates: ['showSummary'],
		},
	};
};

// Preset message builders
export const userMessage = (content: string) => ({ type: 'user', content });
export const assistantMessage = (content: string) => ({ type: 'assistant', content });
export const thinkingMessage = (content: string) => ({ type: 'assistant_cot', content });
export const diaryMessage = (content: string) => ({ type: 'assistant_diary', content });
export const toolCall = (content: string) => ({ type: 'tool_use', content });
export const toolResult = (content: string) => ({ type: 'tool_result', content });
export const summaryMessage = (content: string) => ({ type: 'summary', content }); 