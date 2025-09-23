package keyword_classification

import (
	"testing"

	"kyb-platform/internal/architecture"

	"github.com/stretchr/testify/assert"
)

func TestNewKeywordClassificationModule(t *testing.T) {
	module := NewKeywordClassificationModule()

	assert.NotNil(t, module)
	assert.Equal(t, "keyword_classification_module", module.ID())
	assert.False(t, module.IsRunning())
}

func TestKeywordClassificationModuleMetadata(t *testing.T) {
	module := NewKeywordClassificationModule()
	metadata := module.Metadata()

	assert.Equal(t, "Keyword Classification Module", metadata.Name)
	assert.Equal(t, "1.0.0", metadata.Version)
	assert.Equal(t, "Performs business classification using keyword analysis", metadata.Description)
	assert.Contains(t, metadata.Capabilities, architecture.CapabilityClassification)
	assert.Equal(t, architecture.PriorityHigh, metadata.Priority)
}

func TestKeywordClassificationModuleCanHandle(t *testing.T) {
	module := NewKeywordClassificationModule()

	// Test supported request type
	req := architecture.ModuleRequest{
		Type: "classify_by_keywords",
	}
	assert.True(t, module.CanHandle(req))

	// Test unsupported request type
	req.Type = "unsupported_type"
	assert.False(t, module.CanHandle(req))
}

func TestKeywordClassificationModuleHealth(t *testing.T) {
	module := NewKeywordClassificationModule()

	// Health when not running
	health := module.Health()
	assert.Equal(t, architecture.ModuleStatusStopped, health.Status)
	assert.Contains(t, health.Message, "Keyword classification module")

	// Health when running
	module.running = true
	health = module.Health()
	assert.Equal(t, architecture.ModuleStatusRunning, health.Status)
}
