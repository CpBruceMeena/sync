#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Starting sync setup...${NC}"

# Configure git hooks
if [ -d ".githooks" ]; then
    git config core.hooksPath .githooks
    echo -e "${GREEN}Git hooks configured${NC}"
fi

# Load .env if it exists
if [ -f ".env" ]; then
    set -a
    source .env
    set +a
    echo -e "${GREEN}Loaded environment from .env${NC}"
fi"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    exit 1
fi

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo -e "${RED}Error: Node.js is not installed${NC}"
    exit 1
fi

# Install Go dependencies
echo -e "${YELLOW}Installing Go dependencies...${NC}"
cd backend
go mod tidy
cd ..

# Install frontend dependencies
if [ ! -d "frontend/node_modules" ]; then
    echo -e "${YELLOW}Installing frontend dependencies...${NC}"
    cd frontend
    npm install
    cd ..
else
    echo -e "${GREEN}Frontend dependencies already installed${NC}"
fi

# Build the Go binary
echo -e "${YELLOW}Building Go backend...${NC}"
cd backend
go build -o sync ./cmd/server
cd ..

# Check if build was successful
if [ $? -eq 0 ]; then
    echo -e "${GREEN}Build successful!${NC}"
    echo -e "${YELLOW}Starting backend server...${NC}"
    echo -e "${GREEN}Backend API: http://localhost:${SERVER_PORT:-8080}${NC}"
    echo -e "${GREEN}Swagger Docs: http://localhost:${SERVER_PORT:-8080}/swagger/index.html${NC}"
    echo -e "${YELLOW}Start frontend separately: cd frontend && npm run dev${NC}"
    echo -e "${YELLOW}Press Ctrl+C to stop the server${NC}"
    cd backend
    ./sync
else
    echo -e "${RED}Build failed!${NC}"
    exit 1
fi
