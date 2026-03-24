#!/bin/bash
# Terminal Intelligence Build Script for Linux/macOS

set -e

BINARY_NAME="scmd"
VERSION="2.0.1"
BUILD_DIR="build"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get build number from git
if git rev-list --count HEAD &>/dev/null; then
    BUILD_NUMBER=$(($(git rev-list --count HEAD) + 1))
else
    BUILD_NUMBER=1
fi

LDFLAGS="-s -w -X main.version=$VERSION -X main.buildNumber=$BUILD_NUMBER"

ensure_build_dir() {
    if [ ! -d "$BUILD_DIR" ]; then
        mkdir -p "$BUILD_DIR"
    fi
}

build_current() {
    echo -e "${CYAN}Building $BINARY_NAME for current platform...${NC}"
    ensure_build_dir
    go build -ldflags="$LDFLAGS" -o "$BUILD_DIR/$BINARY_NAME" .
    echo -e "${GREEN}Build complete: $BUILD_DIR/$BINARY_NAME${NC}"
}

build_windows() {
    echo -e "${CYAN}Building $BINARY_NAME for Windows...${NC}"
    ensure_build_dir
    GOOS=windows GOARCH=amd64 go build -ldflags="$LDFLAGS" -o "$BUILD_DIR/$BINARY_NAME-windows-amd64.exe" .
    echo -e "${GREEN}Build complete: $BUILD_DIR/$BINARY_NAME-windows-amd64.exe${NC}"
}

build_linux() {
    echo -e "${CYAN}Building $BINARY_NAME for Linux...${NC}"
    ensure_build_dir
    GOOS=linux GOARCH=amd64 go build -ldflags="$LDFLAGS" -o "$BUILD_DIR/$BINARY_NAME-linux-amd64" .
    GOOS=linux GOARCH=arm64 go build -ldflags="$LDFLAGS" -o "$BUILD_DIR/$BINARY_NAME-linux-aarch64" .
    echo -e "${GREEN}Build complete: $BUILD_DIR/$BINARY_NAME-linux-amd64 and $BUILD_DIR/$BINARY_NAME-linux-aarch64${NC}"
}

build_darwin() {
    echo -e "${CYAN}Building $BINARY_NAME for macOS...${NC}"
    ensure_build_dir
    GOOS=darwin GOARCH=amd64 go build -ldflags="$LDFLAGS" -o "$BUILD_DIR/$BINARY_NAME-darwin-amd64" .
    GOOS=darwin GOARCH=arm64 go build -ldflags="$LDFLAGS" -o "$BUILD_DIR/$BINARY_NAME-darwin-arm64" .
    echo -e "${GREEN}Build complete: $BUILD_DIR/$BINARY_NAME-darwin-amd64 and $BUILD_DIR/$BINARY_NAME-darwin-arm64${NC}"
}

build_all() {
    build_windows
    build_linux
    build_darwin
    echo -e "${GREEN}All platform builds complete!${NC}"
}

run_tests() {
    echo -e "${CYAN}Running tests...${NC}"
    go test ./... -v
}

run_clean() {
    echo -e "${CYAN}Cleaning build artifacts...${NC}"
    rm -rf "$BUILD_DIR"
    rm -f coverage.out coverage.html
    echo -e "${GREEN}Clean complete${NC}"
}

show_help() {
    echo -e "${CYAN}Terminal Intelligence (TI) Build Script${NC}"
    echo ""
    echo -e "${YELLOW}Usage: ./build.sh [target]${NC}"
    echo ""
    echo -e "${YELLOW}Targets:${NC}"
    echo "  build            - Build for current platform (default)"
    echo "  windows          - Build for Windows (amd64)"
    echo "  linux            - Build for Linux (amd64 and arm64)"
    echo "  darwin           - Build for macOS (amd64 and arm64)"
    echo "  all              - Build for all platforms"
    echo "  test             - Run all tests"
    echo "  clean            - Remove build artifacts"
    echo "  help             - Show this help message"
}

# Execute target
TARGET="${1:-build}"

case "$TARGET" in
    build)
        build_current
        ;;
    windows)
        build_windows
        ;;
    linux)
        build_linux
        ;;
    darwin)
        build_darwin
        ;;
    all)
        build_all
        ;;
    test)
        run_tests
        ;;
    clean)
        run_clean
        ;;
    help)
        show_help
        ;;
    *)
        echo -e "${RED}Unknown target: $TARGET${NC}"
        show_help
        exit 1
        ;;
esac
