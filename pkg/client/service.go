package client

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/kyleterry/tenyks/internal/config"
	internalpb "github.com/kyleterry/tenyks/internal/pb"
	"github.com/kyleterry/tenyks/internal/tlsconfig"
)

// pbStream is the gRPC bidirectional stream type.
type pbStream = internalpb.MessageService_StreamMessagesClient

// Config holds the configuration for a tenyks service client.
type Config struct {
	// Name is the service identifier.
	Name string
	// Version is an optional version string.
	Version string
	// Description describes what this service does.
	Description string
	// HelpText is shown in response to help queries.
	HelpText string
	// Addr is the tenyks gRPC server address. Defaults to localhost:50001.
	Addr string
	// TLS holds the mTLS certificate paths.
	TLS TLSConfig
	// Debug enables debug-level logging to stderr.
	Debug bool
}

// TLSConfig holds file paths for mTLS certificates.
type TLSConfig struct {
	CAFile   string
	CertFile string
	KeyFile  string
}

// Service connects to tenyks and dispatches incoming messages to registered handlers.
type Service struct {
	config         *Config
	log            *slog.Logger
	handlers       []MsgHandler
	defaultHandler MsgHandler
}

// New creates a new Service with the given config.
func New(cfg *Config) *Service {
	level := slog.LevelWarn
	if cfg.Debug {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})).
		With("service", cfg.Name)

	s := &Service{config: cfg, log: logger}
	s.DefaultHandle(NoopHandler)
	return s
}

// Handle registers a MsgHandler. Panics if both DirectOnly and PrivateOnly are set.
func (s *Service) Handle(h MsgHandler) {
	if h.DirectOnly && h.PrivateOnly {
		panic("client: cannot set both DirectOnly and PrivateOnly on a handler")
	}
	s.handlers = append(s.handlers, h)
}

// DefaultHandle sets the handler called when no registered handler matches.
func (s *Service) DefaultHandle(h MsgHandler) {
	s.defaultHandler = h
}

// Run connects to tenyks and blocks until a signal or stream error occurs.
func (s *Service) Run() error {
	certs, err := tlsconfig.Load(&config.TLS{
		CAFile:         s.config.TLS.CAFile,
		CertFile:       s.config.TLS.CertFile,
		PrivateKeyFile: s.config.TLS.KeyFile,
	})
	if err != nil {
		return err
	}

	addr := s.config.Addr
	c := newGRPCClient(addr, tlsconfig.NewClientConfig(certs))
	if err := c.dial(context.Background()); err != nil {
		return err
	}
	defer c.close()

	stream, err := c.stream(context.Background())
	if err != nil {
		return err
	}

	s.log.Info("connected", "addr", addr)

	if err := stream.Send(&internalpb.Message{
		Payload: &internalpb.Message_Control{
			Control: &internalpb.Control{
				Type:        internalpb.Control_TYPE_REGISTER,
				Name:        s.config.Name,
				Description: s.config.Description,
				HelpText:    s.config.HelpText,
			},
		},
	}); err != nil {
		return err
	}

	recvErrs := make(chan error, 1)
	go func() {
		for {
			pbMsg, err := stream.Recv()
			if err != nil {
				recvErrs <- err
				return
			}
			if pbMsg.GetChat() == nil {
				continue
			}
			msg := messageFromPB(pbMsg)
			s.log.Debug("message received", "nick", msg.Nick, "target", msg.Target, "payload", msg.Payload)
			s.dispatch(stream, msg)
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-sigs:
	case err := <-recvErrs:
		if err != io.EOF {
			s.log.Error("stream error", "err", err)
		}
	}

	return nil
}

func (s *Service) dispatch(stream pbStream, msg Message) {
	com := &Communication{stream: stream}

	for _, h := range s.handlers {
		if h.DirectOnly && !msg.Direct && !msg.Mention {
			continue
		}
		if h.PrivateOnly && msg.FromChannel {
			continue
		}
		if h.MatcherFunc == nil {
			s.log.Debug("dispatching message", "nick", msg.Nick, "target", msg.Target, "handler", "catch-all")
			go h.MatchHandler.HandleMatch(nil, msg, com)
			return
		}
		res := h.MatcherFunc.Match(msg)
		if res == nil {
			continue
		}
		s.log.Debug("dispatching message", "nick", msg.Nick, "target", msg.Target, "handler", "matched")
		go h.MatchHandler.HandleMatch(res, msg, com)
		return
	}

	s.log.Debug("dispatching message", "nick", msg.Nick, "target", msg.Target, "handler", "default")
	go s.defaultHandler.MatchHandler.HandleMatch(nil, msg, com)
}
