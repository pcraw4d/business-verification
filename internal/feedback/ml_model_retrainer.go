package feedback

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

// NewMLModelRetrainer creates a new ML model retrainer
func NewMLModelRetrainer(config *AdvancedLearningConfig, logger *zap.Logger) *MLModelRetrainer {
	return &MLModelRetrainer{
		config:         config,
		logger:         logger,
		trainingData:   make([]*TrainingDataPoint, 0),
		modelVersions:  make(map[string]*ModelVersion),
		retrainingJobs: make(map[string]*RetrainingJob),
	}
}

// RetrainModels retrains ML models with new feedback data
func (mmr *MLModelRetrainer) RetrainModels(trainingData []*TrainingDataPoint) error {
	mmr.mu.Lock()
	defer mmr.mu.Unlock()

	mmr.logger.Info("Starting ML model retraining",
		zap.Int("training_data_count", len(trainingData)))

	// Validate training data
	if err := mmr.validateTrainingData(trainingData); err != nil {
		return fmt.Errorf("failed to validate training data: %w", err)
	}

	// Add new training data to existing data
	mmr.trainingData = append(mmr.trainingData, trainingData...)

	// Maintain training data size
	if len(mmr.trainingData) > mmr.config.MaxLearningBatchSize*2 {
		// Keep most recent data
		mmr.trainingData = mmr.trainingData[len(mmr.trainingData)-mmr.config.MaxLearningBatchSize:]
	}

	// Check if retraining is needed
	if !mmr.shouldRetrain() {
		mmr.logger.Info("Retraining not needed based on current criteria")
		return nil
	}

	// Start retraining jobs for each model type
	modelsToRetrain := []string{"bert_classifier", "ensemble_classifier", "uncertainty_estimator"}

	for _, modelID := range modelsToRetrain {
		if err := mmr.startRetrainingJob(modelID); err != nil {
			mmr.logger.Error("Failed to start retraining job",
				zap.String("model_id", modelID),
				zap.Error(err))
			continue
		}
	}

	return nil
}

// validateTrainingData validates training data quality
func (mmr *MLModelRetrainer) validateTrainingData(trainingData []*TrainingDataPoint) error {
	if len(trainingData) == 0 {
		return fmt.Errorf("no training data provided")
	}

	// Check data quality
	validDataCount := 0
	for _, data := range trainingData {
		if mmr.isValidTrainingDataPoint(data) {
			validDataCount++
		}
	}

	if validDataCount < int(float64(len(trainingData))*0.8) { // At least 80% valid data
		return fmt.Errorf("insufficient valid training data: %d/%d", validDataCount, len(trainingData))
	}

	// Check class distribution
	classDistribution := make(map[string]int)
	for _, data := range trainingData {
		if data.TrueClassification != "" {
			classDistribution[data.TrueClassification]++
		}
	}

	// Ensure minimum samples per class
	minSamplesPerClass := 10
	for class, count := range classDistribution {
		if count < minSamplesPerClass {
			mmr.logger.Warn("Insufficient samples for class",
				zap.String("class", class),
				zap.Int("count", count),
				zap.Int("min_required", minSamplesPerClass))
		}
	}

	return nil
}

// isValidTrainingDataPoint checks if a training data point is valid
func (mmr *MLModelRetrainer) isValidTrainingDataPoint(data *TrainingDataPoint) bool {
	// Check required fields
	if data.BusinessName == "" || data.TrueClassification == "" {
		return false
	}

	// Check feedback type
	if data.FeedbackType != FeedbackTypeAccuracy && data.FeedbackType != FeedbackTypeCorrection {
		return false
	}

	// Check confidence score range
	if data.ConfidenceScore < 0.0 || data.ConfidenceScore > 1.0 {
		return false
	}

	// Check timestamp
	if data.Timestamp.IsZero() || data.Timestamp.After(time.Now()) {
		return false
	}

	return true
}

// shouldRetrain determines if retraining is needed
func (mmr *MLModelRetrainer) shouldRetrain() bool {
	// Check if we have enough new data
	if len(mmr.trainingData) < mmr.config.MinDataForRetraining {
		return false
	}

	// Check if enough time has passed since last retraining
	lastRetraining := mmr.getLastRetrainingTime()
	if time.Since(lastRetraining) < mmr.config.RetrainingInterval {
		return false
	}

	// Check if there are active retraining jobs
	if mmr.hasActiveRetrainingJobs() {
		return false
	}

	// Check performance degradation (if available)
	if mmr.hasPerformanceDegradation() {
		mmr.logger.Info("Retraining triggered due to performance degradation")
		return true
	}

	// Check data drift (if available)
	if mmr.hasDataDrift() {
		mmr.logger.Info("Retraining triggered due to data drift")
		return true
	}

	return true
}

// startRetrainingJob starts a retraining job for a specific model
func (mmr *MLModelRetrainer) startRetrainingJob(modelID string) error {
	// Create retraining job
	job := &RetrainingJob{
		ID:               generateID(),
		ModelID:          modelID,
		Status:           "pending",
		Progress:         0.0,
		TrainingDataSize: len(mmr.trainingData),
		StartedAt:        time.Now(),
	}

	// Add to retraining jobs
	mmr.retrainingJobs[job.ID] = job

	// Start retraining in background
	go mmr.executeRetrainingJob(job)

	mmr.logger.Info("Retraining job started",
		zap.String("job_id", job.ID),
		zap.String("model_id", modelID),
		zap.Int("training_data_size", job.TrainingDataSize))

	return nil
}

// executeRetrainingJob executes a retraining job
func (mmr *MLModelRetrainer) executeRetrainingJob(job *RetrainingJob) {
	mmr.logger.Info("Executing retraining job",
		zap.String("job_id", job.ID),
		zap.String("model_id", job.ModelID))

	// Update job status
	mmr.updateJobStatus(job.ID, "running", 0.1)

	// Prepare training data
	trainingData, err := mmr.prepareTrainingData(job.ModelID)
	if err != nil {
		mmr.logger.Error("Failed to prepare training data",
			zap.String("job_id", job.ID),
			zap.Error(err))
		mmr.updateJobStatus(job.ID, "failed", 0.0)
		mmr.setJobError(job.ID, err.Error())
		return
	}

	mmr.updateJobStatus(job.ID, "running", 0.3)

	// Train the model
	newModel, err := mmr.trainModel(job.ModelID, trainingData)
	if err != nil {
		mmr.logger.Error("Failed to train model",
			zap.String("job_id", job.ID),
			zap.Error(err))
		mmr.updateJobStatus(job.ID, "failed", 0.0)
		mmr.setJobError(job.ID, err.Error())
		return
	}

	mmr.updateJobStatus(job.ID, "running", 0.7)

	// Evaluate the new model
	accuracy, err := mmr.evaluateModel(newModel, trainingData)
	if err != nil {
		mmr.logger.Error("Failed to evaluate model",
			zap.String("job_id", job.ID),
			zap.Error(err))
		mmr.updateJobStatus(job.ID, "failed", 0.0)
		mmr.setJobError(job.ID, err.Error())
		return
	}

	mmr.updateJobStatus(job.ID, "running", 0.9)

	// Deploy the new model if it meets criteria
	if accuracy >= mmr.config.RetrainingThreshold {
		err = mmr.deployModel(newModel)
		if err != nil {
			mmr.logger.Error("Failed to deploy model",
				zap.String("job_id", job.ID),
				zap.Error(err))
			mmr.updateJobStatus(job.ID, "failed", 0.0)
			mmr.setJobError(job.ID, err.Error())
			return
		}

		// Update job with final results
		mmr.updateJobStatus(job.ID, "completed", 1.0)
		mmr.setJobAccuracy(job.ID, accuracy)
		mmr.setJobCompleted(job.ID)

		mmr.logger.Info("Retraining job completed successfully",
			zap.String("job_id", job.ID),
			zap.String("model_id", job.ModelID),
			zap.Float64("accuracy", accuracy))
	} else {
		mmr.logger.Warn("New model accuracy below threshold, not deploying",
			zap.String("job_id", job.ID),
			zap.Float64("accuracy", accuracy),
			zap.Float64("threshold", mmr.config.RetrainingThreshold))

		mmr.updateJobStatus(job.ID, "completed", 1.0)
		mmr.setJobAccuracy(job.ID, accuracy)
		mmr.setJobCompleted(job.ID)
	}
}

// prepareTrainingData prepares training data for a specific model
func (mmr *MLModelRetrainer) prepareTrainingData(modelID string) ([]*TrainingDataPoint, error) {
	// Filter and prepare data based on model type
	var filteredData []*TrainingDataPoint

	switch modelID {
	case "bert_classifier":
		filteredData = mmr.prepareBERTTrainingData()
	case "ensemble_classifier":
		filteredData = mmr.prepareEnsembleTrainingData()
	case "uncertainty_estimator":
		filteredData = mmr.prepareUncertaintyTrainingData()
	default:
		return nil, fmt.Errorf("unknown model ID: %s", modelID)
	}

	if len(filteredData) == 0 {
		return nil, fmt.Errorf("no training data available for model: %s", modelID)
	}

	// Shuffle data for better training
	mmr.shuffleTrainingData(filteredData)

	return filteredData, nil
}

// prepareBERTTrainingData prepares training data for BERT classifier
func (mmr *MLModelRetrainer) prepareBERTTrainingData() []*TrainingDataPoint {
	var bertData []*TrainingDataPoint

	for _, data := range mmr.trainingData {
		// BERT needs text data (business name and description)
		if data.BusinessName != "" && data.BusinessDescription != "" {
			bertData = append(bertData, data)
		}
	}

	return bertData
}

// prepareEnsembleTrainingData prepares training data for ensemble classifier
func (mmr *MLModelRetrainer) prepareEnsembleTrainingData() []*TrainingDataPoint {
	// Ensemble classifier can use all available data
	return mmr.trainingData
}

// prepareUncertaintyTrainingData prepares training data for uncertainty estimator
func (mmr *MLModelRetrainer) prepareUncertaintyTrainingData() []*TrainingDataPoint {
	var uncertaintyData []*TrainingDataPoint

	for _, data := range mmr.trainingData {
		// Uncertainty estimator needs confidence scores and feedback
		if data.ConfidenceScore > 0 && data.FeedbackType != "" {
			uncertaintyData = append(uncertaintyData, data)
		}
	}

	return uncertaintyData
}

// shuffleTrainingData shuffles training data for better training
func (mmr *MLModelRetrainer) shuffleTrainingData(data []*TrainingDataPoint) {
	// Simple shuffle implementation
	for i := len(data) - 1; i > 0; i-- {
		j := int(time.Now().UnixNano()) % (i + 1)
		data[i], data[j] = data[j], data[i]
	}
}

// trainModel trains a model with the provided data
func (mmr *MLModelRetrainer) trainModel(modelID string, trainingData []*TrainingDataPoint) (*ModelVersion, error) {
	mmr.logger.Info("Training model",
		zap.String("model_id", modelID),
		zap.Int("training_data_size", len(trainingData)))

	// Create new model version
	modelVersion := &ModelVersion{
		ID:               generateID(),
		Version:          fmt.Sprintf("v%d", time.Now().Unix()),
		ModelType:        modelID,
		TrainingDataSize: len(trainingData),
		CreatedAt:        time.Now(),
		Status:           "training",
	}

	// Simulate training process
	// In a real implementation, this would call the actual ML training pipeline
	time.Sleep(2 * time.Second) // Simulate training time

	// Calculate initial accuracy (placeholder)
	modelVersion.Accuracy = mmr.calculateInitialAccuracy(trainingData)

	// Store model version
	mmr.modelVersions[modelVersion.ID] = modelVersion

	mmr.logger.Info("Model training completed",
		zap.String("model_id", modelID),
		zap.String("version", modelVersion.Version),
		zap.Float64("accuracy", modelVersion.Accuracy))

	return modelVersion, nil
}

// evaluateModel evaluates a trained model
func (mmr *MLModelRetrainer) evaluateModel(model *ModelVersion, testData []*TrainingDataPoint) (float64, error) {
	mmr.logger.Info("Evaluating model",
		zap.String("model_id", model.ID),
		zap.String("version", model.Version))

	// Split data for evaluation (80% train, 20% test)
	splitIndex := int(float64(len(testData)) * 0.8)
	evalData := testData[splitIndex:]

	if len(evalData) == 0 {
		return 0.0, fmt.Errorf("no evaluation data available")
	}

	// Simulate model evaluation
	// In a real implementation, this would run the model on test data
	time.Sleep(1 * time.Second) // Simulate evaluation time

	// Calculate accuracy based on feedback
	correct := 0
	for _, data := range evalData {
		if data.FeedbackType == FeedbackTypeAccuracy || data.FeedbackType == FeedbackTypeClassification {
			correct++
		}
	}

	accuracy := float64(correct) / float64(len(evalData))

	mmr.logger.Info("Model evaluation completed",
		zap.String("model_id", model.ID),
		zap.Float64("accuracy", accuracy),
		zap.Int("test_samples", len(evalData)))

	return accuracy, nil
}

// deployModel deploys a trained model
func (mmr *MLModelRetrainer) deployModel(model *ModelVersion) error {
	mmr.logger.Info("Deploying model",
		zap.String("model_id", model.ID),
		zap.String("version", model.Version))

	// Update model status
	model.Status = "deployed"
	model.DeployedAt = time.Now()

	// In a real implementation, this would:
	// 1. Save the model to the model registry
	// 2. Update the model serving endpoint
	// 3. Update configuration to use the new model
	// 4. Perform health checks

	mmr.logger.Info("Model deployed successfully",
		zap.String("model_id", model.ID),
		zap.String("version", model.Version))

	return nil
}

// Helper methods for job management

func (mmr *MLModelRetrainer) updateJobStatus(jobID, status string, progress float64) {
	mmr.mu.Lock()
	defer mmr.mu.Unlock()

	if job, exists := mmr.retrainingJobs[jobID]; exists {
		job.Status = status
		job.Progress = progress
	}
}

func (mmr *MLModelRetrainer) setJobError(jobID, errorMsg string) {
	mmr.mu.Lock()
	defer mmr.mu.Unlock()

	if job, exists := mmr.retrainingJobs[jobID]; exists {
		job.Error = errorMsg
	}
}

func (mmr *MLModelRetrainer) setJobAccuracy(jobID string, accuracy float64) {
	mmr.mu.Lock()
	defer mmr.mu.Unlock()

	if job, exists := mmr.retrainingJobs[jobID]; exists {
		job.NewAccuracy = accuracy
	}
}

func (mmr *MLModelRetrainer) setJobCompleted(jobID string) {
	mmr.mu.Lock()
	defer mmr.mu.Unlock()

	if job, exists := mmr.retrainingJobs[jobID]; exists {
		job.CompletedAt = time.Now()
	}
}

// Helper methods for retraining criteria

func (mmr *MLModelRetrainer) getLastRetrainingTime() time.Time {
	mmr.mu.RLock()
	defer mmr.mu.RUnlock()

	var lastTime time.Time
	for _, job := range mmr.retrainingJobs {
		if job.CompletedAt.After(lastTime) {
			lastTime = job.CompletedAt
		}
	}

	// If no completed jobs, return epoch time
	if lastTime.IsZero() {
		return time.Unix(0, 0)
	}

	return lastTime
}

func (mmr *MLModelRetrainer) hasActiveRetrainingJobs() bool {
	mmr.mu.RLock()
	defer mmr.mu.RUnlock()

	for _, job := range mmr.retrainingJobs {
		if job.Status == "pending" || job.Status == "running" {
			return true
		}
	}

	return false
}

func (mmr *MLModelRetrainer) hasPerformanceDegradation() bool {
	// TODO: Implement performance degradation detection
	// This would compare current model performance with historical performance
	return false
}

func (mmr *MLModelRetrainer) hasDataDrift() bool {
	// TODO: Implement data drift detection
	// This would compare current data distribution with training data distribution
	return false
}

func (mmr *MLModelRetrainer) calculateInitialAccuracy(trainingData []*TrainingDataPoint) float64 {
	if len(trainingData) == 0 {
		return 0.0
	}

	correct := 0
	for _, data := range trainingData {
		if data.FeedbackType == FeedbackTypeAccuracy || data.FeedbackType == FeedbackTypeClassification {
			correct++
		}
	}

	return float64(correct) / float64(len(trainingData))
}

// GetRetrainingMetrics returns retraining metrics
func (mmr *MLModelRetrainer) GetRetrainingMetrics() *ModelRetrainingMetrics {
	mmr.mu.RLock()
	defer mmr.mu.RUnlock()

	metrics := &ModelRetrainingMetrics{
		TotalRetrainingJobs: len(mmr.retrainingJobs),
	}

	// Calculate successful and failed jobs
	for _, job := range mmr.retrainingJobs {
		if job.Status == "completed" && job.Error == "" {
			metrics.SuccessfulJobs++
		} else if job.Status == "failed" || job.Error != "" {
			metrics.FailedJobs++
		}
	}

	// Calculate average accuracy gain
	totalAccuracyGain := 0.0
	accuracyGainCount := 0

	for _, job := range mmr.retrainingJobs {
		if job.Status == "completed" && job.NewAccuracy > 0 {
			// Calculate accuracy gain (simplified)
			accuracyGain := job.NewAccuracy - 0.8 // Assuming baseline accuracy of 0.8
			totalAccuracyGain += accuracyGain
			accuracyGainCount++
		}
	}

	if accuracyGainCount > 0 {
		metrics.AverageAccuracyGain = totalAccuracyGain / float64(accuracyGainCount)
	}

	// Get last retraining time
	metrics.LastRetraining = mmr.getLastRetrainingTime()

	return metrics
}

// GetRetrainingJobs returns retraining jobs
func (mmr *MLModelRetrainer) GetRetrainingJobs(limit int) []*RetrainingJob {
	mmr.mu.RLock()
	defer mmr.mu.RUnlock()

	jobs := make([]*RetrainingJob, 0, len(mmr.retrainingJobs))
	for _, job := range mmr.retrainingJobs {
		jobs = append(jobs, job)
	}

	// Sort by start time (most recent first)
	// In a real implementation, you'd sort by timestamp
	if limit > 0 && limit < len(jobs) {
		jobs = jobs[:limit]
	}

	return jobs
}
