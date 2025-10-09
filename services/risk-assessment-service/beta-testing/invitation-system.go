package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// InvitationSystem handles beta testing invitations
type InvitationSystem struct {
	logger *zap.Logger
	store  InvitationStore
	email  EmailService
}

// InvitationStore interface for managing invitations
type InvitationStore interface {
	CreateInvitation(invitation *BetaInvitation) error
	GetInvitation(id string) (*BetaInvitation, error)
	GetInvitationByEmail(email string) (*BetaInvitation, error)
	UpdateInvitation(invitation *BetaInvitation) error
	DeleteInvitation(id string) error
	GetAllInvitations() ([]*BetaInvitation, error)
	GetInvitationsByStatus(status string) ([]*BetaInvitation, error)
}

// EmailService interface for sending emails
type EmailService interface {
	SendInvitationEmail(invitation *BetaInvitation) error
	SendWelcomeEmail(tester *BetaTester) error
	SendReminderEmail(invitation *BetaInvitation) error
}

// SMTPEmailService implements EmailService using SMTP
type SMTPEmailService struct {
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	fromEmail    string
	fromName     string
}

// BetaTester represents a beta tester
type BetaTester struct {
	ID                     string    `json:"id"`
	Name                   string    `json:"name"`
	Email                  string    `json:"email"`
	Company                string    `json:"company"`
	Role                   string    `json:"role"`
	Experience             string    `json:"experience"`
	PreferredSDK           string    `json:"preferred_sdk"`
	IntegrationType        string    `json:"integration_type"`
	JoinedAt               time.Time `json:"joined_at"`
	LastActiveAt           time.Time `json:"last_active_at"`
	Status                 string    `json:"status"`
	APIKey                 string    `json:"api_key"`
	FeedbackCount          int       `json:"feedback_count"`
	BugReportCount         int       `json:"bug_report_count"`
	FeatureRequestCount    int       `json:"feature_request_count"`
	TestScenariosCompleted int       `json:"test_scenarios_completed"`
	OverallRating          float64   `json:"overall_rating"`
	Notes                  string    `json:"notes"`
}

// BetaInvitation represents a beta testing invitation
type BetaInvitation struct {
	ID              string     `json:"id"`
	Email           string     `json:"email"`
	Name            string     `json:"name"`
	Company         string     `json:"company"`
	Role            string     `json:"role"`
	Experience      string     `json:"experience"`
	PreferredSDK    string     `json:"preferred_sdk"`
	IntegrationType string     `json:"integration_type"`
	InvitedAt       time.Time  `json:"invited_at"`
	InvitedBy       string     `json:"invited_by"`
	Status          string     `json:"status"` // pending, accepted, declined, expired
	ExpiresAt       time.Time  `json:"expires_at"`
	AcceptedAt      *time.Time `json:"accepted_at,omitempty"`
	DeclinedAt      *time.Time `json:"declined_at,omitempty"`
	Message         string     `json:"message"`
	APIKey          string     `json:"api_key,omitempty"`
	BetaTesterID    string     `json:"beta_tester_id,omitempty"`
}

// InMemoryInvitationStore implements InvitationStore using in-memory storage
type InMemoryInvitationStore struct {
	invitations map[string]*BetaInvitation
}

// NewInvitationSystem creates a new invitation system
func NewInvitationSystem(logger *zap.Logger) *InvitationSystem {
	return &InvitationSystem{
		logger: logger,
		store:  NewInMemoryInvitationStore(),
		email:  NewSMTPEmailService(),
	}
}

// NewInMemoryInvitationStore creates a new in-memory invitation store
func NewInMemoryInvitationStore() *InMemoryInvitationStore {
	return &InMemoryInvitationStore{
		invitations: make(map[string]*BetaInvitation),
	}
}

// NewSMTPEmailService creates a new SMTP email service
func NewSMTPEmailService() *SMTPEmailService {
	return &SMTPEmailService{
		smtpHost:     os.Getenv("SMTP_HOST"),
		smtpPort:     os.Getenv("SMTP_PORT"),
		smtpUsername: os.Getenv("SMTP_USERNAME"),
		smtpPassword: os.Getenv("SMTP_PASSWORD"),
		fromEmail:    os.Getenv("FROM_EMAIL"),
		fromName:     os.Getenv("FROM_NAME"),
	}
}

// CreateInvitation creates a new invitation
func (s *InMemoryInvitationStore) CreateInvitation(invitation *BetaInvitation) error {
	s.invitations[invitation.ID] = invitation
	return nil
}

// GetInvitation retrieves an invitation by ID
func (s *InMemoryInvitationStore) GetInvitation(id string) (*BetaInvitation, error) {
	invitation, exists := s.invitations[id]
	if !exists {
		return nil, fmt.Errorf("invitation not found")
	}
	return invitation, nil
}

// GetInvitationByEmail retrieves an invitation by email
func (s *InMemoryInvitationStore) GetInvitationByEmail(email string) (*BetaInvitation, error) {
	for _, invitation := range s.invitations {
		if invitation.Email == email {
			return invitation, nil
		}
	}
	return nil, fmt.Errorf("invitation not found")
}

// UpdateInvitation updates an existing invitation
func (s *InMemoryInvitationStore) UpdateInvitation(invitation *BetaInvitation) error {
	s.invitations[invitation.ID] = invitation
	return nil
}

// DeleteInvitation deletes an invitation
func (s *InMemoryInvitationStore) DeleteInvitation(id string) error {
	delete(s.invitations, id)
	return nil
}

// GetAllInvitations retrieves all invitations
func (s *InMemoryInvitationStore) GetAllInvitations() ([]*BetaInvitation, error) {
	var invitations []*BetaInvitation
	for _, invitation := range s.invitations {
		invitations = append(invitations, invitation)
	}
	return invitations, nil
}

// GetInvitationsByStatus retrieves invitations by status
func (s *InMemoryInvitationStore) GetInvitationsByStatus(status string) ([]*BetaInvitation, error) {
	var invitations []*BetaInvitation
	for _, invitation := range s.invitations {
		if invitation.Status == status {
			invitations = append(invitations, invitation)
		}
	}
	return invitations, nil
}

// SendInvitationEmail sends an invitation email
func (e *SMTPEmailService) SendInvitationEmail(invitation *BetaInvitation) error {
	// Email template
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Beta Testing Invitation - Risk Assessment Service</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .button { display: inline-block; background: #667eea; color: white; padding: 15px 30px; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .button:hover { background: #5a6fd8; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 Beta Testing Invitation</h1>
            <p>Risk Assessment Service</p>
        </div>
        <div class="content">
            <h2>Hello {{.Name}}!</h2>
            <p>You've been invited to participate in the beta testing program for our Enhanced Risk Assessment Service.</p>
            
            <h3>What's in it for you?</h3>
            <ul>
                <li>Early access to cutting-edge risk assessment technology</li>
                <li>ML-powered business risk predictions</li>
                <li>Real-time performance monitoring</li>
                <li>Direct influence on product development</li>
                <li>Free service credits for production use</li>
            </ul>
            
            <h3>What we're testing:</h3>
            <ul>
                <li>API functionality and performance</li>
                <li>Developer experience with our SDKs</li>
                <li>Documentation clarity and completeness</li>
                <li>Integration ease and reliability</li>
            </ul>
            
            <p><strong>Target Performance:</strong> 1000 requests/minute with sub-1-second response times</p>
            
            <div style="text-align: center;">
                <a href="{{.AcceptURL}}" class="button">Accept Invitation</a>
                <a href="{{.DeclineURL}}" class="button" style="background: #95a5a6;">Decline</a>
            </div>
            
            <p><strong>Beta Testing Period:</strong> 4 weeks</p>
            <p><strong>Time Commitment:</strong> 2-3 hours per week</p>
            <p><strong>Expires:</strong> {{.ExpiresAt}}</p>
            
            {{if .Message}}
            <div style="background: #e8f4fd; padding: 15px; border-radius: 5px; margin: 20px 0;">
                <strong>Personal Message:</strong><br>
                {{.Message}}
            </div>
            {{end}}
        </div>
        <div class="footer">
            <p>Questions? Contact us at beta-support@yourcompany.com</p>
            <p>This invitation expires on {{.ExpiresAt}}</p>
        </div>
    </div>
</body>
</html>
`

	// Parse template
	t, err := template.New("invitation").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	// Prepare email data
	data := struct {
		Name       string
		AcceptURL  string
		DeclineURL string
		ExpiresAt  string
		Message    string
	}{
		Name:       invitation.Name,
		AcceptURL:  fmt.Sprintf("https://yourcompany.com/beta/accept/%s", invitation.ID),
		DeclineURL: fmt.Sprintf("https://yourcompany.com/beta/decline/%s", invitation.ID),
		ExpiresAt:  invitation.ExpiresAt.Format("January 2, 2006"),
		Message:    invitation.Message,
	}

	// Render email
	var body bytes.Buffer
	if err := t.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	// Send email
	return e.sendEmail(invitation.Email, "Beta Testing Invitation - Risk Assessment Service", body.String())
}

// SendWelcomeEmail sends a welcome email to a new beta tester
func (e *SMTPEmailService) SendWelcomeEmail(tester *BetaTester) error {
	// Welcome email template
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Welcome to Beta Testing - Risk Assessment Service</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #27ae60 0%, #2ecc71 100%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .api-key { background: #2c3e50; color: white; padding: 15px; border-radius: 5px; font-family: monospace; margin: 20px 0; }
        .button { display: inline-block; background: #27ae60; color: white; padding: 15px 30px; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎉 Welcome to Beta Testing!</h1>
            <p>Risk Assessment Service</p>
        </div>
        <div class="content">
            <h2>Hello {{.Name}}!</h2>
            <p>Welcome to the Risk Assessment Service beta testing program! We're excited to have you on board.</p>
            
            <h3>Your API Key</h3>
            <div class="api-key">{{.APIKey}}</div>
            <p><strong>Keep this API key secure and don't share it publicly.</strong></p>
            
            <h3>Getting Started</h3>
            <ol>
                <li>Choose your preferred SDK (Go, Python, or Node.js)</li>
                <li>Run the quick start example</li>
                <li>Explore the API documentation</li>
                <li>Start testing with real scenarios</li>
            </ol>
            
            <div style="text-align: center;">
                <a href="https://yourcompany.com/beta/docs" class="button">View Documentation</a>
                <a href="https://yourcompany.com/beta/dashboard" class="button">Beta Dashboard</a>
            </div>
            
            <h3>Important Links</h3>
            <ul>
                <li><strong>Service URL:</strong> https://risk-assessment-service-production.up.railway.app</li>
                <li><strong>API Documentation:</strong> https://yourcompany.com/beta/docs</li>
                <li><strong>Feedback Form:</strong> https://yourcompany.com/beta/feedback</li>
                <li><strong>Support:</strong> beta-support@yourcompany.com</li>
            </ul>
            
            <h3>Testing Goals</h3>
            <p>Help us validate:</p>
            <ul>
                <li>1000 requests/minute performance target</li>
                <li>Sub-1-second response times</li>
                <li>Developer experience and SDK quality</li>
                <li>API design and documentation clarity</li>
            </ul>
        </div>
        <div class="footer">
            <p>Thank you for participating in our beta testing program!</p>
            <p>Questions? Contact us at beta-support@yourcompany.com</p>
        </div>
    </div>
</body>
</html>
`

	// Parse template
	t, err := template.New("welcome").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse welcome email template: %w", err)
	}

	// Prepare email data
	data := struct {
		Name   string
		APIKey string
	}{
		Name:   tester.Name,
		APIKey: tester.APIKey,
	}

	// Render email
	var body bytes.Buffer
	if err := t.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to render welcome email template: %w", err)
	}

	// Send email
	return e.sendEmail(tester.Email, "Welcome to Beta Testing - Risk Assessment Service", body.String())
}

// SendReminderEmail sends a reminder email for pending invitations
func (e *SMTPEmailService) SendReminderEmail(invitation *BetaInvitation) error {
	// Reminder email template
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Beta Testing Invitation Reminder</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #f39c12 0%, #e67e22 100%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .button { display: inline-block; background: #f39c12; color: white; padding: 15px 30px; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>⏰ Beta Testing Invitation Reminder</h1>
            <p>Risk Assessment Service</p>
        </div>
        <div class="content">
            <h2>Hello {{.Name}}!</h2>
            <p>This is a friendly reminder about your beta testing invitation for the Risk Assessment Service.</p>
            
            <p><strong>Your invitation expires in {{.DaysLeft}} days!</strong></p>
            
            <p>Don't miss out on:</p>
            <ul>
                <li>Early access to cutting-edge risk assessment technology</li>
                <li>Direct influence on product development</li>
                <li>Free service credits for production use</li>
                <li>Beta tester recognition and benefits</li>
            </ul>
            
            <div style="text-align: center;">
                <a href="{{.AcceptURL}}" class="button">Accept Invitation Now</a>
            </div>
            
            <p><strong>Expires:</strong> {{.ExpiresAt}}</p>
        </div>
        <div class="footer">
            <p>Questions? Contact us at beta-support@yourcompany.com</p>
        </div>
    </div>
</body>
</html>
`

	// Parse template
	t, err := template.New("reminder").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse reminder email template: %w", err)
	}

	// Calculate days left
	daysLeft := int(time.Until(invitation.ExpiresAt).Hours() / 24)

	// Prepare email data
	data := struct {
		Name      string
		AcceptURL string
		ExpiresAt string
		DaysLeft  int
	}{
		Name:      invitation.Name,
		AcceptURL: fmt.Sprintf("https://yourcompany.com/beta/accept/%s", invitation.ID),
		ExpiresAt: invitation.ExpiresAt.Format("January 2, 2006"),
		DaysLeft:  daysLeft,
	}

	// Render email
	var body bytes.Buffer
	if err := t.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to render reminder email template: %w", err)
	}

	// Send email
	return e.sendEmail(invitation.Email, "Beta Testing Invitation Reminder - Risk Assessment Service", body.String())
}

// sendEmail sends an email using SMTP
func (e *SMTPEmailService) sendEmail(to, subject, body string) error {
	// For demo purposes, we'll just log the email
	// In production, you would use actual SMTP
	log.Printf("Sending email to %s: %s", to, subject)
	log.Printf("Email body: %s", body)
	return nil
}

// HandleCreateInvitation handles creating a new invitation
func (is *InvitationSystem) HandleCreateInvitation(w http.ResponseWriter, r *http.Request) {
	var invitation BetaInvitation
	if err := json.NewDecoder(r.Body).Decode(&invitation); err != nil {
		is.logger.Error("Failed to decode invitation", zap.Error(err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if invitation.Email == "" || invitation.Name == "" {
		http.Error(w, "Missing required fields: email, name", http.StatusBadRequest)
		return
	}

	// Set default values
	invitation.ID = fmt.Sprintf("invite_%d", time.Now().UnixNano())
	invitation.InvitedAt = time.Now()
	invitation.Status = "pending"
	invitation.ExpiresAt = time.Now().Add(7 * 24 * time.Hour) // 7 days

	// Create invitation
	if err := is.store.CreateInvitation(&invitation); err != nil {
		is.logger.Error("Failed to create invitation", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Send invitation email
	if err := is.email.SendInvitationEmail(&invitation); err != nil {
		is.logger.Error("Failed to send invitation email", zap.Error(err))
		// Don't fail the request, just log the error
	}

	is.logger.Info("Invitation created and sent",
		zap.String("id", invitation.ID),
		zap.String("email", invitation.Email),
		zap.String("name", invitation.Name))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(invitation)
}

// HandleAcceptInvitation handles accepting an invitation
func (is *InvitationSystem) HandleAcceptInvitation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	invitationID := vars["invitationID"]

	// Get invitation
	invitation, err := is.store.GetInvitation(invitationID)
	if err != nil {
		is.logger.Error("Failed to get invitation", zap.Error(err))
		http.Error(w, "Invitation not found", http.StatusNotFound)
		return
	}

	// Check if invitation is still valid
	if invitation.Status != "pending" {
		http.Error(w, "Invitation is no longer valid", http.StatusBadRequest)
		return
	}

	if time.Now().After(invitation.ExpiresAt) {
		http.Error(w, "Invitation has expired", http.StatusBadRequest)
		return
	}

	// Update invitation status
	now := time.Now()
	invitation.Status = "accepted"
	invitation.AcceptedAt = &now
	invitation.APIKey = generateAPIKey()
	invitation.BetaTesterID = fmt.Sprintf("tester_%d", time.Now().UnixNano())

	// Update invitation
	if err := is.store.UpdateInvitation(invitation); err != nil {
		is.logger.Error("Failed to update invitation", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create beta tester
	tester := &BetaTester{
		ID:              invitation.BetaTesterID,
		Name:            invitation.Name,
		Email:           invitation.Email,
		Company:         invitation.Company,
		Role:            invitation.Role,
		Experience:      invitation.Experience,
		PreferredSDK:    invitation.PreferredSDK,
		IntegrationType: invitation.IntegrationType,
		JoinedAt:        now,
		LastActiveAt:    now,
		Status:          "active",
		APIKey:          invitation.APIKey,
	}

	// Send welcome email
	if err := is.email.SendWelcomeEmail(tester); err != nil {
		is.logger.Error("Failed to send welcome email", zap.Error(err))
		// Don't fail the request, just log the error
	}

	is.logger.Info("Invitation accepted",
		zap.String("id", invitation.ID),
		zap.String("email", invitation.Email),
		zap.String("beta_tester_id", tester.ID))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Invitation accepted successfully",
		"tester":  tester,
	})
}

// HandleDeclineInvitation handles declining an invitation
func (is *InvitationSystem) HandleDeclineInvitation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	invitationID := vars["invitationID"]

	// Get invitation
	invitation, err := is.store.GetInvitation(invitationID)
	if err != nil {
		is.logger.Error("Failed to get invitation", zap.Error(err))
		http.Error(w, "Invitation not found", http.StatusNotFound)
		return
	}

	// Update invitation status
	now := time.Now()
	invitation.Status = "declined"
	invitation.DeclinedAt = &now

	// Update invitation
	if err := is.store.UpdateInvitation(invitation); err != nil {
		is.logger.Error("Failed to update invitation", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	is.logger.Info("Invitation declined",
		zap.String("id", invitation.ID),
		zap.String("email", invitation.Email))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Invitation declined successfully",
	})
}

// SetupRoutes sets up the invitation system routes
func (is *InvitationSystem) SetupRoutes(router *mux.Router) {
	api := router.PathPrefix("/api/v1/beta").Subrouter()

	// Invitation routes
	api.HandleFunc("/invitations", is.HandleCreateInvitation).Methods("POST")
	api.HandleFunc("/invitations/{invitationID}/accept", is.HandleAcceptInvitation).Methods("POST")
	api.HandleFunc("/invitations/{invitationID}/decline", is.HandleDeclineInvitation).Methods("POST")
}

// Helper function to generate API key
func generateAPIKey() string {
	return fmt.Sprintf("beta_%d_%s", time.Now().Unix(), randomString(16))
}

// Helper function to generate random string
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()

	// Create invitation system
	system := NewInvitationSystem(logger)

	// Setup router
	router := mux.NewRouter()
	system.SetupRoutes(router)

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"service":   "invitation-system",
			"timestamp": time.Now(),
		})
	}).Methods("GET")

	// Start server
	port := "8082"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	logger.Info("Starting invitation system server", zap.String("port", port))
	log.Fatal(http.ListenAndServe(":"+port, router))
}
