// Unified Navigation System for KYB Platform
class KYBNavigation {
    constructor() {
        this.currentPage = this.getCurrentPage();
        this.init();
    }

    init() {
        this.createNavigation();
        this.addNavigationToPage();
        this.setActivePage();
        this.bindEvents();
    }

    getCurrentPage() {
        const path = window.location.pathname;
        const filename = path.split('/').pop().replace('.html', '');
        
        const pageMap = {
            'index': 'home',
            'dashboard-hub': 'home',
            'add-merchant': 'add-merchant',
            'merchant-details': 'merchant-details',
            'dashboard': 'business-intelligence',
            'risk-dashboard': 'risk-assessment',
            'compliance-dashboard': 'compliance-status',
            'compliance-gap-analysis': 'compliance-gaps',
            'compliance-progress-tracking': 'compliance-progress',
            'market-analysis-dashboard': 'market-analysis',
            'competitive-analysis-dashboard': 'competitive-analysis',
            'business-growth-analytics': 'growth-analytics',
            'enhanced-risk-indicators': 'risk-indicators',
            'merchant-hub-integration': 'merchant-hub',
            'merchant-portfolio': 'merchant-portfolio',
            'merchant-detail': 'merchant-detail',
            'risk-assessment-portfolio': 'risk-assessment-portfolio'
        };

        return pageMap[filename] || 'home';
    }

    createNavigation() {
        const navigationHTML = `
            <div class="kyb-sidebar">
                <div class="sidebar-header">
                    <a href="index.html" class="brand-link">
                        <i class="fas fa-shield-alt"></i>
                        <span class="brand-text">KYB Platform</span>
                    </a>
                    <button class="sidebar-toggle" id="sidebarToggle">
                        <i class="fas fa-bars"></i>
                    </button>
                </div>
                
                <div class="sidebar-content">
                    <div class="nav-section">
                        <h3 class="nav-section-title">Platform</h3>
                        <ul class="nav-list">
                            <li class="nav-item">
                                <a href="index.html" class="nav-link" data-page="home">
                                    <i class="fas fa-home"></i>
                                    <span class="nav-text">Home</span>
                                </a>
                            </li>
                            <li class="nav-item">
                                <a href="dashboard-hub.html" class="nav-link" data-page="home">
                                    <i class="fas fa-th-large"></i>
                                    <span class="nav-text">Dashboard Hub</span>
                                </a>
                            </li>
                        </ul>
                    </div>
                    
                    <div class="nav-section">
                        <h3 class="nav-section-title">Merchant Verification & Risk</h3>
                        <ul class="nav-list">
                            <li class="nav-item">
                                <a href="add-merchant.html" class="nav-link" data-page="add-merchant">
                                    <i class="fas fa-plus-circle"></i>
                                    <span class="nav-text">Add Merchant</span>
                                    <span class="nav-badge new">NEW</span>
                                </a>
                            </li>
                            <li class="nav-item">
                                <a href="dashboard.html" class="nav-link" data-page="business-intelligence">
                                    <i class="fas fa-chart-line"></i>
                                    <span class="nav-text">Business Intelligence</span>
                                </a>
                            </li>
                            <li class="nav-item">
                                <a href="risk-dashboard.html" class="nav-link" data-page="risk-assessment">
                                    <i class="fas fa-exclamation-triangle"></i>
                                    <span class="nav-text">Risk Assessment</span>
                                </a>
                            </li>
                            <li class="nav-item">
                                <a href="enhanced-risk-indicators.html" class="nav-link" data-page="risk-indicators">
                                    <i class="fas fa-gauge-high"></i>
                                    <span class="nav-text">Risk Indicators</span>
                                </a>
                            </li>
                        </ul>
                    </div>
                    
                    <div class="nav-section">
                        <h3 class="nav-section-title">Compliance</h3>
                        <ul class="nav-list">
                            <li class="nav-item">
                                <a href="compliance-dashboard.html" class="nav-link" data-page="compliance-status">
                                    <i class="fas fa-clipboard-check"></i>
                                    <span class="nav-text">Compliance Status</span>
                                </a>
                            </li>
                            <li class="nav-item">
                                <a href="compliance-gap-analysis.html" class="nav-link" data-page="compliance-gaps">
                                    <i class="fas fa-search-minus"></i>
                                    <span class="nav-text">Gap Analysis</span>
                                    <span class="nav-badge new">NEW</span>
                                </a>
                            </li>
                            <li class="nav-item">
                                <a href="compliance-progress-tracking.html" class="nav-link" data-page="compliance-progress">
                                    <i class="fas fa-tasks"></i>
                                    <span class="nav-text">Progress Tracking</span>
                                </a>
                            </li>
                        </ul>
                    </div>
                    
                    <div class="nav-section">
                        <h3 class="nav-section-title">Merchant Management</h3>
                        <ul class="nav-list">
                            <li class="nav-item">
                                <a href="merchant-hub-integration.html" class="nav-link" data-page="merchant-hub">
                                    <i class="fas fa-sitemap"></i>
                                    <span class="nav-text">Merchant Hub</span>
                                    <span class="nav-badge new">NEW</span>
                                </a>
                            </li>
                            <li class="nav-item">
                                <a href="merchant-portfolio.html" class="nav-link" data-page="merchant-portfolio">
                                    <i class="fas fa-store"></i>
                                    <span class="nav-text">Merchant Portfolio</span>
                                </a>
                            </li>
                            <li class="nav-item">
                                <a href="risk-assessment-portfolio.html" class="nav-link" data-page="risk-assessment-portfolio">
                                    <i class="fas fa-shield-alt"></i>
                                    <span class="nav-text">Risk Assessment Portfolio</span>
                                </a>
                            </li>
                            <li class="nav-item">
                                <a href="merchant-detail.html" class="nav-link" data-page="merchant-detail">
                                    <i class="fas fa-user-tie"></i>
                                    <span class="nav-text">Merchant Detail</span>
                                </a>
                            </li>
                        </ul>
                    </div>
                    
                    <div class="nav-section">
                        <h3 class="nav-section-title">Market Intelligence</h3>
                        <ul class="nav-list">
                            <li class="nav-item">
                                <a href="market-analysis-dashboard.html" class="nav-link" data-page="market-analysis">
                                    <i class="fas fa-chart-bar"></i>
                                    <span class="nav-text">Market Analysis</span>
                                </a>
                            </li>
                            <li class="nav-item">
                                <a href="competitive-analysis-dashboard.html" class="nav-link" data-page="competitive-analysis">
                                    <i class="fas fa-users"></i>
                                    <span class="nav-text">Competitive Analysis</span>
                                </a>
                            </li>
                            <li class="nav-item">
                                <a href="business-growth-analytics.html" class="nav-link" data-page="growth-analytics">
                                    <i class="fas fa-trending-up"></i>
                                    <span class="nav-text">Growth Analytics</span>
                                </a>
                            </li>
                        </ul>
                    </div>
                </div>
                
                <div class="sidebar-footer">
                    <div class="nav-status">
                        <span class="status-indicator live"></span>
                        <span class="status-text">Live</span>
                    </div>
                </div>
            </div>
            
            <div class="main-content-wrapper">
                <div class="main-content" id="mainContent">
                    <!-- Main content will be inserted here -->
                </div>
            </div>
        `;

        this.navigationHTML = navigationHTML;
    }

    addNavigationToPage() {
        // Check if navigation already exists
        if (document.querySelector('.kyb-sidebar')) {
            return;
        }

        // Create navigation element
        const navElement = document.createElement('div');
        navElement.innerHTML = this.navigationHTML;
        
        // Get the sidebar and main content wrapper
        const sidebar = navElement.querySelector('.kyb-sidebar');
        const mainContentWrapper = navElement.querySelector('.main-content-wrapper');
        const mainContent = navElement.querySelector('.main-content');

        // Move existing body content to main content area
        const existingContent = document.body.innerHTML;
        mainContent.innerHTML = existingContent;

        // Clear body and add new structure
        document.body.innerHTML = '';
        document.body.appendChild(sidebar);
        document.body.appendChild(mainContentWrapper);

        // Add navigation styles
        this.addNavigationStyles();
    }

    addNavigationStyles() {
        const styles = `
            <style>
                /* Sidebar Layout */
                body {
                    margin: 0;
                    padding: 0;
                    display: flex;
                    min-height: 100vh;
                }

                .kyb-sidebar {
                    width: 280px;
                    background: rgba(255, 255, 255, 0.95);
                    backdrop-filter: blur(10px);
                    border-right: 1px solid rgba(0, 0, 0, 0.1);
                    position: fixed;
                    top: 0;
                    left: 0;
                    height: 100vh;
                    z-index: 1000;
                    box-shadow: 2px 0 20px rgba(0, 0, 0, 0.1);
                    display: flex;
                    flex-direction: column;
                    transition: transform 0.3s ease;
                }

                .sidebar-header {
                    padding: 20px;
                    border-bottom: 1px solid rgba(0, 0, 0, 0.1);
                    display: flex;
                    align-items: center;
                    justify-content: space-between;
                }

                .brand-link {
                    display: flex;
                    align-items: center;
                    gap: 12px;
                    text-decoration: none;
                    color: #2c3e50;
                    font-size: 1.3rem;
                    font-weight: 700;
                    transition: color 0.3s ease;
                }

                .brand-link:hover {
                    color: #3498db;
                }

                .brand-link i {
                    font-size: 1.5rem;
                    color: #3498db;
                }

                .brand-text {
                    white-space: nowrap;
                }

                .sidebar-toggle {
                    display: none;
                    background: none;
                    border: none;
                    font-size: 1.2rem;
                    color: #5a6c7d;
                    cursor: pointer;
                    padding: 8px;
                    border-radius: 8px;
                    transition: all 0.3s ease;
                }

                .sidebar-toggle:hover {
                    background: rgba(52, 152, 219, 0.1);
                    color: #3498db;
                }

                .sidebar-content {
                    flex: 1;
                    padding: 20px 0;
                    overflow-y: auto;
                }

                .nav-section {
                    margin-bottom: 30px;
                }

                .nav-section-title {
                    font-size: 0.75rem;
                    font-weight: 600;
                    color: #7f8c8d;
                    text-transform: uppercase;
                    letter-spacing: 0.5px;
                    margin: 0 20px 15px;
                }

                .nav-list {
                    list-style: none;
                    margin: 0;
                    padding: 0;
                }

                .nav-item {
                    margin: 0 10px 5px;
                }

                .nav-link {
                    display: flex;
                    align-items: center;
                    gap: 12px;
                    padding: 12px 16px;
                    text-decoration: none;
                    color: #5a6c7d;
                    font-weight: 500;
                    border-radius: 10px;
                    transition: all 0.3s ease;
                    position: relative;
                }

                .nav-link:hover {
                    background: rgba(52, 152, 219, 0.1);
                    color: #3498db;
                    transform: translateX(5px);
                }

                .nav-link.active {
                    background: linear-gradient(135deg, #3498db, #2980b9);
                    color: white;
                    box-shadow: 0 4px 12px rgba(52, 152, 219, 0.3);
                }

                .nav-link i {
                    font-size: 1.1rem;
                    width: 20px;
                    text-align: center;
                }

                .nav-text {
                    flex: 1;
                }

                .nav-badge {
                    padding: 2px 6px;
                    border-radius: 10px;
                    font-size: 0.65rem;
                    font-weight: 600;
                    text-transform: uppercase;
                    letter-spacing: 0.5px;
                }

                .nav-badge.new {
                    background: linear-gradient(135deg, #e74c3c, #c0392b);
                    color: white;
                }

                .sidebar-footer {
                    padding: 20px;
                    border-top: 1px solid rgba(0, 0, 0, 0.1);
                }

                .nav-status {
                    display: flex;
                    align-items: center;
                    gap: 8px;
                    padding: 8px 12px;
                    background: rgba(46, 204, 113, 0.1);
                    border-radius: 20px;
                    border: 1px solid rgba(46, 204, 113, 0.2);
                }

                .status-indicator {
                    width: 8px;
                    height: 8px;
                    border-radius: 50%;
                    animation: pulse 2s infinite;
                }

                .status-indicator.live {
                    background: #2ecc71;
                }

                .status-text {
                    font-size: 0.8rem;
                    font-weight: 600;
                    color: #27ae60;
                }

                @keyframes pulse {
                    0% { opacity: 1; }
                    50% { opacity: 0.5; }
                    100% { opacity: 1; }
                }

                /* Main Content Area */
                .main-content-wrapper {
                    flex: 1;
                    margin-left: 280px;
                    min-height: 100vh;
                    background: transparent;
                }

                .main-content {
                    width: 100%;
                    min-height: 100vh;
                }

                /* Enhanced Mobile Responsive Design */
                @media (max-width: 1024px) {
                    .kyb-sidebar {
                        transform: translateX(-100%);
                        transition: transform 0.3s ease-in-out;
                    }

                    .kyb-sidebar.open {
                        transform: translateX(0);
                    }

                    .main-content-wrapper {
                        margin-left: 0;
                    }

                    .sidebar-toggle {
                        display: block;
                        min-width: 44px;
                        min-height: 44px;
                        touch-action: manipulation;
                        -webkit-tap-highlight-color: rgba(0, 0, 0, 0.1);
                    }

                    .brand-text {
                        display: none;
                    }

                    .kyb-sidebar {
                        width: 260px;
                    }

                    /* Enhanced touch interactions */
                    .nav-link {
                        min-height: 44px;
                        touch-action: manipulation;
                        -webkit-tap-highlight-color: rgba(0, 0, 0, 0.1);
                        transition: background-color 0.2s ease;
                    }

                    .nav-link:active {
                        background-color: rgba(52, 152, 219, 0.1);
                        transform: scale(0.98);
                    }
                }

                @media (max-width: 768px) {
                    .kyb-sidebar {
                        width: 100%;
                        max-width: 300px;
                    }

                    .sidebar-header {
                        padding: 16px 20px;
                        min-height: 60px;
                    }

                    .brand-link {
                        font-size: 1.2rem;
                        min-height: 44px;
                        display: flex;
                        align-items: center;
                        gap: 12px;
                    }

                    .brand-link i {
                        font-size: 1.5rem;
                    }

                    .nav-section {
                        margin-bottom: 25px;
                    }

                    .nav-section-title {
                        margin: 0 15px 12px;
                        font-size: 0.9rem;
                        padding: 12px 20px 8px;
                    }

                    .nav-item {
                        margin: 0 5px 3px;
                    }

                    .nav-link {
                        padding: 16px 20px;
                        font-size: 16px;
                        min-height: 48px;
                    }

                    .nav-text {
                        font-size: 1rem;
                    }

                    /* Enhanced mobile sidebar toggle */
                    .sidebar-toggle {
                        min-width: 48px;
                        min-height: 48px;
                        font-size: 1.3rem;
                        border-radius: 12px;
                    }
                }

                @media (max-width: 480px) {
                    .kyb-sidebar {
                        width: 100%;
                        max-width: 100%;
                    }

                    .sidebar-header {
                        padding: 20px;
                        min-height: 70px;
                    }

                    .brand-link {
                        font-size: 1.3rem;
                    }

                    .brand-link i {
                        font-size: 1.8rem;
                    }

                    .nav-section-title {
                        font-size: 1rem;
                        padding: 16px 20px 12px;
                    }

                    .nav-link {
                        padding: 18px 20px;
                        min-height: 52px;
                    }

                    .nav-text {
                        font-size: 1.1rem;
                    }

                    .sidebar-toggle {
                        min-width: 52px;
                        min-height: 52px;
                        font-size: 1.4rem;
                    }
                }

                /* Landscape mobile optimization */
                @media (max-width: 768px) and (orientation: landscape) {
                    .kyb-sidebar {
                        width: 280px;
                        max-width: 280px;
                    }

                    .nav-link {
                        padding: 12px 20px;
                        min-height: 44px;
                    }

                    .nav-text {
                        font-size: 0.95rem;
                    }

                    .sidebar-header {
                        padding: 12px 20px;
                        min-height: 50px;
                    }
                }

                /* Overlay for mobile */
                .sidebar-overlay {
                    position: fixed;
                    top: 0;
                    left: 0;
                    width: 100%;
                    height: 100%;
                    background: rgba(0, 0, 0, 0.5);
                    z-index: 999;
                    display: none;
                }

                .sidebar-overlay.active {
                    display: block;
                }
            </style>
        `;

        // Add styles to head
        const styleElement = document.createElement('div');
        styleElement.innerHTML = styles;
        document.head.appendChild(styleElement.firstElementChild);
    }

    setActivePage() {
        const activeLink = document.querySelector(`.nav-link[data-page="${this.currentPage}"]`);
        if (activeLink) {
            activeLink.classList.add('active');
        }
    }

    bindEvents() {
        // Sidebar toggle for mobile
        const sidebarToggle = document.getElementById('sidebarToggle');
        const sidebar = document.querySelector('.kyb-sidebar');

        if (sidebarToggle && sidebar) {
            sidebarToggle.addEventListener('click', () => {
                sidebar.classList.toggle('open');
                const icon = sidebarToggle.querySelector('i');
                if (sidebar.classList.contains('open')) {
                    icon.className = 'fas fa-times';
                    this.createOverlay();
                } else {
                    icon.className = 'fas fa-bars';
                    this.removeOverlay();
                }
            });
        }

        // Close sidebar when clicking outside on mobile
        document.addEventListener('click', (e) => {
            if (window.innerWidth <= 1024 && sidebar && sidebar.classList.contains('open')) {
                if (!sidebar.contains(e.target) && !sidebarToggle.contains(e.target)) {
                    sidebar.classList.remove('open');
                    const icon = sidebarToggle.querySelector('i');
                    icon.className = 'fas fa-bars';
                    this.removeOverlay();
                }
            }
        });

        // Handle window resize
        window.addEventListener('resize', () => {
            if (window.innerWidth > 1024) {
                sidebar.classList.remove('open');
                const icon = sidebarToggle.querySelector('i');
                if (icon) icon.className = 'fas fa-bars';
                this.removeOverlay();
            }
        });

        // Smooth scrolling for anchor links
        document.querySelectorAll('a[href^="#"]').forEach(anchor => {
            anchor.addEventListener('click', function (e) {
                e.preventDefault();
                const target = document.querySelector(this.getAttribute('href'));
                if (target) {
                    target.scrollIntoView({
                        behavior: 'smooth',
                        block: 'start'
                    });
                }
            });
        });
    }

    createOverlay() {
        if (document.querySelector('.sidebar-overlay')) return;
        
        const overlay = document.createElement('div');
        overlay.className = 'sidebar-overlay active';
        document.body.appendChild(overlay);
        
        overlay.addEventListener('click', () => {
            const sidebar = document.querySelector('.kyb-sidebar');
            const sidebarToggle = document.getElementById('sidebarToggle');
            if (sidebar && sidebarToggle) {
                sidebar.classList.remove('open');
                const icon = sidebarToggle.querySelector('i');
                icon.className = 'fas fa-bars';
                this.removeOverlay();
            }
        });
    }

    removeOverlay() {
        const overlay = document.querySelector('.sidebar-overlay');
        if (overlay) {
            overlay.remove();
        }
    }

    // Method to update navigation when switching pages
    updateActivePage(page) {
        // Remove active class from all links
        document.querySelectorAll('.nav-link').forEach(link => {
            link.classList.remove('active');
        });

        // Add active class to current page
        const activeLink = document.querySelector(`.nav-link[data-page="${page}"]`);
        if (activeLink) {
            activeLink.classList.add('active');
        }
    }

    // Method to show notification badge
    showNotification(page, message) {
        const link = document.querySelector(`.nav-link[data-page="${page}"]`);
        if (link) {
            const badge = link.querySelector('.nav-badge');
            if (!badge) {
                const notificationBadge = document.createElement('span');
                notificationBadge.className = 'nav-badge notification';
                notificationBadge.textContent = '!';
                notificationBadge.style.background = 'linear-gradient(135deg, #e74c3c, #c0392b)';
                notificationBadge.style.color = 'white';
                link.appendChild(notificationBadge);
            }
        }
    }

    // Method to hide notification badge
    hideNotification(page) {
        const link = document.querySelector(`.nav-link[data-page="${page}"]`);
        if (link) {
            const badge = link.querySelector('.nav-badge.notification');
            if (badge) {
                badge.remove();
            }
        }
    }
}

// Initialize navigation when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.kybNavigation = new KYBNavigation();
});

// Export for use in other scripts
if (typeof module !== 'undefined' && module.exports) {
    module.exports = KYBNavigation;
}
