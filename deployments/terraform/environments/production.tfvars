# KYB Platform - Production Environment Configuration
# Terraform variables for production deployment

# Environment
environment = "production"
aws_region  = "us-west-2"

# Networking
vpc_cidr = "10.0.0.0/16"
availability_zones = [
  "us-west-2a",
  "us-west-2b", 
  "us-west-2c"
]
private_subnet_cidrs = [
  "10.0.1.0/24",
  "10.0.2.0/24",
  "10.0.3.0/24"
]
public_subnet_cidrs = [
  "10.0.101.0/24",
  "10.0.102.0/24",
  "10.0.103.0/24"
]

# Domain and SSL
domain_name = "kybplatform.com"
certificate_arn = "arn:aws:acm:us-west-2:123456789012:certificate/your-certificate-id"

# Database
db_username = "kyb_admin"
# db_password should be provided via environment variable or secrets manager
rds_instance_class = "db.t3.small"
rds_allocated_storage = 50

# EKS Configuration
eks_cluster_version = "1.28"
eks_node_instance_types = ["t3.medium", "t3.small"]
eks_node_desired_capacity = 3
eks_node_max_capacity = 15

# Redis Configuration
redis_node_type = "cache.t3.micro"

# Features
enable_spot_instances = true
enable_auto_scaling = true
enable_monitoring = true
enable_backup = true
enable_encryption = true
enable_ssl = true
enable_deletion_protection = true

# Retention
backup_retention_days = 14
log_retention_days = 90

# Additional tags
tags = {
  Environment = "production"
  Project     = "kyb-platform"
  Owner       = "platform-team"
  CostCenter  = "engineering"
  Compliance  = "soc2-pci-gdpr"
}

# Load Balancer Configuration
admin_ip_addresses = [
  "192.168.1.0/24",  # Office network
  "10.0.0.0/8"       # VPC network
]

alert_emails = [
  "alerts@kybplatform.com",
  "ops@kybplatform.com"
]

waf_rate_limit = 5000
health_check_path = "/health"
health_check_interval = 30
health_check_timeout = 5
alb_idle_timeout = 60
enable_sticky_sessions = true
sticky_session_duration = 86400

# Auto Scaling Configuration
api_min_replicas = 3
api_max_replicas = 30
api_target_cpu_utilization = 70
api_target_memory_utilization = 80
api_target_rps = 150
api_response_time_threshold = 1.5

enable_rds_autoscaling = true
rds_min_capacity = 1
rds_max_capacity = 5
rds_target_cpu_utilization = 70

enable_redis_autoscaling = true
redis_min_capacity = 1
redis_max_capacity = 3
redis_target_cpu_utilization = 70

cluster_cpu_threshold = 80
cluster_memory_threshold = 80

autoscaling_emails = [
  "autoscaling@kybplatform.com",
  "ops@kybplatform.com"
]

# Monitoring and Alerting Configuration
error_rate_threshold = 5
response_time_threshold = 1.5
db_connection_threshold = 80
redis_memory_threshold = 80
security_event_threshold = 3

critical_alert_emails = [
  "critical@kybplatform.com",
  "oncall@kybplatform.com"
]

warning_alert_emails = [
  "alerts@kybplatform.com",
  "ops@kybplatform.com"
]

info_alert_emails = [
  "info@kybplatform.com"
]

enable_slack_alerts = true
slack_webhook_url = "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"

enable_anomaly_detection = true
monitoring_retention_days = 90
enable_prometheus_metrics = true
enable_grafana_dashboards = true

# Backup and Disaster Recovery Configuration
enable_cross_region_backup = true
dr_region = "us-east-1"
enable_rds_backup = true
enable_eks_backup = true
enable_aws_backup = true
enable_backup_monitoring = true
enable_dr_health_check = true
enable_dr_failover = true

backup_schedule = "0 2 * * *"  # Daily at 2 AM
backup_window = "02:00-04:00"
backup_encryption_algorithm = "AES256"
backup_compression = true
backup_validation = true

dr_rto_minutes = 60
dr_rpo_minutes = 15

backup_notification_emails = [
  "backup@kybplatform.com",
  "ops@kybplatform.com"
]

enable_backup_testing = true
backup_test_schedule = "0 5 * * 0"  # Weekly on Sunday at 5 AM
enable_backup_reporting = true
backup_report_schedule = "0 6 * * 1"  # Weekly on Monday at 6 AM
