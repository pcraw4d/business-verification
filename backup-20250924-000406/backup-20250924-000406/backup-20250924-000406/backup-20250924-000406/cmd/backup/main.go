package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/database"
	"github.com/pcraw4d/business-verification/internal/database/backup"
)

func main() {
	var (
		outputDir       = flag.String("output", "./backups", "Output directory for backups")
		retentionDays   = flag.Int("retention", 30, "Number of days to retain backups")
		compress        = flag.Bool("compress", false, "Compress backup files")
		verifyIntegrity = flag.Bool("verify", true, "Verify backup integrity")
		timeout         = flag.Duration("timeout", 30*time.Minute, "Backup operation timeout")
		cleanup         = flag.Bool("cleanup", false, "Clean up old backups")
		list            = flag.Bool("list", false, "List available backups")
		help            = flag.Bool("help", false, "Show help message")
	)

	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Initialize logger
	logger := log.New(os.Stdout, "[backup] ", log.LstdFlags)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("❌ Failed to load configuration: %v", err)
	}

	// Validate Supabase configuration
	if cfg.Supabase.URL == "" || cfg.Supabase.APIKey == "" || cfg.Supabase.ServiceRoleKey == "" {
		logger.Fatalf("❌ Supabase configuration incomplete. Required: SUPABASE_URL, SUPABASE_API_KEY, SUPABASE_SERVICE_ROLE_KEY")
	}

	// Create Supabase client
	supabaseConfig := &database.SupabaseConfig{
		URL:            cfg.Supabase.URL,
		APIKey:         cfg.Supabase.APIKey,
		ServiceRoleKey: cfg.Supabase.ServiceRoleKey,
		JWTSecret:      cfg.Supabase.JWTSecret,
	}

	supabaseClient, err := database.NewSupabaseClient(supabaseConfig, logger)
	if err != nil {
		logger.Fatalf("❌ Failed to create Supabase client: %v", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := supabaseClient.Connect(ctx); err != nil {
		logger.Fatalf("❌ Failed to connect to Supabase: %v", err)
	}

	logger.Println("✅ Successfully connected to Supabase")

	// Create backup configuration
	backupConfig := &backup.BackupConfig{
		OutputDir:       *outputDir,
		RetentionDays:   *retentionDays,
		CompressBackup:  *compress,
		VerifyIntegrity: *verifyIntegrity,
		Timeout:         *timeout,
	}

	// Create backup manager
	backupManager := backup.NewSupabaseBackupManager(supabaseClient, backupConfig, logger)

	// Handle different operations
	if *list {
		listBackups(backupManager, logger)
		return
	}

	if *cleanup {
		cleanupBackups(backupManager, logger)
		return
	}

	// Create backup
	createBackup(backupManager, logger)
}

func createBackup(backupManager *backup.SupabaseBackupManager, logger *log.Logger) {
	logger.Println("🚀 Starting Supabase database backup...")
	logger.Printf("📁 Output directory: %s", backupManager.GetBackupStatus().Configuration.OutputDir)
	logger.Printf("⏱️ Timeout: %v", backupManager.GetBackupStatus().Configuration.Timeout)
	logger.Printf("🔍 Verify integrity: %t", backupManager.GetBackupStatus().Configuration.VerifyIntegrity)

	// Create backup with timeout
	ctx, cancel := context.WithTimeout(context.Background(), backupManager.GetBackupStatus().Configuration.Timeout)
	defer cancel()

	metadata, err := backupManager.CreateFullBackup(ctx)
	if err != nil {
		logger.Fatalf("❌ Backup failed: %v", err)
	}

	// Print backup summary
	printBackupSummary(metadata, logger)

	// Cleanup old backups
	if err := backupManager.CleanupOldBackups(); err != nil {
		logger.Printf("⚠️ Warning: Failed to cleanup old backups: %v", err)
	}
}

func listBackups(backupManager *backup.SupabaseBackupManager, logger *log.Logger) {
	logger.Println("📋 Listing available backups...")

	backups, err := backupManager.ListBackups()
	if err != nil {
		logger.Fatalf("❌ Failed to list backups: %v", err)
	}

	if len(backups) == 0 {
		logger.Println("📭 No backups found")
		return
	}

	fmt.Printf("\n%-20s %-20s %-10s %-15s %-10s\n", "Backup ID", "Timestamp", "Status", "Records", "Size")
	fmt.Println(strings.Repeat("-", 80))

	for _, backup := range backups {
		sizeStr := formatBytes(backup.BackupSize)
		fmt.Printf("%-20s %-20s %-10s %-15d %-10s\n",
			backup.BackupID,
			backup.Timestamp.Format("2006-01-02 15:04:05"),
			backup.Status,
			backup.TotalRecords,
			sizeStr,
		)
	}

	fmt.Printf("\nTotal backups: %d\n", len(backups))
}

func cleanupBackups(backupManager *backup.SupabaseBackupManager, logger *log.Logger) {
	logger.Println("🧹 Cleaning up old backups...")

	if err := backupManager.CleanupOldBackups(); err != nil {
		logger.Fatalf("❌ Cleanup failed: %v", err)
	}

	logger.Println("✅ Cleanup completed successfully")
}

func printBackupSummary(metadata *backup.BackupMetadata, logger *log.Logger) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎉 BACKUP COMPLETED SUCCESSFULLY")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("📋 Backup ID: %s\n", metadata.BackupID)
	fmt.Printf("⏰ Timestamp: %s\n", metadata.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("🌐 Database: %s\n", metadata.DatabaseURL)
	fmt.Printf("📊 Total Records: %d\n", metadata.TotalRecords)
	fmt.Printf("💾 Backup Size: %s\n", formatBytes(metadata.BackupSize))
	fmt.Printf("🔍 Checksum: %s\n", metadata.Checksum)
	fmt.Printf("📁 Output Directory: %s\n", filepath.Join(metadata.Configuration.OutputDir, metadata.BackupID))
	fmt.Printf("🏷️ Environment: %s\n", metadata.Environment)
	fmt.Printf("📋 Tables Backed Up: %d\n", len(metadata.Tables))

	if len(metadata.Tables) > 0 {
		fmt.Println("\n📋 Table Details:")
		fmt.Printf("%-30s %-10s %-15s\n", "Table Name", "Records", "Size")
		fmt.Println(strings.Repeat("-", 55))
		for _, table := range metadata.Tables {
			fmt.Printf("%-30s %-10d %-15s\n", table.Name, table.Records, formatBytes(table.Size))
		}
	}

	fmt.Println(strings.Repeat("=", 60))
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func showHelp() {
	fmt.Println("🗄️ Supabase Database Backup Tool")
	fmt.Println("==================================")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  go run cmd/backup/main.go [OPTIONS]")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  -output string")
	fmt.Println("        Output directory for backups (default: ./backups)")
	fmt.Println("  -retention int")
	fmt.Println("        Number of days to retain backups (default: 30)")
	fmt.Println("  -compress")
	fmt.Println("        Compress backup files (default: false)")
	fmt.Println("  -verify")
	fmt.Println("        Verify backup integrity (default: true)")
	fmt.Println("  -timeout duration")
	fmt.Println("        Backup operation timeout (default: 30m)")
	fmt.Println("  -cleanup")
	fmt.Println("        Clean up old backups")
	fmt.Println("  -list")
	fmt.Println("        List available backups")
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Create a backup")
	fmt.Println("  go run cmd/backup/main.go")
	fmt.Println()
	fmt.Println("  # Create a backup with custom settings")
	fmt.Println("  go run cmd/backup/main.go -output /path/to/backups -retention 7 -compress")
	fmt.Println()
	fmt.Println("  # List available backups")
	fmt.Println("  go run cmd/backup/main.go -list")
	fmt.Println()
	fmt.Println("  # Clean up old backups")
	fmt.Println("  go run cmd/backup/main.go -cleanup")
	fmt.Println()
	fmt.Println("ENVIRONMENT VARIABLES:")
	fmt.Println("  SUPABASE_URL              Supabase project URL")
	fmt.Println("  SUPABASE_API_KEY          Supabase API key")
	fmt.Println("  SUPABASE_SERVICE_ROLE_KEY Supabase service role key")
	fmt.Println("  SUPABASE_JWT_SECRET       Supabase JWT secret")
	fmt.Println()
}
