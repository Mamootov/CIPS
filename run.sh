#!/usr/bin/env bash
set -e

echo "==================================================="
echo "  Launching Clean IP Scanner (CIPS)..."
echo "  Embedded Dashboard: http://127.0.0.1:19090"
echo "==================================================="

if [ -f "./cips" ]; then
    chmod +x ./cips
    ./cips
else
    go run ./cmd/cips
fi
