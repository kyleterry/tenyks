package irc

type Nick struct {
	Name string
}

type NickManager struct {
	nicks      []string
	deadletter []string
	current    string
}
