package irc

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/kyleterry/tenyks/internal/logger"
)

func defaultLoginFunc(ctx context.Context, c *Connection) error {
	// TODO figure out how to create a blocking and async message sending API.
	// blocking can beused for sending messages (like PASS) and checking if there
	// is an error coming back from the server allowing us to respond to it immediately.

	if c.password != "" {
		passCmd := NewPassCommand(c.password)
		c.out <- passCmd
	}

	// TODO support MODE bitmasks: https://tools.ietf.org/html/rfc2812#page-11
	userCmd := NewUserCommand(c.user, 0, c.realName)
	c.out <- userCmd

	nickCmd := NewNickCommand(c.nicks[0])
	c.out <- nickCmd
	c.Status.CurrentNick = c.nicks[0]

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
				channelName := cmd.Message().Params[0]

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
	if err == io.EOF {
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

func defaultPrivmsgHandler(ctx context.Context, c *Connection, command Command) error {
	switch cmd := command.(type) {
	case *PrivmsgCommand:
		direct := logger.Param{Key: "directMessage", Value: cmd.IsDirect()}
		mention := logger.Param{Key: "mentionMessage", Value: cmd.IsMention()}
		c.log.Debug(cmd.Message().RawMsg, direct, mention)

		e := &tenyksChatMessageEncoder{}
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
