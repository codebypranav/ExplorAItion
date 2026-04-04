package ingest

import (
	"context"
	"fmt"
	"strings"

	"github.com/codebypranav/exploraition/internal/embeddings"
)

// TransformedPOI couples a POI with the text we will embed and metadata for Pinecone
type TransformedPOI struct {
	Id       string
	Vector   []float32
	Metadata map[string]interface{}
}

// BuildTextForEmbedding creates a textual summary for the POI
func BuildTextForEmbedding(p POI) string {
	parts := []string{p.Name}
	if p.Description != "" {
		parts = append(parts, p.Description)
	}
	if p.Kinds != "" {
		parts = append(parts, fmt.Sprintf("types: %s", p.Kinds))
	}
	if p.Country != "" {
		parts = append(parts, fmt.Sprintf("country: %s", p.Country))
	}
	return strings.Join(parts, " \n")
}

// TransformAll converts a slice of POI to TransformedPOI by generating sparse embeddings using Pinecone
func TransformAll(ctx context.Context, pois []POI) ([]TransformedPOI, error) {
	out := make([]TransformedPOI, 0, len(pois))
	for _, p := range pois {
		text := BuildTextForEmbedding(p)
		vec, err := embeddings.GenerateEmbedding(ctx, text)
		if err != nil {
			return nil, err
		}
		m := map[string]interface{}{
			"name":    p.Name,
			"xid":     p.XID,
			"country": p.Country,
			"kinds":   p.Kinds,
			"lat":     p.Lat,
			"lon":     p.Lon,
		}
		if p.Description != "" {
			m["description"] = p.Description
		}
		if len(p.Tags) > 0 {
			m["tags"] = p.Tags
		}
		if p.Image != "" {
			m["image"] = p.Image
		}
		if p.Rate != 0 {
			m["rate"] = p.Rate
		}

		// Derive a simple `suitable_for_hiking` flag from kinds
		suitable := false
		if strings.Contains(strings.ToLower(p.Kinds), "hiking") || strings.Contains(strings.ToLower(p.Kinds), "trail") || strings.Contains(strings.ToLower(p.Kinds), "park") {
			suitable = true
		}
		m["suitable_for_hiking"] = suitable
		out = append(out, TransformedPOI{
			Id:       p.XID,
			Vector:   vec,
			Metadata: m,
		})
	}
	return out, nil
}
