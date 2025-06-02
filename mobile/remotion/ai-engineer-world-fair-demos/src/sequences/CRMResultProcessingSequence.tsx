import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	spring,
	useCurrentFrame,
	useVideoConfig,
} from 'remotion';

export const CRMResultProcessingSequence: React.FC = () => {
	const frame = useCurrentFrame();
	const {fps} = useVideoConfig();

	const stepOpacity = interpolate(frame, [0, 30], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const llmOpacity = interpolate(frame, [0, 30], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const processingOpacity = interpolate(frame, [40, 70], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const inefficiencyOpacity = interpolate(frame, [100, 130], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const responseOpacity = interpolate(frame, [180, 210], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const arrowOpacity = interpolate(frame, [240, 270], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const arrowLength = interpolate(frame, [240, 270], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const userOpacity = interpolate(frame, [260, 290], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const finalMessageOpacity = interpolate(frame, [320, 360], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const statsOpacity = interpolate(frame, [380, 420], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const processingScale = spring({
		frame: frame - 40,
		fps,
		config: {
			damping: 8,
			stiffness: 80,
		},
	});

	const responseScale = spring({
		frame: frame - 180,
		fps,
		config: {
			damping: 8,
			stiffness: 80,
		},
	});

	const finalMessageScale = spring({
		frame: frame - 320,
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
					top: '8%',
					left: '50%',
					transform: 'translateX(-50%)',
					color: 'white',
					fontSize: '28px',
					fontWeight: 'bold',
					opacity: stepOpacity,
				}}
			>
				Step 4: LLM struggles to process massive response
			</div>

			{/* LLM with stress animation */}
			<div
				style={{
					position: 'absolute',
					top: '20%',
					left: '15%',
					opacity: llmOpacity,
				}}
			>
				<div
					style={{
						width: '120px',
						height: '120px',
						borderRadius: '20px',
						backgroundColor: frame > 70 ? '#e74c3c' : '#8e44ad',
						display: 'flex',
						alignItems: 'center',
						justifyContent: 'center',
						fontSize: '60px',
						color: 'white',
						boxShadow: '0 6px 20px rgba(0,0,0,0.2)',
						transform: frame > 70 && frame < 250 ? `scale(${1 + 0.1 * Math.sin(frame * 0.5)}) rotate(${Math.sin(frame * 0.3) * 3}deg)` : 'scale(1)',
					}}
				>
					{frame > 70 && frame < 250 ? '🤯' : '🧠'}
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

			{/* Processing thought */}
			<div
				style={{
					position: 'absolute',
					top: '15%',
					left: '35%',
					opacity: processingOpacity,
					transform: `scale(${processingScale})`,
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
						maxWidth: '300px',
						lineHeight: 1.4,
					}}
				>
					"I need to scan through all {36} companies to find just OpenAI... this is inefficient!"
				</div>
			</div>

			{/* Inefficiency visualization */}
			{inefficiencyOpacity > 0 && (
				<div
					style={{
						position: 'absolute',
						top: '40%',
						left: '20%',
						right: '20%',
						opacity: inefficiencyOpacity,
					}}
				>
					<div
						style={{
							backgroundColor: 'rgba(231, 76, 60, 0.9)',
							borderRadius: '15px',
							padding: '20px',
							color: 'white',
							textAlign: 'center',
						}}
					>
						<div style={{fontSize: '24px', fontWeight: 'bold', marginBottom: '15px'}}>
							🔥 Processing Overhead
						</div>
						<div style={{display: 'flex', justifyContent: 'space-around', fontSize: '16px'}}>
							<div>
								<div style={{fontSize: '20px', fontWeight: 'bold'}}>3,600+</div>
								<div>Input Tokens</div>
							</div>
							<div>
								<div style={{fontSize: '20px', fontWeight: 'bold'}}>36</div>
								<div>Companies Scanned</div>
							</div>
							<div>
								<div style={{fontSize: '20px', fontWeight: 'bold'}}>1</div>
								<div>Needed Result</div>
							</div>
						</div>
					</div>
				</div>
			)}

			{/* Generated Response */}
			<div
				style={{
					position: 'absolute',
					top: '65%',
					left: '25%',
					opacity: responseOpacity,
					transform: `scale(${responseScale})`,
				}}
			>
				<div
					style={{
						backgroundColor: '#27ae60',
						borderRadius: '20px',
						padding: '20px',
						boxShadow: '0 6px 20px rgba(0,0,0,0.15)',
						fontSize: '16px',
						color: 'white',
						maxWidth: '450px',
						lineHeight: 1.4,
					}}
				>
					<div style={{fontWeight: 'bold', marginBottom: '10px'}}>
						🤖 LLM Response (finally!):
					</div>
					"OpenAI contact information:
					<br />📧 contact@openai.com
					<br />📞 +1-555-7892
					<br />📍 3180 Tech St, Silicon Valley, CA"
				</div>
			</div>

			{/* Smooth Arrow to user */}
			<div
				style={{
					position: 'absolute',
					top: '85%',
					left: '15%',
					opacity: arrowOpacity,
				}}
			>
				<svg width="180" height="40" viewBox="0 0 180 40">
					<defs>
						<linearGradient id="returnArrowGradient" x1="0%" y1="0%" x2="100%" y2="0%">
							<stop offset="0%" stopColor="#27ae60" />
							<stop offset="100%" stopColor="#2ecc71" />
						</linearGradient>
					</defs>
					<path
						d={`M 10 20 L ${10 + 140 * arrowLength} 20`}
						stroke="url(#returnArrowGradient)"
						strokeWidth="4"
						strokeLinecap="round"
						fill="none"
					/>
					<polygon
						points={`${10 + 140 * arrowLength},20 ${10 + 140 * arrowLength - 12},15 ${10 + 140 * arrowLength - 12},25`}
						fill="url(#returnArrowGradient)"
						opacity={arrowLength}
					/>
				</svg>
			</div>

			{/* User */}
			<div
				style={{
					position: 'absolute',
					top: '80%',
					right: '15%',
					opacity: userOpacity,
				}}
			>
				<div
					style={{
						width: '100px',
						height: '100px',
						borderRadius: '50%',
						backgroundColor: '#3498db',
						display: 'flex',
						alignItems: 'center',
						justifyContent: 'center',
						fontSize: '50px',
						color: 'white',
						boxShadow: '0 6px 20px rgba(0,0,0,0.2)',
					}}
				>
					👤
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
					User
				</div>
			</div>

			{/* Final message about inefficiency */}
			<div
				style={{
					position: 'absolute',
					bottom: '8%',
					left: '50%',
					transform: `translateX(-50%) scale(${finalMessageScale})`,
					opacity: finalMessageOpacity,
				}}
			>
				<div
					style={{
						backgroundColor: 'rgba(231, 76, 60, 0.9)',
						borderRadius: '25px',
						padding: '20px 40px',
						color: 'white',
						fontSize: '22px',
						fontWeight: 'bold',
						textAlign: 'center',
						boxShadow: '0 6px 20px rgba(0,0,0,0.3)',
					}}
				>
					❌ Inefficient! Got the answer, but wasted thousands of tokens
				</div>
			</div>

			{/* Final stats */}
			{statsOpacity > 0 && (
				<div
					style={{
						position: 'absolute',
						bottom: '2%',
						left: '50%',
						transform: 'translateX(-50%)',
						opacity: statsOpacity,
						color: 'rgba(255,255,255,0.8)',
						fontSize: '16px',
						textAlign: 'center',
					}}
				>
					Better approach: Use search_companies(name="OpenAI") instead!
				</div>
			)}
		</AbsoluteFill>
	);
};
