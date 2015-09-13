package main

import (
	"io/ioutil"
	"log"
	"os"
	"sync"

	"gopkg.in/yaml.v1"
)

const defaultConfig = `
---
hosts:
`

// example.org -> go-import example.org/foo git https://github.com/example/foo
// example.org -> go-import example.org/bar hg  https://code.google.com/p/bar
type Host struct {
	Imports  []Import `yaml:"imports"`
	Default  Import   `yaml:"default"`
	Defaults []Import `yaml:"defaults"`

	mutex     *sync.Mutex
	generated []Import
}

type Import struct {
	Prefix string `yaml:"prefix"`
	VCS    string `yaml:"vcs"`
	URL    string `yaml:"url"`
	Docs   string `yaml:"docs"`
	Source string `yaml:"source"`
}

type Config struct {
	Hosts map[string]Host `yaml:"hosts"`
}

func loadConfiguration(cfgFile string) Config {
	var err error
	var buf []byte

	if _, err := os.Stat(cfgFile); err == nil {
		log.Println("Loading config from ", cfgFile)
		buf, err = ioutil.ReadFile(cfgFile)
		if err != nil {
			log.Println("Could not read config from ", cfgFile)
			buf = []byte(defaultConfig)
		}
	} else {
		log.Println("Loading default config, due to error ", err)
		buf = []byte(defaultConfig)
	}

	var cfg Config
	err = yaml.Unmarshal(buf, &cfg)
	if err != nil {
		log.Panic("Could not load config file", err)
	}

	return cfg
}
