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

  it('should render risk score panel with assessment data', () => {
    render(<RiskScorePanel assessment={mockAssessment} />);
    
    expect(screen.getByText('Why This Score?')).toBeInTheDocument();
    expect(screen.getByText('Risk score breakdown and factors')).toBeInTheDocument();
    expect(screen.getByText('7.5')).toBeInTheDocument();
    expect(screen.getByText('Medium')).toBeInTheDocument();
  });

  it('should render risk factors when provided', () => {
    render(<RiskScorePanel assessment={mockAssessment} />);
    
    expect(screen.getByText('Risk Factors')).toBeInTheDocument();
    expect(screen.getByText('Financial Risk')).toBeInTheDocument();
    expect(screen.getByText('Operational Risk')).toBeInTheDocument();
    expect(screen.getByText('Compliance Risk')).toBeInTheDocument();
  });

  it('should display factor scores and weights', () => {
    render(<RiskScorePanel assessment={mockAssessment} />);
    
    expect(screen.getByText('8.0')).toBeInTheDocument(); // Financial Risk score
    expect(screen.getByText('(weight: 0.40)')).toBeInTheDocument();
  });

  it('should toggle collapsible content', async () => {
    const user = userEvent.setup();
    render(<RiskScorePanel assessment={mockAssessment} collapsed={true} />);
    
    const toggleButton = screen.getByLabelText('Toggle risk score breakdown');
    await user.click(toggleButton);
    
    // Content should still be visible (collapsible behavior)
    expect(screen.getByText('Why This Score?')).toBeInTheDocument();
  });

  it('should apply correct badge variant for low risk', () => {
    const lowRiskAssessment: RiskAssessment = {
      ...mockAssessment,
      result: {
        ...mockAssessment.result,
        riskLevel: 'Low',
        overallScore: 3.0,
      },
    };
    
    render(<RiskScorePanel assessment={lowRiskAssessment} />);
    expect(screen.getByText('Low')).toBeInTheDocument();
  });

  it('should apply correct badge variant for high risk', () => {
    const highRiskAssessment: RiskAssessment = {
      ...mockAssessment,
      result: {
        ...mockAssessment.result,
        riskLevel: 'High',
        overallScore: 9.0,
      },
    };
    
    render(<RiskScorePanel assessment={highRiskAssessment} />);
    expect(screen.getByText('High')).toBeInTheDocument();
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

