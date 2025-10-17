package main

import (
	"context"
	"fmt"
	"log"

	"kyb-platform/internal/modules/classification_optimization"
	"go.uber.org/zap"
)

func main() {
	logger := zap.NewNop()

	// Set up dependencies
	algorithmRegistry := classification_optimization.NewAlgorithmRegistry(logger)
	performanceTracker := classification_optimization.NewPerformanceTracker(logger)
	accuracyValidator := classification_optimization.NewAccuracyValidator(nil, logger)
	accuracyValidator.SetAlgorithmRegistry(algorithmRegistry)

	// Create workflow
	workflow := classification_optimization.NewImprovementWorkflow(nil, logger)
	workflow.SetDependencies(algorithmRegistry, performanceTracker, accuracyValidator, nil)

	// Register test algorithm
	algorithm := &classification_optimization.ClassificationAlgorithm{
		ID:                  "test-algorithm",
		Name:                "Test Algorithm",
		Category:            "test-category",
		ConfidenceThreshold: 0.7,
		IsActive:            true,
	}
	algorithmRegistry.RegisterAlgorithm(algorithm)

	// Start continuous improvement
	fmt.Println("Starting continuous improvement workflow...")
	execution, err := workflow.StartContinuousImprovement(context.Background(), "test-algorithm")
	if err != nil {
		log.Fatalf("Failed to start workflow: %v", err)
	}

	fmt.Printf("Workflow completed successfully!\n")
	fmt.Printf("Workflow ID: %s\n", execution.ID)
	fmt.Printf("Status: %s\n", execution.Status)
	fmt.Printf("Type: %s\n", execution.Type)
	fmt.Printf("Improvement Score: %.4f\n", execution.ImprovementScore)
	fmt.Printf("Iterations: %d\n", len(execution.Iterations))

	// Test A/B testing
	fmt.Println("\nStarting A/B testing workflow...")

	// Register second algorithm
	algorithmB := &classification_optimization.ClassificationAlgorithm{
		ID:                  "algorithm-b",
		Name:                "Algorithm B",
		Category:            "test-category",
		ConfidenceThreshold: 0.8,
		IsActive:            true,
	}
	algorithmRegistry.RegisterAlgorithm(algorithmB)

	// Create test cases
	testCases := make([]*classification_optimization.TestCase, 500)
	for i := 0; i < 500; i++ {
		testCases[i] = &classification_optimization.TestCase{
			ID:             fmt.Sprintf("test-%d", i),
			Input:          map[string]interface{}{"name": fmt.Sprintf("Company %d", i)},
			ExpectedOutput: "technology",
			TestCaseType:   "standard",
			Difficulty:     "easy",
		}
	}

	abExecution, err := workflow.StartABTesting(context.Background(), "test-algorithm", "algorithm-b", testCases)
	if err != nil {
		log.Fatalf("Failed to start A/B testing: %v", err)
	}

	fmt.Printf("A/B Testing completed successfully!\n")
	fmt.Printf("Workflow ID: %s\n", abExecution.ID)
	fmt.Printf("Status: %s\n", abExecution.Status)
	fmt.Printf("Type: %s\n", abExecution.Type)
	fmt.Printf("Improvement Score: %.4f\n", abExecution.ImprovementScore)

	// Get workflow history
	history := workflow.GetWorkflowHistory()
	fmt.Printf("\nWorkflow History: %d workflows\n", len(history))

	// Get active workflows
	active := workflow.GetActiveWorkflows()
	fmt.Printf("Active Workflows: %d workflows\n", len(active))

	fmt.Println("\nAll tests completed successfully!")
}
