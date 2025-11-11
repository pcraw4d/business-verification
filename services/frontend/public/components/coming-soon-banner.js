/**
 * Coming Soon Banner Component
 * Displays feature indicators for coming soon features with descriptions and timelines
 * Integrates with placeholder service to show mock data warnings
 */
class ComingSoonBanner {
    constructor(options = {}) {
        this.container = options.container || document.body;
        this.apiBaseUrl = options.apiBaseUrl || '/api/v1';
        this.featureId = options.featureId || null;
        this.category = options.category || null;
        this.status = options.status || 'coming_soon';
        this.showMockDataWarning = options.showMockDataWarning !== false;
        this.autoRefresh = options.autoRefresh || false;
        this.refreshInterval = options.refreshInterval || 300000; // 5 minutes
        this.isVisible = false;
        this.features = [];
        this.currentFeature = null;
        this.refreshTimer = null;
        
        // Event callbacks
        this.onFeatureClick = options.onFeatureClick || null;
        this.onBannerClose = options.onBannerClose || null;
        this.onMockDataWarning = options.onMockDataWarning || null;
        
        this.init();
    }

    init() {
        this.createBannerInterface();
        this.bindEvents();
        this.loadFeatureData();
        
        if (this.autoRefresh) {
            this.startAutoRefresh();
        }
    }

    createBannerInterface() {
        const bannerHTML = `
            <div class="coming-soon-banner" id="comingSoonBanner" style="display: none;">
                <div class="banner-content">
                    <div class="banner-header">
                        <div class="banner-icon">
                            <i class="fas fa-rocket"></i>
                        </div>
                        <div class="banner-title">
                            <h3 class="banner-title-text">Coming Soon</h3>
                            <p class="banner-subtitle" id="bannerSubtitle">Exciting new features are on the way!</p>
                        </div>
                        <div class="banner-actions">
                            <button class="banner-close-btn" id="bannerCloseBtn" title="Close banner">
                                <i class="fas fa-times"></i>
                            </button>
                        </div>
                    </div>
                    
                    <div class="banner-body">
                        <div class="feature-info" id="featureInfo">
                            <div class="feature-details">
                                <h4 class="feature-name" id="featureName">Loading...</h4>
                                <p class="feature-description" id="featureDescription">Loading feature details...</p>
                                <div class="feature-meta">
                                    <div class="feature-category" id="featureCategory">
                                        <i class="fas fa-tag"></i>
                                        <span class="category-text">Loading...</span>
                                    </div>
                                    <div class="feature-priority" id="featurePriority">
                                        <i class="fas fa-star"></i>
                                        <span class="priority-text">Loading...</span>
                                    </div>
                                    <div class="feature-eta" id="featureETA">
                                        <i class="fas fa-calendar-alt"></i>
                                        <span class="eta-text">Loading...</span>
                                    </div>
                                </div>
                            </div>
                            
                            <div class="feature-actions">
                                <button class="btn btn-primary" id="notifyMeBtn">
                                    <i class="fas fa-bell"></i>
                                    Notify Me
                                </button>
                                <button class="btn btn-outline" id="learnMoreBtn">
                                    <i class="fas fa-info-circle"></i>
                                    Learn More
                                </button>
                            </div>
                        </div>
                        
                        <div class="mock-data-warning" id="mockDataWarning" style="display: none;">
                            <div class="warning-icon">
                                <i class="fas fa-exclamation-triangle"></i>
                            </div>
                            <div class="warning-content">
                                <h5 class="warning-title">Mock Data Warning</h5>
                                <p class="warning-text">This feature is currently using mock data for demonstration purposes.</p>
                                <div class="warning-actions">
                                    <button class="btn btn-warning" id="dismissWarningBtn">
                                        <i class="fas fa-check"></i>
                                        Dismiss
                                    </button>
                                </div>
                            </div>
                        </div>
                        
                        <div class="progress-indicator" id="progressIndicator" style="display: none;">
                            <div class="progress-bar">
                                <div class="progress-fill" id="progressFill"></div>
                            </div>
                            <div class="progress-text" id="progressText">Development Progress</div>
                        </div>
                    </div>
                    
                    <div class="banner-footer">
                        <div class="banner-stats" id="bannerStats">
                            <div class="stat-item">
                                <span class="stat-label">Features Coming Soon:</span>
                                <span class="stat-value" id="comingSoonCount">0</span>
                            </div>
                            <div class="stat-item">
                                <span class="stat-label">In Development:</span>
                                <span class="stat-value" id="inDevelopmentCount">0</span>
                            </div>
                        </div>
                        <div class="banner-links">
                            <a href="#" class="banner-link" id="viewAllFeaturesLink">
                                <i class="fas fa-list"></i>
                                View All Features
                            </a>
                        </div>
                    </div>
                </div>
            </div>
        `;

        // If container is document.body, create a wrapper div to avoid clearing the body
        // This prevents the banner from clearing all page content
        if (this.container === document.body) {
            let wrapper = document.getElementById('coming-soon-banner-wrapper');
            if (!wrapper) {
                wrapper = document.createElement('div');
                wrapper.id = 'coming-soon-banner-wrapper';
                wrapper.style.cssText = 'position: fixed; top: 0; right: 0; z-index: 10000;';
                document.body.appendChild(wrapper);
            }
            this.container = wrapper;
        }
        
        // Only set innerHTML on the container (which is now the wrapper, not body)
        this.container.innerHTML = bannerHTML;
        this.addStyles();
    }

    addStyles() {
        const styles = `
            <style>
                .coming-soon-banner {
                    position: fixed;
                    top: 20px;
                    right: 20px;
                    width: 400px;
                    max-width: calc(100vw - 40px);
                    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                    border-radius: 16px;
                    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
                    z-index: 10000;
                    animation: slideInRight 0.5s ease-out;
                    backdrop-filter: blur(10px);
                    border: 1px solid rgba(255, 255, 255, 0.1);
                }

                .banner-content {
                    padding: 24px;
                    color: white;
                }

                .banner-header {
                    display: flex;
                    align-items: flex-start;
                    gap: 16px;
                    margin-bottom: 20px;
                }

                .banner-icon {
                    width: 48px;
                    height: 48px;
                    background: rgba(255, 255, 255, 0.2);
                    border-radius: 12px;
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    font-size: 1.5rem;
                    flex-shrink: 0;
                }

                .banner-title {
                    flex: 1;
                }

                .banner-title-text {
                    margin: 0 0 4px 0;
                    font-size: 1.3rem;
                    font-weight: 700;
                    color: white;
                }

                .banner-subtitle {
                    margin: 0;
                    font-size: 0.9rem;
                    color: rgba(255, 255, 255, 0.8);
                    line-height: 1.4;
                }

                .banner-actions {
                    display: flex;
                    gap: 8px;
                }

                .banner-close-btn {
                    width: 32px;
                    height: 32px;
                    background: rgba(255, 255, 255, 0.1);
                    border: none;
                    border-radius: 8px;
                    color: white;
                    cursor: pointer;
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    transition: all 0.3s ease;
                    font-size: 0.9rem;
                }

                .banner-close-btn:hover {
                    background: rgba(255, 255, 255, 0.2);
                    transform: scale(1.1);
                }

                .banner-body {
                    margin-bottom: 20px;
                }

                .feature-info {
                    background: rgba(255, 255, 255, 0.1);
                    border-radius: 12px;
                    padding: 20px;
                    margin-bottom: 16px;
                }

                .feature-details {
                    margin-bottom: 16px;
                }

                .feature-name {
                    margin: 0 0 8px 0;
                    font-size: 1.1rem;
                    font-weight: 600;
                    color: white;
                }

                .feature-description {
                    margin: 0 0 16px 0;
                    font-size: 0.9rem;
                    color: rgba(255, 255, 255, 0.9);
                    line-height: 1.5;
                }

                .feature-meta {
                    display: flex;
                    flex-wrap: wrap;
                    gap: 16px;
                    margin-bottom: 16px;
                }

                .feature-category,
                .feature-priority,
                .feature-eta {
                    display: flex;
                    align-items: center;
                    gap: 6px;
                    font-size: 0.8rem;
                    color: rgba(255, 255, 255, 0.8);
                }

                .feature-category i,
                .feature-priority i,
                .feature-eta i {
                    font-size: 0.7rem;
                    opacity: 0.7;
                }

                .feature-actions {
                    display: flex;
                    gap: 12px;
                }

                .btn {
                    padding: 10px 16px;
                    border: none;
                    border-radius: 8px;
                    font-size: 0.85rem;
                    font-weight: 600;
                    cursor: pointer;
                    transition: all 0.3s ease;
                    display: flex;
                    align-items: center;
                    gap: 6px;
                    text-decoration: none;
                    flex: 1;
                    justify-content: center;
                }

                .btn:disabled {
                    opacity: 0.6;
                    cursor: not-allowed;
                }

                .btn-primary {
                    background: rgba(255, 255, 255, 0.2);
                    color: white;
                    border: 1px solid rgba(255, 255, 255, 0.3);
                }

                .btn-primary:hover:not(:disabled) {
                    background: rgba(255, 255, 255, 0.3);
                    transform: translateY(-2px);
                }

                .btn-outline {
                    background: transparent;
                    color: white;
                    border: 1px solid rgba(255, 255, 255, 0.3);
                }

                .btn-outline:hover:not(:disabled) {
                    background: rgba(255, 255, 255, 0.1);
                    transform: translateY(-2px);
                }

                .btn-warning {
                    background: rgba(255, 193, 7, 0.2);
                    color: #ffc107;
                    border: 1px solid rgba(255, 193, 7, 0.3);
                }

                .btn-warning:hover:not(:disabled) {
                    background: rgba(255, 193, 7, 0.3);
                    transform: translateY(-2px);
                }

                .mock-data-warning {
                    background: rgba(255, 193, 7, 0.1);
                    border: 1px solid rgba(255, 193, 7, 0.3);
                    border-radius: 12px;
                    padding: 16px;
                    margin-bottom: 16px;
                    display: flex;
                    gap: 12px;
                    align-items: flex-start;
                }

                .warning-icon {
                    color: #ffc107;
                    font-size: 1.2rem;
                    margin-top: 2px;
                }

                .warning-content {
                    flex: 1;
                }

                .warning-title {
                    margin: 0 0 4px 0;
                    font-size: 0.9rem;
                    font-weight: 600;
                    color: #ffc107;
                }

                .warning-text {
                    margin: 0 0 12px 0;
                    font-size: 0.8rem;
                    color: rgba(255, 193, 7, 0.9);
                    line-height: 1.4;
                }

                .warning-actions {
                    display: flex;
                    gap: 8px;
                }

                .progress-indicator {
                    background: rgba(255, 255, 255, 0.1);
                    border-radius: 12px;
                    padding: 16px;
                    margin-bottom: 16px;
                }

                .progress-bar {
                    width: 100%;
                    height: 8px;
                    background: rgba(255, 255, 255, 0.2);
                    border-radius: 4px;
                    overflow: hidden;
                    margin-bottom: 8px;
                }

                .progress-fill {
                    height: 100%;
                    background: linear-gradient(90deg, #4CAF50, #8BC34A);
                    border-radius: 4px;
                    transition: width 0.5s ease;
                    width: 0%;
                }

                .progress-text {
                    font-size: 0.8rem;
                    color: rgba(255, 255, 255, 0.8);
                    text-align: center;
                }

                .banner-footer {
                    border-top: 1px solid rgba(255, 255, 255, 0.1);
                    padding-top: 16px;
                    display: flex;
                    justify-content: space-between;
                    align-items: center;
                    flex-wrap: wrap;
                    gap: 12px;
                }

                .banner-stats {
                    display: flex;
                    gap: 16px;
                    flex-wrap: wrap;
                }

                .stat-item {
                    display: flex;
                    align-items: center;
                    gap: 6px;
                    font-size: 0.8rem;
                }

                .stat-label {
                    color: rgba(255, 255, 255, 0.7);
                }

                .stat-value {
                    color: white;
                    font-weight: 600;
                }

                .banner-links {
                    display: flex;
                    gap: 12px;
                }

                .banner-link {
                    color: rgba(255, 255, 255, 0.8);
                    text-decoration: none;
                    font-size: 0.8rem;
                    display: flex;
                    align-items: center;
                    gap: 4px;
                    transition: color 0.3s ease;
                }

                .banner-link:hover {
                    color: white;
                }

                /* Animations */
                @keyframes slideInRight {
                    from {
                        opacity: 0;
                        transform: translateX(100%);
                    }
                    to {
                        opacity: 1;
                        transform: translateX(0);
                    }
                }

                @keyframes slideOutRight {
                    from {
                        opacity: 1;
                        transform: translateX(0);
                    }
                    to {
                        opacity: 0;
                        transform: translateX(100%);
                    }
                }

                .banner-closing {
                    animation: slideOutRight 0.3s ease-in forwards;
                }

                /* Responsive Design */
                @media (max-width: 768px) {
                    .coming-soon-banner {
                        top: 10px;
                        right: 10px;
                        left: 10px;
                        width: auto;
                        max-width: none;
                    }

                    .banner-content {
                        padding: 20px;
                    }

                    .banner-header {
                        gap: 12px;
                    }

                    .banner-icon {
                        width: 40px;
                        height: 40px;
                        font-size: 1.2rem;
                    }

                    .banner-title-text {
                        font-size: 1.1rem;
                    }

                    .feature-meta {
                        flex-direction: column;
                        gap: 8px;
                    }

                    .feature-actions {
                        flex-direction: column;
                    }

                    .banner-footer {
                        flex-direction: column;
                        align-items: flex-start;
                        gap: 12px;
                    }

                    .banner-stats {
                        flex-direction: column;
                        gap: 8px;
                    }
                }

                /* Dark mode support */
                @media (prefers-color-scheme: dark) {
                    .coming-soon-banner {
                        background: linear-gradient(135deg, #2c3e50 0%, #34495e 100%);
                    }
                }

                /* High contrast mode */
                @media (prefers-contrast: high) {
                    .coming-soon-banner {
                        border: 2px solid white;
                    }

                    .banner-close-btn {
                        border: 1px solid white;
                    }
                }

                /* Reduced motion */
                @media (prefers-reduced-motion: reduce) {
                    .coming-soon-banner {
                        animation: none;
                    }

                    .banner-closing {
                        animation: none;
                    }

                    .btn:hover:not(:disabled) {
                        transform: none;
                    }

                    .banner-close-btn:hover {
                        transform: none;
                    }
                }
            </style>
        `;

        // Add styles to head if not already added
        if (!document.querySelector('#coming-soon-banner-styles')) {
            const styleElement = document.createElement('style');
            styleElement.id = 'coming-soon-banner-styles';
            styleElement.textContent = styles;
            document.head.appendChild(styleElement);
        }
    }

    bindEvents() {
        const bannerCloseBtn = document.getElementById('bannerCloseBtn');
        const notifyMeBtn = document.getElementById('notifyMeBtn');
        const learnMoreBtn = document.getElementById('learnMoreBtn');
        const dismissWarningBtn = document.getElementById('dismissWarningBtn');
        const viewAllFeaturesLink = document.getElementById('viewAllFeaturesLink');

        // Close banner
        bannerCloseBtn.addEventListener('click', () => {
            this.hide();
        });

        // Notify me button
        notifyMeBtn.addEventListener('click', () => {
            this.handleNotifyMe();
        });

        // Learn more button
        learnMoreBtn.addEventListener('click', () => {
            this.handleLearnMore();
        });

        // Dismiss warning
        dismissWarningBtn.addEventListener('click', () => {
            this.dismissMockDataWarning();
        });

        // View all features link
        viewAllFeaturesLink.addEventListener('click', (e) => {
            e.preventDefault();
            this.handleViewAllFeatures();
        });

        // Click outside to close
        document.addEventListener('click', (e) => {
            const banner = document.getElementById('comingSoonBanner');
            if (this.isVisible && banner && !banner.contains(e.target)) {
                // Don't auto-close on outside click for better UX
                // this.hide();
            }
        });

        // Keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            if (this.isVisible) {
                if (e.key === 'Escape') {
                    this.hide();
                }
            }
        });
    }

    async loadFeatureData() {
        try {
            let features = [];
            
            if (this.featureId) {
                // Load specific feature
                const feature = await this.getFeature(this.featureId);
                if (feature) {
                    features = [feature];
                }
            } else if (this.category) {
                // Load features by category
                features = await this.getFeaturesByCategory(this.category);
            } else {
                // Load features by status
                features = await this.getFeaturesByStatus(this.status);
            }

            if (features.length > 0) {
                this.features = features;
                this.currentFeature = features[0]; // Show first feature
                this.updateBannerContent();
                this.updateBannerStats();
                
                // Show mock data warning if applicable
                if (this.showMockDataWarning && this.currentFeature.mock_data) {
                    this.showMockDataWarning();
                }
                
                this.show();
            } else {
                this.hide();
            }
        } catch (error) {
            console.error('Error loading feature data:', error);
            this.hide();
        }
    }

    async getFeature(featureId) {
        try {
            const response = await fetch(`${this.apiBaseUrl}/features/${featureId}`, {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${this.getAuthToken()}`
                }
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            return data.success ? data.data : null;
        } catch (error) {
            console.error('Error fetching feature:', error);
            return null;
        }
    }

    async getFeaturesByCategory(category) {
        try {
            const response = await fetch(`${this.apiBaseUrl}/features/category/${category}`, {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${this.getAuthToken()}`
                }
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            return data.success ? data.data.features : [];
        } catch (error) {
            console.error('Error fetching features by category:', error);
            return [];
        }
    }

    async getFeaturesByStatus(status) {
        try {
            const response = await fetch(`${this.apiBaseUrl}/features/status/${status}`, {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${this.getAuthToken()}`
                }
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            // Check if response is JSON before parsing
            const contentType = response.headers.get('content-type');
            if (!contentType || !contentType.includes('application/json')) {
                console.warn('⚠️ API returned non-JSON response for features, using empty array');
                console.warn('🔍 Response details:', {
                    status: response.status,
                    statusText: response.statusText,
                    contentType: contentType,
                    url: response.url
                });
                // Try to read response text for debugging (clone response first)
                try {
                    const clonedResponse = response.clone();
                    const text = await clonedResponse.text();
                    console.warn('🔍 Response body (first 500 chars):', text.substring(0, 500));
                } catch (e) {
                    console.warn('🔍 Could not read response body:', e.message);
                }
                return [];
            }

            const data = await response.json();
            return data.success ? data.data.features : [];
        } catch (error) {
            console.warn('⚠️ Error fetching features by status, using empty array:', error.message);
            return [];
        }
    }

    async getFeatureStatistics() {
        try {
            const response = await fetch(`${this.apiBaseUrl}/features/statistics`, {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${this.getAuthToken()}`
                }
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            return data.success ? data.data.statistics : null;
        } catch (error) {
            console.error('Error fetching feature statistics:', error);
            return null;
        }
    }

    updateBannerContent() {
        if (!this.currentFeature) return;

        const feature = this.currentFeature;
        
        // Update feature details
        document.getElementById('featureName').textContent = feature.name;
        document.getElementById('featureDescription').textContent = feature.description;
        
        // Update category
        const categoryElement = document.getElementById('featureCategory');
        categoryElement.querySelector('.category-text').textContent = this.formatCategory(feature.category);
        
        // Update priority
        const priorityElement = document.getElementById('featurePriority');
        priorityElement.querySelector('.priority-text').textContent = this.formatPriority(feature.priority);
        
        // Update ETA
        const etaElement = document.getElementById('featureETA');
        etaElement.querySelector('.eta-text').textContent = this.formatETA(feature.eta);
        
        // Update subtitle based on status
        const subtitle = document.getElementById('bannerSubtitle');
        subtitle.textContent = this.getStatusSubtitle(feature.status);
        
        // Show/hide progress indicator based on status
        const progressIndicator = document.getElementById('progressIndicator');
        if (feature.status === 'in_development') {
            progressIndicator.style.display = 'block';
            this.updateProgressIndicator(feature);
        } else {
            progressIndicator.style.display = 'none';
        }
    }

    updateBannerStats() {
        // This would typically come from the statistics API
        // For now, we'll use the loaded features to calculate stats
        const comingSoonCount = this.features.filter(f => f.status === 'coming_soon').length;
        const inDevelopmentCount = this.features.filter(f => f.status === 'in_development').length;
        
        document.getElementById('comingSoonCount').textContent = comingSoonCount;
        document.getElementById('inDevelopmentCount').textContent = inDevelopmentCount;
    }

    updateProgressIndicator(feature) {
        // Calculate progress based on ETA and current date
        if (!feature.eta) return;
        
        const now = new Date();
        const eta = new Date(feature.eta);
        const created = new Date(feature.created_at);
        
        const totalTime = eta.getTime() - created.getTime();
        const elapsedTime = now.getTime() - created.getTime();
        
        let progress = Math.min(Math.max((elapsedTime / totalTime) * 100, 0), 100);
        
        // For in_development status, show at least 25% progress
        if (feature.status === 'in_development' && progress < 25) {
            progress = 25;
        }
        
        const progressFill = document.getElementById('progressFill');
        const progressText = document.getElementById('progressText');
        
        progressFill.style.width = `${progress}%`;
        progressText.textContent = `Development Progress: ${Math.round(progress)}%`;
    }

    showMockDataWarning() {
        const warning = document.getElementById('mockDataWarning');
        warning.style.display = 'flex';
        
        // Call callback if provided
        if (this.onMockDataWarning) {
            this.onMockDataWarning(this.currentFeature);
        }
    }

    dismissMockDataWarning() {
        const warning = document.getElementById('mockDataWarning');
        warning.style.display = 'none';
    }

    show() {
        const banner = document.getElementById('comingSoonBanner');
        if (banner) {
            banner.style.display = 'block';
            this.isVisible = true;
            
            // Add entrance animation
            banner.style.animation = 'slideInRight 0.5s ease-out';
        }
    }

    hide() {
        const banner = document.getElementById('comingSoonBanner');
        if (banner) {
            banner.classList.add('banner-closing');
            
            setTimeout(() => {
                banner.style.display = 'none';
                banner.classList.remove('banner-closing');
                this.isVisible = false;
                
                // Call callback if provided
                if (this.onBannerClose) {
                    this.onBannerClose();
                }
            }, 300);
        }
    }

    handleNotifyMe() {
        // Implement notification signup logic
        console.log('Notify me clicked for feature:', this.currentFeature?.name);
        
        // Show success message
        this.showNotification('You will be notified when this feature is available!', 'success');
        
        // Call callback if provided
        if (this.onFeatureClick) {
            this.onFeatureClick('notify', this.currentFeature);
        }
    }

    handleLearnMore() {
        // Implement learn more logic
        console.log('Learn more clicked for feature:', this.currentFeature?.name);
        
        // Call callback if provided
        if (this.onFeatureClick) {
            this.onFeatureClick('learn_more', this.currentFeature);
        }
    }

    handleViewAllFeatures() {
        // Implement view all features logic
        console.log('View all features clicked');
        
        // Call callback if provided
        if (this.onFeatureClick) {
            this.onFeatureClick('view_all', null);
        }
    }

    startAutoRefresh() {
        this.refreshTimer = setInterval(() => {
            this.loadFeatureData();
        }, this.refreshInterval);
    }

    stopAutoRefresh() {
        if (this.refreshTimer) {
            clearInterval(this.refreshTimer);
            this.refreshTimer = null;
        }
    }

    showNotification(message, type = 'info') {
        const notification = document.createElement('div');
        notification.className = `notification notification-${type}`;
        notification.style.cssText = `
            position: fixed;
            top: 20px;
            left: 50%;
            transform: translateX(-50%);
            background: ${type === 'success' ? '#4CAF50' : type === 'error' ? '#f44336' : '#2196F3'};
            color: white;
            padding: 16px 24px;
            border-radius: 8px;
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
            z-index: 10001;
            animation: slideInDown 0.3s ease-out;
            max-width: 400px;
            text-align: center;
        `;
        notification.textContent = message;
        
        document.body.appendChild(notification);
        
        setTimeout(() => {
            notification.remove();
        }, 3000);
    }

    formatCategory(category) {
        const categories = {
            'analytics': 'Analytics',
            'reporting': 'Reporting',
            'integration': 'Integration',
            'automation': 'Automation',
            'monitoring': 'Monitoring',
            'security': 'Security',
            'mobile': 'Mobile'
        };
        return categories[category] || category;
    }

    formatPriority(priority) {
        const priorities = {
            1: 'High Priority',
            2: 'Medium Priority',
            3: 'Low Priority',
            4: 'Future',
            5: 'Backlog'
        };
        return priorities[priority] || `Priority ${priority}`;
    }

    formatETA(eta) {
        if (!eta) return 'TBD';
        
        const etaDate = new Date(eta);
        const now = new Date();
        const diffTime = etaDate.getTime() - now.getTime();
        const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
        
        if (diffDays < 0) {
            return 'Overdue';
        } else if (diffDays === 0) {
            return 'Today';
        } else if (diffDays === 1) {
            return 'Tomorrow';
        } else if (diffDays < 7) {
            return `${diffDays} days`;
        } else if (diffDays < 30) {
            const weeks = Math.ceil(diffDays / 7);
            return `${weeks} week${weeks > 1 ? 's' : ''}`;
        } else {
            const months = Math.ceil(diffDays / 30);
            return `${months} month${months > 1 ? 's' : ''}`;
        }
    }

    getStatusSubtitle(status) {
        const subtitles = {
            'coming_soon': 'Exciting new features are on the way!',
            'in_development': 'This feature is currently being developed',
            'available': 'This feature is now available!',
            'deprecated': 'This feature is no longer supported'
        };
        return subtitles[status] || 'Feature status update';
    }

    getAuthToken() {
        // Get auth token from localStorage or cookie
        return localStorage.getItem('auth_token') || 
               document.cookie.split('; ').find(row => row.startsWith('auth_token='))?.split('=')[1] || 
               '';
    }

    // Public methods for external control
    setFeature(featureId) {
        this.featureId = featureId;
        this.category = null;
        this.loadFeatureData();
    }

    setCategory(category) {
        this.category = category;
        this.featureId = null;
        this.loadFeatureData();
    }

    setStatus(status) {
        this.status = status;
        this.featureId = null;
        this.category = null;
        this.loadFeatureData();
    }

    refresh() {
        this.loadFeatureData();
    }

    toggle() {
        if (this.isVisible) {
            this.hide();
        } else {
            this.show();
        }
    }

    destroy() {
        this.stopAutoRefresh();
        if (this.container) {
            this.container.innerHTML = '';
        }
        this.isVisible = false;
    }
}

// Export for use in other scripts
if (typeof module !== 'undefined' && module.exports) {
    module.exports = ComingSoonBanner;
}

// Auto-initialize if container is found
document.addEventListener('DOMContentLoaded', () => {
    const bannerContainer = document.getElementById('comingSoonBannerContainer');
    if (bannerContainer && !window.comingSoonBanner) {
        window.comingSoonBanner = new ComingSoonBanner({
            container: bannerContainer
        });
    }
});
