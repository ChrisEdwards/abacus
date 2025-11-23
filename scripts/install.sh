#!/usr/bin/env bash
set -e

# Abacus Installation Script for Unix/macOS/Linux
# Usage: curl -fsSL https://raw.githubusercontent.com/ChrisEdwards/abacus/main/scripts/install.sh | bash

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="ChrisEdwards/abacus"
BINARY_NAME="abacus"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

# Platform detection
detect_platform() {
    OS="$(uname -s)"
    ARCH="$(uname -m)"

    case "$OS" in
        Darwin)
            OS="darwin"
            ;;
        Linux)
            OS="linux"
            ;;
        *)
            echo -e "${RED}✗${NC} Unsupported operating system: $OS"
            exit 1
            ;;
    esac

    case "$ARCH" in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            echo -e "${RED}✗${NC} Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac

    echo -e "${BLUE}ℹ${NC} Detected platform: $OS/$ARCH"
}

# Get latest version from GitHub API
get_latest_version() {
    echo -e "${BLUE}ℹ${NC} Fetching latest version..."

    if command -v curl >/dev/null 2>&1; then
        VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
        VERSION=$(wget -qO- "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        echo -e "${RED}✗${NC} Neither curl nor wget is available. Please install one of them."
        exit 1
    fi

    if [ -z "$VERSION" ]; then
        echo -e "${RED}✗${NC} Failed to fetch latest version"
        exit 1
    fi

    echo -e "${GREEN}✓${NC} Latest version: $VERSION"
}

# Download and install binary
install_binary() {
    VERSION_NUMBER="${VERSION#v}"
    ARCHIVE_NAME="${BINARY_NAME}_${VERSION_NUMBER}_${OS}_${ARCH}.tar.gz"
    DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/$ARCHIVE_NAME"

    echo -e "${BLUE}ℹ${NC} Downloading $ARCHIVE_NAME..."

    TMP_DIR=$(mktemp -d)
    cd "$TMP_DIR"

    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$DOWNLOAD_URL" -o "$ARCHIVE_NAME"
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$DOWNLOAD_URL" -O "$ARCHIVE_NAME"
    fi

    if [ ! -f "$ARCHIVE_NAME" ]; then
        echo -e "${RED}✗${NC} Failed to download $ARCHIVE_NAME"
        rm -rf "$TMP_DIR"
        exit 1
    fi

    echo -e "${GREEN}✓${NC} Downloaded $ARCHIVE_NAME"

    echo -e "${BLUE}ℹ${NC} Extracting binary..."
    tar -xzf "$ARCHIVE_NAME"

    if [ ! -f "$BINARY_NAME" ]; then
        echo -e "${RED}✗${NC} Binary not found in archive"
        rm -rf "$TMP_DIR"
        exit 1
    fi

    # Create install directory if it doesn't exist
    mkdir -p "$INSTALL_DIR"

    # Check if binary already exists
    if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
        echo -e "${YELLOW}⚠${NC} Existing $BINARY_NAME found, backing up..."
        mv "$INSTALL_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME.backup"
    fi

    # Install binary
    echo -e "${BLUE}ℹ${NC} Installing to $INSTALL_DIR..."
    mv "$BINARY_NAME" "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"

    rm -rf "$TMP_DIR"
    echo -e "${GREEN}✓${NC} Installed $BINARY_NAME to $INSTALL_DIR"
}

# Try go install as fallback
try_go_install() {
    echo -e "${YELLOW}⚠${NC} Trying go install as fallback..."

    if ! command -v go >/dev/null 2>&1; then
        echo -e "${RED}✗${NC} Go is not installed"
        return 1
    fi

    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    echo -e "${BLUE}ℹ${NC} Found Go $GO_VERSION"

    echo -e "${BLUE}ℹ${NC} Running: go install github.com/$REPO/cmd/$BINARY_NAME@latest"
    go install "github.com/$REPO/cmd/$BINARY_NAME@latest"

    GOBIN="${GOBIN:-$(go env GOPATH)/bin}"
    if [ -f "$GOBIN/$BINARY_NAME" ]; then
        echo -e "${GREEN}✓${NC} Installed via go install to $GOBIN"
        INSTALL_DIR="$GOBIN"
        return 0
    fi

    return 1
}

# Verify installation
verify_installation() {
    echo -e "${BLUE}ℹ${NC} Verifying installation..."

    if ! command -v "$BINARY_NAME" >/dev/null 2>&1; then
        echo -e "${YELLOW}⚠${NC} $BINARY_NAME is not in PATH"
        echo ""
        echo "Add $INSTALL_DIR to your PATH by adding this line to your shell profile:"
        echo ""
        echo -e "  ${BLUE}export PATH=\"\$PATH:$INSTALL_DIR\"${NC}"
        echo ""
        echo "For bash: ~/.bashrc or ~/.bash_profile"
        echo "For zsh: ~/.zshrc"
        echo ""
        echo "Then run: source ~/.bashrc (or ~/.zshrc)"
        echo ""
        echo "Or run directly: $INSTALL_DIR/$BINARY_NAME"
        return 1
    fi

    VERSION_OUTPUT=$("$BINARY_NAME" --version 2>&1 || true)
    echo -e "${GREEN}✓${NC} $BINARY_NAME is installed and in PATH"
    echo ""
    echo "$VERSION_OUTPUT"
    return 0
}

# Check for existing installations
check_existing() {
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        EXISTING_PATH=$(command -v "$BINARY_NAME")
        EXISTING_VERSION=$("$BINARY_NAME" --version 2>&1 | head -1 || echo "unknown")
        echo -e "${YELLOW}⚠${NC} Found existing installation:"
        echo "  Location: $EXISTING_PATH"
        echo "  Version:  $EXISTING_VERSION"
        echo ""
        read -p "Continue with installation? [y/N] " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "Installation cancelled"
            exit 0
        fi
    fi
}

# Main installation flow
main() {
    echo ""
    echo "═══════════════════════════════════════"
    echo "  Abacus Installation Script"
    echo "═══════════════════════════════════════"
    echo ""

    check_existing
    detect_platform
    get_latest_version

    echo ""
    if install_binary; then
        echo ""
        verify_installation
        echo ""
        echo -e "${GREEN}✓${NC} Installation complete!"
    else
        echo ""
        echo -e "${YELLOW}⚠${NC} Direct installation failed"
        if try_go_install; then
            echo ""
            verify_installation
            echo ""
            echo -e "${GREEN}✓${NC} Installation complete via go install!"
        else
            echo ""
            echo -e "${RED}✗${NC} Installation failed"
            echo ""
            echo "Please try one of these alternatives:"
            echo "  1. Install via Homebrew: brew install ChrisEdwards/tap/abacus"
            echo "  2. Install via Go: go install github.com/$REPO/cmd/$BINARY_NAME@latest"
            echo "  3. Download manually from: https://github.com/$REPO/releases"
            exit 1
        fi
    fi

    echo ""
    echo "For more information:"
    echo "  Documentation: https://github.com/$REPO"
    echo "  Report issues: https://github.com/$REPO/issues"
    echo ""
}

main
