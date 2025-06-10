import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	spring,
	useCurrentFrame,
	useVideoConfig,
} from 'remotion';

export const FutureEfficiencySequence: React.FC = () => {
	const frame = useCurrentFrame();
	const {fps} = useVideoConfig();

	const stepOpacity = interpolate(frame, [0, 30], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const llmOpacity = interpolate(frame, [0, 30], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const newRequestOpacity = interpolate(frame, [50, 80], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const instantToolOpacity = interpolate(frame, [100, 130], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const resultOpacity = interpolate(frame, [150, 180], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const evolutionOpacity = interpolate(frame, [200, 230], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const finalMessageOpacity = interpolate(frame, [250, 270], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const newRequestScale = spring({
		frame: frame - 50,
		fps,
		config: {
			damping: 8,
			stiffness: 80,
		},
	});

	const instantToolScale = spring({
		frame: frame - 100,
		fps,
		config: {
			damping: 8,
			stiffness: 80,
		},
	});

	const evolutionScale = spring({
		frame: frame - 200,
		fps,
		config: {
			damping: 10,
			stiffness: 100,
		},
	});

	const finalMessageScale = spring({
		frame: frame - 250,
		fps,
		config: {
			damping: 10,
			stiffness: 100,
		},
	});

	return (
		<AbsoluteFill
			style={{
				background: 'linear-gradient(135deg, #27ae60 0%, #2ecc71 100%)',
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
				Step 4: Future efficiency - Instant tool access
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
						backgroundColor: '#2c3e50',
						display: 'flex',
						alignItems: 'center',
						justifyContent: 'center',
						fontSize: '60px',
						color: 'white',
						boxShadow: '0 6px 20px rgba(0,0,0,0.2)',
						transform: frame > 80 && frame < 180 ? `scale(${1 + 0.02 * Math.sin(frame * 0.3)})` : 'scale(1)',
					}}
				>
					{frame > 150 ? '😊' : '🧠'}
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

			{/* New user request */}
			<div
				style={{
					position: 'absolute',
					top: '15%',
					left: '35%',
					opacity: newRequestOpacity,
					transform: `scale(${newRequestScale})`,
				}}
			>
				<div
					style={{
						backgroundColor: 'white',
						borderRadius: '20px',
						padding: '20px',
						boxShadow: '0 6px 20px rgba(0,0,0,0.15)',
						fontSize: '16px',
						color: '#2c3e50',
						maxWidth: '350px',
						lineHeight: 1.4,
					}}
				>
					"Show me total revenue for each customer this quarter"
				</div>
			</div>

			{/* Instant tool selection */}
			<div
				style={{
					position: 'absolute',
					top: '40%',
					left: '10%',
					opacity: instantToolOpacity,
					transform: `scale(${instantToolScale})`,
				}}
			>
				<div
					style={{
						backgroundColor: '#2c3e50',
						borderRadius: '15px',
						padding: '20px',
						boxShadow: '0 8px 25px rgba(0,0,0,0.4)',
						fontSize: '16px',
						color: '#ecf0f1',
						fontFamily: 'monospace',
						maxWidth: '500px',
						border: '3px solid #f39c12',
					}}
				>
					<div style={{color: '#f39c12', fontWeight: 'bold', marginBottom: '15px', fontSize: '18px'}}>
						⚡ Instant Tool Access:
					</div>
					<div style={{color: '#e74c3c'}}>query_customer_orders_view(</div>
					<div style={{paddingLeft: '20px', color: '#f39c12', lineHeight: 1.6}}>
						sql: "SELECT customer_name, SUM(amount) as total_revenue<br/>
						FROM customer_orders_view<br/>
						WHERE order_date &gt;= '2024-10-01'<br/>
						GROUP BY customer_name<br/>
						ORDER BY total_revenue DESC"
					</div>
					<div style={{color: '#e74c3c'}}>)</div>
					
					<div style={{marginTop: '15px', padding: '10px', backgroundColor: 'rgba(241, 196, 15, 0.2)', borderRadius: '8px', fontSize: '14px'}}>
						<div style={{color: '#f39c12', fontWeight: 'bold'}}>🚀 No exploration needed:</div>
						<div style={{color: '#ecf0f1'}}>• View already exists and is discoverable</div>
						<div style={{color: '#ecf0f1'}}>• LLM knows the schema from startup</div>
						<div style={{color: '#ecf0f1'}}>• Direct query execution - instant results</div>
					</div>
				</div>
			</div>

			{/* Quick result */}
			{resultOpacity > 0 && (
				<div
					style={{
						position: 'absolute',
						top: '40%',
						right: '10%',
						opacity: resultOpacity,
					}}
				>
					<div
						style={{
							backgroundColor: '#27ae60',
							borderRadius: '15px',
							padding: '20px',
							color: 'white',
							fontSize: '16px',
							textAlign: 'center',
							boxShadow: '0 8px 25px rgba(0,0,0,0.3)',
							border: '3px solid #2ecc71',
							minWidth: '200px',
						}}
					>
						<div style={{fontSize: '20px', marginBottom: '10px'}}>⚡</div>
						<div style={{fontWeight: 'bold', marginBottom: '10px'}}>
							Instant Results
						</div>
						<div style={{fontSize: '14px', opacity: 0.9}}>
							✅ 25 customers analyzed<br/>
							⏱️ 0.2 seconds response<br/>
							🔧 1 optimized query
						</div>
					</div>
				</div>
			)}

			{/* Evolution comparison */}
			<div
				style={{
					position: 'absolute',
					bottom: '25%',
					left: '50%',
					transform: `translateX(-50%) scale(${evolutionScale})`,
					opacity: evolutionOpacity,
				}}
			>
				<div
					style={{
						backgroundColor: 'rgba(255,255,255,0.95)',
						borderRadius: '20px',
						padding: '25px',
						color: '#2c3e50',
						fontSize: '16px',
						textAlign: 'center',
						boxShadow: '0 8px 25px rgba(0,0,0,0.3)',
						border: '3px solid #fff',
						minWidth: '700px',
					}}
				>
					<div style={{fontSize: '24px', marginBottom: '20px', fontWeight: 'bold'}}>🔄 Evolution Timeline</div>
					<div style={{display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '25px', fontSize: '14px'}}>
						<div>
							<div style={{fontSize: '18px', marginBottom: '10px'}}>❌</div>
							<div style={{fontWeight: 'bold', marginBottom: '8px', color: '#e74c3c'}}>Past: Inefficient</div>
							<div style={{fontSize: '12px', lineHeight: 1.4}}>
								• Bulk data calls<br/>
								• 3,600+ tokens wasted<br/>
								• Single-use queries<br/>
								• No infrastructure
							</div>
						</div>
						<div>
							<div style={{fontSize: '18px', marginBottom: '10px'}}>🏗️</div>
							<div style={{fontWeight: 'bold', marginBottom: '8px', color: '#f39c12'}}>Learning: Building</div>
							<div style={{fontSize: '12px', lineHeight: 1.4}}>
								• Schema exploration<br/>
								• View creation<br/>
								• Metadata documentation<br/>
								• 400 tokens invested
							</div>
						</div>
						<div>
							<div style={{fontSize: '18px', marginBottom: '10px'}}>⚡</div>
							<div style={{fontWeight: 'bold', marginBottom: '8px', color: '#27ae60'}}>Future: Optimized</div>
							<div style={{fontSize: '12px', lineHeight: 1.4}}>
								• Instant tool access<br/>
								• 50 tokens per query<br/>
								• Reusable infrastructure<br/>
								• Compounding returns
							</div>
						</div>
					</div>
				</div>
			</div>

			{/* Final message */}
			<div
				style={{
					position: 'absolute',
					bottom: '5%',
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
						fontSize: '22px',
						fontWeight: 'bold',
						textAlign: 'center',
						boxShadow: '0 8px 25px rgba(0,0,0,0.3)',
						border: '3px solid #2ecc71',
					}}
				>
					🎯 Intelligence investment pays dividends: Infrastructure → Tools → Efficiency
				</div>
			</div>
		</AbsoluteFill>
	);
};
