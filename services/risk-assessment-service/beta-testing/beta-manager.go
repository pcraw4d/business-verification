package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// BetaManager manages the beta testing program
type BetaManager struct {
	logger *zap.Logger
	store  BetaTesterStore
}

// BetaTesterStore interface for managing beta testers
type BetaTesterStore interface {
	CreateBetaTester(tester *BetaTester) error
	GetBetaTester(id string) (*BetaTester, error)
	GetAllBetaTesters() ([]*BetaTester, error)
	UpdateBetaTester(tester *BetaTester) error
	DeleteBetaTester(id string) error
	GetBetaTesterByEmail(email string) (*BetaTester, error)
}

// InMemoryBetaTesterStore implements BetaTesterStore using in-memory storage
type InMemoryBetaTesterStore struct {
	testers map[string]*BetaTester
}

// BetaTester represents a beta tester
type BetaTester struct {
	ID                     string    `json:"id"`
	Name                   string    `json:"name"`
	Email                  string    `json:"email"`
	Company                string    `json:"company"`
	Role                   string    `json:"role"`
	Experience             string    `json:"experience"`       // beginner, intermediate, advanced
	PreferredSDK           string    `json:"preferred_sdk"`    // go, python, nodejs
	IntegrationType        string    `json:"integration_type"` // web, mobile, desktop, api
	JoinedAt               time.Time `json:"joined_at"`
	LastActiveAt           time.Time `json:"last_active_at"`
	Status                 string    `json:"status"` // active, inactive, completed
	APIKey                 string    `json:"api_key"`
	FeedbackCount          int       `json:"feedback_count"`
	BugReportCount         int       `json:"bug_report_count"`
	FeatureRequestCount    int       `json:"feature_request_count"`
	TestScenariosCompleted int       `json:"test_scenarios_completed"`
	OverallRating          float64   `json:"overall_rating"`
	Notes                  string    `json:"notes"`
}

// BetaTesterInvite represents an invitation to join the beta program
type BetaTesterInvite struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Company   string    `json:"company"`
	InvitedAt time.Time `json:"invited_at"`
	InvitedBy string    `json:"invited_by"`
	Status    string    `json:"status"` // pending, accepted, declined, expired
	ExpiresAt time.Time `json:"expires_at"`
	Message   string    `json:"message"`
}

// BetaProgramStats represents statistics about the beta program
type BetaProgramStats struct {
	TotalInvites         int            `json:"total_invites"`
	AcceptedInvites      int            `json:"accepted_invites"`
	ActiveTesters        int            `json:"active_testers"`
	CompletedTesters     int            `json:"completed_testers"`
	TotalFeedback        int            `json:"total_feedback"`
	TotalBugReports      int            `json:"total_bug_reports"`
	TotalFeatureRequests int            `json:"total_feature_requests"`
	AverageRating        float64        `json:"average_rating"`
	SDKUsage             map[string]int `json:"sdk_usage"`
	IntegrationTypes     map[string]int `json:"integration_types"`
	ExperienceLevels     map[string]int `json:"experience_levels"`
	RecentActivity       []*BetaTester  `json:"recent_activity"`
}

// NewBetaManager creates a new beta manager
func NewBetaManager(logger *zap.Logger) *BetaManager {
	return &BetaManager{
		logger: logger,
		store:  NewInMemoryBetaTesterStore(),
	}
}

// NewInMemoryBetaTesterStore creates a new in-memory beta tester store
func NewInMemoryBetaTesterStore() *InMemoryBetaTesterStore {
	return &InMemoryBetaTesterStore{
		testers: make(map[string]*BetaTester),
	}
}

// CreateBetaTester creates a new beta tester
func (s *InMemoryBetaTesterStore) CreateBetaTester(tester *BetaTester) error {
	s.testers[tester.ID] = tester
	return nil
}

// GetBetaTester retrieves a beta tester by ID
func (s *InMemoryBetaTesterStore) GetBetaTester(id string) (*BetaTester, error) {
	tester, exists := s.testers[id]
	if !exists {
		return nil, fmt.Errorf("beta tester not found")
	}
	return tester, nil
}

// GetAllBetaTesters retrieves all beta testers
func (s *InMemoryBetaTesterStore) GetAllBetaTesters() ([]*BetaTester, error) {
	var testers []*BetaTester
	for _, tester := range s.testers {
		testers = append(testers, tester)
	}
	return testers, nil
}

// UpdateBetaTester updates an existing beta tester
func (s *InMemoryBetaTesterStore) UpdateBetaTester(tester *BetaTester) error {
	s.testers[tester.ID] = tester
	return nil
}

// DeleteBetaTester deletes a beta tester
func (s *InMemoryBetaTesterStore) DeleteBetaTester(id string) error {
	delete(s.testers, id)
	return nil
}

// GetBetaTesterByEmail retrieves a beta tester by email
func (s *InMemoryBetaTesterStore) GetBetaTesterByEmail(email string) (*BetaTester, error) {
	for _, tester := range s.testers {
		if tester.Email == email {
			return tester, nil
		}
	}
	return nil, fmt.Errorf("beta tester not found")
}

// HandleCreateBetaTester handles creating a new beta tester
func (bm *BetaManager) HandleCreateBetaTester(w http.ResponseWriter, r *http.Request) {
	var tester BetaTester
	if err := json.NewDecoder(r.Body).Decode(&tester); err != nil {
		bm.logger.Error("Failed to decode beta tester", zap.Error(err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if tester.Name == "" || tester.Email == "" || tester.Company == "" {
		http.Error(w, "Missing required fields: name, email, company", http.StatusBadRequest)
		return
	}

	// Set default values
	tester.ID = fmt.Sprintf("tester_%d", time.Now().UnixNano())
	tester.JoinedAt = time.Now()
	tester.LastActiveAt = time.Now()
	tester.Status = "active"
	tester.APIKey = generateAPIKey()

	// Create beta tester
	if err := bm.store.CreateBetaTester(&tester); err != nil {
		bm.logger.Error("Failed to create beta tester", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	bm.logger.Info("Beta tester created",
		zap.String("id", tester.ID),
		zap.String("name", tester.Name),
		zap.String("email", tester.Email),
		zap.String("company", tester.Company))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tester)
}

// HandleGetBetaTester handles retrieving a beta tester
func (bm *BetaManager) HandleGetBetaTester(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	testerID := vars["testerID"]

	tester, err := bm.store.GetBetaTester(testerID)
	if err != nil {
		bm.logger.Error("Failed to get beta tester", zap.Error(err))
		http.Error(w, "Beta tester not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tester)
}

// HandleGetAllBetaTesters handles retrieving all beta testers
func (bm *BetaManager) HandleGetAllBetaTesters(w http.ResponseWriter, r *http.Request) {
	testers, err := bm.store.GetAllBetaTesters()
	if err != nil {
		bm.logger.Error("Failed to get all beta testers", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"testers": testers,
		"count":   len(testers),
	})
}

// HandleUpdateBetaTester handles updating a beta tester
func (bm *BetaManager) HandleUpdateBetaTester(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	testerID := vars["testerID"]

	var updates BetaTester
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		bm.logger.Error("Failed to decode beta tester updates", zap.Error(err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Get existing tester
	tester, err := bm.store.GetBetaTester(testerID)
	if err != nil {
		bm.logger.Error("Failed to get beta tester", zap.Error(err))
		http.Error(w, "Beta tester not found", http.StatusNotFound)
		return
	}

	// Update fields
	if updates.Name != "" {
		tester.Name = updates.Name
	}
	if updates.Company != "" {
		tester.Company = updates.Company
	}
	if updates.Role != "" {
		tester.Role = updates.Role
	}
	if updates.Experience != "" {
		tester.Experience = updates.Experience
	}
	if updates.PreferredSDK != "" {
		tester.PreferredSDK = updates.PreferredSDK
	}
	if updates.IntegrationType != "" {
		tester.IntegrationType = updates.IntegrationType
	}
	if updates.Status != "" {
		tester.Status = updates.Status
	}
	if updates.Notes != "" {
		tester.Notes = updates.Notes
	}

	tester.LastActiveAt = time.Now()

	// Update tester
	if err := bm.store.UpdateBetaTester(tester); err != nil {
		bm.logger.Error("Failed to update beta tester", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	bm.logger.Info("Beta tester updated",
		zap.String("id", tester.ID),
		zap.String("name", tester.Name))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tester)
}

// HandleDeleteBetaTester handles deleting a beta tester
func (bm *BetaManager) HandleDeleteBetaTester(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	testerID := vars["testerID"]

	if err := bm.store.DeleteBetaTester(testerID); err != nil {
		bm.logger.Error("Failed to delete beta tester", zap.Error(err))
		http.Error(w, "Beta tester not found", http.StatusNotFound)
		return
	}

	bm.logger.Info("Beta tester deleted", zap.String("id", testerID))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Beta tester deleted successfully",
		"id":      testerID,
	})
}

// HandleGetBetaProgramStats handles retrieving beta program statistics
func (bm *BetaManager) HandleGetBetaProgramStats(w http.ResponseWriter, r *http.Request) {
	testers, err := bm.store.GetAllBetaTesters()
	if err != nil {
		bm.logger.Error("Failed to get all beta testers", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	stats := &BetaProgramStats{
		TotalInvites:     len(testers),
		AcceptedInvites:  len(testers),
		SDKUsage:         make(map[string]int),
		IntegrationTypes: make(map[string]int),
		ExperienceLevels: make(map[string]int),
		RecentActivity:   make([]*BetaTester, 0),
	}

	var totalRating float64
	var ratingCount int

	// Get recent activity (last 10)
	recentCount := 10
	if len(testers) < recentCount {
		recentCount = len(testers)
	}
	stats.RecentActivity = testers[len(testers)-recentCount:]

	for _, tester := range testers {
		if tester.Status == "active" {
			stats.ActiveTesters++
		} else if tester.Status == "completed" {
			stats.CompletedTesters++
		}

		stats.TotalFeedback += tester.FeedbackCount
		stats.TotalBugReports += tester.BugReportCount
		stats.TotalFeatureRequests += tester.FeatureRequestCount

		if tester.OverallRating > 0 {
			totalRating += tester.OverallRating
			ratingCount++
		}

		// SDK usage
		if tester.PreferredSDK != "" {
			stats.SDKUsage[tester.PreferredSDK]++
		}

		// Integration types
		if tester.IntegrationType != "" {
			stats.IntegrationTypes[tester.IntegrationType]++
		}

		// Experience levels
		if tester.Experience != "" {
			stats.ExperienceLevels[tester.Experience]++
		}
	}

	if ratingCount > 0 {
		stats.AverageRating = totalRating / float64(ratingCount)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// HandleSendInvite handles sending an invitation to a potential beta tester
func (bm *BetaManager) HandleSendInvite(w http.ResponseWriter, r *http.Request) {
	var invite BetaTesterInvite
	if err := json.NewDecoder(r.Body).Decode(&invite); err != nil {
		bm.logger.Error("Failed to decode invite", zap.Error(err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if invite.Email == "" || invite.Name == "" {
		http.Error(w, "Missing required fields: email, name", http.StatusBadRequest)
		return
	}

	// Set default values
	invite.ID = fmt.Sprintf("invite_%d", time.Now().UnixNano())
	invite.InvitedAt = time.Now()
	invite.Status = "pending"
	invite.ExpiresAt = time.Now().Add(7 * 24 * time.Hour) // 7 days

	// TODO: Send actual email invitation
	bm.logger.Info("Beta tester invitation sent",
		zap.String("id", invite.ID),
		zap.String("email", invite.Email),
		zap.String("name", invite.Name))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Invitation sent successfully",
		"id":         invite.ID,
		"email":      invite.Email,
		"expires_at": invite.ExpiresAt,
	})
}

// SetupRoutes sets up the beta management routes
func (bm *BetaManager) SetupRoutes(router *mux.Router) {
	api := router.PathPrefix("/api/v1/beta").Subrouter()

	// Beta tester management routes
	api.HandleFunc("/testers", bm.HandleCreateBetaTester).Methods("POST")
	api.HandleFunc("/testers/{testerID}", bm.HandleGetBetaTester).Methods("GET")
	api.HandleFunc("/testers", bm.HandleGetAllBetaTesters).Methods("GET")
	api.HandleFunc("/testers/{testerID}", bm.HandleUpdateBetaTester).Methods("PUT")
	api.HandleFunc("/testers/{testerID}", bm.HandleDeleteBetaTester).Methods("DELETE")

	// Beta program management routes
	api.HandleFunc("/stats", bm.HandleGetBetaProgramStats).Methods("GET")
	api.HandleFunc("/invites", bm.HandleSendInvite).Methods("POST")
}

// generateAPIKey generates a simple API key for beta testers
func generateAPIKey() string {
	return fmt.Sprintf("beta_%d_%s", time.Now().Unix(), randomString(16))
}

// randomString generates a random string of specified length
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

	// Create beta manager
	manager := NewBetaManager(logger)

	// Setup router
	router := mux.NewRouter()
	manager.SetupRoutes(router)

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"service":   "beta-manager",
			"timestamp": time.Now(),
		})
	}).Methods("GET")

	// Start server
	port := "8081"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	logger.Info("Starting beta manager server", zap.String("port", port))
	log.Fatal(http.ListenAndServe(":"+port, router))
}
