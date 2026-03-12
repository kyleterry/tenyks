package irc

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/kyleterry/tenyks/internal/logger"
)

func defaultLoginFunc(ctx context.Context, c *Connection) error {
	if c.password != "" {
		// Start capability negotiation for SASL PLAIN. NICK and USER are sent
		// alongside CAP LS so the server can process them while we negotiate;
		// the server holds off on sending 001 until we send CAP END.
		c.out <- NewCAPLSCommand()
	}

	c.out <- NewNickCommand(c.nicks[0])
	c.Status.CurrentNick = c.nicks[0]

	// TODO support MODE bitmasks: https://tools.ietf.org/html/rfc2812#page-11
	c.out <- NewUserCommand(c.user, 0, c.realName)

	return nil
}

func defaultSASLCommandHandler(ctx context.Context, c *Connection, command Command) error {
	switch cmd := command.(type) {
	case *CAPCommand:
		switch cmd.Subcommand() {
		case "LS":
			if strings.Contains(cmd.Capabilities(), "sasl") {
				c.out <- NewCAPReqCommand("sasl")
			} else {
				c.log.Error("server does not support SASL", logger.Param{Key: "connection", Value: c.Name})
				c.out <- NewCAPEndCommand()
			}
		case "ACK":
			if strings.Contains(cmd.Capabilities(), "sasl") {
				c.out <- NewAuthenticateCommand("PLAIN")
			}
		case "NAK":
			c.log.Error("SASL capability rejected", logger.Param{Key: "connection", Value: c.Name})
			c.out <- NewCAPEndCommand()
		}
	case *AuthenticateCommand:
		if cmd.Payload() == "+" {
			// Server is ready — send base64(\0nick\0password)
			payload := base64.StdEncoding.EncodeToString(
				[]byte("\x00" + c.nicks[0] + "\x00" + c.password),
			)
			c.out <- NewAuthenticateCommand(payload)
		}
	}
	return nil
}

func defaultSASLReplyHandler(ctx context.Context, c *Connection, reply Reply) error {
	switch reply.(type) {
	case *SASLSuccessReply:
		c.log.Debug("SASL authentication successful", logger.Param{Key: "connection", Value: c.Name})
		c.out <- NewCAPEndCommand()
	case *SASLFailReply:
		c.log.Error("SASL authentication failed", logger.Param{Key: "connection", Value: c.Name})
		c.out <- NewCAPEndCommand()
	}
	return nil
}

func defaultJoinFunc(ctx context.Context, c *Connection) error {
	if len(c.channels) > 0 {
		channels := []string{}
		for channel := range c.channels {
			channels = append(channels, channel)
		}

		c.out <- NewJoinCommand(channels...)
	}

	return nil
}

func defaultConnectionStatusUpdater(ctx context.Context, c *Connection, reply Reply) error {
	switch reply.(type) {
	case *WelcomeReply:
		c.WithWriteLock(ctx, func(conn *Connection) {
			conn.Status.Connected = true
		})
		return defaultJoinFunc(ctx, c)
	}

	return nil
}

func defaultPingResponder(ctx context.Context, c *Connection, command Command) error {
	switch cmd := command.(type) {
	case *PingCommand:
		pongCmd := NewPongCommand(cmd.Message().Trail)

		c.out <- pongCmd
	}

	return nil
}

func defaultJoinChannelStatusUpdater(ctx context.Context, c *Connection, command Command) error {
	switch cmd := command.(type) {
	case *JoinCommand:
		msg := cmd.Message()

		if msg.Prefix != nil {
			if msg.Prefix.Nick == c.Status.CurrentNick {
				channelName := msg.Trail
				if len(msg.Params) > 0 {
					channelName = msg.Params[0]
				}

				c.WithWriteLock(ctx, func(conn *Connection) {
					if channel, ok := conn.channels[channelName]; ok {
						channel.Status.Status = ChannelStatusJoined
						conn.channels[channelName] = channel
					}
				})
			}
		}
	}

	return nil
}

// defaultChannelMemberUpdater finds RPL_NAMREPLY messages and updates our copy
// of a channel's member list.
func defaultChannelMemberUpdater(ctx context.Context, c *Connection, reply Reply) error {
	switch r := reply.(type) {
	case *NamesReply:
		channelName := r.Channel()
		names := r.Names()

		c.WithWriteLock(ctx, func(conn *Connection) {
			if channel, ok := conn.channels[channelName]; ok {
				for _, name := range names {
					channel.Status.Nicks[name] = &Nick{Name: name}
				}

				conn.channels[channelName] = channel
			}
		})
	case *EndOfNamesReply:
		members := []string{}
		channelName := r.Channel()

		c.WithReadLock(ctx, func(conn *Connection) {
			if channel, ok := conn.channels[channelName]; ok {
				for nick := range channel.Status.Nicks {
					members = append(members, nick)
				}
			}
		})

		logParam := logger.Param{Key: "members", Value: strings.Join(members, ", ")}
		c.log.Debug(fmt.Sprintf("%s", channelName), logParam)
	}

	return nil
}

func defaultErrorHandlerFunc(ctx context.Context, c *Connection, err error) error {
	if errors.Is(err, io.EOF) {
		return c.Close(ctx)
	}

	return nil
}

func defaultUnknownHandler(ctx context.Context, c *Connection, command Command) error {
	switch cmd := command.(type) {
	case *UnknownCommand:
		c.log.Debug("unknown command", logger.Param{Key: "msg", Value: cmd.Message().RawMsg})
	}

	return nil
}

func cleanupChannels(ctx context.Context, c *Connection) error {
	c.WithWriteLock(ctx, func(conn *Connection) {
		for channel := range conn.channels {
			conn.channels[channel] = NewChannel(channel)
		}
	})

	return nil
}

func defaultNoticeHandler(ctx context.Context, c *Connection, command Command) error {
	switch cmd := command.(type) {
	case *NoticeCommand:
		c.log.Debug(cmd.Message().RawMsg)
	}

	return nil
}

func defaultPrivmsgHandler(ctx context.Context, c *Connection, command Command) error {
	switch cmd := command.(type) {
	case *PrivmsgCommand:
		direct := logger.Param{Key: "directMessage", Value: cmd.IsDirect()}
		mention := logger.Param{Key: "mentionMessage", Value: cmd.IsMention()}
		c.log.Debug(cmd.Message().RawMsg, direct, mention)

		e := &tenyksChatMessageEncoder{serverName: c.Name}
		msg, err := e.Encode(cmd)
		if err != nil {
			return err
		}

		for _, h := range c.chatMessageHandlers {
			h(msg)
		}
	}

	return nil
}
