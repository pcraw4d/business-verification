-- Migration: Add Classification Accuracy Tracking
-- Purpose: Track classification accuracy for confidence calibration
-- OPTIMIZATION #5.2: Confidence Calibration

-- Create accuracy tracking table
CREATE TABLE IF NOT EXISTS classification_accuracy_tracking (
    id BIGSERIAL PRIMARY KEY,
    request_id VARCHAR(255) UNIQUE NOT NULL,
    business_name VARCHAR(500) NOT NULL,
    website_url TEXT,
    predicted_industry VARCHAR(255) NOT NULL,
    actual_industry VARCHAR(255), -- NULL until validated
    predicted_confidence DECIMAL(5,4) NOT NULL CHECK (predicted_confidence >= 0 AND predicted_confidence <= 1),
    actual_confidence DECIMAL(5,4), -- NULL until validated
    is_correct BOOLEAN, -- NULL until validated
    confidence_bin INTEGER NOT NULL, -- Bin index (0-9 for 10 bins)
    classification_method VARCHAR(100), -- e.g., "multi_strategy", "python_ml", "go_classification"
    keywords_count INTEGER DEFAULT 0,
    processing_time_ms INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    validated_at TIMESTAMP WITH TIME ZONE, -- When actual industry was validated
    validated_by VARCHAR(255) -- Who validated (e.g., "manual", "api", "automated")
);

-- Create indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_accuracy_tracking_confidence_bin ON classification_accuracy_tracking(confidence_bin);
CREATE INDEX IF NOT EXISTS idx_accuracy_tracking_created_at ON classification_accuracy_tracking(created_at);
CREATE INDEX IF NOT EXISTS idx_accuracy_tracking_is_correct ON classification_accuracy_tracking(is_correct) WHERE is_correct IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_accuracy_tracking_predicted_industry ON classification_accuracy_tracking(predicted_industry);
CREATE INDEX IF NOT EXISTS idx_accuracy_tracking_actual_industry ON classification_accuracy_tracking(actual_industry) WHERE actual_industry IS NOT NULL;

-- Create calibration results table
CREATE TABLE IF NOT EXISTS classification_calibration_results (
    id BIGSERIAL PRIMARY KEY,
    calibration_date TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    overall_accuracy DECIMAL(5,4) NOT NULL CHECK (overall_accuracy >= 0 AND overall_accuracy <= 1),
    target_accuracy DECIMAL(5,4) NOT NULL DEFAULT 0.95 CHECK (target_accuracy >= 0 AND target_accuracy <= 1),
    recommended_threshold DECIMAL(5,4) NOT NULL CHECK (recommended_threshold >= 0 AND recommended_threshold <= 1),
    is_well_calibrated BOOLEAN NOT NULL,
    total_classifications BIGINT NOT NULL,
    correct_classifications BIGINT NOT NULL,
    calibration_bins JSONB, -- Store bin data as JSON
    adjustments JSONB, -- Store adjustment factors as JSON
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

-- Create index on calibration date for time-series queries
CREATE INDEX IF NOT EXISTS idx_calibration_results_date ON classification_calibration_results(calibration_date DESC);

-- Create view for accuracy statistics by confidence bin
CREATE OR REPLACE VIEW classification_accuracy_by_bin AS
SELECT 
    confidence_bin,
    COUNT(*) as total_classifications,
    COUNT(CASE WHEN is_correct = true THEN 1 END) as correct_classifications,
    COUNT(CASE WHEN is_correct = false THEN 1 END) as incorrect_classifications,
    ROUND(AVG(predicted_confidence)::numeric, 4) as avg_predicted_confidence,
    ROUND(AVG(CASE WHEN is_correct THEN 1.0 ELSE 0.0 END)::numeric, 4) as actual_accuracy,
    ROUND(AVG(predicted_confidence - CASE WHEN is_correct THEN 1.0 ELSE 0.0 END)::numeric, 4) as calibration_error,
    MIN(created_at) as first_classification,
    MAX(created_at) as last_classification
FROM classification_accuracy_tracking
WHERE is_correct IS NOT NULL
GROUP BY confidence_bin
ORDER BY confidence_bin;

-- Create view for accuracy by industry
CREATE OR REPLACE VIEW classification_accuracy_by_industry AS
SELECT 
    predicted_industry,
    COUNT(*) as total_classifications,
    COUNT(CASE WHEN is_correct = true THEN 1 END) as correct_classifications,
    COUNT(CASE WHEN is_correct = false THEN 1 END) as incorrect_classifications,
    ROUND(AVG(CASE WHEN is_correct THEN 1.0 ELSE 0.0 END)::numeric, 4) as accuracy,
    ROUND(AVG(predicted_confidence)::numeric, 4) as avg_confidence,
    MIN(created_at) as first_classification,
    MAX(created_at) as last_classification
FROM classification_accuracy_tracking
WHERE is_correct IS NOT NULL
GROUP BY predicted_industry
ORDER BY accuracy DESC, total_classifications DESC;

-- Create function to update accuracy tracking when actual industry is provided
CREATE OR REPLACE FUNCTION update_classification_accuracy(
    p_request_id VARCHAR(255),
    p_actual_industry VARCHAR(255),
    p_validated_by VARCHAR(255) DEFAULT 'manual'
)
RETURNS VOID AS $$
BEGIN
    UPDATE classification_accuracy_tracking
    SET 
        actual_industry = p_actual_industry,
        is_correct = (predicted_industry = p_actual_industry),
        validated_at = NOW(),
        validated_by = p_validated_by
    WHERE request_id = p_request_id;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Classification with request_id % not found', p_request_id;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Create function to get calibration statistics
CREATE OR REPLACE FUNCTION get_calibration_statistics(
    p_start_date TIMESTAMP WITH TIME ZONE DEFAULT NOW() - INTERVAL '30 days',
    p_end_date TIMESTAMP WITH TIME ZONE DEFAULT NOW()
)
RETURNS TABLE (
    confidence_bin INTEGER,
    total_classifications BIGINT,
    correct_classifications BIGINT,
    predicted_accuracy DECIMAL(5,4),
    actual_accuracy DECIMAL(5,4),
    calibration_error DECIMAL(5,4)
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        cat.confidence_bin,
        COUNT(*)::BIGINT as total_classifications,
        COUNT(CASE WHEN cat.is_correct = true THEN 1 END)::BIGINT as correct_classifications,
        ROUND(AVG(cat.predicted_confidence)::numeric, 4) as predicted_accuracy,
        ROUND(AVG(CASE WHEN cat.is_correct THEN 1.0 ELSE 0.0 END)::numeric, 4) as actual_accuracy,
        ROUND(AVG(cat.predicted_confidence - CASE WHEN cat.is_correct THEN 1.0 ELSE 0.0 END)::numeric, 4) as calibration_error
    FROM classification_accuracy_tracking cat
    WHERE cat.created_at BETWEEN p_start_date AND p_end_date
      AND cat.is_correct IS NOT NULL
    GROUP BY cat.confidence_bin
    ORDER BY cat.confidence_bin;
END;
$$ LANGUAGE plpgsql;

-- Add comments for documentation
COMMENT ON TABLE classification_accuracy_tracking IS 'Tracks classification results for accuracy analysis and confidence calibration';
COMMENT ON TABLE classification_calibration_results IS 'Stores calibration analysis results and recommendations';
COMMENT ON VIEW classification_accuracy_by_bin IS 'Accuracy statistics grouped by confidence bin';
COMMENT ON VIEW classification_accuracy_by_industry IS 'Accuracy statistics grouped by predicted industry';
COMMENT ON FUNCTION update_classification_accuracy IS 'Updates accuracy tracking when actual industry is validated';
COMMENT ON FUNCTION get_calibration_statistics IS 'Returns calibration statistics for a date range';

