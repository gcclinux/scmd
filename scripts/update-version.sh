#!/bin/bash
# Update version across all project files
# Usage: ./scripts/update-version.sh <new_version>
# Example: ./scripts/update-version.sh 2.0.6

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
NC='\033[0m'

NEW_VERSION="$1"

if [ -z "$NEW_VERSION" ]; then
    echo -e "${RED}Error: No version specified${NC}"
    echo "Usage: ./scripts/update-version.sh <new_version>"
    echo "Example: ./scripts/update-version.sh 2.0.6"
    exit 1
fi

# Validate version format (e.g. 1.2.3)
if ! echo "$NEW_VERSION" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+$'; then
    echo -e "${RED}Error: Invalid version format '$NEW_VERSION'. Expected format: X.Y.Z${NC}"
    exit 1
fi

# Get current version from the release file
CURRENT_VERSION=$(tr -d '[:space:]' < release)
echo -e "${CYAN}Updating version: ${CURRENT_VERSION} -> ${NEW_VERSION}${NC}"

# 1. release file
echo "$NEW_VERSION" > release
echo -e "${GREEN}Updated: release${NC}"

# 2. internal/updater/version.go
sed -i "s/const Release = \"$CURRENT_VERSION\"/const Release = \"$NEW_VERSION\"/" internal/updater/version.go
echo -e "${GREEN}Updated: internal/updater/version.go${NC}"

# 3. docker/Dockerfile
sed -i "s/app.version=\"$CURRENT_VERSION\"/app.version=\"$NEW_VERSION\"/" docker/Dockerfile
echo -e "${GREEN}Updated: docker/Dockerfile${NC}"

# 4. scripts/build.sh
sed -i "s/^VERSION=\"$CURRENT_VERSION\"/VERSION=\"$NEW_VERSION\"/" scripts/build.sh
echo -e "${GREEN}Updated: scripts/build.sh${NC}"

# 5. scripts/build.ps1
sed -i "s/^\$VERSION = \"$CURRENT_VERSION\"/\$VERSION = \"$NEW_VERSION\"/" scripts/build.ps1
echo -e "${GREEN}Updated: scripts/build.ps1${NC}"

echo ""
echo -e "${GREEN}Version updated to ${NEW_VERSION} in all files.${NC}"
