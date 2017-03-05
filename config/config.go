package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	consul "github.com/hashicorp/consul/api"
)

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
	Logging     LogConfig          `json:"logging"`
	Version     string
}

type LogConfig struct {
	Debug bool `json:"debug"`
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

func NewConfigFromConsulKey(key, address string) (*Config, error) {
	consulConfig := consul.DefaultConfig()
	consulConfig.Address = address
	client, err := consul.NewClient(consulConfig)
	if err != nil {
		return nil, err
	}

	kv := client.KV()

	pair, _, err := kv.Get(key, nil)
	if err != nil {
		return nil, err
	}

	if pair == nil {
		return nil, errors.New(fmt.Sprintf("No such consul key: %s", key))
	}

	return NewConfig(pair.Value)
}

func NewConfigAutoDiscover(configPath *string) (*Config, error) {
	var filename string

	if *configPath == "" {
		filename = discoverConfig()
	} else {
		filename = *configPath
	}

	if filename == "" {
		return nil, errors.New("No configuration file found.")
	}

	input, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return NewConfig(input)
}

func NewConfig(input []byte) (*Config, error) {
	conf := new(Config)
	if err := json.Unmarshal(input, &conf); err != nil {
		return nil, err
	}

	return conf, nil
}
