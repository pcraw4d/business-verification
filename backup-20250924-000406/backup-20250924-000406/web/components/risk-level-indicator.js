/**
 * Risk Level Indicator Component
 * Provides risk level visualization with color coding and icons
 * Supports filtering and integrates with other merchant components
 */
class RiskLevelIndicator {
    constructor(options = {}) {
        this.container = options.container || document.body;
        this.mode = options.mode || 'single'; // 'single' or 'multiple'
        this.allowClear = options.allowClear !== false;
        this.showCounts = options.showCounts !== false;
        this.allowAll = options.allowAll !== false;
        this.initialValue = options.initialValue || null;
        this.disabled = options.disabled || false;
        this.showDescriptions = options.showDescriptions !== false;
        this.compactMode = options.compactMode || false;
        
        // Risk level definitions with visual indicators
        this.riskLevels = {
            'low': {
                label: 'Low Risk',
                description: 'Minimal risk factors identified',
                icon: 'fas fa-shield-alt',
                color: '#27ae60',
                bgColor: '#d5f4e6',
                borderColor: '#27ae60',
                gradient: 'linear-gradient(135deg, #27ae60, #2ecc71)',
                count: 0,
                priority: 1
            },
            'medium': {
                label: 'Medium Risk',
                description: 'Moderate risk factors present',
                icon: 'fas fa-exclamation-triangle',
                color: '#f39c12',
                bgColor: '#fef9e7',
                borderColor: '#f39c12',
                gradient: 'linear-gradient(135deg, #f39c12, #e67e22)',
                count: 0,
                priority: 2
            },
            'high': {
                label: 'High Risk',
                description: 'Significant risk factors identified',
                icon: 'fas fa-exclamation-circle',
                color: '#e74c3c',
                bgColor: '#fadbd8',
                borderColor: '#e74c3c',
                gradient: 'linear-gradient(135deg, #e74c3c, #c0392b)',
                count: 0,
                priority: 3
            }
        };
        
        // Current selection state
        this.selectedValues = this.mode === 'multiple' ? [] : null;
        this.isExpanded = false;
        
        // Event callbacks
        this.onChange = options.onChange || null;
        this.onFilter = options.onFilter || null;
        this.onClear = options.onClear || null;
        this.onRiskLevelClick = options.onRiskLevelClick || null;
        
        // API integration
        this.apiBaseUrl = options.apiBaseUrl || '/api/v1';
        this.updateCountsOnInit = options.updateCountsOnInit !== false;
        
        this.init();
    }

    init() {
        this.createIndicatorInterface();
        this.bindEvents();
        this.setInitialValue();
        
        if (this.updateCountsOnInit) {
            this.updateCounts();
        }
    }

    createIndicatorInterface() {
        const indicatorHTML = `
            <div class="risk-level-indicator-container ${this.compactMode ? 'compact' : ''}">
                <div class="indicator-header">
                    <label class="indicator-label">
                        <i class="fas fa-shield-alt"></i>
                        Risk Level
                        ${this.mode === 'multiple' ? '<span class="multi-indicator">(Multiple)</span>' : ''}
                    </label>
                    ${this.allowClear && this.hasSelection() ? `
                        <button class="clear-indicator-btn" id="clearRiskIndicator" title="Clear selection">
                            <i class="fas fa-times"></i>
                        </button>
                    ` : ''}
                </div>

                ${this.compactMode ? this.createCompactView() : this.createFullView()}
            </div>
        `;

        this.container.innerHTML = indicatorHTML;
        this.addStyles();
    }

    createFullView() {
        return `
            <div class="risk-indicator-dropdown" id="riskIndicatorDropdown">
                <button class="indicator-trigger" id="riskIndicatorTrigger" ${this.disabled ? 'disabled' : ''}>
                    <div class="trigger-content">
                        <span class="trigger-text" id="riskIndicatorText">
                            ${this.getTriggerText()}
                        </span>
                        <div class="trigger-selection" id="riskIndicatorSelection">
                            ${this.renderSelectionPreview()}
                        </div>
                    </div>
                    <i class="fas fa-chevron-down trigger-icon"></i>
                </button>

                <div class="indicator-dropdown-menu" id="riskIndicatorMenu" style="display: none;">
                    ${this.allowAll ? `
                        <div class="indicator-option all-option" data-value="">
                            <div class="option-content">
                                <div class="option-indicator">
                                    <i class="fas fa-globe"></i>
                                </div>
                                <div class="option-info">
                                    <span class="option-label">All Risk Levels</span>
                                    <span class="option-description">Show all risk levels</span>
                                </div>
                                ${this.showCounts ? `<span class="option-count" id="allRiskCount">-</span>` : ''}
                            </div>
                            <div class="option-checkbox">
                                <i class="fas fa-check" style="display: none;"></i>
                            </div>
                        </div>
                    ` : ''}

                    ${Object.entries(this.riskLevels).map(([key, level]) => `
                        <div class="indicator-option" data-value="${key}">
                            <div class="option-content">
                                <div class="option-indicator" style="background: ${level.gradient}; border-color: ${level.borderColor};">
                                    <i class="${level.icon}" style="color: white;"></i>
                                </div>
                                <div class="option-info">
                                    <span class="option-label">${level.label}</span>
                                    ${this.showDescriptions ? `<span class="option-description">${level.description}</span>` : ''}
                                </div>
                                ${this.showCounts ? `<span class="option-count" id="riskCount-${key}">${level.count}</span>` : ''}
                            </div>
                            <div class="option-checkbox">
                                <i class="fas fa-check" style="display: none;"></i>
                            </div>
                        </div>
                    `).join('')}

                    ${this.mode === 'multiple' ? `
                        <div class="indicator-actions">
                            <button class="btn btn-secondary btn-sm" id="selectAllRiskLevels">
                                <i class="fas fa-check-double"></i>
                                Select All
                            </button>
                            <button class="btn btn-outline btn-sm" id="clearAllRiskLevels">
                                <i class="fas fa-eraser"></i>
                                Clear All
                            </button>
                        </div>
                    ` : ''}
                </div>
            </div>

            ${this.mode === 'multiple' && this.hasSelection() ? `
                <div class="selected-risk-tags" id="riskSelectedTags">
                    ${this.renderSelectedTags()}
                </div>
            ` : ''}
        `;
    }

    createCompactView() {
        return `
            <div class="risk-level-badges">
                ${Object.entries(this.riskLevels).map(([key, level]) => `
                    <div class="risk-badge ${this.isSelected(key) ? 'selected' : ''}" 
                         data-value="${key}" 
                         style="background: ${level.gradient}; border-color: ${level.borderColor};">
                        <i class="${level.icon}"></i>
                        <span class="badge-label">${level.label}</span>
                        ${this.showCounts ? `<span class="badge-count">${level.count}</span>` : ''}
                    </div>
                `).join('')}
            </div>
        `;
    }

    addStyles() {
        const styles = `
            <style>
                .risk-level-indicator-container {
                    position: relative;
                    width: 100%;
                    max-width: 400px;
                }

                .risk-level-indicator-container.compact {
                    max-width: 100%;
                }

                .indicator-header {
                    display: flex;
                    justify-content: space-between;
                    align-items: center;
                    margin-bottom: 8px;
                }

                .indicator-label {
                    display: flex;
                    align-items: center;
                    gap: 8px;
                    font-weight: 600;
                    color: #495057;
                    font-size: 0.9rem;
                    margin: 0;
                }

                .indicator-label i {
                    color: #6c757d;
                    font-size: 0.8rem;
                }

                .multi-indicator {
                    font-size: 0.8rem;
                    color: #6c757d;
                    font-weight: 400;
                    font-style: italic;
                }

                .clear-indicator-btn {
                    background: none;
                    border: none;
                    color: #6c757d;
                    cursor: pointer;
                    padding: 4px;
                    border-radius: 4px;
                    transition: all 0.2s ease;
                    font-size: 0.8rem;
                }

                .clear-indicator-btn:hover {
                    background: #e9ecef;
                    color: #495057;
                }

                .risk-indicator-dropdown {
                    position: relative;
                }

                .indicator-trigger {
                    width: 100%;
                    padding: 12px 16px;
                    border: 2px solid #e9ecef;
                    border-radius: 8px;
                    background: white;
                    cursor: pointer;
                    transition: all 0.3s ease;
                    display: flex;
                    justify-content: space-between;
                    align-items: center;
                    text-align: left;
                }

                .indicator-trigger:hover:not(:disabled) {
                    border-color: #3498db;
                    box-shadow: 0 0 0 3px rgba(52, 152, 219, 0.1);
                }

                .indicator-trigger:focus {
                    outline: none;
                    border-color: #3498db;
                    box-shadow: 0 0 0 3px rgba(52, 152, 219, 0.1);
                }

                .indicator-trigger:disabled {
                    background: #f8f9fa;
                    color: #6c757d;
                    cursor: not-allowed;
                }

                .trigger-content {
                    flex: 1;
                    display: flex;
                    flex-direction: column;
                    gap: 4px;
                }

                .trigger-text {
                    font-size: 0.9rem;
                    color: #495057;
                    font-weight: 500;
                }

                .trigger-selection {
                    display: flex;
                    flex-wrap: wrap;
                    gap: 4px;
                }

                .selection-tag {
                    display: inline-flex;
                    align-items: center;
                    gap: 4px;
                    padding: 2px 8px;
                    border-radius: 12px;
                    font-size: 0.8rem;
                    font-weight: 500;
                    color: white;
                }

                .trigger-icon {
                    color: #6c757d;
                    font-size: 0.8rem;
                    transition: transform 0.3s ease;
                }

                .risk-indicator-dropdown.open .trigger-icon {
                    transform: rotate(180deg);
                }

                .indicator-dropdown-menu {
                    position: absolute;
                    top: 100%;
                    left: 0;
                    right: 0;
                    background: white;
                    border: 2px solid #e9ecef;
                    border-radius: 8px;
                    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
                    z-index: 1000;
                    max-height: 300px;
                    overflow-y: auto;
                    margin-top: 4px;
                }

                .indicator-option {
                    display: flex;
                    align-items: center;
                    padding: 12px 16px;
                    cursor: pointer;
                    transition: all 0.2s ease;
                    border-bottom: 1px solid #f8f9fa;
                }

                .indicator-option:last-child {
                    border-bottom: none;
                }

                .indicator-option:hover {
                    background: #f8f9fa;
                }

                .indicator-option.selected {
                    background: #e3f2fd;
                }

                .indicator-option.all-option {
                    border-bottom: 2px solid #e9ecef;
                    font-weight: 600;
                }

                .option-content {
                    flex: 1;
                    display: flex;
                    align-items: center;
                    gap: 12px;
                }

                .option-indicator {
                    width: 32px;
                    height: 32px;
                    border-radius: 50%;
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    border: 2px solid;
                    flex-shrink: 0;
                }

                .option-indicator i {
                    font-size: 0.9rem;
                }

                .option-info {
                    flex: 1;
                    display: flex;
                    flex-direction: column;
                    gap: 2px;
                }

                .option-label {
                    font-weight: 600;
                    color: #2c3e50;
                    font-size: 0.9rem;
                }

                .option-description {
                    font-size: 0.8rem;
                    color: #6c757d;
                }

                .option-count {
                    background: #e9ecef;
                    color: #495057;
                    padding: 2px 8px;
                    border-radius: 12px;
                    font-size: 0.8rem;
                    font-weight: 600;
                    min-width: 24px;
                    text-align: center;
                }

                .option-checkbox {
                    width: 20px;
                    height: 20px;
                    border: 2px solid #e9ecef;
                    border-radius: 4px;
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    transition: all 0.2s ease;
                    flex-shrink: 0;
                }

                .indicator-option.selected .option-checkbox {
                    background: #3498db;
                    border-color: #3498db;
                    color: white;
                }

                .indicator-option.selected .option-checkbox i {
                    display: block !important;
                    font-size: 0.7rem;
                }

                .indicator-actions {
                    padding: 12px 16px;
                    border-top: 1px solid #e9ecef;
                    display: flex;
                    gap: 8px;
                    background: #f8f9fa;
                }

                .btn {
                    padding: 8px 12px;
                    border: none;
                    border-radius: 6px;
                    font-size: 0.8rem;
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
                    padding: 6px 10px;
                    font-size: 0.75rem;
                }

                .btn-secondary {
                    background: #6c757d;
                    color: white;
                }

                .btn-secondary:hover:not(:disabled) {
                    background: #5a6268;
                }

                .btn-outline {
                    background: transparent;
                    color: #6c757d;
                    border: 1px solid #6c757d;
                }

                .btn-outline:hover:not(:disabled) {
                    background: #6c757d;
                    color: white;
                }

                .selected-risk-tags {
                    margin-top: 8px;
                    display: flex;
                    flex-wrap: wrap;
                    gap: 6px;
                }

                .selected-tag {
                    display: inline-flex;
                    align-items: center;
                    gap: 6px;
                    padding: 6px 12px;
                    border-radius: 16px;
                    font-size: 0.8rem;
                    font-weight: 500;
                    position: relative;
                    color: white;
                }

                .tag-remove {
                    background: none;
                    border: none;
                    color: inherit;
                    cursor: pointer;
                    padding: 2px;
                    border-radius: 50%;
                    transition: all 0.2s ease;
                    opacity: 0.7;
                }

                .tag-remove:hover {
                    opacity: 1;
                    background: rgba(0, 0, 0, 0.1);
                }

                /* Compact View Styles */
                .risk-level-badges {
                    display: flex;
                    gap: 8px;
                    flex-wrap: wrap;
                }

                .risk-badge {
                    display: inline-flex;
                    align-items: center;
                    gap: 6px;
                    padding: 8px 12px;
                    border-radius: 20px;
                    border: 2px solid;
                    cursor: pointer;
                    transition: all 0.3s ease;
                    font-size: 0.8rem;
                    font-weight: 600;
                    color: white;
                    position: relative;
                }

                .risk-badge:hover {
                    transform: translateY(-2px);
                    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
                }

                .risk-badge.selected {
                    transform: scale(1.05);
                    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.3);
                }

                .risk-badge i {
                    font-size: 0.9rem;
                }

                .badge-label {
                    font-weight: 600;
                }

                .badge-count {
                    background: rgba(255, 255, 255, 0.2);
                    padding: 2px 6px;
                    border-radius: 10px;
                    font-size: 0.7rem;
                    font-weight: 700;
                }

                /* Responsive Design */
                @media (max-width: 768px) {
                    .risk-level-indicator-container {
                        max-width: 100%;
                    }

                    .indicator-dropdown-menu {
                        max-height: 250px;
                    }

                    .option-content {
                        gap: 8px;
                    }

                    .option-indicator {
                        width: 28px;
                        height: 28px;
                    }

                    .indicator-actions {
                        flex-direction: column;
                    }

                    .risk-level-badges {
                        flex-direction: column;
                    }

                    .risk-badge {
                        justify-content: center;
                    }
                }

                /* Animation for dropdown */
                .indicator-dropdown-menu {
                    animation: slideDown 0.2s ease-out;
                }

                @keyframes slideDown {
                    from {
                        opacity: 0;
                        transform: translateY(-10px);
                    }
                    to {
                        opacity: 1;
                        transform: translateY(0);
                    }
                }

                /* Loading state */
                .indicator-loading {
                    opacity: 0.6;
                    pointer-events: none;
                }

                .indicator-loading::after {
                    content: '';
                    position: absolute;
                    top: 0;
                    left: 0;
                    right: 0;
                    bottom: 0;
                    background: rgba(255, 255, 255, 0.8);
                    display: flex;
                    align-items: center;
                    justify-content: center;
                }

                /* Risk level specific animations */
                .risk-badge.low-risk {
                    animation: pulseGreen 2s infinite;
                }

                .risk-badge.medium-risk {
                    animation: pulseOrange 2s infinite;
                }

                .risk-badge.high-risk {
                    animation: pulseRed 2s infinite;
                }

                @keyframes pulseGreen {
                    0%, 100% { box-shadow: 0 0 0 0 rgba(39, 174, 96, 0.4); }
                    50% { box-shadow: 0 0 0 8px rgba(39, 174, 96, 0); }
                }

                @keyframes pulseOrange {
                    0%, 100% { box-shadow: 0 0 0 0 rgba(243, 156, 18, 0.4); }
                    50% { box-shadow: 0 0 0 8px rgba(243, 156, 18, 0); }
                }

                @keyframes pulseRed {
                    0%, 100% { box-shadow: 0 0 0 0 rgba(231, 76, 60, 0.4); }
                    50% { box-shadow: 0 0 0 8px rgba(231, 76, 60, 0); }
                }
            </style>
        `;

        // Add styles to head if not already added
        if (!document.querySelector('#risk-level-indicator-styles')) {
            const styleElement = document.createElement('style');
            styleElement.id = 'risk-level-indicator-styles';
            styleElement.textContent = styles;
            document.head.appendChild(styleElement);
        }
    }

    bindEvents() {
        if (this.compactMode) {
            this.bindCompactEvents();
        } else {
            this.bindFullEvents();
        }
    }

    bindFullEvents() {
        const trigger = document.getElementById('riskIndicatorTrigger');
        const menu = document.getElementById('riskIndicatorMenu');
        const clearBtn = document.getElementById('clearRiskIndicator');
        const selectAllBtn = document.getElementById('selectAllRiskLevels');
        const clearAllBtn = document.getElementById('clearAllRiskLevels');

        // Toggle dropdown
        if (trigger) {
            trigger.addEventListener('click', (e) => {
                e.stopPropagation();
                this.toggleDropdown();
            });
        }

        // Close dropdown when clicking outside
        document.addEventListener('click', (e) => {
            if (!this.container.contains(e.target)) {
                this.closeDropdown();
            }
        });

        // Clear indicator
        if (clearBtn) {
            clearBtn.addEventListener('click', (e) => {
                e.stopPropagation();
                this.clearSelection();
            });
        }

        // Select all (multiple mode)
        if (selectAllBtn) {
            selectAllBtn.addEventListener('click', (e) => {
                e.stopPropagation();
                this.selectAll();
            });
        }

        // Clear all (multiple mode)
        if (clearAllBtn) {
            clearAllBtn.addEventListener('click', (e) => {
                e.stopPropagation();
                this.clearAll();
            });
        }

        // Option selection
        if (menu) {
            menu.addEventListener('click', (e) => {
                const option = e.target.closest('.indicator-option');
                if (option) {
                    e.stopPropagation();
                    const value = option.dataset.value;
                    this.selectOption(value);
                }
            });
        }

        // Keyboard navigation
        if (trigger) {
            trigger.addEventListener('keydown', (e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    this.toggleDropdown();
                } else if (e.key === 'Escape') {
                    this.closeDropdown();
                }
            });
        }

        // Tag removal (multiple mode)
        this.container.addEventListener('click', (e) => {
            if (e.target.classList.contains('tag-remove')) {
                e.stopPropagation();
                const tag = e.target.closest('.selected-tag');
                const value = tag.dataset.value;
                this.removeSelection(value);
            }
        });
    }

    bindCompactEvents() {
        const badges = this.container.querySelectorAll('.risk-badge');
        badges.forEach(badge => {
            badge.addEventListener('click', (e) => {
                e.stopPropagation();
                const value = badge.dataset.value;
                this.selectOption(value);
            });
        });
    }

    toggleDropdown() {
        if (this.disabled || this.compactMode) return;
        
        this.isExpanded = !this.isExpanded;
        const dropdown = document.getElementById('riskIndicatorDropdown');
        const menu = document.getElementById('riskIndicatorMenu');
        
        if (this.isExpanded) {
            dropdown.classList.add('open');
            menu.style.display = 'block';
            this.updateOptionStates();
        } else {
            this.closeDropdown();
        }
    }

    closeDropdown() {
        this.isExpanded = false;
        const dropdown = document.getElementById('riskIndicatorDropdown');
        const menu = document.getElementById('riskIndicatorMenu');
        
        if (dropdown) dropdown.classList.remove('open');
        if (menu) menu.style.display = 'none';
    }

    selectOption(value) {
        if (this.mode === 'multiple') {
            this.toggleMultipleSelection(value);
        } else {
            this.setSingleSelection(value);
        }
        
        this.updateUI();
        this.triggerChange();
        
        if (this.mode === 'single' && !this.compactMode) {
            this.closeDropdown();
        }

        // Trigger risk level click callback
        if (this.onRiskLevelClick) {
            const riskLevel = this.riskLevels[value];
            this.onRiskLevelClick({
                value: value,
                level: riskLevel,
                selected: this.isSelected(value)
            });
        }
    }

    setSingleSelection(value) {
        this.selectedValues = value || null;
    }

    toggleMultipleSelection(value) {
        if (!this.selectedValues) {
            this.selectedValues = [];
        }
        
        const index = this.selectedValues.indexOf(value);
        if (index > -1) {
            this.selectedValues.splice(index, 1);
        } else {
            this.selectedValues.push(value);
        }
    }

    selectAll() {
        if (this.mode === 'multiple') {
            this.selectedValues = Object.keys(this.riskLevels);
            this.updateUI();
            this.triggerChange();
        }
    }

    clearAll() {
        this.selectedValues = this.mode === 'multiple' ? [] : null;
        this.updateUI();
        this.triggerChange();
    }

    clearSelection() {
        this.clearAll();
        if (this.onClear) {
            this.onClear();
        }
    }

    removeSelection(value) {
        if (this.mode === 'multiple' && this.selectedValues) {
            const index = this.selectedValues.indexOf(value);
            if (index > -1) {
                this.selectedValues.splice(index, 1);
                this.updateUI();
                this.triggerChange();
            }
        }
    }

    updateUI() {
        if (this.compactMode) {
            this.updateCompactView();
        } else {
            this.updateFullView();
        }
    }

    updateFullView() {
        this.updateTriggerText();
        this.updateSelectedTags();
        this.updateClearButton();
        this.updateOptionStates();
    }

    updateCompactView() {
        const badges = this.container.querySelectorAll('.risk-badge');
        badges.forEach(badge => {
            const value = badge.dataset.value;
            const isSelected = this.isSelected(value);
            badge.classList.toggle('selected', isSelected);
            badge.classList.toggle(`${value}-risk`, isSelected);
        });
    }

    updateTriggerText() {
        const triggerText = document.getElementById('riskIndicatorText');
        const triggerSelection = document.getElementById('riskIndicatorSelection');
        
        if (this.hasSelection()) {
            if (this.mode === 'single') {
                const selectedLevel = this.riskLevels[this.selectedValues];
                triggerText.textContent = selectedLevel ? selectedLevel.label : 'All Risk Levels';
                triggerSelection.innerHTML = '';
            } else {
                triggerText.textContent = `${this.selectedValues.length} selected`;
                triggerSelection.innerHTML = this.renderSelectionPreview();
            }
        } else {
            triggerText.textContent = this.allowAll ? 'All Risk Levels' : 'Select risk level';
            triggerSelection.innerHTML = '';
        }
    }

    updateSelectedTags() {
        const selectedTags = document.getElementById('riskSelectedTags');
        if (selectedTags) {
            if (this.hasSelection() && this.mode === 'multiple') {
                selectedTags.innerHTML = this.renderSelectedTags();
                selectedTags.style.display = 'flex';
            } else {
                selectedTags.style.display = 'none';
            }
        }
    }

    updateClearButton() {
        const clearBtn = document.getElementById('clearRiskIndicator');
        if (clearBtn) {
            clearBtn.style.display = this.hasSelection() ? 'block' : 'none';
        }
    }

    updateOptionStates() {
        const options = this.container.querySelectorAll('.indicator-option');
        options.forEach(option => {
            const value = option.dataset.value;
            const isSelected = this.isSelected(value);
            option.classList.toggle('selected', isSelected);
        });
    }

    renderSelectionPreview() {
        if (this.mode === 'multiple' && this.selectedValues && this.selectedValues.length > 0) {
            return this.selectedValues.slice(0, 2).map(value => {
                const level = this.riskLevels[value];
                return `
                    <span class="selection-tag" style="background: ${level.gradient};">
                        <i class="${level.icon}"></i>
                        ${level.label}
                    </span>
                `;
            }).join('') + (this.selectedValues.length > 2 ? `<span class="selection-tag" style="background: #e9ecef; color: #6c757d;">+${this.selectedValues.length - 2}</span>` : '');
        }
        return '';
    }

    renderSelectedTags() {
        if (this.mode === 'multiple' && this.selectedValues) {
            return this.selectedValues.map(value => {
                const level = this.riskLevels[value];
                return `
                    <span class="selected-tag" data-value="${value}" style="background: ${level.gradient};">
                        <i class="${level.icon}"></i>
                        ${level.label}
                        <button class="tag-remove" type="button">
                            <i class="fas fa-times"></i>
                        </button>
                    </span>
                `;
            }).join('');
        }
        return '';
    }

    getTriggerText() {
        if (this.hasSelection()) {
            if (this.mode === 'single') {
                const selectedLevel = this.riskLevels[this.selectedValues];
                return selectedLevel ? selectedLevel.label : 'All Risk Levels';
            } else {
                return `${this.selectedValues.length} selected`;
            }
        }
        return this.allowAll ? 'All Risk Levels' : 'Select risk level';
    }

    hasSelection() {
        if (this.mode === 'multiple') {
            return this.selectedValues && this.selectedValues.length > 0;
        } else {
            return this.selectedValues !== null && this.selectedValues !== '';
        }
    }

    isSelected(value) {
        if (this.mode === 'multiple') {
            return this.selectedValues && this.selectedValues.includes(value);
        } else {
            return this.selectedValues === value;
        }
    }

    setInitialValue() {
        if (this.initialValue !== null) {
            if (this.mode === 'multiple' && Array.isArray(this.initialValue)) {
                this.selectedValues = [...this.initialValue];
            } else if (this.mode === 'single') {
                this.selectedValues = this.initialValue;
            }
            this.updateUI();
        }
    }

    async updateCounts() {
        try {
            const response = await fetch(`${this.apiBaseUrl}/merchants/risk-counts`, {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${this.getAuthToken()}`
                }
            });

            if (response.ok) {
                const data = await response.json();
                this.updateCountsFromData(data);
            }
        } catch (error) {
            console.warn('Failed to update risk level counts:', error);
        }
    }

    updateCountsFromData(data) {
        if (data.risk_levels) {
            Object.keys(this.riskLevels).forEach(key => {
                if (data.risk_levels[key] !== undefined) {
                    this.riskLevels[key].count = data.risk_levels[key];
                }
            });
            
            // Update UI if counts are shown
            if (this.showCounts) {
                Object.keys(this.riskLevels).forEach(key => {
                    const countElement = document.getElementById(`riskCount-${key}`);
                    if (countElement) {
                        countElement.textContent = this.riskLevels[key].count;
                    }
                });
                
                const allCountElement = document.getElementById('allRiskCount');
                if (allCountElement && data.total !== undefined) {
                    allCountElement.textContent = data.total;
                }
            }
        }
    }

    triggerChange() {
        const value = this.getValue();
        
        if (this.onChange) {
            this.onChange(value);
        }
        
        if (this.onFilter) {
            this.onFilter({
                risk_level: value,
                mode: this.mode
            });
        }
    }

    getValue() {
        return this.selectedValues;
    }

    setValue(value) {
        if (this.mode === 'multiple' && Array.isArray(value)) {
            this.selectedValues = [...value];
        } else if (this.mode === 'single') {
            this.selectedValues = value;
        }
        this.updateUI();
    }

    setDisabled(disabled) {
        this.disabled = disabled;
        const trigger = document.getElementById('riskIndicatorTrigger');
        if (trigger) {
            trigger.disabled = disabled;
        }
    }

    getAuthToken() {
        return localStorage.getItem('auth_token') || 
               document.cookie.split('; ').find(row => row.startsWith('auth_token='))?.split('=')[1] || 
               '';
    }

    // Public methods for external control
    refresh() {
        if (this.updateCountsOnInit) {
            this.updateCounts();
        }
    }

    reset() {
        this.clearAll();
    }

    // Utility methods
    getRiskLevelInfo(value) {
        return this.riskLevels[value] || null;
    }

    getSelectedRiskLevels() {
        if (this.mode === 'multiple' && this.selectedValues) {
            return this.selectedValues.map(value => ({
                value: value,
                level: this.riskLevels[value]
            }));
        } else if (this.mode === 'single' && this.selectedValues) {
            return [{
                value: this.selectedValues,
                level: this.riskLevels[this.selectedValues]
            }];
        }
        return [];
    }

    setCompactMode(compact) {
        this.compactMode = compact;
        this.createIndicatorInterface();
        this.bindEvents();
        this.updateUI();
    }

    destroy() {
        // Clean up event listeners and DOM elements
        if (this.container) {
            this.container.innerHTML = '';
        }
    }
}

// Export for use in other scripts
if (typeof module !== 'undefined' && module.exports) {
    module.exports = RiskLevelIndicator;
}

// Auto-initialize if container is found
document.addEventListener('DOMContentLoaded', () => {
    const indicatorContainer = document.getElementById('riskLevelIndicatorContainer');
    if (indicatorContainer && !window.riskLevelIndicator) {
        window.riskLevelIndicator = new RiskLevelIndicator({
            container: indicatorContainer
        });
    }
});
