#!/bin/bash

# Continuous Service Monitoring Script for Haven
# This script monitors Haven and Nginx services and restarts them if they fail

echo "👁️  Starting continuous monitoring of Haven and Nginx services..."
echo "Press Ctrl+C to stop monitoring"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
CHECK_INTERVAL=30  # Check every 30 seconds
LOG_FILE="service-monitor.log"

# Function to log messages
log_message() {
    local message="$1"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] $message" | tee -a "$LOG_FILE"
}

# Function to check if a process is running
check_process() {
    local process_name=$1
    pgrep -f "$process_name" > /dev/null
}

# Function to check if a port is listening
check_port() {
    local port=$1
    lsof -i :$port > /dev/null 2>&1
}

# Function to restart Haven
restart_haven() {
    log_message "🔄 Restarting Haven service..."
    
    # Stop Haven if running
    if check_process "haven"; then
        pkill -f "haven"
        sleep 2
        if check_process "haven"; then
            pkill -9 -f "haven"
            sleep 1
        fi
    fi
    
    # Start Haven
    cd "$(dirname "$0")"
    nohup ./haven > haven.log 2>&1 &
    
    # Wait for Haven to start
    local max_attempts=30
    local attempt=1
    while [ $attempt -le $max_attempts ]; do
        if check_process "haven" && check_port "3355"; then
            log_message "✅ Haven restarted successfully"
            return 0
        fi
        sleep 1
        ((attempt++))
    done
    
    log_message "❌ Failed to restart Haven"
    return 1
}

# Function to reload Nginx
reload_nginx() {
    log_message "🔄 Reloading Nginx configuration..."
    if sudo nginx -t > /dev/null 2>&1; then
        sudo nginx -s reload
        log_message "✅ Nginx configuration reloaded"
        return 0
    else
        log_message "❌ Nginx configuration test failed"
        return 1
    fi
}

# Main monitoring loop
while true; do
    # Check Haven
    if ! check_process "haven"; then
        log_message "❌ Haven process is not running"
        restart_haven
    elif ! check_port "3355"; then
        log_message "❌ Haven is not listening on port 3355"
        restart_haven
    else
        log_message "✅ Haven is running correctly"
    fi
    
    # Check Nginx
    if ! check_process "nginx"; then
        log_message "❌ Nginx process is not running"
        # Note: Nginx restart might require sudo, so we'll just log the issue
        log_message "⚠️  Nginx restart requires manual intervention"
    elif ! check_port "80" || ! check_port "443"; then
        log_message "❌ Nginx is not listening on required ports"
        reload_nginx
    else
        log_message "✅ Nginx is running correctly"
    fi
    
    # Check SSL certificate (less frequently)
    if [ $((SECONDS % 300)) -eq 0 ]; then  # Every 5 minutes
        if ! openssl s_client -connect podtards.com:443 -servername podtards.com < /dev/null 2>/dev/null | openssl x509 -noout -dates > /dev/null 2>&1; then
            log_message "❌ SSL certificate for podtards.com is invalid or not accessible"
        else
            log_message "✅ SSL certificate is valid"
        fi
    fi
    
    # Wait before next check
    sleep $CHECK_INTERVAL
done 