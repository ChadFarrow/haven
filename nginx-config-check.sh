#!/bin/bash

# Nginx Configuration Check Script
NGINX_SERVERS_DIR="/opt/homebrew/etc/nginx/servers"

echo "Checking nginx configuration for podtards.com..."

# Check for duplicate server_name entries
echo "Looking for duplicate server_name entries:"
grep -r "server_name.*podtards.com" "$NGINX_SERVERS_DIR" 2>/dev/null || echo "No duplicate entries found"

echo ""
echo "Testing nginx configuration:"
nginx -t

echo ""
echo "Active nginx server blocks for podtards.com:"
ls -la "$NGINX_SERVERS_DIR"/*podtards* 2>/dev/null || echo "No podtards config files found"