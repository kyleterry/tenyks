package version

import(
	"fmt"
)

const (
	TenyksVersion = "1.0"
)

var Info []string

func init() {
	Info = append(Info,
		fmt.Sprintf("I am Tenyks version %s", TenyksVersion),
		"I am written in the Go programming language",
		"You can find my code at https://github.com/kyleterry/tenyks")
}

func GetInfo() []string {
	return Info
}
