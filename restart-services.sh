#!/bin/bash

# Service Restart Script for Haven
# This script safely restarts Haven and Nginx services

echo "🔄 Restarting Haven and Nginx services..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to check if a process is running
check_process() {
    local process_name=$1
    pgrep -f "$process_name" > /dev/null
}

# Function to wait for a process to start
wait_for_process() {
    local process_name=$1
    local max_attempts=30
    local attempt=1
    
    echo -n "Waiting for $process_name to start..."
    while [ $attempt -le $max_attempts ]; do
        if check_process "$process_name"; then
            echo -e "${GREEN} ✅${NC}"
            return 0
        fi
        echo -n "."
        sleep 1
        ((attempt++))
    done
    echo -e "${RED} ❌${NC}"
    return 1
}

# Stop Haven if running
echo "🛑 Stopping Haven service..."
if check_process "haven"; then
    pkill -f "haven"
    sleep 2
    if check_process "haven"; then
        echo -e "${YELLOW}⚠️  Haven is still running, force killing...${NC}"
        pkill -9 -f "haven"
        sleep 1
    fi
else
    echo -e "${YELLOW}⚠️  Haven was not running${NC}"
fi

# Reload Nginx configuration
echo "🔄 Reloading Nginx configuration..."
if sudo nginx -t > /dev/null 2>&1; then
    sudo nginx -s reload
    echo -e "${GREEN}✅ Nginx configuration reloaded${NC}"
else
    echo -e "${RED}❌ Nginx configuration test failed${NC}"
    exit 1
fi

# Start Haven
echo "🚀 Starting Haven service..."
cd "$(dirname "$0")"
nohup ./haven > haven.log 2>&1 &
HAVEN_PID=$!

# Wait for Haven to start
if wait_for_process "haven"; then
    echo -e "${GREEN}✅ Haven started successfully (PID: $HAVEN_PID)${NC}"
else
    echo -e "${RED}❌ Failed to start Haven${NC}"
    echo "Check haven.log for details"
    exit 1
fi

# Wait a bit for Haven to fully initialize
echo "⏳ Waiting for Haven to initialize..."
sleep 5

# Check if port 3355 is listening
echo "🔍 Checking if Haven is listening on port 3355..."
if lsof -i :3355 > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Haven is listening on port 3355${NC}"
else
    echo -e "${RED}❌ Haven is not listening on port 3355${NC}"
    echo "Check haven.log for details"
    exit 1
fi

echo ""
echo -e "${GREEN}🎉 All services restarted successfully!${NC}"
echo "You can now access your relay at: https://podtards.com" 