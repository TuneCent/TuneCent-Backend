# Postman Collection Guide

## 📦 Import Collection

### Step 1: Import to Postman

1. Open Postman
2. Click **Import** button (top left)
3. Select **File** tab
4. Choose `TuneCent-API.postman_collection.json`
5. Click **Import**

### Step 2: Verify Import

You should see a new collection named **"TuneCent Backend API"** with 6 folders:
- Health & System
- Music
- Campaigns
- Royalties
- Users

---

## 🔧 Configuration

The collection uses a variable:
- **base_url** = `http://localhost:8080`

If your server runs on a different port, update it:
1. Click on the collection name
2. Go to **Variables** tab
3. Change `base_url` value
4. Click **Save**

---

## 📚 Collection Contents

### 1. Health & System (1 endpoint)
- **GET** `/health` - Check API status

### 2. Music (4 endpoints)
- **POST** `/api/v1/music/register` - Register music with NFT
- **GET** `/api/v1/music/:tokenId` - Get music details
- **GET** `/api/v1/music` - List all music (with pagination)
- **GET** `/api/v1/music/:tokenId/analytics` - Get usage analytics

### 3. Campaigns (4 endpoints)
- **POST** `/api/v1/campaigns` - Create campaign
- **GET** `/api/v1/campaigns/:campaignId` - Get campaign
- **GET** `/api/v1/campaigns` - List campaigns
- **POST** `/api/v1/campaigns/:campaignId/contribute` - Contribute

### 4. Royalties (2 endpoints)
- **GET** `/api/v1/royalties/token/:tokenId` - Get royalties
- **POST** `/api/v1/royalties/simulate` - Simulate payment

### 5. Users (2 endpoints)
- **GET** `/api/v1/users/:address` - Get user profile
- **GET** `/api/v1/users/:address/reputation` - Get reputation

**Total: 13 endpoints**

---

## 🎯 Example Responses

Each endpoint includes **example responses** for:
- ✅ Success cases
- ❌ Error cases
- 📊 Different scenarios

Click on any request → **Examples** tab to see saved responses.

---

## 🚀 Quick Test Flow

### 1. Health Check
```
GET /health
```
Verify server is running.

### 2. Register Music
```
POST /api/v1/music/register
```
Note the returned `token_id` (e.g., 1729123456)

### 3. Get Music Details
```
GET /api/v1/music/1729123456
```
Replace with your token_id

### 4. Create Campaign
```
POST /api/v1/campaigns
Body: {
  "token_id": 1729123456,
  "creator_address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
  "goal_amount": "5000000000000000000",
  "royalty_percentage": 2000,
  "duration_days": 30,
  "lockup_days": 90
}
```

### 5. Contribute to Campaign
```
POST /api/v1/campaigns/1/contribute
Body: {
  "contributor_address": "0x8ba1f109551bD432803012645Ac136ddd64DBA72",
  "amount": "1000000000000000000"
}
```

### 6. Simulate Royalty Payment
```
POST /api/v1/royalties/simulate
Body: {
  "token_id": 1729123456,
  "platform": "TikTok",
  "amount": "100000000000000000"
}
```

### 7. Check Royalties
```
GET /api/v1/royalties/token/1729123456
```

### 8. Get Analytics
```
GET /api/v1/music/1729123456/analytics
```

---

## 📝 Request Details

### Music Registration

**Endpoint:** `POST /api/v1/music/register`

**Body Type:** `multipart/form-data`

**Fields:**
- `creator_address` (required) - Ethereum wallet address
- `title` (required) - Song title
- `artist` (required) - Artist name
- `genre` (optional) - Music genre
- `description` (optional) - Description
- `duration` (optional) - Duration in seconds
- `audio_file` (required) - Audio file

**Note:** To test file upload:
1. Select the request
2. Go to **Body** tab
3. Click on `audio_file` field
4. Click **Select Files** and choose an MP3/WAV file

---

### Campaign Creation

**Endpoint:** `POST /api/v1/campaigns`

**Body Type:** `application/json`

**Sample Body:**
```json
{
  "token_id": 1729123456,
  "creator_address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
  "goal_amount": "5000000000000000000",
  "royalty_percentage": 2000,
  "duration_days": 30,
  "lockup_days": 90
}
```

**Field Explanations:**
- `goal_amount`: In wei (5 ETH = 5000000000000000000)
- `royalty_percentage`: In basis points (2000 = 20%)
- `duration_days`: Campaign duration (1-90 days)
- `lockup_days`: Lock-up period for investors

---

### Royalty Simulation

**Endpoint:** `POST /api/v1/royalties/simulate`

**Sample Body:**
```json
{
  "token_id": 1729123456,
  "platform": "TikTok",
  "amount": "100000000000000000"
}
```

**Amount Examples:**
- 0.01 ETH = `10000000000000000`
- 0.1 ETH = `100000000000000000`
- 1 ETH = `1000000000000000000`

---

## 🔄 Testing Workflow

### Scenario 1: Full Music Registration Flow

1. **Register Music** → Get `token_id`
2. **Get Music** → Verify registration
3. **List Music** → See it in the list
4. **Get Analytics** → Check initial stats

### Scenario 2: Crowdfunding Flow

1. **Register Music** → Get `token_id`
2. **Create Campaign** → Get `campaign_id`
3. **List Campaigns** → See active campaigns
4. **Contribute** → Add funds
5. **Get Campaign** → Check raised amount

### Scenario 3: Royalty Flow

1. **Register Music** → Get `token_id`
2. **Simulate Payment** → Add royalty (repeat multiple times)
3. **Get Royalties** → See payment history
4. **Get Analytics** → Check total royalties

---

## 💡 Tips

### Use Path Variables

Endpoints with `:tokenId` or `:campaignId` use **path variables**.

To change them:
1. Click on the request
2. Go to **Params** tab
3. Edit **Path Variables** section

### Save Responses

Right-click on any request → **Add Example** to save custom responses.

### Environment Variables

Create environments for different setups:
- **Local** - http://localhost:8080
- **VPS Dev** - http://your-vps:8080
- **Production** - https://api.tunecent.com

---

## 🐛 Troubleshooting

### "Could not get response"
- Check if backend is running: `curl http://localhost:8080/health`
- Verify `base_url` variable matches your server

### "404 Not Found"
- Ensure endpoint path is correct
- Check server logs for routing issues

### "400 Bad Request"
- Verify request body format
- Check required fields are present
- Ensure JSON is valid

### "500 Internal Server Error"
- Check database connection
- View server logs for details

---

## 📊 Response Examples

All requests include example responses showing:

**Success Response:**
```json
{
  "token_id": 1729123456,
  "message": "Success message",
  "data": {...}
}
```

**Error Response:**
```json
{
  "error": "Error description"
}
```

**Paginated Response:**
```json
{
  "data": [...],
  "total": 100,
  "limit": 20,
  "offset": 0
}
```

---

## 🎨 Collection Features

✅ Complete endpoint coverage (13 endpoints)
✅ Detailed descriptions for each request
✅ Example requests with real data
✅ Saved example responses (success + errors)
✅ Path variables pre-configured
✅ Request bodies pre-filled
✅ Ready to use immediately

---

## 📚 Additional Resources

- **API Documentation**: See README.md
- **Backend Code**: `cmd/server/main_complete.go`
- **Database Schema**: `schema.sql`

---

**Ready to test! Start with the Health Check endpoint.** ✨
