package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	BaseURL  string `json:"base_url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func LoadConfig() (*Config, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	configPath := filepath.Join(configDir, "linksaver", "settings.json")

	configData, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				BaseURL: "http://localhost:8080",
			}, nil
		}
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func (c *Config) Save() error {
	configData, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	dirPath := filepath.Join(configDir, "linksaver")
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return err
	}
	configPath := filepath.Join(dirPath, "settings.json")

	return os.WriteFile(configPath, configData, 0700)
}
