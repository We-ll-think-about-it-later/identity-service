package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Client struct {
	*mongo.Client
	Ctx context.Context
}

func New(ctx context.Context, creds options.Credential, host string, port int) (*Client, error) {
	client, err := mongo.Connect(
		ctx,
		options.Client().
			ApplyURI(fmt.Sprintf("mongodb://%s:%d", host, port)).
			SetAuth(creds),
	)
	if err != nil {
		return nil, err
	}
	return &Client{client, ctx}, nil
}

func (m *Client) Disconnect() error {
	return m.Client.Disconnect(m.Ctx)
}
