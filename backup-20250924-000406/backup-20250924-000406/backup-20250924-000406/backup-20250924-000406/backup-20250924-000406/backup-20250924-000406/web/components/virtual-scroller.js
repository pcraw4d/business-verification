/**
 * Virtual Scroller Component
 * Provides efficient rendering of large merchant lists with virtual scrolling
 * Optimizes performance for thousands of merchants
 */

class VirtualScroller {
    constructor(options = {}) {
        this.options = {
            container: null,
            itemHeight: 80,
            bufferSize: 5,
            threshold: 0.1,
            scrollDebounce: 16, // ~60fps
            ...options
        };

        this.container = this.options.container;
        this.itemHeight = this.options.itemHeight;
        this.bufferSize = this.options.bufferSize;
        
        this.data = [];
        this.filteredData = [];
        this.visibleItems = [];
        this.scrollTop = 0;
        this.containerHeight = 0;
        this.totalHeight = 0;
        this.startIndex = 0;
        this.endIndex = 0;
        this.visibleCount = 0;
        
        this.scrollTimeout = null;
        this.isScrolling = false;
        this.renderTimeout = null;
        
        this.itemRenderer = null;
        this.itemUpdater = null;
        this.onItemClick = null;
        this.onItemSelect = null;
        
        this.init();
    }

    /**
     * Initialize the virtual scroller
     */
    init() {
        if (!this.container) {
            throw new Error('VirtualScroller: Container element is required');
        }

        this.setupContainer();
        this.bindEvents();
        this.updateDimensions();
    }

    /**
     * Setup container element
     */
    setupContainer() {
        this.container.style.position = 'relative';
        this.container.style.overflow = 'auto';
        this.container.style.height = '100%';
        
        // Create viewport
        this.viewport = document.createElement('div');
        this.viewport.style.position = 'relative';
        this.viewport.style.height = '100%';
        this.container.appendChild(this.viewport);
        
        // Create scrollable area
        this.scrollableArea = document.createElement('div');
        this.scrollableArea.style.position = 'absolute';
        this.scrollableArea.style.top = '0';
        this.scrollableArea.style.left = '0';
        this.scrollableArea.style.right = '0';
        this.viewport.appendChild(this.scrollableArea);
        
        // Create items container
        this.itemsContainer = document.createElement('div');
        this.itemsContainer.style.position = 'relative';
        this.scrollableArea.appendChild(this.itemsContainer);
    }

    /**
     * Bind scroll events
     */
    bindEvents() {
        this.container.addEventListener('scroll', this.handleScroll.bind(this), { passive: true });
        window.addEventListener('resize', this.handleResize.bind(this));
        
        // Handle item interactions
        this.itemsContainer.addEventListener('click', this.handleItemClick.bind(this));
        this.itemsContainer.addEventListener('change', this.handleItemChange.bind(this));
    }

    /**
     * Handle scroll events
     */
    handleScroll() {
        if (this.scrollTimeout) {
            clearTimeout(this.scrollTimeout);
        }
        
        this.scrollTimeout = setTimeout(() => {
            this.updateScrollPosition();
        }, this.options.scrollDebounce);
    }

    /**
     * Handle resize events
     */
    handleResize() {
        this.updateDimensions();
        this.render();
    }

    /**
     * Handle item click events
     */
    handleItemClick(event) {
        const itemElement = event.target.closest('[data-item-index]');
        if (itemElement && this.onItemClick) {
            const index = parseInt(itemElement.dataset.itemIndex);
            const item = this.filteredData[index];
            this.onItemClick(item, index, event);
        }
    }

    /**
     * Handle item change events (for checkboxes, etc.)
     */
    handleItemChange(event) {
        const itemElement = event.target.closest('[data-item-index]');
        if (itemElement && this.onItemSelect) {
            const index = parseInt(itemElement.dataset.itemIndex);
            const item = this.filteredData[index];
            this.onItemSelect(item, index, event);
        }
    }

    /**
     * Update scroll position and render visible items
     */
    updateScrollPosition() {
        this.scrollTop = this.container.scrollTop;
        this.calculateVisibleRange();
        this.render();
    }

    /**
     * Update container dimensions
     */
    updateDimensions() {
        const rect = this.container.getBoundingClientRect();
        this.containerHeight = rect.height;
        this.visibleCount = Math.ceil(this.containerHeight / this.itemHeight) + this.bufferSize * 2;
        this.totalHeight = this.filteredData.length * this.itemHeight;
        
        this.scrollableArea.style.height = `${this.totalHeight}px`;
        this.calculateVisibleRange();
    }

    /**
     * Calculate visible item range
     */
    calculateVisibleRange() {
        this.startIndex = Math.max(0, Math.floor(this.scrollTop / this.itemHeight) - this.bufferSize);
        this.endIndex = Math.min(
            this.filteredData.length - 1,
            this.startIndex + this.visibleCount
        );
    }

    /**
     * Set data for the virtual scroller
     * @param {Array} data - Array of items to display
     */
    setData(data) {
        this.data = data || [];
        this.filteredData = [...this.data];
        this.updateDimensions();
        this.render();
    }

    /**
     * Filter data
     * @param {Function} filterFn - Filter function
     */
    filterData(filterFn) {
        this.filteredData = this.data.filter(filterFn);
        this.updateDimensions();
        this.render();
    }

    /**
     * Sort data
     * @param {Function} sortFn - Sort function
     */
    sortData(sortFn) {
        this.filteredData.sort(sortFn);
        this.render();
    }

    /**
     * Set item renderer function
     * @param {Function} renderer - Function to render individual items
     */
    setItemRenderer(renderer) {
        this.itemRenderer = renderer;
    }

    /**
     * Set item updater function
     * @param {Function} updater - Function to update individual items
     */
    setItemUpdater(updater) {
        this.itemUpdater = updater;
    }

    /**
     * Set item click handler
     * @param {Function} handler - Click handler function
     */
    setItemClickHandler(handler) {
        this.onItemClick = handler;
    }

    /**
     * Set item select handler
     * @param {Function} handler - Select handler function
     */
    setItemSelectHandler(handler) {
        this.onItemSelect = handler;
    }

    /**
     * Render visible items
     */
    render() {
        if (this.renderTimeout) {
            clearTimeout(this.renderTimeout);
        }
        
        this.renderTimeout = setTimeout(() => {
            this.performRender();
        }, 0);
    }

    /**
     * Perform the actual rendering
     */
    performRender() {
        if (!this.itemRenderer) {
            console.warn('VirtualScroller: No item renderer set');
            return;
        }

        // Clear existing items
        this.itemsContainer.innerHTML = '';
        
        // Render visible items
        for (let i = this.startIndex; i <= this.endIndex; i++) {
            if (i >= this.filteredData.length) break;
            
            const item = this.filteredData[i];
            const itemElement = this.createItemElement(item, i);
            this.itemsContainer.appendChild(itemElement);
        }
        
        // Update scrollable area position
        this.scrollableArea.style.transform = `translateY(${this.startIndex * this.itemHeight}px)`;
        
        // Trigger render event
        this.container.dispatchEvent(new CustomEvent('virtualScrollRender', {
            detail: {
                startIndex: this.startIndex,
                endIndex: this.endIndex,
                visibleCount: this.endIndex - this.startIndex + 1,
                totalCount: this.filteredData.length
            }
        }));
    }

    /**
     * Create item element
     * @param {Object} item - Item data
     * @param {number} index - Item index
     */
    createItemElement(item, index) {
        const itemElement = document.createElement('div');
        itemElement.className = 'virtual-scroll-item';
        itemElement.style.position = 'absolute';
        itemElement.style.top = '0';
        itemElement.style.left = '0';
        itemElement.style.right = '0';
        itemElement.style.height = `${this.itemHeight}px`;
        itemElement.dataset.itemIndex = index;
        
        // Render item content
        if (this.itemRenderer) {
            const content = this.itemRenderer(item, index);
            if (typeof content === 'string') {
                itemElement.innerHTML = content;
            } else if (content instanceof HTMLElement) {
                itemElement.appendChild(content);
            }
        }
        
        return itemElement;
    }

    /**
     * Scroll to specific item
     * @param {number} index - Item index
     * @param {boolean} smooth - Smooth scrolling
     */
    scrollToItem(index, smooth = true) {
        const targetScrollTop = index * this.itemHeight;
        
        if (smooth) {
            this.container.scrollTo({
                top: targetScrollTop,
                behavior: 'smooth'
            });
        } else {
            this.container.scrollTop = targetScrollTop;
            this.updateScrollPosition();
        }
    }

    /**
     * Scroll to top
     * @param {boolean} smooth - Smooth scrolling
     */
    scrollToTop(smooth = true) {
        this.scrollToItem(0, smooth);
    }

    /**
     * Scroll to bottom
     * @param {boolean} smooth - Smooth scrolling
     */
    scrollToBottom(smooth = true) {
        this.scrollToItem(this.filteredData.length - 1, smooth);
    }

    /**
     * Get visible items
     */
    getVisibleItems() {
        return this.filteredData.slice(this.startIndex, this.endIndex + 1);
    }

    /**
     * Get scroll statistics
     */
    getStats() {
        return {
            totalItems: this.filteredData.length,
            visibleItems: this.endIndex - this.startIndex + 1,
            startIndex: this.startIndex,
            endIndex: this.endIndex,
            scrollTop: this.scrollTop,
            containerHeight: this.containerHeight,
            totalHeight: this.totalHeight,
            scrollPercentage: this.totalHeight > 0 ? (this.scrollTop / (this.totalHeight - this.containerHeight)) * 100 : 0
        };
    }

    /**
     * Update item at specific index
     * @param {number} index - Item index
     * @param {Object} newData - New item data
     */
    updateItem(index, newData) {
        if (index >= 0 && index < this.filteredData.length) {
            this.filteredData[index] = { ...this.filteredData[index], ...newData };
            
            // Re-render if item is visible
            if (index >= this.startIndex && index <= this.endIndex) {
                this.render();
            }
        }
    }

    /**
     * Add item
     * @param {Object} item - Item to add
     * @param {number} index - Insertion index (optional)
     */
    addItem(item, index = null) {
        if (index === null) {
            this.filteredData.push(item);
        } else {
            this.filteredData.splice(index, 0, item);
        }
        
        this.updateDimensions();
        this.render();
    }

    /**
     * Remove item
     * @param {number} index - Item index to remove
     */
    removeItem(index) {
        if (index >= 0 && index < this.filteredData.length) {
            this.filteredData.splice(index, 1);
            this.updateDimensions();
            this.render();
        }
    }

    /**
     * Clear all items
     */
    clear() {
        this.data = [];
        this.filteredData = [];
        this.updateDimensions();
        this.render();
    }

    /**
     * Destroy the virtual scroller
     */
    destroy() {
        if (this.scrollTimeout) {
            clearTimeout(this.scrollTimeout);
        }
        if (this.renderTimeout) {
            clearTimeout(this.renderTimeout);
        }
        
        this.container.removeEventListener('scroll', this.handleScroll);
        window.removeEventListener('resize', this.handleResize);
        
        if (this.container && this.viewport) {
            this.container.removeChild(this.viewport);
        }
    }
}

/**
 * Merchant Virtual Scroller
 * Specialized virtual scroller for merchant lists
 */
class MerchantVirtualScroller extends VirtualScroller {
    constructor(options = {}) {
        super({
            itemHeight: 100,
            bufferSize: 10,
            ...options
        });
        
        this.selectedItems = new Set();
        this.bulkMode = false;
        this.searchQuery = '';
        this.filters = {
            portfolioType: '',
            riskLevel: '',
            industry: ''
        };
        
        this.initMerchantScroller();
    }

    /**
     * Initialize merchant-specific functionality
     */
    initMerchantScroller() {
        this.setItemRenderer(this.renderMerchantItem.bind(this));
        this.setItemClickHandler(this.handleMerchantClick.bind(this));
        this.setItemSelectHandler(this.handleMerchantSelect.bind(this));
    }

    /**
     * Render merchant item
     * @param {Object} merchant - Merchant data
     * @param {number} index - Item index
     */
    renderMerchantItem(merchant, index) {
        const isSelected = this.selectedItems.has(merchant.id);
        const isVisible = index >= this.startIndex && index <= this.endIndex;
        
        return `
            <div class="merchant-item ${isSelected ? 'selected' : ''} ${merchant.riskLevel}" 
                 data-merchant-id="${merchant.id}">
                <div class="merchant-item-content">
                    ${this.bulkMode ? `
                        <div class="merchant-checkbox">
                            <input type="checkbox" ${isSelected ? 'checked' : ''} 
                                   data-merchant-id="${merchant.id}">
                        </div>
                    ` : ''}
                    
                    <div class="merchant-info">
                        <div class="merchant-header">
                            <h3 class="merchant-name">${merchant.name}</h3>
                            <span class="merchant-risk-level ${merchant.riskLevel}">
                                ${merchant.riskLevel.toUpperCase()}
                            </span>
                        </div>
                        
                        <div class="merchant-details">
                            <div class="merchant-detail">
                                <span class="label">Industry:</span>
                                <span class="value">${merchant.industry}</span>
                            </div>
                            <div class="merchant-detail">
                                <span class="label">Portfolio:</span>
                                <span class="value portfolio-type ${merchant.portfolioType}">
                                    ${merchant.portfolioType}
                                </span>
                            </div>
                            <div class="merchant-detail">
                                <span class="label">Location:</span>
                                <span class="value">${merchant.location}</span>
                            </div>
                        </div>
                    </div>
                    
                    <div class="merchant-actions">
                        <button class="btn btn-sm btn-primary" data-action="view" data-merchant-id="${merchant.id}">
                            View Details
                        </button>
                        <button class="btn btn-sm btn-secondary" data-action="compare" data-merchant-id="${merchant.id}">
                            Compare
                        </button>
                    </div>
                </div>
            </div>
        `;
    }

    /**
     * Handle merchant click
     * @param {Object} merchant - Merchant data
     * @param {number} index - Item index
     * @param {Event} event - Click event
     */
    handleMerchantClick(merchant, index, event) {
        const action = event.target.dataset.action;
        
        if (action === 'view') {
            this.container.dispatchEvent(new CustomEvent('merchantView', {
                detail: { merchant, index }
            }));
        } else if (action === 'compare') {
            this.container.dispatchEvent(new CustomEvent('merchantCompare', {
                detail: { merchant, index }
            }));
        } else {
            // Default click behavior
            this.container.dispatchEvent(new CustomEvent('merchantClick', {
                detail: { merchant, index }
            }));
        }
    }

    /**
     * Handle merchant selection
     * @param {Object} merchant - Merchant data
     * @param {number} index - Item index
     * @param {Event} event - Change event
     */
    handleMerchantSelect(merchant, index, event) {
        if (event.target.type === 'checkbox') {
            const isSelected = event.target.checked;
            
            if (isSelected) {
                this.selectedItems.add(merchant.id);
            } else {
                this.selectedItems.delete(merchant.id);
            }
            
            this.container.dispatchEvent(new CustomEvent('merchantSelectionChange', {
                detail: {
                    merchant,
                    index,
                    selected: isSelected,
                    selectedCount: this.selectedItems.size
                }
            }));
        }
    }

    /**
     * Set bulk mode
     * @param {boolean} enabled - Enable bulk mode
     */
    setBulkMode(enabled) {
        this.bulkMode = enabled;
        if (!enabled) {
            this.selectedItems.clear();
        }
        this.render();
    }

    /**
     * Select all visible items
     */
    selectAllVisible() {
        const visibleItems = this.getVisibleItems();
        visibleItems.forEach(item => {
            this.selectedItems.add(item.id);
        });
        this.render();
        
        this.container.dispatchEvent(new CustomEvent('merchantSelectionChange', {
            detail: {
                selectedCount: this.selectedItems.size
            }
        }));
    }

    /**
     * Deselect all items
     */
    deselectAll() {
        this.selectedItems.clear();
        this.render();
        
        this.container.dispatchEvent(new CustomEvent('merchantSelectionChange', {
            detail: {
                selectedCount: 0
            }
        }));
    }

    /**
     * Get selected merchants
     */
    getSelectedMerchants() {
        return this.filteredData.filter(merchant => this.selectedItems.has(merchant.id));
    }

    /**
     * Search merchants
     * @param {string} query - Search query
     */
    search(query) {
        this.searchQuery = query.toLowerCase();
        this.applyFilters();
    }

    /**
     * Set filter
     * @param {string} key - Filter key
     * @param {string} value - Filter value
     */
    setFilter(key, value) {
        this.filters[key] = value;
        this.applyFilters();
    }

    /**
     * Apply all filters
     */
    applyFilters() {
        this.filterData(merchant => {
            // Search filter
            if (this.searchQuery && !merchant.name.toLowerCase().includes(this.searchQuery) &&
                !merchant.industry.toLowerCase().includes(this.searchQuery) &&
                !merchant.location.toLowerCase().includes(this.searchQuery)) {
                return false;
            }
            
            // Portfolio type filter
            if (this.filters.portfolioType && merchant.portfolioType !== this.filters.portfolioType) {
                return false;
            }
            
            // Risk level filter
            if (this.filters.riskLevel && merchant.riskLevel !== this.filters.riskLevel) {
                return false;
            }
            
            // Industry filter
            if (this.filters.industry && merchant.industry !== this.filters.industry) {
                return false;
            }
            
            return true;
        });
    }

    /**
     * Clear all filters
     */
    clearFilters() {
        this.searchQuery = '';
        this.filters = {
            portfolioType: '',
            riskLevel: '',
            industry: ''
        };
        this.setData(this.data);
    }
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { VirtualScroller, MerchantVirtualScroller };
}
