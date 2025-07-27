#!/usr/bin/env aish

# 🎭 Epic Story Generator - Interactive Story Creator
# Advanced AI-powered story generation with multiple genres, characters, and styles

# Color definitions
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

print_banner() {
  # Display epic banner
  echo -e "${CYAN}"
  echo "╔══════════════════════════════════════════════════════════════╗"
  echo "║                 🎭 Epic Story Generator 🎭                   ║"
  echo "║                  AI-Powered AISH Script                      ║"
  echo "╚══════════════════════════════════════════════════════════════╝"
  echo -e "${NC}"
}

generate_story() {
  # add ai: prefix to ensure this command is executed under AI mode
  ai: write me a story about "$@" within 200 words
}

print_banner
generate_story "$@" 