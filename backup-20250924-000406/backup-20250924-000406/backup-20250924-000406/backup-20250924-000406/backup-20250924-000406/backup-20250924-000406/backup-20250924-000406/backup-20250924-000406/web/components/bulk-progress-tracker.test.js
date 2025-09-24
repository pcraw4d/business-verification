/**
 * Unit Tests for Bulk Progress Tracker Component
 * 
 * Tests all functionality of the BulkProgressTracker class including:
 * - Initialization and configuration
 * - Progress tracking and updates
 * - Operation control (start, pause, resume, stop)
 * - Error handling and retry logic
 * - Event system
 * - UI updates and display
 */

// Mock DOM elements for testing
const mockElements = {
    progressBar: { style: { width: '0%' } },
    progressText: { textContent: '' },
    progressPercentage: { textContent: '0%' },
    progressCompleted: { textContent: '0' },
    progressFailed: { textContent: '0' },
    operationStatus: { className: '', innerHTML: '' },
    startButton: { addEventListener: jest.fn(), disabled: true, style: { display: 'none' } },
    pauseButton: { addEventListener: jest.fn(), disabled: true, style: { display: 'none' } },
    resumeButton: { addEventListener: jest.fn(), disabled: true, style: { display: 'none' } },
    stopButton: { addEventListener: jest.fn(), disabled: true, style: { display: 'none' } },
    exportButton: { addEventListener: jest.fn(), disabled: true, style: { display: 'none' } },
    operationLog: { appendChild: jest.fn(), scrollTop: 0, scrollHeight: 100 }
};

// Mock document.getElementById
global.document = {
    getElementById: jest.fn((id) => {
        const elementMap = {
            'progressBarFill': mockElements.progressBar,
            'progressText': mockElements.progressText,
            'progressPercentage': mockElements.progressPercentage,
            'progressCompleted': mockElements.progressCompleted,
            'progressFailed': mockElements.progressFailed,
            'operationStatus': mockElements.operationStatus,
            'startOperation': mockElements.startButton,
            'pauseOperation': mockElements.pauseButton,
            'resumeOperation': mockElements.resumeButton,
            'stopOperation': mockElements.stopButton,
            'exportResults': mockElements.exportButton,
            'operationLog': mockElements.operationLog
        };
        return elementMap[id] || null;
    }),
    createElement: jest.fn(() => ({
        href: '',
        download: '',
        click: jest.fn()
    })),
    body: {
        appendChild: jest.fn(),
        removeChild: jest.fn()
    }
};

// Mock window.URL
global.window = {
    URL: {
        createObjectURL: jest.fn(() => 'mock-url'),
        revokeObjectURL: jest.fn()
    }
};

// Mock fetch
global.fetch = jest.fn();

// Mock console methods
global.console = {
    ...console,
    warn: jest.fn(),
    error: jest.fn()
};

// Mock timers
jest.useFakeTimers();

describe('BulkProgressTracker', () => {
    let tracker;
    let mockOptions;

    beforeEach(() => {
        // Reset all mocks
        jest.clearAllMocks();
        
        // Reset mock elements
        Object.values(mockElements).forEach(element => {
            if (element.style) {
                element.style.width = '0%';
                element.style.display = 'none';
            }
            if (element.textContent !== undefined) {
                element.textContent = '';
            }
            if (element.className !== undefined) {
                element.className = '';
            }
            if (element.innerHTML !== undefined) {
                element.innerHTML = '';
            }
            if (element.disabled !== undefined) {
                element.disabled = true;
            }
        });

        // Default options
        mockOptions = {
            updateInterval: 1000,
            maxRetries: 3,
            retryDelay: 2000,
            autoRefresh: true,
            showDetailedLog: true,
            enablePauseResume: true,
            enableExport: true
        };

        // Create new tracker instance
        tracker = new BulkProgressTracker(mockOptions);
    });

    afterEach(() => {
        if (tracker) {
            tracker.destroy();
        }
        jest.clearAllTimers();
    });

    describe('Initialization', () => {
        test('should initialize with default options', () => {
            const defaultTracker = new BulkProgressTracker();
            expect(defaultTracker.options.updateInterval).toBe(1000);
            expect(defaultTracker.options.maxRetries).toBe(3);
            expect(defaultTracker.options.enablePauseResume).toBe(true);
        });

        test('should initialize with custom options', () => {
            const customOptions = {
                updateInterval: 500,
                maxRetries: 5,
                enablePauseResume: false
            };
            const customTracker = new BulkProgressTracker(customOptions);
            expect(customTracker.options.updateInterval).toBe(500);
            expect(customTracker.options.maxRetries).toBe(5);
            expect(customTracker.options.enablePauseResume).toBe(false);
        });

        test('should initialize with ready status', () => {
            expect(tracker.state.status).toBe('ready');
            expect(tracker.state.progress).toBe(0);
            expect(tracker.state.totalItems).toBe(0);
        });

        test('should bind DOM elements', () => {
            expect(tracker.elements.progressBar).toBe(mockElements.progressBar);
            expect(tracker.elements.progressText).toBe(mockElements.progressText);
            expect(tracker.elements.operationStatus).toBe(mockElements.operationStatus);
        });

        test('should bind event listeners', () => {
            expect(mockElements.startButton.addEventListener).toHaveBeenCalledWith('click', expect.any(Function));
            expect(mockElements.pauseButton.addEventListener).toHaveBeenCalledWith('click', expect.any(Function));
            expect(mockElements.resumeButton.addEventListener).toHaveBeenCalledWith('click', expect.any(Function));
            expect(mockElements.stopButton.addEventListener).toHaveBeenCalledWith('click', expect.any(Function));
            expect(mockElements.exportButton.addEventListener).toHaveBeenCalledWith('click', expect.any(Function));
        });
    });

    describe('Operation Control', () => {
        test('should start operation successfully', async () => {
            const operationData = {
                operationId: 'test-op-123',
                totalItems: 100
            };

            await tracker.startOperation(operationData);

            expect(tracker.state.operationId).toBe('test-op-123');
            expect(tracker.state.totalItems).toBe(100);
            expect(tracker.state.status).toBe('running');
            expect(tracker.state.startTime).toBeInstanceOf(Date);
        });

        test('should generate operation ID if not provided', async () => {
            await tracker.startOperation({ totalItems: 50 });

            expect(tracker.state.operationId).toMatch(/^bulk-op-\d+-[a-z0-9]+$/);
            expect(tracker.state.totalItems).toBe(50);
        });

        test('should pause running operation', async () => {
            await tracker.startOperation({ totalItems: 100 });
            tracker.pauseOperation();

            expect(tracker.state.status).toBe('paused');
        });

        test('should not pause non-running operation', () => {
            tracker.pauseOperation();
            expect(tracker.state.status).toBe('ready');
        });

        test('should resume paused operation', async () => {
            await tracker.startOperation({ totalItems: 100 });
            tracker.pauseOperation();
            tracker.resumeOperation();

            expect(tracker.state.status).toBe('running');
        });

        test('should not resume non-paused operation', async () => {
            await tracker.startOperation({ totalItems: 100 });
            tracker.resumeOperation();

            expect(tracker.state.status).toBe('running'); // Should remain running
        });

        test('should stop operation', async () => {
            await tracker.startOperation({ totalItems: 100 });
            tracker.stopOperation();

            expect(tracker.state.status).toBe('cancelled');
            expect(tracker.state.endTime).toBeInstanceOf(Date);
        });

        test('should not stop ready or completed operation', () => {
            tracker.stopOperation();
            expect(tracker.state.status).toBe('ready');

            tracker.state.status = 'completed';
            tracker.stopOperation();
            expect(tracker.state.status).toBe('completed');
        });
    });

    describe('Progress Updates', () => {
        beforeEach(async () => {
            await tracker.startOperation({ totalItems: 100 });
        });

        test('should update progress correctly', () => {
            const progressData = {
                completed: 25,
                failed: 5,
                currentItem: 'Test Item',
                errors: []
            };

            tracker.updateProgress(progressData);

            expect(tracker.state.completedItems).toBe(25);
            expect(tracker.state.failedItems).toBe(5);
            expect(tracker.state.currentItem).toBe('Test Item');
            expect(tracker.state.progress).toBe(30); // (25 + 5) / 100 * 100
        });

        test('should complete operation when all items processed', () => {
            const progressData = {
                completed: 95,
                failed: 5,
                currentItem: null,
                errors: []
            };

            tracker.updateProgress(progressData);

            expect(tracker.state.status).toBe('completed');
            expect(tracker.state.progress).toBe(100);
            expect(tracker.state.endTime).toBeInstanceOf(Date);
        });

        test('should calculate estimated time remaining', () => {
            // Simulate some progress
            tracker.state.startTime = new Date(Date.now() - 10000); // 10 seconds ago
            tracker.state.completedItems = 10;

            tracker.calculateEstimatedTime();

            expect(tracker.state.estimatedTimeRemaining).toBeGreaterThan(0);
        });

        test('should handle zero total items', () => {
            tracker.state.totalItems = 0;
            const progressData = { completed: 0, failed: 0 };

            tracker.updateProgress(progressData);

            expect(tracker.state.progress).toBe(0);
        });
    });

    describe('Display Updates', () => {
        test('should update progress bar width', () => {
            tracker.state.progress = 50;
            tracker.updateDisplay();

            expect(mockElements.progressBar.style.width).toBe('50%');
        });

        test('should update progress text', () => {
            tracker.state.status = 'running';
            tracker.state.completedItems = 25;
            tracker.state.totalItems = 100;
            tracker.updateDisplay();

            expect(mockElements.progressText.textContent).toContain('Processing 25 of 100');
        });

        test('should update progress percentage', () => {
            tracker.state.progress = 75;
            tracker.updateDisplay();

            expect(mockElements.progressPercentage.textContent).toBe('75%');
        });

        test('should update completed and failed counts', () => {
            tracker.state.completedItems = 20;
            tracker.state.failedItems = 5;
            tracker.updateDisplay();

            expect(mockElements.progressCompleted.textContent).toBe('20');
            expect(mockElements.progressFailed.textContent).toBe('5');
        });

        test('should update operation status', () => {
            tracker.state.status = 'running';
            tracker.updateDisplay();

            expect(mockElements.operationStatus.className).toContain('status-running');
            expect(mockElements.operationStatus.innerHTML).toContain('Running');
        });
    });

    describe('Control Buttons', () => {
        test('should show start button when ready', () => {
            tracker.state.status = 'ready';
            tracker.updateControlButtons();

            expect(mockElements.startButton.disabled).toBe(false);
            expect(mockElements.startButton.style.display).toBe('inline-flex');
        });

        test('should show pause and stop buttons when running', () => {
            tracker.state.status = 'running';
            tracker.updateControlButtons();

            expect(mockElements.pauseButton.disabled).toBe(false);
            expect(mockElements.pauseButton.style.display).toBe('inline-flex');
            expect(mockElements.stopButton.disabled).toBe(false);
            expect(mockElements.stopButton.style.display).toBe('inline-flex');
        });

        test('should show resume and stop buttons when paused', () => {
            tracker.state.status = 'paused';
            tracker.updateControlButtons();

            expect(mockElements.resumeButton.disabled).toBe(false);
            expect(mockElements.resumeButton.style.display).toBe('inline-flex');
            expect(mockElements.stopButton.disabled).toBe(false);
            expect(mockElements.stopButton.style.display).toBe('inline-flex');
        });

        test('should show export button when completed', () => {
            tracker.state.status = 'completed';
            tracker.updateControlButtons();

            expect(mockElements.exportButton.disabled).toBe(false);
            expect(mockElements.exportButton.style.display).toBe('inline-flex');
        });

        test('should respect pause/resume disable option', () => {
            tracker.options.enablePauseResume = false;
            tracker.state.status = 'running';
            tracker.updateControlButtons();

            expect(mockElements.pauseButton.style.display).toBe('none');
        });
    });

    describe('Progress Monitoring', () => {
        beforeEach(async () => {
            await tracker.startOperation({ totalItems: 100 });
        });

        test('should start progress monitoring', () => {
            expect(tracker.updateTimer).toBeDefined();
        });

        test('should stop progress monitoring', () => {
            tracker.stopProgressMonitoring();
            expect(tracker.updateTimer).toBeNull();
        });

        test('should fetch progress updates', async () => {
            const mockResponse = {
                ok: true,
                json: jest.fn().mockResolvedValue({
                    completed: 10,
                    failed: 2,
                    currentItem: 'Test Item'
                })
            };
            fetch.mockResolvedValue(mockResponse);

            await tracker.fetchProgressUpdate();

            expect(fetch).toHaveBeenCalledWith(`/api/bulk-operations/${tracker.state.operationId}/progress`);
            expect(tracker.state.completedItems).toBe(10);
            expect(tracker.state.failedItems).toBe(2);
        });

        test('should handle fetch errors with retry', async () => {
            fetch.mockRejectedValue(new Error('Network error'));

            // First attempt
            await tracker.fetchProgressUpdate();
            expect(tracker.retryCount).toBe(1);

            // Second attempt
            await tracker.fetchProgressUpdate();
            expect(tracker.retryCount).toBe(2);

            // Third attempt
            await tracker.fetchProgressUpdate();
            expect(tracker.retryCount).toBe(3);

            // Fourth attempt should fail the operation
            await tracker.fetchProgressUpdate();
            expect(tracker.state.status).toBe('failed');
        });

        test('should reset retry count on successful fetch', async () => {
            tracker.retryCount = 2;
            
            const mockResponse = {
                ok: true,
                json: jest.fn().mockResolvedValue({ completed: 5, failed: 0 })
            };
            fetch.mockResolvedValue(mockResponse);

            await tracker.fetchProgressUpdate();

            expect(tracker.retryCount).toBe(0);
        });
    });

    describe('Export Functionality', () => {
        test('should export results successfully', async () => {
            tracker.state.operationId = 'test-op-123';
            
            const mockBlob = new Blob(['test data'], { type: 'text/csv' });
            const mockResponse = {
                ok: true,
                blob: jest.fn().mockResolvedValue(mockBlob)
            };
            fetch.mockResolvedValue(mockResponse);

            await tracker.exportResults();

            expect(fetch).toHaveBeenCalledWith('/api/bulk-operations/test-op-123/export');
            expect(document.createElement).toHaveBeenCalledWith('a');
        });

        test('should handle export errors', async () => {
            tracker.state.operationId = 'test-op-123';
            fetch.mockRejectedValue(new Error('Export failed'));

            await tracker.exportResults();

            expect(console.error).toHaveBeenCalled();
        });

        test('should not export without operation ID', async () => {
            await tracker.exportResults();

            expect(fetch).not.toHaveBeenCalled();
        });
    });

    describe('Event System', () => {
        test('should add event listeners', () => {
            const callback = jest.fn();
            tracker.addEventListener('testEvent', callback);

            expect(tracker.eventListeners.has('testEvent')).toBe(true);
            expect(tracker.eventListeners.get('testEvent')).toContain(callback);
        });

        test('should remove event listeners', () => {
            const callback = jest.fn();
            tracker.addEventListener('testEvent', callback);
            tracker.removeEventListener('testEvent', callback);

            expect(tracker.eventListeners.get('testEvent')).not.toContain(callback);
        });

        test('should emit events', () => {
            const callback = jest.fn();
            tracker.addEventListener('testEvent', callback);

            tracker.emit('testEvent', { test: 'data' });

            expect(callback).toHaveBeenCalledWith({ test: 'data' });
        });

        test('should handle event listener errors', () => {
            const errorCallback = jest.fn().mockImplementation(() => {
                throw new Error('Event error');
            });
            tracker.addEventListener('testEvent', errorCallback);

            tracker.emit('testEvent', {});

            expect(console.error).toHaveBeenCalled();
        });
    });

    describe('Utility Methods', () => {
        test('should generate unique operation ID', () => {
            const id1 = tracker.generateOperationId();
            const id2 = tracker.generateOperationId();

            expect(id1).toMatch(/^bulk-op-\d+-[a-z0-9]+$/);
            expect(id2).toMatch(/^bulk-op-\d+-[a-z0-9]+$/);
            expect(id1).not.toBe(id2);
        });

        test('should format duration correctly', () => {
            expect(tracker.formatDuration(1000)).toBe('1s');
            expect(tracker.formatDuration(61000)).toBe('1m 1s');
            expect(tracker.formatDuration(3661000)).toBe('1h 1m 1s');
        });

        test('should reset state', () => {
            tracker.state.operationId = 'test';
            tracker.state.progress = 50;
            tracker.state.status = 'running';

            tracker.resetState();

            expect(tracker.state.operationId).toBeNull();
            expect(tracker.state.progress).toBe(0);
            expect(tracker.state.status).toBe('ready');
        });

        test('should get current state', () => {
            tracker.state.progress = 75;
            const state = tracker.getState();

            expect(state.progress).toBe(75);
            expect(state).not.toBe(tracker.state); // Should be a copy
        });

        test('should update options', () => {
            tracker.updateOptions({ updateInterval: 500 });

            expect(tracker.options.updateInterval).toBe(500);
            expect(tracker.options.maxRetries).toBe(3); // Should preserve other options
        });
    });

    describe('Logging', () => {
        test('should add log entries', () => {
            tracker.log('Test message', 'info');

            expect(mockElements.operationLog.appendChild).toHaveBeenCalled();
        });

        test('should respect showDetailedLog option', () => {
            tracker.options.showDetailedLog = false;
            tracker.log('Test message', 'info');

            expect(mockElements.operationLog.appendChild).not.toHaveBeenCalled();
        });

        test('should handle missing operation log element', () => {
            tracker.elements.operationLog = null;
            
            expect(() => {
                tracker.log('Test message', 'info');
            }).not.toThrow();
        });
    });

    describe('Error Handling', () => {
        test('should handle errors gracefully', () => {
            const error = new Error('Test error');
            tracker.handleError('Test message', error);

            expect(console.error).toHaveBeenCalledWith('Test message', error);
        });

        test('should handle operation errors', () => {
            const error = new Error('Operation failed');
            tracker.handleOperationError(error);

            expect(tracker.state.status).toBe('failed');
            expect(tracker.state.errors).toContain(error);
        });
    });

    describe('Cleanup', () => {
        test('should destroy component properly', () => {
            tracker.destroy();

            expect(tracker.updateTimer).toBeNull();
            expect(tracker.eventListeners.size).toBe(0);
        });
    });
});
