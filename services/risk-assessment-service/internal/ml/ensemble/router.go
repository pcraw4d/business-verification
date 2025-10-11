package ensemble

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	mlmodels "kyb-platform/services/risk-assessment-service/internal/ml/models"
	"kyb-platform/services/risk-assessment-service/internal/models"
)

// EnsembleRouter routes prediction requests to appropriate models based on horizon
type EnsembleRouter struct {
	xgboostModel mlmodels.RiskModel
	lstmModel    mlmodels.RiskModel
	combiner     *EnsembleCombiner
	logger       *zap.Logger
}

// NewEnsembleRouter creates a new ensemble router
func NewEnsembleRouter(xgboostModel, lstmModel mlmodels.RiskModel, logger *zap.Logger) *EnsembleRouter {
	return &EnsembleRouter{
		xgboostModel: xgboostModel,
		lstmModel:    lstmModel,
		combiner:     NewEnsembleCombiner(),
		logger:       logger,
	}
}

// Route determines which model to use based on prediction horizon
func (er *EnsembleRouter) Route(horizonMonths int) string {
	switch {
	case horizonMonths <= 3:
		return "xgboost" // Best for short-term predictions
	case horizonMonths >= 6:
		return "lstm" // Best for long-term predictions
	default:
		return "ensemble" // Combine both for medium-term predictions
	}
}

// PredictWithEnsemble performs ensemble prediction for the given business
func (er *EnsembleRouter) PredictWithEnsemble(ctx context.Context, business *models.RiskAssessmentRequest) (*models.RiskAssessment, error) {
	startTime := time.Now()

	er.logger.Info("Performing ensemble prediction",
		zap.String("business_name", business.BusinessName),
		zap.Int("horizon_months", business.PredictionHorizon))

	// Get predictions from both models
	var xgbPrediction, lstmPrediction *models.RiskAssessment

	// Run predictions in parallel
	xgbChan := make(chan *models.RiskAssessment, 1)
	lstmChan := make(chan *models.RiskAssessment, 1)
	errChan := make(chan error, 2)

	// XGBoost prediction
	go func() {
		pred, err := er.xgboostModel.Predict(ctx, business)
		if err != nil {
			errChan <- fmt.Errorf("XGBoost prediction failed: %w", err)
			return
		}
		xgbChan <- pred
	}()

	// LSTM prediction
	go func() {
		pred, err := er.lstmModel.Predict(ctx, business)
		if err != nil {
			errChan <- fmt.Errorf("LSTM prediction failed: %w", err)
			return
		}
		lstmChan <- pred
	}()

	// Collect results
	completed := 0
	for completed < 2 {
		select {
		case pred := <-xgbChan:
			xgbPrediction = pred
			completed++
		case pred := <-lstmChan:
			lstmPrediction = pred
			completed++
		case err := <-errChan:
			er.logger.Warn("Model prediction failed", zap.Error(err))
			completed++
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Check if we have at least one successful prediction
	if xgbPrediction == nil && lstmPrediction == nil {
		return nil, fmt.Errorf("both model predictions failed")
	}

	// If only one model succeeded, return its prediction
	if xgbPrediction == nil {
		er.logger.Info("Using LSTM prediction only (XGBoost failed)")
		return lstmPrediction, nil
	}
	if lstmPrediction == nil {
		er.logger.Info("Using XGBoost prediction only (LSTM failed)")
		return xgbPrediction, nil
	}

	// Combine predictions from both models
	ensemblePrediction, err := er.combiner.CombinePredictions(xgbPrediction, lstmPrediction, business.PredictionHorizon)
	if err != nil {
		er.logger.Warn("Ensemble combination failed, using XGBoost prediction", zap.Error(err))
		return xgbPrediction, nil
	}

	duration := time.Since(startTime)
	er.logger.Info("Ensemble prediction completed",
		zap.Duration("duration", duration),
		zap.Float64("ensemble_risk_score", ensemblePrediction.RiskScore),
		zap.String("ensemble_risk_level", string(ensemblePrediction.RiskLevel)))

	return ensemblePrediction, nil
}

// PredictFutureWithEnsemble performs ensemble future prediction
func (er *EnsembleRouter) PredictFutureWithEnsemble(ctx context.Context, business *models.RiskAssessmentRequest, horizonMonths int) (*models.RiskPrediction, error) {
	startTime := time.Now()

	er.logger.Info("Performing ensemble future prediction",
		zap.String("business_name", business.BusinessName),
		zap.Int("horizon_months", horizonMonths))

	// Get predictions from both models
	var xgbPrediction, lstmPrediction *models.RiskPrediction

	// Run predictions in parallel
	xgbChan := make(chan *models.RiskPrediction, 1)
	lstmChan := make(chan *models.RiskPrediction, 1)
	errChan := make(chan error, 2)

	// XGBoost prediction
	go func() {
		pred, err := er.xgboostModel.PredictFuture(ctx, business, horizonMonths)
		if err != nil {
			errChan <- fmt.Errorf("XGBoost future prediction failed: %w", err)
			return
		}
		xgbChan <- pred
	}()

	// LSTM prediction
	go func() {
		pred, err := er.lstmModel.PredictFuture(ctx, business, horizonMonths)
		if err != nil {
			errChan <- fmt.Errorf("LSTM future prediction failed: %w", err)
			return
		}
		lstmChan <- pred
	}()

	// Collect results
	completed := 0
	for completed < 2 {
		select {
		case pred := <-xgbChan:
			xgbPrediction = pred
			completed++
		case pred := <-lstmChan:
			lstmPrediction = pred
			completed++
		case err := <-errChan:
			er.logger.Warn("Model future prediction failed", zap.Error(err))
			completed++
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Check if we have at least one successful prediction
	if xgbPrediction == nil && lstmPrediction == nil {
		return nil, fmt.Errorf("both model future predictions failed")
	}

	// If only one model succeeded, return its prediction
	if xgbPrediction == nil {
		er.logger.Info("Using LSTM future prediction only (XGBoost failed)")
		return lstmPrediction, nil
	}
	if lstmPrediction == nil {
		er.logger.Info("Using XGBoost future prediction only (LSTM failed)")
		return xgbPrediction, nil
	}

	// Combine future predictions from both models
	ensemblePrediction, err := er.combiner.CombineFuturePredictions(xgbPrediction, lstmPrediction, horizonMonths)
	if err != nil {
		er.logger.Warn("Ensemble future combination failed, using XGBoost prediction", zap.Error(err))
		return xgbPrediction, nil
	}

	duration := time.Since(startTime)
	er.logger.Info("Ensemble future prediction completed",
		zap.Duration("duration", duration),
		zap.Float64("ensemble_predicted_score", ensemblePrediction.PredictedScore),
		zap.String("ensemble_predicted_level", string(ensemblePrediction.PredictedLevel)))

	return ensemblePrediction, nil
}

// GetModelWeights returns the weights for ensemble combination based on horizon
func (er *EnsembleRouter) GetModelWeights(horizonMonths int) (float64, float64) {
	switch {
	case horizonMonths <= 3:
		// Short-term: heavily favor XGBoost
		return 0.8, 0.2
	case horizonMonths >= 6:
		// Long-term: heavily favor LSTM
		return 0.2, 0.8
	default:
		// Medium-term: balanced ensemble
		return 0.5, 0.5
	}
}

// GetModelInfo returns information about the ensemble router
func (er *EnsembleRouter) GetModelInfo() map[string]interface{} {
	info := map[string]interface{}{
		"type":             "ensemble_router",
		"available_models": []string{"xgboost", "lstm"},
		"routing_logic": map[string]interface{}{
			"short_term": map[string]interface{}{
				"horizon_months": "1-3",
				"model":          "xgboost",
				"reason":         "Best for short-term tabular predictions",
			},
			"medium_term": map[string]interface{}{
				"horizon_months": "3-6",
				"model":          "ensemble",
				"reason":         "Combined predictions for balanced accuracy",
			},
			"long_term": map[string]interface{}{
				"horizon_months": "6-12",
				"model":          "lstm",
				"reason":         "Best for long-term temporal patterns",
			},
		},
		"ensemble_weights": map[string]interface{}{
			"short_term":  map[string]float64{"xgboost": 0.8, "lstm": 0.2},
			"medium_term": map[string]float64{"xgboost": 0.5, "lstm": 0.5},
			"long_term":   map[string]float64{"xgboost": 0.2, "lstm": 0.8},
		},
	}

	// Add individual model info if available
	if er.xgboostModel != nil {
		info["xgboost_model"] = er.xgboostModel.GetModelInfo()
	}
	if er.lstmModel != nil {
		info["lstm_model"] = er.lstmModel.GetModelInfo()
	}

	return info
}

// Health checks the health of all models in the ensemble
func (er *EnsembleRouter) Health(ctx context.Context) error {
	var errors []error

	// Check XGBoost model
	if er.xgboostModel == nil {
		errors = append(errors, fmt.Errorf("XGBoost model is nil"))
	} else {
		// Try a simple prediction to check health
		testBusiness := &models.RiskAssessmentRequest{
			BusinessName:    "Health Check Test",
			BusinessAddress: "123 Test St",
			Industry:        "technology",
			Country:         "US",
		}
		_, err := er.xgboostModel.Predict(ctx, testBusiness)
		if err != nil {
			errors = append(errors, fmt.Errorf("XGBoost model health check failed: %w", err))
		}
	}

	// Check LSTM model
	if er.lstmModel == nil {
		errors = append(errors, fmt.Errorf("LSTM model is nil"))
	} else {
		// Try a simple prediction to check health
		testBusiness := &models.RiskAssessmentRequest{
			BusinessName:    "Health Check Test",
			BusinessAddress: "123 Test St",
			Industry:        "technology",
			Country:         "US",
		}
		_, err := er.lstmModel.Predict(ctx, testBusiness)
		if err != nil {
			errors = append(errors, fmt.Errorf("LSTM model health check failed: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("ensemble health check failed: %v", errors)
	}

	return nil
}
