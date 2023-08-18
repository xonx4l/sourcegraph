package tst

import (
	"encoding/json"
	"os"
)

type CodeHost struct {
	Kind     string `json:"Kind"`
	Token    string `json:"Token"`
	Org      string `json:"Org"`
	URL      string `json:"URL"`
	User     string `json:"User"`
	Password string `json:"Password"`
}

type SourcegraphCfg struct {
	URL   string `json:"URL"`
	User  string `json:"User"`
	Token string `json:"Token"`
}

type Config struct {
	CodeHost    CodeHost       `json:"CodeHost"`
	Sourcegraph SourcegraphCfg `json:"Sourcegraph"`
}

func LoadConfig(filename string) (*Config, error) {
	var c Config

	fd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	if err := json.NewDecoder(fd).Decode(&c); err != nil {
		return nil, err
	}

	return &c, nil
}
