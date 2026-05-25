#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

PID_FILE=".sync-backend.pid"

# Load .env if it exists
load_env() {
    if [ -f ".env" ]; then
        set -a
        source .env
        set +a
        echo -e "${GREEN}Loaded environment from .env${NC}"
    fi
}

# Configure git hooks
setup_githooks() {
    if [ -d ".githooks" ]; then
        git config core.hooksPath .githooks
        echo -e "${GREEN}Git hooks configured${NC}"
    fi
}

# Install dependencies
install_deps() {
    echo -e "${YELLOW}Installing Go dependencies...${NC}"
    cd backend
    go mod tidy
    cd ..

    if [ ! -d "frontend/node_modules" ]; then
        echo -e "${YELLOW}Installing frontend dependencies...${NC}"
        cd frontend
        npm install
        cd ..
    else
        echo -e "${GREEN}Frontend dependencies already installed${NC}"
    fi
}

# Build the Go binary
build_backend() {
    echo -e "${YELLOW}Building Go backend...${NC}"
    cd backend
    go build -o ../server ./cmd/server
    cd ..
}

start_backend() {
    if [ -f "$PID_FILE" ] && kill -0 $(cat "$PID_FILE") 2>/dev/null; then
        echo -e "${YELLOW}Backend is already running (PID $(cat $PID_FILE))${NC}"
        return
    fi

    load_env
    build_backend

    if [ $? -ne 0 ]; then
        echo -e "${RED}Build failed!${NC}"
        exit 1
    fi

    echo -e "${YELLOW}Starting backend server...${NC}"
    nohup ./server > .backend.log 2>&1 &
    echo $! > "$PID_FILE"

    sleep 2
    if kill -0 $(cat "$PID_FILE") 2>/dev/null; then
        echo -e "${GREEN}Backend started (PID $(cat $PID_FILE))${NC}"
        echo -e "${GREEN}Backend API: http://localhost:${SERVER_PORT:-8080}${NC}"
        echo -e "${GREEN}Swagger Docs: http://localhost:${SERVER_PORT:-8080}/swagger/index.html${NC}"
        echo -e "${GREEN}Frontend: http://localhost:3000${NC}"
        echo -e "${GREEN}Logs: .backend.log${NC}"
    else
        echo -e "${RED}Backend failed to start. Check .backend.log${NC}"
        rm -f "$PID_FILE"
        exit 1
    fi
}

stop_backend() {
    if [ ! -f "$PID_FILE" ]; then
        echo -e "${YELLOW}No backend PID file found${NC}"
        return
    fi

    PID=$(cat "$PID_FILE")
    if kill -0 "$PID" 2>/dev/null; then
        echo -e "${YELLOW}Stopping backend (PID $PID)...${NC}"
        kill "$PID" 2>/dev/null
        sleep 1
        if kill -0 "$PID" 2>/dev/null; then
            kill -9 "$PID" 2>/dev/null
        fi
        echo -e "${GREEN}Backend stopped${NC}"
    else
        echo -e "${YELLOW}No running backend process found${NC}"
    fi
    rm -f "$PID_FILE"
}

start_frontend() {
    echo -e "${YELLOW}Starting frontend dev server...${NC}"
    cd frontend
    npm run dev &
    FRONTEND_PID=$!
    cd ..
    echo -e "${GREEN}Frontend starting (PID $FRONTEND_PID)${NC}"
    echo -e "${GREEN}Frontend: http://localhost:3000${NC}"
}

show_status() {
    echo -e "${YELLOW}=== Sync Status ===${NC}"
    
    # Backend
    if [ -f "$PID_FILE" ] && kill -0 $(cat "$PID_FILE") 2>/dev/null; then
        echo -e "Backend:  ${GREEN}Running${NC} (PID $(cat $PID_FILE)) on :${SERVER_PORT:-8080}"
    else
        echo -e "Backend:  ${RED}Stopped${NC}"
    fi

    # Frontend
    FRONTEND_PID=$(lsof -ti :3000 2>/dev/null | head -1)
    if [ -n "$FRONTEND_PID" ]; then
        echo -e "Frontend: ${GREEN}Running${NC} (PID $FRONTEND_PID) on :3000"
    else
        echo -e "Frontend: ${RED}Stopped${NC}"
    fi

    # Database
    DB_PID=$(lsof -ti :5432 2>/dev/null | head -1)
    if [ -n "$DB_PID" ]; then
        echo -e "Database: ${GREEN}Running${NC} (PID $DB_PID) on :5432"
    else
        echo -e "Database: ${RED}Stopped${NC}"
    fi
}

case "${1:-setup}" in
    setup)
        echo -e "${YELLOW}Starting sync setup...${NC}"
        setup_githooks
        install_deps
        build_backend
        echo -e "${GREEN}Setup complete! Run '$0 start' to start the servers${NC}"
        ;;
    start)
        echo -e "${YELLOW}Starting sync...${NC}"
        setup_githooks
        install_deps
        start_backend
        start_frontend
        echo -e "${GREEN}All services started!${NC}"
        echo -e "${GREEN}Frontend: http://localhost:3000${NC}"
        echo -e "${GREEN}Backend:  http://localhost:${SERVER_PORT:-8080}${NC}"
        ;;
    stop)
        echo -e "${YELLOW}Stopping sync...${NC}"
        stop_backend
        FRONTEND_PID=$(lsof -ti :3000 2>/dev/null | head -1)
        if [ -n "$FRONTEND_PID" ]; then
            echo -e "${YELLOW}Stopping frontend (PID $FRONTEND_PID)...${NC}"
            kill "$FRONTEND_PID" 2>/dev/null
            echo -e "${GREEN}Frontend stopped${NC}"
        fi
        echo -e "${GREEN}All services stopped${NC}"
        ;;
    restart)
        echo -e "${YELLOW}Restarting sync...${NC}"
        stop_backend
        FRONTEND_PID=$(lsof -ti :3000 2>/dev/null | head -1)
        if [ -n "$FRONTEND_PID" ]; then
            kill "$FRONTEND_PID" 2>/dev/null
        fi
        sleep 1
        start_backend
        start_frontend
        echo -e "${GREEN}All services restarted!${NC}"
        ;;
    status)
        show_status
        ;;
    *)
        echo -e "Usage: $0 {setup|start|stop|restart|status}"
        echo -e ""
        echo -e "Commands:"
        echo -e "  setup    Install dependencies and build (default)"
        echo -e "  start    Build and start all services (backend + frontend)"
        echo -e "  stop     Stop all services"
        echo -e "  restart  Restart all services"
        echo -e "  status   Show running status of all services"
        exit 1
        ;;
esac
