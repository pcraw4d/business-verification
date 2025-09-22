package ml_classification

import (
	"testing"

	"kyb-platform/internal/architecture"
	"github.com/stretchr/testify/assert"
)

func TestNewMLClassificationModule(t *testing.T) {
	module := NewMLClassificationModule()

	assert.NotNil(t, module)
	assert.Equal(t, "ml_classification_module", module.ID())
	assert.False(t, module.IsRunning())
}

func TestMLClassificationModuleMetadata(t *testing.T) {
	module := NewMLClassificationModule()
	metadata := module.Metadata()

	assert.Equal(t, "ML Classification Module", metadata.Name)
	assert.Equal(t, "1.0.0", metadata.Version)
	assert.Equal(t, "Performs business classification using machine learning models", metadata.Description)
	assert.Contains(t, metadata.Capabilities, architecture.CapabilityClassification)
	assert.Contains(t, metadata.Capabilities, architecture.CapabilityMLPrediction)
	assert.Equal(t, architecture.PriorityHigh, metadata.Priority)
}

func TestMLClassificationModuleCanHandle(t *testing.T) {
	module := NewMLClassificationModule()

	// Test supported request type
	req := architecture.ModuleRequest{
		Type: "classify_by_ml",
	}
	assert.True(t, module.CanHandle(req))

	// Test unsupported request type
	req.Type = "unsupported_type"
	assert.False(t, module.CanHandle(req))
}

func TestMLClassificationModuleHealth(t *testing.T) {
	module := NewMLClassificationModule()

	// Health when not running
	health := module.Health()
	assert.Equal(t, architecture.ModuleStatusStopped, health.Status)
	assert.Contains(t, health.Message, "ML classification module")

	// Health when running
	module.running = true
	health = module.Health()
	assert.Equal(t, architecture.ModuleStatusRunning, health.Status)
}

func TestMLClassificationModuleFactory(t *testing.T) {
	// Create factory
	factory := NewMLClassificationFactory(nil, nil, nil, nil, nil)

	assert.NotNil(t, factory)
	assert.Equal(t, "ml_classification", factory.GetModuleType())

	// Create module through factory
	moduleConfig := architecture.ModuleConfig{
		Enabled: true,
	}

	module, err := factory.CreateModule(moduleConfig)
	assert.NoError(t, err)
	assert.NotNil(t, module)
	assert.Equal(t, "ml_classification_module", module.ID())

	// Verify module implements Module interface
	_, ok := module.(architecture.Module)
	assert.True(t, ok, "Module should implement architecture.Module interface")
}
