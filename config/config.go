package config

import (
	"io/ioutil"
	"os"
	"errors"
	"encoding/json"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("tenyks")

type Config struct {
	Debug       bool
	Redis       RedisConfig
	Connections []ConnectionConfig
	LogLocation string
}

type RedisConfig struct {
	Host          string `json:"host",localhost`
	Port          int
	Db            int
	Password      string
	TenyksPrefix  string
	ServicePrefix string
}

type ConnectionConfig struct {
	Name     string
	Host     string
	Port     int
	Retries  int
	Password string
	Nicks    []string
	Ident    string
	Realname string
	Commands []string
	Admins   []string
	Channels []string
	Ssl      bool
}

func discoverConfig() string {
	// TODO: This is temporary. Please refactor
	var filename string
	if len(os.Args) > 1 {
		filename = os.Args[1]
	} else {
		if _, err := os.Stat("/etc/tenyks/config.json"); err == nil {
			filename = "/etc/tenyks/config.json"
		} else if _, err := os.Stat(os.Getenv("HOME") + "/.config/tenyks/config.json"); err == nil {
			filename = os.Getenv("HOME") + "/.config/tenyks/config.json"
		} else {
			return ""
		}
	}
	return filename
}

func NewConfigAutoDiscover() (conf *Config, err error) {
	filename := discoverConfig()
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
	//fmt.Printf("%+v\n", conf)
	return
}
