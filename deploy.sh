#!/bin/bash

# Simple deployment script for the camera-detection project
# This script is intended to be run on the home server where the application is deployed.

set -e # Exit immediately if a command exits with a non-zero status.

echo "🚀 Starting deployment..."

# 1. Go to the project directory
# The script assumes it's being run from the root of the project.
# If not, you would add: cd /path/to/your/project

echo "🔄 Pulling latest changes from the main branch..."
# Ensure the current branch is main or master
# git checkout main
git pull origin main

echo "🏗️ Building and restarting Docker services..."
# Stop any running containers to avoid conflicts
docker-compose down

# Build the images and start the services in detached mode
# --build flag forces a rebuild of the images if the Dockerfile or context has changed
docker-compose up -d --build

echo "✅ Deployment finished successfully!"
echo "📈 To view logs, run: docker-compose logs -f"
