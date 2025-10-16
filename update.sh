#!/bin/bash

# TuneCent Backend Update Script
# Usage: sudo ./update.sh

set -e

APP_NAME="tunecent-backend"
APP_DIR="/opt/tunecent"
SERVICE_NAME="tunecent-backend"

echo "=========================================="
echo "TuneCent Backend Update"
echo "=========================================="

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo "This script must be run as root (use sudo)"
   exit 1
fi

# Stop service
echo "Stopping service..."
systemctl stop $SERVICE_NAME

# Backup current binary
echo "Backing up current binary..."
if [ -f "$APP_DIR/bin/$APP_NAME" ]; then
    cp $APP_DIR/bin/$APP_NAME $APP_DIR/bin/${APP_NAME}.backup.$(date +%Y%m%d%H%M%S)
fi

# Pull latest code (if using git)
echo "Updating code..."
cd $APP_DIR/app
# git pull origin main  # Uncomment if using git

# Rebuild application
echo "Building application..."
export GO111MODULE=on
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o $APP_DIR/bin/$APP_NAME ./cmd/server/main_complete.go

# Set permissions
chown tunecent:tunecent $APP_DIR/bin/$APP_NAME
chmod +x $APP_DIR/bin/$APP_NAME

# Start service
echo "Starting service..."
systemctl start $SERVICE_NAME

# Check status
echo ""
echo "Checking service status..."
sleep 2
systemctl status $SERVICE_NAME --no-pager

echo ""
echo "✅ Update complete!"
echo ""
echo "View logs: journalctl -u $SERVICE_NAME -f"
echo "Health check: curl http://localhost:8080/health"
