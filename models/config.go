package models

import (
	"code.google.com/p/gcfg"
)

type Config struct {
	General struct {
		Approval bool
		Origin   []string
		Markdown bool
		Secret   string
	}
	Database struct {
		Driver string
		Access string
	}
	Rate_Limit struct {
		Enable       bool
		Max_Comments int64
		Seconds      int64
	}
	Email struct {
		Notify   bool
		From     string
		To       []string
		Username string
		Password string
		Host     string
		Port     int
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
