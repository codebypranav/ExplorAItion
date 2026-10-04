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

// Input types for Pinecone's sparse model: documents are embedded as
// "passage", user queries as "query".
const (
	InputPassage = "passage"
	InputQuery   = "query"
)

const model = "pinecone-sparse-english-v0"

// embedURL is a variable so tests can point it at a fake server.
var embedURL = "https://api.pinecone.io/embed"

// Sparse is a sparse embedding: Values[i] is the weight of token Indices[i].
type Sparse struct {
	Indices []uint32
	Values  []float32
}

// GenerateEmbedding generates a sparse embedding with Pinecone's hosted inference API.
// inputType must be InputPassage (when indexing) or InputQuery (when searching).
// Uses PINECONE_API_KEY from the environment for authentication.
func GenerateEmbedding(ctx context.Context, input, inputType string) (*Sparse, error) {
	apiKey := os.Getenv("PINECONE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("PINECONE_API_KEY not set")
	}

	bodyBytes, err := json.Marshal(map[string]interface{}{
		"model":      model,
		"parameters": map[string]string{"input_type": inputType, "truncate": "END"},
		"inputs":     []map[string]string{{"text": input}},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, embedURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", apiKey)
	req.Header.Set("X-Pinecone-API-Version", "2025-01")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call embedding API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Sparse models return sparse_indices/sparse_values, not "values".
	var respData struct {
		Data []struct {
			SparseIndices []uint32  `json:"sparse_indices"`
			SparseValues  []float32 `json:"sparse_values"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &respData); err != nil {
		return nil, fmt.Errorf("failed to parse embedding response: %w", err)
	}
	if len(respData.Data) == 0 {
		return nil, fmt.Errorf("embedding API returned no data")
	}
	d := respData.Data[0]
	if len(d.SparseValues) == 0 || len(d.SparseValues) != len(d.SparseIndices) {
		return nil, fmt.Errorf("embedding has no usable sparse values (%d indices, %d values)", len(d.SparseIndices), len(d.SparseValues))
	}
	return &Sparse{Indices: d.SparseIndices, Values: d.SparseValues}, nil
}
