package seed

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/codebypranav/exploraition/internal/ingest"
	pineconeio "github.com/pinecone-io/go-pinecone/v3/pinecone"
	openai "github.com/sashabaranov/go-openai"
	"google.golang.org/protobuf/types/known/structpb"
)

// SeedIndex fetches places from OpenTripMap, generates embeddings, and upserts into Pinecone.
// It accepts an openai client implicitly via env var OPENAI_API_KEY -- see GenerateEmbedding usage.
func SeedIndex(ctx context.Context, idxConn *pineconeio.IndexConnection, openaiClient *openai.Client) error {
	// List of cities to fetch data for -- extendable or configurable via env
	var cities []string
	if v := os.Getenv("SEED_CITIES"); v != "" {
		// comma separated list
		for _, c := range strings.Split(v, ",") {
			cities = append(cities, strings.TrimSpace(c))
		}
	} else {
		cities = []string{"Paris", "New York", "Tokyo", "San Francisco", "Sydney"}
	}
	batchVectors := []*pineconeio.Vector{}
	apiKey := os.Getenv("OPEN_TRIP_MAP_KEY")
	if apiKey == "" {
		return fmt.Errorf("OPEN_TRIP_MAP_KEY not set; cannot seed index")
	}

	for _, city := range cities {
		pois, err := ingest.FetchPOIsForCity(ctx, city, 20000, 200)
		if err != nil {
			log.Printf("failed to fetch POIs for %s: %v", city, err)
			continue
		}
		transformed, err := ingest.TransformAll(ctx, openaiClient, pois)
		if err != nil {
			log.Printf("transforming embeddings failed for %s: %v", city, err)
			continue
		}
		for _, t := range transformed {
			metaStruct, err := structpb.NewStruct(t.Metadata)
			if err != nil {
				log.Printf("failed to convert metadata for %s: %v", t.Id, err)
				continue
			}
			v := &pineconeio.Vector{
				Id:       t.Id,
				Values:   &t.Vector,
				Metadata: metaStruct,
			}
			batchVectors = append(batchVectors, v)
			// Upsert in batches of 100
			if len(batchVectors) >= 100 {
				if _, err := idxConn.UpsertVectors(ctx, batchVectors); err != nil {
					log.Printf("upsert error for batch: %v", err)
				} else {
					log.Printf("upserted batch of %d vectors", len(batchVectors))
				}
				batchVectors = []*pineconeio.Vector{}
			}
		}
	}
	// upsert remainder
	if len(batchVectors) > 0 {
		if _, err := idxConn.UpsertVectors(ctx, batchVectors); err != nil {
			return err
		}
		log.Printf("upserted final batch of %d vectors", len(batchVectors))
	}
	return nil
}
