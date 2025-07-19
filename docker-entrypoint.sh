#!/bin/sh
set -e

# Ensure data directory exists and has correct permissions
mkdir -p /app/data
chown -R appuser:appuser /app/data
chmod -R 755 /app/data

# Switch to appuser and execute the main command
exec su-exec appuser "$@" 