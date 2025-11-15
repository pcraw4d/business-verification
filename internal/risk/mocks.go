package risk

import (
	"github.com/stretchr/testify/mock"
)

// MockBackupService is a mock implementation of BackupService for testing
type MockBackupService struct {
	mock.Mock
}

// MockExportService is a mock implementation of ExportService for testing
type MockExportService struct {
	mock.Mock
}

// MockExportJobManager is a mock implementation of ExportJobManager for testing
type MockExportJobManager struct {
	mock.Mock
}

// ValidateExportRequest validates an export request (mock method)
func (m *MockExportService) ValidateExportRequest(ctx interface{}) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// validateBackupRequest validates a backup request (mock method)
// Note: This is a private method in the real service, but we expose it for mocking
func (m *MockBackupService) validateBackupRequest(ctx interface{}) error {
	args := m.Called(ctx)
	return args.Error(0)
}

