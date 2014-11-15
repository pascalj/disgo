package models

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config represents the user's config that's read from a gcfg file.
type Config struct {
	General struct {
		Approval  bool     `json:"approval"`
		Origin    []string `json:"origin"`
		Markdown  bool     `json:"markdown"`
		Secret    string   `json:"secret"`
		Templates string   `json:"templates"`
		Prefix    string   `json:"prefix"`
	} `json:"general"`
	Database struct {
		Driver   string
		Host     string
		Port     string
		Username string
		Password string
		Database string
		Path     string
	} `json:"-"`
	Rate_Limit struct {
		Enable       bool  `json:"enable"`
		Max_Comments int64 `json:"maxComments"`
		Seconds      int64 `json:"seconds"`
	} `json:"rateLimit"`
	Email struct {
		Notify   bool     `json:"notify"`
		From     string   `json:"from"`
		To       []string `json:"to"`
		Username string   `json:"username"`
		Password string   `json:"-"`
		Host     string   `json:"host"`
		Port     int      `json:"port"`
	} `json:"email"`
}

// LoadConfig loads the config from disc and outputs an error if the file could no be read.
func LoadConfig(path string) (Config, error) {
	var cfg Config
	configFile, err := ioutil.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	err = yaml.Unmarshal(configFile, &cfg)
	return cfg, err
}

// func (cfg *Config) UnmarshalJSON(b []byte) error {
// 	return json.Unmarshal(b, &d.jsonData)
// }
