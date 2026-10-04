package embeddings

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func fakeServer(t *testing.T, status int, body string, gotReq *map[string]interface{}) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if gotReq != nil {
			_ = json.NewDecoder(r.Body).Decode(gotReq)
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	old := embedURL
	embedURL = srv.URL
	t.Cleanup(func() { embedURL = old; srv.Close() })
}

func TestGenerateEmbeddingParsesSparseResponse(t *testing.T) {
	t.Setenv("PINECONE_API_KEY", "test")
	var got map[string]interface{}
	fakeServer(t, 200, `{"model":"pinecone-sparse-english-v0","vector_type":"sparse","data":[{"vector_type":"sparse","sparse_indices":[10,42],"sparse_values":[0.5,1.5]}]}`, &got)

	emb, err := GenerateEmbedding(context.Background(), "art museums", InputQuery)
	if err != nil {
		t.Fatal(err)
	}
	if len(emb.Indices) != 2 || emb.Indices[1] != 42 || emb.Values[0] != 0.5 {
		t.Fatalf("unexpected embedding: %+v", emb)
	}
	if p := got["parameters"].(map[string]interface{}); p["input_type"] != "query" {
		t.Fatalf("input_type not sent: %v", got["parameters"])
	}
}

func TestGenerateEmbeddingRejectsEmptyEmbedding(t *testing.T) {
	t.Setenv("PINECONE_API_KEY", "test")
	// This is what the old code silently accepted: a response with no dense "values".
	fakeServer(t, 200, `{"data":[{"vector_type":"sparse","sparse_indices":[],"sparse_values":[]}]}`, nil)
	if _, err := GenerateEmbedding(context.Background(), "x", InputQuery); err == nil {
		t.Fatal("expected error for empty embedding")
	}
}

func TestGenerateEmbeddingAPIError(t *testing.T) {
	t.Setenv("PINECONE_API_KEY", "test")
	fakeServer(t, 400, `{"error":"bad"}`, nil)
	if _, err := GenerateEmbedding(context.Background(), "x", InputQuery); err == nil {
		t.Fatal("expected error for non-200")
	}
}
