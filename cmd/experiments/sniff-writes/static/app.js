let ws;
let eventCount = 0;
let filteredEventCount = 0;
let allEvents = []; // Store all events for filtering
let maxStoredEvents = 2000; // Limit stored events for performance (configurable)
let showFilteredEvents = false; // Toggle to show filtered out events
let processFilters = []; // Array of process filter objects
let filenameFilters = []; // Array of filename filter objects
let selectedOperations = ['open', 'read', 'write', 'close', 'lseek']; // Selected operations
let displayMode = 'content'; // 'content', 'diff', 'none'
let contentFilter = 'all'; // 'all', 'with-content', 'without-content'
let cliOperations = []; // Operations set via CLI flags

function connectWebSocket() {
	const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
	const wsUrl = protocol + '//' + window.location.host + '/ws';
	
	ws = new WebSocket(wsUrl);
	
	ws.onopen = function() {
		updateConnectionStatus(true);
	};
	
	ws.onmessage = function(event) {
		const eventData = JSON.parse(event.data);
		addEvent(eventData);
	};
	
	ws.onclose = function() {
		updateConnectionStatus(false);
		// Attempt to reconnect after 3 seconds
		setTimeout(connectWebSocket, 3000);
	};
	
	ws.onerror = function() {
		updateConnectionStatus(false);
	};
}

function updateConnectionStatus(connected) {
	const statusIndicator = document.getElementById('connection-status');
	const statusText = document.getElementById('connection-text');
	
	if (connected) {
		statusIndicator.className = 'status-indicator status-connected';
		statusText.textContent = 'Connected';
	} else {
		statusIndicator.className = 'status-indicator status-disconnected';
		statusText.textContent = 'Disconnected';
	}
}

function addEvent(eventData) {
	eventCount++;
	
	// Store event for filtering
	allEvents.push(eventData);
	
	// Limit stored events for performance
	if (allEvents.length > maxStoredEvents) {
		allEvents.shift();
	}
	
	// Re-render the filtered view
	renderFilteredEvents();
}

function renderFilteredEvents() {
	const eventLog = document.getElementById('event-log');
	eventLog.innerHTML = ''; // Clear existing events
	
	filteredEventCount = 0;
	
	// Filter and display events
	for (const eventData of allEvents) {
		const passesFilter = shouldShowEvent(eventData);
		
		if (passesFilter) {
			filteredEventCount++;
		}
		
		// Show event if it passes filter, or if we're showing filtered events and it doesn't pass
		if (passesFilter || showFilteredEvents) {
			const eventItem = createEventElement(eventData, !passesFilter);
			eventLog.appendChild(eventItem);
		}
	}
	
	// Auto-scroll to bottom
	eventLog.scrollTop = eventLog.scrollHeight;
	
	updateStats();
}

function createEventElement(eventData, isFiltered = false) {
	const eventItem = document.createElement('div');
	eventItem.className = isFiltered ? 'event-item text-muted' : 'event-item';
	
	if (isFiltered) {
		eventItem.style.opacity = '0.5';
		eventItem.style.backgroundColor = '#f8f9fa';
	}
	
	const timestamp = new Date(eventData.timestamp).toLocaleTimeString();
	const operationClass = 'operation-' + eventData.operation;
	
	let content = '';
	if (isFiltered) {
		content += '<span class="badge bg-secondary me-2">FILTERED</span>';
	}
	content += '<span class="text-muted">[' + timestamp + ']</span> ';
	content += '<span class="' + operationClass + '"><strong>' + eventData.operation.toUpperCase() + '</strong></span> ';
	content += '<span class="text-primary">' + eventData.process + '</span> ';
	content += '<span class="text-muted">(PID: ' + eventData.pid + ')</span> ';
	
	if (eventData.filename) {
		content += '<span class="text-dark">' + eventData.filename + '</span>';
	}
	
	// Show size information for read/write operations
	if (eventData.write_size > 0) {
		content += ' <span class="text-info">(' + eventData.write_size + ' bytes)</span>';
	}
	
	// Show offset information
	if (eventData.file_offset > 0) {
		content += ' <span class="text-secondary">@' + eventData.file_offset + '</span>';
	}
	
	// Show lseek specific information
	if (eventData.operation === 'lseek' && eventData.whence) {
		content += ' <span class="text-warning">(' + eventData.whence + ')</span>';
	}
	
	// Handle content/diff display based on mode
	if (displayMode === 'content' && eventData.content) {
		// Escape HTML in content for safety
		const escapedContent = eventData.content
			.replace(/&/g, '&amp;')
			.replace(/</g, '&lt;')
			.replace(/>/g, '&gt;')
			.replace(/"/g, '&quot;')
			.replace(/'/g, '&#x27;');
		
		content += '<div class="content-container ms-3 mt-2">';
		content += '<div class="d-flex align-items-center mb-1">';
		content += '<small class="text-muted me-2"><strong>Content:</strong></small>';
		if (eventData.truncated) {
			content += '<span class="badge bg-warning text-dark">TRUNCATED</span>';
		}
		content += '</div>';
		const contentWithLineNumbers = addLineNumbers(eventData.content);
		content += '<div class="content-display-with-lines">' + contentWithLineNumbers + '</div>';
		content += '</div>';
	} else if (displayMode === 'diff') {
		// Handle diff display mode
		if (eventData.diff && eventData.diff.trim()) {
			// Show diff if available
			content += '<div class="diff-container ms-3 mt-2">';
			content += '<div class="d-flex align-items-center mb-1">';
			content += '<small class="text-muted me-2"><strong>Diff:</strong></small>';
			content += '<span class="badge bg-info">CHANGES DETECTED</span>';
			content += '</div>';
			
			// Check if diff is already HTML formatted (contains div tags)
			if (eventData.diff.includes('<div class="diff-')) {
				// Already formatted HTML diff
				content += '<div class="diff-display">' + eventData.diff + '</div>';
			} else {
				// Plain text diff, escape and wrap in pre/code
				const escapedDiff = eventData.diff
					.replace(/&/g, '&amp;')
					.replace(/</g, '&lt;')
					.replace(/>/g, '&gt;')
					.replace(/"/g, '&quot;')
					.replace(/'/g, '&#x27;');
				content += '<pre class="diff-display"><code>' + escapedDiff + '</code></pre>';
			}
			content += '</div>';
		} else if (eventData.operation === 'write' && eventData.content) {
			// Show "no changes" message for write operations with content but no diff
			content += '<div class="diff-container ms-3 mt-2">';
			content += '<div class="d-flex align-items-center mb-1">';
			content += '<small class="text-muted me-2"><strong>Diff:</strong></small>';
			content += '<span class="badge bg-secondary">NO CHANGES DETECTED</span>';
			content += '</div>';
			content += '<div class="text-muted fst-italic">Content identical to previous read</div>';
			content += '</div>';
		}
	}
	// If displayMode is 'none', we don't show content or diff
	
	eventItem.innerHTML = content;
	return eventItem;
}

function shouldShowEvent(eventData) {
	// Operation filter (multi-select)
	if (!selectedOperations.includes(eventData.operation)) {
		return false;
	}
	
	// Process filters (with negative syntax support)
	for (const filter of processFilters) {
		const processName = eventData.process.toLowerCase();
		const filterText = filter.text.toLowerCase();
		
		if (filter.negative) {
			// Negative filter: exclude if matches
			if (processName.includes(filterText)) {
				return false;
			}
		} else {
			// Positive filter: must match at least one
			if (!processName.includes(filterText)) {
				return false;
			}
		}
	}
	
	// If there are positive process filters, at least one must match
	const positiveProcessFilters = processFilters.filter(f => !f.negative);
	if (positiveProcessFilters.length > 0) {
		const processName = eventData.process.toLowerCase();
		const hasMatch = positiveProcessFilters.some(filter => 
			processName.includes(filter.text.toLowerCase())
		);
		if (!hasMatch) {
			return false;
		}
	}
	
	// Filename filters (with negative syntax support)
	if (eventData.filename) {
		for (const filter of filenameFilters) {
			const filename = eventData.filename.toLowerCase();
			const filterText = filter.text.toLowerCase();
			
			if (filter.negative) {
				// Negative filter: exclude if matches
				if (filename.includes(filterText)) {
					return false;
				}
			} else {
				// Positive filter: must match at least one
				if (!filename.includes(filterText)) {
					return false;
				}
			}
		}
		
		// If there are positive filename filters, at least one must match
		const positiveFilenameFilters = filenameFilters.filter(f => !f.negative);
		if (positiveFilenameFilters.length > 0) {
			const filename = eventData.filename.toLowerCase();
			const hasMatch = positiveFilenameFilters.some(filter => 
				filename.includes(filter.text.toLowerCase())
			);
			if (!hasMatch) {
				return false;
			}
		}
	}
	
	// Content filter
	if (contentFilter !== 'all') {
		const hasContent = eventData.content && eventData.content.length > 0;
		const hasDiff = eventData.diff && eventData.diff.length > 0;
		const hasDisplayableContent = hasContent || hasDiff;
		
		if (contentFilter === 'with-content' && !hasDisplayableContent) {
			return false;
		}
		if (contentFilter === 'without-content' && hasDisplayableContent) {
			return false;
		}
	}
	
	return true;
}

function updateStats() {
	const filteredOutCount = allEvents.length - filteredEventCount;
	document.getElementById('total-events').textContent = eventCount;
	document.getElementById('filtered-events').textContent = filteredEventCount;
	document.getElementById('filtered-out-events').textContent = filteredOutCount;
}

function clearEvents() {
	allEvents = [];
	eventCount = 0;
	filteredEventCount = 0;
	document.getElementById('event-log').innerHTML = '';
	updateStats();
}

function clearAllFilters() {
	processFilters = [];
	filenameFilters = [];
	selectedOperations = ['open', 'read', 'write', 'close', 'lseek'];
	showFilteredEvents = false;
	displayMode = 'content';
	contentFilter = 'all';
	
	renderFilterPills();
	updateOperationCheckboxes();
	updateContentFilterButtons();
	
	// Reset show filtered button
	const btn = document.getElementById('show-filtered-btn');
	btn.textContent = 'Show Filtered';
	btn.classList.remove('btn-info');
	btn.classList.add('btn-outline-info');
	
	// Reset display mode buttons
	setDisplayMode('content');
	
	saveSettings();
	renderFilteredEvents();
}

function applyFilters() {
	// Re-render all events with current filters
	renderFilteredEvents();
}

function toggleShowFiltered() {
	showFilteredEvents = !showFilteredEvents;
	const btn = document.getElementById('show-filtered-btn');
	
	if (showFilteredEvents) {
		btn.textContent = 'Hide Filtered';
		btn.classList.remove('btn-outline-info');
		btn.classList.add('btn-info');
	} else {
		btn.textContent = 'Show Filtered';
		btn.classList.remove('btn-info');
		btn.classList.add('btn-outline-info');
	}
	
	renderFilteredEvents();
}

function setDisplayMode(mode) {
	displayMode = mode;
	
	// Update button states
	const buttons = ['content', 'diff', 'none'];
	buttons.forEach(btnMode => {
		const btn = document.getElementById(`display-mode-${btnMode}`);
		if (btn) {
			if (btnMode === mode) {
				btn.classList.remove('btn-outline-secondary');
				btn.classList.add('btn-primary');
			} else {
				btn.classList.remove('btn-primary');
				btn.classList.add('btn-outline-secondary');
			}
		}
	});
	
	saveSettings();
	renderFilteredEvents();
}

function changeContentFilter(filter) {
	contentFilter = filter;
	updateContentFilterButtons();
	saveSettings();
	renderFilteredEvents();
}

function updateContentFilterButtons() {
	['all', 'with-content', 'without-content'].forEach(filter => {
		const btn = document.getElementById('content-filter-' + filter);
		if (btn) {
			if (contentFilter === filter) {
				btn.classList.remove('btn-outline-secondary');
				btn.classList.add('btn-secondary');
			} else {
				btn.classList.remove('btn-secondary');
				btn.classList.add('btn-outline-secondary');
			}
		}
	});
}

function changeMemoryLimit() {
	const select = document.getElementById('memory-limit');
	maxStoredEvents = parseInt(select.value);
	
	// Trim stored events if necessary
	if (allEvents.length > maxStoredEvents) {
		allEvents = allEvents.slice(allEvents.length - maxStoredEvents);
	}
	
	saveSettings();
	renderFilteredEvents();
}

// Filter management functions
function addProcessFilter(text) {
	if (!text) return;
	
	const negative = text.startsWith('!');
	const filterText = negative ? text.slice(1) : text;
	
	if (filterText) {
		processFilters.push({ text: filterText, negative });
		renderFilterPills();
		saveSettings();
		renderFilteredEvents();
	}
}

function addFilenameFilter(text) {
	if (!text) return;
	
	const negative = text.startsWith('!');
	const filterText = negative ? text.slice(1) : text;
	
	if (filterText) {
		filenameFilters.push({ text: filterText, negative });
		renderFilterPills();
		saveSettings();
		renderFilteredEvents();
	}
}

function removeProcessFilter(index) {
	processFilters.splice(index, 1);
	renderFilterPills();
	saveSettings();
	renderFilteredEvents();
}

function removeFilenameFilter(index) {
	filenameFilters.splice(index, 1);
	renderFilterPills();
	saveSettings();
	renderFilteredEvents();
}

function toggleOperation(operation) {
	const index = selectedOperations.indexOf(operation);
	if (index > -1) {
		selectedOperations.splice(index, 1);
	} else {
		selectedOperations.push(operation);
	}
	updateOperationCheckboxes();
	saveSettings();
	renderFilteredEvents();
}

function updateOperationCheckboxes() {
	['open', 'read', 'write', 'close', 'lseek'].forEach(op => {
		const checkbox = document.getElementById('op-' + op);
		if (checkbox) {
			checkbox.checked = selectedOperations.includes(op);
		}
	});
}

function renderFilterPills() {
	const processContainer = document.getElementById('process-pills');
	const filenameContainer = document.getElementById('filename-pills');
	
	// Render process filter pills
	processContainer.innerHTML = '';
	processFilters.forEach((filter, index) => {
		const pill = document.createElement('span');
		pill.className = 'filter-pill ' + (filter.negative ? 'negative' : '');
		pill.innerHTML = (filter.negative ? '!' : '') + filter.text + 
			'<button class="remove-pill" onclick="removeProcessFilter(' + index + ')">&times;</button>';
		processContainer.appendChild(pill);
	});
	
	// Add "add filter" button
	const addBtn = document.createElement('button');
	addBtn.className = 'add-filter-btn';
	addBtn.textContent = '+ Add process filter';
	addBtn.onclick = () => {
		const text = prompt('Enter process filter (prefix with ! for negative filter):');
		if (text) addProcessFilter(text.trim());
	};
	processContainer.appendChild(addBtn);
	
	// Render filename filter pills
	filenameContainer.innerHTML = '';
	filenameFilters.forEach((filter, index) => {
		const pill = document.createElement('span');
		pill.className = 'filter-pill ' + (filter.negative ? 'negative' : '');
		pill.innerHTML = (filter.negative ? '!' : '') + filter.text + 
			'<button class="remove-pill" onclick="removeFilenameFilter(' + index + ')">&times;</button>';
		filenameContainer.appendChild(pill);
	});
	
	// Add "add filter" button
	const addBtn2 = document.createElement('button');
	addBtn2.className = 'add-filter-btn';
	addBtn2.textContent = '+ Add filename filter';
	addBtn2.onclick = () => {
		const text = prompt('Enter filename filter (prefix with ! for negative filter):');
		if (text) addFilenameFilter(text.trim());
	};
	filenameContainer.appendChild(addBtn2);
}

// localStorage functions
function saveSettings() {
	const settings = {
		processFilters,
		filenameFilters,
		selectedOperations,
		showFilteredEvents,
		displayMode,
		contentFilter,
		maxStoredEvents
	};
	localStorage.setItem('sniff-writes-settings', JSON.stringify(settings));
}

function loadSettings() {
	try {
		const saved = localStorage.getItem('sniff-writes-settings');
		if (saved) {
			const settings = JSON.parse(saved);
			processFilters = settings.processFilters || [];
			filenameFilters = settings.filenameFilters || [];
			selectedOperations = settings.selectedOperations || ['open', 'read', 'write', 'close', 'lseek'];
			showFilteredEvents = settings.showFilteredEvents || false;
			displayMode = settings.displayMode || 'content';
			contentFilter = settings.contentFilter || 'all';
			maxStoredEvents = settings.maxStoredEvents || 2000;
			
			renderFilterPills();
			updateOperationCheckboxes();
			updateContentFilterButtons();
			setDisplayMode(displayMode);
			
			// Update memory limit dropdown
			const memorySelect = document.getElementById('memory-limit');
			if (memorySelect) {
				memorySelect.value = maxStoredEvents.toString();
			}
			
			// Update show filtered button
			const btn = document.getElementById('show-filtered-btn');
			if (showFilteredEvents) {
				btn.textContent = 'Hide Filtered';
				btn.classList.remove('btn-outline-info');
				btn.classList.add('btn-info');
			}
			
			// Update content toggle button
			const contentBtn = document.getElementById('toggle-content-btn');
			if (contentBtn) {
				if (showContent) {
					contentBtn.textContent = 'Hide Content';
					contentBtn.classList.remove('btn-outline-primary');
					contentBtn.classList.add('btn-primary');
				} else {
					contentBtn.textContent = 'Show Content';
					contentBtn.classList.remove('btn-primary');
					contentBtn.classList.add('btn-outline-primary');
				}
			}
		}
	} catch (e) {
		console.warn('Failed to load settings from localStorage:', e);
	}
}

// History search state
let currentPage = 1;
let currentSearchParams = {};
let systemStatus = null;

// Initialize WebSocket connection when page loads
document.addEventListener('DOMContentLoaded', function() {
	loadSettings();
	connectWebSocket();
	updateStats();
	checkSystemStatus();
	
	// Add tab switching event listeners
	document.getElementById('live-tab').addEventListener('click', function() {
		showLiveView();
	});
	
	document.getElementById('history-tab').addEventListener('click', function() {
		showHistoryView();
	});
});

function showLiveView() {
	document.getElementById('live-stats').style.display = 'block';
	document.getElementById('live-log').style.display = 'block';
}

function showHistoryView() {
	document.getElementById('live-stats').style.display = 'none';
	document.getElementById('live-log').style.display = 'none';
}

function checkSystemStatus() {
	fetch('/api/status')
		.then(response => response.json())
		.then(data => {
			systemStatus = data;
			updateHistorySearchUI();
			updateCLIOperationsDisplay();
		})
		.catch(error => {
			console.error('Error checking system status:', error);
			showSystemStatusError();
		});
}

function updateCLIOperationsDisplay() {
	const cliOpsElement = document.getElementById('cli-operations');
	if (systemStatus && systemStatus.operations) {
		cliOperations = systemStatus.operations;
		cliOpsElement.textContent = cliOperations.join(', ');
		cliOpsElement.className = 'text-info';
	} else {
		cliOpsElement.textContent = 'Unknown';
		cliOpsElement.className = 'text-muted';
	}
}

function updateHistorySearchUI() {
	const historySection = document.querySelector('.history-section');
	const searchForm = document.getElementById('history-form');
	const exportSection = document.querySelector('.card:has(.btn-group)');
	
	if (!systemStatus.can_search) {
		// Database not available - show warning and disable form
		const warningDiv = document.createElement('div');
		warningDiv.className = 'alert alert-warning';
		warningDiv.innerHTML = `
			<h6><strong>History Search Unavailable</strong></h6>
			<p>${systemStatus.message}</p>
			<small class="text-muted">
				<strong>Database Status:</strong> ${systemStatus.database_status}
				${systemStatus.database_error ? `<br><strong>Error:</strong> ${systemStatus.database_error}` : ''}
			</small>
		`;
		
		// Insert warning at the top of history section
		historySection.insertBefore(warningDiv, searchForm);
		
		// Disable form elements
		const formElements = searchForm.querySelectorAll('input, button, select');
		formElements.forEach(element => {
			element.disabled = true;
		});
		
		// Disable export buttons
		if (exportSection) {
			const exportButtons = exportSection.querySelectorAll('button');
			exportButtons.forEach(button => {
				button.disabled = true;
			});
		}
	}
}

function showSystemStatusError() {
	const historySection = document.querySelector('.history-section');
	const warningDiv = document.createElement('div');
	warningDiv.className = 'alert alert-danger';
	warningDiv.innerHTML = `
		<h6><strong>System Status Check Failed</strong></h6>
		<p>Unable to determine system status. History search functionality may not work properly.</p>
	`;
	historySection.insertBefore(warningDiv, document.getElementById('history-form'));
}

function displaySearchError(error) {
	const resultsContainer = document.getElementById('history-results');
	const countBadge = document.getElementById('history-count');
	
	countBadge.textContent = 'Error';
	
	let errorHtml = '<div class="alert alert-danger m-3">';
	
	try {
		// Try to parse structured error response
		const errorData = JSON.parse(error.message);
		errorHtml += `
			<h6><strong>Search Error: ${errorData.error || 'Unknown'}</strong></h6>
			<p>${errorData.message || 'An error occurred while searching.'}</p>
		`;
		
		if (errorData.details) {
			errorHtml += '<div class="mt-2">';
			errorHtml += '<strong>Details:</strong><br>';
			if (typeof errorData.details === 'object') {
				for (const [key, value] of Object.entries(errorData.details)) {
					errorHtml += `<small class="text-muted">${key}: ${value}</small><br>`;
				}
			} else {
				errorHtml += `<small class="text-muted">${errorData.details}</small>`;
			}
			errorHtml += '</div>';
		}
	} catch (parseError) {
		// Fallback for non-structured errors
		errorHtml += `
			<h6><strong>Search Error</strong></h6>
			<p>An error occurred while searching: ${error.message}</p>
		`;
	}
	
	errorHtml += '</div>';
	resultsContainer.innerHTML = errorHtml;
}

function searchHistory(page = 1) {
	currentPage = page;
	
	// Build search parameters
	const params = new URLSearchParams();
	
	const startTime = document.getElementById('start-time').value;
	const endTime = document.getElementById('end-time').value;
	const process = document.getElementById('history-process').value;
	const filename = document.getElementById('history-filename').value;
	const pid = document.getElementById('history-pid').value;
	const limit = document.getElementById('history-limit').value;
	
	if (startTime) params.append('start_time', new Date(startTime).toISOString());
	if (endTime) params.append('end_time', new Date(endTime).toISOString());
	if (process) params.append('process', process);
	if (filename) params.append('filename', filename);
	if (pid) params.append('pid', pid);
	
	// Get selected operations
	const operations = [];
	if (document.getElementById('hist-op-open').checked) operations.push('open');
	if (document.getElementById('hist-op-read').checked) operations.push('read');
	if (document.getElementById('hist-op-write').checked) operations.push('write');
	if (document.getElementById('hist-op-close').checked) operations.push('close');
	
	if (operations.length > 0) {
		params.append('operations', operations.join(','));
	}
	
	params.append('limit', limit);
	params.append('offset', (page - 1) * parseInt(limit));
	
	// Store current search params for pagination
	currentSearchParams = Object.fromEntries(params);
	
	// Make API request
	fetch('/api/events?' + params.toString())
		.then(response => {
			if (!response.ok) {
				return response.json().then(errorData => {
					throw new Error(JSON.stringify(errorData));
				});
			}
			return response.json();
		})
		.then(data => {
			displayHistoryResults(data);
			updatePagination(data);
		})
		.catch(error => {
			console.error('Error searching history:', error);
			displaySearchError(error);
		});
}

function displayHistoryResults(data) {
	const resultsContainer = document.getElementById('history-results');
	const countBadge = document.getElementById('history-count');
	
	countBadge.textContent = `${data.events.length} of ${data.total} results`;
	
	if (data.events.length === 0) {
		resultsContainer.innerHTML = '<div class="text-center p-4 text-muted">No events found for the given criteria</div>';
		return;
	}
	
	resultsContainer.innerHTML = '';
	
	data.events.forEach(event => {
		const eventElement = createHistoryEventElement(event);
		resultsContainer.appendChild(eventElement);
	});
}

function updatePagination(data) {
	const paginationContainer = document.getElementById('pagination');
	const paginationRow = document.getElementById('pagination-row');
	
	if (data.total_pages <= 1) {
		paginationRow.style.display = 'none';
		return;
	}
	
	paginationRow.style.display = 'block';
	paginationContainer.innerHTML = '';
	
	// Previous button
	const prevLi = document.createElement('li');
	prevLi.className = `page-item ${currentPage === 1 ? 'disabled' : ''}`;
	prevLi.innerHTML = `<a class="page-link" href="#" onclick="searchHistory(${currentPage - 1})">Previous</a>`;
	paginationContainer.appendChild(prevLi);
	
	// Page numbers
	const startPage = Math.max(1, currentPage - 2);
	const endPage = Math.min(data.total_pages, currentPage + 2);
	
	for (let i = startPage; i <= endPage; i++) {
		const li = document.createElement('li');
		li.className = `page-item ${i === currentPage ? 'active' : ''}`;
		li.innerHTML = `<a class="page-link" href="#" onclick="searchHistory(${i})">${i}</a>`;
		paginationContainer.appendChild(li);
	}
	
	// Next button
	const nextLi = document.createElement('li');
	nextLi.className = `page-item ${currentPage === data.total_pages ? 'disabled' : ''}`;
	nextLi.innerHTML = `<a class="page-link" href="#" onclick="searchHistory(${currentPage + 1})">Next</a>`;
	paginationContainer.appendChild(nextLi);
}

function exportData(format) {
	// Check if search is available
	if (systemStatus && !systemStatus.can_search) {
		alert('Export is not available because the database is not connected.');
		return;
	}
	
	// Build export URL with current search parameters
	const params = new URLSearchParams(currentSearchParams);
	params.set('format', format);
	params.delete('limit');  // Remove pagination for export
	params.delete('offset');
	
	// Test the export URL first to handle errors gracefully
	const url = '/api/events/export?' + params.toString();
	
	fetch(url, { method: 'HEAD' })
		.then(response => {
			if (!response.ok) {
				return response.json().then(errorData => {
					throw new Error(JSON.stringify(errorData));
				});
			}
			
			// If HEAD request succeeds, proceed with download
			const link = document.createElement('a');
			link.href = url;
			link.download = `file_events.${format}`;
			document.body.appendChild(link);
			link.click();
			document.body.removeChild(link);
		})
		.catch(error => {
			try {
				const errorData = JSON.parse(error.message);
				alert(`Export failed: ${errorData.message}\n\nDetails: ${errorData.details?.error_details || 'Unknown error'}`);
			} catch (parseError) {
				alert(`Export failed: ${error.message}`);
			}
		});
}

function createHistoryEventElement(event) {
	const div = document.createElement('div');
	div.className = 'event-item';
	
	const timestamp = new Date(event.timestamp).toLocaleString();
	const processInfo = `${event.process} (PID: ${event.pid})`;
	const operationClass = `operation-${event.operation}`;
	
	let contentDisplay = '';
	if (event.content && event.content.trim()) {
		const truncated = event.truncated ? ' (truncated)' : '';
		const contentWithLineNumbers = addLineNumbers(event.content);
		contentDisplay = `
			<div class="event-content">
				<strong>Content${truncated}:</strong>
				<div class="content-display-with-lines">${contentWithLineNumbers}</div>
			</div>
		`;
	}
	
	let diffDisplay = '';
	if (event.diff && event.diff.trim()) {
		// Check if diff is already HTML formatted
		if (event.diff.includes('<div class="diff-')) {
			diffDisplay = `
				<div class="event-diff">
					<strong>Diff:</strong>
					<div class="diff-display">${event.diff}</div>
				</div>
			`;
		} else {
			diffDisplay = `
				<div class="event-diff">
					<strong>Diff:</strong>
					<pre class="diff-display"><code>${escapeHtml(event.diff)}</code></pre>
				</div>
			`;
		}
	}
	
	let offsetInfo = '';
	if (event.file_offset > 0) {
		offsetInfo = `<div class="file-offset">Offset: ${event.file_offset}</div>`;
	}
	
	let whenceInfo = '';
	if (event.operation === 'lseek' && event.whence) {
		whenceInfo = `<div class="whence-info">Mode: ${event.whence}</div>`;
	}
	
	div.innerHTML = `
		<div class="event-header">
			<span class="timestamp">${timestamp}</span>
			<span class="operation ${operationClass}">${event.operation.toUpperCase()}</span>
			<span class="process">${escapeHtml(processInfo)}</span>
		</div>
		<div class="event-details">
			<div class="filename">${escapeHtml(event.filename || 'N/A')}</div>
			${event.write_size > 0 ? `<div class="write-size">Size: ${event.write_size} bytes</div>` : ''}
			${offsetInfo}
			${whenceInfo}
			${contentDisplay}
			${diffDisplay}
		</div>
	`;
	
	return div;
}

function escapeHtml(text) {
	const div = document.createElement('div');
	div.textContent = text;
	return div.innerHTML;
}

function addLineNumbers(content) {
	const lines = content.split('\n');
	const maxLineNum = lines.length;
	const maxDigits = maxLineNum.toString().length;
	
	return lines.map((line, index) => {
		const lineNum = (index + 1).toString().padStart(maxDigits, ' ');
		const escapedLine = escapeHtml(line);
		return `<div class="content-line"><span class="line-number">${lineNum}</span><span class="line-content">${escapedLine}</span></div>`;
	}).join('');
}