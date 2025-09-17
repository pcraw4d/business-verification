-- Security Validation Performance Monitoring for Business Classification System
-- This script provides comprehensive security validation performance monitoring and analysis

-- 1. Create security validation performance monitoring table
CREATE TABLE IF NOT EXISTS security_validation_performance_log (
    id SERIAL PRIMARY KEY,
    validation_id VARCHAR(255) NOT NULL,
    validation_type VARCHAR(100) NOT NULL,
    validation_name VARCHAR(255) NOT NULL,
    execution_count BIGINT DEFAULT 0,
    total_execution_time_ms DECIMAL(15,4) DEFAULT 0,
    average_execution_time_ms DECIMAL(15,4) DEFAULT 0,
    min_execution_time_ms DECIMAL(15,4) DEFAULT 0,
    max_execution_time_ms DECIMAL(15,4) DEFAULT 0,
    p50_execution_time_ms DECIMAL(15,4),
    p95_execution_time_ms DECIMAL(15,4),
    p99_execution_time_ms DECIMAL(15,4),
    success_count BIGINT DEFAULT 0,
    failure_count BIGINT DEFAULT 0,
    timeout_count BIGINT DEFAULT 0,
    error_count BIGINT DEFAULT 0,
    security_violation_count BIGINT DEFAULT 0,
    compliance_violation_count BIGINT DEFAULT 0,
    threat_detection_count BIGINT DEFAULT 0,
    vulnerability_count BIGINT DEFAULT 0,
    trust_score DECIMAL(5,4),
    confidence_level DECIMAL(5,4),
    risk_level VARCHAR(20),
    performance_category VARCHAR(20),
    security_category VARCHAR(20),
    last_executed TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    first_executed TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 2. Create security validation alerts table
CREATE TABLE IF NOT EXISTS security_validation_alerts (
    id VARCHAR(255) PRIMARY KEY,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    alert_type VARCHAR(100) NOT NULL,
    severity VARCHAR(50) NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    validation_type VARCHAR(100) NOT NULL,
    validation_id VARCHAR(255),
    validation_name VARCHAR(255),
    threshold DECIMAL(15,4),
    actual_value DECIMAL(15,4),
    message TEXT,
    security_impact VARCHAR(100),
    recommendations TEXT[],
    resolved BOOLEAN DEFAULT FALSE,
    resolved_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 3. Create security performance metrics table
CREATE TABLE IF NOT EXISTS security_performance_metrics (
    id VARCHAR(255) PRIMARY KEY,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metric_type VARCHAR(100) NOT NULL,
    validation_type VARCHAR(100) NOT NULL,
    validation_name VARCHAR(255) NOT NULL,
    execution_time_ms DECIMAL(15,4),
    success_rate DECIMAL(5,2),
    failure_rate DECIMAL(5,2),
    timeout_rate DECIMAL(5,2),
    error_rate DECIMAL(5,2),
    security_violation_rate DECIMAL(5,2),
    compliance_violation_rate DECIMAL(5,2),
    threat_detection_rate DECIMAL(5,2),
    vulnerability_rate DECIMAL(5,2),
    trust_score DECIMAL(5,4),
    confidence_level DECIMAL(5,4),
    risk_level VARCHAR(20),
    performance_score DECIMAL(5,2),
    security_score DECIMAL(5,2),
    overall_score DECIMAL(5,2),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 4. Create security system health table
CREATE TABLE IF NOT EXISTS security_system_health (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    overall_security_score DECIMAL(5,2),
    overall_performance_score DECIMAL(5,2),
    overall_risk_level VARCHAR(20),
    active_threats INTEGER,
    security_violations INTEGER,
    compliance_violations INTEGER,
    vulnerabilities INTEGER,
    slow_validations INTEGER,
    failed_validations INTEGER,
    high_risk_validations INTEGER,
    trust_score_average DECIMAL(5,4),
    confidence_level_average DECIMAL(5,4),
    validation_count INTEGER,
    success_rate DECIMAL(5,2),
    failure_rate DECIMAL(5,2),
    average_execution_time_ms DECIMAL(15,4),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 5. Create indexes for faster querying
CREATE INDEX IF NOT EXISTS idx_svpl_validation_id ON security_validation_performance_log (validation_id);
CREATE INDEX IF NOT EXISTS idx_svpl_validation_type ON security_validation_performance_log (validation_type);
CREATE INDEX IF NOT EXISTS idx_svpl_performance_category ON security_validation_performance_log (performance_category);
CREATE INDEX IF NOT EXISTS idx_svpl_security_category ON security_validation_performance_log (security_category);
CREATE INDEX IF NOT EXISTS idx_svpl_risk_level ON security_validation_performance_log (risk_level);
CREATE INDEX IF NOT EXISTS idx_svpl_last_executed ON security_validation_performance_log (last_executed);
CREATE INDEX IF NOT EXISTS idx_svpl_created_at ON security_validation_performance_log (created_at);

CREATE INDEX IF NOT EXISTS idx_sva_timestamp ON security_validation_alerts (timestamp);
CREATE INDEX IF NOT EXISTS idx_sva_alert_type ON security_validation_alerts (alert_type);
CREATE INDEX IF NOT EXISTS idx_sva_severity ON security_validation_alerts (severity);
CREATE INDEX IF NOT EXISTS idx_sva_resolved ON security_validation_alerts (resolved);
CREATE INDEX IF NOT EXISTS idx_sva_validation_type ON security_validation_alerts (validation_type);
CREATE INDEX IF NOT EXISTS idx_sva_validation_id ON security_validation_alerts (validation_id);

CREATE INDEX IF NOT EXISTS idx_spm_timestamp ON security_performance_metrics (timestamp);
CREATE INDEX IF NOT EXISTS idx_spm_metric_type ON security_performance_metrics (metric_type);
CREATE INDEX IF NOT EXISTS idx_spm_validation_type ON security_performance_metrics (validation_type);
CREATE INDEX IF NOT EXISTS idx_spm_risk_level ON security_performance_metrics (risk_level);

CREATE INDEX IF NOT EXISTS idx_ssh_timestamp ON security_system_health (timestamp);
CREATE INDEX IF NOT EXISTS idx_ssh_overall_risk_level ON security_system_health (overall_risk_level);

-- 6. Create function to analyze security validation performance
CREATE OR REPLACE FUNCTION analyze_security_validation_performance(
    p_validation_type VARCHAR,
    p_validation_name VARCHAR,
    p_execution_time_ms DECIMAL DEFAULT 0,
    p_success BOOLEAN DEFAULT TRUE,
    p_security_violation BOOLEAN DEFAULT FALSE,
    p_compliance_violation BOOLEAN DEFAULT FALSE,
    p_threat_detected BOOLEAN DEFAULT FALSE,
    p_vulnerability_found BOOLEAN DEFAULT FALSE,
    p_trust_score DECIMAL DEFAULT 1.0,
    p_confidence_level DECIMAL DEFAULT 1.0
) RETURNS TABLE (
    validation_id VARCHAR,
    performance_score DECIMAL,
    security_score DECIMAL,
    overall_score DECIMAL,
    performance_category VARCHAR,
    security_category VARCHAR,
    risk_level VARCHAR,
    recommendations TEXT[]
) AS $$
DECLARE
    v_validation_id VARCHAR;
    v_performance_score DECIMAL;
    v_security_score DECIMAL;
    v_overall_score DECIMAL;
    v_performance_category VARCHAR;
    v_security_category VARCHAR;
    v_risk_level VARCHAR;
    v_recommendations TEXT[];
    v_slow_validation_threshold DECIMAL := 200.0; -- 200ms threshold
BEGIN
    -- Generate validation ID
    v_validation_id := p_validation_type || '_' || p_validation_name || '_' || extract(epoch from now())::bigint;
    
    -- Calculate performance score (0-100, higher is better)
    v_performance_score := 100.0;
    
    -- Deduct points for slow execution
    IF p_execution_time_ms > v_slow_validation_threshold THEN
        v_performance_score := v_performance_score - ((p_execution_time_ms - v_slow_validation_threshold) / 10);
    END IF;
    
    -- Deduct points for failure
    IF NOT p_success THEN
        v_performance_score := v_performance_score - 30;
    END IF;
    
    -- Ensure performance score is between 0 and 100
    v_performance_score := GREATEST(0, LEAST(100, v_performance_score));
    
    -- Calculate security score (0-100, higher is better)
    v_security_score := 100.0;
    
    -- Deduct points for security violations
    IF p_security_violation THEN
        v_security_score := v_security_score - 100; -- Critical impact
    END IF;
    
    -- Deduct points for compliance violations
    IF p_compliance_violation THEN
        v_security_score := v_security_score - 80; -- High impact
    END IF;
    
    -- Deduct points for threats detected
    IF p_threat_detected THEN
        v_security_score := v_security_score - 90; -- Critical impact
    END IF;
    
    -- Deduct points for vulnerabilities
    IF p_vulnerability_found THEN
        v_security_score := v_security_score - 70; -- High impact
    END IF;
    
    -- Deduct points for low trust score
    IF p_trust_score < 0.5 THEN
        v_security_score := v_security_score - 20;
    ELSIF p_trust_score < 0.8 THEN
        v_security_score := v_security_score - 10;
    END IF;
    
    -- Deduct points for low confidence level
    IF p_confidence_level < 0.5 THEN
        v_security_score := v_security_score - 15;
    ELSIF p_confidence_level < 0.8 THEN
        v_security_score := v_security_score - 5;
    END IF;
    
    -- Ensure security score is between 0 and 100
    v_security_score := GREATEST(0, LEAST(100, v_security_score));
    
    -- Calculate overall score
    v_overall_score := (v_performance_score + v_security_score) / 2;
    
    -- Determine performance category
    IF v_performance_score >= 90 THEN
        v_performance_category := 'excellent';
    ELSIF v_performance_score >= 75 THEN
        v_performance_category := 'good';
    ELSIF v_performance_score >= 60 THEN
        v_performance_category := 'fair';
    ELSIF v_performance_score >= 40 THEN
        v_performance_category := 'poor';
    ELSE
        v_performance_category := 'critical';
    END IF;
    
    -- Determine security category
    IF p_security_violation OR p_threat_detected THEN
        v_security_category := 'at_risk';
    ELSIF p_compliance_violation OR p_vulnerability_found THEN
        v_security_category := 'monitored';
    ELSIF p_trust_score < 0.8 OR p_confidence_level < 0.8 THEN
        v_security_category := 'compliant';
    ELSE
        v_security_category := 'secure';
    END IF;
    
    -- Determine risk level
    IF p_security_violation OR p_threat_detected THEN
        v_risk_level := 'critical';
    ELSIF p_compliance_violation OR p_vulnerability_found THEN
        v_risk_level := 'high';
    ELSIF p_trust_score < 0.8 OR p_confidence_level < 0.8 THEN
        v_risk_level := 'medium';
    ELSE
        v_risk_level := 'low';
    END IF;
    
    -- Generate recommendations
    v_recommendations := ARRAY[]::TEXT[];
    
    IF p_execution_time_ms > v_slow_validation_threshold THEN
        v_recommendations := array_append(v_recommendations, 'Optimize validation algorithm performance');
        v_recommendations := array_append(v_recommendations, 'Consider caching validation results');
    END IF;
    
    IF NOT p_success THEN
        v_recommendations := array_append(v_recommendations, 'Review validation logic for potential issues');
        v_recommendations := array_append(v_recommendations, 'Add proper error handling');
    END IF;
    
    IF p_security_violation THEN
        v_recommendations := array_append(v_recommendations, 'Immediately investigate the security violation');
        v_recommendations := array_append(v_recommendations, 'Review security policies and validation criteria');
    END IF;
    
    IF p_compliance_violation THEN
        v_recommendations := array_append(v_recommendations, 'Review compliance requirements');
        v_recommendations := array_append(v_recommendations, 'Update validation logic to meet compliance standards');
    END IF;
    
    IF p_threat_detected THEN
        v_recommendations := array_append(v_recommendations, 'Implement threat response procedures');
        v_recommendations := array_append(v_recommendations, 'Review and update threat detection rules');
    END IF;
    
    IF p_vulnerability_found THEN
        v_recommendations := array_append(v_recommendations, 'Assess vulnerability severity and impact');
        v_recommendations := array_append(v_recommendations, 'Implement appropriate remediation measures');
    END IF;
    
    IF p_trust_score < 0.8 THEN
        v_recommendations := array_append(v_recommendations, 'Review data source trustworthiness');
        v_recommendations := array_append(v_recommendations, 'Implement additional trust validation');
    END IF;
    
    IF p_confidence_level < 0.8 THEN
        v_recommendations := array_append(v_recommendations, 'Improve validation accuracy');
        v_recommendations := array_append(v_recommendations, 'Add additional validation checks');
    END IF;
    
    -- Return results
    RETURN QUERY SELECT 
        v_validation_id,
        v_performance_score,
        v_security_score,
        v_overall_score,
        v_performance_category,
        v_security_category,
        v_risk_level,
        v_recommendations;
END;
$$ LANGUAGE plpgsql;

-- 7. Create function to collect security system health metrics
CREATE OR REPLACE FUNCTION collect_security_system_health()
RETURNS TABLE (
    overall_security_score DECIMAL,
    overall_performance_score DECIMAL,
    overall_risk_level VARCHAR,
    active_threats INTEGER,
    security_violations INTEGER,
    compliance_violations INTEGER,
    vulnerabilities INTEGER,
    slow_validations INTEGER,
    failed_validations INTEGER,
    high_risk_validations INTEGER,
    trust_score_average DECIMAL,
    confidence_level_average DECIMAL,
    validation_count INTEGER,
    success_rate DECIMAL,
    failure_rate DECIMAL,
    average_execution_time_ms DECIMAL
) AS $$
DECLARE
    v_overall_security_score DECIMAL;
    v_overall_performance_score DECIMAL;
    v_overall_risk_level VARCHAR;
    v_active_threats INTEGER;
    v_security_violations INTEGER;
    v_compliance_violations INTEGER;
    v_vulnerabilities INTEGER;
    v_slow_validations INTEGER;
    v_failed_validations INTEGER;
    v_high_risk_validations INTEGER;
    v_trust_score_average DECIMAL;
    v_confidence_level_average DECIMAL;
    v_validation_count INTEGER;
    v_success_rate DECIMAL;
    v_failure_rate DECIMAL;
    v_average_execution_time_ms DECIMAL;
    v_total_validations INTEGER;
    v_total_success INTEGER;
    v_total_failures INTEGER;
BEGIN
    -- Get validation counts and statistics
    SELECT 
        count(*),
        sum(success_count),
        sum(failure_count),
        sum(security_violation_count),
        sum(compliance_violation_count),
        sum(threat_detection_count),
        sum(vulnerability_count),
        count(*) FILTER (WHERE performance_category IN ('poor', 'critical')),
        count(*) FILTER (WHERE risk_level IN ('high', 'critical')),
        avg(trust_score),
        avg(confidence_level),
        avg(average_execution_time_ms)
    INTO 
        v_total_validations,
        v_total_success,
        v_total_failures,
        v_security_violations,
        v_compliance_violations,
        v_active_threats,
        v_vulnerabilities,
        v_slow_validations,
        v_high_risk_validations,
        v_trust_score_average,
        v_confidence_level_average,
        v_average_execution_time_ms
    FROM security_validation_performance_log
    WHERE last_executed >= now() - interval '24 hours';
    
    -- Set counts
    v_validation_count := v_total_validations;
    v_failed_validations := v_total_failures;
    
    -- Calculate rates
    IF v_total_validations > 0 THEN
        v_success_rate := (v_total_success::DECIMAL / v_total_validations) * 100;
        v_failure_rate := (v_total_failures::DECIMAL / v_total_validations) * 100;
    ELSE
        v_success_rate := 0;
        v_failure_rate := 0;
    END IF;
    
    -- Calculate overall scores (simplified calculation)
    v_overall_performance_score := 100.0;
    IF v_total_validations > 0 THEN
        -- Deduct points for slow validations
        v_overall_performance_score := v_overall_performance_score - ((v_slow_validations::DECIMAL / v_total_validations) * 30);
        -- Deduct points for failed validations
        v_overall_performance_score := v_overall_performance_score - ((v_total_failures::DECIMAL / v_total_validations) * 50);
        -- Deduct points for slow average execution time
        IF v_average_execution_time_ms > 200 THEN
            v_overall_performance_score := v_overall_performance_score - ((v_average_execution_time_ms - 200) / 10);
        END IF;
    END IF;
    v_overall_performance_score := GREATEST(0, LEAST(100, v_overall_performance_score));
    
    v_overall_security_score := 100.0;
    IF v_total_validations > 0 THEN
        -- Deduct points for security violations
        v_overall_security_score := v_overall_security_score - ((v_security_violations::DECIMAL / v_total_validations) * 100);
        -- Deduct points for compliance violations
        v_overall_security_score := v_overall_security_score - ((v_compliance_violations::DECIMAL / v_total_validations) * 80);
        -- Deduct points for active threats
        v_overall_security_score := v_overall_security_score - ((v_active_threats::DECIMAL / v_total_validations) * 90);
        -- Deduct points for vulnerabilities
        v_overall_security_score := v_overall_security_score - ((v_vulnerabilities::DECIMAL / v_total_validations) * 70);
        -- Deduct points for high risk validations
        v_overall_security_score := v_overall_security_score - ((v_high_risk_validations::DECIMAL / v_total_validations) * 60);
    END IF;
    v_overall_security_score := GREATEST(0, LEAST(100, v_overall_security_score));
    
    -- Determine overall risk level
    IF v_security_violations > 0 OR v_active_threats > 0 THEN
        v_overall_risk_level := 'critical';
    ELSIF v_compliance_violations > 0 OR v_vulnerabilities > 0 THEN
        v_overall_risk_level := 'high';
    ELSIF v_high_risk_validations > 0 THEN
        v_overall_risk_level := 'medium';
    ELSE
        v_overall_risk_level := 'low';
    END IF;
    
    -- Return results
    RETURN QUERY SELECT 
        v_overall_security_score,
        v_overall_performance_score,
        v_overall_risk_level,
        v_active_threats,
        v_security_violations,
        v_compliance_violations,
        v_vulnerabilities,
        v_slow_validations,
        v_failed_validations,
        v_high_risk_validations,
        v_trust_score_average,
        v_confidence_level_average,
        v_validation_count,
        v_success_rate,
        v_failure_rate,
        v_average_execution_time_ms;
END;
$$ LANGUAGE plpgsql;

-- 8. Create function to get security performance dashboard data
CREATE OR REPLACE FUNCTION get_security_performance_dashboard(
    p_hours_back INTEGER DEFAULT 24
) RETURNS TABLE (
    metric_name TEXT,
    metric_value DECIMAL,
    metric_unit TEXT,
    metric_category TEXT,
    status TEXT,
    trend TEXT,
    recommendations TEXT[]
) AS $$
DECLARE
    v_start_time TIMESTAMP WITH TIME ZONE;
    v_avg_execution_time DECIMAL;
    v_slow_validation_count INTEGER;
    v_failure_rate DECIMAL;
    v_security_violation_count INTEGER;
    v_compliance_violation_count INTEGER;
    v_threat_count INTEGER;
    v_vulnerability_count INTEGER;
    v_trust_score_avg DECIMAL;
    v_confidence_level_avg DECIMAL;
    v_overall_security_score DECIMAL;
    v_overall_performance_score DECIMAL;
BEGIN
    v_start_time := now() - (p_hours_back || ' hours')::INTERVAL;
    
    -- Get average execution time
    SELECT COALESCE(avg(average_execution_time_ms), 0) INTO v_avg_execution_time
    FROM security_validation_performance_log 
    WHERE last_executed >= v_start_time;
    
    -- Get slow validation count
    SELECT count(*)::INTEGER INTO v_slow_validation_count
    FROM security_validation_performance_log 
    WHERE last_executed >= v_start_time 
    AND average_execution_time_ms > 200;
    
    -- Get failure rate
    SELECT 
        CASE 
            WHEN sum(execution_count) > 0 
            THEN round((sum(failure_count)::DECIMAL / sum(execution_count)) * 100, 2)
            ELSE 0 
        END
    INTO v_failure_rate
    FROM security_validation_performance_log 
    WHERE last_executed >= v_start_time;
    
    -- Get security violation count
    SELECT sum(security_violation_count)::INTEGER INTO v_security_violation_count
    FROM security_validation_performance_log 
    WHERE last_executed >= v_start_time;
    
    -- Get compliance violation count
    SELECT sum(compliance_violation_count)::INTEGER INTO v_compliance_violation_count
    FROM security_validation_performance_log 
    WHERE last_executed >= v_start_time;
    
    -- Get threat count
    SELECT sum(threat_detection_count)::INTEGER INTO v_threat_count
    FROM security_validation_performance_log 
    WHERE last_executed >= v_start_time;
    
    -- Get vulnerability count
    SELECT sum(vulnerability_count)::INTEGER INTO v_vulnerability_count
    FROM security_validation_performance_log 
    WHERE last_executed >= v_start_time;
    
    -- Get average trust score
    SELECT COALESCE(avg(trust_score), 0) INTO v_trust_score_avg
    FROM security_validation_performance_log 
    WHERE last_executed >= v_start_time;
    
    -- Get average confidence level
    SELECT COALESCE(avg(confidence_level), 0) INTO v_confidence_level_avg
    FROM security_validation_performance_log 
    WHERE last_executed >= v_start_time;
    
    -- Get overall scores from system health
    SELECT 
        COALESCE(avg(overall_security_score), 0),
        COALESCE(avg(overall_performance_score), 0)
    INTO v_overall_security_score, v_overall_performance_score
    FROM security_system_health 
    WHERE timestamp >= v_start_time;
    
    -- Return dashboard metrics
    RETURN QUERY
    SELECT 'Average Validation Execution Time'::TEXT, v_avg_execution_time, 'ms'::TEXT, 'Performance'::TEXT,
           CASE WHEN v_avg_execution_time < 100 THEN 'OK'::TEXT
                WHEN v_avg_execution_time < 200 THEN 'WARNING'::TEXT
                ELSE 'CRITICAL'::TEXT END,
           'stable'::TEXT,
           CASE WHEN v_avg_execution_time > 200 THEN ARRAY['Optimize validation algorithms', 'Implement caching']::TEXT[]
                ELSE ARRAY[]::TEXT[] END
    
    UNION ALL
    
    SELECT 'Slow Validation Count'::TEXT, v_slow_validation_count::DECIMAL, 'validations'::TEXT, 'Performance'::TEXT,
           CASE WHEN v_slow_validation_count < 5 THEN 'OK'::TEXT
                WHEN v_slow_validation_count < 20 THEN 'WARNING'::TEXT
                ELSE 'CRITICAL'::TEXT END,
           'stable'::TEXT,
           CASE WHEN v_slow_validation_count > 5 THEN ARRAY['Review slow validations', 'Optimize performance']::TEXT[]
                ELSE ARRAY[]::TEXT[] END
    
    UNION ALL
    
    SELECT 'Validation Failure Rate'::TEXT, v_failure_rate, '%'::TEXT, 'Reliability'::TEXT,
           CASE WHEN v_failure_rate < 1 THEN 'OK'::TEXT
                WHEN v_failure_rate < 5 THEN 'WARNING'::TEXT
                ELSE 'CRITICAL'::TEXT END,
           'stable'::TEXT,
           CASE WHEN v_failure_rate > 1 THEN ARRAY['Investigate validation failures', 'Improve error handling']::TEXT[]
                ELSE ARRAY[]::TEXT[] END
    
    UNION ALL
    
    SELECT 'Security Violations'::TEXT, v_security_violation_count::DECIMAL, 'violations'::TEXT, 'Security'::TEXT,
           CASE WHEN v_security_violation_count = 0 THEN 'OK'::TEXT
                WHEN v_security_violation_count < 3 THEN 'WARNING'::TEXT
                ELSE 'CRITICAL'::TEXT END,
           'stable'::TEXT,
           CASE WHEN v_security_violation_count > 0 THEN ARRAY['Investigate security violations', 'Review security policies']::TEXT[]
                ELSE ARRAY[]::TEXT[] END
    
    UNION ALL
    
    SELECT 'Compliance Violations'::TEXT, v_compliance_violation_count::DECIMAL, 'violations'::TEXT, 'Compliance'::TEXT,
           CASE WHEN v_compliance_violation_count = 0 THEN 'OK'::TEXT
                WHEN v_compliance_violation_count < 5 THEN 'WARNING'::TEXT
                ELSE 'CRITICAL'::TEXT END,
           'stable'::TEXT,
           CASE WHEN v_compliance_violation_count > 0 THEN ARRAY['Review compliance requirements', 'Update validation logic']::TEXT[]
                ELSE ARRAY[]::TEXT[] END
    
    UNION ALL
    
    SELECT 'Threats Detected'::TEXT, v_threat_count::DECIMAL, 'threats'::TEXT, 'Security'::TEXT,
           CASE WHEN v_threat_count = 0 THEN 'OK'::TEXT
                WHEN v_threat_count < 2 THEN 'WARNING'::TEXT
                ELSE 'CRITICAL'::TEXT END,
           'stable'::TEXT,
           CASE WHEN v_threat_count > 0 THEN ARRAY['Implement threat response', 'Update threat detection']::TEXT[]
                ELSE ARRAY[]::TEXT[] END
    
    UNION ALL
    
    SELECT 'Vulnerabilities Found'::TEXT, v_vulnerability_count::DECIMAL, 'vulnerabilities'::TEXT, 'Security'::TEXT,
           CASE WHEN v_vulnerability_count = 0 THEN 'OK'::TEXT
                WHEN v_vulnerability_count < 3 THEN 'WARNING'::TEXT
                ELSE 'CRITICAL'::TEXT END,
           'stable'::TEXT,
           CASE WHEN v_vulnerability_count > 0 THEN ARRAY['Assess vulnerability impact', 'Implement remediation']::TEXT[]
                ELSE ARRAY[]::TEXT[] END
    
    UNION ALL
    
    SELECT 'Average Trust Score'::TEXT, v_trust_score_avg, 'score'::TEXT, 'Security'::TEXT,
           CASE WHEN v_trust_score_avg > 0.9 THEN 'OK'::TEXT
                WHEN v_trust_score_avg > 0.8 THEN 'WARNING'::TEXT
                ELSE 'CRITICAL'::TEXT END,
           'stable'::TEXT,
           CASE WHEN v_trust_score_avg < 0.8 THEN ARRAY['Review data source trust', 'Implement trust validation']::TEXT[]
                ELSE ARRAY[]::TEXT[] END
    
    UNION ALL
    
    SELECT 'Average Confidence Level'::TEXT, v_confidence_level_avg, 'level'::TEXT, 'Quality'::TEXT,
           CASE WHEN v_confidence_level_avg > 0.9 THEN 'OK'::TEXT
                WHEN v_confidence_level_avg > 0.8 THEN 'WARNING'::TEXT
                ELSE 'CRITICAL'::TEXT END,
           'stable'::TEXT,
           CASE WHEN v_confidence_level_avg < 0.8 THEN ARRAY['Improve validation accuracy', 'Add validation checks']::TEXT[]
                ELSE ARRAY[]::TEXT[] END
    
    UNION ALL
    
    SELECT 'Overall Security Score'::TEXT, v_overall_security_score, 'score'::TEXT, 'Security'::TEXT,
           CASE WHEN v_overall_security_score > 90 THEN 'OK'::TEXT
                WHEN v_overall_security_score > 75 THEN 'WARNING'::TEXT
                ELSE 'CRITICAL'::TEXT END,
           'stable'::TEXT,
           CASE WHEN v_overall_security_score < 75 THEN ARRAY['Review security posture', 'Implement security improvements']::TEXT[]
                ELSE ARRAY[]::TEXT[] END
    
    UNION ALL
    
    SELECT 'Overall Performance Score'::TEXT, v_overall_performance_score, 'score'::TEXT, 'Performance'::TEXT,
           CASE WHEN v_overall_performance_score > 90 THEN 'OK'::TEXT
                WHEN v_overall_performance_score > 75 THEN 'WARNING'::TEXT
                ELSE 'CRITICAL'::TEXT END,
           'stable'::TEXT,
           CASE WHEN v_overall_performance_score < 75 THEN ARRAY['Optimize performance', 'Review slow operations']::TEXT[]
                ELSE ARRAY[]::TEXT[] END;
END;
$$ LANGUAGE plpgsql;

-- 9. Create view for security validation summary
CREATE OR REPLACE VIEW security_validation_summary AS
SELECT 
    validation_id,
    validation_type,
    validation_name,
    execution_count,
    average_execution_time_ms,
    performance_category,
    security_category,
    risk_level,
    last_executed,
    success_count,
    failure_count,
    security_violation_count,
    compliance_violation_count,
    threat_detection_count,
    vulnerability_count,
    trust_score,
    confidence_level,
    CASE 
        WHEN execution_count > 0 
        THEN round((failure_count::DECIMAL / execution_count) * 100, 2)
        ELSE 0 
    END as failure_rate_percent,
    CASE 
        WHEN execution_count > 0 
        THEN round((success_count::DECIMAL / execution_count) * 100, 2)
        ELSE 0 
    END as success_rate_percent
FROM security_validation_performance_log
WHERE last_executed >= now() - interval '24 hours'
ORDER BY 
    CASE risk_level 
        WHEN 'critical' THEN 1
        WHEN 'high' THEN 2
        WHEN 'medium' THEN 3
        WHEN 'low' THEN 4
    END,
    average_execution_time_ms DESC;

-- 10. Create view for active security alerts
CREATE OR REPLACE VIEW active_security_alerts AS
SELECT 
    id,
    timestamp,
    alert_type,
    severity,
    validation_type,
    validation_id,
    validation_name,
    message,
    security_impact,
    recommendations,
    metadata
FROM security_validation_alerts
WHERE resolved = FALSE
ORDER BY 
    CASE severity 
        WHEN 'critical' THEN 1
        WHEN 'high' THEN 2
        WHEN 'medium' THEN 3
        WHEN 'low' THEN 4
    END,
    timestamp DESC;

-- 11. Create view for security performance metrics summary
CREATE OR REPLACE VIEW security_performance_metrics_summary AS
SELECT 
    spm.validation_type,
    spm.validation_name,
    spm.performance_score,
    spm.security_score,
    spm.overall_score,
    spm.risk_level,
    spm.execution_time_ms,
    spm.success_rate,
    spm.failure_rate,
    spm.security_violation_rate,
    spm.compliance_violation_rate,
    spm.threat_detection_rate,
    spm.vulnerability_rate,
    spm.trust_score,
    spm.confidence_level,
    spm.timestamp
FROM security_performance_metrics spm
WHERE spm.timestamp >= now() - interval '7 days'
ORDER BY spm.overall_score ASC, spm.timestamp DESC;

-- 12. Create function to clean up old security data
CREATE OR REPLACE FUNCTION cleanup_old_security_data(
    p_days_to_keep INTEGER DEFAULT 30
) RETURNS TABLE (
    table_name TEXT,
    records_deleted BIGINT
) AS $$
DECLARE
    v_cutoff_date TIMESTAMP WITH TIME ZONE;
    v_deleted_count BIGINT;
BEGIN
    v_cutoff_date := now() - (p_days_to_keep || ' days')::INTERVAL;
    
    -- Clean up old validation performance logs
    DELETE FROM security_validation_performance_log WHERE created_at < v_cutoff_date;
    GET DIAGNOSTICS v_deleted_count = ROW_COUNT;
    RETURN QUERY SELECT 'security_validation_performance_log'::TEXT, v_deleted_count;
    
    -- Clean up old resolved alerts
    DELETE FROM security_validation_alerts 
    WHERE resolved = TRUE AND resolved_at < v_cutoff_date;
    GET DIAGNOSTICS v_deleted_count = ROW_COUNT;
    RETURN QUERY SELECT 'security_validation_alerts'::TEXT, v_deleted_count;
    
    -- Clean up old performance metrics
    DELETE FROM security_performance_metrics WHERE timestamp < v_cutoff_date;
    GET DIAGNOSTICS v_deleted_count = ROW_COUNT;
    RETURN QUERY SELECT 'security_performance_metrics'::TEXT, v_deleted_count;
    
    -- Clean up old system health data
    DELETE FROM security_system_health WHERE timestamp < v_cutoff_date;
    GET DIAGNOSTICS v_deleted_count = ROW_COUNT;
    RETURN QUERY SELECT 'security_system_health'::TEXT, v_deleted_count;
END;
$$ LANGUAGE plpgsql;

-- 13. Create trigger to update updated_at timestamp
CREATE TRIGGER update_security_validation_performance_log_updated_at
    BEFORE UPDATE ON security_validation_performance_log
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 14. Create function to get top security issues
CREATE OR REPLACE FUNCTION get_top_security_issues(
    p_limit INTEGER DEFAULT 10,
    p_hours_back INTEGER DEFAULT 24
) RETURNS TABLE (
    validation_id VARCHAR,
    validation_type VARCHAR,
    validation_name VARCHAR,
    risk_level VARCHAR,
    security_violation_count BIGINT,
    compliance_violation_count BIGINT,
    threat_detection_count BIGINT,
    vulnerability_count BIGINT,
    trust_score DECIMAL,
    confidence_level DECIMAL,
    last_executed TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        svpl.validation_id,
        svpl.validation_type,
        svpl.validation_name,
        svpl.risk_level,
        svpl.security_violation_count,
        svpl.compliance_violation_count,
        svpl.threat_detection_count,
        svpl.vulnerability_count,
        svpl.trust_score,
        svpl.confidence_level,
        svpl.last_executed
    FROM security_validation_performance_log svpl
    WHERE svpl.last_executed >= now() - (p_hours_back || ' hours')::INTERVAL
    AND (svpl.security_violation_count > 0 
         OR svpl.compliance_violation_count > 0 
         OR svpl.threat_detection_count > 0 
         OR svpl.vulnerability_count > 0
         OR svpl.risk_level IN ('high', 'critical'))
    ORDER BY 
        CASE svpl.risk_level 
            WHEN 'critical' THEN 1
            WHEN 'high' THEN 2
            WHEN 'medium' THEN 3
            WHEN 'low' THEN 4
        END,
        (svpl.security_violation_count + svpl.compliance_violation_count + 
         svpl.threat_detection_count + svpl.vulnerability_count) DESC
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

-- 15. Create function to get security trends
CREATE OR REPLACE FUNCTION get_security_trends(
    p_hours_back INTEGER DEFAULT 24,
    p_interval_hours INTEGER DEFAULT 1
) RETURNS TABLE (
    time_bucket TIMESTAMP WITH TIME ZONE,
    avg_execution_time_ms DECIMAL,
    validation_count BIGINT,
    success_count BIGINT,
    failure_count BIGINT,
    security_violation_count BIGINT,
    compliance_violation_count BIGINT,
    threat_count BIGINT,
    vulnerability_count BIGINT,
    avg_trust_score DECIMAL,
    avg_confidence_level DECIMAL
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        date_trunc('hour', svpl.last_executed) + 
        (extract(hour from svpl.last_executed)::INTEGER / p_interval_hours * p_interval_hours) * interval '1 hour' as time_bucket,
        avg(svpl.average_execution_time_ms) as avg_execution_time_ms,
        sum(svpl.execution_count) as validation_count,
        sum(svpl.success_count) as success_count,
        sum(svpl.failure_count) as failure_count,
        sum(svpl.security_violation_count) as security_violation_count,
        sum(svpl.compliance_violation_count) as compliance_violation_count,
        sum(svpl.threat_detection_count) as threat_count,
        sum(svpl.vulnerability_count) as vulnerability_count,
        avg(svpl.trust_score) as avg_trust_score,
        avg(svpl.confidence_level) as avg_confidence_level
    FROM security_validation_performance_log svpl
    WHERE svpl.last_executed >= now() - (p_hours_back || ' hours')::INTERVAL
    GROUP BY time_bucket
    ORDER BY time_bucket;
END;
$$ LANGUAGE plpgsql;
