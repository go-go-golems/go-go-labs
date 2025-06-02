import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	spring,
	useCurrentFrame,
	useVideoConfig,
} from 'remotion';

export const ToolDiscoverySequence: React.FC = () => {
	const frame = useCurrentFrame();
	const {fps} = useVideoConfig();

	const stepOpacity = interpolate(frame, [0, 30], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const systemOpacity = interpolate(frame, [50, 80], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const startupOpacity = interpolate(frame, [90, 120], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const scanOpacity = interpolate(frame, [130, 160], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const discoveryOpacity = interpolate(frame, [170, 200], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const registrationOpacity = interpolate(frame, [220, 250], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const completionOpacity = interpolate(frame, [260, 270], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const systemScale = spring({
		frame: frame - 50,
		fps,
		config: {
			damping: 8,
			stiffness: 80,
		},
	});

	const registrationScale = spring({
		frame: frame - 220,
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
				Step 3: Automatic tool discovery on startup
			</div>

			{/* System startup */}
			<div
				style={{
					position: 'absolute',
					top: '25%',
					left: '15%',
					opacity: systemOpacity,
					transform: `scale(${systemScale})`,
				}}
			>
				<div
					style={{
						backgroundColor: 'rgba(44, 62, 80, 0.9)',
						borderRadius: '15px',
						padding: '20px',
						boxShadow: '0 8px 25px rgba(0,0,0,0.3)',
						color: 'white',
						maxWidth: '300px',
						border: '2px solid #34495e',
					}}
				>
					<div style={{fontSize: '18px', fontWeight: 'bold', marginBottom: '15px', textAlign: 'center'}}>
						🚀 LLM System Startup
					</div>
					<div style={{fontSize: '14px', lineHeight: 1.6}}>
						<div style={{marginBottom: '8px'}}>✅ Loading core tools</div>
						<div style={{marginBottom: '8px'}}>✅ Connecting to database</div>
						<div style={{marginBottom: '8px', color: '#f39c12'}}>🔄 Scanning for views...</div>
						<div style={{fontSize: '12px', fontStyle: 'italic', color: '#bdc3c7', marginTop: '10px'}}>
							Just like loading tool schemas, but for SQL views
						</div>
					</div>
				</div>
			</div>

			{/* Startup query */}
			{startupOpacity > 0 && (
				<div
					style={{
						position: 'absolute',
						top: '25%',
						right: '15%',
						opacity: startupOpacity,
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
							🔍 Startup Discovery Query:
						</div>
						<div style={{color: '#f39c12', fontSize: '12px'}}>
							SELECT name, sql<br/>
							FROM sqlite_master<br/>
							WHERE type = 'view'<br/>
							ORDER BY name;
						</div>
					</div>
				</div>
			)}

			{/* Scanning animation */}
			{scanOpacity > 0 && (
				<div
					style={{
						position: 'absolute',
						top: '50%',
						left: '50%',
						transform: 'translateX(-50%)',
						opacity: scanOpacity,
					}}
				>
					<div
						style={{
							color: 'white',
							fontSize: '18px',
							textAlign: 'center',
						}}
					>
						<div
							style={{
								display: 'flex',
								alignItems: 'center',
								gap: '10px',
								justifyContent: 'center',
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
							Scanning database for views...
						</div>
					</div>
				</div>
			)}

			{/* Discovered views */}
			{discoveryOpacity > 0 && (
				<div
					style={{
						position: 'absolute',
						top: '55%',
						left: '50%',
						transform: 'translateX(-50%)',
						opacity: discoveryOpacity,
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
							textAlign: 'center',
							minWidth: '500px',
						}}
					>
						<div style={{fontWeight: 'bold', marginBottom: '15px', color: '#27ae60', fontSize: '16px'}}>
							📋 Views Discovered:
						</div>
						<div style={{display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '15px'}}>
							<div style={{padding: '10px', backgroundColor: 'rgba(39, 174, 96, 0.1)', borderRadius: '8px'}}>
								<div style={{fontWeight: 'bold'}}>customer_orders_view</div>
								<div style={{fontSize: '12px', color: '#7f8c8d'}}>Pre-joined customer data</div>
							</div>
							<div style={{padding: '10px', backgroundColor: 'rgba(52, 152, 219, 0.1)', borderRadius: '8px'}}>
								<div style={{fontWeight: 'bold'}}>sales_summary_view</div>
								<div style={{fontSize: '12px', color: '#7f8c8d'}}>Monthly sales rollups</div>
							</div>
							<div style={{padding: '10px', backgroundColor: 'rgba(155, 89, 182, 0.1)', borderRadius: '8px'}}>
								<div style={{fontWeight: 'bold'}}>inventory_status_view</div>
								<div style={{fontSize: '12px', color: '#7f8c8d'}}>Stock level analysis</div>
							</div>
						</div>
					</div>
				</div>
			)}

			{/* Tool registration */}
			<div
				style={{
					position: 'absolute',
					bottom: '20%',
					left: '50%',
					transform: `translateX(-50%) scale(${registrationScale})`,
					opacity: registrationOpacity,
				}}
			>
				<div
					style={{
						backgroundColor: 'rgba(39, 174, 96, 0.9)',
						borderRadius: '20px',
						padding: '25px',
						color: 'white',
						fontSize: '16px',
						textAlign: 'center',
						boxShadow: '0 8px 25px rgba(0,0,0,0.3)',
						border: '3px solid #2ecc71',
						minWidth: '600px',
					}}
				>
					<div style={{fontSize: '24px', marginBottom: '20px', fontWeight: 'bold'}}>🔧 Dynamic Tool Registration</div>
					<div style={{display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '25px', fontSize: '14px'}}>
						<div>
							<div style={{fontSize: '20px', marginBottom: '10px'}}>📜</div>
							<div style={{fontWeight: 'bold', marginBottom: '5px'}}>Schema Generation</div>
							<div style={{fontSize: '12px', opacity: 0.9}}>Auto-generate tool definitions<br/>from view metadata</div>
						</div>
						<div>
							<div style={{fontSize: '20px', marginBottom: '10px'}}>🎯</div>
							<div style={{fontWeight: 'bold', marginBottom: '5px'}}>Smart Descriptions</div>
							<div style={{fontSize: '12px', opacity: 0.9}}>Extract use cases from<br/>view comments & structure</div>
						</div>
						<div>
							<div style={{fontSize: '20px', marginBottom: '10px'}}>⚡</div>
							<div style={{fontWeight: 'bold', marginBottom: '5px'}}>Instant Access</div>
							<div style={{fontSize: '12px', opacity: 0.9}}>Views become callable tools<br/>immediately available</div>
						</div>
					</div>
				</div>
			</div>

			{/* Completion message */}
			{completionOpacity > 0 && (
				<div
					style={{
						position: 'absolute',
						bottom: '8%',
						left: '50%',
						transform: 'translateX(-50%)',
						opacity: completionOpacity,
					}}
				>
					<div
						style={{
							backgroundColor: 'rgba(39, 174, 96, 0.1)',
							borderRadius: '15px',
							padding: '15px 30px',
							color: 'white',
							fontSize: '18px',
							fontWeight: 'bold',
							textAlign: 'center',
							border: '2px solid rgba(39, 174, 96, 0.5)',
						}}
					>
						🎉 Views are now discoverable tools - Infrastructure becomes intelligence!
					</div>
				</div>
			)}

			<style jsx>{`
				@keyframes spin {
					0% { transform: rotate(0deg); }
					100% { transform: rotate(360deg); }
				}
			`}</style>
		</AbsoluteFill>
	);
};
