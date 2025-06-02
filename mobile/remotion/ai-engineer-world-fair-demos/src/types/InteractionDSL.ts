// DSL for declarative interaction scripting

export interface MessageTypeConfig {
	bg: string;
	icon: string;
	label: string;
	fontSize?: string;
	padding?: string;
	boxShadow?: string;
	border?: string;
	fontWeight?: string;
	fontStyle?: string;
}

export interface MessageTypeRegistry {
	[key: string]: MessageTypeConfig;
}

export interface StateTransition {
	name: string;
	startFrame: number;
	endFrame: number;
	easing?: 'linear' | 'ease-in' | 'ease-out' | 'ease-in-out';
}

export interface MessageDefinition {
	id: string;
	type: string;
	content: string;
	column?: 'left' | 'right' | 'auto'; // auto fills left to right
	visibleStates: string[]; // which states this message is visible in
	fadeOutStates?: string[]; // states where this message should fade out
	customOpacity?: number; // override opacity
}

export interface OverlayElement {
	id: string;
	content: string;
	position: {
		bottom?: string;
		top?: string;
		left?: string;
		right?: string;
	};
	style?: React.CSSProperties;
	visibleStates: string[];
}

export interface InteractionSequence {
	title: string;
	subtitle?: string;
	messageTypes: MessageTypeRegistry;
	states: StateTransition[];
	messages: MessageDefinition[];
	overlays?: OverlayElement[];
	layout: {
		columns: number;
		autoFill?: boolean; // automatically distribute messages across columns
		maxMessagesPerColumn?: number;
	};
	tokenCounter?: {
		enabled: boolean;
		initialTokens: number;
		maxTokens: number;
		stateTokenCounts: { [stateName: string]: number };
		optimizedStates?: string[]; // states that show optimization
	};
}

// Helper functions for creating common configurations
export const createMessageType = (
	bg: string,
	icon: string,
	label: string,
	options: Partial<MessageTypeConfig> = {}
): MessageTypeConfig => ({
	bg,
	icon,
	label,
	fontSize: '13px',
	padding: '12px 15px',
	boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
	border: 'none',
	fontWeight: 'normal',
	fontStyle: 'normal',
	...options,
});

export const createState = (
	name: string,
	startFrame: number,
	duration: number = 10,
	easing: StateTransition['easing'] = 'linear'
): StateTransition => ({
	name,
	startFrame,
	endFrame: startFrame + duration,
	easing,
});

export const createMessage = (
	id: string,
	type: string,
	content: string,
	visibleStates: string[],
	options: Partial<MessageDefinition> = {}
): MessageDefinition => ({
	id,
	type,
	content,
	visibleStates,
	column: 'auto',
	...options,
});

// Predefined message type sets
export const DEFAULT_MESSAGE_TYPES: MessageTypeRegistry = {
	user: createMessageType('#3498db', '👤', 'User'),
	assistant: createMessageType('#9b59b6', '🧠', 'Assistant'),
	assistant_cot: createMessageType('#e74c3c', '🤔', 'Chain of Thought', {
		fontSize: '11px',
		padding: '8px 12px',
		boxShadow: '0 3px 10px rgba(0,0,0,0.2)',
		border: '1px solid rgba(255, 255, 255, 0.2)',
		fontStyle: 'italic',
	}),
	assistant_diary: createMessageType('#8e44ad', '📔', 'Diary', {
		fontSize: '11px',
		padding: '8px 12px',
		boxShadow: '0 3px 10px rgba(0,0,0,0.2)',
		border: '1px solid rgba(255, 255, 255, 0.2)',
		fontWeight: 'bold',
		fontStyle: 'italic',
	}),
	tool_use: createMessageType('#e67e22', '⚡', 'Tool Use'),
	tool_result: createMessageType('#27ae60', '📊', 'Tool Result'),
	summary: createMessageType('#6c3483', '📝', 'Summary', {
		fontSize: '13px',
		padding: '14px 18px',
		boxShadow: '0 4px 15px rgba(108, 52, 131, 0.4)',
		border: '1px solid rgba(255, 255, 255, 0.2)',
		fontWeight: '500',
	}),
}; 