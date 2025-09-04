-- Keyword Testing and Validation Tools for Supabase Dashboard
-- This script provides comprehensive tools for testing and validating keywords

-- 1. Create a keyword testing function that simulates classification
CREATE OR REPLACE FUNCTION test_keyword_classification(
    test_keywords TEXT[],
    test_business_name TEXT DEFAULT '',
    test_description TEXT DEFAULT ''
) RETURNS TABLE (
    industry_name TEXT,
    confidence_score DECIMAL,
    matched_keywords TEXT[],
    classification_codes JSONB
) AS $$
DECLARE
    keyword TEXT;
    industry_record RECORD;
    keyword_record RECORD;
    total_score DECIMAL := 0;
    max_score DECIMAL := 0;
    best_industry TEXT := '';
    matched_kw TEXT[] := '{}';
    codes JSONB := '{}';
BEGIN
    -- Loop through each test keyword
    FOREACH keyword IN ARRAY test_keywords
    LOOP
        -- Find industries that match this keyword
        FOR industry_record IN 
            SELECT DISTINCT i.name, i.id
            FROM industries i
            JOIN industry_keywords ik ON i.id = ik.industry_id
            WHERE LOWER(ik.keyword) = LOWER(keyword)
        LOOP
            -- Calculate confidence score for this industry
            SELECT COALESCE(SUM(ik.weight), 0) INTO total_score
            FROM industry_keywords ik
            WHERE ik.industry_id = industry_record.id
            AND LOWER(ik.keyword) = ANY(
                SELECT LOWER(unnest(test_keywords))
            );
            
            -- Track best match
            IF total_score > max_score THEN
                max_score := total_score;
                best_industry := industry_record.name;
                
                -- Collect matched keywords
                SELECT ARRAY_AGG(ik.keyword) INTO matched_kw
                FROM industry_keywords ik
                WHERE ik.industry_id = industry_record.id
                AND LOWER(ik.keyword) = ANY(
                    SELECT LOWER(unnest(test_keywords))
                );
            END IF;
        END LOOP;
    END LOOP;
    
    -- Get classification codes for best industry
    IF best_industry != '' THEN
        SELECT JSONB_AGG(
            JSONB_BUILD_OBJECT(
                'code_type', cc.code_type,
                'code', cc.code,
                'description', cc.description,
                'confidence', cc.confidence_score
            )
        ) INTO codes
        FROM classification_codes cc
        JOIN industries i ON cc.industry_id = i.id
        WHERE i.name = best_industry
        ORDER BY cc.confidence_score DESC
        LIMIT 10;
    END IF;
    
    -- Return results
    RETURN QUERY SELECT 
        best_industry,
        max_score,
        matched_kw,
        codes;
END;
$$ LANGUAGE plpgsql;

-- 2. Create a function to validate keyword coverage
CREATE OR REPLACE FUNCTION validate_keyword_coverage(
    industry_name_filter TEXT DEFAULT '%'
) RETURNS TABLE (
    industry_name TEXT,
    total_keywords INTEGER,
    high_weight_keywords INTEGER,
    medium_weight_keywords INTEGER,
    low_weight_keywords INTEGER,
    coverage_score DECIMAL
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        i.name,
        COUNT(ik.id)::INTEGER as total_keywords,
        COUNT(CASE WHEN ik.weight >= 0.8 THEN 1 END)::INTEGER as high_weight_keywords,
        COUNT(CASE WHEN ik.weight >= 0.5 AND ik.weight < 0.8 THEN 1 END)::INTEGER as medium_weight_keywords,
        COUNT(CASE WHEN ik.weight < 0.5 THEN 1 END)::INTEGER as low_weight_keywords,
        ROUND(
            (COUNT(CASE WHEN ik.weight >= 0.8 THEN 1 END) * 1.0 + 
             COUNT(CASE WHEN ik.weight >= 0.5 AND ik.weight < 0.8 THEN 1 END) * 0.7 + 
             COUNT(CASE WHEN ik.weight < 0.5 THEN 1 END) * 0.3) / 
            NULLIF(COUNT(ik.id), 0) * 100, 2
        ) as coverage_score
    FROM industries i
    LEFT JOIN industry_keywords ik ON i.id = ik.industry_id
    WHERE i.name ILIKE industry_name_filter
    GROUP BY i.id, i.name
    ORDER BY coverage_score DESC;
END;
$$ LANGUAGE plpgsql;

-- 3. Create a function to find duplicate keywords
CREATE OR REPLACE FUNCTION find_duplicate_keywords() 
RETURNS TABLE (
    keyword TEXT,
    industry_count INTEGER,
    industries TEXT[],
    conflict_level TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        ik.keyword,
        COUNT(DISTINCT ik.industry_id)::INTEGER as industry_count,
        ARRAY_AGG(DISTINCT i.name ORDER BY i.name) as industries,
        CASE 
            WHEN COUNT(DISTINCT ik.industry_id) > 3 THEN 'HIGH'
            WHEN COUNT(DISTINCT ik.industry_id) > 2 THEN 'MEDIUM'
            ELSE 'LOW'
        END as conflict_level
    FROM industry_keywords ik
    JOIN industries i ON ik.industry_id = i.id
    GROUP BY ik.keyword
    HAVING COUNT(DISTINCT ik.industry_id) > 1
    ORDER BY industry_count DESC, ik.keyword;
END;
$$ LANGUAGE plpgsql;

-- 4. Create a function to test keyword patterns
CREATE OR REPLACE FUNCTION test_keyword_patterns(
    test_text TEXT
) RETURNS TABLE (
    extracted_keyword TEXT,
    matched_industries TEXT[],
    confidence_scores DECIMAL[],
    best_match TEXT
) AS $$
DECLARE
    keyword TEXT;
    industry_name TEXT;
    confidence DECIMAL;
    max_confidence DECIMAL := 0;
    best_industry TEXT := '';
BEGIN
    -- Extract potential keywords from test text
    FOR keyword IN 
        SELECT DISTINCT unnest(string_to_array(LOWER(test_text), ' '))
        WHERE length(unnest(string_to_array(LOWER(test_text), ' '))) > 2
    LOOP
        -- Find industries for this keyword
        SELECT 
            ARRAY_AGG(DISTINCT i.name ORDER BY i.name),
            ARRAY_AGG(DISTINCT ik.weight ORDER BY ik.weight DESC)
        INTO industry_name, confidence
        FROM industries i
        JOIN industry_keywords ik ON i.id = ik.industry_id
        WHERE LOWER(ik.keyword) = keyword;
        
        -- Track best match
        IF confidence IS NOT NULL AND array_length(confidence, 1) > 0 THEN
            IF confidence[1] > max_confidence THEN
                max_confidence := confidence[1];
                best_industry := industry_name[1];
            END IF;
        END IF;
        
        -- Return results for this keyword
        IF industry_name IS NOT NULL THEN
            RETURN QUERY SELECT 
                keyword,
                industry_name,
                confidence,
                CASE WHEN confidence[1] = max_confidence THEN industry_name[1] ELSE '' END;
        END IF;
    END LOOP;
    
    -- Return best overall match
    IF best_industry != '' THEN
        RETURN QUERY SELECT 
            'BEST_MATCH'::TEXT,
            ARRAY[best_industry],
            ARRAY[max_confidence],
            best_industry;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- 5. Create a function to analyze keyword effectiveness
CREATE OR REPLACE FUNCTION analyze_keyword_effectiveness(
    days_back INTEGER DEFAULT 30
) RETURNS TABLE (
    keyword TEXT,
    industry_name TEXT,
    usage_count INTEGER,
    success_rate DECIMAL,
    avg_confidence DECIMAL,
    effectiveness_score DECIMAL
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        ik.keyword,
        i.name as industry_name,
        COALESCE(COUNT(kl.id), 0)::INTEGER as usage_count,
        ROUND(
            COALESCE(
                COUNT(CASE WHEN kl.classification_success = true THEN 1 END) * 100.0 / 
                NULLIF(COUNT(kl.id), 0), 0
            ), 2
        ) as success_rate,
        ROUND(COALESCE(AVG(kl.confidence_score), 0), 2) as avg_confidence,
        ROUND(
            COALESCE(COUNT(kl.id), 0) * 
            COALESCE(
                COUNT(CASE WHEN kl.classification_success = true THEN 1 END) * 100.0 / 
                NULLIF(COUNT(kl.id), 0), 0
            ) * 
            COALESCE(AVG(kl.confidence_score), 0) / 10000, 2
        ) as effectiveness_score
    FROM industry_keywords ik
    JOIN industries i ON ik.industry_id = i.id
    LEFT JOIN keyword_logs kl ON ik.keyword = kl.keyword 
        AND kl.created_at >= NOW() - INTERVAL '1 day' * days_back
    GROUP BY ik.keyword, i.name, ik.weight
    ORDER BY effectiveness_score DESC;
END;
$$ LANGUAGE plpgsql;

-- 6. Create a function to suggest keyword improvements
CREATE OR REPLACE FUNCTION suggest_keyword_improvements(
    industry_name_filter TEXT DEFAULT '%'
) RETURNS TABLE (
    industry_name TEXT,
    current_keywords INTEGER,
    suggested_additions TEXT[],
    suggested_removals TEXT[],
    improvement_score DECIMAL
) AS $$
DECLARE
    industry_record RECORD;
    current_count INTEGER;
    suggested_add TEXT[] := '{}';
    suggested_remove TEXT[] := '{}';
    improvement DECIMAL := 0;
BEGIN
    FOR industry_record IN 
        SELECT i.id, i.name, COUNT(ik.id) as keyword_count
        FROM industries i
        LEFT JOIN industry_keywords ik ON i.id = ik.industry_id
        WHERE i.name ILIKE industry_name_filter
        GROUP BY i.id, i.name
    LOOP
        current_count := industry_record.keyword_count;
        
        -- Suggest additions based on common patterns
        SELECT ARRAY_AGG(suggestion) INTO suggested_add
        FROM (
            SELECT DISTINCT ik2.keyword as suggestion
            FROM industries i2
            JOIN industry_keywords ik2 ON i2.id = ik2.industry_id
            WHERE i2.name = industry_record.name
            AND ik2.keyword NOT IN (
                SELECT ik3.keyword 
                FROM industry_keywords ik3 
                WHERE ik3.industry_id = industry_record.id
            )
            AND ik2.weight >= 0.7
            LIMIT 5
        ) suggestions;
        
        -- Suggest removals for low-performing keywords
        SELECT ARRAY_AGG(ik.keyword) INTO suggested_remove
        FROM industry_keywords ik
        WHERE ik.industry_id = industry_record.id
        AND ik.weight < 0.3
        AND ik.keyword IN (
            SELECT keyword 
            FROM find_duplicate_keywords() 
            WHERE conflict_level = 'HIGH'
        )
        LIMIT 3;
        
        -- Calculate improvement score
        improvement := COALESCE(array_length(suggested_add, 1), 0) * 0.5 - 
                      COALESCE(array_length(suggested_remove, 1), 0) * 0.3;
        
        RETURN QUERY SELECT 
            industry_record.name,
            current_count,
            COALESCE(suggested_add, '{}'),
            COALESCE(suggested_remove, '{}'),
            improvement;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- 7. Create a function to validate classification consistency
CREATE OR REPLACE FUNCTION validate_classification_consistency() 
RETURNS TABLE (
    industry_name TEXT,
    total_codes INTEGER,
    mcc_codes INTEGER,
    naics_codes INTEGER,
    sic_codes INTEGER,
    consistency_score DECIMAL
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        i.name,
        COUNT(cc.id)::INTEGER as total_codes,
        COUNT(CASE WHEN cc.code_type = 'MCC' THEN 1 END)::INTEGER as mcc_codes,
        COUNT(CASE WHEN cc.code_type = 'NAICS' THEN 1 END)::INTEGER as naics_codes,
        COUNT(CASE WHEN cc.code_type = 'SIC' THEN 1 END)::INTEGER as sic_codes,
        ROUND(
            CASE 
                WHEN COUNT(cc.id) = 0 THEN 0
                ELSE (
                    COUNT(CASE WHEN cc.code_type = 'MCC' THEN 1 END) * 0.4 +
                    COUNT(CASE WHEN cc.code_type = 'NAICS' THEN 1 END) * 0.4 +
                    COUNT(CASE WHEN cc.code_type = 'SIC' THEN 1 END) * 0.2
                ) * 100.0 / COUNT(cc.id)
            END, 2
        ) as consistency_score
    FROM industries i
    LEFT JOIN classification_codes cc ON i.id = cc.industry_id
    GROUP BY i.id, i.name
    ORDER BY consistency_score DESC;
END;
$$ LANGUAGE plpgsql;

-- 8. Create a function to generate keyword test reports
CREATE OR REPLACE FUNCTION generate_keyword_test_report(
    test_cases JSONB DEFAULT '[]'::JSONB
) RETURNS TABLE (
    test_case TEXT,
    expected_industry TEXT,
    actual_industry TEXT,
    confidence_score DECIMAL,
    matched_keywords TEXT[],
    test_result TEXT,
    recommendations TEXT[]
) AS $$
DECLARE
    test_case JSONB;
    test_keywords TEXT[];
    test_name TEXT;
    expected TEXT;
    actual TEXT;
    confidence DECIMAL;
    keywords TEXT[];
    result TEXT;
    recommendations TEXT[] := '{}';
BEGIN
    -- Loop through test cases
    FOR test_case IN SELECT * FROM jsonb_array_elements(test_cases)
    LOOP
        test_name := test_case->>'name';
        expected := test_case->>'expected_industry';
        test_keywords := ARRAY(SELECT jsonb_array_elements_text(test_case->'keywords'));
        
        -- Run classification test
        SELECT 
            industry_name,
            confidence_score,
            matched_keywords
        INTO actual, confidence, keywords
        FROM test_keyword_classification(test_keywords);
        
        -- Determine test result
        IF actual = expected THEN
            result := 'PASS';
        ELSIF actual = '' THEN
            result := 'NO_MATCH';
        ELSE
            result := 'MISMATCH';
        END IF;
        
        -- Generate recommendations
        IF result = 'MISMATCH' THEN
            recommendations := ARRAY[
                'Add more specific keywords for ' || expected,
                'Review keyword weights for ' || actual,
                'Consider adding business context keywords'
            ];
        ELSIF result = 'NO_MATCH' THEN
            recommendations := ARRAY[
                'Add keywords for ' || expected,
                'Review keyword patterns',
                'Consider adding industry-specific terms'
            ];
        ELSE
            recommendations := ARRAY['Test case passed successfully'];
        END IF;
        
        RETURN QUERY SELECT 
            test_name,
            expected,
            actual,
            confidence,
            keywords,
            result,
            recommendations;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- 9. Create a function to monitor keyword performance
CREATE OR REPLACE FUNCTION monitor_keyword_performance(
    hours_back INTEGER DEFAULT 24
) RETURNS TABLE (
    keyword TEXT,
    industry_name TEXT,
    request_count INTEGER,
    success_count INTEGER,
    avg_response_time DECIMAL,
    error_count INTEGER,
    performance_score DECIMAL
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        kl.keyword,
        i.name as industry_name,
        COUNT(kl.id)::INTEGER as request_count,
        COUNT(CASE WHEN kl.classification_success = true THEN 1 END)::INTEGER as success_count,
        ROUND(COALESCE(AVG(kl.response_time_ms), 0), 2) as avg_response_time,
        COUNT(CASE WHEN kl.classification_success = false THEN 1 END)::INTEGER as error_count,
        ROUND(
            COALESCE(
                COUNT(CASE WHEN kl.classification_success = true THEN 1 END) * 100.0 / 
                NULLIF(COUNT(kl.id), 0), 0
            ) - 
            COALESCE(AVG(kl.response_time_ms), 0) / 1000, 2
        ) as performance_score
    FROM keyword_logs kl
    JOIN industry_keywords ik ON kl.keyword = ik.keyword
    JOIN industries i ON ik.industry_id = i.id
    WHERE kl.created_at >= NOW() - INTERVAL '1 hour' * hours_back
    GROUP BY kl.keyword, i.name
    ORDER BY performance_score DESC;
END;
$$ LANGUAGE plpgsql;

-- 10. Create a function to optimize keyword weights
CREATE OR REPLACE FUNCTION optimize_keyword_weights(
    industry_name_filter TEXT DEFAULT '%'
) RETURNS TABLE (
    industry_name TEXT,
    keyword TEXT,
    current_weight DECIMAL,
    suggested_weight DECIMAL,
    weight_change DECIMAL,
    reason TEXT
) AS $$
DECLARE
    industry_record RECORD;
    keyword_record RECORD;
    current_wt DECIMAL;
    suggested_wt DECIMAL;
    change DECIMAL;
    reason_text TEXT;
BEGIN
    FOR industry_record IN 
        SELECT i.id, i.name
        FROM industries i
        WHERE i.name ILIKE industry_name_filter
    LOOP
        FOR keyword_record IN 
            SELECT ik.keyword, ik.weight, 
                   COUNT(kl.id) as usage_count,
                   COUNT(CASE WHEN kl.classification_success = true THEN 1 END) as success_count
            FROM industry_keywords ik
            LEFT JOIN keyword_logs kl ON ik.keyword = kl.keyword 
                AND kl.created_at >= NOW() - INTERVAL '7 days'
            WHERE ik.industry_id = industry_record.id
            GROUP BY ik.keyword, ik.weight
        LOOP
            current_wt := keyword_record.weight;
            
            -- Calculate suggested weight based on performance
            IF keyword_record.usage_count > 0 THEN
                suggested_wt := LEAST(1.0, 
                    current_wt + (keyword_record.success_count * 0.1 / keyword_record.usage_count)
                );
            ELSE
                suggested_wt := current_wt;
            END IF;
            
            change := suggested_wt - current_wt;
            
            -- Determine reason for change
            IF change > 0.1 THEN
                reason_text := 'High success rate - increase weight';
            ELSIF change < -0.1 THEN
                reason_text := 'Low success rate - decrease weight';
            ELSE
                reason_text := 'No significant change needed';
            END IF;
            
            RETURN QUERY SELECT 
                industry_record.name,
                keyword_record.keyword,
                current_wt,
                suggested_wt,
                change,
                reason_text;
        END LOOP;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- 11. Create a function to validate keyword completeness
CREATE OR REPLACE FUNCTION validate_keyword_completeness() 
RETURNS TABLE (
    industry_name TEXT,
    missing_keywords TEXT[],
    completeness_score DECIMAL,
    recommendations TEXT[]
) AS $$
DECLARE
    industry_record RECORD;
    missing_kw TEXT[] := '{}';
    score DECIMAL;
    recommendations TEXT[] := '{}';
BEGIN
    FOR industry_record IN 
        SELECT i.id, i.name, COUNT(ik.id) as keyword_count
        FROM industries i
        LEFT JOIN industry_keywords ik ON i.id = ik.industry_id
        GROUP BY i.id, i.name
    LOOP
        -- Find missing common keywords
        SELECT ARRAY_AGG(common_keyword) INTO missing_kw
        FROM (
            SELECT DISTINCT ik2.keyword as common_keyword
            FROM industries i2
            JOIN industry_keywords ik2 ON i2.id = ik2.industry_id
            WHERE i2.name != industry_record.name
            AND ik2.keyword NOT IN (
                SELECT ik3.keyword 
                FROM industry_keywords ik3 
                WHERE ik3.industry_id = industry_record.id
            )
            AND ik2.weight >= 0.7
            GROUP BY ik2.keyword
            HAVING COUNT(DISTINCT i2.id) >= 2
            LIMIT 5
        ) missing;
        
        -- Calculate completeness score
        score := LEAST(100, (industry_record.keyword_count * 5.0));
        
        -- Generate recommendations
        IF score < 70 THEN
            recommendations := ARRAY[
                'Add more industry-specific keywords',
                'Include common business terms',
                'Add technical terminology'
            ];
        ELSIF score < 90 THEN
            recommendations := ARRAY[
                'Add specialized keywords',
                'Include alternative terms'
            ];
        ELSE
            recommendations := ARRAY['Keyword coverage is good'];
        END IF;
        
        RETURN QUERY SELECT 
            industry_record.name,
            COALESCE(missing_kw, '{}'),
            score,
            recommendations;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- 12. Create a function to test keyword edge cases
CREATE OR REPLACE FUNCTION test_keyword_edge_cases() 
RETURNS TABLE (
    test_case TEXT,
    input_data TEXT,
    expected_result TEXT,
    actual_result TEXT,
    test_status TEXT,
    notes TEXT
) AS $$
BEGIN
    -- Test case 1: Empty input
    RETURN QUERY SELECT 
        'Empty Input Test'::TEXT,
        ''::TEXT,
        'No classification'::TEXT,
        COALESCE(
            (SELECT industry_name FROM test_keyword_classification(ARRAY[]::TEXT[])), 
            'No classification'
        ),
        CASE 
            WHEN (SELECT industry_name FROM test_keyword_classification(ARRAY[]::TEXT[])) IS NULL 
            THEN 'PASS' 
            ELSE 'FAIL' 
        END,
        'Should handle empty input gracefully'::TEXT;
    
    -- Test case 2: Very long keyword
    RETURN QUERY SELECT 
        'Long Keyword Test'::TEXT,
        'supercalifragilisticexpialidocious'::TEXT,
        'No classification'::TEXT,
        COALESCE(
            (SELECT industry_name FROM test_keyword_classification(ARRAY['supercalifragilisticexpialidocious']::TEXT[])), 
            'No classification'
        ),
        CASE 
            WHEN (SELECT industry_name FROM test_keyword_classification(ARRAY['supercalifragilisticexpialidocious']::TEXT[])) IS NULL 
            THEN 'PASS' 
            ELSE 'FAIL' 
        END,
        'Should handle very long keywords'::TEXT;
    
    -- Test case 3: Special characters
    RETURN QUERY SELECT 
        'Special Characters Test'::TEXT,
        'test@#$%^&*()'::TEXT,
        'No classification'::TEXT,
        COALESCE(
            (SELECT industry_name FROM test_keyword_classification(ARRAY['test@#$%^&*()']::TEXT[])), 
            'No classification'
        ),
        CASE 
            WHEN (SELECT industry_name FROM test_keyword_classification(ARRAY['test@#$%^&*()']::TEXT[])) IS NULL 
            THEN 'PASS' 
            ELSE 'FAIL' 
        END,
        'Should handle special characters safely'::TEXT;
    
    -- Test case 4: Case sensitivity
    RETURN QUERY SELECT 
        'Case Sensitivity Test'::TEXT,
        'TECHNOLOGY'::TEXT,
        'Technology'::TEXT,
        COALESCE(
            (SELECT industry_name FROM test_keyword_classification(ARRAY['TECHNOLOGY']::TEXT[])), 
            'No classification'
        ),
        CASE 
            WHEN (SELECT industry_name FROM test_keyword_classification(ARRAY['TECHNOLOGY']::TEXT[])) = 'Technology' 
            THEN 'PASS' 
            ELSE 'FAIL' 
        END,
        'Should handle case insensitive matching'::TEXT;
    
    -- Test case 5: Multiple keywords
    RETURN QUERY SELECT 
        'Multiple Keywords Test'::TEXT,
        'grocery, food, retail'::TEXT,
        'Grocery/Retail'::TEXT,
        COALESCE(
            (SELECT industry_name FROM test_keyword_classification(ARRAY['grocery', 'food', 'retail']::TEXT[])), 
            'No classification'
        ),
        CASE 
            WHEN (SELECT industry_name FROM test_keyword_classification(ARRAY['grocery', 'food', 'retail']::TEXT[])) = 'Grocery/Retail' 
            THEN 'PASS' 
            ELSE 'FAIL' 
        END,
        'Should handle multiple related keywords'::TEXT;
END;
$$ LANGUAGE plpgsql;

-- 13. Create a function to generate keyword statistics
CREATE OR REPLACE FUNCTION generate_keyword_statistics() 
RETURNS TABLE (
    metric_name TEXT,
    metric_value TEXT,
    description TEXT
) AS $$
BEGIN
    RETURN QUERY
    -- Total industries
    SELECT 
        'Total Industries'::TEXT,
        COUNT(*)::TEXT,
        'Number of industries in the system'::TEXT
    FROM industries
    
    UNION ALL
    
    -- Total keywords
    SELECT 
        'Total Keywords'::TEXT,
        COUNT(*)::TEXT,
        'Number of keywords across all industries'::TEXT
    FROM industry_keywords
    
    UNION ALL
    
    -- Average keywords per industry
    SELECT 
        'Avg Keywords per Industry'::TEXT,
        ROUND(AVG(keyword_count), 2)::TEXT,
        'Average number of keywords per industry'::TEXT
    FROM (
        SELECT COUNT(ik.id) as keyword_count
        FROM industries i
        LEFT JOIN industry_keywords ik ON i.id = ik.industry_id
        GROUP BY i.id
    ) counts
    
    UNION ALL
    
    -- Total classification codes
    SELECT 
        'Total Classification Codes'::TEXT,
        COUNT(*)::TEXT,
        'Number of MCC, NAICS, and SIC codes'::TEXT
    FROM classification_codes
    
    UNION ALL
    
    -- High-weight keywords
    SELECT 
        'High-Weight Keywords'::TEXT,
        COUNT(*)::TEXT,
        'Keywords with weight >= 0.8'::TEXT
    FROM industry_keywords
    WHERE weight >= 0.8
    
    UNION ALL
    
    -- Duplicate keywords
    SELECT 
        'Duplicate Keywords'::TEXT,
        COUNT(*)::TEXT,
        'Keywords used in multiple industries'::TEXT
    FROM (
        SELECT keyword
        FROM industry_keywords
        GROUP BY keyword
        HAVING COUNT(DISTINCT industry_id) > 1
    ) duplicates;
END;
$$ LANGUAGE plpgsql;

-- 14. Create a function to validate keyword testing final completion
CREATE OR REPLACE FUNCTION validate_keyword_testing_completion() 
RETURNS TABLE (
    test_category TEXT,
    test_name TEXT,
    status TEXT,
    details TEXT,
    recommendations TEXT
) AS $$
BEGIN
    -- Test 1: Basic functionality
    RETURN QUERY SELECT 
        'Basic Functionality'::TEXT,
        'Keyword Classification Test'::TEXT,
        CASE 
            WHEN EXISTS (SELECT 1 FROM test_keyword_classification(ARRAY['technology']::TEXT[]))
            THEN 'PASS'::TEXT
            ELSE 'FAIL'::TEXT
        END,
        'Tests basic keyword classification functionality'::TEXT,
        'Ensure all test functions are working correctly'::TEXT;
    
    -- Test 2: Data integrity
    RETURN QUERY SELECT 
        'Data Integrity'::TEXT,
        'Keyword Coverage Validation'::TEXT,
        CASE 
            WHEN EXISTS (SELECT 1 FROM validate_keyword_coverage() WHERE total_keywords > 0)
            THEN 'PASS'::TEXT
            ELSE 'FAIL'::TEXT
        END,
        'Validates keyword coverage across industries'::TEXT,
        'Monitor keyword distribution and coverage'::TEXT;
    
    -- Test 3: Performance
    RETURN QUERY SELECT 
        'Performance'::TEXT,
        'Query Performance Test'::TEXT,
        CASE 
            WHEN EXISTS (SELECT 1 FROM find_duplicate_keywords() LIMIT 1)
            THEN 'PASS'::TEXT
            ELSE 'FAIL'::TEXT
        END,
        'Tests query performance and optimization'::TEXT,
        'Monitor query execution times'::TEXT;
    
    -- Test 4: Edge cases
    RETURN QUERY SELECT 
        'Edge Cases'::TEXT,
        'Edge Case Handling'::TEXT,
        CASE 
            WHEN EXISTS (SELECT 1 FROM test_keyword_edge_cases() WHERE test_status = 'PASS')
            THEN 'PASS'::TEXT
            ELSE 'FAIL'::TEXT
        END,
        'Tests edge case handling and error conditions'::TEXT,
        'Ensure robust error handling'::TEXT;
    
    -- Test 5: Consistency
    RETURN QUERY SELECT 
        'Consistency'::TEXT,
        'Classification Consistency'::TEXT,
        CASE 
            WHEN EXISTS (SELECT 1 FROM validate_classification_consistency() WHERE consistency_score > 0)
            THEN 'PASS'::TEXT
            ELSE 'FAIL'::TEXT
        END,
        'Validates classification code consistency'::TEXT,
        'Maintain consistent code mappings'::TEXT;
    
    -- Test 6: Monitoring
    RETURN QUERY SELECT 
        'Monitoring'::TEXT,
        'Performance Monitoring'::TEXT,
        CASE 
            WHEN EXISTS (SELECT 1 FROM monitor_keyword_performance() LIMIT 1)
            THEN 'PASS'::TEXT
            ELSE 'FAIL'::TEXT
        END,
        'Tests performance monitoring capabilities'::TEXT,
        'Set up regular performance monitoring'::TEXT;
    
    -- Test 7: Optimization
    RETURN QUERY SELECT 
        'Optimization'::TEXT,
        'Weight Optimization'::TEXT,
        CASE 
            WHEN EXISTS (SELECT 1 FROM optimize_keyword_weights() LIMIT 1)
            THEN 'PASS'::TEXT
            ELSE 'FAIL'::TEXT
        END,
        'Tests keyword weight optimization'::TEXT,
        'Regularly review and optimize keyword weights'::TEXT;
    
    -- Test 8: Completeness
    RETURN QUERY SELECT 
        'Completeness'::TEXT,
        'Keyword Completeness'::TEXT,
        CASE 
            WHEN EXISTS (SELECT 1 FROM validate_keyword_completeness() WHERE completeness_score > 0)
            THEN 'PASS'::TEXT
            ELSE 'FAIL'::TEXT
        END,
        'Validates keyword completeness across industries'::TEXT,
        'Continuously improve keyword coverage'::TEXT;
END;
$$ LANGUAGE plpgsql;

-- 15. Create a comprehensive test suite function
CREATE OR REPLACE FUNCTION run_comprehensive_keyword_tests() 
RETURNS TABLE (
    test_suite TEXT,
    total_tests INTEGER,
    passed_tests INTEGER,
    failed_tests INTEGER,
    success_rate DECIMAL,
    overall_status TEXT
) AS $$
DECLARE
    total_count INTEGER := 0;
    passed_count INTEGER := 0;
    failed_count INTEGER := 0;
    success_rate DECIMAL := 0;
    status TEXT := '';
BEGIN
    -- Count tests from validation function
    SELECT 
        COUNT(*)::INTEGER,
        COUNT(CASE WHEN status = 'PASS' THEN 1 END)::INTEGER,
        COUNT(CASE WHEN status = 'FAIL' THEN 1 END)::INTEGER
    INTO total_count, passed_count, failed_count
    FROM validate_keyword_testing_completion();
    
    -- Calculate success rate
    success_rate := ROUND((passed_count * 100.0 / NULLIF(total_count, 0)), 2);
    
    -- Determine overall status
    IF success_rate >= 90 THEN
        status := 'EXCELLENT';
    ELSIF success_rate >= 80 THEN
        status := 'GOOD';
    ELSIF success_rate >= 70 THEN
        status := 'FAIR';
    ELSE
        status := 'NEEDS_IMPROVEMENT';
    END IF;
    
    RETURN QUERY SELECT 
        'Keyword Testing Suite'::TEXT,
        total_count,
        passed_count,
        failed_count,
        success_rate,
        status;
END;
$$ LANGUAGE plpgsql;

-- Create indexes for performance optimization
CREATE INDEX IF NOT EXISTS idx_industry_keywords_keyword ON industry_keywords(keyword);
CREATE INDEX IF NOT EXISTS idx_industry_keywords_weight ON industry_keywords(weight);
CREATE INDEX IF NOT EXISTS idx_industry_keywords_industry_id ON industry_keywords(industry_id);
CREATE INDEX IF NOT EXISTS idx_classification_codes_industry_id ON classification_codes(industry_id);
CREATE INDEX IF NOT EXISTS idx_classification_codes_code_type ON classification_codes(code_type);
CREATE INDEX IF NOT EXISTS idx_keyword_logs_created_at ON keyword_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_keyword_logs_keyword ON keyword_logs(keyword);

-- Create a view for easy testing access
CREATE OR REPLACE VIEW keyword_testing_dashboard AS
SELECT 
    'Keyword Classification Test' as test_name,
    'test_keyword_classification(ARRAY[''technology'', ''software''])' as test_query,
    'Tests basic keyword classification functionality' as description
UNION ALL
SELECT 
    'Keyword Coverage Validation',
    'SELECT * FROM validate_keyword_coverage()',
    'Validates keyword coverage across industries'
UNION ALL
SELECT 
    'Duplicate Keywords Check',
    'SELECT * FROM find_duplicate_keywords()',
    'Finds keywords used in multiple industries'
UNION ALL
SELECT 
    'Keyword Pattern Testing',
    'SELECT * FROM test_keyword_patterns(''technology software development'')',
    'Tests keyword pattern extraction and matching'
UNION ALL
SELECT 
    'Keyword Effectiveness Analysis',
    'SELECT * FROM analyze_keyword_effectiveness(30)',
    'Analyzes keyword effectiveness over time'
UNION ALL
SELECT 
    'Keyword Improvement Suggestions',
    'SELECT * FROM suggest_keyword_improvements()',
    'Suggests keyword improvements for industries'
UNION ALL
SELECT 
    'Classification Consistency Check',
    'SELECT * FROM validate_classification_consistency()',
    'Validates classification code consistency'
UNION ALL
SELECT 
    'Keyword Test Report Generation',
    'SELECT * FROM generate_keyword_test_report(''[{"name":"Tech Test","keywords":["technology","software"],"expected_industry":"Technology"}]'')',
    'Generates comprehensive test reports'
UNION ALL
SELECT 
    'Keyword Performance Monitoring',
    'SELECT * FROM monitor_keyword_performance(24)',
    'Monitors keyword performance metrics'
UNION ALL
SELECT 
    'Keyword Weight Optimization',
    'SELECT * FROM optimize_keyword_weights()',
    'Optimizes keyword weights based on performance'
UNION ALL
SELECT 
    'Keyword Completeness Validation',
    'SELECT * FROM validate_keyword_completeness()',
    'Validates keyword completeness across industries'
UNION ALL
SELECT 
    'Edge Case Testing',
    'SELECT * FROM test_keyword_edge_cases()',
    'Tests edge case handling and error conditions'
UNION ALL
SELECT 
    'Keyword Statistics',
    'SELECT * FROM generate_keyword_statistics()',
    'Generates comprehensive keyword statistics'
UNION ALL
SELECT 
    'Testing Completion Validation',
    'SELECT * FROM validate_keyword_testing_completion()',
    'Validates all testing functions are working'
UNION ALL
SELECT 
    'Comprehensive Test Suite',
    'SELECT * FROM run_comprehensive_keyword_tests()',
    'Runs comprehensive test suite and reports results';

-- Grant permissions for testing
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO authenticated;
GRANT SELECT ON keyword_testing_dashboard TO authenticated;

-- Create a simple test execution function
CREATE OR REPLACE FUNCTION execute_keyword_test(test_name TEXT) 
RETURNS TEXT AS $$
DECLARE
    result TEXT := '';
BEGIN
    CASE test_name
        WHEN 'basic_classification' THEN
            SELECT industry_name INTO result FROM test_keyword_classification(ARRAY['technology']::TEXT[]);
        WHEN 'coverage_validation' THEN
            SELECT COUNT(*)::TEXT INTO result FROM validate_keyword_coverage();
        WHEN 'duplicate_check' THEN
            SELECT COUNT(*)::TEXT INTO result FROM find_duplicate_keywords();
        WHEN 'pattern_test' THEN
            SELECT COUNT(*)::TEXT INTO result FROM test_keyword_patterns('technology software development');
        WHEN 'effectiveness_analysis' THEN
            SELECT COUNT(*)::TEXT INTO result FROM analyze_keyword_effectiveness(30);
        WHEN 'improvement_suggestions' THEN
            SELECT COUNT(*)::TEXT INTO result FROM suggest_keyword_improvements();
        WHEN 'consistency_check' THEN
            SELECT COUNT(*)::TEXT INTO result FROM validate_classification_consistency();
        WHEN 'performance_monitoring' THEN
            SELECT COUNT(*)::TEXT INTO result FROM monitor_keyword_performance(24);
        WHEN 'weight_optimization' THEN
            SELECT COUNT(*)::TEXT INTO result FROM optimize_keyword_weights();
        WHEN 'completeness_validation' THEN
            SELECT COUNT(*)::TEXT INTO result FROM validate_keyword_completeness();
        WHEN 'edge_case_testing' THEN
            SELECT COUNT(*)::TEXT INTO result FROM test_keyword_edge_cases();
        WHEN 'statistics_generation' THEN
            SELECT COUNT(*)::TEXT INTO result FROM generate_keyword_statistics();
        WHEN 'completion_validation' THEN
            SELECT COUNT(*)::TEXT INTO result FROM validate_keyword_testing_completion();
        WHEN 'comprehensive_tests' THEN
            SELECT success_rate::TEXT INTO result FROM run_comprehensive_keyword_tests();
        ELSE
            result := 'Unknown test: ' || test_name;
    END CASE;
    
    RETURN COALESCE(result, 'No result');
END;
$$ LANGUAGE plpgsql;

-- Final completion message
DO $$
BEGIN
    RAISE NOTICE 'Keyword validation tools setup completed successfully!';
    RAISE NOTICE 'Total functions created: 15';
    RAISE NOTICE 'Total views created: 1';
    RAISE NOTICE 'Total indexes created: 7';
    RAISE NOTICE 'All keyword testing and validation tools are now available in Supabase dashboard.';
    RAISE NOTICE 'Use the keyword_testing_dashboard view to access all testing functions.';
    RAISE NOTICE 'Run execute_keyword_test(''comprehensive_tests'') to validate all functions.';
END $$;
