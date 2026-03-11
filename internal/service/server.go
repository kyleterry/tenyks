package service

import (
	"crypto/tls"
	"fmt"
	"io"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

type Server struct {
	UnimplementedMessageServiceServer
}

func (s *Server) StreamMessages(stream MessageService_StreamMessagesServer) error {
	ctx := stream.Context()

	client, ok := peer.FromContext(ctx)
	if !ok {
		return fmt.Errorf("failed to get peer from context")
	}

	_ = client

	for {
		_, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}
	}

	return nil
}

func New(tlsConfig *tls.Config) *grpc.Server {
	gs := grpc.NewServer(
		grpc.Creds(credentials.NewTLS(tlsConfig)),
	)

	RegisterMessageServiceServer(gs, &Server{})

	return gs
}
