#!/bin/bash

# Exit on error
set -e

# Build the CLI
make dev

# Set permissions for the binary
chmod +x ./bin/tgfs

# Run the CLI with any passed arguments
./bin/tgfs "$@" 