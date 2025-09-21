package testing

import (
	"context"
	"testing"
	"time"
)

// TestBackupRecoveryIntegration runs the complete backup and recovery test suite
func TestBackupRecoveryIntegration(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Get configuration from environment variables
	config := &BackupTestConfig{
		SupabaseURL:       getEnvOrDefault("SUPABASE_URL", "postgresql://postgres:password@localhost:5432/kyb_platform"),
		TestDatabaseURL:   getEnvOrDefault("TEST_DATABASE_URL", "postgresql://postgres:password@localhost:5432/kyb_platform_test"),
		BackupDirectory:   getEnvOrDefault("BACKUP_DIRECTORY", "/tmp/backup_recovery_test"),
		TestDataSize:      1000,
		RecoveryTimeout:   10 * time.Minute,
		ValidationRetries: 3,
	}

	// Create test runner
	runner, err := NewBackupRecoveryTestRunner(config)
	if err != nil {
		t.Fatalf("Failed to create test runner: %v", err)
	}
	defer runner.Close()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Run all tests
	if err := runner.RunAllTests(ctx); err != nil {
		t.Errorf("Backup and recovery tests failed: %v", err)
	}

	// Validate results
	validateTestResults(t, runner.results)
}

// TestBackupProceduresOnly tests only the backup procedures
func TestBackupProceduresOnly(t *testing.T) {
	config := &BackupTestConfig{
		SupabaseURL:       getEnvOrDefault("SUPABASE_URL", "postgresql://postgres:password@localhost:5432/kyb_platform"),
		TestDatabaseURL:   getEnvOrDefault("TEST_DATABASE_URL", "postgresql://postgres:password@localhost:5432/kyb_platform_test"),
		BackupDirectory:   getEnvOrDefault("BACKUP_DIRECTORY", "/tmp/backup_recovery_test"),
		TestDataSize:      100,
		RecoveryTimeout:   5 * time.Minute,
		ValidationRetries: 2,
	}

	tester, err := NewBackupRecoveryTester(config)
	if err != nil {
		t.Fatalf("Failed to create backup recovery tester: %v", err)
	}
	defer tester.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	result, err := tester.TestBackupProcedures(ctx)
	if err != nil {
		t.Errorf("Backup procedures test failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Backup procedures test failed: %s", result.ErrorMessage)
	}

	t.Logf("Backup procedures test completed in %v", result.Duration)
}

// TestRecoveryScenariosOnly tests only the recovery scenarios
func TestRecoveryScenariosOnly(t *testing.T) {
	config := &BackupTestConfig{
		SupabaseURL:       getEnvOrDefault("SUPABASE_URL", "postgresql://postgres:password@localhost:5432/kyb_platform"),
		TestDatabaseURL:   getEnvOrDefault("TEST_DATABASE_URL", "postgresql://postgres:password@localhost:5432/kyb_platform_test"),
		BackupDirectory:   getEnvOrDefault("BACKUP_DIRECTORY", "/tmp/backup_recovery_test"),
		TestDataSize:      100,
		RecoveryTimeout:   5 * time.Minute,
		ValidationRetries: 2,
	}

	tester, err := NewBackupRecoveryTester(config)
	if err != nil {
		t.Fatalf("Failed to create backup recovery tester: %v", err)
	}
	defer tester.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	result, err := tester.TestRecoveryScenarios(ctx)
	if err != nil {
		t.Errorf("Recovery scenarios test failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Recovery scenarios test failed: %s", result.ErrorMessage)
	}

	t.Logf("Recovery scenarios test completed in %v", result.Duration)
}

// TestDataRestorationOnly tests only the data restoration validation
func TestDataRestorationOnly(t *testing.T) {
	config := &BackupTestConfig{
		SupabaseURL:       getEnvOrDefault("SUPABASE_URL", "postgresql://postgres:password@localhost:5432/kyb_platform"),
		TestDatabaseURL:   getEnvOrDefault("TEST_DATABASE_URL", "postgresql://postgres:password@localhost:5432/kyb_platform_test"),
		BackupDirectory:   getEnvOrDefault("BACKUP_DIRECTORY", "/tmp/backup_recovery_test"),
		TestDataSize:      100,
		RecoveryTimeout:   5 * time.Minute,
		ValidationRetries: 2,
	}

	tester, err := NewBackupRecoveryTester(config)
	if err != nil {
		t.Fatalf("Failed to create backup recovery tester: %v", err)
	}
	defer tester.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	result, err := tester.TestDataRestoration(ctx)
	if err != nil {
		t.Errorf("Data restoration test failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Data restoration test failed: %s", result.ErrorMessage)
	}

	if !result.DataIntegrity {
		t.Errorf("Data integrity validation failed")
	}

	t.Logf("Data restoration test completed in %v with validation score: %.2f%%",
		result.Duration, result.ValidationScore*100)
}

// TestPointInTimeRecoveryOnly tests only the point-in-time recovery
func TestPointInTimeRecoveryOnly(t *testing.T) {
	config := &BackupTestConfig{
		SupabaseURL:       getEnvOrDefault("SUPABASE_URL", "postgresql://postgres:password@localhost:5432/kyb_platform"),
		TestDatabaseURL:   getEnvOrDefault("TEST_DATABASE_URL", "postgresql://postgres:password@localhost:5432/kyb_platform_test"),
		BackupDirectory:   getEnvOrDefault("BACKUP_DIRECTORY", "/tmp/backup_recovery_test"),
		TestDataSize:      100,
		RecoveryTimeout:   5 * time.Minute,
		ValidationRetries: 2,
	}

	tester, err := NewBackupRecoveryTester(config)
	if err != nil {
		t.Fatalf("Failed to create backup recovery tester: %v", err)
	}
	defer tester.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	result, err := tester.TestPointInTimeRecovery(ctx)
	if err != nil {
		t.Errorf("Point-in-time recovery test failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Point-in-time recovery test failed: %s", result.ErrorMessage)
	}

	t.Logf("Point-in-time recovery test completed in %v with recovery time: %v",
		result.Duration, result.RecoveryTime)
}

// validateTestResults validates the overall test results
func validateTestResults(t *testing.T, results []*BackupTestResult) {
	if len(results) == 0 {
		t.Fatal("No test results to validate")
	}

	var totalTests, passedTests int
	var totalDuration time.Duration
	var totalScore float64

	for _, result := range results {
		totalTests++
		totalDuration += result.Duration
		totalScore += result.ValidationScore

		if result.Success {
			passedTests++
		} else {
			t.Errorf("Test %s failed: %s", result.TestName, result.ErrorMessage)
		}

		// Validate individual test metrics
		if result.ValidationScore < 0.95 {
			t.Errorf("Test %s has low validation score: %.2f%% (target: 95%%)",
				result.TestName, result.ValidationScore*100)
		}

		if result.RecoveryTime > 5*time.Minute && result.RecoveryTime > 0 {
			t.Errorf("Test %s has slow recovery time: %v (target: <5 minutes)",
				result.TestName, result.RecoveryTime)
		}
	}

	// Validate overall metrics
	passRate := float64(passedTests) / float64(totalTests)
	if passRate < 1.0 {
		t.Errorf("Overall pass rate is %.2f%% (target: 100%%)", passRate*100)
	}

	avgScore := totalScore / float64(totalTests)
	if avgScore < 0.95 {
		t.Errorf("Average validation score is %.2f%% (target: 95%%)", avgScore*100)
	}

	t.Logf("Test Summary: %d/%d tests passed (%.2f%%), Average score: %.2f%%, Total duration: %v",
		passedTests, totalTests, passRate*100, avgScore*100, totalDuration)
}

// BenchmarkBackupProcedures benchmarks backup procedures
func BenchmarkBackupProcedures(b *testing.B) {
	config := &BackupTestConfig{
		SupabaseURL:       getEnvOrDefault("SUPABASE_URL", "postgresql://postgres:password@localhost:5432/kyb_platform"),
		TestDatabaseURL:   getEnvOrDefault("TEST_DATABASE_URL", "postgresql://postgres:password@localhost:5432/kyb_platform_test"),
		BackupDirectory:   getEnvOrDefault("BACKUP_DIRECTORY", "/tmp/backup_recovery_benchmark"),
		TestDataSize:      1000,
		RecoveryTimeout:   5 * time.Minute,
		ValidationRetries: 2,
	}

	tester, err := NewBackupRecoveryTester(config)
	if err != nil {
		b.Fatalf("Failed to create backup recovery tester: %v", err)
	}
	defer tester.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := tester.TestBackupProcedures(ctx)
		if err != nil {
			b.Errorf("Backup procedures test failed: %v", err)
		}
		if !result.Success {
			b.Errorf("Backup procedures test failed: %s", result.ErrorMessage)
		}
	}
}

// BenchmarkRecoveryProcedures benchmarks recovery procedures
func BenchmarkRecoveryProcedures(b *testing.B) {
	config := &BackupTestConfig{
		SupabaseURL:       getEnvOrDefault("SUPABASE_URL", "postgresql://postgres:password@localhost:5432/kyb_platform"),
		TestDatabaseURL:   getEnvOrDefault("TEST_DATABASE_URL", "postgresql://postgres:password@localhost:5432/kyb_platform_test"),
		BackupDirectory:   getEnvOrDefault("BACKUP_DIRECTORY", "/tmp/backup_recovery_benchmark"),
		TestDataSize:      1000,
		RecoveryTimeout:   5 * time.Minute,
		ValidationRetries: 2,
	}

	tester, err := NewBackupRecoveryTester(config)
	if err != nil {
		b.Fatalf("Failed to create backup recovery tester: %v", err)
	}
	defer tester.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := tester.TestRecoveryScenarios(ctx)
		if err != nil {
			b.Errorf("Recovery scenarios test failed: %v", err)
		}
		if !result.Success {
			b.Errorf("Recovery scenarios test failed: %s", result.ErrorMessage)
		}
	}
}
