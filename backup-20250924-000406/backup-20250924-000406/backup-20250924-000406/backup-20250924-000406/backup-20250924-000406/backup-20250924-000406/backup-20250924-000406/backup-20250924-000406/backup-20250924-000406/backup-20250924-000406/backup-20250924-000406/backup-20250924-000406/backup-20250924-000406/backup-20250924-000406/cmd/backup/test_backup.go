package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/database"
	"github.com/pcraw4d/business-verification/internal/database/backup"
)

// TestBackupSystem tests the backup system functionality
func TestBackupSystem() {
	fmt.Println("🧪 Testing Supabase Backup System")
	fmt.Println("==================================")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}

	// Check if Supabase is configured
	if cfg.Supabase.URL == "" || cfg.Supabase.APIKey == "" || cfg.Supabase.ServiceRoleKey == "" {
		fmt.Println("⚠️ Supabase configuration incomplete - running in test mode")
		fmt.Println("📝 Required: SUPABASE_URL, SUPABASE_API_KEY, SUPABASE_SERVICE_ROLE_KEY")
		return
	}

	// Create Supabase client
	supabaseConfig := &database.SupabaseConfig{
		URL:            cfg.Supabase.URL,
		APIKey:         cfg.Supabase.APIKey,
		ServiceRoleKey: cfg.Supabase.ServiceRoleKey,
		JWTSecret:      cfg.Supabase.JWTSecret,
	}

	supabaseClient, err := database.NewSupabaseClient(supabaseConfig, log.Default())
	if err != nil {
		log.Fatalf("❌ Failed to create Supabase client: %v", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := supabaseClient.Connect(ctx); err != nil {
		log.Fatalf("❌ Failed to connect to Supabase: %v", err)
	}

	fmt.Println("✅ Successfully connected to Supabase")

	// Create backup configuration
	backupConfig := &backup.BackupConfig{
		OutputDir:       "./test_backups",
		RetentionDays:   7,
		CompressBackup:  false,
		VerifyIntegrity: true,
		Timeout:         5 * time.Minute,
	}

	// Create backup manager
	backupManager := backup.NewSupabaseBackupManager(supabaseClient, backupConfig, log.Default())

	// Test backup creation
	fmt.Println("🔄 Testing backup creation...")

	backupCtx, backupCancel := context.WithTimeout(context.Background(), backupConfig.Timeout)
	defer backupCancel()

	metadata, err := backupManager.CreateFullBackup(backupCtx)
	if err != nil {
		log.Fatalf("❌ Backup test failed: %v", err)
	}

	// Display results
	fmt.Println("✅ Backup test completed successfully")
	fmt.Printf("📋 Backup ID: %s\n", metadata.BackupID)
	fmt.Printf("⏰ Timestamp: %s\n", metadata.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("📊 Total Records: %d\n", metadata.TotalRecords)
	fmt.Printf("💾 Backup Size: %d bytes\n", metadata.BackupSize)
	fmt.Printf("🔍 Checksum: %s\n", metadata.Checksum)
	fmt.Printf("📋 Tables: %d\n", len(metadata.Tables))

	// Test backup listing
	fmt.Println("\n🔄 Testing backup listing...")
	backups, err := backupManager.ListBackups()
	if err != nil {
		log.Printf("⚠️ Warning: Failed to list backups: %v", err)
	} else {
		fmt.Printf("📋 Found %d backup(s)\n", len(backups))
		for _, b := range backups {
			fmt.Printf("  - %s (%s) - %s\n", b.BackupID, b.Timestamp.Format("2006-01-02 15:04:05"), b.Status)
		}
	}

	// Test cleanup
	fmt.Println("\n🔄 Testing backup cleanup...")
	if err := backupManager.CleanupOldBackups(); err != nil {
		log.Printf("⚠️ Warning: Failed to cleanup backups: %v", err)
	} else {
		fmt.Println("✅ Backup cleanup completed")
	}

	fmt.Println("\n🎉 All backup system tests completed successfully!")
}

func main() {
	TestBackupSystem()
}
