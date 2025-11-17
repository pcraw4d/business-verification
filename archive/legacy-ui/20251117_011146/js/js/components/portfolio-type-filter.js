/**
 * Portfolio Type Filter Component
 * Provides filtering functionality for merchant portfolio types with visual indicators
 * Supports multiple selection modes and integrates with other components
 */
class PortfolioTypeFilter {
    constructor(options = {}) {
        this.container = options.container || document.body;
        this.mode = options.mode || 'single'; // 'single' or 'multiple'
        this.allowClear = options.allowClear !== false;
        this.showCounts = options.showCounts !== false;
        this.allowAll = options.allowAll !== false;
        this.initialValue = options.initialValue || null;
        this.disabled = options.disabled || false;
        
        // Portfolio type definitions with visual indicators
        this.portfolioTypes = {
            'onboarded': {
                label: 'Onboarded',
                description: 'Fully verified and active merchants',
                icon: 'fas fa-check-circle',
                color: '#27ae60',
                bgColor: '#d5f4e6',
                borderColor: '#27ae60',
                count: 0
            },
            'deactivated': {
                label: 'Deactivated',
                description: 'Merchants that have been deactivated',
                icon: 'fas fa-times-circle',
                color: '#e74c3c',
                bgColor: '#fadbd8',
                borderColor: '#e74c3c',
                count: 0
            },
            'prospective': {
                label: 'Prospective',
                description: 'Potential merchants under evaluation',
                icon: 'fas fa-eye',
                color: '#f39c12',
                bgColor: '#fef9e7',
                borderColor: '#f39c12',
                count: 0
            },
            'pending': {
                label: 'Pending',
                description: 'Merchants awaiting verification',
                icon: 'fas fa-clock',
                color: '#3498db',
                bgColor: '#ebf3fd',
                borderColor: '#3498db',
                count: 0
            }
        };
        
        // Current selection state
        this.selectedValues = this.mode === 'multiple' ? [] : null;
        this.isExpanded = false;
        
        // Event callbacks
        this.onChange = options.onChange || null;
        this.onFilter = options.onFilter || null;
        this.onClear = options.onClear || null;
        
        // API integration
        this.apiBaseUrl = options.apiBaseUrl || '/api/v1';
        this.updateCountsOnInit = options.updateCountsOnInit !== false;
        
        this.init();
    }

    init() {
        this.createFilterInterface();
        this.bindEvents();
        this.setInitialValue();
        
        if (this.updateCountsOnInit) {
            this.updateCounts();
        }
    }

    createFilterInterface() {
        const filterHTML = `
            <div class="portfolio-type-filter-container">
                <div class="filter-header">
                    <label class="filter-label">
                        <i class="fas fa-folder"></i>
                        Portfolio Type
                        ${this.mode === 'multiple' ? '<span class="multi-indicator">(Multiple)</span>' : ''}
                    </label>
                    ${this.allowClear && this.hasSelection() ? `
                        <button class="clear-filter-btn" id="clearPortfolioFilter" title="Clear selection">
                            <i class="fas fa-times"></i>
                        </button>
                    ` : ''}
                </div>

                <div class="filter-dropdown" id="portfolioFilterDropdown">
                    <button class="filter-trigger" id="portfolioFilterTrigger" ${this.disabled ? 'disabled' : ''}>
                        <div class="trigger-content">
                            <span class="trigger-text" id="portfolioFilterText">
                                ${this.getTriggerText()}
                            </span>
                            <div class="trigger-selection" id="portfolioFilterSelection">
                                ${this.renderSelectionPreview()}
                            </div>
                        </div>
                        <i class="fas fa-chevron-down trigger-icon"></i>
                    </button>

                    <div class="filter-dropdown-menu" id="portfolioFilterMenu" style="display: none;">
                        ${this.allowAll ? `
                            <div class="filter-option all-option" data-value="">
                                <div class="option-content">
                                    <div class="option-indicator">
                                        <i class="fas fa-globe"></i>
                                    </div>
                                    <div class="option-info">
                                        <span class="option-label">All Types</span>
                                        <span class="option-description">Show all portfolio types</span>
                                    </div>
                                    ${this.showCounts ? `<span class="option-count" id="allCount">-</span>` : ''}
                                </div>
                                <div class="option-checkbox">
                                    <i class="fas fa-check" style="display: none;"></i>
                                </div>
                            </div>
                        ` : ''}

                        ${Object.entries(this.portfolioTypes).map(([key, type]) => `
                            <div class="filter-option" data-value="${key}">
                                <div class="option-content">
                                    <div class="option-indicator" style="background-color: ${type.bgColor}; border-color: ${type.borderColor};">
                                        <i class="${type.icon}" style="color: ${type.color};"></i>
                                    </div>
                                    <div class="option-info">
                                        <span class="option-label">${type.label}</span>
                                        <span class="option-description">${type.description}</span>
                                    </div>
                                    ${this.showCounts ? `<span class="option-count" id="count-${key}">${type.count}</span>` : ''}
                                </div>
                                <div class="option-checkbox">
                                    <i class="fas fa-check" style="display: none;"></i>
                                </div>
                            </div>
                        `).join('')}

                        ${this.mode === 'multiple' ? `
                            <div class="filter-actions">
                                <button class="btn btn-secondary btn-sm" id="selectAllPortfolioTypes">
                                    <i class="fas fa-check-double"></i>
                                    Select All
                                </button>
                                <button class="btn btn-outline btn-sm" id="clearAllPortfolioTypes">
                                    <i class="fas fa-eraser"></i>
                                    Clear All
                                </button>
                            </div>
                        ` : ''}
                    </div>
                </div>

                ${this.mode === 'multiple' && this.hasSelection() ? `
                    <div class="selected-tags" id="portfolioSelectedTags">
                        ${this.renderSelectedTags()}
                    </div>
                ` : ''}
            </div>
        `;

        this.container.innerHTML = filterHTML;
        this.addStyles();
    }

    addStyles() {
        const styles = `
            <style>
                .portfolio-type-filter-container {
                    position: relative;
                    width: 100%;
                    max-width: 400px;
                }

                .filter-header {
                    display: flex;
                    justify-content: space-between;
                    align-items: center;
                    margin-bottom: 8px;
                }

                .filter-label {
                    display: flex;
                    align-items: center;
                    gap: 8px;
                    font-weight: 600;
                    color: #495057;
                    font-size: 0.9rem;
                    margin: 0;
                }

                .filter-label i {
                    color: #6c757d;
                    font-size: 0.8rem;
                }

                .multi-indicator {
                    font-size: 0.8rem;
                    color: #6c757d;
                    font-weight: 400;
                    font-style: italic;
                }

                .clear-filter-btn {
                    background: none;
                    border: none;
                    color: #6c757d;
                    cursor: pointer;
                    padding: 4px;
                    border-radius: 4px;
                    transition: all 0.2s ease;
                    font-size: 0.8rem;
                }

                .clear-filter-btn:hover {
                    background: #e9ecef;
                    color: #495057;
                }

                .filter-dropdown {
                    position: relative;
                }

                .filter-trigger {
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

                .filter-trigger:hover:not(:disabled) {
                    border-color: #3498db;
                    box-shadow: 0 0 0 3px rgba(52, 152, 219, 0.1);
                }

                .filter-trigger:focus {
                    outline: none;
                    border-color: #3498db;
                    box-shadow: 0 0 0 3px rgba(52, 152, 219, 0.1);
                }

                .filter-trigger:disabled {
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
                }

                .trigger-icon {
                    color: #6c757d;
                    font-size: 0.8rem;
                    transition: transform 0.3s ease;
                }

                .filter-dropdown.open .trigger-icon {
                    transform: rotate(180deg);
                }

                .filter-dropdown-menu {
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

                .filter-option {
                    display: flex;
                    align-items: center;
                    padding: 12px 16px;
                    cursor: pointer;
                    transition: all 0.2s ease;
                    border-bottom: 1px solid #f8f9fa;
                }

                .filter-option:last-child {
                    border-bottom: none;
                }

                .filter-option:hover {
                    background: #f8f9fa;
                }

                .filter-option.selected {
                    background: #e3f2fd;
                }

                .filter-option.all-option {
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

                .filter-option.selected .option-checkbox {
                    background: #3498db;
                    border-color: #3498db;
                    color: white;
                }

                .filter-option.selected .option-checkbox i {
                    display: block !important;
                    font-size: 0.7rem;
                }

                .filter-actions {
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

                .selected-tags {
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

                /* Responsive Design */
                @media (max-width: 768px) {
                    .portfolio-type-filter-container {
                        max-width: 100%;
                    }

                    .filter-dropdown-menu {
                        max-height: 250px;
                    }

                    .option-content {
                        gap: 8px;
                    }

                    .option-indicator {
                        width: 28px;
                        height: 28px;
                    }

                    .filter-actions {
                        flex-direction: column;
                    }
                }

                /* Animation for dropdown */
                .filter-dropdown-menu {
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
                .filter-loading {
                    opacity: 0.6;
                    pointer-events: none;
                }

                .filter-loading::after {
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
            </style>
        `;

        // Add styles to head if not already added
        if (!document.querySelector('#portfolio-type-filter-styles')) {
            const styleElement = document.createElement('style');
            styleElement.id = 'portfolio-type-filter-styles';
            styleElement.textContent = styles;
            document.head.appendChild(styleElement);
        }
    }

    bindEvents() {
        const trigger = document.getElementById('portfolioFilterTrigger');
        const menu = document.getElementById('portfolioFilterMenu');
        const clearBtn = document.getElementById('clearPortfolioFilter');
        const selectAllBtn = document.getElementById('selectAllPortfolioTypes');
        const clearAllBtn = document.getElementById('clearAllPortfolioTypes');

        // Toggle dropdown
        trigger.addEventListener('click', (e) => {
            e.stopPropagation();
            this.toggleDropdown();
        });

        // Close dropdown when clicking outside
        document.addEventListener('click', (e) => {
            if (!this.container.contains(e.target)) {
                this.closeDropdown();
            }
        });

        // Clear filter
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
        menu.addEventListener('click', (e) => {
            const option = e.target.closest('.filter-option');
            if (option) {
                e.stopPropagation();
                const value = option.dataset.value;
                this.selectOption(value);
            }
        });

        // Keyboard navigation
        trigger.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                this.toggleDropdown();
            } else if (e.key === 'Escape') {
                this.closeDropdown();
            }
        });

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

    toggleDropdown() {
        if (this.disabled) return;
        
        this.isExpanded = !this.isExpanded;
        const dropdown = document.getElementById('portfolioFilterDropdown');
        const menu = document.getElementById('portfolioFilterMenu');
        
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
        const dropdown = document.getElementById('portfolioFilterDropdown');
        const menu = document.getElementById('portfolioFilterMenu');
        
        dropdown.classList.remove('open');
        menu.style.display = 'none';
    }

    selectOption(value) {
        if (this.mode === 'multiple') {
            this.toggleMultipleSelection(value);
        } else {
            this.setSingleSelection(value);
        }
        
        this.updateUI();
        this.triggerChange();
        
        if (this.mode === 'single') {
            this.closeDropdown();
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
            this.selectedValues = Object.keys(this.portfolioTypes);
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
        this.updateTriggerText();
        this.updateSelectedTags();
        this.updateClearButton();
        this.updateOptionStates();
    }

    updateTriggerText() {
        const triggerText = document.getElementById('portfolioFilterText');
        const triggerSelection = document.getElementById('portfolioFilterSelection');
        
        if (this.hasSelection()) {
            if (this.mode === 'single') {
                const selectedType = this.portfolioTypes[this.selectedValues];
                triggerText.textContent = selectedType ? selectedType.label : 'All Types';
                triggerSelection.innerHTML = '';
            } else {
                triggerText.textContent = `${this.selectedValues.length} selected`;
                triggerSelection.innerHTML = this.renderSelectionPreview();
            }
        } else {
            triggerText.textContent = this.allowAll ? 'All Types' : 'Select portfolio type';
            triggerSelection.innerHTML = '';
        }
    }

    updateSelectedTags() {
        const selectedTags = document.getElementById('portfolioSelectedTags');
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
        const clearBtn = document.getElementById('clearPortfolioFilter');
        if (clearBtn) {
            clearBtn.style.display = this.hasSelection() ? 'block' : 'none';
        }
    }

    updateOptionStates() {
        const options = this.container.querySelectorAll('.filter-option');
        options.forEach(option => {
            const value = option.dataset.value;
            const isSelected = this.isSelected(value);
            option.classList.toggle('selected', isSelected);
        });
    }

    renderSelectionPreview() {
        if (this.mode === 'multiple' && this.selectedValues && this.selectedValues.length > 0) {
            return this.selectedValues.slice(0, 2).map(value => {
                const type = this.portfolioTypes[value];
                return `
                    <span class="selection-tag" style="background-color: ${type.bgColor}; color: ${type.color};">
                        <i class="${type.icon}"></i>
                        ${type.label}
                    </span>
                `;
            }).join('') + (this.selectedValues.length > 2 ? `<span class="selection-tag" style="background-color: #e9ecef; color: #6c757d;">+${this.selectedValues.length - 2}</span>` : '');
        }
        return '';
    }

    renderSelectedTags() {
        if (this.mode === 'multiple' && this.selectedValues) {
            return this.selectedValues.map(value => {
                const type = this.portfolioTypes[value];
                return `
                    <span class="selected-tag" data-value="${value}" style="background-color: ${type.bgColor}; color: ${type.color};">
                        <i class="${type.icon}"></i>
                        ${type.label}
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
                const selectedType = this.portfolioTypes[this.selectedValues];
                return selectedType ? selectedType.label : 'All Types';
            } else {
                return `${this.selectedValues.length} selected`;
            }
        }
        return this.allowAll ? 'All Types' : 'Select portfolio type';
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
            const response = await fetch(`${this.apiBaseUrl}/merchants/counts`, {
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
            console.warn('Failed to update portfolio type counts:', error);
        }
    }

    updateCountsFromData(data) {
        if (data.portfolio_types) {
            Object.keys(this.portfolioTypes).forEach(key => {
                if (data.portfolio_types[key] !== undefined) {
                    this.portfolioTypes[key].count = data.portfolio_types[key];
                }
            });
            
            // Update UI if counts are shown
            if (this.showCounts) {
                Object.keys(this.portfolioTypes).forEach(key => {
                    const countElement = document.getElementById(`count-${key}`);
                    if (countElement) {
                        countElement.textContent = this.portfolioTypes[key].count;
                    }
                });
                
                const allCountElement = document.getElementById('allCount');
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
                portfolio_type: value,
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
        const trigger = document.getElementById('portfolioFilterTrigger');
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

    destroy() {
        // Clean up event listeners and DOM elements
        if (this.container) {
            this.container.innerHTML = '';
        }
    }
}

// Export for use in other scripts
if (typeof module !== 'undefined' && module.exports) {
    module.exports = PortfolioTypeFilter;
}

// Auto-initialize if container is found
document.addEventListener('DOMContentLoaded', () => {
    const filterContainer = document.getElementById('portfolioTypeFilterContainer');
    if (filterContainer && !window.portfolioTypeFilter) {
        window.portfolioTypeFilter = new PortfolioTypeFilter({
            container: filterContainer
        });
    }
});
