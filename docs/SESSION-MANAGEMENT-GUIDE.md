# Session Management Guide

## Overview

Session Management allows users to view and manage their active sessions across different devices and browsers.

## Features

- View all active sessions
- View session details (device, IP address, user agent)
- Switch between sessions
- Terminate sessions
- View session metrics

## Access

Navigate to `/sessions` to access the session management page.

## Usage

### Viewing Sessions

1. Navigate to the Sessions page
2. View all active sessions in the list
3. See session details including:
   - Device information
   - IP address
   - User agent
   - Creation time
   - Last activity

### Managing Sessions

- **Switch Session**: Click "Switch" to switch to a different session
- **Terminate Session**: Click "Terminate" to end a session
- **Create Session**: Click "Create New Session" to create a new session

## Session Metrics

The dashboard displays:
- Total active sessions
- Total sessions
- Average session duration
- Last activity time

## API Endpoints

- `GET /api/v1/sessions` - List all sessions
- `POST /api/v1/sessions` - Create new session
- `DELETE /api/v1/sessions` - Delete/terminate session
- `GET /api/v1/sessions/metrics` - Get session metrics
- `GET /api/v1/sessions/activity` - Get session activity

## Security

- Sessions are automatically tracked when users log in
- Sessions expire after a period of inactivity
- Users can terminate their own sessions
- Admin users can view all sessions

## Best Practices

1. Regularly review active sessions
2. Terminate sessions from unknown devices
3. Use session switching for multi-device access
4. Monitor session metrics for security

