#!/bin/bash

# Exit on error
set -e

# Build the CLI
make dev

# Run the CLI with any passed arguments
./bin/tgfs "$@" 