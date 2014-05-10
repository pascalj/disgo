package models

import (
	"code.google.com/p/gcfg"
)

// Config represents the user's config that's read from a gcfg file.
type Config struct {
	General struct {
		Approval  bool
		Origin    []string
		Markdown  bool
		Secret    string
		Templates string
		Prefix    string
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

// LoadConfig loads the config from disc and outputs an error if the file could no be read.
func LoadConfig(path string) (Config, error) {
	var cfg Config
	err := gcfg.ReadFileInto(&cfg, path)
	return cfg, err
}
