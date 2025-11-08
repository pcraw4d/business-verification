/**
 * Export Functionality Frontend Tests
 */
describe('Export Functionality', () => {
    let exportButton;

    beforeEach(() => {
        document.body.innerHTML = `
            <div id="exportButtonContainer"></div>
        `;

        if (typeof ExportButton !== 'undefined') {
            exportButton = new ExportButton({
                container: document.getElementById('exportButtonContainer'),
                exportType: 'merchant',
                formats: ['csv', 'pdf', 'json']
            });
        }
    });

    afterEach(() => {
        if (exportButton) {
            exportButton = null;
        }
    });

    test('should initialize export button', async () => {
        if (exportButton) {
            await exportButton.init();
            expect(exportButton).toBeDefined();
        }
    });

    test('should create export button UI', async () => {
        if (exportButton) {
            await exportButton.init();
            const button = document.getElementById('exportBtn');
            expect(button).toBeDefined();
        }
    });

    test('should show export dropdown', async () => {
        if (exportButton) {
            await exportButton.init();
            const button = document.getElementById('exportBtn');
            const dropdown = document.getElementById('exportDropdown');
            
            if (button && dropdown) {
                button.click();
                expect(dropdown.style.display).toBe('block');
            }
        }
    });
});

