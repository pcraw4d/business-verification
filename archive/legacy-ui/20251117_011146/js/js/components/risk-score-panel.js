/**
 * Risk Score Panel Component
 * Displays risk score explanation and breakdown
 */
class RiskScorePanel {
    constructor(containerId, options = {}) {
        this.containerId = containerId;
        this.container = document.getElementById(containerId);
        this.options = {
            collapsed: true,
            showBreakdown: true,
            showFactors: true,
            ...options
        };
        this.isVisible = false;
        this.scoreData = null;
    }

    /**
     * Initialize the score panel
     */
    init() {
        if (!this.container) {
            console.error(`Score panel container ${this.containerId} not found`);
            return;
        }

        this.render();
    }

    /**
     * Render the score panel
     */
    render() {
        if (!this.container) return;

        this.container.innerHTML = `
            <div class="risk-score-panel" style="display: ${this.isVisible ? 'block' : 'none'};">
                <div class="score-panel-header">
                    <h3><i class="fas fa-info-circle"></i> Why This Score?</h3>
                    <button class="score-panel-toggle" onclick="riskScorePanel.toggle()">
                        <i class="fas fa-${this.options.collapsed ? 'chevron-down' : 'chevron-up'}"></i>
                    </button>
                </div>
                <div class="score-panel-content" style="display: ${this.options.collapsed ? 'none' : 'block'};">
                    <div id="scoreBreakdown"></div>
                    <div id="scoreFactors"></div>
                </div>
            </div>
        `;

        this.addStyles();
    }

    /**
     * Add panel styles
     */
    addStyles() {
        if (document.getElementById('riskScorePanelStyles')) {
            return;
        }

        const style = document.createElement('style');
        style.id = 'riskScorePanelStyles';
        style.textContent = `
            .risk-score-panel {
                background: white;
                border-radius: 8px;
                padding: 1.5rem;
                margin: 1rem 0;
                box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            }

            .score-panel-header {
                display: flex;
                justify-content: space-between;
                align-items: center;
                margin-bottom: 1rem;
            }

            .score-panel-header h3 {
                margin: 0;
                color: #1a1a1a;
                font-size: 1.1rem;
            }

            .score-panel-toggle {
                background: none;
                border: none;
                cursor: pointer;
                color: #4a90e2;
                font-size: 1.2rem;
                padding: 0.5rem;
            }

            .score-panel-content {
                margin-top: 1rem;
            }

            .score-breakdown {
                margin-bottom: 1.5rem;
            }

            .score-breakdown-item {
                display: flex;
                justify-content: space-between;
                padding: 0.75rem;
                border-bottom: 1px solid #e0e0e0;
            }

            .score-breakdown-item:last-child {
                border-bottom: none;
            }

            .score-breakdown-label {
                font-weight: 500;
                color: #333;
            }

            .score-breakdown-value {
                color: #4a90e2;
                font-weight: 600;
            }

            .score-factors {
                margin-top: 1.5rem;
            }

            .score-factor {
                padding: 0.75rem;
                margin-bottom: 0.5rem;
                background: #f8f9fa;
                border-radius: 4px;
                border-left: 3px solid #4a90e2;
            }

            .score-factor-label {
                font-weight: 600;
                margin-bottom: 0.25rem;
            }

            .score-factor-impact {
                font-size: 0.9rem;
                color: #666;
            }
        `;
        document.head.appendChild(style);
    }

    /**
     * Toggle panel visibility
     */
    toggle() {
        this.options.collapsed = !this.options.collapsed;
        const content = this.container.querySelector('.score-panel-content');
        const toggle = this.container.querySelector('.score-panel-toggle i');
        
        if (content) {
            content.style.display = this.options.collapsed ? 'none' : 'block';
        }
        
        if (toggle) {
            toggle.className = `fas fa-${this.options.collapsed ? 'chevron-down' : 'chevron-up'}`;
        }
    }

    /**
     * Show the panel
     */
    show() {
        this.isVisible = true;
        if (this.container) {
            const panel = this.container.querySelector('.risk-score-panel');
            if (panel) {
                panel.style.display = 'block';
            }
        }
    }

    /**
     * Hide the panel
     */
    hide() {
        this.isVisible = false;
        if (this.container) {
            const panel = this.container.querySelector('.risk-score-panel');
            if (panel) {
                panel.style.display = 'none';
            }
        }
    }

    /**
     * Update score data
     */
    updateScore(scoreData) {
        this.scoreData = scoreData;
        this.renderScoreBreakdown();
        if (this.options.showFactors) {
            this.renderScoreFactors();
        }
    }

    /**
     * Render score breakdown
     */
    renderScoreBreakdown() {
        const breakdownContainer = document.getElementById('scoreBreakdown');
        if (!breakdownContainer || !this.scoreData) return;

        const breakdown = this.scoreData.breakdown || [];
        const totalScore = this.scoreData.totalScore || this.scoreData.score || 0;

        breakdownContainer.innerHTML = `
            <div class="score-breakdown">
                <h4 style="margin-bottom: 1rem;">Score Breakdown</h4>
                ${breakdown.map(item => `
                    <div class="score-breakdown-item">
                        <span class="score-breakdown-label">${item.label || item.name}</span>
                        <span class="score-breakdown-value">${item.value || item.score}</span>
                    </div>
                `).join('')}
                <div class="score-breakdown-item" style="border-top: 2px solid #4a90e2; margin-top: 0.5rem; padding-top: 1rem;">
                    <span class="score-breakdown-label" style="font-weight: 600;">Total Score</span>
                    <span class="score-breakdown-value" style="font-size: 1.2rem;">${totalScore.toFixed(2)}</span>
                </div>
            </div>
        `;
    }

    /**
     * Render score factors
     */
    renderScoreFactors() {
        const factorsContainer = document.getElementById('scoreFactors');
        if (!factorsContainer || !this.scoreData) return;

        const factors = this.scoreData.factors || [];

        factorsContainer.innerHTML = `
            <div class="score-factors">
                <h4 style="margin-bottom: 1rem;">Key Factors</h4>
                ${factors.map(factor => `
                    <div class="score-factor">
                        <div class="score-factor-label">${factor.label || factor.name}</div>
                        <div class="score-factor-impact">${factor.impact || factor.description || ''}</div>
                    </div>
                `).join('')}
            </div>
        `;
    }
}

// Global function for backward compatibility
function toggleWhyScorePanel() {
    if (window.riskScorePanel) {
        window.riskScorePanel.toggle();
    }
}

// Export for use in other modules
if (typeof window !== 'undefined') {
    window.RiskScorePanel = RiskScorePanel;
}

