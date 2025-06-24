#!/bin/bash

# Update Crontab for Better Sleep/Wake Handling
echo "📅 Updating crontab for better sleep/wake handling..."

# Get the current directory
HAVEN_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Create a temporary crontab file
TEMP_CRON=$(mktemp)

# Get existing crontab entries (if any)
crontab -l 2>/dev/null > "$TEMP_CRON" || true

# Remove old Haven entries and add new ones
grep -v "haven" "$TEMP_CRON" > "$TEMP_CRON.tmp" && mv "$TEMP_CRON.tmp" "$TEMP_CRON"

# Add new crontab entries with better sleep/wake handling
cat >> "$TEMP_CRON" << EOF

# Haven Service Management (Updated for Sleep/Wake)
# Check services every 2 minutes (more frequent)
*/2 * * * * cd $HAVEN_DIR && ./check-services.sh >> service-checks.log 2>&1

# Restart services if they're down (every 5 minutes)
*/5 * * * * cd $HAVEN_DIR && ./check-services.sh | grep -q "❌" && ./restart-services.sh >> service-restarts.log 2>&1

# Force restart Haven every 6 hours to prevent sleep-related issues
0 */6 * * * cd $HAVEN_DIR && ./restart-services.sh >> scheduled-restarts.log 2>&1

# Daily service health report at 6 AM
0 6 * * * cd $HAVEN_DIR && echo "=== Daily Haven Service Report ===" >> daily-report.log && ./check-services.sh >> daily-report.log 2>&1

# Weekly log rotation (every Sunday at 2 AM)
0 2 * * 0 cd $HAVEN_DIR && mv haven.log haven.log.old && mv service-monitor.log service-monitor.log.old 2>/dev/null || true

# Monthly SSL certificate check (1st of month at 3 AM)
0 3 1 * * cd $HAVEN_DIR && echo "=== SSL Certificate Check ===" >> ssl-checks.log && openssl s_client -connect podtards.com:443 -servername podtards.com < /dev/null 2>/dev/null | openssl x509 -noout -dates >> ssl-checks.log 2>&1

# Clean up old log files (keep last 30 days)
0 4 * * * cd $HAVEN_DIR && find . -name "*.log.old" -mtime +30 -delete 2>/dev/null || true
EOF

# Install the updated crontab
if crontab "$TEMP_CRON"; then
    echo "✅ Updated crontab entries installed successfully!"
    echo ""
    echo "📋 Updated crontab entries:"
    echo "   • Service health check: Every 2 minutes (was 5)"
    echo "   • Auto-restart if down: Every 5 minutes (was 10)"
    echo "   • Force restart: Every 6 hours (new - prevents sleep issues)"
    echo "   • Daily health report: 6:00 AM daily"
    echo "   • Log rotation: 2:00 AM every Sunday"
    echo "   • SSL certificate check: 3:00 AM on 1st of month"
    echo "   • Log cleanup: 4:00 AM daily (keeps 30 days)"
    echo ""
    echo "💡 The force restart every 6 hours will help prevent sleep-related issues"
else
    echo "❌ Failed to update crontab entries"
    exit 1
fi

# Clean up temporary file
rm "$TEMP_CRON" 