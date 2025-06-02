import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	spring,
	useCurrentFrame,
	useVideoConfig,
} from 'remotion';

export const TokenEfficiencyComparisonSequence: React.FC = () => {
	const frame = useCurrentFrame();
	const {fps} = useVideoConfig();

	const stepOpacity = interpolate(frame, [0, 30], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const inefficientOpacity = interpolate(frame, [50, 80], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const efficientOpacity = interpolate(frame, [120, 150], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const comparisonOpacity = interpolate(frame, [200, 230], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const conclusionOpacity = interpolate(frame, [270, 300], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const inefficientScale = spring({
		frame: frame - 50,
		fps,
		config: {
			damping: 8,
			stiffness: 80,
		},
	});

	const efficientScale = spring({
		frame: frame - 120,
		fps,
		config: {
			damping: 8,
			stiffness: 80,
		},
	});

	const comparisonScale = spring({
		frame: frame - 200,
		fps,
		config: {
			damping: 10,
			stiffness: 100,
		},
	});

	const conclusionScale = spring({
		frame: frame - 270,
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
					top: '15%',
					left: '50%',
					transform: 'translateX(-50%)',
					color: 'white',
					fontSize: '28px',
					fontWeight: 'bold',
					opacity: stepOpacity,
				}}
			>
				Step 1: Token efficiency comparison
			</div>

			{/* Inefficient Approach */}
			<div
				style={{
					position: 'absolute',
					top: '25%',
					left: '5%',
					opacity: inefficientOpacity,
					transform: `scale(${inefficientScale})`,
				}}
			>
				<div
					style={{
						backgroundColor: '#e74c3c',
						borderRadius: '15px',
						padding: '20px',
						boxShadow: '0 8px 25px rgba(0,0,0,0.3)',
						color: 'white',
						maxWidth: '400px',
						border: '3px solid #c0392b',
					}}
				>
					<div style={{fontSize: '20px', fontWeight: 'bold', marginBottom: '15px', textAlign: 'center'}}>
						❌ Inefficient: Single Bulk Call
					</div>
					<div style={{fontSize: '14px', lineHeight: 1.6}}>
						<div style={{marginBottom: '10px', fontFamily: 'monospace', backgroundColor: 'rgba(0,0,0,0.2)', padding: '8px', borderRadius: '5px'}}>
							get_crm_companies() → Returns ALL 36 companies
						</div>
						<div style={{display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px', fontSize: '13px'}}>
							<div>
								<strong>Input:</strong> 50 tokens<br/>
								<strong>Output:</strong> 3,600 tokens<br/>
								<strong>Processing:</strong> Heavy
							</div>
							<div>
								<strong>Total:</strong> 3,650 tokens<br/>
								<strong>Useful data:</strong> 1 company<br/>
								<strong>Waste:</strong> 98% unused
							</div>
						</div>
					</div>
					<div style={{marginTop: '15px', padding: '10px', backgroundColor: 'rgba(192, 57, 43, 0.3)', borderRadius: '8px', textAlign: 'center'}}>
						<div style={{fontWeight: 'bold', color: '#fff'}}>🔥 Massive token waste for simple query</div>
					</div>
				</div>
			</div>

			{/* Efficient Approach */}
			<div
				style={{
					position: 'absolute',
					top: '25%',
					right: '5%',
					opacity: efficientOpacity,
					transform: `scale(${efficientScale})`,
				}}
			>
				<div
					style={{
						backgroundColor: '#2c3e50',
						borderRadius: '15px',
						padding: '20px',
						boxShadow: '0 8px 25px rgba(0,0,0,0.3)',
						color: 'white',
						maxWidth: '400px',
						border: '3px solid #34495e',
					}}
				>
					<div style={{fontSize: '20px', fontWeight: 'bold', marginBottom: '15px', textAlign: 'center'}}>
						✅ Efficient: Smart Exploration
					</div>
					<div style={{fontSize: '14px', lineHeight: 1.6}}>
						<div style={{marginBottom: '10px', fontFamily: 'monospace', backgroundColor: 'rgba(255,255,255,0.1)', padding: '8px', borderRadius: '5px', fontSize: '12px'}}>
							1. sqlite_query("SELECT name FROM sqlite_master")<br/>
							2. sqlite_query("PRAGMA table_info(customers)")<br/>
							3. sqlite_query("PRAGMA table_info(orders)")<br/>
							4. sqlite_query("CREATE VIEW customer_orders_view...")<br/>
							5. sqlite_query("SELECT COUNT(*) FROM customer_orders_view...")
						</div>
						<div style={{display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px', fontSize: '13px'}}>
							<div>
								<strong>Input:</strong> 250 tokens<br/>
								<strong>Output:</strong> 150 tokens<br/>
								<strong>Processing:</strong> Smart
							</div>
							<div>
								<strong>Total:</strong> 400 tokens<br/>
								<strong>Useful data:</strong> All relevant<br/>
								<strong>Waste:</strong> 0% unused
							</div>
						</div>
					</div>
					<div style={{marginTop: '15px', padding: '10px', backgroundColor: 'rgba(52, 73, 94, 0.5)', borderRadius: '8px', textAlign: 'center'}}>
						<div style={{fontWeight: 'bold', color: '#fff'}}>🎯 Multiple queries, still 90% more efficient!</div>
					</div>
				</div>
			</div>

			{/* Visual Comparison */}
			<div
				style={{
					position: 'absolute',
					bottom: '30%',
					left: '50%',
					transform: `translateX(-50%) scale(${comparisonScale})`,
					opacity: comparisonOpacity,
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
						minWidth: '600px',
					}}
				>
					<div style={{fontSize: '24px', marginBottom: '20px', fontWeight: 'bold'}}>📊 Token Usage Comparison</div>
					
					{/* Visual token bars */}
					<div style={{display: 'flex', justifyContent: 'center', gap: '40px', marginBottom: '20px'}}>
						<div style={{textAlign: 'center'}}>
							<div style={{fontSize: '14px', marginBottom: '10px', fontWeight: 'bold'}}>Inefficient CRM</div>
							<div style={{width: '150px', height: '20px', backgroundColor: '#e74c3c', borderRadius: '10px', position: 'relative'}}>
								<div style={{position: 'absolute', top: '-25px', right: '0', fontSize: '12px', fontWeight: 'bold'}}>3,650 tokens</div>
							</div>
						</div>
						<div style={{textAlign: 'center'}}>
							<div style={{fontSize: '14px', marginBottom: '10px', fontWeight: 'bold'}}>Smart SQLite</div>
							<div style={{width: '16px', height: '20px', backgroundColor: '#27ae60', borderRadius: '10px', position: 'relative'}}>
								<div style={{position: 'absolute', top: '-25px', left: '0', fontSize: '12px', fontWeight: 'bold', whiteSpace: 'nowrap'}}>400 tokens</div>
							</div>
						</div>
					</div>

					<div style={{display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '25px', fontSize: '14px'}}>
						<div>
							<div style={{fontSize: '20px', fontWeight: 'bold', color: '#27ae60', marginBottom: '5px'}}>90%</div>
							<div style={{color: '#2c3e50'}}>Fewer Tokens</div>
							<div style={{fontSize: '12px', color: '#7f8c8d'}}>Even with exploration</div>
						</div>
						<div>
							<div style={{fontSize: '20px', fontWeight: 'bold', color: '#3498db', marginBottom: '5px'}}>5</div>
							<div style={{color: '#2c3e50'}}>Smart Queries</div>
							<div style={{fontSize: '12px', color: '#7f8c8d'}}>vs 1 wasteful call</div>
						</div>
						<div>
							<div style={{fontSize: '20px', fontWeight: 'bold', color: '#9b59b6', marginBottom: '5px'}}>∞</div>
							<div style={{color: '#2c3e50'}}>Future Value</div>
							<div style={{fontSize: '12px', color: '#7f8c8d'}}>Reusable infrastructure</div>
						</div>
					</div>
				</div>
			</div>

			{/* Conclusion */}
			<div
				style={{
					position: 'absolute',
					bottom: '8%',
					left: '50%',
					transform: `translateX(-50%) scale(${conclusionScale})`,
					opacity: conclusionOpacity,
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
						boxShadow: '0 8px 25px rgba(0,0,0,0.3)',
						border: '3px solid #2ecc71',
					}}
				>
					💡 Intelligence beats brute force: Exploration + Infrastructure = Efficiency
				</div>
			</div>
		</AbsoluteFill>
	);
};
