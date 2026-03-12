package client

import (
	"context"
	"crypto/tls"

	servicepb "github.com/kyleterry/tenyks/internal/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const DefaultAddr = "localhost:50001"

type grpcClient struct {
	addr      string
	tlsConfig *tls.Config
	gclient   servicepb.MessageServiceClient
	conn      *grpc.ClientConn
}

func (c *grpcClient) dial(ctx context.Context) error {
	conn, err := grpc.NewClient(c.addr,
		grpc.WithTransportCredentials(credentials.NewTLS(c.tlsConfig)))
	if err != nil {
		return err
	}

	c.gclient = servicepb.NewMessageServiceClient(conn)
	c.conn = conn

	return nil
}

func (c *grpcClient) stream(ctx context.Context) (servicepb.MessageService_StreamMessagesClient, error) {
	return c.gclient.StreamMessages(ctx)
}

func (c *grpcClient) close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func newGRPCClient(addr string, tlsConfig *tls.Config) *grpcClient {
	if addr == "" {
		addr = DefaultAddr
	}
	return &grpcClient{
		addr:      addr,
		tlsConfig: tlsConfig,
	}
}
