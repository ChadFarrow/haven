#!/bin/bash

# Crontab Setup Script for Haven Services
# This script sets up automated monitoring and maintenance tasks

echo "📅 Setting up crontab entries for Haven services..."

# Get the current directory (where Haven is installed)
HAVEN_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Create a temporary crontab file
TEMP_CRON=$(mktemp)

# Get existing crontab entries (if any)
crontab -l 2>/dev/null > "$TEMP_CRON" || true

# Add header comment
echo "# Haven Service Management Crontab Entries" >> "$TEMP_CRON"
echo "# Generated on $(date)" >> "$TEMP_CRON"
echo "# Haven directory: $HAVEN_DIR" >> "$TEMP_CRON"
echo "" >> "$TEMP_CRON"

# Add crontab entries
cat >> "$TEMP_CRON" << EOF
# Check services every 5 minutes
*/5 * * * * cd $HAVEN_DIR && ./check-services.sh >> service-checks.log 2>&1

# Restart services if they're down (every 10 minutes)
*/10 * * * * cd $HAVEN_DIR && ./check-services.sh | grep -q "❌" && ./restart-services.sh >> service-restarts.log 2>&1

# Daily service health report at 6 AM
0 6 * * * cd $HAVEN_DIR && echo "=== Daily Haven Service Report ===" >> daily-report.log && ./check-services.sh >> daily-report.log 2>&1

# Weekly log rotation (every Sunday at 2 AM)
0 2 * * 0 cd $HAVEN_DIR && mv haven.log haven.log.old && mv service-monitor.log service-monitor.log.old 2>/dev/null || true

# Monthly SSL certificate check (1st of month at 3 AM)
0 3 1 * * cd $HAVEN_DIR && echo "=== SSL Certificate Check ===" >> ssl-checks.log && openssl s_client -connect podtards.com:443 -servername podtards.com < /dev/null 2>/dev/null | openssl x509 -noout -dates >> ssl-checks.log 2>&1

# Clean up old log files (keep last 30 days)
0 4 * * * cd $HAVEN_DIR && find . -name "*.log.old" -mtime +30 -delete 2>/dev/null || true
EOF

# Install the new crontab
if crontab "$TEMP_CRON"; then
    echo "✅ Crontab entries installed successfully!"
    echo ""
    echo "📋 Installed crontab entries:"
    echo "   • Service health check: Every 5 minutes"
    echo "   • Auto-restart if down: Every 10 minutes"
    echo "   • Daily health report: 6:00 AM daily"
    echo "   • Log rotation: 2:00 AM every Sunday"
    echo "   • SSL certificate check: 3:00 AM on 1st of month"
    echo "   • Log cleanup: 4:00 AM daily (keeps 30 days)"
    echo ""
    echo "📁 Log files will be created in: $HAVEN_DIR"
    echo "   • service-checks.log - Health check results"
    echo "   • service-restarts.log - Restart attempts"
    echo "   • daily-report.log - Daily health reports"
    echo "   • ssl-checks.log - SSL certificate status"
    echo ""
    echo "🔍 To view current crontab: crontab -l"
    echo "📝 To edit crontab manually: crontab -e"
    echo "🗑️  To remove all crontab entries: crontab -r"
else
    echo "❌ Failed to install crontab entries"
    exit 1
fi

# Clean up temporary file
rm "$TEMP_CRON" 