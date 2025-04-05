/**
 * Severance-inspired Museum Display Webapp
 * A self-contained frontend application for displaying museum content from JSON files
 */

// Global state
const state = {
    loadedFiles: [],
    currentFileIndex: -1,
    currentPageIndex: -1,
    currentSlideIndices: {},
    currentExampleIndices: {},
    currentQuestionIndices: {},
    quizAnswers: {},
    navigationVisible: true
};

// DOM Elements
const elements = {
    fileInterface: document.getElementById('fileInterface'),
    mainInterface: document.getElementById('mainInterface'),
    fileUpload: document.getElementById('fileUpload'),
    fileInfo: document.getElementById('fileInfo'),
    fileList: document.getElementById('fileList'),
    fileSelector: document.getElementById('fileSelector'),
    uploadButton: document.getElementById('uploadButton'),
    displayTitle: document.getElementById('displayTitle'),
    menuToggle: document.getElementById('menuToggle'),
    closeNav: document.getElementById('closeNav'),
    navigation: document.getElementById('navigation'),
    navList: document.getElementById('navList'),
    progressContainer: document.getElementById('progressContainer'),
    progressBar: document.getElementById('progressBar'),
    progressText: document.getElementById('progressText'),
    contentContainer: document.getElementById('contentContainer'),
    footerText: document.getElementById('footerText'),
    footerLogos: document.getElementById('footerLogos'),
    loadingOverlay: document.getElementById('loadingOverlay')
};

// Templates
const templates = {
    slideDeck: document.getElementById('slideDeckTemplate'),
    tutorial: document.getElementById('tutorialTemplate'),
    tutorialStep: document.getElementById('tutorialStepTemplate'),
    interactiveCode: document.getElementById('interactiveCodeTemplate'),
    hardwareVisual: document.getElementById('hardwareVisualTemplate'),
    hardwarePanel: document.getElementById('hardwarePanelTemplate'),
    bioGallery: document.getElementById('bioGalleryTemplate'),
    bioCard: document.getElementById('bioCardTemplate'),
    resourceList: document.getElementById('resourceListTemplate'),
    resourceItem: document.getElementById('resourceItemTemplate'),
    quiz: document.getElementById('quizTemplate'),
    option: document.getElementById('optionTemplate')
};

// Initialize the application
document.addEventListener('DOMContentLoaded', () => {
    // Hide loading overlay
    toggleLoading(false);
    
    // Set up event listeners
    setupEventListeners();
});

// Set up all event listeners
function setupEventListeners() {
    // File upload
    elements.fileUpload.addEventListener('change', handleFileUpload);
    
    // File selection from dropdown
    elements.fileSelector.addEventListener('change', (e) => {
        const selectedIndex = parseInt(e.target.value);
        if (!isNaN(selectedIndex) && selectedIndex >= 0) {
            loadFile(selectedIndex);
        }
    });
    
    // Upload button in main interface
    elements.uploadButton.addEventListener('click', () => {
        elements.fileUpload.click();
    });
    
    // Navigation toggle for mobile
    elements.menuToggle.addEventListener('click', () => {
        elements.navigation.classList.add('active');
    });
    
    elements.closeNav.addEventListener('click', () => {
        elements.navigation.classList.remove('active');
    });
    
    // Close navigation when clicking on a link (mobile)
    elements.navList.addEventListener('click', (e) => {
        if (e.target.tagName === 'A') {
            elements.navigation.classList.remove('active');
        }
    });
}

// Handle file upload
async function handleFileUpload(event) {
    const files = event.target.files;
    if (!files || files.length === 0) return;
    
    toggleLoading(true);
    
    try {
        for (let i = 0; i < files.length; i++) {
            const file = files[i];
            const fileContent = await readFileContent(file);
            const jsonData = JSON.parse(fileContent);
            
            // Validate JSON structure
            if (validateJsonStructure(jsonData)) {
                // Add to loaded files
                state.loadedFiles.push({
                    name: file.name,
                    data: jsonData
                });
                
                // Update file list
                updateFileList();
                
                // Update file selector
                updateFileSelector();
                
                // If this is the first file, load it
                if (state.loadedFiles.length === 1) {
                    loadFile(0);
                }
            } else {
                console.error('Invalid JSON structure in file:', file.name);
                alert(`Invalid JSON structure in file: ${file.name}`);
            }
        }
    } catch (error) {
        console.error('Error processing files:', error);
        alert('Error processing files: ' + error.message);
    } finally {
        toggleLoading(false);
        // Reset file input
        event.target.value = '';
    }
}

// Read file content as text
function readFileContent(file) {
    return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = (e) => resolve(e.target.result);
        reader.onerror = (e) => reject(new Error('Error reading file'));
        reader.readAsText(file);
    });
}

// Validate JSON structure
function validateJsonStructure(jsonData) {
    // Check if it's an array with at least one object
    if (!Array.isArray(jsonData) || jsonData.length === 0) {
        return false;
    }
    
    // Get the first object
    const firstObject = jsonData[0];
    
    // Check if it has a museum display key
    const displayKey = Object.keys(firstObject)[0];
    if (!displayKey) return false;
    
    const display = firstObject[displayKey];
    
    // Check required properties
    if (!display.title || !display.pages || !Array.isArray(display.pages) || display.pages.length === 0) {
        return false;
    }
    
    return true;
}

// Update file list in the file interface
function updateFileList() {
    elements.fileList.innerHTML = '';
    elements.fileInfo.textContent = state.loadedFiles.length > 0 
        ? `${state.loadedFiles.length} file(s) loaded` 
        : 'No file selected';
    
    state.loadedFiles.forEach((file, index) => {
        const li = document.createElement('li');
        
        const displayName = document.createElement('span');
        displayName.textContent = file.name;
        li.appendChild(displayName);
        
        const buttonContainer = document.createElement('div');
        
        const loadButton = document.createElement('button');
        loadButton.textContent = 'LOAD';
        loadButton.addEventListener('click', () => loadFile(index));
        buttonContainer.appendChild(loadButton);
        
        const removeButton = document.createElement('button');
        removeButton.textContent = 'REMOVE';
        removeButton.addEventListener('click', () => removeFile(index));
        buttonContainer.appendChild(removeButton);
        
        li.appendChild(buttonContainer);
        elements.fileList.appendChild(li);
    });
}

// Update file selector in the main interface
function updateFileSelector() {
    elements.fileSelector.innerHTML = '';
    
    const defaultOption = document.createElement('option');
    defaultOption.value = '';
    defaultOption.textContent = 'SELECT DISPLAY';
    elements.fileSelector.appendChild(defaultOption);
    
    state.loadedFiles.forEach((file, index) => {
        const option = document.createElement('option');
        option.value = index;
        option.textContent = file.name;
        elements.fileSelector.appendChild(option);
    });
}

// Remove a file from the loaded files
function removeFile(index) {
    if (index < 0 || index >= state.loadedFiles.length) return;
    
    state.loadedFiles.splice(index, 1);
    updateFileList();
    updateFileSelector();
    
    // If we removed the current file
    if (index === state.currentFileIndex) {
        if (state.loadedFiles.length > 0) {
            // Load the first file
            loadFile(0);
        } else {
            // No files left, show file interface
            showFileInterface();
        }
    } else if (index < state.currentFileIndex) {
        // Adjust current index
        state.currentFileIndex--;
    }
}

// Load a file by index
function loadFile(index) {
    if (index < 0 || index >= state.loadedFiles.length) return;
    
    toggleLoading(true);
    
    try {
        state.currentFileIndex = index;
        elements.fileSelector.value = index;
        
        // Get the display data
        const fileData = state.loadedFiles[index].data;
        const displayKey = Object.keys(fileData[0])[0];
        const displayData = fileData[0][displayKey];
        
        // Set display title
        elements.displayTitle.textContent = displayData.title;
        
        // Set up navigation
        setupNavigation(displayData);
        
        // Set up footer
        setupFooter(displayData);
        
        // Show main interface
        showMainInterface();
        
        // Load first page
        if (displayData.pages.length > 0) {
            loadPage(0);
        }
    } catch (error) {
        console.error('Error loading file:', error);
        alert('Error loading file: ' + error.message);
    } finally {
        toggleLoading(false);
    }
}

// Set up navigation
function setupNavigation(displayData) {
    // Clear navigation
    elements.navList.innerHTML = '';
    
    // Add pages to navigation
    displayData.pages.forEach((page, index) => {
        const li = document.createElement('li');
        const a = document.createElement('a');
        a.href = '#';
        a.textContent = page.title;
        a.dataset.pageIndex = index;
        a.addEventListener('click', (e) => {
            e.preventDefault();
            loadPage(index);
        });
        li.appendChild(a);
        elements.navList.appendChild(li);
    });
    
    // Set up navigation type
    if (displayData.navigation) {
        // Set navigation type
        if (displayData.navigation.type) {
            elements.navigation.className = 'navigation ' + displayData.navigation.type;
        }
        
        // Show/hide progress
        if (displayData.navigation.show_progress) {
            elements.progressContainer.classList.remove('hidden');
            updateProgress(0, displayData.pages.length);
        } else {
            elements.progressContainer.classList.add('hidden');
        }
    }
}

// Set up footer
function setupFooter(displayData) {
    // Clear footer
    elements.footerText.innerHTML = '';
    elements.footerLogos.innerHTML = '';
    
    // Set footer text
    if (displayData.footer && displayData.footer.text) {
        elements.footerText.textContent = displayData.footer.text;
    }
    
    // Set footer logos
    if (displayData.footer && displayData.footer.logos && Array.isArray(displayData.footer.logos)) {
        displayData.footer.logos.forEach(logo => {
            const img = document.createElement('img');
            img.src = logo;
            img.alt = 'Logo';
            elements.footerLogos.appendChild(img);
        });
    }
}

// Load a page by index
function loadPage(index) {
    toggleLoading(true);
    
    try {
        // Get the current file data
        const fileData = state.loadedFiles[state.currentFileIndex].data;
        const displayKey = Object.keys(fileData[0])[0];
        const displayData = fileData[0][displayKey];
        
        // Check if index is valid
        if (index < 0 || index >= displayData.pages.length) {
            throw new Error('Invalid page index');
        }
        
        // Update current page index
        state.currentPageIndex = index;
        
        // Update active navigation item
        const navLinks = elements.navList.querySelectorAll('a');
        navLinks.forEach(link => {
            if (parseInt(link.dataset.pageIndex) === index) {
                link.classList.add('active');
            } else {
                link.classList.remove('active');
            }
        });
        
        // Update progress
        if (displayData.navigation && displayData.navigation.show_progress) {
            updateProgress(index + 1, displayData.pages.length);
        }
        
        // Get the page data
        const pageData = displayData.pages[index];
        
        // Clear content container
        elements.contentContainer.innerHTML = '';
        
        // Render page based on type
        switch (pageData.type) {
            case 'slide_deck':
                renderSlideDeck(pageData);
                break;
            case 'tutorial':
                renderTutorial(pageData);
                break;
            case 'interactive_code':
                renderInteractiveCode(pageData);
                break;
            case 'hardware_visual':
                renderHardwareVisual(pageData);
                break;
            case 'bio_gallery':
                renderBioGallery(pageData);
                break;
            case 'resource_list':
                renderResourceList(pageData);
                break;
            case 'quiz':
                renderQuiz(pageData);
                break;
            default:
                elements.contentContainer.innerHTML = `<div class="error">Unknown page type: ${pageData.type}</div>`;
        }
    } catch (error) {
        console.error('Error loading page:', error);
        elements.contentContainer.innerHTML = `<div class="error">Error loading page: ${error.message}</div>`;
    } finally {
        toggleLoading(false);
    }
}

// Update progress bar and text
function updateProgress(current, total) {
    const percentage = (current / total) * 100;
    elements.progressBar.innerHTML = `<div style="width: ${percentage}%"></div>`;
    elements.progressText.textContent = `${current}/${total}`;
}

// Render slide deck
function renderSlideDeck(pageData) {
    // Clone template
    const template = templates.slideDeck.content.cloneNode(true);
    
    // Set page title
    template.querySelector('.page-title').textContent = pageData.title;
    
    // Get slide container
    const slideContent = template.querySelector('.slide-content');
    const slideCounter = template.querySelector('.slide-counter');
    const prevButton = template.querySelector('.prev-slide');
    const nextButton = template.querySelector('.next-slide');
    
    // Initialize slide index
    if (!state.currentSlideIndices[pageData.id]) {
        state.currentSlideIndices[pageData.id] = 0;
    }
    
    // Function to render current slide
    const renderSlide = () => {
        const currentIndex = state.currentSlideIndices[pageData.id];
        const slide = pageData.slides[currentIndex];
        
        // Clear slide content
        slideContent.innerHTML = '';
        
        // Add slide title
        const slideTitle = document.createElement('h3');
        slideTitle.textContent = slide.title;
        slideContent.appendChild(slideTitle);
        
        // Add slide content
        const slideContentDiv = document.createElement('div');
        slideContentDiv.className = 'markdown';
        
        // Check if slide has code property
        if (slide.code) {
            // Create code block
            const codeBlock = document.createElement('pre');
            const code = document.createElement('code');
            code.textContent = slide.code;
            codeBlock.appendChild(code);
            slideContentDiv.appendChild(codeBlock);
        } else if (slide.content) {
            // Regular content
            slideContentDiv.innerHTML = formatMarkdown(slide.content);
        }
        
        slideContent.appendChild(slideContentDiv);
        
        // Update slide counter
        slideCounter.textContent = `${currentIndex + 1}/${pageData.slides.length}`;
        
        // Update button states
        prevButton.disabled = currentIndex === 0;
        nextButton.disabled = currentIndex === pageData.slides.length - 1;
    };
    
    // Set up navigation buttons
    prevButton.addEventListener('click', () => {
        if (state.currentSlideIndices[pageData.id] > 0) {
            state.currentSlideIndices[pageData.id]--;
            renderSlide();
        }
    });
    
    nextButton.addEventListener('click', () => {
        if (state.currentSlideIndices[pageData.id] < pageData.slides.length - 1) {
            state.currentSlideIndices[pageData.id]++;
            renderSlide();
        }
    });
    
    // Render initial slide
    renderSlide();
    
    // Add to content container
    elements.contentContainer.appendChild(template);
}

// Render tutorial
function renderTutorial(pageData) {
    // Clone template
    const template = templates.tutorial.content.cloneNode(true);
    
    // Set page title
    template.querySelector('.page-title').textContent = pageData.title;
    
    // Get steps container
    const stepsContainer = template.querySelector('.steps-container');
    
    // Add steps
    pageData.steps.forEach(step => {
        // Clone step template
        const stepTemplate = templates.tutorialStep.content.cloneNode(true);
        
        // Set step title
        stepTemplate.querySelector('.step-title').textContent = step.title;
        
        // Set step description
        stepTemplate.querySelector('.step-description').innerHTML = formatMarkdown(step.description);
        
        // Add to steps container
        stepsContainer.appendChild(stepTemplate);
    });
    
    // Add to content container
    elements.contentContainer.appendChild(template);
}

// Render interactive code
function renderInteractiveCode(pageData) {
    // Clone template
    const template = templates.interactiveCode.content.cloneNode(true);
    
    // Set page title
    template.querySelector('.page-title').textContent = pageData.title;
    
    // Get code selector
    const codeSelector = template.querySelector('.code-selector');
    
    // Get code example elements
    const exampleTitle = template.querySelector('.example-title');
    const exampleDescription = template.querySelector('.example-description');
    const codeElement = template.querySelector('code');
    const copyButton = template.querySelector('.copy-code');
    
    // Initialize example index
    if (!state.currentExampleIndices[pageData.id]) {
        state.currentExampleIndices[pageData.id] = 0;
    }
    
    // Add examples to selector
    pageData.examples.forEach((example, index) => {
        const option = document.createElement('option');
        option.value = index;
        option.textContent = example.title;
        codeSelector.appendChild(option);
    });
    
    // Set initial selection
    codeSelector.value = state.currentExampleIndices[pageData.id];
    
    // Function to render current example
    const renderExample = () => {
        const currentIndex = state.currentExampleIndices[pageData.id];
        const example = pageData.examples[currentIndex];
        
        // Set example title
        exampleTitle.textContent = example.title;
        
        // Set example description
        exampleDescription.innerHTML = formatMarkdown(example.description);
        
        // Set code content
        codeElement.textContent = example.code;
        
        // Set language class if specified
        if (pageData.language) {
            codeElement.className = `language-${pageData.language}`;
        }
    };
    
    // Set up code selector
    codeSelector.addEventListener('change', (e) => {
        state.currentExampleIndices[pageData.id] = parseInt(e.target.value);
        renderExample();
    });
    
    // Set up copy button
    copyButton.addEventListener('click', () => {
        const currentIndex = state.currentExampleIndices[pageData.id];
        const example = pageData.examples[currentIndex];
        
        navigator.clipboard.writeText(example.code)
            .then(() => {
                copyButton.textContent = 'COPIED!';
                setTimeout(() => {
                    copyButton.textContent = 'COPY';
                }, 2000);
            })
            .catch(err => {
                console.error('Failed to copy code:', err);
                alert('Failed to copy code to clipboard');
            });
    });
    
    // Render initial example
    renderExample();
    
    // Add to content container
    elements.contentContainer.appendChild(template);
}

// Render hardware visual
function renderHardwareVisual(pageData) {
    // Clone template
    const template = templates.hardwareVisual.content.cloneNode(true);
    
    // Set page title
    template.querySelector('.page-title').textContent = pageData.title;
    
    // Get panels container
    const panelsContainer = template.querySelector('.panels-container');
    
    // Add panels
    pageData.panels.forEach(panel => {
        // Clone panel template
        const panelTemplate = templates.hardwarePanel.content.cloneNode(true);
        
        // Set panel name
        panelTemplate.querySelector('.panel-name').textContent = panel.name;
        
        // Set panel image
        const panelImage = panelTemplate.querySelector('.panel-image');
        panelImage.src = panel.image;
        panelImage.alt = panel.name;
        
        // Set panel description
        panelTemplate.querySelector('.panel-description').innerHTML = formatMarkdown(panel.description);
        
        // Add interactions
        const interactionsContainer = panelTemplate.querySelector('.panel-interactions');
        
        if (panel.interactions && Array.isArray(panel.interactions)) {
            panel.interactions.forEach(interaction => {
                const button = document.createElement('button');
                button.textContent = interaction.label;
                button.title = interaction.action;
                button.addEventListener('click', () => {
                    alert(`Interaction: ${interaction.action}`);
                });
                interactionsContainer.appendChild(button);
            });
        }
        
        // Add to panels container
        panelsContainer.appendChild(panelTemplate);
    });
    
    // Add to content container
    elements.contentContainer.appendChild(template);
}

// Render bio gallery
function renderBioGallery(pageData) {
    // Clone template
    const template = templates.bioGallery.content.cloneNode(true);
    
    // Set page title
    template.querySelector('.page-title').textContent = pageData.title;
    
    // Get bios container
    const biosContainer = template.querySelector('.bios-container');
    
    // Add bios
    pageData.bios.forEach(bio => {
        // Clone bio card template
        const bioTemplate = templates.bioCard.content.cloneNode(true);
        
        // Set bio name
        bioTemplate.querySelector('.bio-name').textContent = bio.name;
        
        // Set bio role
        bioTemplate.querySelector('.bio-role').textContent = bio.role;
        
        // Set bio image
        const bioImage = bioTemplate.querySelector('.bio-image');
        bioImage.src = bio.image;
        bioImage.alt = bio.name;
        
        // Set bio quote
        bioTemplate.querySelector('.bio-quote').textContent = bio.quote;
        
        // Add to bios container
        biosContainer.appendChild(bioTemplate);
    });
    
    // Add to content container
    elements.contentContainer.appendChild(template);
}

// Render resource list
function renderResourceList(pageData) {
    // Clone template
    const template = templates.resourceList.content.cloneNode(true);
    
    // Set page title
    template.querySelector('.page-title').textContent = pageData.title;
    
    // Get resources container
    const resourcesContainer = template.querySelector('.resources-container');
    
    // Add resources
    pageData.resources.forEach(resource => {
        // Clone resource item template
        const resourceTemplate = templates.resourceItem.content.cloneNode(true);
        
        // Set resource link
        const resourceLink = resourceTemplate.querySelector('.resource-link');
        resourceLink.href = resource.link;
        resourceLink.textContent = resource.title;
        
        // Add to resources container
        resourcesContainer.appendChild(resourceTemplate);
    });
    
    // Add to content container
    elements.contentContainer.appendChild(template);
}

// Render quiz
function renderQuiz(pageData) {
    // Clone template
    const template = templates.quiz.content.cloneNode(true);
    
    // Set page title
    template.querySelector('.page-title').textContent = pageData.title;
    
    // Get quiz elements
    const questionContainer = template.querySelector('.question-container');
    const questionText = template.querySelector('.question-text');
    const optionsContainer = template.querySelector('.options-container');
    const questionCounter = template.querySelector('.question-counter');
    const prevButton = template.querySelector('.prev-question');
    const nextButton = template.querySelector('.next-question');
    const quizResults = template.querySelector('.quiz-results');
    const scoreDisplay = template.querySelector('.score-display');
    const restartButton = template.querySelector('.restart-quiz');
    
    // Initialize quiz state
    if (!state.currentQuestionIndices[pageData.id]) {
        state.currentQuestionIndices[pageData.id] = 0;
    }
    
    if (!state.quizAnswers[pageData.id]) {
        state.quizAnswers[pageData.id] = new Array(pageData.questions.length).fill(null);
    }
    
    // Function to render current question
    const renderQuestion = () => {
        const currentIndex = state.currentQuestionIndices[pageData.id];
        const question = pageData.questions[currentIndex];
        
        // Set question text
        questionText.textContent = question.question;
        
        // Clear options
        optionsContainer.innerHTML = '';
        
        // Add options
        question.options.forEach((option, optionIndex) => {
            // Clone option template
            const optionTemplate = templates.option.content.cloneNode(true);
            const optionButton = optionTemplate.querySelector('.option-button');
            
            // Set option text
            optionButton.textContent = option;
            
            // Check if this option is selected
            if (state.quizAnswers[pageData.id][currentIndex] === option) {
                optionButton.classList.add('selected');
                
                // Check if correct
                if (option === question.answer) {
                    optionButton.classList.add('correct');
                } else {
                    optionButton.classList.add('incorrect');
                }
            }
            
            // Add click handler
            optionButton.addEventListener('click', () => {
                // Save answer
                state.quizAnswers[pageData.id][currentIndex] = option;
                
                // Update UI
                const optionButtons = optionsContainer.querySelectorAll('.option-button');
                optionButtons.forEach(btn => {
                    btn.classList.remove('selected', 'correct', 'incorrect');
                });
                
                optionButton.classList.add('selected');
                
                // Check if correct
                if (option === question.answer) {
                    optionButton.classList.add('correct');
                } else {
                    optionButton.classList.add('incorrect');
                }
            });
            
            // Add to options container
            optionsContainer.appendChild(optionTemplate);
        });
        
        // Update question counter
        questionCounter.textContent = `${currentIndex + 1}/${pageData.questions.length}`;
        
        // Update button states
        prevButton.disabled = currentIndex === 0;
        nextButton.disabled = currentIndex === pageData.questions.length - 1;
        
        // If this is the last question and all questions are answered, show results button
        if (currentIndex === pageData.questions.length - 1) {
            nextButton.textContent = 'RESULTS';
            nextButton.disabled = state.quizAnswers[pageData.id].includes(null);
            
            nextButton.onclick = () => {
                showQuizResults();
            };
        } else {
            nextButton.textContent = '▶';
            nextButton.onclick = null;
        }
    };
    
    // Function to show quiz results
    const showQuizResults = () => {
        // Hide question container
        questionContainer.classList.add('hidden');
        
        // Hide navigation
        prevButton.parentElement.classList.add('hidden');
        
        // Show results
        quizResults.classList.remove('hidden');
        
        // Calculate score
        let correctCount = 0;
        pageData.questions.forEach((question, index) => {
            if (state.quizAnswers[pageData.id][index] === question.answer) {
                correctCount++;
            }
        });
        
        // Display score
        scoreDisplay.textContent = `${correctCount}/${pageData.questions.length}`;
    };
    
    // Set up navigation buttons
    prevButton.addEventListener('click', () => {
        if (state.currentQuestionIndices[pageData.id] > 0) {
            state.currentQuestionIndices[pageData.id]--;
            renderQuestion();
        }
    });
    
    nextButton.addEventListener('click', () => {
        if (state.currentQuestionIndices[pageData.id] < pageData.questions.length - 1) {
            state.currentQuestionIndices[pageData.id]++;
            renderQuestion();
        }
    });
    
    // Set up restart button
    restartButton.addEventListener('click', () => {
        // Reset quiz state
        state.quizAnswers[pageData.id] = new Array(pageData.questions.length).fill(null);
        state.currentQuestionIndices[pageData.id] = 0;
        
        // Show question container
        questionContainer.classList.remove('hidden');
        
        // Show navigation
        prevButton.parentElement.classList.remove('hidden');
        
        // Hide results
        quizResults.classList.add('hidden');
        
        // Render first question
        renderQuestion();
    });
    
    // Render initial question
    renderQuestion();
    
    // Add to content container
    elements.contentContainer.appendChild(template);
}

// Format markdown content
function formatMarkdown(content) {
    if (!content) return '';
    
    // Simple markdown formatting
    let formatted = content
        // Code blocks
        .replace(/```([\s\S]*?)```/g, '<pre><code>$1</code></pre>')
        // Inline code
        .replace(/`([^`]+)`/g, '<code>$1</code>')
        // Bold
        .replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
        // Italic
        .replace(/\*([^*]+)\*/g, '<em>$1</em>')
        // Lists
        .replace(/^- (.*?)$/gm, '<li>$1</li>')
        // Headers
        .replace(/^### (.*?)$/gm, '<h3>$1</h3>')
        .replace(/^## (.*?)$/gm, '<h2>$1</h2>')
        .replace(/^# (.*?)$/gm, '<h1>$1</h1>');
    
    // Wrap lists
    if (formatted.includes('<li>')) {
        formatted = formatted.replace(/<li>[\s\S]*?<\/li>/g, match => {
            return '<ul>' + match + '</ul>';
        });
        
        // Fix nested lists
        formatted = formatted.replace(/<\/ul>\s*<ul>/g, '');
    }
    
    // Convert line breaks to paragraphs
    const paragraphs = formatted.split('\n\n');
    formatted = paragraphs.map(p => {
        // Skip if already wrapped in HTML tag
        if (p.trim().startsWith('<') && !p.trim().startsWith('<li>')) {
            return p;
        }
        return `<p>${p}</p>`;
    }).join('');
    
    return formatted;
}

// Show file interface
function showFileInterface() {
    elements.fileInterface.classList.remove('hidden');
    elements.mainInterface.classList.add('hidden');
}

// Show main interface
function showMainInterface() {
    elements.fileInterface.classList.add('hidden');
    elements.mainInterface.classList.remove('hidden');
}

// Toggle loading overlay
function toggleLoading(show) {
    if (show) {
        elements.loadingOverlay.classList.remove('hidden');
    } else {
        elements.loadingOverlay.classList.add('hidden');
    }
}
