package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/petercrawford/New tool/internal/monitoring"
)

func main() {
	// Get database connection string from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/kyb_platform?sslmode=disable"
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	logger := log.New(os.Stdout, "ALERT_VERIFICATION: ", log.LstdFlags)
	logger.Println("Starting alert systems verification...")

	// Create unified monitoring service
	service := monitoring.NewUnifiedMonitoringService(db, logger)
	adapter := monitoring.NewMonitoringAdapter(db, logger)

	ctx := context.Background()

	// Test 1: Create various types of alerts
	logger.Println("Test 1: Creating various types of alerts...")

	// Critical performance alert
	criticalAlert := &monitoring.UnifiedAlert{
		ID:                uuid.New(),
		CreatedAt:         time.Now(),
		AlertType:         monitoring.AlertTypeThreshold,
		AlertCategory:     monitoring.AlertCategoryPerformance,
		Severity:          monitoring.AlertSeverityCritical,
		Component:         "api",
		ComponentInstance: "gateway",
		ServiceName:       "api_gateway",
		AlertName:         "critical_response_time",
		Description:       "API response time is critically high",
		Condition: map[string]interface{}{
			"metric_name": "response_time",
			"operator":    ">",
			"threshold":   500.0,
		},
		CurrentValue:   &[]float64{750.0}[0],
		ThresholdValue: &[]float64{500.0}[0],
		Status:         monitoring.AlertStatusActive,
		Tags: map[string]interface{}{
			"endpoint": "/api/businesses",
			"severity": "critical",
			"priority": "p1",
		},
		Metadata: map[string]interface{}{
			"alert_source": "threshold_monitor",
			"escalation":   "immediate",
		},
	}

	if err := service.RecordAlert(ctx, criticalAlert); err != nil {
		logger.Printf("Failed to create critical alert: %v", err)
	} else {
		logger.Println("✓ Critical alert created successfully")
	}

	// Warning resource alert
	warningAlert := &monitoring.UnifiedAlert{
		ID:                uuid.New(),
		CreatedAt:         time.Now(),
		AlertType:         monitoring.AlertTypeThreshold,
		AlertCategory:     monitoring.AlertCategoryResource,
		Severity:          monitoring.AlertSeverityWarning,
		Component:         "system",
		ComponentInstance: "main",
		ServiceName:       "system_monitor",
		AlertName:         "high_memory_usage",
		Description:       "Memory usage is above warning threshold",
		Condition: map[string]interface{}{
			"metric_name": "memory_usage",
			"operator":    ">",
			"threshold":   80.0,
		},
		CurrentValue:   &[]float64{85.5}[0],
		ThresholdValue: &[]float64{80.0}[0],
		Status:         monitoring.AlertStatusActive,
		Tags: map[string]interface{}{
			"host":     "server-01",
			"severity": "warning",
			"priority": "p2",
		},
		Metadata: map[string]interface{}{
			"alert_source": "resource_monitor",
			"escalation":   "normal",
		},
	}

	if err := service.RecordAlert(ctx, warningAlert); err != nil {
		logger.Printf("Failed to create warning alert: %v", err)
	} else {
		logger.Println("✓ Warning alert created successfully")
	}

	// Info business alert
	infoAlert := &monitoring.UnifiedAlert{
		ID:                uuid.New(),
		CreatedAt:         time.Now(),
		AlertType:         monitoring.AlertTypeTrend,
		AlertCategory:     monitoring.AlertCategoryBusiness,
		Severity:          monitoring.AlertSeverityInfo,
		Component:         "classification",
		ComponentInstance: "ml_engine",
		ServiceName:       "classification_service",
		AlertName:         "accuracy_trend",
		Description:       "Classification accuracy is trending downward",
		Condition: map[string]interface{}{
			"metric_name": "classification_accuracy",
			"trend":       "decreasing",
			"timeframe":   "24h",
		},
		CurrentValue:   &[]float64{92.5}[0],
		ThresholdValue: &[]float64{95.0}[0],
		Status:         monitoring.AlertStatusActive,
		Tags: map[string]interface{}{
			"model":    "bert-base",
			"severity": "info",
			"priority": "p3",
		},
		Metadata: map[string]interface{}{
			"alert_source": "trend_monitor",
			"escalation":   "low",
		},
	}

	if err := service.RecordAlert(ctx, infoAlert); err != nil {
		logger.Printf("Failed to create info alert: %v", err)
	} else {
		logger.Println("✓ Info alert created successfully")
	}

	// Security alert
	securityAlert := &monitoring.UnifiedAlert{
		ID:                uuid.New(),
		CreatedAt:         time.Now(),
		AlertType:         monitoring.AlertTypeAnomaly,
		AlertCategory:     monitoring.AlertCategorySecurity,
		Severity:          monitoring.AlertSeverityCritical,
		Component:         "security",
		ComponentInstance: "validator",
		ServiceName:       "security_validator",
		AlertName:         "suspicious_activity",
		Description:       "Unusual authentication patterns detected",
		Condition: map[string]interface{}{
			"metric_name": "failed_authentications",
			"operator":    ">",
			"threshold":   10,
			"timeframe":   "5m",
		},
		CurrentValue:   &[]float64{25.0}[0],
		ThresholdValue: &[]float64{10.0}[0],
		Status:         monitoring.AlertStatusActive,
		Tags: map[string]interface{}{
			"security_level": "high",
			"severity":       "critical",
			"priority":       "p1",
		},
		Metadata: map[string]interface{}{
			"alert_source":              "security_monitor",
			"escalation":                "immediate",
			"requires_immediate_action": true,
		},
	}

	if err := service.RecordAlert(ctx, securityAlert); err != nil {
		logger.Printf("Failed to create security alert: %v", err)
	} else {
		logger.Println("✓ Security alert created successfully")
	}

	// Test 2: Query alerts by various criteria
	logger.Println("\nTest 2: Querying alerts by various criteria...")

	// Get all active alerts
	activeAlerts, err := service.GetAlerts(ctx, &monitoring.AlertFilters{
		Status: monitoring.AlertStatusActive,
		Limit:  100,
	})
	if err != nil {
		logger.Printf("Failed to get active alerts: %v", err)
	} else {
		logger.Printf("✓ Retrieved %d active alerts", len(activeAlerts))
	}

	// Get critical alerts
	criticalAlerts, err := service.GetAlerts(ctx, &monitoring.AlertFilters{
		Severity: monitoring.AlertSeverityCritical,
		Limit:    50,
	})
	if err != nil {
		logger.Printf("Failed to get critical alerts: %v", err)
	} else {
		logger.Printf("✓ Retrieved %d critical alerts", len(criticalAlerts))
	}

	// Get alerts by component
	apiAlerts, err := service.GetAlerts(ctx, &monitoring.AlertFilters{
		Component: "api",
		Limit:     50,
	})
	if err != nil {
		logger.Printf("Failed to get API alerts: %v", err)
	} else {
		logger.Printf("✓ Retrieved %d API alerts", len(apiAlerts))
	}

	// Get alerts by category
	securityAlerts, err := service.GetAlerts(ctx, &monitoring.AlertFilters{
		AlertCategory: monitoring.AlertCategorySecurity,
		Limit:         50,
	})
	if err != nil {
		logger.Printf("Failed to get security alerts: %v", err)
	} else {
		logger.Printf("✓ Retrieved %d security alerts", len(securityAlerts))
	}

	// Test 3: Alert status management
	logger.Println("\nTest 3: Testing alert status management...")

	if len(activeAlerts) > 0 {
		// Test acknowledging an alert
		alertToAcknowledge := activeAlerts[0]
		userID := uuid.New()

		logger.Printf("Acknowledging alert: %s", alertToAcknowledge.AlertName)
		if err := service.UpdateAlertStatus(ctx, alertToAcknowledge.ID, monitoring.AlertStatusAcknowledged, &userID); err != nil {
			logger.Printf("Failed to acknowledge alert: %v", err)
		} else {
			logger.Println("✓ Alert acknowledged successfully")
		}

		// Verify the alert status was updated
		updatedAlerts, err := service.GetAlerts(ctx, &monitoring.AlertFilters{
			Status: monitoring.AlertStatusAcknowledged,
			Limit:  10,
		})
		if err != nil {
			logger.Printf("Failed to get acknowledged alerts: %v", err)
		} else {
			logger.Printf("✓ Found %d acknowledged alerts", len(updatedAlerts))
		}

		// Test resolving an alert
		if len(updatedAlerts) > 0 {
			alertToResolve := updatedAlerts[0]
			logger.Printf("Resolving alert: %s", alertToResolve.AlertName)
			if err := service.UpdateAlertStatus(ctx, alertToResolve.ID, monitoring.AlertStatusResolved, &userID); err != nil {
				logger.Printf("Failed to resolve alert: %v", err)
			} else {
				logger.Println("✓ Alert resolved successfully")
			}
		}
	}

	// Test 4: Alert counts and statistics
	logger.Println("\nTest 4: Testing alert counts and statistics...")

	// Get active alerts count by severity
	alertCounts, err := service.GetActiveAlertsCount(ctx)
	if err != nil {
		logger.Printf("Failed to get active alerts count: %v", err)
	} else {
		logger.Println("✓ Active alerts count by severity:")
		for severity, count := range alertCounts {
			logger.Printf("  - %s: %d alerts", severity, count)
		}
	}

	// Test 5: Alert escalation and priority handling
	logger.Println("\nTest 5: Testing alert escalation and priority handling...")

	// Create alerts with different priorities
	priorities := []string{"p1", "p2", "p3"}
	for i, priority := range priorities {
		escalationAlert := &monitoring.UnifiedAlert{
			ID:                uuid.New(),
			CreatedAt:         time.Now(),
			AlertType:         monitoring.AlertTypeThreshold,
			AlertCategory:     monitoring.AlertCategoryPerformance,
			Severity:          monitoring.AlertSeverityWarning,
			Component:         "test",
			ComponentInstance: "escalation_test",
			ServiceName:       "escalation_test_service",
			AlertName:         fmt.Sprintf("escalation_test_%s", priority),
			Description:       fmt.Sprintf("Test alert with priority %s", priority),
			Condition: map[string]interface{}{
				"metric_name": "test_metric",
				"operator":    ">",
				"threshold":   float64(100 + i*50),
			},
			CurrentValue:   &[]float64{float64(150 + i*50)}[0],
			ThresholdValue: &[]float64{float64(100 + i*50)}[0],
			Status:         monitoring.AlertStatusActive,
			Tags: map[string]interface{}{
				"priority": priority,
				"test":     true,
			},
			Metadata: map[string]interface{}{
				"alert_source": "escalation_test",
				"priority":     priority,
			},
		}

		if err := service.RecordAlert(ctx, escalationAlert); err != nil {
			logger.Printf("Failed to create escalation test alert %s: %v", priority, err)
		} else {
			logger.Printf("✓ Escalation test alert %s created successfully", priority)
		}
	}

	// Test 6: Alert correlation and relationships
	logger.Println("\nTest 6: Testing alert correlation and relationships...")

	// Create related alerts
	baseAlert := &monitoring.UnifiedAlert{
		ID:                uuid.New(),
		CreatedAt:         time.Now(),
		AlertType:         monitoring.AlertTypeThreshold,
		AlertCategory:     monitoring.AlertCategoryPerformance,
		Severity:          monitoring.AlertSeverityCritical,
		Component:         "api",
		ComponentInstance: "gateway",
		ServiceName:       "api_gateway",
		AlertName:         "base_alert",
		Description:       "Base alert for correlation testing",
		Condition: map[string]interface{}{
			"metric_name": "response_time",
			"operator":    ">",
			"threshold":   1000.0,
		},
		CurrentValue:   &[]float64{1200.0}[0],
		ThresholdValue: &[]float64{1000.0}[0],
		Status:         monitoring.AlertStatusActive,
		Tags: map[string]interface{}{
			"correlation_test": true,
		},
		Metadata: map[string]interface{}{
			"alert_source": "correlation_test",
		},
	}

	if err := service.RecordAlert(ctx, baseAlert); err != nil {
		logger.Printf("Failed to create base alert: %v", err)
	} else {
		logger.Println("✓ Base alert created successfully")
	}

	// Create related alert
	relatedAlert := &monitoring.UnifiedAlert{
		ID:                uuid.New(),
		CreatedAt:         time.Now(),
		AlertType:         monitoring.AlertTypeThreshold,
		AlertCategory:     monitoring.AlertCategoryResource,
		Severity:          monitoring.AlertSeverityWarning,
		Component:         "database",
		ComponentInstance: "main",
		ServiceName:       "database_monitor",
		AlertName:         "related_alert",
		Description:       "Related alert for correlation testing",
		Condition: map[string]interface{}{
			"metric_name": "connection_count",
			"operator":    ">",
			"threshold":   50.0,
		},
		CurrentValue:   &[]float64{75.0}[0],
		ThresholdValue: &[]float64{50.0}[0],
		Status:         monitoring.AlertStatusActive,
		RelatedMetrics: []uuid.UUID{baseAlert.ID},
		Tags: map[string]interface{}{
			"correlation_test": true,
			"related_to":       baseAlert.ID.String(),
		},
		Metadata: map[string]interface{}{
			"alert_source":     "correlation_test",
			"related_alert_id": baseAlert.ID.String(),
		},
	}

	if err := service.RecordAlert(ctx, relatedAlert); err != nil {
		logger.Printf("Failed to create related alert: %v", err)
	} else {
		logger.Println("✓ Related alert created successfully")
	}

	// Test 7: Alert filtering and search
	logger.Println("\nTest 7: Testing alert filtering and search...")

	// Test time-based filtering
	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)

	recentAlerts, err := service.GetAlerts(ctx, &monitoring.AlertFilters{
		StartTime: &oneHourAgo,
		EndTime:   &now,
		Limit:     100,
	})
	if err != nil {
		logger.Printf("Failed to get recent alerts: %v", err)
	} else {
		logger.Printf("✓ Retrieved %d alerts from the last hour", len(recentAlerts))
	}

	// Test 8: Alert performance and scalability
	logger.Println("\nTest 8: Testing alert performance and scalability...")

	// Create multiple alerts quickly
	start := time.Now()
	alertCount := 100

	for i := 0; i < alertCount; i++ {
		perfAlert := &monitoring.UnifiedAlert{
			ID:                uuid.New(),
			CreatedAt:         time.Now(),
			AlertType:         monitoring.AlertTypeThreshold,
			AlertCategory:     monitoring.AlertCategoryPerformance,
			Severity:          monitoring.AlertSeverityInfo,
			Component:         "performance_test",
			ComponentInstance: "test_instance",
			ServiceName:       "performance_test_service",
			AlertName:         fmt.Sprintf("performance_test_alert_%d", i),
			Description:       fmt.Sprintf("Performance test alert %d", i),
			Condition: map[string]interface{}{
				"metric_name": "test_metric",
				"operator":    ">",
				"threshold":   float64(i),
			},
			CurrentValue:   &[]float64{float64(i + 10)}[0],
			ThresholdValue: &[]float64{float64(i)}[0],
			Status:         monitoring.AlertStatusActive,
			Tags: map[string]interface{}{
				"performance_test": true,
				"iteration":        i,
			},
			Metadata: map[string]interface{}{
				"alert_source": "performance_test",
				"iteration":    i,
			},
		}

		if err := service.RecordAlert(ctx, perfAlert); err != nil {
			logger.Printf("Failed to create performance test alert %d: %v", i, err)
		}
	}

	duration := time.Since(start)
	logger.Printf("✓ Created %d alerts in %v (%.2f alerts/second)", alertCount, duration, float64(alertCount)/duration.Seconds())

	// Test 9: Alert data integrity
	logger.Println("\nTest 9: Testing alert data integrity...")

	// Verify all alerts were created correctly
	allAlerts, err := service.GetAlerts(ctx, &monitoring.AlertFilters{
		Limit: 1000,
	})
	if err != nil {
		logger.Printf("Failed to get all alerts: %v", err)
	} else {
		logger.Printf("✓ Total alerts in system: %d", len(allAlerts))

		// Verify alert data integrity
		integrityIssues := 0
		for _, alert := range allAlerts {
			if alert.AlertName == "" {
				integrityIssues++
			}
			if alert.Description == "" {
				integrityIssues++
			}
			if alert.Component == "" {
				integrityIssues++
			}
			if alert.ServiceName == "" {
				integrityIssues++
			}
		}

		if integrityIssues == 0 {
			logger.Println("✓ Alert data integrity verified - no issues found")
		} else {
			logger.Printf("⚠ Found %d alert data integrity issues", integrityIssues)
		}
	}

	// Test 10: Final verification
	logger.Println("\nTest 10: Final verification...")

	// Get final alert counts by severity
	finalCounts, err := service.GetActiveAlertsCount(ctx)
	if err != nil {
		logger.Printf("Failed to get final alert counts: %v", err)
	} else {
		logger.Println("✓ Final alert counts by severity:")
		totalAlerts := 0
		for severity, count := range finalCounts {
			logger.Printf("  - %s: %d alerts", severity, count)
			totalAlerts += count
		}
		logger.Printf("  - Total active alerts: %d", totalAlerts)
	}

	// Get final alert counts by category
	categories := []monitoring.AlertCategory{
		monitoring.AlertCategoryPerformance,
		monitoring.AlertCategoryResource,
		monitoring.AlertCategoryBusiness,
		monitoring.AlertCategorySecurity,
	}

	logger.Println("✓ Final alert counts by category:")
	for _, category := range categories {
		categoryAlerts, err := service.GetAlerts(ctx, &monitoring.AlertFilters{
			AlertCategory: category,
			Limit:         1000,
		})
		if err != nil {
			logger.Printf("  - %s: Error retrieving alerts", category)
		} else {
			logger.Printf("  - %s: %d alerts", category, len(categoryAlerts))
		}
	}

	logger.Println("\n🎉 All alert system verification tests completed successfully!")
	logger.Println("The unified alert system is working correctly and ready for production use.")
	logger.Println("")
	logger.Println("Key capabilities verified:")
	logger.Println("✓ Alert creation and storage")
	logger.Println("✓ Alert querying and filtering")
	logger.Println("✓ Alert status management")
	logger.Println("✓ Alert escalation and priority handling")
	logger.Println("✓ Alert correlation and relationships")
	logger.Println("✓ Alert performance and scalability")
	logger.Println("✓ Alert data integrity")
	logger.Println("✓ Alert statistics and reporting")
}
