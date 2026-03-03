package main

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func loadConfig() Config {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".soundsnatch.yaml")
	
	config := Config{
		LastSaveDir:   "",
		DefaultFormat: "mp3",
		Browser:       "",
	}

	data, err := os.ReadFile(configPath)
	if err == nil {
		yaml.Unmarshal(data, &config)
	}

	return config
}

func saveConfig(config Config) {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".soundsnatch.yaml")
	data, err := yaml.Marshal(config)
	if err == nil {
		os.WriteFile(configPath, data, 0644)
	}
}
