import React from 'react';
import {
	AbsoluteFill,
	interpolate,
	spring,
	useCurrentFrame,
	useVideoConfig,
} from 'remotion';

export const CRMToolExecutionSequence: React.FC = () => {
	const frame = useCurrentFrame();
	const {fps} = useVideoConfig();

	const stepOpacity = interpolate(frame, [0, 30], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const llmOpacity = interpolate(frame, [0, 30], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const requestOpacity = interpolate(frame, [50, 80], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const arrowOpacity = interpolate(frame, [80, 110], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const arrowLength = interpolate(frame, [80, 110], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const crmOpacity = interpolate(frame, [100, 130], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const loadingOpacity = interpolate(frame, [130, 160], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const dataFloodOpacity = interpolate(frame, [180, 210], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const dataScrollProgress = interpolate(frame, [210, 330], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const warningOpacity = interpolate(frame, [300, 360], [0, 1], {
		extrapolateRight: 'clamp',
	});

	const requestScale = spring({
		frame: frame - 50,
		fps,
		config: {
			damping: 8,
			stiffness: 80,
		},
	});

	// Generate massive data entries
	const generateCompanyData = () => {
		const companies = [
			'OpenAI', 'Microsoft', 'Google', 'Apple', 'Amazon', 'Meta', 'Tesla', 'Netflix',
			'Salesforce', 'Oracle', 'IBM', 'Intel', 'NVIDIA', 'Adobe', 'Uber', 'Airbnb',
			'Spotify', 'Zoom', 'Slack', 'Dropbox', 'Twitter', 'LinkedIn', 'PayPal', 'Square',
			'Stripe', 'Shopify', 'ServiceNow', 'Snowflake', 'Palantir', 'Datadog',
			'MongoDB', 'Atlassian', 'Twilio', 'Okta', 'CrowdStrike', 'Zscaler'
		];
		
		return companies.map((company, index) => ({
			id: 1000 + index,
			name: company,
			email: `contact@${company.toLowerCase().replace(/\s+/g, '')}.com`,
			phone: `+1-555-${String(Math.floor(Math.random() * 9000) + 1000)}`,
			address: `${Math.floor(Math.random() * 9999) + 1} Tech St, Silicon Valley, CA`,
			employees: Math.floor(Math.random() * 50000) + 100,
			revenue: `$${Math.floor(Math.random() * 100)}B`,
			founded: Math.floor(Math.random() * 30) + 1990,
			industry: 'Technology',
		}));
	};

	const allCompanies = generateCompanyData();
	const scrollOffset = dataScrollProgress * (allCompanies.length - 8) * 60;

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
				Step 3: Tool execution returns MASSIVE dataset
			</div>

			{/* LLM */}
			<div
				style={{
					position: 'absolute',
					top: '20%',
					left: '10%',
					opacity: llmOpacity,
				}}
			>
				<div
					style={{
						width: '100px',
						height: '100px',
						borderRadius: '20px',
						backgroundColor: '#8e44ad',
						display: 'flex',
						alignItems: 'center',
						justifyContent: 'center',
						fontSize: '50px',
						color: 'white',
						boxShadow: '0 6px 20px rgba(0,0,0,0.2)',
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

			{/* API Request */}
			<div
				style={{
					position: 'absolute',
					top: '15%',
					left: '25%',
					opacity: requestOpacity,
					transform: `scale(${requestScale})`,
				}}
			>
				<div
					style={{
						backgroundColor: 'white',
						borderRadius: '12px',
						padding: '15px',
						boxShadow: '0 6px 20px rgba(0,0,0,0.15)',
						fontSize: '14px',
						color: '#2c3e50',
						maxWidth: '250px',
						fontFamily: 'monospace',
					}}
				>
					<div style={{fontWeight: 'bold', marginBottom: '8px'}}>API Call:</div>
					<div>get_crm_companies()</div>
					<div style={{fontSize: '12px', color: '#7f8c8d', marginTop: '5px'}}>
						No filters applied!
					</div>
				</div>
			</div>

			{/* Smooth Arrow to CRM */}
			<div
				style={{
					position: 'absolute',
					top: '25%',
					left: '50%',
					opacity: arrowOpacity,
				}}
			>
				<svg width="120" height="40" viewBox="0 0 120 40">
					<defs>
						<linearGradient id="crmArrowGradient" x1="0%" y1="0%" x2="100%" y2="0%">
							<stop offset="0%" stopColor="#e74c3c" />
							<stop offset="100%" stopColor="#c0392b" />
						</linearGradient>
					</defs>
					<path
						d={`M 10 20 L ${10 + 80 * arrowLength} 20`}
						stroke="url(#crmArrowGradient)"
						strokeWidth="4"
						strokeLinecap="round"
						fill="none"
					/>
					<polygon
						points={`${10 + 80 * arrowLength},20 ${10 + 80 * arrowLength - 12},15 ${10 + 80 * arrowLength - 12},25`}
						fill="url(#crmArrowGradient)"
						opacity={arrowLength}
					/>
				</svg>
			</div>

			{/* CRM Database */}
			<div
				style={{
					position: 'absolute',
					top: '20%',
					right: '10%',
					opacity: crmOpacity,
				}}
			>
				<div
					style={{
						width: '100px',
						height: '100px',
						borderRadius: '20px',
						backgroundColor: '#27ae60',
						display: 'flex',
						alignItems: 'center',
						justifyContent: 'center',
						fontSize: '50px',
						color: 'white',
						boxShadow: '0 6px 20px rgba(0,0,0,0.2)',
					}}
				>
					🗄️
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
					CRM DB
				</div>
			</div>

			{/* Loading indicator */}
			{loadingOpacity > 0 && (
				<div
					style={{
						position: 'absolute',
						top: '40%',
						left: '50%',
						transform: 'translateX(-50%)',
						color: 'white',
						fontSize: '18px',
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
						Querying database...
					</div>
				</div>
			)}

			{/* MASSIVE Data Response */}
			{dataFloodOpacity > 0 && (
				<div
					style={{
						position: 'absolute',
						top: '45%',
						left: '5%',
						right: '5%',
						height: '50%',
						opacity: dataFloodOpacity,
						overflow: 'hidden',
					}}
				>
					<div
						style={{
							backgroundColor: 'rgba(255,255,255,0.95)',
							borderRadius: '15px',
							padding: '20px',
							boxShadow: '0 6px 20px rgba(0,0,0,0.3)',
							fontSize: '12px',
							color: '#2c3e50',
							fontFamily: 'monospace',
							height: '100%',
							overflow: 'hidden',
							position: 'relative',
						}}
					>
						<div style={{fontWeight: 'bold', marginBottom: '15px', fontSize: '16px', color: '#e74c3c'}}>
							🚨 RESPONSE: {allCompanies.length} companies returned (Only need 1!)
						</div>
						<div
							style={{
								transform: `translateY(-${scrollOffset}px)`,
								lineHeight: 1.8,
							}}
						>
							{allCompanies.map((company, index) => (
								<div
									key={index}
									style={{
										marginBottom: '15px',
										padding: '10px',
										backgroundColor: company.name === 'OpenAI' ? 'rgba(39, 174, 96, 0.2)' : 'rgba(0,0,0,0.05)',
										borderRadius: '8px',
										borderLeft: company.name === 'OpenAI' ? '4px solid #27ae60' : '4px solid transparent',
									}}
								>
									<div style={{fontWeight: 'bold', color: company.name === 'OpenAI' ? '#27ae60' : '#2c3e50'}}>
										{company.name} {company.name === 'OpenAI' ? '← TARGET!' : ''}
									</div>
									<div>ID: {company.id}</div>
									<div>Email: {company.email}</div>
									<div>Phone: {company.phone}</div>
									<div>Address: {company.address}</div>
									<div>Employees: {company.employees}</div>
									<div>Revenue: {company.revenue}</div>
									<div>Founded: {company.founded}</div>
								</div>
							))}
						</div>
						
						{/* Scroll indicator */}
						<div
							style={{
								position: 'absolute',
								right: '10px',
								top: '50px',
								bottom: '10px',
								width: '8px',
								backgroundColor: 'rgba(0,0,0,0.1)',
								borderRadius: '4px',
							}}
						>
							<div
								style={{
									position: 'absolute',
									top: `${(scrollOffset / ((allCompanies.length - 8) * 60)) * 100}%`,
									width: '8px',
									height: '40px',
									backgroundColor: '#e74c3c',
									borderRadius: '4px',
								}}
							/>
						</div>
					</div>
				</div>
			)}

			{/* Warning */}
			{warningOpacity > 0 && (
				<div
					style={{
						position: 'absolute',
						bottom: '5%',
						left: '50%',
						transform: 'translateX(-50%)',
						color: '#f39c12',
						fontSize: '20px',
						fontWeight: 'bold',
						opacity: warningOpacity,
						textAlign: 'center',
						backgroundColor: 'rgba(0,0,0,0.7)',
						padding: '15px 30px',
						borderRadius: '15px',
					}}
				>
					⚠️ Token overload! Processing {allCompanies.length * 100}+ tokens for 1 answer
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
