#!/bin/bash

# Service Health Check Script for Haven
# This script checks if Haven and Nginx are running properly

echo "🔍 Checking Haven and Nginx services..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to check if a process is running
check_process() {
    local process_name=$1
    local display_name=$2
    if pgrep -f "$process_name" > /dev/null; then
        echo -e "${GREEN}✅ $display_name is running${NC}"
        return 0
    else
        echo -e "${RED}❌ $display_name is not running${NC}"
        return 1
    fi
}

# Function to check if a port is listening
check_port() {
    local port=$1
    local service_name=$2
    if lsof -i :$port > /dev/null 2>&1; then
        echo -e "${GREEN}✅ $service_name is listening on port $port${NC}"
        return 0
    else
        echo -e "${RED}❌ $service_name is not listening on port $port${NC}"
        return 1
    fi
}

# Function to check SSL certificate
check_ssl() {
    local domain=$1
    if openssl s_client -connect $domain:443 -servername $domain < /dev/null 2>/dev/null | openssl x509 -noout -dates > /dev/null 2>&1; then
        echo -e "${GREEN}✅ SSL certificate for $domain is valid${NC}"
        return 0
    else
        echo -e "${RED}❌ SSL certificate for $domain is invalid or not accessible${NC}"
        return 1
    fi
}

# Initialize counters
total_checks=0
passed_checks=0

echo ""
echo "=== Process Checks ==="
check_process "haven" "Haven Service" && ((passed_checks++)) || true
((total_checks++))

check_process "nginx" "Nginx" && ((passed_checks++)) || true
((total_checks++))

echo ""
echo "=== Port Checks ==="
check_port "3355" "Haven" && ((passed_checks++)) || true
((total_checks++))

check_port "80" "Nginx HTTP" && ((passed_checks++)) || true
((total_checks++))

check_port "443" "Nginx HTTPS" && ((passed_checks++)) || true
((total_checks++))

echo ""
echo "=== SSL Certificate Check ==="
check_ssl "podtards.com" && ((passed_checks++)) || true
((total_checks++))

echo ""
echo "=== Nginx Configuration ==="
if sudo nginx -t > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Nginx configuration is valid${NC}"
    ((passed_checks++))
else
    echo -e "${RED}❌ Nginx configuration has errors${NC}"
fi
((total_checks++))

echo ""
echo "=== Summary ==="
echo "Passed: $passed_checks/$total_checks checks"

if [ $passed_checks -eq $total_checks ]; then
    echo -e "${GREEN}🎉 All services are running correctly!${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠️  Some issues detected. Check the output above.${NC}"
    exit 1
fi 