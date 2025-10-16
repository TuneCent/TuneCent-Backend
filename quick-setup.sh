#!/bin/bash

# TuneCent Backend - Quick Setup Script for VPS
# One-command setup for fresh Ubuntu/Debian VPS

set -e

echo "=========================================="
echo "🚀 TuneCent Backend - Quick Setup"
echo "=========================================="
echo ""
echo "This script will:"
echo "  - Install all dependencies (Go, MySQL, Nginx)"
echo "  - Setup the application"
echo "  - Configure the database"
echo "  - Start the service"
echo ""
read -p "Continue? (y/n): " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 1
fi

# Check root
if [[ $EUID -ne 0 ]]; then
   echo "❌ This script must be run as root"
   echo "Run: sudo ./quick-setup.sh"
   exit 1
fi

echo ""
echo "📦 Step 1: Installing system packages..."
apt-get update
apt-get install -y wget curl git build-essential mysql-server nginx ufw

echo ""
echo "🔧 Step 2: Installing Go..."
if ! command -v go &> /dev/null; then
    wget -q https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
    tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
    rm go1.21.6.linux-amd64.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    export PATH=$PATH:/usr/local/go/bin
    echo "✅ Go installed: $(go version)"
else
    echo "✅ Go already installed: $(go version)"
fi

echo ""
echo "👤 Step 3: Creating application user..."
if ! id -u tunecent > /dev/null 2>&1; then
    useradd -r -s /bin/bash -d /opt/tunecent tunecent
    echo "✅ User created"
else
    echo "✅ User already exists"
fi

echo ""
echo "📁 Step 4: Setting up directories..."
mkdir -p /opt/tunecent/{bin,logs,app}

echo ""
echo "🏗️  Step 5: Building application..."
cp -r . /opt/tunecent/app/
cd /opt/tunecent/app
go mod download
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o /opt/tunecent/bin/tunecent-backend ./cmd/server/main_complete.go
echo "✅ Binary built"

echo ""
echo "⚙️  Step 6: Configuring environment..."
if [ ! -f /opt/tunecent/.env ]; then
    cp .env.example /opt/tunecent/.env

    # Generate random JWT secret
    JWT_SECRET=$(openssl rand -base64 32)
    sed -i "s/your_jwt_secret_here/$JWT_SECRET/" /opt/tunecent/.env

    echo "✅ Environment file created"
fi

echo ""
echo "🗄️  Step 7: Setting up MySQL..."
# Generate random password
DB_PASSWORD=$(openssl rand -base64 16 | tr -d "=+/" | cut -c1-16)

# Update .env with password
sed -i "s/your_password_here/$DB_PASSWORD/" /opt/tunecent/.env

# Create database
mysql -e "CREATE DATABASE IF NOT EXISTS tunecent_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
mysql -e "CREATE USER IF NOT EXISTS 'tunecent'@'localhost' IDENTIFIED BY '$DB_PASSWORD';"
mysql -e "GRANT ALL PRIVILEGES ON tunecent_db.* TO 'tunecent'@'localhost';"
mysql -e "FLUSH PRIVILEGES;"

# Load schema
mysql tunecent_db < schema.sql

echo "✅ Database configured"
echo "   Database: tunecent_db"
echo "   User: tunecent"
echo "   Password: $DB_PASSWORD (saved in .env)"

echo ""
echo "🔐 Step 8: Setting permissions..."
chown -R tunecent:tunecent /opt/tunecent
chmod 600 /opt/tunecent/.env
chmod +x /opt/tunecent/bin/tunecent-backend

echo ""
echo "🔧 Step 9: Creating systemd service..."
cat > /etc/systemd/system/tunecent-backend.service <<'EOF'
[Unit]
Description=TuneCent Backend API Service
After=network.target mysql.service

[Service]
Type=simple
User=tunecent
Group=tunecent
WorkingDirectory=/opt/tunecent
EnvironmentFile=/opt/tunecent/.env
ExecStart=/opt/tunecent/bin/tunecent-backend
Restart=always
RestartSec=10
StandardOutput=append:/opt/tunecent/logs/app.log
StandardError=append:/opt/tunecent/logs/error.log
NoNewPrivileges=true
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable tunecent-backend
echo "✅ Service configured"

echo ""
echo "🛡️  Step 10: Configuring firewall..."
ufw --force enable
ufw allow 22/tcp comment 'SSH'
ufw allow 80/tcp comment 'HTTP'
ufw allow 443/tcp comment 'HTTPS'
ufw allow 8080/tcp comment 'TuneCent API'
echo "✅ Firewall configured"

echo ""
echo "🚀 Step 11: Starting service..."
systemctl start tunecent-backend
sleep 2

echo ""
echo "=========================================="
echo "✅ Installation Complete!"
echo "=========================================="
echo ""

# Check if service is running
if systemctl is-active --quiet tunecent-backend; then
    echo "✅ Service is running!"
    echo ""
    echo "🧪 Testing health endpoint..."
    if curl -s http://localhost:8080/health > /dev/null; then
        echo "✅ API is responding!"
    else
        echo "⚠️  API not responding yet (may need a moment)"
    fi
else
    echo "⚠️  Service failed to start"
    echo "Check logs: journalctl -u tunecent-backend -n 50"
fi

echo ""
echo "📋 Next Steps:"
echo ""
echo "1. Edit configuration (optional):"
echo "   nano /opt/tunecent/.env"
echo ""
echo "2. Deploy smart contracts and update addresses in .env:"
echo "   MUSIC_REGISTRY_ADDRESS=0x..."
echo "   ROYALTY_DISTRIBUTOR_ADDRESS=0x..."
echo "   etc."
echo ""
echo "3. Add IPFS/Pinata credentials in .env:"
echo "   PINATA_API_KEY=..."
echo "   PINATA_SECRET_KEY=..."
echo ""
echo "4. Restart after config changes:"
echo "   systemctl restart tunecent-backend"
echo ""
echo "5. Setup Nginx reverse proxy (recommended):"
echo "   cp /opt/tunecent/app/nginx.conf /etc/nginx/sites-available/tunecent"
echo "   # Edit domain name in the file"
echo "   ln -s /etc/nginx/sites-available/tunecent /etc/nginx/sites-enabled/"
echo "   nginx -t && systemctl reload nginx"
echo ""
echo "6. Setup SSL with Let's Encrypt:"
echo "   certbot --nginx -d api.yourdomain.com"
echo ""
echo "=========================================="
echo "📊 Useful Commands:"
echo "=========================================="
echo ""
echo "Service Management:"
echo "  systemctl status tunecent-backend    # Check status"
echo "  systemctl restart tunecent-backend   # Restart"
echo "  journalctl -u tunecent-backend -f    # View logs"
echo ""
echo "Health Check:"
echo "  curl http://localhost:8080/health"
echo ""
echo "View Configuration:"
echo "  cat /opt/tunecent/.env"
echo ""
echo "Application Logs:"
echo "  tail -f /opt/tunecent/logs/app.log"
echo ""
echo "=========================================="
echo ""
echo "🎉 Your TuneCent backend is ready!"
echo ""
echo "Database Password: $DB_PASSWORD"
echo "(Also saved in /opt/tunecent/.env)"
echo ""
echo "=========================================="
