package rpcserver

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "github.com/wholesome-ghoul/persona-prototype-6/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const EXTERNAL_SERVER_PORT = 50052

type contentServer struct {
	pb.UnimplementedContentServer
}

func (c *contentServer) Pop(ctx context.Context, empty *pb.Empty) (*pb.ContentIDs, error) {
	contentIDs := pb.ContentIDs{Items: []string{"oi1", "oi2"}}
	return &contentIDs, nil
}

func newServer() *contentServer {
	s := &contentServer{}
	return s
}

func StartServer(port int) {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var opts []grpc.ServerOption
	opts = append(opts, grpc.Creds(insecure.NewCredentials()))

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterContentServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
