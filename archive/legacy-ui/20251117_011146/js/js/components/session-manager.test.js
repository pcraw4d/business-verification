/**
 * Unit Tests for Session Manager Component
 * Tests session management, state persistence, and switching functionality
 */

// Mock DOM environment for testing
const { JSDOM } = require('jsdom');
const fs = require('fs');
const path = require('path');

// Setup DOM environment
const dom = new JSDOM(`
    <!DOCTYPE html>
    <html>
        <head></head>
        <body>
            <div id="testContainer"></div>
        </body>
    </html>
`, {
    url: 'http://localhost',
    pretendToBeVisual: true,
    resources: 'usable'
});

global.window = dom.window;
global.document = dom.window.document;
global.localStorage = {
    data: {},
    getItem: function(key) { return this.data[key] || null; },
    setItem: function(key, value) { this.data[key] = value; },
    removeItem: function(key) { delete this.data[key]; },
    clear: function() { this.data = {}; }
};

// Load the SessionManager class
const sessionManagerCode = fs.readFileSync(path.join(__dirname, 'session-manager.js'), 'utf8');
eval(sessionManagerCode);

describe('SessionManager', () => {
    let sessionManager;
    let container;

    beforeEach(() => {
        // Clear localStorage
        localStorage.clear();
        
        // Create fresh container
        container = document.getElementById('testContainer');
        container.innerHTML = '';
        
        // Create new session manager instance
        sessionManager = new SessionManager({
            container: container,
            sessionTimeout: 5000, // 5 seconds for testing
            maxHistorySize: 3
        });
    });

    afterEach(() => {
        if (sessionManager) {
            sessionManager.destroy();
        }
    });

    describe('Initialization', () => {
        test('should initialize with default values', () => {
            expect(sessionManager.isSessionActive).toBe(false);
            expect(sessionManager.currentSession).toBeNull();
            expect(sessionManager.sessionHistory).toEqual([]);
            expect(sessionManager.sessionTimeout).toBe(5000);
        });

        test('should create session interface elements', () => {
            expect(document.getElementById('sessionManagerContainer')).toBeTruthy();
            expect(document.getElementById('sessionStatus')).toBeTruthy();
            expect(document.getElementById('sessionDetails')).toBeTruthy();
            expect(document.getElementById('endSessionBtn')).toBeTruthy();
        });

        test('should load session from storage if valid', () => {
            const mockSession = {
                merchant: { id: '1', name: 'Test Merchant' },
                startTime: new Date().toISOString(),
                sessionId: 'test-session'
            };

            localStorage.setItem('merchant_session', JSON.stringify({
                currentSession: mockSession,
                sessionHistory: [],
                timestamp: new Date().toISOString()
            }));

            const newSessionManager = new SessionManager({
                container: document.createElement('div'),
                sessionTimeout: 5000
            });

            expect(newSessionManager.isSessionActive).toBe(true);
            expect(newSessionManager.currentSession).toBeTruthy();
        });
    });

    describe('Session Management', () => {
        const mockMerchant = {
            id: '1',
            name: 'Test Merchant',
            industry: 'Technology'
        };

        test('should start a new session', () => {
            const onSessionStart = jest.fn();
            sessionManager.onSessionStart = onSessionStart;

            sessionManager.startSession(mockMerchant);

            expect(sessionManager.isSessionActive).toBe(true);
            expect(sessionManager.currentSession).toBeTruthy();
            expect(sessionManager.currentSession.merchant).toEqual(mockMerchant);
            expect(sessionManager.currentSession.sessionId).toBeTruthy();
            expect(onSessionStart).toHaveBeenCalledWith(sessionManager.currentSession);
        });

        test('should end current session', () => {
            const onSessionEnd = jest.fn();
            const onOverviewReset = jest.fn();
            sessionManager.onSessionEnd = onSessionEnd;
            sessionManager.onOverviewReset = onOverviewReset;

            sessionManager.startSession(mockMerchant);
            sessionManager.endCurrentSession();

            expect(sessionManager.isSessionActive).toBe(false);
            expect(sessionManager.currentSession).toBeNull();
            expect(sessionManager.sessionHistory).toHaveLength(1);
            expect(onSessionEnd).toHaveBeenCalled();
            expect(onOverviewReset).toHaveBeenCalled();
        });

        test('should show switch modal when starting session while another is active', () => {
            sessionManager.startSession(mockMerchant);
            
            const anotherMerchant = { id: '2', name: 'Another Merchant' };
            sessionManager.startSession(anotherMerchant);

            const modal = document.getElementById('sessionSwitchModal');
            expect(modal.style.display).toBe('block');
        });

        test('should switch sessions when confirmed', () => {
            const onSessionSwitch = jest.fn();
            sessionManager.onSessionSwitch = onSessionSwitch;

            sessionManager.startSession(mockMerchant);
            
            const anotherMerchant = { id: '2', name: 'Another Merchant' };
            sessionManager.startSession(anotherMerchant);
            sessionManager.confirmSessionSwitch();

            expect(sessionManager.currentSession.merchant).toEqual(anotherMerchant);
            expect(sessionManager.sessionHistory).toHaveLength(1);
            expect(onSessionSwitch).toHaveBeenCalledWith(anotherMerchant);
        });
    });

    describe('Session Timer', () => {
        const mockMerchant = { id: '1', name: 'Test Merchant' };

        test('should start session timer', () => {
            sessionManager.startSession(mockMerchant);
            expect(sessionManager.sessionTimer).toBeTruthy();
        });

        test('should clear session timer when session ends', () => {
            sessionManager.startSession(mockMerchant);
            const timer = sessionManager.sessionTimer;
            
            sessionManager.endCurrentSession();
            expect(sessionManager.sessionTimer).toBeNull();
        });

        test('should handle session timeout', (done) => {
            const onSessionTimeout = jest.fn();
            sessionManager.onSessionTimeout = onSessionTimeout;

            sessionManager.startSession(mockMerchant);

            // Wait for timeout
            setTimeout(() => {
                expect(sessionManager.isSessionActive).toBe(false);
                expect(onSessionTimeout).toHaveBeenCalled();
                done();
            }, 6000);
        });

        test('should update timer display', () => {
            sessionManager.startSession(mockMerchant);
            sessionManager.updateTimer();

            const timerText = document.getElementById('timerText');
            const timerBar = document.getElementById('timerBar');
            
            expect(timerText.textContent).toMatch(/^\d{2}:\d{2}$/);
            expect(timerBar.style.width).toBeTruthy();
        });
    });

    describe('Session History', () => {
        const mockMerchant = { id: '1', name: 'Test Merchant' };

        test('should add session to history when ended', () => {
            sessionManager.startSession(mockMerchant);
            sessionManager.endCurrentSession();

            expect(sessionManager.sessionHistory).toHaveLength(1);
            expect(sessionManager.sessionHistory[0].merchant).toEqual(mockMerchant);
        });

        test('should limit history size', () => {
            // Add more sessions than maxHistorySize
            for (let i = 0; i < 5; i++) {
                const merchant = { id: i.toString(), name: `Merchant ${i}` };
                sessionManager.startSession(merchant);
                sessionManager.endCurrentSession();
            }

            expect(sessionManager.sessionHistory).toHaveLength(3);
        });

        test('should show and hide history', () => {
            sessionManager.startSession(mockMerchant);
            sessionManager.endCurrentSession();

            sessionManager.showHistory();
            const history = document.getElementById('sessionHistory');
            expect(history.style.display).toBe('block');

            sessionManager.hideHistory();
            expect(history.style.display).toBe('none');
        });

        test('should clear history', () => {
            sessionManager.startSession(mockMerchant);
            sessionManager.endCurrentSession();
            sessionManager.clearHistory();

            expect(sessionManager.sessionHistory).toHaveLength(0);
        });
    });

    describe('Storage Persistence', () => {
        const mockMerchant = { id: '1', name: 'Test Merchant' };

        test('should save session to storage', () => {
            sessionManager.startSession(mockMerchant);
            sessionManager.saveSessionToStorage();

            const stored = localStorage.getItem('merchant_session');
            expect(stored).toBeTruthy();
            
            const parsed = JSON.parse(stored);
            expect(parsed.currentSession).toBeTruthy();
            expect(parsed.currentSession.merchant).toEqual(mockMerchant);
        });

        test('should load session from storage', () => {
            const mockSession = {
                merchant: mockMerchant,
                startTime: new Date().toISOString(),
                sessionId: 'test-session'
            };

            localStorage.setItem('merchant_session', JSON.stringify({
                currentSession: mockSession,
                sessionHistory: [],
                timestamp: new Date().toISOString()
            }));

            const newSessionManager = new SessionManager({
                container: document.createElement('div'),
                sessionTimeout: 5000
            });

            expect(newSessionManager.isSessionActive).toBe(true);
            expect(newSessionManager.currentSession.merchant).toEqual(mockMerchant);
        });

        test('should not load expired session from storage', () => {
            const expiredSession = {
                merchant: mockMerchant,
                startTime: new Date(Date.now() - 10000).toISOString(), // 10 seconds ago
                sessionId: 'expired-session'
            };

            localStorage.setItem('merchant_session', JSON.stringify({
                currentSession: expiredSession,
                sessionHistory: [],
                timestamp: new Date().toISOString()
            }));

            const newSessionManager = new SessionManager({
                container: document.createElement('div'),
                sessionTimeout: 5000
            });

            expect(newSessionManager.isSessionActive).toBe(false);
            expect(newSessionManager.sessionHistory).toHaveLength(1); // Expired session added to history
        });
    });

    describe('UI Updates', () => {
        const mockMerchant = { id: '1', name: 'Test Merchant' };

        test('should update UI when session starts', () => {
            sessionManager.startSession(mockMerchant);
            sessionManager.updateSessionUI();

            const sessionDetails = document.getElementById('sessionDetails');
            const endSessionBtn = document.getElementById('endSessionBtn');
            const statusText = document.getElementById('statusText');

            expect(sessionDetails.style.display).toBe('block');
            expect(endSessionBtn.disabled).toBe(false);
            expect(statusText.textContent).toBe('Active session');
        });

        test('should update UI when session ends', () => {
            sessionManager.startSession(mockMerchant);
            sessionManager.endCurrentSession();
            sessionManager.updateSessionUI();

            const sessionDetails = document.getElementById('sessionDetails');
            const endSessionBtn = document.getElementById('endSessionBtn');
            const statusText = document.getElementById('statusText');

            expect(sessionDetails.style.display).toBe('none');
            expect(endSessionBtn.disabled).toBe(true);
            expect(statusText.textContent).toBe('No active session');
        });

        test('should update merchant info in UI', () => {
            sessionManager.startSession(mockMerchant);
            sessionManager.updateSessionUI();

            const merchantName = document.getElementById('merchantName');
            const merchantId = document.getElementById('merchantId');

            expect(merchantName.textContent).toBe('Test Merchant');
            expect(merchantId.textContent).toBe('ID: 1');
        });
    });

    describe('Event Handling', () => {
        const mockMerchant = { id: '1', name: 'Test Merchant' };

        test('should handle end session button click', () => {
            sessionManager.startSession(mockMerchant);
            
            const endSessionBtn = document.getElementById('endSessionBtn');
            endSessionBtn.click();

            expect(sessionManager.isSessionActive).toBe(false);
        });

        test('should handle history button click', () => {
            sessionManager.startSession(mockMerchant);
            sessionManager.endCurrentSession();

            const historyBtn = document.getElementById('sessionHistoryBtn');
            historyBtn.click();

            const history = document.getElementById('sessionHistory');
            expect(history.style.display).toBe('block');
        });

        test('should handle keyboard shortcuts', () => {
            sessionManager.startSession(mockMerchant);

            // Test Ctrl+E (end session)
            const keyEvent = new KeyboardEvent('keydown', {
                key: 'e',
                ctrlKey: true
            });
            document.dispatchEvent(keyEvent);

            expect(sessionManager.isSessionActive).toBe(false);
        });
    });

    describe('Utility Methods', () => {
        test('should generate unique session IDs', () => {
            const id1 = sessionManager.generateSessionId();
            const id2 = sessionManager.generateSessionId();

            expect(id1).not.toBe(id2);
            expect(id1).toMatch(/^session_\d+_[a-z0-9]+$/);
        });

        test('should format dates correctly', () => {
            const date = new Date('2023-01-15T10:30:00Z');
            const formatted = sessionManager.formatDate(date);

            expect(formatted).toMatch(/Jan 15, 2023/);
        });

        test('should format duration correctly', () => {
            expect(sessionManager.formatDuration(30000)).toBe('0m'); // 30 seconds
            expect(sessionManager.formatDuration(120000)).toBe('2m'); // 2 minutes
            expect(sessionManager.formatDuration(3660000)).toBe('1h 1m'); // 1 hour 1 minute
        });

        test('should get current session', () => {
            const mockMerchant = { id: '1', name: 'Test Merchant' };
            sessionManager.startSession(mockMerchant);

            const currentSession = sessionManager.getCurrentSession();
            expect(currentSession).toBeTruthy();
            expect(currentSession.merchant).toEqual(mockMerchant);
        });

        test('should check if session is active', () => {
            expect(sessionManager.isActive()).toBe(false);

            const mockMerchant = { id: '1', name: 'Test Merchant' };
            sessionManager.startSession(mockMerchant);

            expect(sessionManager.isActive()).toBe(true);
        });

        test('should get session history', () => {
            const mockMerchant = { id: '1', name: 'Test Merchant' };
            sessionManager.startSession(mockMerchant);
            sessionManager.endCurrentSession();

            const history = sessionManager.getSessionHistory();
            expect(history).toHaveLength(1);
            expect(history[0].merchant).toEqual(mockMerchant);
        });

        test('should set session timeout', () => {
            sessionManager.setSessionTimeout(10000);
            expect(sessionManager.sessionTimeout).toBe(10000);
        });
    });

    describe('Error Handling', () => {
        test('should handle localStorage errors gracefully', () => {
            // Mock localStorage to throw error
            const originalSetItem = localStorage.setItem;
            localStorage.setItem = jest.fn(() => {
                throw new Error('Storage quota exceeded');
            });

            const mockMerchant = { id: '1', name: 'Test Merchant' };
            
            // Should not throw error
            expect(() => {
                sessionManager.startSession(mockMerchant);
            }).not.toThrow();

            // Restore original method
            localStorage.setItem = originalSetItem;
        });

        test('should handle invalid stored data gracefully', () => {
            localStorage.setItem('merchant_session', 'invalid json');

            const newSessionManager = new SessionManager({
                container: document.createElement('div'),
                sessionTimeout: 5000
            });

            expect(newSessionManager.isSessionActive).toBe(false);
            expect(newSessionManager.sessionHistory).toEqual([]);
        });
    });

    describe('Cleanup', () => {
        test('should destroy session manager properly', () => {
            const mockMerchant = { id: '1', name: 'Test Merchant' };
            sessionManager.startSession(mockMerchant);

            sessionManager.destroy();

            expect(sessionManager.sessionTimer).toBeNull();
            expect(container.innerHTML).toBe('');
        });
    });
});

// Integration tests
describe('SessionManager Integration', () => {
    let sessionManager;
    let container;

    beforeEach(() => {
        localStorage.clear();
        container = document.getElementById('testContainer');
        container.innerHTML = '';
        sessionManager = new SessionManager({
            container: container,
            sessionTimeout: 10000
        });
    });

    afterEach(() => {
        if (sessionManager) {
            sessionManager.destroy();
        }
    });

    test('should handle complete session lifecycle', () => {
        const mockMerchant = { id: '1', name: 'Test Merchant' };
        const onSessionStart = jest.fn();
        const onSessionEnd = jest.fn();
        const onOverviewReset = jest.fn();

        sessionManager.onSessionStart = onSessionStart;
        sessionManager.onSessionEnd = onSessionEnd;
        sessionManager.onOverviewReset = onOverviewReset;

        // Start session
        sessionManager.startSession(mockMerchant);
        expect(sessionManager.isSessionActive).toBe(true);
        expect(onSessionStart).toHaveBeenCalled();

        // End session
        sessionManager.endCurrentSession();
        expect(sessionManager.isSessionActive).toBe(false);
        expect(onSessionEnd).toHaveBeenCalled();
        expect(onOverviewReset).toHaveBeenCalled();
        expect(sessionManager.sessionHistory).toHaveLength(1);
    });

    test('should handle session switching workflow', () => {
        const merchant1 = { id: '1', name: 'Merchant 1' };
        const merchant2 = { id: '2', name: 'Merchant 2' };
        const onSessionSwitch = jest.fn();

        sessionManager.onSessionSwitch = onSessionSwitch;

        // Start first session
        sessionManager.startSession(merchant1);
        expect(sessionManager.currentSession.merchant).toEqual(merchant1);

        // Switch to second session
        sessionManager.startSession(merchant2);
        sessionManager.confirmSessionSwitch();

        expect(sessionManager.currentSession.merchant).toEqual(merchant2);
        expect(sessionManager.sessionHistory).toHaveLength(1);
        expect(onSessionSwitch).toHaveBeenCalledWith(merchant2);
    });

    test('should persist session across page reloads', () => {
        const mockMerchant = { id: '1', name: 'Test Merchant' };
        
        // Start session
        sessionManager.startSession(mockMerchant);
        const sessionId = sessionManager.currentSession.sessionId;

        // Simulate page reload by creating new instance
        const newSessionManager = new SessionManager({
            container: document.createElement('div'),
            sessionTimeout: 10000
        });

        expect(newSessionManager.isSessionActive).toBe(true);
        expect(newSessionManager.currentSession.sessionId).toBe(sessionId);
        expect(newSessionManager.currentSession.merchant).toEqual(mockMerchant);
    });
});
