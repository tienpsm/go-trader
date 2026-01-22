#!/bin/bash
# Download sample ITCH data file from CppTrader repository

set -e

TESTDATA_DIR="testdata"
SAMPLE_FILE="$TESTDATA_DIR/sample.itch"
DOWNLOAD_URL="https://github.com/chronoxor/CppTrader/raw/master/tools/itch/sample.itch"

echo "Downloading sample ITCH data file..."
echo "URL: $DOWNLOAD_URL"
echo "Destination: $SAMPLE_FILE"
echo ""

# Create testdata directory if it doesn't exist
mkdir -p "$TESTDATA_DIR"

# Download the file
if command -v curl &> /dev/null; then
    curl -L -o "$SAMPLE_FILE" "$DOWNLOAD_URL"
elif command -v wget &> /dev/null; then
    wget -O "$SAMPLE_FILE" "$DOWNLOAD_URL"
else
    echo "Error: Neither curl nor wget is available. Please install one of them."
    exit 1
fi

# Check if download was successful
if [ -f "$SAMPLE_FILE" ]; then
    SIZE=$(du -h "$SAMPLE_FILE" | cut -f1)
    echo ""
    echo "✅ Download complete!"
    echo "File size: $SIZE"
    echo ""
    echo "You can now:"
    echo "  1. Run tests: go test ./itch -v -run TestParser_SampleFile"
    echo "  2. Run benchmarks: go test ./itch -bench=BenchmarkParser_ParseFile"
    echo "  3. Use analyzer: go run ./cmd/itch-analyzer/main.go $SAMPLE_FILE"
else
    echo "❌ Download failed!"
    exit 1
fi
