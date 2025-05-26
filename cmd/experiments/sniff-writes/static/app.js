let ws;
let eventCount = 0;
let filteredEventCount = 0;
let allEvents = []; // Store all events for filtering
const maxStoredEvents = 2000; // Limit stored events for performance
let showFilteredEvents = false; // Toggle to show filtered out events
let processFilters = []; // Array of process filter objects
let filenameFilters = []; // Array of filename filter objects
let selectedOperations = ['open', 'read', 'write', 'close']; // Selected operations
let showContent = true; // Show content in events
let contentFilter = 'all'; // 'all', 'with-content', 'without-content'

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
	if (eventData.write_size > 0) {
		content += ' <span class="text-info">(' + eventData.write_size + ' bytes)</span>';
	}
	if (eventData.content && showContent) {
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
		content += '<div class="content-display">' + escapedContent + '</div>';
		content += '</div>';
	}
	
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
		if (contentFilter === 'with-content' && !hasContent) {
			return false;
		}
		if (contentFilter === 'without-content' && hasContent) {
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
	selectedOperations = ['open', 'read', 'write', 'close'];
	showFilteredEvents = false;
	showContent = true;
	contentFilter = 'all';
	
	renderFilterPills();
	updateOperationCheckboxes();
	updateContentFilterButtons();
	
	// Reset show filtered button
	const btn = document.getElementById('show-filtered-btn');
	btn.textContent = 'Show Filtered';
	btn.classList.remove('btn-info');
	btn.classList.add('btn-outline-info');
	
	// Reset content toggle button
	const contentBtn = document.getElementById('toggle-content-btn');
	if (contentBtn) {
		contentBtn.textContent = 'Hide Content';
		contentBtn.classList.remove('btn-outline-primary');
		contentBtn.classList.add('btn-primary');
	}
	
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

function toggleContentDisplay() {
	showContent = !showContent;
	const btn = document.getElementById('toggle-content-btn');
	
	if (showContent) {
		btn.textContent = 'Hide Content';
		btn.classList.remove('btn-outline-primary');
		btn.classList.add('btn-primary');
	} else {
		btn.textContent = 'Show Content';
		btn.classList.remove('btn-primary');
		btn.classList.add('btn-outline-primary');
	}
	
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
	['open', 'read', 'write', 'close'].forEach(op => {
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
		showContent,
		contentFilter
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
			selectedOperations = settings.selectedOperations || ['open', 'read', 'write', 'close'];
			showFilteredEvents = settings.showFilteredEvents || false;
			showContent = settings.showContent !== undefined ? settings.showContent : true;
			contentFilter = settings.contentFilter || 'all';
			
			renderFilterPills();
			updateOperationCheckboxes();
			updateContentFilterButtons();
			
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

// Initialize WebSocket connection when page loads
document.addEventListener('DOMContentLoaded', function() {
	loadSettings();
	connectWebSocket();
	updateStats();
});