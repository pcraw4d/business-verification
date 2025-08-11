# KYB Platform - Backup and Disaster Recovery Configuration
# Comprehensive backup strategies, disaster recovery, and business continuity

# S3 Bucket for Backup Storage
resource "aws_s3_bucket" "backup_storage" {
  bucket = "kyb-platform-backup-${var.environment}-${random_string.bucket_suffix.result}"
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "backup"
    Purpose     = "disaster-recovery"
  }
}

# Random string for bucket naming
resource "random_string" "bucket_suffix" {
  length  = 8
  special = false
  upper   = false
}

# S3 Bucket Versioning for Backup Protection
resource "aws_s3_bucket_versioning" "backup_storage" {
  bucket = aws_s3_bucket.backup_storage.id
  
  versioning_configuration {
    status = "Enabled"
  }
}

# S3 Bucket Encryption for Backup Security
resource "aws_s3_bucket_server_side_encryption_configuration" "backup_storage" {
  bucket = aws_s3_bucket.backup_storage.id
  
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# S3 Bucket Lifecycle Policy for Backup Retention
resource "aws_s3_bucket_lifecycle_configuration" "backup_storage" {
  bucket = aws_s3_bucket.backup_storage.id
  
  rule {
    id     = "backup-retention-policy"
    status = "Enabled"
    
    # Transition to IA after 30 days
    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }
    
    # Transition to Glacier after 90 days
    transition {
      days          = 90
      storage_class = "GLACIER"
    }
    
    # Transition to Deep Archive after 365 days
    transition {
      days          = 365
      storage_class = "DEEP_ARCHIVE"
    }
    
    # Delete after 7 years
    expiration {
      days = 2555  # 7 years
    }
    
    # Delete incomplete multipart uploads after 7 days
    abort_incomplete_multipart_upload {
      days_after_initiation = 7
    }
  }
}

# S3 Bucket Public Access Block
resource "aws_s3_bucket_public_access_block" "backup_storage" {
  bucket = aws_s3_bucket.backup_storage.id
  
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# S3 Bucket Policy for Backup Access Control
resource "aws_s3_bucket_policy" "backup_storage" {
  bucket = aws_s3_bucket.backup_storage.id
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "DenyUnencryptedObjectUploads"
        Effect = "Deny"
        Principal = {
          AWS = "*"
        }
        Action = [
          "s3:PutObject"
        ]
        Resource = "${aws_s3_bucket.backup_storage.arn}/*"
        Condition = {
          StringNotEquals = {
            "s3:x-amz-server-side-encryption" = "AES256"
          }
        }
      },
      {
        Sid    = "DenyIncorrectEncryptionHeader"
        Effect = "Deny"
        Principal = {
          AWS = "*"
        }
        Action = [
          "s3:PutObject"
        ]
        Resource = "${aws_s3_bucket.backup_storage.arn}/*"
        Condition = {
          StringNotEquals = {
            "s3:x-amz-server-side-encryption" = "AES256"
          }
        }
      },
      {
        Sid    = "DenyUnencryptedObjectUploads"
        Effect = "Deny"
        Principal = {
          AWS = "*"
        }
        Action = [
          "s3:PutObject"
        ]
        Resource = "${aws_s3_bucket.backup_storage.arn}/*"
        Condition = {
          Null = {
            "s3:x-amz-server-side-encryption" = "true"
          }
        }
      }
    ]
  })
}

# Cross-Region Replication for Disaster Recovery
resource "aws_s3_bucket" "backup_replica" {
  count  = var.enable_cross_region_backup ? 1 : 0
  bucket = "kyb-platform-backup-replica-${var.environment}-${random_string.bucket_suffix.result}"
  provider = aws.dr_region
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "backup"
    Purpose     = "disaster-recovery-replica"
  }
}

# S3 Bucket Versioning for Replica
resource "aws_s3_bucket_versioning" "backup_replica" {
  count  = var.enable_cross_region_backup ? 1 : 0
  bucket = aws_s3_bucket.backup_replica[0].id
  
  versioning_configuration {
    status = "Enabled"
  }
}

# S3 Bucket Encryption for Replica
resource "aws_s3_bucket_server_side_encryption_configuration" "backup_replica" {
  count  = var.enable_cross_region_backup ? 1 : 0
  bucket = aws_s3_bucket.backup_replica[0].id
  
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# Cross-Region Replication Configuration
resource "aws_s3_bucket_replication_configuration" "backup_storage" {
  count  = var.enable_cross_region_backup ? 1 : 0
  bucket = aws_s3_bucket.backup_storage.id
  role   = aws_iam_role.s3_replication[0].arn
  
  rule {
    id     = "cross-region-backup-replication"
    status = "Enabled"
    
    destination {
      bucket        = aws_s3_bucket.backup_replica[0].arn
      storage_class = "STANDARD"
    }
    
    source_selection_criteria {
      sse_kms_encrypted_objects {
        status = "Enabled"
      }
    }
  }
}

# IAM Role for S3 Replication
resource "aws_iam_role" "s3_replication" {
  count = var.enable_cross_region_backup ? 1 : 0
  name  = "kyb-platform-s3-replication-role"
  
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "s3.amazonaws.com"
        }
      }
    ]
  })
}

# IAM Policy for S3 Replication
resource "aws_iam_role_policy" "s3_replication" {
  count = var.enable_cross_region_backup ? 1 : 0
  name  = "kyb-platform-s3-replication-policy"
  role  = aws_iam_role.s3_replication[0].id
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetReplicationConfiguration",
          "s3:ListBucket"
        ]
        Resource = aws_s3_bucket.backup_storage.arn
      },
      {
        Effect = "Allow"
        Action = [
          "s3:GetObjectVersion",
          "s3:GetObjectVersionAcl"
        ]
        Resource = "${aws_s3_bucket.backup_storage.arn}/*"
      },
      {
        Effect = "Allow"
        Action = [
          "s3:ReplicateObject",
          "s3:ReplicateDelete"
        ]
        Resource = "${aws_s3_bucket.backup_replica[0].arn}/*"
      }
    ]
  })
}

# RDS Automated Backups
resource "aws_db_instance" "rds_backup" {
  count = var.enable_rds_backup ? 1 : 0
  
  identifier = "kyb-platform-rds-backup"
  
  # Use the same configuration as main RDS
  engine               = "postgres"
  engine_version       = "15.4"
  instance_class       = var.rds_instance_class
  allocated_storage    = var.rds_allocated_storage
  storage_type         = "gp3"
  storage_encrypted    = true
  
  # Backup configuration
  backup_retention_period = var.backup_retention_days
  backup_window          = "03:00-04:00"
  maintenance_window     = "sun:04:00-sun:05:00"
  
  # Enable automated backups
  backup_retention_period = 35  # 5 weeks
  delete_automated_backups = false
  
  # Enable point-in-time recovery
  storage_encrypted = true
  
  # Enable deletion protection
  deletion_protection = true
  
  # Enable performance insights
  performance_insights_enabled = true
  performance_insights_retention_period = 7
  
  # Enable monitoring
  monitoring_interval = 60
  monitoring_role_arn = aws_iam_role.rds_monitoring.arn
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "backup"
    Purpose     = "disaster-recovery"
  }
}

# RDS Snapshot Schedule
resource "aws_db_instance_automated_backups_replication" "rds_backup" {
  count = var.enable_cross_region_backup ? 1 : 0
  
  source_db_instance_arn = module.rds.db_instance_arn
  kms_key_id             = aws_kms_key.backup[0].arn
  
  # Replicate to DR region
  destination_region = var.dr_region
  
  # Retention period for replicated backups
  retention_period = 35
}

# KMS Key for Backup Encryption
resource "aws_kms_key" "backup" {
  count = var.enable_cross_region_backup ? 1 : 0
  
  description             = "KMS key for KYB Platform backup encryption"
  deletion_window_in_days = 7
  enable_key_rotation     = true
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "backup"
  }
}

# KMS Key Alias
resource "aws_kms_alias" "backup" {
  count = var.enable_cross_region_backup ? 1 : 0
  
  name          = "alias/kyb-platform-backup"
  target_key_id = aws_kms_key.backup[0].key_id
}

# EKS Cluster Backup
resource "aws_eks_cluster" "backup" {
  count = var.enable_eks_backup ? 1 : 0
  
  name     = "kyb-platform-backup-cluster"
  role_arn = aws_iam_role.eks_backup[0].arn
  version  = var.eks_cluster_version
  
  vpc_config {
    subnet_ids = module.vpc.private_subnets
  }
  
  # Enable control plane logging
  enabled_cluster_log_types = ["api", "audit", "authenticator", "controllerManager", "scheduler"]
  
  # Encryption configuration
  encryption_config {
    provider {
      key_arn = aws_kms_key.backup[0].arn
    }
    resources = ["secrets"]
  }
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "backup"
    Purpose     = "disaster-recovery"
  }
}

# IAM Role for EKS Backup
resource "aws_iam_role" "eks_backup" {
  count = var.enable_eks_backup ? 1 : 0
  name  = "kyb-platform-eks-backup-role"
  
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "eks.amazonaws.com"
        }
      }
    ]
  })
}

# IAM Policy for EKS Backup
resource "aws_iam_role_policy_attachment" "eks_backup" {
  count = var.enable_eks_backup ? 1 : 0
  
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.eks_backup[0].name
}

# AWS Backup Vault
resource "aws_backup_vault" "main" {
  count = var.enable_aws_backup ? 1 : 0
  
  name = "kyb-platform-backup-vault"
  
  # Enable encryption
  encryption_key_arn = aws_kms_key.backup[0].arn
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "backup"
  }
}

# AWS Backup Plan
resource "aws_backup_plan" "main" {
  count = var.enable_aws_backup ? 1 : 0
  
  name = "kyb-platform-backup-plan"
  
  rule {
    rule_name         = "daily_backup"
    target_vault_name = aws_backup_vault.main[0].name
    
    schedule = "cron(0 2 * * ? *)"  # Daily at 2 AM
    
    lifecycle {
      delete_after = var.backup_retention_days
    }
    
    copy_action {
      destination_vault_arn = var.enable_cross_region_backup ? aws_backup_vault.replica[0].arn : null
    }
  }
  
  rule {
    rule_name         = "weekly_backup"
    target_vault_name = aws_backup_vault.main[0].name
    
    schedule = "cron(0 3 ? * SUN *)"  # Weekly on Sunday at 3 AM
    
    lifecycle {
      delete_after = 90  # 3 months
    }
  }
  
  rule {
    rule_name         = "monthly_backup"
    target_vault_name = aws_backup_vault.main[0].name
    
    schedule = "cron(0 4 1 * ? *)"  # Monthly on 1st at 4 AM
    
    lifecycle {
      delete_after = 365  # 1 year
    }
  }
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "backup"
  }
}

# AWS Backup Vault for Cross-Region Replication
resource "aws_backup_vault" "replica" {
  count = var.enable_cross_region_backup && var.enable_aws_backup ? 1 : 0
  
  provider = aws.dr_region
  name     = "kyb-platform-backup-vault-replica"
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "backup"
    Purpose     = "disaster-recovery"
  }
}

# AWS Backup Selection
resource "aws_backup_selection" "main" {
  count = var.enable_aws_backup ? 1 : 0
  
  name         = "kyb-platform-backup-selection"
  iam_role_arn = aws_iam_role.backup[0].arn
  plan_id      = aws_backup_plan.main[0].id
  
  resources = [
    module.rds.db_instance_arn,
    aws_elasticache_replication_group.redis.arn,
    aws_s3_bucket.application_data.arn,
    aws_s3_bucket.backup_storage.arn
  ]
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "backup"
  }
}

# IAM Role for AWS Backup
resource "aws_iam_role" "backup" {
  count = var.enable_aws_backup ? 1 : 0
  name  = "kyb-platform-backup-role"
  
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "backup.amazonaws.com"
        }
      }
    ]
  })
}

# IAM Policy for AWS Backup
resource "aws_iam_role_policy_attachment" "backup" {
  count = var.enable_aws_backup ? 1 : 0
  
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSBackupServiceRolePolicyForBackup"
  role       = aws_iam_role.backup[0].name
}

# IAM Policy for AWS Backup Restore
resource "aws_iam_role_policy_attachment" "backup_restore" {
  count = var.enable_aws_backup ? 1 : 0
  
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSBackupServiceRolePolicyForRestores"
  role       = aws_iam_role.backup[0].name
}

# CloudWatch Alarms for Backup Monitoring
resource "aws_cloudwatch_metric_alarm" "backup_failure" {
  count = var.enable_backup_monitoring ? 1 : 0
  
  alarm_name          = "kyb-platform-backup-failure"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "1"
  metric_name         = "BackupJobsFailed"
  namespace           = "AWS/Backup"
  period              = "300"
  statistic           = "Sum"
  threshold           = "1"
  alarm_description   = "This metric monitors backup job failures"
  
  alarm_actions = [aws_sns_topic.critical_alerts.arn]
  ok_actions    = [aws_sns_topic.critical_alerts.arn]
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "backup"
    Severity    = "critical"
  }
}

resource "aws_cloudwatch_metric_alarm" "backup_success" {
  count = var.enable_backup_monitoring ? 1 : 0
  
  alarm_name          = "kyb-platform-backup-success"
  comparison_operator = "LessThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "BackupJobsCompleted"
  namespace           = "AWS/Backup"
  period              = "3600"
  statistic           = "Sum"
  threshold           = "1"
  alarm_description   = "This metric monitors backup job completion"
  
  alarm_actions = [aws_sns_topic.warning_alerts.arn]
  ok_actions    = [aws_sns_topic.warning_alerts.arn]
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "backup"
    Severity    = "warning"
  }
}

# CloudWatch Dashboard for Backup Monitoring
resource "aws_cloudwatch_dashboard" "backup_monitoring" {
  count = var.enable_backup_monitoring ? 1 : 0
  
  dashboard_name = "kyb-platform-backup-monitoring"
  
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
            ["AWS/Backup", "BackupJobsCompleted"],
            [".", "BackupJobsFailed"]
          ]
          period = 3600
          stat   = "Sum"
          region = var.aws_region
          title  = "Backup Job Status"
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
            ["AWS/Backup", "BackupJobsRunning"],
            [".", "RestoreJobsRunning"]
          ]
          period = 300
          stat   = "Sum"
          region = var.aws_region
          title  = "Active Backup/Restore Jobs"
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
            ["AWS/S3", "NumberOfObjects", "BucketName", aws_s3_bucket.backup_storage.id],
            [".", "BucketSizeBytes", ".", "."]
          ]
          period = 3600
          stat   = "Average"
          region = var.aws_region
          title  = "Backup Storage Usage"
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
            ["AWS/RDS", "BackupRetentionPeriodStorageUsage", "DBInstanceIdentifier", module.rds.db_instance_id]
          ]
          period = 3600
          stat   = "Average"
          region = var.aws_region
          title  = "RDS Backup Storage Usage"
        }
      }
    ]
  })
}

# Disaster Recovery Route53 Configuration
resource "aws_route53_health_check" "primary_region" {
  count = var.enable_dr_health_check ? 1 : 0
  
  fqdn              = var.domain_name
  port              = 443
  type              = "HTTPS"
  resource_path     = "/health"
  failure_threshold = "3"
  request_interval  = "30"
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "disaster-recovery"
  }
}

# Route53 Failover Configuration
resource "aws_route53_record" "dr_failover" {
  count = var.enable_dr_failover ? 1 : 0
  
  zone_id = data.aws_route53_zone.main.zone_id
  name    = var.domain_name
  type    = "A"
  
  failover_routing_policy {
    type = "PRIMARY"
  }
  
  set_identifier = "primary"
  health_check_id = aws_route53_health_check.primary_region[0].id
  
  alias {
    name                   = aws_lb.main.dns_name
    zone_id                = aws_lb.main.zone_id
    evaluate_target_health = true
  }
}

# Disaster Recovery Load Balancer (in DR region)
resource "aws_lb" "dr" {
  count = var.enable_dr_failover ? 1 : 0
  
  provider = aws.dr_region
  
  name               = "kyb-platform-dr-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.dr_alb[0].id]
  subnets            = data.aws_subnets.dr_public[0].ids
  
  enable_deletion_protection = true
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "disaster-recovery"
  }
}

# Security Group for DR Load Balancer
resource "aws_security_group" "dr_alb" {
  count = var.enable_dr_failover ? 1 : 0
  
  provider = aws.dr_region
  
  name        = "kyb-platform-dr-alb-sg"
  description = "Security group for DR load balancer"
  vpc_id      = data.aws_vpc.dr[0].id
  
  ingress {
    description = "HTTPS from Internet"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  
  ingress {
    description = "HTTP from Internet"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "disaster-recovery"
  }
}

# Data sources for DR region
data "aws_vpc" "dr" {
  count = var.enable_dr_failover ? 1 : 0
  
  provider = aws.dr_region
  
  tags = {
    Name = "kyb-platform-dr-vpc"
  }
}

data "aws_subnets" "dr_public" {
  count = var.enable_dr_failover ? 1 : 0
  
  provider = aws.dr_region
  
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.dr[0].id]
  }
  
  filter {
    name   = "tag:Type"
    values = ["public"]
  }
}

# Outputs
output "backup_bucket_name" {
  description = "Name of the backup S3 bucket"
  value       = aws_s3_bucket.backup_storage.bucket
}

output "backup_bucket_arn" {
  description = "ARN of the backup S3 bucket"
  value       = aws_s3_bucket.backup_storage.arn
}

output "backup_replica_bucket_name" {
  description = "Name of the backup replica S3 bucket"
  value       = var.enable_cross_region_backup ? aws_s3_bucket.backup_replica[0].bucket : null
}

output "backup_replica_bucket_arn" {
  description = "ARN of the backup replica S3 bucket"
  value       = var.enable_cross_region_backup ? aws_s3_bucket.backup_replica[0].arn : null
}

output "backup_vault_arn" {
  description = "ARN of the AWS Backup vault"
  value       = var.enable_aws_backup ? aws_backup_vault.main[0].arn : null
}

output "backup_plan_arn" {
  description = "ARN of the AWS Backup plan"
  value       = var.enable_aws_backup ? aws_backup_plan.main[0].arn : null
}

output "dr_load_balancer_dns" {
  description = "DNS name of the DR load balancer"
  value       = var.enable_dr_failover ? aws_lb.dr[0].dns_name : null
}

output "backup_dashboard_url" {
  description = "URL of the backup monitoring CloudWatch dashboard"
  value       = var.enable_backup_monitoring ? "https://${var.aws_region}.console.aws.amazon.com/cloudwatch/home?region=${var.aws_region}#dashboards:name=kyb-platform-backup-monitoring" : null
}
