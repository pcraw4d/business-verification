package feedback

import (
	"context"
	"testing"
	"time"
)

func TestFeedbackValidator_ValidateUserFeedback(t *testing.T) {
	validator := NewFeedbackValidator(&MockSecurityValidator{})

	tests := []struct {
		name     string
		feedback UserFeedback
		wantErr  bool
	}{
		{
			name: "valid user feedback",
			feedback: UserFeedback{
				ID:                       "feedback123",
				UserID:                   "user123",
				BusinessName:             "Test Business",
				OriginalClassificationID: "classification123",
				FeedbackType:             FeedbackTypeAccuracy,
				FeedbackValue:            map[string]interface{}{"accuracy": 0.9},
				FeedbackText:             "This classification was accurate",
				ConfidenceScore:          0.9,
				Status:                   FeedbackStatusPending,
				ProcessingTimeMs:         100,
				CreatedAt:                time.Now(),
				Metadata:                 map[string]interface{}{"source": "web"},
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			feedback: UserFeedback{
				UserID:                   "user123",
				BusinessName:             "Test Business",
				OriginalClassificationID: "classification123",
				FeedbackType:             FeedbackTypeAccuracy,
			},
			wantErr: true,
		},
		{
			name: "missing user ID",
			feedback: UserFeedback{
				ID:                       "feedback123",
				BusinessName:             "Test Business",
				OriginalClassificationID: "classification123",
				FeedbackType:             FeedbackTypeAccuracy,
			},
			wantErr: true,
		},
		{
			name: "missing business name",
			feedback: UserFeedback{
				ID:                       "feedback123",
				UserID:                   "user123",
				OriginalClassificationID: "classification123",
				FeedbackType:             FeedbackTypeAccuracy,
			},
			wantErr: true,
		},
		{
			name: "missing original classification ID",
			feedback: UserFeedback{
				ID:           "feedback123",
				UserID:       "user123",
				BusinessName: "Test Business",
				FeedbackType: FeedbackTypeAccuracy,
			},
			wantErr: true,
		},
		{
			name: "invalid feedback type",
			feedback: UserFeedback{
				ID:                       "feedback123",
				UserID:                   "user123",
				BusinessName:             "Test Business",
				OriginalClassificationID: "classification123",
				FeedbackType:             "invalid_type",
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			feedback: UserFeedback{
				ID:                       "feedback123",
				UserID:                   "user123",
				BusinessName:             "Test Business",
				OriginalClassificationID: "classification123",
				FeedbackType:             FeedbackTypeAccuracy,
				Status:                   "invalid_status",
			},
			wantErr: true,
		},
		{
			name: "invalid confidence score - too high",
			feedback: UserFeedback{
				ID:                       "feedback123",
				UserID:                   "user123",
				BusinessName:             "Test Business",
				OriginalClassificationID: "classification123",
				FeedbackType:             FeedbackTypeAccuracy,
				ConfidenceScore:          1.5,
			},
			wantErr: true,
		},
		{
			name: "invalid confidence score - negative",
			feedback: UserFeedback{
				ID:                       "feedback123",
				UserID:                   "user123",
				BusinessName:             "Test Business",
				OriginalClassificationID: "classification123",
				FeedbackType:             FeedbackTypeAccuracy,
				ConfidenceScore:          -0.1,
			},
			wantErr: true,
		},
		{
			name: "invalid business name - too long",
			feedback: UserFeedback{
				ID:                       "feedback123",
				UserID:                   "user123",
				BusinessName:             string(make([]byte, 256)), // 256 characters
				OriginalClassificationID: "classification123",
				FeedbackType:             FeedbackTypeAccuracy,
			},
			wantErr: true,
		},
		{
			name: "invalid business name - empty",
			feedback: UserFeedback{
				ID:                       "feedback123",
				UserID:                   "user123",
				BusinessName:             "",
				OriginalClassificationID: "classification123",
				FeedbackType:             FeedbackTypeAccuracy,
			},
			wantErr: true,
		},
		{
			name: "invalid business name - only whitespace",
			feedback: UserFeedback{
				ID:                       "feedback123",
				UserID:                   "user123",
				BusinessName:             "   ",
				OriginalClassificationID: "classification123",
				FeedbackType:             FeedbackTypeAccuracy,
			},
			wantErr: true,
		},
		{
			name: "invalid business name - invalid characters",
			feedback: UserFeedback{
				ID:                       "feedback123",
				UserID:                   "user123",
				BusinessName:             "Test@Business#123",
				OriginalClassificationID: "classification123",
				FeedbackType:             FeedbackTypeAccuracy,
			},
			wantErr: true,
		},
		{
			name: "feedback text too long",
			feedback: UserFeedback{
				ID:                       "feedback123",
				UserID:                   "user123",
				BusinessName:             "Test Business",
				OriginalClassificationID: "classification123",
				FeedbackType:             FeedbackTypeAccuracy,
				FeedbackText:             string(make([]byte, 5001)), // 5001 characters
			},
			wantErr: true,
		},
		{
			name: "negative processing time",
			feedback: UserFeedback{
				ID:                       "feedback123",
				UserID:                   "user123",
				BusinessName:             "Test Business",
				OriginalClassificationID: "classification123",
				FeedbackType:             FeedbackTypeAccuracy,
				ProcessingTimeMs:         -1,
			},
			wantErr: true,
		},
		{
			name: "processed at before created at",
			feedback: UserFeedback{
				ID:                       "feedback123",
				UserID:                   "user123",
				BusinessName:             "Test Business",
				OriginalClassificationID: "classification123",
				FeedbackType:             FeedbackTypeAccuracy,
				CreatedAt:                time.Now(),
				ProcessedAt:              &time.Time{}, // Zero time, which is before CreatedAt
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := validator.ValidateUserFeedback(ctx, tt.feedback)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateUserFeedback() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateUserFeedback() unexpected error: %v", err)
			}
		})
	}
}

func TestFeedbackValidator_ValidateMLModelFeedback(t *testing.T) {
	validator := NewFeedbackValidator(&MockSecurityValidator{})

	tests := []struct {
		name     string
		feedback MLModelFeedback
		wantErr  bool
	}{
		{
			name: "valid ML model feedback",
			feedback: MLModelFeedback{
				ID:                   "feedback123",
				ModelVersionID:       "model_v1.0.0",
				ModelType:            "bert",
				ClassificationMethod: MethodML,
				PredictionID:         "prediction123",
				ActualResult:         map[string]interface{}{"industry": "technology"},
				PredictedResult:      map[string]interface{}{"industry": "technology"},
				AccuracyScore:        0.95,
				ConfidenceScore:      0.90,
				ProcessingTimeMs:     100,
				Status:               FeedbackStatusPending,
				CreatedAt:            time.Now(),
				Metadata:             map[string]interface{}{"model_version": "v1.0.0"},
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			feedback: MLModelFeedback{
				ModelVersionID:       "model_v1.0.0",
				ModelType:            "bert",
				ClassificationMethod: MethodML,
				PredictionID:         "prediction123",
			},
			wantErr: true,
		},
		{
			name: "missing model version ID",
			feedback: MLModelFeedback{
				ID:                   "feedback123",
				ModelType:            "bert",
				ClassificationMethod: MethodML,
				PredictionID:         "prediction123",
			},
			wantErr: true,
		},
		{
			name: "missing model type",
			feedback: MLModelFeedback{
				ID:                   "feedback123",
				ModelVersionID:       "model_v1.0.0",
				ClassificationMethod: MethodML,
				PredictionID:         "prediction123",
			},
			wantErr: true,
		},
		{
			name: "missing classification method",
			feedback: MLModelFeedback{
				ID:             "feedback123",
				ModelVersionID: "model_v1.0.0",
				ModelType:      "bert",
				PredictionID:   "prediction123",
			},
			wantErr: true,
		},
		{
			name: "missing prediction ID",
			feedback: MLModelFeedback{
				ID:                   "feedback123",
				ModelVersionID:       "model_v1.0.0",
				ModelType:            "bert",
				ClassificationMethod: MethodML,
			},
			wantErr: true,
		},
		{
			name: "invalid classification method",
			feedback: MLModelFeedback{
				ID:                   "feedback123",
				ModelVersionID:       "model_v1.0.0",
				ModelType:            "bert",
				ClassificationMethod: "invalid_method",
				PredictionID:         "prediction123",
			},
			wantErr: true,
		},
		{
			name: "invalid accuracy score - too high",
			feedback: MLModelFeedback{
				ID:                   "feedback123",
				ModelVersionID:       "model_v1.0.0",
				ModelType:            "bert",
				ClassificationMethod: MethodML,
				PredictionID:         "prediction123",
				AccuracyScore:        1.5,
			},
			wantErr: true,
		},
		{
			name: "invalid accuracy score - negative",
			feedback: MLModelFeedback{
				ID:                   "feedback123",
				ModelVersionID:       "model_v1.0.0",
				ModelType:            "bert",
				ClassificationMethod: MethodML,
				PredictionID:         "prediction123",
				AccuracyScore:        -0.1,
			},
			wantErr: true,
		},
		{
			name: "invalid confidence score - too high",
			feedback: MLModelFeedback{
				ID:                   "feedback123",
				ModelVersionID:       "model_v1.0.0",
				ModelType:            "bert",
				ClassificationMethod: MethodML,
				PredictionID:         "prediction123",
				ConfidenceScore:      1.5,
			},
			wantErr: true,
		},
		{
			name: "invalid confidence score - negative",
			feedback: MLModelFeedback{
				ID:                   "feedback123",
				ModelVersionID:       "model_v1.0.0",
				ModelType:            "bert",
				ClassificationMethod: MethodML,
				PredictionID:         "prediction123",
				ConfidenceScore:      -0.1,
			},
			wantErr: true,
		},
		{
			name: "empty actual result",
			feedback: MLModelFeedback{
				ID:                   "feedback123",
				ModelVersionID:       "model_v1.0.0",
				ModelType:            "bert",
				ClassificationMethod: MethodML,
				PredictionID:         "prediction123",
				ActualResult:         map[string]interface{}{},
			},
			wantErr: true,
		},
		{
			name: "empty predicted result",
			feedback: MLModelFeedback{
				ID:                   "feedback123",
				ModelVersionID:       "model_v1.0.0",
				ModelType:            "bert",
				ClassificationMethod: MethodML,
				PredictionID:         "prediction123",
				ActualResult:         map[string]interface{}{"industry": "technology"},
				PredictedResult:      map[string]interface{}{},
			},
			wantErr: true,
		},
		{
			name: "negative processing time",
			feedback: MLModelFeedback{
				ID:                   "feedback123",
				ModelVersionID:       "model_v1.0.0",
				ModelType:            "bert",
				ClassificationMethod: MethodML,
				PredictionID:         "prediction123",
				ActualResult:         map[string]interface{}{"industry": "technology"},
				PredictedResult:      map[string]interface{}{"industry": "technology"},
				ProcessingTimeMs:     -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := validator.ValidateMLModelFeedback(ctx, tt.feedback)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateMLModelFeedback() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateMLModelFeedback() unexpected error: %v", err)
			}
		})
	}
}

func TestFeedbackValidator_ValidateSecurityValidationFeedback(t *testing.T) {
	validator := NewFeedbackValidator(&MockSecurityValidator{})

	tests := []struct {
		name     string
		feedback SecurityValidationFeedback
		wantErr  bool
	}{
		{
			name: "valid security validation feedback",
			feedback: SecurityValidationFeedback{
				ID:                 "feedback123",
				ValidationType:     "website_verification",
				DataSourceType:     "domain_analysis",
				WebsiteURL:         "https://example.com",
				ValidationResult:   map[string]interface{}{"verified": true},
				TrustScore:         0.95,
				VerificationStatus: "verified",
				SecurityViolations: []string{},
				ProcessingTimeMs:   50,
				Status:             FeedbackStatusPending,
				CreatedAt:          time.Now(),
				Metadata:           map[string]interface{}{"source": "whois"},
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			feedback: SecurityValidationFeedback{
				ValidationType:   "website_verification",
				DataSourceType:   "domain_analysis",
				ValidationResult: map[string]interface{}{"verified": true},
			},
			wantErr: true,
		},
		{
			name: "missing validation type",
			feedback: SecurityValidationFeedback{
				ID:               "feedback123",
				DataSourceType:   "domain_analysis",
				ValidationResult: map[string]interface{}{"verified": true},
			},
			wantErr: true,
		},
		{
			name: "missing data source type",
			feedback: SecurityValidationFeedback{
				ID:               "feedback123",
				ValidationType:   "website_verification",
				ValidationResult: map[string]interface{}{"verified": true},
			},
			wantErr: true,
		},
		{
			name: "missing verification status",
			feedback: SecurityValidationFeedback{
				ID:               "feedback123",
				ValidationType:   "website_verification",
				DataSourceType:   "domain_analysis",
				ValidationResult: map[string]interface{}{"verified": true},
			},
			wantErr: true,
		},
		{
			name: "invalid trust score - too high",
			feedback: SecurityValidationFeedback{
				ID:               "feedback123",
				ValidationType:   "website_verification",
				DataSourceType:   "domain_analysis",
				ValidationResult: map[string]interface{}{"verified": true},
				TrustScore:       1.5,
			},
			wantErr: true,
		},
		{
			name: "invalid trust score - negative",
			feedback: SecurityValidationFeedback{
				ID:               "feedback123",
				ValidationType:   "website_verification",
				DataSourceType:   "domain_analysis",
				ValidationResult: map[string]interface{}{"verified": true},
				TrustScore:       -0.1,
			},
			wantErr: true,
		},
		{
			name: "invalid website URL",
			feedback: SecurityValidationFeedback{
				ID:               "feedback123",
				ValidationType:   "website_verification",
				DataSourceType:   "domain_analysis",
				WebsiteURL:       "invalid-url",
				ValidationResult: map[string]interface{}{"verified": true},
			},
			wantErr: true,
		},
		{
			name: "website URL too long",
			feedback: SecurityValidationFeedback{
				ID:               "feedback123",
				ValidationType:   "website_verification",
				DataSourceType:   "domain_analysis",
				WebsiteURL:       "https://" + string(make([]byte, 1000)) + ".com",
				ValidationResult: map[string]interface{}{"verified": true},
			},
			wantErr: true,
		},
		{
			name: "empty validation result",
			feedback: SecurityValidationFeedback{
				ID:               "feedback123",
				ValidationType:   "website_verification",
				DataSourceType:   "domain_analysis",
				ValidationResult: map[string]interface{}{},
			},
			wantErr: true,
		},
		{
			name: "negative processing time",
			feedback: SecurityValidationFeedback{
				ID:               "feedback123",
				ValidationType:   "website_verification",
				DataSourceType:   "domain_analysis",
				ValidationResult: map[string]interface{}{"verified": true},
				ProcessingTimeMs: -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := validator.ValidateSecurityValidationFeedback(ctx, tt.feedback)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateSecurityValidationFeedback() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateSecurityValidationFeedback() unexpected error: %v", err)
			}
		})
	}
}

func TestFeedbackValidator_ValidateFeedbackCollectionRequest(t *testing.T) {
	validator := NewFeedbackValidator(&MockSecurityValidator{})

	tests := []struct {
		name    string
		request FeedbackCollectionRequest
		wantErr bool
	}{
		{
			name: "valid feedback collection request",
			request: FeedbackCollectionRequest{
				UserID:                   "user123",
				BusinessName:             "Test Business",
				OriginalClassificationID: "classification123",
				FeedbackType:             FeedbackTypeAccuracy,
				FeedbackValue:            map[string]interface{}{"accuracy": 0.9},
				FeedbackText:             "This classification was accurate",
				ConfidenceScore:          0.9,
				Metadata:                 map[string]interface{}{"source": "web"},
			},
			wantErr: false,
		},
		{
			name: "missing user ID",
			request: FeedbackCollectionRequest{
				BusinessName:             "Test Business",
				OriginalClassificationID: "classification123",
				FeedbackType:             FeedbackTypeAccuracy,
			},
			wantErr: true,
		},
		{
			name: "missing business name",
			request: FeedbackCollectionRequest{
				UserID:                   "user123",
				OriginalClassificationID: "classification123",
				FeedbackType:             FeedbackTypeAccuracy,
			},
			wantErr: true,
		},
		{
			name: "missing original classification ID",
			request: FeedbackCollectionRequest{
				UserID:       "user123",
				BusinessName: "Test Business",
				FeedbackType: FeedbackTypeAccuracy,
			},
			wantErr: true,
		},
		{
			name: "missing feedback type",
			request: FeedbackCollectionRequest{
				UserID:                   "user123",
				BusinessName:             "Test Business",
				OriginalClassificationID: "classification123",
			},
			wantErr: true,
		},
		{
			name: "invalid confidence score - too high",
			request: FeedbackCollectionRequest{
				UserID:                   "user123",
				BusinessName:             "Test Business",
				OriginalClassificationID: "classification123",
				FeedbackType:             FeedbackTypeAccuracy,
				ConfidenceScore:          1.5,
			},
			wantErr: true,
		},
		{
			name: "invalid confidence score - negative",
			request: FeedbackCollectionRequest{
				UserID:                   "user123",
				BusinessName:             "Test Business",
				OriginalClassificationID: "classification123",
				FeedbackType:             FeedbackTypeAccuracy,
				ConfidenceScore:          -0.1,
			},
			wantErr: true,
		},
		{
			name: "feedback text too long",
			request: FeedbackCollectionRequest{
				UserID:                   "user123",
				BusinessName:             "Test Business",
				OriginalClassificationID: "classification123",
				FeedbackType:             FeedbackTypeAccuracy,
				FeedbackText:             string(make([]byte, 5001)), // 5001 characters
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := validator.ValidateFeedbackCollectionRequest(ctx, tt.request)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateFeedbackCollectionRequest() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateFeedbackCollectionRequest() unexpected error: %v", err)
			}
		})
	}
}

func TestFeedbackValidator_validateBusinessName(t *testing.T) {
	validator := NewFeedbackValidator(&MockSecurityValidator{})

	tests := []struct {
		name         string
		businessName string
		wantErr      bool
	}{
		{
			name:         "valid business name",
			businessName: "Acme Corporation",
			wantErr:      false,
		},
		{
			name:         "valid business name with special characters",
			businessName: "Smith & Associates, LLC",
			wantErr:      false,
		},
		{
			name:         "valid business name with numbers",
			businessName: "Company 123 Inc.",
			wantErr:      false,
		},
		{
			name:         "empty business name",
			businessName: "",
			wantErr:      true,
		},
		{
			name:         "business name too long",
			businessName: string(make([]byte, 256)), // 256 characters
			wantErr:      true,
		},
		{
			name:         "business name only whitespace",
			businessName: "   ",
			wantErr:      true,
		},
		{
			name:         "business name with invalid characters",
			businessName: "Test@Business#123",
			wantErr:      true,
		},
		{
			name:         "business name with leading/trailing whitespace",
			businessName: "  Test Business  ",
			wantErr:      false, // Should be valid after trimming
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateBusinessName(tt.businessName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validateBusinessName() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("validateBusinessName() unexpected error: %v", err)
			}
		})
	}
}

func TestFeedbackValidator_validateWebsiteURL(t *testing.T) {
	validator := NewFeedbackValidator(&MockSecurityValidator{})

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid HTTP URL",
			url:     "http://example.com",
			wantErr: false,
		},
		{
			name:    "valid HTTPS URL",
			url:     "https://example.com",
			wantErr: false,
		},
		{
			name:    "valid URL with path",
			url:     "https://example.com/path/to/page",
			wantErr: false,
		},
		{
			name:    "valid URL with subdomain",
			url:     "https://www.example.com",
			wantErr: false,
		},
		{
			name:    "invalid URL - no protocol",
			url:     "example.com",
			wantErr: true,
		},
		{
			name:    "invalid URL - wrong protocol",
			url:     "ftp://example.com",
			wantErr: true,
		},
		{
			name:    "invalid URL - malformed",
			url:     "not-a-url",
			wantErr: true,
		},
		{
			name:    "URL too long",
			url:     "https://" + string(make([]byte, 1000)) + ".com",
			wantErr: true,
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateWebsiteURL(tt.url)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validateWebsiteURL() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("validateWebsiteURL() unexpected error: %v", err)
			}
		})
	}
}
