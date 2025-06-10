import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	spring,
	useCurrentFrame,
	useVideoConfig,
} from 'remotion';

export const SQLiteFinalResponseSequence: React.FC = () => {
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

	const responseOpacity = interpolate(frame, [90, 120], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const arrowOpacity = interpolate(frame, [150, 180], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const arrowLength = interpolate(frame, [150, 180], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const userOpacity = interpolate(frame, [170, 200], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const summaryOpacity = interpolate(frame, [230, 270], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const statsOpacity = interpolate(frame, [300, 340], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const finalMessageOpacity = interpolate(frame, [370, 420], [0, 1], {
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
		frame: frame - 90,
		fps,
		config: {
			damping: 8,
			stiffness: 80,
		},
	});

	const summaryScale = spring({
		frame: frame - 230,
		fps,
		config: {
			damping: 10,
			stiffness: 100,
		},
	});

	const finalMessageScale = spring({
		frame: frame - 370,
		fps,
		config: {
			damping: 10,
			stiffness: 100,
		},
	});

	return (
		<AbsoluteFill
			style={{
				background: 'linear-gradient(135deg, #2c3e50 0%, #34495e 100%)',
				fontFamily: 'Arial, sans-serif',
			}}
		>
			{/* Step indicator */}
			<div
				style={{
					position: 'absolute',
					top: '10%',
					left: '50%',
					transform: 'translateX(-50%)',
					color: 'white',
					fontSize: '28px',
					fontWeight: 'bold',
					opacity: stepOpacity,
				}}
			>
				Step 5: Efficient response delivery
			</div>

			{/* LLM */}
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
						backgroundColor: '#27ae60',
						display: 'flex',
						alignItems: 'center',
						justifyContent: 'center',
						fontSize: '60px',
						color: 'white',
						boxShadow: '0 6px 20px rgba(0,0,0,0.2)',
						transform: frame > 40 && frame < 120 ? `scale(${1 + 0.03 * Math.sin(frame * 0.2)})` : 'scale(1)',
					}}
				>
					{frame > 120 ? '😊' : '🧠'}
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
					"Perfect! I got exactly the data I needed. Now I can give a precise answer to the user."
				</div>
			</div>

			{/* Generated Response */}
			<div
				style={{
					position: 'absolute',
					top: '45%',
					left: '25%',
					opacity: responseOpacity,
					transform: `scale(${responseScale})`,
				}}
			>
				<div
					style={{
						backgroundColor: '#27ae60',
						borderRadius: '20px',
						padding: '25px',
						boxShadow: '0 8px 25px rgba(0,0,0,0.3)',
						fontSize: '18px',
						color: 'white',
						maxWidth: '500px',
						lineHeight: 1.5,
					}}
				>
					<div style={{fontWeight: 'bold', marginBottom: '15px', fontSize: '20px'}}>
						🤖 LLM Response:
					</div>
					"Based on the database query, John Smith placed <strong>7 orders</strong> last month (November 2024). 
					I analyzed the customer and order tables to get this precise count."
				</div>
			</div>

			{/* Arrow to user */}
			<div
				style={{
					position: 'absolute',
					top: '75%',
					left: '15%',
					opacity: arrowOpacity,
				}}
			>
				<svg width="200" height="40" viewBox="0 0 200 40">
					<defs>
						<linearGradient id="finalArrowGradient" x1="0%" y1="0%" x2="100%" y2="0%">
							<stop offset="0%" stopColor="#27ae60" />
							<stop offset="100%" stopColor="#2ecc71" />
						</linearGradient>
					</defs>
					<path
						d={`M 10 20 L ${10 + 160 * arrowLength} 20`}
						stroke="url(#finalArrowGradient)"
						strokeWidth="5"
						strokeLinecap="round"
						fill="none"
					/>
					<polygon
						points={`${10 + 160 * arrowLength},20 ${10 + 160 * arrowLength - 15},12 ${10 + 160 * arrowLength - 15},28`}
						fill="url(#finalArrowGradient)"
						opacity={arrowLength}
					/>
				</svg>
			</div>

			{/* User */}
			<div
				style={{
					position: 'absolute',
					top: '70%',
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

			{/* Summary box */}
			<div
				style={{
					position: 'absolute',
					bottom: '25%',
					left: '50%',
					transform: `translateX(-50%) scale(${summaryScale})`,
					opacity: summaryOpacity,
				}}
			>
				<div
					style={{
						backgroundColor: 'rgba(44, 62, 80, 0.95)',
						borderRadius: '20px',
						padding: '25px',
						color: 'white',
						fontSize: '16px',
						textAlign: 'center',
						boxShadow: '0 8px 25px rgba(0,0,0,0.3)',
						border: '2px solid #3498db',
						maxWidth: '600px',
					}}
				>
					<div style={{fontSize: '24px', marginBottom: '15px'}}>📊 Process Summary</div>
					<div style={{display: 'grid', gridTemplateColumns: '1fr 1fr 1fr 1fr', gap: '20px', fontSize: '14px'}}>
						<div>
							<div style={{fontSize: '18px', fontWeight: 'bold', color: '#3498db'}}>4</div>
							<div>Tool Calls</div>
						</div>
						<div>
							<div style={{fontSize: '18px', fontWeight: 'bold', color: '#27ae60'}}>Smart</div>
							<div>Exploration</div>
						</div>
						<div>
							<div style={{fontSize: '18px', fontWeight: 'bold', color: '#f39c12'}}>Precise</div>
							<div>Result</div>
						</div>
						<div>
							<div style={{fontSize: '18px', fontWeight: 'bold', color: '#e74c3c'}}>Minimal</div>
							<div>Tokens</div>
						</div>
					</div>
				</div>
			</div>

			{/* Efficiency stats */}
			{statsOpacity > 0 && (
				<div
					style={{
						position: 'absolute',
						bottom: '12%',
						left: '50%',
						transform: 'translateX(-50%)',
						opacity: statsOpacity,
					}}
				>
					<div
						style={{
							backgroundColor: 'rgba(39, 174, 96, 0.9)',
							borderRadius: '15px',
							padding: '15px 30px',
							color: 'white',
							fontSize: '14px',
							textAlign: 'center',
							display: 'flex',
							gap: '30px',
							alignItems: 'center',
						}}
					>
						<div>
							<div style={{fontWeight: 'bold'}}>Schema Discovery:</div>
							<div>2 exploration queries</div>
						</div>
						<div>
							<div style={{fontWeight: 'bold'}}>Final Query:</div>
							<div>1 targeted result</div>
						</div>
						<div>
							<div style={{fontWeight: 'bold'}}>Total Tokens:</div>
							<div>~300 (vs 3,600+ in bulk approach)</div>
						</div>
					</div>
				</div>
			)}

			{/* Final message */}
			<div
				style={{
					position: 'absolute',
					bottom: '2%',
					left: '50%',
					transform: `translateX(-50%) scale(${finalMessageScale})`,
					opacity: finalMessageOpacity,
				}}
			>
				<div
					style={{
						backgroundColor: 'rgba(39, 174, 96, 0.9)',
						borderRadius: '25px',
						padding: '20px 40px',
						color: 'white',
						fontSize: '20px',
						fontWeight: 'bold',
						textAlign: 'center',
						boxShadow: '0 6px 20px rgba(0,0,0,0.3)',
					}}
				>
					✅ Intelligent multi-step approach: Maximum precision, minimal waste!
				</div>
			</div>
		</AbsoluteFill>
	);
};
