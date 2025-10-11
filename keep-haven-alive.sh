#!/bin/bash

# Keep Haven Alive Script
# This script runs in the background and restarts Haven if it stops working

echo "🔄 Starting Haven keep-alive monitor..."

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
CHECK_INTERVAL=30  # Check every 30 seconds
LOG_FILE="keep-alive.log"

# Function to log messages
log_message() {
    local message="$1"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] $message" | tee -a "$LOG_FILE"
}

# Function to check if Haven is working properly
check_haven() {
    # Check if process is running
    if ! pgrep -f "haven" > /dev/null; then
        return 1
    fi
    
    # Check if port is listening
    if ! lsof -i :3355 > /dev/null 2>&1; then
        return 1
    fi
    
    # Check if Haven is responding (simple HTTP check)
    if ! curl -s http://localhost:3355 > /dev/null 2>&1; then
        return 1
    fi
    
    return 0
}

# Function to restart Haven
restart_haven() {
    log_message "🔄 Restarting Haven..."
    
    # Kill existing Haven process
    pkill -f "haven" 2>/dev/null
    sleep 2
    
    # Force kill if still running
    if pgrep -f "haven" > /dev/null; then
        pkill -9 -f "haven" 2>/dev/null
        sleep 1
    fi
    
    # Start Haven
    cd "$(dirname "$0")"
    nohup ./haven > haven.log 2>&1 &
    
    # Wait for Haven to start
    local max_attempts=30
    local attempt=1
    while [ $attempt -le $max_attempts ]; do
        if check_haven; then
            log_message "✅ Haven restarted successfully"
            return 0
        fi
        sleep 1
        ((attempt++))
    done
    
    log_message "❌ Failed to restart Haven"
    return 1
}

# Main monitoring loop
log_message "🚀 Starting Haven keep-alive monitor..."

while true; do
    if ! check_haven; then
        log_message "❌ Haven is not working properly, restarting..."
        restart_haven
    else
        log_message "✅ Haven is running correctly"
    fi
    
    sleep $CHECK_INTERVAL
done 