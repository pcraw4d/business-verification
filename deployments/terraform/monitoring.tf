# KYB Platform - Monitoring and Alerting Configuration
# Comprehensive monitoring setup with CloudWatch, Prometheus, and Grafana

# CloudWatch Log Groups for Application Logging
resource "aws_cloudwatch_log_group" "application_logs" {
  name              = "/aws/eks/kyb-platform/application"
  retention_in_days = var.log_retention_days
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "monitoring"
  }
}

resource "aws_cloudwatch_log_group" "system_logs" {
  name              = "/aws/eks/kyb-platform/system"
  retention_in_days = var.log_retention_days
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "monitoring"
  }
}

resource "aws_cloudwatch_log_group" "security_logs" {
  name              = "/aws/eks/kyb-platform/security"
  retention_in_days = var.log_retention_days
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "monitoring"
  }
}

resource "aws_cloudwatch_log_group" "audit_logs" {
  name              = "/aws/eks/kyb-platform/audit"
  retention_in_days = var.log_retention_days
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "monitoring"
  }
}

# CloudWatch Metrics Filters for Log Analysis
resource "aws_cloudwatch_log_metric_filter" "error_logs" {
  name           = "kyb-platform-error-logs"
  pattern        = "[timestamp, level=ERROR, service, message]"
  log_group_name = aws_cloudwatch_log_group.application_logs.name
  
  metric_transformation {
    name      = "ErrorCount"
    namespace = "KYBPlatform/Application"
    value     = "1"
  }
}

resource "aws_cloudwatch_log_metric_filter" "security_events" {
  name           = "kyb-platform-security-events"
  pattern        = "[timestamp, level=WARN, service, message=*security*]"
  log_group_name = aws_cloudwatch_log_group.security_logs.name
  
  metric_transformation {
    name      = "SecurityEventCount"
    namespace = "KYBPlatform/Security"
    value     = "1"
  }
}

resource "aws_cloudwatch_log_metric_filter" "api_requests" {
  name           = "kyb-platform-api-requests"
  pattern        = "[timestamp, method, path, status_code, response_time]"
  log_group_name = aws_cloudwatch_log_group.application_logs.name
  
  metric_transformation {
    name      = "APIRequestCount"
    namespace = "KYBPlatform/API"
    value     = "1"
  }
}

# CloudWatch Alarms for Application Monitoring
resource "aws_cloudwatch_metric_alarm" "high_error_rate" {
  alarm_name          = "kyb-platform-high-error-rate"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "ErrorCount"
  namespace           = "KYBPlatform/Application"
  period              = "300"
  statistic           = "Sum"
  threshold           = var.error_rate_threshold
  alarm_description   = "This metric monitors application error rate"
  
  alarm_actions = [aws_sns_topic.critical_alerts.arn]
  ok_actions    = [aws_sns_topic.critical_alerts.arn]
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "monitoring"
    Severity    = "critical"
  }
}

resource "aws_cloudwatch_metric_alarm" "high_response_time" {
  alarm_name          = "kyb-platform-high-response-time"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "TargetResponseTime"
  namespace           = "AWS/ApplicationELB"
  period              = "300"
  statistic           = "Average"
  threshold           = var.response_time_threshold
  alarm_description   = "This metric monitors API response time"
  
  dimensions = {
    LoadBalancer = aws_lb.main.arn_suffix
    TargetGroup  = aws_lb_target_group.api.arn_suffix
  }
  
  alarm_actions = [aws_sns_topic.warning_alerts.arn]
  ok_actions    = [aws_sns_topic.warning_alerts.arn]
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "monitoring"
    Severity    = "warning"
  }
}

resource "aws_cloudwatch_metric_alarm" "database_connections" {
  alarm_name          = "kyb-platform-database-connections"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "DatabaseConnections"
  namespace           = "AWS/RDS"
  period              = "300"
  statistic           = "Average"
  threshold           = var.db_connection_threshold
  alarm_description   = "This metric monitors database connection count"
  
  dimensions = {
    DBClusterIdentifier = module.rds.db_cluster_id
  }
  
  alarm_actions = [aws_sns_topic.warning_alerts.arn]
  ok_actions    = [aws_sns_topic.warning_alerts.arn]
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "monitoring"
    Severity    = "warning"
  }
}

resource "aws_cloudwatch_metric_alarm" "redis_memory_usage" {
  alarm_name          = "kyb-platform-redis-memory-usage"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "DatabaseMemoryUsagePercentage"
  namespace           = "AWS/ElastiCache"
  period              = "300"
  statistic           = "Average"
  threshold           = var.redis_memory_threshold
  alarm_description   = "This metric monitors Redis memory usage"
  
  dimensions = {
    CacheClusterId = aws_elasticache_replication_group.redis.id
  }
  
  alarm_actions = [aws_sns_topic.warning_alerts.arn]
  ok_actions    = [aws_sns_topic.warning_alerts.arn]
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "monitoring"
    Severity    = "warning"
  }
}

resource "aws_cloudwatch_metric_alarm" "security_events" {
  alarm_name          = "kyb-platform-security-events"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "1"
  metric_name         = "SecurityEventCount"
  namespace           = "KYBPlatform/Security"
  period              = "300"
  statistic           = "Sum"
  threshold           = var.security_event_threshold
  alarm_description   = "This metric monitors security events"
  
  alarm_actions = [aws_sns_topic.critical_alerts.arn]
  ok_actions    = [aws_sns_topic.critical_alerts.arn]
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "monitoring"
    Severity    = "critical"
  }
}

# SNS Topics for Different Alert Severities
resource "aws_sns_topic" "critical_alerts" {
  name = "kyb-platform-critical-alerts"
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "notifications"
    Severity    = "critical"
  }
}

resource "aws_sns_topic" "warning_alerts" {
  name = "kyb-platform-warning-alerts"
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "notifications"
    Severity    = "warning"
  }
}

resource "aws_sns_topic" "info_alerts" {
  name = "kyb-platform-info-alerts"
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "notifications"
    Severity    = "info"
  }
}

# SNS Topic Subscriptions
resource "aws_sns_topic_subscription" "critical_email" {
  count     = length(var.critical_alert_emails)
  topic_arn = aws_sns_topic.critical_alerts.arn
  protocol  = "email"
  endpoint  = var.critical_alert_emails[count.index]
}

resource "aws_sns_topic_subscription" "warning_email" {
  count     = length(var.warning_alert_emails)
  topic_arn = aws_sns_topic.warning_alerts.arn
  protocol  = "email"
  endpoint  = var.warning_alert_emails[count.index]
}

resource "aws_sns_topic_subscription" "info_email" {
  count     = length(var.info_alert_emails)
  topic_arn = aws_sns_topic.info_alerts.arn
  protocol  = "email"
  endpoint  = var.info_alert_emails[count.index]
}

# Slack Integration (if webhook URL is provided)
resource "aws_sns_topic_subscription" "critical_slack" {
  count     = var.enable_slack_alerts ? 1 : 0
  topic_arn = aws_sns_topic.critical_alerts.arn
  protocol  = "https"
  endpoint  = var.slack_webhook_url
}

resource "aws_sns_topic_subscription" "warning_slack" {
  count     = var.enable_slack_alerts ? 1 : 0
  topic_arn = aws_sns_topic.warning_alerts.arn
  protocol  = "https"
  endpoint  = var.slack_webhook_url
}

# CloudWatch Dashboard for Application Monitoring
resource "aws_cloudwatch_dashboard" "application_monitoring" {
  dashboard_name = "kyb-platform-application-monitoring"
  
  dashboard_body = jsonencode({
    widgets = [
      {
        type   = "metric"
        x      = 0
        y      = 0
        width  = 12
        height = 6
        
        properties = {
          metrics = [
            ["KYBPlatform/Application", "ErrorCount"],
            [".", "APIRequestCount"]
          ]
          period = 300
          stat   = "Sum"
          region = var.aws_region
          title  = "Application Metrics"
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 0
        width  = 12
        height = 6
        
        properties = {
          metrics = [
            ["AWS/ApplicationELB", "TargetResponseTime", "LoadBalancer", aws_lb.main.arn_suffix],
            [".", "RequestCount", ".", "."]
          ]
          period = 300
          stat   = "Average"
          region = var.aws_region
          title  = "API Performance"
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 6
        width  = 12
        height = 6
        
        properties = {
          metrics = [
            ["AWS/RDS", "CPUUtilization", "DBClusterIdentifier", module.rds.db_cluster_id],
            [".", "DatabaseConnections", ".", "."]
          ]
          period = 300
          stat   = "Average"
          region = var.aws_region
          title  = "Database Performance"
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 6
        width  = 12
        height = 6
        
        properties = {
          metrics = [
            ["AWS/ElastiCache", "CPUUtilization", "CacheClusterId", aws_elasticache_replication_group.redis.id],
            [".", "DatabaseMemoryUsagePercentage", ".", "."]
          ]
          period = 300
          stat   = "Average"
          region = var.aws_region
          title  = "Redis Performance"
        }
      },
      {
        type   = "log"
        x      = 0
        y      = 12
        width  = 24
        height = 6
        
        properties = {
          query = "SOURCE '/aws/eks/kyb-platform/application'\n| fields @timestamp, @message\n| filter @message like /ERROR/\n| sort @timestamp desc\n| limit 100"
          region = var.aws_region
          title  = "Recent Error Logs"
        }
      }
    ]
  })
}

# CloudWatch Dashboard for Security Monitoring
resource "aws_cloudwatch_dashboard" "security_monitoring" {
  dashboard_name = "kyb-platform-security-monitoring"
  
  dashboard_body = jsonencode({
    widgets = [
      {
        type   = "metric"
        x      = 0
        y      = 0
        width  = 12
        height = 6
        
        properties = {
          metrics = [
            ["KYBPlatform/Security", "SecurityEventCount"]
          ]
          period = 300
          stat   = "Sum"
          region = var.aws_region
          title  = "Security Events"
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 0
        width  = 12
        height = 6
        
        properties = {
          metrics = [
            ["AWS/WAFV2", "AllowedRequests", "WebACL", aws_wafv2_web_acl.alb.name],
            [".", "BlockedRequests", ".", "."]
          ]
          period = 300
          stat   = "Sum"
          region = var.aws_region
          title  = "WAF Activity"
        }
      },
      {
        type   = "log"
        x      = 0
        y      = 6
        width  = 24
        height = 6
        
        properties = {
          query = "SOURCE '/aws/eks/kyb-platform/security'\n| fields @timestamp, @message\n| sort @timestamp desc\n| limit 100"
          region = var.aws_region
          title  = "Security Logs"
        }
      }
    ]
  })
}

# CloudWatch Dashboard for Infrastructure Monitoring
resource "aws_cloudwatch_dashboard" "infrastructure_monitoring" {
  dashboard_name = "kyb-platform-infrastructure-monitoring"
  
  dashboard_body = jsonencode({
    widgets = [
      {
        type   = "metric"
        x      = 0
        y      = 0
        width  = 12
        height = 6
        
        properties = {
          metrics = [
            ["AWS/ECS", "CPUUtilization", "ClusterName", module.eks.cluster_name],
            [".", "MemoryUtilization", ".", "."]
          ]
          period = 300
          stat   = "Average"
          region = var.aws_region
          title  = "EKS Cluster Resources"
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 0
        width  = 12
        height = 6
        
        properties = {
          metrics = [
            ["AWS/ApplicationELB", "HealthyHostCount", "LoadBalancer", aws_lb.main.arn_suffix],
            [".", "UnHealthyHostCount", ".", "."]
          ]
          period = 300
          stat   = "Average"
          region = var.aws_region
          title  = "Load Balancer Health"
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 6
        width  = 12
        height = 6
        
        properties = {
          metrics = [
            ["AWS/EC2", "CPUUtilization"],
            [".", "NetworkIn", ".", "."],
            [".", "NetworkOut", ".", "."]
          ]
          period = 300
          stat   = "Average"
          region = var.aws_region
          title  = "EC2 Instance Metrics"
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 6
        width  = 12
        height = 6
        
        properties = {
          metrics = [
            ["AWS/S3", "NumberOfObjects", "BucketName", aws_s3_bucket.application_data.id],
            [".", "BucketSizeBytes", ".", "."]
          ]
          period = 300
          stat   = "Average"
          region = var.aws_region
          title  = "S3 Storage Metrics"
        }
      }
    ]
  })
}

# CloudWatch Anomaly Detection
resource "aws_cloudwatch_metric_alarm" "response_time_anomaly" {
  count               = var.enable_anomaly_detection ? 1 : 0
  alarm_name          = "kyb-platform-response-time-anomaly"
  comparison_operator = "GreaterThanUpperThreshold"
  evaluation_periods  = "2"
  threshold_metric_id = "e1"
  
  metric_query {
    id          = "e1"
    expression  = "ANOMALY_DETECTION_BAND(m1, 2)"
    label       = "ResponseTime (Expected)"
    return_data = "true"
  }
  
  metric_query {
    id          = "m1"
    return_data = "true"
    metric {
      metric_name = "TargetResponseTime"
      namespace   = "AWS/ApplicationELB"
      period      = "300"
      stat        = "Average"
      
      dimensions = {
        LoadBalancer = aws_lb.main.arn_suffix
        TargetGroup  = aws_lb_target_group.api.arn_suffix
      }
    }
  }
  
  alarm_actions = [aws_sns_topic.warning_alerts.arn]
  ok_actions    = [aws_sns_topic.warning_alerts.arn]
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "monitoring"
    Severity    = "warning"
  }
}

# CloudWatch Composite Alarm for Critical Issues
resource "aws_cloudwatch_composite_alarm" "critical_system_issue" {
  alarm_name = "kyb-platform-critical-system-issue"
  
  alarm_rule = "ALARM(${aws_cloudwatch_metric_alarm.high_error_rate.alarm_name}) OR ALARM(${aws_cloudwatch_metric_alarm.security_events.alarm_name})"
  
  alarm_actions = [aws_sns_topic.critical_alerts.arn]
  ok_actions    = [aws_sns_topic.critical_alerts.arn]
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "monitoring"
    Severity    = "critical"
  }
}

# CloudWatch Insights Queries for Log Analysis
resource "aws_cloudwatch_query_definition" "error_analysis" {
  name = "kyb-platform-error-analysis"
  
  log_group_names = [
    aws_cloudwatch_log_group.application_logs.name
  ]
  
  query_string = <<EOF
fields @timestamp, @message
| filter @message like /ERROR/
| stats count() by bin(5m)
| sort @timestamp desc
EOF
}

resource "aws_cloudwatch_query_definition" "security_analysis" {
  name = "kyb-platform-security-analysis"
  
  log_group_names = [
    aws_cloudwatch_log_group.security_logs.name
  ]
  
  query_string = <<EOF
fields @timestamp, @message
| filter @message like /security/ or @message like /auth/ or @message like /permission/
| stats count() by bin(5m)
| sort @timestamp desc
EOF
}

# Outputs
output "critical_alerts_topic_arn" {
  description = "ARN of the critical alerts SNS topic"
  value       = aws_sns_topic.critical_alerts.arn
}

output "warning_alerts_topic_arn" {
  description = "ARN of the warning alerts SNS topic"
  value       = aws_sns_topic.warning_alerts.arn
}

output "application_dashboard_url" {
  description = "URL of the application monitoring CloudWatch dashboard"
  value       = "https://${var.aws_region}.console.aws.amazon.com/cloudwatch/home?region=${var.aws_region}#dashboards:name=kyb-platform-application-monitoring"
}

output "security_dashboard_url" {
  description = "URL of the security monitoring CloudWatch dashboard"
  value       = "https://${var.aws_region}.console.aws.amazon.com/cloudwatch/home?region=${var.aws_region}#dashboards:name=kyb-platform-security-monitoring"
}

output "infrastructure_dashboard_url" {
  description = "URL of the infrastructure monitoring CloudWatch dashboard"
  value       = "https://${var.aws_region}.console.aws.amazon.com/cloudwatch/home?region=${var.aws_region}#dashboards:name=kyb-platform-infrastructure-monitoring"
}
