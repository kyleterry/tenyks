package irc

import (
	"testing"
)

func TestCommandEncoding(t *testing.T) {
	cases := []struct {
		cmd      Command
		expected string
	}{
		{
			cmd:      NewPassCommand("testpassword"),
			expected: "PASS testpassword\r\n",
		},
		{
			cmd:      NewUserCommand("test-user", 0, "Test User"),
			expected: "USER test-user 0 * :Test User\r\n",
		},
		{
			cmd:      NewNickCommand("test-nick"),
			expected: "NICK test-nick\r\n",
		},
		{
			cmd:      NewJoinCommand("#test-channel"),
			expected: "JOIN #test-channel\r\n",
		},
		{
			cmd:      NewJoinCommand("#test1", "&test2", "#test3"),
			expected: "JOIN #test1,&test2,#test3\r\n",
		},
		{
			cmd:      NewPrivmsgCommand("test-target", "hello, world!"),
			expected: "PRIVMSG test-target :hello, world!\r\n",
		},
		{
			cmd:      NewPrivmsgCommand("#test-channel", "hello, Channel!"),
			expected: "PRIVMSG #test-channel :hello, Channel!\r\n",
		},
	}

	for _, c := range cases {
		result, err := c.cmd.Encode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != c.expected {
			t.Errorf("got %q, want %q", result, c.expected)
		}
	}
}
