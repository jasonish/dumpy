// Simple date formatting function
function formatDate(date) {
    const pad = (n) => n.toString().padStart(2, '0');
    const year = date.getFullYear();
    const month = pad(date.getMonth() + 1);
    const day = pad(date.getDate());
    const hours = pad(date.getHours());
    const minutes = pad(date.getMinutes());
    const seconds = pad(date.getSeconds());
    
    // Get timezone offset
    const offset = -date.getTimezoneOffset();
    const offsetHours = Math.floor(Math.abs(offset) / 60);
    const offsetMinutes = Math.abs(offset) % 60;
    const offsetSign = offset >= 0 ? '+' : '-';
    const offsetString = `${offsetSign}${pad(offsetHours)}:${pad(offsetMinutes)}`;
    
    return `${year}-${month}-${day}T${hours}:${minutes}:${seconds}${offsetString}`;
}

// Get timezone offset in format like "+05:00"
function getTimezoneOffset() {
    const offset = -new Date().getTimezoneOffset();
    const hours = Math.floor(Math.abs(offset) / 60);
    const minutes = Math.abs(offset) % 60;
    const sign = offset >= 0 ? '+' : '-';
    return `${sign}${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}`;
}

// Calculate time relative to now
function getRelativeTime(unit, value) {
    const date = new Date();
    switch(unit) {
        case 'minute':
            date.setMinutes(date.getMinutes() - value);
            break;
        case 'hour':
            date.setHours(date.getHours() - value);
            break;
    }
    return date;
}

// Set start time based on selection
function setStartTime(unit, value) {
    const input = document.getElementById('start-time');
    if (unit === null) {
        input.value = formatDate(new Date());
    } else {
        input.value = formatDate(getRelativeTime(unit, value));
    }
}

// Set value of an input element
function setElementValue(elementId, value) {
    document.getElementById(elementId).value = value;
}

// Initialize the application
document.addEventListener('DOMContentLoaded', function() {
    // Set timezone offset
    document.getElementById('timezone-offset').value = getTimezoneOffset();
    
    // Set initial start time
    document.getElementById('start-time').value = formatDate(new Date());
    
    // Check URL parameters for event
    const params = new URLSearchParams(window.location.search);
    if (params.get('event')) {
        document.getElementById('event-input').value = params.get('event');
        // Switch to event tab
        const eventTab = document.querySelector('[data-tab="event"]');
        eventTab.click();
    }
    
    // Tab handling
    const tabs = document.querySelectorAll('.tab');
    const tabContents = document.querySelectorAll('.tab-content');
    
    tabs.forEach(tab => {
        tab.addEventListener('click', function() {
            // Remove active class from all tabs and contents
            tabs.forEach(t => t.classList.remove('active'));
            tabContents.forEach(tc => tc.classList.remove('active'));
            
            // Add active class to clicked tab
            this.classList.add('active');
            
            // Show corresponding content
            const tabName = this.getAttribute('data-tab');
            const content = document.querySelector(`[data-content="${tabName}"]`);
            if (content) {
                content.classList.add('active');
            }
            
            // Update query type
            document.getElementById('query-type').value = this.getAttribute('data-query-type');
        });
    });
    
    // Dropdown handling
    const dropdownToggles = document.querySelectorAll('.dropdown-toggle');
    
    dropdownToggles.forEach(toggle => {
        toggle.addEventListener('click', function(e) {
            e.preventDefault();
            const menu = this.nextElementSibling;
            
            // Close all other dropdowns
            document.querySelectorAll('.dropdown-menu').forEach(m => {
                if (m !== menu) m.classList.remove('show');
            });
            
            // Toggle this dropdown
            menu.classList.toggle('show');
        });
    });
    
    // Close dropdowns when clicking outside
    document.addEventListener('click', function(e) {
        if (!e.target.closest('.input-group')) {
            document.querySelectorAll('.dropdown-menu').forEach(menu => {
                menu.classList.remove('show');
            });
        }
    });
    
    // Handle start time dropdown items
    document.querySelectorAll('[data-time]').forEach(item => {
        item.addEventListener('click', function(e) {
            e.preventDefault();
            const time = this.getAttribute('data-time');
            
            if (time === 'now') {
                setStartTime(null, null);
            } else if (time === '1m') {
                setStartTime('minute', 1);
            } else if (time === '5m') {
                setStartTime('minute', 5);
            } else if (time === '15m') {
                setStartTime('minute', 15);
            } else if (time === '1h') {
                setStartTime('hour', 1);
            }
            
            // Close dropdown
            this.closest('.dropdown-menu').classList.remove('show');
        });
    });
    
    // Handle duration dropdown items
    document.querySelectorAll('[data-duration]').forEach(item => {
        item.addEventListener('click', function(e) {
            e.preventDefault();
            const duration = this.getAttribute('data-duration');
            const inputGroup = this.closest('.input-group');
            const input = inputGroup.querySelector('.input');
            
            input.value = duration;
            
            // Close dropdown
            this.closest('.dropdown-menu').classList.remove('show');
        });
    });
    
    // Load spools from server
    fetch('/api/spools')
        .then(response => response.json())
        .then(spools => {
            const selector = document.getElementById('spool');
            spools.forEach(spool => {
                const option = document.createElement('option');
                option.value = spool;
                option.textContent = spool;
                selector.appendChild(option);
            });
        })
        .catch(error => {
            console.error('Failed to load spools:', error);
        });
});