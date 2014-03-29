package main

import (
	"code.google.com/p/gcfg"
)

type Config struct {
	General struct {
		Approval bool
		Origin   []string
	}
}

func LoadConfig() Config {
	var cfg Config
	err := gcfg.ReadFileInto(&cfg, "disgo.gcfg")
	if err != nil {
		panic(err)
	}
	return cfg
}
