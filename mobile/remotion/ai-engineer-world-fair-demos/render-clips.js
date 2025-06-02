#!/usr/bin/env node

const { execSync } = require('child_process');
const path = require('path');

// Define all the clip compositions
const clips = [
  // Full animations
  { id: 'ToolCallingAnimation', name: 'weather-full' },
  { id: 'CRMQueryAnimation', name: 'crm-full' },
  { id: 'SQLiteQueryAnimation', name: 'sqlite-full' },
  { id: 'SQLiteViewOptimizationAnimation', name: 'sqlite-view-optimization-full' },
  { id: 'ComprehensiveComparisonAnimation', name: 'comprehensive-comparison-full' },
  
  // Weather API steps
  { id: 'Weather-Step1-UserRequest', name: 'weather-step1-user-request' },
  { id: 'Weather-Step2-ToolAnalysis', name: 'weather-step2-tool-analysis' },
  { id: 'Weather-Step3-ToolExecution', name: 'weather-step3-tool-execution' },
  { id: 'Weather-Step4-ResultIntegration', name: 'weather-step4-result-integration' },
  
  // CRM steps
  { id: 'CRM-Step1-UserRequest', name: 'crm-step1-user-request' },
  { id: 'CRM-Step2-ToolAnalysis', name: 'crm-step2-tool-analysis' },
  { id: 'CRM-Step3-ToolExecution', name: 'crm-step3-tool-execution' },
  { id: 'CRM-Step4-ResultProcessing', name: 'crm-step4-result-processing' },
  
  // SQLite steps
  { id: 'SQLite-Step1-UserRequest', name: 'sqlite-step1-user-request' },
  { id: 'SQLite-Step2-SchemaDiscovery', name: 'sqlite-step2-schema-discovery' },
  { id: 'SQLite-Step3-TableExploration', name: 'sqlite-step3-table-exploration' },
  { id: 'SQLite-Step4-TargetedQuery', name: 'sqlite-step4-targeted-query' },
  { id: 'SQLite-Step5-FinalResponse', name: 'sqlite-step5-final-response' },
  
  // SQLite View Optimization steps
  { id: 'SQLiteView-Step1-ViewCreation', name: 'sqlite-view-step1-view-creation' },
  { id: 'SQLiteView-Step2-MultipleQueries', name: 'sqlite-view-step2-multiple-queries' },
  { id: 'SQLiteView-Step3-PerformanceComparison', name: 'sqlite-view-step3-performance-comparison' },
  
  // Comprehensive Comparison steps
  { id: 'Comparison-Step1-TokenEfficiency', name: 'comparison-step1-token-efficiency' },
  { id: 'Comparison-Step2-ViewPersistence', name: 'comparison-step2-view-persistence' },
  { id: 'Comparison-Step3-ToolDiscovery', name: 'comparison-step3-tool-discovery' },
  { id: 'Comparison-Step4-FutureEfficiency', name: 'comparison-step4-future-efficiency' },
];

// Parse command line arguments
const args = process.argv.slice(2);
const shouldRenderAll = args.includes('--all');
const requestedClip = args.find(arg => !arg.startsWith('--'));

function renderClip(clip) {
  const outputPath = `out/${clip.name}.mp4`;
  const command = `npx remotion render ${clip.id} ${outputPath}`;
  
  console.log(`\n🎬 Rendering: ${clip.name}`);
  console.log(`📝 Command: ${command}`);
  
  try {
    execSync(command, { stdio: 'inherit' });
    console.log(`✅ Successfully rendered: ${outputPath}`);
  } catch (error) {
    console.error(`❌ Failed to render: ${clip.name}`);
    console.error(error.message);
  }
}

function listClips() {
  console.log('\n📋 Available clips:');
  console.log('\n🎯 Full Animations:');
  clips.filter(c => c.name.includes('full')).forEach(clip => {
    console.log(`  • ${clip.name} (${clip.id})`);
  });
  
  console.log('\n🌤️  Weather API Steps:');
  clips.filter(c => c.name.startsWith('weather-step')).forEach(clip => {
    console.log(`  • ${clip.name} (${clip.id})`);
  });
  
  console.log('\n🗄️  CRM Query Steps:');
  clips.filter(c => c.name.startsWith('crm-step')).forEach(clip => {
    console.log(`  • ${clip.name} (${clip.id})`);
  });
  
  console.log('\n🗃️  SQLite Query Steps:');
  clips.filter(c => c.name.startsWith('sqlite-step')).forEach(clip => {
    console.log(`  • ${clip.name} (${clip.id})`);
  });
  
  console.log('\n🏗️  SQLite View Optimization Steps:');
  clips.filter(c => c.name.startsWith('sqlite-view-step')).forEach(clip => {
    console.log(`  • ${clip.name} (${clip.id})`);
  });
  
  console.log('\n🎯  Comprehensive Comparison Steps:');
  clips.filter(c => c.name.startsWith('comparison-step')).forEach(clip => {
    console.log(`  • ${clip.name} (${clip.id})`);
  });
}

function showUsage() {
  console.log('\n🎬 Remotion Clip Renderer\n');
  console.log('Usage:');
  console.log('  node render-clips.js --list                 # List all available clips');
  console.log('  node render-clips.js --all                  # Render all clips');
  console.log('  node render-clips.js <clip-name>           # Render specific clip');
  console.log('  node render-clips.js weather-step1         # Example: render weather step 1');
  console.log('\nExamples:');
  console.log('  node render-clips.js weather-step1-user-request');
  console.log('  node render-clips.js crm-step3-tool-execution');
  console.log('  node render-clips.js sqlite-step2-schema-discovery');
}

// Main execution
if (args.includes('--help') || args.includes('-h')) {
  showUsage();
} else if (args.includes('--list')) {
  listClips();
} else if (shouldRenderAll) {
  console.log('🚀 Rendering all clips...');
  clips.forEach(renderClip);
  console.log('\n🎉 All clips rendered!');
} else if (requestedClip) {
  const clip = clips.find(c => c.name === requestedClip || c.id === requestedClip);
  if (clip) {
    renderClip(clip);
  } else {
    console.error(`❌ Clip not found: ${requestedClip}`);
    console.log('💡 Use --list to see all available clips');
  }
} else {
  showUsage();
}
