package pinecone

import (
	"context"
	"os"
	"github.com/pinecone-io/go-pinecone/pinecone"
)

func NewClient() (*pinecone.Client, error) {
	apiKey := os.Getenv("PINECONE_API_KEY")
	env := os.Getenv("PINECONE_ENVIRONMENT")
	return pinecone.NewClient(apiKey, env)
}

func GetIndex(client *pinecone.Client) (*pinecone.Index, error) {
	return client.Index(os.Getenv("PINECONE_INDEX_NAME"))
}

