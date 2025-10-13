package reporting

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// DefaultReportScheduler implements ReportScheduler
type DefaultReportScheduler struct {
	scheduledReports map[string]*ScheduledReport
	mu               sync.RWMutex
	ticker           *time.Ticker
	stopChan         chan struct{}
	logger           *zap.Logger
}

// NewDefaultReportScheduler creates a new default report scheduler
func NewDefaultReportScheduler(logger *zap.Logger) *DefaultReportScheduler {
	scheduler := &DefaultReportScheduler{
		scheduledReports: make(map[string]*ScheduledReport),
		stopChan:         make(chan struct{}),
		logger:           logger,
	}

	// Start the scheduler
	scheduler.start()

	return scheduler
}

// ScheduleReport schedules a report for execution
func (s *DefaultReportScheduler) ScheduleReport(ctx context.Context, scheduledReport *ScheduledReport) error {
	s.logger.Info("Scheduling report",
		zap.String("scheduled_report_id", scheduledReport.ID),
		zap.String("name", scheduledReport.Name),
		zap.String("frequency", string(scheduledReport.Schedule.Frequency)))

	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate schedule
	if err := s.validateSchedule(scheduledReport.Schedule); err != nil {
		return fmt.Errorf("invalid schedule: %w", err)
	}

	// Calculate next run time
	nextRunAt := s.calculateNextRunTime(scheduledReport.Schedule)
	scheduledReport.NextRunAt = &nextRunAt

	// Store the scheduled report
	s.scheduledReports[scheduledReport.ID] = scheduledReport

	s.logger.Info("Report scheduled successfully",
		zap.String("scheduled_report_id", scheduledReport.ID),
		zap.Time("next_run_at", *scheduledReport.NextRunAt))

	return nil
}

// UnscheduleReport removes a scheduled report
func (s *DefaultReportScheduler) UnscheduleReport(ctx context.Context, scheduledReportID string) error {
	s.logger.Info("Unscheduling report",
		zap.String("scheduled_report_id", scheduledReportID))

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.scheduledReports[scheduledReportID]; !exists {
		return fmt.Errorf("scheduled report not found: %s", scheduledReportID)
	}

	delete(s.scheduledReports, scheduledReportID)

	s.logger.Info("Report unscheduled successfully",
		zap.String("scheduled_report_id", scheduledReportID))

	return nil
}

// GetScheduledReports returns all scheduled reports
func (s *DefaultReportScheduler) GetScheduledReports(ctx context.Context) ([]*ScheduledReport, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	reports := make([]*ScheduledReport, 0, len(s.scheduledReports))
	for _, report := range s.scheduledReports {
		reports = append(reports, report)
	}

	return reports, nil
}

// RunScheduledReports checks and runs scheduled reports
func (s *DefaultReportScheduler) RunScheduledReports(ctx context.Context) error {
	s.logger.Debug("Checking scheduled reports")

	s.mu.RLock()
	reports := make([]*ScheduledReport, 0, len(s.scheduledReports))
	for _, report := range s.scheduledReports {
		if report.IsActive && report.NextRunAt != nil && time.Now().After(*report.NextRunAt) {
			reports = append(reports, report)
		}
	}
	s.mu.RUnlock()

	// Run reports that are due
	for _, report := range reports {
		if err := s.runScheduledReport(ctx, report); err != nil {
			s.logger.Error("Failed to run scheduled report",
				zap.String("scheduled_report_id", report.ID),
				zap.Error(err))
		}
	}

	return nil
}

// start starts the scheduler background process
func (s *DefaultReportScheduler) start() {
	s.ticker = time.NewTicker(1 * time.Minute) // Check every minute

	go func() {
		for {
			select {
			case <-s.ticker.C:
				ctx := context.Background()
				if err := s.RunScheduledReports(ctx); err != nil {
					s.logger.Error("Error running scheduled reports", zap.Error(err))
				}
			case <-s.stopChan:
				s.ticker.Stop()
				return
			}
		}
	}()

	s.logger.Info("Report scheduler started")
}

// Stop stops the scheduler
func (s *DefaultReportScheduler) Stop() {
	s.logger.Info("Stopping report scheduler")
	close(s.stopChan)
}

// runScheduledReport runs a single scheduled report
func (s *DefaultReportScheduler) runScheduledReport(ctx context.Context, report *ScheduledReport) error {
	s.logger.Info("Running scheduled report",
		zap.String("scheduled_report_id", report.ID),
		zap.String("name", report.Name))

	// Update last run time
	now := time.Now()
	report.LastRunAt = &now

	// Calculate next run time
	nextRunAt := s.calculateNextRunTime(report.Schedule)
	report.NextRunAt = &nextRunAt

	// Update the scheduled report
	s.mu.Lock()
	s.scheduledReports[report.ID] = report
	s.mu.Unlock()

	// In a real implementation, you would trigger the actual report generation
	// This could be done by:
	// 1. Sending a message to a queue
	// 2. Making an HTTP call to the report service
	// 3. Using a job scheduler like cron

	s.logger.Info("Scheduled report executed",
		zap.String("scheduled_report_id", report.ID),
		zap.Time("next_run_at", *report.NextRunAt))

	return nil
}

// validateSchedule validates a report schedule
func (s *DefaultReportScheduler) validateSchedule(schedule ReportSchedule) error {
	// Validate frequency
	validFrequencies := map[ScheduleFrequency]bool{
		ScheduleFrequencyOnce:      true,
		ScheduleFrequencyDaily:     true,
		ScheduleFrequencyWeekly:    true,
		ScheduleFrequencyMonthly:   true,
		ScheduleFrequencyQuarterly: true,
		ScheduleFrequencyYearly:    true,
	}

	if !validFrequencies[schedule.Frequency] {
		return fmt.Errorf("invalid frequency: %s", schedule.Frequency)
	}

	// Validate time of day format
	if schedule.TimeOfDay != "" {
		if _, err := time.Parse("15:04", schedule.TimeOfDay); err != nil {
			return fmt.Errorf("invalid time of day format: %s", schedule.TimeOfDay)
		}
	}

	// Validate days of week (0-6)
	for _, day := range schedule.DaysOfWeek {
		if day < 0 || day > 6 {
			return fmt.Errorf("invalid day of week: %d (must be 0-6)", day)
		}
	}

	// Validate days of month (1-31)
	for _, day := range schedule.DaysOfMonth {
		if day < 1 || day > 31 {
			return fmt.Errorf("invalid day of month: %d (must be 1-31)", day)
		}
	}

	// Validate interval
	if schedule.Interval < 1 {
		return fmt.Errorf("interval must be greater than 0")
	}

	return nil
}

// calculateNextRunTime calculates the next run time for a schedule
func (s *DefaultReportScheduler) calculateNextRunTime(schedule ReportSchedule) time.Time {
	now := time.Now()

	// Parse time of day
	var hour, minute int
	if schedule.TimeOfDay != "" {
		if t, err := time.Parse("15:04", schedule.TimeOfDay); err == nil {
			hour, minute = t.Hour(), t.Minute()
		}
	}

	// Calculate next run time based on frequency
	switch schedule.Frequency {
	case ScheduleFrequencyOnce:
		// For one-time schedules, return a time far in the future
		return now.AddDate(100, 0, 0)

	case ScheduleFrequencyDaily:
		next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
		if next.Before(now) {
			next = next.AddDate(0, 0, schedule.Interval)
		}
		return next

	case ScheduleFrequencyWeekly:
		// Find the next occurrence of the specified days of week
		next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())

		// If no days specified, use current day
		if len(schedule.DaysOfWeek) == 0 {
			schedule.DaysOfWeek = []int{int(now.Weekday())}
		}

		// Find the next valid day
		for i := 0; i < 7*schedule.Interval; i++ {
			day := int(next.Weekday())
			for _, validDay := range schedule.DaysOfWeek {
				if day == validDay && next.After(now) {
					return next
				}
			}
			next = next.AddDate(0, 0, 1)
		}
		return next

	case ScheduleFrequencyMonthly:
		// Find the next occurrence of the specified days of month
		next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())

		// If no days specified, use current day
		if len(schedule.DaysOfMonth) == 0 {
			schedule.DaysOfMonth = []int{now.Day()}
		}

		// Find the next valid day
		for i := 0; i < 12*schedule.Interval; i++ {
			day := next.Day()
			for _, validDay := range schedule.DaysOfMonth {
				if day == validDay && next.After(now) {
					return next
				}
			}
			next = next.AddDate(0, 1, 0)
		}
		return next

	case ScheduleFrequencyQuarterly:
		// Run every 3 months
		next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
		next = next.AddDate(0, 3*schedule.Interval, 0)
		if next.Before(now) {
			next = next.AddDate(0, 3*schedule.Interval, 0)
		}
		return next

	case ScheduleFrequencyYearly:
		// Run every year
		next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
		next = next.AddDate(schedule.Interval, 0, 0)
		if next.Before(now) {
			next = next.AddDate(schedule.Interval, 0, 0)
		}
		return next

	default:
		// Default to daily
		next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
		if next.Before(now) {
			next = next.AddDate(0, 0, 1)
		}
		return next
	}
}
