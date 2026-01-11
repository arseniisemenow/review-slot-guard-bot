#!/bin/bash
set -e

# Script is in project root, so SCRIPT_DIR is the project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$SCRIPT_DIR"

echo "Building functions for Yandex Cloud..."

# Clean previous builds
rm -rf "${SCRIPT_DIR}/build"
mkdir -p "${SCRIPT_DIR}/build"

# Helper function to build a function
build_function() {
    local FUNC_NAME=$1
    local FUNC_DIR="${PROJECT_ROOT}/functions/${FUNC_NAME}"

    echo "Building ${FUNC_NAME} function..."

    # Create build directory
    local BUILD_DIR="${SCRIPT_DIR}/build/${FUNC_NAME}"
    rm -rf "$BUILD_DIR"
    mkdir -p "$BUILD_DIR"

    # Copy function source files
    find "${FUNC_DIR}" -name "*.go" -type f | while read -r file; do
        rel_path="${file#${FUNC_DIR}/}"
        mkdir -p "$BUILD_DIR/$(dirname "$rel_path")"
        cp "$file" "$BUILD_DIR/$rel_path"
    done

    # Copy go.mod and go.sum
    if [ -f "${FUNC_DIR}/go.mod" ]; then
        cp "${FUNC_DIR}/go.mod" "$BUILD_DIR/go.mod"
    fi
    if [ -f "${FUNC_DIR}/go.sum" ]; then
        cp "${FUNC_DIR}/go.sum" "$BUILD_DIR/"
    fi

    cd "$BUILD_DIR"

    # Run go mod download to populate go.sum if needed
    if command -v go &> /dev/null; then
        echo "  Downloading dependencies for ${FUNC_NAME}..."
        go mod download 2>/dev/null || true
    fi

    # Create the zip file
    zip -qr "${SCRIPT_DIR}/terraform/${FUNC_NAME}.zip" \
        ./*.go \
        ./*.mod \
        ./*.sum \
        ./internal/

    echo "  Created ${FUNC_NAME}.zip"
}

# Build both functions
build_function "periodic_job"
build_function "telegram_handler"

echo "Build complete!"
echo ""
echo "Created zip files:"
ls -lh "${SCRIPT_DIR}/terraform/"*.zip
