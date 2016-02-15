package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("tenyks")

type configPaths struct {
	paths []string
}

var ConfigSearch configPaths

func (self *configPaths) AddPath(path string) {
	ConfigSearch.paths = append(ConfigSearch.paths, path)
}

type Config struct {
	Debug       bool               `json:"debug"`
	Service     ServiceConfig      `json:"service"`
	Connections []ConnectionConfig `json:"connections"`
	Control     ControlConfig      `json:"control"`
	LogLocation string             `json:"log_location"`
	Version     string
}

// TODO(kt) look into zmq channels later
type ServiceConfig struct {
	SenderBind   string `json:"sender_bind"`
	ReceiverBind string `json:"receiver_bind"`
}

type ConnectionConfig struct {
	Name            string   `json:"name"`
	Host            string   `json:"host"`
	Port            int      `json:"port"`
	Retries         int      `json:"retries"`
	FloodProtection bool     `json:"flood_protection"`
	Password        string   `json:"password"`
	Nicks           []string `json:"nicks"`
	Ident           string   `json:"ident"`
	Realname        string   `json:"real_name"`
	Commands        []string `json:"commands"`
	Admins          []string `json:"admins"`
	Channels        []string `json:"channels"`
	Ssl             bool     `json:"ssl"`
}

type ControlConfig struct {
	Enabled bool   `json:"enabled"`
	Bind    string `json:"bind"`
}

// discoverConfig will check to see if a config has been passed to tenyks on
// the command line or it will iterate over ConfigSearch paths and look for a
// config in the paths made with *configPaths.AddPath().
// It will return a string of either the path found to have a config or "".
func discoverConfig() string {
	for _, path := range ConfigSearch.paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

func NewConfigAutoDiscover(configPath *string) (conf *Config, err error) {
	var filename string
	if *configPath == "" {
		filename = discoverConfig()
	} else {
		filename = *configPath
	}
	if filename == "" {
		return nil, errors.New("No configuration file found.")
	}
	log.Info("Loading configuration from %s", filename)
	input, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return NewConfig(input)
}

func NewConfig(input []byte) (conf *Config, err error) {
	conf = new(Config)
	jsonerr := json.Unmarshal(input, &conf)
	err = nil
	if jsonerr != nil {
		err = jsonerr
	}
	return
}
