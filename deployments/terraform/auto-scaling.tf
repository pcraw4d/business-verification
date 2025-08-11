# KYB Platform - Auto Scaling Configuration
# Comprehensive auto-scaling setup for EKS and application workloads

# EKS Cluster Autoscaler
resource "aws_eks_addon" "cluster_autoscaler" {
  cluster_name = module.eks.cluster_name
  addon_name   = "aws-cluster-autoscaler"
  
  addon_version = "v1.28.1-eksbuild.1"
  
  resolve_conflicts_on_create = "OVERWRITE"
  resolve_conflicts_on_update = "OVERWRITE"
  
  configuration_values = jsonencode({
    "clusterName" = module.eks.cluster_name
    "autoDiscovery" = {
      "clusterName" = module.eks.cluster_name
    }
    "awsRegion" = var.aws_region
    "rbac" = {
      "serviceAccount" = {
        "create" = true
        "annotations" = {
          "eks.amazonaws.com/role-arn" = aws_iam_role.cluster_autoscaler.arn
        }
      }
    }
    "replicaCount" = 1
    "resources" = {
      "requests" = {
        "cpu"    = "100m"
        "memory" = "300Mi"
      }
      "limits" = {
        "cpu"    = "100m"
        "memory" = "300Mi"
      }
    }
    "nodeSelector" = {
      "kubernetes.io/os" = "linux"
    }
    "tolerations" = [
      {
        "key"    = "node.kubernetes.io/not-ready"
        "effect" = "NoExecute"
        "operator" = "Exists"
      },
      {
        "key"    = "node.kubernetes.io/unreachable"
        "effect" = "NoExecute"
        "operator" = "Exists"
      }
    ]
  })
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "autoscaler"
  }
}

# IAM Role for Cluster Autoscaler
resource "aws_iam_role" "cluster_autoscaler" {
  name = "kyb-platform-cluster-autoscaler"
  
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRoleWithWebIdentity"
        Effect = "Allow"
        Principal = {
          Federated = module.eks.oidc_provider_arn
        }
        Condition = {
          StringEquals = {
            "${module.eks.oidc_provider}:aud" : "sts.amazonaws.com",
            "${module.eks.oidc_provider}:sub" : "system:serviceaccount:kube-system:cluster-autoscaler"
          }
        }
      }
    ]
  })
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "autoscaler"
  }
}

# IAM Policy for Cluster Autoscaler
resource "aws_iam_role_policy" "cluster_autoscaler" {
  name = "kyb-platform-cluster-autoscaler-policy"
  role = aws_iam_role.cluster_autoscaler.id
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "autoscaling:DescribeAutoScalingGroups",
          "autoscaling:DescribeAutoScalingInstances",
          "autoscaling:DescribeLaunchConfigurations",
          "autoscaling:DescribeTags",
          "autoscaling:SetDesiredCapacity",
          "autoscaling:TerminateInstanceInAutoScalingGroup",
          "ec2:DescribeLaunchTemplateVersions"
        ]
        Resource = "*"
      }
    ]
  })
}

# Horizontal Pod Autoscaler for API Deployment
resource "kubernetes_horizontal_pod_autoscaler" "api" {
  metadata {
    name      = "kyb-platform-api-hpa"
    namespace = "default"
    
    labels = {
      app       = "kyb-platform-api"
      component = "autoscaler"
    }
  }
  
  spec {
    max_replicas = var.api_max_replicas
    min_replicas = var.api_min_replicas
    
    target_cpu_utilization_percentage = var.api_target_cpu_utilization
    
    scale_target_ref {
      api_version = "apps/v1"
      kind        = "Deployment"
      name        = "kyb-platform-api"
    }
    
    # Custom metrics for scaling
    metric {
      type = "Resource"
      resource {
        name = "cpu"
        target {
          type                = "Utilization"
          average_utilization = var.api_target_cpu_utilization
        }
      }
    }
    
    metric {
      type = "Resource"
      resource {
        name = "memory"
        target {
          type                = "Utilization"
          average_utilization = var.api_target_memory_utilization
        }
      }
    }
    
    # Custom metrics for request rate
    metric {
      type = "Object"
      object {
        metric {
          name = "requests_per_second"
        }
        described_object {
          api_version = "v1"
          kind        = "Service"
          name        = "kyb-platform-api"
        }
        target {
          type  = "AverageValue"
          value = var.api_target_rps
        }
      }
    }
    
    behavior {
      scale_up {
        stabilization_window_seconds = 60
        select_policy               = "Max"
        policies {
          type          = "Percent"
          value         = 100
          period_seconds = 15
        }
      }
      
      scale_down {
        stabilization_window_seconds = 300
        select_policy               = "Min"
        policies {
          type          = "Percent"
          value         = 10
          period_seconds = 60
        }
      }
    }
  }
}

# Vertical Pod Autoscaler for API Deployment
resource "kubernetes_vertical_pod_autoscaler" "api" {
  metadata {
    name      = "kyb-platform-api-vpa"
    namespace = "default"
    
    labels = {
      app       = "kyb-platform-api"
      component = "autoscaler"
    }
  }
  
  spec {
    target_ref {
      api_version = "apps/v1"
      kind        = "Deployment"
      name        = "kyb-platform-api"
    }
    
    update_policy {
      update_mode = "Auto"
    }
    
    resource_policy {
      container_policies {
        container_name = "kyb-platform-api"
        min_allowed {
          cpu    = "100m"
          memory = "128Mi"
        }
        max_allowed {
          cpu    = "1"
          memory = "1Gi"
        }
        controlled_resources = ["cpu", "memory"]
      }
    }
  }
}

# Application Auto Scaling for RDS
resource "aws_appautoscaling_target" "rds" {
  count              = var.enable_rds_autoscaling ? 1 : 0
  max_capacity       = var.rds_max_capacity
  min_capacity       = var.rds_min_capacity
  resource_id        = "cluster:${module.rds.db_cluster_id}"
  scalable_dimension = "rds:db:ReadReplicaCount"
  service_namespace  = "rds"
}

resource "aws_appautoscaling_policy" "rds_cpu" {
  count              = var.enable_rds_autoscaling ? 1 : 0
  name               = "kyb-platform-rds-cpu-autoscaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.rds[0].resource_id
  scalable_dimension = aws_appautoscaling_target.rds[0].scalable_dimension
  service_namespace  = aws_appautoscaling_target.rds[0].service_namespace
  
  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "RDSReaderAverageCPUUtilization"
    }
    target_value = var.rds_target_cpu_utilization
  }
}

# Application Auto Scaling for ElastiCache Redis
resource "aws_appautoscaling_target" "redis" {
  count              = var.enable_redis_autoscaling ? 1 : 0
  max_capacity       = var.redis_max_capacity
  min_capacity       = var.redis_min_capacity
  resource_id        = "replication-group/${aws_elasticache_replication_group.redis.id}"
  scalable_dimension = "elasticache:replication-group:NodeGroups"
  service_namespace  = "elasticache"
}

resource "aws_appautoscaling_policy" "redis_cpu" {
  count              = var.enable_redis_autoscaling ? 1 : 0
  name               = "kyb-platform-redis-cpu-autoscaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.redis[0].resource_id
  scalable_dimension = aws_appautoscaling_target.redis[0].scalable_dimension
  service_namespace  = aws_appautoscaling_target.redis[0].service_namespace
  
  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ElastiCacheReplicaEngineCPUUtilization"
    }
    target_value = var.redis_target_cpu_utilization
  }
}

# CloudWatch Alarms for Auto Scaling
resource "aws_cloudwatch_metric_alarm" "cluster_cpu_high" {
  alarm_name          = "kyb-platform-cluster-cpu-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "CPUUtilization"
  namespace           = "AWS/ECS"
  period              = "300"
  statistic           = "Average"
  threshold           = var.cluster_cpu_threshold
  alarm_description   = "This metric monitors cluster CPU utilization"
  
  dimensions = {
    ClusterName = module.eks.cluster_name
  }
  
  alarm_actions = [aws_sns_topic.alerts.arn]
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "autoscaler"
  }
}

resource "aws_cloudwatch_metric_alarm" "cluster_memory_high" {
  alarm_name          = "kyb-platform-cluster-memory-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "MemoryUtilization"
  namespace           = "AWS/ECS"
  period              = "300"
  statistic           = "Average"
  threshold           = var.cluster_memory_threshold
  alarm_description   = "This metric monitors cluster memory utilization"
  
  dimensions = {
    ClusterName = module.eks.cluster_name
  }
  
  alarm_actions = [aws_sns_topic.alerts.arn]
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "autoscaler"
  }
}

resource "aws_cloudwatch_metric_alarm" "api_response_time_high" {
  alarm_name          = "kyb-platform-api-response-time-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "TargetResponseTime"
  namespace           = "AWS/ApplicationELB"
  period              = "300"
  statistic           = "Average"
  threshold           = var.api_response_time_threshold
  alarm_description   = "This metric monitors API response time"
  
  dimensions = {
    LoadBalancer = aws_lb.main.arn_suffix
    TargetGroup  = aws_lb_target_group.api.arn_suffix
  }
  
  alarm_actions = [aws_sns_topic.alerts.arn]
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "autoscaler"
  }
}

# Auto Scaling Dashboard
resource "aws_cloudwatch_dashboard" "autoscaling" {
  dashboard_name = "kyb-platform-autoscaling"
  
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
          title  = "Cluster Resource Utilization"
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
            [".", "CurrConnections", ".", "."]
          ]
          period = 300
          stat   = "Average"
          region = var.aws_region
          title  = "Redis Performance"
        }
      }
    ]
  })
}

# Auto Scaling Notification SNS Topic
resource "aws_sns_topic" "autoscaling_notifications" {
  name = "kyb-platform-autoscaling-notifications"
  
  tags = {
    Environment = var.environment
    Project     = "kyb-platform"
    Component   = "notifications"
  }
}

# Auto Scaling Notification Subscriptions
resource "aws_sns_topic_subscription" "autoscaling_email" {
  count     = length(var.autoscaling_emails)
  topic_arn = aws_sns_topic.autoscaling_notifications.arn
  protocol  = "email"
  endpoint  = var.autoscaling_emails[count.index]
}

# Auto Scaling Group Notifications
resource "aws_autoscaling_notification" "eks_notifications" {
  count = length(var.eks_managed_node_groups)
  
  group_names = [for group in module.eks.eks_managed_node_groups : group.resource_id]
  
  notifications = [
    "autoscaling:EC2_INSTANCE_LAUNCH",
    "autoscaling:EC2_INSTANCE_TERMINATE",
    "autoscaling:EC2_INSTANCE_LAUNCH_ERROR",
    "autoscaling:EC2_INSTANCE_TERMINATE_ERROR"
  ]
  
  topic_arn = aws_sns_topic.autoscaling_notifications.arn
}

# Outputs
output "cluster_autoscaler_role_arn" {
  description = "ARN of the cluster autoscaler IAM role"
  value       = aws_iam_role.cluster_autoscaler.arn
}

output "autoscaling_dashboard_url" {
  description = "URL of the auto scaling CloudWatch dashboard"
  value       = "https://${var.aws_region}.console.aws.amazon.com/cloudwatch/home?region=${var.aws_region}#dashboards:name=kyb-platform-autoscaling"
}

output "autoscaling_notifications_topic_arn" {
  description = "ARN of the auto scaling notifications SNS topic"
  value       = aws_sns_topic.autoscaling_notifications.arn
}
