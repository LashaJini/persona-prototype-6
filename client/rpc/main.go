package rpcclient

import (
	"context"
	"fmt"

	pb "github.com/wholesome-ghoul/persona-prototype-6/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	client pb.ContentClient
}

func NewClient(port int) *Client {
	addr := fmt.Sprintf("localhost:%d", port)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, _ := grpc.Dial(addr, opts...)
	client := pb.NewContentClient(conn)

	return &Client{client}
}

func (c *Client) Personalize(ctx context.Context, contents *pb.Contents) (*pb.ContentScheme, error) {
	return c.client.Personalize(ctx, contents)
}

func (c *Client) InsertContents(ctx context.Context, contents *pb.Contents) error {
	_, err := c.client.InsertContents(ctx, contents)
	return err
}

func (c *Client) Rollback(ctx context.Context, contentIDs *pb.ContentIDs) error {
	_, err := c.client.Rollback(ctx, contentIDs)
	return err
}

func (c *Client) SemanticSearch(ctx context.Context, search *pb.Search) (*pb.ContentIDs, error) {
	return c.client.SemanticSearch(ctx, search)
}

func (c *Client) CalculateSpamProbs(ctx context.Context, sentences *pb.Sentences) (*pb.SpamProbs, error) {
	return c.client.CalculateSpamProbs(ctx, sentences)
}
