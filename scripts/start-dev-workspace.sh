#!/bin/bash

# Vim Actions Development Workspace Startup Script
# Simple script to start the vim actions development environment

set -e

# Configuration
SESSION_NAME="vim-actions-dev"

echo "🚀 Starting Vim Actions development environment..."

# Start development environment
echo "Starting services with Tilt..."
tilt up

echo "✅ Development environment started!"
echo ""
echo "Services available at:"
echo "  📊 Tilt Dashboard: http://localhost:10350"
echo "  🔧 Server API: http://localhost:8288"
echo "  🎨 Client App: http://localhost:3020"
echo "  💾 Valkey DB: localhost:6399"
echo ""
echo "Useful commands:"
echo "  tilt down           - Stop all services"
echo "  tilt up --stream    - Start with streaming logs"