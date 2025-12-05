#!/bin/bash

# Build the application
echo "Building..."
go build -o multi-cmd ./cmd/multi-cmd

if [ $? -eq 0 ]; then
    echo "✅ Build successful!"
    echo ""
    echo "Usage examples:"
    echo "  ./multi-cmd                            # Scan current directory"
    echo "  ./multi-cmd /path/to/repos             # Scan specific directory"
    echo "  ./multi-cmd . commands.yaml            # Use custom config"
    echo "  ./multi-cmd . commands.yaml out.md     # Custom config and output"
    echo ""
    echo "Try it now:"
    echo "  ./multi-cmd .."
else
    echo "❌ Build failed"
    exit 1
fi
