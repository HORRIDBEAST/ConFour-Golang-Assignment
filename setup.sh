#!/bin/bash

# ============================================
# 🎮 4 in a Row - Setup Script
# ============================================
# This script initializes the Go project and
# downloads all required dependencies.
# ============================================

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Header
echo ""
echo -e "${PURPLE}╔═══════════════════════════════════════════╗${NC}"
echo -e "${PURPLE}║                                           ║${NC}"
echo -e "${PURPLE}║     ${CYAN}🎮  4 in a Row - Setup Script${PURPLE}       ║${NC}"
echo -e "${PURPLE}║                                           ║${NC}"
echo -e "${PURPLE}╚═══════════════════════════════════════════╝${NC}"
echo ""

# ============================================
# Step 1: Check Go Installation
# ============================================
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}📋 Step 1: Checking Prerequisites${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Error: Go is not installed.${NC}"
    echo -e "${YELLOW}   Please install Go 1.25 or higher from:${NC}"
    echo -e "${YELLOW}   https://golang.org/dl/${NC}"
    exit 1
fi

GO_VERSION=$(go version)
echo -e "${GREEN}✅ Go is installed: ${GO_VERSION}${NC}"
echo ""

# ============================================
# Step 2: Initialize Go Module
# ============================================
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}📦 Step 2: Initializing Go Module${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

if [ ! -f "go.mod" ]; then
    echo -e "${YELLOW}⚙️  Creating go.mod file...${NC}"
    # IMPORTANT: Change 'hello-go' if your module name is different
    go mod init hello-go
    echo -e "${GREEN}✅ Go module initialized successfully${NC}"
else
    echo -e "${GREEN}✅ go.mod already exists${NC}"
fi
echo ""

# ============================================
# Step 3: Download Dependencies
# ============================================
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}📥 Step 3: Downloading Dependencies${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

echo -e "${YELLOW}⚙️  Running go mod tidy...${NC}"
go mod tidy

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Dependencies downloaded successfully${NC}"
else
    echo -e "${RED}❌ Failed to download dependencies${NC}"
    exit 1
fi
echo ""

# ============================================
# Setup Complete
# ============================================
echo -e "${GREEN}╔═══════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║                                           ║${NC}"
echo -e "${GREEN}║        ✅  Setup Complete! 🎉            ║${NC}"
echo -e "${GREEN}║                                           ║${NC}"
echo -e "${GREEN}╚═══════════════════════════════════════════╝${NC}"
echo ""

# ============================================
# Next Steps
# ============================================
echo -e "${PURPLE}┌───────────────────────────────────────────┐${NC}"
echo -e "${PURPLE}│  ${CYAN}🚀 Next Steps${PURPLE}                          │${NC}"
echo -e "${PURPLE}└───────────────────────────────────────────┘${NC}"
echo ""
echo -e "${YELLOW}┌─────────────────────────────────────────────────────┐${NC}"
echo -e "${YELLOW}│  Option 1: Using Docker (Recommended) ⭐            │${NC}"
echo -e "${YELLOW}└─────────────────────────────────────────────────────┘${NC}"
echo -e "${CYAN}   docker-compose up --build${NC}"
echo ""
echo -e "   ${GREEN}Then open:${NC} ${BLUE}http://localhost:8080${NC}"
echo ""
echo -e "${YELLOW}┌─────────────────────────────────────────────────────┐${NC}"
echo -e "${YELLOW}│  Option 2: Local Development                        │${NC}"
echo -e "${YELLOW}└─────────────────────────────────────────────────────┘${NC}"
echo -e "${CYAN}   # 1. Start Docker services:${NC}"
echo -e "      docker-compose up -d db zookeeper kafka"
echo ""
echo -e "${CYAN}   # 2. Set environment variables:${NC}"
echo -e "      export DATABASE_URL=\"postgres://postgres:postgres@localhost:5432/connect4?sslmode=disable\""
echo -e "      export KAFKA_BROKERS=\"localhost:9092\""
echo ""
echo -e "${CYAN}   # 3. Run the application:${NC}"
echo -e "      go run ."
echo ""
echo -e "${YELLOW}┌─────────────────────────────────────────────────────┐${NC}"
echo -e "${YELLOW}│  📚 Documentation                                    │${NC}"
echo -e "${YELLOW}└─────────────────────────────────────────────────────┘${NC}"
echo -e "   Check ${CYAN}README.md${NC} for detailed instructions!"
echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}Happy Gaming! 🎮${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""