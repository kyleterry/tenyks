package client

import (
	"context"
	"crypto/tls"

	servicepb "github.com/kyleterry/tenyks/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const DefaultAddr = "localhost:50001"

type Client struct {
	addr      string
	tlsConfig *tls.Config
	gclient   servicepb.MessageServiceClient
	conn      *grpc.ClientConn
}

func (c *Client) Dial(ctx context.Context) error {
	conn, err := grpc.NewClient(c.addr,
		grpc.WithTransportCredentials(credentials.NewTLS(c.tlsConfig)))
	if err != nil {
		return err
	}

	c.gclient = servicepb.NewMessageServiceClient(conn)
	c.conn = conn

	return nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}

	return nil
}

func NewClient(addr string, tlsConfig *tls.Config) *Client {
	if addr == "" {
		addr = DefaultAddr
	}

	return &Client{
		addr:      addr,
		tlsConfig: tlsConfig,
	}
}
