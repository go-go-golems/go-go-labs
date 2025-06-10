import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	spring,
	useCurrentFrame,
	useVideoConfig,
} from 'remotion';

export const MultipleQueriesSequence: React.FC = () => {
	const frame = useCurrentFrame();
	const {fps} = useVideoConfig();

	const stepOpacity = interpolate(frame, [0, 30], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const llmOpacity = interpolate(frame, [0, 30], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const query1Opacity = interpolate(frame, [50, 80], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const result1Opacity = interpolate(frame, [90, 120], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const query2Opacity = interpolate(frame, [140, 170], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const result2Opacity = interpolate(frame, [180, 210], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const query3Opacity = interpolate(frame, [230, 260], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const result3Opacity = interpolate(frame, [270, 300], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const query4Opacity = interpolate(frame, [320, 350], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const result4Opacity = interpolate(frame, [360, 390], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const summaryOpacity = interpolate(frame, [400, 420], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const query1Scale = spring({
		frame: frame - 50,
		fps,
		config: {
			damping: 8,
			stiffness: 80,
		},
	});

	const query2Scale = spring({
		frame: frame - 140,
		fps,
		config: {
			damping: 8,
			stiffness: 80,
		},
	});

	const query3Scale = spring({
		frame: frame - 230,
		fps,
		config: {
			damping: 8,
			stiffness: 80,
		},
	});

	const query4Scale = spring({
		frame: frame - 320,
		fps,
		config: {
			damping: 8,
			stiffness: 80,
		},
	});

	return (
		<AbsoluteFill
			style={{
				background: 'linear-gradient(135deg, #8e44ad 0%, #9b59b6 100%)',
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
				Step 2: Running multiple efficient queries
			</div>

			{/* LLM */}
			<div
				style={{
					position: 'absolute',
					top: '15%',
					left: '8%',
					opacity: llmOpacity,
				}}
			>
				<div
					style={{
						width: '100px',
						height: '100px',
						borderRadius: '20px',
						backgroundColor: '#9b59b6',
						display: 'flex',
						alignItems: 'center',
						justifyContent: 'center',
						fontSize: '50px',
						color: 'white',
						boxShadow: '0 6px 20px rgba(0,0,0,0.2)',
						transform: frame > 40 ? `scale(${1 + 0.02 * Math.sin(frame * 0.1)})` : 'scale(1)',
					}}
				>
					🧠
				</div>
				<div
					style={{
						textAlign: 'center',
						color: 'white',
						marginTop: '5px',
						fontSize: '16px',
						fontWeight: 'bold',
					}}
				>
					LLM
				</div>
			</div>

			{/* Query 1 - Count orders for John Smith */}
			<div
				style={{
					position: 'absolute',
					top: '20%',
					left: '20%',
					opacity: query1Opacity,
					transform: `scale(${query1Scale})`,
				}}
			>
				<div
					style={{
						backgroundColor: '#2c3e50',
						borderRadius: '12px',
						padding: '15px',
						boxShadow: '0 6px 20px rgba(0,0,0,0.3)',
						fontSize: '14px',
						color: '#ecf0f1',
						fontFamily: 'monospace',
						maxWidth: '350px',
						border: '2px solid #3498db',
					}}
				>
					<div style={{color: '#3498db', fontWeight: 'bold', marginBottom: '8px'}}>
						🔍 Query #1:
					</div>
					<div style={{color: '#f39c12', fontSize: '12px'}}>
						SELECT COUNT(*) FROM customer_orders_view<br/>
						WHERE customer_name = 'John Smith';
					</div>
				</div>
			</div>

			{/* Result 1 */}
			{result1Opacity > 0 && (
				<div
					style={{
						position: 'absolute',
						top: '20%',
						right: '15%',
						opacity: result1Opacity,
					}}
				>
					<div
						style={{
							backgroundColor: '#27ae60',
							borderRadius: '10px',
							padding: '12px 20px',
							color: 'white',
							fontSize: '16px',
							fontWeight: 'bold',
							textAlign: 'center',
						}}
					>
						7 orders
					</div>
				</div>
			)}

			{/* Query 2 - Total revenue from customer */}
			<div
				style={{
					position: 'absolute',
					top: '35%',
					left: '20%',
					opacity: query2Opacity,
					transform: `scale(${query2Scale})`,
				}}
			>
				<div
					style={{
						backgroundColor: '#2c3e50',
						borderRadius: '12px',
						padding: '15px',
						boxShadow: '0 6px 20px rgba(0,0,0,0.3)',
						fontSize: '14px',
						color: '#ecf0f1',
						fontFamily: 'monospace',
						maxWidth: '350px',
						border: '2px solid #e74c3c',
					}}
				>
					<div style={{color: '#e74c3c', fontWeight: 'bold', marginBottom: '8px'}}>
						💰 Query #2:
					</div>
					<div style={{color: '#f39c12', fontSize: '12px'}}>
						SELECT SUM(amount) FROM customer_orders_view<br/>
						WHERE customer_name = 'John Smith';
					</div>
				</div>
			</div>

			{/* Result 2 */}
			{result2Opacity > 0 && (
				<div
					style={{
						position: 'absolute',
						top: '35%',
						right: '15%',
						opacity: result2Opacity,
					}}
				>
					<div
						style={{
							backgroundColor: '#27ae60',
							borderRadius: '10px',
							padding: '12px 20px',
							color: 'white',
							fontSize: '16px',
							fontWeight: 'bold',
							textAlign: 'center',
						}}
					>
						$2,847.50
					</div>
				</div>
			)}

			{/* Query 3 - Average order value */}
			<div
				style={{
					position: 'absolute',
					top: '50%',
					left: '20%',
					opacity: query3Opacity,
					transform: `scale(${query3Scale})`,
				}}
			>
				<div
					style={{
						backgroundColor: '#2c3e50',
						borderRadius: '12px',
						padding: '15px',
						boxShadow: '0 6px 20px rgba(0,0,0,0.3)',
						fontSize: '14px',
						color: '#ecf0f1',
						fontFamily: 'monospace',
						maxWidth: '350px',
						border: '2px solid #f39c12',
					}}
				>
					<div style={{color: '#f39c12', fontWeight: 'bold', marginBottom: '8px'}}>
						📊 Query #3:
					</div>
					<div style={{color: '#f39c12', fontSize: '12px'}}>
						SELECT AVG(amount) FROM customer_orders_view<br/>
						WHERE customer_name = 'John Smith';
					</div>
				</div>
			</div>

			{/* Result 3 */}
			{result3Opacity > 0 && (
				<div
					style={{
						position: 'absolute',
						top: '50%',
						right: '15%',
						opacity: result3Opacity,
					}}
				>
					<div
						style={{
							backgroundColor: '#27ae60',
							borderRadius: '10px',
							padding: '12px 20px',
							color: 'white',
							fontSize: '16px',
							fontWeight: 'bold',
							textAlign: 'center',
						}}
					>
						$406.79
					</div>
				</div>
			)}

			{/* Query 4 - Latest order date */}
			<div
				style={{
					position: 'absolute',
					top: '65%',
					left: '20%',
					opacity: query4Opacity,
					transform: `scale(${query4Scale})`,
				}}
			>
				<div
					style={{
						backgroundColor: '#2c3e50',
						borderRadius: '12px',
						padding: '15px',
						boxShadow: '0 6px 20px rgba(0,0,0,0.3)',
						fontSize: '14px',
						color: '#ecf0f1',
						fontFamily: 'monospace',
						maxWidth: '350px',
						border: '2px solid #9b59b6',
					}}
				>
					<div style={{color: '#9b59b6', fontWeight: 'bold', marginBottom: '8px'}}>
						📅 Query #4:
					</div>
					<div style={{color: '#f39c12', fontSize: '12px'}}>
						SELECT MAX(order_date) FROM customer_orders_view<br/>
						WHERE customer_name = 'John Smith';
					</div>
				</div>
			</div>

			{/* Result 4 */}
			{result4Opacity > 0 && (
				<div
					style={{
						position: 'absolute',
						top: '65%',
						right: '15%',
						opacity: result4Opacity,
					}}
				>
					<div
						style={{
							backgroundColor: '#27ae60',
							borderRadius: '10px',
							padding: '12px 20px',
							color: 'white',
							fontSize: '16px',
							fontWeight: 'bold',
							textAlign: 'center',
						}}
					>
						2024-11-28
					</div>
				</div>
			)}

			{/* Summary efficiency box */}
			{summaryOpacity > 0 && (
				<div
					style={{
						position: 'absolute',
						bottom: '8%',
						left: '50%',
						transform: 'translateX(-50%)',
						opacity: summaryOpacity,
					}}
				>
					<div
						style={{
							backgroundColor: 'rgba(155, 89, 182, 0.9)',
							borderRadius: '20px',
							padding: '20px 30px',
							color: 'white',
							fontSize: '18px',
							textAlign: 'center',
							boxShadow: '0 8px 25px rgba(0,0,0,0.3)',
							border: '2px solid #8e44ad',
						}}
					>
						<div style={{fontSize: '28px', marginBottom: '10px'}}>⚡</div>
						<div style={{fontWeight: 'bold', marginBottom: '10px'}}>
							4 Queries, 0 JOINs Needed!
						</div>
						<div style={{fontSize: '16px', display: 'flex', gap: '30px', justifyContent: 'center'}}>
							<div>
								<div style={{fontWeight: 'bold'}}>🚀 Fast</div>
								<div style={{fontSize: '14px'}}>Pre-joined data</div>
							</div>
							<div>
								<div style={{fontWeight: 'bold'}}>🎯 Simple</div>
								<div style={{fontSize: '14px'}}>Clean syntax</div>
							</div>
							<div>
								<div style={{fontWeight: 'bold'}}>🔄 Reusable</div>
								<div style={{fontSize: '14px'}}>One view, many uses</div>
							</div>
						</div>
					</div>
				</div>
			)}

			{/* Progress indicator */}
			<div
				style={{
					position: 'absolute',
					top: '15%',
					right: '8%',
					opacity: llmOpacity,
				}}
			>
				<div
					style={{
						backgroundColor: 'rgba(255,255,255,0.1)',
						borderRadius: '10px',
						padding: '15px',
						color: 'white',
						fontSize: '14px',
						textAlign: 'center',
						minWidth: '100px',
					}}
				>
					<div style={{fontWeight: 'bold', marginBottom: '10px'}}>Progress</div>
					<div style={{display: 'flex', flexDirection: 'column', gap: '5px'}}>
						<div style={{opacity: result1Opacity > 0 ? 1 : 0.3}}>✅ Count</div>
						<div style={{opacity: result2Opacity > 0 ? 1 : 0.3}}>✅ Sum</div>
						<div style={{opacity: result3Opacity > 0 ? 1 : 0.3}}>✅ Average</div>
						<div style={{opacity: result4Opacity > 0 ? 1 : 0.3}}>✅ Latest</div>
					</div>
				</div>
			</div>
		</AbsoluteFill>
	);
};
