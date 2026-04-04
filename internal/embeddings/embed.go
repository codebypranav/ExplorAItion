package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// GenerateEmbedding generates a sparse embedding using Pinecone's serverless embedding service
// Uses PINECONE_API_KEY from environment for authentication
func GenerateEmbedding(ctx context.Context, input string) ([]float32, error) {
	// Call Pinecone's embedding service via HTTP API
	apiKey := os.Getenv("PINECONE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("PINECONE_API_KEY not set")
	}

	// Pinecone embedding API endpoint
	embeddingURL := "https://api.pinecone.io/embed"

	// Prepare the request body
	reqBody := map[string]interface{}{
		"model": "pinecone-sparse-english-v0",
		"inputs": []string{input},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", embeddingURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", apiKey)

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call embedding API: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var respData struct {
		Data []struct {
			Values []float32 `json:"values"`
		} `json:"data"`
	}

	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedding response: %w", err)
	}

	if len(respData.Data) == 0 {
		return []float32{}, nil
	}

	return respData.Data[0].Values, nil
}
