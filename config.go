package main

import (
	"code.google.com/p/gcfg"
)

type Config struct {
	General struct {
		Approval bool
		Origin   []string
	}
	Rate_Limit struct {
		Enable       bool
		Max_Comments int64
		Seconds      int64
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
