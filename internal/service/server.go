package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"sort"
	"strings"
	"sync"

	"github.com/kyleterry/tenyks/internal/adapter"
	"github.com/kyleterry/tenyks/internal/certutil"
	pb "github.com/kyleterry/tenyks/internal/pb"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type streamEntry struct {
	ch          chan *pb.Message
	perms       certutil.Permissions
	name        string
	description string
	helpText    string
}

// Server is the gRPC MessageService implementation. It fans incoming IRC
// messages out to every connected service client and routes messages received
// from service clients back to the IRC adapters.
type Server struct {
	pb.UnimplementedMessageServiceServer

	mu       sync.RWMutex
	streams  map[string]*streamEntry
	adapters adapter.Registry
}

// Broadcast sends msg to every connected service client whose certificate
// permissions allow the message's destination path. If the message is a
// mention that matches a built-in command, tenyks handles it directly and
// does not forward to service clients.
func (s *Server) Broadcast(msg *pb.Message) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	chat := msg.GetChat()
	if chat == nil {
		return
	}

	if chat.Mention {
		if _, after, ok := strings.Cut(msg.Content, ": "); ok {
			cmd := strings.TrimSpace(after)
			if s.handleBuiltinCmdLocked(cmd, chat) {
				return
			}
		}
	}

	for _, entry := range s.streams {
		if chat.DestinationPath != "" && !entry.perms.AllowsPath(chat.DestinationPath) {
			continue
		}
		select {
		case entry.ch <- msg:
		default:
			// drop rather than block if the client is slow
		}
	}
}

// handleBuiltinCmdLocked matches cmd against tenyks' built-in commands and
// replies via the IRC adapters. Returns true if the command was handled.
// Caller must hold s.mu.RLock().
func (s *Server) handleBuiltinCmdLocked(cmd string, chat *pb.Chat) bool {
	switch {
	case cmd == "list-services":
		var names []string
		for _, entry := range s.streams {
			if entry.name != "" {
				names = append(names, entry.name)
			}
		}
		sort.Strings(names)
		s.replyTo(chat.DestinationPath, "["+strings.Join(names, ", ")+"]")
		return true
	default:
		if svcName, ok := strings.CutPrefix(cmd, "help "); ok {
			for _, entry := range s.streams {
				if entry.name == svcName {
					text := entry.description
					if entry.helpText != "" {
						text += " | " + entry.helpText
					}
					s.replyTo(chat.DestinationPath, text)
					return true
				}
			}
			s.replyTo(chat.DestinationPath, "unknown service: "+svcName)
			return true
		}
	}
	return false
}

// replyTo sends a chat message to destPath via all registered IRC adapters.
func (s *Server) replyTo(destPath, text string) {
	msg := &pb.Message{
		Payload: &pb.Message_Chat{
			Chat: &pb.Chat{
				DestinationPath: destPath,
			},
		},
		Content:   text,
		CreatedAt: timestamppb.Now(),
	}
	for _, a := range s.adapters.GetAdaptersFor(adapter.AdapterTypeIRC) {
		if err := a.SendAsync(context.Background(), msg); err != nil {
			log.Printf("failed to send built-in reply to adapter %s: %v", a.GetName(), err)
		}
	}
}

func (s *Server) StreamMessages(stream pb.MessageService_StreamMessagesServer) error {
	ctx := stream.Context()

	p, ok := peer.FromContext(ctx)
	if !ok {
		return fmt.Errorf("failed to get peer from context")
	}

	perms, err := permsFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to read client permissions: %w", err)
	}

	id := p.Addr.String()
	entry := &streamEntry{
		ch:    make(chan *pb.Message, 16),
		perms: perms,
	}

	s.mu.Lock()
	s.streams[id] = entry
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.streams, id)
		close(entry.ch)
		s.mu.Unlock()
	}()

	if len(perms.Paths) == 0 {
		log.Printf("service client connected: %s (all paths)", id)
	} else {
		log.Printf("service client connected: %s (paths: %v)", id, perms.Paths)
	}

	// Fan outbound messages to this stream.
	sendErrs := make(chan error, 1)
	go func() {
		for msg := range entry.ch {
			if err := stream.Send(msg); err != nil {
				sendErrs <- err
				return
			}
		}
	}()

	// Read messages from the service client and route to adapters.
	for {
		select {
		case err := <-sendErrs:
			return err
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		msg, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		if ctrl := msg.GetControl(); ctrl != nil {
			if ctrl.Type == pb.Control_TYPE_REGISTER {
				s.mu.Lock()
				entry.name = ctrl.Name
				entry.description = ctrl.Description
				entry.helpText = ctrl.HelpText
				s.mu.Unlock()
				log.Printf("service registered: %s (%s)", ctrl.Name, id)
			}
			continue
		}

		for _, a := range s.adapters.GetAdaptersFor(adapter.AdapterTypeIRC) {
			if err := a.SendAsync(context.Background(), msg); err != nil {
				log.Printf("failed to route message to adapter %s: %v", a.GetName(), err)
			}
		}
	}
}

// permsFromContext extracts the client's Permissions from the TLS peer
// certificate embedded in the gRPC context.
func permsFromContext(ctx context.Context) (certutil.Permissions, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return certutil.Permissions{}, nil
	}
	tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok || len(tlsInfo.State.PeerCertificates) == 0 {
		return certutil.Permissions{}, nil
	}
	return certutil.DecodePerms(tlsInfo.State.PeerCertificates[0])
}

// New creates the gRPC server with mTLS and returns the service Server (for
// registering as an adapter handler) and the gRPC server (for Serve).
func New(tlsConfig *tls.Config, adapters adapter.Registry) (*Server, *grpc.Server) {
	srv := &Server{
		streams:  make(map[string]*streamEntry),
		adapters: adapters,
	}

	gs := grpc.NewServer(
		grpc.Creds(credentials.NewTLS(tlsConfig)),
	)

	pb.RegisterMessageServiceServer(gs, srv)

	return srv, gs
}
