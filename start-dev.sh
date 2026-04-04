#!/bin/bash

# ExplorAItion Development Server Starter
# Runs both the Go backend and Next.js frontend

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Set PATH to include local binaries
export PATH=/usr/local/bin:/usr/local/Cellar/go/1.26.1/bin:$PATH

# Get the script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Ensure .env exists
if [ ! -f .env ]; then
    echo -e "${RED}✗ Error: .env file not found in $SCRIPT_DIR${NC}"
    echo "Please ensure .env is in the project root directory"
    exit 1
fi

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}  ExplorAItion Development Server Starter${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Function to cleanup on exit
cleanup() {
    echo ""
    echo -e "${YELLOW}Shutting down servers...${NC}"

    if [ ! -z "$BACKEND_PID" ]; then
        echo -e "  Stopping backend (PID: $BACKEND_PID)..."
        kill $BACKEND_PID 2>/dev/null || true
    fi

    if [ ! -z "$FRONTEND_PID" ]; then
        echo -e "  Stopping frontend (PID: $FRONTEND_PID)..."
        kill $FRONTEND_PID 2>/dev/null || true
    fi

    echo -e "${GREEN}✓ Servers stopped${NC}"
    exit 0
}

# Trap signals for graceful shutdown
trap cleanup SIGINT SIGTERM

# Check if ports are available
check_port() {
    local port=$1
    local name=$2

    if lsof -i ":$port" >/dev/null 2>&1; then
        echo -e "${YELLOW}⚠ Warning: Port $port is already in use${NC}"
        echo -e "${YELLOW}  $name might not start properly${NC}"
        echo ""
    fi
}

echo "Checking prerequisites..."
check_port 8080 "Backend (Go)"
check_port 3000 "Frontend (Next.js)"

# Verify binaries exist
if [ ! -f "./exploraition_new" ]; then
    echo -e "${YELLOW}Building backend binary...${NC}"
    go build -o exploraition_new main.go
    echo -e "${GREEN}✓ Backend built${NC}"
else
    echo -e "${GREEN}✓ Backend binary found${NC}"
fi

if [ ! -d "./frontend/node_modules" ]; then
    echo -e "${YELLOW}Installing frontend dependencies...${NC}"
    npm install --prefix frontend
    echo -e "${GREEN}✓ Frontend dependencies installed${NC}"
else
    echo -e "${GREEN}✓ Frontend dependencies found${NC}"
fi

echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}Starting servers...${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Start backend
echo -e "${BLUE}Backend:${NC}"
echo "  Command: ./exploraition_new"
echo "  Port: 8080"
echo "  .env: Loaded"
./exploraition_new > backend.log 2>&1 &
BACKEND_PID=$!
echo -e "${GREEN}  PID: $BACKEND_PID${NC}"
echo ""

# Wait a moment for backend to start
sleep 2

# Start frontend
echo -e "${BLUE}Frontend:${NC}"
echo "  Command: npm run dev"
echo "  Port: 3000"
npm run dev --prefix frontend > frontend.log 2>&1 &
FRONTEND_PID=$!
echo -e "${GREEN}  PID: $FRONTEND_PID${NC}"
echo ""

# Wait for frontend to start
sleep 3

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✓ Both servers started!${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${YELLOW}Access your application:${NC}"
echo -e "  🌐 Frontend:  ${GREEN}http://localhost:3000${NC}"
echo -e "  🔧 Backend:   ${GREEN}http://localhost:8080${NC}"
echo ""
echo -e "${YELLOW}Available pages:${NC}"
echo -e "  • Search: ${GREEN}http://localhost:3000/search${NC}"
echo -e "  • Itinerary: ${GREEN}http://localhost:3000/itinerary${NC}"
echo ""
echo -e "${YELLOW}Log files:${NC}"
echo -e "  Backend:  ./backend.log"
echo -e "  Frontend: ./frontend.log"
echo ""
echo -e "${YELLOW}To stop servers: Press Ctrl+C${NC}"
echo ""

# Keep the script running
wait
