#!/bin/bash
# GitScrum CLI Installer
# https://raw.githubusercontent.com/gitscrum-core/cli/main/install.sh
#
# Usage:
#   curl -sL https://raw.githubusercontent.com/gitscrum-core/cli/main/install.sh | sh
#   curl -sL https://raw.githubusercontent.com/gitscrum-core/cli/main/install.sh | sh -s -- --version 1.0.0
#
# Environment Variables:
#   GITSCRUM_INSTALL_DIR  - Installation directory (default: ~/.local/bin)
#   GITSCRUM_VERSION      - Specific version to install (default: latest)

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="gitscrum-core/cli"
BINARY_NAME="gitscrum"
INSTALL_DIR="${GITSCRUM_INSTALL_DIR:-$HOME/.local/bin}"
VERSION="${GITSCRUM_VERSION:-latest}"
GITHUB_API="https://api.github.com"
GITHUB_RELEASES="https://github.com/$REPO/releases"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --version|-v)   VERSION="$2"; shift 2 ;;
        --dir|-d)       INSTALL_DIR="$2"; shift 2 ;;
        --help|-h)      
            echo "GitScrum CLI Installer"
            echo ""
            echo "Usage: curl -sL https://raw.githubusercontent.com/gitscrum-core/cli/main/install.sh | sh"
            echo ""
            echo "Options:"
            echo "  --version, -v    Install specific version (e.g., 1.0.0)"
            echo "  --dir, -d        Installation directory"
            echo "  --help, -h       Show this help"
            echo ""
            echo "Environment Variables:"
            echo "  GITSCRUM_INSTALL_DIR  Installation directory"
            echo "  GITSCRUM_VERSION      Version to install"
            exit 0
            ;;
        *)              shift ;;
    esac
done

# Detect OS and Architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case "$OS" in
        linux*)   OS="linux" ;;
        darwin*)  OS="darwin" ;;
        mingw*|msys*|cygwin*) OS="windows" ;;
        *)        
            echo -e "${RED}Error: Unsupported operating system: $OS${NC}"
            exit 1
            ;;
    esac
    
    case "$ARCH" in
        x86_64|amd64)  ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        armv7*)        ARCH="arm" ;;
        i686|i386)     ARCH="386" ;;
        *)
            echo -e "${RED}Error: Unsupported architecture: $ARCH${NC}"
            exit 1
            ;;
    esac
    
    PLATFORM="${OS}_${ARCH}"
    
    # File extension
    EXT="tar.gz"
    if [ "$OS" = "windows" ]; then
        EXT="zip"
        BINARY_NAME="gitscrum.exe"
    fi
}

# Get latest version from GitHub
get_latest_version() {
    if [ "$VERSION" = "latest" ]; then
        VERSION=$(curl -s "$GITHUB_API/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"v?([^"]+)".*/\1/')
        if [ -z "$VERSION" ]; then
            echo -e "${RED}Error: Could not determine latest version${NC}"
            exit 1
        fi
    fi
    # Remove 'v' prefix if present
    VERSION="${VERSION#v}"
}

# Download and install
install() {
    echo -e "${BLUE}GitScrum CLI Installer${NC}"
    echo ""
    
    detect_platform
    get_latest_version
    
    echo -e "Version:  ${GREEN}v$VERSION${NC}"
    echo -e "Platform: ${GREEN}$PLATFORM${NC}"
    echo -e "Install:  ${GREEN}$INSTALL_DIR${NC}"
    echo ""
    
    # Create install directory
    mkdir -p "$INSTALL_DIR"
    
    # Download URL
    DOWNLOAD_URL="$GITHUB_RELEASES/download/v$VERSION/${BINARY_NAME}_${VERSION}_${PLATFORM}.${EXT}"
    
    echo -e "${YELLOW}Downloading...${NC}"
    
    # Create temp directory
    TMP_DIR=$(mktemp -d)
    trap "rm -rf $TMP_DIR" EXIT
    
    # Download
    ARCHIVE="$TMP_DIR/gitscrum.${EXT}"
    if command -v curl &> /dev/null; then
        curl -sL "$DOWNLOAD_URL" -o "$ARCHIVE"
    elif command -v wget &> /dev/null; then
        wget -q "$DOWNLOAD_URL" -O "$ARCHIVE"
    else
        echo -e "${RED}Error: curl or wget required${NC}"
        exit 1
    fi
    
    # Verify download
    if [ ! -f "$ARCHIVE" ] || [ ! -s "$ARCHIVE" ]; then
        echo -e "${RED}Error: Download failed${NC}"
        echo "URL: $DOWNLOAD_URL"
        exit 1
    fi
    
    # Extract
    echo -e "${YELLOW}Extracting...${NC}"
    cd "$TMP_DIR"
    
    if [ "$EXT" = "tar.gz" ]; then
        tar -xzf "$ARCHIVE"
    else
        unzip -q "$ARCHIVE"
    fi
    
    # Find binary
    if [ -f "$BINARY_NAME" ]; then
        BINARY_PATH="$BINARY_NAME"
    elif [ -f "${BINARY_NAME%.exe}" ]; then
        BINARY_PATH="${BINARY_NAME%.exe}"
    else
        echo -e "${RED}Error: Binary not found in archive${NC}"
        exit 1
    fi
    
    # Install
    echo -e "${YELLOW}Installing...${NC}"
    chmod +x "$BINARY_PATH"
    mv "$BINARY_PATH" "$INSTALL_DIR/$BINARY_NAME"
    
    # Verify installation
    if [ ! -f "$INSTALL_DIR/$BINARY_NAME" ]; then
        echo -e "${RED}Error: Installation failed${NC}"
        exit 1
    fi
    
    echo ""
    echo -e "${GREEN}✓ GitScrum CLI v$VERSION installed successfully!${NC}"
    echo ""
    
    # Check if in PATH
    if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
        echo -e "${YELLOW}Note: Add $INSTALL_DIR to your PATH:${NC}"
        echo ""
        
        SHELL_NAME=$(basename "$SHELL")
        case "$SHELL_NAME" in
            bash)
                echo "  echo 'export PATH=\"\$PATH:$INSTALL_DIR\"' >> ~/.bashrc"
                echo "  source ~/.bashrc"
                ;;
            zsh)
                echo "  echo 'export PATH=\"\$PATH:$INSTALL_DIR\"' >> ~/.zshrc"
                echo "  source ~/.zshrc"
                ;;
            fish)
                echo "  set -U fish_user_paths $INSTALL_DIR \$fish_user_paths"
                ;;
            *)
                echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
                ;;
        esac
        echo ""
    fi
    
    # Next steps
    echo -e "${BLUE}Next steps:${NC}"
    echo ""
    echo "  1. Login to GitScrum:"
    echo "     gitscrum auth login"
    echo ""
    echo "  2. Set your default workspace:"
    echo "     gitscrum config set workspace <your-workspace>"
    echo ""
    echo "  3. Initialize project (optional):"
    echo "     gitscrum init"
    echo ""
    echo -e "${BLUE}Documentation: https://gitscrum.com/docs/cli${NC}"
}

# Run installer
install
