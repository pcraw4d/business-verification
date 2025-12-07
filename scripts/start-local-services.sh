#!/bin/bash

# Start Local Services for Phase 1 Testing
# This script starts the Playwright and Classification services using Docker Compose

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=== Starting Local Services for Phase 1 Testing ===${NC}\n"

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker is not installed${NC}"
    echo "Please install Docker: https://docs.docker.com/get-docker/"
    exit 1
fi

# Check if docker-compose is available
if ! docker compose version &> /dev/null && ! docker-compose version &> /dev/null; then
    echo -e "${RED}❌ docker-compose is not installed${NC}"
    exit 1
fi

# Use docker compose (newer) or docker-compose (older)
if docker compose version &> /dev/null; then
    DOCKER_COMPOSE="docker compose"
else
    DOCKER_COMPOSE="docker-compose"
fi

# Check if .env exists
if [ ! -f ".env" ]; then
    echo -e "${YELLOW}⚠️  .env file not found${NC}"
    echo "Creating .env from template..."
    if [ -f ".env.example" ]; then
        cp .env.example .env
        echo -e "${YELLOW}Please update .env with your Supabase credentials${NC}"
    else
        echo -e "${RED}❌ .env.example not found. Please create .env manually.${NC}"
        exit 1
    fi
fi

echo -e "${GREEN}✅ Docker is available${NC}"
echo ""

# Start services (only Playwright and Classification for Phase 1 testing)
echo -e "${BLUE}Starting Playwright and Classification services...${NC}"
$DOCKER_COMPOSE -f docker-compose.local.yml up -d redis-cache playwright-scraper classification-service

# Wait for services to be ready
echo ""
echo -e "${BLUE}Waiting for services to start...${NC}"

# Wait for Playwright service
echo -n "Waiting for Playwright service..."
for i in {1..60}; do
    if curl -s -f "http://localhost:3000/health" > /dev/null 2>&1; then
        echo -e " ${GREEN}✅${NC}"
        break
    fi
    if [ $i -eq 60 ]; then
        echo -e " ${RED}❌ Timeout${NC}"
        echo "Check logs: $DOCKER_COMPOSE -f docker-compose.local.yml logs playwright-scraper"
        exit 1
    fi
    sleep 1
    echo -n "."
done

# Wait for Classification service
echo -n "Waiting for Classification service..."
for i in {1..120}; do
    if curl -s -f "http://localhost:8081/health" > /dev/null 2>&1; then
        echo -e " ${GREEN}✅${NC}"
        break
    fi
    if [ $i -eq 120 ]; then
        echo -e " ${RED}❌ Timeout${NC}"
        echo "Check logs: $DOCKER_COMPOSE -f docker-compose.local.yml logs classification-service"
        exit 1
    fi
    sleep 1
    if [ $((i % 10)) -eq 0 ]; then
        echo -n "."
    fi
done

echo ""
echo -e "${GREEN}✅ All services are ready!${NC}"
echo ""
echo -e "${BLUE}Service URLs:${NC}"
echo -e "  Classification: http://localhost:8081"
echo -e "  Playwright: http://localhost:3000"
echo ""
echo -e "${BLUE}Useful Commands:${NC}"
echo -e "  View logs: ${YELLOW}$DOCKER_COMPOSE -f docker-compose.local.yml logs -f${NC}"
echo -e "  Stop services: ${YELLOW}$DOCKER_COMPOSE -f docker-compose.local.yml down${NC}"
echo -e "  Restart services: ${YELLOW}$DOCKER_COMPOSE -f docker-compose.local.yml restart${NC}"
echo ""
echo -e "${BLUE}Test the services:${NC}"
echo -e "  ${YELLOW}curl http://localhost:3000/health${NC}"
echo -e "  ${YELLOW}curl http://localhost:8081/health${NC}"
echo -e "  ${YELLOW}./scripts/test-phase1-local.sh${NC}"
