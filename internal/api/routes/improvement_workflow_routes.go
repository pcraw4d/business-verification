package routes

import (
	"github.com/gorilla/mux"

	"github.com/pcraw4d/business-verification/internal/api/handlers"
)

// RegisterImprovementWorkflowRoutes registers all improvement workflow API routes
func RegisterImprovementWorkflowRoutes(router *mux.Router, handler *handlers.ImprovementWorkflowHandler) {
	// Base path for improvement workflow endpoints
	workflowRouter := router.PathPrefix("/api/v1/workflows").Subrouter()

	// Workflow execution endpoints
	workflowRouter.HandleFunc("/continuous-improvement", handler.StartContinuousImprovement).Methods("POST")
	workflowRouter.HandleFunc("/ab-testing", handler.StartABTesting).Methods("POST")
	workflowRouter.HandleFunc("/history", handler.GetWorkflowHistory).Methods("GET")
	workflowRouter.HandleFunc("/active", handler.GetActiveWorkflows).Methods("GET")
	workflowRouter.HandleFunc("/{workflow_id}", handler.GetWorkflowExecution).Methods("GET")
	workflowRouter.HandleFunc("/{workflow_id}/stop", handler.StopWorkflow).Methods("POST")

	// Analytics and monitoring endpoints
	workflowRouter.HandleFunc("/statistics", handler.GetWorkflowStatistics).Methods("GET")
	workflowRouter.HandleFunc("/recommendations", handler.GetWorkflowRecommendations).Methods("GET")
	workflowRouter.HandleFunc("/metrics", handler.GetWorkflowMetrics).Methods("GET")
}
