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
	Debug       bool
	Redis       RedisConfig
	Connections []ConnectionConfig
	LogLocation string
	Version     string
}

type RedisConfig struct {
	Host           string `json:"host","localhost"`
	Port           int
	Db             int
	Password       string
	TenyksChannel  string
	ServiceChannel string
}

type ConnectionConfig struct {
	Name            string
	Host            string
	Port            int
	Retries         int
	FloodProtection bool
	Password        string
	Nicks           []string
	Ident           string
	Realname        string
	Commands        []string
	Admins          []string
	Channels        []string
	Ssl             bool
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
	log.Debug("Loading configuration from %s", filename)
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
