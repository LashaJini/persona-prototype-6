package vdb

import (
	"context"
	"log"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
)

type ContentVDB struct {
	client.Client
}

func InitClient(ctx context.Context, collectionName string) *ContentVDB {
	client, err := client.NewClient(ctx, client.Config{
		Address: "localhost:19530",
	})

	if err != nil {
		log.Fatal(err)
	}

	// if exists, _ := client.HasCollection(ctx, collectionName); !exists {
	// 	schema := models.GetContentVDBSchema(collectionName)
	//
	// 	err = client.CreateCollection(
	// 		ctx,
	// 		schema,
	// 		2, // TODO: wtf is this
	// 	)
	//
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// }

	return &ContentVDB{client}
}
