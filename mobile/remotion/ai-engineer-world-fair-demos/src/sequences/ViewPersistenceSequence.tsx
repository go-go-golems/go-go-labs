import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	spring,
	useCurrentFrame,
	useVideoConfig,
} from 'remotion';

export const ViewPersistenceSequence: React.FC = () => {
	const frame = useCurrentFrame();
	const {fps} = useVideoConfig();

	const stepOpacity = interpolate(frame, [0, 30], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const llmOpacity = interpolate(frame, [0, 30], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const thoughtOpacity = interpolate(frame, [40, 70], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const saveOpacity = interpolate(frame, [90, 120], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const metadataOpacity = interpolate(frame, [150, 180], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const infrastructureOpacity = interpolate(frame, [210, 240], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const benefitsOpacity = interpolate(frame, [250, 270], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const thoughtScale = spring({
		frame: frame - 40,
		fps,
		config: {
			damping: 8,
			stiffness: 80,
		},
	});

	const saveScale = spring({
		frame: frame - 90,
		fps,
		config: {
			damping: 8,
			stiffness: 80,
		},
	});

	const infrastructureScale = spring({
		frame: frame - 210,
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
					top: '12%',
					left: '50%',
					transform: 'translateX(-50%)',
					color: 'white',
					fontSize: '28px',
					fontWeight: 'bold',
					opacity: stepOpacity,
				}}
			>
				Step 2: Persisting views as reusable infrastructure
			</div>

			{/* LLM */}
			<div
				style={{
					position: 'absolute',
					top: '22%',
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
						transform: frame > 70 && frame < 180 ? `scale(${1 + 0.03 * Math.sin(frame * 0.2)})` : 'scale(1)',
					}}
				>
					{frame > 70 && frame < 180 ? '💡' : '🧠'}
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
					top: '18%',
					left: '35%',
					opacity: thoughtOpacity,
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
						maxWidth: '300px',
						lineHeight: 1.4,
					}}
				>
					"This view will be useful for future customer analytics. I should save it with proper metadata so it can be discovered later."
				</div>
			</div>

			{/* Save View Command */}
			<div
				style={{
					position: 'absolute',
					top: '45%',
					left: '10%',
					opacity: saveOpacity,
					transform: `scale(${saveScale})`,
				}}
			>
				<div
					style={{
						backgroundColor: '#2c3e50',
						borderRadius: '15px',
						padding: '25px',
						boxShadow: '0 8px 25px rgba(0,0,0,0.4)',
						fontSize: '16px',
						color: '#ecf0f1',
						fontFamily: 'monospace',
						maxWidth: '600px',
						border: '3px solid #27ae60',
					}}
				>
					<div style={{color: '#27ae60', fontWeight: 'bold', marginBottom: '15px', fontSize: '18px'}}>
						💾 Saving View with Metadata:
					</div>
					<div style={{color: '#e74c3c'}}>sqlite_query(</div>
					<div style={{paddingLeft: '20px', color: '#f39c12', lineHeight: 1.6}}>
						"-- View: customer_orders_view<br/>
						-- Description: Pre-joined customer and order data<br/>
						-- Use cases: Customer analytics, order summaries<br/>
						-- Created: 2024-11-30<br/>
						-- Tags: customers, orders, analytics<br/>
						<br/>
						CREATE VIEW customer_orders_view AS<br/>
						SELECT c.id as customer_id, c.name as customer_name,<br/>
						&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;c.email, o.id as order_id,<br/>
						&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;o.amount, o.order_date<br/>
						FROM customers c JOIN orders o ON c.id = o.customer_id;"
					</div>
					<div style={{color: '#e74c3c'}}>)</div>
				</div>
			</div>

			{/* Metadata Registration */}
			{metadataOpacity > 0 && (
				<div
					style={{
						position: 'absolute',
						top: '45%',
						right: '10%',
						opacity: metadataOpacity,
					}}
				>
					<div
						style={{
							backgroundColor: 'rgba(255,255,255,0.95)',
							borderRadius: '15px',
							padding: '20px',
							boxShadow: '0 6px 20px rgba(0,0,0,0.15)',
							fontSize: '14px',
							color: '#2c3e50',
							maxWidth: '300px',
						}}
					>
						<div style={{fontWeight: 'bold', marginBottom: '15px', color: '#27ae60', fontSize: '16px'}}>
							📝 View Registry Updated:
						</div>
						<div style={{marginBottom: '10px'}}>
							<strong>Name:</strong> customer_orders_view<br/>
							<strong>Type:</strong> Customer Analytics<br/>
							<strong>Columns:</strong> 6 fields<br/>
							<strong>Performance:</strong> Pre-joined<br/>
							<strong>Tags:</strong> customers, orders
						</div>
						<div style={{padding: '10px', backgroundColor: 'rgba(39, 174, 96, 0.1)', borderRadius: '8px', fontSize: '13px'}}>
							<strong>Status:</strong> ✅ Available for discovery
						</div>
					</div>
				</div>
			)}

			{/* Infrastructure visualization */}
			<div
				style={{
					position: 'absolute',
					bottom: '25%',
					left: '50%',
					transform: `translateX(-50%) scale(${infrastructureScale})`,
					opacity: infrastructureOpacity,
				}}
			>
				<div
					style={{
						backgroundColor: 'rgba(44, 62, 80, 0.9)',
						borderRadius: '20px',
						padding: '25px',
						color: 'white',
						fontSize: '16px',
						textAlign: 'center',
						boxShadow: '0 8px 25px rgba(0,0,0,0.3)',
						border: '2px solid #34495e',
						minWidth: '600px',
					}}
				>
					<div style={{fontSize: '24px', marginBottom: '20px', fontWeight: 'bold'}}>🏗️ Infrastructure Layer</div>
					<div style={{display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '25px', fontSize: '14px'}}>
						<div>
							<div style={{fontSize: '30px', marginBottom: '10px'}}>🗃️</div>
							<div style={{fontWeight: 'bold', marginBottom: '5px'}}>Database Views</div>
							<div style={{fontSize: '12px', opacity: 0.8}}>Pre-computed joins<br/>Ready for queries</div>
						</div>
						<div>
							<div style={{fontSize: '30px', marginBottom: '10px'}}>📋</div>
							<div style={{fontWeight: 'bold', marginBottom: '5px'}}>Metadata Registry</div>
							<div style={{fontSize: '12px', opacity: 0.8}}>Searchable descriptions<br/>Use case documentation</div>
						</div>
						<div>
							<div style={{fontSize: '30px', marginBottom: '10px'}}>🔄</div>
							<div style={{fontWeight: 'bold', marginBottom: '5px'}}>Auto-Discovery</div>
							<div style={{fontSize: '12px', opacity: 0.8}}>Tool enumeration<br/>Dynamic schemas</div>
						</div>
					</div>
				</div>
			</div>

			{/* Benefits */}
			{benefitsOpacity > 0 && (
				<div
					style={{
						position: 'absolute',
						bottom: '8%',
						left: '50%',
						transform: 'translateX(-50%)',
						opacity: benefitsOpacity,
					}}
				>
					<div
						style={{
							backgroundColor: 'rgba(39, 174, 96, 0.1)',
							borderRadius: '15px',
							padding: '15px 25px',
							color: 'white',
							fontSize: '16px',
							textAlign: 'center',
							border: '2px solid rgba(39, 174, 96, 0.5)',
							display: 'flex',
							gap: '30px',
							alignItems: 'center',
						}}
					>
						<div>
							<div style={{fontSize: '18px', marginBottom: '5px'}}>💾</div>
							<div style={{fontWeight: 'bold'}}>Persistent</div>
							<div style={{fontSize: '14px'}}>Survives sessions</div>
						</div>
						<div>
							<div style={{fontSize: '18px', marginBottom: '5px'}}>🔍</div>
							<div style={{fontWeight: 'bold'}}>Discoverable</div>
							<div style={{fontSize: '14px'}}>Auto-enumerated</div>
						</div>
						<div>
							<div style={{fontSize: '18px', marginBottom: '5px'}}>📚</div>
							<div style={{fontWeight: 'bold'}}>Documented</div>
							<div style={{fontSize: '14px'}}>Self-describing</div>
						</div>
						<div>
							<div style={{fontSize: '18px', marginBottom: '5px'}}>🚀</div>
							<div style={{fontWeight: 'bold'}}>Reusable</div>
							<div style={{fontSize: '14px'}}>Cross-sessions</div>
						</div>
					</div>
				</div>
			)}
		</AbsoluteFill>
	);
};
