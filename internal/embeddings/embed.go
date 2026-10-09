package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	embedURL = "https://api.pinecone.io/embed"
	// The embed endpoint 404s without an explicit API version header.
	apiVersion = "2025-04"
	model      = "pinecone-sparse-english-v0"

	// InputQuery is for search text, InputPassage for documents being indexed.
	// The model requires one of the two; it has no default.
	InputQuery   = "query"
	InputPassage = "passage"

	// The embed endpoint accepts at most 96 inputs per request.
	maxBatch = 64
)

// SparseEmbedding is a sparse vector: Indices and Values are parallel slices
// holding only the non-zero dimensions.
type SparseEmbedding struct {
	Indices []uint32
	Values  []float32
}

// Empty reports whether the embedding has no non-zero dimensions, which happens
// when the input has no terms the model recognises.
func (s SparseEmbedding) Empty() bool {
	return len(s.Indices) == 0
}

// GenerateEmbedding embeds a single search query.
func GenerateEmbedding(ctx context.Context, input string) (SparseEmbedding, error) {
	out, err := GenerateEmbeddings(ctx, []string{input}, InputQuery)
	if err != nil {
		return SparseEmbedding{}, err
	}
	if len(out) == 0 {
		return SparseEmbedding{}, fmt.Errorf("embedding API returned no vectors")
	}
	return out[0], nil
}

// GenerateEmbeddings embeds a batch of inputs, splitting into as many requests
// as the endpoint's batch limit requires. inputType must be InputQuery or
// InputPassage.
func GenerateEmbeddings(ctx context.Context, inputs []string, inputType string) ([]SparseEmbedding, error) {
	apiKey := os.Getenv("PINECONE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("PINECONE_API_KEY not set")
	}
	if len(inputs) == 0 {
		return nil, nil
	}

	out := make([]SparseEmbedding, 0, len(inputs))
	for start := 0; start < len(inputs); start += maxBatch {
		end := start + maxBatch
		if end > len(inputs) {
			end = len(inputs)
		}
		batch, err := embedBatch(ctx, apiKey, inputs[start:end], inputType)
		if err != nil {
			return nil, err
		}
		out = append(out, batch...)
	}
	return out, nil
}

func embedBatch(ctx context.Context, apiKey string, inputs []string, inputType string) ([]SparseEmbedding, error) {
	// Inputs are objects, not bare strings: [{"text": "..."}].
	items := make([]map[string]string, 0, len(inputs))
	for _, in := range inputs {
		items = append(items, map[string]string{"text": in})
	}
	reqBody := map[string]interface{}{
		"model":      model,
		"parameters": map[string]string{"input_type": inputType},
		"inputs":     items,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, embedURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", apiKey)
	req.Header.Set("X-Pinecone-API-Version", apiVersion)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
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

	var respData struct {
		Data []struct {
			VectorType    string    `json:"vector_type"`
			SparseIndices []uint32  `json:"sparse_indices"`
			SparseValues  []float32 `json:"sparse_values"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &respData); err != nil {
		return nil, fmt.Errorf("failed to parse embedding response: %w", err)
	}

	out := make([]SparseEmbedding, 0, len(respData.Data))
	for _, d := range respData.Data {
		out = append(out, SparseEmbedding{Indices: d.SparseIndices, Values: d.SparseValues})
	}
	return out, nil
}
