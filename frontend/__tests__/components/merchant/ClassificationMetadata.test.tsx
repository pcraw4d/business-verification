import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { ClassificationMetadata } from '@/components/merchant/ClassificationMetadata';
import { ClassificationData } from '@/types/merchant';

describe('ClassificationMetadata', () => {
  describe('DisplaysPagesAnalyzed', () => {
    it('displays pages analyzed count when metadata is provided', () => {
      const metadata: ClassificationData['metadata'] = {
        pageAnalysis: {
          pagesAnalyzed: 12,
          analysisMethod: 'multi_page',
          structuredDataFound: true,
        },
      };

      render(<ClassificationMetadata metadata={metadata} />);

      expect(screen.getByText(/12 pages/i)).toBeInTheDocument();
    });

    it('displays analysis method badge', () => {
      const metadata: ClassificationData['metadata'] = {
        pageAnalysis: {
          pagesAnalyzed: 5,
          analysisMethod: 'single_page',
          structuredDataFound: false,
        },
      };

      render(<ClassificationMetadata metadata={metadata} />);

      expect(screen.getByText(/single-page/i)).toBeInTheDocument();
    });
  });

  describe('DisplaysStructuredData', () => {
    it('displays structured data indicator when found', () => {
      const metadata: ClassificationData['metadata'] = {
        pageAnalysis: {
          pagesAnalyzed: 8,
          analysisMethod: 'multi_page',
          structuredDataFound: true,
        },
      };

      render(<ClassificationMetadata metadata={metadata} />);

      expect(screen.getByText(/structured data found/i)).toBeInTheDocument();
    });

    it('does not display structured data indicator when not found', () => {
      const metadata: ClassificationData['metadata'] = {
        pageAnalysis: {
          pagesAnalyzed: 8,
          analysisMethod: 'multi_page',
          structuredDataFound: false,
        },
      };

      render(<ClassificationMetadata metadata={metadata} />);

      expect(screen.queryByText(/structured data found/i)).not.toBeInTheDocument();
    });
  });

  describe('DisplaysBrandMatch', () => {
    it('displays brand match badge for known hotel brands', () => {
      const metadata: ClassificationData['metadata'] = {
        brandMatch: {
          isBrandMatch: true,
          brandName: 'Hilton',
          confidence: 0.95,
        },
      };

      render(<ClassificationMetadata metadata={metadata} />);

      expect(screen.getByText(/brand match/i)).toBeInTheDocument();
      expect(screen.getByText(/Hilton/i)).toBeInTheDocument();
    });

    it('displays confidence score when provided', () => {
      const metadata: ClassificationData['metadata'] = {
        brandMatch: {
          isBrandMatch: true,
          brandName: 'Marriott',
          confidence: 0.85,
        },
      };

      render(<ClassificationMetadata metadata={metadata} />);

      expect(screen.getByText(/85%/i)).toBeInTheDocument();
    });

    it('does not display brand match when not matched', () => {
      const metadata: ClassificationData['metadata'] = {
        brandMatch: {
          isBrandMatch: false,
        },
      };

      render(<ClassificationMetadata metadata={metadata} />);

      expect(screen.queryByText(/brand match/i)).not.toBeInTheDocument();
    });
  });

  describe('DisplaysDataSource', () => {
    it('displays data source priority indicators', () => {
      const metadata: ClassificationData['metadata'] = {
        dataSourcePriority: {
          websiteContent: 'primary',
          businessName: 'secondary',
        },
      };

      render(<ClassificationMetadata metadata={metadata} />);

      expect(screen.getByText(/data source priority/i)).toBeInTheDocument();
      expect(screen.getByText(/website content/i)).toBeInTheDocument();
      expect(screen.getByText(/business name/i)).toBeInTheDocument();
    });

    it('displays primary badge for primary data source', () => {
      const metadata: ClassificationData['metadata'] = {
        dataSourcePriority: {
          websiteContent: 'primary',
          businessName: 'none',
        },
      };

      render(<ClassificationMetadata metadata={metadata} />);

      const primaryBadges = screen.getAllByText(/primary/i);
      expect(primaryBadges.length).toBeGreaterThan(0);
    });
  });

  describe('CompactMode', () => {
    it('displays compact version when compact prop is true', () => {
      const metadata: ClassificationData['metadata'] = {
        pageAnalysis: {
          pagesAnalyzed: 10,
          analysisMethod: 'multi_page',
          structuredDataFound: true,
        },
        brandMatch: {
          isBrandMatch: true,
          brandName: 'Hilton',
        },
      };

      render(<ClassificationMetadata metadata={metadata} compact />);

      // In compact mode, should show badges but not full details
      expect(screen.getByText(/10 pages/i)).toBeInTheDocument();
    });
  });

  describe('BackwardCompatibility', () => {
    it('returns null when metadata is not provided', () => {
      const { container } = render(<ClassificationMetadata metadata={undefined} />);

      expect(container.firstChild).toBeNull();
    });

    it('handles missing pageAnalysis gracefully', () => {
      const metadata: ClassificationData['metadata'] = {
        brandMatch: {
          isBrandMatch: true,
          brandName: 'Hilton',
        },
      };

      render(<ClassificationMetadata metadata={metadata} />);

      // Should still render brand match even without pageAnalysis
      expect(screen.getByText(/brand match/i)).toBeInTheDocument();
    });

    it('handles missing brandMatch gracefully', () => {
      const metadata: ClassificationData['metadata'] = {
        pageAnalysis: {
          pagesAnalyzed: 5,
          analysisMethod: 'single_page',
          structuredDataFound: false,
        },
      };

      render(<ClassificationMetadata metadata={metadata} />);

      // Should still render page analysis even without brandMatch
      expect(screen.getByText(/5 pages/i)).toBeInTheDocument();
    });

    it('handles empty metadata object', () => {
      const metadata: ClassificationData['metadata'] = {};

      render(<ClassificationMetadata metadata={metadata} />);

      // Should not crash, but may not display anything
      expect(screen.queryByText(/analysis metadata/i)).not.toBeInTheDocument();
    });
  });
});

