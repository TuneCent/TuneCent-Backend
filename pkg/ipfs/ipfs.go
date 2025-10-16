package ipfs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/tunecent/backend/internal/config"
)

type Service struct {
	apiKey    string
	apiSecret string
	gateway   string
	client    *http.Client
}

type PinataResponse struct {
	IpfsHash  string `json:"IpfsHash"`
	PinSize   int    `json:"PinSize"`
	Timestamp string `json:"Timestamp"`
}

type MusicMetadata struct {
	Title           string `json:"title"`
	Artist          string `json:"artist"`
	Genre           string `json:"genre,omitempty"`
	Description     string `json:"description,omitempty"`
	Duration        int    `json:"duration,omitempty"`
	FingerprintHash string `json:"fingerprint_hash"`
	Creator         string `json:"creator"`
	Timestamp       int64  `json:"timestamp"`
}

func NewService(cfg *config.Config) *Service {
	return &Service{
		apiKey:    cfg.IPFS.PinataAPIKey,
		apiSecret: cfg.IPFS.PinataSecret,
		gateway:   cfg.IPFS.Gateway,
		client:    &http.Client{},
	}
}

// UploadJSON uploads JSON metadata to IPFS via Pinata
func (s *Service) UploadJSON(metadata interface{}) (string, error) {
	jsonData, err := json.Marshal(metadata)
	if err != nil {
		return "", fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "metadata.json")
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = part.Write(jsonData)
	if err != nil {
		return "", fmt.Errorf("failed to write metadata: %w", err)
	}

	writer.Close()

	// Create request
	req, err := http.NewRequest("POST", "https://api.pinata.cloud/pinning/pinFileToIPFS", body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("pinata_api_key", s.apiKey)
	req.Header.Set("pinata_secret_api_key", s.apiSecret)

	// Send request
	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload to IPFS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("pinata API error: %s", string(bodyBytes))
	}

	// Parse response
	var pinataResp PinataResponse
	if err := json.NewDecoder(resp.Body).Decode(&pinataResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return pinataResp.IpfsHash, nil
}

// UploadFile uploads a file to IPFS via Pinata
func (s *Service) UploadFile(fileData []byte, filename string) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = part.Write(fileData)
	if err != nil {
		return "", fmt.Errorf("failed to write file data: %w", err)
	}

	writer.Close()

	req, err := http.NewRequest("POST", "https://api.pinata.cloud/pinning/pinFileToIPFS", body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("pinata_api_key", s.apiKey)
	req.Header.Set("pinata_secret_api_key", s.apiSecret)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload to IPFS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("pinata API error: %s", string(bodyBytes))
	}

	var pinataResp PinataResponse
	if err := json.NewDecoder(resp.Body).Decode(&pinataResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return pinataResp.IpfsHash, nil
}

// GetURL returns the gateway URL for an IPFS CID
func (s *Service) GetURL(cid string) string {
	return fmt.Sprintf("%s%s", s.gateway, cid)
}

// FetchMetadata retrieves metadata from IPFS
func (s *Service) FetchMetadata(cid string) (*MusicMetadata, error) {
	url := s.GetURL(cid)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from IPFS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("IPFS gateway error: status %d", resp.StatusCode)
	}

	var metadata MusicMetadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("failed to decode metadata: %w", err)
	}

	return &metadata, nil
}
