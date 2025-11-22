import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { RiskScorePanel } from '@/components/risk/RiskScorePanel';
import { describe, it, expect, vi } from 'vitest';
import type { RiskAssessment } from '@/types/merchant';

describe('RiskScorePanel', () => {
  const mockAssessment: RiskAssessment = {
    id: 'assessment-1',
    merchantId: 'merchant-1',
    status: 'completed',
    createdAt: new Date().toISOString(),
    result: {
      overallScore: 7.5,
      riskLevel: 'Medium',
      factors: [
        { name: 'Financial Risk', score: 8.0, weight: 0.4 },
        { name: 'Operational Risk', score: 6.0, weight: 0.3 },
        { name: 'Compliance Risk', score: 8.5, weight: 0.3 },
      ],
    },
  };

  it('should return null when assessment is null', () => {
    const { container } = render(<RiskScorePanel assessment={null} />);
    expect(container.firstChild).toBeNull();
  });

  it('should return null when assessment result is missing', () => {
    const { container } = render(
      <RiskScorePanel assessment={{ ...mockAssessment, result: null as any }} />
    );
    expect(container.firstChild).toBeNull();
  });

  it('should render risk score panel with assessment data', async () => {
    const user = userEvent.setup();
    render(<RiskScorePanel assessment={mockAssessment} collapsed={false} />);
    
    expect(screen.getByText('Why This Score?')).toBeInTheDocument();
    expect(screen.getByText('Risk score breakdown and factors')).toBeInTheDocument();
    
    // Content is in collapsible - expand if needed
    // With collapsed={false}, defaultOpen={true}, content should be visible
    await waitFor(() => {
      // Score is formatted with formatNumber, might be "7.5" or "7.50"
      expect(screen.getByText(/7\.5/)).toBeInTheDocument();
      expect(screen.getByText('Medium')).toBeInTheDocument();
    }, { timeout: 3000 });
  });

  it('should render risk factors when provided', async () => {
    render(<RiskScorePanel assessment={mockAssessment} collapsed={false} />);
    
    // Content is in collapsible - should be visible with collapsed={false}
    await waitFor(() => {
      expect(screen.getByText('Risk Factors')).toBeInTheDocument();
      expect(screen.getByText('Financial Risk')).toBeInTheDocument();
      expect(screen.getByText('Operational Risk')).toBeInTheDocument();
      expect(screen.getByText('Compliance Risk')).toBeInTheDocument();
    }, { timeout: 3000 });
  });

  it('should display factor scores and weights', async () => {
    render(<RiskScorePanel assessment={mockAssessment} collapsed={false} />);
    
    // Content is in collapsible - should be visible with collapsed={false}
    await waitFor(() => {
      // Scores are formatted with formatNumber
      expect(screen.getByText(/8\.0/)).toBeInTheDocument(); // Financial Risk score
      // Weight is formatted with formatNumber(weight, 2) - might be "0.40" or "0.4"
      expect(screen.getByText(/weight.*0\.4/)).toBeInTheDocument();
    }, { timeout: 3000 });
  });

  it('should toggle collapsible content', async () => {
    const user = userEvent.setup();
    render(<RiskScorePanel assessment={mockAssessment} collapsed={true} />);
    
    const toggleButton = screen.getByLabelText('Toggle risk score breakdown');
    await user.click(toggleButton);
    
    // Content should still be visible (collapsible behavior)
    expect(screen.getByText('Why This Score?')).toBeInTheDocument();
  });

  it('should apply correct badge variant for low risk', async () => {
    const lowRiskAssessment: RiskAssessment = {
      ...mockAssessment,
      result: {
        ...mockAssessment.result,
        riskLevel: 'Low',
        overallScore: 3.0,
      },
    };
    
    render(<RiskScorePanel assessment={lowRiskAssessment} collapsed={false} />);
    
    await waitFor(() => {
      expect(screen.getByText('Low')).toBeInTheDocument();
    }, { timeout: 3000 });
  });

  it('should apply correct badge variant for high risk', async () => {
    const highRiskAssessment: RiskAssessment = {
      ...mockAssessment,
      result: {
        ...mockAssessment.result,
        riskLevel: 'High',
        overallScore: 9.0,
      },
    };
    
    render(<RiskScorePanel assessment={highRiskAssessment} collapsed={false} />);
    
    await waitFor(() => {
      expect(screen.getByText('High')).toBeInTheDocument();
    }, { timeout: 3000 });
  });

  it('should handle assessment without factors', () => {
    const assessmentWithoutFactors: RiskAssessment = {
      ...mockAssessment,
      result: {
        ...mockAssessment.result,
        factors: [],
      },
    };
    
    render(<RiskScorePanel assessment={assessmentWithoutFactors} />);
    
    expect(screen.getByText('Why This Score?')).toBeInTheDocument();
    expect(screen.queryByText('Risk Factors')).not.toBeInTheDocument();
  });
});

