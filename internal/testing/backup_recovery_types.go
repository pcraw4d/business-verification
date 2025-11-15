package testing

import (
	"database/sql"
	"log"
	"time"
)

// BackupRecoveryTester handles comprehensive backup and recovery testing
type BackupRecoveryTester struct {
	db        *sql.DB
	testDB    *sql.DB
	backupDir string
	logger    *log.Logger
	config    *BackupTestConfig
}

// BackupTestResult contains results of backup testing
type BackupTestResult struct {
	TestName        string
	Success         bool
	Duration        time.Duration
	ErrorMessage    string
	DataIntegrity   bool
	RecoveryTime    time.Duration
	ValidationScore float64
}

// NewBackupRecoveryTester implementation is in backup_recovery_test.go
// This file provides the type definitions for use in non-test files

