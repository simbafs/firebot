package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type ChatConfig struct {
	Source string `yaml:"source"`
	ChatID int64  `yaml:"chat_id"`
	URL    string `yaml:"url"`
}

type Config struct {
	Chats []ChatConfig `yaml:"chats"`
}

// LoadConfig reads the YAML config file from path (default: "./config.yaml").
// If CONFIG_PATH env var is set, it takes precedence.
func LoadConfig() (*Config, error) {
	path := Getenv("CONFIG_PATH", "./config.yaml")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
