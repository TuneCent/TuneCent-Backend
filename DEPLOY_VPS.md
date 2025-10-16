# 🚀 VPS Deployment - Quick Start

Deploy TuneCent Backend to your VPS in **5 minutes** without Docker.

---

## One-Command Setup

```bash
# 1. Upload to VPS
scp -r TuneCent-Backend root@YOUR_VPS_IP:/tmp/

# 2. SSH and run
ssh root@YOUR_VPS_IP
cd /tmp/TuneCent-Backend
chmod +x quick-setup.sh
./quick-setup.sh
```

That's it! The script automatically:
- ✅ Installs Go, MySQL, Nginx
- ✅ Creates database with random password
- ✅ Builds and deploys application
- ✅ Configures systemd service
- ✅ Sets up firewall
- ✅ Starts the API

---

## After Installation

### 1. Test the API

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "ok",
  "service": "TuneCent Backend API",
  "version": "1.0.0-poc",
  "database": "ok",
  "blockchain": "not_configured"
}
```

### 2. Update Configuration

```bash
nano /opt/tunecent/.env
```

Add your contract addresses (after deployment):
```env
MUSIC_REGISTRY_ADDRESS=0x...
ROYALTY_DISTRIBUTOR_ADDRESS=0x...
CROWDFUNDING_POOL_ADDRESS=0x...
REPUTATION_SCORE_ADDRESS=0x...
```

Add IPFS credentials (from Pinata.cloud):
```env
PINATA_API_KEY=your_key
PINATA_SECRET_KEY=your_secret
```

Restart:
```bash
systemctl restart tunecent-backend
```

### 3. Setup Domain & SSL (Optional)

Edit nginx config:
```bash
nano /opt/tunecent/app/nginx.conf
# Change api.tunecent.com to your domain
```

Enable:
```bash
cp /opt/tunecent/app/nginx.conf /etc/nginx/sites-available/tunecent
ln -s /etc/nginx/sites-available/tunecent /etc/nginx/sites-enabled/
nginx -t
systemctl reload nginx
```

Get SSL certificate:
```bash
apt-get install certbot python3-certbot-nginx
certbot --nginx -d api.yourdomain.com
```

---

## Service Management

```bash
# Start/stop/restart
systemctl start tunecent-backend
systemctl stop tunecent-backend
systemctl restart tunecent-backend

# View status
systemctl status tunecent-backend

# View logs (live)
journalctl -u tunecent-backend -f

# View application logs
tail -f /opt/tunecent/logs/app.log
```

---

## Update Application

```bash
# Upload new code to /opt/tunecent/app/
# Then run:
cd /opt/tunecent/app
./update.sh
```

---

## Troubleshooting

### Service won't start?
```bash
# Check logs
journalctl -u tunecent-backend -n 50

# Test binary
sudo -u tunecent /opt/tunecent/bin/tunecent-backend
```

### Can't connect to database?
```bash
# Test connection
mysql -u tunecent -p tunecent_db

# Check credentials
cat /opt/tunecent/.env | grep DB_
```

### Port already in use?
```bash
# Check what's using port 8080
lsof -i :8080

# Change port in .env
nano /opt/tunecent/.env
# PORT=8081
```

---

## File Locations

```
/opt/tunecent/
├── bin/tunecent-backend           # Binary
├── app/                           # Source code
├── logs/                          # Log files
│   ├── app.log
│   └── error.log
└── .env                           # Configuration

/etc/systemd/system/tunecent-backend.service
/etc/nginx/sites-available/tunecent
```

---

## Security Checklist

After deployment:

- [ ] Change default database password in `.env`
- [ ] Set strong `JWT_SECRET` in `.env`
- [ ] Enable firewall (UFW)
- [ ] Setup SSL certificate (Let's Encrypt)
- [ ] Restrict `.env` permissions (`chmod 600`)
- [ ] Enable automatic security updates
- [ ] Setup backup strategy

---

## Quick Reference

### Health Check
```bash
curl http://localhost:8080/health
```

### API Test
```bash
# Register music (example)
curl -X POST http://localhost:8080/api/v1/music/register \
  -F "creator_address=0x123..." \
  -F "title=Test Song" \
  -F "artist=Artist" \
  -F "audio_file=@song.mp3"
```

### Database Backup
```bash
mysqldump -u tunecent -p tunecent_db > backup.sql
```

### View Logs
```bash
journalctl -u tunecent-backend -f
```

---

## Need More Details?

See [DEPLOYMENT_GUIDE.md](../DEPLOYMENT_GUIDE.md) for:
- Manual installation steps
- Advanced configuration
- Nginx setup details
- Monitoring & maintenance
- Complete troubleshooting guide

---

## Support

If issues persist:
1. Check logs: `journalctl -u tunecent-backend -n 100`
2. Verify config: `cat /opt/tunecent/.env`
3. Test database: `mysql -u tunecent -p`
4. Check service: `systemctl status tunecent-backend`

---

**Your TuneCent backend is now live on VPS!** 🎉

Access it at:
- HTTP: `http://YOUR_VPS_IP:8080`
- With domain: `https://api.yourdomain.com`
