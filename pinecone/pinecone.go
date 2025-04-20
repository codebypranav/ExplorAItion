package pinecone

import (
	"context"
	"os"

	"github.com/pinecone-io/go-pinecone/v3/pinecone"
)

func NewClient() (*pinecone.Client, error) {
	params := pinecone.NewClientParams{
		ApiKey: os.Getenv("PINECONE_API_KEY"),
	}
	return pinecone.NewClient(params)
}

func GetIndexConnection(ctx context.Context) (*pinecone.IndexConnection, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}
	idxDesc, err := client.DescribeIndex(ctx, os.Getenv("PINECONE_INDEX_NAME"))
	if err != nil {
		return nil, err
	}
	connParams := pinecone.NewIndexConnParams{
		Host:      idxDesc.Host,
		Namespace: "",
	}
	return client.Index(connParams)
}
