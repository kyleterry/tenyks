package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Unmarshaler interface {
	UnmarshalConfig(b []byte) error
}

type Config struct {
	Logging *Logging  `json:"logging"`
	Servers []*Server `json:"servers"`
	Service *Service  `json:"service"`
}

type Logging struct {
	Debug    bool   `json:"debug"`
	Location string `json:"location"`
}

type Server struct {
	Kind   string `json:"kind"`
	Name   string `json:"name"`
	Config any    `json:"config"`
}

func (sc *Server) UnmarshalJSON(b []byte) error {
	var j map[string]*json.RawMessage
	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}

	var name string
	if err := json.Unmarshal(*j["name"], &name); err != nil {
		return err
	}

	sc.Name = name

	var kind string
	if err := json.Unmarshal(*j["kind"], &kind); err != nil {
		return err
	}

	sc.Kind = kind

	switch sc.Kind {
	case "irc":
		var irc IRCServer

		if err := json.Unmarshal(*j["config"], &irc); err != nil {
			return err
		}

		irc.Name = name

		sc.Config = irc
	default:
		return fmt.Errorf("no such server kind %s", sc.Kind)
	}

	return nil
}

type IRCServer struct {
	Name       string   `json:"-"`
	ServerAddr string   `json:"server_addr"`
	Password   string   `json:"password"`
	Nicks      []string `json:"nicks"`
	User       string   `json:"user"`
	RealName   string   `json:"real_name"`
	Channels   []string `json:"channels"`
	Commands   []string `json:"commands"`
	UseTLS     bool     `json:"use_tls"`
	RootCAPath string   `json:"root_ca"`
}

// TLS holds the file paths to certificate PEM blocks used to configure mTLS.
type TLS struct {
	// CAFile is the path to the CA certificate file in PEM format. This CA must
	// be the authority that signed the client certificates.
	CAFile string `json:"ca_file"`
	// CertFile is the path to the server certificate file in PEM format.
	CertFile string `json:"cert_file"`
	// PrivateKeyFile is the path to the server private key file in PEM format.
	PrivateKeyFile string `json:"private_key_file"`
}

type Service struct {
	BindAddr string `json:"bind_addr"`
	TLS      *TLS   `json:"tls"`
}

func NewFromFile(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config

	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
