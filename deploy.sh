#!/bin/bash

# TuneCent Backend Deployment Script for VPS
# Usage: ./deploy.sh [production|staging]

set -e  # Exit on error

ENV=${1:-production}
APP_NAME="tunecent-backend"
APP_USER="tunecent"
APP_DIR="/opt/tunecent"
SERVICE_NAME="tunecent-backend"

echo "=========================================="
echo "TuneCent Backend Deployment"
echo "Environment: $ENV"
echo "=========================================="

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo "This script must be run as root (use sudo)"
   exit 1
fi

# Step 1: Install dependencies
echo ""
echo "Step 1: Installing system dependencies..."
apt-get update
apt-get install -y wget curl git build-essential mysql-client

# Step 2: Install Go if not present
if ! command -v go &> /dev/null; then
    echo ""
    echo "Step 2: Installing Go 1.21..."
    wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
    rm -rf /usr/local/go
    tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
    rm go1.21.6.linux-amd64.tar.gz

    # Add Go to PATH
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    export PATH=$PATH:/usr/local/go/bin

    echo "Go installed: $(go version)"
else
    echo "Go already installed: $(go version)"
fi

# Step 3: Create application user
echo ""
echo "Step 3: Creating application user..."
if ! id -u $APP_USER > /dev/null 2>&1; then
    useradd -r -s /bin/bash -d $APP_DIR $APP_USER
    echo "User $APP_USER created"
else
    echo "User $APP_USER already exists"
fi

# Step 4: Create application directory
echo ""
echo "Step 4: Setting up application directory..."
mkdir -p $APP_DIR
mkdir -p $APP_DIR/logs
mkdir -p $APP_DIR/bin

# Step 5: Copy application files
echo ""
echo "Step 5: Copying application files..."
cp -r . $APP_DIR/app
cd $APP_DIR/app

# Step 6: Build application
echo ""
echo "Step 6: Building application..."
export GO111MODULE=on
go mod download
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o $APP_DIR/bin/$APP_NAME ./cmd/server/main_complete.go

echo "Binary built: $APP_DIR/bin/$APP_NAME"

# Step 7: Setup environment file
echo ""
echo "Step 7: Setting up environment configuration..."
if [ ! -f "$APP_DIR/.env" ]; then
    cp .env.example $APP_DIR/.env
    echo "Created .env file at $APP_DIR/.env"
    echo "⚠️  IMPORTANT: Edit $APP_DIR/.env with your configuration!"
else
    echo ".env file already exists"
fi

# Step 8: Set permissions
echo ""
echo "Step 8: Setting permissions..."
chown -R $APP_USER:$APP_USER $APP_DIR
chmod +x $APP_DIR/bin/$APP_NAME

# Step 9: Create systemd service
echo ""
echo "Step 9: Creating systemd service..."
cat > /etc/systemd/system/$SERVICE_NAME.service <<EOF
[Unit]
Description=TuneCent Backend API Service
After=network.target mysql.service
Wants=mysql.service

[Service]
Type=simple
User=$APP_USER
Group=$APP_USER
WorkingDirectory=$APP_DIR
EnvironmentFile=$APP_DIR/.env
ExecStart=$APP_DIR/bin/$APP_NAME
Restart=always
RestartSec=10
StandardOutput=append:$APP_DIR/logs/app.log
StandardError=append:$APP_DIR/logs/error.log

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$APP_DIR

[Install]
WantedBy=multi-user.target
EOF

echo "Systemd service created: /etc/systemd/system/$SERVICE_NAME.service"

# Step 10: Reload systemd and enable service
echo ""
echo "Step 10: Enabling service..."
systemctl daemon-reload
systemctl enable $SERVICE_NAME

# Step 11: Setup log rotation
echo ""
echo "Step 11: Setting up log rotation..."
cat > /etc/logrotate.d/$SERVICE_NAME <<EOF
$APP_DIR/logs/*.log {
    daily
    rotate 14
    compress
    delaycompress
    notifempty
    create 0640 $APP_USER $APP_USER
    sharedscripts
    postrotate
        systemctl reload $SERVICE_NAME > /dev/null 2>&1 || true
    endscript
}
EOF

echo "Log rotation configured"

# Step 12: Setup firewall (if ufw is available)
if command -v ufw &> /dev/null; then
    echo ""
    echo "Step 12: Configuring firewall..."
    ufw allow 8080/tcp comment 'TuneCent Backend API'
    echo "Firewall rule added for port 8080"
else
    echo "UFW not found, skipping firewall configuration"
fi

# Step 13: Display next steps
echo ""
echo "=========================================="
echo "✅ Deployment Complete!"
echo "=========================================="
echo ""
echo "Next Steps:"
echo "1. Edit configuration: nano $APP_DIR/.env"
echo "2. Setup MySQL database (see instructions below)"
echo "3. Start the service: systemctl start $SERVICE_NAME"
echo "4. Check status: systemctl status $SERVICE_NAME"
echo "5. View logs: journalctl -u $SERVICE_NAME -f"
echo ""
echo "MySQL Setup:"
echo "  mysql -u root -p"
echo "  CREATE DATABASE tunecent_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
echo "  CREATE USER 'tunecent'@'localhost' IDENTIFIED BY 'your_password';"
echo "  GRANT ALL PRIVILEGES ON tunecent_db.* TO 'tunecent'@'localhost';"
echo "  FLUSH PRIVILEGES;"
echo "  EXIT;"
echo "  mysql -u tunecent -p tunecent_db < $APP_DIR/app/schema.sql"
echo ""
echo "Service Management Commands:"
echo "  systemctl start $SERVICE_NAME     # Start service"
echo "  systemctl stop $SERVICE_NAME      # Stop service"
echo "  systemctl restart $SERVICE_NAME   # Restart service"
echo "  systemctl status $SERVICE_NAME    # Check status"
echo "  journalctl -u $SERVICE_NAME -f    # View logs"
echo ""
echo "Health Check:"
echo "  curl http://localhost:8080/health"
echo ""
echo "=========================================="
