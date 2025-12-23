#!/bin/bash

# 50-Sample E2E Test for Quick Metrics
# Tests against Railway production to measure improvements

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

RAILWAY_API_URL="${RAILWAY_API_URL:-https://classification-service-production.up.railway.app}"
NUM_SAMPLES="${NUM_SAMPLES:-50}"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}50-Sample E2E Metrics Test${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "API URL: ${YELLOW}$RAILWAY_API_URL${NC}"
echo -e "Samples: ${YELLOW}$NUM_SAMPLES${NC}"
echo ""

# Create results directory
mkdir -p test/results

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RESULTS_FILE="test/results/50_sample_e2e_metrics_${TIMESTAMP}.json"

echo -e "${GREEN}🚀 Running $NUM_SAMPLES sample E2E test...${NC}"
echo ""

# Run the comprehensive test but limit to 50 samples
# We'll use the Go test but modify it to use fewer samples
export RAILWAY_API_URL="$RAILWAY_API_URL"
export TEST_SAMPLE_LIMIT="$NUM_SAMPLES"

# Create a temporary test file that limits samples
cat > /tmp/test_50_samples.go << 'EOF'
//go:build e2e_railway
// +build e2e_railway

package integration

import (
	"os"
	"testing"
	"time"
)

func TestRailway50SampleE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping 50-sample Railway E2E test in short mode")
	}

	apiURL := os.Getenv("RAILWAY_API_URL")
	if apiURL == "" {
		apiURL = "https://classification-service-production.up.railway.app"
	}

	if !verifyServiceHealth(t, apiURL) {
		t.Fatalf("Service at %s is not accessible", apiURL)
	}

	allSamples := generateComprehensiveTestSamples()
	
	// Limit to 50 samples
	sampleLimit := 50
	if len(allSamples) > sampleLimit {
		allSamples = allSamples[:sampleLimit]
	}
	
	t.Logf("🚀 Starting Railway E2E test with %d samples", len(allSamples))

	runner := NewRailwayE2ETestRunner(t, apiURL)
	startTime := time.Now()
	_ = runner.RunComprehensiveTests(allSamples)
	totalDuration := time.Since(startTime)

	t.Logf("✅ Completed all tests in %v", totalDuration)
	runner.CalculateMetrics()
	report := runner.GenerateComprehensiveReport(totalDuration)
	validateE2EResults(t, report)

	timestamp := time.Now().Format("20060102_150405")
	reportPath := fmt.Sprintf("test/results/railway_e2e_classification_50_%s.json", timestamp)
	analysisPath := fmt.Sprintf("test/results/railway_e2e_analysis_50_%s.json", timestamp)
	
	saveReport(report, reportPath)
	saveAnalysis(analysis, analysisPath)
	
	t.Logf("📊 Reports saved:")
	t.Logf("  - %s", reportPath)
	t.Logf("  - %s", analysisPath)
}
EOF

# Run the test
echo "Running Go test with 50 samples..."
go test -v -timeout 30m -tags e2e_railway ./test/integration -run TestRailway50SampleE2E 2>&1 | tee "test/results/50_sample_e2e_output_${TIMESTAMP}.txt"

TEST_EXIT_CODE=$?

echo ""
if [ $TEST_EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✅ 50-sample test completed successfully!${NC}"
    echo ""
    echo -e "${BLUE}📊 Results saved to:${NC}"
    echo "  - test/results/railway_e2e_classification_50_*.json"
    echo "  - test/results/railway_e2e_analysis_50_*.json"
else
    echo -e "${RED}❌ Test failed with exit code: $TEST_EXIT_CODE${NC}"
fi

exit $TEST_EXIT_CODE

