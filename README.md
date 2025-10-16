# TuneCent Backend - Golang REST API

Complete backend system for TuneCent music rights management platform.

## 🏗️ Architecture

```
TuneCent-Backend/
├── cmd/
│   └── server/
│       ├── main.go              # Original template
│       └── main_complete.go     # Complete implementation
├── internal/
│   ├── config/                  # Configuration management
│   ├── database/                # Database connection & migrations
│   ├── blockchain/              # Blockchain client & service
│   ├── handlers/                # HTTP request handlers
│   ├── services/                # Business logic
│   └── models/                  # Database models (GORM)
├── pkg/
│   ├── ipfs/                    # IPFS service (Pinata)
│   └── fingerprint/             # Audio fingerprinting (mock)
├── deploy.sh                    # VPS deployment script
├── quick-setup.sh               # One-command setup
├── update.sh                    # Update script
├── nginx.conf                   # Nginx reverse proxy config
├── Makefile                     # Build automation
└── schema.sql                   # MySQL schema
```

## ✨ Features Implemented

### Core Services
- ✅ **Config Management** - Environment-based configuration
- ✅ **Database Layer** - MySQL with GORM ORM
- ✅ **Blockchain Integration** - go-ethereum client
- ✅ **IPFS Storage** - Pinata API integration
- ✅ **Audio Fingerprinting** - Mock implementation (ready for real algorithm)

### API Endpoints

#### Music Registration
- `POST /api/v1/music/register` - Register music with NFT minting
- `GET /api/v1/music/:tokenId` - Get music metadata
- `GET /api/v1/music` - List all music (with pagination)
- `GET /api/v1/music/:tokenId/analytics` - Get usage analytics

#### Crowdfunding Campaigns
- `POST /api/v1/campaigns` - Create funding campaign
- `GET /api/v1/campaigns/:campaignId` - Get campaign details
- `GET /api/v1/campaigns` - List campaigns (filterable by status)
- `POST /api/v1/campaigns/:campaignId/contribute` - Contribute to campaign

#### Royalty Management
- `GET /api/v1/royalties/token/:tokenId` - Get royalty payments
- `POST /api/v1/royalties/simulate` - Simulate payment (PoC demo)

#### User & Reputation
- `GET /api/v1/users/:address` - Get user profile
- `GET /api/v1/users/:address/reputation` - Get reputation score

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- MySQL 8.0+

### Local Development

1. **Install Dependencies**
   ```bash
   make install
   ```

2. **Setup Database**
   ```bash
   # Create database
   mysql -u root -p -e "CREATE DATABASE tunecent_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

   # Load schema
   mysql -u root -p tunecent_db < schema.sql
   ```

3. **Configure Environment**
   ```bash
   cp .env.example .env
   # Edit .env with your settings
   ```

4. **Run Server**
   ```bash
   make run
   ```

5. **Test API**
   ```bash
   curl http://localhost:8080/health
   ```

### VPS Deployment

See [DEPLOY_VPS.md](DEPLOY_VPS.md) for one-command VPS deployment:

```bash
# Upload to VPS
scp -r . root@YOUR_VPS_IP:/tmp/tunecent-backend

# SSH and run
ssh root@YOUR_VPS_IP
cd /tmp/tunecent-backend
./quick-setup.sh
```

## 📖 Usage Examples

### Register Music

```bash
curl -X POST http://localhost:8080/api/v1/music/register \
  -F "creator_address=0x1234567890123456789012345678901234567890" \
  -F "title=My Song" \
  -F "artist=Artist Name" \
  -F "genre=Pop" \
  -F "duration=180" \
  -F "audio_file=@/path/to/song.mp3"
```

### Get Music Details

```bash
curl http://localhost:8080/api/v1/music/1
```

### Create Campaign

```bash
curl -X POST http://localhost:8080/api/v1/campaigns \
  -H "Content-Type: application/json" \
  -d '{
    "token_id": 1,
    "creator_address": "0x123...",
    "goal_amount": "1000000000000000000",
    "royalty_percentage": 2000,
    "duration_days": 30,
    "lockup_days": 90
  }'
```

### Simulate Royalty Payment

```bash
curl -X POST http://localhost:8080/api/v1/royalties/simulate \
  -H "Content-Type: application/json" \
  -d '{
    "token_id": 1,
    "platform": "TikTok",
    "amount": "10000000000000000"
  }'
```

## 🛠️ Development

### Available Make Commands

```bash
make help              # Show all commands
make install           # Install dependencies
make build             # Build binary
make run               # Run development server
make test              # Run tests
make lint              # Run linter
make fmt               # Format code
make clean             # Clean build artifacts
make prod              # Build for production
```

### Project Structure Explained

- **cmd/server/** - Application entry point
- **internal/** - Private application code
  - **config/** - Environment configuration
  - **database/** - Database setup & migrations
  - **blockchain/** - Smart contract interaction
  - **handlers/** - HTTP handlers (controllers)
  - **services/** - Business logic layer
  - **models/** - Data models (GORM entities)
- **pkg/** - Public libraries (can be imported by other projects)
  - **ipfs/** - IPFS/Pinata integration
  - **fingerprint/** - Audio fingerprinting logic

## 🔌 Database Schema

8 main tables:
- `users` - Platform users
- `music_metadata` - Music NFT metadata
- `campaigns` - Crowdfunding campaigns
- `contributions` - Campaign contributions
- `royalty_payments` - Incoming royalty payments
- `royalty_distributions` - Outgoing distributions
- `usage_detections` - Platform usage tracking
- `analytics` - Aggregated analytics

See `schema.sql` for full schema.

## 🔗 Integration with Smart Contracts

To enable full blockchain integration:

1. **Deploy Contracts** (see TuneCent-SmartContract/README.md)

2. **Update .env** with contract addresses:
   ```
   MUSIC_REGISTRY_ADDRESS=0x...
   ROYALTY_DISTRIBUTOR_ADDRESS=0x...
   CROWDFUNDING_POOL_ADDRESS=0x...
   REPUTATION_SCORE_ADDRESS=0x...
   ```

3. **Generate Contract Bindings** (requires abigen):
   ```bash
   make generate-bindings
   ```

## 📊 API Response Format

### Success Response
```json
{
  "token_id": 1,
  "ipfs_cid": "QmXxxx...",
  "fingerprint_hash": "0x123...",
  "message": "Music registered successfully"
}
```

### Error Response
```json
{
  "error": "Error message here"
}
```

### Paginated Response
```json
{
  "data": [...],
  "total": 100,
  "limit": 20,
  "offset": 0
}
```

## 🧪 Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run linter
make lint

# Run all checks
make check
```

## 🔐 Environment Variables

See `.env.example` for all available configuration options:

- **Server**: `PORT`, `ENV`
- **Database**: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- **Blockchain**: `RPC_URL`, `CHAIN_ID`, contract addresses
- **IPFS**: `IPFS_GATEWAY`, `PINATA_API_KEY`, `PINATA_SECRET_KEY`
- **Security**: `JWT_SECRET`

## 🚀 Deployment

### VPS Deployment (Recommended)

See [DEPLOY_VPS.md](DEPLOY_VPS.md) for quick deployment guide.

Or see [../DEPLOYMENT_GUIDE.md](../DEPLOYMENT_GUIDE.md) for complete deployment instructions with:
- Automated installation
- MySQL setup
- Nginx configuration
- SSL with Let's Encrypt
- Security hardening
- Monitoring & maintenance

## 🎯 PoC Limitations

- Audio fingerprinting is mocked (SHA256 hash)
- Smart contract calls are simulated
- IPFS upload requires valid Pinata credentials
- No authentication/authorization implemented
- Basic error handling

## 🚧 Production Readiness Checklist

- [ ] Implement real audio fingerprinting (Chromaprint/AcoustID)
- [ ] Complete smart contract integration with abigen bindings
- [ ] Add JWT authentication middleware
- [ ] Implement rate limiting
- [ ] Add comprehensive error handling
- [ ] Write unit & integration tests
- [ ] Setup logging (structured logging with Zap/Zerolog)
- [ ] Add metrics & monitoring (Prometheus)
- [ ] Implement caching layer (Redis)
- [ ] Add API documentation (Swagger/OpenAPI)
- [ ] Security audit
- [ ] Load testing

## 📝 API Documentation

Health check endpoint provides system status:

```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "ok",
  "service": "TuneCent Backend API",
  "version": "1.0.0-poc",
  "database": "ok",
  "blockchain": "not_configured"
}
```

## 🤝 Contributing

This is a hackathon PoC. For production use, please implement the production readiness checklist above.

## 📄 License

MIT License - See LICENSE file

---

**Built with Go, Gin, GORM, and go-ethereum for TuneCent PoC**
