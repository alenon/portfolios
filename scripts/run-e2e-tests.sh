#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="docker-compose.e2e.yml"
PROJECT_NAME="portfolios-e2e"
TEST_TIMEOUT="${TEST_TIMEOUT:-10m}"
CLEANUP="${CLEANUP:-true}"

echo -e "${GREEN}=== Portfolio E2E Test Runner ===${NC}"
echo ""

# Function to cleanup
cleanup() {
    if [ "$CLEANUP" = "true" ]; then
        echo -e "${YELLOW}Cleaning up Docker containers...${NC}"
        docker compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" down -v --remove-orphans
        echo -e "${GREEN}Cleanup complete${NC}"
    else
        echo -e "${YELLOW}Skipping cleanup (CLEANUP=false)${NC}"
    fi
}

# Trap cleanup on exit
trap cleanup EXIT

# Check if docker compose is available
if ! docker compose version &> /dev/null; then
    echo -e "${RED}Error: 'docker compose' is not available${NC}"
    echo -e "${YELLOW}Please ensure Docker Compose V2 is installed${NC}"
    exit 1
fi

# Step 1: Build images
echo -e "${GREEN}Step 1: Building Docker images...${NC}"
docker compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" build --no-cache

# Step 2: Start services
echo -e "${GREEN}Step 2: Starting services...${NC}"
docker compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" up -d

# Step 3: Wait for services to be healthy
echo -e "${GREEN}Step 3: Waiting for services to be healthy...${NC}"
MAX_WAIT=120
ELAPSED=0
INTERVAL=5

while [ $ELAPSED -lt $MAX_WAIT ]; do
    # Check if backend is healthy
    BACKEND_HEALTH=$(docker compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" ps backend-e2e --format json 2>/dev/null | grep -o '"Health":"[^"]*"' | cut -d'"' -f4 || echo "starting")

    if [ "$BACKEND_HEALTH" = "healthy" ]; then
        echo -e "${GREEN}Backend is healthy!${NC}"
        break
    fi

    echo "Waiting for backend (health: $BACKEND_HEALTH)... ${ELAPSED}s/${MAX_WAIT}s"
    sleep $INTERVAL
    ELAPSED=$((ELAPSED + INTERVAL))
done

if [ $ELAPSED -ge $MAX_WAIT ]; then
    echo -e "${RED}Error: Services did not become healthy in time${NC}"
    echo -e "${YELLOW}Showing logs:${NC}"
    docker compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" logs backend-e2e
    exit 1
fi

# Wait a bit more for full initialization
sleep 5

# Step 4: Run tests inside the CLI container
echo -e "${GREEN}Step 4: Running E2E tests...${NC}"
echo ""

# Run the tests with proper Go test flags
TEST_EXIT_CODE=0
docker compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" exec -T cli-e2e \
    sh -c "cd /root && go test -v -timeout $TEST_TIMEOUT ./tests/e2e/..." || TEST_EXIT_CODE=$?

echo ""

# Step 5: Show results
if [ $TEST_EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}=== E2E Tests PASSED ===${NC}"
else
    echo -e "${RED}=== E2E Tests FAILED ===${NC}"
    echo -e "${YELLOW}Showing backend logs:${NC}"
    docker compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" logs backend-e2e | tail -50
fi

# Step 6: Optionally show logs
if [ "${SHOW_LOGS}" = "true" ]; then
    echo -e "${YELLOW}=== Service Logs ===${NC}"
    docker compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" logs
fi

exit $TEST_EXIT_CODE
