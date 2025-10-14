package routes

import (
	"github.com/gorilla/mux"

	"kyb-platform/services/risk-assessment-service/internal/handlers"
)

// RegisterWebhookRoutes registers webhook-related routes
func RegisterWebhookRoutes(router *mux.Router, webhookHandlers *handlers.SimpleWebhookHandlers) {
	// Webhook management routes
	webhookRouter := router.PathPrefix("/webhooks").Subrouter()

	// CRUD operations
	webhookRouter.HandleFunc("", webhookHandlers.CreateWebhook).Methods("POST")
	webhookRouter.HandleFunc("", webhookHandlers.ListWebhooks).Methods("GET")
	webhookRouter.HandleFunc("/{id}", webhookHandlers.GetWebhook).Methods("GET")
	webhookRouter.HandleFunc("/{id}", webhookHandlers.UpdateWebhook).Methods("PUT", "PATCH")
	webhookRouter.HandleFunc("/{id}", webhookHandlers.DeleteWebhook).Methods("DELETE")

	// Note: Additional webhook operations (test, stats, deliveries, retry)
	// are not implemented in SimpleWebhookHandlers
	// These would need to be added to the handler or a more advanced webhook handler
}
