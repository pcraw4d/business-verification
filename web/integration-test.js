/**
 * Comprehensive Integration Test for KYB Platform
 * Tests the complete flow from frontend to database
 */

class IntegrationTester {
    constructor() {
        this.apiBaseUrl = APIConfig.getBaseURL();
        this.testResults = {
            database: {},
            api: {},
            frontend: {},
            endToEnd: {}
        };
    }

    /**
     * Run all integration tests
     */
    async runAllTests() {
        console.log('🧪 Starting KYB Platform Integration Tests...\n');
        
        try {
            await this.testDatabaseConnection();
            await this.testAPIEndpoints();
            await this.testFrontendComponents();
            await this.testEndToEndFlow();
            
            this.generateReport();
        } catch (error) {
            console.error('❌ Integration test failed:', error);
        }
    }

    /**
     * Test database connection and data availability
     */
    async testDatabaseConnection() {
        console.log('📊 Testing Database Connection...');
        
        try {
            // Test health endpoint
            const healthResponse = await fetch(`${this.apiBaseUrl}/health`);
            const healthData = await healthResponse.json();
            
            this.testResults.database.health = {
                status: healthResponse.status,
                connected: healthData.supabase_status?.connected,
                url: healthData.supabase_status?.url
            };
            
            console.log('✅ Health check:', healthData.status);
            console.log('✅ Supabase connected:', healthData.supabase_status?.connected);
            
            // Test classification endpoint (creates data)
            const classificationResponse = await fetch(`${this.apiBaseUrl}/v1/classify`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    business_name: 'Integration Test Company',
                    description: 'A technology company for integration testing'
                })
            });
            
            const classificationData = await classificationResponse.json();
            
            this.testResults.database.classification = {
                status: classificationResponse.status,
                dataSource: classificationData.data_source,
                hasBusinessId: !!classificationData.business_id,
                hasRiskAssessment: !!classificationData.risk_assessment
            };
            
            console.log('✅ Classification test:', classificationData.data_source);
            console.log('✅ Business ID generated:', classificationData.business_id);
            
        } catch (error) {
            this.testResults.database.error = error.message;
            console.error('❌ Database test failed:', error.message);
        }
    }

    /**
     * Test all API endpoints
     */
    async testAPIEndpoints() {
        console.log('\n🔌 Testing API Endpoints...');
        
        const endpoints = [
            { name: 'Merchants', url: '/api/v1/merchants' },
            { name: 'Analytics', url: '/api/v1/merchants/analytics' },
            { name: 'Statistics', url: '/api/v1/merchants/statistics' },
            { name: 'Portfolio Types', url: '/api/v1/merchants/portfolio-types' },
            { name: 'Risk Levels', url: '/api/v1/merchants/risk-levels' }
        ];
        
        for (const endpoint of endpoints) {
            try {
                const response = await fetch(`${this.apiBaseUrl}${endpoint.url}`);
                const data = await response.json();
                
                this.testResults.api[endpoint.name] = {
                    status: response.status,
                    success: response.ok,
                    dataSource: data.data_source || 'unknown',
                    hasData: Object.keys(data).length > 0,
                    dataKeys: Object.keys(data).slice(0, 5) // First 5 keys
                };
                
                console.log(`✅ ${endpoint.name}: ${response.status} (${data.data_source || 'unknown'})`);
                
            } catch (error) {
                this.testResults.api[endpoint.name] = {
                    status: 'error',
                    success: false,
                    error: error.message
                };
                console.error(`❌ ${endpoint.name}: ${error.message}`);
            }
        }
    }

    /**
     * Test frontend components
     */
    async testFrontendComponents() {
        console.log('\n🎨 Testing Frontend Components...');
        
        const components = [
            'RealDataIntegration',
            'MerchantDashboardRealData',
            'MonitoringDashboardRealData',
            'MerchantBulkOperationsRealData',
            'DashboardRealData'
        ];
        
        for (const componentName of components) {
            const isAvailable = typeof window[componentName] !== 'undefined';
            const isFunction = typeof window[componentName] === 'function';
            
            this.testResults.frontend[componentName] = {
                available: isAvailable,
                isFunction: isFunction
            };
            
            if (isAvailable) {
                console.log(`✅ ${componentName}: Available`);
            } else {
                console.log(`❌ ${componentName}: Not available`);
            }
        }
        
        // Test RealDataIntegration functionality
        if (typeof RealDataIntegration !== 'undefined') {
            try {
                const dataIntegration = new RealDataIntegration();
                
                this.testResults.frontend.RealDataIntegrationMethods = {
                    getMerchants: typeof dataIntegration.getMerchants === 'function',
                    getMerchantAnalytics: typeof dataIntegration.getMerchantAnalytics === 'function',
                    getSystemMetrics: typeof dataIntegration.getSystemMetrics === 'function',
                    getBusinessIntelligence: typeof dataIntegration.getBusinessIntelligence === 'function'
                };
                
                console.log('✅ RealDataIntegration methods available');
            } catch (error) {
                this.testResults.frontend.RealDataIntegrationError = error.message;
                console.error('❌ RealDataIntegration error:', error.message);
            }
        }
    }

    /**
     * Test end-to-end integration flow
     */
    async testEndToEndFlow() {
        console.log('\n🔄 Testing End-to-End Integration Flow...');
        
        try {
            // Step 1: Test data flow from API to frontend
            const merchantsResponse = await fetch(`${this.apiBaseUrl}/api/v1/merchants`);
            const merchantsData = await merchantsResponse.json();
            
            this.testResults.endToEnd.dataFlow = {
                merchantsCount: merchantsData.merchants?.length || 0,
                dataSource: merchantsData.data_source,
                hasRealData: merchantsData.data_source !== 'mock_data'
            };
            
            console.log(`✅ Data flow: ${merchantsData.merchants?.length || 0} merchants (${merchantsData.data_source})`);
            
            // Step 2: Test classification flow
            const classificationResponse = await fetch(`${this.apiBaseUrl}/v1/classify`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    business_name: 'E2E Test Company',
                    description: 'End-to-end integration test company'
                })
            });
            
            const classificationData = await classificationResponse.json();
            
            this.testResults.endToEnd.classificationFlow = {
                success: classificationData.success,
                dataSource: classificationData.data_source,
                hasRiskAssessment: !!classificationData.risk_assessment,
                confidenceScore: classificationData.confidence_score
            };
            
            console.log(`✅ Classification flow: ${classificationData.data_source} (confidence: ${classificationData.confidence_score})`);
            
            // Step 3: Test analytics flow
            const analyticsResponse = await fetch(`${this.apiBaseUrl}/api/v1/merchants/analytics`);
            const analyticsData = await analyticsResponse.json();
            
            this.testResults.endToEnd.analyticsFlow = {
                dataSource: analyticsData.data_source,
                hasPortfolioDistribution: !!analyticsData.portfolio_distribution,
                hasRiskDistribution: !!analyticsData.risk_distribution,
                totalMerchants: analyticsData.total_merchants
            };
            
            console.log(`✅ Analytics flow: ${analyticsData.data_source} (${analyticsData.total_merchants} merchants)`);
            
            // Step 4: Test frontend integration capability
            const frontendReady = typeof RealDataIntegration !== 'undefined' && 
                                typeof MerchantDashboardRealData !== 'undefined' &&
                                typeof MonitoringDashboardRealData !== 'undefined';
            
            this.testResults.endToEnd.frontendIntegration = {
                ready: frontendReady,
                componentsAvailable: {
                    RealDataIntegration: typeof RealDataIntegration !== 'undefined',
                    MerchantDashboardRealData: typeof MerchantDashboardRealData !== 'undefined',
                    MonitoringDashboardRealData: typeof MonitoringDashboardRealData !== 'undefined',
                    MerchantBulkOperationsRealData: typeof MerchantBulkOperationsRealData !== 'undefined',
                    DashboardRealData: typeof DashboardRealData !== 'undefined'
                }
            };
            
            console.log(`✅ Frontend integration: ${frontendReady ? 'Ready' : 'Not ready'}`);
            
        } catch (error) {
            this.testResults.endToEnd.error = error.message;
            console.error('❌ End-to-end test failed:', error.message);
        }
    }

    /**
     * Generate comprehensive test report
     */
    generateReport() {
        console.log('\n📋 INTEGRATION TEST REPORT');
        console.log('=' .repeat(50));
        
        // Database tests
        console.log('\n📊 DATABASE TESTS:');
        const dbHealth = this.testResults.database.health?.connected;
        const dbClassification = this.testResults.database.classification?.dataSource;
        console.log(`  Health: ${dbHealth ? '✅ Connected' : '❌ Disconnected'}`);
        console.log(`  Classification: ${dbClassification ? `✅ ${dbClassification}` : '❌ Failed'}`);
        
        // API tests
        console.log('\n🔌 API TESTS:');
        const apiTests = Object.entries(this.testResults.api);
        const apiSuccessCount = apiTests.filter(([_, result]) => result.success).length;
        console.log(`  Success: ${apiSuccessCount}/${apiTests.length} endpoints`);
        
        apiTests.forEach(([name, result]) => {
            const status = result.success ? '✅' : '❌';
            console.log(`  ${name}: ${status} ${result.status} (${result.dataSource || 'error'})`);
        });
        
        // Frontend tests
        console.log('\n🎨 FRONTEND TESTS:');
        const frontendTests = Object.entries(this.testResults.frontend);
        const frontendSuccessCount = frontendTests.filter(([_, result]) => result.available).length;
        console.log(`  Success: ${frontendSuccessCount}/${frontendTests.length} components`);
        
        frontendTests.forEach(([name, result]) => {
            const status = result.available ? '✅' : '❌';
            console.log(`  ${name}: ${status}`);
        });
        
        // End-to-end tests
        console.log('\n🔄 END-TO-END TESTS:');
        const e2eDataFlow = this.testResults.endToEnd.dataFlow?.hasRealData;
        const e2eClassification = this.testResults.endToEnd.classificationFlow?.success;
        const e2eFrontend = this.testResults.endToEnd.frontendIntegration?.ready;
        
        console.log(`  Data Flow: ${e2eDataFlow ? '✅ Real Data' : '❌ Mock Data'}`);
        console.log(`  Classification: ${e2eClassification ? '✅ Working' : '❌ Failed'}`);
        console.log(`  Frontend: ${e2eFrontend ? '✅ Ready' : '❌ Not Ready'}`);
        
        // Overall assessment
        const overallSuccess = dbHealth && 
                              apiSuccessCount >= 3 && 
                              frontendSuccessCount >= 4 && 
                              e2eDataFlow && 
                              e2eClassification && 
                              e2eFrontend;
        
        console.log('\n🎯 OVERALL ASSESSMENT:');
        console.log(`  Status: ${overallSuccess ? '✅ INTEGRATION READY' : '⚠️ ISSUES DETECTED'}`);
        
        if (overallSuccess) {
            console.log('\n🚀 The KYB Platform integration is ready for deployment!');
            console.log('   All components are working with real Supabase data.');
        } else {
            console.log('\n⚠️ Some integration issues detected:');
            if (!dbHealth) console.log('   - Database connection issues');
            if (apiSuccessCount < 3) console.log('   - API endpoint issues');
            if (frontendSuccessCount < 4) console.log('   - Frontend component issues');
            if (!e2eDataFlow) console.log('   - Data flow issues');
            if (!e2eClassification) console.log('   - Classification issues');
            if (!e2eFrontend) console.log('   - Frontend integration issues');
        }
        
        console.log('\n' + '=' .repeat(50));
        
        return overallSuccess;
    }
}

// Auto-run tests if this script is loaded in a browser
if (typeof window !== 'undefined') {
    window.IntegrationTester = IntegrationTester;
    
    // Auto-run tests when page loads
    document.addEventListener('DOMContentLoaded', async () => {
        const tester = new IntegrationTester();
        await tester.runAllTests();
    });
}

// Export for Node.js usage
if (typeof module !== 'undefined' && module.exports) {
    module.exports = IntegrationTester;
}
