import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	spring,
	useCurrentFrame,
	useVideoConfig,
} from 'remotion';

export const ToolExecutionSequence: React.FC = () => {
	const frame = useCurrentFrame();
	const {fps} = useVideoConfig();

	const stepOpacity = interpolate(frame, [0, 30], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const llmOpacity = interpolate(frame, [0, 30], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const requestOpacity = interpolate(frame, [60, 90], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const arrowProgress = interpolate(frame, [90, 150], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const apiOpacity = interpolate(frame, [120, 150], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const loadingOpacity = interpolate(frame, [150, 180], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const responseOpacity = interpolate(frame, [210, 240], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const returnArrowProgress = interpolate(frame, [240, 300], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const requestScale = spring({
		frame: frame - 60,
		fps,
		config: {
			damping: 8,
			stiffness: 80,
		},
	});

	const responseScale = spring({
		frame: frame - 210,
		fps,
		config: {
			damping: 8,
			stiffness: 80,
		},
	});

	return (
		<AbsoluteFill
			style={{
				background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
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
					fontSize: '32px',
					fontWeight: 'bold',
					opacity: stepOpacity,
				}}
			>
				Step 3: Tool execution
			</div>

			{/* LLM */}
			<div
				style={{
					position: 'absolute',
					top: '25%',
					left: '15%',
					opacity: llmOpacity,
				}}
			>
				<div
					style={{
						width: '120px',
						height: '120px',
						borderRadius: '20px',
						backgroundColor: '#7B68EE',
						display: 'flex',
						alignItems: 'center',
						justifyContent: 'center',
						fontSize: '60px',
						color: 'white',
						boxShadow: '0 4px 12px rgba(0,0,0,0.3)',
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

			{/* API Request */}
			<div
				style={{
					position: 'absolute',
					top: '15%',
					left: '35%',
					opacity: requestOpacity,
					transform: `scale(${requestScale})`,
				}}
			>
				<div
					style={{
						backgroundColor: 'white',
						borderRadius: '15px',
						padding: '15px',
						boxShadow: '0 4px 12px rgba(0,0,0,0.3)',
						fontSize: '14px',
						color: '#333',
						maxWidth: '300px',
						fontFamily: 'monospace',
					}}
				>
					<div style={{fontWeight: 'bold', marginBottom: '10px'}}>API Call:</div>
					<div>GET weather.api.com/current</div>
					<div>location: "San Francisco"</div>
					<div>units: "fahrenheit"</div>
				</div>
			</div>

			{/* Arrow to API */}
			<div
				style={{
					position: 'absolute',
					top: '30%',
					left: '65%',
					width: '150px',
					height: '4px',
					backgroundColor: '#FFD700',
					borderRadius: '2px',
					transform: `scaleX(${arrowProgress})`,
					transformOrigin: 'left center',
				}}
			>
				<div
					style={{
						position: 'absolute',
						right: '-10px',
						top: '-8px',
						width: 0,
						height: 0,
						borderTop: '10px solid transparent',
						borderBottom: '10px solid transparent',
						borderLeft: '20px solid #FFD700',
						opacity: arrowProgress,
					}}
				/>
			</div>

			{/* Weather API */}
			<div
				style={{
					position: 'absolute',
					top: '25%',
					right: '10%',
					opacity: apiOpacity,
				}}
			>
				<div
					style={{
						width: '120px',
						height: '120px',
						borderRadius: '20px',
						backgroundColor: '#32CD32',
						display: 'flex',
						alignItems: 'center',
						justifyContent: 'center',
						fontSize: '60px',
						color: 'white',
						boxShadow: '0 4px 12px rgba(0,0,0,0.3)',
					}}
				>
					🌤️
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
					Weather API
				</div>
			</div>

			{/* Loading indicator */}
			{loadingOpacity > 0 && (
				<div
					style={{
						position: 'absolute',
						top: '50%',
						left: '50%',
						transform: 'translateX(-50%)',
						color: 'white',
						fontSize: '20px',
						opacity: loadingOpacity,
					}}
				>
					<div
						style={{
							display: 'flex',
							alignItems: 'center',
							gap: '10px',
						}}
					>
						<div
							style={{
								width: '20px',
								height: '20px',
								border: '3px solid rgba(255,255,255,0.3)',
								borderTop: '3px solid white',
								borderRadius: '50%',
								animation: 'spin 1s linear infinite',
							}}
						/>
						Processing request...
					</div>
				</div>
			)}

			{/* API Response */}
			<div
				style={{
					position: 'absolute',
					top: '60%',
					right: '25%',
					opacity: responseOpacity,
					transform: `scale(${responseScale})`,
				}}
			>
				<div
					style={{
						backgroundColor: 'white',
						borderRadius: '15px',
						padding: '15px',
						boxShadow: '0 4px 12px rgba(0,0,0,0.3)',
						fontSize: '14px',
						color: '#333',
						maxWidth: '300px',
						fontFamily: 'monospace',
					}}
				>
					<div style={{fontWeight: 'bold', marginBottom: '10px'}}>Response:</div>
					<div>temperature: 72°F</div>
					<div>condition: "Sunny"</div>
					<div>humidity: 65%</div>
					<div>wind: 8 mph</div>
				</div>
			</div>

			{/* Return Arrow */}
			<div
				style={{
					position: 'absolute',
					top: '75%',
					left: '15%',
					width: '150px',
					height: '4px',
					backgroundColor: '#FF69B4',
					borderRadius: '2px',
					transform: `scaleX(${returnArrowProgress})`,
					transformOrigin: 'right center',
				}}
			>
				<div
					style={{
						position: 'absolute',
						left: '-10px',
						top: '-8px',
						width: 0,
						height: 0,
						borderTop: '10px solid transparent',
						borderBottom: '10px solid transparent',
						borderRight: '20px solid #FF69B4',
						opacity: returnArrowProgress,
					}}
				/>
			</div>

			<style jsx>{`
				@keyframes spin {
					0% { transform: rotate(0deg); }
					100% { transform: rotate(360deg); }
				}
			`}</style>
		</AbsoluteFill>
	);
};
