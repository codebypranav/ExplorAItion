package embeddings

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
)

func GenerateEmbedding(ctx context.Context, client *openai.Client, input string) ([]float32, error) {
	resp, err := client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Model: openai.SmallEmbedding3,
		Input: []string{input},
	})
	if err != nil {
		return nil, err
	}
	return resp.Data[0].Embedding, nil
}
