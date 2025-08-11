# KYB Platform - Terraform Variables
# Input variables for infrastructure configuration

variable "aws_region" {
  description = "AWS region for infrastructure deployment"
  type        = string
  default     = "us-west-2"
}

variable "environment" {
  description = "Environment name (production, staging, development)"
  type        = string
  default     = "production"
  
  validation {
    condition     = contains(["production", "staging", "development"], var.environment)
    error_message = "Environment must be one of: production, staging, development."
  }
}

variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "List of availability zones"
  type        = list(string)
  default     = ["us-west-2a", "us-west-2b", "us-west-2c"]
}

variable "private_subnet_cidrs" {
  description = "CIDR blocks for private subnets"
  type        = list(string)
  default     = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
}

variable "public_subnet_cidrs" {
  description = "CIDR blocks for public subnets"
  type        = list(string)
  default     = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]
}

variable "domain_name" {
  description = "Domain name for the application"
  type        = string
  default     = "kybplatform.com"
}

variable "certificate_arn" {
  description = "ARN of SSL certificate for HTTPS"
  type        = string
  default     = ""
}

variable "db_username" {
  description = "Database master username"
  type        = string
  default     = "kyb_admin"
  
  validation {
    condition     = length(var.db_username) >= 3
    error_message = "Database username must be at least 3 characters long."
  }
}

variable "db_password" {
  description = "Database master password"
  type        = string
  sensitive   = true
  
  validation {
    condition     = length(var.db_password) >= 8
    error_message = "Database password must be at least 8 characters long."
  }
}

variable "eks_cluster_version" {
  description = "EKS cluster version"
  type        = string
  default     = "1.28"
}

variable "eks_node_instance_types" {
  description = "Instance types for EKS nodes"
  type        = list(string)
  default     = ["t3.medium"]
}

variable "eks_node_desired_capacity" {
  description = "Desired number of EKS nodes"
  type        = number
  default     = 2
  
  validation {
    condition     = var.eks_node_desired_capacity >= 1
    error_message = "Desired capacity must be at least 1."
  }
}

variable "eks_node_max_capacity" {
  description = "Maximum number of EKS nodes"
  type        = number
  default     = 10
  
  validation {
    condition     = var.eks_node_max_capacity >= var.eks_node_desired_capacity
    error_message = "Max capacity must be greater than or equal to desired capacity."
  }
}

variable "rds_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t3.micro"
}

variable "rds_allocated_storage" {
  description = "RDS allocated storage in GB"
  type        = number
  default     = 20
  
  validation {
    condition     = var.rds_allocated_storage >= 20
    error_message = "RDS allocated storage must be at least 20 GB."
  }
}

variable "redis_node_type" {
  description = "ElastiCache Redis node type"
  type        = string
  default     = "cache.t3.micro"
}

variable "enable_spot_instances" {
  description = "Enable spot instances for cost optimization"
  type        = bool
  default     = true
}

variable "enable_auto_scaling" {
  description = "Enable auto scaling for EKS nodes"
  type        = bool
  default     = true
}

variable "enable_monitoring" {
  description = "Enable CloudWatch monitoring"
  type        = bool
  default     = true
}

variable "enable_backup" {
  description = "Enable automated backups"
  type        = bool
  default     = true
}

variable "backup_retention_days" {
  description = "Number of days to retain backups"
  type        = number
  default     = 7
  
  validation {
    condition     = var.backup_retention_days >= 1 && var.backup_retention_days <= 35
    error_message = "Backup retention days must be between 1 and 35."
  }
}

variable "log_retention_days" {
  description = "Number of days to retain CloudWatch logs"
  type        = number
  default     = 30
  
  validation {
    condition     = var.log_retention_days >= 1 && var.log_retention_days <= 365
    error_message = "Log retention days must be between 1 and 365."
  }
}

variable "enable_encryption" {
  description = "Enable encryption at rest"
  type        = bool
  default     = true
}

variable "enable_ssl" {
  description = "Enable SSL/TLS encryption in transit"
  type        = bool
  default     = true
}

variable "enable_deletion_protection" {
  description = "Enable deletion protection for critical resources"
  type        = bool
  default     = true
}

variable "tags" {
  description = "Additional tags for resources"
  type        = map(string)
  default     = {}
}

# Load Balancer Variables
variable "admin_ip_addresses" {
  description = "IP addresses allowed to access admin endpoints"
  type        = list(string)
  default     = []
}

variable "alert_emails" {
  description = "Email addresses for alert notifications"
  type        = list(string)
  default     = []
  
  validation {
    condition     = alltrue([for email in var.alert_emails : can(regex("^[^@]+@[^@]+\\.[^@]+$", email))])
    error_message = "All alert emails must be valid email addresses."
  }
}

variable "waf_rate_limit" {
  description = "Rate limit for WAF (requests per 5 minutes)"
  type        = number
  default     = 2000
  
  validation {
    condition     = var.waf_rate_limit >= 100 && var.waf_rate_limit <= 10000
    error_message = "WAF rate limit must be between 100 and 10000 requests per 5 minutes."
  }
}

variable "health_check_path" {
  description = "Path for health checks"
  type        = string
  default     = "/health"
}

variable "health_check_interval" {
  description = "Health check interval in seconds"
  type        = number
  default     = 30
  
  validation {
    condition     = var.health_check_interval >= 5 && var.health_check_interval <= 300
    error_message = "Health check interval must be between 5 and 300 seconds."
  }
}

variable "health_check_timeout" {
  description = "Health check timeout in seconds"
  type        = number
  default     = 5
  
  validation {
    condition     = var.health_check_timeout >= 2 && var.health_check_timeout <= 60
    error_message = "Health check timeout must be between 2 and 60 seconds."
  }
}

variable "alb_idle_timeout" {
  description = "ALB idle timeout in seconds"
  type        = number
  default     = 60
  
  validation {
    condition     = var.alb_idle_timeout >= 1 && var.alb_idle_timeout <= 4000
    error_message = "ALB idle timeout must be between 1 and 4000 seconds."
  }
}

variable "enable_sticky_sessions" {
  description = "Enable sticky sessions for API consistency"
  type        = bool
  default     = true
}

variable "sticky_session_duration" {
  description = "Sticky session duration in seconds"
  type        = number
  default     = 86400
  
  validation {
    condition     = var.sticky_session_duration >= 1 && var.sticky_session_duration <= 604800
    error_message = "Sticky session duration must be between 1 and 604800 seconds (7 days)."
  }
}

# Auto Scaling Variables
variable "api_min_replicas" {
  description = "Minimum number of API replicas"
  type        = number
  default     = 2
  
  validation {
    condition     = var.api_min_replicas >= 1 && var.api_min_replicas <= 10
    error_message = "API min replicas must be between 1 and 10."
  }
}

variable "api_max_replicas" {
  description = "Maximum number of API replicas"
  type        = number
  default     = 20
  
  validation {
    condition     = var.api_max_replicas >= var.api_min_replicas && var.api_max_replicas <= 100
    error_message = "API max replicas must be greater than min replicas and less than 100."
  }
}

variable "api_target_cpu_utilization" {
  description = "Target CPU utilization percentage for API scaling"
  type        = number
  default     = 70
  
  validation {
    condition     = var.api_target_cpu_utilization >= 10 && var.api_target_cpu_utilization <= 90
    error_message = "API target CPU utilization must be between 10 and 90 percent."
  }
}

variable "api_target_memory_utilization" {
  description = "Target memory utilization percentage for API scaling"
  type        = number
  default     = 80
  
  validation {
    condition     = var.api_target_memory_utilization >= 10 && var.api_target_memory_utilization <= 90
    error_message = "API target memory utilization must be between 10 and 90 percent."
  }
}

variable "api_target_rps" {
  description = "Target requests per second for API scaling"
  type        = number
  default     = 100
  
  validation {
    condition     = var.api_target_rps >= 10 && var.api_target_rps <= 1000
    error_message = "API target RPS must be between 10 and 1000."
  }
}

variable "enable_rds_autoscaling" {
  description = "Enable auto scaling for RDS"
  type        = bool
  default     = true
}

variable "rds_min_capacity" {
  description = "Minimum RDS capacity"
  type        = number
  default     = 1
  
  validation {
    condition     = var.rds_min_capacity >= 1 && var.rds_min_capacity <= 10
    error_message = "RDS min capacity must be between 1 and 10."
  }
}

variable "rds_max_capacity" {
  description = "Maximum RDS capacity"
  type        = number
  default     = 5
  
  validation {
    condition     = var.rds_max_capacity >= var.rds_min_capacity && var.rds_max_capacity <= 15
    error_message = "RDS max capacity must be greater than min capacity and less than 15."
  }
}

variable "rds_target_cpu_utilization" {
  description = "Target CPU utilization percentage for RDS scaling"
  type        = number
  default     = 70
  
  validation {
    condition     = var.rds_target_cpu_utilization >= 10 && var.rds_target_cpu_utilization <= 90
    error_message = "RDS target CPU utilization must be between 10 and 90 percent."
  }
}

variable "enable_redis_autoscaling" {
  description = "Enable auto scaling for Redis"
  type        = bool
  default     = true
}

variable "redis_min_capacity" {
  description = "Minimum Redis capacity"
  type        = number
  default     = 1
  
  validation {
    condition     = var.redis_min_capacity >= 1 && var.redis_min_capacity <= 5
    error_message = "Redis min capacity must be between 1 and 5."
  }
}

variable "redis_max_capacity" {
  description = "Maximum Redis capacity"
  type        = number
  default     = 3
  
  validation {
    condition     = var.redis_max_capacity >= var.redis_min_capacity && var.redis_max_capacity <= 10
    error_message = "Redis max capacity must be greater than min capacity and less than 10."
  }
}

variable "redis_target_cpu_utilization" {
  description = "Target CPU utilization percentage for Redis scaling"
  type        = number
  default     = 70
  
  validation {
    condition     = var.redis_target_cpu_utilization >= 10 && var.redis_target_cpu_utilization <= 90
    error_message = "Redis target CPU utilization must be between 10 and 90 percent."
  }
}

variable "cluster_cpu_threshold" {
  description = "CPU threshold for cluster scaling alarms"
  type        = number
  default     = 80
  
  validation {
    condition     = var.cluster_cpu_threshold >= 50 && var.cluster_cpu_threshold <= 95
    error_message = "Cluster CPU threshold must be between 50 and 95 percent."
  }
}

variable "cluster_memory_threshold" {
  description = "Memory threshold for cluster scaling alarms"
  type        = number
  default     = 80
  
  validation {
    condition     = var.cluster_memory_threshold >= 50 && var.cluster_memory_threshold <= 95
    error_message = "Cluster memory threshold must be between 50 and 95 percent."
  }
}

variable "api_response_time_threshold" {
  description = "Response time threshold for API scaling alarms (seconds)"
  type        = number
  default     = 2.0
  
  validation {
    condition     = var.api_response_time_threshold >= 0.5 && var.api_response_time_threshold <= 10.0
    error_message = "API response time threshold must be between 0.5 and 10.0 seconds."
  }
}

variable "autoscaling_emails" {
  description = "Email addresses for auto scaling notifications"
  type        = list(string)
  default     = []
  
  validation {
    condition     = alltrue([for email in var.autoscaling_emails : can(regex("^[^@]+@[^@]+\\.[^@]+$", email))])
    error_message = "All auto scaling emails must be valid email addresses."
  }
}

# Monitoring and Alerting Variables
variable "error_rate_threshold" {
  description = "Error rate threshold for critical alerts"
  type        = number
  default     = 10
  
  validation {
    condition     = var.error_rate_threshold >= 1 && var.error_rate_threshold <= 100
    error_message = "Error rate threshold must be between 1 and 100."
  }
}

variable "response_time_threshold" {
  description = "Response time threshold for warning alerts (seconds)"
  type        = number
  default     = 2.0
  
  validation {
    condition     = var.response_time_threshold >= 0.5 && var.response_time_threshold <= 10.0
    error_message = "Response time threshold must be between 0.5 and 10.0 seconds."
  }
}

variable "db_connection_threshold" {
  description = "Database connection threshold for warning alerts"
  type        = number
  default     = 80
  
  validation {
    condition     = var.db_connection_threshold >= 10 && var.db_connection_threshold <= 100
    error_message = "Database connection threshold must be between 10 and 100."
  }
}

variable "redis_memory_threshold" {
  description = "Redis memory usage threshold for warning alerts (percentage)"
  type        = number
  default     = 80
  
  validation {
    condition     = var.redis_memory_threshold >= 50 && var.redis_memory_threshold <= 95
    error_message = "Redis memory threshold must be between 50 and 95 percent."
  }
}

variable "security_event_threshold" {
  description = "Security event threshold for critical alerts"
  type        = number
  default     = 5
  
  validation {
    condition     = var.security_event_threshold >= 1 && var.security_event_threshold <= 50
    error_message = "Security event threshold must be between 1 and 50."
  }
}

variable "critical_alert_emails" {
  description = "Email addresses for critical alerts"
  type        = list(string)
  default     = []
  
  validation {
    condition     = alltrue([for email in var.critical_alert_emails : can(regex("^[^@]+@[^@]+\\.[^@]+$", email))])
    error_message = "All critical alert emails must be valid email addresses."
  }
}

variable "warning_alert_emails" {
  description = "Email addresses for warning alerts"
  type        = list(string)
  default     = []
  
  validation {
    condition     = alltrue([for email in var.warning_alert_emails : can(regex("^[^@]+@[^@]+\\.[^@]+$", email))])
    error_message = "All warning alert emails must be valid email addresses."
  }
}

variable "info_alert_emails" {
  description = "Email addresses for info alerts"
  type        = list(string)
  default     = []
  
  validation {
    condition     = alltrue([for email in var.info_alert_emails : can(regex("^[^@]+@[^@]+\\.[^@]+$", email))])
    error_message = "All info alert emails must be valid email addresses."
  }
}

variable "enable_slack_alerts" {
  description = "Enable Slack alerts"
  type        = bool
  default     = false
}

variable "slack_webhook_url" {
  description = "Slack webhook URL for alerts"
  type        = string
  default     = ""
  sensitive   = true
}

variable "enable_anomaly_detection" {
  description = "Enable CloudWatch anomaly detection"
  type        = bool
  default     = true
}

variable "monitoring_retention_days" {
  description = "Number of days to retain monitoring data"
  type        = number
  default     = 30
  
  validation {
    condition     = var.monitoring_retention_days >= 1 && var.monitoring_retention_days <= 365
    error_message = "Monitoring retention days must be between 1 and 365."
  }
}

variable "enable_prometheus_metrics" {
  description = "Enable Prometheus metrics collection"
  type        = bool
  default     = true
}

variable "enable_grafana_dashboards" {
  description = "Enable Grafana dashboards"
  type        = bool
  default     = true
}

# Backup and Disaster Recovery Variables
variable "enable_cross_region_backup" {
  description = "Enable cross-region backup replication"
  type        = bool
  default     = true
}

variable "dr_region" {
  description = "Disaster recovery region"
  type        = string
  default     = "us-east-1"
  
  validation {
    condition     = can(regex("^[a-z]{2}-[a-z]+-[0-9]+$", var.dr_region))
    error_message = "DR region must be a valid AWS region."
  }
}

variable "enable_rds_backup" {
  description = "Enable RDS automated backups"
  type        = bool
  default     = true
}

variable "enable_eks_backup" {
  description = "Enable EKS cluster backup"
  type        = bool
  default     = true
}

variable "enable_aws_backup" {
  description = "Enable AWS Backup service"
  type        = bool
  default     = true
}

variable "enable_backup_monitoring" {
  description = "Enable backup monitoring and alerting"
  type        = bool
  default     = true
}

variable "enable_dr_health_check" {
  description = "Enable disaster recovery health checks"
  type        = bool
  default     = true
}

variable "enable_dr_failover" {
  description = "Enable disaster recovery failover"
  type        = bool
  default     = true
}

variable "backup_schedule" {
  description = "Backup schedule in cron format"
  type        = string
  default     = "0 2 * * *"  # Daily at 2 AM
  
  validation {
    condition     = can(regex("^[0-9*/, -]+$", var.backup_schedule))
    error_message = "Backup schedule must be in valid cron format."
  }
}

variable "backup_window" {
  description = "Backup window in HH:MM-HH:MM format"
  type        = string
  default     = "02:00-04:00"
  
  validation {
    condition     = can(regex("^([01]?[0-9]|2[0-3]):[0-5][0-9]-([01]?[0-9]|2[0-3]):[0-5][0-9]$", var.backup_window))
    error_message = "Backup window must be in HH:MM-HH:MM format."
  }
}

variable "backup_encryption_algorithm" {
  description = "Backup encryption algorithm"
  type        = string
  default     = "AES256"
  
  validation {
    condition     = contains(["AES256", "aws:kms"], var.backup_encryption_algorithm)
    error_message = "Backup encryption algorithm must be AES256 or aws:kms."
  }
}

variable "backup_compression" {
  description = "Enable backup compression"
  type        = bool
  default     = true
}

variable "backup_validation" {
  description = "Enable backup validation"
  type        = bool
  default     = true
}

variable "dr_rto_minutes" {
  description = "Disaster recovery RTO in minutes"
  type        = number
  default     = 60
  
  validation {
    condition     = var.dr_rto_minutes >= 15 && var.dr_rto_minutes <= 1440
    error_message = "DR RTO must be between 15 and 1440 minutes."
  }
}

variable "dr_rpo_minutes" {
  description = "Disaster recovery RPO in minutes"
  type        = number
  default     = 15
  
  validation {
    condition     = var.dr_rpo_minutes >= 1 && var.dr_rpo_minutes <= 1440
    error_message = "DR RPO must be between 1 and 1440 minutes."
  }
}

variable "backup_notification_emails" {
  description = "Email addresses for backup notifications"
  type        = list(string)
  default     = []
  
  validation {
    condition     = alltrue([for email in var.backup_notification_emails : can(regex("^[^@]+@[^@]+\\.[^@]+$", email))])
    error_message = "All backup notification emails must be valid email addresses."
  }
}

variable "enable_backup_testing" {
  description = "Enable automated backup testing"
  type        = bool
  default     = true
}

variable "backup_test_schedule" {
  description = "Backup testing schedule in cron format"
  type        = string
  default     = "0 5 * * 0"  # Weekly on Sunday at 5 AM
  
  validation {
    condition     = can(regex("^[0-9*/, -]+$", var.backup_test_schedule))
    error_message = "Backup test schedule must be in valid cron format."
  }
}

variable "enable_backup_reporting" {
  description = "Enable backup reporting"
  type        = bool
  default     = true
}

variable "backup_report_schedule" {
  description = "Backup report schedule in cron format"
  type        = string
  default     = "0 6 * * 1"  # Weekly on Monday at 6 AM
  
  validation {
    condition     = can(regex("^[0-9*/, -]+$", var.backup_report_schedule))
    error_message = "Backup report schedule must be in valid cron format."
  }
}
