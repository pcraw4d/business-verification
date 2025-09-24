/**
 * Simple test to verify Jest setup is working
 */

describe('Jest Setup Verification', () => {
    test('should have jsdom environment', () => {
        expect(typeof document).toBe('object');
        expect(typeof window).toBe('object');
        expect(typeof localStorage).toBe('object');
    });

    test('should have mocked fetch', () => {
        expect(typeof fetch).toBe('function');
        expect(fetch).toBeDefined();
    });

    test('should have mocked console', () => {
        expect(typeof console.log).toBe('function');
        expect(console.log).toBeDefined();
    });

    test('should be able to create DOM elements', () => {
        const div = document.createElement('div');
        div.textContent = 'Test';
        document.body.appendChild(div);
        
        expect(div.textContent).toBe('Test');
        expect(document.body.contains(div)).toBe(true);
        
        document.body.removeChild(div);
    });
});
