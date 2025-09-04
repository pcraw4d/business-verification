-- Classification Accuracy and Response Time Monitoring for Business Classification System
-- This script provides comprehensive monitoring for classification accuracy and response times

-- 1. Create a classification accuracy monitoring table
CREATE TABLE IF NOT EXISTS classification_accuracy_metrics (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    request_id VARCHAR(255) NOT NULL,
    business_name TEXT,
    business_description TEXT,
    website_url TEXT,
    predicted_industry VARCHAR(255) NOT NULL,
    predicted_confidence DECIMAL(5,2) NOT NULL,
    actual_industry VARCHAR(255),
    actual_confidence DECIMAL(5,2),
    accuracy_score DECIMAL(5,2),
    response_time_ms DECIMAL(10,2) NOT NULL,
    processing_time_ms DECIMAL(10,2),
    classification_method VARCHAR(100),
    keywords_used TEXT[],
    confidence_threshold DECIMAL(5,2) DEFAULT 70.0,
    is_correct BOOLEAN,
    error_message TEXT,
    user_feedback VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 2. Create a function to log classification accuracy metrics
CREATE OR REPLACE FUNCTION log_classification_accuracy_metrics(
    p_request_id VARCHAR(255),
    p_business_name TEXT,
    p_business_description TEXT,
    p_website_url TEXT,
    p_predicted_industry VARCHAR(255),
    p_predicted_confidence DECIMAL,
    p_actual_industry VARCHAR(255) DEFAULT NULL,
    p_actual_confidence DECIMAL DEFAULT NULL,
    p_response_time_ms DECIMAL,
    p_processing_time_ms DECIMAL DEFAULT NULL,
    p_classification_method VARCHAR(100) DEFAULT NULL,
    p_keywords_used TEXT[] DEFAULT NULL,
    p_confidence_threshold DECIMAL DEFAULT 70.0,
    p_error_message TEXT DEFAULT NULL,
    p_user_feedback VARCHAR(50) DEFAULT NULL
) RETURNS INTEGER AS $$
DECLARE
    v_log_id INTEGER;
    v_accuracy_score DECIMAL;
    v_is_correct BOOLEAN;
BEGIN
    -- Calculate accuracy score if actual industry is provided
    IF p_actual_industry IS NOT NULL THEN
        IF p_predicted_industry = p_actual_industry THEN
            v_accuracy_score := 100.0;
            v_is_correct := TRUE;
        ELSE
            v_accuracy_score := 0.0;
            v_is_correct := FALSE;
        END IF;
    ELSE
        v_accuracy_score := NULL;
        v_is_correct := NULL;
    END IF;
    
    INSERT INTO classification_accuracy_metrics (
        request_id,
        business_name,
        business_description,
        website_url,
        predicted_industry,
        predicted_confidence,
        actual_industry,
        actual_confidence,
        accuracy_score,
        response_time_ms,
        processing_time_ms,
        classification_method,
        keywords_used,
        confidence_threshold,
        is_correct,
        error_message,
        user_feedback
    ) VALUES (
        p_request_id,
        p_business_name,
        p_business_description,
        p_website_url,
        p_predicted_industry,
        p_predicted_confidence,
        p_actual_industry,
        p_actual_confidence,
        v_accuracy_score,
        p_response_time_ms,
        p_processing_time_ms,
        p_classification_method,
        p_keywords_used,
        p_confidence_threshold,
        v_is_correct,
        p_error_message,
        p_user_feedback
    ) RETURNING id INTO v_log_id;
    
    RETURN v_log_id;
END;
$$ LANGUAGE plpgsql;

-- 3. Create a function to get classification accuracy statistics
CREATE OR REPLACE FUNCTION get_classification_accuracy_stats(
    p_hours_back INTEGER DEFAULT 24
) RETURNS TABLE (
    total_classifications BIGINT,
    correct_classifications BIGINT,
    accuracy_percentage DECIMAL,
    avg_response_time_ms DECIMAL,
    avg_processing_time_ms DECIMAL,
    avg_confidence DECIMAL,
    confidence_distribution JSONB,
    method_accuracy JSONB,
    error_rate DECIMAL,
    user_feedback_distribution JSONB
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        COUNT(*) as total_classifications,
        COUNT(*) FILTER (WHERE is_correct = TRUE) as correct_classifications,
        ROUND(
            (COUNT(*) FILTER (WHERE is_correct = TRUE)::DECIMAL / 
             NULLIF(COUNT(*) FILTER (WHERE is_correct IS NOT NULL), 0)) * 100, 
            2
        ) as accuracy_percentage,
        ROUND(AVG(response_time_ms), 2) as avg_response_time_ms,
        ROUND(AVG(processing_time_ms), 2) as avg_processing_time_ms,
        ROUND(AVG(predicted_confidence), 2) as avg_confidence,
        jsonb_build_object(
            'high_confidence', COUNT(*) FILTER (WHERE predicted_confidence >= 90),
            'medium_confidence', COUNT(*) FILTER (WHERE predicted_confidence >= 70 AND predicted_confidence < 90),
            'low_confidence', COUNT(*) FILTER (WHERE predicted_confidence < 70)
        ) as confidence_distribution,
        jsonb_object_agg(
            classification_method, 
            ROUND(
                (COUNT(*) FILTER (WHERE is_correct = TRUE AND classification_method = cam.classification_method)::DECIMAL / 
                 NULLIF(COUNT(*) FILTER (WHERE classification_method = cam.classification_method), 0)) * 100, 
                2
            )
        ) as method_accuracy,
        ROUND(
            (COUNT(*) FILTER (WHERE error_message IS NOT NULL)::DECIMAL / COUNT(*)) * 100, 
            2
        ) as error_rate,
        jsonb_object_agg(
            COALESCE(user_feedback, 'no_feedback'), 
            COUNT(*)
        ) as user_feedback_distribution
    FROM classification_accuracy_metrics cam
    WHERE timestamp >= NOW() - INTERVAL '1 hour' * p_hours_back;
END;
$$ LANGUAGE plpgsql;

-- 4. Create a function to get classification accuracy trends
CREATE OR REPLACE FUNCTION get_classification_accuracy_trends(
    p_hours_back INTEGER DEFAULT 168
) RETURNS TABLE (
    hour_bucket TIMESTAMPTZ,
    total_classifications BIGINT,
    correct_classifications BIGINT,
    accuracy_percentage DECIMAL,
    avg_response_time_ms DECIMAL,
    avg_processing_time_ms DECIMAL,
    avg_confidence DECIMAL,
    error_count BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        date_trunc('hour', timestamp) as hour_bucket,
        COUNT(*) as total_classifications,
        COUNT(*) FILTER (WHERE is_correct = TRUE) as correct_classifications,
        ROUND(
            (COUNT(*) FILTER (WHERE is_correct = TRUE)::DECIMAL / 
             NULLIF(COUNT(*) FILTER (WHERE is_correct IS NOT NULL), 0)) * 100, 
            2
        ) as accuracy_percentage,
        ROUND(AVG(response_time_ms), 2) as avg_response_time_ms,
        ROUND(AVG(processing_time_ms), 2) as avg_processing_time_ms,
        ROUND(AVG(predicted_confidence), 2) as avg_confidence,
        COUNT(*) FILTER (WHERE error_message IS NOT NULL) as error_count
    FROM classification_accuracy_metrics
    WHERE timestamp >= NOW() - INTERVAL '1 hour' * p_hours_back
    GROUP BY date_trunc('hour', timestamp)
    ORDER BY hour_bucket DESC;
END;
$$ LANGUAGE plpgsql;

-- 5. Create a function to get classification accuracy alerts
CREATE OR REPLACE FUNCTION get_classification_accuracy_alerts(
    p_hours_back INTEGER DEFAULT 1
) RETURNS TABLE (
    alert_id SERIAL,
    alert_type VARCHAR(50),
    alert_level VARCHAR(20),
    alert_message TEXT,
    metric_value DECIMAL,
    threshold_value DECIMAL,
    recommendations TEXT,
    created_at TIMESTAMPTZ
) AS $$
BEGIN
    RETURN QUERY
    -- Low accuracy alerts
    SELECT 
        cam.id as alert_id,
        'LOW_ACCURACY' as alert_type,
        CASE 
            WHEN accuracy_percentage < 60 THEN 'CRITICAL'
            WHEN accuracy_percentage < 75 THEN 'HIGH'
            WHEN accuracy_percentage < 85 THEN 'MEDIUM'
            ELSE 'LOW'
        END as alert_level,
        'Low classification accuracy: ' || ROUND(accuracy_percentage, 2) || '%' as alert_message,
        accuracy_percentage as metric_value,
        85.0 as threshold_value,
        'Review classification algorithms and training data' as recommendations,
        cam.timestamp as created_at
    FROM (
        SELECT 
            id,
            timestamp,
            ROUND(
                (COUNT(*) FILTER (WHERE is_correct = TRUE)::DECIMAL / 
                 NULLIF(COUNT(*) FILTER (WHERE is_correct IS NOT NULL), 0)) * 100, 
                2
            ) as accuracy_percentage
        FROM classification_accuracy_metrics
        WHERE timestamp >= NOW() - INTERVAL '1 hour' * p_hours_back
        GROUP BY id, timestamp
    ) cam
    WHERE cam.accuracy_percentage < 85
    
    UNION ALL
    
    -- High response time alerts
    SELECT 
        cam.id as alert_id,
        'HIGH_RESPONSE_TIME' as alert_type,
        CASE 
            WHEN avg_response_time > 5000 THEN 'CRITICAL'
            WHEN avg_response_time > 2000 THEN 'HIGH'
            WHEN avg_response_time > 1000 THEN 'MEDIUM'
            ELSE 'LOW'
        END as alert_level,
        'High response time: ' || ROUND(avg_response_time, 2) || 'ms' as alert_message,
        avg_response_time as metric_value,
        1000.0 as threshold_value,
        'Optimize classification algorithms and database queries' as recommendations,
        cam.timestamp as created_at
    FROM (
        SELECT 
            id,
            timestamp,
            AVG(response_time_ms) as avg_response_time
        FROM classification_accuracy_metrics
        WHERE timestamp >= NOW() - INTERVAL '1 hour' * p_hours_back
        GROUP BY id, timestamp
    ) cam
    WHERE cam.avg_response_time > 1000
    
    UNION ALL
    
    -- High error rate alerts
    SELECT 
        cam.id as alert_id,
        'HIGH_ERROR_RATE' as alert_type,
        CASE 
            WHEN error_rate > 20 THEN 'CRITICAL'
            WHEN error_rate > 10 THEN 'HIGH'
            WHEN error_rate > 5 THEN 'MEDIUM'
            ELSE 'LOW'
        END as alert_level,
        'High error rate: ' || ROUND(error_rate, 2) || '%' as alert_message,
        error_rate as metric_value,
        5.0 as threshold_value,
        'Investigate and fix classification errors' as recommendations,
        cam.timestamp as created_at
    FROM (
        SELECT 
            id,
            timestamp,
            ROUND(
                (COUNT(*) FILTER (WHERE error_message IS NOT NULL)::DECIMAL / COUNT(*)) * 100, 
                2
            ) as error_rate
        FROM classification_accuracy_metrics
        WHERE timestamp >= NOW() - INTERVAL '1 hour' * p_hours_back
        GROUP BY id, timestamp
    ) cam
    WHERE cam.error_rate > 5
    
    UNION ALL
    
    -- Low confidence alerts
    SELECT 
        cam.id as alert_id,
        'LOW_CONFIDENCE' as alert_type,
        CASE 
            WHEN avg_confidence < 50 THEN 'CRITICAL'
            WHEN avg_confidence < 70 THEN 'HIGH'
            WHEN avg_confidence < 80 THEN 'MEDIUM'
            ELSE 'LOW'
        END as alert_level,
        'Low average confidence: ' || ROUND(avg_confidence, 2) || '%' as alert_message,
        avg_confidence as metric_value,
        80.0 as threshold_value,
        'Improve classification algorithms and training data' as recommendations,
        cam.timestamp as created_at
    FROM (
        SELECT 
            id,
            timestamp,
            AVG(predicted_confidence) as avg_confidence
        FROM classification_accuracy_metrics
        WHERE timestamp >= NOW() - INTERVAL '1 hour' * p_hours_back
        GROUP BY id, timestamp
    ) cam
    WHERE cam.avg_confidence < 80
    
    ORDER BY created_at DESC;
END;
$$ LANGUAGE plpgsql;

-- 6. Create a function to get classification accuracy dashboard
CREATE OR REPLACE FUNCTION get_classification_accuracy_dashboard() 
RETURNS TABLE (
    metric_name TEXT,
    current_value TEXT,
    target_value TEXT,
    status TEXT,
    trend TEXT,
    recommendations TEXT
) AS $$
DECLARE
    v_total_classifications BIGINT;
    v_correct_classifications BIGINT;
    v_accuracy_percentage DECIMAL;
    v_avg_response_time_ms DECIMAL;
    v_avg_processing_time_ms DECIMAL;
    v_avg_confidence DECIMAL;
    v_error_rate DECIMAL;
BEGIN
    -- Get current metrics
    SELECT 
        total_classifications,
        correct_classifications,
        accuracy_percentage,
        avg_response_time_ms,
        avg_processing_time_ms,
        avg_confidence,
        error_rate
    INTO 
        v_total_classifications,
        v_correct_classifications,
        v_accuracy_percentage,
        v_avg_response_time_ms,
        v_avg_processing_time_ms,
        v_avg_confidence,
        v_error_rate
    FROM get_classification_accuracy_stats(24);
    
    RETURN QUERY
    -- Classification Accuracy
    SELECT 
        'Classification Accuracy' as metric_name,
        COALESCE(v_accuracy_percentage::TEXT, 'N/A') || '%' as current_value,
        '85%' as target_value,
        CASE 
            WHEN v_accuracy_percentage IS NULL THEN 'UNKNOWN'
            WHEN v_accuracy_percentage < 60 THEN 'CRITICAL'
            WHEN v_accuracy_percentage < 75 THEN 'WARNING'
            WHEN v_accuracy_percentage < 85 THEN 'FAIR'
            ELSE 'GOOD'
        END as status,
        'STABLE' as trend,
        CASE 
            WHEN v_accuracy_percentage IS NULL THEN 'No accuracy data available'
            WHEN v_accuracy_percentage < 60 THEN 'Critical: Classification accuracy is very low'
            WHEN v_accuracy_percentage < 75 THEN 'Warning: Classification accuracy needs improvement'
            WHEN v_accuracy_percentage < 85 THEN 'Fair: Classification accuracy is acceptable'
            ELSE 'Good: Classification accuracy is excellent'
        END as recommendations
    
    UNION ALL
    
    -- Response Time
    SELECT 
        'Response Time' as metric_name,
        COALESCE(v_avg_response_time_ms::TEXT, 'N/A') || ' ms' as current_value,
        '1000 ms' as target_value,
        CASE 
            WHEN v_avg_response_time_ms IS NULL THEN 'UNKNOWN'
            WHEN v_avg_response_time_ms > 5000 THEN 'CRITICAL'
            WHEN v_avg_response_time_ms > 2000 THEN 'WARNING'
            WHEN v_avg_response_time_ms > 1000 THEN 'FAIR'
            ELSE 'GOOD'
        END as status,
        'STABLE' as trend,
        CASE 
            WHEN v_avg_response_time_ms IS NULL THEN 'No response time data available'
            WHEN v_avg_response_time_ms > 5000 THEN 'Critical: Response times are very high'
            WHEN v_avg_response_time_ms > 2000 THEN 'Warning: Response times are high'
            WHEN v_avg_response_time_ms > 1000 THEN 'Fair: Response times are acceptable'
            ELSE 'Good: Response times are excellent'
        END as recommendations
    
    UNION ALL
    
    -- Processing Time
    SELECT 
        'Processing Time' as metric_name,
        COALESCE(v_avg_processing_time_ms::TEXT, 'N/A') || ' ms' as current_value,
        '500 ms' as target_value,
        CASE 
            WHEN v_avg_processing_time_ms IS NULL THEN 'UNKNOWN'
            WHEN v_avg_processing_time_ms > 2000 THEN 'CRITICAL'
            WHEN v_avg_processing_time_ms > 1000 THEN 'WARNING'
            WHEN v_avg_processing_time_ms > 500 THEN 'FAIR'
            ELSE 'GOOD'
        END as status,
        'STABLE' as trend,
        CASE 
            WHEN v_avg_processing_time_ms IS NULL THEN 'No processing time data available'
            WHEN v_avg_processing_time_ms > 2000 THEN 'Critical: Processing times are very high'
            WHEN v_avg_processing_time_ms > 1000 THEN 'Warning: Processing times are high'
            WHEN v_avg_processing_time_ms > 500 THEN 'Fair: Processing times are acceptable'
            ELSE 'Good: Processing times are excellent'
        END as recommendations
    
    UNION ALL
    
    -- Average Confidence
    SELECT 
        'Average Confidence' as metric_name,
        COALESCE(v_avg_confidence::TEXT, 'N/A') || '%' as current_value,
        '80%' as target_value,
        CASE 
            WHEN v_avg_confidence IS NULL THEN 'UNKNOWN'
            WHEN v_avg_confidence < 50 THEN 'CRITICAL'
            WHEN v_avg_confidence < 70 THEN 'WARNING'
            WHEN v_avg_confidence < 80 THEN 'FAIR'
            ELSE 'GOOD'
        END as status,
        'STABLE' as trend,
        CASE 
            WHEN v_avg_confidence IS NULL THEN 'No confidence data available'
            WHEN v_avg_confidence < 50 THEN 'Critical: Average confidence is very low'
            WHEN v_avg_confidence < 70 THEN 'Warning: Average confidence needs improvement'
            WHEN v_avg_confidence < 80 THEN 'Fair: Average confidence is acceptable'
            ELSE 'Good: Average confidence is excellent'
        END as recommendations
    
    UNION ALL
    
    -- Error Rate
    SELECT 
        'Error Rate' as metric_name,
        COALESCE(v_error_rate::TEXT, 'N/A') || '%' as current_value,
        '5%' as target_value,
        CASE 
            WHEN v_error_rate IS NULL THEN 'UNKNOWN'
            WHEN v_error_rate > 20 THEN 'CRITICAL'
            WHEN v_error_rate > 10 THEN 'WARNING'
            WHEN v_error_rate > 5 THEN 'FAIR'
            ELSE 'GOOD'
        END as status,
        'STABLE' as trend,
        CASE 
            WHEN v_error_rate IS NULL THEN 'No error data available'
            WHEN v_error_rate > 20 THEN 'Critical: Error rate is very high'
            WHEN v_error_rate > 10 THEN 'Warning: Error rate is high'
            WHEN v_error_rate > 5 THEN 'Fair: Error rate is acceptable'
            ELSE 'Good: Error rate is excellent'
        END as recommendations;
END;
$$ LANGUAGE plpgsql;

-- 7. Create a function to get classification accuracy insights
CREATE OR REPLACE FUNCTION get_classification_accuracy_insights() 
RETURNS TABLE (
    insight_type TEXT,
    insight_title TEXT,
    insight_description TEXT,
    insight_priority VARCHAR(20),
    insight_recommendations TEXT,
    affected_classifications BIGINT,
    potential_improvement DECIMAL
) AS $$
BEGIN
    RETURN QUERY
    -- Low accuracy insights
    SELECT 
        'LOW_ACCURACY' as insight_type,
        'Classification Accuracy Improvement' as insight_title,
        'Low classification accuracy detected' as insight_description,
        'HIGH' as insight_priority,
        'Review and improve classification algorithms and training data' as insight_recommendations,
        COUNT(*) as affected_classifications,
        ROUND(AVG(85 - accuracy_score), 2) as potential_improvement
    FROM classification_accuracy_metrics
    WHERE timestamp >= NOW() - INTERVAL '24 hours'
    AND accuracy_score < 85
    AND accuracy_score IS NOT NULL
    
    UNION ALL
    
    -- High response time insights
    SELECT 
        'HIGH_RESPONSE_TIME' as insight_type,
        'Response Time Optimization' as insight_title,
        'High response times detected' as insight_description,
        'MEDIUM' as insight_priority,
        'Optimize classification algorithms and database queries' as insight_recommendations,
        COUNT(*) as affected_classifications,
        ROUND(AVG(response_time_ms - 1000), 2) as potential_improvement
    FROM classification_accuracy_metrics
    WHERE timestamp >= NOW() - INTERVAL '24 hours'
    AND response_time_ms > 1000
    
    UNION ALL
    
    -- Low confidence insights
    SELECT 
        'LOW_CONFIDENCE' as insight_type,
        'Confidence Score Improvement' as insight_title,
        'Low confidence scores detected' as insight_description,
        'MEDIUM' as insight_priority,
        'Improve classification algorithms and training data quality' as insight_recommendations,
        COUNT(*) as affected_classifications,
        ROUND(AVG(80 - predicted_confidence), 2) as potential_improvement
    FROM classification_accuracy_metrics
    WHERE timestamp >= NOW() - INTERVAL '24 hours'
    AND predicted_confidence < 80
    
    UNION ALL
    
    -- High error rate insights
    SELECT 
        'HIGH_ERROR_RATE' as insight_type,
        'Error Rate Reduction' as insight_title,
        'High error rate detected' as insight_description,
        'HIGH' as insight_priority,
        'Investigate and fix classification errors' as insight_recommendations,
        COUNT(*) as affected_classifications,
        ROUND(AVG(CASE WHEN error_message IS NOT NULL THEN 1 ELSE 0 END) * 100, 2) as potential_improvement
    FROM classification_accuracy_metrics
    WHERE timestamp >= NOW() - INTERVAL '24 hours'
    AND error_message IS NOT NULL;
END;
$$ LANGUAGE plpgsql;

-- 8. Create a function to analyze classification performance
CREATE OR REPLACE FUNCTION analyze_classification_performance(
    p_hours_back INTEGER DEFAULT 24
) RETURNS TABLE (
    performance_metric TEXT,
    current_value DECIMAL,
    target_value DECIMAL,
    performance_score DECIMAL,
    status TEXT,
    recommendations TEXT
) AS $$
DECLARE
    v_total_classifications BIGINT;
    v_correct_classifications BIGINT;
    v_accuracy_percentage DECIMAL;
    v_avg_response_time_ms DECIMAL;
    v_avg_processing_time_ms DECIMAL;
    v_avg_confidence DECIMAL;
    v_error_rate DECIMAL;
BEGIN
    -- Get current metrics
    SELECT 
        total_classifications,
        correct_classifications,
        accuracy_percentage,
        avg_response_time_ms,
        avg_processing_time_ms,
        avg_confidence,
        error_rate
    INTO 
        v_total_classifications,
        v_correct_classifications,
        v_accuracy_percentage,
        v_avg_response_time_ms,
        v_avg_processing_time_ms,
        v_avg_confidence,
        v_error_rate
    FROM get_classification_accuracy_stats(p_hours_back);
    
    RETURN QUERY
    -- Accuracy Performance
    SELECT 
        'Accuracy' as performance_metric,
        COALESCE(v_accuracy_percentage, 0) as current_value,
        85.0 as target_value,
        CASE 
            WHEN v_accuracy_percentage IS NULL THEN 0
            WHEN v_accuracy_percentage >= 85 THEN 100
            WHEN v_accuracy_percentage >= 75 THEN 80
            WHEN v_accuracy_percentage >= 60 THEN 60
            ELSE 40
        END as performance_score,
        CASE 
            WHEN v_accuracy_percentage IS NULL THEN 'UNKNOWN'
            WHEN v_accuracy_percentage >= 85 THEN 'EXCELLENT'
            WHEN v_accuracy_percentage >= 75 THEN 'GOOD'
            WHEN v_accuracy_percentage >= 60 THEN 'FAIR'
            ELSE 'POOR'
        END as status,
        CASE 
            WHEN v_accuracy_percentage IS NULL THEN 'No accuracy data available'
            WHEN v_accuracy_percentage >= 85 THEN 'Accuracy is excellent'
            WHEN v_accuracy_percentage >= 75 THEN 'Accuracy is good, consider minor improvements'
            WHEN v_accuracy_percentage >= 60 THEN 'Accuracy needs improvement'
            ELSE 'Accuracy requires significant improvement'
        END as recommendations
    
    UNION ALL
    
    -- Response Time Performance
    SELECT 
        'Response Time' as performance_metric,
        COALESCE(v_avg_response_time_ms, 0) as current_value,
        1000.0 as target_value,
        CASE 
            WHEN v_avg_response_time_ms IS NULL THEN 0
            WHEN v_avg_response_time_ms <= 1000 THEN 100
            WHEN v_avg_response_time_ms <= 2000 THEN 80
            WHEN v_avg_response_time_ms <= 5000 THEN 60
            ELSE 40
        END as performance_score,
        CASE 
            WHEN v_avg_response_time_ms IS NULL THEN 'UNKNOWN'
            WHEN v_avg_response_time_ms <= 1000 THEN 'EXCELLENT'
            WHEN v_avg_response_time_ms <= 2000 THEN 'GOOD'
            WHEN v_avg_response_time_ms <= 5000 THEN 'FAIR'
            ELSE 'POOR'
        END as status,
        CASE 
            WHEN v_avg_response_time_ms IS NULL THEN 'No response time data available'
            WHEN v_avg_response_time_ms <= 1000 THEN 'Response time is excellent'
            WHEN v_avg_response_time_ms <= 2000 THEN 'Response time is good, consider minor optimizations'
            WHEN v_avg_response_time_ms <= 5000 THEN 'Response time needs optimization'
            ELSE 'Response time requires significant optimization'
        END as recommendations
    
    UNION ALL
    
    -- Confidence Performance
    SELECT 
        'Confidence' as performance_metric,
        COALESCE(v_avg_confidence, 0) as current_value,
        80.0 as target_value,
        CASE 
            WHEN v_avg_confidence IS NULL THEN 0
            WHEN v_avg_confidence >= 80 THEN 100
            WHEN v_avg_confidence >= 70 THEN 80
            WHEN v_avg_confidence >= 50 THEN 60
            ELSE 40
        END as performance_score,
        CASE 
            WHEN v_avg_confidence IS NULL THEN 'UNKNOWN'
            WHEN v_avg_confidence >= 80 THEN 'EXCELLENT'
            WHEN v_avg_confidence >= 70 THEN 'GOOD'
            WHEN v_avg_confidence >= 50 THEN 'FAIR'
            ELSE 'POOR'
        END as status,
        CASE 
            WHEN v_avg_confidence IS NULL THEN 'No confidence data available'
            WHEN v_avg_confidence >= 80 THEN 'Confidence is excellent'
            WHEN v_avg_confidence >= 70 THEN 'Confidence is good, consider minor improvements'
            WHEN v_avg_confidence >= 50 THEN 'Confidence needs improvement'
            ELSE 'Confidence requires significant improvement'
        END as recommendations
    
    UNION ALL
    
    -- Error Rate Performance
    SELECT 
        'Error Rate' as performance_metric,
        COALESCE(v_error_rate, 0) as current_value,
        5.0 as target_value,
        CASE 
            WHEN v_error_rate IS NULL THEN 0
            WHEN v_error_rate <= 5 THEN 100
            WHEN v_error_rate <= 10 THEN 80
            WHEN v_error_rate <= 20 THEN 60
            ELSE 40
        END as performance_score,
        CASE 
            WHEN v_error_rate IS NULL THEN 'UNKNOWN'
            WHEN v_error_rate <= 5 THEN 'EXCELLENT'
            WHEN v_error_rate <= 10 THEN 'GOOD'
            WHEN v_error_rate <= 20 THEN 'FAIR'
            ELSE 'POOR'
        END as status,
        CASE 
            WHEN v_error_rate IS NULL THEN 'No error data available'
            WHEN v_error_rate <= 5 THEN 'Error rate is excellent'
            WHEN v_error_rate <= 10 THEN 'Error rate is good, consider minor improvements'
            WHEN v_error_rate <= 20 THEN 'Error rate needs improvement'
            ELSE 'Error rate requires significant improvement'
        END as recommendations;
END;
$$ LANGUAGE plpgsql;

-- 9. Create a function to cleanup old classification accuracy metrics
CREATE OR REPLACE FUNCTION cleanup_classification_accuracy_metrics(
    p_days_to_keep INTEGER DEFAULT 30
) RETURNS INTEGER AS $$
DECLARE
    v_deleted_count INTEGER;
BEGIN
    DELETE FROM classification_accuracy_metrics
    WHERE timestamp < NOW() - INTERVAL '1 day' * p_days_to_keep;
    
    GET DIAGNOSTICS v_deleted_count = ROW_COUNT;
    
    RETURN v_deleted_count;
END;
$$ LANGUAGE plpgsql;

-- 10. Create a function to validate classification accuracy monitoring setup
CREATE OR REPLACE FUNCTION validate_classification_accuracy_monitoring_setup() 
RETURNS TABLE (
    component TEXT,
    status TEXT,
    details TEXT,
    recommendation TEXT
) AS $$
BEGIN
    RETURN QUERY
    -- Check if classification_accuracy_metrics table exists
    SELECT 
        'Classification Accuracy Metrics Table' as component,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'classification_accuracy_metrics') 
            THEN 'OK' 
            ELSE 'MISSING' 
        END as status,
        'Table for storing classification accuracy metrics' as details,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'classification_accuracy_metrics') 
            THEN 'Table exists and ready for use' 
            ELSE 'Create classification_accuracy_metrics table' 
        END as recommendation
    
    UNION ALL
    
    -- Check if monitoring functions exist
    SELECT 
        'Classification Accuracy Functions' as component,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.routines WHERE routine_name = 'get_classification_accuracy_stats') 
            THEN 'OK' 
            ELSE 'MISSING' 
        END as status,
        'Functions for monitoring classification accuracy' as details,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.routines WHERE routine_name = 'get_classification_accuracy_stats') 
            THEN 'All classification accuracy functions are available' 
            ELSE 'Create classification accuracy monitoring functions' 
        END as recommendation
    
    UNION ALL
    
    -- Check if logging function exists
    SELECT 
        'Classification Accuracy Logging' as component,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.routines WHERE routine_name = 'log_classification_accuracy_metrics') 
            THEN 'OK' 
            ELSE 'MISSING' 
        END as status,
        'Function for logging classification accuracy metrics' as details,
        CASE 
            WHEN EXISTS (SELECT 1 FROM information_schema.routines WHERE routine_name = 'log_classification_accuracy_metrics') 
            THEN 'Logging function is ready' 
            ELSE 'Create classification accuracy logging function' 
        END as recommendation;
END;
$$ LANGUAGE plpgsql;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_classification_accuracy_metrics_timestamp ON classification_accuracy_metrics(timestamp);
CREATE INDEX IF NOT EXISTS idx_classification_accuracy_metrics_request_id ON classification_accuracy_metrics(request_id);
CREATE INDEX IF NOT EXISTS idx_classification_accuracy_metrics_predicted_industry ON classification_accuracy_metrics(predicted_industry);
CREATE INDEX IF NOT EXISTS idx_classification_accuracy_metrics_accuracy_score ON classification_accuracy_metrics(accuracy_score);
CREATE INDEX IF NOT EXISTS idx_classification_accuracy_metrics_response_time ON classification_accuracy_metrics(response_time_ms);
CREATE INDEX IF NOT EXISTS idx_classification_accuracy_metrics_is_correct ON classification_accuracy_metrics(is_correct);

-- Create a view for easy classification accuracy dashboard access
CREATE OR REPLACE VIEW classification_accuracy_dashboard AS
SELECT 
    'Classification Accuracy Overview' as metric_name,
    (SELECT COALESCE(accuracy_percentage::TEXT, 'N/A') || '%' FROM get_classification_accuracy_stats(24)) as current_value,
    '85%' as target_value,
    (SELECT 
        CASE 
            WHEN accuracy_percentage IS NULL THEN 'UNKNOWN'
            WHEN accuracy_percentage < 60 THEN 'CRITICAL'
            WHEN accuracy_percentage < 75 THEN 'WARNING'
            ELSE 'OK'
        END
    FROM get_classification_accuracy_stats(24)) as status
UNION ALL
SELECT 
    'Response Time' as metric_name,
    (SELECT COALESCE(avg_response_time_ms::TEXT, 'N/A') || ' ms' FROM get_classification_accuracy_stats(24)) as current_value,
    '1000 ms' as target_value,
    (SELECT 
        CASE 
            WHEN avg_response_time_ms IS NULL THEN 'UNKNOWN'
            WHEN avg_response_time_ms > 5000 THEN 'CRITICAL'
            WHEN avg_response_time_ms > 2000 THEN 'WARNING'
            ELSE 'OK'
        END
    FROM get_classification_accuracy_stats(24)) as status
UNION ALL
SELECT 
    'Error Rate' as metric_name,
    (SELECT COALESCE(error_rate::TEXT, 'N/A') || '%' FROM get_classification_accuracy_stats(24)) as current_value,
    '5%' as target_value,
    (SELECT 
        CASE 
            WHEN error_rate IS NULL THEN 'UNKNOWN'
            WHEN error_rate > 20 THEN 'CRITICAL'
            WHEN error_rate > 10 THEN 'WARNING'
            ELSE 'OK'
        END
    FROM get_classification_accuracy_stats(24)) as status;

-- Grant permissions
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO authenticated;
GRANT SELECT ON classification_accuracy_dashboard TO authenticated;
GRANT SELECT, INSERT ON classification_accuracy_metrics TO authenticated;

-- Initial setup message
DO $$
BEGIN
    RAISE NOTICE 'Classification accuracy monitoring setup completed successfully!';
    RAISE NOTICE 'Total functions created: 10';
    RAISE NOTICE 'Total tables created: 1';
    RAISE NOTICE 'Total views created: 1';
    RAISE NOTICE 'Total indexes created: 6';
    RAISE NOTICE 'All classification accuracy monitoring tools are now available.';
    RAISE NOTICE 'Use classification_accuracy_dashboard view to access current classification accuracy metrics.';
    RAISE NOTICE 'Call get_classification_accuracy_stats() to get current classification accuracy statistics.';
    RAISE NOTICE 'Call log_classification_accuracy_metrics() to log classification accuracy metrics.';
END $$;
