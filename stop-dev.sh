#!/bin/bash

# Stop ExplorAItion development servers

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}  Stopping ExplorAItion Servers${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Kill processes on specific ports
kill_by_port() {
    local port=$1
    local name=$2

    echo -e "Stopping ${YELLOW}$name${NC} on port $port..."

    # Find and kill process on port
    local pids=$(lsof -ti:$port 2>/dev/null || true)

    if [ -z "$pids" ]; then
        echo -e "  ${YELLOW}No process found on port $port${NC}"
    else
        echo $pids | xargs kill -9 2>/dev/null
        echo -e "  ${GREEN}✓ Killed (PID: $pids)${NC}"
    fi
}

kill_by_port 8080 "Backend (Go)"
kill_by_port 3000 "Frontend (Next.js)"

echo ""
echo -e "${GREEN}✓ All servers stopped${NC}"
echo ""
