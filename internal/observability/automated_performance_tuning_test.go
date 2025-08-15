package observability

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewAutomatedPerformanceTuningSystem(t *testing.T) {
	logger := zap.NewNop()
	config := TuningConfig{
		TuningInterval:       5 * time.Minute,
		MaxConcurrentTunings: 3,
		TuningTimeout:        30 * time.Minute,
		DefaultPolicy:        "balanced",
		ConservativePolicy:   "conservative",
		AggressivePolicy:     "aggressive",
		MaxTuningAttempts:    5,
		RollbackThreshold:    -5.0,
		SafetyMargin:         10.0,
		MinImprovement:       2.0,
		MaxDegradation:       3.0,
		StabilizationPeriod:  2 * time.Minute,
		EnableTuningAlerts:   true,
		AlertThresholds:      make(map[string]float64),
	}

	performanceMonitor := &PerformanceMonitor{}
	optimizationSystem := &PerformanceOptimizationSystem{}
	predictiveAnalytics := &PredictiveAnalytics{}
	regressionDetection := &RegressionDetectionSystem{}

	apts := NewAutomatedPerformanceTuningSystem(performanceMonitor, optimizationSystem, predictiveAnalytics, regressionDetection, config, logger)

	assert.NotNil(t, apts)
	assert.Equal(t, performanceMonitor, apts.performanceMonitor)
	assert.Equal(t, optimizationSystem, apts.optimizationSystem)
	assert.Equal(t, predictiveAnalytics, apts.predictiveAnalytics)
	assert.Equal(t, regressionDetection, apts.regressionDetection)
	assert.Equal(t, config, apts.config)
	assert.NotNil(t, apts.tuningEngine)
	assert.NotNil(t, apts.tuningHistory)
	assert.NotNil(t, apts.activeTunings)
	assert.NotNil(t, apts.tuningPolicies)
}

func TestAutomatedPerformanceTuningSystem_StartStop(t *testing.T) {
	logger := zap.NewNop()
	config := TuningConfig{
		TuningInterval: 100 * time.Millisecond, // Short interval for testing
	}

	performanceMonitor := &PerformanceMonitor{}
	optimizationSystem := &PerformanceOptimizationSystem{}
	predictiveAnalytics := &PredictiveAnalytics{}
	regressionDetection := &RegressionDetectionSystem{}

	apts := NewAutomatedPerformanceTuningSystem(performanceMonitor, optimizationSystem, predictiveAnalytics, regressionDetection, config, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the system
	err := apts.Start(ctx)
	assert.NoError(t, err)

	// Wait a bit for goroutines to start
	time.Sleep(50 * time.Millisecond)

	// Stop the system
	err = apts.Stop()
	assert.NoError(t, err)
}

func TestAutomatedPerformanceTuningSystem_ShouldTune(t *testing.T) {
	logger := zap.NewNop()
	config := TuningConfig{
		SafetyMargin: 10.0,
	}

	performanceMonitor := &PerformanceMonitor{}
	optimizationSystem := &PerformanceOptimizationSystem{}
	predictiveAnalytics := &PredictiveAnalytics{}
	regressionDetection := &RegressionDetectionSystem{}

	apts := NewAutomatedPerformanceTuningSystem(performanceMonitor, optimizationSystem, predictiveAnalytics, regressionDetection, config, logger)

	// Test metrics that should trigger tuning
	metrics := &PerformanceMetrics{
		ResponseTime: struct {
			Current     time.Duration `json:"current"`
			Expected    time.Duration `json:"expected"`
			Improvement float64       `json:"improvement"`
		}{
			Current:  600 * time.Millisecond, // 20% above expected (should trigger)
			Expected: 500 * time.Millisecond,
		},
		Throughput: struct {
			Current     float64 `json:"current"`
			Expected    float64 `json:"expected"`
			Improvement float64 `json:"improvement"`
		}{
			Current:  800.0, // 20% below expected (should trigger)
			Expected: 1000.0,
		},
		SuccessRate: struct {
			Current     float64 `json:"current"`
			Expected    float64 `json:"expected"`
			Improvement float64 `json:"improvement"`
		}{
			Current:  0.85, // 10% below expected (should trigger)
			Expected: 0.95,
		},
		ResourceUsage: struct {
			CPU struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			} `json:"cpu"`
			Memory struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			} `json:"memory"`
			Disk struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			} `json:"disk"`
		}{
			CPU: struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			}{
				Current: 85.0, // Above safety margin (should trigger)
			},
			Memory: struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			}{
				Current: 90.0, // Above safety margin (should trigger)
			},
			Disk: struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			}{
				Current: 70.0, // Below safety margin (should not trigger)
			},
		},
	}

	// Should trigger tuning
	shouldTune := apts.shouldTune(metrics)
	assert.True(t, shouldTune, "Metrics should trigger tuning")

	// Test metrics that should not trigger tuning
	goodMetrics := &PerformanceMetrics{
		ResponseTime: struct {
			Current     time.Duration `json:"current"`
			Expected    time.Duration `json:"expected"`
			Improvement float64       `json:"improvement"`
		}{
			Current:  450 * time.Millisecond, // 10% above expected (within safety margin)
			Expected: 500 * time.Millisecond,
		},
		Throughput: struct {
			Current     float64 `json:"current"`
			Expected    float64 `json:"expected"`
			Improvement float64 `json:"improvement"`
		}{
			Current:  920.0, // 8% below expected (within safety margin)
			Expected: 1000.0,
		},
		SuccessRate: struct {
			Current     float64 `json:"current"`
			Expected    float64 `json:"expected"`
			Improvement float64 `json:"improvement"`
		}{
			Current:  0.87, // 8% below expected (within safety margin)
			Expected: 0.95,
		},
		ResourceUsage: struct {
			CPU struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			} `json:"cpu"`
			Memory struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			} `json:"memory"`
			Disk struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			} `json:"disk"`
		}{
			CPU: struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			}{
				Current: 75.0, // Below safety margin
			},
			Memory: struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			}{
				Current: 80.0, // Below safety margin
			},
			Disk: struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			}{
				Current: 70.0, // Below safety margin
			},
		},
	}

	// Should not trigger tuning
	shouldTune = apts.shouldTune(goodMetrics)
	assert.False(t, shouldTune, "Metrics should not trigger tuning")
}

func TestAutomatedPerformanceTuningSystem_SelectTuningPolicy(t *testing.T) {
	logger := zap.NewNop()
	config := TuningConfig{}

	performanceMonitor := &PerformanceMonitor{}
	optimizationSystem := &PerformanceOptimizationSystem{}
	predictiveAnalytics := &PredictiveAnalytics{}
	regressionDetection := &RegressionDetectionSystem{}

	apts := NewAutomatedPerformanceTuningSystem(performanceMonitor, optimizationSystem, predictiveAnalytics, regressionDetection, config, logger)

	// Test critical performance issue (should select aggressive policy)
	criticalMetrics := &PerformanceMetrics{
		ResponseTime: struct {
			Current     time.Duration `json:"current"`
			Expected    time.Duration `json:"expected"`
			Improvement float64       `json:"improvement"`
		}{
			Current:  1200 * time.Millisecond, // 2x expected
			Expected: 500 * time.Millisecond,
		},
	}

	policy := apts.selectTuningPolicy(criticalMetrics)
	assert.NotNil(t, policy)
	assert.Equal(t, "aggressive", policy.Type)

	// Test moderate performance issue (should select balanced policy)
	moderateMetrics := &PerformanceMetrics{
		ResponseTime: struct {
			Current     time.Duration `json:"current"`
			Expected    time.Duration `json:"expected"`
			Improvement float64       `json:"improvement"`
		}{
			Current:  800 * time.Millisecond, // 1.6x expected
			Expected: 500 * time.Millisecond,
		},
	}

	policy = apts.selectTuningPolicy(moderateMetrics)
	assert.NotNil(t, policy)
	assert.Equal(t, "balanced", policy.Type)

	// Test minor performance issue (should select conservative policy)
	minorMetrics := &PerformanceMetrics{
		ResponseTime: struct {
			Current     time.Duration `json:"current"`
			Expected    time.Duration `json:"expected"`
			Improvement float64       `json:"improvement"`
		}{
			Current:  600 * time.Millisecond, // 1.2x expected
			Expected: 500 * time.Millisecond,
		},
	}

	policy = apts.selectTuningPolicy(minorMetrics)
	assert.NotNil(t, policy)
	assert.Equal(t, "conservative", policy.Type)
}

func TestAutomatedPerformanceTuningSystem_CreateTuningSession(t *testing.T) {
	logger := zap.NewNop()
	config := TuningConfig{}

	performanceMonitor := &PerformanceMonitor{}
	optimizationSystem := &PerformanceOptimizationSystem{}
	predictiveAnalytics := &PredictiveAnalytics{}
	regressionDetection := &RegressionDetectionSystem{}

	apts := NewAutomatedPerformanceTuningSystem(performanceMonitor, optimizationSystem, predictiveAnalytics, regressionDetection, config, logger)

	// Get a policy
	policies := apts.GetTuningPolicies()
	policy := policies["balanced"]
	assert.NotNil(t, policy)

	// Create metrics
	metrics := &PerformanceMetrics{
		ResponseTime: struct {
			Current     time.Duration `json:"current"`
			Expected    time.Duration `json:"expected"`
			Improvement float64       `json:"improvement"`
		}{
			Current:  600 * time.Millisecond,
			Expected: 500 * time.Millisecond,
		},
		Throughput: struct {
			Current     float64 `json:"current"`
			Expected    float64 `json:"expected"`
			Improvement float64 `json:"improvement"`
		}{
			Current:  800.0,
			Expected: 1000.0,
		},
		SuccessRate: struct {
			Current     float64 `json:"current"`
			Expected    float64 `json:"expected"`
			Improvement float64 `json:"improvement"`
		}{
			Current:  0.90,
			Expected: 0.95,
		},
		ResourceUsage: struct {
			CPU struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			} `json:"cpu"`
			Memory struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			} `json:"memory"`
			Disk struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			} `json:"disk"`
		}{
			CPU: struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			}{
				Current: 85.0,
			},
			Memory: struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			}{
				Current: 90.0,
			},
			Disk: struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			}{
				Current: 70.0,
			},
		},
	}

	// Create tuning session
	session := apts.createTuningSession(policy, metrics)

	assert.NotNil(t, session)
	assert.Equal(t, policy.ID, session.PolicyID)
	assert.Equal(t, "active", session.Status)
	assert.Equal(t, "system_performance", session.Target)
	assert.Equal(t, "optimize_performance", session.Objective)
	assert.Equal(t, policy.Priority, session.Priority)
	assert.NotNil(t, session.InitialMetrics)
	assert.NotEmpty(t, session.Actions)
	assert.Equal(t, len(session.Actions), session.TotalActions)
	assert.NotNil(t, session.NextAction)

	// Check that session was stored
	sessions := apts.GetTuningSessions()
	assert.Contains(t, sessions, session.ID)
}

func TestAutomatedPerformanceTuningSystem_CalculateImprovement(t *testing.T) {
	logger := zap.NewNop()
	config := TuningConfig{}

	performanceMonitor := &PerformanceMonitor{}
	optimizationSystem := &PerformanceOptimizationSystem{}
	predictiveAnalytics := &PredictiveAnalytics{}
	regressionDetection := &RegressionDetectionSystem{}

	apts := NewAutomatedPerformanceTuningSystem(performanceMonitor, optimizationSystem, predictiveAnalytics, regressionDetection, config, logger)

	// Test improvement calculation
	before := &PerformanceMetrics{
		ResponseTime: struct {
			Current     time.Duration `json:"current"`
			Expected    time.Duration `json:"expected"`
			Improvement float64       `json:"improvement"`
		}{
			Current:  600 * time.Millisecond,
			Expected: 500 * time.Millisecond,
		},
		Throughput: struct {
			Current     float64 `json:"current"`
			Expected    float64 `json:"expected"`
			Improvement float64 `json:"improvement"`
		}{
			Current:  800.0,
			Expected: 1000.0,
		},
		SuccessRate: struct {
			Current     float64 `json:"current"`
			Expected    float64 `json:"expected"`
			Improvement float64 `json:"improvement"`
		}{
			Current:  0.90,
			Expected: 0.95,
		},
		ResourceUsage: struct {
			CPU struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			} `json:"cpu"`
			Memory struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			} `json:"memory"`
			Disk struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			} `json:"disk"`
		}{
			CPU: struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			}{
				Current: 85.0,
			},
			Memory: struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			}{
				Current: 90.0,
			},
			Disk: struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			}{
				Current: 70.0,
			},
		},
	}

	after := &PerformanceMetrics{
		ResponseTime: struct {
			Current     time.Duration `json:"current"`
			Expected    time.Duration `json:"expected"`
			Improvement float64       `json:"improvement"`
		}{
			Current:  500 * time.Millisecond, // 16.67% improvement
			Expected: 500 * time.Millisecond,
		},
		Throughput: struct {
			Current     float64 `json:"current"`
			Expected    float64 `json:"expected"`
			Improvement float64 `json:"improvement"`
		}{
			Current:  900.0, // 12.5% improvement
			Expected: 1000.0,
		},
		SuccessRate: struct {
			Current     float64 `json:"current"`
			Expected    float64 `json:"expected"`
			Improvement float64 `json:"improvement"`
		}{
			Current:  0.93, // 3.33% improvement
			Expected: 0.95,
		},
		ResourceUsage: struct {
			CPU struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			} `json:"cpu"`
			Memory struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			} `json:"memory"`
			Disk struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			} `json:"disk"`
		}{
			CPU: struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			}{
				Current: 75.0,
			},
			Memory: struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			}{
				Current: 80.0,
			},
			Disk: struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			}{
				Current: 70.0,
			},
		},
	}

	improvement := apts.calculateImprovement(before, after)
	// Expected: (16.67 + 12.5 + 3.33) / 3 = 10.83%
	assert.Greater(t, improvement, 10.0)
	assert.Less(t, improvement, 11.0)
}

func TestAutomatedPerformanceTuningSystem_GetTuningSessions(t *testing.T) {
	logger := zap.NewNop()
	config := TuningConfig{}

	performanceMonitor := &PerformanceMonitor{}
	optimizationSystem := &PerformanceOptimizationSystem{}
	predictiveAnalytics := &PredictiveAnalytics{}
	regressionDetection := &RegressionDetectionSystem{}

	apts := NewAutomatedPerformanceTuningSystem(performanceMonitor, optimizationSystem, predictiveAnalytics, regressionDetection, config, logger)

	// Initially no active sessions
	sessions := apts.GetTuningSessions()
	assert.Empty(t, sessions)

	// Add a test session
	testSession := &TuningSession{
		ID:        "test_session_1",
		PolicyID:  "balanced",
		StartedAt: time.Now().UTC(),
		Status:    "active",
		Target:    "test_target",
		Objective: "test_objective",
		Priority:  "medium",
		Actions:   make([]*TuningAction, 0),
		Tags:      make(map[string]string),
	}

	apts.mu.Lock()
	apts.activeTunings[testSession.ID] = testSession
	apts.mu.Unlock()

	// Check that session is returned
	sessions = apts.GetTuningSessions()
	assert.Len(t, sessions, 1)
	assert.Contains(t, sessions, testSession.ID)
}

func TestAutomatedPerformanceTuningSystem_GetTuningHistory(t *testing.T) {
	logger := zap.NewNop()
	config := TuningConfig{}

	performanceMonitor := &PerformanceMonitor{}
	optimizationSystem := &PerformanceOptimizationSystem{}
	predictiveAnalytics := &PredictiveAnalytics{}
	regressionDetection := &RegressionDetectionSystem{}

	apts := NewAutomatedPerformanceTuningSystem(performanceMonitor, optimizationSystem, predictiveAnalytics, regressionDetection, config, logger)

	// Initially no history
	history := apts.GetTuningHistory()
	assert.Empty(t, history)

	// Add a test action
	testAction := &TuningAction{
		ID:          "test_action_1",
		SessionID:   "test_session_1",
		Type:        "parameter",
		Category:    "response_time",
		Action:      "optimize_response_time",
		ExecutedAt:  time.Now().UTC(),
		Status:      "completed",
		Description: "Test tuning action",
		Tags:        make(map[string]string),
	}

	apts.mu.Lock()
	apts.tuningHistory = append(apts.tuningHistory, testAction)
	apts.mu.Unlock()

	// Check that action is returned
	history = apts.GetTuningHistory()
	assert.Len(t, history, 1)
	assert.Equal(t, testAction.ID, history[0].ID)
}

func TestAutomatedPerformanceTuningSystem_GetTuningPolicies(t *testing.T) {
	logger := zap.NewNop()
	config := TuningConfig{}

	performanceMonitor := &PerformanceMonitor{}
	optimizationSystem := &PerformanceOptimizationSystem{}
	predictiveAnalytics := &PredictiveAnalytics{}
	regressionDetection := &RegressionDetectionSystem{}

	apts := NewAutomatedPerformanceTuningSystem(performanceMonitor, optimizationSystem, predictiveAnalytics, regressionDetection, config, logger)

	// Get tuning policies
	policies := apts.GetTuningPolicies()

	// Should have default policies
	assert.Contains(t, policies, "conservative")
	assert.Contains(t, policies, "balanced")
	assert.Contains(t, policies, "aggressive")

	// Check policy details
	conservative := policies["conservative"]
	assert.Equal(t, "Conservative Tuning Policy", conservative.Name)
	assert.Equal(t, "conservative", conservative.Type)
	assert.True(t, conservative.IsActive)

	balanced := policies["balanced"]
	assert.Equal(t, "Balanced Tuning Policy", balanced.Name)
	assert.Equal(t, "balanced", balanced.Type)
	assert.True(t, balanced.IsActive)

	aggressive := policies["aggressive"]
	assert.Equal(t, "Aggressive Tuning Policy", aggressive.Name)
	assert.Equal(t, "aggressive", aggressive.Type)
	assert.True(t, aggressive.IsActive)
}

func TestAutomatedPerformanceTuningSystem_CancelTuningSession(t *testing.T) {
	logger := zap.NewNop()
	config := TuningConfig{}

	performanceMonitor := &PerformanceMonitor{}
	optimizationSystem := &PerformanceOptimizationSystem{}
	predictiveAnalytics := &PredictiveAnalytics{}
	regressionDetection := &RegressionDetectionSystem{}

	apts := NewAutomatedPerformanceTuningSystem(performanceMonitor, optimizationSystem, predictiveAnalytics, regressionDetection, config, logger)

	// Add a test session
	testSession := &TuningSession{
		ID:        "test_session_1",
		PolicyID:  "balanced",
		StartedAt: time.Now().UTC(),
		Status:    "active",
		Target:    "test_target",
		Objective: "test_objective",
		Priority:  "medium",
		Actions:   make([]*TuningAction, 0),
		Tags:      make(map[string]string),
	}

	apts.mu.Lock()
	apts.activeTunings[testSession.ID] = testSession
	apts.mu.Unlock()

	// Cancel the session
	err := apts.CancelTuningSession(testSession.ID)
	assert.NoError(t, err)

	// Check that session was cancelled
	sessions := apts.GetTuningSessions()
	assert.NotContains(t, sessions, testSession.ID)

	// Try to cancel non-existent session
	err = apts.CancelTuningSession("non_existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPerformanceTuningEngine_GenerateTuningActions(t *testing.T) {
	logger := zap.NewNop()
	config := TuningConfig{}

	engine := NewPerformanceTuningEngine(config, logger)

	// Create test policy
	policy := &TuningPolicy{
		ID:   "test_policy",
		Name: "Test Policy",
		Type: "balanced",
		Parameters: TuningParameters{
			ResponseTime: struct {
				TargetImprovement float64 `json:"target_improvement"`
				MaxDegradation    float64 `json:"max_degradation"`
				AdjustmentStep    float64 `json:"adjustment_step"`
			}{
				TargetImprovement: 10.0,
				MaxDegradation:    3.0,
				AdjustmentStep:    3.0,
			},
			Throughput: struct {
				TargetImprovement float64 `json:"target_improvement"`
				MaxDegradation    float64 `json:"max_degradation"`
				AdjustmentStep    float64 `json:"adjustment_step"`
			}{
				TargetImprovement: 8.0,
				MaxDegradation:    2.0,
				AdjustmentStep:    2.5,
			},
			Resource: struct {
				CPUTarget     float64 `json:"cpu_target"`
				MemoryTarget  float64 `json:"memory_target"`
				DiskTarget    float64 `json:"disk_target"`
				NetworkTarget float64 `json:"network_target"`
			}{
				CPUTarget:     75.0,
				MemoryTarget:  80.0,
				DiskTarget:    85.0,
				NetworkTarget: 70.0,
			},
			Frequency: struct {
				CheckInterval     time.Duration `json:"check_interval"`
				AdjustmentDelay   time.Duration `json:"adjustment_delay"`
				StabilizationTime time.Duration `json:"stabilization_time"`
			}{
				CheckInterval:     5 * time.Minute,
				AdjustmentDelay:   3 * time.Minute,
				StabilizationTime: 2 * time.Minute,
			},
		},
	}

	// Create test metrics that should generate actions
	metrics := &PerformanceMetrics{
		ResponseTime: struct {
			Current     time.Duration `json:"current"`
			Expected    time.Duration `json:"expected"`
			Improvement float64       `json:"improvement"`
		}{
			Current:  600 * time.Millisecond, // 20% above expected
			Expected: 500 * time.Millisecond,
		},
		Throughput: struct {
			Current     float64 `json:"current"`
			Expected    float64 `json:"expected"`
			Improvement float64 `json:"improvement"`
		}{
			Current:  800.0, // 20% below expected
			Expected: 1000.0,
		},
		SuccessRate: struct {
			Current     float64 `json:"current"`
			Expected    float64 `json:"expected"`
			Improvement float64 `json:"improvement"`
		}{
			Current:  0.90,
			Expected: 0.95,
		},
		ResourceUsage: struct {
			CPU struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			} `json:"cpu"`
			Memory struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			} `json:"memory"`
			Disk struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			} `json:"disk"`
		}{
			CPU: struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			}{
				Current: 85.0, // Above target
			},
			Memory: struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			}{
				Current: 90.0, // Above target
			},
			Disk: struct {
				Current     float64 `json:"current"`
				Expected    float64 `json:"expected"`
				Improvement float64 `json:"improvement"`
			}{
				Current: 70.0, // Below target
			},
		},
	}

	// Generate tuning actions
	actions := engine.GenerateTuningActions(policy, metrics)

	// Should generate multiple actions
	assert.NotEmpty(t, actions)

	// Check for specific action types
	foundResponseTime := false
	foundThroughput := false
	foundCPU := false
	foundMemory := false

	for _, action := range actions {
		switch action.Category {
		case "response_time":
			foundResponseTime = true
			assert.Equal(t, "parameter", action.Type)
			assert.Equal(t, "optimize_response_time", action.Action)
		case "throughput":
			foundThroughput = true
			assert.Equal(t, "parameter", action.Type)
			assert.Equal(t, "optimize_throughput", action.Action)
		case "cpu_optimization":
			foundCPU = true
			assert.Equal(t, "resource", action.Type)
			assert.Equal(t, "optimize_cpu_usage", action.Action)
		case "memory_optimization":
			foundMemory = true
			assert.Equal(t, "resource", action.Type)
			assert.Equal(t, "optimize_memory_usage", action.Action)
		}
	}

	// Verify that expected actions were generated
	assert.True(t, foundResponseTime, "Response time tuning action should be generated")
	assert.True(t, foundThroughput, "Throughput tuning action should be generated")
	assert.True(t, foundCPU, "CPU optimization action should be generated")
	assert.True(t, foundMemory, "Memory optimization action should be generated")
}

func TestPerformanceTuningEngine_ExecuteAction(t *testing.T) {
	logger := zap.NewNop()
	config := TuningConfig{}

	engine := NewPerformanceTuningEngine(config, logger)

	// Test different action types
	testCases := []struct {
		category string
		expected bool
	}{
		{"response_time", true},
		{"throughput", true},
		{"cpu_optimization", true},
		{"memory_optimization", true},
		{"unknown", false},
	}

	for _, tc := range testCases {
		action := &TuningAction{
			ID:       fmt.Sprintf("test_action_%s", tc.category),
			Category: tc.category,
			Action:   fmt.Sprintf("test_%s_action", tc.category),
		}

		success := engine.ExecuteAction(action)
		assert.Equal(t, tc.expected, success, "Action execution for %s should return %v", tc.category, tc.expected)
	}
}

func TestPerformanceTuningEngine_RollbackAction(t *testing.T) {
	logger := zap.NewNop()
	config := TuningConfig{}

	engine := NewPerformanceTuningEngine(config, logger)

	// Create test action
	action := &TuningAction{
		ID:          "test_action_1",
		Category:    "response_time",
		Action:      "optimize_response_time",
		Parameter:   "response_time_target",
		OldValue:    500 * time.Millisecond,
		NewValue:    400 * time.Millisecond,
		Description: "Optimize response time",
	}

	// Rollback action
	success := engine.RollbackAction(action)
	assert.True(t, success)

	// Check that values were swapped
	assert.Equal(t, 400*time.Millisecond, action.OldValue)
	assert.Equal(t, 500*time.Millisecond, action.NewValue)
	assert.Contains(t, action.Description, "Rolled back")
}

func TestTuningPolicy_Validation(t *testing.T) {
	// Test conservative policy validation
	conservativePolicy := &TuningPolicy{
		ID:          "conservative",
		Name:        "Conservative Tuning Policy",
		Description: "Conservative performance tuning with minimal risk",
		Type:        "conservative",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		IsActive:    true,
		Environment: "production",
		Priority:    "low",
		Tags:        make(map[string]string),
		Parameters: TuningParameters{
			ResponseTime: struct {
				TargetImprovement float64 `json:"target_improvement"`
				MaxDegradation    float64 `json:"max_degradation"`
				AdjustmentStep    float64 `json:"adjustment_step"`
			}{
				TargetImprovement: 5.0,
				MaxDegradation:    2.0,
				AdjustmentStep:    2.0,
			},
			Throughput: struct {
				TargetImprovement float64 `json:"target_improvement"`
				MaxDegradation    float64 `json:"max_degradation"`
				AdjustmentStep    float64 `json:"adjustment_step"`
			}{
				TargetImprovement: 3.0,
				MaxDegradation:    1.0,
				AdjustmentStep:    1.5,
			},
			Resource: struct {
				CPUTarget     float64 `json:"cpu_target"`
				MemoryTarget  float64 `json:"memory_target"`
				DiskTarget    float64 `json:"disk_target"`
				NetworkTarget float64 `json:"network_target"`
			}{
				CPUTarget:     70.0,
				MemoryTarget:  75.0,
				DiskTarget:    80.0,
				NetworkTarget: 60.0,
			},
			Frequency: struct {
				CheckInterval     time.Duration `json:"check_interval"`
				AdjustmentDelay   time.Duration `json:"adjustment_delay"`
				StabilizationTime time.Duration `json:"stabilization_time"`
			}{
				CheckInterval:     10 * time.Minute,
				AdjustmentDelay:   5 * time.Minute,
				StabilizationTime: 3 * time.Minute,
			},
		},
		SafetyLimits: SafetyLimits{
			MaxCPUUsage:     85.0,
			MaxMemoryUsage:  90.0,
			MaxDiskUsage:    95.0,
			MaxNetworkUsage: 80.0,
			MinResponseTime: 100 * time.Millisecond,
			MaxResponseTime: 2000 * time.Millisecond,
			MinSuccessRate:  0.95,
			MaxErrorRate:    0.05,
		},
	}

	// Validate policy
	assert.NotEmpty(t, conservativePolicy.ID)
	assert.NotEmpty(t, conservativePolicy.Name)
	assert.NotEmpty(t, conservativePolicy.Description)
	assert.NotEmpty(t, conservativePolicy.Type)
	assert.NotZero(t, conservativePolicy.CreatedAt)
	assert.NotZero(t, conservativePolicy.UpdatedAt)
	assert.True(t, conservativePolicy.IsActive)
	assert.NotEmpty(t, conservativePolicy.Environment)
	assert.NotEmpty(t, conservativePolicy.Priority)

	// Validate parameters
	assert.Greater(t, conservativePolicy.Parameters.ResponseTime.TargetImprovement, 0.0)
	assert.Greater(t, conservativePolicy.Parameters.Throughput.TargetImprovement, 0.0)
	assert.Greater(t, conservativePolicy.Parameters.Resource.CPUTarget, 0.0)
	assert.Less(t, conservativePolicy.Parameters.Resource.CPUTarget, 100.0)
	assert.Greater(t, conservativePolicy.Parameters.Frequency.CheckInterval, 0)

	// Validate safety limits
	assert.Greater(t, conservativePolicy.SafetyLimits.MaxCPUUsage, 0.0)
	assert.Less(t, conservativePolicy.SafetyLimits.MaxCPUUsage, 100.0)
	assert.Greater(t, conservativePolicy.SafetyLimits.MinResponseTime, 0)
	assert.Greater(t, conservativePolicy.SafetyLimits.MinSuccessRate, 0.0)
	assert.Less(t, conservativePolicy.SafetyLimits.MinSuccessRate, 1.0)
}
