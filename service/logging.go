package service

import (
	"github.com/inconshreveable/log15"
)

var (
	Logger log15.Logger
)

func init() {
	Logger = log15.New()
	Logger.SetHandler(log15.StdoutHandler)
}
