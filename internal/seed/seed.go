package seed

import (
	"context"
	"log"

	pineconeio "github.com/pinecone-io/go-pinecone/v3/pinecone"
	"google.golang.org/protobuf/types/known/structpb"
)

type Location struct {
	Id       string
	Vector   []float32
	Metadata map[string]interface{}
}

func SeedIndex(ctx context.Context, idxConn *pineconeio.IndexConnection) error {
	locations := []Location{
		{
			Id:     "loc1",
			Vector: []float32{0.12, 0.34, 0.56, 0.78},
			Metadata: map[string]interface{}{
				"name":    "Eiffel Tower",
				"country": "France",
			},
		},
		{
			Id:     "loc2",
			Vector: []float32{0.22, 0.44, 0.66, 0.88},
			Metadata: map[string]interface{}{
				"name":    "Grand Canyon",
				"country": "USA",
			},
		},
	}

	var vectors []*pineconeio.Vector
	for _, loc := range locations {
		metaStruct, err := structpb.NewStruct(loc.Metadata)
		if err != nil {
			return err
		}

		v := &pineconeio.Vector{
			Id:       loc.Id,
			Values:   &loc.Vector,
			Metadata: metaStruct,
		}
		vectors = append(vectors, v)
	}

	count, err := idxConn.UpsertVectors(ctx, vectors)
	if err != nil {
		return err
	}
	log.Printf("upserted %d vectors", count)
	return nil
}
