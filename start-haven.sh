#!/bin/bash

# Haven Complete Startup Script
HAVEN_DIR="$(dirname "$0")"
cd "$HAVEN_DIR"

echo "🚀 Starting Haven infrastructure..."

# Check nginx configuration first
echo "📋 Checking nginx configuration..."
if ! ./nginx-config-check.sh; then
    echo "❌ Nginx configuration has issues. Please fix before starting Haven."
    exit 1
fi

# Stop any existing Haven processes
echo "🛑 Stopping existing Haven processes..."
./haven-service.sh stop

# Start Haven service properly
echo "🔄 Starting Haven service..."
./haven-service.sh start

# Check if service started successfully
sleep 3
if ./haven-service.sh status | grep -q "running"; then
    echo "✅ Haven started successfully!"
    echo "🌐 Available at: https://podtards.com"
else
    echo "❌ Haven failed to start. Check logs:"
    tail -20 haven.log
    exit 1
fi 