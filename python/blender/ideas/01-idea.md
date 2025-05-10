https://claude.ai/chat/7b88d39c-e916-4a03-9efb-eceabd6fc9bd

<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>AI YouTube Shorts Generator</title>
    <style>
        body {
            margin: 0;
            padding: 0;
            font-family: Arial, sans-serif;
            background-color: #3c3c3c;
            color: #e0e0e0;
        }
        
        .addon-panel {
            width: 320px;
            background-color: #2b2b2b;
            border: 1px solid #505050;
            border-radius: 4px;
            padding: 10px;
            margin: 10px;
        }
        
        .panel-header {
            background-color: #4a4a4a;
            padding: 8px 10px;
            margin: -10px -10px 10px -10px;
            border-radius: 4px 4px 0 0;
            font-weight: bold;
            display: flex;
            align-items: center;
            justify-content: space-between;
        }
        
        .section {
            margin-bottom: 15px;
            padding: 10px;
            background-color: #3a3a3a;
            border-radius: 4px;
        }
        
        .section-header {
            font-weight: bold;
            margin-bottom: 10px;
            color: #b4b4b4;
            font-size: 14px;
            display: flex;
            align-items: center;
            gap: 5px;
        }
        
        .button {
            background-color: #5a5a5a;
            border: 1px solid #6a6a6a;
            color: #e0e0e0;
            padding: 6px 12px;
            border-radius: 3px;
            cursor: pointer;
            margin: 3px;
            font-size: 12px;
            width: 100%;
            text-align: left;
            display: flex;
            align-items: center;
            gap: 5px;
        }
        
        .button:hover {
            background-color: #6a6a6a;
        }
        
        .button.primary {
            background-color: #5680c2;
            border-color: #6690d2;
        }
        
        .button.primary:hover {
            background-color: #6690d2;
        }
        
        .button.ai {
            background-color: #8b5cc2;
            border-color: #9b6cd2;
        }
        
        .button.ai:hover {
            background-color: #9b6cd2;
        }
        
        .button.icon {
            padding: 4px 8px;
            width: auto;
        }
        
        .input-group {
            display: flex;
            align-items: center;
            margin-bottom: 8px;
        }
        
        .input-group label {
            flex: 1;
            font-size: 12px;
        }
        
        .input-group input, .input-group select {
            background-color: #4a4a4a;
            border: 1px solid #5a5a5a;
            color: #e0e0e0;
            padding: 4px;
            border-radius: 3px;
            width: 120px;
        }
        
        .checkbox-group {
            display: flex;
            align-items: center;
            margin-bottom: 8px;
        }
        
        .checkbox-group input[type="checkbox"] {
            margin-right: 8px;
        }
        
        .progress-section {
            margin-top: 10px;
            padding: 10px;
            background-color: #2a2a2a;
            border-radius: 4px;
        }
        
        .progress-bar {
            width: 100%;
            height: 6px;
            background-color: #4a4a4a;
            border-radius: 3px;
            overflow: hidden;
            margin: 10px 0;
        }
        
        .progress-fill {
            height: 100%;
            background-color: #8b5cc2;
            width: 0%;
            transition: width 0.3s ease;
        }
        
        .status-text {
            font-size: 11px;
            color: #888;
            text-align: center;
        }
        
        .asset-list {
            max-height: 120px;
            overflow-y: auto;
            margin-top: 8px;
            padding: 5px;
            background-color: #2a2a2a;
            border-radius: 3px;
        }
        
        .asset-item {
            background-color: #4a4a4a;
            padding: 6px;
            margin-bottom: 4px;
            border-radius: 3px;
            font-size: 11px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        
        .asset-controls {
            display: flex;
            gap: 3px;
        }
        
        .tab-container {
            display: flex;
            margin-bottom: 10px;
            background-color: #3a3a3a;
            border-radius: 4px;
            padding: 2px;
        }
        
        .tab {
            flex: 1;
            padding: 6px;
            text-align: center;
            cursor: pointer;
            border-radius: 3px;
            font-size: 12px;
        }
        
        .tab.active {
            background-color: #5680c2;
        }
        
        .preset-grid {
            display: grid;
            grid-template-columns: repeat(2, 1fr);
            gap: 5px;
            margin-top: 10px;
        }
        
        .preset-card {
            background-color: #4a4a4a;
            padding: 8px;
            border-radius: 4px;
            text-align: center;
            cursor: pointer;
            font-size: 11px;
        }
        
        .preset-card:hover {
            background-color: #5a5a5a;
        }
        
        .timeline-preview {
            height: 60px;
            background-color: #2a2a2a;
            border: 1px solid #4a4a4a;
            border-radius: 3px;
            margin-top: 10px;
            position: relative;
            overflow: hidden;
        }
        
        .timeline-track {
            position: absolute;
            height: 12px;
            top: 4px;
            border-radius: 2px;
        }
        
        .track-main { background-color: #507950; top: 4px; }
        .track-broll { background-color: #795050; top: 18px; }
        .track-text { background-color: #506079; top: 32px; }
        .track-effects { background-color: #796050; top: 46px; }
        
        .ai-settings {
            background-color: #2a2a2a;
            padding: 8px;
            border-radius: 3px;
            margin-top: 8px;
        }
        
        .ai-model-selector {
            display: flex;
            gap: 5px;
            margin-bottom: 8px;
        }
        
        .model-option {
            flex: 1;
            padding: 4px;
            text-align: center;
            background-color: #4a4a4a;
            border-radius: 3px;
            cursor: pointer;
            font-size: 11px;
        }
        
        .model-option.selected {
            background-color: #8b5cc2;
        }
    </style>
</head>
<body>
    <div class="addon-panel">
        <div class="panel-header">
            <span>🤖 AI YouTube Shorts Generator</span>
            <span>⚙️</span>
        </div>
        
        <div class="section">
            <div class="section-header">🚀 Quick Generate</div>
            <button class="button primary ai" onclick="generateShort()">✨ Generate YouTube Short</button>
            <button class="button ai" onclick="analyzeContent()">🔍 Analyze Content</button>
            <button class="button" onclick="previewShort()">👁️ Preview Result</button>
        </div>
        
        <div class="section">
            <div class="section-header">📝 Content Sources</div>
            <div class="input-group">
                <label>Transcript:</label>
                <button class="button icon" onclick="loadTranscript()">📄</button>
            </div>
            <div class="input-group">
                <label>Main Video:</label>
                <button class="button icon" onclick="selectMainVideo()">🎥</button>
            </div>
            <div class="input-group">
                <label>B-Roll Folder:</label>
                <button class="button icon" onclick="selectBRollFolder()">📁</button>
            </div>
            <div class="checkbox-group">
                <input type="checkbox" id="autoTranscribe" checked>
                <label for="autoTranscribe">Auto-transcribe from video</label>
            </div>
        </div>
        
        <div class="section">
            <div class="section-header">🎨 Animated Elements</div>
            <div class="tab-container">
                <div class="tab active" onclick="switchTab('templates')">Templates</div>
                <div class="tab" onclick="switchTab('custom')">Custom</div>
            </div>
            <div id="templatesTab">
                <div class="preset-grid">
                    <div class="preset-card">
                        <div>🔥 Fire Title</div>
                    </div>
                    <div class="preset-card">
                        <div>💫 Particle Burst</div>
                    </div>
                    <div class="preset-card">
                        <div>🌊 Wave Effect</div>
                    </div>
                    <div class="preset-card">
                        <div>📊 Data Viz</div>
                    </div>
                </div>
            </div>
            <div id="customTab" style="display: none;">
                <button class="button" onclick="importBlenderScene()">📦 Import Scene</button>
                <button class="button" onclick="createCustomEffect()">✨ Create Effect</button>
            </div>
        </div>
        
        <div class="section">
            <div class="section-header">🤖 AI Settings</div>
            <div class="ai-settings">
                <label style="font-size: 11px; display: block; margin-bottom: 5px;">AI Model:</label>
                <div class="ai-model-selector">
                    <div class="model-option selected">GPT-4</div>
                    <div class="model-option">Claude</div>
                    <div class="model-option">Local</div>
                </div>
                <div class="input-group">
                    <label>Style:</label>
                    <select>
                        <option>Viral Hook</option>
                        <option>Educational</option>
                        <option>Entertainment</option>
                        <option>Tutorial</option>
                        <option>Story Time</option>
                    </select>
                </div>
                <div class="input-group">
                    <label>Duration (s):</label>
                    <input type="number" value="60" min="15" max="60">
                </div>
                <div class="checkbox-group">
                    <input type="checkbox" id="autoSubtitles" checked>
                    <label for="autoSubtitles">Generate captions</label>
                </div>
                <div class="checkbox-group">
                    <input type="checkbox" id="musicGen" checked>
                    <label for="musicGen">Generate background music</label>
                </div>
            </div>
        </div>
        
        <div class="section">
            <div class="section-header">📊 Content Analysis</div>
            <div class="asset-list">
                <div class="asset-item">
                    <span>✅ Hook detected: "Did you know..."</span>
                    <div class="asset-controls">
                        <button class="button icon" onclick="editHook()">✏️</button>
                    </div>
                </div>
                <div class="asset-item">
                    <span>🎯 Key moments: 3 found</span>
                    <div class="asset-controls">
                        <button class="button icon" onclick="viewMoments()">👁️</button>
                    </div>
                </div>
                <div class="asset-item">
                    <span>🎵 Music sync points: 5</span>
                    <div class="asset-controls">
                        <button class="button icon" onclick="adjustSync()">🔧</button>
                    </div>
                </div>
            </div>
        </div>
        
        <div class="section">
            <div class="section-header">🎬 Timeline Preview</div>
            <div class="timeline-preview">
                <div class="timeline-track track-main" style="width: 100%; left: 0;"></div>
                <div class="timeline-track track-broll" style="width: 30%; left: 20%;"></div>
                <div class="timeline-track track-broll" style="width: 25%; left: 60%;"></div>
                <div class="timeline-track track-text" style="width: 15%; left: 0;"></div>
                <div class="timeline-track track-text" style="width: 20%; left: 40%;"></div>
                <div class="timeline-track track-effects" style="width: 10%; left: 0;"></div>
                <div class="timeline-track track-effects" style="width: 15%; left: 50%;"></div>
            </div>
        </div>
        
        <div class="section">
            <div class="section-header">⚡ Generation Queue</div>
            <div class="progress-section">
                <div class="status-text">Processing: Analyzing transcript...</div>
                <div class="progress-bar">
                    <div class="progress-fill" id="progressBar"></div>
                </div>
                <div style="display: flex; gap: 5px; margin-top: 10px;">
                    <button class="button" onclick="pauseGeneration()">⏸️ Pause</button>
                    <button class="button" onclick="cancelGeneration()">❌ Cancel</button>
                </div>
            </div>
        </div>
        
        <div class="section">
            <div class="section-header">📤 Export Options</div>
            <div class="input-group">
                <label>Format:</label>
                <select>
                    <option>YouTube Shorts (9:16)</option>
                    <option>TikTok (9:16)</option>
                    <option>Instagram Reels (9:16)</option>
                    <option>Square (1:1)</option>
                </select>
            </div>
            <div class="input-group">
                <label>Quality:</label>
                <select>
                    <option>1080p</option>
                    <option>4K</option>
                    <option>720p</option>
                </select>
            </div>
            <button class="button primary" onclick="exportShort()">🎯 Export YouTube Short</button>
        </div>
    </div>

    <script>
        // Mock functions for UI demonstration
        function generateShort() {
            alert('Starting AI generation process...');
            animateProgress();
        }
        
        function analyzeContent() {
            alert('Analyzing transcript and video content...');
            animateProgress();
        }
        
        function previewShort() {
            alert('Opening preview window...');
        }
        
        function loadTranscript() {
            alert('Loading transcript file...');
        }
        
        function selectMainVideo() {
            alert('Selecting main video file...');
        }
        
        function selectBRollFolder() {
            alert('Selecting B-roll folder...');
        }
        
        function switchTab(tab) {
            const tabs = document.querySelectorAll('.tab');
            tabs.forEach(t => t.classList.remove('active'));
            event.target.classList.add('active');
            
            document.getElementById('templatesTab').style.display = tab === 'templates' ? 'block' : 'none';
            document.getElementById('customTab').style.display = tab === 'custom' ? 'block' : 'none';
        }
        
        function importBlenderScene() {
            alert('Importing Blender scene...');
        }
        
        function createCustomEffect() {
            alert('Opening effect creator...');
        }
        
        function editHook() {
            alert('Editing hook text...');
        }
        
        function viewMoments() {
            alert('Viewing key moments...');
        }
        
        function adjustSync() {
            alert('Adjusting music sync points...');
        }
        
        function pauseGeneration() {
            alert('Pausing generation...');
        }
        
        function cancelGeneration() {
            alert('Cancelling generation...');
        }
        
        function exportShort() {
            alert('Exporting YouTube Short...');
            animateProgress();
        }
        
        function animateProgress() {
            const progressBar = document.getElementById('progressBar');
            progressBar.style.width = '0%';
            
            let progress = 0;
            const interval = setInterval(() => {
                progress += 2;
                progressBar.style.width = progress + '%';
                
                if (progress >= 100) {
                    clearInterval(interval);
                    setTimeout(() => {
                        progressBar.style.width = '0%';
                    }, 1000);
                }
            }, 50);
        }
        
        // Model selector functionality
        document.querySelectorAll('.model-option').forEach(option => {
            option.addEventListener('click', function() {
                document.querySelectorAll('.model-option').forEach(o => o.classList.remove('selected'));
                this.classList.add('selected');
            });
        });
        
        // Preset card functionality
        document.querySelectorAll('.preset-card').forEach(card => {
            card.addEventListener('click', function() {
                alert(`Selected preset: ${this.innerText}`);
            });
        });
    </script>
</body>
</html>