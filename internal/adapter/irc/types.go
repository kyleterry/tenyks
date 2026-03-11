package irc

// CommandType represents an IRC command
type CommandType int

// Incomplete list of IRC commands
const (
	CommandTypeUser CommandType = iota
	CommandTypeNick
	CommandTypeJoin
	CommandTypePart
	CommandTypePass
	CommandTypePrivmsg
	CommandTypePing
	CommandTypePong
	CommandTypeCTCP
	CommandTypeUnknown
)

var CommandTypeMapping = map[string]CommandType{
	"USER":    CommandTypeUser,
	"NICK":    CommandTypeNick,
	"JOIN":    CommandTypeJoin,
	"PART":    CommandTypePart,
	"PASS":    CommandTypePass,
	"PRIVMSG": CommandTypePrivmsg,
	"PING":    CommandTypePing,
	"PONG":    CommandTypePong,
	"CTCP":    CommandTypeCTCP,
}

// ReplyType represents a reply to a command. These can be successful replies
// or errors.
type ReplyType int

const (
	ReplyTypeWelcome ReplyType = iota
	ReplyTypeYourHost
	ReplyTypeCreated
	ReplyTypeMyInfo
	ReplyTypeBounce
	ReplyTypeNoTopic
	ReplyTypeTopic
	ReplyTypeNames
	ReplyTypeEndOfNames
	ReplyTypeErrNoSuchNick
	ReplyTypeErrErroneusNickname
	ReplyTypeErrNickInUse
)

var ReplyTypeMapping = map[string]ReplyType{
	"001": ReplyTypeWelcome,
	"002": ReplyTypeYourHost,
	"003": ReplyTypeCreated,
	"004": ReplyTypeMyInfo,
	"005": ReplyTypeBounce,
	"331": ReplyTypeNoTopic,
	"332": ReplyTypeTopic,
	"353": ReplyTypeNames,
	"366": ReplyTypeEndOfNames,
	"401": ReplyTypeErrNoSuchNick,
	"432": ReplyTypeErrErroneusNickname,
	"433": ReplyTypeErrNickInUse,
}
