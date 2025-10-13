package routes

import (
	"github.com/gorilla/mux"

	"kyb-platform/services/risk-assessment-service/internal/handlers"
)

// RegisterWebhookRoutes registers webhook-related routes
func RegisterWebhookRoutes(router *mux.Router, webhookHandlers *handlers.WebhookHandlers) {
	// Webhook management routes
	webhookRouter := router.PathPrefix("/webhooks").Subrouter()

	// CRUD operations
	webhookRouter.HandleFunc("", webhookHandlers.CreateWebhook).Methods("POST")
	webhookRouter.HandleFunc("", webhookHandlers.ListWebhooks).Methods("GET")
	webhookRouter.HandleFunc("/{id}", webhookHandlers.GetWebhook).Methods("GET")
	webhookRouter.HandleFunc("/{id}", webhookHandlers.UpdateWebhook).Methods("PUT", "PATCH")
	webhookRouter.HandleFunc("/{id}", webhookHandlers.DeleteWebhook).Methods("DELETE")

	// Webhook operations
	webhookRouter.HandleFunc("/{id}/test", webhookHandlers.TestWebhook).Methods("POST")
	webhookRouter.HandleFunc("/{id}/stats", webhookHandlers.GetWebhookStats).Methods("GET")
	webhookRouter.HandleFunc("/{id}/deliveries", webhookHandlers.GetWebhookDeliveries).Methods("GET")

	// Delivery management
	deliveryRouter := router.PathPrefix("/webhook-deliveries").Subrouter()
	deliveryRouter.HandleFunc("/{delivery_id}/retry", webhookHandlers.RetryWebhookDelivery).Methods("POST")
}
