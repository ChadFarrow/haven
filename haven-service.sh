#!/bin/bash

# Haven Service Management Script
SERVICE_NAME="com.haven.relay"
PLIST_PATH="$HOME/haven/com.haven.relay.plist"
HAVEN_DIR="$HOME/haven"

case "$1" in
    start)
        echo "Starting Haven service..."
        cd "$HAVEN_DIR"
        launchctl load "$PLIST_PATH"
        ;;
    stop)
        echo "Stopping Haven service..."
        launchctl unload "$PLIST_PATH" 2>/dev/null || true
        pkill -f haven 2>/dev/null || true
        ;;
    restart)
        echo "Restarting Haven service..."
        $0 stop
        sleep 2
        $0 start
        ;;
    status)
        if launchctl list | grep -q "$SERVICE_NAME"; then
            echo "Haven service is running"
            launchctl list | grep "$SERVICE_NAME"
        else
            echo "Haven service is not running"
        fi
        ;;
    import)
        echo "Running Haven import..."
        $0 stop
        sleep 2
        cd "$HAVEN_DIR"
        ./haven --import
        $0 start
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|status|import}"
        exit 1
        ;;
esac