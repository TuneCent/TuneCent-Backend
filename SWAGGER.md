# TuneCent API - Swagger Documentation

This project includes comprehensive Swagger/OpenAPI documentation for all API endpoints.

## Accessing the Documentation

### Local Development

1. Start the backend server:
   ```bash
   go run cmd/server/main.go
   ```

2. Open your browser and navigate to:
   ```
   http://localhost:8080/swagger/index.html
   ```

### Production

When deployed, access the Swagger UI at:
```
https://tunecent.fabian.web.id/swagger/index.html
```

## Features

- **Interactive API Documentation**: Test all endpoints directly from the browser
- **68 Documented Endpoints** organized by category:
  - Health Check
  - Music NFT Management (4 endpoints)
  - Campaigns (4 endpoints)
  - Royalties (2 endpoints)
  - Users (2 endpoints)
  - Dashboard (8 endpoints)
  - Analytics (8 endpoints)
  - Wallet (4 endpoints)
  - Leaderboard (3 endpoints)
  - Portfolio (4 endpoints)
  - Distribution (5 endpoints)
  - Notifications (7 endpoints)
  - Ledger (4 endpoints)
  - Audit (3 endpoints)
  - Reinvestment (4 endpoints)

## Regenerating Documentation

If you make changes to API endpoints or add new annotations, regenerate the docs:

```bash
# Install swag CLI (if not already installed)
go install github.com/swaggo/swag/cmd/swag@latest

# Generate/update documentation
~/go/bin/swag init -g cmd/server/main.go -o docs
```

## Adding Documentation to New Endpoints

To document a new endpoint, add Swagger annotations above the handler function:

```go
// HandlerName godoc
// @Summary Short description
// @Description Longer description of what this endpoint does
// @Tags CategoryName
// @Accept json
// @Produce json
// @Param paramName path/query/body type true/false "Description"
// @Success 200 {object} ResponseType "Success message"
// @Failure 400 {object} map[string]interface{} "Error message"
// @Router /endpoint/path [get/post/put/delete]
func HandlerName(c *gin.Context) {
    // Handler implementation
}
```

## Swagger Annotations Guide

### Common Annotations

- `@Summary`: Brief one-line description
- `@Description`: Detailed description
- `@Tags`: Category/group for the endpoint
- `@Accept`: Request content type (json, multipart/form-data, etc.)
- `@Produce`: Response content type
- `@Param`: Parameter definition
  - Format: `name location type required "description"`
  - Locations: path, query, header, body, formData
- `@Success`: Success response (code, type, description)
- `@Failure`: Error response (code, type, description)
- `@Router`: Route path and HTTP method

### Example: Music Registration Endpoint

```go
// RegisterMusic handles POST /api/v1/music/register
// @Summary Register new music NFT
// @Description Upload and register a new music NFT with metadata and audio file
// @Tags Music
// @Accept multipart/form-data
// @Produce json
// @Param creator_address formData string true "Creator's wallet address"
// @Param title formData string true "Music title"
// @Param artist formData string true "Artist name"
// @Param audio_file formData file true "Audio file"
// @Success 201 {object} map[string]interface{} "Music registered successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Router /music/register [post]
func (h *MusicHandler) RegisterMusic(c *gin.Context) {
    // Implementation
}
```

## API Information

- **Version**: 1.0
- **Base Path**: /api/v1
- **Schemes**: http, https
- **License**: MIT

## Support

For issues or questions about the API documentation:
- GitHub: https://github.com/tunecent
- Email: support@tunecent.com
