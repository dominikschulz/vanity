package main

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/dominikschulz/vanity/server"
	"github.com/go-kit/kit/log"
)

const defaultConfig = `
---
hosts:
`

// example.org -> go-import example.org/foo git https://github.com/example/foo
// example.org -> go-import example.org/bar hg  https://code.google.com/p/bar

// Config holds the server configuration
type Config struct {
	Hosts map[string]*server.Host `yaml:"hosts"`
}

func loadConfiguration(l log.Logger, cfgFile string) (Config, error) {
	if l == nil {
		l = log.NewNopLogger()
	}

	var err error
	var buf []byte

	if _, err := os.Stat(cfgFile); err == nil {
		l.Log("level", "debug", "msg", "Loading config", "source", cfgFile)
		buf, err = ioutil.ReadFile(cfgFile)
		if err != nil {
			l.Log("level", "error", "msg", "Could not load config", "source", cfgFile)
			buf = []byte(defaultConfig)
		}
	} else {
		l.Log("level", "error", "msg", "Loading default config due to error", "err", err)
		buf = []byte(defaultConfig)
	}

	var cfg Config
	err = yaml.Unmarshal(buf, &cfg)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}
