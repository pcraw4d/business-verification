/**
 * Session Manager Component
 * Manages single merchant session with state persistence and switching capabilities
 * Ensures only one merchant is active at a time with proper overview reset
 */
class SessionManager {
    constructor(options = {}) {
        this.container = options.container || document.body;
        this.apiBaseUrl = options.apiBaseUrl || '/api/v1';
        this.storageKey = options.storageKey || 'merchant_session';
        this.sessionTimeout = options.sessionTimeout || 30 * 60 * 1000; // 30 minutes
        this.currentSession = null;
        this.sessionHistory = [];
        this.maxHistorySize = options.maxHistorySize || 10;
        this.isSessionActive = false;
        this.sessionTimer = null;
        
        // Event callbacks
        this.onSessionStart = options.onSessionStart || null;
        this.onSessionEnd = options.onSessionEnd || null;
        this.onSessionSwitch = options.onSessionSwitch || null;
        this.onSessionTimeout = options.onSessionTimeout || null;
        this.onOverviewReset = options.onOverviewReset || null;
        
        this.init();
    }

    init() {
        this.loadSessionFromStorage();
        this.createSessionInterface();
        this.bindEvents();
        this.startSessionMonitoring();
    }

    createSessionInterface() {
        const sessionHTML = `
            <div class="session-manager-container" id="sessionManagerContainer">
                <div class="session-header">
                    <div class="session-info">
                        <h3 class="session-title">
                            <i class="fas fa-user-circle"></i>
                            Active Session
                        </h3>
                        <div class="session-status" id="sessionStatus">
                            <span class="status-indicator" id="statusIndicator"></span>
                            <span class="status-text" id="statusText">No active session</span>
                        </div>
                    </div>
                    <div class="session-actions">
                        <button class="btn btn-outline btn-sm" id="sessionHistoryBtn" disabled>
                            <i class="fas fa-history"></i>
                            History
                        </button>
                        <button class="btn btn-danger btn-sm" id="endSessionBtn" disabled>
                            <i class="fas fa-sign-out-alt"></i>
                            End Session
                        </button>
                    </div>
                </div>

                <div class="session-details" id="sessionDetails" style="display: none;">
                    <div class="merchant-info">
                        <div class="merchant-avatar" id="merchantAvatar">
                            <i class="fas fa-building"></i>
                        </div>
                        <div class="merchant-details">
                            <div class="merchant-name" id="merchantName">-</div>
                            <div class="merchant-meta">
                                <span class="merchant-id" id="merchantId">-</span>
                                <span class="session-duration" id="sessionDuration">-</span>
                            </div>
                        </div>
                    </div>
                    <div class="session-timer">
                        <div class="timer-label">Session Timeout</div>
                        <div class="timer-progress">
                            <div class="timer-bar" id="timerBar"></div>
                        </div>
                        <div class="timer-text" id="timerText">30:00</div>
                    </div>
                </div>

                <div class="session-history" id="sessionHistory" style="display: none;">
                    <div class="history-header">
                        <h4>Recent Sessions</h4>
                        <button class="btn btn-outline btn-sm" id="closeHistoryBtn">
                            <i class="fas fa-times"></i>
                        </button>
                    </div>
                    <div class="history-list" id="historyList">
                        <div class="no-history">
                            <i class="fas fa-history"></i>
                            <p>No recent sessions</p>
                        </div>
                    </div>
                </div>

                <div class="session-switch-modal" id="sessionSwitchModal" style="display: none;">
                    <div class="modal-overlay">
                        <div class="modal-content">
                            <div class="modal-header">
                                <h4>Switch Merchant Session</h4>
                                <button class="btn btn-outline btn-sm" id="closeSwitchModalBtn">
                                    <i class="fas fa-times"></i>
                                </button>
                            </div>
                            <div class="modal-body">
                                <div class="current-session-warning">
                                    <i class="fas fa-exclamation-triangle"></i>
                                    <p>You are about to switch from the current merchant session. This will reset the overview and start a new session.</p>
                                </div>
                                <div class="new-merchant-info" id="newMerchantInfo">
                                    <!-- New merchant details will be populated here -->
                                </div>
                            </div>
                            <div class="modal-footer">
                                <button class="btn btn-secondary" id="cancelSwitchBtn">Cancel</button>
                                <button class="btn btn-primary" id="confirmSwitchBtn">Switch Session</button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        `;

        this.container.innerHTML = sessionHTML;
        this.addStyles();
    }

    addStyles() {
        const styles = `
            <style>
                .session-manager-container {
                    background: white;
                    border-radius: 12px;
                    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
                    padding: 20px;
                    margin-bottom: 20px;
                    border-left: 4px solid #3498db;
                }

                .session-header {
                    display: flex;
                    justify-content: space-between;
                    align-items: center;
                    margin-bottom: 16px;
                }

                .session-info {
                    display: flex;
                    flex-direction: column;
                    gap: 8px;
                }

                .session-title {
                    display: flex;
                    align-items: center;
                    gap: 8px;
                    margin: 0;
                    color: #2c3e50;
                    font-size: 1.2rem;
                    font-weight: 600;
                }

                .session-title i {
                    color: #3498db;
                }

                .session-status {
                    display: flex;
                    align-items: center;
                    gap: 8px;
                }

                .status-indicator {
                    width: 12px;
                    height: 12px;
                    border-radius: 50%;
                    background: #95a5a6;
                    transition: all 0.3s ease;
                }

                .status-indicator.active {
                    background: #27ae60;
                    box-shadow: 0 0 8px rgba(39, 174, 96, 0.4);
                }

                .status-indicator.warning {
                    background: #f39c12;
                    box-shadow: 0 0 8px rgba(243, 156, 18, 0.4);
                }

                .status-indicator.danger {
                    background: #e74c3c;
                    box-shadow: 0 0 8px rgba(231, 76, 60, 0.4);
                }

                .status-text {
                    font-size: 0.9rem;
                    color: #6c757d;
                    font-weight: 500;
                }

                .session-actions {
                    display: flex;
                    gap: 8px;
                }

                .btn {
                    padding: 8px 16px;
                    border: none;
                    border-radius: 6px;
                    font-size: 0.85rem;
                    font-weight: 600;
                    cursor: pointer;
                    transition: all 0.3s ease;
                    display: flex;
                    align-items: center;
                    gap: 6px;
                    text-decoration: none;
                }

                .btn:disabled {
                    opacity: 0.6;
                    cursor: not-allowed;
                }

                .btn-sm {
                    padding: 6px 12px;
                    font-size: 0.8rem;
                }

                .btn-outline {
                    background: transparent;
                    color: #3498db;
                    border: 2px solid #3498db;
                }

                .btn-outline:hover:not(:disabled) {
                    background: #3498db;
                    color: white;
                }

                .btn-danger {
                    background: #e74c3c;
                    color: white;
                }

                .btn-danger:hover:not(:disabled) {
                    background: #c0392b;
                    transform: translateY(-1px);
                }

                .btn-secondary {
                    background: #6c757d;
                    color: white;
                }

                .btn-secondary:hover:not(:disabled) {
                    background: #5a6268;
                }

                .btn-primary {
                    background: #3498db;
                    color: white;
                }

                .btn-primary:hover:not(:disabled) {
                    background: #2980b9;
                }

                .session-details {
                    background: #f8f9fa;
                    border-radius: 8px;
                    padding: 16px;
                    margin-top: 16px;
                }

                .merchant-info {
                    display: flex;
                    align-items: center;
                    gap: 16px;
                    margin-bottom: 16px;
                }

                .merchant-avatar {
                    width: 50px;
                    height: 50px;
                    border-radius: 50%;
                    background: linear-gradient(135deg, #3498db, #2980b9);
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    color: white;
                    font-size: 1.5rem;
                    flex-shrink: 0;
                }

                .merchant-details {
                    flex: 1;
                }

                .merchant-name {
                    font-size: 1.1rem;
                    font-weight: 600;
                    color: #2c3e50;
                    margin-bottom: 4px;
                }

                .merchant-meta {
                    display: flex;
                    gap: 16px;
                    font-size: 0.85rem;
                    color: #6c757d;
                }

                .session-timer {
                    background: white;
                    border-radius: 6px;
                    padding: 12px;
                }

                .timer-label {
                    font-size: 0.8rem;
                    color: #6c757d;
                    margin-bottom: 8px;
                    font-weight: 600;
                }

                .timer-progress {
                    width: 100%;
                    height: 6px;
                    background: #e9ecef;
                    border-radius: 3px;
                    overflow: hidden;
                    margin-bottom: 8px;
                }

                .timer-bar {
                    height: 100%;
                    background: linear-gradient(90deg, #27ae60, #f39c12, #e74c3c);
                    width: 100%;
                    transition: width 1s linear;
                }

                .timer-text {
                    font-size: 0.9rem;
                    font-weight: 600;
                    color: #2c3e50;
                    text-align: center;
                }

                .session-history {
                    background: #f8f9fa;
                    border-radius: 8px;
                    padding: 16px;
                    margin-top: 16px;
                }

                .history-header {
                    display: flex;
                    justify-content: space-between;
                    align-items: center;
                    margin-bottom: 16px;
                }

                .history-header h4 {
                    margin: 0;
                    color: #2c3e50;
                    font-size: 1rem;
                }

                .history-list {
                    max-height: 300px;
                    overflow-y: auto;
                }

                .history-item {
                    display: flex;
                    align-items: center;
                    gap: 12px;
                    padding: 12px;
                    background: white;
                    border-radius: 6px;
                    margin-bottom: 8px;
                    cursor: pointer;
                    transition: all 0.3s ease;
                }

                .history-item:hover {
                    background: #e3f2fd;
                    transform: translateX(4px);
                }

                .history-avatar {
                    width: 35px;
                    height: 35px;
                    border-radius: 50%;
                    background: linear-gradient(135deg, #95a5a6, #7f8c8d);
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    color: white;
                    font-size: 0.9rem;
                    flex-shrink: 0;
                }

                .history-details {
                    flex: 1;
                }

                .history-name {
                    font-weight: 600;
                    color: #2c3e50;
                    font-size: 0.9rem;
                    margin-bottom: 2px;
                }

                .history-meta {
                    font-size: 0.8rem;
                    color: #6c757d;
                }

                .no-history {
                    text-align: center;
                    padding: 40px 20px;
                    color: #6c757d;
                }

                .no-history i {
                    font-size: 2rem;
                    color: #dee2e6;
                    margin-bottom: 12px;
                }

                .no-history p {
                    margin: 0;
                    font-size: 0.9rem;
                }

                .session-switch-modal {
                    position: fixed;
                    top: 0;
                    left: 0;
                    right: 0;
                    bottom: 0;
                    z-index: 10000;
                }

                .modal-overlay {
                    position: absolute;
                    top: 0;
                    left: 0;
                    right: 0;
                    bottom: 0;
                    background: rgba(0, 0, 0, 0.5);
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    padding: 20px;
                }

                .modal-content {
                    background: white;
                    border-radius: 12px;
                    max-width: 500px;
                    width: 100%;
                    max-height: 90vh;
                    overflow-y: auto;
                    box-shadow: 0 20px 40px rgba(0, 0, 0, 0.2);
                }

                .modal-header {
                    display: flex;
                    justify-content: space-between;
                    align-items: center;
                    padding: 20px;
                    border-bottom: 2px solid #f8f9fa;
                }

                .modal-header h4 {
                    margin: 0;
                    color: #2c3e50;
                    font-size: 1.2rem;
                }

                .modal-body {
                    padding: 20px;
                }

                .current-session-warning {
                    display: flex;
                    align-items: flex-start;
                    gap: 12px;
                    background: #fff3cd;
                    border: 1px solid #ffeaa7;
                    border-radius: 8px;
                    padding: 16px;
                    margin-bottom: 20px;
                }

                .current-session-warning i {
                    color: #f39c12;
                    font-size: 1.2rem;
                    margin-top: 2px;
                }

                .current-session-warning p {
                    margin: 0;
                    color: #856404;
                    font-size: 0.9rem;
                    line-height: 1.4;
                }

                .new-merchant-info {
                    background: #f8f9fa;
                    border-radius: 8px;
                    padding: 16px;
                }

                .modal-footer {
                    display: flex;
                    justify-content: flex-end;
                    gap: 12px;
                    padding: 20px;
                    border-top: 2px solid #f8f9fa;
                }

                /* Responsive Design */
                @media (max-width: 768px) {
                    .session-manager-container {
                        padding: 16px;
                    }

                    .session-header {
                        flex-direction: column;
                        gap: 12px;
                        align-items: flex-start;
                    }

                    .session-actions {
                        width: 100%;
                        justify-content: flex-end;
                    }

                    .merchant-info {
                        flex-direction: column;
                        align-items: flex-start;
                        gap: 12px;
                    }

                    .merchant-meta {
                        flex-direction: column;
                        gap: 4px;
                    }

                    .modal-content {
                        margin: 10px;
                        max-width: none;
                    }

                    .modal-footer {
                        flex-direction: column;
                    }
                }

                /* Animation for session status changes */
                .status-indicator {
                    animation: pulse 2s infinite;
                }

                @keyframes pulse {
                    0% { opacity: 1; }
                    50% { opacity: 0.7; }
                    100% { opacity: 1; }
                }

                /* Loading state */
                .loading {
                    opacity: 0.6;
                    pointer-events: none;
                }
            </style>
        `;

        // Add styles to head if not already added
        if (!document.querySelector('#session-manager-styles')) {
            const styleElement = document.createElement('style');
            styleElement.id = 'session-manager-styles';
            styleElement.textContent = styles;
            document.head.appendChild(styleElement);
        }
    }

    bindEvents() {
        const endSessionBtn = document.getElementById('endSessionBtn');
        const sessionHistoryBtn = document.getElementById('sessionHistoryBtn');
        const closeHistoryBtn = document.getElementById('closeHistoryBtn');
        const closeSwitchModalBtn = document.getElementById('closeSwitchModalBtn');
        const cancelSwitchBtn = document.getElementById('cancelSwitchBtn');
        const confirmSwitchBtn = document.getElementById('confirmSwitchBtn');

        // End session button
        endSessionBtn.addEventListener('click', () => {
            this.endCurrentSession();
        });

        // Session history button
        sessionHistoryBtn.addEventListener('click', () => {
            this.toggleHistory();
        });

        // Close history button
        closeHistoryBtn.addEventListener('click', () => {
            this.hideHistory();
        });

        // Close switch modal buttons
        closeSwitchModalBtn.addEventListener('click', () => {
            this.hideSwitchModal();
        });

        cancelSwitchBtn.addEventListener('click', () => {
            this.hideSwitchModal();
        });

        confirmSwitchBtn.addEventListener('click', () => {
            this.confirmSessionSwitch();
        });

        // Keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            if (e.ctrlKey || e.metaKey) {
                switch (e.key) {
                    case 'e':
                        e.preventDefault();
                        if (this.isSessionActive) {
                            this.endCurrentSession();
                        }
                        break;
                    case 'h':
                        e.preventDefault();
                        this.toggleHistory();
                        break;
                }
            }
        });

        // Handle page visibility changes
        document.addEventListener('visibilitychange', () => {
            if (document.hidden) {
                this.pauseSessionMonitoring();
            } else {
                this.resumeSessionMonitoring();
            }
        });

        // Handle beforeunload to save session state
        window.addEventListener('beforeunload', () => {
            this.saveSessionToStorage();
        });
    }

    // Session Management Methods
    startSession(merchant) {
        if (this.isSessionActive) {
            this.showSwitchModal(merchant);
            return;
        }

        this.currentSession = {
            merchant: merchant,
            startTime: new Date(),
            lastActivity: new Date(),
            sessionId: this.generateSessionId()
        };

        this.isSessionActive = true;
        this.updateSessionUI();
        this.startSessionTimer();
        this.saveSessionToStorage();

        // Call callback
        if (this.onSessionStart) {
            this.onSessionStart(this.currentSession);
        }

        console.log(`Session started for merchant: ${merchant.name}`);
    }

    endCurrentSession() {
        if (!this.isSessionActive) return;

        // Add to history
        this.addToHistory(this.currentSession);

        // Call callback
        if (this.onSessionEnd) {
            this.onSessionEnd(this.currentSession);
        }

        // Reset overview
        if (this.onOverviewReset) {
            this.onOverviewReset();
        }

        this.currentSession = null;
        this.isSessionActive = false;
        this.clearSessionTimer();
        this.updateSessionUI();
        this.saveSessionToStorage();

        console.log('Session ended');
    }

    switchSession(merchant) {
        if (!this.isSessionActive) {
            this.startSession(merchant);
            return;
        }

        this.showSwitchModal(merchant);
    }

    showSwitchModal(merchant) {
        const modal = document.getElementById('sessionSwitchModal');
        const newMerchantInfo = document.getElementById('newMerchantInfo');

        newMerchantInfo.innerHTML = `
            <div class="merchant-info">
                <div class="merchant-avatar">
                    ${merchant.name.charAt(0).toUpperCase()}
                </div>
                <div class="merchant-details">
                    <div class="merchant-name">${merchant.name}</div>
                    <div class="merchant-meta">
                        <span>ID: ${merchant.id}</span>
                        <span>Industry: ${merchant.industry || 'N/A'}</span>
                    </div>
                </div>
            </div>
        `;

        modal.style.display = 'block';
        this.pendingMerchant = merchant;
    }

    hideSwitchModal() {
        const modal = document.getElementById('sessionSwitchModal');
        modal.style.display = 'none';
        this.pendingMerchant = null;
    }

    confirmSessionSwitch() {
        if (!this.pendingMerchant) return;

        // End current session
        this.endCurrentSession();

        // Start new session
        this.startSession(this.pendingMerchant);

        // Call callback
        if (this.onSessionSwitch) {
            this.onSessionSwitch(this.pendingMerchant);
        }

        this.hideSwitchModal();
    }

    // Session Timer Management
    startSessionTimer() {
        this.clearSessionTimer();
        this.updateTimer();
        this.sessionTimer = setInterval(() => {
            this.updateTimer();
            this.checkSessionTimeout();
        }, 1000);
    }

    clearSessionTimer() {
        if (this.sessionTimer) {
            clearInterval(this.sessionTimer);
            this.sessionTimer = null;
        }
    }

    updateTimer() {
        if (!this.isSessionActive || !this.currentSession) return;

        const now = new Date();
        const elapsed = now - this.currentSession.startTime;
        const remaining = this.sessionTimeout - elapsed;

        if (remaining <= 0) {
            this.handleSessionTimeout();
            return;
        }

        const minutes = Math.floor(remaining / 60000);
        const seconds = Math.floor((remaining % 60000) / 1000);
        const progress = (remaining / this.sessionTimeout) * 100;

        const timerText = document.getElementById('timerText');
        const timerBar = document.getElementById('timerBar');
        const statusIndicator = document.getElementById('statusIndicator');

        if (timerText) {
            timerText.textContent = `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
        }

        if (timerBar) {
            timerBar.style.width = `${progress}%`;
        }

        // Update status indicator based on remaining time
        if (statusIndicator) {
            statusIndicator.className = 'status-indicator';
            if (remaining < 5 * 60 * 1000) { // Less than 5 minutes
                statusIndicator.classList.add('danger');
            } else if (remaining < 10 * 60 * 1000) { // Less than 10 minutes
                statusIndicator.classList.add('warning');
            } else {
                statusIndicator.classList.add('active');
            }
        }
    }

    checkSessionTimeout() {
        if (!this.isSessionActive || !this.currentSession) return;

        const now = new Date();
        const elapsed = now - this.currentSession.startTime;

        if (elapsed >= this.sessionTimeout) {
            this.handleSessionTimeout();
        }
    }

    handleSessionTimeout() {
        console.log('Session timeout reached');
        
        // Call callback
        if (this.onSessionTimeout) {
            this.onSessionTimeout(this.currentSession);
        }

        // End session
        this.endCurrentSession();
    }

    // Session History Management
    addToHistory(session) {
        const historyItem = {
            merchant: session.merchant,
            startTime: session.startTime,
            endTime: new Date(),
            duration: new Date() - session.startTime,
            sessionId: session.sessionId
        };

        this.sessionHistory.unshift(historyItem);

        // Limit history size
        if (this.sessionHistory.length > this.maxHistorySize) {
            this.sessionHistory = this.sessionHistory.slice(0, this.maxHistorySize);
        }

        this.saveSessionToStorage();
    }

    toggleHistory() {
        const history = document.getElementById('sessionHistory');
        if (history.style.display === 'none') {
            this.showHistory();
        } else {
            this.hideHistory();
        }
    }

    showHistory() {
        const history = document.getElementById('sessionHistory');
        const historyList = document.getElementById('historyList');

        if (this.sessionHistory.length === 0) {
            historyList.innerHTML = `
                <div class="no-history">
                    <i class="fas fa-history"></i>
                    <p>No recent sessions</p>
                </div>
            `;
        } else {
            historyList.innerHTML = this.sessionHistory.map(item => `
                <div class="history-item" data-session-id="${item.sessionId}">
                    <div class="history-avatar">
                        ${item.merchant.name.charAt(0).toUpperCase()}
                    </div>
                    <div class="history-details">
                        <div class="history-name">${item.merchant.name}</div>
                        <div class="history-meta">
                            ${this.formatDate(item.startTime)} - ${this.formatDuration(item.duration)}
                        </div>
                    </div>
                </div>
            `).join('');

            // Add click handlers for history items
            historyList.querySelectorAll('.history-item').forEach(item => {
                item.addEventListener('click', () => {
                    const sessionId = item.dataset.sessionId;
                    const historyItem = this.sessionHistory.find(h => h.sessionId === sessionId);
                    if (historyItem) {
                        this.switchSession(historyItem.merchant);
                    }
                });
            });
        }

        history.style.display = 'block';
    }

    hideHistory() {
        const history = document.getElementById('sessionHistory');
        history.style.display = 'none';
    }

    // UI Update Methods
    updateSessionUI() {
        const sessionDetails = document.getElementById('sessionDetails');
        const endSessionBtn = document.getElementById('endSessionBtn');
        const sessionHistoryBtn = document.getElementById('sessionHistoryBtn');
        const statusIndicator = document.getElementById('statusIndicator');
        const statusText = document.getElementById('statusText');

        if (this.isSessionActive && this.currentSession) {
            // Show session details
            sessionDetails.style.display = 'block';
            endSessionBtn.disabled = false;
            sessionHistoryBtn.disabled = false;

            // Update merchant info
            document.getElementById('merchantAvatar').innerHTML = 
                this.currentSession.merchant.name.charAt(0).toUpperCase();
            document.getElementById('merchantName').textContent = this.currentSession.merchant.name;
            document.getElementById('merchantId').textContent = `ID: ${this.currentSession.merchant.id}`;

            // Update status
            statusIndicator.className = 'status-indicator active';
            statusText.textContent = 'Active session';
        } else {
            // Hide session details
            sessionDetails.style.display = 'none';
            endSessionBtn.disabled = true;
            sessionHistoryBtn.disabled = this.sessionHistory.length === 0;

            // Update status
            statusIndicator.className = 'status-indicator';
            statusText.textContent = 'No active session';
        }
    }

    // Storage Methods
    saveSessionToStorage() {
        const sessionData = {
            currentSession: this.currentSession,
            sessionHistory: this.sessionHistory,
            timestamp: new Date().toISOString()
        };

        try {
            localStorage.setItem(this.storageKey, JSON.stringify(sessionData));
        } catch (error) {
            console.error('Failed to save session to storage:', error);
        }
    }

    loadSessionFromStorage() {
        try {
            const sessionData = localStorage.getItem(this.storageKey);
            if (sessionData) {
                const parsed = JSON.parse(sessionData);
                
                // Check if session is still valid (not expired)
                if (parsed.currentSession) {
                    const sessionStart = new Date(parsed.currentSession.startTime);
                    const now = new Date();
                    const elapsed = now - sessionStart;

                    if (elapsed < this.sessionTimeout) {
                        this.currentSession = parsed.currentSession;
                        this.isSessionActive = true;
                    } else {
                        // Session expired, add to history
                        this.addToHistory(parsed.currentSession);
                    }
                }

                this.sessionHistory = parsed.sessionHistory || [];
            }
        } catch (error) {
            console.error('Failed to load session from storage:', error);
            this.sessionHistory = [];
        }
    }

    // Monitoring Methods
    startSessionMonitoring() {
        // Monitor for user activity
        const activityEvents = ['mousedown', 'mousemove', 'keypress', 'scroll', 'touchstart'];
        
        activityEvents.forEach(event => {
            document.addEventListener(event, () => {
                this.updateLastActivity();
            }, true);
        });
    }

    pauseSessionMonitoring() {
        // Pause timer when page is not visible
        this.clearSessionTimer();
    }

    resumeSessionMonitoring() {
        // Resume timer when page becomes visible
        if (this.isSessionActive) {
            this.startSessionTimer();
        }
    }

    updateLastActivity() {
        if (this.isSessionActive && this.currentSession) {
            this.currentSession.lastActivity = new Date();
        }
    }

    // Utility Methods
    generateSessionId() {
        return 'session_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
    }

    formatDate(date) {
        return new Date(date).toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    }

    formatDuration(milliseconds) {
        const minutes = Math.floor(milliseconds / 60000);
        const hours = Math.floor(minutes / 60);
        
        if (hours > 0) {
            return `${hours}h ${minutes % 60}m`;
        } else {
            return `${minutes}m`;
        }
    }

    // Public API Methods
    getCurrentSession() {
        return this.currentSession;
    }

    isActive() {
        return this.isSessionActive;
    }

    getSessionHistory() {
        return [...this.sessionHistory];
    }

    clearHistory() {
        this.sessionHistory = [];
        this.saveSessionToStorage();
        this.updateSessionUI();
    }

    setSessionTimeout(timeout) {
        this.sessionTimeout = timeout;
        if (this.isSessionActive) {
            this.startSessionTimer();
        }
    }

    destroy() {
        this.clearSessionTimer();
        this.saveSessionToStorage();
        if (this.container) {
            this.container.innerHTML = '';
        }
    }
}

// Export for use in other scripts
if (typeof module !== 'undefined' && module.exports) {
    module.exports = SessionManager;
}

// Auto-initialize if container is found
document.addEventListener('DOMContentLoaded', () => {
    const sessionContainer = document.getElementById('sessionManagerContainer');
    if (sessionContainer && !window.sessionManager) {
        window.sessionManager = new SessionManager({
            container: sessionContainer
        });
    }
});
