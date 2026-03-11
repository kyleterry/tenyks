package irc

type ChannelStatusType int

const (
	ChannelStatusParted ChannelStatusType = iota
	ChannelStatusJoined
	ChannelStatusErr
)

// ChannelStatus is the current state of the channel. There are 3 possible
// states the channel can be in: parted, joined and error.
//
// Parted means the channel was never joined or it was left. In this case the
// Nicks map will be empty.
//
// Joined means the channel is active and the connection is a member. The Nicks
// map will have a current reflection of the other channel members.
//
// Err means the connection received an error reply when attempting to join, or
// the connection was banned or kicked at some point. The Nicks map will then
// be empty and a message about the error will be set on Message.
type ChannelStatus struct {
	Status  ChannelStatusType
	Message string
	Nicks   map[string]*Nick
}

type Channel struct {
	Name   string
	Status *ChannelStatus
}

func NewChannel(name string) *Channel {
	return &Channel{
		Name: name,
		Status: &ChannelStatus{
			Status: ChannelStatusParted,
			Nicks:  make(map[string]*Nick),
		},
	}
}
