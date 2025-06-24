# Custom Haven Modifications

This document describes the custom modifications made to the Haven relay project.

## Overview
Custom enhancements to improve reliability, monitoring, and sleep/wake handling for the Haven Nostr relay server.

## Modifications Made

### 1. Enhanced LaunchAgent Configuration (`com.haven.relay.plist`)
- Added `ProcessType: Background` for better background operation
- Added `ThrottleInterval: 10` to prevent excessive restarts
- Added `ExitTimeOut: 10` for graceful shutdown handling
- Reordered properties for better organization

### 2. Keep-Alive Monitoring Script (`keep-haven-alive.sh`)
- **Purpose**: Continuously monitors Haven and automatically restarts it if it stops working
- **Features**:
  - Checks process status, port availability, and HTTP response
  - Automatic restart with graceful shutdown handling
  - Comprehensive logging with timestamps
  - Configurable check intervals (default: 30 seconds)
  - Color-coded output for better visibility

### 3. Enhanced Crontab Management (`update-crontab.sh`)
- **Purpose**: Improved scheduling for better sleep/wake handling and service reliability
- **Enhancements**:
  - More frequent service checks (every 2 minutes vs 5)
  - Faster auto-restart (every 5 minutes vs 10)
  - Force restart every 6 hours to prevent sleep-related issues
  - Daily health reports at 6:00 AM
  - Weekly log rotation
  - Monthly SSL certificate checks
  - Automatic log cleanup (keeps 30 days)

### 4. Monitoring Interface
- Added monitoring interface commits (see git log for details)
- Enhanced service monitoring capabilities

### 5. Additional Log Files
- `daily-report.log`: Daily service health reports
- `service-checks.log`: Service check results
- `service-restarts.log`: Service restart events
- `scheduled-restarts.log`: Scheduled restart events

## Installation Instructions

### To Apply These Changes:
1. **Backup Branch**: Switch to `custom-haven-backup` branch
   ```bash
   git checkout custom-haven-backup
   ```

2. **Apply Patch**: Use the patch file
   ```bash
   git apply custom-haven-changes.patch
   ```

3. **Manual Setup**: Follow the documentation above to manually implement changes

### To Use Keep-Alive Script:
```bash
chmod +x keep-haven-alive.sh
./keep-haven-alive.sh &
```

### To Update Crontab:
```bash
chmod +x update-crontab.sh
./update-crontab.sh
```

## Benefits

1. **Improved Reliability**: Automatic restart on failures
2. **Better Sleep/Wake Handling**: Prevents issues when Mac goes to sleep
3. **Enhanced Monitoring**: Comprehensive logging and health checks
4. **Reduced Manual Intervention**: Automated maintenance tasks
5. **Better Resource Management**: Automatic log rotation and cleanup

## Notes

- These modifications are designed for personal use and may not be suitable for all environments
- The keep-alive script runs continuously and should be monitored
- Crontab changes are system-wide and may affect other scheduled tasks
- Always test in a development environment before applying to production

## Reverting Changes

To revert to the original state:
```bash
git checkout master
git reset --hard origin/master
```

## Backup Information

- **Backup Branch**: `custom-haven-backup`
- **Patch File**: `custom-haven-changes.patch`
- **Created**: $(date)
- **Original Repository**: Haven relay project 