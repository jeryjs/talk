#!/bin/bash
# Build script for Nero
# Usage: ./build.sh [version]

VERSION="${1:-dev}"
OUTPUT="nero"

echo "🔨 Building Nero v$VERSION..."

# Get build information
COMMIT_SHA=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(date +"%Y-%m-%d_%H:%M:%S")

# Build flags
LDFLAGS="-X main.Version=$VERSION -X main.CommitSHA=$COMMIT_SHA -X main.BuildDate=$BUILD_DATE"

# Build
echo "📦 Compiling..."
go build -ldflags "$LDFLAGS" -o "$OUTPUT" ./cmd/nero/

if [ $? -eq 0 ]; then
    echo "✅ Build successful: $OUTPUT"
    echo ""
    echo "Run with: ./$OUTPUT"
else
    echo "❌ Build failed"
    exit 1
fi
