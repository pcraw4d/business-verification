# Admin Dashboard Guide

## Overview

The Admin Dashboard provides system monitoring and management capabilities for administrators. It includes memory monitoring, system metrics, and administrative controls.

## Access

The admin dashboard is accessible at `/admin` and requires admin role privileges.

## Features

### Memory Monitoring

- **Memory Profile**: View current memory usage and heap allocation
- **Memory History**: Track memory usage over time with interactive charts
- **GC Cycles**: Monitor garbage collection cycles
- **Memory Optimization**: Trigger memory optimization manually

### System Metrics

- **CPU Usage**: Monitor CPU utilization
- **Memory Thresholds**: View and manage memory thresholds
- **Request Rate**: Track API request rates
- **Error Rate**: Monitor system error rates

## Usage

1. Navigate to `/admin` in your browser
2. The dashboard will automatically check your admin privileges
3. View real-time metrics and memory profiles
4. Use the "Optimize Memory" button to trigger memory optimization
5. Use the "Refresh Metrics" button to manually refresh all metrics

## API Endpoints

- `GET /api/v1/memory/profile` - Get current memory profile
- `GET /api/v1/memory/profile/history` - Get memory history
- `POST /api/v1/memory/optimize` - Optimize memory
- `GET /api/v1/thresholds` - Get system thresholds
- `PUT /api/v1/thresholds` - Update thresholds
- `GET /api/v1/system` - Get system information

## Authentication

All admin endpoints require:
- Valid JWT token in Authorization header
- Admin role in token claims

## Troubleshooting

If you cannot access the admin dashboard:
1. Verify your user has admin role
2. Check that your JWT token includes the admin role
3. Ensure the token is not expired

