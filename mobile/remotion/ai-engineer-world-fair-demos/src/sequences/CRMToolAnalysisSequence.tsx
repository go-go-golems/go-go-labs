import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	spring,
	useCurrentFrame,
	useVideoConfig,
} from 'remotion';

export const CRMToolAnalysisSequence: React.FC = () => {
	const frame = useCurrentFrame();
	const {fps} = useVideoConfig();

	const stepOpacity = interpolate(frame, [0, 30], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const llmOpacity = interpolate(frame, [0, 30], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const thoughtBubbleOpacity = interpolate(frame, [40, 70], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const toolsHeaderOpacity = interpolate(frame, [80, 110], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const tool1Opacity = interpolate(frame, [110, 140], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const tool2Opacity = interpolate(frame, [130, 160], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const selectionGlow = interpolate(frame, [150, 180], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const thoughtScale = spring({
		frame: frame - 40,
		fps,
		config: {
			damping: 10,
			stiffness: 100,
		},
	});

	return (
		<AbsoluteFill
			style={{
				background: 'linear-gradient(135deg, #e74c3c 0%, #c0392b 100%)',
				fontFamily: 'Arial, sans-serif',
			}}
		>
			{/* Step indicator */}
			<div
				style={{
					position: 'absolute',
					top: '22%',
					left: '50%',
					transform: 'translateX(-50%)',
					color: 'white',
					fontSize: '28px',
					fontWeight: 'bold',
					opacity: stepOpacity,
				}}
			>
				Step 2: LLM chooses available tool
			</div>

			{/* LLM with thinking animation */}
			<div
				style={{
					position: 'absolute',
					top: '35%',
					left: '20%',
					opacity: llmOpacity,
				}}
			>
				<div
					style={{
						width: '120px',
						height: '120px',
						borderRadius: '20px',
						backgroundColor: '#8e44ad',
						display: 'flex',
						alignItems: 'center',
						justifyContent: 'center',
						fontSize: '60px',
						color: 'white',
						boxShadow: '0 6px 20px rgba(0,0,0,0.2)',
						transform: frame > 70 ? `scale(${1 + 0.05 * Math.sin(frame * 0.3)})` : 'scale(1)',
					}}
				>
					🧠
				</div>
				<div
					style={{
						textAlign: 'center',
						color: 'white',
						marginTop: '10px',
						fontSize: '18px',
						fontWeight: 'bold',
					}}
				>
					LLM
				</div>
			</div>

			{/* Thought bubble */}
			<div
				style={{
					position: 'absolute',
					top: '30%',
					left: '38%',
					opacity: thoughtBubbleOpacity,
					transform: `scale(${thoughtScale})`,
				}}
			>
				<div
					style={{
						backgroundColor: 'rgba(255,255,255,0.95)',
						borderRadius: '20px',
						padding: '20px',
						boxShadow: '0 6px 20px rgba(0,0,0,0.15)',
						fontSize: '16px',
						color: '#2c3e50',
						maxWidth: '280px',
						lineHeight: 1.4,
					}}
				>
					"I need company data... I'll use the CRM tool to find OpenAI."
				</div>
			</div>

			{/* Available Tools Header */}
			<div
				style={{
					position: 'absolute',
					top: '60%',
					left: '50%',
					transform: 'translateX(-50%)',
					color: 'white',
					fontSize: '24px',
					fontWeight: 'bold',
					opacity: toolsHeaderOpacity,
				}}
			>
				Available Tools:
			</div>

			{/* Tool 1: CRM Database */}
			<div
				style={{
					position: 'absolute',
					top: '70%',
					left: '25%',
					opacity: tool1Opacity,
				}}
			>
				<div
					style={{
						backgroundColor: selectionGlow > 0.5 ? '#27ae60' : '#e74c3c',
						borderRadius: '15px',
						padding: '25px',
						boxShadow: `0 6px 20px rgba(0,0,0,0.2) ${selectionGlow > 0.5 ? ', 0 0 25px rgba(39, 174, 96, 0.6)' : ''}`,
						textAlign: 'center',
						color: 'white',
						fontSize: '16px',
						fontWeight: 'bold',
						minWidth: '180px',
						transform: selectionGlow > 0.5 ? 'scale(1.05)' : 'scale(1)',
						transition: 'transform 0.3s ease',
					}}
				>
					<div style={{fontSize: '35px', marginBottom: '10px'}}>🗄️</div>
					get_crm_companies
					<div style={{fontSize: '12px', marginTop: '5px', opacity: 0.9}}>
						Returns ALL company data
					</div>
				</div>
			</div>

			{/* Tool 2: Web Search */}
			<div
				style={{
					position: 'absolute',
					top: '70%',
					right: '25%',
					opacity: tool2Opacity,
				}}
			>
				<div
					style={{
						backgroundColor: '#3498db',
						borderRadius: '15px',
						padding: '25px',
						boxShadow: '0 6px 20px rgba(0,0,0,0.2)',
						textAlign: 'center',
						color: 'white',
						fontSize: '16px',
						fontWeight: 'bold',
						minWidth: '180px',
					}}
				>
					<div style={{fontSize: '35px', marginBottom: '10px'}}>🔍</div>
					web_search
					<div style={{fontSize: '12px', marginTop: '5px', opacity: 0.9}}>
						Search the web
					</div>
				</div>
			</div>

			{/* Selection indicator */}
			{selectionGlow > 0 && (
				<div
					style={{
						position: 'absolute',
						top: '90%',
						left: '50%',
						transform: 'translateX(-50%)',
						color: '#27ae60',
						fontSize: '20px',
						fontWeight: 'bold',
						opacity: selectionGlow,
						textShadow: '0 0 10px rgba(39, 174, 96, 0.8)',
					}}
				>
					✓ CRM database selected - this will return ALL companies!
				</div>
			)}
		</AbsoluteFill>
	);
};
