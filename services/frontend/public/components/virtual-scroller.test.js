/**
 * Virtual Scroller Unit Tests
 * Tests for virtual scrolling functionality and performance optimization
 */

const { VirtualScroller, MerchantVirtualScroller } = require('./virtual-scroller');

describe('VirtualScroller', () => {
    let container;
    let virtualScroller;
    let mockData;

    beforeEach(() => {
        // Create container element
        container = document.createElement('div');
        container.style.height = '400px';
        container.style.width = '300px';
        document.body.appendChild(container);

        // Create mock data
        mockData = Array.from({ length: 1000 }, (_, i) => ({
            id: `item-${i}`,
            name: `Item ${i}`,
            value: i
        }));

        // Create virtual scroller
        virtualScroller = new VirtualScroller({
            container,
            itemHeight: 50,
            bufferSize: 5
        });
    });

    afterEach(() => {
        if (virtualScroller) {
            virtualScroller.destroy();
        }
        if (container && container.parentNode) {
            container.parentNode.removeChild(container);
        }
    });

    describe('Initialization', () => {
        test('should initialize with default options', () => {
            expect(virtualScroller.itemHeight).toBe(50);
            expect(virtualScroller.bufferSize).toBe(5);
            expect(virtualScroller.container).toBe(container);
        });

        test('should throw error without container', () => {
            expect(() => {
                new VirtualScroller({ container: null });
            }).toThrow('VirtualScroller: Container element is required');
        });

        test('should setup container structure', () => {
            expect(container.children.length).toBe(1); // viewport
            expect(container.firstChild.children.length).toBe(1); // scrollableArea
            expect(container.firstChild.firstChild.children.length).toBe(0); // itemsContainer (empty initially)
        });
    });

    describe('Data Management', () => {
        test('should set data correctly', () => {
            virtualScroller.setData(mockData);
            
            expect(virtualScroller.data).toEqual(mockData);
            expect(virtualScroller.filteredData).toEqual(mockData);
        });

        test('should handle empty data', () => {
            virtualScroller.setData([]);
            
            expect(virtualScroller.data).toEqual([]);
            expect(virtualScroller.filteredData).toEqual([]);
        });

        test('should filter data correctly', () => {
            virtualScroller.setData(mockData);
            virtualScroller.filterData(item => item.value % 2 === 0);
            
            expect(virtualScroller.filteredData.length).toBe(500);
            expect(virtualScroller.filteredData.every(item => item.value % 2 === 0)).toBe(true);
        });

        test('should sort data correctly', () => {
            virtualScroller.setData(mockData);
            virtualScroller.sortData((a, b) => b.value - a.value);
            
            expect(virtualScroller.filteredData[0].value).toBe(999);
            expect(virtualScroller.filteredData[999].value).toBe(0);
        });
    });

    describe('Rendering', () => {
        test('should render items with custom renderer', () => {
            const renderer = jest.fn((item, index) => `<div>Item ${index}: ${item.name}</div>`);
            virtualScroller.setItemRenderer(renderer);
            virtualScroller.setData(mockData.slice(0, 10));
            
            // Trigger render
            virtualScroller.render();
            
            expect(renderer).toHaveBeenCalled();
        });

        test('should handle missing renderer gracefully', () => {
            const consoleSpy = jest.spyOn(console, 'warn').mockImplementation();
            
            virtualScroller.setData(mockData.slice(0, 10));
            virtualScroller.render();
            
            expect(consoleSpy).toHaveBeenCalledWith('VirtualScroller: No item renderer set');
            
            consoleSpy.mockRestore();
        });

        test('should create item elements with correct structure', () => {
            const renderer = jest.fn((item, index) => `<div>Item ${index}</div>`);
            virtualScroller.setItemRenderer(renderer);
            virtualScroller.setData(mockData.slice(0, 5));
            
            virtualScroller.render();
            
            const items = virtualScroller.itemsContainer.children;
            expect(items.length).toBeGreaterThan(0);
            
            for (let i = 0; i < items.length; i++) {
                const item = items[i];
                expect(item.classList.contains('virtual-scroll-item')).toBe(true);
                expect(item.dataset.itemIndex).toBeDefined();
                expect(item.style.height).toBe('50px');
            }
        });
    });

    describe('Scrolling', () => {
        test('should calculate visible range correctly', () => {
            virtualScroller.setData(mockData);
            virtualScroller.updateDimensions();
            
            const stats = virtualScroller.getStats();
            expect(stats.startIndex).toBeGreaterThanOrEqual(0);
            expect(stats.endIndex).toBeLessThan(mockData.length);
            expect(stats.visibleItems).toBeGreaterThan(0);
        });

        test('should scroll to specific item', () => {
            virtualScroller.setData(mockData);
            
            virtualScroller.scrollToItem(100, false);
            
            expect(virtualScroller.container.scrollTop).toBe(100 * 50);
        });

        test('should scroll to top', () => {
            virtualScroller.setData(mockData);
            virtualScroller.scrollToItem(100, false);
            
            virtualScroller.scrollToTop(false);
            
            expect(virtualScroller.container.scrollTop).toBe(0);
        });

        test('should scroll to bottom', () => {
            virtualScroller.setData(mockData);
            
            virtualScroller.scrollToBottom(false);
            
            expect(virtualScroller.container.scrollTop).toBe((mockData.length - 1) * 50);
        });
    });

    describe('Item Management', () => {
        test('should update item at specific index', () => {
            const renderer = jest.fn((item, index) => `<div>Item ${index}</div>`);
            virtualScroller.setItemRenderer(renderer);
            virtualScroller.setData(mockData.slice(0, 10));
            
            virtualScroller.updateItem(5, { name: 'Updated Item' });
            
            expect(virtualScroller.filteredData[5].name).toBe('Updated Item');
        });

        test('should add item', () => {
            virtualScroller.setData(mockData.slice(0, 10));
            const newItem = { id: 'new-item', name: 'New Item', value: 999 };
            
            virtualScroller.addItem(newItem);
            
            expect(virtualScroller.filteredData.length).toBe(11);
            expect(virtualScroller.filteredData[10]).toEqual(newItem);
        });

        test('should add item at specific index', () => {
            virtualScroller.setData(mockData.slice(0, 10));
            const newItem = { id: 'new-item', name: 'New Item', value: 999 };
            
            virtualScroller.addItem(newItem, 5);
            
            expect(virtualScroller.filteredData.length).toBe(11);
            expect(virtualScroller.filteredData[5]).toEqual(newItem);
        });

        test('should remove item', () => {
            virtualScroller.setData(mockData.slice(0, 10));
            const originalLength = virtualScroller.filteredData.length;
            
            virtualScroller.removeItem(5);
            
            expect(virtualScroller.filteredData.length).toBe(originalLength - 1);
        });

        test('should clear all items', () => {
            virtualScroller.setData(mockData);
            
            virtualScroller.clear();
            
            expect(virtualScroller.data).toEqual([]);
            expect(virtualScroller.filteredData).toEqual([]);
        });
    });

    describe('Event Handling', () => {
        test('should handle item click events', () => {
            const clickHandler = jest.fn();
            virtualScroller.setItemClickHandler(clickHandler);
            virtualScroller.setData(mockData.slice(0, 5));
            
            // Simulate click on item
            const event = new Event('click');
            Object.defineProperty(event, 'target', {
                value: { closest: () => ({ dataset: { itemIndex: '2' } }) }
            });
            
            virtualScroller.handleItemClick(event);
            
            expect(clickHandler).toHaveBeenCalledWith(mockData[2], 2, event);
        });

        test('should handle item change events', () => {
            const changeHandler = jest.fn();
            virtualScroller.setItemSelectHandler(changeHandler);
            virtualScroller.setData(mockData.slice(0, 5));
            
            // Simulate change on item
            const event = new Event('change');
            Object.defineProperty(event, 'target', {
                value: { 
                    closest: () => ({ dataset: { itemIndex: '2' } }),
                    type: 'checkbox'
                }
            });
            
            virtualScroller.handleItemChange(event);
            
            expect(changeHandler).toHaveBeenCalledWith(mockData[2], 2, event);
        });
    });

    describe('Statistics', () => {
        test('should return correct statistics', () => {
            virtualScroller.setData(mockData);
            virtualScroller.updateDimensions();
            
            const stats = virtualScroller.getStats();
            
            expect(stats.totalItems).toBe(mockData.length);
            expect(stats.visibleItems).toBeGreaterThan(0);
            expect(stats.startIndex).toBeGreaterThanOrEqual(0);
            expect(stats.endIndex).toBeLessThan(mockData.length);
            expect(stats.containerHeight).toBeGreaterThan(0);
            expect(stats.totalHeight).toBe(mockData.length * 50);
        });
    });
});

describe('MerchantVirtualScroller', () => {
    let container;
    let merchantScroller;
    let mockMerchants;

    beforeEach(() => {
        // Create container element
        container = document.createElement('div');
        container.style.height = '400px';
        container.style.width = '300px';
        document.body.appendChild(container);

        // Create mock merchant data
        mockMerchants = Array.from({ length: 100 }, (_, i) => ({
            id: `merchant-${i}`,
            name: `Merchant ${i}`,
            industry: i % 2 === 0 ? 'Technology' : 'Finance',
            portfolioType: ['onboarded', 'pending', 'deactivated'][i % 3],
            riskLevel: ['low', 'medium', 'high'][i % 3],
            location: `City ${i}`
        }));

        // Create merchant virtual scroller
        merchantScroller = new MerchantVirtualScroller({
            container,
            itemHeight: 100,
            bufferSize: 5
        });
    });

    afterEach(() => {
        if (merchantScroller) {
            merchantScroller.destroy();
        }
        if (container && container.parentNode) {
            container.parentNode.removeChild(container);
        }
    });

    describe('Initialization', () => {
        test('should initialize with merchant-specific settings', () => {
            expect(merchantScroller.itemHeight).toBe(100);
            expect(merchantScroller.bufferSize).toBe(5);
            expect(merchantScroller.selectedItems).toBeInstanceOf(Set);
            expect(merchantScroller.bulkMode).toBe(false);
        });

        test('should have merchant item renderer set', () => {
            expect(merchantScroller.itemRenderer).toBeDefined();
            expect(typeof merchantScroller.itemRenderer).toBe('function');
        });
    });

    describe('Merchant Rendering', () => {
        test('should render merchant items correctly', () => {
            merchantScroller.setData(mockMerchants.slice(0, 5));
            merchantScroller.render();
            
            const items = merchantScroller.itemsContainer.children;
            expect(items.length).toBeGreaterThan(0);
            
            const firstItem = items[0];
            expect(firstItem.innerHTML).toContain('Merchant 0');
            expect(firstItem.innerHTML).toContain('Technology');
            expect(firstItem.innerHTML).toContain('onboarded');
        });

        test('should render selected merchants with selected class', () => {
            merchantScroller.setData(mockMerchants.slice(0, 5));
            merchantScroller.selectedItems.add('merchant-0');
            merchantScroller.render();
            
            const items = merchantScroller.itemsContainer.children;
            const firstItem = items[0];
            expect(firstItem.innerHTML).toContain('selected');
        });

        test('should render bulk mode with checkboxes', () => {
            merchantScroller.setData(mockMerchants.slice(0, 5));
            merchantScroller.setBulkMode(true);
            merchantScroller.render();
            
            const items = merchantScroller.itemsContainer.children;
            const firstItem = items[0];
            expect(firstItem.innerHTML).toContain('checkbox');
        });
    });

    describe('Selection Management', () => {
        test('should handle merchant selection', () => {
            merchantScroller.setData(mockMerchants.slice(0, 5));
            
            const event = new Event('change');
            Object.defineProperty(event, 'target', {
                value: { 
                    closest: () => ({ dataset: { itemIndex: '0' } }),
                    type: 'checkbox',
                    checked: true
                }
            });
            
            merchantScroller.handleMerchantSelect(mockMerchants[0], 0, event);
            
            expect(merchantScroller.selectedItems.has('merchant-0')).toBe(true);
        });

        test('should handle merchant deselection', () => {
            merchantScroller.setData(mockMerchants.slice(0, 5));
            merchantScroller.selectedItems.add('merchant-0');
            
            const event = new Event('change');
            Object.defineProperty(event, 'target', {
                value: { 
                    closest: () => ({ dataset: { itemIndex: '0' } }),
                    type: 'checkbox',
                    checked: false
                }
            });
            
            merchantScroller.handleMerchantSelect(mockMerchants[0], 0, event);
            
            expect(merchantScroller.selectedItems.has('merchant-0')).toBe(false);
        });

        test('should set bulk mode correctly', () => {
            merchantScroller.setData(mockMerchants.slice(0, 5));
            
            merchantScroller.setBulkMode(true);
            expect(merchantScroller.bulkMode).toBe(true);
            
            merchantScroller.setBulkMode(false);
            expect(merchantScroller.bulkMode).toBe(false);
            expect(merchantScroller.selectedItems.size).toBe(0);
        });

        test('should select all visible items', () => {
            merchantScroller.setData(mockMerchants.slice(0, 10));
            merchantScroller.updateDimensions();
            
            merchantScroller.selectAllVisible();
            
            const visibleItems = merchantScroller.getVisibleItems();
            expect(merchantScroller.selectedItems.size).toBe(visibleItems.length);
        });

        test('should deselect all items', () => {
            merchantScroller.setData(mockMerchants.slice(0, 5));
            merchantScroller.selectedItems.add('merchant-0');
            merchantScroller.selectedItems.add('merchant-1');
            
            merchantScroller.deselectAll();
            
            expect(merchantScroller.selectedItems.size).toBe(0);
        });

        test('should get selected merchants', () => {
            merchantScroller.setData(mockMerchants.slice(0, 5));
            merchantScroller.selectedItems.add('merchant-0');
            merchantScroller.selectedItems.add('merchant-2');
            
            const selected = merchantScroller.getSelectedMerchants();
            
            expect(selected.length).toBe(2);
            expect(selected.map(m => m.id)).toContain('merchant-0');
            expect(selected.map(m => m.id)).toContain('merchant-2');
        });
    });

    describe('Search and Filtering', () => {
        test('should search merchants by name', () => {
            merchantScroller.setData(mockMerchants);
            
            merchantScroller.search('Merchant 1');
            
            expect(merchantScroller.filteredData.length).toBe(1);
            expect(merchantScroller.filteredData[0].name).toBe('Merchant 1');
        });

        test('should search merchants by industry', () => {
            merchantScroller.setData(mockMerchants);
            
            merchantScroller.search('Technology');
            
            expect(merchantScroller.filteredData.length).toBe(50);
            expect(merchantScroller.filteredData.every(m => m.industry === 'Technology')).toBe(true);
        });

        test('should filter by portfolio type', () => {
            merchantScroller.setData(mockMerchants);
            
            merchantScroller.setFilter('portfolioType', 'onboarded');
            
            expect(merchantScroller.filteredData.every(m => m.portfolioType === 'onboarded')).toBe(true);
        });

        test('should filter by risk level', () => {
            merchantScroller.setData(mockMerchants);
            
            merchantScroller.setFilter('riskLevel', 'high');
            
            expect(merchantScroller.filteredData.every(m => m.riskLevel === 'high')).toBe(true);
        });

        test('should apply multiple filters', () => {
            merchantScroller.setData(mockMerchants);
            
            merchantScroller.setFilter('industry', 'Technology');
            merchantScroller.setFilter('riskLevel', 'low');
            
            expect(merchantScroller.filteredData.every(m => 
                m.industry === 'Technology' && m.riskLevel === 'low'
            )).toBe(true);
        });

        test('should clear all filters', () => {
            merchantScroller.setData(mockMerchants);
            merchantScroller.setFilter('industry', 'Technology');
            merchantScroller.search('test');
            
            merchantScroller.clearFilters();
            
            expect(merchantScroller.filteredData.length).toBe(mockMerchants.length);
            expect(merchantScroller.searchQuery).toBe('');
        });
    });

    describe('Event Handling', () => {
        test('should handle merchant view action', () => {
            merchantScroller.setData(mockMerchants.slice(0, 5));
            
            const event = new Event('click');
            Object.defineProperty(event, 'target', {
                value: { 
                    closest: () => ({ dataset: { itemIndex: '0' } }),
                    dataset: { action: 'view', merchantId: 'merchant-0' }
                }
            });
            
            const viewHandler = jest.fn();
            container.addEventListener('merchantView', viewHandler);
            
            merchantScroller.handleMerchantClick(mockMerchants[0], 0, event);
            
            expect(viewHandler).toHaveBeenCalled();
        });

        test('should handle merchant compare action', () => {
            merchantScroller.setData(mockMerchants.slice(0, 5));
            
            const event = new Event('click');
            Object.defineProperty(event, 'target', {
                value: { 
                    closest: () => ({ dataset: { itemIndex: '0' } }),
                    dataset: { action: 'compare', merchantId: 'merchant-0' }
                }
            });
            
            const compareHandler = jest.fn();
            container.addEventListener('merchantCompare', compareHandler);
            
            merchantScroller.handleMerchantClick(mockMerchants[0], 0, event);
            
            expect(compareHandler).toHaveBeenCalled();
        });
    });
});
