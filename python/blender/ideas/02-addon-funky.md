https://claude.ai/chat/14b874ae-9910-4c96-ad7b-9a7c2c8cbbd3

<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>AI Shorts Magic - Blender Addon</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: 'Arial', sans-serif;
            background: #1e1e2e;
            color: #e0e0e0;
            line-height: 1.6;
        }

        .addon-panel {
            width: 350px;
            background: #2a2a3a;
            height: 100vh;
            overflow-y: auto;
            padding: 10px;
        }

        .panel-header {
            background: linear-gradient(45deg, #ff4081, #8e24aa, #448aff);
            padding: 12px;
            border-radius: 8px;
            margin-bottom: 15px;
            text-align: center;
            position: relative;
            overflow: hidden;
        }

        .panel-header::before {
            content: "";
            position: absolute;
            top: 0;
            left: -100%;
            width: 100%;
            height: 100%;
            background: linear-gradient(90deg, transparent, rgba(255,255,255,0.2), transparent);
            animation: shine 3s infinite;
        }

        @keyframes shine {
            to {
                left: 100%;
            }
        }

        .panel-header h1 {
            font-size: 20px;
            font-weight: bold;
            color: white;
            text-shadow: 2px 2px 4px rgba(0,0,0,0.3);
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 8px;
        }

        .section {
            background: #33334a;
            border-radius: 8px;
            padding: 12px;
            margin-bottom: 12px;
            border: 1px solid #444466;
        }

        .section-header {
            font-weight: bold;
            font-size: 14px;
            margin-bottom: 10px;
            color: #ffffff;
            display: flex;
            align-items: center;
            gap: 6px;
        }

        .button {
            padding: 8px 12px;
            border: none;
            border-radius: 6px;
            cursor: pointer;
            font-weight: bold;
            font-size: 13px;
            transition: all 0.3s ease;
            margin: 4px 0;
            width: 100%;
            display: flex;
            align-items: center;
            gap: 8px;
            justify-content: center;
        }

        .primary-button {
            background: linear-gradient(45deg, #ff4081, #8e24aa);
            color: white;
        }

        .primary-button:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(255, 64, 129, 0.4);
        }

        .secondary-button {
            background: #4a4a6a;
            color: white;
            border: 1px solid #6a6a8a;
        }

        .secondary-button:hover {
            background: #5a5a7a;
            transform: translateY(-1px);
        }

        .accent-button {
            background: linear-gradient(45deg, #00bcd4, #2196f3);
            color: white;
        }

        .accent-button:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(33, 150, 243, 0.4);
        }

        .input-group {
            margin-bottom: 10px;
        }

        .input-group label {
            display: block;
            margin-bottom: 4px;
            font-size: 12px;
            color: #a0a0a0;
        }

        .input-group input,
        .input-group select,
        .input-group textarea {
            width: 100%;
            padding: 8px;
            border-radius: 4px;
            border: 1px solid #444466;
            background: #1e1e2e;
            color: #e0e0e0;
            font-size: 13px;
        }

        .input-group textarea {
            min-height: 80px;
            resize: vertical;
        }

        .timeline-strip {
            background: #1e1e2e;
            border-radius: 4px;
            padding: 8px;
            margin: 8px 0;
            display: flex;
            align-items: center;
            gap: 8px;
            border: 1px solid #444466;
        }

        .strip-icon {
            width: 30px;
            height: 30px;
            background: #4a4a6a;
            border-radius: 4px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 16px;
        }

        .strip-info {
            flex: 1;
        }

        .strip-title {
            font-weight: bold;
            font-size: 12px;
        }

        .strip-duration {
            font-size: 11px;
            color: #888;
        }

        .tabs {
            display: flex;
            gap: 4px;
            margin-bottom: 12px;
        }

        .tab {
            flex: 1;
            padding: 8px;
            text-align: center;
            background: #3a3a4a;
            border: none;
            color: #a0a0a0;
            cursor: pointer;
            border-radius: 4px 4px 0 0;
            font-size: 12px;
            transition: all 0.3s ease;
        }

        .tab.active {
            background: #4a4a6a;
            color: white;
            border-bottom: 2px solid #ff4081;
        }

        .progress-bar {
            width: 100%;
            height: 6px;
            background: #1e1e2e;
            border-radius: 3px;
            overflow: hidden;
            margin: 10px 0;
        }

        .progress-fill {
            height: 100%;
            background: linear-gradient(90deg, #ff4081, #8e24aa, #448aff);
            width: 0%;
            transition: width 0.5s ease;
        }

        .modal {
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0,0,0,0.8);
            display: none;
            align-items: center;
            justify-content: center;
            z-index: 1000;
        }

        .modal-content {
            background: #2a2a3a;
            padding: 20px;
            border-radius: 12px;
            max-width: 500px;
            width: 90%;
            border: 1px solid #444466;
        }

        .preset-grid {
            display: grid;
            grid-template-columns: repeat(2, 1fr);
            gap: 8px;
            margin-top: 10px;
        }

        .preset-card {
            background: #1e1e2e;
            padding: 10px;
            border-radius: 6px;
            border: 1px solid #444466;
            cursor: pointer;
            transition: all 0.3s ease;
            text-align: center;
        }

        .preset-card:hover {
            border-color: #ff4081;
            transform: translateY(-2px);
        }

        .preset-icon {
            font-size: 24px;
            margin-bottom: 6px;
        }

        .preset-name {
            font-size: 12px;
            font-weight: bold;
        }

        .pill {
            display: inline-block;
            padding: 4px 8px;
            border-radius: 12px;
            background: #4a4a6a;
            font-size: 11px;
            margin: 2px;
        }

        .ai-status {
            padding: 8px;
            background: #1e1e2e;
            border-radius: 6px;
            text-align: center;
            font-size: 12px;
            margin: 10px 0;
            border: 1px solid transparent;
            border-image: linear-gradient(45deg, #ff4081, #8e24aa, #448aff) 1;
        }

        .close-button {
            float: right;
            background: none;
            border: none;
            color: #a0a0a0;
            font-size: 20px;
            cursor: pointer;
            padding: 0;
            width: auto;
        }

        .close-button:hover {
            color: #ffffff;
        }
    </style>
</head>
<body>
    <div class="addon-panel">
        <div class="panel-header">
            <h1>🚀 AI Shorts Magic ✨</h1>
        </div>

        <div class="tabs">
            <button class="tab active" onclick="switchTab('setup')">Setup</button>
            <button class="tab" onclick="switchTab('generate')">Generate</button>
            <button class="tab" onclick="switchTab('edit')">Edit</button>
        </div>

        <!-- Setup Tab -->
        <div id="setup-tab" class="tab-content">
            <div class="section">
                <div class="section-header">📁 Import Assets</div>
                <button class="secondary-button" onclick="selectVideo()">
                    🎬 Select Main Video
                </button>
                <button class="secondary-button" onclick="selectBRoll()">
                    🎥 Add B-Roll Footage
                </button>
                <button class="secondary-button" onclick="importTranscript()">
                    📝 Import Transcript
                </button>
            </div>

            <div class="section">
                <div class="section-header">🎨 Presets</div>
                <button class="accent-button" onclick="openPresetModal()">
                    ⚡ Quick Start Templates
                </button>
            </div>

            <div class="section">
                <div class="section-header">🎬 Current Timeline</div>
                <div class="timeline-strip">
                    <div class="strip-icon">🎥</div>
                    <div class="strip-info">
                        <div class="strip-title">Main Video</div>
                        <div class="strip-duration">2:34</div>
                    </div>
                </div>
                <div class="timeline-strip">
                    <div class="strip-icon">🌟</div>
                    <div class="strip-info">
                        <div class="strip-title">B-Roll 1</div>
                        <div class="strip-duration">0:45</div>
                    </div>
                </div>
                <div class="timeline-strip">
                    <div class="strip-icon">✨</div>
                    <div class="strip-info">
                        <div class="strip-title">B-Roll 2</div>
                        <div class="strip-duration">0:30</div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Generate Tab -->
        <div id="generate-tab" class="tab-content" style="display: none;">
            <div class="section">
                <div class="section-header">🤖 AI Settings</div>
                <div class="input-group">
                    <label>Video Style</label>
                    <select>
                        <option>💫 Energetic Coding Tutorial</option>
                        <option>🔥 Fast-Paced Tech News</option>
                        <option>🎯 Problem-Solution Format</option>
                        <option>🚀 Startup Vibes</option>
                        <option>🎮 Gaming Tech</option>
                    </select>
                </div>
                <div class="input-group">
                    <label>Duration</label>
                    <select>
                        <option>⚡ 15-30 seconds</option>
                        <option>🎯 30-60 seconds</option>
                        <option>📺 60-90 seconds</option>
                    </select>
                </div>
                <div class="input-group">
                    <label>Hook Type</label>
                    <select>
                        <option>❓ Question Hook</option>
                        <option>🤯 Mind-Blowing Fact</option>
                        <option>⚠️ Problem Statement</option>
                        <option>🎬 Action Scene</option>
                    </select>
                </div>
            </div>

            <div class="section">
                <div class="section-header">✍️ AI Script Assistant</div>
                <div class="input-group">
                    <label>Key Points (AI will enhance)</label>
                    <textarea placeholder="Enter your main points here..."></textarea>
                </div>
                <button class="primary-button" onclick="generateScript()">
                    🧠 Generate Smart Script
                </button>
            </div>

            <div class="section">
                <div class="section-header">🎭 Title Sequences</div>
                <button class="secondary-button" onclick="generateTitleSequence()">
                    ✨ Generate Title Animation
                </button>
                <div class="input-group">
                    <label>Title Text</label>
                    <input type="text" placeholder="Your awesome title...">
                </div>
            </div>

            <div class="ai-status">
                <div>🤖 AI Ready to Generate Magic!</div>
                <div class="progress-bar">
                    <div class="progress-fill" id="progress"></div>
                </div>
            </div>

            <button class="primary-button" onclick="generateShort()">
                🚀 Generate YouTube Short
            </button>
        </div>

        <!-- Edit Tab -->
        <div id="edit-tab" class="tab-content" style="display: none;">
            <div class="section">
                <div class="section-header">✂️ Quick Split Tools</div>
                <button class="secondary-button" onclick="autoSplitByBeats()">
                    🎵 Split by Beats
                </button>
                <button class="secondary-button" onclick="autoSplitByTranscript()">
                    💬 Split by Transcript
                </button>
                <button class="secondary-button" onclick="manualSplit()">
                    ✂️ Manual Split Mode
                </button>
            </div>

            <div class="section">
                <div class="section-header">🎨 B-Roll Editor</div>
                <div class="input-group">
                    <label>B-Roll Timing Mode</label>
                    <select>
                        <option>🎯 Auto-detect Good Moments</option>
                        <option>⏱️ Every X Seconds</option>
                        <option>💬 On Keywords</option>
                        <option>✋ Manual</option>
                    </select>
                </div>
                <button class="accent-button" onclick="previewBRoll()">
                    👁️ Preview B-Roll Placement
                </button>
            </div>

            <div class="section">
                <div class="section-header">🎭 Effects & Transitions</div>
                <div class="preset-grid">
                    <div class="preset-card" onclick="applyEffect('glitch')">
                        <div class="preset-icon">💥</div>
                        <div class="preset-name">Glitch</div>
                    </div>
                    <div class="preset-card" onclick="applyEffect('zoom')">
                        <div class="preset-icon">🔍</div>
                        <div class="preset-name">Zoom</div>
                    </div>
                    <div class="preset-card" onclick="applyEffect('rgb')">
                        <div class="preset-icon">🌈</div>
                        <div class="preset-name">RGB Split</div>
                    </div>
                    <div class="preset-card" onclick="applyEffect('shake')">
                        <div class="preset-icon">📳</div>
                        <div class="preset-name">Shake</div>
                    </div>
                </div>
            </div>

            <button class="primary-button" onclick="finalizEdit()">
                ✨ Finalize Edit
            </button>
        </div>
    </div>

    <!-- Preset Modal -->
    <div class="modal" id="presetModal">
        <div class="modal-content">
            <button class="close-button" onclick="closePresetModal()">×</button>
            <h2 style="margin-bottom: 16px;">⚡ Quick Start Templates</h2>
            <div class="preset-grid">
                <div class="preset-card" onclick="selectPreset('tutorial')">
                    <div class="preset-icon">👨‍💻</div>
                    <div class="preset-name">Coding Tutorial</div>
                </div>
                <div class="preset-card" onclick="selectPreset('tips')">
                    <div class="preset-icon">💡</div>
                    <div class="preset-name">Tech Tips</div>
                </div>
                <div class="preset-card" onclick="selectPreset('news')">
                    <div class="preset-icon">📰</div>
                    <div class="preset-name">Tech News</div>
                </div>
                <div class="preset-card" onclick="selectPreset('challenge')">
                    <div class="preset-icon">🏆</div>
                    <div class="preset-name">Code Challenge</div>
                </div>
                <div class="preset-card" onclick="selectPreset('review')">
                    <div class="preset-icon">⭐</div>
                    <div class="preset-name">Tool Review</div>
                </div>
                <div class="preset-card" onclick="selectPreset('showcase')">
                    <div class="preset-icon">🎪</div>
                    <div class="preset-name">Project Showcase</div>
                </div>
            </div>
        </div>
    </div>

    <script>
        function switchTab(tab) {
            // Hide all tabs
            document.querySelectorAll('.tab-content').forEach(content => {
                content.style.display = 'none';
            });
            
            // Remove active class from all tabs
            document.querySelectorAll('.tab').forEach(tabBtn => {
                tabBtn.classList.remove('active');
            });
            
            // Show selected tab
            document.getElementById(tab + '-tab').style.display = 'block';
            
            // Add active class to clicked tab
            event.target.classList.add('active');
        }

        function selectVideo() {
            console.log('Opening video selector...');
            // Simulate file selection
            setTimeout(() => {
                alert('Video selected: coding_tutorial_raw.mp4');
            }, 500);
        }

        function selectBRoll() {
            console.log('Opening B-roll selector...');
            setTimeout(() => {
                alert('B-roll added: screen_recording_01.mp4');
            }, 500);
        }

        function importTranscript() {
            console.log('Importing transcript...');
            setTimeout(() => {
                alert('Transcript imported successfully!');
            }, 500);
        }

        function openPresetModal() {
            document.getElementById('presetModal').style.display = 'flex';
        }

        function closePresetModal() {
            document.getElementById('presetModal').style.display = 'none';
        }

        function selectPreset(preset) {
            console.log('Selected preset:', preset);
            closePresetModal();
            alert(`Preset "${preset}" applied! Settings configured for optimal ${preset} content.`);
        }

        function generateScript() {
            console.log('Generating AI script...');
            // Simulate progress
            let progress = 0;
            const progressBar = document.getElementById('progress');
            const interval = setInterval(() => {
                progress += 10;
                progressBar.style.width = progress + '%';
                if (progress >= 100) {
                    clearInterval(interval);
                    alert('AI Script generated! Your content is now optimized for engagement!');
                    progressBar.style.width = '0%';
                }
            }, 200);
        }

        function generateTitleSequence() {
            console.log('Generating title sequence...');
            setTimeout(() => {
                alert('Title sequence created with sick animations! 🔥');
            }, 1000);
        }

        function generateShort() {
            console.log('Generating YouTube short...');
            let progress = 0;
            const progressBar = document.getElementById('progress');
            const interval = setInterval(() => {
                progress += 5;
                progressBar.style.width = progress + '%';
                if (progress >= 100) {
                    clearInterval(interval);
                    alert('YouTube Short generated! Ready for viral success! 🚀');
                    progressBar.style.width = '0%';
                }
            }, 150);
        }

        function autoSplitByBeats() {
            console.log('Auto-splitting by beats...');
            setTimeout(() => {
                alert('Video split into 12 segments based on audio beats!');
            }, 800);
        }

        function autoSplitByTranscript() {
            console.log('Auto-splitting by transcript...');
            setTimeout(() => {
                alert('Video split into 8 segments based on transcript sections!');
            }, 800);
        }

        function manualSplit() {
            console.log('Entering manual split mode...');
            alert('Manual split mode activated! Click on timeline to add cut points.');
        }

        function previewBRoll() {
            console.log('Previewing B-roll placement...');
            setTimeout(() => {
                alert('B-roll preview ready! 5 optimal insertion points found.');
            }, 600);
        }

        function applyEffect(effect) {
            console.log('Applying effect:', effect);
            alert(`${effect} effect applied! Your video just got 200% cooler! 😎`);
        }

        function finalizEdit() {
            console.log('Finalizing edit...');
            setTimeout(() => {
                alert('Edit finalized! Your YouTube Short is ready to go viral! 🎉');
            }, 1000);
        }
    </script>
</body>
</html>