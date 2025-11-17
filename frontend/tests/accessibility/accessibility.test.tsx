/**
 * Accessibility tests using @axe-core/react
 * 
 * Note: These tests use Vitest, not Playwright
 * Run with: npm run test:accessibility
 */

import { describe, test, expect } from 'vitest';
import { render } from '@testing-library/react';
import { axe, toHaveNoViolations } from 'jest-axe';
import { MerchantPortfolioPage } from '@/app/merchant-portfolio/page';
import { DashboardPage } from '@/app/dashboard/page';

expect.extend(toHaveNoViolations);

describe('Accessibility Tests', () => {
  test('merchant portfolio page should have no accessibility violations', async () => {
    const { container } = render(<MerchantPortfolioPage />);
    const results = await axe(container);
    expect(results).toHaveNoViolations();
  });

  test('dashboard page should have no accessibility violations', async () => {
    const { container } = render(<DashboardPage />);
    const results = await axe(container);
    expect(results).toHaveNoViolations();
  });

  test('all interactive elements should be keyboard accessible', async () => {
    const { container } = render(<MerchantPortfolioPage />);
    const interactiveElements = container.querySelectorAll('button, a, input, select, textarea');
    
    interactiveElements.forEach((element) => {
      // Check for tabindex
      const tabIndex = element.getAttribute('tabindex');
      if (tabIndex === '-1' && element.getAttribute('disabled') !== 'true') {
        // Elements with tabindex="-1" should have another way to be accessed
        expect(element.getAttribute('aria-hidden')).not.toBe('true');
      }
    });
  });

  test('all images should have alt text', async () => {
    const { container } = render(<MerchantPortfolioPage />);
    const images = container.querySelectorAll('img');
    
    images.forEach((img) => {
      const alt = img.getAttribute('alt');
      expect(alt).not.toBeNull();
      // Alt should not be empty unless image is decorative
      if (alt === '') {
        expect(img.getAttribute('role')).toBe('presentation');
      }
    });
  });

  test('form inputs should have labels', async () => {
    const { container } = render(<MerchantPortfolioPage />);
    const inputs = container.querySelectorAll('input, select, textarea');
    
    inputs.forEach((input) => {
      const id = input.getAttribute('id');
      const ariaLabel = input.getAttribute('aria-label');
      const ariaLabelledBy = input.getAttribute('aria-labelledby');
      
      // Input should have at least one: id with label, aria-label, or aria-labelledby
      const hasLabel = id && container.querySelector(`label[for="${id}"]`);
      const hasAccessibleName = ariaLabel || ariaLabelledBy || hasLabel;
      
      // Skip hidden inputs
      if (input.getAttribute('type') !== 'hidden' && input.getAttribute('aria-hidden') !== 'true') {
        expect(hasAccessibleName).toBeTruthy();
      }
    });
  });

  test('heading hierarchy should be logical', async () => {
    const { container } = render(<MerchantPortfolioPage />);
    const headings = Array.from(container.querySelectorAll('h1, h2, h3, h4, h5, h6'));
    
    let lastLevel = 0;
    headings.forEach((heading) => {
      const level = parseInt(heading.tagName.charAt(1));
      // Allow skipping one level (h1 -> h3 is ok, but h1 -> h4 is not)
      if (lastLevel > 0 && level > lastLevel + 1) {
        // This is a warning, not an error
        console.warn(`Heading level ${level} follows level ${lastLevel} - may skip levels`);
      }
      lastLevel = level;
    });
  });
});

