package industry_codes

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"
)

func TestDebugCategorizer(t *testing.T) {
	logger := zap.NewNop()
	categorizer := NewErrorCategorizer(logger, nil)

	err := errors.New("connection timeout")
	errorContext := map[string]interface{}{
		"source":    "api_client",
		"operation": "fetch_data",
	}

	result := categorizer.CategorizeError(context.Background(), err, errorContext)

	if result == nil {
		t.Fatal("Categorization returned nil")
	}

	t.Logf("Categorization result: %+v", result)
	t.Logf("ID: %s", result.ID)
	t.Logf("Category: %s", result.Category)
	t.Logf("Severity: %s", result.Severity)
}
